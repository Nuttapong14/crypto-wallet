# Database Migration Research for Multi-Database Go Application

## Executive Summary

**Recommended Tool**: **pressly/goose**

For managing 4 separate PostgreSQL databases (core_db, kyc_db, rates_db, audit_db) in a Go application, pressly/goose is the optimal choice due to:
- Dual support for SQL and Go-based migrations (essential for complex seed data)
- Out-of-order migration support (helpful when teams work on different databases)
- Embedded migrations capability (deploy single binary with migrations included)
- Active maintenance with 3.2k GitHub stars
- Simple CLI and library interface
- Excellent PostgreSQL support via pgx/v5

---

## 1. Migration Tool Comparison

| Feature | golang-migrate | pressly/goose | sql-migrate |
|---------|----------------|---------------|-------------|
| **GitHub Stars** | 10.3k | 3.2k | 3.1k |
| **Last Updated** | Active (2024) | Active (2024) | Active (2024) |
| **Language** | Go | Go | Go |
| **Migration Types** | SQL only | SQL + Go functions | SQL only |
| **Database Support** | 15+ databases | 10+ databases | 5 databases |
| **PostgreSQL Support** | ✅ Excellent | ✅ Excellent (pgx/v5) | ✅ Good |
| **CLI Tool** | ✅ Yes | ✅ Yes | ✅ Yes |
| **Library Usage** | ✅ Yes | ✅ Yes | ✅ Yes |
| **Embedded Migrations** | ✅ Yes | ✅ Yes | ❌ Limited |
| **Versioning** | Timestamp/Sequential | Timestamp/Sequential | Timestamp |
| **Out-of-Order** | ❌ No | ✅ Yes | ❌ No |
| **Rollback** | Down migrations | Down migrations | Down migrations |
| **Transaction Support** | ✅ Yes | ✅ Yes | ✅ Yes |
| **Go Migrations** | ❌ No | ✅ Yes | ❌ No |
| **Multi-DB Support** | Manual (separate runs) | Manual (separate runs) | Manual (separate runs) |
| **Learning Curve** | Medium | Low | Medium |
| **CI/CD Integration** | Excellent | Excellent | Good |

### Detailed Analysis

#### golang-migrate/migrate
**Pros**:
- Most popular (10.3k stars)
- Broadest database driver support (15+ databases)
- Multiple migration source support (filesystem, GitHub, AWS S3, etc.)
- Excellent for language-agnostic teams
- Strong CI/CD integration
- Well-documented

**Cons**:
- SQL-only migrations (no Go functions)
- No out-of-order migration support
- More complex setup
- Cannot embed complex seed logic

**Best For**: Teams needing maximum database compatibility or language-agnostic tooling

#### pressly/goose
**Pros**:
- Go function migrations (powerful for seed data)
- Out-of-order migration support
- Embedded migrations in binary
- Simple, minimal setup
- Both CLI and library usage
- pgx/v5 pool support
- Atlas integration for declarative migrations

**Cons**:
- Fewer database drivers than golang-migrate
- Smaller community than golang-migrate
- Less flexibility in migration sources

**Best For**: Go-native projects needing complex seed logic and flexible migration workflows

#### sql-migrate
**Pros**:
- Clean API
- Simple configuration
- Good for basic migration needs
- Go-migrate syntax compatible

**Cons**:
- Forked from old goose (pre-Pressly)
- No Go migrations
- Limited database support (5 databases)
- Smaller feature set
- Less active development

**Best For**: Simple projects with basic migration needs

---

## 2. Recommended Architecture for 4 Databases

### Project Structure

```
crypto-wallet/
├── cmd/
│   └── migrate/
│       └── main.go                 # Migration CLI tool
├── internal/
│   └── database/
│       ├── migrations/
│       │   ├── core_db/            # Core application database
│       │   │   ├── 00001_init.sql
│       │   │   ├── 00002_users.sql
│       │   │   └── seeds/
│       │   │       └── 00010_seed_admin.go
│       │   ├── kyc_db/             # KYC verification database
│       │   │   ├── 00001_init.sql
│       │   │   ├── 00002_kyc_documents.sql
│       │   │   └── seeds/
│       │   │       └── 00010_seed_test_users.go
│       │   ├── rates_db/           # Exchange rates database
│       │   │   ├── 00001_init.sql
│       │   │   ├── 00002_exchange_rates.sql
│       │   │   └── seeds/
│       │   │       └── 00010_seed_currency_pairs.go
│       │   └── audit_db/           # Audit log database
│       │       ├── 00001_init.sql
│       │       ├── 00002_audit_logs.sql
│       │       └── seeds/
│       │           └── 00010_seed_audit_types.go
│       └── migrate.go              # Migration helper functions
├── scripts/
│   ├── migrate-all.sh              # Migrate all databases
│   ├── migrate-up.sh               # Run up migrations
│   ├── migrate-down.sh             # Run down migrations
│   └── seed-dev.sh                 # Seed development data
├── docker-compose.yml              # Local development databases
├── Makefile                        # Migration commands
└── .github/
    └── workflows/
        └── migrations.yml          # CI/CD pipeline
```

