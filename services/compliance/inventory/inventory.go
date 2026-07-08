package inventory

type AssetType string

const (
	AssetService  AssetType = "SERVICE"
	AssetDatabase AssetType = "DATABASE"
	AssetFrontend AssetType = "FRONTEND"
	AssetWallet   AssetType = "WALLET"
	AssetVendor   AssetType = "VENDOR"
)

type Classification string

const (
	ClassificationPublic       Classification = "PUBLIC"
	ClassificationInternal     Classification = "INTERNAL"
	ClassificationConfidential Classification = "CONFIDENTIAL"
	ClassificationRestricted   Classification = "RESTRICTED"
)

type Asset struct {
	ID             string
	Name           string
	Type           AssetType
	Owner          string
	Classification Classification
	DataCategories []string
	Critical       bool
}

type Register struct {
	assets []Asset
}

func NewRegister(assets []Asset) Register {
	return Register{assets: assets}
}

func (r Register) MissingOwners() []Asset {
	var missing []Asset
	for _, asset := range r.assets {
		if asset.Owner == "" {
			missing = append(missing, asset)
		}
	}
	return missing
}

func (r Register) MissingCritical(required []string) []string {
	seen := make(map[string]bool)
	for _, asset := range r.assets {
		if asset.Critical {
			seen[asset.ID] = true
		}
	}
	var missing []string
	for _, id := range required {
		if !seen[id] {
			missing = append(missing, id)
		}
	}
	return missing
}

func BaselineAssets() []Asset {
	return []Asset{
		{ID: "apps/web", Name: "User Web App", Type: AssetFrontend, Owner: "product", Classification: ClassificationConfidential, DataCategories: []string{"account", "orders", "wallet"}, Critical: true},
		{ID: "apps/admin", Name: "Admin Web App", Type: AssetFrontend, Owner: "compliance", Classification: ClassificationRestricted, DataCategories: []string{"kyc", "aml", "withdrawals", "audit"}, Critical: true},
		{ID: "services/gateway", Name: "API Gateway", Type: AssetService, Owner: "backend", Classification: ClassificationConfidential, DataCategories: []string{"api_requests", "sessions"}, Critical: true},
		{ID: "services/matching-engine", Name: "Matching Engine", Type: AssetService, Owner: "trading-core", Classification: ClassificationRestricted, DataCategories: []string{"orders", "trades"}, Critical: true},
		{ID: "services/ledger", Name: "Ledger Service", Type: AssetService, Owner: "finance-core", Classification: ClassificationRestricted, DataCategories: []string{"balances", "ledger_entries"}, Critical: true},
		{ID: "services/wallet-gateway", Name: "Wallet Gateway", Type: AssetWallet, Owner: "custody", Classification: ClassificationRestricted, DataCategories: []string{"deposit_addresses", "withdrawals"}, Critical: true},
		{ID: "services/kyc", Name: "KYC Service", Type: AssetService, Owner: "compliance", Classification: ClassificationRestricted, DataCategories: []string{"identity", "documents"}, Critical: true},
		{ID: "services/aml", Name: "AML Service", Type: AssetService, Owner: "compliance", Classification: ClassificationRestricted, DataCategories: []string{"risk_scores", "cases"}, Critical: true},
		{ID: "postgres", Name: "PostgreSQL", Type: AssetDatabase, Owner: "platform", Classification: ClassificationRestricted, DataCategories: []string{"accounts", "orders", "ledger", "compliance"}, Critical: true},
	}
}
