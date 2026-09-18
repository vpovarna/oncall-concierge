# oncall-concierge

A hands-on learning project for **Google ADK (Go)**, **A2A**, and **agent memory**.

Three Go ADK agents cooperate over A2A to help an on-call engineer. Two of them keep long-term memory. The project's main lesson: **memory is private to each agent; only messages cross the A2A boundary.**

---

## Architecture

```
            engineer
               │
               ▼
      ┌──────────────────┐
      │   Coordinator    │  memory: user preferences
      │   (:8000)        │
      └──────┬─────┬─────┘
         A2A │     │ A2A
             ▼     ▼
┌──────────────────┐  ┌──────────────────┐
│ Incident         │  │ Runbook Agent    │
│ Historian (:8001)│  │ (:8002)          │
│ memory: incidents│  │ no memory        │
└────────┬─────────┘  └──────────────────┘
         │
         ▼
  Postgres + pgvector (from M6)
```

Each agent is a separate binary with its own runner, session service and memory service.

---

## Repo structure

```
oncall-concierge/
├── go.mod
├── Makefile
├── docker-compose.yml
├── .env.example
├── cmd/                     # wiring only: config → services → agent → launcher
│   ├── coordinator/main.go
│   ├── historian/main.go
│   ├── runbook/main.go
│   └── seed/main.go
├── internal/
│   ├── agents/
│   │   ├── coordinator/     # agent.go, prompt.go
│   │   ├── historian/       # agent.go, prompt.go, tools.go
│   │   └── runbook/         # agent.go, prompt.go, tools.go
│   ├── memory/
│   │   ├── pgvector/        # memory.Service implementation + tests
│   │   └── ingest/          # when sessions are written to memory
│   ├── identity/            # user ID in context, propagation over A2A
│   └── platform/            # config, model factory, logging, telemetry
├── mockdata/                # incidents.json, runbooks.json
├── migrations/001_memory.sql
├── test/e2e/                # curl scripts and scripted scenarios
└── docs/adr/                # one decision record per milestone
```

**Rule:** `cmd/` contains only wiring. All logic lives in `internal/`, so it can be tested without starting servers.

---

## Responsibilities

| Component | Owns | Does not |
|---|---|---|
| **Coordinator** | User conversation, routing to remote agents, user-preference memory, passing identity on every A2A call | Hold domain knowledge |
| **Historian** | Incident memory; tools `search_incidents`, `record_incident` | Serve users directly |
| **Runbook** | Stateless tools `get_runbook(service)`, `list_runbooks` | Remember anything |
| **memory/pgvector** | Implementing ADK's memory interface (add session, search), scoped by app + user | Decide *when* to write |
| **memory/ingest** | When to write memory (per turn vs. session end), deduplication | Storage details |
| **identity** | Putting the user ID into the context; restoring it after an A2A hop | Authentication itself |
| **platform** | Config, model setup, structured logs, OpenTelemetry | Agent logic |

---

## Milestones

Don't move on until the milestone's definition of done (DoD) is met. Write a short ADR at the end of each one.

### M0 — Setup (½ day)
- [X] Init module, pin ADK version
- [X] Run the official ADK Go quickstart unchanged

**DoD:** the quickstart agent answers in the dev web UI.

### M1 — Runbook agent, local (½ day)
- [ ] Write tools as plain Go functions over `mockdata/runbooks.json`
- [ ] Unit-test the tools
- [ ] Wire them into an LLM agent

**DoD:** "How do I restart payments?" returns the mock runbook, and the tool call is visible in the UI's event view.

### M2 — Historian with in-memory memory (1 day)
- [ ] Runner with in-memory session service and memory service
- [ ] `ingest.OnSessionEnd` adds the session to memory
- [ ] Memory search tool on the agent

**DoD:** an incident described in session 1 is recalled in session 2. After a restart it's gone (verify this).

### M3 — A2A servers (½ day)
- [ ] Serve Historian (:8001) and Runbook (:8002) with the A2A launcher
- [ ] Save curl commands in `test/e2e/`

**DoD:** `curl` returns each Agent Card, and a raw A2A message gets a reply without the Coordinator.

### M4 — Coordinator (1 day)
- [ ] LLM agent with both remote agents as sub-agents (`remoteagent`), URLs from config
- [ ] Its own memory for user preferences

**DoD:** "Payments is timing out, what should I do?" combines past incidents and the runbook.

### M5 — Identity and sessions across A2A (1 day)
- [ ] Log which user ID and session the Historian sees on A2A calls
- [ ] Check whether follow-ups reuse the same remote session (A2A context ID)
- [ ] Implement identity propagation
- [ ] ADR-001: identity propagation approach

**DoD:** e2e test where Alice and Bob each record an incident and neither can retrieve the other's.

### M6 — Persistent memory with pgvector (1–2 days)
- [ ] Migration (below)
- [ ] `memory/pgvector.Service` with an `Embedder` interface (fake in tests)
- [ ] Tests against real Postgres (testcontainers)

**DoD:** memory survives a restart, and "connection pool" finds "DB connections exhausted."

### M7 — Hardening (1 day)
- [ ] Before-tool callback: block `record_incident` without a root cause
- [ ] OpenTelemetry spans across all three services
- [ ] 15-question eval script with expected answers

**DoD:** `make e2e` passes, and one question appears as one trace in Jaeger.

---

## Memory schema (M6)

```sql
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE memories (
  id         BIGSERIAL PRIMARY KEY,
  app_name   TEXT NOT NULL,
  user_id    TEXT NOT NULL,
  session_id TEXT NOT NULL,
  event_id   TEXT NOT NULL,
  author     TEXT,
  content    TEXT NOT NULL,
  embedding  vector(768),            -- match your embedding model
  created_at TIMESTAMPTZ DEFAULT now(),
  UNIQUE (session_id, event_id)      -- re-ingesting a session is safe
);

CREATE INDEX ON memories (app_name, user_id);
```

## Memory service sketch

```go
type Service struct {
    db    *pgxpool.Pool
    embed Embedder // interface, so tests can use a fake
}

// Compile-time check against the pinned ADK version.
var _ memory.Service = (*Service)(nil)

// Embed each event's text and upsert; ON CONFLICT DO NOTHING.
func (s *Service) AddSession(ctx context.Context, sess session.Session) error

// Embed query; ORDER BY embedding <=> $1 LIMIT 5, filtered by app_name + user_id.
func (s *Service) Search(ctx context.Context, req *memory.SearchRequest) (*memory.SearchResponse, error)
```

Method names differ between ADK versions (`AddSession`/`Search` vs. `AddSessionToMemory`/`SearchMemory`). The compile-time check catches mismatches.

---

## Conventions

- **Test tools and memory as plain Go.** Only e2e tests call a real LLM.
- **Pin the ADK version.** The API is still changing; upgrade deliberately.
- **Surprises are findings.** Write them down in the milestone's ADR.
- **Prompts live in `prompt.go`**, next to the agent, never inline in `main.go`.

---

## Getting started

```bash
cp .env.example .env         # add your GOOGLE_API_KEY
docker compose up -d db      # needed from M6
make run-runbook             # M1
```

## ADR log

| # | Title | Milestone | Status |
|---|---|---|---|
| 000 | Template | — | — |
| 001 | Identity propagation over A2A | M5 | todo |
