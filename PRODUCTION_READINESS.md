# Production Readiness: Document & Messaging MCP Tools

This document explains the hackathon demo architecture, what is mocked vs. real,
and what is required to ship these tools to production.

---

## Architecture: Demo vs. Production

### Demo (Current State)

```mermaid
graph TD
  subgraph Browser
    UI["Chat UI<br/>localhost:8093"]
  end

  subgraph Proxy["Chat UI Go Proxy :8093"]
    JWT["Dev-secret JWT generator"]
    SKIP["Sets X-Skip-Guardrails: true<br/>Sets User-Agent: Local-CLI"]
    DOCPROXY["Proxies /doc-content/ to mock"]
  end

  subgraph Docker["AOR Docker :8080"]
    GC["GuardedConversation<br/>⚠️ SKIPPED"]
    LLM["Concierge Agent<br/>Gemini LLM"]
    TS1["ConciergeDocumentsMcpToolset"]
    TS2["ConciergeMessagingMcpToolset"]
    TS3["ConciergeUIMcpToolset"]
  end

  subgraph CCMCP["Connected Care MCP :18006"]
    CC_TOOLS["get_health_document_filters<br/>search_health_documents<br/>read_document_content<br/>get_health_document_link"]
    CC_CALLER["⚠️ MOCK directExtensionsCaller<br/>Direct HTTP, no OAuth"]
  end

  subgraph MSGMCP["Messaging MCP :18005"]
    MSG_TOOLS["search_conversation_attachments<br/>get_attachment_link"]
    MSG_SVC["⚠️ MOCK IMessagingService<br/>In-memory: 3 threads, 4 attachments"]
  end

  subgraph MOCK["Demo Mock Server :8092"]
    MOCK_API["⚠️ Hardcoded JSON responses<br/>/v1/documents/filters<br/>/v1/documents/search<br/>/v1/documents/content"]
  end

  UI -->|"SSE stream"| Proxy
  Proxy -->|"/v1/run-test-agent"| Docker
  LLM --> TS1 & TS2 & TS3
  TS1 -->|"MCP over HTTP"| CCMCP
  TS2 -->|"MCP over HTTP"| MSGMCP
  CC_TOOLS --> CC_CALLER
  CC_CALLER -->|"Raw HTTP"| MOCK
  MSG_TOOLS --> MSG_SVC
```

### Production (Target State)

```mermaid
graph TD
  subgraph Client["League Mobile / Web App"]
    APP["Chat UI Component"]
  end

  subgraph AOR["AOR on GKE"]
    GC_PROD["GuardedConversation<br/>✅ ACTIVE"]
    LLM_PROD["Concierge Agent<br/>Gemini LLM"]
    TS1_PROD["ConciergeDocumentsMcpToolset"]
    TS2_PROD["ConciergeMessagingMcpToolset"]
  end

  subgraph CC_PROD["Connected Care Service on GKE"]
    CC_MCP_PROD["MCP Server embedded"]
    CC_TOOLS_PROD["get_health_document_filters<br/>search_health_documents<br/>read_document_content<br/>get_health_document_link"]
    CC_CALLER_PROD["✅ Real ExtensionsCaller<br/>OAuth + tenant routing"]
  end

  subgraph MSG_PROD["Messaging Service on GKE"]
    MSG_MCP_PROD["MCP Server embedded"]
    MSG_TOOLS_PROD["search_conversation_attachments<br/>get_attachment_link"]
    MSG_SVC_PROD["✅ Real IMessagingService<br/>MongoDB backed, per-user"]
  end

  subgraph EXT["Tenant Extension Backend"]
    EXT_API["Real document storage<br/>Cloud storage / EHR"]
  end

  APP -->|"Real OAuth token"| AOR
  GC_PROD --> LLM_PROD
  LLM_PROD --> TS1_PROD & TS2_PROD
  TS1_PROD -->|"Internal K8s network"| CC_PROD
  TS2_PROD -->|"Internal K8s network"| MSG_PROD
  CC_TOOLS_PROD --> CC_CALLER_PROD
  CC_CALLER_PROD -->|"Authenticated call"| EXT
  MSG_TOOLS_PROD --> MSG_SVC_PROD
```

### Request Flow: Search Conversation Attachments

