#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND_DIR="${ROOT_DIR}/crypto-wallet-backend"
FRONTEND_DIR="${ROOT_DIR}/crypto-wallet-frontend"

log() {
  printf "[integration] %s\n" "$1"
}

log "Running Go unit tests"
(cd "$BACKEND_DIR" && go test ./...)

log "Running TypeScript checks"
(cd "$FRONTEND_DIR" && npm install >/dev/null 2>&1 && npm run lint)

log "Running frontend unit tests"
(cd "$FRONTEND_DIR" && npm run test -- --watch=false)

log "Integration suite completed successfully"
