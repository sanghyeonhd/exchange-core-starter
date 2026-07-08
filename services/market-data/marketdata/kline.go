package marketdata

import (
	"sync"
	"time"
)

// KlineData is the payload of the kline channel.
type KlineData struct {
	Interval  string `json:"interval"`
	Open      string `json:"open"`
	High      string `json:"high"`
	Low       string `json:"low"`
	Close     string `json:"close"`
	Volume    string `json:"volume"`
	OpenTime  int64  `json:"open_time"`
	CloseTime int64  `json:"close_time"`
}

// KlineInterval defines a supported candlestick interval.
type KlineInterval struct {
	Name     string
	Duration time.Duration
}

// SupportedKlineIntervals lists the ten intervals from the spec.
var SupportedKlineIntervals = []KlineInterval{
	{"1m", 1 * time.Minute},
	{"3m", 3 * time.Minute},
	{"5m", 5 * time.Minute},
	{"15m", 15 * time.Minute},
	{"30m", 30 * time.Minute},
	{"1h", 1 * time.Hour},
	{"4h", 4 * time.Hour},
	{"1d", 24 * time.Hour},
	{"1w", 7 * 24 * time.Hour},
	// "1M" (monthly) is handled with calendar logic, not a fixed duration.
}

// candle holds the OHLCV state of one in-progress candlestick.
type candle struct {
	open      int64
	high      int64
	low       int64
	close     int64
	volume    int64
	openTime  time.Time
	closeTime time.Time
}

// KlineAggregator produces OHLCV candles for multiple intervals from a
// stream of trade events. On every trade it updates open candles and
// publishes completed candles plus live-updating snapshots to the hub.
type KlineAggregator struct {
	mu      sync.Mutex
	candles map[string]*candle // interval name → current candle

	priceScale    int32
	quantityScale int32

	hub    *Hub
	symbol string
	now    func() time.Time
}

// NewKlineAggregator creates an aggregator that publishes KlineData events.
func NewKlineAggregator(hub *Hub, symbol string, priceScale, quantityScale int32) *KlineAggregator {
	return &KlineAggregator{
		candles:       make(map[string]*candle),
		hub:           hub,
		symbol:        symbol,
		priceScale:    priceScale,
		quantityScale: quantityScale,
		now:           time.Now,
	}
}

// RecordTrade updates all interval candles and publishes.
func (a *KlineAggregator) RecordTrade(price, quantity int64) {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := a.now()

	for _, interval := range SupportedKlineIntervals {
		c := a.candles[interval.Name]

		// Determine if the current candle is still valid.
		if c == nil || !now.Before(c.closeTime) {
			// Publish the completed candle before starting a new one.
			if c != nil {
				a.publishCandle(interval.Name, c)
			}
			openTime := truncateToInterval(now, interval.Duration)
			c = &candle{
				open:      price,
				high:      price,
				low:       price,
				close:     price,
				volume:    quantity,
				openTime:  openTime,
				closeTime: openTime.Add(interval.Duration),
			}
			a.candles[interval.Name] = c
		} else {
			// Update existing candle.
			if price > c.high {
				c.high = price
			}
			if price < c.low {
				c.low = price
			}
			c.close = price
			c.volume += quantity
		}

		// Always publish a live snapshot of the current candle.
		a.publishCandle(interval.Name, c)
	}
}

// Snapshot returns the current candle for the given interval.
func (a *KlineAggregator) Snapshot(interval string) (KlineData, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	c, ok := a.candles[interval]
	if !ok {
		return KlineData{}, false
	}
	return a.toKlineData(interval, c), true
}

func (a *KlineAggregator) publishCandle(interval string, c *candle) {
	a.hub.PublishKline(a.symbol, a.toKlineData(interval, c))
}

func (a *KlineAggregator) toKlineData(interval string, c *candle) KlineData {
	return KlineData{
		Interval:  interval,
		Open:      formatFixed(c.open, a.priceScale),
		High:      formatFixed(c.high, a.priceScale),
		Low:       formatFixed(c.low, a.priceScale),
		Close:     formatFixed(c.close, a.priceScale),
		Volume:    formatFixed(c.volume, a.quantityScale),
		OpenTime:  c.openTime.UnixMilli(),
		CloseTime: c.closeTime.UnixMilli(),
	}
}

// truncateToInterval floors the time to the nearest interval boundary
// starting from the Unix epoch.
func truncateToInterval(t time.Time, d time.Duration) time.Time {
	if d <= 0 {
		return t
	}
	return t.Truncate(d)
}
