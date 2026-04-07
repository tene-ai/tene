#!/bin/bash
# Local development: start all services with tene-managed secrets
#
# Usage:
#   ./scripts/dev.sh              # Start everything (infra + apps)
#   ./scripts/dev.sh infra        # Start Docker infra only (DB + S3)
#   ./scripts/dev.sh api          # Start API only (needs infra)
#   ./scripts/dev.sh dashboard    # Start Dashboard only
#   ./scripts/dev.sh landing      # Start Landing page only
#   ./scripts/dev.sh apps         # Start all apps (no infra)
#   ./scripts/dev.sh stop         # Stop everything
#   ./scripts/dev.sh status       # Show running services
#
# Ports:
#   3000  Landing page    (apps/web)
#   3001  Dashboard       (apps/dashboard)
#   5432  PostgreSQL      (Docker)
#   8080  Go API server   (tene run)
#   9000  MinIO S3 API    (Docker)
#   9001  MinIO Console   (Docker)

set -euo pipefail
cd "$(dirname "$0")/.."

PIDS_FILE="/tmp/tene-dev-pids"

# ── Infra (Docker) ──────────────────────────────
start_infra() {
  echo "  Starting infra (PostgreSQL + MinIO)..."
  docker compose -f docker-compose.dev.yml up -d
  echo "  ✓ PostgreSQL: localhost:5432"
  echo "  ✓ MinIO S3:   localhost:9000"
  echo "  ✓ MinIO UI:   localhost:9001 (minioadmin/minioadmin)"

  echo "  Waiting for PostgreSQL..."
  for i in $(seq 1 15); do
    if docker exec tene-db pg_isready -U tene_admin -d tene > /dev/null 2>&1; then
      echo "  ✓ PostgreSQL ready"

      echo "  Running migrations..."
      for f in migrations/*.up.sql; do
        docker exec -i tene-db psql -U tene_admin -d tene < "$f" 2>/dev/null || true
      done
      echo "  ✓ Migrations applied"
      return
    fi
    sleep 1
  done
  echo "  ⚠ PostgreSQL not ready after 15s"
}

stop_infra() {
  docker compose -f docker-compose.dev.yml down 2>/dev/null || true
}

# ── App Servers ─────────────────────────────────
start_api() {
  echo "  Starting Go API (port 8080)..."
  tene run --env local -- go run ./cmd/server > /tmp/tene-api.log 2>&1 &
  echo $! >> "$PIDS_FILE"
  sleep 2
  if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "  ✓ API: http://localhost:8080/health"
  else
    echo "  ⏳ API starting... (check /tmp/tene-api.log)"
  fi
}

start_dashboard() {
  echo "  Starting Dashboard (port 3001)..."
  cd apps/dashboard
  [ ! -d "node_modules" ] && npm install --silent
  NEXT_PUBLIC_API_URL=http://localhost:8080 NEXT_PUBLIC_DASHBOARD_URL=http://localhost:3001 npx next dev --port 3001 > /tmp/tene-dashboard.log 2>&1 &
  echo $! >> "$PIDS_FILE"
  cd ../..
  echo "  ✓ Dashboard: http://localhost:3001"
}

start_landing() {
  echo "  Starting Landing (port 3000)..."
  cd apps/web
  [ ! -d "node_modules" ] && npm install --silent
  NEXT_PUBLIC_API_URL=http://localhost:8080 NEXT_PUBLIC_DASHBOARD_URL=http://localhost:3001 npx next dev --port 3000 > /tmp/tene-landing.log 2>&1 &
  echo $! >> "$PIDS_FILE"
  cd ../..
  echo "  ✓ Landing: http://localhost:3000"
}

# ── Stop ────────────────────────────────────────
stop_apps() {
  if [ -f "$PIDS_FILE" ]; then
    while read -r pid; do
      kill "$pid" 2>/dev/null || true
    done < "$PIDS_FILE"
    rm -f "$PIDS_FILE"
  fi
  pkill -f "go run ./cmd/server" 2>/dev/null || true
  pkill -f "next dev --port 300" 2>/dev/null || true
}

stop_all() {
  echo "  Stopping all..."
  stop_apps
  stop_infra
  echo "  ✓ All stopped"
}

# ── Status ──────────────────────────────────────
show_status() {
  echo ""
  echo "  ┌─ tene dev status ─────────────────────────────────────┐"
  printf "  │  %-12s %-30s %-8s │\n" "SERVICE" "URL" "STATUS"
  echo "  │─────────────────────────────────────────────────────│"

  # PostgreSQL
  if docker exec tene-db pg_isready -U tene_admin > /dev/null 2>&1; then
    printf "  │  %-12s %-30s %-8s │\n" "PostgreSQL" "localhost:5432" "✓ UP"
  else
    printf "  │  %-12s %-30s %-8s │\n" "PostgreSQL" "localhost:5432" "✗ DOWN"
  fi

  # MinIO
  if curl -s http://localhost:9000/minio/health/live > /dev/null 2>&1; then
    printf "  │  %-12s %-30s %-8s │\n" "MinIO S3" "localhost:9000" "✓ UP"
  else
    printf "  │  %-12s %-30s %-8s │\n" "MinIO S3" "localhost:9000" "✗ DOWN"
  fi

  # API
  if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    printf "  │  %-12s %-30s %-8s │\n" "API" "http://localhost:8080" "✓ UP"
  else
    printf "  │  %-12s %-30s %-8s │\n" "API" "http://localhost:8080" "✗ DOWN"
  fi

  # Dashboard
  if curl -s http://localhost:3001 > /dev/null 2>&1; then
    printf "  │  %-12s %-30s %-8s │\n" "Dashboard" "http://localhost:3001" "✓ UP"
  else
    printf "  │  %-12s %-30s %-8s │\n" "Dashboard" "http://localhost:3001" "✗ DOWN"
  fi

  # Landing
  if curl -s http://localhost:3000 > /dev/null 2>&1; then
    printf "  │  %-12s %-30s %-8s │\n" "Landing" "http://localhost:3000" "✓ UP"
  else
    printf "  │  %-12s %-30s %-8s │\n" "Landing" "http://localhost:3000" "✗ DOWN"
  fi

  echo "  └─────────────────────────────────────────────────────┘"
  echo ""
  echo "  Secrets: tene list --env local"
  echo "  Logs:    tail -f /tmp/tene-{api,dashboard,landing}.log"
  echo ""
}

# ── Main ────────────────────────────────────────
case "${1:-all}" in
  infra)       start_infra ;;
  api)         > "$PIDS_FILE"; start_api; wait ;;
  dashboard)   > "$PIDS_FILE"; start_dashboard; wait ;;
  landing)     > "$PIDS_FILE"; start_landing; wait ;;
  apps)        stop_apps; > "$PIDS_FILE"; start_api; start_dashboard; start_landing; wait ;;
  stop)        stop_all ;;
  status|ps)   show_status ;;
  all|start)
    stop_all 2>/dev/null
    > "$PIDS_FILE"
    echo ""
    echo "  ┌─ tene dev ──────────────────────────────────────────┐"
    echo "  │  Environment: local ($(./tene list --env local 2>&1 | grep -c 'ago\|now') secrets via tene)        │"
    echo "  └──────────────────────────────────────────────────────┘"
    echo ""
    start_infra
    echo ""
    start_api
    start_dashboard
    start_landing
    echo ""
    echo "  Waiting for Next.js servers to compile..."
    sleep 8
    show_status
    echo "  Press Ctrl+C to stop all"
    trap "stop_all; exit 0" INT TERM
    wait
    ;;
  *)
    echo "Usage: ./scripts/dev.sh [all|infra|api|dashboard|landing|apps|stop|status]"
    exit 1
    ;;
esac
