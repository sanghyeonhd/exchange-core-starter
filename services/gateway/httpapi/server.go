// Package httpapi is the REST edge of the exchange. It converts decimal
// strings at the API boundary to fixed-point int64 internally, and maps
// core errors to the spec error format.
//
// Authentication is a development placeholder: private endpoints identify
// the caller by the X-USER-ID header. API key + HMAC signatures from the
// REST spec are pending the auth service integration.
package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/exchange-core-starter/exchange-core-starter/libs/decimal"
	"github.com/exchange-core-starter/exchange-core-starter/libs/ws"
	"github.com/exchange-core-starter/exchange-core-starter/services/account/account"
	"github.com/exchange-core-starter/exchange-core-starter/services/compliance/readiness"
	"github.com/exchange-core-starter/exchange-core-starter/services/market-data/marketdata"
	"github.com/exchange-core-starter/exchange-core-starter/services/matching-engine/engine"
	"github.com/exchange-core-starter/exchange-core-starter/services/oms/oms"
	"github.com/exchange-core-starter/exchange-core-starter/services/oms/spotexchange"
	"github.com/exchange-core-starter/exchange-core-starter/services/wallet-gateway/wallet"
)

type Deps struct {
	Exchange    *spotexchange.Exchange
	Wallet      *wallet.Coordinator
	MarketData  *marketdata.Hub
	PrivateData *marketdata.PrivateHub
	Readiness   readiness.Status
}

type Server struct {
	exchange    *spotexchange.Exchange
	wallet      *wallet.Coordinator
	marketData  *marketdata.Hub
	privateData *marketdata.PrivateHub
	status      readiness.Status
}

func NewServer(deps Deps) *Server {
	return &Server{
		exchange:    deps.Exchange,
		wallet:      deps.Wallet,
		marketData:  deps.MarketData,
		privateData: deps.PrivateData,
		status:      deps.Readiness,
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", s.health)
	mux.HandleFunc("GET /api/v1/time", s.serverTime)
	mux.HandleFunc("GET /api/v1/markets", s.markets)
	mux.HandleFunc("GET /api/v1/orderbook/{symbol}", s.orderbook)
	mux.HandleFunc("GET /api/v1/trades/{symbol}", s.trades)
	mux.HandleFunc("GET /api/v1/readiness/spot-launch", s.spotLaunchReadiness)
	mux.HandleFunc("GET /api/v1/readiness/derivatives", s.derivativesReadiness)
	mux.HandleFunc("GET /ws/public", s.publicStream)
	mux.HandleFunc("GET /ws/private", s.authenticated(s.privateStream))

	mux.HandleFunc("POST /api/v1/orders", s.authenticated(s.placeOrder))
	mux.HandleFunc("DELETE /api/v1/orders/{order_id}", s.authenticated(s.cancelOrder))
	mux.HandleFunc("GET /api/v1/orders/{order_id}", s.authenticated(s.getOrder))
	mux.HandleFunc("GET /api/v1/open-orders", s.authenticated(s.openOrders))
	mux.HandleFunc("GET /api/v1/account/balances", s.authenticated(s.balances))
	mux.HandleFunc("GET /api/v1/wallet/deposit-address/{asset}", s.authenticated(s.depositAddress))
	mux.HandleFunc("POST /api/v1/wallet/withdraw", s.authenticated(s.withdraw))

	return withCORS(withJSON(mux))
}

// --- public ---

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) serverTime(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]int64{"server_time_ms": time.Now().UnixMilli()})
}

type marketView struct {
	Symbol        string `json:"symbol"`
	BaseAsset     string `json:"base_asset"`
	QuoteAsset    string `json:"quote_asset"`
	Type          string `json:"type"`
	Status        string `json:"status"`
	PriceScale    int32  `json:"price_scale"`
	QuantityScale int32  `json:"quantity_scale"`
	TickSize      string `json:"tick_size"`
	LotSize       string `json:"lot_size"`
	MinOrderQty   string `json:"min_order_qty"`
	MinNotional   string `json:"min_notional"`
	MakerFeePPM   int64  `json:"maker_fee_ppm"`
	TakerFeePPM   int64  `json:"taker_fee_ppm"`
}

