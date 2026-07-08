package wallet

import (
	"context"
	"fmt"

	"github.com/exchange-core-starter/exchange-core-starter/libs/audit"
)

type Service struct {
	config    Config
	adapter   *MockAdapter
	whitelist *Whitelist
	auditLog  *audit.Log
}

func NewService(config Config, adapter *MockAdapter, whitelist *Whitelist, auditLog *audit.Log) *Service {
	return &Service{config: config, adapter: adapter, whitelist: whitelist, auditLog: auditLog}
}

func (s *Service) CreateDepositAddress(ctx context.Context, userID int64, asset string, network string) (DepositAddress, error) {
	if s.config.MainnetEnabled {
		return DepositAddress{}, fmt.Errorf("%w: mainnet wallet requires production signer review", ErrWalletDisabled)
	}
	return s.adapter.CreateDepositAddress(ctx, userID, asset, network)
}

func (s *Service) RequestWithdrawal(req WithdrawalRequest) (Withdrawal, error) {
	if !s.config.WithdrawalsEnabled {
		return Withdrawal{}, fmt.Errorf("%w: withdrawals disabled", ErrWalletDisabled)
	}
	if req.ID <= 0 || req.UserID <= 0 || req.Asset == "" || req.Network == "" || req.Address == "" || req.Amount <= 0 || req.Fee < 0 {
		return Withdrawal{}, ErrInvalidWalletRequest
	}
	if !s.whitelist.IsAllowed(req.UserID, req.Asset, req.Network, req.Address) {
		return Withdrawal{}, ErrAddressNotWhitelisted
	}
	withdrawal := Withdrawal{Request: req, Status: WithdrawalPendingReview}
	s.adapter.StoreWithdrawal(withdrawal)
	return withdrawal, nil
}

func (s *Service) ApproveWithdrawal(id int64, adminID string, requestID string, reason string) (Withdrawal, error) {
	withdrawal, ok := s.adapter.Withdrawal(id)
	if !ok {
		return Withdrawal{}, ErrWithdrawalNotFound
	}
	if withdrawal.Status != WithdrawalPendingReview {
		return Withdrawal{}, ErrWithdrawalState
	}
	if _, err := s.auditLog.Append(audit.Event{
		ID:           fmt.Sprintf("withdrawal-approval-%d", id),
		ActorID:      adminID,
		ActorType:    "ADMIN",
		Action:       "WITHDRAWAL_APPROVE",
		ResourceType: "WITHDRAWAL",
		ResourceID:   fmt.Sprintf("%d", id),
		RequestID:    requestID,
		Reason:       reason,
	}); err != nil {
		return Withdrawal{}, err
	}
	withdrawal.Status = WithdrawalApproved
	s.adapter.StoreWithdrawal(withdrawal)
	return withdrawal, nil
}

func (s *Service) RejectWithdrawal(id int64, adminID string, requestID string, reason string) (Withdrawal, error) {
	withdrawal, ok := s.adapter.Withdrawal(id)
	if !ok {
		return Withdrawal{}, ErrWithdrawalNotFound
	}
	if withdrawal.Status != WithdrawalPendingReview {
		return Withdrawal{}, ErrWithdrawalState
	}
	if _, err := s.auditLog.Append(audit.Event{
		ID:           fmt.Sprintf("withdrawal-reject-%d", id),
		ActorID:      adminID,
		ActorType:    "ADMIN",
		Action:       "WITHDRAWAL_REJECT",
		ResourceType: "WITHDRAWAL",
		ResourceID:   fmt.Sprintf("%d", id),
		RequestID:    requestID,
		Reason:       reason,
	}); err != nil {
		return Withdrawal{}, err
	}
	withdrawal.Status = WithdrawalRejected
	s.adapter.StoreWithdrawal(withdrawal)
	return withdrawal, nil
}

func (s *Service) Withdrawal(id int64) (Withdrawal, bool) {
	return s.adapter.Withdrawal(id)
}

func (s *Service) MockBroadcast(id int64) (Withdrawal, error) {
	withdrawal, ok := s.adapter.Withdrawal(id)
	if !ok {
		return Withdrawal{}, ErrWithdrawalNotFound
	}
	if withdrawal.Status != WithdrawalApproved {
		return Withdrawal{}, ErrWithdrawalState
	}
	withdrawal.Status = WithdrawalBroadcasted
	withdrawal.TxID = fmt.Sprintf("mock_tx_%d", id)
	s.adapter.StoreWithdrawal(withdrawal)
	return withdrawal, nil
}
