# Recovery Runbook

## Current Status

The first MVP includes in-memory matching and ledger invariant checks. Production recovery requires WAL, snapshots, event archive, and idempotent settlement consumers.

## Matching Recovery Target

1. Stop accepting new commands for the affected symbol.
2. Load the latest orderbook snapshot.
3. Replay WAL commands after the snapshot sequence.
4. Compare final book hash and last sequence with the expected event log.
5. Resume command intake only after consistency checks pass.

## Settlement Recovery Target

1. Read the last successfully settled trade id.
2. Reconsume trade events from the durable event stream.
3. Use `trade_id` as the settlement idempotency key.
4. Let ledger unique constraints reject duplicate settlement attempts.
5. Verify debit and credit totals after replay.

## Wallet Recovery Target

1. Resume scanners from the last finalized block height.
2. Rescan a confirmation overlap window to handle reorgs.
3. Deduplicate deposits by network/txid/address/memo.
4. Never rebroadcast a withdrawal without checking its current chain and database state.

## Release Gate

No production financial workflow is complete until replay tests prove that repeated recovery produces identical state and no duplicate ledger movements.

