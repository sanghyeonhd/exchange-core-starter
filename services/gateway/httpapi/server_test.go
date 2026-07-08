package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/exchange-core-starter/exchange-core-starter/libs/audit"
	"github.com/exchange-core-starter/exchange-core-starter/libs/ws"
	"github.com/exchange-core-starter/exchange-core-starter/services/compliance/readiness"
	"github.com/exchange-core-starter/exchange-core-starter/services/market-data/marketdata"
	"github.com/exchange-core-starter/exchange-core-starter/services/oms/oms"
	"github.com/exchange-core-starter/exchange-core-starter/services/oms/spotexchange"
	"github.com/exchange-core-starter/exchange-core-starter/services/wallet-gateway/wallet"
)

func newTestServer(t *testing.T, status readiness.Status) *Server {
	t.Helper()
	exchange, err := spotexchange.New(oms.Market{
		Symbol:        "BTC-USDT",
		BaseAsset:     "BTC",
		QuoteAsset:    "USDT",
		Status:        oms.MarketTrading,
		PriceScale:    2,
		QuantityScale: 8,
		MinOrderQty:   100_000,
		MinNotional:   10_00,
		TickSize:      1_00,
		LotSize:       100_000,
		MakerFeePPM:   100,
		TakerFeePPM:   200,
	})
	if err != nil {
		t.Fatalf("spotexchange.New: %v", err)
	}
	if err := exchange.Deposit("seed-1-usdt", 1, "USDT", 1_000_000); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := exchange.Deposit("seed-2-btc", 2, "BTC", 100_000_000); err != nil {
		t.Fatalf("seed: %v", err)
	}

	hub := marketdata.NewHub()
	exchange.SetMarketData(hub)

	privateHub := marketdata.NewPrivateHub()
	exchange.SetPrivateData(privateHub)

	walletService := wallet.NewService(
		wallet.Config{MainnetEnabled: false, WithdrawalsEnabled: false},
		wallet.NewMockAdapter(),
		wallet.NewWhitelist(),
		audit.NewLog(),
	)
	return NewServer(Deps{
		Exchange:    exchange,
		Wallet:      wallet.NewCoordinator(walletService, exchange),
		MarketData:  hub,
		PrivateData: privateHub,
		Readiness:   status,
	})
}

func TestPublicWebSocketStream(t *testing.T) {
	server := newTestServer(t, readiness.Status{})
	httpServer := httptest.NewServer(server.Handler())
	defer httpServer.Close()

	conn, err := ws.Dial("ws" + strings.TrimPrefix(httpServer.URL, "http") + "/ws/public")
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer conn.Close()

	readEnvelope := func() marketdata.Envelope {
		t.Helper()
		_, payload, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("ReadMessage: %v", err)
		}
		var envelope marketdata.Envelope
		if err := json.Unmarshal(payload, &envelope); err != nil {
			t.Fatalf("decode %s: %v", payload, err)
		}
		return envelope
	}

	// Initial snapshot arrives immediately with sequence 0.
	initial := readEnvelope()
	if initial.Channel != marketdata.ChannelOrderbook || initial.Sequence != 0 {
		t.Fatalf("initial = %+v", initial)
	}

	// A resting order, then a matching order produce live events.
	request(t, server, http.MethodPost, "/api/v1/orders",
		`{"symbol":"BTC-USDT","side":"BUY","type":"LIMIT","price":"50000.00","quantity":"0.00100000"}`, 1)
	request(t, server, http.MethodPost, "/api/v1/orders",
		`{"symbol":"BTC-USDT","side":"SELL","type":"LIMIT","price":"50000.00","quantity":"0.00100000"}`, 2)

	afterRest := readEnvelope()
	if afterRest.Channel != marketdata.ChannelOrderbook || afterRest.Sequence != 1 {
		t.Fatalf("afterRest = %+v", afterRest)
	}
	trade := readEnvelope()
	if trade.Channel != marketdata.ChannelTrades || trade.Sequence != 1 {
		t.Fatalf("trade = %+v", trade)
	}
	afterMatch := readEnvelope()
	if afterMatch.Channel != marketdata.ChannelOrderbook || afterMatch.Sequence != 2 {
		t.Fatalf("afterMatch = %+v", afterMatch)
	}
}

