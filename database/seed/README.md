# Database Seed Scripts

This directory contains idempotent seed scripts that populate the core project
databases with deterministic sample data. The scripts are safe to run multiple
times and are intended to provide meaningful fixtures for local development,
manual testing, and automated integration checks.

## Available Seeds

| Database | Script | Description |
|----------|--------|-------------|
| `core_db`  | `core_db_seed.sql`  | Demo users, wallets, transactions, ledger entries, and reference data |
| `kyc_db`   | `kyc_db_seed.sql`   | Example KYC profiles, documents, and risk scores |
| `rates_db` | `rates_db_seed.sql` | Baseline market data, trading pairs, and price history |
| `audit_db` | `audit_db_seed.sql` | Representative audit trails and security events |

## Usage

1. Ensure the databases have been created and migrations applied:

   ```bash
   make migrate-up
   ```

2. Export DSNs that point at each logical database (the same values used by the
   application). The script understands either `postgres://` or `postgresql://`
   URIs:

   ```bash
   export CORE_DB_DSN="postgresql://user:pass@localhost:5432/core_db?sslmode=disable"
   export KYC_DB_DSN="postgresql://user:pass@localhost:5433/kyc_db?sslmode=disable"
   export RATES_DB_DSN="postgresql://user:pass@localhost:5434/rates_db?sslmode=disable"
   export AUDIT_DB_DSN="postgresql://user:pass@localhost:5435/audit_db?sslmode=disable"
   ```

3. Run the helper script to seed every database in a single step:

   ```bash
   ./database/seed/run-seeds.sh
   ```

   You can also execute an individual seed by piping it into `psql`, for
   example:

   ```bash
   psql "$CORE_DB_DSN" -f database/seed/core_db_seed.sql
   ```

## Notes

- The scripts use `ON CONFLICT DO NOTHING` guards on natural keys so repeated
  executions do not generate duplicate data.
- Seeded timestamps are UTC and deterministic which simplifies assertions in
  tests.
- The sample data mirrors the schema defined in `data-model.md` and is suitable
  for exercising the end-to-end flows covered by the user stories.
