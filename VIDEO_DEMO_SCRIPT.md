# Video Demo Script — Document Concierge

Run `./demo-scripts/video-demo.sh` and follow along. Each step waits for Enter.

---

## ACT 1 — One Source (Health Documents)

> Only health_records is enabled. We show the core loop: discover → search → read.

### 1.1 Discovery
**Prompt:** "What health documents do I have?"

**What to say:** "We start with a single document source. The agent calls `search_documents`, finds 10 documents across 3 categories — Statements, Bills, Plan Documents — and offers to filter."

### 1.2 Filtered Search
**Prompt:** "Show me my monthly statements"

**What to say:** "The agent maps natural language to a filter and returns only the 4 monthly statements."

### 1.3 Consent-Gated Read
**Prompt:** "Read my January 2026 statement and summarize it"

**What to say:** "This is the privacy pattern. The agent finds the document but asks for explicit consent before reading the content. This is a HIPAA-aligned design — the agent never reads sensitive data without permission."

### 1.4 Grant Consent
**Prompt:** "Yes, go ahead"

**What to say:** "Now the agent calls `read_document`, gets the full statement content, and returns key insights — not a data dump. Things like: what you owe, what changed, what action to take."

---

## ACT 2 — Two Sources (+ Conversation Attachments)

> We toggle messaging on. Same 5 tools, zero code changes.

### 2.1 Multi-Source Discovery
**Prompt:** "What documents or files do I have?"

**What to say:** "We just enabled a second source — Conversation Attachments. The agent now discovers both sources and shows document counts for each. No new tools were added. The same `search_documents` tool routes to multiple backends."

### 2.2 Source-Specific Search
**Prompt:** "Do I have a medication guide in my messages?"

**What to say:** "The agent routes this query to the messaging source based on context. It finds the Medication Guide attachment. This is the LLM doing the routing — we didn't write routing logic."

---

## ACT 3 — Three Sources (+ Benefits & Claims)

> Benefits enabled. 3 backends, still 5 tools.

### 3.1 Three-Source Discovery
**Prompt:** "What documents or files do I have?"

**What to say:** "Third source added — Benefits & Claims. The discovery now shows all three with document counts. The architecture scales without tool proliferation."

### 3.2 Benefits Search
**Prompt:** "What benefits documents do I have?"

**What to say:** "The agent searches the benefits source and returns EOBs, claim summaries, and coverage letters. Each source can have different capabilities — search, read, upload, filters — and the tools adapt automatically."

---

## Closing Points

- **5 generic tools** replace what would have been 15+ source-specific tools
- **New sources** plug in by implementing a Go interface — no tool code changes
- **Privacy** is built in — consent-gated document access
- **Observability** is automatic — Prometheus metrics on every tool call, zero custom instrumentation
- **The LLM does the routing** — tool descriptions update dynamically based on enabled sources
