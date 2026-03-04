#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SERVICES_DIR="$HOME/GoProjects/services"
SERVICES_EL="$SERVICES_DIR/src/el"
AOR_DIR="$HOME/GoProjects/agent-orchestrator"
MOCK_DIR="$SCRIPT_DIR/demo-mock-server"
MCP_DIR="$SERVICES_EL/apps/connected_care/local_mcp_server"
CHAT_UI_DIR="$SCRIPT_DIR/chat-ui"
MCP_TEST_UI_DIR="$SCRIPT_DIR/mcp-test-ui"
VENV="$AOR_DIR/.hackathon-venv"

PIDS_TO_KILL=()

cleanup() {
  if [[ ${#PIDS_TO_KILL[@]} -gt 0 ]]; then
    echo ""
    echo "Shutting down..."
    for pid in "${PIDS_TO_KILL[@]}"; do
      kill "$pid" 2>/dev/null && echo "  Stopped pid $pid"
    done
    echo "  (AOR containers left running — stop with: cd $AOR_DIR && docker compose down)"
  fi
}
trap cleanup EXIT

kill_port() {
  local port=$1
  local pids
  pids=$(lsof -ti :"$port" 2>/dev/null || true)
  if [[ -n "$pids" ]]; then
    echo "$pids" | xargs kill -9 2>/dev/null || true
    sleep 1
  fi
}

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Hackathon Agent Documents — Full Stack Launcher"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# ── 1. Mock extension server (:8092) ──
echo "[..] Rebuilding mock server..."
kill_port 8092
(cd "$MOCK_DIR" && go build -o demo-mock-server .)
"$MOCK_DIR/demo-mock-server" &
PIDS_TO_KILL+=($!)
sleep 1
echo "[OK] Mock server on :8092"

# ── 2. Documents MCP server (:18006) ──
echo "[..] Rebuilding Documents MCP server..."
kill_port 18006
(cd "$SERVICES_EL" && go build -o "$MCP_DIR/local-mcp-server" ./apps/connected_care/local_mcp_server/)
(cd "$MCP_DIR" && ./local-mcp-server) &
PIDS_TO_KILL+=($!)
sleep 2
echo "[OK] Documents MCP server on :18006 (unified gateway: health_records + messaging)"

# ── 3. AOR containers ──
if docker ps --format '{{.Names}}' 2>/dev/null | grep -q agent-orchestrator-agent-orchestrator; then
  echo "[OK] AOR containers already running"
else
  echo "[..] Starting AOR containers..."
  cp ~/.config/gcloud/application_default_credentials.json /tmp/gcloud_adc.json 2>/dev/null || true
  (cd "$AOR_DIR" && docker compose up -d agent-orchestrator mongodb jaeger mocked-mcp-server otel-collector 2>&1 | tail -3)
  echo "[..] Waiting for AOR health check..."
  for i in $(seq 1 30); do
    if curl -s http://localhost:8080/healthz 2>/dev/null | grep -q ok; then
      break
    fi
    sleep 2
  done
fi

if curl -s http://localhost:8080/healthz 2>/dev/null | grep -q ok; then
  echo "[OK] AOR healthy on :8080"
else
  echo "[!!] AOR not responding — check: docker logs agent-orchestrator-agent-orchestrator-1"
  exit 1
fi

# ── 4. Chat UI (:8093) ──
echo "[..] Rebuilding Chat UI..."
kill_port 8093
(cd "$CHAT_UI_DIR" && go build -o chat-ui-server .)
"$CHAT_UI_DIR/chat-ui-server" &
PIDS_TO_KILL+=($!)
sleep 1
echo "[OK] Chat UI on :8093"

# ── 5. MCP Test UI (:8091) ──
echo "[..] Rebuilding MCP Test UI..."
kill_port 8091
(cd "$MCP_TEST_UI_DIR" && go build -o mcp-test-ui .)
"$MCP_TEST_UI_DIR/mcp-test-ui" &
PIDS_TO_KILL+=($!)
sleep 1
echo "[OK] MCP Test UI on :8091"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  All services ready!"
echo ""
echo "  Chat UI:     http://localhost:8093"
echo "  MCP Test UI: http://localhost:8091"
echo "  Jaeger:      http://localhost:16686"
echo ""
echo "  Demo scripts:"
echo "    $SCRIPT_DIR/demo-scripts/demo-1-autonomous-thinking.sh"
echo "    $SCRIPT_DIR/demo-scripts/demo-2-privacy-consent.sh"
echo "    $SCRIPT_DIR/demo-scripts/demo-3-problem-solving.sh"
echo ""
echo "  Terminal chat:"
echo "    $VENV/bin/python $SCRIPT_DIR/chat.py"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Press Ctrl+C to stop local servers."
echo ""
wait
