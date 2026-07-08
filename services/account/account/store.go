package account

import (
	"errors"
	"fmt"
	"sync"

	"github.com/exchange-core-starter/exchange-core-starter/services/ledger/ledger"
)

var (
	ErrAccountNotFound     = errors.New("account not found")
	ErrInvalidAmount       = errors.New("invalid amount")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInvalidTransaction  = errors.New("invalid account transaction")
)

type Type string

const (
	TypeSpot       Type = "SPOT"
	TypeFeeRevenue Type = "FEE_REVENUE"
	// TypeExternal represents value outside the exchange (custody omnibus).
	// It is the double-entry counterpart of deposits and withdrawals and is
	// the only account type allowed to hold a negative available balance.
	TypeExternal Type = "EXTERNAL"
)

type Balance struct {
	AccountID int64
	UserID    int64
	Type      Type
	Asset     string
	Available int64
	Locked    int64
}

type Store struct {
	mu      sync.Mutex
	entries []ledger.Entry

	accounts map[int64]*Balance
	applied  map[string]struct{}
}

func NewStore() *Store {
	return &Store{
		accounts: make(map[int64]*Balance),
		applied:  make(map[string]struct{}),
	}
}

func (s *Store) CreateAccount(accountID, userID int64, typ Type, asset string, available int64) error {
	if accountID <= 0 || asset == "" || available < 0 {
		return ErrInvalidAmount
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.accounts[accountID]; exists {
		return fmt.Errorf("%w: duplicate account id %d", ErrInvalidTransaction, accountID)
	}
	s.accounts[accountID] = &Balance{
		AccountID: accountID,
		UserID:    userID,
		Type:      typ,
		Asset:     asset,
		Available: available,
	}
	return nil
}

func (s *Store) Reserve(accountID, amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	balance, ok := s.accounts[accountID]
	if !ok {
		return ErrAccountNotFound
	}
	if balance.Available < amount {
		return ErrInsufficientBalance
	}
	balance.Available -= amount
	balance.Locked += amount
	return nil
}

func (s *Store) Release(accountID, amount int64) error {
	if amount < 0 {
		return ErrInvalidAmount
	}
	if amount == 0 {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	balance, ok := s.accounts[accountID]
	if !ok {
		return ErrAccountNotFound
	}
	if balance.Locked < amount {
		return ErrInsufficientBalance
	}
	balance.Locked -= amount
	balance.Available += amount
	return nil
}

func (s *Store) ApplyTransaction(tx ledger.Transaction) error {
	if tx.IdempotencyKey == "" {
		return ErrInvalidTransaction
	}
	if err := ledger.ValidateTransaction(tx); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.applied[tx.IdempotencyKey]; exists {
		return nil
	}

	for _, entry := range tx.Entries {
		balance, ok := s.accounts[entry.AccountID]
		if !ok {
			return fmt.Errorf("%w: %d", ErrAccountNotFound, entry.AccountID)
		}
		if balance.Asset != entry.Asset {
			return fmt.Errorf("%w: account %d asset %s entry asset %s", ErrInvalidTransaction, entry.AccountID, balance.Asset, entry.Asset)
		}
	}

	next := make(map[int64]Balance)
	for id, balance := range s.accounts {
		next[id] = *balance
	}

	for _, entry := range tx.Entries {
		balance := next[entry.AccountID]
		if entry.Debit > 0 {
			balance.Available += entry.Debit
		}
		if entry.Credit > 0 {
			if balance.Locked >= entry.Credit {
				balance.Locked -= entry.Credit
			} else {
				remaining := entry.Credit - balance.Locked
				balance.Locked = 0
				if balance.Available < remaining && balance.Type != TypeExternal {
					return ErrInsufficientBalance
				}
				balance.Available -= remaining
			}
		}
		next[entry.AccountID] = balance
	}

	for id, balance := range next {
		current := s.accounts[id]
		current.Available = balance.Available
		current.Locked = balance.Locked
	}
	for _, entry := range tx.Entries {
		s.entries = append(s.entries, entry)
	}
	s.applied[tx.IdempotencyKey] = struct{}{}
	return nil
}

func (s *Store) Balance(accountID int64) (Balance, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	balance, ok := s.accounts[accountID]
	if !ok {
		return Balance{}, ErrAccountNotFound
	}
	return *balance, nil
}

func (s *Store) BalancesByUser(userID int64) []Balance {
	s.mu.Lock()
	defer s.mu.Unlock()

	var balances []Balance
	for _, balance := range s.accounts {
		if balance.UserID == userID {
			balances = append(balances, *balance)
		}
	}
	return balances
}

func (s *Store) Entries() []ledger.Entry {
	s.mu.Lock()
	defer s.mu.Unlock()

	copied := make([]ledger.Entry, len(s.entries))
	copy(copied, s.entries)
	return copied
}
