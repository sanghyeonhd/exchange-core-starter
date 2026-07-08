package marketdata

import (
	"testing"
	"time"
)

func TestPrivateHubRouting(t *testing.T) {
	hub := NewPrivateHub()
	hub.now = func() time.Time { return time.UnixMilli(1_720_000_000_000) }

	sub1 := hub.Subscribe(1, 16)
	defer sub1.Cancel()

	sub2 := hub.Subscribe(2, 16)
	defer sub2.Cancel()

	// Publish to user 1
	hub.PublishOrder(1, "BTC-USDT", OrderData{OrderID: 100})
	
	// Publish to user 2
	hub.PublishFill(2, "BTC-USDT", FillData{TradeID: 200})

	// User 1 should get order update, not fill
	env1 := <-sub1.C
	if env1.Channel != ChannelOrders || env1.Sequence != 1 {
		t.Fatalf("expected order for user 1, got %+v", env1)
	}
	if data := env1.Data.(OrderData); data.OrderID != 100 {
		t.Fatalf("expected OrderID 100, got %d", data.OrderID)
	}

	// User 2 should get fill update, not order
	env2 := <-sub2.C
	if env2.Channel != ChannelFills || env2.Sequence != 1 {
		t.Fatalf("expected fill for user 2, got %+v", env2)
	}
	if data := env2.Data.(FillData); data.TradeID != 200 {
		t.Fatalf("expected TradeID 200, got %d", data.TradeID)
	}
}

func TestPrivateHubSequencing(t *testing.T) {
	hub := NewPrivateHub()
	sub := hub.Subscribe(1, 16)
	defer sub.Cancel()

	hub.PublishBalance(1, BalanceData{Asset: "BTC", Available: "1.0"})
	hub.PublishBalance(1, BalanceData{Asset: "USDT", Available: "50000.0"})

	env1 := <-sub.C
	if env1.Sequence != 1 {
		t.Fatalf("expected sequence 1, got %d", env1.Sequence)
	}
	env2 := <-sub.C
	if env2.Sequence != 2 {
		t.Fatalf("expected sequence 2, got %d", env2.Sequence)
	}
}
