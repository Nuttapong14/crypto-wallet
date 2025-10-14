# 🎯 Project Goal

Build a **Full-Stack Multi-Chain Crypto Wallet & Transaction Aggregator** using **Golang (Backend)** and **Next.js (Frontend)**, designed with **DDD + Hexagonal Architecture (Hexagonal)** principles.
The system must securely manage multi-chain wallets (BTC, ETH, SOL, XLM), transactions, KYC/AML, exchange rates, and event-based auditing.

---

## 🧱 1. Architecture Requirements

- Follow **Domain Driven Design (DDD)** + **Hexagonal Architecture**
- Layer separation:
  - `/internal/domain` → Entities, Repositories, Value Objects
  - `/internal/application` → Use Cases (wallet, tx, swap)
  - `/internal/infrastructure` → Database, Blockchain Adapters, Encryption
  - `/internal/interfaces/http` → REST Handlers & Middleware
- Support 4 PostgreSQL databases:
  1. `core_db` — users, wallets, transactions, chains, tokens, ledger_entries
  2. `kyc_db` — kyc_profiles, documents, risk_score, alert_rules
  3. `rates_db` — exchange_rates, trading_pairs, price_history, summary
  4. `audit_db` — audit_logs, api_audit, security_logs
- Containerized with **Docker Compose**
- Configs via `.env` + `/configs/config.yaml`
- Support hot reload via `air` or `CompileDaemon`

---

## 🪙 2. Blockchain Integration Layer

### Supported Chains

| Chain    | Library                    | Protocol           |
| -------- | -------------------------- | ------------------ |
| Bitcoin  | `btcsuite/btcd`, `btcutil` | JSON-RPC           |
| Ethereum | `go-ethereum`              | JSON-RPC           |
| Solana   | `portto/solana-go-sdk`     | JSON-RPC           |
| Stellar  | `stellar/go`               | REST (Horizon API) |

### Requirements

- Implement unified interface `BlockchainAdapter`:
  ```go
  type BlockchainAdapter interface {
      GetBalance(address string) (float64, error)
      GetTransactions(address string) ([]domain.Transaction, error)
      SendTransaction(tx domain.Transaction) (string, error)
  }
  ```
- Create concrete adapters:
  `/internal/infrastructure/blockchain/{bitcoin,ethereum,solana,stellar}_adapter.go`
- Configurable RPC endpoints via `configs/blockchain.yaml`
- Include **mock adapters** for testing

---

## 🧩 3. Core Features

### Wallet System

- Hierarchical Deterministic (HD) wallets using **BIP32/BIP44**
- Secure key encryption (AES-GCM or NaCl Secretbox)
- `wallet_addresses` table for address derivations
- Generate new addresses per transaction

### Transaction Handling

- Store and broadcast transactions via blockchain adapters
- Implement **Transaction Queue & Worker System**:
  - `tx_queue` table
  - Background worker (Go routine or cron)
  - Poll node status and mark `pending`, `confirmed`, `failed`

### Exchange & Rates

- `ExchangeService` with off-chain swap (mock)
- Update live rates via CoinGecko/Binance APIs
- Maintain tables: `exchange_rates`, `trading_pairs`, `price_history`
- Support rate caching + WebSocket updates (optional)

### Event Logging

- All actions (wallet create, tx send, KYC, swap) → send to Event Bus (Kafka/NATS)
- Write immutable audit records in `audit_db`
- Asynchronous worker for log persistence

---

## 🔐 4. Security & Compliance

- Passwords hashed via `bcrypt`
- Private keys encrypted using `AES-GCM`
- JWT-based authentication (`github.com/golang-jwt/jwt/v5`)
- Support external KMS/Vault integration (optional)
- KYC & AML in `kyc_db` with basic rule engine
- GDPR-compliant masking for PII

---

## 💾 5. Infrastructure & DevOps

### Docker Compose Stack

Include:

- `api` → Go service
- `nextjs` → Frontend
- `postgres` → with 4 DBs + `pgadmin`
- `n8n` → workflow automation (rate updater, notifier)
- `prometheus`, `grafana` → metrics & visualization

### Monitoring

- Custom Prometheus metrics:
  - API latency
  - RPC response time
  - Worker queue depth
- Grafana dashboards for all metrics

---

## 💻 6. Frontend Requirements (Next.js 15 + TypeScript)

- Wallet dashboard (balance, transactions, swap)
- Auth (JWT-based)
- Real-time status updates via polling or WebSocket
- Design system:
  - TailwindCSS + ShadCN/UI
  - Recharts for analytics
  - Framer Motion animations
- Connect to backend via `/api/v1/*` endpoints

---

## 🧠 7. Folder Structure Example

```
crypto-wallet-system/
│
├── backend/
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── domain/
│   │   ├── application/
│   │   ├── infrastructure/
│   │   │   ├── blockchain/
│   │   │   ├── repository/
│   │   │   └── security/
│   │   └── interfaces/http/
│   ├── configs/
│   ├── db/migrations/
│   ├── pkg/utils/
│   └── go.mod
│
├── frontend/
│   ├── app/
│   ├── components/
│   ├── lib/
│   └── package.json
│
├── docker-compose.yml
└── Makefile
```

---

## 📡 8. Endpoints Summary

| Endpoint                      | Method   | Description                 |
| ----------------------------- | -------- | --------------------------- |
| `/api/v1/auth/login`          | POST     | Authenticate user           |
| `/api/v1/wallets`             | GET/POST | Create/list user wallets    |
| `/api/v1/wallets/:id/balance` | GET      | Query wallet balance        |
| `/api/v1/transactions`        | POST     | Broadcast a new transaction |
| `/api/v1/transactions/:id`    | GET      | Fetch transaction details   |
| `/api/v1/exchange/swap`       | POST     | Off-chain swap execution    |
| `/api/v1/kyc/upload`          | POST     | Upload KYC documents        |
| `/api/v1/rates`               | GET      | Get exchange rates          |
| `/api/v1/notifications`       | GET      | Fetch latest system events  |

---

## 🧾 9. Deliverables Claude Should Generate

1. 📂 Folder structure (backend + frontend)
2. 🐳 `docker-compose.yml` for all services
3. ⚙️ Go backend with DDD structure
4. 🧩 Blockchain adapter implementations (BTC/ETH/SOL/XLM)
5. 🧱 PostgreSQL schema (all 4 DBs)
6. 🔐 Encryption + JWT middleware
7. 🔄 Transaction Queue Worker
8. 💱 ExchangeService (mock swap)
9. 🧾 Audit & Event system (Kafka/NATS optional)
10. 💻 Next.js Frontend UI (wallet dashboard)
11. 📈 Prometheus metrics & Grafana dashboards

---

## 🧩 10. Tone & Output Style

- Use **Go Fiber or Echo** for REST API.
- Write **Hexagonal, idiomatic Go code**.
- Add meaningful comments and domain-driven naming.
- Create minimal but working code stubs (runnable via Docker).
- Prioritize correctness, modularity, and clarity over brevity.
- If unsure, generate **scaffold first**, then detail each layer iteratively.

---

🧠 **In short:**

> Generate the full project scaffold (Go backend + Next.js frontend)
> using DDD + Hexagonal Architecture
> for a 4-chain crypto wallet & transaction aggregator
> with complete database, adapters, and security layers.

Start with backend scaffold and database migration first.
