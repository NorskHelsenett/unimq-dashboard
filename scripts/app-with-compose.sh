#!/bin/bash

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

cd "$ROOT_DIR"


# Start docker compose in detached mode in the background
if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
  echo "  Starting Docker Compose (Dex auth)..."
  docker compose up -d
else
  echo "  WARN: Docker Compose not available; skipping 'docker compose up'."
fi

# Start the Go application
echo "  Starting Go backend..."
go run ./cmd/unimq/main.go &
BACKEND_PID=$!

# Start the React application
echo "  Starting Vite dev server..."
npm run dev --prefix ./web-src &
FRONTEND_PID=$!

cleanup() {
  kill "$BACKEND_PID" "$FRONTEND_PID" 2>/dev/null

  if [[ "${KEEP_COMPOSE:-}" == "1" ]]; then
    echo "KEEP_COMPOSE=1 set; leaving Docker Compose services running."
    return
  fi

  if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
    echo "Stopping Docker Compose services..."
    docker compose down
  fi
}

trap cleanup EXIT

sleep 1

ITALIC='\033[3m'
ORANGE='\033[38;2;249;115;22m'
BOLD='\033[1m'
RESET='\033[0m'

echo ""
echo ""
echo -e "${BOLD}${ORANGE}  ██╗   ██╗███╗   ██╗██╗███╗   ███╗ ██████╗ ${RESET}"
echo -e "${BOLD}${ORANGE}  ██║   ██║████╗  ██║██║████╗ ████║██╔═══██╗${RESET}"
echo -e "${BOLD}${ORANGE}  ██║   ██║██╔██╗ ██║██║██╔████╔██║██║   ██║${RESET}"
echo -e "${BOLD}${ORANGE}  ██║   ██║██║╚██╗██║██║██║╚██╔╝██║██║▄▄ ██║${RESET}"
echo -e "${BOLD}${ORANGE}  ╚██████╔╝██║ ╚████║██║██║ ╚═╝ ██║╚██████╔╝${RESET}"
echo -e "${BOLD}${ORANGE}   ╚═════╝ ╚═╝  ╚═══╝╚═╝╚═╝     ╚═╝ ╚══▀▀═╝${RESET}"
echo ""
echo -e "  ${BOLD}${ITALIC}development mode - enjoy coding${RESET}"
echo ""
echo ""
echo "  All services started:"
echo -e "    ${ORANGE}›${RESET} Go backend       → http://localhost:8080"
echo -e "    ${ORANGE}›${RESET} Vite dev server  → http://localhost:5173"
echo ""
echo -e "  ${ORANGE}›${RESET} Run ${BOLD}docker compose down${RESET} to stop docker container."
echo ""
wait
