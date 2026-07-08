package aml

import "testing"

func TestAMLBlocksSanctionsHit(t *testing.T) {
	profile := Profile{UserID: 1, RiskLevel: RiskLow, SanctionsHit: true}
	if profile.CanTrade() || profile.CanWithdraw() {
		t.Fatal("sanctions hit must block trade and withdrawal")
	}
}

func TestNeedsCaseForLargeWithdrawal(t *testing.T) {
	if !NeedsCase(10_000_000, 5_000_000, Profile{RiskLevel: RiskLow}) {
		t.Fatal("large withdrawal should create AML case")
	}
}
