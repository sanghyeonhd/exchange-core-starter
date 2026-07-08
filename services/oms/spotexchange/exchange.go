// Package spotexchange composes the spot trading core: OMS validation,
// account reservation, orderbook matching, per-trade settlement, and
// balance projection. It is the in-process backend used by the gateway
// until services are split behind transports.
package spotexchange

import (
	"errors"
	"fmt"
	"sync"

	"github.com/exchange-core-starter/exchange-core-starter/libs/decimal"
	"github.com/exchange-core-starter/exchange-core-starter/services/account/account"
	"github.com/exchange-core-starter/exchange-core-starter/services/market-data/marketdata"
	"github.com/exchange-core-starter/exchange-core-starter/services/matching-engine/engine"
	"github.com/exchange-core-starter/exchange-core-starter/services/matching-engine/wal"
	"github.com/exchange-core-starter/exchange-core-starter/services/oms/oms"
	"github.com/exchange-core-starter/exchange-core-starter/services/settlement/funding"
	spotsettlement "github.com/exchange-core-starter/exchange-core-starter/services/settlement/spot"
)

var (
	ErrOrderNotFound  = errors.New("order not found")
	ErrOrderNotOwned  = errors.New("order does not belong to user")
	ErrOrderNotOpen   = errors.New("order is not open")
	ErrUnknownSymbol  = errors.New("unknown symbol")
	ErrInvalidRequest = errors.New("invalid exchange request")
)

type PlaceOrderRequest struct {
	UserID        int64
	ClientOrderID string
	Symbol        string
	Side          engine.Side
	Type          engine.OrderType
	TimeInForce   engine.TimeInForce
	Price         int64
	Quantity      int64
}

type PlaceOrderResult struct {
	Order  engine.Order
	Trades []engine.Trade
	// Duplicate is true when ClientOrderID was already accepted and the
	// existing order is returned instead of creating a new one.
	Duplicate bool
}

type orderState struct {
	order         engine.Order
	clientOrderID string
	// reserved tracks the remaining locked amount backing this order:
	// quote notional plus taker-fee bound for buys, base quantity for sells.
	reserved       int64
	reserveAccount int64
}

type Exchange struct {
	mu     sync.Mutex
	market oms.Market
	book   *engine.OrderBook

	accounts      account.AccountRepository
	userAccounts  map[string]int64
	feeAccounts   map[string]int64
	extAccounts   map[string]int64
	nextAccountID int64

	nextOrderID int64
	orders      map[int64]*orderState
	byClientID  map[string]int64
	trades      []engine.Trade

	journal        wal.Appender
	lastJournalSeq int64

	marketData MarketDataPublisher
	tickerAgg  *marketdata.TickerAggregator
	klineAgg   *marketdata.KlineAggregator
}

// MarketDataPublisher receives public market events. Implemented by
// marketdata.Hub; publishing must never block the matching path.
type MarketDataPublisher interface {
	PublishTrade(symbol string, data marketdata.TradeData)
	PublishOrderbook(symbol string, data marketdata.OrderbookData)
}

func New(market oms.Market) (*Exchange, error) {
	if market.Symbol == "" || market.BaseAsset == "" || market.QuoteAsset == "" {
		return nil, ErrInvalidRequest
	}
	return &Exchange{
		market:       market,
		book:         engine.NewOrderBook(market.Symbol),
		accounts:     account.NewStore(),
		userAccounts: make(map[string]int64),
		feeAccounts:  make(map[string]int64),
		extAccounts:  make(map[string]int64),
		orders:       make(map[int64]*orderState),
		byClientID:   make(map[string]int64),
	}, nil
}

func (e *Exchange) Market() oms.Market {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.market
}

// HaltMarket sets the market status to HALTED. New orders are rejected
// while halted; existing resting orders remain on the book.
func (e *Exchange) HaltMarket() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.market.Status = oms.MarketHalted
}

// ResumeMarket sets the market status back to TRADING.
func (e *Exchange) ResumeMarket() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.market.Status = oms.MarketTrading
}

// Users returns the unique set of user IDs that have accounts.
func (e *Exchange) Users() []int64 {
	e.mu.Lock()
	defer e.mu.Unlock()

	seen := make(map[int64]struct{})
	for key := range e.userAccounts {
		// key format: "userID:asset"
		for i := 0; i < len(key); i++ {
			if key[i] == ':' {
				uid := parseUserID(key[:i])
				if uid > 0 {
					seen[uid] = struct{}{}
				}
				break
			}
		}
	}
	users := make([]int64, 0, len(seen))
	for uid := range seen {
		users = append(users, uid)
	}
	return users
}

