# Repository Comparison Matrix

| Repository | Purpose | License Status | Secret Scan | Code Reuse Policy | Risk |
|---|---|---|---|---|---|
| opex-core | CEX core architecture, gateway, wallet, accounting split | MIT | Leak candidates found | Concepts only until focused review | High |
| coinglobalvip | Broad feature inventory including wallet/futures/OTC | Apache-2.0 | Leak candidates found | Structure reference only | High |
| coincoin-crypto-exchange | Java/SpringCloud exchange layout | Apache-2.0 | Leak candidates found | Structure reference only | High |
| frizo-exchange | Go spot/futures service boundaries | No root license found | No leaks found | Concepts only, no code reuse | High |
| futures-engine | Go futures engine approach | MIT | No leaks found | Algorithmic ideas only | Medium |
| omx-engine | Derivatives matching/settlement concepts | No root license found | No leaks found | Concepts only, no code reuse | High |
| opencex | Feature checklist | Apache-2.0 | No leaks found | No production base | High |
| billy-coinexchange | CoinExchange mirror | Apache-2.0 | Leak candidates found | Structure reference only | High |
| exchange-core | Java high-performance matching structures | Apache-2.0 | No leaks found | Benchmark/data-structure reference | Medium |

Detailed per-repository reports are generated after `scripts/clone-sources.sh` and `scripts/audit-sources.sh` run.
