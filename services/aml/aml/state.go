package aml

type RiskLevel string

const (
	RiskLow      RiskLevel = "LOW"
	RiskMedium   RiskLevel = "MEDIUM"
	RiskHigh     RiskLevel = "HIGH"
	RiskCritical RiskLevel = "CRITICAL"
)

type CaseStatus string

const (
	CaseOpen                CaseStatus = "OPEN"
	CaseNeedsReview         CaseStatus = "NEEDS_REVIEW"
	CaseEscalated           CaseStatus = "ESCALATED"
	CaseClosedFalsePositive CaseStatus = "CLOSED_FALSE_POSITIVE"
	CaseClosedReported      CaseStatus = "CLOSED_REPORTED"
	CaseClosedNoAction      CaseStatus = "CLOSED_NO_ACTION"
)

type Profile struct {
	UserID            int64
	RiskLevel         RiskLevel
	SanctionsHit      bool
	PEP               bool
	OpenCase          bool
	WithdrawalsLocked bool
}

func (p Profile) CanTrade() bool {
	return !p.SanctionsHit && p.RiskLevel != RiskCritical
}

func (p Profile) CanWithdraw() bool {
	if p.SanctionsHit || p.WithdrawalsLocked || p.OpenCase {
		return false
	}
	return p.RiskLevel == RiskLow || p.RiskLevel == RiskMedium
}

func NeedsCase(amount int64, highRiskThreshold int64, profile Profile) bool {
	return profile.SanctionsHit || profile.PEP || profile.RiskLevel == RiskHigh || profile.RiskLevel == RiskCritical || amount >= highRiskThreshold
}
