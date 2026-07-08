package marketdata

import (
	"testing"
	"time"
)

func TestTickerRecordTradeUpdatesFields(t *testing.T) {
	hub := NewHub()
	hub.now = func() time.Time { return time.UnixMilli(1_720_000_000_000) }
	sub := hub.Subscribe(64)
	defer sub.Cancel()

	agg := NewTickerAggregator(hub, "BTC-USDT", 2, 8)
	agg.now = func() time.Time { return time.UnixMilli(1_720_000_000_000) }

	// First trade: last=high=low=50000.00, volume=0.01000000
	agg.RecordTrade(50_000_00, 1_000_000) // 50000.00, 0.01 BTC

	snap := agg.Snapshot()
	if snap.LastPrice != "50000.00" {
		t.Fatalf("last price = %s, want 50000.00", snap.LastPrice)
	}
	if snap.High24h != "50000.00" || snap.Low24h != "50000.00" {
		t.Fatalf("high=%s low=%s, want both 50000.00", snap.High24h, snap.Low24h)
	}
	if snap.Volume24h != "0.01000000" {
		t.Fatalf("volume = %s, want 0.01000000", snap.Volume24h)
	}

	// Second trade: higher price
	agg.RecordTrade(51_000_00, 2_000_000) // 51000.00, 0.02 BTC
	snap = agg.Snapshot()
	if snap.LastPrice != "51000.00" {
		t.Fatalf("last price = %s, want 51000.00", snap.LastPrice)
	}
	if snap.High24h != "51000.00" {
		t.Fatalf("high = %s, want 51000.00", snap.High24h)
	}
	if snap.Low24h != "50000.00" {
		t.Fatalf("low = %s, want 50000.00", snap.Low24h)
	}
	if snap.Volume24h != "0.03000000" {
		t.Fatalf("volume = %s, want 0.03000000", snap.Volume24h)
	}

	// Verify hub received ticker envelopes.
	var tickerCount int
	for len(sub.C) > 0 {
		env := <-sub.C
		if env.Channel == ChannelTicker {
			tickerCount++
		}
	}
	if tickerCount != 2 {
		t.Fatalf("ticker publishes = %d, want 2", tickerCount)
	}
}

func TestTickerEvictsExpiredTrades(t *testing.T) {
	hub := NewHub()
	agg := NewTickerAggregator(hub, "BTC-USDT", 2, 8)

	// Two trades at time 0.
	baseTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	agg.now = func() time.Time { return baseTime }
	agg.RecordTrade(50_000_00, 1_000_000)
	agg.RecordTrade(49_000_00, 1_000_000)

	// Advance 25 hours — the old trades should be evicted.
	agg.now = func() time.Time { return baseTime.Add(25 * time.Hour) }
	agg.RecordTrade(52_000_00, 500_000)

	snap := agg.Snapshot()
	if snap.Low24h != "52000.00" {
		t.Fatalf("low should be 52000.00 after eviction, got %s", snap.Low24h)
	}
	if snap.Volume24h != "0.00500000" {
		t.Fatalf("volume should be 0.00500000 after eviction, got %s", snap.Volume24h)
	}
}

func TestFormatFixed(t *testing.T) {
	tests := []struct {
		value int64
		scale int32
		want  string
	}{
		{50_000_00, 2, "50000.00"},
		{1_000_000, 8, "0.01000000"},
		{0, 2, "0.00"},
		{100, 2, "1.00"},
		{1, 2, "0.01"},
		{-50_000_00, 2, "-50000.00"},
	}
	for _, tt := range tests {
		got := formatFixed(tt.value, tt.scale)
		if got != tt.want {
			t.Errorf("formatFixed(%d, %d) = %s, want %s", tt.value, tt.scale, got, tt.want)
		}
	}
}