func parseUserID(s string) int64 {
	var n int64
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int64(c-'0')
	}
	return n
}

// SetJournal enables write-ahead logging of matching commands. Must be set
// before the exchange serves traffic; accepted commands are recorded before
// they reach the book, and a journal failure rejects the command.
func (e *Exchange) SetJournal(journal wal.Appender) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.journal = journal
}

// SetMarketData enables public trade/orderbook publication. Must be set
// before the exchange serves traffic.
func (e *Exchange) SetMarketData(publisher MarketDataPublisher) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.marketData = publisher
}

// SetAggregators enables ticker and kline aggregation. Must be set before
// the exchange serves traffic.
func (e *Exchange) SetAggregators(ticker *marketdata.TickerAggregator, kline *marketdata.KlineAggregator) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.tickerAgg = ticker
	e.klineAgg = kline
}

// publishTrade and publishOrderbook are called with e.mu held so events are
// emitted in matching order; the hub never blocks on subscribers.
func (e *Exchange) publishTrade(trade engine.Trade) {
	if e.marketData != nil {
		e.marketData.PublishTrade(e.market.Symbol, marketdata.TradeData{
			Price:     decimal.New(trade.Price, e.market.PriceScale).String(),
			Quantity:  decimal.New(trade.Quantity, e.market.QuantityScale).String(),
			TakerSide: string(trade.Side),
		})
	}
	if e.tickerAgg != nil {
		e.tickerAgg.RecordTrade(trade.Price, trade.Quantity)
	}
	if e.klineAgg != nil {
		e.klineAgg.RecordTrade(trade.Price, trade.Quantity)
	}
}

const publishedDepthLevels = 50

func (e *Exchange) publishOrderbook() {
	if e.marketData == nil {
		return
	}
	bids, asks := e.book.Depth(publishedDepthLevels)
	e.marketData.PublishOrderbook(e.market.Symbol, marketdata.OrderbookData{
		Type: "snapshot",
		Bids: e.priceLevels(bids),
		Asks: e.priceLevels(asks),
	})
}

func (e *Exchange) priceLevels(levels []engine.Level) []marketdata.PriceLevel {
	out := make([]marketdata.PriceLevel, 0, len(levels))
	for _, level := range levels {
		out = append(out, marketdata.PriceLevel{
			Price:    decimal.New(level.Price, e.market.PriceScale).String(),
			Quantity: decimal.New(level.Quantity, e.market.QuantityScale).String(),
		})
	}
	return out
}

// Checkpoint pairs an orderbook snapshot with the journal sequence of the
// last command applied to it, taken atomically. Recovery restores the book
// from the snapshot and replays journal commands after WALSequence.
type Checkpoint struct {
	Book        engine.OrderBookSnapshot
	WALSequence int64
}

func (e *Exchange) Checkpoint() Checkpoint {
	e.mu.Lock()
	defer e.mu.Unlock()
	return Checkpoint{Book: e.book.Snapshot(), WALSequence: e.lastJournalSeq}
}

func (e *Exchange) journalCommand(cmd wal.Command) error {
	if e.journal == nil {
		return nil
	}
	recorded, err := e.journal.Append(cmd)
	if err != nil {
		return err
	}
	e.lastJournalSeq = recorded.Sequence
	return nil
}

// Deposit applies one confirmed deposit as a balanced ledger transaction.
// Applying the same deposit id twice is a no-op.
func (e *Exchange) Deposit(depositID string, userID int64, asset string, amount int64) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	userAccount, err := e.ensureUserAccount(userID, asset)
	if err != nil {
		return err
	}
	externalAccount, err := e.ensureSystemAccount(e.extAccounts, account.TypeExternal, asset)
	if err != nil {
		return err
	}
	tx, err := funding.BuildDepositTransaction(funding.DepositRequest{
		DepositID:         depositID,
		Asset:             asset,
		UserAccountID:     userAccount,
		ExternalAccountID: externalAccount,
		Amount:            amount,
	})
	if err != nil {
		return err
	}
	return e.accounts.ApplyTransaction(tx)
}