### Key Design Decisions

1. **Separate Folders per Database**: Each database has its own migration folder for clear separation of concerns
2. **Seed Subfolder**: Seed migrations use higher version numbers (00010+) and Go migrations for complex logic
3. **Independent Versioning**: Each database maintains its own version sequence
4. **Shared Tooling**: Common migration utilities in `internal/database/migrate.go`
5. **Script Automation**: Shell scripts for common operations across all databases

---

## 3. Migration File Examples

### Example: core_db Schema Migration

**File**: `internal/database/migrations/core_db/00001_init.sql`

```sql
-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_created_at ON users(created_at);

-- Trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS users;
DROP EXTENSION IF EXISTS "uuid-ossp";
-- +goose StatementEnd
```

**File**: `internal/database/migrations/core_db/00002_wallets.sql`

```sql
-- +goose Up
-- +goose StatementBegin
CREATE TYPE wallet_status AS ENUM ('active', 'suspended', 'closed');
CREATE TYPE currency_code AS ENUM ('BTC', 'ETH', 'USDT', 'USD', 'EUR');

CREATE TABLE wallets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    currency currency_code NOT NULL,
    balance DECIMAL(20, 8) NOT NULL DEFAULT 0 CHECK (balance >= 0),
    status wallet_status NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, currency)
);

CREATE INDEX idx_wallets_user_id ON wallets(user_id);
CREATE INDEX idx_wallets_currency ON wallets(currency);
CREATE INDEX idx_wallets_status ON wallets(status);

CREATE TRIGGER update_wallets_updated_at
    BEFORE UPDATE ON wallets
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_wallets_updated_at ON wallets;
DROP TABLE IF EXISTS wallets;
DROP TYPE IF EXISTS currency_code;
DROP TYPE IF EXISTS wallet_status;
-- +goose StatementEnd
```

### Example: kyc_db Schema Migration

**File**: `internal/database/migrations/kyc_db/00001_init.sql`

```sql
-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE kyc_status AS ENUM ('pending', 'under_review', 'approved', 'rejected', 'expired');
CREATE TYPE document_type AS ENUM ('passport', 'national_id', 'drivers_license', 'proof_of_address');

CREATE TABLE kyc_submissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,  -- Foreign key reference to core_db.users
    status kyc_status NOT NULL DEFAULT 'pending',
    submitted_at TIMESTAMP NOT NULL DEFAULT NOW(),
    reviewed_at TIMESTAMP,
    reviewed_by UUID,
    rejection_reason TEXT,
    expires_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE kyc_documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    submission_id UUID NOT NULL REFERENCES kyc_submissions(id) ON DELETE CASCADE,
    document_type document_type NOT NULL,
    file_url TEXT NOT NULL,
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_kyc_submissions_user_id ON kyc_submissions(user_id);
CREATE INDEX idx_kyc_submissions_status ON kyc_submissions(status);
CREATE INDEX idx_kyc_documents_submission_id ON kyc_documents(submission_id);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_kyc_submissions_updated_at
    BEFORE UPDATE ON kyc_submissions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_kyc_documents_updated_at
    BEFORE UPDATE ON kyc_documents
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_kyc_documents_updated_at ON kyc_documents;
DROP TRIGGER IF EXISTS update_kyc_submissions_updated_at ON kyc_submissions;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS kyc_documents;
DROP TABLE IF EXISTS kyc_submissions;
DROP TYPE IF EXISTS document_type;
DROP TYPE IF EXISTS kyc_status;
DROP EXTENSION IF EXISTS "uuid-ossp";
-- +goose StatementEnd
```

### Example: Go Migration for Seed Data

**File**: `internal/database/migrations/core_db/seeds/00010_seed_admin.go`

```go
package migrations

import (
    "context"
    "database/sql"
    "fmt"

    "github.com/pressly/goose/v3"
    "golang.org/x/crypto/bcrypt"
)

func init() {
    goose.AddMigrationContext(upSeedAdmin, downSeedAdmin)
}

func upSeedAdmin(ctx context.Context, tx *sql.Tx) error {
    // Only seed in development environment
    if isDevelopment() {
        // Hash password for admin user
        passwordHash, err := bcrypt.GenerateFromPassword(
            []byte("admin123"),
            bcrypt.DefaultCost,
        )
        if err != nil {
            return fmt.Errorf("failed to hash password: %w", err)
        }

        // Insert admin user
        query := `
            INSERT INTO users (email, password_hash, full_name)
            VALUES ($1, $2, $3)
            ON CONFLICT (email) DO NOTHING
        `
        _, err = tx.ExecContext(
            ctx,
            query,
            "admin@cryptowallet.com",
            string(passwordHash),
            "System Administrator",
        )
        if err != nil {
            return fmt.Errorf("failed to insert admin user: %w", err)
        }

        fmt.Println("✓ Seeded admin user: admin@cryptowallet.com")
    }

    return nil
}

func downSeedAdmin(ctx context.Context, tx *sql.Tx) error {
    // Remove seed data
    query := `DELETE FROM users WHERE email = $1`
    _, err := tx.ExecContext(ctx, query, "admin@cryptowallet.com")
    if err != nil {
        return fmt.Errorf("failed to remove admin user: %w", err)
    }

    fmt.Println("✓ Removed admin user seed data")
    return nil
}

func isDevelopment() bool {
    // Check environment variable
    env := os.Getenv("APP_ENV")
    return env == "development" || env == "dev" || env == ""
}
```

