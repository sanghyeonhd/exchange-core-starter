package listing

import "testing"

func TestListingPolicyBlocksUnlistedMarket(t *testing.T) {
	policy := Policy{Markets: map[string]Market{"BTC-USDT": {Symbol: "BTC-USDT", State: StateApproved}}}
	if policy.CanTrade("BTC-USDT") {
		t.Fatal("approved but not listed market should not trade")
	}
}

func TestListingPolicyAllowsListedAssetNetworkCounterparty(t *testing.T) {
	policy := Policy{
		Assets:         map[string]Asset{"USDT": {Symbol: "USDT", State: StateListed}},
		Networks:       map[string]Network{"USDT:TRON": {Asset: "USDT", Network: "TRON", State: StateListed}},
		Counterparties: map[string]CounterpartyVASP{"vasp-1": {ID: "vasp-1", State: StateApproved}},
	}
	if !policy.CanWithdraw("USDT", "TRON", "vasp-1") {
		t.Fatal("listed asset/network and approved counterparty should allow withdrawal")
	}
}
