#!/usr/bin/env bash
# Demo 1: Agent Thinks By Itself
#
# Shows the agent autonomously chaining multiple tool calls
# without being told which tools to use.
set -euo pipefail
source "$(dirname "$0")/common.sh"

echo -e "${BOLD}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Demo 1: The Agent Thinks By Itself"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${RESET}"
echo "This demo shows the agent autonomously discovering and"
echo "chaining tools without being told what to do."
echo ""

check_services

JWT=$(generate_jwt)
SESSION=$(create_session)

echo -e "${BOLD}Step 1: Vague request — agent must figure out what to do${RESET}"
echo "We just ask a broad question. The agent should:"
echo "  1. Call get_health_document_filters to discover categories"
echo "  2. Realize it needs to search and call search_health_documents"
echo "  3. Present results with UI cards"
pause

send_message "$SESSION" "What health documents do I have available?"

echo -e "${BOLD}Step 2: Follow-up — agent uses context from previous turn${RESET}"
echo "The agent should remember the filter categories and search"
echo "the specific one we ask about."
pause

send_message "$SESSION" "Show me my premium bills"

echo -e "${BOLD}Step 3: Complex request — agent chains multiple tools${RESET}"
echo "A multi-part request that requires the agent to plan,"
echo "search, and synthesize."
pause

send_message "$SESSION" "I need to compare my November and December 2025 statements. Can you find them?"

echo -e "\n${GREEN}${BOLD}Demo 1 Complete!${RESET}"
echo "Key takeaway: The agent autonomously selected the right tools,"
echo "chained them together, and maintained context across turns."
