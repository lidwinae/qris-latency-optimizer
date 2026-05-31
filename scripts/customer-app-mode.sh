#!/bin/sh
# Switch the Docker customer app between direct backend and rural proxy modes.

set -e

MODE="$1"
ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
ENV_FILE="$ROOT_DIR/.env"

usage() {
  echo "Usage: ./scripts/customer-app-mode.sh <normal|rural|status>"
}

set_env_value() {
  key="$1"
  value="$2"

  if [ ! -f "$ENV_FILE" ]; then
    touch "$ENV_FILE"
    echo "Created .env"
  fi

  if grep -q "^${key}=" "$ENV_FILE"; then
    sed -i "s/^${key}=.*/${key}=${value}/" "$ENV_FILE"
  else
    printf "\n%s=%s\n" "$key" "$value" >> "$ENV_FILE"
  fi
}

current_mode() {
  port="$(grep '^CUSTOMER_APP_API_PORT=' "$ENV_FILE" 2>/dev/null | tail -n 1 | cut -d= -f2-)"
  if [ "$port" = "8081" ]; then
    echo "rural"
  else
    echo "normal"
  fi
}

recreate_customer_app() {
  cd "$ROOT_DIR"
  docker compose up -d --force-recreate customer_app
}

case "$MODE" in
  normal)
    set_env_value "CUSTOMER_APP_API_PORT" "8080"
    recreate_customer_app
    echo "Customer app switched to normal mode: API port 8080."
    echo "API target: http://localhost:8080/api"
    echo "Open http://localhost:5174"
    ;;
  rural)
    set_env_value "CUSTOMER_APP_API_PORT" "8081"
    cd "$ROOT_DIR"
    docker compose up -d toxiproxy golang customer_app
    "$ROOT_DIR/k6/rural_test_setup.sh"
    docker compose up -d --force-recreate customer_app
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
