# Asset Inventory

The canonical baseline is implemented in `services/compliance/inventory`.

Critical assets currently tracked:

| Asset | Type | Owner | Classification |
|---|---|---|---|
| `apps/web` | Frontend | product | confidential |
| `apps/admin` | Frontend | compliance | restricted |
| `services/gateway` | Service | backend | confidential |
| `services/matching-engine` | Service | trading-core | restricted |
| `services/ledger` | Service | finance-core | restricted |
| `services/wallet-gateway` | Wallet | custody | restricted |
| `services/kyc` | Service | compliance | restricted |
| `services/aml` | Service | compliance | restricted |
| `postgres` | Database | platform | restricted |

ISMS/ISO implication:

- Every critical asset must have an owner.
- Restricted assets need access reviews.
- Assets with identity, wallet, ledger, or compliance data need retention and audit controls.

