package engine

import "sort"

type priceLevel struct {
	price  int64
	orders []*Order
}

type OrderBook struct {
	Symbol string

	bids   []priceLevel
	asks   []priceLevel
	orders map[int64]*Order
	seq    int64
}

func NewOrderBook(symbol string) *OrderBook {
	return &OrderBook{
		Symbol: symbol,
		orders: make(map[int64]*Order),
	}
}

func (b *OrderBook) Submit(order Order) (MatchResult, error) {
	if err := validateOrder(b.Symbol, order); err != nil {
		order.Status = OrderRejected
		return MatchResult{Accepted: &order}, err
	}
	if _, exists := b.orders[order.ID]; exists {
		order.Status = OrderRejected
		return MatchResult{Accepted: &order}, ErrDuplicateOrder
	}

	b.seq++
	order.Sequence = b.seq
	order.Status = OrderAccepted
	order.RemainingQuantity = order.Quantity - order.ExecutedQuantity

	result := MatchResult{Accepted: cloneOrder(&order)}
	resting := &order
	if order.Side == Buy {
		result.Trades = b.matchBuy(resting)
	} else {
		result.Trades = b.matchSell(resting)
	}

	if resting.RemainingQuantity == 0 {
		resting.Status = OrderFilled
		result.Filled = true
		return result, nil
	}
	if order.Type == Market {
		resting.Status = OrderFilled
		result.Filled = true
		return result, nil
	}

	resting.Status = OrderAccepted
	b.addResting(resting)
	result.Rested = cloneOrder(resting)
	return result, nil
}

func (b *OrderBook) Cancel(orderID int64) (*Order, error) {
	order, ok := b.orders[orderID]
	if !ok {
		return nil, ErrOrderNotFound
	}

	levels := &b.bids
	if order.Side == Sell {
		levels = &b.asks
	}

	for levelIndex := range *levels {
		level := &(*levels)[levelIndex]
		for orderIndex, candidate := range level.orders {
			if candidate.ID == orderID {
				level.orders = append(level.orders[:orderIndex], level.orders[orderIndex+1:]...)
				if len(level.orders) == 0 {
					*levels = append((*levels)[:levelIndex], (*levels)[levelIndex+1:]...)
				}
				delete(b.orders, orderID)
				order.Status = OrderCanceled
				return cloneOrder(order), nil
			}
		}
	}

	delete(b.orders, orderID)
	return nil, ErrOrderNotFound
}

func (b *OrderBook) BestBid() (int64, bool) {
	if len(b.bids) == 0 {
		return 0, false
	}
	return b.bids[0].price, true
}

func (b *OrderBook) BestAsk() (int64, bool) {
	if len(b.asks) == 0 {
		return 0, false
	}
	return b.asks[0].price, true
}

func (b *OrderBook) OpenOrder(orderID int64) (*Order, bool) {
	order, ok := b.orders[orderID]
	if !ok {
		return nil, false
	}
	return cloneOrder(order), true
}

func (b *OrderBook) matchBuy(taker *Order) []Trade {
	var trades []Trade
	for taker.RemainingQuantity > 0 && len(b.asks) > 0 {
		best := &b.asks[0]
		if taker.Type == Limit && best.price > taker.Price {
			break
		}
		trades = append(trades, b.matchLevel(taker, best)...)
		if len(best.orders) == 0 {
			b.asks = b.asks[1:]
		}
	}
	return trades
}

func (b *OrderBook) matchSell(taker *Order) []Trade {
	var trades []Trade
	for taker.RemainingQuantity > 0 && len(b.bids) > 0 {
		best := &b.bids[0]
		if taker.Type == Limit && best.price < taker.Price {
			break
		}
		trades = append(trades, b.matchLevel(taker, best)...)
		if len(best.orders) == 0 {
			b.bids = b.bids[1:]
		}
	}
	return trades
}

func (b *OrderBook) matchLevel(taker *Order, level *priceLevel) []Trade {
	var trades []Trade
	for taker.RemainingQuantity > 0 && len(level.orders) > 0 {
		maker := level.orders[0]
		qty := min64(taker.RemainingQuantity, maker.RemainingQuantity)

		b.seq++
		trades = append(trades, Trade{
			Sequence:     b.seq,
			Symbol:       b.Symbol,
			MakerOrderID: maker.ID,
			TakerOrderID: taker.ID,
			MakerUserID:  maker.UserID,
			TakerUserID:  taker.UserID,
			Side:         taker.Side,
			Price:        maker.Price,
			Quantity:     qty,
		})

		maker.ExecutedQuantity += qty
		maker.RemainingQuantity -= qty
		taker.ExecutedQuantity += qty
		taker.RemainingQuantity -= qty

		if maker.RemainingQuantity == 0 {
			maker.Status = OrderFilled
			delete(b.orders, maker.ID)
			level.orders = level.orders[1:]
		} else {
			maker.Status = OrderPartiallyFilled
		}
	}
	return trades
}

func (b *OrderBook) addResting(order *Order) {
	levels := &b.bids
	desc := true
	if order.Side == Sell {
		levels = &b.asks
		desc = false
	}

	for i := range *levels {
		if (*levels)[i].price == order.Price {
			(*levels)[i].orders = append((*levels)[i].orders, order)
			b.orders[order.ID] = order
			return
		}
	}

	*levels = append(*levels, priceLevel{price: order.Price, orders: []*Order{order}})
	sort.SliceStable(*levels, func(i, j int) bool {
		if desc {
			return (*levels)[i].price > (*levels)[j].price
		}
		return (*levels)[i].price < (*levels)[j].price
	})
	b.orders[order.ID] = order
}

func validateOrder(symbol string, order Order) error {
	if order.ID <= 0 || order.UserID <= 0 || order.Symbol != symbol || order.Quantity <= 0 {
		return ErrInvalidOrder
	}
	if order.Side != Buy && order.Side != Sell {
		return ErrInvalidOrder
	}
	if order.Type != Limit && order.Type != Market {
		return ErrInvalidOrder
	}
	if order.Type == Limit && order.Price <= 0 {
		return ErrInvalidOrder
	}
	if order.Type == Market && order.Price != 0 {
		return ErrInvalidOrder
	}
	return nil
}

func cloneOrder(order *Order) *Order {
	if order == nil {
		return nil
	}
	copied := *order
	return &copied
}

func min64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
