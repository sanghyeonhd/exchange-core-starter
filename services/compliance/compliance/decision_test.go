package compliance

import (
	"testing"

	"github.com/exchange-core-starter/exchange-core-starter/services/aml/aml"
	"github.com/exchange-core-starter/exchange-core-starter/services/kyc/kyc"
	"github.com/exchange-core-starter/exchange-core-starter/services/listing/listing"
)

func TestTradeAllowedWhenKYCAMLAndMarketListed(t *testing.T) {
	decision := Decide(Request{Action: ActionTrade, UserID: 1, Market: "BTC-USDT"}, validInputs())
	if !decision.Allowed {
		t.Fatalf("decision = %+v, want allowed", decision)
	}
}

func TestWithdrawalBlockedByMultipleControls(t *testing.T) {
	inputs := validInputs()
	inputs.KYC.Level = 1
	inputs.AML.OpenCase = true
	inputs.WithdrawalsGloballyEnabled = false

	decision := Decide(Request{
		Action:         ActionWithdraw,
		UserID:         1,
		Asset:          "USDT",
		Network:        "TRON",
		CounterpartyID: "vasp-1",
		Amount:         1_000_000,
	}, inputs)
	if decision.Allowed {
		t.Fatal("withdrawal should be blocked")
	}
	wantReasons := map[string]bool{
		"WITHDRAWALS_DISABLED": true,
		"KYC_WITHDRAW_BLOCKED": true,
		"AML_WITHDRAW_BLOCKED": true,
	}
	for _, reason := range decision.Reasons {
		delete(wantReasons, reason)
	}
	if len(wantReasons) != 0 {
		t.Fatalf("missing reasons: %+v in decision %+v", wantReasons, decision)
	}
}

func validInputs() Inputs {
	return Inputs{
		KYC: kyc.Profile{UserID: 1, Level: 2, Status: kyc.StatusApproved},
		AML: aml.Profile{UserID: 1, RiskLevel: aml.RiskLow},
		Listing: listing.Policy{
			Assets:         map[string]listing.Asset{"USDT": {Symbol: "USDT", State: listing.StateListed}},
			Networks:       map[string]listing.Network{"USDT:TRON": {Asset: "USDT", Network: "TRON", State: listing.StateListed}},
			Markets:        map[string]listing.Market{"BTC-USDT": {Symbol: "BTC-USDT", State: listing.StateListed}},
			Counterparties: map[string]listing.CounterpartyVASP{"vasp-1": {ID: "vasp-1", State: listing.StateApproved}},
		},
		HighRiskAmount:             10_000_000,
		WithdrawalsGloballyEnabled: true,
	}
}
