package engine

import "testing"

func TestLimitOrderRestsWhenNotMarketable(t *testing.T) {
	book := NewOrderBook("BTC-USDT")
	result, err := book.Submit(Order{
		ID:          1,
		UserID:      10,
		Symbol:      "BTC-USDT",
		Side:        Buy,
		Type:        Limit,
		TimeInForce: GTC,
		Price:       50_000_00,
		Quantity:    100_000_000,
	})
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}
	if len(result.Trades) != 0 {
		t.Fatalf("trades = %d, want 0", len(result.Trades))
	}
	if result.Rested == nil || result.Rested.RemainingQuantity != 100_000_000 {
		t.Fatalf("rested order missing or wrong quantity: %+v", result.Rested)
	}
	bestBid, ok := book.BestBid()
	if !ok || bestBid != 50_000_00 {
		t.Fatalf("best bid = %d/%v, want 5000000/true", bestBid, ok)
	}
}

func TestPriceTimePriority(t *testing.T) {
	book := NewOrderBook("BTC-USDT")
	mustSubmit(t, book, Order{ID: 1, UserID: 10, Symbol: "BTC-USDT", Side: Sell, Type: Limit, Price: 50_000_00, Quantity: 4})
	mustSubmit(t, book, Order{ID: 2, UserID: 20, Symbol: "BTC-USDT", Side: Sell, Type: Limit, Price: 49_900_00, Quantity: 3})
	mustSubmit(t, book, Order{ID: 3, UserID: 30, Symbol: "BTC-USDT", Side: Sell, Type: Limit, Price: 49_900_00, Quantity: 2})

	result, err := book.Submit(Order{ID: 4, UserID: 40, Symbol: "BTC-USDT", Side: Buy, Type: Limit, Price: 50_000_00, Quantity: 6})
	if err != nil {
		t.Fatalf("Submit taker: %v", err)
	}
	if len(result.Trades) != 3 {
		t.Fatalf("trades = %d, want 3", len(result.Trades))
	}
	wantMakerIDs := []int64{2, 3, 1}
	wantQty := []int64{3, 2, 1}
	for i, trade := range result.Trades {
		if trade.MakerOrderID != wantMakerIDs[i] || trade.Quantity != wantQty[i] {
			t.Fatalf("trade[%d] = %+v, want maker=%d qty=%d", i, trade, wantMakerIDs[i], wantQty[i])
		}
	}
}

func TestMarketOrderDoesNotRest(t *testing.T) {
	book := NewOrderBook("BTC-USDT")
	mustSubmit(t, book, Order{ID: 1, UserID: 10, Symbol: "BTC-USDT", Side: Sell, Type: Limit, Price: 50_000_00, Quantity: 2})

	result, err := book.Submit(Order{ID: 2, UserID: 20, Symbol: "BTC-USDT", Side: Buy, Type: Market, Quantity: 5})
	if err != nil {
		t.Fatalf("Submit market: %v", err)
	}
	if len(result.Trades) != 1 || result.Trades[0].Quantity != 2 {
		t.Fatalf("trades = %+v, want one partial fill qty 2", result.Trades)
	}
	if _, ok := book.OpenOrder(2); ok {
		t.Fatal("market order rested in book")
	}
}

func TestCancel(t *testing.T) {
	book := NewOrderBook("BTC-USDT")
	mustSubmit(t, book, Order{ID: 1, UserID: 10, Symbol: "BTC-USDT", Side: Buy, Type: Limit, Price: 50_000_00, Quantity: 2})

	canceled, err := book.Cancel(1)
	if err != nil {
		t.Fatalf("Cancel: %v", err)
	}
	if canceled.Status != OrderCanceled {
		t.Fatalf("status = %s, want %s", canceled.Status, OrderCanceled)
	}
	if _, ok := book.BestBid(); ok {
		t.Fatal("best bid exists after cancel")
	}
}

func TestDeterministicReplay(t *testing.T) {
	commands := []Order{
		{ID: 1, UserID: 10, Symbol: "BTC-USDT", Side: Sell, Type: Limit, Price: 50_000_00, Quantity: 2},
		{ID: 2, UserID: 20, Symbol: "BTC-USDT", Side: Buy, Type: Limit, Price: 50_100_00, Quantity: 1},
		{ID: 3, UserID: 30, Symbol: "BTC-USDT", Side: Buy, Type: Limit, Price: 49_900_00, Quantity: 3},
	}

	first := replay(t, commands)
	for i := 0; i < 100; i++ {
		next := replay(t, commands)
		if len(next) != len(first) {
			t.Fatalf("run %d trade count = %d, want %d", i, len(next), len(first))
		}
		for j := range first {
			if next[j] != first[j] {
				t.Fatalf("run %d trade %d = %+v, want %+v", i, j, next[j], first[j])
			}
		}
	}
}

func mustSubmit(t *testing.T, book *OrderBook, order Order) MatchResult {
	t.Helper()
	result, err := book.Submit(order)
	if err != nil {
		t.Fatalf("Submit(%+v): %v", order, err)
	}
	return result
}

func replay(t *testing.T, commands []Order) []Trade {
	t.Helper()
	book := NewOrderBook("BTC-USDT")
	var trades []Trade
	for _, command := range commands {
		result, err := book.Submit(command)
		if err != nil {
			t.Fatalf("replay submit: %v", err)
		}
		trades = append(trades, result.Trades...)
	}
	return trades
}
