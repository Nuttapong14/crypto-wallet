# Implementation Plan: Multi-Chain Crypto Wallet and Exchange Platform

**Branch**: `001-build-a-web` | **Date**: 2025-10-14 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-build-a-web/spec.md`

## Summary

This implementation plan covers the development of a full-stack multi-chain cryptocurrency wallet and exchange platform supporting Bitcoin (BTC), Ethereum (ETH), Solana (SOL), and Stellar (XLM). The system follows Domain-Driven Design (DDD) with Hexagonal Architecture principles, featuring a Go backend with 4 PostgreSQL databases and a Next.js 15 frontend with TypeScript.

**Primary Requirements**:

- Multi-chain wallet creation and management with encrypted private keys
- Send/receive cryptocurrency transactions with real-time status tracking
- Live cryptocurrency price feeds via WebSocket
- Off-chain cryptocurrency exchange functionality
- KYC verification with document upload and tiered access levels
- Transaction history, portfolio analytics, and notifications
- Comprehensive audit logging and security monitoring

**Technical Approach**:

- **Backend**: Golang 1.23+ with DDD + Hexagonal Architecture, separated into domain, application, infrastructure, and interface layers
- **Frontend**: Next.js 15 (App Router) with TypeScript, TailwindCSS, ShadCN/UI, and Framer Motion
- **Databases**: 4 PostgreSQL 16 databases (core_db, kyc_db, rates_db, audit_db) for separation of concerns
- **Blockchain**: Adapters for BTC (btcsuite), ETH (go-ethereum), SOL (solana-go-sdk), XLM (stellar/go)
- **Real-time**: WebSocket for live price updates via Redis Pub/Sub
- **Security**: AES-GCM encryption for private keys, bcrypt for passwords, JWT authentication
- **Deployment**: Docker Compose orchestration with monitoring via Prometheus + Grafana

## Technical Context

**Language/Version**:

- Backend: Go 1.23+
- Frontend: Node.js 20+, TypeScript 5.3+, Next.js 15 (App Router)

**Primary Dependencies**:

_Backend_:

- **Web Framework**: Fiber v2 (high-performance Go web framework)
- **Database**: PostgreSQL 16 with `pgx` driver
- **Blockchain SDKs**:
  - Bitcoin: `github.com/btcsuite/btcd`, `github.com/btcsuite/btcutil`
  - Ethereum: `github.com/ethereum/go-ethereum`
  - Solana: `github.com/portto/solana-go-sdk`
  - Stellar: `github.com/stellar/go`
- **Cryptography**: `golang.org/x/crypto` (bcrypt, AES-GCM), `github.com/kevinburke/nacl/secretbox`
- **Authentication**: `github.com/golang-jwt/jwt/v5`
- **Configuration**: `github.com/spf13/viper`
- **Migration**: `github.com/golang-migrate/migrate/v4`
- **Message Queue**: Redis for Pub/Sub
- **Testing**: `github.com/stretchr/testify`, `github.com/DATA-DOG/go-sqlmock`

_Frontend_:

- **Framework**: Next.js 15 with App Router
- **UI System**: TailwindCSS 3.4+, ShadCN/UI components
- **State Management**: Zustand 4.x
- **Forms**: React Hook Form + Zod validation
- **API Client**: Axios with interceptors
- **Charts**: Recharts for analytics
- **Animation**: Framer Motion
- **i18n**: next-intl (English/Thai)
- **Testing**: Playwright for E2E

**Storage**:

- 4 PostgreSQL 16 databases with distinct schemas:
  - `core_db`: Operational data (users, wallets, transactions, chains, tokens, ledger)
  - `kyc_db`: Compliance data (profiles, documents, risk scores, verification status)
  - `rates_db`: Market data (exchange rates, trading pairs, price history, analytics)
  - `audit_db`: Security logs (audit trails, API logs, security events)
- Redis for real-time price caching and Pub/Sub messaging

**Testing**:

- Backend: Go's `testing` package with `testify/assert`, `testify/mock` for unit tests; `testcontainers-go` for integration tests
- Frontend: Playwright for E2E testing, Jest + React Testing Library for component tests
- Contract testing for API endpoints using OpenAPI validation

**Target Platform**:

- Backend: Linux containers (Docker) deployable to cloud platforms (AWS, GCP, Azure) or on-premise
- Frontend: Web application (responsive design for desktop, tablet, mobile browsers)
- Minimum browser support: Chrome 90+, Firefox 88+, Safari 14+, Edge 90+

**Project Type**: Web application (full-stack: backend API + frontend SPA)

**Performance Goals**:

- API response time: <200ms for 95th percentile (excluding blockchain calls)
- Blockchain RPC latency: <2s for balance queries, <5s for transaction submission
- WebSocket price updates: <5 seconds latency from market changes
- Database queries: <50ms for 95th percentile
- Frontend initial load: <3s on 3G networks, <1s on WiFi
- Support 10,000 concurrent active users without degradation

**Constraints**:

- Private key encryption must be military-grade (AES-256-GCM minimum)
- Zero plaintext storage of sensitive data (keys, passwords, PII)
- All financial transactions must be atomically consistent
- Blockchain transaction immutability must be respected
- KYC data must be isolated with restricted access
- Audit logs must be immutable and tamper-evident
- Session timeout: 30 minutes of inactivity
- Password requirements: minimum 12 characters, complexity rules enforced
- Two-factor authentication strongly recommended for high-value accounts
- GDPR and financial regulation compliance required

**Scale/Scope**:

- Expected users: 10,000-50,000 registered users at launch
- Transaction volume: 1,000-5,000 transactions per day initially
- Database size: ~100GB over first year (transaction history + audit logs)
- API endpoints: ~40 REST endpoints across 8 resource domains
- Frontend pages: 7 major pages with 30+ reusable components
- Blockchain support: 4 chains (BTC, ETH, SOL, XLM) with potential for future expansion
- Supported cryptocurrencies: 4 native currencies initially, ERC-20 tokens in future iterations

## Constitution Check

_GATE: Must pass before Phase 0 research. Re-check after Phase 1 design._

**Current Project Status**: This is a new greenfield project with no existing constitution defined. The default constitution template contains only placeholders.

**Constitution Actions Required**:

1. ✅ **No gates to check**: Constitution file contains only placeholder content with no defined principles or constraints
2. ⚠️ **Recommendation**: Consider establishing project constitution principles before Phase 2 implementation, particularly around:
   - Testing requirements (TDD approach, coverage thresholds)
   - Security standards (encryption, audit logging, access control)
   - API design principles (REST conventions, versioning strategy)
   - Database transaction patterns (isolation levels, consistency guarantees)
   - Error handling and observability standards

**Gate Status**: **PASS** (no gates defined to block progress)

**Post-Design Re-check**: Will verify any constitutional principles established during Phase 0-1 research and design phases.

## Project Structure

### Documentation (this feature)

```
specs/001-build-a-web/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
│   ├── openapi.yaml     # OpenAPI 3.0 specification for all REST endpoints
│   ├── websocket.md     # WebSocket protocol documentation
│   └── blockchain.md    # Blockchain adapter interface contracts
├── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
└── checklists/          # Quality validation checklists
    └── requirements.md  # Specification quality checklist (already created)
