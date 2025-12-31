-- +goose Up
-- Core database initial schema
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Enumerations
CREATE TYPE user_status AS ENUM ('active', 'suspended', 'deleted');
CREATE TYPE currency_code AS ENUM ('USD', 'EUR', 'THB', 'GBP', 'JPY');
CREATE TYPE blockchain_chain AS ENUM ('BTC', 'ETH', 'SOL', 'XLM');
CREATE TYPE wallet_status AS ENUM ('active', 'archived');
CREATE TYPE transaction_type AS ENUM ('send', 'receive', 'swap_in', 'swap_out');
CREATE TYPE transaction_status AS ENUM ('pending', 'confirming', 'confirmed', 'failed', 'cancelled');
CREATE TYPE entry_type AS ENUM ('debit', 'credit');
CREATE TYPE exchange_status AS ENUM ('pending', 'processing', 'completed', 'failed', 'cancelled');

-- Users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    phone_number VARCHAR(20),
    status user_status NOT NULL DEFAULT 'active',
    preferred_currency currency_code NOT NULL DEFAULT 'USD',
    two_factor_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    two_factor_secret VARCHAR(255),
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    email_verified_at TIMESTAMP WITH TIME ZONE,
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_created_at ON users(created_at);

-- Wallets
CREATE TABLE wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    chain blockchain_chain NOT NULL,
    address VARCHAR(255) NOT NULL,
    encrypted_private_key TEXT NOT NULL,
    derivation_path VARCHAR(100),
    label VARCHAR(100),
    balance DECIMAL(36, 18) NOT NULL DEFAULT 0,
    balance_updated_at TIMESTAMP WITH TIME ZONE,
    status wallet_status NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(chain, address)
);

CREATE INDEX idx_wallets_user_id ON wallets(user_id);
CREATE INDEX idx_wallets_chain ON wallets(chain);
CREATE INDEX idx_wallets_address ON wallets(address);
CREATE INDEX idx_wallets_status ON wallets(status);

-- Transactions
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    chain blockchain_chain NOT NULL,
    tx_hash VARCHAR(255) NOT NULL,
    type transaction_type NOT NULL,
    amount DECIMAL(36, 18) NOT NULL,
    fee DECIMAL(36, 18) NOT NULL DEFAULT 0,
    status transaction_status NOT NULL DEFAULT 'pending',
    from_address VARCHAR(255) NOT NULL,
    to_address VARCHAR(255) NOT NULL,
    block_number BIGINT,
    confirmations INTEGER NOT NULL DEFAULT 0,
    error_message TEXT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    confirmed_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(chain, tx_hash)
);

CREATE INDEX idx_transactions_wallet_id ON transactions(wallet_id);
CREATE INDEX idx_transactions_chain ON transactions(chain);
CREATE INDEX idx_transactions_tx_hash ON transactions(tx_hash);
CREATE INDEX idx_transactions_status ON transactions(status);
CREATE INDEX idx_transactions_type ON transactions(type);
CREATE INDEX idx_transactions_created_at ON transactions(created_at DESC);
CREATE INDEX idx_transactions_from_address ON transactions(from_address);
CREATE INDEX idx_transactions_to_address ON transactions(to_address);

