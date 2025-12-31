# Atlas Wallet Platform

Atlas Wallet is a full-stack multi-chain cryptocurrency wallet that supports
portfolio analytics, KYC workflows, secure authentication, and rich operational
tooling. The backend is written in Go with a PostgreSQL/Redis foundation while
the frontend is a TypeScript/Next.js application.

## Features

- Multi-chain wallet creation (BTC, ETH, SOL, XLM) with encrypted key storage
- Transaction execution with audit logging, rate limiting, and retry support
- Portfolio analytics dashboard featuring performance charts and allocation
- KYC profile/document management backed by dedicated compliance storage
- Production-ready deployment scripts, database seed fixtures, and observability

## Getting Started From Zero

Follow these steps the first time you clone the repository.

1. **Install prerequisites**

   - Go 1.25+
   - Node.js 20+
   - Docker & Docker Compose (optional for deployment)
   - PostgreSQL 16+

2. **Bootstrap the project**

   ```bash
   make setup           # installs deps, creates databases
   make migrate-up      # applies all migrations across logical databases
   ./database/seed/run-seeds.sh
   ```

3. **Run the stack**

   ```bash
   make run-backend     # starts the Fiber HTTP API on :8080
   npm --prefix crypto-wallet-frontend run dev  # starts the Next.js frontend
   ```

4. **Visit the app**: <http://localhost:3000>

5. **Verify everything works**

   ```bash
   ./scripts/run-integration.sh
   ```

   This runs Go unit tests, frontend linting, and the Jest suite in one go.

## Database Seeding

Idempotent seed scripts populate the four logical databases with sample users,
wallets, transactions, compliance records, audit trails, and market data. Set
the DSNs used by the backend and run:

```bash
export CORE_DB_DSN="postgresql://user:pass@localhost:5432/core_db?sslmode=disable"
export KYC_DB_DSN="postgresql://user:pass@localhost:5432/kyc_db?sslmode=disable"
export RATES_DB_DSN="postgresql://user:pass@localhost:5432/rates_db?sslmode=disable"
export AUDIT_DB_DSN="postgresql://user:pass@localhost:5432/audit_db?sslmode=disable"
./database/seed/run-seeds.sh
```

## Production Deployment

A production-oriented Docker Compose harness is available under `deploy/`:

```bash
cd deploy
./deploy.sh           # build images and start containers
./deploy.sh logs      # tail service logs
./deploy.sh down      # stop the stack
```

Environment templates live in `deploy/env/`. Update the secrets before
deploying to a real environment.

## Testing & QA

- **Backend**: `make test`
- **Frontend**: `npm --prefix crypto-wallet-frontend run lint && npm --prefix crypto-wallet-frontend test`
- **Integration**: `./scripts/run-integration.sh`

## Additional Documentation

- `specs/` – implementation plans, data models, API contracts
- `PHASE_0_RESEARCH_SYNTHESIS.md` – research and design rationale
- `tech_stack.md` – runtime and tooling overview
