//go:build postgres

// Package ledger PostgreSQL implementation. Build with -tags postgres and
// provide DATABASE_URL. This is a skeleton.
package ledger

import (
	"database/sql"
	"fmt"
)

// PgLedgerStore implements LedgerRepository against PostgreSQL.
type PgLedgerStore struct {
	db *sql.DB
}

func NewPgLedgerStore(db *sql.DB) *PgLedgerStore {
	return &PgLedgerStore{db: db}
}

func (s *PgLedgerStore) ValidateAndStore(tx Transaction) error {
	return fmt.Errorf("PgLedgerStore.ValidateAndStore: not yet implemented")
}

func (s *PgLedgerStore) Entries() []Entry {
	return nil
}

// Compile-time check.
var _ LedgerRepository = (*PgLedgerStore)(nil)
