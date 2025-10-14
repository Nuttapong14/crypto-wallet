
-- =====================================================================
-- Crypto Wallet & Fintech Platform (Simplified: 4 Databases)
-- PostgreSQL DDL — Production-Ready (Enums, Constraints, Indexes, Triggers)
-- =====================================================================

-- NOTE:
-- - Requires PostgreSQL 13+
-- - Uses pgcrypto for gen_random_uuid()
-- - Execute per-database section after creating the database or run with psql \connect.
-- - Timezones: use TIMESTAMPTZ to avoid ambiguity.
-- - Numeric precision chosen for crypto assets (36,18). Adjust if needed per asset.
-- =====================================================================

-- =====================================================================
-- Create Databases (optional; run with superuser if desired)
-- =====================================================================
-- CREATE DATABASE core_db;
-- CREATE DATABASE kyc_db;
-- CREATE DATABASE rates_db;
-- CREATE DATABASE audit_db;

-- =====================================================================
-- Common helper: function to auto-update updated_at
-- (Create in each database where used)
-- =====================================================================
-- Example:
-- CREATE OR REPLACE FUNCTION set_updated_at()
-- RETURNS TRIGGER AS $$
-- BEGIN
--   NEW.updated_at := NOW();
--   RETURN NEW;
-- END;
-- $$ LANGUAGE plpgsql;

-- =====================================================================
-- ============================== 1) core_db ============================
-- =====================================================================
-- \connect core_db

CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext;

-- ---------- Enums ----------
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'chain_enum') THEN
    CREATE TYPE chain_enum AS ENUM ('BTC','ETH','BSC','POLYGON');
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'tx_status_enum') THEN
    CREATE TYPE tx_status_enum AS ENUM ('pending','confirmed','failed');
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'account_type_enum') THEN
    CREATE TYPE account_type_enum AS ENUM ('spot','savings','staking');
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'entry_type_enum') THEN
    CREATE TYPE entry_type_enum AS ENUM ('debit','credit');
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'node_status_enum') THEN
    CREATE TYPE node_status_enum AS ENUM ('active','inactive','degraded');
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'job_status_enum') THEN
    CREATE TYPE job_status_enum AS ENUM ('idle','running','success','error');
  END IF;
END $$;

-- ---------- Helper trigger for updated_at ----------
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at := NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ---------- Tables ----------