**File**: `internal/database/migrations/rates_db/seeds/00010_seed_currency_pairs.go`

```go
package migrations

import (
    "context"
    "database/sql"
    "fmt"
    "time"

    "github.com/pressly/goose/v3"
)

func init() {
    goose.AddMigrationContext(upSeedCurrencyPairs, downSeedCurrencyPairs)
}

func upSeedCurrencyPairs(ctx context.Context, tx *sql.Tx) error {
    // Define initial currency pairs with mock rates
    pairs := []struct {
        Base   string
        Quote  string
        Rate   float64
        Source string
    }{
        {"BTC", "USD", 42000.00, "coinbase"},
        {"ETH", "USD", 2200.00, "coinbase"},
        {"USDT", "USD", 1.00, "binance"},
        {"BTC", "EUR", 38000.00, "kraken"},
        {"ETH", "EUR", 2000.00, "kraken"},
    }

    query := `
        INSERT INTO exchange_rates (base_currency, quote_currency, rate, source, timestamp)
        VALUES ($1, $2, $3, $4, $5)
        ON CONFLICT (base_currency, quote_currency, source)
        DO UPDATE SET rate = EXCLUDED.rate, timestamp = EXCLUDED.timestamp
    `

    now := time.Now()
    for _, pair := range pairs {
        _, err := tx.ExecContext(
            ctx,
            query,
            pair.Base,
            pair.Quote,
            pair.Rate,
            pair.Source,
            now,
        )
        if err != nil {
            return fmt.Errorf("failed to insert rate %s/%s: %w",
                pair.Base, pair.Quote, err)
        }
    }

    fmt.Printf("✓ Seeded %d currency pairs\n", len(pairs))
    return nil
}

func downSeedCurrencyPairs(ctx context.Context, tx *sql.Tx) error {
    query := `TRUNCATE TABLE exchange_rates CASCADE`
    _, err := tx.ExecContext(ctx, query)
    if err != nil {
        return fmt.Errorf("failed to truncate exchange_rates: %w", err)
    }

    fmt.Println("✓ Removed currency pair seed data")
    return nil
}
```

---

## 4. Migration Helper Code

**File**: `internal/database/migrate.go`

