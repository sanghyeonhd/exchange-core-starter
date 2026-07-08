package marketdata

import (
	"sync"
	"time"
)

// TickerData is the payload of the ticker channel.
type TickerData struct {
	LastPrice      string `json:"last_price"`
	Volume24h      string `json:"volume_24h"`
	High24h        string `json:"high_24h"`
	Low24h         string `json:"low_24h"`
	PriceChange    string `json:"price_change"`
	PriceChangePct string `json:"price_change_pct"`
}

// trade records a single trade for rolling window calculations.
type trade struct {
	price    int64
	quantity int64
	ts       time.Time
}

// TickerAggregator accumulates trades and produces a 24-hour rolling ticker.
// It is safe for concurrent use and never blocks the matching path: the hub
// publish call is the only external side effect and is non-blocking by design.
type TickerAggregator struct {
	mu     sync.Mutex
	trades []trade

	priceScale    int32
	quantityScale int32

	hub    *Hub
	symbol string
	now    func() time.Time
}

// NewTickerAggregator creates an aggregator that publishes TickerData to the
// given hub each time a trade is recorded.
func NewTickerAggregator(hub *Hub, symbol string, priceScale, quantityScale int32) *TickerAggregator {
	return &TickerAggregator{
		hub:           hub,
		symbol:        symbol,
		priceScale:    priceScale,
		quantityScale: quantityScale,
		now:           time.Now,
	}
}

// RecordTrade adds a trade and publishes an updated ticker snapshot.
func (a *TickerAggregator) RecordTrade(price, quantity int64) {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := a.now()
	a.trades = append(a.trades, trade{price: price, quantity: quantity, ts: now})
	a.evictOlderThan(now.Add(-24 * time.Hour))

	data := a.computeLocked()
	a.hub.PublishTicker(a.symbol, data)
}

// Snapshot returns the current ticker without publishing.
func (a *TickerAggregator) Snapshot() TickerData {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := a.now()
	a.evictOlderThan(now.Add(-24 * time.Hour))
	return a.computeLocked()
}

func (a *TickerAggregator) evictOlderThan(cutoff time.Time) {
	idx := 0
	for idx < len(a.trades) && a.trades[idx].ts.Before(cutoff) {
		idx++
	}
	if idx > 0 {
		a.trades = a.trades[idx:]
	}
}

func (a *TickerAggregator) computeLocked() TickerData {
	if len(a.trades) == 0 {
		return TickerData{}
	}

	last := a.trades[len(a.trades)-1]
	first := a.trades[0]

	high := first.price
	low := first.price
	var volumeRaw int64
	for _, t := range a.trades {
		if t.price > high {
			high = t.price
		}
		if t.price < low {
			low = t.price
		}
		volumeRaw += t.quantity
	}

	priceChange := last.price - first.price
	var pctStr string
	if first.price != 0 {
		// percentage as a string with 2 decimals: (change * 10000 / first) / 100.0
		pctBps := priceChange * 10000 / first.price
		sign := ""
		if pctBps < 0 {
			sign = "-"
			pctBps = -pctBps
		}
		pctStr = sign + formatFixed(pctBps, 2)
	}

	return TickerData{
		LastPrice:      formatFixed(last.price, a.priceScale),
		Volume24h:      formatFixed(volumeRaw, a.quantityScale),
		High24h:        formatFixed(high, a.priceScale),
		Low24h:         formatFixed(low, a.priceScale),
		PriceChange:    formatFixed(priceChange, a.priceScale),
		PriceChangePct: pctStr,
	}
}

// formatFixed converts an int64 fixed-point value to a decimal string using
// the given scale (number of decimal places).
func formatFixed(value int64, scale int32) string {
	if scale <= 0 {
		return intToStr(value)
	}

	negative := false
	if value < 0 {
		negative = true
		value = -value
	}

	divisor := int64(1)
	for i := int32(0); i < scale; i++ {
		divisor *= 10
	}

	whole := value / divisor
	frac := value % divisor

	fracStr := intToStr(frac)
	for int32(len(fracStr)) < scale {
		fracStr = "0" + fracStr
	}

	prefix := ""
	if negative {
		prefix = "-"
	}
	return prefix + intToStr(whole) + "." + fracStr
}

func intToStr(v int64) string {
	if v == 0 {
		return "0"
	}
	negative := false
	if v < 0 {
		negative = true
		v = -v
	}
	var buf [20]byte
	pos := len(buf)
	for v > 0 {
		pos--
		buf[pos] = byte('0' + v%10)
		v /= 10
	}
	if negative {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