func TestPrivateWebSocketStream(t *testing.T) {
	server := newTestServer(t, readiness.Status{})
	httpServer := httptest.NewServer(server.Handler())
	defer httpServer.Close()

	headers := http.Header{}
	headers.Add("X-USER-ID", "1")
	conn, err := ws.DialWithHeader("ws"+strings.TrimPrefix(httpServer.URL, "http")+"/ws/private", headers)
	if err != nil {
		t.Fatalf("DialWithHeader: %v", err)
	}
	defer conn.Close()

	readEnvelope := func() marketdata.Envelope {
		t.Helper()
		_, payload, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("ReadMessage: %v", err)
		}
		var envelope marketdata.Envelope
		if err := json.Unmarshal(payload, &envelope); err != nil {
			t.Fatalf("decode %s: %v", payload, err)
		}
		return envelope
	}

	request(t, server, http.MethodPost, "/api/v1/orders",
		`{"client_order_id":"a1","symbol":"BTC-USDT","side":"BUY","type":"LIMIT","price":"50000.00","quantity":"0.00100000"}`, 1)

	sawOrder := false
	sawBalance := false
	for i := 0; i < 5; i++ {
		env := readEnvelope()
		if env.Channel == marketdata.ChannelOrders {
			sawOrder = true
		} else if env.Channel == marketdata.ChannelBalances {
			sawBalance = true
		}
		if sawOrder && sawBalance {
			break
		}
	}
	if !sawOrder || !sawBalance {
		t.Fatalf("expected to see order and balance updates, got order=%v balance=%v", sawOrder, sawBalance)
	}
}

func TestHealthAndMarkets(t *testing.T) {
	server := newTestServer(t, readiness.Status{})

	health := request(t, server, http.MethodGet, "/api/v1/health", "", 0)
	if health.Code != http.StatusOK {
		t.Fatalf("health status = %d, want 200", health.Code)
	}

	markets := request(t, server, http.MethodGet, "/api/v1/markets", "", 0)
	if markets.Code != http.StatusOK {
		t.Fatalf("markets status = %d, want 200", markets.Code)
	}
	var body struct {
		Markets []struct {
			Symbol   string `json:"symbol"`
			TickSize string `json:"tick_size"`
		} `json:"markets"`
	}
	decode(t, markets, &body)
	if len(body.Markets) != 1 || body.Markets[0].Symbol != "BTC-USDT" || body.Markets[0].TickSize != "1.00" {
		t.Fatalf("markets = %+v", body.Markets)
	}
}

func TestPrivateEndpointsRequireUser(t *testing.T) {
	server := newTestServer(t, readiness.Status{})
	res := request(t, server, http.MethodGet, "/api/v1/account/balances", "", 0)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", res.Code)
	}
	var body struct {
		Code string `json:"code"`
	}
	decode(t, res, &body)
	if body.Code != "UNAUTHENTICATED" {
		t.Fatalf("code = %s", body.Code)
	}
}

func TestOrderFlowThroughAPI(t *testing.T) {
	server := newTestServer(t, readiness.Status{})

	place := request(t, server, http.MethodPost, "/api/v1/orders",
		`{"client_order_id":"a1","symbol":"BTC-USDT","side":"BUY","type":"LIMIT","time_in_force":"GTC","price":"50000.00","quantity":"0.00100000"}`, 1)
	if place.Code != http.StatusCreated {
		t.Fatalf("place status = %d body=%s", place.Code, place.Body.String())
	}
	var placed struct {
		Order struct {
			OrderID int64  `json:"order_id"`
			Status  string `json:"status"`
			Price   string `json:"price"`
		} `json:"order"`
	}
	decode(t, place, &placed)
	if placed.Order.Status != "ACCEPTED" || placed.Order.Price != "50000.00" {
		t.Fatalf("placed = %+v", placed)
	}

	// Duplicate client order id returns the existing order with 200.
	duplicate := request(t, server, http.MethodPost, "/api/v1/orders",
		`{"client_order_id":"a1","symbol":"BTC-USDT","side":"BUY","type":"LIMIT","time_in_force":"GTC","price":"50000.00","quantity":"0.00100000"}`, 1)
	if duplicate.Code != http.StatusOK {
		t.Fatalf("duplicate status = %d", duplicate.Code)
	}

	book := request(t, server, http.MethodGet, "/api/v1/orderbook/BTC-USDT", "", 0)
	var depth struct {
		Bids []struct {
			Price    string `json:"price"`
			Quantity string `json:"quantity"`
		} `json:"bids"`
	}
	decode(t, book, &depth)
	if len(depth.Bids) != 1 || depth.Bids[0].Price != "50000.00" || depth.Bids[0].Quantity != "0.00100000" {
		t.Fatalf("depth = %+v", depth)
	}

	fill := request(t, server, http.MethodPost, "/api/v1/orders",
		`{"client_order_id":"b1","symbol":"BTC-USDT","side":"SELL","type":"LIMIT","time_in_force":"GTC","price":"50000.00","quantity":"0.00100000"}`, 2)
	if fill.Code != http.StatusCreated {
		t.Fatalf("fill status = %d body=%s", fill.Code, fill.Body.String())
	}
	var filled struct {
		Order struct {
			Status string `json:"status"`
		} `json:"order"`
		Trades int `json:"trades"`
	}
	decode(t, fill, &filled)
	if filled.Order.Status != "FILLED" || filled.Trades != 1 {
		t.Fatalf("filled = %+v", filled)
	}

	trades := request(t, server, http.MethodGet, "/api/v1/trades/BTC-USDT", "", 0)
	var tradeList struct {
		Trades []struct {
			Price string `json:"price"`
		} `json:"trades"`
	}
	decode(t, trades, &tradeList)
	if len(tradeList.Trades) != 1 || tradeList.Trades[0].Price != "50000.00" {
		t.Fatalf("trades = %+v", tradeList)
	}

	balances := request(t, server, http.MethodGet, "/api/v1/account/balances", "", 1)
	var balanceBody struct {
		Balances []balanceView `json:"balances"`
	}
	decode(t, balances, &balanceBody)
	byAsset := map[string]balanceView{}
	for _, b := range balanceBody.Balances {
		byAsset[b.Asset] = b
	}
	// Buyer paid 50.00 notional; the 100 ppm maker fee on 50.00 floors to 0.
	if byAsset["USDT"].Available != "9950.00" || byAsset["BTC"].Available != "0.00100000" {
		t.Fatalf("balances = %+v", byAsset)
	}
}