```go
package database

import (
    "database/sql"
    "embed"
    "fmt"
    "log"

    "github.com/pressly/goose/v3"
    _ "github.com/lib/pq"
)

//go:embed migrations/core_db/*.sql
var coreDBMigrations embed.FS

//go:embed migrations/kyc_db/*.sql
var kycDBMigrations embed.FS

//go:embed migrations/rates_db/*.sql
var ratesDBMigrations embed.FS

//go:embed migrations/audit_db/*.sql
var auditDBMigrations embed.FS

type DatabaseConfig struct {
    Name        string
    DSN         string
    Migrations  embed.FS
    MigrationDir string
}

var Databases = []DatabaseConfig{
    {
        Name:        "core_db",
        DSN:         getEnv("CORE_DB_DSN", "postgres://user:pass@localhost:5432/core_db?sslmode=disable"),
        Migrations:  coreDBMigrations,
        MigrationDir: "migrations/core_db",
    },
    {
        Name:        "kyc_db",
        DSN:         getEnv("KYC_DB_DSN", "postgres://user:pass@localhost:5432/kyc_db?sslmode=disable"),
        Migrations:  kycDBMigrations,
        MigrationDir: "migrations/kyc_db",
    },
    {
        Name:        "rates_db",
        DSN:         getEnv("RATES_DB_DSN", "postgres://user:pass@localhost:5432/rates_db?sslmode=disable"),
        Migrations:  ratesDBMigrations,
        MigrationDir: "migrations/rates_db",
    },
    {
        Name:        "audit_db",
        DSN:         getEnv("AUDIT_DB_DSN", "postgres://user:pass@localhost:5432/audit_db?sslmode=disable"),
        Migrations:  auditDBMigrations,
        MigrationDir: "migrations/audit_db",
    },
}

// MigrateUp runs all pending migrations for a specific database
func MigrateUp(dbName string) error {
    config := findDatabaseConfig(dbName)
    if config == nil {
        return fmt.Errorf("database %s not found", dbName)
    }

    db, err := sql.Open("postgres", config.DSN)
    if err != nil {
        return fmt.Errorf("failed to connect to %s: %w", dbName, err)
    }
    defer db.Close()

    goose.SetBaseFS(config.Migrations)

    if err := goose.SetDialect("postgres"); err != nil {
        return fmt.Errorf("failed to set dialect: %w", err)
    }

    if err := goose.Up(db, config.MigrationDir); err != nil {
        return fmt.Errorf("failed to run migrations for %s: %w", dbName, err)
    }

    log.Printf("✓ Successfully migrated %s", dbName)
    return nil
}

// MigrateDown rolls back the last migration for a specific database
func MigrateDown(dbName string) error {
    config := findDatabaseConfig(dbName)
    if config == nil {
        return fmt.Errorf("database %s not found", dbName)
    }

    db, err := sql.Open("postgres", config.DSN)
    if err != nil {
        return fmt.Errorf("failed to connect to %s: %w", dbName, err)
    }
    defer db.Close()

    goose.SetBaseFS(config.Migrations)

    if err := goose.SetDialect("postgres"); err != nil {
        return fmt.Errorf("failed to set dialect: %w", err)
    }

    if err := goose.Down(db, config.MigrationDir); err != nil {
        return fmt.Errorf("failed to rollback migration for %s: %w", dbName, err)
    }

    log.Printf("✓ Successfully rolled back %s", dbName)
    return nil
}

// MigrateAllUp runs migrations for all databases
func MigrateAllUp() error {
    for _, config := range Databases {
        if err := MigrateUp(config.Name); err != nil {
            return err
        }
    }
    return nil
}

// MigrateStatus shows migration status for a database
func MigrateStatus(dbName string) error {
    config := findDatabaseConfig(dbName)
    if config == nil {
        return fmt.Errorf("database %s not found", dbName)
    }

    db, err := sql.Open("postgres", config.DSN)
    if err != nil {
        return fmt.Errorf("failed to connect to %s: %w", dbName, err)
    }
    defer db.Close()

    goose.SetBaseFS(config.Migrations)

    if err := goose.SetDialect("postgres"); err != nil {
        return fmt.Errorf("failed to set dialect: %w", err)
    }

    if err := goose.Status(db, config.MigrationDir); err != nil {
        return fmt.Errorf("failed to get status for %s: %w", dbName, err)
    }

    return nil
}

func findDatabaseConfig(name string) *DatabaseConfig {
    for _, config := range Databases {
        if config.Name == name {
            return &config
        }
    }
    return nil
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

**File**: `cmd/migrate/main.go`

```go
package main

import (
    "flag"
    "log"
    "os"

    "github.com/yourusername/crypto-wallet/internal/database"
)

func main() {
    var (
        command = flag.String("command", "up", "Command: up, down, status")
        dbName  = flag.String("db", "all", "Database name or 'all'")
    )
    flag.Parse()

    switch *command {
    case "up":
        if *dbName == "all" {
            if err := database.MigrateAllUp(); err != nil {
                log.Fatalf("Migration failed: %v", err)
            }
        } else {
            if err := database.MigrateUp(*dbName); err != nil {
                log.Fatalf("Migration failed: %v", err)
            }
        }
    case "down":
        if *dbName == "all" {
            log.Fatal("Cannot rollback all databases at once. Specify a database.")
        }
        if err := database.MigrateDown(*dbName); err != nil {
            log.Fatalf("Rollback failed: %v", err)
        }
    case "status":
        if *dbName == "all" {
            for _, config := range database.Databases {
                log.Printf("\n=== %s ===", config.Name)
                if err := database.MigrateStatus(config.Name); err != nil {
                    log.Printf("Status check failed: %v", err)
                }
            }
        } else {
            if err := database.MigrateStatus(*dbName); err != nil {
                log.Fatalf("Status check failed: %v", err)
            }
        }
    default:
        log.Fatalf("Unknown command: %s", *command)
    }
}
```

---

## 5. Rollback Strategy

### Rollback Principles

1. **Every Up Has a Down**: Every migration must have a corresponding rollback
2. **Transaction Safety**: Wrap migrations in transactions when possible
3. **Backup Before Rollback**: Always backup production data before rollback
4. **Test Rollbacks**: Test down migrations in staging before production
5. **Manual Review**: Critical rollbacks should be manually reviewed
6. **Limited Scope**: Roll back one database at a time, not all at once

### Rollback Scenarios

#### Scenario 1: Development Rollback (Safe)
```bash
# Roll back last migration on core_db
go run cmd/migrate/main.go -command=down -db=core_db

# Or using goose CLI
goose -dir internal/database/migrations/core_db postgres "DSN" down
```

#### Scenario 2: Production Rollback (Requires Backup)
```bash
# 1. Take backup
pg_dump -h localhost -U user -d core_db -F c -f backup_core_db_$(date +%Y%m%d_%H%M%S).dump

