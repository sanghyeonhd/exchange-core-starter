package readiness

import "testing"

func TestSpotLaunchRequiresRegulatoryAndOperationalPrerequisites(t *testing.T) {
	decision := Evaluate(ScopeSpotLaunch, Status{
		LegalApproved:           true,
		ComplianceApproved:      true,
		SecurityApproved:        true,
		ISMSReady:               true,
		VASPFilingReady:         true,
		AMLOperational:          true,
		KYCOperational:          true,
		ClosedPilotPassed:       true,
		IncidentDrillPassed:     true,
		RecoveryDrillPassed:     true,
		MarketSurveillanceReady: true,
	})
	if !decision.Allowed {
		t.Fatalf("decision = %+v, want allowed", decision)
	}
}

func TestFiatEnablementRequiresBankReadiness(t *testing.T) {
	decision := Evaluate(ScopeFiatEnablement, Status{
		LegalApproved:            true,
		ComplianceApproved:       true,
		VASPFilingReady:          true,
		BankRealNameAccountReady: false,
	})
	if decision.Allowed {
		t.Fatal("fiat enablement should be blocked without bank real-name account readiness")
	}
	assertMissing(t, decision, "BANK_REAL_NAME_ACCOUNT_READY")
}

func TestDerivativesRemainBlockedWithoutSeparateApprovals(t *testing.T) {
	decision := Evaluate(ScopeDerivatives, Status{
		LegalApproved:           true,
		MarketSurveillanceReady: true,
	})
	if decision.Allowed {
		t.Fatal("derivatives should be blocked without separate legal/risk/suitability approvals")
	}
	assertMissing(t, decision, "DERIVATIVES_LEGAL_APPROVAL")
	assertMissing(t, decision, "DERIVATIVES_RISK_APPROVAL")
	assertMissing(t, decision, "CUSTOMER_SUITABILITY_READY")
}

func assertMissing(t *testing.T, decision Decision, code string) {
	t.Helper()
	for _, missing := range decision.Missing {
		if missing == code {
			return
		}
	}
	t.Fatalf("missing code %s not found in %+v", code, decision)
}
