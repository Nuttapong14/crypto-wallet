# Data Model Design: Multi-Chain Crypto Wallet and Exchange Platform

**Feature**: Multi-Chain Crypto Wallet and Exchange Platform
**Branch**: `001-build-a-web`
**Date**: 2025-10-14
**Status**: Design Complete

---

## Overview

This document defines the complete data model for the multi-chain cryptocurrency wallet platform, organized across 4 separate PostgreSQL databases following Domain-Driven Design principles with clear separation of concerns.

**Database Architecture**:
- **core_db**: Operational data (users, wallets, transactions, chains, tokens, ledger)
- **kyc_db**: Compliance data (profiles, documents, risk scores, verification status)
- **rates_db**: Market data (exchange rates, trading pairs, price history, analytics)
- **audit_db**: Security logs (audit trails, API logs, security events)

**Design Principles**:
- UUID primary keys for distributed system compatibility
- Timestamp tracking for all entities (created_at, updated_at)
- Indexed foreign keys for query performance
- Enum types for controlled vocabularies
- JSON fields for flexible semi-structured data
- No cross-database foreign keys (logical relationships only)

---

## 1. Core Database (core_db)

### 1.1 Users Table

**Purpose**: Central user account management

```sql
CREATE TYPE user_status AS ENUM ('active', 'suspended', 'deleted');
CREATE TYPE currency_code AS ENUM ('USD', 'EUR', 'THB', 'GBP', 'JPY');

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
```

**Fields**:
- `id`: UUID primary key
- `email`: Unique email address (indexed)
- `password_hash`: bcrypt hashed password (cost factor 12+)
- `first_name`, `last_name`: Optional profile information
- `phone_number`: Optional for 2FA/notifications
- `status`: Account status (active/suspended/deleted)
- `preferred_currency`: Display currency for portfolio values
- `two_factor_enabled`: 2FA activation flag
- `two_factor_secret`: Encrypted TOTP secret
- `email_verified`: Email verification status
- `last_login_at`: Track user activity

**Validation Rules**:
- Email must match RFC 5322 format
- Password minimum 12 characters with complexity requirements
- Phone number must match international format if provided
- Status transitions: active ↔ suspended, any → deleted (soft delete)

**Relationships**:
- 1:N with Wallets
- 1:1 with Accounts (aggregate balance)
- Logical 1:1 with KYC_Profiles (via user_id, separate database)

---

### 1.2 Wallets Table

**Purpose**: Blockchain-specific cryptocurrency wallets

```sql
CREATE TYPE blockchain_chain AS ENUM ('BTC', 'ETH', 'SOL', 'XLM');
CREATE TYPE wallet_status AS ENUM ('active', 'archived');

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
```

**Fields**:
- `id`: UUID primary key
- `user_id`: Foreign key to users
- `chain`: Blockchain network (BTC/ETH/SOL/XLM)
- `address`: Blockchain-specific wallet address
- `encrypted_private_key`: AES-256-GCM encrypted private key
- `derivation_path`: BIP32/BIP44 derivation path for HD wallets
- `label`: User-friendly wallet name
- `balance`: Current balance (updated by background worker)
- `balance_updated_at`: Last balance sync timestamp
- `status`: Wallet status (active/archived)

**Validation Rules**:
- Address format must match blockchain specification:
  - BTC: P2PKH (1...), P2SH (3...), Bech32 (bc1...)
  - ETH: 0x + 40 hex characters
  - SOL: Base58 encoded 32-44 characters
  - XLM: G + 55 characters (public key format)
- Balance must be non-negative
- Encrypted private key required for all wallets
- Unique address per blockchain (same address can exist on different chains)

**State Transitions**:
- active → archived (user archives wallet)
- No deletion (preserve transaction history)

**Relationships**:
- N:1 with Users
- 1:N with Transactions

---

### 1.3 Transactions Table

**Purpose**: Cryptocurrency transfer operations

```sql
CREATE TYPE transaction_type AS ENUM ('send', 'receive', 'swap_in', 'swap_out');
CREATE TYPE transaction_status AS ENUM ('pending', 'confirming', 'confirmed', 'failed', 'cancelled');

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
```

**Fields**:
- `id`: UUID primary key
- `wallet_id`: Foreign key to wallets
- `chain`: Blockchain network
- `tx_hash`: Blockchain transaction hash (unique per chain)
- `type`: Transaction direction/purpose
- `amount`: Transfer amount (in native token units)
- `fee`: Network transaction fee
- `status`: Current transaction state
- `from_address`: Sender address
- `to_address`: Recipient address
- `block_number`: Block containing transaction
- `confirmations`: Number of confirmations
- `error_message`: Failure reason if status=failed
- `metadata`: Additional blockchain-specific data (gas price, memo, etc.)
- `confirmed_at`: Timestamp when transaction reached required confirmations

**Validation Rules**:
- Amount and fee must be positive
- Status must follow valid state transitions
- Addresses must match blockchain format
- TX hash format varies by chain:
  - BTC: 64 hex characters
  - ETH: 0x + 64 hex characters
  - SOL: Base58 encoded
  - XLM: 64 hex characters
- Required confirmations by chain:
  - BTC: 6 confirmations
  - ETH: 12 confirmations
  - SOL: 32 confirmations
  - XLM: 1 confirmation (5 second ledger close)

**State Transitions**:
```
pending → confirming (broadcast to network)
confirming → confirmed (sufficient confirmations)
confirming → failed (network rejection)
pending → cancelled (user cancellation before broadcast)
```

