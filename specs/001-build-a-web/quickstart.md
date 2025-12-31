# Quickstart Guide: Multi-Chain Crypto Wallet Development

**Version**: 1.0.0
**Date**: 2025-10-14
**Estimated Setup Time**: 15 minutes

---

## Prerequisites

Ensure you have the following installed:

- **Go** 1.23+ ([https://golang.org/dl/](https://golang.org/dl/))
- **Node.js** 20+ and npm ([https://nodejs.org/](https://nodejs.org/))
- **Docker** and Docker Compose ([https://docs.docker.com/get-docker/](https://docs.docker.com/get-docker/))
- **PostgreSQL** client tools (psql) for manual DB access
- **Make** utility (usually pre-installed on macOS/Linux)
- **Git** for version control

### Verify Prerequisites
```bash
go version     # Should show go1.23 or higher
node --version # Should show v20 or higher
docker --version
docker-compose --version
make --version
psql --version
```

---

## Quick Start (5 minutes)

### Option 1: Docker Compose (Recommended)

```bash
# 1. Clone repository
cd crypto-wallet

# 2. Start all services
docker-compose up -d

# 3. Wait for services to initialize (~30 seconds)
docker-compose logs -f

# 4. Access applications
# Frontend: http://localhost:3000
# Backend API: http://localhost:8080
# pgAdmin: http://localhost:5050
# Grafana: http://localhost:3001
```

### Option 2: Local Development

For development with hot-reload and debugging:

```bash
# Terminal 1: Backend
cd crypto-wallet-backend
make setup    # Install deps, create DBs, run migrations
./database/seed/run-seeds.sh  # Populate the core, KYC, rates, and audit databases
make run      # Start backend on :8080

# Terminal 2: Frontend
cd crypto-wallet-frontend
npm install
npm run dev   # Start frontend on :3000

# Terminal 3: Redis (required for WebSocket)
docker run -p 6379:6379 redis:alpine
```

### 3. Verify the Installation

Execute the integration helper script from the repository root. It runs the
Go unit tests, TypeScript lint checks, and the frontend Jest suite:

```bash
./scripts/run-integration.sh
```

---

## Backend Setup (5 minutes)

### 1. Environment Configuration

```bash
cd crypto-wallet-backend

# Copy environment template
cp .env.example .env
```

### 2. Edit `.env` File

```bash
# Database Configuration (4 databases)
CORE_DB_DSN=postgresql://user:pass@localhost:5432/core_db?sslmode=disable
KYC_DB_DSN=postgresql://user:pass@localhost:5433/kyc_db?sslmode=disable
RATES_DB_DSN=postgresql://user:pass@localhost:5434/rates_db?sslmode=disable
AUDIT_DB_DSN=postgresql://user:pass@localhost:5435/audit_db?sslmode=disable

# Redis Configuration
REDIS_URL=redis://localhost:6379

# JWT Secret (generate with: openssl rand -base64 32)
JWT_SECRET=your-secret-key-here

# Blockchain RPC Endpoints
BTC_RPC_URL=https://bitcoin-rpc.example.com
BTC_RPC_USER=user
BTC_RPC_PASSWORD=pass

ETH_RPC_URL=https://ethereum-rpc.example.com

SOL_RPC_URL=https://solana-rpc.example.com

XLM_HORIZON_URL=https://horizon.stellar.org

# External APIs
COINGECKO_API_KEY=your-coingecko-api-key
KYC_PROVIDER_API_KEY=your-kyc-provider-key

# Server Configuration
SERVER_PORT=8080
SERVER_HOST=localhost
ENVIRONMENT=development
```

### 3. Initialize Project

```bash
# Install Go dependencies
go mod download

# Create databases
make db-create

# Run migrations
make migrate-up

# Seed initial data
make db-seed

# Verify setup
make test
```

### 4. Start Backend Server

```bash
make run

# Or for development with hot-reload
make dev

# Verify health
curl http://localhost:8080/health
```

**Expected Output**:
```json
{
  "status": "healthy",
  "timestamp": "2025-10-14T10:30:00Z",
  "services": {
    "database": "healthy",
    "redis": "healthy",
    "blockchain_btc": "healthy",
    "blockchain_eth": "healthy"
  }
}
```

---

## Frontend Setup (3 minutes)

### 1. Environment Configuration

```bash
cd crypto-wallet-frontend

# Copy environment template
cp .env.local.example .env.local
```

### 2. Edit `.env.local` File

```bash
# API Configuration
NEXT_PUBLIC_API_URL=http://localhost:8080/v1
NEXT_PUBLIC_WS_URL=ws://localhost:8080/ws

# Feature Flags
NEXT_PUBLIC_ENABLE_2FA=true
NEXT_PUBLIC_ENABLE_KYC=true

# Analytics (optional)
NEXT_PUBLIC_GA_ID=G-XXXXXXXXXX
```

### 3. Install Dependencies and Start

```bash
# Install dependencies
npm install

# Start development server
npm run dev

# Open browser
# http://localhost:3000
```

### 4. Build for Production

```bash
npm run build
npm start # Runs production build on port 3000
```

---

## Database Management

### Create New Migration

```bash
cd crypto-wallet-backend

# Create migration file
make migration name=add_withdrawal_whitelist

# Edit generated files in db/migrations/core_db/
# - 00X_add_withdrawal_whitelist.up.sql
# - 00X_add_withdrawal_whitelist.down.sql

# Apply migration
make migrate-up

# Rollback if needed
make migrate-down
```

### Database Access

```bash
# Connect to databases
make db-console-core    # Core database
make db-console-kyc     # KYC database
make db-console-rates   # Rates database
make db-console-audit   # Audit database

# Or use Docker
docker-compose exec postgres-core psql -U user -d core_db
```

### Reset Databases

```bash
# ⚠️ WARNING: This will DELETE all data
make db-reset

# Recreate with seed data
make db-seed
```

---

## Common Development Tasks

### Code Quality

```bash
# Backend
cd crypto-wallet-backend
make lint      # Run golangci-lint
make fmt       # Format code with gofmt
make test      # Run all tests
make test-unit # Run unit tests only
make test-integration # Run integration tests

# Frontend
cd crypto-wallet-frontend
npm run lint   # ESLint
npm run format # Prettier
npm test       # Jest unit tests
npm run test:e2e # Playwright E2E tests
```

### Running Tests

```bash
# Backend tests
make test                    # All tests
make test-coverage           # With coverage report
make test-integration        # Integration tests (requires Docker)

# Frontend tests
npm test                     # Unit tests (Jest)
npm run test:watch          # Watch mode
npm run test:e2e            # E2E tests (Playwright)
npm run test:e2e:headed     # E2E with browser visible
```

### View Logs

```bash
# Docker Compose logs
docker-compose logs -f [service]

# Examples
docker-compose logs -f backend
docker-compose logs -f frontend
docker-compose logs -f postgres-core

# Local development
# Backend logs to stdout
# Frontend logs to terminal
```

### Database Utilities

```bash
# Backup database
make db-backup db=core_db

# Restore database
make db-restore db=core_db file=backup.sql

# Export schema
make db-schema db=core_db > schema.sql

# View migration status
make migrate-status
```

---

## API Testing

### Using cURL

```bash
# Register user
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"SecurePass123!"}'

# Login
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -c cookies.txt \
  -d '{"email":"user@example.com","password":"SecurePass123!"}'

# Get user profile (authenticated)
curl -X GET http://localhost:8080/v1/auth/me \
  -b cookies.txt

# Create wallet
curl -X POST http://localhost:8080/v1/wallets \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{"chain":"BTC","label":"Main BTC Wallet"}'

# Get wallet balance
curl -X GET http://localhost:8080/v1/wallets/{wallet-id}/balance \
  -b cookies.txt
```

### Using Postman/Insomnia

1. Import OpenAPI spec: `specs/001-build-a-web/contracts/openapi.yaml`
2. Set base URL: `http://localhost:8080/v1`
3. Authentication: Cookie-based (automatic after login)

---

## Troubleshooting

### Backend Won't Start

**Problem**: `connection refused` errors

**Solution**:
```bash
# Check if databases are running
docker-compose ps

# Check database connectivity
make db-ping

# Verify environment variables
cat .env | grep DB_DSN

# Check port availability
lsof -i :8080
```

### Frontend Build Errors

**Problem**: Module not found errors

**Solution**:
```bash
# Clear cache and reinstall
rm -rf node_modules package-lock.json
npm install

# Clear Next.js cache
rm -rf .next

# Rebuild
npm run build
```

### Database Migration Fails

**Problem**: Migration error or stuck

**Solution**:
```bash
# Check current version
make migrate-status

# Force version (if necessary)
make migrate-force version=X

# Reset and reapply
make db-reset
make migrate-up
```

### WebSocket Connection Issues

**Problem**: WebSocket connection refused

**Solution**:
```bash
# Verify Redis is running
docker-compose ps redis

# Check Redis connection
redis-cli ping

# Verify WebSocket endpoint
curl -i -N -H "Connection: Upgrade" \
  -H "Upgrade: websocket" \
  -H "Sec-WebSocket-Version: 13" \
  -H "Sec-WebSocket-Key: test" \
  http://localhost:8080/ws
```

### Port Already in Use

**Problem**: `bind: address already in use`

**Solution**:
```bash
# Find process using port
lsof -i :8080  # Backend
lsof -i :3000  # Frontend

# Kill process
kill -9 <PID>

# Or change port in .env/.env.local
```

---

## Development Workflow

### Daily Development Cycle

```bash
# 1. Pull latest changes
git pull origin main

# 2. Update dependencies
cd crypto-wallet-backend && go mod tidy
cd crypto-wallet-frontend && npm install

# 3. Run migrations
make migrate-up

# 4. Start services
docker-compose up -d redis postgres-*
make dev  # Backend
npm run dev  # Frontend

# 5. Run tests before commit
make test
npm test

# 6. Commit changes
git add .
git commit -m "feat: add withdrawal whitelist"
git push
```

### Feature Development Workflow

1. **Create feature branch**
   ```bash
   git checkout -b feature/withdrawal-whitelist
   ```

2. **Create migration (if needed)**
   ```bash
   make migration name=add_withdrawal_whitelist
   ```

3. **Implement backend changes**
   - Domain entities in `internal/domain/`
   - Use cases in `internal/application/usecases/`
   - Handlers in `internal/interfaces/http/handlers/`

4. **Implement frontend changes**
   - Components in `src/components/`
   - Pages in `src/app/`
   - API integration in `src/lib/api.ts`

5. **Write tests**
   ```bash
   # Backend
   make test

   # Frontend
   npm test
   npm run test:e2e
   ```

6. **Create pull request**
   ```bash
   git push origin feature/withdrawal-whitelist
   # Create PR on GitHub
   ```

---

## Production Deployment

### Build for Production

```bash
# Backend
cd crypto-wallet-backend
make build
# Binary created at: bin/server

# Frontend
cd crypto-wallet-frontend
npm run build
# Static files in: .next/
```

### Docker Production Build

```bash
# Build images
docker-compose -f docker-compose.prod.yml build

# Push to registry
docker-compose -f docker-compose.prod.yml push

# Deploy
docker-compose -f docker-compose.prod.yml up -d

# Alternatively, use the helper script which wraps these commands:
cd deploy
./deploy.sh
```

### Environment Variables for Production

**Critical settings to change for production**:
```bash
# Security
JWT_SECRET=<strong-random-secret>  # Generate new secret
ENVIRONMENT=production

# Databases
# Use managed database services (AWS RDS, etc.)
CORE_DB_DSN=postgresql://user:pass@prod-db:5432/core_db?sslmode=require

# TLS/SSL
SERVER_TLS_ENABLED=true
SERVER_TLS_CERT=/path/to/cert.pem
SERVER_TLS_KEY=/path/to/key.pem

# Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m

# Monitoring
SENTRY_DSN=<sentry-project-dsn>
PROMETHEUS_ENABLED=true
```

---

## Useful Make Commands

```bash
# Backend Development
make help              # Show all available commands
make setup             # Complete project setup
make run               # Start server
make dev               # Start with hot-reload
make test              # Run all tests
make test-coverage     # Generate coverage report
make lint              # Run linters
make fmt               # Format code
make build             # Build production binary
make clean             # Clean build artifacts

# Database Management
make db-create         # Create all databases
make db-drop           # Drop all databases
make db-reset          # Drop, create, migrate, seed
make db-seed           # Seed initial data
make migrate-up        # Apply migrations
make migrate-down      # Rollback last migration
make migrate-status    # Show migration status
make migration name=X  # Create new migration

# Docker Management
make docker-build      # Build Docker images
make docker-up         # Start Docker services
make docker-down       # Stop Docker services
make docker-logs       # View Docker logs
```

---

## Next Steps

1. **Review Architecture**: Read `specs/001-build-a-web/plan.md`
2. **Understand Data Model**: Study `specs/001-build-a-web/data-model.md`
3. **Explore API**: Review `specs/001-build-a-web/contracts/openapi.yaml`
4. **Start Implementing**: Follow task breakdown in `specs/001-build-a-web/tasks.md` (when generated)

---

## Support

- **Documentation**: [Project Wiki](https://github.com/project/wiki)
- **API Docs**: [OpenAPI Spec](./contracts/openapi.yaml)
- **Issues**: [GitHub Issues](https://github.com/project/issues)
- **Discussions**: [GitHub Discussions](https://github.com/project/discussions)

---

## Security Notes

⚠️ **Development vs Production**:
- Development uses HTTP, production must use HTTPS
- Development JWT secrets are weak, production requires strong secrets
- Development databases have weak passwords, production must use strong passwords
- Never commit `.env` files with real credentials to version control
- Use secrets management services (AWS Secrets Manager, HashiCorp Vault) in production

---

**Quickstart Complete!** 🎉

You now have a fully functional development environment for the Multi-Chain Crypto Wallet platform.