-- Chains
CREATE TABLE chains (
    id SERIAL PRIMARY KEY,
    symbol blockchain_chain NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    rpc_url VARCHAR(500) NOT NULL,
    rpc_url_fallback VARCHAR(500),
    explorer_url VARCHAR(500) NOT NULL,
    native_token_symbol VARCHAR(10) NOT NULL,
    native_token_decimals INTEGER NOT NULL DEFAULT 18,
    confirmation_threshold INTEGER NOT NULL,
    average_block_time_seconds INTEGER NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    network_type VARCHAR(20) NOT NULL DEFAULT 'mainnet',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_chains_is_active ON chains(is_active);

-- Tokens
CREATE TABLE tokens (
    id SERIAL PRIMARY KEY,
    symbol VARCHAR(20) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    chain_id INTEGER NOT NULL REFERENCES chains(id) ON DELETE CASCADE,
    contract_address VARCHAR(255),
    decimals INTEGER NOT NULL,
    is_native BOOLEAN NOT NULL DEFAULT FALSE,
    logo_url VARCHAR(500),
    coingecko_id VARCHAR(100),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tokens_chain_id ON tokens(chain_id);
CREATE INDEX idx_tokens_is_active ON tokens(is_active);
CREATE INDEX idx_tokens_is_native ON tokens(is_native);

-- Accounts
CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    total_balance_usd DECIMAL(20, 2) NOT NULL DEFAULT 0,
    last_calculated_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_accounts_user_id ON accounts(user_id);
CREATE INDEX idx_accounts_total_balance_usd ON accounts(total_balance_usd DESC);

-- Ledger entries
CREATE TABLE ledger_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    transaction_id UUID REFERENCES transactions(id) ON DELETE SET NULL,
    entry_type entry_type NOT NULL,
    amount DECIMAL(36, 18) NOT NULL,
    currency VARCHAR(10) NOT NULL,
    description TEXT NOT NULL,
    balance_after DECIMAL(36, 18) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ledger_entries_account_id ON ledger_entries(account_id);
CREATE INDEX idx_ledger_entries_transaction_id ON ledger_entries(transaction_id);
CREATE INDEX idx_ledger_entries_created_at ON ledger_entries(created_at DESC);
CREATE INDEX idx_ledger_entries_currency ON ledger_entries(currency);

-- Exchange operations
CREATE TABLE exchange_operations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    from_wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    to_wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    from_amount DECIMAL(36, 18) NOT NULL,
    to_amount DECIMAL(36, 18) NOT NULL,
    exchange_rate DECIMAL(36, 18) NOT NULL,
    fee_percentage DECIMAL(5, 4) NOT NULL,
    fee_amount DECIMAL(36, 18) NOT NULL,
    status exchange_status NOT NULL DEFAULT 'pending',
    from_transaction_id UUID REFERENCES transactions(id) ON DELETE SET NULL,
    to_transaction_id UUID REFERENCES transactions(id) ON DELETE SET NULL,
    quote_expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    executed_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_exchange_operations_user_id ON exchange_operations(user_id);
CREATE INDEX idx_exchange_operations_status ON exchange_operations(status);
CREATE INDEX idx_exchange_operations_created_at ON exchange_operations(created_at DESC);
CREATE INDEX idx_exchange_operations_from_wallet_id ON exchange_operations(from_wallet_id);
CREATE INDEX idx_exchange_operations_to_wallet_id ON exchange_operations(to_wallet_id);

-- Seed essential chain and token data
INSERT INTO chains (symbol, name, rpc_url, rpc_url_fallback, explorer_url, native_token_symbol, native_token_decimals, confirmation_threshold, average_block_time_seconds)
VALUES
    ('BTC', 'Bitcoin', 'https://bitcoin-rpc.example.com', NULL, 'https://blockchain.info/tx/', 'BTC', 8, 6, 600),
    ('ETH', 'Ethereum', 'https://ethereum-rpc.example.com', NULL, 'https://etherscan.io/tx/', 'ETH', 18, 12, 12),
    ('SOL', 'Solana', 'https://solana-rpc.example.com', NULL, 'https://explorer.solana.com/tx/', 'SOL', 9, 32, 0),
    ('XLM', 'Stellar', 'https://horizon.stellar.org', NULL, 'https://stellarchain.io/tx/', 'XLM', 7, 1, 5);

INSERT INTO tokens (symbol, name, chain_id, contract_address, decimals, is_native, coingecko_id)
VALUES
    ('BTC', 'Bitcoin', (SELECT id FROM chains WHERE symbol = 'BTC'), NULL, 8, TRUE, 'bitcoin'),
    ('ETH', 'Ethereum', (SELECT id FROM chains WHERE symbol = 'ETH'), NULL, 18, TRUE, 'ethereum'),
    ('SOL', 'Solana', (SELECT id FROM chains WHERE symbol = 'SOL'), NULL, 9, TRUE, 'solana'),
    ('XLM', 'Stellar Lumens', (SELECT id FROM chains WHERE symbol = 'XLM'), NULL, 7, TRUE, 'stellar');
