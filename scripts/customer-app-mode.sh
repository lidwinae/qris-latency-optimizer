#!/bin/sh
# Switch the Docker customer app between direct backend and rural proxy modes.

set -e

MODE="$1"
ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"

usage() {
  echo "Usage: ./scripts/customer-app-mode.sh <normal|rural|status>"
}

current_mode() {
  port="$(docker inspect -f '{{range .Config.Env}}{{println .}}{{end}}' qris_customer_app 2>/dev/null | grep '^VITE_API_PORT=' | tail -n 1 | cut -d= -f2-)"
  if [ "$port" = "8081" ]; then
    echo "rural"
  else
    echo "normal"
  fi
}

recreate_customer_app() {
  cd "$ROOT_DIR"
  CUSTOMER_APP_API_PORT="${1:-8080}" docker compose up -d --force-recreate customer_app
}

case "$MODE" in
  normal)
    recreate_customer_app 8080
    echo "Customer app switched to normal mode: API port 8080."
    echo "API target: http://localhost:8080/api"
    echo "Open http://localhost:5174"
    ;;
  rural)
    cd "$ROOT_DIR"
    CUSTOMER_APP_API_PORT=8081 docker compose up -d toxiproxy golang customer_app
    "$ROOT_DIR/k6/rural_test_setup.sh"
    CUSTOMER_APP_API_PORT=8081 docker compose up -d --force-recreate customer_app
    echo "Customer app switched to rural mode: API port 8081 via Toxiproxy."
    echo "API target: http://localhost:8081/api"
    echo "Open http://localhost:5174"
    ;;
  status)
    echo "Customer app mode: $(current_mode)"
    ;;
  *)
    usage
    exit 1
    ;;
esac
