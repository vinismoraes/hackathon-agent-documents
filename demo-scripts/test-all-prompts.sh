#!/usr/bin/env bash
# Automated test harness for all welcome screen prompts + benefits.
# Verifies tool routing, text output, and card rendering via the AOR API.
set -euo pipefail
source "$(dirname "$0")/common.sh"

check_services

JWT=$(generate_jwt)

run_prompt() {
  local num="$1"
  local prompt="$2"

  local session
  session=$(create_session)

  echo -e "${BOLD}--- PROMPT $num: $prompt ---${RESET}"

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
tools, text_parts, cards, chips = [], [], [], []
for m in re.finditer(r'^data: ({.*})\$', raw, re.MULTILINE):
    try:
        data = json.loads(m.group(1))
        for part in data.get('message', {}).get('parts', []):
            fc = part.get('function_call')
            fr = part.get('function_response')
            if fc:
                tools.append(fc.get('name',''))
                if 'chip' in fc.get('name','').lower():
                    args = fc.get('args', {})
                    if isinstance(args, str):
                        try: args = json.loads(args)
                        except: pass
                    for o in args.get('options', []):
                        chips.append(o.get('label',''))
            if fr:
                resp = fr.get('response', {})
                if isinstance(resp, str):
                    try: resp = json.loads(resp)
                    except: pass
                sc = resp.get('result', {}).get('structuredContent', {})
                for doc in sc.get('documents', []):
                    cards.append(doc.get('title',''))
            if part.get('text'):
                text_parts.append(part['text'])
    except:
        pass
print(f'  Tools: {tools}')
txt = ' '.join(text_parts)[:500]
print(f'  Text: {txt}')
if cards:
    print(f'  Cards: {cards[:6]}')
if chips:
    print(f'  Chips: {chips}')
print()
"
}

echo -e "${BOLD}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Test All Prompts — Document Gateway"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${RESET}"

run_prompt 1 "What documents or files do I have?"
run_prompt 2 "What health documents do I have?"
run_prompt 3 "Show me my monthly statements"
run_prompt 4 "Read my January 2026 statement and summarize it"
run_prompt 5 "I want to upload a health document"
run_prompt 6 "What files were shared in my conversations?"
run_prompt 7 "Do I have a medication guide in my messages?"
run_prompt 8 "What benefits documents do I have?"

echo -e "${GREEN}${BOLD}All 8 prompts tested.${RESET}"
