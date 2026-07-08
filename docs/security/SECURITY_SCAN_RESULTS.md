# Security Scan Results

## Current Project

Date: 2026-07-09

| Check | Command | Result |
|---|---|---|
| Unit and package tests | `go test ./...` | Passed |
| Go static checks | `go vet ./...` | Passed |
| Secret scan | `gitleaks detect --source . --no-banner --redact` | No leaks found |
| Go vulnerability reachability | `govulncheck ./...` | 0 called vulnerabilities |
| Floating-point finance scan | `rg "float32|float64|double|float"` | No Go/proto financial float types; only documentation references |

`govulncheck` reported vulnerable symbols in dependency modules that are not called by this code. This is acceptable for the current skeleton but should be rechecked before release.

## Reference Sources

Reference repositories are analysis inputs, not code inputs. After installing `gitleaks`, `scripts/audit-sources.sh` was rerun against `~/exchange-lab/sources`.

Observed redacted leak candidates during the rerun:

| Repository | Result |
|---|---|
| billy-coinexchange | Leak candidates found |
| coincoin-crypto-exchange | Leak candidates found |
| coinglobalvip | Leak candidates found |
| exchange-core | No leaks found |
| frizo-exchange | No leaks found |
| futures-engine | No leaks found |
| omx-engine | No leaks found |
| opencex | No leaks found |
| opex-core | Leak candidates found |

Policy impact:

- CoinExchange/Gitee-family repositories remain structure-reference only.
- Leak candidates in any reference repository are not copied, normalized, or used.
- OPEX remains architecture-reference only until a focused review confirms the findings are harmless test fixtures or removes them from consideration.