# 2. Verify backup
pg_restore --list backup_core_db_*.dump | head -20

# 3. Roll back migration
go run cmd/migrate/main.go -command=down -db=core_db

# 4. Verify database state
psql -h localhost -U user -d core_db -c "\dt"
```

#### Scenario 3: Multiple Migration Rollback
```bash
# Roll back to specific version using goose
goose -dir internal/database/migrations/core_db postgres "DSN" down-to 00005
```

#### Scenario 4: Emergency Rollback with Restore
```bash
# If rollback fails, restore from backup
dropdb core_db
createdb core_db
pg_restore -h localhost -U user -d core_db backup_core_db_*.dump
```

### Rollback Safety Checklist

```markdown
## Pre-Rollback Checklist
- [ ] Backup database (pg_dump)
- [ ] Verify backup integrity (pg_restore --list)
- [ ] Review down migration SQL
- [ ] Check for data loss in down migration
- [ ] Notify team of rollback window
- [ ] Stop application writes to database
- [ ] Document rollback reason

## Post-Rollback Checklist
- [ ] Verify schema state (psql \dt, \d table_name)
- [ ] Run smoke tests
- [ ] Check application logs
- [ ] Verify data integrity
- [ ] Resume application writes
- [ ] Document rollback outcome
- [ ] Update migration files if needed
```

### Handling Non-Reversible Migrations

Some migrations cannot be safely reversed (e.g., dropping columns with data). Handle these cases explicitly:

```sql
-- +goose Up
-- +goose StatementBegin
-- Add new column
ALTER TABLE users ADD COLUMN phone_number VARCHAR(20);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- WARNING: This will cause data loss!
-- Consider creating a backup table before running this migration
DO $$
BEGIN
    RAISE NOTICE 'WARNING: Dropping phone_number column will cause data loss!';
    RAISE NOTICE 'Ensure you have a backup before proceeding.';
END $$;

ALTER TABLE users DROP COLUMN IF EXISTS phone_number;
-- +goose StatementEnd
```

---

## 6. Seed Data Management Strategy

### Separation of Concerns

1. **Schema Migrations**: Versions 00001-00009 (structural changes only)
2. **Seed Migrations**: Versions 00010+ (data seeding)
3. **Environment-Aware**: Seeds only run in development/staging
4. **Idempotent**: Seeds can be run multiple times safely

### Seed Data Patterns

#### Pattern 1: SQL Seeds (Simple Data)

```sql
-- +goose Up
-- +goose StatementBegin
INSERT INTO currency_pairs (base, quote, enabled) VALUES
    ('BTC', 'USD', true),
    ('ETH', 'USD', true),
    ('USDT', 'USD', true)
ON CONFLICT (base, quote) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM currency_pairs WHERE base IN ('BTC', 'ETH', 'USDT');
-- +goose StatementEnd
```

#### Pattern 2: Go Seeds (Complex Data)

Use Go migrations for:
- Password hashing
- Complex calculations
- External API calls
- Conditional logic
- Large datasets

Example: See `00010_seed_admin.go` above

#### Pattern 3: Conditional Seeds

```go
func upSeedData(ctx context.Context, tx *sql.Tx) error {
    env := os.Getenv("APP_ENV")

    switch env {
    case "development", "dev":
        return seedDevelopmentData(ctx, tx)
    case "staging":
        return seedStagingData(ctx, tx)
    case "production":
        // No seeding in production
        log.Println("Skipping seed data in production")
        return nil
    default:
        return seedDevelopmentData(ctx, tx)
    }
}
```

### Seed Data Best Practices

1. **Idempotency**: Use `ON CONFLICT DO NOTHING` or `INSERT ... WHERE NOT EXISTS`
2. **Environment Guards**: Never seed production automatically
3. **Realistic Data**: Use realistic data for development testing
4. **Versioning**: Keep seeds in sync with schema versions
5. **Documentation**: Document what each seed provides

---

## 7. CI/CD Integration

### GitHub Actions Workflow

**File**: `.github/workflows/migrations.yml`

```yaml
name: Database Migrations

on:
  pull_request:
    paths:
      - 'internal/database/migrations/**'
      - '.github/workflows/migrations.yml'
  push:
    branches:
      - main
      - develop