```

### Source Code (repository root)

```
crypto-wallet/
├── backend/                          # Go backend service
│   ├── cmd/
│   │   └── server/
│   │       └── main.go              # Application entry point
│   ├── internal/                     # Private application code
│   │   ├── domain/                  # Domain layer (DDD)
│   │   │   ├── entities/            # Core business entities
│   │   │   │   ├── user.go
│   │   │   │   ├── wallet.go
│   │   │   │   ├── transaction.go
│   │   │   │   ├── kyc_profile.go
│   │   │   │   └── exchange_operation.go
│   │   │   ├── repositories/        # Repository interfaces
│   │   │   │   ├── user_repository.go
│   │   │   │   ├── wallet_repository.go
│   │   │   │   ├── transaction_repository.go
│   │   │   │   ├── kyc_repository.go
│   │   │   │   └── rate_repository.go
│   │   │   └── services/            # Domain services
│   │   │       ├── wallet_service.go
│   │   │       ├── transaction_service.go
│   │   │       └── exchange_service.go
│   │   ├── application/             # Application layer (Use Cases)
│   │   │   ├── usecases/
│   │   │   │   ├── auth/            # Authentication use cases
│   │   │   │   │   ├── login.go
│   │   │   │   │   ├── register.go
│   │   │   │   │   └── logout.go
│   │   │   │   ├── wallet/          # Wallet management use cases
│   │   │   │   │   ├── create_wallet.go
│   │   │   │   │   ├── get_balance.go
│   │   │   │   │   └── list_wallets.go
│   │   │   │   ├── transaction/     # Transaction use cases
│   │   │   │   │   ├── send_transaction.go
│   │   │   │   │   ├── get_transaction_status.go
│   │   │   │   │   └── list_transactions.go
│   │   │   │   ├── exchange/        # Exchange use cases
│   │   │   │   │   ├── get_exchange_rate.go
│   │   │   │   │   ├── swap_tokens.go
│   │   │   │   │   └── list_swap_history.go
│   │   │   │   ├── kyc/             # KYC use cases
│   │   │   │   │   ├── submit_kyc.go
│   │   │   │   │   ├── upload_document.go
│   │   │   │   │   └── get_kyc_status.go
│   │   │   │   └── analytics/       # Analytics use cases
│   │   │   │       ├── get_portfolio_summary.go
│   │   │   │       └── get_transaction_history.go
│   │   │   └── dto/                 # Data Transfer Objects
│   │   │       ├── wallet_dto.go
│   │   │       ├── transaction_dto.go
│   │   │       └── kyc_dto.go
│   │   ├── infrastructure/          # Infrastructure layer
│   │   │   ├── blockchain/          # Blockchain adapters
│   │   │   │   ├── adapter.go       # Common interface
│   │   │   │   ├── bitcoin_adapter.go
│   │   │   │   ├── ethereum_adapter.go
│   │   │   │   ├── solana_adapter.go
│   │   │   │   └── stellar_adapter.go
│   │   │   ├── repository/          # Repository implementations
│   │   │   │   ├── postgres/
│   │   │   │   │   ├── user_repo.go
│   │   │   │   │   ├── wallet_repo.go
│   │   │   │   │   ├── transaction_repo.go
│   │   │   │   │   ├── kyc_repo.go
│   │   │   │   │   └── rate_repo.go
│   │   │   │   └── cache/
│   │   │   │       └── redis_cache.go
│   │   │   ├── security/            # Security utilities
│   │   │   │   ├── encryption.go    # AES-GCM encryption
│   │   │   │   ├── hashing.go       # bcrypt password hashing
│   │   │   │   └── jwt.go           # JWT token handling
│   │   │   ├── messaging/           # Message queue
│   │   │   │   └── redis_pubsub.go  # Redis Pub/Sub for WebSocket
│   │   │   └── external/            # External service clients
│   │   │       ├── coingecko.go     # CoinGecko API client
│   │   │       └── kyc_provider.go  # KYC service integration
│   │   └── interfaces/              # Interface layer
│   │       ├── http/                # HTTP handlers
│   │       │   ├── handlers/
│   │       │   │   ├── auth_handler.go
│   │       │   │   ├── wallet_handler.go
│   │       │   │   ├── transaction_handler.go
│   │       │   │   ├── exchange_handler.go
│   │       │   │   ├── kyc_handler.go
│   │       │   │   ├── rate_handler.go
│   │       │   │   └── analytics_handler.go
│   │       │   ├── middleware/
│   │       │   │   ├── auth_middleware.go
│   │       │   │   ├── cors_middleware.go
│   │       │   │   ├── rate_limit_middleware.go
│   │       │   │   └── logging_middleware.go
│   │       │   └── routes.go        # Route definitions
│   │       └── websocket/           # WebSocket handlers
│   │           └── rates_handler.go # Real-time price updates
│   ├── configs/                     # Configuration files
│   │   ├── config.yaml              # Main configuration
│   │   └── blockchain.yaml          # Blockchain RPC endpoints
│   ├── db/                          # Database files
│   │   └── migrations/              # SQL migration files
│   │       ├── core_db/
│   │       │   ├── 001_initial_schema.up.sql
│   │       │   └── 001_initial_schema.down.sql
│   │       ├── kyc_db/
│   │       │   ├── 001_initial_schema.up.sql
│   │       │   └── 001_initial_schema.down.sql
│   │       ├── rates_db/
│   │       │   ├── 001_initial_schema.up.sql
│   │       │   └── 001_initial_schema.down.sql
│   │       └── audit_db/
│   │           ├── 001_initial_schema.up.sql
│   │           └── 001_initial_schema.down.sql
│   ├── pkg/                         # Public libraries (if any)
│   │   └── utils/
│   │       ├── validator.go
│   │       └── error_handler.go
│   ├── tests/                       # Test files
│   │   ├── integration/
│   │   │   ├── wallet_test.go
│   │   │   ├── transaction_test.go
│   │   │   └── exchange_test.go
│   │   ├── unit/
│   │   │   ├── domain/
│   │   │   ├── application/
│   │   │   └── infrastructure/
│   │   └── mocks/                   # Mock implementations
│   │       ├── blockchain_mock.go
│   │       └── repository_mock.go
│   ├── scripts/                     # Utility scripts
│   │   ├── setup.sh                 # Initial setup script
│   │   └── seed_data.sh             # Database seeding
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   ├── Makefile                     # Build and run commands
│   └── README.md
│
├── frontend/                        # Next.js frontend (crypto-wallet-frontend/)
│   ├── src/
│   │   ├── app/                     # Next.js 15 App Router pages
│   │   │   ├── layout.tsx           # Root layout
│   │   │   ├── page.tsx             # Dashboard (home)
│   │   │   ├── login/
│   │   │   │   └── page.tsx         # Login page
│   │   │   ├── register/
│   │   │   │   └── page.tsx         # Registration page
│   │   │   ├── wallets/
│   │   │   │   ├── page.tsx         # Wallet list
│   │   │   │   └── [id]/
│   │   │   │       └── page.tsx     # Wallet details
│   │   │   ├── transactions/
│   │   │   │   ├── page.tsx         # Transaction list
│   │   │   │   └── [txid]/
│   │   │   │       └── page.tsx     # Transaction details
│   │   │   ├── swap/
│   │   │   │   └── page.tsx         # Swap interface
│   │   │   ├── kyc/
│   │   │   │   └── page.tsx         # KYC verification
│   │   │   ├── notifications/
│   │   │   │   └── page.tsx         # Notifications
│   │   │   ├── analytics/
│   │   │   │   └── page.tsx         # Portfolio analytics
│   │   │   └── settings/
│   │   │       └── page.tsx         # User settings
│   │   ├── components/              # React components
│   │   │   ├── ui/                  # ShadCN/UI base components
│   │   │   │   ├── button.tsx
│   │   │   │   ├── card.tsx
│   │   │   │   ├── input.tsx
│   │   │   │   ├── dialog.tsx
│   │   │   │   ├── toast.tsx
│   │   │   │   └── ...
│   │   │   ├── layout/              # Layout components
│   │   │   │   ├── Sidebar.tsx
│   │   │   │   ├── Navbar.tsx
│   │   │   │   ├── Footer.tsx
│   │   │   │   └── ThemeToggle.tsx
│   │   │   ├── wallet/              # Wallet components
│   │   │   │   ├── WalletCard.tsx
│   │   │   │   ├── WalletList.tsx
│   │   │   │   ├── AddWalletModal.tsx
│   │   │   │   └── WalletBalance.tsx
│   │   │   ├── transaction/         # Transaction components
│   │   │   │   ├── TxTable.tsx
│   │   │   │   ├── TxStatusBadge.tsx
│   │   │   │   ├── TxDetails.tsx
│   │   │   │   └── SendTxForm.tsx
│   │   │   ├── swap/                # Swap components
│   │   │   │   ├── SwapForm.tsx
│   │   │   │   ├── SwapRateCard.tsx
│   │   │   │   └── SwapHistory.tsx
│   │   │   ├── kyc/                 # KYC components
│   │   │   │   ├── KYCForm.tsx
│   │   │   │   ├── DocumentUpload.tsx
│   │   │   │   └── KYCStatus.tsx
│   │   │   ├── charts/              # Chart components
│   │   │   │   ├── RateChart.tsx
│   │   │   │   ├── BalanceChart.tsx
│   │   │   │   └── TxVolumeChart.tsx
│   │   │   └── auth/                # Auth components
│   │   │       ├── LoginForm.tsx
│   │   │       └── RegisterForm.tsx
│   │   ├── lib/                     # Utility libraries
│   │   │   ├── api.ts               # Axios API client
│   │   │   ├── auth.ts              # Auth utilities
│   │   │   ├── websocket.ts         # WebSocket client
│   │   │   ├── constants.ts         # Constants and enums
│   │   │   ├── validators.ts        # Zod validation schemas
│   │   │   ├── helpers.ts           # Helper functions
│   │   │   └── stores/              # Zustand stores
│   │   │       ├── useAuthStore.ts
│   │   │       ├── useWalletStore.ts
│   │   │       ├── useRateStore.ts
│   │   │       └── useNotificationStore.ts
│   │   ├── hooks/                   # Custom React hooks
│   │   │   ├── useWallets.ts
│   │   │   ├── useTransactions.ts
│   │   │   ├── useRates.ts
│   │   │   └── useWebSocket.ts
│   │   ├── types/                   # TypeScript types
│   │   │   ├── wallet.ts
│   │   │   ├── transaction.ts
│   │   │   ├── user.ts
│   │   │   └── api.ts
│   │   └── styles/                  # Global styles
│   │       └── globals.css
│   ├── public/                      # Static assets
│   │   ├── icons/
│   │   └── images/
│   ├── tests/                       # Frontend tests
│   │   ├── e2e/                     # Playwright E2E tests
│   │   │   ├── auth.spec.ts
│   │   │   ├── wallet.spec.ts
│   │   │   └── transaction.spec.ts
│   │   └── unit/                    # Jest unit tests
│   │       └── components/
│   ├── next.config.ts
│   ├── tailwind.config.ts
│   ├── tsconfig.json
│   ├── package.json
│   ├── Dockerfile
│   └── README.md
│
├── database/                        # Database initialization scripts
│   ├── init-databases.sh            # Create 4 databases
│   └── seed/                        # Seed data scripts
│       ├── core_seed.sql
│       ├── kyc_seed.sql
│       ├── rates_seed.sql
│       └── audit_seed.sql
│
├── docker/                          # Docker-related files
│   ├── backend.Dockerfile
│   ├── frontend.Dockerfile
│   └── nginx.conf                   # Nginx reverse proxy config
│
├── monitoring/                      # Monitoring configuration
│   ├── prometheus/
│   │   └── prometheus.yml
│   └── grafana/
│       └── dashboards/
│           ├── api_metrics.json
│           └── blockchain_metrics.json
│
├── docker-compose.yml               # Docker Compose orchestration
├── docker-compose.dev.yml           # Development overrides
├── .env.example                     # Environment variables template
├── Makefile                         # Project-wide make commands
└── README.md                        # Project documentation
```

**Structure Decision**:

This project uses **Option 2: Web Application** structure due to the presence of both backend (Go API) and frontend (Next.js) components detected in the requirements. The structure follows these design decisions:

1. **Backend (crypto-wallet-backend/)**:
   - Implements DDD + Hexagonal Architecture with clear layer separation
   - Domain layer contains pure business logic (entities, repositories, domain services)
   - Application layer orchestrates use cases and defines DTOs
   - Infrastructure layer handles technical concerns (database, blockchain, security, external APIs)
   - Interface layer provides HTTP REST API and WebSocket endpoints
   - Currently empty and needs complete scaffolding

2. **Frontend (crypto-wallet-frontend/)**:
   - Follows Next.js 15 App Router conventions
   - Component-based architecture with ShadCN/UI design system
   - State management via Zustand for global state
   - Separation of concerns: pages, components, hooks, utilities
   - Already has basic Next.js scaffold but needs feature implementation

3. **Database**:
   - 4 separate PostgreSQL databases for separation of concerns
   - Migration-based schema management
   - Seed scripts for development data

4. **Infrastructure**:
   - Docker Compose for local development and deployment
   - Prometheus + Grafana for monitoring
   - Redis for caching and Pub/Sub messaging

5. **Testing**:
   - Backend: Unit tests alongside source, integration tests in dedicated folder
   - Frontend: E2E tests with Playwright, component tests with Jest
   - Contract tests for API validation

The root directory contains shared infrastructure configuration (Docker, monitoring) while keeping backend and frontend codebases separate and independently deployable.

## Complexity Tracking

_This section documents any violations of project constitution principles that require justification._

**Status**: No constitutional principles have been defined yet, so no violations exist to track.

**Anticipated Complexity Justifications** (for future constitution):

If constitutional principles around simplicity are established, the following design decisions may require justification:

| Design Decision             | Business Justification                                                                                                                                             | Simpler Alternative Rejected                                                                                                     |
| --------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------- |
| 4 Separate Databases        | Regulatory compliance requires isolation of KYC data; Performance optimization through separation of operational, compliance, market, and audit data               | Single database with schema separation insufficient for compliance auditing and data access control requirements                 |
| Hexagonal Architecture      | Multi-blockchain support requires pluggable adapters; Future blockchain additions must not require core logic changes; Testing requires mock blockchain services   | Direct blockchain SDK integration would create tight coupling and make testing extremely difficult with real blockchain networks |
| Redis Pub/Sub for WebSocket | Real-time price updates to thousands of concurrent users requires scalable broadcast mechanism; Horizontal scaling of WebSocket servers needs message distribution | Polling REST API would create excessive load (10K users × 5 second intervals = 2K req/s just for prices)                         |
| Separate Frontend/Backend   | Security requires backend validation of all financial operations; Mobile apps planned for future require API-first architecture                                    | Monolithic Next.js with server actions insufficient for mobile client support and blockchain transaction validation              |

---

## Phase 0: Outline & Research

**Objective**: Resolve all technical uncertainties and establish concrete technology choices for backend and frontend implementation.

### Research Tasks

1. **Backend Framework Selection**
   - **Unknown**: Whether to use Fiber or Echo for Go web framework
   - **Research Task**: Compare Fiber vs Echo for crypto wallet API requirements
   - **Success Criteria**: Documented comparison with performance benchmarks, middleware ecosystem, WebSocket support, and final recommendation

2. **Blockchain SDK Integration Patterns**
   - **Unknown**: Best practices for abstracting blockchain-specific implementations behind unified interface
   - **Research Task**: Research blockchain adapter patterns for multi-chain wallet systems
   - **Success Criteria**: Documented design pattern with example implementations for BTC/ETH adapters

3. **Database Migration Strategy**
   - **Unknown**: golang-migrate vs other migration tools, versioning strategy for 4 databases
   - **Research Task**: Research database migration tools and strategies for multi-database Go applications
   - **Success Criteria**: Selected migration tool with rationale, migration file structure, rollback strategy

4. **Private Key Encryption Approach**
   - **Unknown**: Specific implementation of AES-GCM encryption for private keys, key derivation, IV generation
   - **Research Task**: Research industry-standard private key encryption practices for cryptocurrency wallets
   - **Success Criteria**: Documented encryption flow with code examples, security audit considerations

5. **JWT Session Management**
   - **Unknown**: JWT storage strategy (HttpOnly cookies vs Authorization header), refresh token implementation
   - **Research Task**: Research secure JWT implementation patterns for financial applications
   - **Success Criteria**: Documented authentication flow with security considerations, code examples

6. **WebSocket Price Feed Architecture**
   - **Unknown**: Redis Pub/Sub implementation details, connection management, reconnection logic
   - **Research Task**: Research WebSocket + Redis Pub/Sub patterns for real-time data broadcast
   - **Success Criteria**: Architecture diagram with connection flow, error handling, scalability considerations

7. **Frontend State Management**
   - **Unknown**: Zustand store structure for wallet, transaction, rate data; persistence strategy
   - **Research Task**: Research Zustand best practices for complex financial application state
   - **Success Criteria**: Store architecture document with data flow diagrams, persistence patterns

8. **Transaction Queue Worker Design**
   - **Unknown**: Go routine vs external worker (cron), polling intervals, blockchain confirmation thresholds
   - **Research Task**: Research background job processing patterns in Go for blockchain transaction monitoring
   - **Success Criteria**: Worker architecture document with deployment options, error handling, scalability

9. **API Rate Limiting Strategy**
   - **Unknown**: Rate limiting middleware implementation, storage backend, limits per endpoint
   - **Research Task**: Research API rate limiting best practices for fintech applications
   - **Success Criteria**: Rate limiting strategy document with implementation approach, configuration

10. **Testing Infrastructure Setup**
    - **Unknown**: testcontainers-go usage for integration tests, mock blockchain adapter patterns
    - **Research Task**: Research Go testing best practices for applications with external dependencies
    - **Success Criteria**: Testing strategy document with examples for unit, integration, contract tests

11. **Frontend WebSocket Client**
    - **Unknown**: WebSocket library choice (native vs library), reconnection logic, state synchronization
    - **Research Task**: Research WebSocket client patterns for Next.js applications
    - **Success Criteria**: WebSocket integration document with reconnection strategy, error handling

12. **KYC Provider Integration**
    - **Unknown**: Specific KYC service provider (Onfido, SumSub, etc.), integration patterns
    - **Research Task**: Research KYC service providers and integration approaches
    - **Success Criteria**: Provider comparison with integration architecture, API contract definition

### Research Output Format

All research findings will be consolidated in `research.md` using this format:

```markdown
## [Research Topic]

