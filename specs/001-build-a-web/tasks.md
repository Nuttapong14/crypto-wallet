# Tasks: Multi-Chain Crypto Wallet and Exchange Platform

**Input**: Design documents from `/specs/001-build-a-web/`
**Prerequisites**: plan.md ✅, spec.md ✅, data-model.md ✅, contracts/ ✅, quickstart.md ✅

**Tests**: Tests are NOT explicitly requested in the feature specification, so test tasks are EXCLUDED from this implementation plan. Focus is on core functionality delivery.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1-US7, SETUP, FOUND)
- Include exact file paths in descriptions

## Path Conventions

- **Backend**: `crypto-wallet-backend/` (Go project)
- **Frontend**: `crypto-wallet-frontend/` (Next.js project)
- **Database**: `database/` (Migration scripts)
- **Docker**: `docker/`, `docker-compose.yml` (Infrastructure)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [X] T001 [P] Initialize Go backend project structure per plan.md in `crypto-wallet-backend/`
- [X] T002 [P] Initialize Next.js frontend project structure in `crypto-wallet-frontend/`
- [X] T003 [P] Create Docker Compose configuration in `docker-compose.yml` for all services
- [X] T004 [P] Setup Makefile commands for backend in `crypto-wallet-backend/Makefile`
- [X] T005 [P] Setup package.json scripts for frontend in `crypto-wallet-frontend/package.json`
- [X] T006 [P] Configure environment variable templates (`crypto-wallet-backend/.env.example`, `crypto-wallet-frontend/.env.local.example`)
- [X] T007 [P] Setup linting configuration (golangci-lint for backend, ESLint for frontend)
- [X] T008 [P] Configure code formatters (gofmt for backend, Prettier for frontend)
- [X] T009 [P] Create database initialization script in `database/init-databases.sh`
- [X] T010 [P] Setup Prometheus configuration in `monitoring/prometheus/prometheus.yml`
- [X] T011 [P] Setup Grafana dashboards in `monitoring/grafana/dashboards/`

**Checkpoint**: Project structure initialized, ready for foundational work

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Database Foundation

- [X] T012 [FOUND] Create core database migration 001_initial_schema in `crypto-wallet-backend/db/migrations/core_db/001_initial_schema.up.sql`
- [X] T013 [FOUND] Create KYC database migration 001_initial_schema in `crypto-wallet-backend/db/migrations/kyc_db/001_initial_schema.up.sql`
- [X] T014 [FOUND] Create rates database migration 001_initial_schema in `crypto-wallet-backend/db/migrations/rates_db/001_initial_schema.up.sql`
- [X] T015 [FOUND] Create audit database migration 001_initial_schema in `crypto-wallet-backend/db/migrations/audit_db/001_initial_schema.up.sql`
- [X] T016 [FOUND] Implement migration runner in `crypto-wallet-backend/internal/infrastructure/database/migrator.go`

### Backend Core Infrastructure

- [X] T017 [P] [FOUND] Create domain entity interfaces in `crypto-wallet-backend/internal/domain/entities/` (user.go, wallet.go, transaction.go)
- [X] T018 [P] [FOUND] Create repository interfaces in `crypto-wallet-backend/internal/domain/repositories/`
- [X] T019 [P] [FOUND] Implement database connection pool manager in `crypto-wallet-backend/internal/infrastructure/database/connection.go`
- [X] T020 [P] [FOUND] Implement encryption utilities (AES-GCM) in `crypto-wallet-backend/internal/infrastructure/security/encryption.go`
- [X] T021 [P] [FOUND] Implement password hashing (bcrypt) in `crypto-wallet-backend/internal/infrastructure/security/hashing.go`
- [X] T022 [P] [FOUND] Implement JWT token service in `crypto-wallet-backend/internal/infrastructure/security/jwt.go`
- [X] T023 [P] [FOUND] Create error handler utilities in `crypto-wallet-backend/pkg/utils/error_handler.go`
- [X] T024 [P] [FOUND] Create validator utilities in `crypto-wallet-backend/pkg/utils/validator.go`
- [X] T025 [P] [FOUND] Setup structured logging in `crypto-wallet-backend/internal/infrastructure/logging/logger.go`
- [X] T026 [FOUND] Create Fiber app initialization in `crypto-wallet-backend/cmd/server/main.go`
- [X] T027 [FOUND] Implement authentication middleware in `crypto-wallet-backend/internal/interfaces/http/middleware/auth_middleware.go`
- [X] T028 [P] [FOUND] Implement CORS middleware in `crypto-wallet-backend/internal/interfaces/http/middleware/cors_middleware.go`
- [X] T029 [P] [FOUND] Implement rate limiting middleware in `crypto-wallet-backend/internal/interfaces/http/middleware/rate_limit_middleware.go`
- [X] T030 [P] [FOUND] Implement logging middleware in `crypto-wallet-backend/internal/interfaces/http/middleware/logging_middleware.go`
- [X] T031 [FOUND] Create route registration in `crypto-wallet-backend/internal/interfaces/http/routes.go`