// LockWithdrawal reserves amount+fee before a withdrawal enters review.
func (e *Exchange) LockWithdrawal(userID int64, asset string, amount int64) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	accountID, err := e.ensureUserAccount(userID, asset)
	if err != nil {
		return err
	}
	return e.accounts.Reserve(accountID, amount)
}

// ReleaseWithdrawal returns a rejected withdrawal's locked funds.
func (e *Exchange) ReleaseWithdrawal(userID int64, asset string, amount int64) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	accountID, err := e.ensureUserAccount(userID, asset)
	if err != nil {
		return err
	}
	return e.accounts.Release(accountID, amount)
}

// SettleWithdrawal converts a broadcasted withdrawal into ledger entries.
// The user account must hold amount+fee locked. Idempotent per withdrawal id.
func (e *Exchange) SettleWithdrawal(withdrawalID string, userID int64, asset string, amount, fee int64) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	userAccount, err := e.ensureUserAccount(userID, asset)
	if err != nil {
		return err
	}
	externalAccount, err := e.ensureSystemAccount(e.extAccounts, account.TypeExternal, asset)
	if err != nil {
		return err
	}
	feeAccount := int64(0)
	if fee > 0 {
		feeAccount, err = e.ensureSystemAccount(e.feeAccounts, account.TypeFeeRevenue, asset)
		if err != nil {
			return err
		}
	}
	tx, err := funding.BuildWithdrawalTransaction(funding.WithdrawalRequest{
		WithdrawalID:        withdrawalID,
		Asset:               asset,
		UserAccountID:       userAccount,
		ExternalAccountID:   externalAccount,
		FeeRevenueAccountID: feeAccount,
		Amount:              amount,
		Fee:                 fee,
	})
	if err != nil {
		return err
	}
	return e.accounts.ApplyTransaction(tx)
}

func (e *Exchange) PlaceOrder(req PlaceOrderRequest) (PlaceOrderResult, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if req.Symbol != e.market.Symbol {
		return PlaceOrderResult{}, ErrUnknownSymbol
	}
	if req.ClientOrderID != "" {
		if existingID, ok := e.byClientID[clientKey(req.UserID, req.ClientOrderID)]; ok {
			return PlaceOrderResult{Order: e.orders[existingID].order, Duplicate: true}, nil
		}
	}
	if req.TimeInForce == "" {
		req.TimeInForce = engine.GTC
	}

	e.nextOrderID++
	orderReq := oms.OrderRequest{
		OrderID:       e.nextOrderID,
		UserID:        req.UserID,
		Symbol:        req.Symbol,
		Side:          req.Side,
		Type:          req.Type,
		TimeInForce:   req.TimeInForce,
		Price:         req.Price,
		Quantity:      req.Quantity,
		ClientOrderID: req.ClientOrderID,
	}
	reservation, err := oms.ValidateSpotOrder(e.market, orderReq)
	if err != nil {
		return PlaceOrderResult{}, err
	}

	// Both settlement legs need base and quote accounts.
	if _, err := e.ensureUserAccount(req.UserID, e.market.BaseAsset); err != nil {
		return PlaceOrderResult{}, err
	}
	if _, err := e.ensureUserAccount(req.UserID, e.market.QuoteAsset); err != nil {
		return PlaceOrderResult{}, err
	}
	reserveAccount := e.userAccounts[userAssetKey(req.UserID, reservation.Asset)]
	if err := e.accounts.Reserve(reserveAccount, reservation.Amount); err != nil {
		return PlaceOrderResult{}, err
	}

	engineOrder := oms.ToEngineOrder(orderReq)
	if err := e.journalCommand(wal.Command{
		Type:   wal.CommandPlace,
		Symbol: e.market.Symbol,
		Order:  &engineOrder,
	}); err != nil {
		if releaseErr := e.accounts.Release(reserveAccount, reservation.Amount); releaseErr != nil {
			return PlaceOrderResult{}, fmt.Errorf("journal failed (%w) and release failed: %v", err, releaseErr)
		}
		return PlaceOrderResult{}, fmt.Errorf("journal place command: %w", err)
	}

	result, err := e.book.Submit(engineOrder)
	if err != nil {
		if releaseErr := e.accounts.Release(reserveAccount, reservation.Amount); releaseErr != nil {
			return PlaceOrderResult{}, fmt.Errorf("submit failed (%w) and release failed: %v", err, releaseErr)
		}
		return PlaceOrderResult{}, err
	}

	state := &orderState{
		order:          *result.Accepted,
		clientOrderID:  req.ClientOrderID,
		reserved:       reservation.Amount,
		reserveAccount: reserveAccount,
	}
	state.order.RemainingQuantity = state.order.Quantity
	e.orders[state.order.ID] = state
	if req.ClientOrderID != "" {
		e.byClientID[clientKey(req.UserID, req.ClientOrderID)] = state.order.ID
	}

	for _, trade := range result.Trades {
		if err := e.settleTrade(trade); err != nil {
			// Settlement failure after matching is unrecoverable in-process;
			// surface it loudly instead of leaving books and ledger diverged.
			return PlaceOrderResult{}, fmt.Errorf("settle trade %d: %w", trade.Sequence, err)
		}
	}

	if state.order.RemainingQuantity == 0 {
		state.order.Status = engine.OrderFilled
	} else if state.order.ExecutedQuantity > 0 {
		state.order.Status = engine.OrderPartiallyFilled
	}

	e.publishOrderbook()

	return PlaceOrderResult{Order: state.order, Trades: result.Trades}, nil
}