### Decision

[What was chosen]

### Rationale

[Why this choice was made - technical reasons, business requirements, trade-offs]

### Alternatives Considered

[What other options were evaluated]

- **Option A**: [description] - Rejected because [reason]
- **Option B**: [description] - Rejected because [reason]

### Implementation Notes

[Specific guidance for Phase 1 design and Phase 2 implementation]

### References

- [Link to documentation]
- [Link to best practice guide]
- [Link to example implementation]
```

---

## Phase 1: Design & Contracts

**Prerequisites**: `research.md` complete with all research tasks resolved

### 1. Data Model Design

**Input**: Feature spec entities, functional requirements, research findings

**Output**: `data-model.md` containing:

#### Core Database (core_db)

**Users Table**:

- Entity: User account holder
- Fields: id (UUID, PK), email (unique, indexed), password_hash, created_at, updated_at, last_login_at, is_active, preferred_currency
- Relationships: 1:N with Wallets, 1:1 with Accounts
- Validation: Email format, password complexity (12+ chars), unique email
- State Transitions: N/A (simple active/inactive flag)

**Wallets Table**:

- Entity: Blockchain-specific cryptocurrency wallet
- Fields: id (UUID, PK), user_id (FK to Users), chain (enum: BTC/ETH/SOL/XLM), address (unique per chain), encrypted_private_key, derivation_path, balance, created_at, updated_at
- Relationships: N:1 with Users, 1:N with Transactions
- Validation: Valid blockchain address format per chain, encrypted private key required
- State Transitions: N/A (wallets are permanent once created)
- Indexes: user_id, chain, address

**Transactions Table**:

- Entity: Cryptocurrency transfer operation
- Fields: id (UUID, PK), wallet_id (FK to Wallets), chain, tx_hash (unique), type (enum: send/receive), amount, fee, status (enum: pending/confirmed/failed), from_address, to_address, block_number, confirmations, created_at, confirmed_at
- Relationships: N:1 with Wallets
- Validation: Positive amounts, valid addresses per chain, tx_hash format
- State Transitions: pending → confirmed (on blockchain confirmation) OR pending → failed (on error)
- Indexes: wallet_id, status, chain, tx_hash, created_at

**Chains Table**:

- Entity: Supported blockchain networks
- Fields: id (SERIAL, PK), symbol (unique, e.g., BTC), name, rpc_url, explorer_url, is_active, native_token_symbol, confirmation_threshold
- Relationships: 1:N with Tokens
- Validation: Valid RPC URL format, positive confirmation threshold
- State Transitions: N/A (configuration data)

**Tokens Table**:

- Entity: Supported cryptocurrencies
- Fields: id (SERIAL, PK), symbol (unique), name, chain_symbol (FK to Chains), contract_address (nullable for native), decimals, is_native, logo_url, is_active
- Relationships: N:1 with Chains
- Validation: Valid symbol format, contract address format for non-native tokens
- State Transitions: N/A (configuration data)

**Accounts Table**:

- Entity: User balance aggregation
- Fields: id (UUID, PK), user_id (FK to Users, unique), total_balance_usd, created_at, updated_at
- Relationships: 1:1 with Users, 1:N with Ledger_Entries
- Validation: Non-negative balance
- State Transitions: N/A (calculated field)

**Ledger_Entries Table**:

- Entity: Double-entry accounting for balance tracking
- Fields: id (UUID, PK), account_id (FK to Accounts), transaction_id (FK to Transactions, nullable), entry_type (enum: debit/credit), amount, currency, description, created_at
- Relationships: N:1 with Accounts, N:1 with Transactions (optional)
- Validation: Positive amounts, balanced debits/credits per transaction
- State Transitions: N/A (immutable once created)
- Indexes: account_id, transaction_id, created_at

#### KYC Database (kyc_db)

**KYC_Profiles Table**:

- Entity: User identity verification information
- Fields: id (UUID, PK), user_id (UUID, indexed, not FK for data isolation), verification_level (enum: unverified/basic/full), status (enum: not_started/pending/approved/rejected), submitted_at, verified_at, rejection_reason, encrypted_data (JSON)
- Relationships: Logically 1:1 with Users (via user_id, but no foreign key for isolation)
- Validation: Valid verification level and status combinations
- State Transitions: not_started → pending (on submission) → approved/rejected (on review)
- Indexes: user_id (unique), status, verification_level

**KYC_Documents Table**:

- Entity: Uploaded identity verification documents
- Fields: id (UUID, PK), kyc_profile_id (FK to KYC_Profiles), document_type (enum: passport/id_card/drivers_license/proof_of_address), file_path_encrypted, file_hash, uploaded_at, reviewed_at, status (enum: pending/approved/rejected), rejection_reason
- Relationships: N:1 with KYC_Profiles
- Validation: Supported document type, encrypted file path required
- State Transitions: pending → approved/rejected (on review)
- Indexes: kyc_profile_id, status

**User_Risk_Score Table**:

- Entity: User risk assessment for AML compliance
- Fields: id (UUID, PK), user_id (UUID, indexed), risk_score (0-100), risk_level (enum: low/medium/high), last_updated, factors (JSON)
- Relationships: Logically 1:1 with Users (via user_id)
- Validation: Risk score between 0-100, risk level consistent with score
- State Transitions: N/A (updated on events)
- Indexes: user_id (unique), risk_level

**Alert_Rules Table**:

- Entity: AML rule configuration
- Fields: id (UUID, PK), rule_name, rule_type, threshold_value, is_active, created_at, updated_at
- Relationships: N/A (configuration table)
- Validation: Positive threshold values
- State Transitions: N/A (configuration data)

#### Rates Database (rates_db)

**Exchange_Rates Table**:

- Entity: Current cryptocurrency prices
- Fields: id (UUID, PK), symbol (indexed), price_usd, price_change_24h, volume_24h, market_cap, last_updated
- Relationships: N/A (market data)
- Validation: Positive prices and volumes
- State Transitions: N/A (continuously updated)
- Indexes: symbol (unique), last_updated

**Trading_Pairs Table**:

- Entity: Supported exchange pairs
- Fields: id (UUID, PK), base_symbol, quote_symbol, exchange_rate, fee_percentage, is_active, last_updated
- Relationships: N/A (configuration/market data)
- Validation: Positive exchange rate and fees, different base/quote symbols
- State Transitions: N/A (continuously updated)
- Indexes: (base_symbol, quote_symbol) composite unique

**Price_History Table**:

- Entity: Historical price data for charts
- Fields: id (UUID, PK), symbol (indexed), price_usd, timestamp, interval (enum: 1m/5m/1h/1d)
- Relationships: N/A (time-series data)
- Validation: Positive prices, valid interval
- State Transitions: N/A (immutable time-series)
- Indexes: (symbol, timestamp) composite, interval

**Transaction_Summary Table**:

- Entity: Aggregated transaction statistics
- Fields: id (UUID, PK), chain, date, transaction_count, total_volume_usd, unique_users, avg_transaction_value
- Relationships: N/A (aggregated metrics)
- Validation: Positive counts and volumes
- State Transitions: N/A (daily aggregation job)
- Indexes: (chain, date) composite unique

#### Audit Database (audit_db)

**Audit_Logs Table**:

- Entity: Immutable audit trail of all critical operations
- Fields: id (UUID, PK), user_id (UUID, indexed), action (enum: wallet_create/tx_send/kyc_submit/etc), resource_type, resource_id, details (JSON), ip_address, user_agent, timestamp
- Relationships: Logically linked to Users via user_id (no FK for audit independence)
- Validation: Valid action enum, timestamp required
- State Transitions: N/A (immutable, append-only)
- Indexes: user_id, action, timestamp, resource_type

**Security_Logs Table**:

- Entity: Security events and anomalies
- Fields: id (UUID, PK), user_id (UUID, nullable, indexed), event_type (enum: login_success/login_fail/2fa_enabled/etc), severity (enum: info/warning/critical), ip_address, user_agent, details (JSON), timestamp
- Relationships: Logically linked to Users via user_id (optional, no FK)
- Validation: Valid event type and severity, timestamp required
- State Transitions: N/A (immutable, append-only)
- Indexes: user_id, event_type, severity, timestamp

**API_Audit Table**:

- Entity: API request/response logging for compliance
- Fields: id (UUID, PK), user_id (UUID, nullable), endpoint, method, status_code, request_body_hash, response_body_hash, duration_ms, ip_address, timestamp
- Relationships: Logically linked to Users via user_id (optional, no FK)
- Validation: Valid HTTP method and status code, positive duration
- State Transitions: N/A (immutable, append-only)
- Indexes: user_id, endpoint, status_code, timestamp

### 2. API Contract Generation

**Input**: Functional requirements, user scenarios, data model

**Output**: OpenAPI 3.0 specification in `contracts/openapi.yaml` covering:

#### Authentication Endpoints

- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login (sets HttpOnly JWT cookie)
- `POST /api/v1/auth/logout` - User logout (clears cookie)
- `POST /api/v1/auth/refresh` - Refresh JWT token
- `GET /api/v1/auth/me` - Get current user profile

#### Wallet Endpoints

- `GET /api/v1/wallets` - List user wallets
- `POST /api/v1/wallets` - Create new wallet
- `GET /api/v1/wallets/:id` - Get wallet details
- `GET /api/v1/wallets/:id/balance` - Get wallet balance
- `GET /api/v1/wallets/:id/transactions` - Get wallet transaction history
- `POST /api/v1/wallets/:id/export-address` - Export wallet address (with QR code)

#### Transaction Endpoints

- `POST /api/v1/transactions` - Create and broadcast transaction
- `GET /api/v1/transactions/:id` - Get transaction details
- `GET /api/v1/transactions` - List user transactions with filters
- `POST /api/v1/transactions/:id/estimate-fee` - Estimate transaction fee
- `GET /api/v1/transactions/:id/status` - Get transaction status

#### Exchange Endpoints

- `GET /api/v1/exchange/rates` - Get current exchange rates
- `POST /api/v1/exchange/quote` - Get exchange quote
- `POST /api/v1/exchange/swap` - Execute token swap
- `GET /api/v1/exchange/history` - Get swap history

#### KYC Endpoints

- `POST /api/v1/kyc/submit` - Submit KYC profile information
- `POST /api/v1/kyc/documents` - Upload verification document
- `GET /api/v1/kyc/status` - Get KYC verification status
- `GET /api/v1/kyc/requirements` - Get KYC requirements for verification level

#### Rate Endpoints

- `GET /api/v1/rates` - Get current cryptocurrency prices
- `GET /api/v1/rates/history` - Get historical price data
- `GET /api/v1/rates/trading-pairs` - Get supported trading pairs

#### Analytics Endpoints

- `GET /api/v1/analytics/portfolio` - Get portfolio summary
- `GET /api/v1/analytics/performance` - Get portfolio performance over time
- `GET /api/v1/analytics/allocation` - Get asset allocation
- `GET /api/v1/analytics/transactions-summary` - Get transaction summary statistics

#### Notification Endpoints

- `GET /api/v1/notifications` - Get user notifications
- `PUT /api/v1/notifications/:id/read` - Mark notification as read
- `GET /api/v1/notifications/preferences` - Get notification preferences
- `PUT /api/v1/notifications/preferences` - Update notification preferences

#### User Settings Endpoints

- `GET /api/v1/users/profile` - Get user profile
- `PUT /api/v1/users/profile` - Update user profile
- `PUT /api/v1/users/password` - Change password
- `POST /api/v1/users/2fa/enable` - Enable 2FA
- `POST /api/v1/users/2fa/verify` - Verify 2FA code
- `POST /api/v1/users/2fa/disable` - Disable 2FA

#### Health & Monitoring Endpoints

- `GET /health` - Health check
- `GET /metrics` - Prometheus metrics (authenticated, admin only)

**OpenAPI Specification Structure**:

- OpenAPI version 3.0.3
- All endpoints documented with request/response schemas
- Error response schemas (400, 401, 403, 404, 500)
- Security schemes (JWT Bearer token via Cookie)
- Request/response examples for each endpoint
- Data validation rules (min/max length, patterns, enums)

**Additional Contract Documentation**:

**`contracts/websocket.md`** - WebSocket protocol documentation:

- Connection: `ws://api:8080/ws/rates`
- Authentication: JWT token in query parameter
- Message format: JSON with `{ event, data }` structure
- Events: `price_update`, `connection_status`, `error`
- Reconnection logic and heartbeat protocol

