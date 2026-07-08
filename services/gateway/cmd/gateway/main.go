package main

import (
	"log"
	"net/http"
	"os"

	"github.com/exchange-core-starter/exchange-core-starter/libs/audit"
	"github.com/exchange-core-starter/exchange-core-starter/services/gateway/httpapi"
	"github.com/exchange-core-starter/exchange-core-starter/services/market-data/marketdata"
	"github.com/exchange-core-starter/exchange-core-starter/services/matching-engine/wal"
	"github.com/exchange-core-starter/exchange-core-starter/services/oms/oms"
	"github.com/exchange-core-starter/exchange-core-starter/services/oms/spotexchange"
	"github.com/exchange-core-starter/exchange-core-starter/services/wallet-gateway/wallet"
)

func main() {
	exchange, err := spotexchange.New(oms.Market{
		Symbol:        "BTC-USDT",
		BaseAsset:     "BTC",
		QuoteAsset:    "USDT",
		Status:        oms.MarketTrading,
		PriceScale:    2,
		QuantityScale: 8,
		MinOrderQty:   100_000, // 0.001 BTC
		MinNotional:   10_00,   // 10.00 USDT
		TickSize:      1_00,    // 1.00 USDT
		LotSize:       100_000, // 0.001 BTC
		MakerFeePPM:   100,
		TakerFeePPM:   200,
	})
	if err != nil {
		log.Fatal(err)
	}

	marketData := marketdata.NewHub()
	exchange.SetMarketData(marketData)

	// MATCHING_WAL_PATH enables the write-ahead command log; accepted
	// matching commands are fsynced to this file before they reach the book.
	if walPath := os.Getenv("MATCHING_WAL_PATH"); walPath != "" {
		journal, err := wal.OpenFileLog(walPath)
		if err != nil {
			log.Fatalf("open matching WAL %s: %v", walPath, err)
		}
		defer journal.Close()
		exchange.SetJournal(journal)
		log.Printf("matching WAL enabled at %s (resuming after sequence %d)", walPath, journal.LastSequence())
	}

	// Local development seed: mock deposits for demo users 1 and 2.
	// Every balance originates from a balanced deposit ledger transaction.
	seed := []struct {
		depositID string
		userID    int64
		asset     string
		amount    int64
	}{
		{"seed-user1-usdt", 1, "USDT", 100_000_00}, // 100000.00 USDT
		{"seed-user1-btc", 1, "BTC", 1_00_000_000}, // 1 BTC
		{"seed-user2-usdt", 2, "USDT", 100_000_00}, // 100000.00 USDT
		{"seed-user2-btc", 2, "BTC", 1_00_000_000}, // 1 BTC
	}
	for _, deposit := range seed {
		if err := exchange.Deposit(deposit.depositID, deposit.userID, deposit.asset, deposit.amount); err != nil {
			log.Fatalf("seed deposit %s: %v", deposit.depositID, err)
		}
	}

	// Withdrawals stay disabled by default; mainnet stays disabled.
	walletService := wallet.NewService(
		wallet.Config{MainnetEnabled: false, WithdrawalsEnabled: false},
		wallet.NewMockAdapter(),
		wallet.NewWhitelist(),
		audit.NewLog(),
	)
	coordinator := wallet.NewCoordinator(walletService, exchange)

	server := httpapi.NewServer(httpapi.Deps{
		Exchange:   exchange,
		Wallet:     coordinator,
		MarketData: marketData,
	})

	log.Println("gateway listening on :8080 (dev auth: X-USER-ID header, demo users 1 and 2 seeded, ws at /ws/public)")
	if err := http.ListenAndServe(":8080", server.Handler()); err != nil {
		log.Fatal(err)
	}
}
