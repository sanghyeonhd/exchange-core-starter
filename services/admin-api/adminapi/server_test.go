package adminapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/exchange-core-starter/exchange-core-starter/libs/audit"
	"github.com/exchange-core-starter/exchange-core-starter/services/oms/oms"
	"github.com/exchange-core-starter/exchange-core-starter/services/oms/spotexchange"
	"github.com/exchange-core-starter/exchange-core-starter/services/wallet-gateway/wallet"
)

func testSetup(t *testing.T) (*Server, *spotexchange.Exchange, *wallet.Coordinator, *audit.Log) {
	t.Helper()
	exchange, err := spotexchange.New(oms.Market{
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
	})
	if err != nil {
		t.Fatal(err)
	}
	// Seed a demo user.
	if err := exchange.Deposit("seed-1", 1, "USDT", 100_000_00); err != nil {
		t.Fatal(err)
	}
	if err := exchange.Deposit("seed-2", 1, "BTC", 1_00_000_000); err != nil {
		t.Fatal(err)
	}

	auditLog := audit.NewLog()
	walletService := wallet.NewService(
		wallet.Config{MainnetEnabled: false, WithdrawalsEnabled: true},
		wallet.NewMockAdapter(),
		wallet.NewWhitelist(),
		auditLog,
	)
	coordinator := wallet.NewCoordinator(walletService, exchange)

	server := NewServer(Deps{
		Exchange: exchange,
		Wallet:   coordinator,
		AuditLog: auditLog,
	})
	return server, exchange, coordinator, auditLog
}

func TestHealthEndpoint(t *testing.T) {
	server, _, _, _ := testSetup(t)
	req := httptest.NewRequest("GET", "/admin/v1/health", nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "ok" || body["service"] != "admin-api" {
		t.Fatalf("body = %v", body)
	}
}

func TestUsersEndpoint(t *testing.T) {
	server, _, _, _ := testSetup(t)
	req := httptest.NewRequest("GET", "/admin/v1/users", nil)
	req.Header.Set("X-ADMIN-ID", "admin1")
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	users, ok := body["users"].([]any)
	if !ok || len(users) == 0 {
		t.Fatalf("expected at least one user, got %v", body)
	}
}

func TestUserBalancesEndpoint(t *testing.T) {
	server, _, _, _ := testSetup(t)
	req := httptest.NewRequest("GET", "/admin/v1/users/1/balances", nil)
	req.Header.Set("X-ADMIN-ID", "admin1")
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestHaltAndResumeMarket(t *testing.T) {
	server, exchange, _, _ := testSetup(t)

	// Halt.
	req := httptest.NewRequest("POST", "/admin/v1/markets/BTC-USDT/halt", nil)
	req.Header.Set("X-ADMIN-ID", "admin1")
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("halt status = %d", rec.Code)
	}
	if exchange.Market().Status != oms.MarketHalted {
		t.Fatalf("market status = %s, want HALTED", exchange.Market().Status)
	}

	// Resume.
	req = httptest.NewRequest("POST", "/admin/v1/markets/BTC-USDT/resume", nil)
	req.Header.Set("X-ADMIN-ID", "admin1")
	rec = httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("resume status = %d", rec.Code)
	}
	if exchange.Market().Status != oms.MarketTrading {
		t.Fatalf("market status = %s, want TRADING", exchange.Market().Status)
	}
}

func TestAuditLogChain(t *testing.T) {
	server, _, _, auditLog := testSetup(t)

	// Halt creates an audit entry.
	req := httptest.NewRequest("POST", "/admin/v1/markets/BTC-USDT/halt", nil)
	req.Header.Set("X-ADMIN-ID", "admin1")
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	// Resume creates another.
	req = httptest.NewRequest("POST", "/admin/v1/markets/BTC-USDT/resume", nil)
	req.Header.Set("X-ADMIN-ID", "admin1")
	rec = httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	// Query audit log.
	req = httptest.NewRequest("GET", "/admin/v1/audit-log", nil)
	req.Header.Set("X-ADMIN-ID", "admin1")
	rec = httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("audit-log status = %d", rec.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if valid, ok := body["chain_valid"].(bool); !ok || !valid {
		t.Fatalf("chain_valid = %v", body["chain_valid"])
	}

	if err := auditLog.Verify(); err != nil {
		t.Fatalf("audit chain verification failed: %v", err)
	}
}

func TestUnauthenticatedRequestRejected(t *testing.T) {
	server, _, _, _ := testSetup(t)
	req := httptest.NewRequest("GET", "/admin/v1/users", nil)
	// No X-ADMIN-ID header.
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestLedgerEntries(t *testing.T) {
	server, _, _, _ := testSetup(t)
	req := httptest.NewRequest("GET", "/admin/v1/ledger/entries", nil)
	req.Header.Set("X-ADMIN-ID", "admin1")
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	// Two seed deposits should have produced ledger entries.
	count, ok := body["count"].(float64)
	if !ok || count < 4 {
		t.Fatalf("expected at least 4 ledger entries, got %v", body["count"])
	}
}

func TestOrderbookEndpoint(t *testing.T) {
	server, _, _, _ := testSetup(t)
	req := httptest.NewRequest("GET", "/admin/v1/orderbook/BTC-USDT", nil)
	req.Header.Set("X-ADMIN-ID", "admin1")
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestUnknownSymbolReturns404(t *testing.T) {
	server, _, _, _ := testSetup(t)

	for _, path := range []string{
		"/admin/v1/markets/ETH-USDT/halt",
		"/admin/v1/markets/ETH-USDT/resume",
		"/admin/v1/orderbook/ETH-USDT",
	} {
		method := "POST"
		if strings.HasPrefix(path, "/admin/v1/orderbook") {
			method = "GET"
		}
		req := httptest.NewRequest(method, path, nil)
		req.Header.Set("X-ADMIN-ID", "admin1")
		rec := httptest.NewRecorder()
		server.Handler().ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("%s %s status = %d, want 404", method, path, rec.Code)
		}
	}
}