**`contracts/blockchain.md`** - Blockchain adapter interface contracts:

- `BlockchainAdapter` interface definition
- Method signatures for each blockchain operation
- Error codes and handling strategies
- Blockchain-specific configuration parameters

### 3. Quickstart Guide

**Output**: `quickstart.md` containing:

#### Prerequisites

- Go 1.23+ installed
- Node.js 20+ and npm installed
- Docker and Docker Compose installed
- PostgreSQL client tools (psql) for manual DB access
- Make utility installed

#### Backend Setup (5 minutes)

1. Clone repository and navigate to backend directory
2. Copy `.env.example` to `.env` and configure:
   - Database connection strings (4 databases)
   - Blockchain RPC endpoints
   - JWT secret key
   - Redis connection string
   - External API keys (CoinGecko, KYC provider)
3. Run `make setup` to:
   - Install Go dependencies
   - Create databases
   - Run migrations
   - Seed initial data
4. Run `make run` to start backend server on `http://localhost:8080`
5. Verify health: `curl http://localhost:8080/health`

#### Frontend Setup (3 minutes)

1. Navigate to frontend directory
2. Copy `.env.local.example` to `.env.local`:
   - Set `NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1`
3. Run `npm install` to install dependencies
4. Run `npm run dev` to start development server on `http://localhost:3000`
5. Open browser to `http://localhost:3000`