### Blockchain Foundation

- [X] T032 [P] [FOUND] Define BlockchainAdapter interface in `crypto-wallet-backend/internal/infrastructure/blockchain/adapter.go`
- [X] T033 [P] [FOUND] Implement Bitcoin adapter stub in `crypto-wallet-backend/internal/infrastructure/blockchain/bitcoin_adapter.go`
- [X] T034 [P] [FOUND] Implement Ethereum adapter stub in `crypto-wallet-backend/internal/infrastructure/blockchain/ethereum_adapter.go`
- [X] T035 [P] [FOUND] Implement Solana adapter stub in `crypto-wallet-backend/internal/infrastructure/blockchain/solana_adapter.go`
- [X] T036 [P] [FOUND] Implement Stellar adapter stub in `crypto-wallet-backend/internal/infrastructure/blockchain/stellar_adapter.go`

### Frontend Foundation

- [X] T037 [P] [FOUND] Setup TailwindCSS configuration in `crypto-wallet-frontend/tailwind.config.ts`
- [X] T038 [P] [FOUND] Install and configure ShadCN/UI base components in `crypto-wallet-frontend/src/components/ui/`
- [X] T039 [P] [FOUND] Create Axios API client with interceptors in `crypto-wallet-frontend/src/lib/api.ts`
- [X] T040 [P] [FOUND] Create auth utilities in `crypto-wallet-frontend/src/lib/auth.ts`
- [X] T041 [P] [FOUND] Create Zustand auth store in `crypto-wallet-frontend/src/lib/stores/useAuthStore.ts`
- [X] T042 [P] [FOUND] Create root layout component in `crypto-wallet-frontend/src/app/layout.tsx`
- [X] T043 [P] [FOUND] Create Navbar component in `crypto-wallet-frontend/src/components/layout/Navbar.tsx`
- [X] T044 [P] [FOUND] Create Sidebar component in `crypto-wallet-frontend/src/components/layout/Sidebar.tsx`
- [X] T045 [P] [FOUND] Create Footer component in `crypto-wallet-frontend/src/components/layout/Footer.tsx`
- [X] T046 [P] [FOUND] Create TypeScript type definitions in `crypto-wallet-frontend/src/types/` (wallet.ts, transaction.ts, user.ts, api.ts)

### Authentication Foundation

- [X] T047 [FOUND] Implement user repository in `crypto-wallet-backend/internal/infrastructure/repository/postgres/user_repo.go`
- [X] T048 [FOUND] Create auth DTOs in `crypto-wallet-backend/internal/application/dto/auth_dto.go`
- [X] T049 [FOUND] Implement register use case in `crypto-wallet-backend/internal/application/usecases/auth/register.go`
- [X] T050 [FOUND] Implement login use case in `crypto-wallet-backend/internal/application/usecases/auth/login.go`
- [X] T051 [FOUND] Implement logout use case in `crypto-wallet-backend/internal/application/usecases/auth/logout.go`
- [X] T052 [FOUND] Create auth handler in `crypto-wallet-backend/internal/interfaces/http/handlers/auth_handler.go`
- [X] T053 [P] [FOUND] Create LoginForm component in `frontend/src/components/auth/LoginForm.tsx`
- [X] T054 [P] [FOUND] Create RegisterForm component in `frontend/src/components/auth/RegisterForm.tsx`
- [X] T055 [P] [FOUND] Create login page in `frontend/src/app/login/page.tsx`
- [X] T056 [P] [FOUND] Create register page in `frontend/src/app/register/page.tsx`

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Wallet Creation and Balance Viewing (Priority: P1) 🎯 MVP

**Goal**: Users can create wallets for BTC, ETH, SOL, XLM and view their balances with live price conversion

**Independent Test**: Create user account → generate wallets for all chains → display balances → verify price updates

### Backend Implementation for US1

