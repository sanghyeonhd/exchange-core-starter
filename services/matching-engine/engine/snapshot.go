package engine

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// SnapshotSchemaVersion identifies the orderbook snapshot layout.
const SnapshotSchemaVersion = 1

// OrderBookSnapshot is a full serializable copy of the book: the matching
// sequence and every resting order in price-time priority (best price
// first, FIFO within a level). Snapshot + WAL replay after the snapshot
// sequence must rebuild an identical book.
type OrderBookSnapshot struct {
	SchemaVersion int     `json:"schema_version"`
	Symbol        string  `json:"symbol"`
	Sequence      int64   `json:"sequence"`
	Bids          []Order `json:"bids"`
	Asks          []Order `json:"asks"`
}

// Snapshot copies the current book state.
func (b *OrderBook) Snapshot() OrderBookSnapshot {
	return OrderBookSnapshot{
		SchemaVersion: SnapshotSchemaVersion,
		Symbol:        b.Symbol,
		Sequence:      b.seq,
		Bids:          flattenLevels(b.bids),
		Asks:          flattenLevels(b.asks),
	}
}

// RestoreOrderBook rebuilds a book from a snapshot.
func RestoreOrderBook(snapshot OrderBookSnapshot) (*OrderBook, error) {
	if snapshot.SchemaVersion != SnapshotSchemaVersion {
		return nil, fmt.Errorf("%w: unsupported snapshot schema %d", ErrInvalidOrder, snapshot.SchemaVersion)
	}
	if snapshot.Symbol == "" || snapshot.Sequence < 0 {
		return nil, ErrInvalidOrder
	}

	book := NewOrderBook(snapshot.Symbol)
	book.seq = snapshot.Sequence
	for _, orders := range [2][]Order{snapshot.Bids, snapshot.Asks} {
		for _, order := range orders {
			if err := validateRestingOrder(snapshot.Symbol, order); err != nil {
				return nil, err
			}
			if _, exists := book.orders[order.ID]; exists {
				return nil, fmt.Errorf("%w: snapshot order %d", ErrDuplicateOrder, order.ID)
			}
			restored := order
			book.addResting(&restored)
		}
	}
	return book, nil
}

// Hash returns a stable digest of the snapshot for recovery comparison.
func (s OrderBookSnapshot) Hash() (string, error) {
	encoded, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:]), nil
}

func validateRestingOrder(symbol string, order Order) error {
	if err := validateOrder(symbol, order); err != nil {
		return err
	}
	if order.Type != Limit {
		return ErrMarketOrderResting
	}
	if order.RemainingQuantity <= 0 || order.RemainingQuantity != order.Quantity-order.ExecutedQuantity {
		return ErrInvalidOrder
	}
	if order.Sequence <= 0 {
		return ErrInvalidOrder
	}
	return nil
}

func flattenLevels(levels []priceLevel) []Order {
	var out []Order
	for _, level := range levels {
		for _, order := range level.orders {
			out = append(out, *order)
		}
	}
	return out
}

// Level is an aggregated price level for orderbook snapshots.
type Level struct {
	Price    int64
	Quantity int64
}

// Depth returns aggregated bid and ask levels, best price first.
// A non-positive limit returns all levels.
func (b *OrderBook) Depth(limit int) (bids []Level, asks []Level) {
	bids = aggregateLevels(b.bids, limit)
	asks = aggregateLevels(b.asks, limit)
	return bids, asks
}

func aggregateLevels(levels []priceLevel, limit int) []Level {
	out := make([]Level, 0, len(levels))
	for _, level := range levels {
		if limit > 0 && len(out) >= limit {
			break
		}
		var total int64
		for _, order := range level.orders {
			total += order.RemainingQuantity
		}
		if total == 0 {
			continue
		}
		out = append(out, Level{Price: level.price, Quantity: total})
	}
	return out
}