#### Docker Compose Setup (2 minutes)

1. Run `docker-compose up -d` from project root
2. Wait for services to initialize (~30 seconds)
3. Access:
   - Frontend: `http://localhost:3000`
   - Backend API: `http://localhost:8080`
   - pgAdmin: `http://localhost:5050`
   - Grafana: `http://localhost:3001`

#### Running Tests

- Backend: `cd backend && make test`
- Frontend: `cd frontend && npm test`
- E2E: `cd frontend && npm run test:e2e`

#### Common Development Tasks

- Generate migration: `make migration name=description`
- Reset database: `make db-reset`
- View logs: `docker-compose logs -f [service]`
- Lint code: `make lint` (backend), `npm run lint` (frontend)
- Format code: `make fmt` (backend), `npm run format` (frontend)

### 4. Agent Context Update

**Action**: Run `.specify/scripts/bash/update-agent-context.sh claude`

**Purpose**: Update Claude-specific context file with technologies used in this plan without duplicating existing content.

**Expected Output**:

- Detection of AI agent type (Claude in this case)
- Update to `.claude/project_context.md` or similar agent-specific file
- Addition of new technologies: Go, Fiber/Echo, Next.js 15, PostgreSQL 16, Redis, blockchain SDKs, etc.
- Preservation of manually added content between designated markers
- No duplication of technologies already documented