jobs:
  test-migrations:
    name: Test Migrations
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:15-alpine
        env:
          POSTGRES_PASSWORD: testpass
          POSTGRES_USER: testuser
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'
          cache: true

      - name: Install goose
        run: |
          go install github.com/pressly/goose/v3/cmd/goose@latest

      - name: Create test databases
        env:
          PGPASSWORD: testpass
        run: |
          psql -h localhost -U testuser -d postgres -c "CREATE DATABASE core_db;"
          psql -h localhost -U testuser -d postgres -c "CREATE DATABASE kyc_db;"
          psql -h localhost -U testuser -d postgres -c "CREATE DATABASE rates_db;"
          psql -h localhost -U testuser -d postgres -c "CREATE DATABASE audit_db;"

      - name: Run migrations up
        env:
          CORE_DB_DSN: "postgres://testuser:testpass@localhost:5432/core_db?sslmode=disable"
          KYC_DB_DSN: "postgres://testuser:testpass@localhost:5432/kyc_db?sslmode=disable"
          RATES_DB_DSN: "postgres://testuser:testpass@localhost:5432/rates_db?sslmode=disable"
          AUDIT_DB_DSN: "postgres://testuser:testpass@localhost:5432/audit_db?sslmode=disable"
        run: |
          for db in core_db kyc_db rates_db audit_db; do
            echo "Migrating $db..."
            goose -dir internal/database/migrations/$db postgres "$CORE_DB_DSN" up
          done

      - name: Test rollback (down migrations)
        env:
          CORE_DB_DSN: "postgres://testuser:testpass@localhost:5432/core_db?sslmode=disable"
        run: |
          for db in core_db kyc_db rates_db audit_db; do
            echo "Testing rollback for $db..."
            goose -dir internal/database/migrations/$db postgres "$CORE_DB_DSN" down
            goose -dir internal/database/migrations/$db postgres "$CORE_DB_DSN" up
          done

      - name: Run integration tests
        env:
          CORE_DB_DSN: "postgres://testuser:testpass@localhost:5432/core_db?sslmode=disable"
          KYC_DB_DSN: "postgres://testuser:testpass@localhost:5432/kyc_db?sslmode=disable"
          RATES_DB_DSN: "postgres://testuser:testpass@localhost:5432/rates_db?sslmode=disable"
          AUDIT_DB_DSN: "postgres://testuser:testpass@localhost:5432/audit_db?sslmode=disable"
        run: |
          go test -v ./internal/database/... -tags=integration

  validate-migration-naming:
    name: Validate Migration Files
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Validate migration file naming
        run: |
          # Check for proper naming convention
          find internal/database/migrations -type f -name "*.sql" | while read file; do
            basename=$(basename "$file")
            if ! [[ $basename =~ ^[0-9]{5}_[a-z_]+\.(up|down)\.sql$ ]]; then
              echo "ERROR: Invalid migration filename: $file"
              echo "Expected format: 00001_description.up.sql or 00001_description.down.sql"
              exit 1
            fi
          done

      - name: Check for orphaned migrations
        run: |
          # Ensure every .up.sql has a corresponding .down.sql
          for upfile in $(find internal/database/migrations -type f -name "*.up.sql"); do
            downfile="${upfile%.up.sql}.down.sql"
            if [ ! -f "$downfile" ]; then
              echo "ERROR: Missing down migration for $upfile"
              exit 1
            fi
          done
```

### Docker Integration

**File**: `Dockerfile` (Multi-stage build)

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install goose
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/migrate ./cmd/migrate

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install postgresql-client for pg_isready
RUN apk add --no-cache postgresql-client

# Copy binaries from builder
COPY --from=builder /app/server /app/server
COPY --from=builder /app/migrate /app/migrate
COPY --from=builder /go/bin/goose /usr/local/bin/goose

# Copy migrations
COPY internal/database/migrations /app/migrations

# Entrypoint script
COPY docker-entrypoint.sh /app/
RUN chmod +x /app/docker-entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["/app/docker-entrypoint.sh"]
CMD ["/app/server"]
```

**File**: `docker-entrypoint.sh`

```bash
#!/bin/sh
set -e

echo "Waiting for databases to be ready..."

# Wait for all databases
for db in core_db kyc_db rates_db audit_db; do
    until pg_isready -h postgres -p 5432 -U "${POSTGRES_USER}" -d "$db"; do
        echo "Waiting for $db..."
        sleep 2
    done
    echo "✓ $db is ready"
done

echo "Running migrations..."

# Run migrations for all databases
/app/migrate -command=up -db=all

echo "✓ All migrations completed"

# Execute the main command
exec "$@"
```

**File**: `docker-compose.yml`

```yaml
version: '3.9'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: cryptouser
      POSTGRES_PASSWORD: cryptopass
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/init-databases.sh:/docker-entrypoint-initdb.d/init-databases.sh
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U cryptouser"]
      interval: 10s
      timeout: 5s
      retries: 5

  app:
    build: .
    depends_on:
      postgres:
        condition: service_healthy
    environment:
      CORE_DB_DSN: "postgres://cryptouser:cryptopass@postgres:5432/core_db?sslmode=disable"
      KYC_DB_DSN: "postgres://cryptouser:cryptopass@postgres:5432/kyc_db?sslmode=disable"
      RATES_DB_DSN: "postgres://cryptouser:cryptopass@postgres:5432/rates_db?sslmode=disable"
      AUDIT_DB_DSN: "postgres://cryptouser:cryptopass@postgres:5432/audit_db?sslmode=disable"
      APP_ENV: "development"
    ports:
      - "8080:8080"
    volumes:
      - ./internal/database/migrations:/app/migrations

volumes:
  postgres_data:
```

