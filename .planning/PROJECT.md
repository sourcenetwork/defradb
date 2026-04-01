# DefraDB Debug Skill

## What This Is

A Claude Code skill (`/defradb:debug`) that agentically tests and debugs DefraDB through end-to-end black-box testing. Given a prompt describing an area of concern, the skill builds DefraDB (if stale), launches a fresh instance, then iteratively generates and executes GraphQL queries against the HTTP API to find and document bugs. It uses deep codebase understanding to guide query generation while reasoning from first principles about expected behavior.

## Core Value

Find real bugs in DefraDB that unit and integration tests miss, by autonomously generating and executing targeted end-to-end workloads guided by both codebase understanding and independent correctness reasoning.

## Requirements

### Validated

<!-- Existing capabilities the skill builds on -->

- ✓ DefraDB builds via `make build` — existing
- ✓ DefraDB serves GraphQL over HTTP API — existing
- ✓ CLI can start/stop instances (`defradb start`) — existing
- ✓ Multiple store backends supported (memory, badger, leveldb) — existing
- ✓ Integration test corpus exists (`tests/integration/`) — existing
- ✓ GraphQL query language defined in `internal/request/graphql/` — existing
- ✓ Query planner in `internal/planner/` — existing
- ✓ HTTP client interface in `http/` — existing

### Active

- [ ] Skill invocation via `/defradb:debug <prompt>` with optional flags
- [ ] Git-based build staleness detection (HEAD vs last-built commit marker)
- [ ] Automated defradb instance lifecycle (start/stop per session)
- [ ] Memory store as default backend, configurable via `--store` flag
- [ ] Connect to existing remote instance via `--remote` flag
- [ ] User-provided JSON fixtures via `--fixtures` flag
- [ ] Auto-generated schemas and test documents based on target area
- [ ] Codebase-aware query generation (reads planner, request, db internals)
- [ ] First-principles correctness reasoning (independent of code behavior)
- [ ] Integration test corpus as reference (not gospel) for expected behavior
- [ ] Semantic analysis of query results against inserted data
- [ ] Exhaustive testing per area before moving on
- [ ] Hybrid parallel execution (pipeline overall, parallel within stages)
- [ ] Background sub-agents for codebase analysis and query execution
- [ ] Running log during execution + structured summary report at end
- [ ] DEBUG_PROGRESS_<DATE>.md folder for session documentation
- [ ] User interrupts only when anomaly is reproducible AND reasoned
- [ ] Skill definition lives in defradb repo (`.claude/skills/`)

### Out of Scope

- P2P sync testing — complex multi-node orchestration, defer to v2
- ACP (Access Control Policy) testing — requires policy setup infrastructure
- Bug fixing — report-only; no code modifications
- Unit test generation — this is end-to-end black-box only
- Integration test modifications — separate testing system
- Time-bounded sessions — runs until satisfied

## Context

**Codebase:** DefraDB is a decentralized peer-to-peer document database in Go. It uses CRDTs for conflict resolution, IPLD/CID for content-addressed storage, and GraphQL as the primary query interface. The query pipeline flows: GraphQL parse → request AST → planner → plan tree → execution → response.

**Key internal packages for the skill to understand:**
- `internal/request/graphql/` — GraphQL schema definition, parsing, type generation
- `internal/planner/` — Query plan construction (selectNode, updateNode, joins, aggregates)
- `internal/db/` — Core database operations, document CRUD, merge logic
- `internal/db/fetcher/` — Low-level document fetching from datastore
- `tests/integration/` — Existing test corpus with queries and expected outputs

**Critical design insight:** The skill uses codebase knowledge to understand *how* DefraDB works and *where* to probe, but must NOT trust the code as the source of truth for correctness. Code-aligned query generation risks producing queries that unknowingly conform to buggy behavior. The skill must reason independently about what *should* happen using database first principles (CRUD consistency, referential integrity, filter semantics, GraphQL spec).

**Correctness oracle hierarchy:**
1. First principles — database fundamentals and GraphQL semantics
2. User prompt context — what the user says should happen
3. Integration test expectations — reference corpus, not gospel
4. Codebase understanding — for targeting, NOT for defining correctness

## Constraints

- **Interface**: HTTP GraphQL API only (no direct Go API calls) — true black-box testing
- **Instance**: Fresh instance per session, clean state — reproducibility
- **Store**: Memory by default — speed; other backends via flag
- **Build**: Go with CGO_ENABLED=1 required
- **Interrupts**: Only prompt user when anomaly is reproducible AND agent can articulate reasoning — minimize noise
- **Scope**: Single-node only for v1 — no P2P, no ACP

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| HTTP API over CLI client for queries | More programmatic, easier to parse responses, closer to real usage | — Pending |
| Memory store as default | Fast iteration, disposable state, no disk cleanup needed | — Pending |
| Report-only, no auto-fix | Skill should surface bugs reliably before attempting repairs | — Pending |
| First-principles correctness over code-derived expectations | Code may contain the bug; trusting it creates blind spots | — Pending |
| Exhaustive per-area before moving on | Thorough coverage more valuable than broad shallow passes | — Pending |
| Git-based build staleness | Simple, reliable — compare HEAD vs marker file | — Pending |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd:transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd:complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-03-31 after initialization*