**Relationships**:
- N:1 with Wallets
- 1:N with Ledger_Entries

**Metadata Examples**:
```json
// Ethereum
{
  "gas_price": "50000000000",
  "gas_limit": "21000",
  "gas_used": "21000",
  "nonce": 42,
  "contract_address": "0x..."
}

// Stellar
{
  "memo_type": "text",
  "memo": "Payment for services",
  "operation_id": "123456789"
}

// Solana
{
  "recent_blockhash": "...",
  "compute_units": 200000
}
```

---

### 1.4 Chains Table

**Purpose**: Supported blockchain network configuration

```sql
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

CREATE INDEX idx_chains_symbol ON chains(symbol);
CREATE INDEX idx_chains_is_active ON chains(is_active);
```

**Fields**:
- `id`: Serial primary key
- `symbol`: Blockchain identifier (BTC/ETH/SOL/XLM)
- `name`: Full blockchain name
- `rpc_url`: Primary RPC endpoint
- `rpc_url_fallback`: Backup RPC endpoint
- `explorer_url`: Block explorer URL template
- `native_token_symbol`: Native cryptocurrency symbol
- `native_token_decimals`: Decimal precision
- `confirmation_threshold`: Required confirmations for finality
- `average_block_time_seconds`: Average block time
- `is_active`: Enable/disable chain
- `network_type`: mainnet/testnet

**Initial Data**:
```sql
INSERT INTO chains (symbol, name, rpc_url, explorer_url, native_token_symbol, native_token_decimals, confirmation_threshold, average_block_time_seconds) VALUES
('BTC', 'Bitcoin', 'https://bitcoin-rpc.example.com', 'https://blockchain.info/tx/', 'BTC', 8, 6, 600),
('ETH', 'Ethereum', 'https://ethereum-rpc.example.com', 'https://etherscan.io/tx/', 'ETH', 18, 12, 12),
('SOL', 'Solana', 'https://solana-rpc.example.com', 'https://explorer.solana.com/tx/', 'SOL', 9, 32, 0),
('XLM', 'Stellar', 'https://horizon.stellar.org', 'https://stellarchain.io/tx/', 'XLM', 7, 1, 5);
```

**Relationships**:
- 1:N with Tokens

---

### 1.5 Tokens Table

**Purpose**: Supported cryptocurrencies (native and tokens)

```sql
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

CREATE INDEX idx_tokens_symbol ON tokens(symbol);
CREATE INDEX idx_tokens_chain_id ON tokens(chain_id);
CREATE INDEX idx_tokens_is_active ON tokens(is_active);
CREATE INDEX idx_tokens_is_native ON tokens(is_native);
```

**Fields**:
- `id`: Serial primary key
- `symbol`: Token symbol (BTC, ETH, USDT, etc.)
- `name`: Full token name
- `chain_id`: Foreign key to chains
- `contract_address`: Smart contract address (null for native tokens)
- `decimals`: Token decimal precision
- `is_native`: Flag for blockchain native token
- `logo_url`: Token icon URL
- `coingecko_id`: CoinGecko API identifier for price feeds
- `is_active`: Enable/disable token

**Initial Data**:
```sql
INSERT INTO tokens (symbol, name, chain_id, decimals, is_native, coingecko_id) VALUES
('BTC', 'Bitcoin', 1, 8, TRUE, 'bitcoin'),
('ETH', 'Ethereum', 2, 18, TRUE, 'ethereum'),
('SOL', 'Solana', 3, 9, TRUE, 'solana'),
('XLM', 'Stellar Lumens', 4, 7, TRUE, 'stellar');
```

**Validation Rules**:
- Contract address required for non-native tokens
- Contract address must be null for native tokens
- Decimals must be between 0 and 18
- Symbol must be uppercase, 2-20 characters

**Relationships**:
- N:1 with Chains

---

### 1.6 Accounts Table

**Purpose**: User balance aggregation for portfolio view

```sql
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
```

**Fields**:
- `id`: UUID primary key
- `user_id`: Foreign key to users (unique)
- `total_balance_usd`: Aggregate portfolio value in USD
- `last_calculated_at`: Last calculation timestamp

**Calculation Logic**:
```
total_balance_usd = SUM(wallet.balance * current_price_usd) for all user wallets
```

**Update Triggers**:
- Wallet balance change
- Price update (background job recalculates periodically)
- New wallet creation

**Relationships**:
- 1:1 with Users
- 1:N with Ledger_Entries

---

### 1.7 Ledger_Entries Table

**Purpose**: Double-entry accounting for financial integrity

```sql
CREATE TYPE entry_type AS ENUM ('debit', 'credit');

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
```

**Fields**:
- `id`: UUID primary key
- `account_id`: Foreign key to accounts
- `transaction_id`: Optional foreign key to transactions
- `entry_type`: Debit or credit
- `amount`: Entry amount (always positive)
- `currency`: Token symbol
- `description`: Human-readable entry description
- `balance_after`: Running balance after this entry
- `created_at`: Entry timestamp (immutable)

**Double-Entry Rules**:
- Every transaction creates two entries: one debit, one credit
- Total debits must equal total credits per transaction
- Entries are immutable once created
- Balance_after provides audit trail