func (s *Server) markets(w http.ResponseWriter, r *http.Request) {
	if s.exchange == nil {
		writeJSON(w, http.StatusOK, map[string][]marketView{"markets": {}})
		return
	}
	m := s.exchange.Market()
	writeJSON(w, http.StatusOK, map[string][]marketView{"markets": {{
		Symbol:        m.Symbol,
		BaseAsset:     m.BaseAsset,
		QuoteAsset:    m.QuoteAsset,
		Type:          "SPOT",
		Status:        string(m.Status),
		PriceScale:    m.PriceScale,
		QuantityScale: m.QuantityScale,
		TickSize:      decimal.New(m.TickSize, m.PriceScale).String(),
		LotSize:       decimal.New(m.LotSize, m.QuantityScale).String(),
		MinOrderQty:   decimal.New(m.MinOrderQty, m.QuantityScale).String(),
		MinNotional:   decimal.New(m.MinNotional, m.PriceScale).String(),
		MakerFeePPM:   m.MakerFeePPM,
		TakerFeePPM:   m.TakerFeePPM,
	}}})
}

type levelView struct {
	Price    string `json:"price"`
	Quantity string `json:"quantity"`
}

func (s *Server) orderbook(w http.ResponseWriter, r *http.Request) {
	market, ok := s.requireMarket(w, r)
	if !ok {
		return
	}
	limit := queryInt(r, "limit", 50)
	bids, asks := s.exchange.Depth(limit)
	writeJSON(w, http.StatusOK, map[string]any{
		"symbol": market.Symbol,
		"bids":   levelViews(bids, market),
		"asks":   levelViews(asks, market),
	})
}

type tradeView struct {
	Sequence  int64  `json:"sequence"`
	Symbol    string `json:"symbol"`
	Price     string `json:"price"`
	Quantity  string `json:"quantity"`
	TakerSide string `json:"taker_side"`
}

func (s *Server) trades(w http.ResponseWriter, r *http.Request) {
	market, ok := s.requireMarket(w, r)
	if !ok {
		return
	}
	limit := queryInt(r, "limit", 100)
	trades := s.exchange.Trades(limit)
	views := make([]tradeView, 0, len(trades))
	for _, trade := range trades {
		views = append(views, tradeView{
			Sequence:  trade.Sequence,
			Symbol:    trade.Symbol,
			Price:     decimal.New(trade.Price, market.PriceScale).String(),
			Quantity:  decimal.New(trade.Quantity, market.QuantityScale).String(),
			TakerSide: string(trade.Side),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"symbol": market.Symbol, "trades": views})
}

// publicStream streams all public channels (trades, orderbook snapshots)
// over WebSocket per docs/api/WEBSOCKET_SPEC.md. The MVP has no
// subscription protocol: every public message is delivered. On connect the
// client receives a fresh orderbook snapshot so deltas-after-snapshot
// sequencing holds from the first message.
func (s *Server) publicStream(w http.ResponseWriter, r *http.Request) {
	if s.marketData == nil || s.exchange == nil {
		writeError(w, r, http.StatusServiceUnavailable, "MARKET_DATA_UNAVAILABLE", "market data feed not configured")
		return
	}
	conn, err := ws.Accept(w, r)
	if err != nil {
		return
	}
	defer conn.Close()

	subscription := s.marketData.Subscribe(256)
	defer subscription.Cancel()

	// Initial snapshot directly from the book, outside the hub sequence
	// (sequence 0 marks it as pre-stream state).
	market := s.exchange.Market()
	bids, asks := s.exchange.Depth(50)
	initial, err := json.Marshal(marketdata.Envelope{
		Channel:   marketdata.ChannelOrderbook,
		Symbol:    market.Symbol,
		Sequence:  0,
		Timestamp: time.Now().UnixMilli(),
		Data: marketdata.OrderbookData{
			Type: "snapshot",
			Bids: toMarketDataLevels(bids, market),
			Asks: toMarketDataLevels(asks, market),
		},
	})
	if err != nil {
		return
	}
	if err := conn.WriteText(initial); err != nil {
		return
	}

	// Reader goroutine: consume client frames so pings and close are
	// handled; any inbound error tears the stream down.
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	for {
		select {
		case envelope, ok := <-subscription.C:
			if !ok {
				// Dropped as a slow consumer; client must reconnect.
				return
			}
			payload, err := json.Marshal(envelope)
			if err != nil {
				return
			}
			if err := conn.WriteText(payload); err != nil {
				return
			}
		case <-done:
			return
		}
	}
}

// privateStream streams all private channels (orders, fills, balances)
// over WebSocket per docs/api/WEBSOCKET_SPEC.md. 
func (s *Server) privateStream(w http.ResponseWriter, r *http.Request, userID int64) {
	if s.privateData == nil {
		writeError(w, r, http.StatusServiceUnavailable, "PRIVATE_DATA_UNAVAILABLE", "private data feed not configured")
		return
	}
	conn, err := ws.Accept(w, r)
	if err != nil {
		return
	}
	defer conn.Close()

	subscription := s.privateData.Subscribe(userID, 256)
	defer subscription.Cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	for {
		select {
		case envelope, ok := <-subscription.C:
			if !ok {
				return
			}
			payload, err := json.Marshal(envelope)
			if err != nil {
				return
			}
			if err := conn.WriteText(payload); err != nil {
				return
			}
		case <-done:
			return
		}
	}
}

func toMarketDataLevels(levels []engine.Level, market oms.Market) []marketdata.PriceLevel {
	out := make([]marketdata.PriceLevel, 0, len(levels))
	for _, level := range levels {
		out = append(out, marketdata.PriceLevel{
			Price:    decimal.New(level.Price, market.PriceScale).String(),
			Quantity: decimal.New(level.Quantity, market.QuantityScale).String(),
		})
	}
	return out
}

func (s *Server) spotLaunchReadiness(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, readiness.Evaluate(readiness.ScopeSpotLaunch, s.status))
}