func (e *Exchange) CancelOrder(userID, orderID int64) (engine.Order, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	state, ok := e.orders[orderID]
	if !ok {
		return engine.Order{}, ErrOrderNotFound
	}
	if state.order.UserID != userID {
		return engine.Order{}, ErrOrderNotOwned
	}
	if !isOpen(state.order.Status) {
		return engine.Order{}, ErrOrderNotOpen
	}

	// Open orders are always resting in the book, so the cancel below can
	// only fail on internal inconsistency; journaling first keeps the WAL
	// ahead of the book mutation.
	if err := e.journalCommand(wal.Command{
		Type:    wal.CommandCancel,
		Symbol:  e.market.Symbol,
		OrderID: orderID,
	}); err != nil {
		return engine.Order{}, fmt.Errorf("journal cancel command: %w", err)
	}
	if _, err := e.book.Cancel(orderID); err != nil {
		return engine.Order{}, err
	}
	if state.reserved > 0 {
		if err := e.accounts.Release(state.reserveAccount, state.reserved); err != nil {
			return engine.Order{}, err
		}
		state.reserved = 0
	}
	state.order.Status = engine.OrderCanceled
	e.publishOrderbook()
	return state.order, nil
}

func (e *Exchange) Order(userID, orderID int64) (engine.Order, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	state, ok := e.orders[orderID]
	if !ok {
		return engine.Order{}, ErrOrderNotFound
	}
	if state.order.UserID != userID {
		return engine.Order{}, ErrOrderNotOwned
	}
	return state.order, nil
}

func (e *Exchange) OpenOrders(userID int64) []engine.Order {
	e.mu.Lock()
	defer e.mu.Unlock()

	var open []engine.Order
	for _, state := range e.orders {
		if state.order.UserID == userID && isOpen(state.order.Status) {
			open = append(open, state.order)
		}
	}
	return open
}

func (e *Exchange) Trades(limit int) []engine.Trade {
	e.mu.Lock()
	defer e.mu.Unlock()

	trades := e.trades
	if limit > 0 && len(trades) > limit {
		trades = trades[len(trades)-limit:]
	}
	out := make([]engine.Trade, len(trades))
	copy(out, trades)
	return out
}

func (e *Exchange) Depth(limit int) (bids, asks []engine.Level) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.book.Depth(limit)
}

func (e *Exchange) Balances(userID int64) []account.Balance {
	return e.accounts.BalancesByUser(userID)
}

// Accounts exposes the underlying store for reconciliation and tests.
func (e *Exchange) Accounts() account.AccountRepository {
	return e.accounts
}

