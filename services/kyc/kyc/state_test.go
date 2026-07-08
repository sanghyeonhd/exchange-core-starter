package kyc

import "testing"

func TestKYCTradingAndWithdrawalGates(t *testing.T) {
	profile := Profile{UserID: 1, Level: 1, Status: StatusApproved}
	if !profile.CanTrade() {
		t.Fatal("approved level 1 should trade")
	}
	if profile.CanWithdraw() {
		t.Fatal("level 1 should not withdraw")
	}
	profile.Level = 2
	if !profile.CanWithdraw() {
		t.Fatal("approved level 2 should withdraw")
	}
}