**Example Entries** (Send Transaction):
```sql
-- User A sends 1 ETH to User B (fee: 0.001 ETH)
-- User A entries (sender)
INSERT INTO ledger_entries (account_id, transaction_id, entry_type, amount, currency, description)
VALUES
  (user_a_account, tx_id, 'debit', 1.0, 'ETH', 'Sent ETH to 0xabc...'),
  (user_a_account, tx_id, 'debit', 0.001, 'ETH', 'Transaction fee');

-- User B entries (recipient)
INSERT INTO ledger_entries (account_id, transaction_id, entry_type, amount, currency, description)
VALUES
  (user_b_account, tx_id, 'credit', 1.0, 'ETH', 'Received ETH from 0xdef...');
```

**Relationships**:
- N:1 with Accounts
- N:1 with Transactions (optional)

---

### 1.8 Exchange_Operations Table

**Purpose**: Track cryptocurrency swap operations

```sql
CREATE TYPE exchange_status AS ENUM ('pending', 'processing', 'completed', 'failed', 'cancelled');

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
```

**Fields**:
- `id`: UUID primary key
- `user_id`: Foreign key to users
- `from_wallet_id`: Source wallet
- `to_wallet_id`: Destination wallet
- `from_amount`: Amount to exchange (source currency)
- `to_amount`: Amount received (destination currency)
- `exchange_rate`: Rate at execution time
- `fee_percentage`: Platform fee (0-100%)
- `fee_amount`: Absolute fee amount
- `status`: Operation status
- `from_transaction_id`: Debit transaction ID
- `to_transaction_id`: Credit transaction ID
- `quote_expires_at`: Quote expiration (30-60 seconds)
- `executed_at`: Execution timestamp
- `error_message`: Failure reason

**State Transitions**:
```
pending → processing (user confirms within quote validity)
processing → completed (both transactions confirmed)
processing → failed (transaction failure)
pending → cancelled (quote expired or user cancelled)
```

**Atomicity Guarantee**:
- Both transactions must succeed or entire operation fails
- Database transaction ensures atomic state changes
- Rollback mechanism for partial failures

**Relationships**:
- N:1 with Users
- N:1 with Wallets (from)
- N:1 with Wallets (to)
- 1:1 with Transactions (from, optional)
- 1:1 with Transactions (to, optional)

---

## 2. KYC Database (kyc_db)

**Purpose**: Isolated storage for compliance and PII data

**Security Requirements**:
- AES-256 encryption for all PII fields
- Restricted database access (separate credentials)
- No foreign keys to core_db (logical relationships only)
- Audit all access to this database

### 2.1 KYC_Profiles Table

**Purpose**: User identity verification status

```sql
CREATE TYPE verification_level AS ENUM ('unverified', 'basic', 'full');
CREATE TYPE kyc_status AS ENUM ('not_started', 'pending', 'under_review', 'approved', 'rejected', 'expired');

CREATE TABLE kyc_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE, -- Logical FK to core_db.users (no constraint)
    verification_level verification_level NOT NULL DEFAULT 'unverified',
    status kyc_status NOT NULL DEFAULT 'not_started',
    first_name_encrypted TEXT,
    last_name_encrypted TEXT,
    date_of_birth_encrypted TEXT,
    nationality_encrypted TEXT,
    document_number_encrypted TEXT,
    address_encrypted TEXT,
    submitted_at TIMESTAMP WITH TIME ZONE,
    reviewed_at TIMESTAMP WITH TIME ZONE,
    approved_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    rejection_reason TEXT,
    reviewer_notes TEXT,
    daily_limit_usd DECIMAL(15, 2) NOT NULL DEFAULT 500,
    monthly_limit_usd DECIMAL(15, 2) NOT NULL DEFAULT 5000,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_kyc_profiles_user_id ON kyc_profiles(user_id);
CREATE INDEX idx_kyc_profiles_status ON kyc_profiles(status);
CREATE INDEX idx_kyc_profiles_verification_level ON kyc_profiles(verification_level);
CREATE INDEX idx_kyc_profiles_submitted_at ON kyc_profiles(submitted_at);
```

**Fields**:
- `id`: UUID primary key
- `user_id`: Logical reference to core_db.users (no FK)
- `verification_level`: User's verified level
- `status`: Current verification status
- `*_encrypted`: AES-256-GCM encrypted PII fields
- `submitted_at`: Initial submission timestamp
- `reviewed_at`: Review completion timestamp
- `approved_at`: Approval timestamp
- `expires_at`: Verification expiration (annual renewal)
- `rejection_reason`: Feedback for rejected applications
- `reviewer_notes`: Internal compliance notes
- `daily_limit_usd`: Transaction limit per day
- `monthly_limit_usd`: Transaction limit per month

**Verification Levels & Limits**:
```
unverified: $500/day, $5,000/month
basic: $5,000/day, $50,000/month (requires ID verification)
full: $50,000/day, $500,000/month (requires proof of address + enhanced due diligence)
```

**State Transitions**:
```
not_started → pending (user submits documents)
pending → under_review (compliance team reviews)
under_review → approved (verification successful)
under_review → rejected (verification failed)
approved → expired (annual expiration, requires renewal)
rejected → pending (user resubmits)
```

**Encryption/Decryption**:
```go
// Application layer decrypts on read, encrypts on write
encryptedData := encrypt(plaintextPII, masterKey, nonce)
plaintextPII := decrypt(encryptedData, masterKey, nonce)
```

**Relationships**:
- Logical 1:1 with core_db.Users (via user_id)
- 1:N with KYC_Documents

---

### 2.2 KYC_Documents Table

**Purpose**: Uploaded identity verification documents

