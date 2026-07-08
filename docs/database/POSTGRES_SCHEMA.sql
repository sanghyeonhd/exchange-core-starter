CREATE TABLE users (
  id BIGINT PRIMARY KEY,
  email TEXT UNIQUE,
  phone TEXT UNIQUE,
  status TEXT NOT NULL,
  kyc_level INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE markets (
  symbol TEXT PRIMARY KEY,
  base_asset TEXT NOT NULL,
  quote_asset TEXT NOT NULL,
  market_type TEXT NOT NULL,
  status TEXT NOT NULL,
  price_scale INT NOT NULL,
  quantity_scale INT NOT NULL,
  min_order_qty BIGINT NOT NULL,
  min_notional BIGINT NOT NULL,
  tick_size BIGINT NOT NULL,
  lot_size BIGINT NOT NULL,
  maker_fee_rate BIGINT NOT NULL,
  taker_fee_rate BIGINT NOT NULL
);

CREATE TABLE accounts (
  id BIGINT PRIMARY KEY,
  user_id BIGINT,
  account_type TEXT NOT NULL,
  asset TEXT NOT NULL,
  available BIGINT NOT NULL DEFAULT 0 CHECK (available >= 0),
  locked BIGINT NOT NULL DEFAULT 0 CHECK (locked >= 0),
  version BIGINT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  UNIQUE(user_id, account_type, asset)
);

CREATE TABLE ledger_transactions (
  id BIGINT PRIMARY KEY,
  type TEXT NOT NULL,
  reference_type TEXT,
  reference_id TEXT,
  idempotency_key TEXT UNIQUE NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE ledger_entries (
  id BIGINT PRIMARY KEY,
  transaction_id BIGINT NOT NULL REFERENCES ledger_transactions(id),
  account_id BIGINT NOT NULL REFERENCES accounts(id),
  asset TEXT NOT NULL,
  debit BIGINT NOT NULL DEFAULT 0 CHECK (debit >= 0),
  credit BIGINT NOT NULL DEFAULT 0 CHECK (credit >= 0),
  balance_after BIGINT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  CHECK ((debit > 0 AND credit = 0) OR (debit = 0 AND credit > 0))
);

CREATE TABLE orders (
  id BIGINT PRIMARY KEY,
  client_order_id TEXT,
  user_id BIGINT NOT NULL,
  market TEXT NOT NULL REFERENCES markets(symbol),
  side TEXT NOT NULL,
  type TEXT NOT NULL,
  time_in_force TEXT NOT NULL,
  price BIGINT,
  quantity BIGINT NOT NULL,
  executed_quantity BIGINT NOT NULL DEFAULT 0,
  status TEXT NOT NULL,
  reduce_only BOOLEAN NOT NULL DEFAULT false,
  post_only BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  UNIQUE(user_id, market, client_order_id),
  CHECK (quantity > 0),
  CHECK (executed_quantity >= 0 AND executed_quantity <= quantity)
);

CREATE TABLE trades (
  id BIGINT PRIMARY KEY,
  market TEXT NOT NULL REFERENCES markets(symbol),
  maker_order_id BIGINT NOT NULL,
  taker_order_id BIGINT NOT NULL,
  maker_user_id BIGINT NOT NULL,
  taker_user_id BIGINT NOT NULL,
  price BIGINT NOT NULL,
  quantity BIGINT NOT NULL,
  maker_fee BIGINT NOT NULL,
  taker_fee BIGINT NOT NULL,
  match_sequence BIGINT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  UNIQUE(market, match_sequence),
  CHECK (price > 0),
  CHECK (quantity > 0)
);

CREATE TABLE positions (
  id BIGINT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  market TEXT NOT NULL REFERENCES markets(symbol),
  position_mode TEXT NOT NULL,
  margin_mode TEXT NOT NULL,
  side TEXT NOT NULL,
  size BIGINT NOT NULL,
  entry_price BIGINT NOT NULL,
  isolated_margin BIGINT NOT NULL,
  leverage INT NOT NULL,
  realized_pnl BIGINT NOT NULL DEFAULT 0,
  status TEXT NOT NULL,
  version BIGINT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  UNIQUE(user_id, market, side),
  CHECK (leverage >= 1 AND leverage <= 10)
);

CREATE TABLE funding_rates (
  id BIGINT PRIMARY KEY,
  market TEXT NOT NULL REFERENCES markets(symbol),
  funding_time TIMESTAMPTZ NOT NULL,
  index_price BIGINT NOT NULL,
  mark_price BIGINT NOT NULL,
  premium_index BIGINT NOT NULL,
  interest_rate BIGINT NOT NULL,
  funding_rate BIGINT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  UNIQUE(market, funding_time)
);

CREATE TABLE funding_payments (
  id BIGINT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  position_id BIGINT NOT NULL REFERENCES positions(id),
  market TEXT NOT NULL REFERENCES markets(symbol),
  funding_time TIMESTAMPTZ NOT NULL,
  amount BIGINT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  UNIQUE(position_id, funding_time)
);

CREATE TABLE deposit_addresses (
  id BIGINT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  asset TEXT NOT NULL,
  network TEXT NOT NULL,
  address TEXT NOT NULL,
  memo TEXT,
  derivation_path TEXT,
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  UNIQUE NULLS NOT DISTINCT (network, address, memo)
);

CREATE TABLE deposits (
  id BIGINT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  asset TEXT NOT NULL,
  network TEXT NOT NULL,
  txid TEXT NOT NULL,
  address TEXT NOT NULL,
  memo TEXT,
  amount BIGINT NOT NULL,
  confirmations INT NOT NULL DEFAULT 0,
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  UNIQUE NULLS NOT DISTINCT (network, txid, address, memo),
  CHECK (amount > 0)
);

CREATE TABLE withdrawals (
  id BIGINT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  asset TEXT NOT NULL,
  network TEXT NOT NULL,
  address TEXT NOT NULL,
  memo TEXT,
  amount BIGINT NOT NULL,
  fee BIGINT NOT NULL,
  txid TEXT,
  status TEXT NOT NULL,
  risk_score INT NOT NULL DEFAULT 0,
  requested_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  CHECK (amount > 0),
  CHECK (fee >= 0)
);
