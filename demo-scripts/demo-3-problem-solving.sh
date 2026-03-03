#!/usr/bin/env bash
# Demo 3: Effective Problem Solving
#
# Shows the agent solving a real user problem end-to-end:
# finding, reading, and explaining health documents.
set -euo pipefail
source "$(dirname "$0")/common.sh"

echo -e "${BOLD}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Demo 3: Solving Real Problems End-to-End"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${RESET}"
echo "This demo shows a realistic user scenario:"
echo "\"I got a bill I don't understand. Help me figure it out.\""
echo ""

check_services

JWT=$(generate_jwt)
SESSION=$(create_session)

echo -e "${BOLD}Step 1: User has a problem — agent needs to help${RESET}"
echo "A natural, conversational request."
pause

send_message "$SESSION" "I just got charged for something on my health plan and I don't understand why. Can you help me look into it?"

echo -e "${BOLD}Step 2: User provides more context${RESET}"
echo "The agent should narrow down the search."
pause

send_message "$SESSION" "I think it was from January. Can you check my January statement?"

echo -e "${BOLD}Step 3: Agent reads and explains${RESET}"
echo "Consent + read + plain-language explanation."
pause

send_message "$SESSION" "Yes, go ahead and read it. I want to understand what I was charged for."

echo -e "${BOLD}Step 4: Follow-up question${RESET}"
echo "The agent should answer from the document it already read."
pause

send_message "$SESSION" "Is there anything I should follow up on with my provider?"

echo -e "\n${GREEN}${BOLD}Demo 3 Complete!${RESET}"
echo "Key takeaway: The agent solved a real user problem by"
echo "autonomously searching, reading, and explaining documents"
echo "in plain language — no technical knowledge required."
