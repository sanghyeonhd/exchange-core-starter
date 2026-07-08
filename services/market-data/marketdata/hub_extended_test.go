package marketdata

import (
	"testing"
	"time"
)

func TestPublishTickerAndKlineAssignSequences(t *testing.T) {
	hub := NewHub()
	hub.now = func() time.Time { return time.UnixMilli(1_720_000_000_000) }
	sub := hub.Subscribe(16)
	defer sub.Cancel()

	hub.PublishTicker("BTC-USDT", TickerData{LastPrice: "50000.00"})
	hub.PublishKline("BTC-USDT", KlineData{Interval: "1m", Open: "50000.00"})
	hub.PublishTicker("BTC-USDT", TickerData{LastPrice: "51000.00"})

	first := <-sub.C
	if first.Channel != ChannelTicker || first.Sequence != 1 {
		t.Fatalf("first = %+v", first)
	}
	second := <-sub.C
	if second.Channel != ChannelKline || second.Sequence != 1 {
		t.Fatalf("second = %+v", second)
	}
	third := <-sub.C
	if third.Channel != ChannelTicker || third.Sequence != 2 {
		t.Fatalf("third = %+v", third)
	}
}

func TestAllFourChannelsHaveIndependentSequences(t *testing.T) {
	hub := NewHub()
	hub.now = func() time.Time { return time.UnixMilli(1_720_000_000_000) }
	sub := hub.Subscribe(32)
	defer sub.Cancel()

	hub.PublishTrade("BTC-USDT", TradeData{Price: "50000.00"})
	hub.PublishOrderbook("BTC-USDT", OrderbookData{Type: "snapshot"})
	hub.PublishTicker("BTC-USDT", TickerData{LastPrice: "50000.00"})
	hub.PublishKline("BTC-USDT", KlineData{Interval: "1m"})

	// All four should have sequence 1.
	for i := 0; i < 4; i++ {
		env := <-sub.C
		if env.Sequence != 1 {
			t.Fatalf("channel %s sequence = %d, want 1", env.Channel, env.Sequence)
		}
	}

	// Second round.
	hub.PublishTrade("BTC-USDT", TradeData{Price: "51000.00"})
	hub.PublishOrderbook("BTC-USDT", OrderbookData{Type: "snapshot"})
	hub.PublishTicker("BTC-USDT", TickerData{LastPrice: "51000.00"})
	hub.PublishKline("BTC-USDT", KlineData{Interval: "1m"})

	for i := 0; i < 4; i++ {
		env := <-sub.C
		if env.Sequence != 2 {
			t.Fatalf("channel %s sequence = %d, want 2", env.Channel, env.Sequence)
		}
	}
}
