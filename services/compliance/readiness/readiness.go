package readiness

type Scope string

const (
	ScopeSpotLaunch     Scope = "SPOT_LAUNCH"
	ScopeFiatEnablement Scope = "FIAT_ENABLEMENT"
	ScopeMainnetCustody Scope = "MAINNET_CUSTODY"
	ScopeDerivatives    Scope = "DERIVATIVES"
)

type Status struct {
	LegalApproved            bool
	ComplianceApproved       bool
	SecurityApproved         bool
	ISMSReady                bool
	ISO27001Ready            bool
	VASPFilingReady          bool
	BankRealNameAccountReady bool
	AMLOperational           bool
	KYCOperational           bool
	TravelRuleOperational    bool
	CustodyRunbookApproved   bool
	ClosedPilotPassed        bool
	IncidentDrillPassed      bool
	RecoveryDrillPassed      bool
	MarketSurveillanceReady  bool
	DerivativesLegalApproved bool
	DerivativesRiskApproved  bool
	CustomerSuitabilityReady bool
}

type Decision struct {
	Allowed bool
	Missing []string
}

func Evaluate(scope Scope, status Status) Decision {
	var missing []string
	require := func(ok bool, name string) {
		if !ok {
			missing = append(missing, name)
		}
	}

	switch scope {
	case ScopeSpotLaunch:
		require(status.LegalApproved, "LEGAL_APPROVAL")
		require(status.ComplianceApproved, "COMPLIANCE_APPROVAL")
		require(status.SecurityApproved, "SECURITY_APPROVAL")
		require(status.ISMSReady, "ISMS_READY")
		require(status.VASPFilingReady, "VASP_FILING_READY")
		require(status.AMLOperational, "AML_OPERATIONAL")
		require(status.KYCOperational, "KYC_OPERATIONAL")
		require(status.ClosedPilotPassed, "CLOSED_PILOT_PASSED")
		require(status.IncidentDrillPassed, "INCIDENT_DRILL_PASSED")
		require(status.RecoveryDrillPassed, "RECOVERY_DRILL_PASSED")
		require(status.MarketSurveillanceReady, "MARKET_SURVEILLANCE_READY")
	case ScopeFiatEnablement:
		require(status.LegalApproved, "LEGAL_APPROVAL")
		require(status.ComplianceApproved, "COMPLIANCE_APPROVAL")
		require(status.BankRealNameAccountReady, "BANK_REAL_NAME_ACCOUNT_READY")
		require(status.VASPFilingReady, "VASP_FILING_READY")
	case ScopeMainnetCustody:
		require(status.SecurityApproved, "SECURITY_APPROVAL")
		require(status.CustodyRunbookApproved, "CUSTODY_RUNBOOK_APPROVED")
		require(status.TravelRuleOperational, "TRAVEL_RULE_OPERATIONAL")
		require(status.IncidentDrillPassed, "INCIDENT_DRILL_PASSED")
		require(status.RecoveryDrillPassed, "RECOVERY_DRILL_PASSED")
	case ScopeDerivatives:
		require(status.LegalApproved, "LEGAL_APPROVAL")
		require(status.DerivativesLegalApproved, "DERIVATIVES_LEGAL_APPROVAL")
		require(status.DerivativesRiskApproved, "DERIVATIVES_RISK_APPROVAL")
		require(status.CustomerSuitabilityReady, "CUSTOMER_SUITABILITY_READY")
		require(status.MarketSurveillanceReady, "MARKET_SURVEILLANCE_READY")
	default:
		missing = append(missing, "UNKNOWN_SCOPE")
	}

	return Decision{Allowed: len(missing) == 0, Missing: missing}
}
