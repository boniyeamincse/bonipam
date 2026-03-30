#!/usr/bin/env bash
set -euo pipefail

export APP_ENV=${APP_ENV:-development}
export LOG_LEVEL=${LOG_LEVEL:-debug}

echo "Starting auth-service on :8081"
HTTP_PORT=8081 go run ./cmd/auth-service &
AUTH_PID=$!

echo "Starting api-gateway on :8080"
HTTP_PORT=8080 go run ./cmd/api-gateway &
GATEWAY_PID=$!

cleanup() {
  echo "Stopping services..."
  kill "$AUTH_PID" "$GATEWAY_PID" 2>/dev/null || true
}

trap cleanup EXIT INT TERM
wait