func (e *Exchange) settleTrade(trade engine.Trade) error {
	maker, ok := e.orders[trade.MakerOrderID]
	if !ok {
		return fmt.Errorf("%w: maker %d", ErrOrderNotFound, trade.MakerOrderID)
	}
	taker, ok := e.orders[trade.TakerOrderID]
	if !ok {
		return fmt.Errorf("%w: taker %d", ErrOrderNotFound, trade.TakerOrderID)
	}

	buyer, seller := maker, taker
	if trade.Side == engine.Buy {
		buyer, seller = taker, maker
	}

	feeAccount, err := e.ensureSystemAccount(e.feeAccounts, account.TypeFeeRevenue, e.market.QuoteAsset)
	if err != nil {
		return err
	}

	tx, amounts, err := spotsettlement.BuildTradeTransaction(spotsettlement.Request{
		TradeID:              fmt.Sprintf("%s:%d", e.market.Symbol, trade.Sequence),
		Symbol:               e.market.Symbol,
		BaseAsset:            e.market.BaseAsset,
		QuoteAsset:           e.market.QuoteAsset,
		BuyerBaseAccountID:   e.userAccounts[userAssetKey(buyer.order.UserID, e.market.BaseAsset)],
		BuyerQuoteAccountID:  e.userAccounts[userAssetKey(buyer.order.UserID, e.market.QuoteAsset)],
		SellerBaseAccountID:  e.userAccounts[userAssetKey(seller.order.UserID, e.market.BaseAsset)],
		SellerQuoteAccountID: e.userAccounts[userAssetKey(seller.order.UserID, e.market.QuoteAsset)],
		FeeRevenueAccountID:  feeAccount,
		Price:                trade.Price,
		Quantity:             trade.Quantity,
		QuantityScale:        e.market.QuantityScale,
		TakerSide:            spotsettlement.Side(trade.Side),
		MakerFeeRatePPM:      e.market.MakerFeePPM,
		TakerFeeRatePPM:      e.market.TakerFeePPM,
	})
	if err != nil {
		return err
	}
	if err := e.accounts.ApplyTransaction(tx); err != nil {
		return err
	}

	applyFill(&buyer.order, trade.Quantity)
	applyFill(&seller.order, trade.Quantity)

	// Seller lock is consumed exactly by the base quantity.
	seller.reserved -= trade.Quantity

	// Buyer was locked at limit price with the taker-fee upper bound; the
	// fill settled at maker price with the actual fee. Shrink the lock to
	// what the remaining quantity still needs and release the difference.
	actualDebit := amounts.QuoteNotional + amounts.BuyerFee
	remainingNeed := int64(0)
	if buyer.order.RemainingQuantity > 0 {
		remainingNeed, err = oms.BuyReservationAmount(e.market, buyer.order.Price, buyer.order.RemainingQuantity)
		if err != nil {
			return err
		}
	}
	release := buyer.reserved - actualDebit - remainingNeed
	if release < 0 {
		return fmt.Errorf("%w: buyer reservation underflow order=%d", ErrInvalidRequest, buyer.order.ID)
	}
	if release > 0 {
		if err := e.accounts.Release(buyer.reserveAccount, release); err != nil {
			return err
		}
	}
	buyer.reserved = remainingNeed

	e.trades = append(e.trades, trade)
	e.publishTrade(trade)
	return nil
}

func (e *Exchange) ensureUserAccount(userID int64, asset string) (int64, error) {
	if userID <= 0 || asset == "" {
		return 0, ErrInvalidRequest
	}
	key := userAssetKey(userID, asset)
	if id, ok := e.userAccounts[key]; ok {
		return id, nil
	}
	e.nextAccountID++
	if err := e.accounts.CreateAccount(e.nextAccountID, userID, account.TypeSpot, asset, 0); err != nil {
		return 0, err
	}
	e.userAccounts[key] = e.nextAccountID
	return e.nextAccountID, nil
}

func (e *Exchange) ensureSystemAccount(registry map[string]int64, typ account.Type, asset string) (int64, error) {
	if id, ok := registry[asset]; ok {
		return id, nil
	}
	e.nextAccountID++
	if err := e.accounts.CreateAccount(e.nextAccountID, 0, typ, asset, 0); err != nil {
		return 0, err
	}
	registry[asset] = e.nextAccountID
	return e.nextAccountID, nil
}

func applyFill(order *engine.Order, quantity int64) {
	order.ExecutedQuantity += quantity
	order.RemainingQuantity = order.Quantity - order.ExecutedQuantity
	if order.RemainingQuantity == 0 {
		order.Status = engine.OrderFilled
	} else {
		order.Status = engine.OrderPartiallyFilled
	}
}

func isOpen(status engine.OrderStatus) bool {
	return status == engine.OrderAccepted || status == engine.OrderPartiallyFilled
}

func clientKey(userID int64, clientOrderID string) string {
	return fmt.Sprintf("%d:%s", userID, clientOrderID)
}

func userAssetKey(userID int64, asset string) string {
	return fmt.Sprintf("%d:%s", userID, asset)
}
