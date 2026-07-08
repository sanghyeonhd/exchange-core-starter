package funding

import (
	"errors"
	"testing"

	"github.com/exchange-core-starter/exchange-core-starter/services/ledger/ledger"
)

func TestBuildDepositTransaction(t *testing.T) {
	tx, err := BuildDepositTransaction(DepositRequest{
		DepositID:         "dep-1",
		Asset:             "USDT",
		UserAccountID:     10,
		ExternalAccountID: 90,
		Amount:            5_000,
	})
	if err != nil {
		t.Fatalf("BuildDepositTransaction: %v", err)
	}
	if tx.IdempotencyKey != "deposit:dep-1" {
		t.Fatalf("idempotency key = %s", tx.IdempotencyKey)
	}
	if err := ledger.ValidateTransaction(tx); err != nil {
		t.Fatalf("ValidateTransaction: %v", err)
	}
	if len(tx.Entries) != 2 || tx.Entries[0].Debit != 5_000 || tx.Entries[1].Credit != 5_000 {
		t.Fatalf("entries = %+v", tx.Entries)
	}
}

func TestBuildDepositTransactionRejectsInvalid(t *testing.T) {
	cases := []DepositRequest{
		{},
		{DepositID: "dep", Asset: "USDT", UserAccountID: 1, ExternalAccountID: 2, Amount: 0},
		{DepositID: "dep", Asset: "", UserAccountID: 1, ExternalAccountID: 2, Amount: 10},
		{DepositID: "dep", Asset: "USDT", UserAccountID: 0, ExternalAccountID: 2, Amount: 10},
	}
	for i, req := range cases {
		if _, err := BuildDepositTransaction(req); !errors.Is(err, ErrInvalidFunding) {
			t.Fatalf("case %d: err = %v, want ErrInvalidFunding", i, err)
		}
	}
}

func TestBuildWithdrawalTransactionWithFee(t *testing.T) {
	tx, err := BuildWithdrawalTransaction(WithdrawalRequest{
		WithdrawalID:        "wd-1",
		Asset:               "BTC",
		UserAccountID:       10,
		ExternalAccountID:   90,
		FeeRevenueAccountID: 91,
		Amount:              1_000_000,
		Fee:                 10_000,
	})
	if err != nil {
		t.Fatalf("BuildWithdrawalTransaction: %v", err)
	}
	if tx.IdempotencyKey != "withdrawal:wd-1" {
		t.Fatalf("idempotency key = %s", tx.IdempotencyKey)
	}
	if len(tx.Entries) != 3 {
		t.Fatalf("entries = %+v", tx.Entries)
	}
	if tx.Entries[0].Credit != 1_010_000 {
		t.Fatalf("user credit = %d, want 1010000", tx.Entries[0].Credit)
	}
}

func TestBuildWithdrawalTransactionWithoutFee(t *testing.T) {
	tx, err := BuildWithdrawalTransaction(WithdrawalRequest{
		WithdrawalID:      "wd-2",
		Asset:             "BTC",
		UserAccountID:     10,
		ExternalAccountID: 90,
		Amount:            1_000_000,
	})
	if err != nil {
		t.Fatalf("BuildWithdrawalTransaction: %v", err)
	}
	if len(tx.Entries) != 2 {
		t.Fatalf("entries = %+v", tx.Entries)
	}
}

func TestBuildWithdrawalTransactionRequiresFeeAccount(t *testing.T) {
	_, err := BuildWithdrawalTransaction(WithdrawalRequest{
		WithdrawalID:      "wd-3",
		Asset:             "BTC",
		UserAccountID:     10,
		ExternalAccountID: 90,
		Amount:            1_000_000,
		Fee:               1,
	})
	if !errors.Is(err, ErrInvalidFunding) {
		t.Fatalf("err = %v, want ErrInvalidFunding", err)
	}
}
