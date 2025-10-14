
# 🧱 Full SQL DDL — Crypto Wallet & Fintech Platform (4 DB Version)

This document contains **production-grade PostgreSQL DDL** for the simplified architecture of **4 databases**:
`core_db`, `kyc_db`, `rates_db`, and `audit_db`.  
It includes enums, constraints, indexes, and a reusable trigger for `updated_at`.

## How to Use
1. Create each database (optional — or use existing ones).
2. Connect to each database and run its section.
3. Ensure `pgcrypto` and `citext` extensions are available where used.

## Highlights
- **Enums** for constrained fields (chains, statuses, account types)
- **Strong typing** (`NUMERIC(36,18)` for crypto amounts)
- **Cross-db references** are **application-level** (store UUIDs; enforce in app layer)
- **Audit & API logs** segregated in `audit_db`
- **Time-series** tables (`price_history`, `exchange_rates`) indexed for recency queries
- **Trigger** `set_updated_at()` keeps `updated_at` fresh

## File
Use the SQL file below:

**👉 `/mnt/data/crypto_wallet_4db_full_schema.sql`**
