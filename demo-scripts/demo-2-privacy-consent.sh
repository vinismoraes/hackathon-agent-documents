#!/usr/bin/env bash
# Demo 2: Privacy-First Consent Flow
#
# Shows the agent asking for explicit user consent before
# reading sensitive health document content.
set -euo pipefail
source "$(dirname "$0")/common.sh"

echo -e "${BOLD}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Demo 2: Privacy-First — Consent-Gated Document Access"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${RESET}"
echo "This demo shows the agent respecting user privacy by"
echo "asking for explicit consent before reading document content."
echo ""

check_services

JWT=$(generate_jwt)
SESSION=$(create_session)

echo -e "${BOLD}Step 1: Search for documents${RESET}"
echo "First, we find some documents."
pause

send_message "$SESSION" "Find my monthly statements"

echo -e "${BOLD}Step 2: Ask the agent to read a document${RESET}"
echo "Now we ask the agent to read document content."
echo "Watch: the agent should ask for consent BEFORE reading."
pause

send_message "$SESSION" "Can you read my January 2026 statement and tell me what's in it?"

echo -e "${BOLD}Step 3: Grant consent${RESET}"
echo "Now we explicitly say yes. The agent will call"
echo "read_document_content and summarize the contents."
pause

send_message "$SESSION" "Yes, please go ahead and read it for me."

echo -e "${BOLD}Step 4: Follow-up on the content${RESET}"
echo "The agent should be able to answer questions about"
echo "the document it just read."
pause

send_message "$SESSION" "How much did I pay out of pocket that month?"

echo -e "\n${GREEN}${BOLD}Demo 2 Complete!${RESET}"
echo "Key takeaway: The agent gates sensitive data access behind"
echo "explicit user consent — a HIPAA-aligned privacy pattern."
