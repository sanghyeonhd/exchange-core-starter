package wallet

import (
	"context"
	"fmt"
)

// FundsLedger is the account/ledger side of wallet flows. It is implemented
// by the spot exchange composition so deposits and withdrawals always settle
// as balanced ledger transactions.
type FundsLedger interface {
	Deposit(depositID string, userID int64, asset string, amount int64) error
	LockWithdrawal(userID int64, asset string, amount int64) error
	ReleaseWithdrawal(userID int64, asset string, amount int64) error
	SettleWithdrawal(withdrawalID string, userID int64, asset string, amount, fee int64) error
}

// Coordinator binds wallet custody state transitions to ledger effects:
// request locks funds, reject releases them, broadcast settles them.
type Coordinator struct {
	service *Service
	ledger  FundsLedger
}

func NewCoordinator(service *Service, ledger FundsLedger) *Coordinator {
	return &Coordinator{service: service, ledger: ledger}
}

func (c *Coordinator) CreateDepositAddress(ctx context.Context, userID int64, asset string, network string) (DepositAddress, error) {
	return c.service.CreateDepositAddress(ctx, userID, asset, network)
}

// ConfirmDeposit records one confirmed on-chain deposit in the ledger.
// Duplicate confirmations of the same deposit id are no-ops.
func (c *Coordinator) ConfirmDeposit(depositID string, userID int64, asset string, amount int64) error {
	if depositID == "" || userID <= 0 || asset == "" || amount <= 0 {
		return ErrInvalidWalletRequest
	}
	return c.ledger.Deposit(depositID, userID, asset, amount)
}

// RequestWithdrawal locks amount+fee, then registers the withdrawal for
// review. The lock is released if the wallet-side request is refused.
func (c *Coordinator) RequestWithdrawal(req WithdrawalRequest) (Withdrawal, error) {
	if req.Amount <= 0 || req.Fee < 0 {
		return Withdrawal{}, ErrInvalidWalletRequest
	}
	total := req.Amount + req.Fee
	if err := c.ledger.LockWithdrawal(req.UserID, req.Asset, total); err != nil {
		return Withdrawal{}, err
	}
	withdrawal, err := c.service.RequestWithdrawal(req)
	if err != nil {
		if releaseErr := c.ledger.ReleaseWithdrawal(req.UserID, req.Asset, total); releaseErr != nil {
			return Withdrawal{}, fmt.Errorf("request failed (%w) and release failed: %v", err, releaseErr)
		}
		return Withdrawal{}, err
	}
	return withdrawal, nil
}

func (c *Coordinator) ApproveWithdrawal(id int64, adminID string, requestID string, reason string) (Withdrawal, error) {
	return c.service.ApproveWithdrawal(id, adminID, requestID, reason)
}

// RejectWithdrawal releases the locked funds after the audited rejection.
func (c *Coordinator) RejectWithdrawal(id int64, adminID string, requestID string, reason string) (Withdrawal, error) {
	withdrawal, err := c.service.RejectWithdrawal(id, adminID, requestID, reason)
	if err != nil {
		return Withdrawal{}, err
	}
	req := withdrawal.Request
	if err := c.ledger.ReleaseWithdrawal(req.UserID, req.Asset, req.Amount+req.Fee); err != nil {
		return Withdrawal{}, err
	}
	return withdrawal, nil
}

// BroadcastWithdrawal mock-broadcasts an approved withdrawal and settles the
// locked funds into ledger entries. Settlement is idempotent per withdrawal.
func (c *Coordinator) BroadcastWithdrawal(id int64) (Withdrawal, error) {
	withdrawal, err := c.service.MockBroadcast(id)
	if err != nil {
		return Withdrawal{}, err
	}
	req := withdrawal.Request
	if err := c.ledger.SettleWithdrawal(fmt.Sprintf("%d", id), req.UserID, req.Asset, req.Amount, req.Fee); err != nil {
		return Withdrawal{}, err
	}
	return withdrawal, nil
}

func (c *Coordinator) Withdrawal(id int64) (Withdrawal, bool) {
	return c.service.Withdrawal(id)
}