func (s *Server) derivativesReadiness(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, readiness.Evaluate(readiness.ScopeDerivatives, s.status))
}

// --- private ---

type placeOrderBody struct {
	ClientOrderID string `json:"client_order_id"`
	Symbol        string `json:"symbol"`
	Side          string `json:"side"`
	Type          string `json:"type"`
	TimeInForce   string `json:"time_in_force"`
	Price         string `json:"price"`
	Quantity      string `json:"quantity"`
}

type orderView struct {
	OrderID          int64  `json:"order_id"`
	ClientOrderID    string `json:"client_order_id,omitempty"`
	Symbol           string `json:"symbol"`
	Side             string `json:"side"`
	Type             string `json:"type"`
	Price            string `json:"price"`
	Quantity         string `json:"quantity"`
	ExecutedQuantity string `json:"executed_quantity"`
	Status           string `json:"status"`
}

func (s *Server) placeOrder(w http.ResponseWriter, r *http.Request, userID int64) {
	if s.exchange == nil {
		writeError(w, r, http.StatusServiceUnavailable, "TRADING_UNAVAILABLE", "trading core not configured")
		return
	}
	var body placeOrderBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	market := s.exchange.Market()
	price, err := decimal.Parse(body.Price, market.PriceScale)
	if err != nil && body.Type != string(engine.Market) {
		writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "invalid price")
		return
	}
	quantity, err := decimal.Parse(body.Quantity, market.QuantityScale)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "invalid quantity")
		return
	}

	result, err := s.exchange.PlaceOrder(spotexchange.PlaceOrderRequest{
		UserID:        userID,
		ClientOrderID: body.ClientOrderID,
		Symbol:        body.Symbol,
		Side:          engine.Side(body.Side),
		Type:          engine.OrderType(body.Type),
		TimeInForce:   engine.TimeInForce(body.TimeInForce),
		Price:         price.Value,
		Quantity:      quantity.Value,
	})
	if err != nil {
		writeOrderError(w, r, err)
		return
	}
	status := http.StatusCreated
	if result.Duplicate {
		status = http.StatusOK
	}
	writeJSON(w, status, map[string]any{
		"order":  s.orderView(result.Order, market),
		"trades": len(result.Trades),
	})
}

func (s *Server) cancelOrder(w http.ResponseWriter, r *http.Request, userID int64) {
	orderID, ok := s.pathOrderID(w, r)
	if !ok {
		return
	}
	order, err := s.exchange.CancelOrder(userID, orderID)
	if err != nil {
		writeOrderError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"order": s.orderView(order, s.exchange.Market())})
}

