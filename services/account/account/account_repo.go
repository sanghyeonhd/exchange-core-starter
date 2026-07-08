package account

import (
	"github.com/exchange-core-starter/exchange-core-starter/services/ledger/ledger"
)

// AccountRepository abstracts account balance operations so the in-memory
// Store and a future PostgreSQL implementation are interchangeable.
// The spot exchange composition and wallet coordinator depend on this
// interface, not the concrete Store.
type AccountRepository interface {
	CreateAccount(accountID, userID int64, typ Type, asset string, available int64) error
	Reserve(accountID, amount int64) error
	Release(accountID, amount int64) error
	ApplyTransaction(tx ledger.Transaction) error
	Balance(accountID int64) (Balance, error)
	BalancesByUser(userID int64) []Balance
	Entries() []ledger.Entry
}

// Compile-time check: Store implements AccountRepository.
var _ AccountRepository = (*Store)(nil)