-- Users
CREATE TABLE IF NOT EXISTS users (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email             CITEXT UNIQUE NOT NULL,
    password_hash     TEXT NOT NULL,
    first_name        VARCHAR(100),
    last_name         VARCHAR(100),
    last_login        TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS users_last_login_idx ON users (last_login DESC);

CREATE TRIGGER trg_users_updated
BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Wallets
CREATE TABLE IF NOT EXISTS wallets (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id                 UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    chain                   chain_enum NOT NULL,
    address                 VARCHAR(100) UNIQUE NOT NULL,
    public_key              TEXT,
    encrypted_private_key   TEXT,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT wallets_user_chain_addr_unique UNIQUE (user_id, chain, address)
);
CREATE INDEX IF NOT EXISTS wallets_user_chain_idx ON wallets (user_id, chain);

CREATE TRIGGER trg_wallets_updated
BEFORE UPDATE ON wallets
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Transactions
CREATE TABLE IF NOT EXISTS transactions (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id     UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    tx_hash       VARCHAR(100) UNIQUE,
    from_address  VARCHAR(100),
    to_address    VARCHAR(100),
    amount        NUMERIC(36,18) NOT NULL CHECK (amount >= 0),
    chain         chain_enum NOT NULL,
    status        tx_status_enum NOT NULL DEFAULT 'pending',
    fee           NUMERIC(36,18),
    block_number  BIGINT,
    timestamp     TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS tx_wallet_time_idx ON transactions (wallet_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS tx_chain_time_idx ON transactions (chain, timestamp DESC);

CREATE TRIGGER trg_transactions_updated
BEFORE UPDATE ON transactions
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Accounts (internal ledger account per user/currency/type)
CREATE TABLE IF NOT EXISTS accounts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type        account_type_enum NOT NULL,
    currency    VARCHAR(10) NOT NULL, -- e.g. BTC, ETH, USDT, USD
    balance     NUMERIC(36,18) NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT accounts_unique_per_user UNIQUE (user_id, type, currency)
);
CREATE INDEX IF NOT EXISTS accounts_user_idx ON accounts (user_id);

CREATE TRIGGER trg_accounts_updated
BEFORE UPDATE ON accounts
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Ledger Entries (double-entry building blocks)
CREATE TABLE IF NOT EXISTS ledger_entries (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id    UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    entry_type    entry_type_enum NOT NULL, -- debit or credit
    amount        NUMERIC(36,18) NOT NULL CHECK (amount > 0),
    reference_id  UUID,          -- points to a business object (e.g., transactions.id) if applicable
    reference_tag VARCHAR(64),   -- optional discriminator, e.g., 'onchain_tx', 'internal_transfer'
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS ledger_account_time_idx ON ledger_entries (account_id, created_at DESC);
CREATE INDEX IF NOT EXISTS ledger_reference_idx ON ledger_entries (reference_id);

-- Blockchain Nodes (RPC endpoints)
CREATE TABLE IF NOT EXISTS blockchain_nodes (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chain         chain_enum NOT NULL,
    rpc_url       TEXT NOT NULL,
    status        node_status_enum NOT NULL DEFAULT 'active',
    last_checked  TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS nodes_chain_status_idx ON blockchain_nodes (chain, status);

-- Sync Jobs (workers/cron states)
CREATE TABLE IF NOT EXISTS sync_jobs (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_name   VARCHAR(100) UNIQUE NOT NULL,
    last_run   TIMESTAMPTZ,
    next_run   TIMESTAMPTZ,
    status     job_status_enum NOT NULL DEFAULT 'idle',
    details    JSONB DEFAULT '{}'::jsonb
);

-- =====================================================================
-- ============================== 2) kyc_db =============================
-- =====================================================================
-- \connect kyc_db

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- ---------- Enums ----------
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'kyc_status_enum') THEN
    CREATE TYPE kyc_status_enum AS ENUM ('pending','verified','rejected');
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'risk_level_enum') THEN
    CREATE TYPE risk_level_enum AS ENUM ('low','medium','high');
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'doc_type_enum') THEN
    CREATE TYPE doc_type_enum AS ENUM ('id_card','passport','selfie','proof_of_address');
  END IF;
END $$;

