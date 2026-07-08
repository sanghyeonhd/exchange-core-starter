package integration

import (
	"fmt"
	"testing"

	"github.com/exchange-core-starter/exchange-core-starter/services/account/account"
	"github.com/exchange-core-starter/exchange-core-starter/services/ledger/ledger"
	"github.com/exchange-core-starter/exchange-core-starter/services/matching-engine/engine"
	"github.com/exchange-core-starter/exchange-core-starter/services/oms/oms"
	spotsettlement "github.com/exchange-core-starter/exchange-core-starter/services/settlement/spot"
)

const (
	aliceUserID    int64 = 100
	bobUserID      int64 = 200
	aliceUSDT      int64 = 1001
	aliceBTC       int64 = 1002
	bobBTC         int64 = 2001
	bobUSDT        int64 = 2002
	feeRevenueUSDT int64 = 9001
)

func TestSpotOrderMatchSettlementLedgerLoop(t *testing.T) {
	market := oms.Market{
		Symbol:        "BTC-USDT",
		BaseAsset:     "BTC",
		QuoteAsset:    "USDT",
		Status:        oms.MarketTrading,
		PriceScale:    2,
		QuantityScale: 8,
		MinOrderQty:   100_000,
		MinNotional:   10_00,
		TickSize:      1_00,
		LotSize:       100_000,
		MakerFeePPM:   100,
		TakerFeePPM:   200,
	}

	accounts := account.NewStore()
	mustCreateAccount(t, accounts, aliceUSDT, aliceUserID, account.TypeSpot, "USDT", 1_000_000)
	mustCreateAccount(t, accounts, aliceBTC, aliceUserID, account.TypeSpot, "BTC", 0)
	mustCreateAccount(t, accounts, bobBTC, bobUserID, account.TypeSpot, "BTC", 100_000_000)
	mustCreateAccount(t, accounts, bobUSDT, bobUserID, account.TypeSpot, "USDT", 0)
	mustCreateAccount(t, accounts, feeRevenueUSDT, 0, account.TypeFeeRevenue, "USDT", 0)

	buy := oms.OrderRequest{
		OrderID:       1,
		UserID:        aliceUserID,
		Symbol:        "BTC-USDT",
		Side:          engine.Buy,
		Type:          engine.Limit,
		TimeInForce:   engine.GTC,
		Price:         50_000_00,
		Quantity:      10_000_000,
		ClientOrderID: "alice-buy-1",
	}
	buyReservation, err := oms.ValidateSpotOrder(market, buy)
	if err != nil {
		t.Fatalf("ValidateSpotOrder buy: %v", err)
	}
	if buyReservation.Asset != "USDT" {
		t.Fatalf("buy reservation asset = %s, want USDT", buyReservation.Asset)
	}
	mustReserve(t, accounts, aliceUSDT, buyReservation.Amount)

	sell := oms.OrderRequest{
		OrderID:       2,
		UserID:        bobUserID,
		Symbol:        "BTC-USDT",
		Side:          engine.Sell,
		Type:          engine.Limit,
		TimeInForce:   engine.GTC,
		Price:         50_000_00,
		Quantity:      10_000_000,
		ClientOrderID: "bob-sell-1",
	}
	sellReservation, err := oms.ValidateSpotOrder(market, sell)
	if err != nil {
		t.Fatalf("ValidateSpotOrder sell: %v", err)
	}
	if sellReservation.Asset != "BTC" {
		t.Fatalf("sell reservation asset = %s, want BTC", sellReservation.Asset)
	}
	mustReserve(t, accounts, bobBTC, sellReservation.Amount)

	book := engine.NewOrderBook("BTC-USDT")
	if result, err := book.Submit(oms.ToEngineOrder(buy)); err != nil || result.Rested == nil {
		t.Fatalf("submit buy result=%+v err=%v", result, err)
	}
	result, err := book.Submit(oms.ToEngineOrder(sell))
	if err != nil {
		t.Fatalf("submit sell: %v", err)
	}
	if len(result.Trades) != 1 {
		t.Fatalf("trade count = %d, want 1", len(result.Trades))
	}
	trade := result.Trades[0]
	if trade.Price != 50_000_00 || trade.Quantity != 10_000_000 || trade.Side != engine.Sell {
		t.Fatalf("trade = %+v, want taker sell 0.1 BTC at 50000.00", trade)
	}

	tx, amounts, err := spotsettlement.BuildTradeTransaction(spotsettlement.Request{
		TradeID:              fmt.Sprintf("BTC-USDT:%d", trade.Sequence),
		Symbol:               "BTC-USDT",
		BaseAsset:            "BTC",
		QuoteAsset:           "USDT",
		BuyerBaseAccountID:   aliceBTC,
		BuyerQuoteAccountID:  aliceUSDT,
		SellerBaseAccountID:  bobBTC,
		SellerQuoteAccountID: bobUSDT,
		FeeRevenueAccountID:  feeRevenueUSDT,
		Price:                trade.Price,
		Quantity:             trade.Quantity,
		QuantityScale:        market.QuantityScale,
		TakerSide:            spotsettlement.Sell,
		MakerFeeRatePPM:      market.MakerFeePPM,
		TakerFeeRatePPM:      market.TakerFeePPM,
	})
	if err != nil {
		t.Fatalf("BuildTradeTransaction: %v", err)
	}

	if err := ledger.ValidateTransaction(tx); err != nil {
		t.Fatalf("ValidateTransaction: %v", err)
	}
	if err := accounts.ApplyTransaction(tx); err != nil {
		t.Fatalf("ApplyTransaction: %v", err)
	}

	actualBuyerQuoteDebit := amounts.QuoteNotional + amounts.BuyerFee
	if err := accounts.Release(aliceUSDT, buyReservation.Amount-actualBuyerQuoteDebit); err != nil {
		t.Fatalf("Release buyer excess quote reservation: %v", err)
	}

	assertBalance(t, accounts, aliceUSDT, 499_950, 0)
	assertBalance(t, accounts, aliceBTC, 10_000_000, 0)
	assertBalance(t, accounts, bobBTC, 90_000_000, 0)
	assertBalance(t, accounts, bobUSDT, 499_900, 0)
	assertBalance(t, accounts, feeRevenueUSDT, 150, 0)

	if err := accounts.ApplyTransaction(tx); err != nil {
		t.Fatalf("ApplyTransaction duplicate: %v", err)
	}
	assertBalance(t, accounts, aliceUSDT, 499_950, 0)
	assertBalance(t, accounts, feeRevenueUSDT, 150, 0)
}

func mustCreateAccount(t *testing.T, store *account.Store, accountID int64, userID int64, typ account.Type, asset string, available int64) {
	t.Helper()
	if err := store.CreateAccount(accountID, userID, typ, asset, available); err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}
}

func mustReserve(t *testing.T, store *account.Store, accountID int64, amount int64) {
	t.Helper()
	if err := store.Reserve(accountID, amount); err != nil {
		t.Fatalf("Reserve account=%d amount=%d: %v", accountID, amount, err)
	}
}

func assertBalance(t *testing.T, store *account.Store, accountID int64, available int64, locked int64) {
	t.Helper()
	balance, err := store.Balance(accountID)
	if err != nil {
		t.Fatalf("Balance(%d): %v", accountID, err)
	}
	if balance.Available != available || balance.Locked != locked {
		t.Fatalf("Balance(%d) = available=%d locked=%d, want available=%d locked=%d", accountID, balance.Available, balance.Locked, available, locked)
	}
}
