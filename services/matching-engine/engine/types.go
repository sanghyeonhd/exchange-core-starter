package engine

import "errors"

type Side string

const (
	Buy  Side = "BUY"
	Sell Side = "SELL"
)

type OrderType string

const (
	Limit  OrderType = "LIMIT"
	Market OrderType = "MARKET"
)

type TimeInForce string

const (
	GTC TimeInForce = "GTC"
)

type OrderStatus string

const (
	OrderAccepted        OrderStatus = "ACCEPTED"
	OrderPartiallyFilled OrderStatus = "PARTIALLY_FILLED"
	OrderFilled          OrderStatus = "FILLED"
	OrderCanceled        OrderStatus = "CANCELED"
	OrderRejected        OrderStatus = "REJECTED"
)

var (
	ErrInvalidOrder       = errors.New("invalid order")
	ErrDuplicateOrder     = errors.New("duplicate order")
	ErrOrderNotFound      = errors.New("order not found")
	ErrMarketOrderResting = errors.New("market order cannot rest")
)

type Order struct {
	ID                int64
	UserID            int64
	Symbol            string
	Side              Side
	Type              OrderType
	TimeInForce       TimeInForce
	Price             int64
	Quantity          int64
	ExecutedQuantity  int64
	RemainingQuantity int64
	Sequence          int64
	Status            OrderStatus
}

type Trade struct {
	Sequence     int64
	Symbol       string
	MakerOrderID int64
	TakerOrderID int64
	MakerUserID  int64
	TakerUserID  int64
	Side         Side
	Price        int64
	Quantity     int64
}

type MatchResult struct {
	Accepted *Order
	Trades   []Trade
	Rested   *Order
	Filled   bool
}