func (s *Server) getOrder(w http.ResponseWriter, r *http.Request, userID int64) {
	orderID, ok := s.pathOrderID(w, r)
	if !ok {
		return
	}
	order, err := s.exchange.Order(userID, orderID)
	if err != nil {
		writeOrderError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"order": s.orderView(order, s.exchange.Market())})
}

func (s *Server) openOrders(w http.ResponseWriter, r *http.Request, userID int64) {
	if s.exchange == nil {
		writeError(w, r, http.StatusServiceUnavailable, "TRADING_UNAVAILABLE", "trading core not configured")
		return
	}
	market := s.exchange.Market()
	orders := s.exchange.OpenOrders(userID)
	views := make([]orderView, 0, len(orders))
	for _, order := range orders {
		views = append(views, s.orderView(order, market))
	}
	writeJSON(w, http.StatusOK, map[string]any{"orders": views})
}

type balanceView struct {
	Asset     string `json:"asset"`
	Available string `json:"available"`
	Locked    string `json:"locked"`
}

func (s *Server) balances(w http.ResponseWriter, r *http.Request, userID int64) {
	if s.exchange == nil {
		writeError(w, r, http.StatusServiceUnavailable, "TRADING_UNAVAILABLE", "trading core not configured")
		return
	}
	market := s.exchange.Market()
	balances := s.exchange.Balances(userID)
	views := make([]balanceView, 0, len(balances))
	for _, b := range balances {
		scale := assetScale(market, b.Asset)
		views = append(views, balanceView{
			Asset:     b.Asset,
			Available: decimal.New(b.Available, scale).String(),
			Locked:    decimal.New(b.Locked, scale).String(),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"balances": views})
}

func (s *Server) depositAddress(w http.ResponseWriter, r *http.Request, userID int64) {
	if s.wallet == nil {
		writeError(w, r, http.StatusServiceUnavailable, "WALLET_UNAVAILABLE", "wallet gateway not configured")
		return
	}
	asset := r.PathValue("asset")
	network := r.URL.Query().Get("network")
	if network == "" {
		network = "TESTNET"
	}
	address, err := s.wallet.CreateDepositAddress(r.Context(), userID, asset, network)
	if err != nil {
		writeWalletError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"asset":   address.Asset,
		"network": address.Network,
		"address": address.Address,
		"memo":    address.Memo,
	})
}

type withdrawBody struct {
	ID      int64  `json:"id"`
	Asset   string `json:"asset"`
	Network string `json:"network"`
	Address string `json:"address"`
	Memo    string `json:"memo"`
	Amount  string `json:"amount"`
	Fee     string `json:"fee"`
}

func (s *Server) withdraw(w http.ResponseWriter, r *http.Request, userID int64) {
	if s.wallet == nil || s.exchange == nil {
		writeError(w, r, http.StatusServiceUnavailable, "WALLET_UNAVAILABLE", "wallet gateway not configured")
		return
	}
	var body withdrawBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	market := s.exchange.Market()
	scale := assetScale(market, body.Asset)
	amount, err := decimal.Parse(body.Amount, scale)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "invalid amount")
		return
	}
	fee := decimal.Decimal{}
	if body.Fee != "" {
		fee, err = decimal.Parse(body.Fee, scale)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "invalid fee")
			return
		}
	}
	withdrawal, err := s.wallet.RequestWithdrawal(wallet.WithdrawalRequest{
		ID:      body.ID,
		UserID:  userID,
		Asset:   body.Asset,
		Network: body.Network,
		Address: body.Address,
		Memo:    body.Memo,
		Amount:  amount.Value,
		Fee:     fee.Value,
	})
	if err != nil {
		writeWalletError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"id":     withdrawal.Request.ID,
		"status": string(withdrawal.Status),
	})
}

// --- helpers ---

type authenticatedHandler func(http.ResponseWriter, *http.Request, int64)

func (s *Server) authenticated(next authenticatedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := strconv.ParseInt(r.Header.Get("X-USER-ID"), 10, 64)
		if err != nil || userID <= 0 {
			writeError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "missing or invalid X-USER-ID header (development auth placeholder)")
			return
		}
		next(w, r, userID)
	}
}

