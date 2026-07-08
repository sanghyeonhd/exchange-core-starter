# ADR 0004: Wallet Adapter Model

## Status

Accepted

## Decision

Wallet handling is abstracted behind adapters. The default adapter is mock and mainnet is disabled in local development.

## Rationale

Full node operation is not required for early development, and private-key handling must stay isolated from API and settlement services.

## Consequences

- `MockWalletAdapter` is used for the first wallet MVP.
- Full-node, explorer, and custody adapters must share the same interface.
- Withdrawal approval and signing are separate responsibilities.

