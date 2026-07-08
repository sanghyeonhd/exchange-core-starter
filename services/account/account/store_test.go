package account

import (
	"errors"
	"testing"

	"github.com/exchange-core-starter/exchange-core-starter/services/ledger/ledger"
)

func TestReserveMovesAvailableToLocked(t *testing.T) {
	store := NewStore()
	if err := store.CreateAccount(1, 10, TypeSpot, "USDT", 1_000); err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}
	if err := store.Reserve(1, 400); err != nil {
		t.Fatalf("Reserve: %v", err)
	}

	balance, err := store.Balance(1)
	if err != nil {
		t.Fatalf("Balance: %v", err)
	}
	if balance.Available != 600 || balance.Locked != 400 {
		t.Fatalf("balance = %+v, want available=600 locked=400", balance)
	}
}

func TestApplyTransactionIsIdempotent(t *testing.T) {
	store := NewStore()
	mustCreate(t, store, 1, 10, TypeSpot, "USDT", 1_000)
	mustCreate(t, store, 2, 20, TypeSpot, "USDT", 0)
	if err := store.Reserve(1, 100); err != nil {
		t.Fatalf("Reserve: %v", err)
	}

	tx := ledger.Transaction{
		Type:           ledger.EntryTransfer,
		IdempotencyKey: "tx-1",
		Entries: []ledger.Entry{
			{AccountID: 2, Asset: "USDT", Debit: 100, Type: ledger.EntryTransfer},
			{AccountID: 1, Asset: "USDT", Credit: 100, Type: ledger.EntryTransfer},
		},
	}
	if err := store.ApplyTransaction(tx); err != nil {
		t.Fatalf("ApplyTransaction: %v", err)
	}
	if err := store.ApplyTransaction(tx); err != nil {
		t.Fatalf("ApplyTransaction duplicate: %v", err)
	}

	from, _ := store.Balance(1)
	to, _ := store.Balance(2)
	if from.Available != 900 || from.Locked != 0 || to.Available != 100 {
		t.Fatalf("from=%+v to=%+v, want idempotent single apply", from, to)
	}
	if len(store.Entries()) != 2 {
		t.Fatalf("entry count = %d, want 2", len(store.Entries()))
	}
}

func TestReleaseMovesLockedToAvailable(t *testing.T) {
	store := NewStore()
	mustCreate(t, store, 1, 10, TypeSpot, "USDT", 1_000)
	if err := store.Reserve(1, 400); err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	if err := store.Release(1, 150); err != nil {
		t.Fatalf("Release: %v", err)
	}

	balance, err := store.Balance(1)
	if err != nil {
		t.Fatalf("Balance: %v", err)
	}
	if balance.Available != 750 || balance.Locked != 250 {
		t.Fatalf("balance = %+v, want available=750 locked=250", balance)
	}
}

func TestApplyRejectsInsufficientBalance(t *testing.T) {
	store := NewStore()
	mustCreate(t, store, 1, 10, TypeSpot, "USDT", 50)
	mustCreate(t, store, 2, 20, TypeSpot, "USDT", 0)

	tx := ledger.Transaction{
		Type:           ledger.EntryTransfer,
		IdempotencyKey: "tx-1",
		Entries: []ledger.Entry{
			{AccountID: 2, Asset: "USDT", Debit: 100, Type: ledger.EntryTransfer},
			{AccountID: 1, Asset: "USDT", Credit: 100, Type: ledger.EntryTransfer},
		},
	}
	if err := store.ApplyTransaction(tx); !errors.Is(err, ErrInsufficientBalance) {
		t.Fatalf("err = %v, want %v", err, ErrInsufficientBalance)
	}
}

func mustCreate(t *testing.T, store *Store, accountID int64, userID int64, typ Type, asset string, available int64) {
	t.Helper()
	if err := store.CreateAccount(accountID, userID, typ, asset, available); err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}
}
