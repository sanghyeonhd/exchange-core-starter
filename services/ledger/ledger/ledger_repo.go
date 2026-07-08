package ledger

// LedgerRepository abstracts ledger persistence so the in-memory validation
// layer and a future PostgreSQL implementation are interchangeable.
type LedgerRepository interface {
	ValidateAndStore(tx Transaction) error
	Entries() []Entry
}

// InMemoryLedger is a minimal in-memory implementation that validates
// balanced transactions and stores entries.
type InMemoryLedger struct {
	entries []Entry
	applied map[string]struct{}
}

func NewInMemoryLedger() *InMemoryLedger {
	return &InMemoryLedger{applied: make(map[string]struct{})}
}

func (l *InMemoryLedger) ValidateAndStore(tx Transaction) error {
	if err := ValidateTransaction(tx); err != nil {
		return err
	}
	if _, exists := l.applied[tx.IdempotencyKey]; exists {
		return nil
	}
	l.entries = append(l.entries, tx.Entries...)
	l.applied[tx.IdempotencyKey] = struct{}{}
	return nil
}

func (l *InMemoryLedger) Entries() []Entry {
	copied := make([]Entry, len(l.entries))
	copy(copied, l.entries)
	return copied
}

// Compile-time check.
var _ LedgerRepository = (*InMemoryLedger)(nil)