```mermaid
sequenceDiagram
    participant User as User / Chat UI
    participant AOR as AOR (Concierge Agent)
    participant MCP as Messaging MCP Server
    participant SVC as IMessagingService

    User->>AOR: "What attachments are in my conversations?"
    AOR->>AOR: LLM selects search_conversation_attachments
    AOR->>MCP: tools/call search_conversation_attachments
    MCP->>SVC: GetThreads(limit=10)
    SVC-->>MCP: 3 threads
    loop For each thread
        MCP->>SVC: GetMessages(threadId)
        SVC-->>MCP: Messages with Documents[]
    end
    MCP-->>AOR: 4 attachments found
    AOR->>AOR: LLM formats response + render_card
    AOR-->>User: Cards with file names, sizes, download buttons
```

### Request Flow: Consent-Gated Document Read

```mermaid
sequenceDiagram
    participant User as User / Chat UI
    participant AOR as AOR (Concierge Agent)
    participant MCP as Connected Care MCP
    participant EXT as Extensions Backend

    User->>AOR: "Summarize my January statement"
    AOR->>AOR: LLM selects read_document_content
    AOR->>AOR: Tool description requires consent first
    AOR->>MCP: tools/call ui_chip_list_card
    MCP-->>AOR: Chips rendered
    AOR-->>User: "I need to read this document.<br/>Would you like to proceed?"<br/>[Yes, read it] [No, thanks]
    User->>AOR: "Yes, read it"
    AOR->>MCP: tools/call read_document_content(docId)
    MCP->>EXT: GET /v1/documents/content?documentId=...
    EXT-->>MCP: Document bytes
    MCP-->>AOR: Document text content
    AOR->>AOR: LLM summarizes content
    AOR-->>User: Summary of the document
```

---

## What Is Mocked (and Why)

| Component | Demo | Production | Why Mocked |
|-----------|------|------------|------------|
| **ExtensionsCaller** | `directExtensionsCaller` — direct HTTP to mock server, no auth | Real caller with OAuth, Redis token cache, tenant routing | Avoids needing GCP credentials, extension infra, and tenant config |
| **Mock Extension Server** | Go binary returning hardcoded JSON | Real tenant backend (cloud storage, EHR systems) | No access to real patient documents or tenant backends |
| **IMessagingService** | In-memory struct with 3 threads, 4 attachments | Real service backed by MongoDB with per-user data | Avoids needing the full messaging stack (MongoDB, participants, auth) |
| **JWT / Auth** | `dev-secret` HMAC token, hardcoded `demo-user` | Real OAuth/OIDC tokens from identity provider | No access to identity provider locally |
| **GuardedConversation** | Skipped via `X-Skip-Guardrails` + `Local-CLI` UA | Active — safety/escalation service validates all messages | Escalation service not available locally |
| **MCP Auth** | `WithNoAuth()` on all tools | Should use `WithAuth()` + token forwarding from AOR | No real user context in demo |
| **API Hostname** | Hardcoded `api.league.com` / `localhost:5400` | From service config / environment | No real API gateway locally |

---

## Production Checklist

### 1. Authentication & Authorization

- [ ] **Remove `WithNoAuth()`** from all new tools (`search_conversation_attachments`, `get_attachment_link`)
- [ ] **Replace with `WithAuth()`** to require and validate the MCP request's auth context
- [ ] **Ensure user-scoped data access** — `IMessagingService.GetThreads()` must only return threads belonging to the authenticated user (already the case in production since the service extracts user ID from the request context)
- [ ] **Verify AOR forwards auth tokens** to MCP servers correctly (currently handled by `MCP_AUTH_CREDENTIAL` / `MCP_AUTH_SCHEME` in the toolset)

### 2. Connected Care: Remove Mock Bypass

- [ ] **Remove `local_mcp_server` binary** — or keep it for dev/test only (already flagged with `// HACKATHON`)
- [ ] **Remove `directExtensionsCaller`** — production uses the real `ExtensionsCaller` via Wire injection
- [ ] **Remove demo-mock-server** — no longer needed when real extension backends are available
- [ ] **Verify `ExtensionsCaller` OAuth flow** works with the MCP server's request context
- [ ] **Remove hardcoded `APIHostname`** in `messaging.go` — use proper config (e.g., from `connected_care.toml`)

### 3. Messaging: Wire Into Real Service

- [ ] **Remove `local_mcp_server` binary** for messaging (already flagged)
- [ ] The MCP server is **already wired** into the real messaging service via `conf/messaging.go` and Wire — it receives the real `IMessagingService` implementation
- [ ] **Verify `GetThreads` / `GetMessages` pagination** — demo uses defaults, production may need cursor-based pagination for large thread lists
- [ ] **Add `APIHostname` to messaging config** — currently hardcoded as `"api.league.com"` in `conf/messaging.go`

### 4. AOR Configuration

