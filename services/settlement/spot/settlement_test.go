package spot

import (
	"testing"

	"github.com/exchange-core-starter/exchange-core-starter/services/ledger/ledger"
)

func TestBuildTradeTransactionTakerBuy(t *testing.T) {
	tx, amounts, err := BuildTradeTransaction(Request{
		TradeID:              "trd-1",
		Symbol:               "BTC-USDT",
		BaseAsset:            "BTC",
		QuoteAsset:           "USDT",
		BuyerBaseAccountID:   10,
		BuyerQuoteAccountID:  11,
		SellerBaseAccountID:  20,
		SellerQuoteAccountID: 21,
		FeeRevenueAccountID:  30,
		Price:                50_000_00,
		Quantity:             10_000_000,
		QuantityScale:        8,
		TakerSide:            Buy,
		MakerFeeRatePPM:      100,
		TakerFeeRatePPM:      200,
	})
	if err != nil {
		t.Fatalf("BuildTradeTransaction: %v", err)
	}
	if amounts.QuoteNotional != 5_000_00 {
		t.Fatalf("notional = %d, want 500000", amounts.QuoteNotional)
	}
	if amounts.BuyerFee != 100 || amounts.SellerFee != 50 {
		t.Fatalf("fees = buyer %d seller %d, want 100/50", amounts.BuyerFee, amounts.SellerFee)
	}
	assertValid(t, tx)
	assertEntry(t, tx, 10, "BTC", 10_000_000, 0)
	assertEntry(t, tx, 20, "BTC", 0, 10_000_000)
	assertEntry(t, tx, 21, "USDT", 499_950, 0)
	assertEntry(t, tx, 30, "USDT", 150, 0)
	assertEntry(t, tx, 11, "USDT", 0, 500_100)
}

func TestBuildTradeTransactionTakerSell(t *testing.T) {
	tx, amounts, err := BuildTradeTransaction(Request{
		TradeID:              "trd-2",
		Symbol:               "BTC-USDT",
		BaseAsset:            "BTC",
		QuoteAsset:           "USDT",
		BuyerBaseAccountID:   10,
		BuyerQuoteAccountID:  11,
		SellerBaseAccountID:  20,
		SellerQuoteAccountID: 21,
		FeeRevenueAccountID:  30,
		Price:                50_000_00,
		Quantity:             10_000_000,
		QuantityScale:        8,
		TakerSide:            Sell,
		MakerFeeRatePPM:      100,
		TakerFeeRatePPM:      200,
	})
	if err != nil {
		t.Fatalf("BuildTradeTransaction: %v", err)
	}
	if amounts.BuyerFee != 50 || amounts.SellerFee != 100 {
		t.Fatalf("fees = buyer %d seller %d, want 50/100", amounts.BuyerFee, amounts.SellerFee)
	}
	assertValid(t, tx)
	assertEntry(t, tx, 21, "USDT", 499_900, 0)
	assertEntry(t, tx, 30, "USDT", 150, 0)
	assertEntry(t, tx, 11, "USDT", 0, 500_050)
}

func TestRejectInvalidSettlement(t *testing.T) {
	_, _, err := BuildTradeTransaction(Request{
		TradeID:              "trd-3",
		Symbol:               "BTC-USDT",
		BaseAsset:            "BTC",
		QuoteAsset:           "USDT",
		BuyerBaseAccountID:   10,
		BuyerQuoteAccountID:  11,
		SellerBaseAccountID:  20,
		SellerQuoteAccountID: 21,
		FeeRevenueAccountID:  30,
		Price:                50_000_00,
		Quantity:             10_000_000,
		QuantityScale:        8,
		TakerSide:            "HOLD",
	})
	if err != ErrInvalidSettlement {
		t.Fatalf("err = %v, want %v", err, ErrInvalidSettlement)
	}
}

func assertValid(t *testing.T, tx ledger.Transaction) {
	t.Helper()
	if err := ledger.ValidateTransaction(tx); err != nil {
		t.Fatalf("ValidateTransaction: %v", err)
	}
}

func assertEntry(t *testing.T, tx ledger.Transaction, accountID int64, asset string, debit int64, credit int64) {
	t.Helper()
	for _, entry := range tx.Entries {
		if entry.AccountID == accountID && entry.Asset == asset && entry.Debit == debit && entry.Credit == credit {
			return
		}
	}
	t.Fatalf("entry not found: account=%d asset=%s debit=%d credit=%d in %+v", accountID, asset, debit, credit, tx.Entries)
}
