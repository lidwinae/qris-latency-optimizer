#!/bin/bash
# ─────────────────────────────────────────────────────────────
#  QRIS Latency Optimizer — Go Load Test CLI Launcher
# ─────────────────────────────────────────────────────────────
#  Usage:
#    ./loadtest/run.sh              # interactive menu
#    ./loadtest/run.sh --help       # show help
#
#  Environment variables:
#    BASE_URL  — override backend target (default: http://localhost:8080)
# ─────────────────────────────────────────────────────────────

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
export BASE_URL="${BASE_URL:-http://localhost:8080}"

# Show help
if [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
    echo ""
    echo "QRIS Load Test CLI"
    echo "─────────────────────────────────────────"
    echo "Usage: ./loadtest/run.sh"
    echo ""
    echo "Menu Options:"
    echo "  1) 🟢 Light Load:    10 concurrent users for 30s (200-800ms delay)"
    echo "  2) 🟡 Medium Load:   50 concurrent users for 30s (0.5-2s delay)"
    echo "  3) 🔴 Heavy Load:    100 concurrent users for 60s (1-4s delay)"
    echo "  4) 💀 Extreme Load:  200 concurrent users for 60s (2-8s delay)"
    echo "  5) 📊 Quick Bench:   50 concurrent users for 15s (no stress)"
    echo "  6) 🔧 Enable Stress Mode only"
    echo "  7) ✅ Disable Stress Mode"
    echo "  8) 🚪 Exit"
    echo ""
    echo "Environment:"
    echo "  BASE_URL=http://host:port  Override backend target"
    echo ""
    exit 0
fi

echo ""
echo "╔══════════════════════════════════════════════════════╗"
echo "║    QRIS Load Test CLI — Launching...                ║"
echo "║    Target: $BASE_URL                  ║"
echo "╚══════════════════════════════════════════════════════╝"
echo ""

cd "$SCRIPT_DIR"
go run main.go
