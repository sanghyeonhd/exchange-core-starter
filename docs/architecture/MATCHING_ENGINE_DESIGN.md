# Matching Engine Design

## Scope

The initial engine supports spot `LIMIT` and `MARKET` orders with price-time priority.

## Rules

- One orderbook is responsible for one symbol.
- Bids sort by highest price first.
- Asks sort by lowest price first.
- Orders at the same price execute FIFO.
- Trade price is the maker order price.
- Market orders never rest in the book.
- The engine emits trades only; settlement and ledger updates happen outside the matching path.

## Recovery Direction

Implemented: an append-only command WAL (`services/matching-engine/wal`, JSON-lines, fsync before acknowledge), orderbook snapshot/restore with a SHA-256 book hash, and a replay harness with divergence detection. `snapshot + WAL replay` is proven deterministic by `tests/replay`: the same command sequence rebuilds the same book hash and trade sequence, including snapshot-plus-tail and repeated recovery.

Remaining before production: snapshot scheduling and retention, WAL segment rotation/archival, and recovery of settlement/account state (currently only the matching side replays).