- [ ] **Move MCP tool entries from `config.toml` to tenant registry** — currently hardcoded with `host.docker.internal` URLs
- [ ] **Update URLs** to use internal Kubernetes service names (e.g., `http://connected-care:18006/mcp`)
- [ ] **Remove `ConciergeMessagingMcpToolset`** file if messaging tools ship under the existing `ConciergeUIMcpToolset` (since they share the same MCP server in production)
- [ ] **Remove `X-Skip-Guardrails` bypass** — let GuardedConversation run normally

### 5. Privacy & Consent

- [ ] **Review `read_document_content` consent flow** — the tool description instructs the LLM to ask for consent via chips before reading document content. Validate this is sufficient for compliance or add server-side enforcement.
- [ ] **Audit document content access logging** — ensure all `read_document_content` calls are logged with user ID, document ID, and consent status for HIPAA audit trail
- [ ] **Evaluate data retention** — confirm that document content passed through the LLM is not stored beyond the session

### 6. Observability

- [ ] **Verify OpenTelemetry tracing** propagates from AOR → MCP server → extension backend (already instrumented in the MCP framework)
- [ ] **Add metrics** for tool call latency, error rates, and attachment counts
- [ ] **Set up alerts** for MCP server health and tool execution failures

### 7. Known Issue: UI Tool Calling Loop

During demo testing, the LLM occasionally enters a loop calling `render_card` and
`render_chips` repeatedly. This is an **AOR agent-loop issue**, not a client or MCP
tool issue.

```mermaid
sequenceDiagram
    participant LLM as LLM (Gemini)
    participant AOR as AOR Agent Loop
    participant MCP as MCP Server

    LLM->>AOR: Call render_chips(labels)
    AOR->>MCP: tools/call render_chips
    MCP-->>AOR: Response with chips payload
    AOR-->>LLM: function_response + chips in message
    Note over LLM: LLM does not consider<br/>render_chips "resolved"
    LLM->>AOR: Call render_chips(labels) again
    Note over AOR: Loop continues until<br/>max iterations or timeout
```

**Root cause:** The LLM sometimes doesn't treat the `render_chips` / `render_card`
function response as a terminal action, and re-invokes the tool in the next agent
loop iteration.

**Demo workaround:** The chat UI detects when cards or chips arrive and visually
marks the corresponding tool as "Done" to prevent the UI from showing duplicates.
This is cosmetic — the loop still happens server-side.

**Production fix needed (AOR level):**
- [ ] Investigate whether AOR's agent loop should treat `render_card` / `render_chips`
  as terminal tools that end the current turn (similar to how a final text response
  ends the turn)
- [ ] Alternatively, add a `max_consecutive_tool_calls` guard in AOR to break out
  of loops after N identical tool calls
- [ ] Consider whether the tool response schema needs adjustment so the LLM
  reliably recognizes completion

**Impact on Chathub:** This same loop would reproduce in the production Chathub UI
since the behavior originates in the AOR agent loop, not the client. The fix must
be applied at the AOR or tool-description level before production rollout.

### 8. Code Cleanup

All hackathon-specific code is tagged with `// HACKATHON` comments. To find every instance:

```bash
grep -rn "HACKATHON" src/el/connected_care/ src/el/messaging/ src/el/apps/
```

Key items to address:
- [ ] Remove or gate `// HACKATHON: replace with proper auth` markers
- [ ] Remove `// HACKATHON: use proper hostname config` in `messaging.go`
- [ ] Remove `local_mcp_server` binaries (or move to a `tools/` or `cmd/` directory for dev use)
- [ ] Run `wire` to regenerate `wire_gen.go` properly (currently hand-edited)

---

## Files Changed (services repo)

### New Files (Hackathon Only — Remove Before Production)
```
apps/connected_care/local_mcp_server/main.go    ← demo binary
apps/messaging/local_mcp_server/main.go          ← demo binary
```

### New Files (Production Code)
```
connected_care/mcp/mcp_server.go                 ← CC MCP server
connected_care/mcp/tools/get_health_document_filters/
connected_care/mcp/tools/search_health_documents/
connected_care/mcp/tools/get_health_document_link/
connected_care/mcp/tools/read_document_content/
messaging/chathub/mcp/doc_tools/search_conversation_attachments/
messaging/chathub/mcp/doc_tools/get_attachment_link/
```

### Modified Files (Need HACKATHON Markers Removed)
```
messaging/chathub/mcp/mcp_server.go              ← added IMessagingService + doc tools
messaging/conf/messaging.go                       ← hardcoded APIHostname
messaging/conf/wire_gen.go                        ← hand-edited (re-run wire)
```
