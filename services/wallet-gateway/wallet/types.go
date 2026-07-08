package wallet

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var (
	ErrWalletDisabled        = errors.New("wallet operation disabled")
	ErrInvalidWalletRequest  = errors.New("invalid wallet request")
	ErrAddressNotWhitelisted = errors.New("address not whitelisted")
	ErrWithdrawalNotFound    = errors.New("withdrawal not found")
	ErrWithdrawalState       = errors.New("invalid withdrawal state")
)

type WithdrawalStatus string

const (
	WithdrawalRequested     WithdrawalStatus = "REQUESTED"
	WithdrawalLocked        WithdrawalStatus = "LOCKED"
	WithdrawalPendingReview WithdrawalStatus = "PENDING_REVIEW"
	WithdrawalApproved      WithdrawalStatus = "APPROVED"
	WithdrawalSigned        WithdrawalStatus = "SIGNED"
	WithdrawalBroadcasted   WithdrawalStatus = "BROADCASTED"
	WithdrawalConfirmed     WithdrawalStatus = "CONFIRMED"
	WithdrawalRejected      WithdrawalStatus = "REJECTED"
)

type Config struct {
	MainnetEnabled     bool
	WithdrawalsEnabled bool
}

type DepositAddress struct {
	UserID  int64
	Asset   string
	Network string
	Address string
	Memo    string
}

type WithdrawalRequest struct {
	ID      int64
	UserID  int64
	Asset   string
	Network string
	Address string
	Memo    string
	Amount  int64
	Fee     int64
}

type Withdrawal struct {
	Request WithdrawalRequest
	Status  WithdrawalStatus
	TxID    string
}

type Whitelist struct {
	allowed map[string]struct{}
}

func NewWhitelist() *Whitelist {
	return &Whitelist{allowed: make(map[string]struct{})}
}

func (w *Whitelist) Allow(userID int64, asset string, network string, address string) {
	w.allowed[whitelistKey(userID, asset, network, address)] = struct{}{}
}

func (w *Whitelist) IsAllowed(userID int64, asset string, network string, address string) bool {
	_, ok := w.allowed[whitelistKey(userID, asset, network, address)]
	return ok
}

type MockAdapter struct {
	mu          sync.Mutex
	withdrawals map[int64]Withdrawal
	addressSeq  int64
}

func NewMockAdapter() *MockAdapter {
	return &MockAdapter{withdrawals: make(map[int64]Withdrawal)}
}

func (m *MockAdapter) CreateDepositAddress(_ context.Context, userID int64, asset string, network string) (DepositAddress, error) {
	if userID <= 0 || asset == "" || network == "" {
		return DepositAddress{}, ErrInvalidWalletRequest
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.addressSeq++
	return DepositAddress{
		UserID:  userID,
		Asset:   asset,
		Network: network,
		Address: fmt.Sprintf("mock_%s_%s_%d_%d", network, asset, userID, m.addressSeq),
	}, nil
}

func (m *MockAdapter) StoreWithdrawal(withdrawal Withdrawal) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.withdrawals[withdrawal.Request.ID] = withdrawal
}

func (m *MockAdapter) Withdrawal(id int64) (Withdrawal, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	withdrawal, ok := m.withdrawals[id]
	return withdrawal, ok
}

func whitelistKey(userID int64, asset string, network string, address string) string {
	return fmt.Sprintf("%d:%s:%s:%s", userID, asset, network, address)
}
