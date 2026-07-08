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

The current implementation is in-memory. Before production use, command WAL and periodic orderbook snapshots are required so `snapshot + WAL replay` produces the same book and trade sequence.