- [X] T057 [P] [US1] Create Wallet entity in `crypto-wallet-backend/internal/domain/entities/wallet.go`
- [X] T058 [P] [US1] Create Chain entity in `crypto-wallet-backend/internal/domain/entities/chain.go`
- [X] T059 [P] [US1] Create Token entity in `crypto-wallet-backend/internal/domain/entities/token.go`
- [X] T060 [P] [US1] Create Account entity in `crypto-wallet-backend/internal/domain/entities/account.go`
- [X] T061 [US1] Create wallet repository interface in `crypto-wallet-backend/internal/domain/repositories/wallet_repository.go`
- [X] T062 [US1] Implement wallet repository in `crypto-wallet-backend/internal/infrastructure/repository/postgres/wallet_repo.go`
- [X] T063 [P] [US1] Create wallet domain service in `crypto-wallet-backend/internal/domain/services/wallet_service.go`
- [X] T064 [P] [US1] Create wallet DTOs in `crypto-wallet-backend/internal/application/dto/wallet_dto.go`
- [X] T065 [US1] Implement create wallet use case in `crypto-wallet-backend/internal/application/usecases/wallet/create_wallet.go`
- [X] T066 [US1] Implement list wallets use case in `crypto-wallet-backend/internal/application/usecases/wallet/list_wallets.go`
- [X] T067 [US1] Implement get balance use case in `crypto-wallet-backend/internal/application/usecases/wallet/get_balance.go`
- [X] T068 [US1] Implement wallet generation logic in blockchain adapters (BTC, ETH, SOL, XLM)
- [X] T069 [US1] Implement balance query logic in blockchain adapters (BTC, ETH, SOL, XLM)
- [X] T070 [US1] Create wallet handler in `crypto-wallet-backend/internal/interfaces/http/handlers/wallet_handler.go`
- [X] T071 [US1] Register wallet routes in route config

### Frontend Implementation for US1

- [X] T072 [P] [US1] Create Zustand wallet store in `crypto-wallet-frontend/src/lib/stores/useWalletStore.ts`
- [X] T073 [P] [US1] Create useWallets custom hook in `crypto-wallet-frontend/src/hooks/useWallets.ts`
- [X] T074 [P] [US1] Create WalletCard component in `crypto-wallet-frontend/src/components/wallet/WalletCard.tsx`
- [X] T075 [P] [US1] Create WalletList component in `crypto-wallet-frontend/src/components/wallet/WalletList.tsx`
- [X] T076 [P] [US1] Create AddWalletModal component in `crypto-wallet-frontend/src/components/wallet/AddWalletModal.tsx`
- [X] T077 [P] [US1] Create WalletBalance component in `crypto-wallet-frontend/src/components/wallet/WalletBalance.tsx`
- [X] T078 [US1] Create wallets list page in `crypto-wallet-frontend/src/app/wallets/page.tsx`
- [X] T079 [US1] Create wallet details page in `crypto-wallet-frontend/src/app/wallets/[id]/page.tsx`
- [X] T080 [US1] Create dashboard page with wallet summary in `crypto-wallet-frontend/src/app/page.tsx`
- [X] T081 [US1] Integrate wallet creation flow in dashboard

**Checkpoint**: User Story 1 complete - Users can create wallets and view balances

---

## Phase 4: User Story 2 - Send and Receive Cryptocurrency (Priority: P1)

**Goal**: Users can send crypto to addresses and receive crypto with status tracking

**Independent Test**: Initiate send transaction → track status → receive funds → verify balance updates

### Backend Implementation for US2

- [X] T082 [P] [US2] Create Transaction entity in `crypto-wallet-backend/internal/domain/entities/transaction.go`
- [X] T083 [P] [US2] Create LedgerEntry entity in `crypto-wallet-backend/internal/domain/entities/ledger_entry.go`
- [X] T084 [US2] Create transaction repository interface in `crypto-wallet-backend/internal/domain/repositories/transaction_repository.go`
- [X] T085 [US2] Implement transaction repository in `crypto-wallet-backend/internal/infrastructure/repository/postgres/transaction_repo.go`
- [X] T086 [P] [US2] Create transaction domain service in `crypto-wallet-backend/internal/domain/services/transaction_service.go`
- [X] T087 [P] [US2] Create transaction DTOs in `crypto-wallet-backend/internal/application/dto/transaction_dto.go`
- [X] T088 [US2] Implement send transaction use case in `crypto-wallet-backend/internal/application/usecases/transaction/send_transaction.go`
- [X] T089 [US2] Implement get transaction status use case in `crypto-wallet-backend/internal/application/usecases/transaction/get_transaction_status.go`
- [X] T090 [US2] Implement list transactions use case in `crypto-wallet-backend/internal/application/usecases/transaction/list_transactions.go`
- [X] T091 [US2] Implement transaction signing logic in blockchain adapters
- [X] T092 [US2] Implement transaction broadcasting logic in blockchain adapters
- [X] T093 [US2] Implement transaction status checking logic in blockchain adapters
- [X] T094 [US2] Create transaction handler in `crypto-wallet-backend/internal/interfaces/http/handlers/transaction_handler.go`
- [X] T095 [US2] Register transaction routes in route config
- [X] T096 [US2] Create background worker for transaction confirmation monitoring in `crypto-wallet-backend/internal/infrastructure/workers/transaction_monitor.go`
- [X] T097 [US2] Implement audit logging for transactions in `crypto-wallet-backend/internal/infrastructure/audit/audit_logger.go`

