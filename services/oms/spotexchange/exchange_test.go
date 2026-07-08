package spotexchange

import (
	"errors"
	"testing"

	"github.com/exchange-core-starter/exchange-core-starter/services/account/account"
	"github.com/exchange-core-starter/exchange-core-starter/services/market-data/marketdata"
	"github.com/exchange-core-starter/exchange-core-starter/services/matching-engine/engine"
	"github.com/exchange-core-starter/exchange-core-starter/services/oms/oms"
)

const (
	alice int64 = 1
	bob   int64 = 2
)

func testMarket() oms.Market {
	return oms.Market{
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
	}
}

func newTestExchange(t *testing.T) *Exchange {
	t.Helper()
	ex, err := New(testMarket())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := ex.Deposit("dep-alice-usdt", alice, "USDT", 1_000_000); err != nil {
		t.Fatalf("Deposit alice: %v", err)
	}
	if err := ex.Deposit("dep-bob-btc", bob, "BTC", 100_000_000); err != nil {
		t.Fatalf("Deposit bob: %v", err)
	}
	return ex
}

func balance(t *testing.T, ex *Exchange, userID int64, asset string) account.Balance {
	t.Helper()
	for _, b := range ex.Balances(userID) {
		if b.Asset == asset {
			return b
		}
	}
	t.Fatalf("no %s balance for user %d", asset, userID)
	return account.Balance{}
}

func TestDepositIsIdempotent(t *testing.T) {
	ex := newTestExchange(t)
	if err := ex.Deposit("dep-alice-usdt", alice, "USDT", 1_000_000); err != nil {
		t.Fatalf("duplicate deposit: %v", err)
	}
	if got := balance(t, ex, alice, "USDT").Available; got != 1_000_000 {
		t.Fatalf("available = %d, want 1000000 (duplicate must not double-credit)", got)
	}
}

func TestPlaceOrderReservesAndRests(t *testing.T) {
	ex := newTestExchange(t)
	result, err := ex.PlaceOrder(PlaceOrderRequest{
		UserID: alice, ClientOrderID: "a1", Symbol: "BTC-USDT",
		Side: engine.Buy, Type: engine.Limit,
		Price: 50_000_00, Quantity: 10_000_000,
	})
	if err != nil {
		t.Fatalf("PlaceOrder: %v", err)
	}
	if result.Order.Status != engine.OrderAccepted || len(result.Trades) != 0 {
		t.Fatalf("result = %+v", result)
	}
	b := balance(t, ex, alice, "USDT")
	// notional 500000 + taker fee bound 100
	if b.Available != 499_900 || b.Locked != 500_100 {
		t.Fatalf("balance = %+v", b)
	}
}

func TestPlaceOrderInsufficientBalance(t *testing.T) {
	ex := newTestExchange(t)
	_, err := ex.PlaceOrder(PlaceOrderRequest{
		UserID: alice, Symbol: "BTC-USDT",
		Side: engine.Buy, Type: engine.Limit,
		Price: 50_000_00, Quantity: 100_000_000, // needs 50M quote, has 1M
	})
	if !errors.Is(err, account.ErrInsufficientBalance) {
		t.Fatalf("err = %v, want ErrInsufficientBalance", err)
	}
}

func TestClientOrderIDIsIdempotent(t *testing.T) {
	ex := newTestExchange(t)
	first, err := ex.PlaceOrder(PlaceOrderRequest{
		UserID: alice, ClientOrderID: "same", Symbol: "BTC-USDT",
		Side: engine.Buy, Type: engine.Limit, Price: 50_000_00, Quantity: 10_000_000,
	})
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	second, err := ex.PlaceOrder(PlaceOrderRequest{
		UserID: alice, ClientOrderID: "same", Symbol: "BTC-USDT",
		Side: engine.Buy, Type: engine.Limit, Price: 50_000_00, Quantity: 10_000_000,
	})
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	if !second.Duplicate || second.Order.ID != first.Order.ID {
		t.Fatalf("second = %+v, want duplicate of order %d", second, first.Order.ID)
	}
	if b := balance(t, ex, alice, "USDT"); b.Locked != 500_100 {
		t.Fatalf("locked = %d, duplicate must not re-reserve", b.Locked)
	}
}