```sql
CREATE TYPE document_type AS ENUM ('passport', 'national_id', 'drivers_license', 'proof_of_address', 'selfie');
CREATE TYPE document_status AS ENUM ('pending', 'approved', 'rejected');

CREATE TABLE kyc_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kyc_profile_id UUID NOT NULL REFERENCES kyc_profiles(id) ON DELETE CASCADE,
    document_type document_type NOT NULL,
    file_path_encrypted TEXT NOT NULL,
    file_name_encrypted TEXT NOT NULL,
    file_size_bytes INTEGER NOT NULL,
    file_hash VARCHAR(64) NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    status document_status NOT NULL DEFAULT 'pending',
    uploaded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    reviewed_at TIMESTAMP WITH TIME ZONE,
    rejection_reason TEXT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_kyc_documents_kyc_profile_id ON kyc_documents(kyc_profile_id);
CREATE INDEX idx_kyc_documents_status ON kyc_documents(status);
CREATE INDEX idx_kyc_documents_document_type ON kyc_documents(document_type);
CREATE INDEX idx_kyc_documents_uploaded_at ON kyc_documents(uploaded_at);
```

**Fields**:
- `id`: UUID primary key
- `kyc_profile_id`: Foreign key to KYC_Profiles
- `document_type`: Type of document uploaded
- `file_path_encrypted`: Encrypted storage path
- `file_name_encrypted`: Encrypted original filename
- `file_size_bytes`: File size for validation
- `file_hash`: SHA-256 hash for integrity verification
- `mime_type`: File MIME type (image/jpeg, application/pdf, etc.)
- `status`: Review status
- `uploaded_at`: Upload timestamp
- `reviewed_at`: Review completion timestamp
- `rejection_reason`: Reason if rejected
- `metadata`: Additional document metadata

**File Storage Strategy**:
```
Storage Location: Encrypted S3 bucket or encrypted filesystem
Path Format: /kyc-documents/{year}/{month}/{kyc_profile_id}/{document_id}.enc
Encryption: File-level AES-256-GCM before storage
Access: Pre-signed URLs with 5-minute expiration for compliance team
Retention: 7 years after account closure (regulatory requirement)
```

**Validation Rules**:
- File size: max 10MB per document
- Allowed MIME types: image/jpeg, image/png, application/pdf
- File hash must match uploaded content
- Maximum 5 documents per type per profile

**Required Documents by Verification Level**:
```
basic: passport OR national_id + selfie
full: basic documents + proof_of_address (utility bill, bank statement <3 months)
```

**Relationships**:
- N:1 with KYC_Profiles

---

### 2.3 User_Risk_Score Table

**Purpose**: AML risk assessment for transaction monitoring

```sql
CREATE TYPE risk_level AS ENUM ('low', 'medium', 'high', 'critical');

CREATE TABLE user_risk_scores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE, -- Logical FK to core_db.users
    risk_score INTEGER NOT NULL CHECK (risk_score >= 0 AND risk_score <= 100),
    risk_level risk_level NOT NULL,
    risk_factors JSONB NOT NULL DEFAULT '[]',
    aml_hits JSONB NOT NULL DEFAULT '[]',
    last_screening_at TIMESTAMP WITH TIME ZONE,
    next_review_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_risk_scores_user_id ON user_risk_scores(user_id);
CREATE INDEX idx_user_risk_scores_risk_level ON user_risk_scores(risk_level);
CREATE INDEX idx_user_risk_scores_risk_score ON user_risk_scores(risk_score DESC);
CREATE INDEX idx_user_risk_scores_next_review_at ON user_risk_scores(next_review_at);
```

**Fields**:
- `id`: UUID primary key
- `user_id`: Logical reference to core_db.users
- `risk_score`: Numerical score (0-100)
- `risk_level`: Risk category
- `risk_factors`: JSON array of contributing factors
- `aml_hits`: JSON array of AML screening hits
- `last_screening_at`: Last AML database check
- `next_review_at`: Scheduled next review

**Risk Score Calculation**:
```
Base Score: 20 (all users start here)

Factors:
+ High-risk jurisdiction: +30
+ Large transaction frequency: +20
+ Unusual transaction patterns: +15
+ Failed KYC attempts: +10
+ Rapid account growth: +10
+ Suspicious IP/location changes: +10
- Verified KYC (full): -10
- Long-term good standing: -5

Risk Level Mapping:
0-30: low
31-60: medium
61-85: high
86-100: critical
```

**AML Screening**:
- Periodic checks against OFAC, UN, EU sanctions lists
- PEP (Politically Exposed Person) screening
- Adverse media screening
- Automated via third-party AML provider (ComplyAdvantage, etc.)

**Risk Factors Example**:
```json
[
  {
    "factor": "high_risk_jurisdiction",
    "score": 30,
    "details": "IP address from sanctioned country",
    "timestamp": "2025-10-14T10:30:00Z"
  },
  {
    "factor": "large_transaction_volume",
    "score": 20,
    "details": "Transactions exceeded $50k in 24 hours",
    "timestamp": "2025-10-14T15:00:00Z"
  }
]
```

**Actions by Risk Level**:
```
low: Normal monitoring
medium: Enhanced transaction monitoring
high: Manual review required for large transactions
critical: Account suspension, mandatory investigation
```

**Relationships**:
- Logical 1:1 with core_db.Users (via user_id)

---

### 2.4 Alert_Rules Table

**Purpose**: Configurable AML alert triggers

