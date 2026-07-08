# Security Checklist

## General API

- Validate every external input.
- Enforce request body size limits.
- Enforce HMAC timestamp drift.
- Enforce nonce or request id replay protection.
- Scope API key permissions.
- Apply IP whitelist where configured.
- Rate limit public and private endpoints separately.

## Trading

- Enforce market status.
- Enforce tick size and lot size.
- Enforce max order size.
- Enforce max open orders per user.
- Enforce price bands.
- Prevent negative available and locked balances.
- Use idempotent order submission.

## Ledger

- All financial movements go through ledger transactions.
- Transactions balance per asset.
- Idempotency keys are unique.
- Manual balance updates are prohibited.

## Wallet

- Mainnet disabled by default.
- Withdrawals disabled by default.
- Signer is isolated.
- Withdrawal approval and signing are separate.
- New withdrawal addresses can be time locked.
- High-value withdrawals require review.
- Full withdrawal addresses are masked in logs.

## Required Tools Before Production

```bash
gitleaks detect
trivy fs .
govulncheck ./...
```