func TestFullMatchSettlesBalances(t *testing.T) {
	ex := newTestExchange(t)
	if _, err := ex.PlaceOrder(PlaceOrderRequest{
		UserID: alice, ClientOrderID: "a1", Symbol: "BTC-USDT",
		Side: engine.Buy, Type: engine.Limit, Price: 50_000_00, Quantity: 10_000_000,
	}); err != nil {
		t.Fatalf("buy: %v", err)
	}
	result, err := ex.PlaceOrder(PlaceOrderRequest{
		UserID: bob, ClientOrderID: "b1", Symbol: "BTC-USDT",
		Side: engine.Sell, Type: engine.Limit, Price: 50_000_00, Quantity: 10_000_000,
	})
	if err != nil {
		t.Fatalf("sell: %v", err)
	}
	if len(result.Trades) != 1 || result.Order.Status != engine.OrderFilled {
		t.Fatalf("result = %+v", result)
	}

	// Same expectations as the spot loop integration test:
	// maker Alice pays 100 ppm fee (50), taker Bob pays 200 ppm (100).
	assertBalance(t, ex, alice, "USDT", 499_950, 0)
	assertBalance(t, ex, alice, "BTC", 10_000_000, 0)
	assertBalance(t, ex, bob, "BTC", 90_000_000, 0)
	assertBalance(t, ex, bob, "USDT", 499_900, 0)

	if trades := ex.Trades(0); len(trades) != 1 {
		t.Fatalf("trades = %+v", trades)
	}
	bids, asks := ex.Depth(0)
	if len(bids) != 0 || len(asks) != 0 {
		t.Fatalf("book not empty: bids=%v asks=%v", bids, asks)
	}
}

func TestPartialFillKeepsReservationForRemainder(t *testing.T) {
	ex := newTestExchange(t)
	if err := ex.Deposit("dep-alice-usdt-2", alice, "USDT", 1_000_000); err != nil {
		t.Fatalf("extra deposit: %v", err)
	}
	if _, err := ex.PlaceOrder(PlaceOrderRequest{
		UserID: alice, Symbol: "BTC-USDT",
		Side: engine.Buy, Type: engine.Limit, Price: 50_000_00, Quantity: 20_000_000,
	}); err != nil {
		t.Fatalf("buy: %v", err)
	}
	if _, err := ex.PlaceOrder(PlaceOrderRequest{
		UserID: bob, Symbol: "BTC-USDT",
		Side: engine.Sell, Type: engine.Limit, Price: 50_000_00, Quantity: 10_000_000,
	}); err != nil {
		t.Fatalf("sell: %v", err)
	}

	// Alice bought 0.1 BTC for 500000 + 50 maker fee; 0.1 BTC remains
	// reserved at 500000 + 100 taker-fee bound.
	b := balance(t, ex, alice, "USDT")
	if b.Locked != 500_100 {
		t.Fatalf("locked = %d, want 500100", b.Locked)
	}
	if b.Available != 2_000_000-500_100-500_050 {
		t.Fatalf("available = %d", b.Available)
	}

	open := ex.OpenOrders(alice)
	if len(open) != 1 || open[0].Status != engine.OrderPartiallyFilled || open[0].RemainingQuantity != 10_000_000 {
		t.Fatalf("open = %+v", open)
	}
}

func TestCancelReleasesReservation(t *testing.T) {
	ex := newTestExchange(t)
	placed, err := ex.PlaceOrder(PlaceOrderRequest{
		UserID: alice, Symbol: "BTC-USDT",
		Side: engine.Buy, Type: engine.Limit, Price: 50_000_00, Quantity: 10_000_000,
	})
	if err != nil {
		t.Fatalf("buy: %v", err)
	}

	if _, err := ex.CancelOrder(bob, placed.Order.ID); !errors.Is(err, ErrOrderNotOwned) {
		t.Fatalf("cancel by other user err = %v, want ErrOrderNotOwned", err)
	}

	canceled, err := ex.CancelOrder(alice, placed.Order.ID)
	if err != nil {
		t.Fatalf("CancelOrder: %v", err)
	}
	if canceled.Status != engine.OrderCanceled {
		t.Fatalf("status = %s", canceled.Status)
	}
	assertBalance(t, ex, alice, "USDT", 1_000_000, 0)

	if _, err := ex.CancelOrder(alice, placed.Order.ID); !errors.Is(err, ErrOrderNotOpen) {
		t.Fatalf("double cancel err = %v, want ErrOrderNotOpen", err)
	}
}