**File**: `scripts/init-databases.sh`

```bash
#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    CREATE DATABASE core_db;
    CREATE DATABASE kyc_db;
    CREATE DATABASE rates_db;
    CREATE DATABASE audit_db;
EOSQL
```

---

## 8. Testing Strategy

### Testing Approach

1. **Unit Tests**: Test migration logic in Go functions
2. **Integration Tests**: Test migrations against real PostgreSQL
3. **CI Tests**: Automated migration testing in GitHub Actions
4. **Rollback Tests**: Verify up/down migration pairs work

### Integration Test with Testcontainers

**File**: `internal/database/migrate_test.go`

```go
//go:build integration
// +build integration

package database

import (
    "context"
    "database/sql"
    "testing"
    "time"

    "github.com/pressly/goose/v3"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/postgres"
    "github.com/testcontainers/testcontainers-go/wait"
)

func TestMigrations(t *testing.T) {
    ctx := context.Background()

    // Start PostgreSQL container
    postgresContainer, err := postgres.RunContainer(ctx,
        testcontainers.WithImage("postgres:15-alpine"),
        postgres.WithDatabase("testdb"),
        postgres.WithUsername("testuser"),
        postgres.WithPassword("testpass"),
        testcontainers.WithWaitStrategy(
            wait.ForLog("database system is ready to accept connections").
                WithOccurrence(2).
                WithStartupTimeout(5*time.Second),
        ),
    )
    require.NoError(t, err)
    defer postgresContainer.Terminate(ctx)

    // Get connection string
    connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
    require.NoError(t, err)

    // Test each database migrations
    databases := []string{"core_db", "kyc_db", "rates_db", "audit_db"}

    for _, dbName := range databases {
        t.Run(dbName, func(t *testing.T) {
            testDatabaseMigrations(t, connStr, dbName)
        })
    }
}

func testDatabaseMigrations(t *testing.T, connStr, dbName string) {
    db, err := sql.Open("postgres", connStr)
    require.NoError(t, err)
    defer db.Close()

    // Set migration directory
    migrationDir := "migrations/" + dbName
    goose.SetBaseFS(nil) // Use filesystem

    err = goose.SetDialect("postgres")
    require.NoError(t, err)

    // Test UP migrations
    t.Run("up_migrations", func(t *testing.T) {
        err := goose.Up(db, migrationDir)
        assert.NoError(t, err, "Up migrations should succeed")

        // Verify version table exists
        var count int
        err = db.QueryRow("SELECT COUNT(*) FROM goose_db_version").Scan(&count)
        assert.NoError(t, err)
        assert.Greater(t, count, 0, "Should have migration records")
    })

    // Test DOWN migrations (rollback)
    t.Run("down_migrations", func(t *testing.T) {
        // Get current version
        version, err := goose.GetDBVersion(db)
        require.NoError(t, err)
        require.Greater(t, version, int64(0))

        // Roll back one migration
        err = goose.Down(db, migrationDir)
        assert.NoError(t, err, "Down migration should succeed")

        // Verify version decreased
        newVersion, err := goose.GetDBVersion(db)
        require.NoError(t, err)
        assert.Less(t, newVersion, version, "Version should decrease after rollback")

        // Migrate back up
        err = goose.Up(db, migrationDir)
        assert.NoError(t, err, "Re-applying migrations should succeed")
    })

    // Test idempotency
    t.Run("idempotency", func(t *testing.T) {
        // Running up twice should not cause errors
        err := goose.Up(db, migrationDir)
        assert.NoError(t, err, "Running up migrations twice should be safe")
    })
}

func TestMigrationValidation(t *testing.T) {
    databases := []string{"core_db", "kyc_db", "rates_db", "audit_db"}

    for _, dbName := range databases {
        t.Run(dbName, func(t *testing.T) {
            migrationDir := "migrations/" + dbName

            // Verify all .up.sql files have corresponding .down.sql files
            upFiles, err := filepath.Glob(filepath.Join(migrationDir, "*.up.sql"))
            require.NoError(t, err)

            for _, upFile := range upFiles {
                downFile := strings.Replace(upFile, ".up.sql", ".down.sql", 1)
                assert.FileExists(t, downFile,
                    "Every up migration must have a down migration")
            }
        })
    }
}
```

### Manual Testing Checklist

```bash
# 1. Test fresh migration
make migrate-up-all
make test-integration

# 2. Test rollback
make migrate-down db=core_db
make migrate-up db=core_db

# 3. Test seed data
APP_ENV=development make migrate-up-all

# 4. Verify schema
psql -h localhost -U user -d core_db -c "\dt"
psql -h localhost -U user -d core_db -c "\d users"

# 5. Test with Docker
docker-compose up --build
```