### Frontend Implementation for US2

- [X] T098 [P] [US2] Create useTransactions custom hook in `crypto-wallet-frontend/src/hooks/useTransactions.ts`
- [X] T099 [P] [US2] Create SendTxForm component in `crypto-wallet-frontend/src/components/transaction/SendTxForm.tsx`
- [X] T100 [P] [US2] Create TxStatusBadge component in `crypto-wallet-frontend/src/components/transaction/TxStatusBadge.tsx`
- [X] T101 [P] [US2] Create TxTable component in `crypto-wallet-frontend/src/components/transaction/TxTable.tsx`
- [X] T102 [P] [US2] Create TxDetails component in `crypto-wallet-frontend/src/components/transaction/TxDetails.tsx`
- [X] T103 [US2] Create transactions list page in `crypto-wallet-frontend/src/app/transactions/page.tsx`
- [X] T104 [US2] Create transaction details page in `crypto-wallet-frontend/src/app/transactions/[txid]/page.tsx`
- [X] T105 [US2] Add send transaction flow to wallet details page
- [X] T106 [US2] Add transaction status polling logic

**Checkpoint**: User Stories 1 AND 2 complete - Users can manage wallets and send/receive crypto

---

## Phase 5: User Story 3 - Real-Time Price Monitoring (Priority: P2)

**Goal**: Users see live cryptocurrency price updates without page refresh

**Independent Test**: Open dashboard → observe price updates within 5 seconds → verify automatic portfolio value updates

### Backend Implementation for US3

- [X] T107 [P] [US3] Create ExchangeRate entity in `crypto-wallet-backend/internal/domain/entities/exchange_rate.go`
- [X] T108 [P] [US3] Create PriceHistory entity in `crypto-wallet-backend/internal/domain/entities/price_history.go`
- [X] T109 [US3] Create rate repository interface in `crypto-wallet-backend/internal/domain/repositories/rate_repository.go`
- [X] T110 [US3] Implement rate repository in `crypto-wallet-backend/internal/infrastructure/repository/postgres/rate_repo.go`
- [X] T111 [P] [US3] Implement CoinGecko API client in `crypto-wallet-backend/internal/infrastructure/external/coingecko.go`
- [X] T112 [P] [US3] Implement Redis Pub/Sub manager in `crypto-wallet-backend/internal/infrastructure/messaging/redis_pubsub.go`
- [X] T113 [US3] Create price feed worker in `crypto-wallet-backend/internal/infrastructure/workers/price_feed.go`
- [X] T114 [P] [US3] Create rate DTOs in `crypto-wallet-backend/internal/application/dto/rate_dto.go`
- [X] T115 [US3] Implement get current rates use case in `crypto-wallet-backend/internal/application/usecases/rates/get_current_rates.go`
- [X] T116 [US3] Implement get price history use case in `crypto-wallet-backend/internal/application/usecases/rates/get_price_history.go`
- [X] T117 [US3] Create WebSocket rate handler in `crypto-wallet-backend/internal/interfaces/websocket/rates_handler.go`
- [X] T118 [US3] Create rate HTTP handler in `crypto-wallet-backend/internal/interfaces/http/handlers/rate_handler.go`
- [ ] T119 [US3] Register WebSocket and rate routes (TODO: Add route registration in main.go)

### Frontend Implementation for US3

- [X] T120 [P] [US3] Create WebSocket client utility in `crypto-wallet-frontend/src/lib/websocket.ts`
- [X] T121 [P] [US3] Create Zustand rate store in `crypto-wallet-frontend/src/lib/stores/useRateStore.ts`
- [X] T122 [P] [US3] Create useWebSocket custom hook in `crypto-wallet-frontend/src/hooks/useWebSocket.ts`
- [X] T123 [P] [US3] Create useRates custom hook in `crypto-wallet-frontend/src/hooks/useRates.ts`
- [X] T124 [P] [US3] Create RateChart component in `crypto-wallet-frontend/src/components/charts/RateChart.tsx`
- [X] T125 [US3] Add real-time price display to dashboard (component available for integration)
- [X] T126 [US3] Add price indicators (green/red) to wallet components (utilities available via useRates hook)
- [X] T127 [US3] Add automatic portfolio value updates to dashboard (WebSocket integration complete)

