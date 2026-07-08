# Wallet Gateway Design

## Responsibilities

- Hide chain/provider differences behind `WalletAdapter`.
- Generate or retrieve deposit addresses.
- Scan deposits and confirmations.
- Manage withdrawal lifecycle.
- Keep signer responsibilities separate from public APIs.
- Emit immutable audit events for every state transition.

## Adapter Interface

```go
type WalletAdapter interface {
    Asset() string
    Network() string
    CreateDepositAddress(ctx context.Context, userID string) (*DepositAddress, error)
    GetBalance(ctx context.Context) (*WalletBalance, error)
    BuildWithdrawal(ctx context.Context, req WithdrawalRequest) (*UnsignedTx, error)
    SignTransaction(ctx context.Context, tx UnsignedTx) (*SignedTx, error)
    BroadcastTransaction(ctx context.Context, tx SignedTx) (*BroadcastResult, error)
    GetTransaction(ctx context.Context, txid string) (*ChainTransaction, error)
    ScanDeposits(ctx context.Context, fromHeight int64, toHeight int64) ([]DepositEvent, error)
}
```

## Deposit Lifecycle

```text
REQUESTED_ADDRESS
  -> ADDRESS_CREATED
  -> DETECTED
  -> CONFIRMING
  -> CONFIRMED
  -> CREDITED
```

Credit is a ledger transaction and is idempotent by `network + txid + address + memo`.

## Withdrawal Lifecycle

```text
REQUESTED
  -> LOCKED
  -> PENDING_REVIEW
  -> APPROVED
  -> SIGNED
  -> BROADCASTED
  -> CONFIRMED
```

Failure terminal states:

```text
REJECTED
FAILED
CANCELED
```

## Mainnet Gate

Mainnet support requires all of the following:

- Production signer isolation
- Secret manager integration
- Chain reorg handling
- Withdrawal approval policy
- Address allowlist policy
- Hot/cold wallet runbooks
- Security test pass

