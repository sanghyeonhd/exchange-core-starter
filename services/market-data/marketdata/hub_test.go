package marketdata

import (
	"testing"
	"time"
)

func TestPublishAssignsPerChannelSequences(t *testing.T) {
	hub := NewHub()
	hub.now = func() time.Time { return time.UnixMilli(1_720_000_000_000) }
	sub := hub.Subscribe(8)
	defer sub.Cancel()

	hub.PublishTrade("BTC-USDT", TradeData{Price: "50000.00", Quantity: "0.01000000", TakerSide: "BUY"})
	hub.PublishOrderbook("BTC-USDT", OrderbookData{Type: "snapshot"})
	hub.PublishTrade("BTC-USDT", TradeData{Price: "50001.00", Quantity: "0.02000000", TakerSide: "SELL"})

	first := <-sub.C
	if first.Channel != ChannelTrades || first.Sequence != 1 || first.Timestamp != 1_720_000_000_000 {
		t.Fatalf("first = %+v", first)
	}
	second := <-sub.C
	if second.Channel != ChannelOrderbook || second.Sequence != 1 {
		t.Fatalf("second = %+v", second)
	}
	third := <-sub.C
	if third.Channel != ChannelTrades || third.Sequence != 2 {
		t.Fatalf("third = %+v", third)
	}
}

func TestSlowSubscriberIsDroppedWithoutBlocking(t *testing.T) {
	hub := NewHub()
	slow := hub.Subscribe(1)
	fast := hub.Subscribe(8)
	defer fast.Cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 5; i++ {
			hub.PublishTrade("BTC-USDT", TradeData{Price: "1.00"})
		}
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("publish blocked on slow subscriber")
	}

	// The slow subscriber's channel must be closed after overflow.
	var received int
	for range slow.C {
		received++
	}
	if received != 1 {
		t.Fatalf("slow subscriber received %d messages, want 1 buffered", received)
	}

	// The fast subscriber saw everything.
	var fastCount int
	for len(fast.C) > 0 {
		<-fast.C
		fastCount++
	}
	if fastCount != 5 {
		t.Fatalf("fast subscriber received %d, want 5", fastCount)
	}
}

func TestCancelIsIdempotent(t *testing.T) {
	hub := NewHub()
	sub := hub.Subscribe(1)
	sub.Cancel()
	sub.Cancel()
	if _, open := <-sub.C; open {
		t.Fatal("channel should be closed after cancel")
	}
	// Publishing after cancel must not panic.
	hub.PublishTrade("BTC-USDT", TradeData{Price: "1.00"})
}
