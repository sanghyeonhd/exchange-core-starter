package oms

import (
	"errors"
	"testing"

	"github.com/exchange-core-starter/exchange-core-starter/services/matching-engine/engine"
)

func TestValidateSpotBuyReservationIncludesTakerFee(t *testing.T) {
	reservation, err := ValidateSpotOrder(testMarket(), OrderRequest{
		OrderID:     1,
		UserID:      10,
		Symbol:      "BTC-USDT",
		Side:        engine.Buy,
		Type:        engine.Limit,
		TimeInForce: engine.GTC,
		Price:       50_000_00,
		Quantity:    10_000_000,
	})
	if err != nil {
		t.Fatalf("ValidateSpotOrder: %v", err)
	}
	if reservation.Asset != "USDT" || reservation.Amount != 500_100 {
		t.Fatalf("reservation = %+v, want USDT 500100", reservation)
	}
}

func TestValidateSpotSellReservationUsesBaseQuantity(t *testing.T) {
	reservation, err := ValidateSpotOrder(testMarket(), OrderRequest{
		OrderID:     1,
		UserID:      20,
		Symbol:      "BTC-USDT",
		Side:        engine.Sell,
		Type:        engine.Limit,
		TimeInForce: engine.GTC,
		Price:       50_000_00,
		Quantity:    10_000_000,
	})
	if err != nil {
		t.Fatalf("ValidateSpotOrder: %v", err)
	}
	if reservation.Asset != "BTC" || reservation.Amount != 10_000_000 {
		t.Fatalf("reservation = %+v, want BTC 10000000", reservation)
	}
}

func TestValidateSpotRejectsBadTick(t *testing.T) {
	_, err := ValidateSpotOrder(testMarket(), OrderRequest{
		OrderID:     1,
		UserID:      10,
		Symbol:      "BTC-USDT",
		Side:        engine.Buy,
		Type:        engine.Limit,
		TimeInForce: engine.GTC,
		Price:       50_000_01,
		Quantity:    10_000_000,
	})
	if !errors.Is(err, ErrInvalidTickSize) {
		t.Fatalf("err = %v, want %v", err, ErrInvalidTickSize)
	}
}

func TestValidateSpotRejectsMinNotional(t *testing.T) {
	market := testMarket()
	market.MinNotional = 1_000_00
	_, err := ValidateSpotOrder(market, OrderRequest{
		OrderID:     1,
		UserID:      10,
		Symbol:      "BTC-USDT",
		Side:        engine.Buy,
		Type:        engine.Limit,
		TimeInForce: engine.GTC,
		Price:       50_000_00,
		Quantity:    100_000,
	})
	if !errors.Is(err, ErrMinNotional) {
		t.Fatalf("err = %v, want %v", err, ErrMinNotional)
	}
}

func testMarket() Market {
	return Market{
		Symbol:        "BTC-USDT",
		BaseAsset:     "BTC",
		QuoteAsset:    "USDT",
		Status:        MarketTrading,
		PriceScale:    2,
		QuantityScale: 8,
		MinOrderQty:   100_000,
		MinNotional:   10_00,
		TickSize:      1_00,
		LotSize:       100_000,
		MakerFeePPM:   100,
		TakerFeePPM:   200,
	}
}
