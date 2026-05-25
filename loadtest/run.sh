#!/bin/bash
# ==========================================
# QRIS Load Test Runner
# Run this while using the Customer App
# on your phone to feel REAL latency!
# ==========================================

set -e

API_BASE="http://localhost:8081"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${CYAN}"
echo "╔═══════════════════════════════════════════════════╗"
echo "║  🔥 QRIS Legacy Stress Test Suite                 ║"
echo "║  Creates REAL latency you can feel on your phone!  ║"
echo "╚═══════════════════════════════════════════════════╝"
echo -e "${NC}"

# Check backend is running
if ! curl -s "$API_BASE/api/legacy/merchants" > /dev/null 2>&1; then
    echo -e "${RED}❌ Backend is not running at $API_BASE${NC}"
    echo "   Start it first: cd backend && go run cmd/main.go"
    exit 1
fi
echo -e "${GREEN}✅ Backend is running${NC}"

show_menu() {
    echo ""
    echo -e "${YELLOW}Choose a test scenario:${NC}"
    echo ""
    echo "  1) 🟢 Light Load    — 10 users,  30s  (slight delay, 200-800ms)"
    echo "  2) 🟡 Medium Load   — 50 users,  30s  (noticeable lag, 0.5-2s)"
    echo "  3) 🔴 Heavy Load    — 100 users, 60s  (very slow, 1-4s)"
    echo "  4) 💀 Extreme Load  — 200 users, 60s  (near-timeout, 2-8s)"
    echo "  5) 📊 Quick Benchmark — 50 users, 15s (fast comparison numbers)"
    echo "  6) 🔧 Enable Stress Mode only (no load test, just enable delays)"
    echo "  7) ✅ Disable Stress Mode"
    echo "  8) 🚪 Exit"
    echo ""
    echo -n "  Select [1-8]: "
}

enable_stress() {
    echo -e "\n${YELLOW}⚡ Enabling stress simulation...${NC}"
    curl -s -X POST "$API_BASE/stress/on" | python3 -m json.tool 2>/dev/null || echo "done"
}

disable_stress() {
    echo -e "\n${GREEN}✅ Disabling stress simulation...${NC}"
    curl -s -X POST "$API_BASE/stress/off" | python3 -m json.tool 2>/dev/null || echo "done"
}

run_load_test() {
    local CONCURRENCY=$1
    local DURATION=$2
    local MODE=$3
    local LABEL=$4

    echo -e "\n${CYAN}════════════════════════════════════════${NC}"
    echo -e "${CYAN}  $LABEL${NC}"
    echo -e "${CYAN}════════════════════════════════════════${NC}"

    # Enable stress mode
    enable_stress

    echo -e "\n${YELLOW}📱 NOW OPEN THE CUSTOMER APP ON YOUR PHONE!${NC}"
    echo -e "${YELLOW}   URL: http://$(hostname -I | awk '{print $1}'):5176${NC}"
    echo -e "${YELLOW}   You should feel the lag while this test runs...${NC}"
    echo ""

    # Run Go load tester
    cd "$SCRIPT_DIR"
    go run main.go -c "$CONCURRENCY" -d "$DURATION" -mode "$MODE"

    echo -e "\n${GREEN}Test complete!${NC}"
    echo -n "Disable stress mode? [Y/n]: "
    read -r answer
    if [[ "$answer" != "n" && "$answer" != "N" ]]; then
        disable_stress
    fi
}

# Main loop
while true; do
    show_menu
    read -r choice

    case $choice in
        1) run_load_test 10 30 "mixed" "🟢 Light Load (10 concurrent users)" ;;
        2) run_load_test 50 30 "mixed" "🟡 Medium Load (50 concurrent users)" ;;
        3) run_load_test 100 60 "mixed" "🔴 Heavy Load (100 concurrent users)" ;;
        4) run_load_test 200 60 "mixed" "💀 Extreme Load (200 concurrent users)" ;;
        5) run_load_test 50 15 "mixed" "📊 Quick Benchmark (50 users, 15s)" ;;
        6) enable_stress ;;
        7) disable_stress ;;
        8) 
            disable_stress
            echo -e "\n${GREEN}Bye! 👋${NC}"
            exit 0
            ;;
        *) echo -e "${RED}Invalid choice${NC}" ;;
    esac
done
