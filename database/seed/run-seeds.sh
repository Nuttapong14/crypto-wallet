#!/usr/bin/env bash
set -euo pipefail

print_usage() {
  cat <<'USAGE'
Usage: ./database/seed/run-seeds.sh

Required environment variables:
  CORE_DB_DSN   - connection string for the core database
  KYC_DB_DSN    - connection string for the KYC database
  RATES_DB_DSN  - connection string for the rates database
  AUDIT_DB_DSN  - connection string for the audit database

Optional flags:
  --dry-run      Display commands without executing them
USAGE
}

dry_run=false
for arg in "$@"; do
  case "$arg" in
    --dry-run) dry_run=true ;;
    --help|-h) print_usage; exit 0 ;;
    *) echo "Unknown argument: $arg" >&2; print_usage; exit 1 ;;
  esac
done

require_dsn() {
  local name="$1"
  local value="$2"
  if [[ -z "${value:-}" ]]; then
    echo "Environment variable ${name} is required." >&2
    exit 1
  fi
}

require_dsn "CORE_DB_DSN" "${CORE_DB_DSN:-}"
require_dsn "KYC_DB_DSN" "${KYC_DB_DSN:-}"
require_dsn "RATES_DB_DSN" "${RATES_DB_DSN:-}"
require_dsn "AUDIT_DB_DSN" "${AUDIT_DB_DSN:-}"

run_psql() {
  local dsn="$1"
  local script="$2"
  local cmd=(psql "$dsn" -v "ON_ERROR_STOP=1" -f "$script")
  if $dry_run; then
    echo "[dry-run] ${cmd[*]}"
  else
    echo "Seeding $(basename "$script")..."
    "${cmd[@]}"
  fi
}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

run_psql "${CORE_DB_DSN}"   "${SCRIPT_DIR}/core_db_seed.sql"
run_psql "${KYC_DB_DSN}"    "${SCRIPT_DIR}/kyc_db_seed.sql"
run_psql "${RATES_DB_DSN}"  "${SCRIPT_DIR}/rates_db_seed.sql"
run_psql "${AUDIT_DB_DSN}"  "${SCRIPT_DIR}/audit_db_seed.sql"

echo "Seed data applied successfully."
