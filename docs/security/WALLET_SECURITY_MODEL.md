# Wallet Security Model

## Default Posture

- `MAINNET_ENABLED=false`
- `WITHDRAWALS_ENABLED=false`
- `WALLET_ADAPTER=mock`
- No private key, mnemonic, exchange API key, or custody credential may be committed.

## Adapter Types

| Adapter | Use |
|---|---|
| MockWalletAdapter | Local development and deterministic tests |
| TestnetAdapter | Pre-production chain integration |
| FullNodeRpcAdapter | Direct node integration |
| ExplorerApiAdapter | Read-only deposit detection fallback |
| ThirdPartyCustodyAdapter | Custody provider integration |
| ManualDepositAdapter | Audited manual credit process |

## Deposit Controls

- Deposit addresses are unique by network/address/memo.
- Deposits are idempotent by network/txid/address/memo.
- Credits occur only after confirmation policy and risk checks pass.
- Reorg handling must be implemented before any mainnet support.

## Withdrawal Controls

1. Auth verifies permission scope, IP whitelist, timestamp, and nonce.
2. Account locks available balance.
3. Risk checks address, amount, velocity, and user status.
4. Admin approval is required while the system is in MVP mode.
5. Signer runs separately from public APIs.
6. Broadcast is idempotent by withdrawal id.
7. Every state transition is written to immutable audit logs.

## Separation of Duties

- Wallet API cannot access private keys.
- Signer cannot approve withdrawals.
- Admin approval cannot sign or broadcast.
- Cold wallet movement remains manual until a separate runbook and approval model exist.

