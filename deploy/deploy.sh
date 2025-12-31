#!/usr/bin/env bash
set -euo pipefail

COMPOSE_FILE="$(dirname "${BASH_SOURCE[0]}")/docker-compose.prod.yml"

function usage() {
  cat <<'USAGE'
Usage: ./deploy/deploy.sh [command]

Commands:
  up        Build images and start the production stack (default)
  down      Stop services and remove containers
  logs      Tail logs for all services

Examples:
  ./deploy/deploy.sh           # builds and starts the stack
  ./deploy/deploy.sh down      # stops containers
  ./deploy/deploy.sh logs      # streams logs
USAGE
}

command="${1:-up}"

case "${command}" in
  up)
    docker compose -f "${COMPOSE_FILE}" build
    docker compose -f "${COMPOSE_FILE}" up -d
    ;;
  down)
    docker compose -f "${COMPOSE_FILE}" down
    ;;
  logs)
    docker compose -f "${COMPOSE_FILE}" logs -f
    ;;
  help|-h|--help)
    usage
    ;;
  *)
    echo "Unknown command: ${command}" >&2
    usage
    exit 1
    ;;
esac
