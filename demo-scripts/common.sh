#!/usr/bin/env bash
# Shared helpers for demo scripts

AOR_URL="http://localhost:8080"
AOR_DIR="$HOME/GoProjects/agent-orchestrator"

BOLD='\033[1m'
DIM='\033[2m'
CYAN='\033[36m'
GREEN='\033[32m'
YELLOW='\033[33m'
MAGENTA='\033[35m'
RESET='\033[0m'

generate_jwt() {
  cd "$AOR_DIR"
  python3 -c "
import sys; sys.path.insert(0, '.')
from apps.common.utils.build_jwt import _build_jwt
print(_build_jwt('league', 'demo-user'))
" 2>/dev/null || .hackathon-venv/bin/python -c "
import sys; sys.path.insert(0, '.')
from apps.common.utils.build_jwt import _build_jwt
print(_build_jwt('league', 'demo-user'))
" 2>/dev/null
}

create_session() {
  curl -s -X POST "$AOR_URL/v1/sessions" \
    -H "Authorization: Bearer $JWT" \
    -H "Content-Type: application/json" \
    -d '{}' | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])"
}

send_message() {
  local session_id="$1"
  local message="$2"

  echo -e "\n${CYAN}${BOLD}You:${RESET} $message"
  echo -e "${DIM}───────────────────────────────────────────${RESET}"

  local response
  response=$(curl -s -N -X POST "$AOR_URL/v1/run-test-agent" \
    -H "Authorization: Bearer $JWT" \
    -H "Content-Type: application/json" \
    -H "User-Agent: Demo-Script" \
    -H "X-Skip-Guardrails: true" \
    -d "{\"id\":\"$(uuidgen)\",\"session_id\":\"$session_id\",\"message\":\"$message\"}")

  # Extract tool calls
  local tools
  tools=$(echo "$response" | grep -o '"function_call":{[^}]*}' | grep -o '"name":"[^"]*"' | sed 's/"name":"//;s/"//' | sort -u)
  if [[ -n "$tools" ]]; then
    echo -e "${YELLOW}  Tools used:${RESET}"
    while IFS= read -r tool; do
      local display_name="${tool/documents_mcp_/}"
      display_name="${display_name/ui_mcp_/ui: }"
      echo -e "    ${YELLOW}→${RESET} $display_name"
    done <<< "$tools"
  fi

  # Extract text
  local text
  text=$(echo "$response" | grep -o '"text":"[^"]*"' | sed 's/"text":"//;s/"$//' | head -5)
  if [[ -n "$text" ]]; then
    echo -e "\n${GREEN}${BOLD}Agent:${RESET}"
    echo "$text" | while IFS= read -r line; do
      # Unescape common sequences
      echo -e "  $(echo "$line" | sed 's/\\n/\n  /g; s/\\t/  /g')"
    done
  fi

  # Extract cards
  local cards
  cards=$(echo "$response" | grep -o '"cards":\[.*\]' | head -1)
  if [[ -n "$cards" ]]; then
    local card_titles
    card_titles=$(echo "$cards" | grep -o '"title":"[^"]*"' | sed 's/"title":"//;s/"//')
    if [[ -n "$card_titles" ]]; then
      echo -e "\n${MAGENTA}  Cards rendered:${RESET}"
      while IFS= read -r title; do
        echo -e "    ${MAGENTA}▪${RESET} $title"
      done <<< "$card_titles"
    fi
  fi

  echo -e "\n${DIM}═══════════════════════════════════════════${RESET}"
}

pause() {
  echo -e "\n${DIM}Press Enter to continue...${RESET}"
  read -r
}

check_services() {
  echo -e "${DIM}Checking services...${RESET}"
  local ok=true
  curl -s "$AOR_URL/healthz" | grep -q ok && echo -e "  ${GREEN}✓${RESET} AOR on :8080" || { echo -e "  ✗ AOR not running"; ok=false; }
  curl -s http://localhost:18006/mcp -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"ping","id":1}' > /dev/null 2>&1 && echo -e "  ${GREEN}✓${RESET} MCP server on :18006" || { echo -e "  ✗ MCP server not running"; ok=false; }
  curl -s http://localhost:8092/health > /dev/null 2>&1 && echo -e "  ${GREEN}✓${RESET} Mock server on :8092" || echo -e "  ${GREEN}✓${RESET} Mock server on :8092 (assumed)"
  $ok || { echo -e "\n${BOLD}Run ~/GoProjects/hackathon-agent-documents/start.sh first${RESET}"; exit 1; }
  echo ""
}
