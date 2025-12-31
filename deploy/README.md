# Production Deployment

The `deploy` directory contains a lightweight deployment harness for running
the platform with Docker Compose in a production-like environment. The stack is
comprised of the backend API, the Next.js frontend, PostgreSQL (hosting the four
logical databases), and Redis.

## Prerequisites

- Docker 24+
- Docker Compose plugin (`docker compose` CLI)
- Seeded databases (optional but recommended): `./database/seed/run-seeds.sh`

## Quick Start

```bash
cd deploy
./deploy.sh         # builds images and starts the stack in the background
./deploy.sh logs    # follow logs for all services
./deploy.sh down    # stop and remove containers
```

### Environment Files

Editable environment templates live under `deploy/env/`:

- `backend.prod.env` – database DSNs, secrets, encryption keys.
- `frontend.prod.env` – public runtime configuration for the Next.js app.
- `postgres.prod.env` – credentials and default database names.

These files ship with sensible defaults for local evaluation. **Update the
secrets before deploying to a real environment.**

## Customisation

- Adjust exposed ports within `docker-compose.prod.yml` to suit your network.
- Override compose variables via `docker compose --env-file` or environment
  variables on invocation.
- Integrate with an external PostgreSQL instance by commenting out the local
  `postgres` service and updating the DSNs in `backend.prod.env`.

## Deployment Script Commands

| Command | Description |
|---------|-------------|
| `./deploy.sh` | Build images and start containers in detached mode |
| `./deploy.sh down` | Stop the stack and remove containers |
| `./deploy.sh logs` | Tail logs from all services |

The script is intentionally minimal and can be integrated into CI/CD pipelines
or invoked manually from an operations workstation.
