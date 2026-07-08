package funding

import (
	"errors"

	"github.com/exchange-core-starter/exchange-core-starter/services/ledger/ledger"
)

var ErrInvalidFunding = errors.New("invalid funding settlement")

// DepositRequest converts one confirmed deposit into one ledger transaction.
// The external account is the custody omnibus counterpart: asset entering a
// user account must leave the external account so every balance change stays
// explainable by balanced ledger entries.
type DepositRequest struct {
	DepositID         string
	Asset             string
	UserAccountID     int64
	ExternalAccountID int64
	Amount            int64
}

// WithdrawalRequest converts one approved withdrawal into one ledger
// transaction. Amount is what leaves the exchange to the external address;
// Fee is retained by the exchange as fee revenue. The user account is
// expected to hold Amount+Fee locked before settlement.
type WithdrawalRequest struct {
	WithdrawalID        string
	Asset               string
	UserAccountID       int64
	ExternalAccountID   int64
	FeeRevenueAccountID int64
	Amount              int64
	Fee                 int64
}

func BuildDepositTransaction(req DepositRequest) (ledger.Transaction, error) {
	if req.DepositID == "" || req.Asset == "" || req.UserAccountID <= 0 || req.ExternalAccountID <= 0 || req.Amount <= 0 {
		return ledger.Transaction{}, ErrInvalidFunding
	}

	tx := ledger.Transaction{
		Type:           ledger.EntryDeposit,
		ReferenceType:  "DEPOSIT",
		ReferenceID:    req.DepositID,
		IdempotencyKey: "deposit:" + req.DepositID,
		Entries: []ledger.Entry{
			{
				AccountID:     req.UserAccountID,
				Asset:         req.Asset,
				Debit:         req.Amount,
				Type:          ledger.EntryDeposit,
				ReferenceType: "DEPOSIT",
				ReferenceID:   req.DepositID,
			},
			{
				AccountID:     req.ExternalAccountID,
				Asset:         req.Asset,
				Credit:        req.Amount,
				Type:          ledger.EntryDeposit,
				ReferenceType: "DEPOSIT",
				ReferenceID:   req.DepositID,
			},
		},
	}

	if err := ledger.ValidateTransaction(tx); err != nil {
		return ledger.Transaction{}, err
	}
	return tx, nil
}

func BuildWithdrawalTransaction(req WithdrawalRequest) (ledger.Transaction, error) {
	if req.WithdrawalID == "" || req.Asset == "" || req.UserAccountID <= 0 || req.ExternalAccountID <= 0 || req.Amount <= 0 || req.Fee < 0 {
		return ledger.Transaction{}, ErrInvalidFunding
	}
	if req.Fee > 0 && req.FeeRevenueAccountID <= 0 {
		return ledger.Transaction{}, ErrInvalidFunding
	}

	entries := []ledger.Entry{
		{
			AccountID:     req.UserAccountID,
			Asset:         req.Asset,
			Credit:        req.Amount + req.Fee,
			Type:          ledger.EntryWithdrawal,
			ReferenceType: "WITHDRAWAL",
			ReferenceID:   req.WithdrawalID,
		},
		{
			AccountID:     req.ExternalAccountID,
			Asset:         req.Asset,
			Debit:         req.Amount,
			Type:          ledger.EntryWithdrawal,
			ReferenceType: "WITHDRAWAL",
			ReferenceID:   req.WithdrawalID,
		},
	}
	if req.Fee > 0 {
		entries = append(entries, ledger.Entry{
			AccountID:     req.FeeRevenueAccountID,
			Asset:         req.Asset,
			Debit:         req.Fee,
			Type:          ledger.EntryFee,
			ReferenceType: "WITHDRAWAL",
			ReferenceID:   req.WithdrawalID,
		})
	}

	tx := ledger.Transaction{
		Type:           ledger.EntryWithdrawal,
		ReferenceType:  "WITHDRAWAL",
		ReferenceID:    req.WithdrawalID,
		IdempotencyKey: "withdrawal:" + req.WithdrawalID,
		Entries:        entries,
	}

	if err := ledger.ValidateTransaction(tx); err != nil {
		return ledger.Transaction{}, err
	}
	return tx, nil
}
