//go:build postgres

// Package account PostgreSQL implementation. Build with -tags postgres and
// provide DATABASE_URL. This is a skeleton; the actual queries will be
// implemented when a PostgreSQL instance is available for integration tests.
package account

import (
	"database/sql"
	"fmt"

	"github.com/exchange-core-starter/exchange-core-starter/services/ledger/ledger"
)

// PgStore implements AccountRepository against PostgreSQL using the schema
// defined in docs/database/POSTGRES_SCHEMA.sql.
type PgStore struct {
	db *sql.DB
}

func NewPgStore(db *sql.DB) *PgStore {
	return &PgStore{db: db}
}

func (s *PgStore) CreateAccount(accountID, userID int64, typ Type, asset string, available int64) error {
	return fmt.Errorf("PgStore.CreateAccount: not yet implemented")
}

func (s *PgStore) Reserve(accountID, amount int64) error {
	return fmt.Errorf("PgStore.Reserve: not yet implemented")
}

func (s *PgStore) Release(accountID, amount int64) error {
	return fmt.Errorf("PgStore.Release: not yet implemented")
}

func (s *PgStore) ApplyTransaction(tx ledger.Transaction) error {
	return fmt.Errorf("PgStore.ApplyTransaction: not yet implemented")
}

func (s *PgStore) Balance(accountID int64) (Balance, error) {
	return Balance{}, fmt.Errorf("PgStore.Balance: not yet implemented")
}

func (s *PgStore) BalancesByUser(userID int64) []Balance {
	return nil
}

func (s *PgStore) Entries() []ledger.Entry {
	return nil
}

// Compile-time check.
var _ AccountRepository = (*PgStore)(nil)
