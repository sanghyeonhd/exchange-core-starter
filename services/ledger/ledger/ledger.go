package ledger

import "errors"

var (
	ErrNoEntries             = errors.New("ledger transaction requires entries")
	ErrImbalancedTransaction = errors.New("ledger transaction debit and credit totals differ")
	ErrInvalidEntry          = errors.New("ledger entry is invalid")
)

type EntryType string

const (
	EntryDeposit     EntryType = "DEPOSIT"
	EntryWithdrawal  EntryType = "WITHDRAWAL"
	EntryTrade       EntryType = "TRADE"
	EntryFee         EntryType = "FEE"
	EntryFunding     EntryType = "FUNDING"
	EntryRealizedPNL EntryType = "REALIZED_PNL"
	EntryLiquidation EntryType = "LIQUIDATION"
	EntryInsurance   EntryType = "INSURANCE"
	EntryTransfer    EntryType = "TRANSFER"
)

type Entry struct {
	ID             int64
	AccountID      int64
	Asset          string
	Debit          int64
	Credit         int64
	BalanceAfter   int64
	Type           EntryType
	ReferenceType  string
	ReferenceID    string
	IdempotencyKey string
}

type Transaction struct {
	ID             int64
	Type           EntryType
	ReferenceType  string
	ReferenceID    string
	IdempotencyKey string
	Entries        []Entry
}

func ValidateTransaction(tx Transaction) error {
	if len(tx.Entries) == 0 {
		return ErrNoEntries
	}

	totalsByAsset := make(map[string]struct {
		debit  int64
		credit int64
	})
	for _, entry := range tx.Entries {
		if entry.AccountID <= 0 || entry.Asset == "" {
			return ErrInvalidEntry
		}
		if entry.Debit < 0 || entry.Credit < 0 {
			return ErrInvalidEntry
		}
		if entry.Debit == 0 && entry.Credit == 0 {
			return ErrInvalidEntry
		}
		if entry.Debit > 0 && entry.Credit > 0 {
			return ErrInvalidEntry
		}
		total := totalsByAsset[entry.Asset]
		total.debit += entry.Debit
		total.credit += entry.Credit
		totalsByAsset[entry.Asset] = total
	}

	for _, total := range totalsByAsset {
		if total.debit != total.credit {
			return ErrImbalancedTransaction
		}
	}
	return nil
}