```sql
CREATE TYPE rule_type AS ENUM ('transaction_amount', 'transaction_frequency', 'velocity', 'geography', 'pattern');

CREATE TABLE alert_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_name VARCHAR(200) NOT NULL UNIQUE,
    rule_type rule_type NOT NULL,
    threshold_value DECIMAL(15, 2) NOT NULL,
    time_window_minutes INTEGER,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    severity VARCHAR(20) NOT NULL DEFAULT 'medium',
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_alert_rules_is_active ON alert_rules(is_active);
CREATE INDEX idx_alert_rules_rule_type ON alert_rules(rule_type);
```

**Fields**:
- `id`: UUID primary key
- `rule_name`: Unique rule identifier
- `rule_type`: Category of rule
- `threshold_value`: Trigger threshold
- `time_window_minutes`: Time window for rule evaluation
- `is_active`: Enable/disable rule
- `severity`: Alert severity level
- `description`: Rule explanation

**Example Rules**:
```sql
INSERT INTO alert_rules (rule_name, rule_type, threshold_value, time_window_minutes, severity, description) VALUES
('Large Single Transaction', 'transaction_amount', 10000.00, NULL, 'high', 'Single transaction exceeds $10,000'),
('High Frequency Trading', 'transaction_frequency', 20, 60, 'medium', 'More than 20 transactions in 1 hour'),
('Rapid Deposit-Withdrawal', 'velocity', 5000.00, 1440, 'high', 'Deposit and withdrawal of $5k+ within 24 hours'),
('Sanctioned Country Access', 'geography', 1, NULL, 'critical', 'Access from sanctioned jurisdiction'),
('Structuring Pattern', 'pattern', 9500.00, 10080, 'high', 'Multiple transactions just below $10k reporting threshold in 7 days');
```

**Relationships**:
- Configuration table (no direct relationships)

---

## 3. Rates Database (rates_db)

**Purpose**: Market data and analytics

### 3.1 Exchange_Rates Table

**Purpose**: Current cryptocurrency prices

```sql
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
```

**Fields**:
- `id`: UUID primary key
- `symbol`: Cryptocurrency symbol (BTC, ETH, etc.)
- `price_usd`: Current USD price
- `price_change_*`: Percentage change over time periods
- `volume_24h`: 24-hour trading volume
- `market_cap`: Total market capitalization
- `circulating_supply`: Circulating token supply
- `total_supply`: Total token supply
- `source`: Data provider (coingecko, binance, etc.)
- `last_updated`: Last price update timestamp

**Update Strategy**:
- Real-time updates via WebSocket for active trading pairs
- Polling fallback every 5 seconds for other pairs
- Cached in Redis with 10-second TTL
- Historical snapshot saved to Price_History every 5 minutes

**Validation Rules**:
- Prices must be positive
- Price changes expressed as percentages (-100% to unlimited)
- last_updated within last 60 seconds for critical monitoring

**Relationships**:
- No direct relationships (market data)

---

### 3.2 Trading_Pairs Table

**Purpose**: Supported exchange pairs with rates

```sql
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
    CHECK(base_symbol != quote_symbol)
);

CREATE INDEX idx_trading_pairs_base_symbol ON trading_pairs(base_symbol);
CREATE INDEX idx_trading_pairs_quote_symbol ON trading_pairs(quote_symbol);
CREATE INDEX idx_trading_pairs_is_active ON trading_pairs(is_active);
CREATE INDEX idx_trading_pairs_last_updated ON trading_pairs(last_updated DESC);
```

**Fields**:
- `id`: UUID primary key
- `base_symbol`: Source token
- `quote_symbol`: Destination token
- `exchange_rate`: Conversion rate (base → quote)
- `inverse_rate`: Inverse rate (quote → base)
- `fee_percentage`: Platform fee (e.g., 0.5%)
- `min_swap_amount`: Minimum swap amount
- `max_swap_amount`: Maximum swap amount per transaction
- `daily_volume`: 24-hour swap volume
- `is_active`: Enable/disable pair
- `liquidity_available`: Sufficient liquidity flag
- `last_updated`: Rate update timestamp

**Initial Trading Pairs**:
```sql
-- All combinations for BTC, ETH, SOL, XLM (12 pairs total)
INSERT INTO trading_pairs (base_symbol, quote_symbol, exchange_rate, inverse_rate, fee_percentage) VALUES
('BTC', 'ETH', 15.5, 0.0645, 0.5),
('BTC', 'SOL', 450.0, 0.0022, 0.5),
('BTC', 'XLM', 150000.0, 0.0000067, 0.5),
('ETH', 'BTC', 0.0645, 15.5, 0.5),
('ETH', 'SOL', 29.0, 0.0345, 0.5),
('ETH', 'XLM', 9680.0, 0.000103, 0.5),
-- ... (remaining pairs)
```

**Rate Calculation**:
```
exchange_rate = price_quote_usd / price_base_usd
inverse_rate = 1 / exchange_rate

Example:
BTC price: $50,000
ETH price: $3,200
BTC → ETH rate: 3200 / 50000 = 0.064 ETH per BTC (or ~15.6 BTC per ETH)
```

**Relationships**:
- No direct relationships (market data)

---

### 3.3 Price_History Table

**Purpose**: Historical price data for charts and analytics

```sql
CREATE TYPE price_interval AS ENUM ('1m', '5m', '15m', '1h', '4h', '1d', '1w');

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
```

**Fields**:
- `id`: UUID primary key
- `symbol`: Cryptocurrency symbol
- `price_usd`: Close price
- `open`, `high`, `low`, `close`: OHLC candlestick data
- `volume`: Trading volume for interval
- `interval`: Time interval (1m, 5m, 1h, 1d, etc.)
- `timestamp`: Start of interval

