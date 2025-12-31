-- +goose Up
-- Indexes to support analytics heavy queries.

CREATE INDEX IF NOT EXISTS idx_price_history_symbol_interval_timestamp
    ON price_history(symbol, interval, timestamp DESC);

CREATE INDEX IF NOT EXISTS idx_transaction_summary_chain_date_desc
    ON transaction_summary(chain, date DESC);