-- ---------- Helper trigger ----------
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at := NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- KYC Profiles (one-to-one with user by external reference)
CREATE TABLE IF NOT EXISTS kyc_profiles (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID UNIQUE NOT NULL,        -- reference to core_db.users.id (cross-db application-level FK)
    full_name   VARCHAR(255) NOT NULL,
    birth_date  DATE,
    id_number   VARCHAR(50) UNIQUE,
    status      kyc_status_enum NOT NULL DEFAULT 'pending',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS kyc_profiles_status_idx ON kyc_profiles (status);

CREATE TRIGGER trg_kyc_profiles_updated
BEFORE UPDATE ON kyc_profiles
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- KYC Documents
CREATE TABLE IF NOT EXISTS kyc_documents (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id   UUID NOT NULL REFERENCES kyc_profiles(id) ON DELETE CASCADE,
    doc_type     doc_type_enum NOT NULL,
    file_path    TEXT NOT NULL,
    verified_by  VARCHAR(100),
    verified_at  TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS kyc_docs_profile_type_idx ON kyc_documents (profile_id, doc_type);

-- Risk Score
CREATE TABLE IF NOT EXISTS user_risk_score (
    user_id     UUID PRIMARY KEY,
    score       NUMERIC(5,2) NOT NULL DEFAULT 0,
    risk_level  risk_level_enum NOT NULL DEFAULT 'low',
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Alert Rules (for AML/Compliance)
CREATE TABLE IF NOT EXISTS alert_rules (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL,
    condition   JSONB NOT NULL, -- e.g., {"field":"amount_usd","op":">","value":100000}
    threshold   NUMERIC(36,8),
    enabled     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS alert_rules_enabled_idx ON alert_rules (enabled);

-- =====================================================================
-- ============================== 3) rates_db ===========================
-- =====================================================================
-- \connect rates_db

-- No special extensions required, but pgcrypto can be enabled if desired
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Exchange Rates (current snapshots)
CREATE TABLE IF NOT EXISTS exchange_rates (
    id               BIGSERIAL PRIMARY KEY,
    base_currency    VARCHAR(10) NOT NULL,  -- e.g., BTC, ETH
    target_currency  VARCHAR(10) NOT NULL,  -- e.g., USD, THB
    rate             NUMERIC(36,8) NOT NULL,
    fetched_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    source           VARCHAR(50) DEFAULT 'coingecko'
);
CREATE INDEX IF NOT EXISTS rates_pair_time_idx ON exchange_rates (base_currency, target_currency, fetched_at DESC);

-- Price History (time-series, can be partitioned by month/day if needed)
CREATE TABLE IF NOT EXISTS price_history (
    id          BIGSERIAL PRIMARY KEY,
    symbol      VARCHAR(10) NOT NULL,       -- e.g., BTC, ETH
    price_usd   NUMERIC(36,8) NOT NULL,
    timestamp   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS price_symbol_time_idx ON price_history (symbol, timestamp DESC);

-- Aggregated Transaction Summary (ETL from core_db)
CREATE TABLE IF NOT EXISTS transaction_summary (
    date               DATE PRIMARY KEY,
    total_volume_usd   NUMERIC(36,2) NOT NULL DEFAULT 0,
    tx_count           BIGINT NOT NULL DEFAULT 0,
    unique_users       BIGINT NOT NULL DEFAULT 0
);

-- User Activity Summary (ETL from core_db)
CREATE TABLE IF NOT EXISTS user_activity_summary (
    user_id            UUID PRIMARY KEY,
    last_login         TIMESTAMPTZ,
    total_tx           BIGINT NOT NULL DEFAULT 0,
    total_volume_usd   NUMERIC(36,2) NOT NULL DEFAULT 0
);

-- Market Sources (optional catalog of providers)
CREATE TABLE IF NOT EXISTS market_sources (
    id        BIGSERIAL PRIMARY KEY,
    name      VARCHAR(50) UNIQUE NOT NULL,  -- e.g., 'coingecko', 'binance'
    base_url  TEXT,
    active    BOOLEAN NOT NULL DEFAULT TRUE
);

-- =====================================================================
-- ============================== 4) audit_db ===========================
-- =====================================================================
-- \connect audit_db

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Audit Logs (generic event trail)
CREATE TABLE IF NOT EXISTS audit_logs (
    id           BIGSERIAL PRIMARY KEY,
    user_id      UUID,                      -- reference to core_db.users.id (application-level)
    event_type   VARCHAR(100) NOT NULL,     -- e.g., 'LOGIN_SUCCESS', 'CREATE_WALLET'
    payload      JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS audit_event_time_idx ON audit_logs (event_type, created_at DESC);
CREATE INDEX IF NOT EXISTS audit_user_time_idx ON audit_logs (user_id, created_at DESC);

-- API Audit (request/response metadata, no bodies to avoid PII)
CREATE TABLE IF NOT EXISTS api_audit (
    id             BIGSERIAL PRIMARY KEY,
    user_id        UUID,
    endpoint       VARCHAR(255) NOT NULL,
    method         VARCHAR(10) NOT NULL,
    response_code  INT NOT NULL,
    latency_ms     INT CHECK (latency_ms >= 0),
    request_id     UUID DEFAULT gen_random_uuid(),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS api_audit_time_idx ON api_audit (created_at DESC);
CREATE INDEX IF NOT EXISTS api_audit_user_idx ON api_audit (user_id);

-- Security Logs (authz/authn events, anomalies)
CREATE TABLE IF NOT EXISTS security_logs (
    id          BIGSERIAL PRIMARY KEY,
    user_id     UUID,
    event       VARCHAR(100) NOT NULL,     -- e.g., 'FAILED_LOGIN', 'TOKEN_REVOKED'
    details     JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS security_event_time_idx ON security_logs (event, created_at DESC);
CREATE INDEX IF NOT EXISTS security_user_time_idx ON security_logs (user_id, created_at DESC);

-- =====================================================================
-- End of Schema
-- =====================================================================
