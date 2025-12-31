#!/bin/bash

# ===============================
# Multi-Chain Crypto Wallet
# Database Initialization Script
# ===============================
#
# Purpose: Create and initialize 4 PostgreSQL databases for the crypto wallet system
# Databases:
#   1. core_db    - Operational data (users, wallets, transactions)
#   2. kyc_db     - Compliance data (KYC profiles, documents, risk scores)
#   3. rates_db   - Market data (exchange rates, price history)
#   4. audit_db   - Security logs (audit trails, API logs)
#
# Usage:
#   ./init-databases.sh
#
# Requirements:
#   - PostgreSQL 16+ running
#   - psql client installed
#   - POSTGRES_USER and POSTGRES_PASSWORD environment variables set
#     OR database connection via environment variables
#
# Environment Variables:
#   POSTGRES_HOST     - Database host (default: localhost)
#   POSTGRES_PORT     - Database port (default: 5432)
#   POSTGRES_USER     - Superuser for creating databases (default: postgres)
#   POSTGRES_PASSWORD - Superuser password
#   DB_USER           - Application database user (default: crypto_user)
#   DB_PASSWORD       - Application database password (default: crypto_pass)
#
# ===============================

set -e  # Exit on error
set -u  # Exit on undefined variable

# ===============================
# Configuration
# ===============================

# PostgreSQL connection settings
POSTGRES_HOST="${POSTGRES_HOST:-localhost}"
POSTGRES_PORT="${POSTGRES_PORT:-5432}"
POSTGRES_USER="${POSTGRES_USER:-postgres}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-}"

# Application database user
DB_USER="${DB_USER:-crypto_user}"
DB_PASSWORD="${DB_PASSWORD:-crypto_pass}"

# Database names
DB_CORE="core_db"
DB_KYC="kyc_db"
DB_RATES="rates_db"
DB_AUDIT="audit_db"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# ===============================
# Helper Functions
# ===============================

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to execute SQL command
execute_sql() {
    local sql="$1"
    local dbname="${2:-postgres}"

    if [ -n "$POSTGRES_PASSWORD" ]; then
        PGPASSWORD="$POSTGRES_PASSWORD" psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$dbname" -c "$sql" 2>/dev/null
    else
        psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$dbname" -c "$sql" 2>/dev/null
    fi
}

