# Hackathon Demo Mock Documents Server

Standalone mock server that mimics the documents backend extension API
with rich, realistic data designed for the hackathon demo.

## Quick Start

```bash
CGO_ENABLED=0 go build -o demo-mock-server .
./demo-mock-server
# Listening on :8090
```

## Endpoints

| Method | Path | Returns |
|--------|------|---------|
| GET | `/v1/document-filters` | 4 filter categories (Statements, Bills, Plan Docs, Uploads) |
| POST | `/v1/documents/search` | Documents filtered by category |
| GET | `/v1/documents/content?documentId=X` | Full document text content |

## Demo Documents

| ID | Title | Key Data |
|----|-------|----------|
| `doc-stmt-001` | January 2026 Monthly Statement | Copay: $25, Total: $35 |
| `doc-stmt-002` | December 2025 Monthly Statement | Copay: $50 |
| `doc-bill-001` | February 2026 Premium Bill | $485, UNPAID |
| `doc-bill-002` | January 2026 Premium Bill | $485, PAID |
| `doc-plan-001` | 2026 Summary of Benefits | Deductible: $1,500 |

## Connected Care Config

Point connected_care's extension config at this server:

```toml
[ExtensionsRouterConfig]
useTestingApis = true

[[ExtensionsRouterConfig.ConnectorConfigs]]
internal = true
tenantId = "league"
baseUrl = "http://localhost:8090"
authType = "none"
```

## Demo Story

1. Agent calls `get_health_document_filters` → gets 4 categories
2. Agent calls `search_health_documents(filter=statements)` → gets 3 statements
3. User: "How much was my copay?" → Agent asks for consent
4. Agent calls `read_document_content(id=doc-stmt-001)` → gets full statement text
5. Agent: "Your copay was $25.00 for your visit with Dr. Sarah Smith on January 15, 2026"