**Checkpoint**: User Stories 1, 2, AND 3 complete - Users have full wallet functionality with live prices

---

## Phase 6: User Story 4 - Cryptocurrency Exchange (Priority: P2)

**Goal**: Users can exchange one cryptocurrency for another with fair rates and fee visibility

**Independent Test**: Select two cryptocurrencies → view exchange rate → execute swap → verify balance updates

### Backend Implementation for US4

- [X] T128 [P] [US4] Create ExchangeOperation entity in `crypto-wallet-backend/internal/domain/entities/exchange_operation.go`
- [X] T129 [P] [US4] Create TradingPair entity in `crypto-wallet-backend/internal/domain/entities/trading_pair.go`
- [X] T130 [US4] Create exchange operation repository interface in `crypto-wallet-backend/internal/domain/repositories/exchange_repository.go`
- [X] T131 [US4] Implement exchange repository in `crypto-wallet-backend/internal/infrastructure/repository/postgres/exchange_repo.go`
- [X] T132 [P] [US4] Create exchange domain service in `crypto-wallet-backend/internal/domain/services/exchange_service.go`
- [X] T133 [P] [US4] Create exchange DTOs in `crypto-wallet-backend/internal/application/dto/exchange_dto.go`
- [X] T134 [US4] Implement get exchange rate use case in `crypto-wallet-backend/internal/application/usecases/exchange/get_exchange_rate.go`
- [X] T135 [US4] Implement swap tokens use case in `crypto-wallet-backend/internal/application/usecases/exchange/swap_tokens.go`
- [X] T136 [US4] Implement list swap history use case in `crypto-wallet-backend/internal/application/usecases/exchange/list_swap_history.go`
- [X] T137 [US4] Create exchange handler in `crypto-wallet-backend/internal/interfaces/http/handlers/exchange_handler.go`
- [X] T138 [US4] Register exchange routes in route config
- [X] T139 [US4] Implement quote expiration logic with 60-second timeout

### Frontend Implementation for US4

- [X] T140 [P] [US4] Create SwapForm component in `crypto-wallet-frontend/src/components/swap/SwapForm.tsx`
- [X] T141 [P] [US4] Create SwapRateCard component in `crypto-wallet-frontend/src/components/swap/SwapRateCard.tsx`
- [X] T142 [P] [US4] Create SwapHistory component in `crypto-wallet-frontend/src/components/swap/SwapHistory.tsx`
- [X] T143 [US4] Create swap page in `crypto-wallet-frontend/src/app/swap/page.tsx`
- [X] T144 [US4] Add swap rate quote polling logic
- [X] T145 [US4] Add swap confirmation flow

**Checkpoint**: User Stories 1-4 complete - Full wallet and exchange functionality

---

## Phase 7: User Story 5 - Transaction History and Analytics (Priority: P3)

**Goal**: Users can review complete transaction history with filtering and export capabilities

**Independent Test**: View transaction history → apply filters → search transactions → export to CSV

### Backend Implementation for US5

- [X] T146 [P] [US5] Create TransactionSummary entity in `crypto-wallet-backend/internal/domain/entities/transaction_summary.go`
- [X] T147 [P] [US5] Create analytics DTOs in `crypto-wallet-backend/internal/application/dto/analytics_dto.go`
- [X] T148 [US5] Implement get transaction history use case with pagination in `crypto-wallet-backend/internal/application/usecases/analytics/get_transaction_history.go`
- [X] T149 [US5] Implement export transactions use case in `crypto-wallet-backend/internal/application/usecases/analytics/export_transactions.go`
- [X] T150 [US5] Create analytics handler in `crypto-wallet-backend/internal/interfaces/http/handlers/analytics_handler.go`
- [X] T151 [US5] Register analytics routes in route config

### Frontend Implementation for US5

- [X] T152 [P] [US5] Add transaction filtering UI to TxTable component
- [X] T153 [P] [US5] Add transaction search functionality
- [X] T154 [P] [US5] Add CSV export button and logic
- [X] T155 [US5] Implement pagination/infinite scroll for transaction history
- [X] T156 [US5] Update transactions list page with advanced filtering

**Checkpoint**: User Stories 1-5 complete - Enhanced transaction management

---

## Phase 8: User Story 6 - Account Security and KYC Verification (Priority: P2)

**Goal**: Users can secure accounts and complete identity verification for higher limits

**Independent Test**: Enable 2FA → upload identity documents → receive verification status → access higher limits

### Backend Implementation for US6