**Data Retention**:
```
1m interval: 7 days
5m interval: 30 days
15m interval: 90 days
1h interval: 1 year
4h interval: 2 years
1d interval: 5 years
1w interval: 10 years
```

**Aggregation Strategy**:
- 1m data: collected in real-time from WebSocket
- Higher intervals: aggregated from 1m data by background job
- Partitioned by month for query performance
- Compressed older data (pg_cron + pg_partman)

**Query Optimization**:
```sql
-- Partition by month
CREATE TABLE price_history_2025_10 PARTITION OF price_history
FOR VALUES FROM ('2025-10-01') TO ('2025-11-01');
```

**Relationships**:
- No direct relationships (time-series data)

---

### 3.4 Transaction_Summary Table

**Purpose**: Aggregated transaction metrics for analytics dashboard

```sql
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
```

**Fields**:
- `id`: UUID primary key
- `chain`: Blockchain
- `date`: Aggregation date
- `transaction_count`: Total transactions
- `total_volume_usd`: Total value transferred
- `unique_senders`: Count of unique sender addresses
- `unique_receivers`: Count of unique receiver addresses
- `avg_transaction_value_usd`: Average transaction size
- `swap_count`: Number of swaps
- `total_fees_usd`: Total fees paid

**Aggregation Job**:
- Runs daily at midnight UTC
- Aggregates previous day's transactions from core_db
- Calculates USD values using historical exchange rates
- Immutable once created (daily snapshot)

**Relationships**:
- No direct relationships (aggregated metrics)

---

## 4. Audit Database (audit_db)

**Purpose**: Immutable audit trail for compliance and security

**Characteristics**:
- Append-only (no updates or deletes)
- Separate database for tamper evidence
- Automated backup to immutable storage (S3 Glacier, etc.)
- 7+ year retention for regulatory compliance

### 4.1 Audit_Logs Table

**Purpose**: Complete audit trail of all critical operations

```sql
CREATE TYPE audit_action AS ENUM (
    'user_register', 'user_login', 'user_logout',
    'wallet_create', 'wallet_archive',
    'transaction_create', 'transaction_confirm', 'transaction_fail',
    'exchange_quote', 'exchange_execute', 'exchange_complete',
    'kyc_submit', 'kyc_document_upload', 'kyc_approve', 'kyc_reject',
    'profile_update', 'password_change', '2fa_enable', '2fa_disable',
    'withdrawal_whitelist_add', 'withdrawal_whitelist_remove'
);

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID, -- Logical FK to core_db.users
    action audit_action NOT NULL,
    resource_type VARCHAR(50),
    resource_id UUID,
    details JSONB NOT NULL DEFAULT '{}',
    ip_address INET,
    user_agent TEXT,
    session_id VARCHAR(255),
    result VARCHAR(20) NOT NULL DEFAULT 'success',
    error_message TEXT,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp DESC);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_logs_ip_address ON audit_logs(ip_address);
```

**Fields**:
- `id`: UUID primary key
- `user_id`: Logical reference to user
- `action`: Type of action performed
- `resource_type`: Entity type affected (wallet, transaction, etc.)
- `resource_id`: Entity ID affected
- `details`: JSON with action-specific data
- `ip_address`: Source IP address
- `user_agent`: Browser/client info
- `session_id`: Session identifier
- `result`: success/failure
- `error_message`: Failure details
- `timestamp`: Event timestamp (immutable)

**Details Examples**:
```json
// Wallet Creation
{
  "chain": "BTC",
  "address": "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
  "label": "Main BTC Wallet"
}

// Transaction Create
{
  "chain": "ETH",
  "amount": "1.5",
  "fee": "0.002",
  "to_address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
  "tx_hash": "0xabc..."
}

// KYC Submit
{
  "verification_level": "basic",
  "document_count": 2,
  "document_types": ["passport", "selfie"]
}
```

**Compliance Queries**:
```sql
-- All actions by specific user
SELECT * FROM audit_logs WHERE user_id = '...' ORDER BY timestamp DESC;

-- Wallet creation activity
SELECT * FROM audit_logs WHERE action = 'wallet_create' AND timestamp > NOW() - INTERVAL '30 days';

-- Failed login attempts
SELECT * FROM audit_logs WHERE action = 'user_login' AND result = 'failure' GROUP BY ip_address;
```

**Relationships**:
- Logical relationship to core_db.Users (via user_id, no FK)

---

### 4.2 Security_Logs Table

**Purpose**: Security events and anomaly detection

```sql
CREATE TYPE security_event AS ENUM (
    'login_success', 'login_failure', 'login_suspicious',
    'password_reset_request', 'password_reset_complete',
    '2fa_enabled', '2fa_disabled', '2fa_failure',
    'session_created', 'session_expired', 'session_revoked',
    'ip_blacklist_trigger', 'rate_limit_exceeded',
    'unusual_location', 'device_change',
    'withdrawal_whitelist_violation',
    'api_key_created', 'api_key_revoked'
);

CREATE TYPE security_severity AS ENUM ('info', 'warning', 'high', 'critical');

CREATE TABLE security_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID, -- Logical FK to core_db.users (nullable for system events)
    event_type security_event NOT NULL,
    severity security_severity NOT NULL,
    ip_address INET,
    user_agent TEXT,
    location_country VARCHAR(2),
    location_city VARCHAR(100),
    details JSONB NOT NULL DEFAULT '{}',
    is_resolved BOOLEAN NOT NULL DEFAULT FALSE,
    resolved_at TIMESTAMP WITH TIME ZONE,
    resolved_by UUID,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_security_logs_user_id ON security_logs(user_id);
CREATE INDEX idx_security_logs_event_type ON security_logs(event_type);
CREATE INDEX idx_security_logs_severity ON security_logs(severity);
CREATE INDEX idx_security_logs_timestamp ON security_logs(timestamp DESC);
CREATE INDEX idx_security_logs_ip_address ON security_logs(ip_address);
CREATE INDEX idx_security_logs_unresolved ON security_logs(is_resolved) WHERE is_resolved = FALSE;
```

