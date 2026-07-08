package oms

import (
	"errors"
	"fmt"
	"math"

	"github.com/exchange-core-starter/exchange-core-starter/services/matching-engine/engine"
	spotsettlement "github.com/exchange-core-starter/exchange-core-starter/services/settlement/spot"
)

var (
	ErrInvalidMarket    = errors.New("invalid market")
	ErrMarketNotTrading = errors.New("market not trading")
	ErrInvalidOrder     = errors.New("invalid order")
	ErrInvalidTickSize  = errors.New("invalid tick size")
	ErrInvalidLotSize   = errors.New("invalid lot size")
	ErrMinNotional      = errors.New("min notional not met")
	ErrAmountOverflow   = errors.New("order amount overflow")
)

type MarketStatus string

const (
	MarketTrading MarketStatus = "TRADING"
	MarketHalted  MarketStatus = "HALTED"
)

type Market struct {
	Symbol        string
	BaseAsset     string
	QuoteAsset    string
	Status        MarketStatus
	PriceScale    int32
	QuantityScale int32
	MinOrderQty   int64
	MinNotional   int64
	TickSize      int64
	LotSize       int64
	MakerFeePPM   int64
	TakerFeePPM   int64
}

type OrderRequest struct {
	OrderID       int64
	UserID        int64
	Symbol        string
	Side          engine.Side
	Type          engine.OrderType
	TimeInForce   engine.TimeInForce
	Price         int64
	Quantity      int64
	ClientOrderID string
}

type Reservation struct {
	Asset  string
	Amount int64
}

func ValidateSpotOrder(market Market, req OrderRequest) (Reservation, error) {
	if err := validateMarket(market); err != nil {
		return Reservation{}, err
	}
	if req.OrderID <= 0 || req.UserID <= 0 || req.Symbol != market.Symbol || req.Quantity <= 0 {
		return Reservation{}, ErrInvalidOrder
	}
	if req.Side != engine.Buy && req.Side != engine.Sell {
		return Reservation{}, ErrInvalidOrder
	}
	if req.Type != engine.Limit && req.Type != engine.Market {
		return Reservation{}, ErrInvalidOrder
	}
	if req.Type == engine.Limit && req.Price <= 0 {
		return Reservation{}, ErrInvalidOrder
	}
	if req.Type == engine.Market {
		return Reservation{}, fmt.Errorf("%w: market orders need quote quantity controls before MVP support", ErrInvalidOrder)
	}
	if market.MinOrderQty > 0 && req.Quantity < market.MinOrderQty {
		return Reservation{}, ErrInvalidOrder
	}
	if market.TickSize > 0 && req.Price%market.TickSize != 0 {
		return Reservation{}, ErrInvalidTickSize
	}
	if market.LotSize > 0 && req.Quantity%market.LotSize != 0 {
		return Reservation{}, ErrInvalidLotSize
	}

	notional, err := notional(req.Price, req.Quantity, market.QuantityScale)
	if err != nil {
		return Reservation{}, err
	}
	if market.MinNotional > 0 && notional < market.MinNotional {
		return Reservation{}, ErrMinNotional
	}

	if req.Side == engine.Sell {
		return Reservation{Asset: market.BaseAsset, Amount: req.Quantity}, nil
	}

	amount, err := BuyReservationAmount(market, req.Price, req.Quantity)
	if err != nil {
		return Reservation{}, err
	}
	return Reservation{Asset: market.QuoteAsset, Amount: amount}, nil
}

// BuyReservationAmount returns the quote amount a buy order must lock for the
// given price and quantity: notional plus the taker-fee upper bound.
func BuyReservationAmount(market Market, price, quantity int64) (int64, error) {
	notionalValue, err := notional(price, quantity, market.QuantityScale)
	if err != nil {
		return 0, err
	}
	feeValue, err := fee(notionalValue, market.TakerFeePPM)
	if err != nil {
		return 0, err
	}
	return notionalValue + feeValue, nil
}

func ToEngineOrder(req OrderRequest) engine.Order {
	return engine.Order{
		ID:          req.OrderID,
		UserID:      req.UserID,
		Symbol:      req.Symbol,
		Side:        req.Side,
		Type:        req.Type,
		TimeInForce: req.TimeInForce,
		Price:       req.Price,
		Quantity:    req.Quantity,
	}
}

func validateMarket(market Market) error {
	if market.Symbol == "" || market.BaseAsset == "" || market.QuoteAsset == "" {
		return ErrInvalidMarket
	}
	if market.Status != MarketTrading {
		return ErrMarketNotTrading
	}
	if market.QuantityScale < 0 || market.PriceScale < 0 {
		return ErrInvalidMarket
	}
	if market.MakerFeePPM < 0 || market.TakerFeePPM < 0 {
		return ErrInvalidMarket
	}
	return nil
}

func notional(price, quantity int64, quantityScale int32) (int64, error) {
	scale, err := pow10(quantityScale)
	if err != nil {
		return 0, err
	}
	if quantity != 0 && price > math.MaxInt64/quantity {
		return 0, ErrAmountOverflow
	}
	return price * quantity / scale, nil
}

func fee(notionalValue, ratePPM int64) (int64, error) {
	if ratePPM != 0 && notionalValue > math.MaxInt64/ratePPM {
		return 0, ErrAmountOverflow
	}
	return notionalValue * ratePPM / spotsettlement.FeeRateDenominator, nil
}

func pow10(scale int32) (int64, error) {
	value := int64(1)
	for i := int32(0); i < scale; i++ {
		if value > math.MaxInt64/10 {
			return 0, ErrAmountOverflow
		}
		value *= 10
	}
	return value, nil
}
