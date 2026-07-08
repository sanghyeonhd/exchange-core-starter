package wallet

import (
	"context"
	"errors"
	"testing"

	"github.com/exchange-core-starter/exchange-core-starter/libs/audit"
)

func TestCreateDepositAddressBlocksMainnet(t *testing.T) {
	service := NewService(Config{MainnetEnabled: true}, NewMockAdapter(), NewWhitelist(), audit.NewLog())
	_, err := service.CreateDepositAddress(context.Background(), 1, "USDT", "TRON")
	if !errors.Is(err, ErrWalletDisabled) {
		t.Fatalf("err = %v, want %v", err, ErrWalletDisabled)
	}
}

func TestWithdrawalRequiresGlobalEnableAndWhitelist(t *testing.T) {
	service := NewService(Config{WithdrawalsEnabled: false}, NewMockAdapter(), NewWhitelist(), audit.NewLog())
	_, err := service.RequestWithdrawal(WithdrawalRequest{ID: 1, UserID: 1, Asset: "USDT", Network: "TRON", Address: "T1", Amount: 100})
	if !errors.Is(err, ErrWalletDisabled) {
		t.Fatalf("err = %v, want disabled", err)
	}

	service = NewService(Config{WithdrawalsEnabled: true}, NewMockAdapter(), NewWhitelist(), audit.NewLog())
	_, err = service.RequestWithdrawal(WithdrawalRequest{ID: 1, UserID: 1, Asset: "USDT", Network: "TRON", Address: "T1", Amount: 100})
	if !errors.Is(err, ErrAddressNotWhitelisted) {
		t.Fatalf("err = %v, want whitelist block", err)
	}
}

func TestWithdrawalApprovalAndMockBroadcastAreAudited(t *testing.T) {
	auditLog := audit.NewLog()
	whitelist := NewWhitelist()
	whitelist.Allow(1, "USDT", "TRON", "T1")
	service := NewService(Config{WithdrawalsEnabled: true}, NewMockAdapter(), whitelist, auditLog)

	withdrawal, err := service.RequestWithdrawal(WithdrawalRequest{ID: 1, UserID: 1, Asset: "USDT", Network: "TRON", Address: "T1", Amount: 100, Fee: 1})
	if err != nil {
		t.Fatalf("RequestWithdrawal: %v", err)
	}
	if withdrawal.Status != WithdrawalPendingReview {
		t.Fatalf("status = %s, want %s", withdrawal.Status, WithdrawalPendingReview)
	}

	withdrawal, err = service.ApproveWithdrawal(1, "wallet-admin-1", "req-1", "test approval")
	if err != nil {
		t.Fatalf("ApproveWithdrawal: %v", err)
	}
	if withdrawal.Status != WithdrawalApproved {
		t.Fatalf("status = %s, want %s", withdrawal.Status, WithdrawalApproved)
	}

	withdrawal, err = service.MockBroadcast(1)
	if err != nil {
		t.Fatalf("MockBroadcast: %v", err)
	}
	if withdrawal.Status != WithdrawalBroadcasted || withdrawal.TxID == "" {
		t.Fatalf("withdrawal = %+v, want broadcasted with txid", withdrawal)
	}
	if err := auditLog.Verify(); err != nil {
		t.Fatalf("audit Verify: %v", err)
	}
}
