package integration

import (
	"context"
	"errors"
	"testing"

	"github.com/exchange-core-starter/exchange-core-starter/libs/audit"
	"github.com/exchange-core-starter/exchange-core-starter/services/oms/oms"
	"github.com/exchange-core-starter/exchange-core-starter/services/oms/spotexchange"
	"github.com/exchange-core-starter/exchange-core-starter/services/wallet-gateway/wallet"
)

func newFundingFixture(t *testing.T) (*spotexchange.Exchange, *wallet.Coordinator, *wallet.Whitelist) {
	t.Helper()
	ex, err := spotexchange.New(oms.Market{
		Symbol:        "BTC-USDT",
		BaseAsset:     "BTC",
		QuoteAsset:    "USDT",
		Status:        oms.MarketTrading,
		PriceScale:    2,
		QuantityScale: 8,
		TickSize:      1_00,
		LotSize:       100_000,
	})
	if err != nil {
		t.Fatalf("spotexchange.New: %v", err)
	}

	whitelist := wallet.NewWhitelist()
	service := wallet.NewService(
		wallet.Config{MainnetEnabled: false, WithdrawalsEnabled: true},
		wallet.NewMockAdapter(),
		whitelist,
		audit.NewLog(),
	)
	return ex, wallet.NewCoordinator(service, ex), whitelist
}

func TestDepositWithdrawalLedgerLoop(t *testing.T) {
	ex, coordinator, whitelist := newFundingFixture(t)
	const user int64 = 42

	address, err := coordinator.CreateDepositAddress(context.Background(), user, "USDT", "TESTNET")
	if err != nil {
		t.Fatalf("CreateDepositAddress: %v", err)
	}
	if address.Address == "" {
		t.Fatal("empty deposit address")
	}

	// Confirmed deposit creates balanced ledger entries; duplicates are no-ops.
	if err := coordinator.ConfirmDeposit("dep-1", user, "USDT", 1_000_000); err != nil {
		t.Fatalf("ConfirmDeposit: %v", err)
	}
	if err := coordinator.ConfirmDeposit("dep-1", user, "USDT", 1_000_000); err != nil {
		t.Fatalf("duplicate ConfirmDeposit: %v", err)
	}
	assertUserBalance(t, ex, user, "USDT", 1_000_000, 0)

	// Non-whitelisted address is refused and the lock is returned.
	if _, err := coordinator.RequestWithdrawal(wallet.WithdrawalRequest{
		ID: 1, UserID: user, Asset: "USDT", Network: "TESTNET",
		Address: "unknown", Amount: 500_000, Fee: 1_000,
	}); !errors.Is(err, wallet.ErrAddressNotWhitelisted) {
		t.Fatalf("err = %v, want ErrAddressNotWhitelisted", err)
	}
	assertUserBalance(t, ex, user, "USDT", 1_000_000, 0)

	whitelist.Allow(user, "USDT", "TESTNET", "allowed-address")
	withdrawal, err := coordinator.RequestWithdrawal(wallet.WithdrawalRequest{
		ID: 2, UserID: user, Asset: "USDT", Network: "TESTNET",
		Address: "allowed-address", Amount: 500_000, Fee: 1_000,
	})
	if err != nil {
		t.Fatalf("RequestWithdrawal: %v", err)
	}
	if withdrawal.Status != wallet.WithdrawalPendingReview {
		t.Fatalf("status = %s", withdrawal.Status)
	}
	assertUserBalance(t, ex, user, "USDT", 499_000, 501_000)

	if _, err := coordinator.ApproveWithdrawal(2, "admin-1", "req-1", "manual review ok"); err != nil {
		t.Fatalf("ApproveWithdrawal: %v", err)
	}
	broadcasted, err := coordinator.BroadcastWithdrawal(2)
	if err != nil {
		t.Fatalf("BroadcastWithdrawal: %v", err)
	}
	if broadcasted.Status != wallet.WithdrawalBroadcasted || broadcasted.TxID == "" {
		t.Fatalf("broadcasted = %+v", broadcasted)
	}
	assertUserBalance(t, ex, user, "USDT", 499_000, 0)
}

func TestRejectedWithdrawalReleasesLock(t *testing.T) {
	ex, coordinator, whitelist := newFundingFixture(t)
	const user int64 = 43

	if err := coordinator.ConfirmDeposit("dep-2", user, "USDT", 300_000); err != nil {
		t.Fatalf("ConfirmDeposit: %v", err)
	}
	whitelist.Allow(user, "USDT", "TESTNET", "addr")

	if _, err := coordinator.RequestWithdrawal(wallet.WithdrawalRequest{
		ID: 7, UserID: user, Asset: "USDT", Network: "TESTNET",
		Address: "addr", Amount: 200_000, Fee: 500,
	}); err != nil {
		t.Fatalf("RequestWithdrawal: %v", err)
	}
	assertUserBalance(t, ex, user, "USDT", 99_500, 200_500)

	rejected, err := coordinator.RejectWithdrawal(7, "admin-1", "req-2", "suspicious destination")
	if err != nil {
		t.Fatalf("RejectWithdrawal: %v", err)
	}
	if rejected.Status != wallet.WithdrawalRejected {
		t.Fatalf("status = %s", rejected.Status)
	}
	assertUserBalance(t, ex, user, "USDT", 300_000, 0)
}

func TestWithdrawalInsufficientBalanceIsRefused(t *testing.T) {
	ex, coordinator, whitelist := newFundingFixture(t)
	const user int64 = 44

	if err := coordinator.ConfirmDeposit("dep-3", user, "USDT", 100); err != nil {
		t.Fatalf("ConfirmDeposit: %v", err)
	}
	whitelist.Allow(user, "USDT", "TESTNET", "addr")

	if _, err := coordinator.RequestWithdrawal(wallet.WithdrawalRequest{
		ID: 8, UserID: user, Asset: "USDT", Network: "TESTNET",
		Address: "addr", Amount: 1_000, Fee: 10,
	}); err == nil {
		t.Fatal("expected insufficient balance error")
	}
	assertUserBalance(t, ex, user, "USDT", 100, 0)
}

func assertUserBalance(t *testing.T, ex *spotexchange.Exchange, userID int64, asset string, available, locked int64) {
	t.Helper()
	for _, b := range ex.Balances(userID) {
		if b.Asset == asset {
			if b.Available != available || b.Locked != locked {
				t.Fatalf("user %d %s = available=%d locked=%d, want available=%d locked=%d",
					userID, asset, b.Available, b.Locked, available, locked)
			}
			return
		}
	}
	t.Fatalf("no %s balance for user %d", asset, userID)
}
