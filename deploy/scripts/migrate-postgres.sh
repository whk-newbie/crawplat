#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
COMPOSE_FILE="${COMPOSE_FILE:-${ROOT_DIR}/deploy/docker-compose/docker-compose.mcp.yml}"
MIGRATIONS_DIR="${ROOT_DIR}/deploy/migrations/postgres"
DB_USER="${POSTGRES_USER:-crawler}"
DB_NAME="${POSTGRES_DB:-crawler}"
STARTUP_TIMEOUT_SECONDS="${STARTUP_TIMEOUT_SECONDS:-60}"

docker compose -f "${COMPOSE_FILE}" up -d postgres

start_time="$(date +%s)"

until docker compose -f "${COMPOSE_FILE}" exec -T postgres pg_isready -U "${DB_USER}" -d "${DB_NAME}" >/dev/null 2>&1; do
  now="$(date +%s)"
  if [ $((now - start_time)) -ge "${STARTUP_TIMEOUT_SECONDS}" ]; then
    echo "postgres did not become ready within ${STARTUP_TIMEOUT_SECONDS}s" >&2
    docker compose -f "${COMPOSE_FILE}" logs postgres >&2 || true
    exit 1
  fi
  sleep 1
done

for migration in "${MIGRATIONS_DIR}"/*.sql; do
  docker compose -f "${COMPOSE_FILE}" exec -T postgres \
    psql -U "${DB_USER}" -d "${DB_NAME}" -v ON_ERROR_STOP=1 \
    -f "/migrations/$(basename "${migration}")"
done
