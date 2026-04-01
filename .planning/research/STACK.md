# Technology Stack

**Project:** DefraDB Debug Skill
**Researched:** 2026-03-31

## Recommended Stack

This skill is a Claude Code skill (markdown-driven, shell-executed). It does not introduce new compiled languages or frameworks. The "stack" is the set of shell tools, Claude Code primitives, and patterns used to orchestrate debugging sessions against a running DefraDB instance.

### Core: Claude Code Skill System

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| Claude Code Skills | Current | Skill definition, invocation, sub-agent orchestration | Native platform -- skills are markdown files with YAML frontmatter that Claude loads and executes. No alternative exists. |
| Claude Code Sub-agents | Current | Parallel codebase analysis and query execution | Built-in isolation mechanism -- each sub-agent gets its own context window, tool restrictions, and model selection. Prevents context exhaustion in the main skill. |
| Claude Code Hooks | Current | Lifecycle management (start/stop DefraDB) | `SubagentStart`/`SubagentStop` hooks or skill-scoped `Stop` hooks enable automated cleanup of background processes. |

**Confidence: HIGH** -- sourced from official Claude Code documentation at [code.claude.com/docs/en/skills](https://code.claude.com/docs/en/skills) and [code.claude.com/docs/en/sub-agents](https://code.claude.com/docs/en/sub-agents).

### Skill Directory Structure

```
.claude/skills/defradb-debug/
  SKILL.md                    # Main skill entry point (<500 lines)
  references/
    query-patterns.md         # GraphQL query templates and patterns
    correctness-oracle.md     # First-principles correctness reasoning guide
    area-catalog.md           # Catalog of testable areas (planner, CRUD, joins, etc.)
  scripts/
    start-defradb.sh          # Start instance, return PID, wait for health
    stop-defradb.sh           # Stop instance by PID file
    exec-graphql.sh           # Execute GraphQL query, return parsed JSON
    check-health.sh           # Health check polling
```

**Rationale:** Progressive disclosure -- SKILL.md stays focused on orchestration logic, references are loaded on-demand when Claude needs domain knowledge, scripts handle deterministic operations.

**Confidence: HIGH** -- structure directly follows official skill documentation patterns.

### HTTP/GraphQL Interaction

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| `curl` | System | HTTP requests to DefraDB API | Universally available, zero-install, Claude Code's Bash tool can execute it directly. No dependency management needed. |
| `jq` | System (1.6+) | JSON response parsing and extraction | Standard CLI JSON processor. Enables precise field extraction, error checking, and data transformation in shell pipelines. |

**Do NOT use:** httpie (extra install), graphqurl (Node dependency), Python scripts (unnecessary complexity). curl + jq is the minimal, universally-available combination that works in any Claude Code environment.

**Confidence: HIGH** -- curl+jq is the standard CLI approach for GraphQL, verified across multiple sources.

#### GraphQL Query Pattern

```bash
# POST query to DefraDB GraphQL endpoint
curl -s -X POST "http://127.0.0.1:9181/api/v0/graphql" \
  -H "Content-Type: application/json" \
  -d "$(jq -c -n --arg query "$GRAPHQL_QUERY" '{"query":$query}')" \
  | jq '.'
```

**Key details verified from DefraDB source code:**
- Endpoint: `POST /api/v0/graphql` (confirmed in `http/handler_store.go:873-874`)
- Default address: `127.0.0.1:9181` (confirmed in `cli/config/config.go:82`)
- Request body: `{"query": "...", "operationName": "...", "variables": {...}}` (confirmed in `http/handler_store.go:338-342`)
- Response: JSON with `data` and optionally `errors` fields (standard GraphQL over HTTP)
- Health check: `GET /health-check` (confirmed in `cli/wizard/callbacks.go:385`)

**Confidence: HIGH** -- verified against DefraDB source code.

### Background Process Management

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| Bash process control | System | Start/stop DefraDB instances | `$!` captures background PID, `kill -0` checks liveness, `kill` stops process. Standard POSIX, no dependencies. |
| PID file pattern | N/A | Track running instance across script invocations | Write `$!` to a temp file after backgrounding DefraDB. Scripts read this file to stop/check the instance. Survives across separate Bash tool invocations in Claude Code. |
| `trap` + cleanup | N/A | Ensure instance stops on skill exit | Skill-scoped `Stop` hook in SKILL.md frontmatter triggers `stop-defradb.sh`. Prevents orphaned processes. |

#### Instance Lifecycle Scripts

**start-defradb.sh:**
```bash
#!/bin/bash
set -euo pipefail
PIDFILE="${1:-/tmp/defradb-debug.pid}"
STORE="${2:-memory}"
PORT="${3:-9181}"

# Kill existing instance if running
if [ -f "$PIDFILE" ] && kill -0 "$(cat "$PIDFILE")" 2>/dev/null; then
  kill "$(cat "$PIDFILE")" 2>/dev/null || true
  sleep 1
fi

# Start DefraDB in background
defradb start --store "$STORE" --url "127.0.0.1:$PORT" &>/tmp/defradb-debug.log &
echo $! > "$PIDFILE"

# Wait for health check (up to 30s)
for i in $(seq 1 30); do
  if curl -sf "http://127.0.0.1:$PORT/health-check" &>/dev/null; then
    echo "DefraDB ready on port $PORT (PID: $(cat "$PIDFILE"))"
    exit 0
  fi
  sleep 1
done

echo "ERROR: DefraDB failed to start within 30s" >&2
cat /tmp/defradb-debug.log >&2
exit 1
```

**stop-defradb.sh:**
```bash
#!/bin/bash
PIDFILE="${1:-/tmp/defradb-debug.pid}"
if [ -f "$PIDFILE" ]; then
  PID=$(cat "$PIDFILE")
  if kill -0 "$PID" 2>/dev/null; then
    kill "$PID"
    echo "Stopped DefraDB (PID: $PID)"
  fi
  rm -f "$PIDFILE"
fi
```

**Confidence: HIGH** -- standard POSIX process management patterns.

### Build Staleness Detection

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| Git commit comparison | N/A | Detect if binary is stale vs source | Compare `git rev-parse HEAD` against a marker file written after each build. Simple, reliable, no external deps. |
| `make build` | N/A | Build DefraDB binary | Existing project Makefile handles CGO_ENABLED=1 and all build flags. |

```bash
# Check if rebuild needed
MARKER="/tmp/defradb-debug-build-marker"
CURRENT_HEAD=$(git rev-parse HEAD)
if [ -f "$MARKER" ] && [ "$(cat "$MARKER")" = "$CURRENT_HEAD" ]; then
  echo "Binary up to date"
else
  make build && echo "$CURRENT_HEAD" > "$MARKER"
fi
```

**Confidence: HIGH** -- trivial git-based pattern.

### Sub-Agent Architecture

| Agent | Model | Tools | Purpose |
|-------|-------|-------|---------|
| `codebase-analyzer` | haiku | Read, Grep, Glob | Read planner/parser/db internals to understand how a feature works. Returns analysis that guides query generation. |
| `query-executor` | haiku | Bash(curl *), Bash(jq *) | Execute GraphQL queries against running instance, parse responses, detect anomalies. Keeps verbose HTTP output out of main context. |
| Main skill (orchestrator) | inherit (opus) | Agent, Read, Bash, Write | Orchestrates the session: decides what to test, dispatches sub-agents, reasons about correctness, writes reports. |

**Sub-agent configuration approach:**

Define sub-agents in `.claude/agents/` as markdown files:

```yaml
# .claude/agents/defradb-codebase-analyzer.md
---
name: defradb-codebase-analyzer
description: Analyzes DefraDB codebase to understand how features work internally. Use when the debug skill needs to understand planner behavior, query parsing, or storage patterns.
tools: Read, Grep, Glob
model: haiku
---

You are a DefraDB codebase analyst. Given a feature area or component name,
find and analyze the relevant source code to explain how it works internally.

Focus on: query planner nodes, GraphQL schema generation, document lifecycle,
CRDT merge logic, and filter/join behavior.

Return structured analysis: what the code does, edge cases, and potential
areas where bugs could hide.
```

```yaml
# .claude/agents/defradb-query-executor.md
---
name: defradb-query-executor
description: Executes GraphQL queries against a running DefraDB instance and analyzes results. Use when the debug skill needs to run queries and check responses.
tools: Bash
permissionMode: acceptEdits
---

You are a DefraDB query execution agent. Given a GraphQL query and expected
behavior description, execute the query against the running instance and
analyze the response.

DefraDB API: POST http://127.0.0.1:9181/api/v0/graphql
Request format: {"query": "..."}

Use curl + jq for all HTTP interactions.
Report: actual response, whether it matches expected behavior, and any
anomalies detected.
```

**Why sub-agents over inline execution:**
1. Context isolation -- codebase analysis produces large outputs that would exhaust the main skill's context window
2. Parallelism -- codebase analysis and query execution can run simultaneously
3. Cost control -- haiku for mechanical tasks, opus for reasoning
4. Tool restriction -- analyzer is read-only, executor only needs Bash

**Confidence: MEDIUM** -- sub-agent architecture follows documented patterns, but the specific agent decomposition for this use case is an architectural decision that needs validation during implementation.

### Report Output

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| Markdown files | N/A | Session reports in `DEBUG_PROGRESS_<DATE>.md` | Human-readable, Claude can write natively, reviewable in any editor or GitHub. |
| Write tool | Built-in | Create/update report files | Claude Code's native file writing -- no shell needed for report generation. |

**Confidence: HIGH** -- straightforward file output.

## Alternatives Considered

| Category | Recommended | Alternative | Why Not |
|----------|-------------|-------------|---------|
| HTTP client | curl | httpie, wget, Python requests | Extra install dependency. curl is universally available and sufficient for simple POST requests. |
| JSON parsing | jq | Python json, Node, grep | jq is purpose-built for CLI JSON processing. grep/awk are fragile for JSON. Python/Node add unnecessary runtime deps. |
| Process management | PID files + kill | Docker containers, systemd | Massive overkill for a single background process. PID file pattern is simple, portable, and sufficient. |
| Sub-agent model | Custom agents (haiku + opus) | All-in-one skill, MCP servers | All-in-one exhausts context window. MCP servers require setup infrastructure. Sub-agents are native to Claude Code. |
| Build detection | Git HEAD comparison | File mtime, checksums | Git HEAD is semantically correct (detects source changes), simple, and DefraDB is a git repo. |
| GraphQL client | curl + jq | graphqurl, Postman CLI, gql-cli | Zero-dependency approach. Claude constructs queries as strings; no need for a client library's type system. |
| Skill type | `disable-model-invocation: true` | Auto-invocation | Debug sessions have side effects (start processes, write reports). User should explicitly trigger. |

## Installation

No installation beyond what DefraDB already requires. The skill uses only system tools:

```bash
# Verify prerequisites (all should be pre-installed)
which curl   # HTTP client
which jq     # JSON parser
which git    # Build staleness detection
which make   # Build system
which defradb || make build  # DefraDB binary
```

If `jq` is missing (unlikely on modern systems):
```bash
# Debian/Ubuntu
sudo apt-get install jq
# macOS
brew install jq
```

## Key Technical Constraints

1. **Claude Code Bash tool resets working directory between calls.** All scripts must use absolute paths or receive paths as arguments. PID files go in `/tmp/` for accessibility.

2. **Sub-agents cannot spawn other sub-agents.** The orchestration is strictly two-level: main skill -> sub-agents. No sub-agent can delegate further.

3. **Background sub-agents auto-deny unpermitted operations.** Pre-approve all needed permissions before launching background query execution.

4. **Skill content loads only when invoked.** Keep SKILL.md under 500 lines. Move reference material to `references/` directory.

5. **`$ARGUMENTS` substitution** provides the user's debug prompt to the skill. The user's area-of-concern description flows through `$ARGUMENTS`.

6. **DefraDB requires CGO_ENABLED=1** for building. The `make build` target already handles this.

## Sources

- [Claude Code Skills Documentation](https://code.claude.com/docs/en/skills) -- HIGH confidence, official docs
- [Claude Code Sub-agents Documentation](https://code.claude.com/docs/en/sub-agents) -- HIGH confidence, official docs
- [Claude Skills Deep Dive](https://leehanchung.github.io/blogs/2025/10/26/claude-skills-deep-dive/) -- MEDIUM confidence, third-party analysis
- [curl for GraphQL](https://til.simonwillison.net/graphql/graphql-with-curl) -- HIGH confidence, practical reference
- [GraphQL with curl](https://www.maxivanov.io/make-graphql-requests-with-curl/) -- HIGH confidence, practical reference
- DefraDB source code (`http/handler_store.go`, `cli/config/config.go`, `cli/start.go`) -- HIGH confidence, primary source
