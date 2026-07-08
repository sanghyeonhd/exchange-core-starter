package inventory

import "testing"

func TestBaselineAssetsHaveOwners(t *testing.T) {
	register := NewRegister(BaselineAssets())
	if missing := register.MissingOwners(); len(missing) != 0 {
		t.Fatalf("assets missing owners: %+v", missing)
	}
}

func TestBaselineAssetsCoverCriticalExchangeSystems(t *testing.T) {
	register := NewRegister(BaselineAssets())
	missing := register.MissingCritical([]string{
		"apps/web",
		"apps/admin",
		"services/gateway",
		"services/matching-engine",
		"services/ledger",
		"services/wallet-gateway",
		"services/kyc",
		"services/aml",
		"postgres",
	})
	if len(missing) != 0 {
		t.Fatalf("missing critical assets: %+v", missing)
	}
}
