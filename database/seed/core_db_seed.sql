-- Seed data for core_db
-- Inserts demo users, wallets, transactions, and supporting reference data.

BEGIN;

-- Reference chains (idempotent)
INSERT INTO chains (id, symbol, name, rpc_url, explorer_url, native_token_symbol,
                    native_token_decimals, confirmation_threshold,
                    average_block_time_seconds, is_active, network_type)
VALUES
    (1, 'BTC', 'Bitcoin', 'https://btc.node.local', 'https://mempool.space', 'BTC', 8, 6, 600, TRUE, 'mainnet'),
    (2, 'ETH', 'Ethereum', 'https://eth.node.local', 'https://etherscan.io', 'ETH', 18, 12, 12, TRUE, 'mainnet'),
    (3, 'SOL', 'Solana', 'https://sol.node.local', 'https://explorer.solana.com', 'SOL', 9, 32, 1, TRUE, 'mainnet'),
    (4, 'XLM', 'Stellar', 'https://xlm.node.local', 'https://stellar.expert', 'XLM', 7, 1, 5, TRUE, 'mainnet')
ON CONFLICT (id) DO NOTHING;

-- Reference tokens (idempotent)
INSERT INTO tokens (id, symbol, name, chain_id, decimals, is_native, is_active)
VALUES
    (1, 'BTC', 'Bitcoin', 1, 8, TRUE, TRUE),
    (2, 'ETH', 'Ethereum', 2, 18, TRUE, TRUE),
    (3, 'SOL', 'Solana', 3, 9, TRUE, TRUE),
    (4, 'XLM', 'Stellar Lumens', 4, 7, TRUE, TRUE),
    (5, 'USDC', 'USD Coin', 2, 6, FALSE, TRUE)
ON CONFLICT (id) DO NOTHING;

-- Demo users
INSERT INTO users (id, email, password_hash, first_name, last_name, status,
                   preferred_currency, two_factor_enabled, email_verified,
                   created_at, updated_at)
VALUES
    ('00000000-0000-0000-0000-000000000001', 'alice@example.com', '$2a$12$examplehashalice',
     'Alice', 'Anderson', 'active', 'USD', TRUE, TRUE, '2025-01-01T10:00:00Z', '2025-01-01T10:00:00Z'),
    ('00000000-0000-0000-0000-000000000002', 'bob@example.com', '$2a$12$examplehashbob',
     'Bob', 'Brown', 'active', 'EUR', FALSE, TRUE, '2025-01-02T11:00:00Z', '2025-01-02T11:00:00Z')
ON CONFLICT (email) DO NOTHING;

-- Accounts (one per user)
INSERT INTO accounts (id, user_id, total_balance_usd, last_calculated_at, created_at, updated_at)
VALUES
    ('10000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 12500.42,
     '2025-01-10T09:30:00Z', '2025-01-01T10:05:00Z', '2025-01-10T09:30:00Z'),
    ('10000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000002', 8050.15,
     '2025-01-10T09:45:00Z', '2025-01-02T11:05:00Z', '2025-01-10T09:45:00Z')
ON CONFLICT (user_id) DO NOTHING;

-- Wallets
INSERT INTO wallets (id, user_id, chain, address, encrypted_private_key,
                     derivation_path, label, balance, status,
                     created_at, updated_at)
VALUES
    ('20000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001',
     'BTC', 'bc1qalice0000000000000000000000000000001',
     'encrypted-key-alice-btc', 'm/84''/0''/0''/0/0', 'Alice BTC',
     1.23456789, 'active', '2025-01-03T09:00:00Z', '2025-01-08T14:00:00Z'),
    ('20000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001',
     'ETH', '0xAliceEthereum00000000000000000000001',
     'encrypted-key-alice-eth', 'm/44''/60''/0''/0/0', 'Alice ETH',
     12.500000000000000000, 'active', '2025-01-03T09:10:00Z', '2025-01-08T14:05:00Z'),
    ('20000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000002',
     'SOL', 'SoLbob11111111111111111111111111111111',
     'encrypted-key-bob-sol', 'm/44''/501''/0''/0', 'Bob SOL',
     250.000000000, 'active', '2025-01-04T12:00:00Z', '2025-01-09T08:00:00Z')
ON CONFLICT (chain, address) DO NOTHING;

-- Transactions
INSERT INTO transactions (id, wallet_id, chain, tx_hash, type, amount, fee,
                          status, from_address, to_address, block_number,
                          confirmations, metadata, created_at, confirmed_at, updated_at)
VALUES
    ('30000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001',
     'BTC', 'btc-demo-hash-0001', 'send', 0.25000000, 0.00020000,
     'confirming', 'bc1qalice0000000000000000000000000000001', 'bc1qreceiveralice000000000000001',
     840000, 0, '{"memo":"Seed transaction 1"}', '2025-01-08T10:00:00Z', NULL, '2025-01-08T10:00:00Z'),
    ('30000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000002',
     'ETH', 'eth-demo-hash-0002', 'receive', 3.500000000000000000, 0.000420000000000000,
     'confirmed', '0xSender000000000000000000000000000001', '0xAliceEthereum00000000000000000000001',
     19500000, 48, '{"source":"staking_rewards"}', '2025-01-07T09:30:00Z', '2025-01-07T09:45:00Z', '2025-01-07T09:45:00Z'),
    ('30000000-0000-0000-0000-000000000003', '20000000-0000-0000-0000-000000000003',
     'SOL', 'sol-demo-hash-0003', 'swap_out', 25.000000000, 0.010000000,
     'confirming', 'SoLbob11111111111111111111111111111111', 'SoLdest000000000000000000000000000001',
     250000000, 12, '{"pair":"SOL/USDC"}', '2025-01-09T08:05:00Z', NULL, '2025-01-09T08:05:00Z')
ON CONFLICT (chain, tx_hash) DO NOTHING;

-- Ledger entries
INSERT INTO ledger_entries (id, account_id, transaction_id, entry_type, amount,
                            currency, description, balance_after, created_at)
VALUES
    ('40000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001',
     '30000000-0000-0000-0000-000000000001', 'credit', 0.25020000, 'BTC',
     'Send BTC to demo contact', 0.98436789, '2025-01-08T10:00:00Z'),
    ('40000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000001',
     '30000000-0000-0000-0000-000000000002', 'debit', 3.500420000000000000, 'ETH',
     'Receive staking rewards', 16.000000000000000000, '2025-01-07T09:45:00Z'),
    ('40000000-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000002',
     '30000000-0000-0000-0000-000000000003', 'credit', 25.010000000, 'SOL',
     'Swap out SOL for USDC', 224.990000000, '2025-01-09T08:05:00Z')
ON CONFLICT (id) DO NOTHING;

COMMIT;
