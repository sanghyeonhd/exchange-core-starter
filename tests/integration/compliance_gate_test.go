package integration

import (
	"testing"
	"time"

	"github.com/exchange-core-starter/exchange-core-starter/libs/audit"
	"github.com/exchange-core-starter/exchange-core-starter/services/aml/aml"
	"github.com/exchange-core-starter/exchange-core-starter/services/auth/auth"
	"github.com/exchange-core-starter/exchange-core-starter/services/compliance/compliance"
	"github.com/exchange-core-starter/exchange-core-starter/services/kyc/kyc"
	"github.com/exchange-core-starter/exchange-core-starter/services/listing/listing"
)

func TestComplianceGateRequiresKYCAMLListingAndAuditedAdminReview(t *testing.T) {
	auditLog := audit.NewLog()
	admin := auth.Subject{ID: "kyc-admin-1", Roles: []auth.Role{auth.RoleKYCAnalyst}, MFAEnabled: true}
	if err := auth.Authorize(admin, auth.PermissionReviewKYC); err != nil {
		t.Fatalf("Authorize KYC admin: %v", err)
	}
	if _, err := auditLog.Append(audit.Event{
		ID:           "audit-1",
		ActorID:      admin.ID,
		ActorType:    "ADMIN",
		Action:       "KYC_APPROVE",
		ResourceType: "KYC_PROFILE",
		ResourceID:   "user-1",
		RequestID:    "req-kyc-1",
		Reason:       "identity verified by test fixture",
		CreatedAt:    time.Unix(1, 0).UTC(),
	}); err != nil {
		t.Fatalf("Append audit: %v", err)
	}

	inputs := compliance.Inputs{
		KYC: kyc.Profile{UserID: 1, Level: 2, Status: kyc.StatusApproved, Country: "KR"},
		AML: aml.Profile{UserID: 1, RiskLevel: aml.RiskLow},
		Listing: listing.Policy{
			Assets: map[string]listing.Asset{
				"USDT": {Symbol: "USDT", State: listing.StateListed},
			},
			Networks: map[string]listing.Network{
				"USDT:TRON": {Asset: "USDT", Network: "TRON", State: listing.StateListed},
			},
			Markets: map[string]listing.Market{
				"BTC-USDT": {Symbol: "BTC-USDT", State: listing.StateListed},
			},
			Counterparties: map[string]listing.CounterpartyVASP{
				"vasp-1": {ID: "vasp-1", State: listing.StateApproved},
			},
		},
		HighRiskAmount:             10_000_000,
		WithdrawalsGloballyEnabled: true,
	}

	tradeDecision := compliance.Decide(compliance.Request{Action: compliance.ActionTrade, UserID: 1, Market: "BTC-USDT"}, inputs)
	if !tradeDecision.Allowed {
		t.Fatalf("trade decision = %+v, want allowed", tradeDecision)
	}

	withdrawDecision := compliance.Decide(compliance.Request{
		Action:         compliance.ActionWithdraw,
		UserID:         1,
		Asset:          "USDT",
		Network:        "TRON",
		CounterpartyID: "vasp-1",
		Amount:         1_000_000,
	}, inputs)
	if !withdrawDecision.Allowed {
		t.Fatalf("withdraw decision = %+v, want allowed", withdrawDecision)
	}
	if err := auditLog.Verify(); err != nil {
		t.Fatalf("audit Verify: %v", err)
	}
}

func TestComplianceGateBlocksUnapprovedUserAndUnlistedMarket(t *testing.T) {
	inputs := compliance.Inputs{
		KYC: kyc.Profile{UserID: 1, Level: 0, Status: kyc.StatusPending},
		AML: aml.Profile{UserID: 1, RiskLevel: aml.RiskLow},
		Listing: listing.Policy{
			Markets: map[string]listing.Market{
				"BTC-USDT": {Symbol: "BTC-USDT", State: listing.StateApproved},
			},
		},
		WithdrawalsGloballyEnabled: true,
	}
	decision := compliance.Decide(compliance.Request{Action: compliance.ActionTrade, UserID: 1, Market: "BTC-USDT"}, inputs)
	if decision.Allowed {
		t.Fatal("trade should be blocked")
	}
	want := map[string]bool{"KYC_TRADE_BLOCKED": true, "MARKET_NOT_LISTED": true}
	for _, reason := range decision.Reasons {
		delete(want, reason)
	}
	if len(want) != 0 {
		t.Fatalf("missing reasons: %+v in %+v", want, decision)
	}
}