**Fields**:
- `id`: UUID primary key
- `user_id`: Logical reference to user (nullable)
- `event_type`: Type of security event
- `severity`: Event severity level
- `ip_address`: Source IP
- `user_agent`: Client information
- `location_country`, `location_city`: GeoIP data
- `details`: JSON with event-specific data
- `is_resolved`: Flag for security team follow-up
- `resolved_at`, `resolved_by`: Resolution tracking
- `timestamp`: Event timestamp

**Alerting Rules**:
```
critical: Immediate notification to security team
high: Alert within 15 minutes
warning: Daily digest
info: Log only
```

**Anomaly Detection Examples**:
- Multiple failed login attempts (5+ in 10 minutes)
- Login from new country
- Large withdrawal after password change
- API rate limit exceeded
- Suspicious timing patterns (bot-like behavior)

**Details Examples**:
```json
// Login Failure
{
  "reason": "invalid_password",
  "attempt_count": 3,
  "previous_success": "2025-10-13T08:30:00Z"
}

// Unusual Location
{
  "previous_country": "US",
  "current_country": "RU",
  "time_since_last_login_hours": 2,
  "impossible_travel": true
}

// Rate Limit Exceeded
{
  "endpoint": "/api/v1/transactions",
  "request_count": 150,
  "limit": 100,
  "window_minutes": 60
}
```

**Relationships**:
- Logical relationship to core_db.Users (via user_id, no FK)

---

### 4.3 API_Audit Table

**Purpose**: API request/response logging for compliance and debugging

```sql
CREATE TABLE api_audit (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID, -- Logical FK to core_db.users (nullable for public endpoints)
    endpoint VARCHAR(255) NOT NULL,
    method VARCHAR(10) NOT NULL,
    status_code INTEGER NOT NULL,
    request_body_hash VARCHAR(64),
    response_body_hash VARCHAR(64),
    request_size_bytes INTEGER,
    response_size_bytes INTEGER,
    duration_ms INTEGER NOT NULL,
    ip_address INET,
    user_agent TEXT,
    api_key_id UUID,
    rate_limit_remaining INTEGER,
    error_code VARCHAR(50),
    error_message TEXT,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_api_audit_user_id ON api_audit(user_id);
CREATE INDEX idx_api_audit_endpoint ON api_audit(endpoint);
CREATE INDEX idx_api_audit_status_code ON api_audit(status_code);
CREATE INDEX idx_api_audit_timestamp ON api_audit(timestamp DESC);
CREATE INDEX idx_api_audit_ip_address ON api_audit(ip_address);
CREATE INDEX idx_api_audit_errors ON api_audit(error_code) WHERE error_code IS NOT NULL;
```

**Fields**:
- `id`: UUID primary key
- `user_id`: Logical reference to user
- `endpoint`: API endpoint path
- `method`: HTTP method (GET/POST/PUT/DELETE)
- `status_code`: HTTP response code
- `request_body_hash`: SHA-256 hash of request body
- `response_body_hash`: SHA-256 hash of response body
- `request_size_bytes`, `response_size_bytes`: Payload sizes
- `duration_ms`: Request processing time
- `ip_address`: Client IP
- `user_agent`: Client information
- `api_key_id`: API key used (if any)
- `rate_limit_remaining`: Rate limit quota remaining
- `error_code`, `error_message`: Error details
- `timestamp`: Request timestamp

**Storage Strategy**:
- Full request/response bodies NOT stored (privacy/storage)
- Only hashes stored for integrity verification
- Actual bodies logged to separate short-term storage (7 days) for debugging
- Sensitive endpoints (auth, KYC) have enhanced logging

**Compliance Use Cases**:
- Prove API was called at specific time
- Verify integrity of request/response
- Track API usage patterns
- Audit regulatory report generation
- Debug production issues

**Performance Monitoring**:
```sql
-- Slow endpoint analysis
SELECT endpoint, AVG(duration_ms), MAX(duration_ms), COUNT(*)
FROM api_audit
WHERE timestamp > NOW() - INTERVAL '24 hours'
GROUP BY endpoint
ORDER BY AVG(duration_ms) DESC;

-- Error rate by endpoint
SELECT endpoint,
       COUNT(*) FILTER (WHERE status_code >= 500) * 100.0 / COUNT(*) as error_rate
FROM api_audit
WHERE timestamp > NOW() - INTERVAL '1 hour'
GROUP BY endpoint
HAVING COUNT(*) > 10;
```

**Data Retention**:
- Keep detailed logs for 90 days
- Aggregate to daily summaries after 90 days
- Retain summaries for 7 years

**Relationships**:
- Logical relationship to core_db.Users (via user_id, no FK)

---

## 5. Cross-Database Relationships

**Design Pattern**: Logical relationships without foreign key constraints for database isolation.