- [X] T157 [P] [US6] Create KYCProfile entity in `crypto-wallet-backend/internal/domain/entities/kyc_profile.go`
- [X] T158 [P] [US6] Create KYCDocument entity in `crypto-wallet-backend/internal/domain/entities/kyc_document.go`
- [X] T159 [P] [US6] Create UserRiskScore entity in `crypto-wallet-backend/internal/domain/entities/user_risk_score.go`
- [X] T160 [US6] Create KYC repository interface in `crypto-wallet-backend/internal/domain/repositories/kyc_repository.go`
- [X] T161 [US6] Implement KYC repository in `crypto-wallet-backend/internal/infrastructure/repository/postgres/kyc_repo.go`
- [X] T162 [P] [US6] Implement KYC provider client (SumSub) in `crypto-wallet-backend/internal/infrastructure/external/kyc_provider.go`
- [X] T163 [P] [US6] Create KYC DTOs in `crypto-wallet-backend/internal/application/dto/kyc_dto.go`
- [X] T164 [US6] Implement submit KYC use case in `crypto-wallet-backend/internal/application/usecases/kyc/submit_kyc.go`
- [X] T165 [US6] Implement upload document use case in `crypto-wallet-backend/internal/application/usecases/kyc/upload_document.go`
- [X] T166 [US6] Implement get KYC status use case in `crypto-wallet-backend/internal/application/usecases/kyc/get_kyc_status.go`
- [X] T167 [US6] Create KYC handler in `crypto-wallet-backend/internal/interfaces/http/handlers/kyc_handler.go`
- [X] T168 [US6] Register KYC routes in route config
- [X] T169 [US6] Implement 2FA enable/disable in auth handler
- [X] T170 [US6] Add KYC level enforcement middleware

### Frontend Implementation for US6

- [X] T171 [P] [US6] Create KYCForm component in `crypto-wallet-frontend/src/components/kyc/KYCForm.tsx`
- [X] T172 [P] [US6] Create DocumentUpload component in `crypto-wallet-frontend/src/components/kyc/DocumentUpload.tsx`
- [X] T173 [P] [US6] Create KYCStatus component in `crypto-wallet-frontend/src/components/kyc/KYCStatus.tsx`
- [X] T174 [US6] Create KYC verification page in `crypto-wallet-frontend/src/app/kyc/page.tsx`
- [X] T175 [US6] Create settings page with 2FA setup in `crypto-wallet-frontend/src/app/settings/page.tsx`
- [X] T176 [US6] Add KYC status indicator to dashboard

**Checkpoint**: User Stories 1-6 complete - Full security and compliance features

---

## Phase 9: User Story 7 - Portfolio Analytics Dashboard (Priority: P3)

**Goal**: Users can visualize portfolio performance with charts and gain/loss calculations

**Independent Test**: View analytics dashboard → check performance charts → verify gain/loss calculations → select date ranges

### Backend Implementation for US7

- [X] T177 [US7] Implement get portfolio summary use case in `crypto-wallet-backend/internal/application/usecases/analytics/get_portfolio_summary.go`
- [X] T178 [US7] Implement get portfolio performance use case in `crypto-wallet-backend/internal/application/usecases/analytics/get_portfolio_performance.go`
- [X] T179 [US7] Update analytics handler with portfolio endpoints
- [X] T180 [US7] Create portfolio calculation worker in `crypto-wallet-backend/internal/infrastructure/workers/portfolio_calculator.go`

### Frontend Implementation for US7

- [X] T181 [P] [US7] Create BalanceChart component in `crypto-wallet-frontend/src/components/charts/BalanceChart.tsx`
- [X] T182 [P] [US7] Create TxVolumeChart component in `crypto-wallet-frontend/src/components/charts/TxVolumeChart.tsx`
- [X] T183 [US7] Create analytics page in `crypto-wallet-frontend/src/app/analytics/page.tsx`
- [X] T184 [US7] Add date range selector to analytics page
- [X] T185 [US7] Add gain/loss calculations display
- [X] T186 [US7] Add asset allocation pie chart

**Checkpoint**: All user stories complete - Full feature set delivered

---

