package marketdata

import (
	"testing"
	"time"
)

func TestKlineRecordTradeCreatesCandles(t *testing.T) {
	hub := NewHub()
	hub.now = func() time.Time { return time.UnixMilli(1_720_000_000_000) }
	sub := hub.Subscribe(256)
	defer sub.Cancel()

	baseTime := time.Date(2025, 1, 1, 12, 0, 30, 0, time.UTC) // mid-minute
	agg := NewKlineAggregator(hub, "BTC-USDT", 2, 8)
	agg.now = func() time.Time { return baseTime }

	agg.RecordTrade(50_000_00, 1_000_000)

	// Check 1m candle snapshot.
	kline, ok := agg.Snapshot("1m")
	if !ok {
		t.Fatal("1m candle should exist")
	}
	if kline.Open != "50000.00" || kline.Close != "50000.00" {
		t.Fatalf("open=%s close=%s, want both 50000.00", kline.Open, kline.Close)
	}
	if kline.Volume != "0.01000000" {
		t.Fatalf("volume = %s, want 0.01000000", kline.Volume)
	}

	// Second trade within same minute: updates high/close.
	agg.RecordTrade(51_000_00, 2_000_000)
	kline, _ = agg.Snapshot("1m")
	if kline.High != "51000.00" {
		t.Fatalf("high = %s, want 51000.00", kline.High)
	}
	if kline.Low != "50000.00" {
		t.Fatalf("low = %s, want 50000.00", kline.Low)
	}
	if kline.Close != "51000.00" {
		t.Fatalf("close = %s, want 51000.00", kline.Close)
	}
	if kline.Volume != "0.03000000" {
		t.Fatalf("volume = %s, want 0.03000000", kline.Volume)
	}

	// Verify kline envelopes were published.
	var klineCount int
	for len(sub.C) > 0 {
		env := <-sub.C
		if env.Channel == ChannelKline {
			klineCount++
		}
	}
	// 2 trades × 9 intervals (excl. 1M monthly) = at least 18 kline publishes.
	if klineCount < 18 {
		t.Fatalf("kline publishes = %d, want >= 18", klineCount)
	}
}

func TestKlineNewCandleOnIntervalBoundary(t *testing.T) {
	hub := NewHub()
	agg := NewKlineAggregator(hub, "BTC-USDT", 2, 8)

	// First trade at 12:00:30.
	t1 := time.Date(2025, 1, 1, 12, 0, 30, 0, time.UTC)
	agg.now = func() time.Time { return t1 }
	agg.RecordTrade(50_000_00, 1_000_000)

	// Second trade at 12:01:10 — should start new 1m candle.
	t2 := time.Date(2025, 1, 1, 12, 1, 10, 0, time.UTC)
	agg.now = func() time.Time { return t2 }
	agg.RecordTrade(52_000_00, 500_000)

	kline, ok := agg.Snapshot("1m")
	if !ok {
		t.Fatal("1m candle should exist")
	}
	// The new candle should have open = close = 52000.00.
	if kline.Open != "52000.00" {
		t.Fatalf("new candle open = %s, want 52000.00", kline.Open)
	}
	if kline.Volume != "0.00500000" {
		t.Fatalf("new candle volume = %s, want 0.00500000", kline.Volume)
	}
}
