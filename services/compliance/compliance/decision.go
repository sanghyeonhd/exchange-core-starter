package compliance

import (
	"strings"

	"github.com/exchange-core-starter/exchange-core-starter/services/aml/aml"
	"github.com/exchange-core-starter/exchange-core-starter/services/kyc/kyc"
	"github.com/exchange-core-starter/exchange-core-starter/services/listing/listing"
)

type Action string

const (
	ActionTrade    Action = "TRADE"
	ActionDeposit  Action = "DEPOSIT"
	ActionWithdraw Action = "WITHDRAW"
)

type Request struct {
	Action         Action
	UserID         int64
	Market         string
	Asset          string
	Network        string
	CounterpartyID string
	Amount         int64
}

type Decision struct {
	Allowed bool
	Reasons []string
}

type Inputs struct {
	KYC                        kyc.Profile
	AML                        aml.Profile
	Listing                    listing.Policy
	HighRiskAmount             int64
	WithdrawalsGloballyEnabled bool
}

func Decide(req Request, inputs Inputs) Decision {
	var reasons []string

	switch req.Action {
	case ActionTrade:
		if !inputs.KYC.CanTrade() {
			reasons = append(reasons, "KYC_TRADE_BLOCKED")
		}
		if !inputs.AML.CanTrade() {
			reasons = append(reasons, "AML_TRADE_BLOCKED")
		}
		if !inputs.Listing.CanTrade(req.Market) {
			reasons = append(reasons, "MARKET_NOT_LISTED")
		}
	case ActionDeposit:
		if !inputs.Listing.CanDeposit(req.Asset, req.Network) {
			reasons = append(reasons, "ASSET_NETWORK_NOT_LISTED")
		}
	case ActionWithdraw:
		if !inputs.WithdrawalsGloballyEnabled {
			reasons = append(reasons, "WITHDRAWALS_DISABLED")
		}
		if !inputs.KYC.CanWithdraw() {
			reasons = append(reasons, "KYC_WITHDRAW_BLOCKED")
		}
		if !inputs.AML.CanWithdraw() {
			reasons = append(reasons, "AML_WITHDRAW_BLOCKED")
		}
		if aml.NeedsCase(req.Amount, inputs.HighRiskAmount, inputs.AML) {
			reasons = append(reasons, "AML_CASE_REQUIRED")
		}
		if !inputs.Listing.CanWithdraw(req.Asset, req.Network, req.CounterpartyID) {
			reasons = append(reasons, "WITHDRAWAL_WHITELIST_BLOCKED")
		}
	default:
		reasons = append(reasons, "UNKNOWN_ACTION")
	}

	return Decision{Allowed: len(reasons) == 0, Reasons: reasons}
}

func (d Decision) Code() string {
	if d.Allowed {
		return "ALLOWED"
	}
	return strings.Join(d.Reasons, ",")
}
