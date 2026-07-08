# Recovery Runbook

## Current Status

Matching recovery primitives are implemented:

- Command WAL (`services/matching-engine/wal`): every accepted place/cancel is fsynced as a JSON-lines envelope before it reaches the book. The gateway enables it with `MATCHING_WAL_PATH=<file>`; reopening the file resumes the sequence.
- Orderbook snapshots (`engine.Snapshot` / `engine.RestoreOrderBook`) with a stable SHA-256 hash for comparison, and `spotexchange.Checkpoint()` pairing a snapshot with the WAL sequence atomically.
- Replay harness (`wal.Replay`) that rejects any divergence between the log and the starting book state.
- Deterministic replay tests in `tests/replay`: full replay, snapshot + tail replay, and repeated recovery all rebuild identical book hashes and trade sequences (`go test ./tests/replay/`).

Still pending for production recovery: durable account/ledger storage, event archive, snapshot scheduling/retention, and recovery of in-flight reservations (accounts are currently rebuilt from seeds, not replayed).

## Matching Recovery Procedure

1. Stop accepting new commands for the affected symbol.
2. Load the latest orderbook snapshot (`engine.RestoreOrderBook`).
3. Load the WAL (`wal.LoadFileLog`) and replay commands after the snapshot's WAL sequence (`wal.Replay`).
4. Compare the final book hash (`OrderBookSnapshot.Hash`) and last sequence with the expected event log.
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

