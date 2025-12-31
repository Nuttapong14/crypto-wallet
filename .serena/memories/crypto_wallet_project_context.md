# Multi-Chain Crypto Wallet Project Context

## Project Overview
Production-grade cryptocurrency wallet and exchange platform supporting Bitcoin (BTC), Ethereum (ETH), Solana (SOL), and Stellar (XLM) with real-time price monitoring, KYC compliance, and secure transaction management.

## Technology Stack

### Backend
- **Language**: Go 1.23+
- **Web Framework**: Fiber v2
- **Architecture**: Domain-Driven Design (DDD) + Hexagonal Architecture
- **Authentication**: JWT with HTTP-only cookies
- **Blockchain SDKs**:
  - Bitcoin: `github.com/btcsuite/btcd`, `github.com/btcsuite/btcutil`
  - Ethereum: `github.com/ethereum/go-ethereum`
  - Solana: `github.com/portto/solana-go-sdk`
  - Stellar: `github.com/stellar/go`

### Frontend
- **Framework**: Next.js 15 with App Router
- **Language**: TypeScript
- **Styling**: TailwindCSS
- **UI Components**: ShadCN/UI
- **State Management**: React Context API
- **Real-time**: WebSocket client with exponential backoff

### Databases
- **Primary Database**: PostgreSQL 16
- **Architecture**: 4 separate databases for compliance isolation
  - `core_db`: Users, wallets, transactions, ledger
  - `kyc_db`: KYC profiles, documents, risk scores
  - `rates_db`: Exchange rates, price history, trading pairs
  - `audit_db`: Immutable audit logs, security events
- **Audit Logging**: pgAudit extension
- **Caching/PubSub**: Redis 7+ (WebSocket message distribution)

### External Services
- **Price Feeds**: 
  - Primary: CoinGecko WebSocket API
  - Backup: Binance REST API
- **KYC Provider**: SumSub (crypto-focused, AML included)
- **Blockchain RPCs**: External RPC nodes for BTC, ETH, SOL, XLM

### Security
- **Private Key Encryption**: AES-256-GCM
- **Password Hashing**: bcrypt
- **HD Wallets**: BIP-32/BIP-44 derivation paths
- **Wallet Generation**: Secure random number generation

### Infrastructure & DevOps
- **Containerization**: Docker, Docker Compose
- **Monitoring**: Prometheus + Grafana
- **Logging**: Structured JSON logging
- **CI/CD**: GitHub Actions (planned)
- **Deployment**: Production-ready Docker images

## Project Structure

### Backend (`crypto-wallet-backend/`)
```
internal/
  domain/           # Domain entities and business rules
  application/      # Use cases and business logic
    usecases/       # Application services
  infrastructure/   # External integrations
    blockchain/     # Blockchain adapters (BTC, ETH, SOL, XLM)
    database/       # Repository implementations
    cache/          # Redis implementations
    kyc/            # SumSub integration
    rates/          # Price feed integrations
  interfaces/       # API handlers and WebSocket
    http/           # REST API handlers
    websocket/      # WebSocket server
```

### Frontend (`crypto-wallet-frontend/`)
```
src/
  app/              # Next.js App Router pages
  components/       # React components (ShadCN/UI)
  lib/              # Utilities and API client
  hooks/            # Custom React hooks
  types/            # TypeScript type definitions
```

## Key Features
1. Multi-chain wallet management (BTC, ETH, SOL, XLM)
2. Real-time cryptocurrency price monitoring (5-second updates)
3. Exchange operations with instant quotes
4. KYC/AML compliance integration
5. Double-entry ledger for financial accuracy
6. WebSocket notifications for transactions and balances
7. Comprehensive audit logging
8. Mobile-responsive progressive web app

## API Specifications
- **REST API**: OpenAPI 3.0 specification (`specs/001-build-a-web/contracts/openapi.yaml`)
- **WebSocket Protocol**: Full specification (`specs/001-build-a-web/contracts/websocket.md`)
- **Blockchain Adapters**: Interface contracts (`specs/001-build-a-web/contracts/blockchain.md`)

## Data Model
Complete schema design with 60+ tables across 4 databases documented in `specs/001-build-a-web/data-model.md`

## Development Setup
See `specs/001-build-a-web/quickstart.md` for:
- Prerequisites (Go 1.23+, Node.js 20+, Docker, PostgreSQL client)
- Docker Compose quick start
- Backend and frontend setup
- Database migrations
- Testing strategies
- Troubleshooting guide

## Performance Requirements
- API response time: <200ms (p95)
- Balance query: <2s
- WebSocket latency: <100ms
- Price update frequency: max 1 per 5 seconds per symbol
- Support: 10,000 concurrent WebSocket connections

## Security Requirements
- No plaintext private key storage
- Separate KYC database with restricted access
- Immutable audit logs
- Rate limiting on all API endpoints
- HTTPS/WSS in production
- Input validation and SQL injection prevention

## Compliance
- Financial audit trail (pgAudit)
- KYC/AML verification (SumSub)
- Data isolation (4 separate databases)
- Privacy protection (GDPR considerations)

## Implementation Status
- ✅ Phase 0: Research completed
- ✅ Phase 1: Design & Contracts completed
- ⏳ Phase 2: Task Generation (pending)
- ⏳ Phase 3-6: Implementation phases (pending)