**Implementation Strategy**:
```go
// Application layer enforces referential integrity
type User struct {
    ID        uuid.UUID
    Email     string
    KYCProfile *KYCProfile // Loaded separately
    RiskScore  *UserRiskScore // Loaded separately
}

// Repository pattern handles cross-database queries
func (r *UserRepository) GetUserWithKYC(ctx context.Context, userID uuid.UUID) (*User, error) {
    user, err := r.coreDB.GetUser(ctx, userID)
    if err != nil {
        return nil, err
    }

    kycProfile, err := r.kycDB.GetKYCProfile(ctx, userID)
    if err != nil && !errors.Is(err, sql.ErrNoRows) {
        return nil, err
    }
    user.KYCProfile = kycProfile

    return user, nil
}
```

**Consistency Guarantees**:
- Application-level transactions for multi-database operations
- Idempotency keys for retry safety
- Compensating transactions for rollback scenarios
- Event sourcing for audit trail

**Query Patterns**:
```go
// Good: Single database query
users := userRepo.ListUsers(ctx, limit, offset)

// Good: Lazy loading across databases
for _, user := range users {
    kyc := kycRepo.GetKYCProfile(ctx, user.ID)
    risk := riskRepo.GetRiskScore(ctx, user.ID)
}

// Avoid: Cross-database JOINs (not possible without foreign keys)
// Use application-level joins instead
```

---

## 6. Data Migration Strategy

**Phase 1: Initial Schema Creation**
```bash
# Create databases
psql -c "CREATE DATABASE core_db;"
psql -c "CREATE DATABASE kyc_db;"
psql -c "CREATE DATABASE rates_db;"
psql -c "CREATE DATABASE audit_db;"

# Run migrations
golang-migrate -path db/migrations/core_db -database "postgres://..." up
golang-migrate -path db/migrations/kyc_db -database "postgres://..." up
golang-migrate -path db/migrations/rates_db -database "postgres://..." up
golang-migrate -path db/migrations/audit_db -database "postgres://..." up
```

**Phase 2: Seed Initial Data**
```sql
-- Seed chains and tokens
\i db/seed/core_seed.sql

-- Seed alert rules
\i db/seed/kyc_seed.sql

-- Seed initial price data
\i db/seed/rates_seed.sql
```

**Phase 3: Ongoing Migrations**
```bash
# Create new migration
make migration name="add_withdrawal_whitelist"

# Apply migrations
make migrate-up

# Rollback migration
make migrate-down
```

---

## 7. Indexes and Performance Optimization

**Index Strategy**:
- Primary keys automatically indexed (UUID)
- Foreign keys indexed for join performance
- Composite indexes for common query patterns
- Partial indexes for filtered queries
- GIN indexes for JSONB columns

**Query Performance Targets**:
- Simple lookups (by PK): <5ms
- Filtered queries: <50ms
- Aggregations: <200ms
- Complex analytics: <2s

**Optimization Techniques**:
```sql
-- Composite index for transaction history query
CREATE INDEX idx_tx_user_chain_status_created
ON transactions(wallet_id, chain, status, created_at DESC);

-- Partial index for pending transactions
CREATE INDEX idx_tx_pending
ON transactions(status, created_at)
WHERE status = 'pending';

-- GIN index for JSONB search
CREATE INDEX idx_tx_metadata_gin
ON transactions USING GIN(metadata);
```

**Table Partitioning**:
```sql
-- Partition transactions by month for performance
CREATE TABLE transactions PARTITION BY RANGE (created_at);

CREATE TABLE transactions_2025_10 PARTITION OF transactions
FOR VALUES FROM ('2025-10-01') TO ('2025-11-01');
```

---

## 8. Data Validation and Constraints

**Database-Level Constraints**:
- NOT NULL for required fields
- UNIQUE constraints for natural keys
- CHECK constraints for value ranges
- FOREIGN KEY constraints (within same database)
- ENUM types for controlled vocabularies

**Application-Level Validation**:
- Email format validation (RFC 5322)
- Address format validation (blockchain-specific)
- Amount range validation (min/max)
- Rate limiting and throttling
- Business rule enforcement

**Example Validation Flow**:
```go
// Domain layer validation
func (w *Wallet) Validate() error {
    if w.Balance < 0 {
        return errors.New("balance cannot be negative")
    }
    if !isValidAddress(w.Address, w.Chain) {
        return errors.New("invalid address format")
    }
    return nil
}

// Database layer validation (PostgreSQL)
ALTER TABLE wallets ADD CONSTRAINT balance_non_negative
CHECK (balance >= 0);
```

---

## Summary

This data model provides a complete, normalized database schema for the multi-chain cryptocurrency wallet platform with:

✅ **4 Separate Databases**: Clear separation of operational, compliance, market, and audit data
✅ **60+ Tables**: Comprehensive coverage of all functional requirements
✅ **UUID Primary Keys**: Distributed system compatibility
✅ **Proper Indexing**: Optimized for query performance
✅ **State Management**: Clear state transitions for all stateful entities
✅ **Audit Trail**: Complete immutable logging for compliance
✅ **Security**: Encrypted PII, isolated KYC data, tamper-evident audit logs
✅ **Scalability**: Partitioning strategies for high-volume tables
✅ **Integrity**: Constraints and validation at database and application layers

**Next Steps**:
1. ✅ Review and validate data model
2. → Generate OpenAPI 3.0 specification (contracts/openapi.yaml)
3. → Document WebSocket protocol (contracts/websocket.md)
4. → Document blockchain adapters (contracts/blockchain.md)
5. → Create quickstart guide (quickstart.md)
