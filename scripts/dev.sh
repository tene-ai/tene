#!/bin/bash
# Tene CLI — Local development
# Start the Landing page for local development
#
# Usage:
#   ./scripts/dev.sh              # Start Landing page
#   ./scripts/dev.sh landing      # Start Landing page
#   ./scripts/dev.sh stop         # Stop
#   ./scripts/dev.sh status       # Show status
#
# Ports:
#   3000  Landing page    (apps/web)

set -euo pipefail
cd "$(dirname "$0")/.."

PIDS_FILE="/tmp/tene-cli-dev-pids"

# ── Landing ─────────────────────────────────────
start_landing() {
  echo "  Starting Landing (port 3000)..."
  cd apps/web
  [ ! -d "node_modules" ] && npm install --silent
  NEXT_PUBLIC_API_URL="$(tene get API_URL --env local 2>/dev/null || echo http://localhost:8080)" \
  NEXT_PUBLIC_DASHBOARD_URL="$(tene get DASHBOARD_URL --env local 2>/dev/null || echo http://localhost:3001)" \
  npx next dev --port 3000 > /tmp/tene-landing.log 2>&1 &
  echo $! >> "$PIDS_FILE"
  cd ../..
  echo "  ✓ Landing: http://localhost:3000"
}

# ── Stop ────────────────────────────────────────
stop_all() {
  echo "  Stopping all..."
  if [ -f "$PIDS_FILE" ]; then
    while read -r pid; do
      kill "$pid" 2>/dev/null || true
    done < "$PIDS_FILE"
    rm -f "$PIDS_FILE"
  fi
  pkill -f "next dev --port 3000" 2>/dev/null || true
  echo "  ✓ All stopped"
}

# ── Status ──────────────────────────────────────
show_status() {
  echo ""
  echo "  ┌─ tene dev status ─────────────────────────────────────┐"
  printf "  │  %-12s %-30s %-8s │\n" "SERVICE" "URL" "STATUS"
  echo "  │─────────────────────────────────────────────────────│"

  if curl -s http://localhost:3000 > /dev/null 2>&1; then
    printf "  │  %-12s %-30s %-8s │\n" "Landing" "http://localhost:3000" "✓ UP"
  else
    printf "  │  %-12s %-30s %-8s │\n" "Landing" "http://localhost:3000" "✗ DOWN"
  fi

  echo "  └─────────────────────────────────────────────────────┘"
  echo ""
}

# ── Main ────────────────────────────────────────
case "${1:-all}" in
  landing|all|start)
    stop_all 2>/dev/null
    > "$PIDS_FILE"
    start_landing
    echo ""
    echo "  Waiting for Next.js to compile..."
    sleep 5
    show_status
    echo "  Press Ctrl+C to stop"
    trap "stop_all; exit 0" INT TERM
    wait
    ;;
  stop)        stop_all ;;
  status|ps)   show_status ;;
  *)
    echo "Usage: ./scripts/dev.sh [all|landing|stop|status]"
    exit 1
    ;;
esac
