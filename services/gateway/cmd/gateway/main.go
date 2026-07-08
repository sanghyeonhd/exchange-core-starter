package main

import (
	"log"
	"net/http"
	"os"

	"github.com/exchange-core-starter/exchange-core-starter/libs/audit"
	"github.com/exchange-core-starter/exchange-core-starter/services/admin-api/adminapi"
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

	privateData := marketdata.NewPrivateHub()
	exchange.SetPrivateData(privateData)

	// Ticker and kline aggregators feed from matching trades and publish
	// through the same market data hub.
	market := exchange.Market()
	tickerAgg := marketdata.NewTickerAggregator(marketData, market.Symbol, market.PriceScale, market.QuantityScale)
	klineAgg := marketdata.NewKlineAggregator(marketData, market.Symbol, market.PriceScale, market.QuantityScale)
	exchange.SetAggregators(tickerAgg, klineAgg)

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
	auditLog := audit.NewLog()
	walletService := wallet.NewService(
		wallet.Config{MainnetEnabled: false, WithdrawalsEnabled: false},
		wallet.NewMockAdapter(),
		wallet.NewWhitelist(),
		auditLog,
	)
	coordinator := wallet.NewCoordinator(walletService, exchange)

	// Public gateway on :8080.
	server := httpapi.NewServer(httpapi.Deps{
		Exchange:    exchange,
		Wallet:      coordinator,
		MarketData:  marketData,
		PrivateData: privateData,
	})

	// Admin API on a separate port (default :8081, override with ADMIN_PORT).
	adminPort := os.Getenv("ADMIN_PORT")
	if adminPort == "" {
		adminPort = ":8081"
	}
	adminServer := adminapi.NewServer(adminapi.Deps{
		Exchange: exchange,
		Wallet:   coordinator,
		AuditLog: auditLog,
	})
	go func() {
		log.Printf("admin API listening on %s (dev auth: X-ADMIN-ID header)", adminPort)
		if err := http.ListenAndServe(adminPort, adminServer.Handler()); err != nil {
			log.Fatalf("admin API: %v", err)
		}
	}()

	log.Println("gateway listening on :8080 (dev auth: X-USER-ID header, demo users 1 and 2 seeded, ws at /ws/public and /ws/private)")
	if err := http.ListenAndServe(":8080", server.Handler()); err != nil {
		log.Fatal(err)
	}
}
