# AI Agent Hackathon - Document MCP Tools

Hackathon project: MCP tools that give the AI agent the ability to find,
navigate, and read health documents through chat.

## Contents

- **[PLAN.md](PLAN.md)** -- Full hackathon plan with architecture diagrams,
  day-by-day checklist, demo strategy, and talking points
- **[demo-mock-server/](demo-mock-server/)** -- Standalone mock documents
  extension server with realistic demo data

## Quick Start

### Run the demo mock server

```bash
cd demo-mock-server
CGO_ENABLED=0 go build -o demo-mock-server .
./demo-mock-server
# Listening on :8090
```

### Test it

```bash
curl http://localhost:8090/v1/document-filters | jq
curl -X POST http://localhost:8090/v1/documents/search \
  -H "Content-Type: application/json" \
  -d '{"included":[{"attributes":{"filterKey":"statements"}}]}' | jq
curl "http://localhost:8090/v1/documents/content?documentId=doc-stmt-001"
```

## The Story

> "I gave our AI agent the ability to help users find and understand their
> health documents -- and with consent, actually read them and answer
> questions. Built in two days with Cursor Agent. No AOR code changes.
> No frontend changes."

See [PLAN.md](PLAN.md) for the full details.
