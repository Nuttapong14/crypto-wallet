-- +goose Up
-- Additional indexes to improve high-volume query paths.

CREATE INDEX IF NOT EXISTS idx_wallets_user_created
    ON wallets(user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_transactions_wallet_created
    ON transactions(wallet_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_transactions_chain_status_created
    ON transactions(chain, status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_ledger_entries_account_created
    ON ledger_entries(account_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_exchange_operations_user_status_created
    ON exchange_operations(user_id, status, created_at DESC);