func (s *Server) requireMarket(w http.ResponseWriter, r *http.Request) (oms.Market, bool) {
	if s.exchange == nil {
		writeError(w, r, http.StatusServiceUnavailable, "TRADING_UNAVAILABLE", "trading core not configured")
		return oms.Market{}, false
	}
	market := s.exchange.Market()
	if symbol := r.PathValue("symbol"); symbol != "" && symbol != market.Symbol {
		writeError(w, r, http.StatusNotFound, "UNKNOWN_SYMBOL", "unknown symbol")
		return oms.Market{}, false
	}
	return market, true
}

func (s *Server) pathOrderID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	if s.exchange == nil {
		writeError(w, r, http.StatusServiceUnavailable, "TRADING_UNAVAILABLE", "trading core not configured")
		return 0, false
	}
	orderID, err := strconv.ParseInt(r.PathValue("order_id"), 10, 64)
	if err != nil || orderID <= 0 {
		writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "invalid order id")
		return 0, false
	}
	return orderID, true
}

func (s *Server) orderView(order engine.Order, market oms.Market) orderView {
	return orderView{
		OrderID:          order.ID,
		Symbol:           order.Symbol,
		Side:             string(order.Side),
		Type:             string(order.Type),
		Price:            decimal.New(order.Price, market.PriceScale).String(),
		Quantity:         decimal.New(order.Quantity, market.QuantityScale).String(),
		ExecutedQuantity: decimal.New(order.ExecutedQuantity, market.QuantityScale).String(),
		Status:           string(order.Status),
	}
}

func assetScale(market oms.Market, asset string) int32 {
	if asset == market.BaseAsset {
		return market.QuantityScale
	}
	return market.PriceScale
}

func levelViews(levels []engine.Level, market oms.Market) []levelView {
	views := make([]levelView, 0, len(levels))
	for _, level := range levels {
		views = append(views, levelView{
			Price:    decimal.New(level.Price, market.PriceScale).String(),
			Quantity: decimal.New(level.Quantity, market.QuantityScale).String(),
		})
	}
	return views
}

func queryInt(r *http.Request, name string, fallback int) int {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func writeOrderError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, account.ErrInsufficientBalance):
		writeError(w, r, http.StatusBadRequest, "ORDER_REJECTED_INSUFFICIENT_BALANCE", "Insufficient available balance")
	case errors.Is(err, spotexchange.ErrOrderNotFound), errors.Is(err, engine.ErrOrderNotFound):
		writeError(w, r, http.StatusNotFound, "ORDER_NOT_FOUND", "Order not found")
	case errors.Is(err, spotexchange.ErrOrderNotOwned):
		// Do not leak other users' order existence.
		writeError(w, r, http.StatusNotFound, "ORDER_NOT_FOUND", "Order not found")
	case errors.Is(err, spotexchange.ErrOrderNotOpen):
		writeError(w, r, http.StatusConflict, "ORDER_NOT_OPEN", "Order is not open")
	case errors.Is(err, spotexchange.ErrUnknownSymbol):
		writeError(w, r, http.StatusNotFound, "UNKNOWN_SYMBOL", "unknown symbol")
	default:
		writeError(w, r, http.StatusBadRequest, "ORDER_REJECTED_VALIDATION", err.Error())
	}
}

func writeWalletError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, wallet.ErrWalletDisabled):
		writeError(w, r, http.StatusForbidden, "WALLET_DISABLED", err.Error())
	case errors.Is(err, wallet.ErrAddressNotWhitelisted):
		writeError(w, r, http.StatusForbidden, "ADDRESS_NOT_WHITELISTED", "withdrawal address is not whitelisted")
	case errors.Is(err, account.ErrInsufficientBalance):
		writeError(w, r, http.StatusBadRequest, "WITHDRAWAL_INSUFFICIENT_BALANCE", "Insufficient available balance")
	default:
		writeError(w, r, http.StatusBadRequest, "WALLET_REJECTED", err.Error())
	}
}

// withCORS is a development-only policy so the static product shells under
// apps/web and apps/admin can call the gateway from another origin. A
// production deployment must replace this with an explicit origin allowlist.
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-USER-ID, X-REQUEST-ID, X-API-KEY, X-SIGNATURE, X-TIMESTAMP")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func withJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, r *http.Request, status int, code string, message string) {
	writeJSON(w, status, map[string]any{
		"code":       code,
		"message":    message,
		"request_id": r.Header.Get("X-REQUEST-ID"),
		"details":    map[string]any{},
	})
}