# Function to check if database exists
database_exists() {
    local dbname="$1"
    local result

    if [ -n "$POSTGRES_PASSWORD" ]; then
        result=$(PGPASSWORD="$POSTGRES_PASSWORD" psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname='$dbname'" 2>/dev/null)
    else
        result=$(psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname='$dbname'" 2>/dev/null)
    fi

    [ "$result" = "1" ]
}

# Function to check if user exists
user_exists() {
    local username="$1"
    local result

    if [ -n "$POSTGRES_PASSWORD" ]; then
        result=$(PGPASSWORD="$POSTGRES_PASSWORD" psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d postgres -tAc "SELECT 1 FROM pg_roles WHERE rolname='$username'" 2>/dev/null)
    else
        result=$(psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d postgres -tAc "SELECT 1 FROM pg_roles WHERE rolname='$username'" 2>/dev/null)
    fi

    [ "$result" = "1" ]
}

# Function to create database
create_database() {
    local dbname="$1"
    local description="$2"

    log_info "Creating database: $dbname ($description)"

    if database_exists "$dbname"; then
        log_warning "Database $dbname already exists, skipping creation"
        return 0
    fi

    execute_sql "CREATE DATABASE $dbname OWNER $DB_USER ENCODING 'UTF8' LC_COLLATE='en_US.UTF-8' LC_CTYPE='en_US.UTF-8' TEMPLATE=template0;" "postgres"

    if [ $? -eq 0 ]; then
        log_success "Database $dbname created successfully"

        # Grant necessary extensions
        log_info "Setting up extensions for $dbname"
        execute_sql "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";" "$dbname"
        execute_sql "CREATE EXTENSION IF NOT EXISTS \"pgcrypto\";" "$dbname"

        log_success "Extensions enabled for $dbname"
    else
        log_error "Failed to create database $dbname"
        return 1
    fi
}

# Function to setup database user
setup_user() {
    log_info "Setting up application database user: $DB_USER"

    if user_exists "$DB_USER"; then
        log_warning "User $DB_USER already exists"

        # Update password
        log_info "Updating password for user $DB_USER"
        execute_sql "ALTER USER $DB_USER WITH PASSWORD '$DB_PASSWORD';" "postgres"
    else
        log_info "Creating user $DB_USER"
        execute_sql "CREATE USER $DB_USER WITH PASSWORD '$DB_PASSWORD' LOGIN CREATEDB;" "postgres"
        log_success "User $DB_USER created successfully"
    fi

    # Grant connection privileges
    execute_sql "GRANT CONNECT ON DATABASE $DB_CORE TO $DB_USER;" "postgres"
    execute_sql "GRANT CONNECT ON DATABASE $DB_KYC TO $DB_USER;" "postgres"
    execute_sql "GRANT CONNECT ON DATABASE $DB_RATES TO $DB_USER;" "postgres"
    execute_sql "GRANT CONNECT ON DATABASE $DB_AUDIT TO $DB_USER;" "postgres"

    log_success "User privileges configured"
}

# Function to grant schema permissions
grant_permissions() {
    local dbname="$1"

    log_info "Granting permissions for $dbname"

    execute_sql "GRANT ALL PRIVILEGES ON DATABASE $dbname TO $DB_USER;" "postgres"
    execute_sql "GRANT ALL ON SCHEMA public TO $DB_USER;" "$dbname"
    execute_sql "GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO $DB_USER;" "$dbname"
    execute_sql "GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO $DB_USER;" "$dbname"
    execute_sql "ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO $DB_USER;" "$dbname"
    execute_sql "ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO $DB_USER;" "$dbname"

    log_success "Permissions granted for $dbname"
}

# Function to test connection
test_connection() {
    log_info "Testing PostgreSQL connection..."

    if [ -n "$POSTGRES_PASSWORD" ]; then
        if PGPASSWORD="$POSTGRES_PASSWORD" psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d postgres -c "SELECT version();" > /dev/null 2>&1; then
            log_success "PostgreSQL connection successful"
            return 0
        fi
    else
        if psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d postgres -c "SELECT version();" > /dev/null 2>&1; then
            log_success "PostgreSQL connection successful"
            return 0
        fi
    fi

    log_error "Failed to connect to PostgreSQL"
    log_error "Please check:"
    log_error "  - PostgreSQL is running"
    log_error "  - POSTGRES_HOST=$POSTGRES_HOST"
    log_error "  - POSTGRES_PORT=$POSTGRES_PORT"
    log_error "  - POSTGRES_USER=$POSTGRES_USER"
    log_error "  - POSTGRES_PASSWORD is set correctly"
    return 1
}

# ===============================
# Main Execution
# ===============================

main() {
    echo ""
    log_info "========================================="
    log_info "Crypto Wallet Database Initialization"
    log_info "========================================="
    echo ""

    # Test connection
    test_connection || exit 1
    echo ""

    # Setup application user
    setup_user
    echo ""

    # Create databases
    log_info "Creating databases..."
    echo ""

    create_database "$DB_CORE" "Core operational data (users, wallets, transactions)"
    create_database "$DB_KYC" "Compliance data (KYC profiles, documents, risk scores)"
    create_database "$DB_RATES" "Market data (exchange rates, price history)"
    create_database "$DB_AUDIT" "Security logs (audit trails, API logs)"
    echo ""

    # Grant permissions
    log_info "Configuring permissions..."
    echo ""

    grant_permissions "$DB_CORE"
    grant_permissions "$DB_KYC"
    grant_permissions "$DB_RATES"
    grant_permissions "$DB_AUDIT"
    echo ""

    # Summary
    log_success "========================================="
    log_success "Database initialization completed!"
    log_success "========================================="
    echo ""
    log_info "Databases created:"
    log_info "  1. $DB_CORE    - Core operational data"
    log_info "  2. $DB_KYC     - Compliance data"
    log_info "  3. $DB_RATES   - Market data"
    log_info "  4. $DB_AUDIT   - Security logs"
    echo ""
    log_info "Database user: $DB_USER"
    log_info "Connection: $POSTGRES_HOST:$POSTGRES_PORT"
    echo ""
    log_info "Next steps:"
    log_info "  1. Run migrations: cd crypto-wallet-backend && make migrate-up"
    log_info "  2. Seed data (optional): cd crypto-wallet-backend && make seed"
    echo ""
}

# Run main function
main "$@"
