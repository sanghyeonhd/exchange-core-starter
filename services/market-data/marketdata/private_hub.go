package marketdata

import (
	"sync"
	"time"
)

const (
	ChannelOrders   = "orders"
	ChannelFills    = "fills"
	ChannelBalances = "balances"
)

// OrderData is the payload for the orders private channel.
type OrderData struct {
	OrderID          int64  `json:"order_id"`
	ClientOrderID    string `json:"client_order_id,omitempty"`
	Symbol           string `json:"symbol"`
	Side             string `json:"side"`
	Type             string `json:"type"`
	Price            string `json:"price"`
	Quantity         string `json:"quantity"`
	ExecutedQuantity string `json:"executed_quantity"`
	Status           string `json:"status"`
}

// FillData is the payload for the fills private channel.
type FillData struct {
	OrderID   int64  `json:"order_id"`
	Symbol    string `json:"symbol"`
	TradeID   int64  `json:"trade_id"`
	Price     string `json:"price"`
	Quantity  string `json:"quantity"`
	Fee       string `json:"fee"`
	FeeAsset  string `json:"fee_asset"`
}

// BalanceData is the payload for the balances private channel.
type BalanceData struct {
	Asset     string `json:"asset"`
	Available string `json:"available"`
	Locked    string `json:"locked"`
}

type userSequenceKey struct {
	userID  int64
	channel string
}

// PrivateSubscription receives envelopes for a specific user.
type PrivateSubscription struct {
	C      <-chan Envelope
	UserID int64
	hub    *PrivateHub
	ch     chan Envelope
	cancel sync.Once
}

// Cancel detaches the subscription and closes its channel.
func (s *PrivateSubscription) Cancel() {
	s.cancel.Do(func() {
		s.hub.unsubscribe(s)
	})
}

// PrivateHub routes private channel envelopes to specific users.
type PrivateHub struct {
	mu          sync.Mutex
	subscribers map[int64]map[*PrivateSubscription]struct{}
	sequences   map[userSequenceKey]int64
	now         func() time.Time
}

func NewPrivateHub() *PrivateHub {
	return &PrivateHub{
		subscribers: make(map[int64]map[*PrivateSubscription]struct{}),
		sequences:   make(map[userSequenceKey]int64),
		now:         time.Now,
	}
}

// Subscribe registers a consumer for a specific user with the given buffer size.
func (h *PrivateHub) Subscribe(userID int64, buffer int) *PrivateSubscription {
	if buffer <= 0 {
		buffer = 64
	}
	ch := make(chan Envelope, buffer)
	sub := &PrivateSubscription{C: ch, ch: ch, hub: h, UserID: userID}

	h.mu.Lock()
	defer h.mu.Unlock()
	if h.subscribers[userID] == nil {
		h.subscribers[userID] = make(map[*PrivateSubscription]struct{})
	}
	h.subscribers[userID][sub] = struct{}{}
	return sub
}

func (h *PrivateHub) unsubscribe(sub *PrivateSubscription) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if userSubs, ok := h.subscribers[sub.UserID]; ok {
		if _, ok := userSubs[sub]; ok {
			delete(userSubs, sub)
			close(sub.ch)
			if len(userSubs) == 0 {
				delete(h.subscribers, sub.UserID)
			}
		}
	}
}

// PublishOrder broadcasts an order update to a specific user.
func (h *PrivateHub) PublishOrder(userID int64, symbol string, data OrderData) {
	h.publish(userID, ChannelOrders, symbol, data)
}

// PublishFill broadcasts a fill event to a specific user.
func (h *PrivateHub) PublishFill(userID int64, symbol string, data FillData) {
	h.publish(userID, ChannelFills, symbol, data)
}

// PublishBalance broadcasts a balance update to a specific user.
func (h *PrivateHub) PublishBalance(userID int64, data BalanceData) {
	h.publish(userID, ChannelBalances, "", data)
}

func (h *PrivateHub) publish(userID int64, channel, symbol string, data any) {
	h.mu.Lock()
	defer h.mu.Unlock()

	key := userSequenceKey{userID: userID, channel: channel}
	h.sequences[key]++
	envelope := Envelope{
		Channel:   channel,
		Symbol:    symbol,
		Sequence:  h.sequences[key],
		Timestamp: h.now().UnixMilli(),
		Data:      data,
	}

	userSubs, ok := h.subscribers[userID]
	if !ok {
		return
	}

	for sub := range userSubs {
		select {
		case sub.ch <- envelope:
		default:
			// Too slow: drop the subscriber
			delete(userSubs, sub)
			close(sub.ch)
		}
	}
	if len(userSubs) == 0 {
		delete(h.subscribers, userID)
	}
}
