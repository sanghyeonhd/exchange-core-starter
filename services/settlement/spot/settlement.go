package spot

import (
	"errors"
	"fmt"
	"math"

	"github.com/exchange-core-starter/exchange-core-starter/services/ledger/ledger"
)

const FeeRateDenominator int64 = 1_000_000

var (
	ErrInvalidSettlement = errors.New("invalid spot settlement")
	ErrAmountOverflow    = errors.New("spot settlement amount overflow")
)

type Side string

const (
	Buy  Side = "BUY"
	Sell Side = "SELL"
)

type Request struct {
	TradeID string
	Symbol  string

	BaseAsset  string
	QuoteAsset string

	BuyerBaseAccountID   int64
	BuyerQuoteAccountID  int64
	SellerBaseAccountID  int64
	SellerQuoteAccountID int64
	FeeRevenueAccountID  int64

	Price         int64
	Quantity      int64
	QuantityScale int32

	TakerSide       Side
	MakerFeeRatePPM int64
	TakerFeeRatePPM int64
}

type Amounts struct {
	QuoteNotional int64
	MakerFee      int64
	TakerFee      int64
	BuyerFee      int64
	SellerFee     int64
}

func BuildTradeTransaction(req Request) (ledger.Transaction, Amounts, error) {
	if err := validateRequest(req); err != nil {
		return ledger.Transaction{}, Amounts{}, err
	}

	notional, err := quoteNotional(req.Price, req.Quantity, req.QuantityScale)
	if err != nil {
		return ledger.Transaction{}, Amounts{}, err
	}
	makerFee, err := fee(notional, req.MakerFeeRatePPM)
	if err != nil {
		return ledger.Transaction{}, Amounts{}, err
	}
	takerFee, err := fee(notional, req.TakerFeeRatePPM)
	if err != nil {
		return ledger.Transaction{}, Amounts{}, err
	}

	amounts := Amounts{
		QuoteNotional: notional,
		MakerFee:      makerFee,
		TakerFee:      takerFee,
	}
	if req.TakerSide == Buy {
		amounts.BuyerFee = takerFee
		amounts.SellerFee = makerFee
	} else {
		amounts.BuyerFee = makerFee
		amounts.SellerFee = takerFee
	}

	tx := ledger.Transaction{
		Type:           ledger.EntryTrade,
		ReferenceType:  "TRADE",
		ReferenceID:    req.TradeID,
		IdempotencyKey: "spot-trade:" + req.TradeID,
		Entries: []ledger.Entry{
			{
				AccountID:     req.BuyerBaseAccountID,
				Asset:         req.BaseAsset,
				Debit:         req.Quantity,
				Type:          ledger.EntryTrade,
				ReferenceType: "TRADE",
				ReferenceID:   req.TradeID,
			},
			{
				AccountID:     req.SellerBaseAccountID,
				Asset:         req.BaseAsset,
				Credit:        req.Quantity,
				Type:          ledger.EntryTrade,
				ReferenceType: "TRADE",
				ReferenceID:   req.TradeID,
			},
			{
				AccountID:     req.SellerQuoteAccountID,
				Asset:         req.QuoteAsset,
				Debit:         notional - amounts.SellerFee,
				Type:          ledger.EntryTrade,
				ReferenceType: "TRADE",
				ReferenceID:   req.TradeID,
			},
			{
				AccountID:     req.FeeRevenueAccountID,
				Asset:         req.QuoteAsset,
				Debit:         amounts.BuyerFee + amounts.SellerFee,
				Type:          ledger.EntryFee,
				ReferenceType: "TRADE",
				ReferenceID:   req.TradeID,
			},
			{
				AccountID:     req.BuyerQuoteAccountID,
				Asset:         req.QuoteAsset,
				Credit:        notional + amounts.BuyerFee,
				Type:          ledger.EntryTrade,
				ReferenceType: "TRADE",
				ReferenceID:   req.TradeID,
			},
		},
	}

	if err := ledger.ValidateTransaction(tx); err != nil {
		return ledger.Transaction{}, Amounts{}, err
	}
	return tx, amounts, nil
}

func validateRequest(req Request) error {
	if req.TradeID == "" || req.Symbol == "" || req.BaseAsset == "" || req.QuoteAsset == "" {
		return ErrInvalidSettlement
	}
	if req.BuyerBaseAccountID <= 0 || req.BuyerQuoteAccountID <= 0 || req.SellerBaseAccountID <= 0 || req.SellerQuoteAccountID <= 0 || req.FeeRevenueAccountID <= 0 {
		return ErrInvalidSettlement
	}
	if req.Price <= 0 || req.Quantity <= 0 || req.QuantityScale < 0 {
		return ErrInvalidSettlement
	}
	if req.TakerSide != Buy && req.TakerSide != Sell {
		return ErrInvalidSettlement
	}
	if req.MakerFeeRatePPM < 0 || req.TakerFeeRatePPM < 0 {
		return ErrInvalidSettlement
	}
	return nil
}

func quoteNotional(price, quantity int64, quantityScale int32) (int64, error) {
	scale, err := pow10(quantityScale)
	if err != nil {
		return 0, err
	}
	if quantity != 0 && price > math.MaxInt64/quantity {
		return 0, ErrAmountOverflow
	}
	return price * quantity / scale, nil
}

func fee(notional, ratePPM int64) (int64, error) {
	if ratePPM != 0 && notional > math.MaxInt64/ratePPM {
		return 0, ErrAmountOverflow
	}
	return notional * ratePPM / FeeRateDenominator, nil
}

func pow10(scale int32) (int64, error) {
	value := int64(1)
	for i := int32(0); i < scale; i++ {
		if value > math.MaxInt64/10 {
			return 0, fmt.Errorf("%w: scale", ErrAmountOverflow)
		}
		value *= 10
	}
	return value, nil
}
