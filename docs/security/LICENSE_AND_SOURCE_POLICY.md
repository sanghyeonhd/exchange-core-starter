# License and Source Policy

## Clean-Room Rule

Reference repositories are used for architecture comparison only until their license, authorship, and security posture are reviewed.

## Allowed License Candidates

- MIT
- Apache-2.0
- BSD family

## High-Risk Sources

- GPL/AGPL without an explicit compatibility decision
- Non-commercial licenses
- No-license repositories
- Repositories with unclear authorship or copied code lineage
- Gitee CoinExchange family code before a dedicated security review

## Current Policy by Source

| Source | Policy |
|---|---|
| OPEX | Architecture ideas allowed; code reuse only with attribution review |
| Gitee CoinExchange variants | Structure reference only |
| Frizo | Concept reference only because root license was not found during initial audit |
| Futures Engine | MIT concepts allowed; code still not copied |
| OMX | Concept reference only because root license was not found during initial audit |
| Exchange-Core | Performance/data-structure reference only |

## Commit Policy

- No copied source files from reference repositories.
- No private keys, mnemonics, real API keys, or production credentials.
- Generated analysis documents must name uncertainty clearly.

