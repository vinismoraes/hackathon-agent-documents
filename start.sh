#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SERVICES_DIR="$HOME/GoProjects/services"
AOR_DIR="$HOME/GoProjects/agent-orchestrator"
MOCK_DIR="$SCRIPT_DIR/demo-mock-server"
MCP_DIR="$SERVICES_DIR/src/el/apps/connected_care/local_mcp_server"
MSG_MCP_DIR="$SERVICES_DIR/src/el/apps/messaging/local_mcp_server"
CHAT_UI_DIR="$SCRIPT_DIR/chat-ui"
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

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Hackathon Agent Documents — Full Stack Launcher"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 1. Mock server
if lsof -ti :8092 > /dev/null 2>&1; then
  echo "[OK] Mock server already running on :8092"
else
  echo "[..] Starting mock server on :8092..."
  if [[ ! -f "$MOCK_DIR/demo-mock-server" ]]; then
    (cd "$MOCK_DIR" && CGO_ENABLED=0 go build -o demo-mock-server .)
  fi
  "$MOCK_DIR/demo-mock-server" &
  PIDS_TO_KILL+=($!)
  sleep 1
  echo "[OK] Mock server started"
fi

# 2. MCP server
if lsof -ti :18006 > /dev/null 2>&1; then
  echo "[OK] MCP server already running on :18006"
else
  echo "[..] Starting MCP server on :18006..."
  (cd "$MCP_DIR" && CGO_ENABLED=0 go build -o local-mcp-server . && ./local-mcp-server) &
  PIDS_TO_KILL+=($!)
  sleep 2
  echo "[OK] MCP server started"
fi

# 3. Messaging MCP server
if lsof -ti :18005 > /dev/null 2>&1; then
  echo "[OK] Messaging MCP server already running on :18005"
else
  echo "[..] Starting Messaging MCP server on :18005..."
  (cd "$MSG_MCP_DIR" && CGO_ENABLED=0 go build -o local-messaging-mcp . && ./local-messaging-mcp) &
  PIDS_TO_KILL+=($!)
  sleep 2
  echo "[OK] Messaging MCP server started"
fi

# 4. AOR containers (renumbered)
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

# 5. Chat UI
if lsof -ti :8093 > /dev/null 2>&1; then
  echo "[OK] Chat UI already running on :8093"
else
  echo "[..] Starting Chat UI on :8093..."
  if [[ ! -f "$CHAT_UI_DIR/chat-ui-server" ]]; then
    (cd "$CHAT_UI_DIR" && CGO_ENABLED=0 go build -o chat-ui-server .)
  fi
  "$CHAT_UI_DIR/chat-ui-server" &
  PIDS_TO_KILL+=($!)
  sleep 1
  echo "[OK] Chat UI started"
fi

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
if [[ ${#PIDS_TO_KILL[@]} -gt 0 ]]; then
  echo "Press Ctrl+C to stop local servers."
  echo ""
  wait
else
  echo "All services were already running. Open the Chat UI and go!"
  echo ""
fi
