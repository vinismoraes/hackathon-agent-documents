#!/usr/bin/env bash
# Tests the demo progression: health-only → +messaging → +benefits
# Verifies prompts route correctly at each stage.
set -euo pipefail
source "$(dirname "$0")/common.sh"

ADMIN_URL="http://localhost:18007"

check_services

JWT=$(generate_jwt)

run_prompt() {
  local prompt="$1"
  local session
  session=$(create_session)

  local resp
  resp=$(curl -s -N -X POST "$AOR_URL/v1/run-test-agent" \
    -H "Authorization: Bearer $JWT" \
    -H "Content-Type: application/json" \
    -H "User-Agent: Local-CLI" \
    -H "X-Skip-Guardrails: true" \
    -d "{\"id\":\"$(uuidgen)\",\"session_id\":\"$session\",\"message\":\"$prompt\"}")

  echo "$resp" | python3 -c "
import sys, json, re
raw = sys.stdin.read()
tools, text_parts, cards = [], [], []
for m in re.finditer(r'^data: ({.*})\$', raw, re.MULTILINE):
    try:
        data = json.loads(m.group(1))
        for part in data.get('message', {}).get('parts', []):
            fc = part.get('function_call')
            fr = part.get('function_response')
            if fc: tools.append(fc.get('name',''))
            if fr:
                resp = fr.get('response', {})
                if isinstance(resp, str):
                    try: resp = json.loads(resp)
                    except: pass
                sc = resp.get('result', {}).get('structuredContent', {})
                for doc in sc.get('documents', []):
                    cards.append(doc.get('title',''))
            if part.get('text'): text_parts.append(part['text'])
    except: pass
print(f'    Tools: {tools}')
txt = ' '.join(text_parts)[:400]
print(f'    Text: {txt}')
if cards: print(f'    Cards: {cards[:5]}')
"
}

set_sources() {
  local health="$1" messaging="$2" benefits="$3"
  curl -s -X POST "$ADMIN_URL/sources/toggle" -H 'Content-Type: application/json' -d "{\"id\":\"health_records\",\"enabled\":$health}" > /dev/null
  curl -s -X POST "$ADMIN_URL/sources/toggle" -H 'Content-Type: application/json' -d "{\"id\":\"messaging\",\"enabled\":$messaging}" > /dev/null
  curl -s -X POST "$ADMIN_URL/sources/toggle" -H 'Content-Type: application/json' -d "{\"id\":\"benefits\",\"enabled\":$benefits}" > /dev/null
  echo -e "  ${DIM}Sources: health=$health messaging=$messaging benefits=$benefits${RESET}"
}

echo -e "${BOLD}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Toggle Progression Test — Document Concierge Scalability"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${RESET}"

# ── Phase 1: Health only ──
echo -e "${BOLD}═══ Phase 1: Health Documents only ═══${RESET}"
set_sources true false false
echo ""

echo -e "  ${CYAN}Prompt: What health documents do I have?${RESET}"
run_prompt "What health documents do I have?"
echo ""

echo -e "  ${CYAN}Prompt: What documents or files do I have? (broad — only health exists)${RESET}"
run_prompt "What documents or files do I have?"
echo ""

echo -e "  ${CYAN}Prompt: What files were shared in my conversations? (messaging OFF)${RESET}"
run_prompt "What files were shared in my conversations?"
echo ""

# ── Phase 2: Health + Messaging ──
echo -e "${BOLD}═══ Phase 2: + Conversation Attachments ═══${RESET}"
set_sources true true false
echo ""

echo -e "  ${CYAN}Prompt: What documents or files do I have? (now 2 sources)${RESET}"
run_prompt "What documents or files do I have?"
echo ""

echo -e "  ${CYAN}Prompt: What files were shared in my conversations? (messaging ON)${RESET}"
run_prompt "What files were shared in my conversations?"
echo ""

# ── Phase 3: All three ──
echo -e "${BOLD}═══ Phase 3: + Benefits & Claims ═══${RESET}"
set_sources true true true
echo ""

echo -e "  ${CYAN}Prompt: What documents or files do I have? (all 3 sources)${RESET}"
run_prompt "What documents or files do I have?"
echo ""

echo -e "  ${CYAN}Prompt: What benefits documents do I have?${RESET}"
run_prompt "What benefits documents do I have?"
echo ""

echo -e "${GREEN}${BOLD}Progression test complete.${RESET}"