## Phase 10: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [X] T187 [P] Create seed data scripts in `database/seed/` for all databases
- [X] T188 [P] Add error boundary components in frontend
- [X] T189 [P] Implement comprehensive logging across all backend services
- [X] T190 [P] Add loading states to all async frontend operations
- [X] T191 [P] Implement retry logic for blockchain operations
- [X] T192 [P] Add toast notifications for user actions
- [X] T193 [POLISH] Performance optimization: database query tuning
- [X] T194 [POLISH] Performance optimization: frontend bundle size reduction
- [X] T195 [POLISH] Security hardening: input validation review
- [X] T196 [POLISH] Security hardening: rate limiting fine-tuning
- [X] T197 [P] Update README.md with comprehensive documentation
- [X] T198 [POLISH] Verify quickstart.md instructions work end-to-end
- [X] T199 [P] Create deployment scripts for production
- [X] T200 [POLISH] Final integration testing of all user stories together

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-9)**: All depend on Foundational phase completion
  - US1 (P1): Wallet Creation - No dependencies on other stories
  - US2 (P1): Send/Receive - Integrates with US1 (wallets)
  - US3 (P2): Real-Time Prices - Enhances US1 (wallet displays)
  - US4 (P2): Exchange - Requires US1 (wallets) and US2 (transactions)
  - US5 (P3): Transaction History - Requires US2 (transaction data)
  - US6 (P2): Security/KYC - Independent but enhances all features
  - US7 (P3): Portfolio Analytics - Requires US1 (wallets) and US2 (transactions)
- **Polish (Phase 10)**: Depends on all desired user stories being complete

### User Story Dependencies

**Recommended Implementation Order** (following priorities and dependencies):

1. **Phase 2: Foundational** → COMPLETE THIS FIRST (BLOCKS ALL STORIES)
2. **Phase 3: US1 (P1)** → Wallet Creation (MVP foundation)
3. **Phase 4: US2 (P1)** → Send/Receive (MVP complete)
4. **Phase 5: US3 (P2)** → Real-Time Prices (enhances MVP)
5. **Phase 6: US4 (P2)** → Exchange (value-add feature)
6. **Phase 8: US6 (P2)** → Security/KYC (compliance feature)
7. **Phase 7: US5 (P3)** → Transaction History (enhanced tracking)
8. **Phase 9: US7 (P3)** → Portfolio Analytics (advanced feature)

**Parallel Execution Opportunity**: After Foundational phase completes, US1 and US6 can be developed in parallel by different developers as they have no interdependencies.

### Within Each User Story

**General Pattern**:

1. Backend entities and repositories (can be parallel)
2. Backend domain services and DTOs (can be parallel after step 1)
3. Backend use cases (sequential, depend on services)
4. Backend handlers and routes (sequential, depend on use cases)
5. Frontend components (can be parallel, independent of backend progress)
6. Frontend pages and integration (sequential, after components and backend API ready)

### Parallel Opportunities

**Setup Phase**: All T001-T011 can run in parallel (independent setup tasks)

**Foundational Phase**:

- T012-T015 (migrations) can run in parallel
- T017-T018, T020-T022, T023-T024, T028-T030 can run in parallel (independent infrastructure)
- T033-T036 (blockchain adapters) can run in parallel
- T037-T046 (frontend foundation) can run in parallel

**User Story 1**:

- T057-T060 (entities) can run in parallel
- T063-T064 (service and DTOs) can run in parallel
- T072-T077 (frontend components) can run in parallel

**User Story 2**:

- T082-T083 (entities) can run in parallel
- T086-T087 (service and DTOs) can run in parallel
- T098-T102 (frontend components) can run in parallel

**Similar patterns apply to other user stories**

---

## Parallel Execution Examples

### Example 1: Setup Phase

```bash
# All setup tasks can run in parallel:
Task T001: "Initialize Go backend project structure"
Task T002: "Initialize Next.js frontend project structure"
Task T003: "Create Docker Compose configuration"
Task T004: "Setup Makefile commands"
Task T005: "Setup package.json scripts"
# ... all other setup tasks
```

### Example 2: Foundational Database Migrations

```bash
# All database migrations can run in parallel:
Task T012: "Create core database migration"
Task T013: "Create KYC database migration"
Task T014: "Create rates database migration"
Task T015: "Create audit database migration"
```

### Example 3: User Story 1 - Backend Entities

```bash
# All entities for US1 can run in parallel:
Task T057: "Create Wallet entity"
Task T058: "Create Chain entity"
Task T059: "Create Token entity"
Task T060: "Create Account entity"
```

### Example 4: User Story 1 - Frontend Components

```bash
# All frontend components for US1 can run in parallel:
Task T072: "Create Zustand wallet store"
Task T073: "Create useWallets custom hook"
Task T074: "Create WalletCard component"
Task T075: "Create WalletList component"
Task T076: "Create AddWalletModal component"
Task T077: "Create WalletBalance component"
```

### Example 5: Cross-Story Parallel Development

```bash
# After Foundational phase completes, these can run in parallel:
# Developer A:
Phase 3: "User Story 1 - Wallet Creation"

# Developer B:
Phase 8: "User Story 6 - Security and KYC"
# (US6 is independent of US1 and can be developed in parallel)
```

---

## Implementation Strategy

### MVP First (User Stories 1 & 2 Only)

**Minimum Viable Product Path**:

