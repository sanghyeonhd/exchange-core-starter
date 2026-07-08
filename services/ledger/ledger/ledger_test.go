package ledger

import "testing"

func TestValidateBalancedTransaction(t *testing.T) {
	tx := Transaction{
		ID:             1,
		Type:           EntryTrade,
		IdempotencyKey: "trade-1",
		Entries: []Entry{
			{AccountID: 10, Asset: "USDT", Debit: 100, Type: EntryTrade},
			{AccountID: 20, Asset: "USDT", Credit: 100, Type: EntryTrade},
		},
	}

	if err := ValidateTransaction(tx); err != nil {
		t.Fatalf("ValidateTransaction: %v", err)
	}
}

func TestRejectImbalancedTransaction(t *testing.T) {
	tx := Transaction{
		ID:             1,
		Type:           EntryTrade,
		IdempotencyKey: "trade-1",
		Entries: []Entry{
			{AccountID: 10, Asset: "USDT", Debit: 100, Type: EntryTrade},
			{AccountID: 20, Asset: "USDT", Credit: 99, Type: EntryTrade},
		},
	}

	if err := ValidateTransaction(tx); err != ErrImbalancedTransaction {
		t.Fatalf("err = %v, want %v", err, ErrImbalancedTransaction)
	}
}

func TestRejectEntryWithBothDebitAndCredit(t *testing.T) {
	tx := Transaction{
		ID:             1,
		Type:           EntryTrade,
		IdempotencyKey: "trade-1",
		Entries: []Entry{
			{AccountID: 10, Asset: "USDT", Debit: 100, Credit: 1, Type: EntryTrade},
			{AccountID: 20, Asset: "USDT", Credit: 101, Type: EntryTrade},
		},
	}

	if err := ValidateTransaction(tx); err != ErrInvalidEntry {
		t.Fatalf("err = %v, want %v", err, ErrInvalidEntry)
	}
}
