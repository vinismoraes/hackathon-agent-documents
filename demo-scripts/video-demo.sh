#!/usr/bin/env bash
# Video Demo — Document Concierge
#
# Self-paced walkthrough for recording. Press Enter to advance each step.
# Shows the incremental story: 1 source → 2 sources → 3 sources.
set -euo pipefail
source "$(dirname "$0")/common.sh"

ADMIN_URL="http://localhost:18007"

toggle_source() {
  curl -s -X POST "$ADMIN_URL/sources/toggle" \
    -H "Content-Type: application/json" \
    -d "{\"id\":\"$1\",\"enabled\":$2}" > /dev/null
}

banner() {
  echo ""
  echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
  echo -e "${BOLD}  $1${RESET}"
  echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
  echo ""
}

narrate() {
  echo -e "${DIM}$1${RESET}"
}

check_services
JWT=$(generate_jwt)

# ── Reset: only health_records on ──
toggle_source "messaging" "false"
toggle_source "benefits" "false"

# ═══════════════════════════════════════════════════════
banner "ACT 1 — One Source: Health Documents"
# ═══════════════════════════════════════════════════════

narrate "Only health_records is enabled. The agent discovers and searches one source."
pause

SESSION=$(create_session)
echo -e "${BOLD}1.1  Discovery — \"What do I have?\"${RESET}"
send_message "$SESSION" "What health documents do I have?"
pause

echo -e "${BOLD}1.2  Filtered search — natural language maps to a filter${RESET}"
send_message "$SESSION" "Show me my monthly statements"
pause

echo -e "${BOLD}1.3  Consent-gated read — agent asks before reading sensitive content${RESET}"
send_message "$SESSION" "Read my January 2026 statement and summarize it"
pause

echo -e "${BOLD}1.4  Grant consent — agent reads and extracts key insights${RESET}"
send_message "$SESSION" "Yes, go ahead"
pause

# ═══════════════════════════════════════════════════════
banner "ACT 2 — Two Sources: + Conversation Attachments"
# ═══════════════════════════════════════════════════════

narrate "We enable messaging. Same 5 tools, but now they route to 2 backends."
narrate "No code changes — just toggle the source on."
toggle_source "messaging" "true"
pause

SESSION=$(create_session)
echo -e "${BOLD}2.1  Multi-source discovery${RESET}"
send_message "$SESSION" "What documents or files do I have?"
pause

echo -e "${BOLD}2.2  Source-specific search — agent routes to messaging${RESET}"
send_message "$SESSION" "Do I have a medication guide in my messages?"
pause

# ═══════════════════════════════════════════════════════
banner "ACT 3 — Three Sources: + Benefits & Claims"
# ═══════════════════════════════════════════════════════

narrate "Benefits & Claims enabled. 3 backends, still 5 tools."
toggle_source "benefits" "true"
pause

SESSION=$(create_session)
echo -e "${BOLD}3.1  Three-source discovery${RESET}"
send_message "$SESSION" "What documents or files do I have?"
pause

echo -e "${BOLD}3.2  Benefits-specific search${RESET}"
send_message "$SESSION" "What benefits documents do I have?"
pause

# ═══════════════════════════════════════════════════════
banner "Demo Complete"
# ═══════════════════════════════════════════════════════

echo "Key takeaways:"
echo "  • 5 generic tools serve any number of document backends"
echo "  • New sources plug in with zero tool changes"
echo "  • LLM autonomously routes queries to the right source"
echo "  • Privacy: consent-gated access to sensitive content"
echo "  • Observability: automatic Prometheus metrics on every tool call"
echo ""