---

## 9. Makefile for Common Operations

**File**: `Makefile`

```makefile
.PHONY: help migrate-up migrate-down migrate-status migrate-create seed-dev

DB ?= all
NAME ?= migration_name

help:
	@echo "Database Migration Commands"
	@echo "  make migrate-up [DB=db_name]     - Run migrations (default: all databases)"
	@echo "  make migrate-down DB=db_name     - Rollback last migration"
	@echo "  make migrate-status [DB=db_name] - Show migration status"
	@echo "  make migrate-create DB=db_name NAME=name - Create new migration"
	@echo "  make seed-dev                    - Seed development data"
	@echo "  make test-migrations             - Run migration tests"

migrate-up:
	go run cmd/migrate/main.go -command=up -db=$(DB)

migrate-down:
	@if [ "$(DB)" = "all" ]; then \
		echo "ERROR: Cannot rollback all databases. Specify DB=db_name"; \
		exit 1; \
	fi
	go run cmd/migrate/main.go -command=down -db=$(DB)

migrate-status:
	go run cmd/migrate/main.go -command=status -db=$(DB)

migrate-create:
	@if [ "$(DB)" = "all" ] || [ -z "$(NAME)" ]; then \
		echo "ERROR: Must specify DB=db_name and NAME=migration_name"; \
		exit 1; \
	fi
	goose -dir internal/database/migrations/$(DB) create $(NAME) sql

seed-dev:
	APP_ENV=development $(MAKE) migrate-up

test-migrations:
	go test -v ./internal/database/... -tags=integration

docker-up:
	docker-compose up --build -d

docker-down:
	docker-compose down -v

docker-logs:
	docker-compose logs -f app
```

---

## 10. Production Deployment Checklist

### Pre-Deployment

```markdown
## Migration Deployment Checklist

### Testing Phase
- [ ] All migrations tested in development
- [ ] All migrations tested in staging
- [ ] Rollback tested for each migration
- [ ] Integration tests passing
- [ ] Performance impact assessed
- [ ] Migration duration estimated

### Backup Phase
- [ ] Full database backup completed
- [ ] Backup integrity verified (pg_restore --list)
- [ ] Backup stored in secure location
- [ ] Backup retention policy confirmed
- [ ] Point-in-time recovery (PITR) enabled

### Communication Phase
- [ ] Team notified of migration window
- [ ] Stakeholders informed of downtime (if any)
- [ ] Rollback plan documented and shared
- [ ] On-call engineer assigned

### Execution Phase
- [ ] Maintenance mode enabled (if required)
- [ ] Migration executed
- [ ] Migration logs reviewed
- [ ] Schema verification completed
- [ ] Application smoke tests passed
- [ ] Monitoring alerts configured

### Post-Deployment Phase
- [ ] Application fully operational
- [ ] Database performance metrics normal
- [ ] Error logs reviewed
- [ ] Migration marked as successful
- [ ] Team notified of completion
- [ ] Documentation updated
```

---

## 11. Summary & Recommendations

### Final Recommendations

1. **Use pressly/goose** for its Go migration support and flexibility
2. **Separate migrations per database** for clear boundaries and independent evolution
3. **Version seed data separately** (00010+) from schema migrations (00001-00009)
4. **Implement comprehensive rollback testing** in CI/CD pipeline
5. **Use embedded migrations** for deployment simplicity
6. **Maintain strict naming conventions** for migration files
7. **Test migrations with testcontainers** for isolation
8. **Automate migration execution** in Docker entrypoint
9. **Monitor migration performance** in production
10. **Document all non-reversible migrations** clearly

### Key Success Factors

- **Idempotency**: Migrations should be safe to run multiple times
- **Atomicity**: Use transactions to ensure all-or-nothing execution
- **Reversibility**: Every up must have a tested down migration
- **Testing**: Comprehensive testing before production deployment
- **Monitoring**: Track migration execution time and failures
- **Documentation**: Clear documentation of migration purpose and impact

### Next Steps

1. Install goose: `go install github.com/pressly/goose/v3/cmd/goose@latest`
2. Create migration directory structure as shown above
3. Implement the migrate helper functions
4. Create initial migrations for all 4 databases
5. Set up GitHub Actions workflow
6. Configure Docker integration
7. Write integration tests
8. Document team migration workflow

---

## References

- [pressly/goose GitHub](https://github.com/pressly/goose)
- [pressly/goose Documentation](https://pressly.github.io/goose/)
- [golang-migrate GitHub](https://github.com/golang-migrate/migrate)
- [PostgreSQL Migration Best Practices](https://www.postgresql.org/docs/current/ddl-alter.html)
- [Testcontainers Go](https://golang.testcontainers.org/)
- [Database Migration Patterns](https://martinfowler.com/articles/evodb.html)