func TestWithdrawalLockSettleFlow(t *testing.T) {
	ex := newTestExchange(t)
	if err := ex.LockWithdrawal(alice, "USDT", 100_050); err != nil {
		t.Fatalf("LockWithdrawal: %v", err)
	}
	assertBalance(t, ex, alice, "USDT", 899_950, 100_050)

	if err := ex.SettleWithdrawal("wd-1", alice, "USDT", 100_000, 50); err != nil {
		t.Fatalf("SettleWithdrawal: %v", err)
	}
	assertBalance(t, ex, alice, "USDT", 899_950, 0)

	// Duplicate settlement must be a no-op.
	if err := ex.SettleWithdrawal("wd-1", alice, "USDT", 100_000, 50); err != nil {
		t.Fatalf("duplicate SettleWithdrawal: %v", err)
	}
	assertBalance(t, ex, alice, "USDT", 899_950, 0)
}

func TestWithdrawalReleaseOnReject(t *testing.T) {
	ex := newTestExchange(t)
	if err := ex.LockWithdrawal(alice, "USDT", 100_000); err != nil {
		t.Fatalf("LockWithdrawal: %v", err)
	}
	if err := ex.ReleaseWithdrawal(alice, "USDT", 100_000); err != nil {
		t.Fatalf("ReleaseWithdrawal: %v", err)
	}
	assertBalance(t, ex, alice, "USDT", 1_000_000, 0)
}

func TestMarketDataPublication(t *testing.T) {
	ex := newTestExchange(t)
	hub := marketdata.NewHub()
	ex.SetMarketData(hub)
	sub := hub.Subscribe(16)
	defer sub.Cancel()

	if _, err := ex.PlaceOrder(PlaceOrderRequest{
		UserID: alice, Symbol: "BTC-USDT",
		Side: engine.Buy, Type: engine.Limit, Price: 50_000_00, Quantity: 10_000_000,
	}); err != nil {
		t.Fatalf("buy: %v", err)
	}
	if _, err := ex.PlaceOrder(PlaceOrderRequest{
		UserID: bob, Symbol: "BTC-USDT",
		Side: engine.Sell, Type: engine.Limit, Price: 50_000_00, Quantity: 10_000_000,
	}); err != nil {
		t.Fatalf("sell: %v", err)
	}

	// Expected stream: orderbook (buy rests), trade, orderbook (after match).
	first := <-sub.C
	if first.Channel != marketdata.ChannelOrderbook {
		t.Fatalf("first = %+v", first)
	}
	book := first.Data.(marketdata.OrderbookData)
	if len(book.Bids) != 1 || book.Bids[0].Price != "50000.00" || book.Bids[0].Quantity != "0.10000000" {
		t.Fatalf("book = %+v", book)
	}

	second := <-sub.C
	if second.Channel != marketdata.ChannelTrades {
		t.Fatalf("second = %+v", second)
	}
	trade := second.Data.(marketdata.TradeData)
	if trade.Price != "50000.00" || trade.Quantity != "0.10000000" || trade.TakerSide != "SELL" {
		t.Fatalf("trade = %+v", trade)
	}

	third := <-sub.C
	if third.Channel != marketdata.ChannelOrderbook {
		t.Fatalf("third = %+v", third)
	}
	if emptied := third.Data.(marketdata.OrderbookData); len(emptied.Bids) != 0 || len(emptied.Asks) != 0 {
		t.Fatalf("book after match = %+v", emptied)
	}
}

func assertBalance(t *testing.T, ex *Exchange, userID int64, asset string, available, locked int64) {
	t.Helper()
	b := balance(t, ex, userID, asset)
	if b.Available != available || b.Locked != locked {
		t.Fatalf("user %d %s balance = available=%d locked=%d, want available=%d locked=%d",
			userID, asset, b.Available, b.Locked, available, locked)
	}
}