func TestCancelOrderThroughAPI(t *testing.T) {
	server := newTestServer(t, readiness.Status{})

	place := request(t, server, http.MethodPost, "/api/v1/orders",
		`{"symbol":"BTC-USDT","side":"BUY","type":"LIMIT","price":"40000.00","quantity":"0.00100000"}`, 1)
	var placed struct {
		Order struct {
			OrderID int64 `json:"order_id"`
		} `json:"order"`
	}
	decode(t, place, &placed)

	// Another user cannot cancel it; existence is not leaked.
	other := request(t, server, http.MethodDelete, "/api/v1/orders/1", "", 2)
	if other.Code != http.StatusNotFound {
		t.Fatalf("other cancel status = %d", other.Code)
	}

	cancel := request(t, server, http.MethodDelete, "/api/v1/orders/1", "", 1)
	if cancel.Code != http.StatusOK {
		t.Fatalf("cancel status = %d body=%s", cancel.Code, cancel.Body.String())
	}

	balances := request(t, server, http.MethodGet, "/api/v1/account/balances", "", 1)
	var balanceBody struct {
		Balances []balanceView `json:"balances"`
	}
	decode(t, balances, &balanceBody)
	for _, b := range balanceBody.Balances {
		if b.Asset == "USDT" && (b.Available != "10000.00" || b.Locked != "0.00") {
			t.Fatalf("USDT after cancel = %+v", b)
		}
	}
}

func TestInsufficientBalanceErrorFormat(t *testing.T) {
	server := newTestServer(t, readiness.Status{})
	res := request(t, server, http.MethodPost, "/api/v1/orders",
		`{"symbol":"BTC-USDT","side":"BUY","type":"LIMIT","price":"50000.00","quantity":"1.00000000"}`, 1)
	if res.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", res.Code)
	}
	var body struct {
		Code string `json:"code"`
	}
	decode(t, res, &body)
	if body.Code != "ORDER_REJECTED_INSUFFICIENT_BALANCE" {
		t.Fatalf("code = %s", body.Code)
	}
}

func TestWithdrawalsDisabledByDefault(t *testing.T) {
	server := newTestServer(t, readiness.Status{})
	res := request(t, server, http.MethodPost, "/api/v1/wallet/withdraw",
		`{"id":1,"asset":"USDT","network":"TESTNET","address":"addr","amount":"10.00"}`, 1)
	if res.Code != http.StatusForbidden {
		t.Fatalf("status = %d body=%s", res.Code, res.Body.String())
	}
	var body struct {
		Code string `json:"code"`
	}
	decode(t, res, &body)
	if body.Code != "WALLET_DISABLED" {
		t.Fatalf("code = %s", body.Code)
	}
}

func TestReadinessEndpointsExposeBlockedDerivatives(t *testing.T) {
	server := newTestServer(t, readiness.Status{
		LegalApproved:           true,
		ComplianceApproved:      true,
		SecurityApproved:        true,
		ISMSReady:               true,
		VASPFilingReady:         true,
		AMLOperational:          true,
		KYCOperational:          true,
		ClosedPilotPassed:       true,
		IncidentDrillPassed:     true,
		RecoveryDrillPassed:     true,
		MarketSurveillanceReady: true,
	})

	spot := request(t, server, http.MethodGet, "/api/v1/readiness/spot-launch", "", 0)
	var spotDecision readiness.Decision
	decode(t, spot, &spotDecision)
	if !spotDecision.Allowed {
		t.Fatalf("spot decision = %+v, want allowed", spotDecision)
	}

	derivatives := request(t, server, http.MethodGet, "/api/v1/readiness/derivatives", "", 0)
	var derivativesDecision readiness.Decision
	decode(t, derivatives, &derivativesDecision)
	if derivativesDecision.Allowed {
		t.Fatalf("derivatives decision = %+v, want blocked", derivativesDecision)
	}
}

func request(t *testing.T, server *Server, method, path, body string, userID int64) *httptest.ResponseRecorder {
	t.Helper()
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	if userID > 0 {
		req.Header.Set("X-USER-ID", fmt.Sprintf("%d", userID))
	}
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	return rec
}

func decode(t *testing.T, rec *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.Unmarshal(rec.Body.Bytes(), target); err != nil {
		t.Fatalf("decode %s: %v", rec.Body.String(), err)
	}
}
