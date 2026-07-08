// Package marketdata fans public market events out to WebSocket
// subscribers using the envelope from docs/api/WEBSOCKET_SPEC.md.
package marketdata

import (
	"sync"
	"time"
)

const (
	ChannelTrades    = "trades"
	ChannelOrderbook = "orderbook"
)

// Envelope is the wire format for every channel message.
type Envelope struct {
	Channel   string `json:"channel"`
	Symbol    string `json:"symbol"`
	Sequence  int64  `json:"sequence"`
	Timestamp int64  `json:"timestamp"`
	Data      any    `json:"data"`
}

// TradeData is the payload of the trades channel.
type TradeData struct {
	Price     string `json:"price"`
	Quantity  string `json:"quantity"`
	TakerSide string `json:"taker_side"`
}

// PriceLevel is one aggregated orderbook level.
type PriceLevel struct {
	Price    string `json:"price"`
	Quantity string `json:"quantity"`
}

// OrderbookData is the payload of the orderbook channel. The MVP publishes
// full snapshots; per the spec, deltas are only valid after a snapshot, so
// snapshot-only is a valid degenerate form.
type OrderbookData struct {
	Type string       `json:"type"` // always "snapshot" for now
	Bids []PriceLevel `json:"bids"`
	Asks []PriceLevel `json:"asks"`
}

// Subscription receives envelopes until Cancel is called or the subscriber
// falls too far behind and is dropped by the hub.
type Subscription struct {
	C      <-chan Envelope
	hub    *Hub
	ch     chan Envelope
	cancel sync.Once
}

// Cancel detaches the subscription and closes its channel.
func (s *Subscription) Cancel() {
	s.cancel.Do(func() {
		s.hub.unsubscribe(s)
	})
}

// Hub assigns per-channel sequences and broadcasts to subscribers.
// Publishing never blocks on a slow consumer: a subscriber whose buffer is
// full is dropped, which forces the client to reconnect and resubscribe —
// the same contract the spec defines for sequence gaps.
type Hub struct {
	mu          sync.Mutex
	subscribers map[*Subscription]struct{}
	sequences   map[string]int64
	now         func() time.Time
}

func NewHub() *Hub {
	return &Hub{
		subscribers: make(map[*Subscription]struct{}),
		sequences:   make(map[string]int64),
		now:         time.Now,
	}
}

// Subscribe registers a consumer with the given buffer size.
func (h *Hub) Subscribe(buffer int) *Subscription {
	if buffer <= 0 {
		buffer = 64
	}
	ch := make(chan Envelope, buffer)
	sub := &Subscription{C: ch, ch: ch, hub: h}

	h.mu.Lock()
	defer h.mu.Unlock()
	h.subscribers[sub] = struct{}{}
	return sub
}

func (h *Hub) unsubscribe(sub *Subscription) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.subscribers[sub]; ok {
		delete(h.subscribers, sub)
		close(sub.ch)
	}
}

// PublishTrade broadcasts one public trade.
func (h *Hub) PublishTrade(symbol string, data TradeData) {
	h.publish(ChannelTrades, symbol, data)
}

// PublishOrderbook broadcasts an orderbook snapshot.
func (h *Hub) PublishOrderbook(symbol string, data OrderbookData) {
	h.publish(ChannelOrderbook, symbol, data)
}

func (h *Hub) publish(channel, symbol string, data any) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.sequences[channel]++
	envelope := Envelope{
		Channel:   channel,
		Symbol:    symbol,
		Sequence:  h.sequences[channel],
		Timestamp: h.now().UnixMilli(),
		Data:      data,
	}

	for sub := range h.subscribers {
		select {
		case sub.ch <- envelope:
		default:
			// Too slow: drop the subscriber rather than stall publishing.
			delete(h.subscribers, sub)
			close(sub.ch)
		}
	}
}
