-- +goose Up
-- Rates database initial schema
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Enumerations
CREATE TYPE blockchain_chain AS ENUM ('BTC', 'ETH', 'SOL', 'XLM');
CREATE TYPE price_interval AS ENUM ('1m', '5m', '15m', '1h', '4h', '1d', '1w');

-- Exchange rates
CREATE TABLE exchange_rates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    symbol VARCHAR(20) NOT NULL,
    price_usd DECIMAL(20, 8) NOT NULL,
    price_change_1h DECIMAL(8, 4),
    price_change_24h DECIMAL(8, 4),
    price_change_7d DECIMAL(8, 4),
    volume_24h DECIMAL(20, 2),
    market_cap DECIMAL(20, 2),
    circulating_supply DECIMAL(30, 2),
    total_supply DECIMAL(30, 2),
    source VARCHAR(50) NOT NULL DEFAULT 'coingecko',
    last_updated TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_exchange_rates_symbol ON exchange_rates(symbol);
CREATE INDEX idx_exchange_rates_last_updated ON exchange_rates(last_updated DESC);
CREATE INDEX idx_exchange_rates_market_cap ON exchange_rates(market_cap DESC);

-- Trading pairs
CREATE TABLE trading_pairs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    base_symbol VARCHAR(20) NOT NULL,
    quote_symbol VARCHAR(20) NOT NULL,
    exchange_rate DECIMAL(36, 18) NOT NULL,
    inverse_rate DECIMAL(36, 18) NOT NULL,
    fee_percentage DECIMAL(5, 4) NOT NULL DEFAULT 0.5,
    min_swap_amount DECIMAL(36, 18) NOT NULL DEFAULT 0,
    max_swap_amount DECIMAL(36, 18),
    daily_volume DECIMAL(20, 2) NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    liquidity_available BOOLEAN NOT NULL DEFAULT TRUE,
    last_updated TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(base_symbol, quote_symbol),
    CHECK (base_symbol <> quote_symbol)
);

CREATE INDEX idx_trading_pairs_base_symbol ON trading_pairs(base_symbol);
CREATE INDEX idx_trading_pairs_quote_symbol ON trading_pairs(quote_symbol);
CREATE INDEX idx_trading_pairs_is_active ON trading_pairs(is_active);
CREATE INDEX idx_trading_pairs_last_updated ON trading_pairs(last_updated DESC);

-- Price history
CREATE TABLE price_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    symbol VARCHAR(20) NOT NULL,
    price_usd DECIMAL(20, 8) NOT NULL,
    open DECIMAL(20, 8) NOT NULL,
    high DECIMAL(20, 8) NOT NULL,
    low DECIMAL(20, 8) NOT NULL,
    close DECIMAL(20, 8) NOT NULL,
    volume DECIMAL(20, 2) NOT NULL,
    interval price_interval NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_price_history_symbol_timestamp ON price_history(symbol, timestamp DESC);
CREATE INDEX idx_price_history_interval ON price_history(interval);
CREATE INDEX idx_price_history_timestamp ON price_history(timestamp DESC);
CREATE UNIQUE INDEX idx_price_history_unique ON price_history(symbol, interval, timestamp);

-- Transaction summary
CREATE TABLE transaction_summary (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chain blockchain_chain NOT NULL,
    date DATE NOT NULL,
    transaction_count INTEGER NOT NULL DEFAULT 0,
    total_volume_usd DECIMAL(20, 2) NOT NULL DEFAULT 0,
    unique_senders INTEGER NOT NULL DEFAULT 0,
    unique_receivers INTEGER NOT NULL DEFAULT 0,
    avg_transaction_value_usd DECIMAL(15, 2) NOT NULL DEFAULT 0,
    swap_count INTEGER NOT NULL DEFAULT 0,
    total_fees_usd DECIMAL(15, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(chain, date)
);

CREATE INDEX idx_transaction_summary_chain_date ON transaction_summary(chain, date DESC);
CREATE INDEX idx_transaction_summary_date ON transaction_summary(date DESC);