---

## Post-Phase-1 Constitution Check

**GATE: Re-evaluate constitutional compliance after design phase**

**Status**: No constitutional principles defined yet, so no post-design violations to check.

**If constitution is established before Phase 2**, verify:

1. ✅ Testing strategy aligns with constitution (TDD requirements, coverage thresholds)
2. ✅ API design follows documented conventions (REST patterns, versioning)
3. ✅ Security measures meet constitutional standards (encryption, authentication)
4. ✅ Data model follows documented architectural principles (DDD, Hexagonal)
5. ✅ Complexity is justified and documented in Complexity Tracking section

---

## Next Steps (Not Part of This Command)

**Phase 2: Task Generation** (`/speckit.tasks` command):

- Break down implementation into atomic, testable tasks
- Assign tasks to appropriate components (backend/frontend/database)
- Order tasks by dependencies
- Estimate effort for each task
- Generate `tasks.md` with complete implementation roadmap

**Phase 3: Implementation** (`/speckit.implement` command):

- Execute tasks from `tasks.md` in dependency order
- Create TDD test cases before implementation
- Implement features following architecture and contracts
- Validate against acceptance criteria
- Update progress in real-time

---

## Summary

This implementation plan provides a comprehensive roadmap for building a full-stack multi-chain cryptocurrency wallet and exchange platform. The plan:

✅ **Defines Technical Context**: Specifies all languages, frameworks, databases, and tools required
✅ **Documents Project Structure**: Provides detailed directory structure for both backend and frontend
✅ **Identifies Research Needs**: Lists 12 specific research tasks to resolve technical uncertainties
✅ **Plans Design Phase**: Outlines data model, API contracts, and quickstart guide deliverables
✅ **Addresses Complexity**: Documents justifications for architectural decisions
✅ **Provides Clear Path Forward**: Establishes foundation for Phase 2 task generation

**Key Architectural Highlights**:

- DDD + Hexagonal Architecture ensures clean separation of concerns and testability
- 4-database design provides regulatory compliance and performance optimization
- Blockchain adapter pattern enables easy addition of new chains
- Real-time WebSocket integration delivers live price updates to users
- Comprehensive security measures protect sensitive financial data
- Full test coverage strategy ensures production readiness

**Current Status**:

- Frontend has basic Next.js scaffold, needs feature implementation
- Backend is empty, needs complete scaffolding from scratch
- Database schemas need to be created via migrations
- Docker Compose infrastructure is defined but needs verification

**Ready for**: Phase 0 research execution to resolve all technical uncertainties and proceed to detailed design.