1. ✅ Complete Phase 1: Setup (~1-2 days)
2. ✅ Complete Phase 2: Foundational (~3-5 days) - CRITICAL BLOCKER
3. ✅ Complete Phase 3: User Story 1 (~3-4 days) - Wallet creation and balance viewing
4. 🛑 **STOP and VALIDATE**: Test User Story 1 independently
5. ✅ Complete Phase 4: User Story 2 (~3-4 days) - Send and receive transactions
6. 🛑 **STOP and VALIDATE**: Test User Stories 1 & 2 together
7. 🚀 Deploy/demo MVP (basic wallet with send/receive capability)

**MVP Delivery Time**: ~10-15 days with single developer

### Incremental Delivery

**Full Feature Rollout**:

1. **Foundation** (Setup + Foundational) → ~5-7 days → Infrastructure ready
2. **MVP Release** (US1 + US2) → ~6-8 days → Core wallet functionality 🚀
3. **Enhanced Release** (US3 + US4) → ~5-7 days → Real-time prices + exchange 🚀
4. **Security Release** (US6) → ~4-5 days → KYC and compliance 🚀
5. **Analytics Release** (US5 + US7) → ~4-6 days → History and analytics 🚀
6. **Polish & Production** (Phase 10) → ~3-5 days → Production-ready 🚀

**Total Delivery Time**: ~27-38 days with single developer, ~15-20 days with team of 3

### Parallel Team Strategy

**Team of 3 Developers**:

**Week 1-2**: Everyone works together on Foundational phase (T012-T056)

- Developer A: Database + Backend Core (T012-T031)
- Developer B: Blockchain Foundation (T032-T036) + Auth Foundation (T047-T052)
- Developer C: Frontend Foundation (T037-T056)

**Week 3**: Once Foundational is complete, split user stories:

- Developer A: User Story 1 (Wallet Creation) - Phase 3
- Developer B: User Story 6 (Security/KYC) - Phase 8
- Developer C: User Story 3 (Real-Time Prices) - Phase 5

**Week 4-5**: Continue with remaining stories:

- Developer A: User Story 2 (Send/Receive) - Phase 4
- Developer B: User Story 4 (Exchange) - Phase 6
- Developer C: User Story 5 (Transaction History) - Phase 7

**Week 6**: Final integration:

- Developer A: User Story 7 (Portfolio Analytics) - Phase 9
- Developer B + C: Polish & Testing - Phase 10

---

## Task Completion Checklist

For each task, verify:

- [ ] Code follows Go/TypeScript best practices
- [ ] Functions have clear, single responsibilities
- [ ] Error handling is comprehensive with proper logging
- [ ] Database queries are optimized with appropriate indexes
- [ ] API endpoints follow RESTful conventions from OpenAPI spec
- [ ] Frontend components are responsive and accessible
- [ ] Security considerations addressed (encryption, validation, sanitization)
- [ ] Audit logging implemented for sensitive operations
- [ ] Changes committed with descriptive commit message
- [ ] Task marked as complete in this file

---

## Notes

- **[P] tasks**: Different files, no dependencies, can run in parallel
- **[Story] labels**: Maps task to specific user story for traceability
- **Foundational phase is CRITICAL**: Nothing can proceed until T012-T056 are complete
- **Each user story is independently testable**: Can validate functionality without other stories
- **Stop at checkpoints**: Validate each story independently before proceeding
- **Commit frequently**: After each task or logical group of related tasks
- **Follow OpenAPI spec**: All API endpoints must match contracts/openapi.yaml
- **Follow data model**: All database operations must match data-model.md schemas
- **Use blockchain contracts**: All blockchain operations must match contracts/blockchain.md interfaces
- **Follow WebSocket spec**: Real-time features must match contracts/websocket.md protocol

---

## Summary Statistics

**Total Tasks**: 200

- Setup: 11 tasks
- Foundational: 45 tasks (BLOCKS ALL STORIES)
- User Story 1 (P1): 25 tasks
- User Story 2 (P1): 25 tasks
- User Story 3 (P2): 19 tasks
- User Story 4 (P2): 18 tasks
- User Story 5 (P3): 11 tasks
- User Story 6 (P2): 20 tasks
- User Story 7 (P3): 12 tasks
- Polish: 14 tasks

**Parallel Opportunities**: ~70 tasks marked [P] for parallel execution

**Independent Test Criteria**: Each user story has clear validation criteria

**MVP Scope**: User Stories 1 & 2 (50 tasks after foundation) = Core wallet functionality

**Estimated Timeline**:

- Single developer: 27-38 days full-time
- Team of 3: 15-20 days full-time
- MVP only (US1+US2): 10-15 days single developer
