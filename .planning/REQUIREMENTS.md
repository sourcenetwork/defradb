# Requirements: DefraDB Debug Skill

**Defined:** 2026-04-01
**Core Value:** Find real bugs in DefraDB that unit and integration tests miss, by autonomously generating and executing targeted end-to-end workloads guided by both codebase understanding and independent correctness reasoning.

## v1 Requirements

Requirements for initial release. Each maps to roadmap phases.

### Instance Lifecycle

- [x] **LIFE-01**: Skill detects build staleness by comparing git HEAD against last-built commit marker
- [x] **LIFE-02**: Skill starts a fresh defradb instance with configurable store backend (memory default)
- [x] **LIFE-03**: Skill polls health check endpoint until instance is ready before proceeding
- [x] **LIFE-04**: Skill cleanly shuts down defradb instance on session completion or error (PID tracking)
- [x] **LIFE-05**: Skill connects to an existing remote defradb instance via `--remote` flag

### Query Execution

- [x] **QEXE-01**: Skill executes GraphQL queries against defradb HTTP API via curl
- [ ] **QEXE-02**: Skill introspects GraphQL schema before generating queries
- [x] **QEXE-03**: Skill accepts user-provided JSON fixtures via `--fixtures` flag

### Correctness Engine

- [x] **CORR-01**: Skill reasons about expected behavior from first principles (CRUD consistency, referential integrity, filter semantics)
- [x] **CORR-02**: Skill classifies errors as parse error, runtime error, or data correctness issue

### Codebase Analysis

- [ ] **CODE-01**: Sub-agents read relevant source code to understand target area implementation
- [ ] **CODE-02**: Skill identifies the query surface area relevant to the user's prompt
- [ ] **CODE-03**: Skill generates edge-case queries from planner/request internals
- [ ] **CODE-04**: Skill uses dual-track reasoning — codebase for WHERE to probe, first-principles for WHAT should happen

### Reporting

- [x] **REPT-01**: Skill creates DEBUG_PROGRESS_<DATE>.md session folder for documentation
- [x] **REPT-02**: Skill writes a running chronological log during execution
- [x] **REPT-03**: Skill produces a structured summary report (bugs found, observations, recommendations)
- [x] **REPT-04**: Skill includes minimal reproduction steps for each bug found

### Skill Invocation

- [x] **INVK-01**: Skill is invoked via `/defradb:debug <prompt>` with optional flags
- [x] **INVK-02**: Skill only prompts user when anomaly is reproducible AND agent can articulate reasoning

## v2 Requirements

Deferred to future release. Tracked but not in current roadmap.

### Instance Lifecycle

- **LIFE-06**: Port conflict detection and auto-assignment
- **LIFE-07**: Orphan process reaper script

### Query Execution

- **QEXE-04**: Auto-generate schemas and seed test documents based on target area
- **QEXE-05**: Batch query execution with progressive complexity

### Correctness Engine

- **CORR-03**: Field-level verification against pre-computed expectations
- **CORR-04**: Integration test corpus mining as reference
- **CORR-05**: Cross-query consistency checking (insert -> query -> verify round-trip)

### Reporting

- **REPT-05**: Session resumption from progress file

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature | Reason |
|---------|--------|
| Bug fixing / code modification | Report-only — skill should surface bugs reliably first |
| Unit test generation | This is E2E black-box only |
| Integration test modification | Separate testing system |
| P2P sync testing | Complex multi-node orchestration, defer to v2 |
| ACP testing | Requires policy setup infrastructure |
| Performance benchmarking | Different tool, different goal |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| LIFE-01 | Phase 1 | Complete |
| LIFE-02 | Phase 1 | Complete |
| LIFE-03 | Phase 1 | Complete |
| LIFE-04 | Phase 1 | Complete |
| LIFE-05 | Phase 1 | Complete |
| QEXE-01 | Phase 1 | Complete |
| QEXE-02 | Phase 3 | Pending |
| QEXE-03 | Phase 1 | Complete |
| CORR-01 | Phase 2 | Complete |
| CORR-02 | Phase 2 | Complete |
| CODE-01 | Phase 3 | Pending |
| CODE-02 | Phase 3 | Pending |
| CODE-03 | Phase 3 | Pending |
| CODE-04 | Phase 3 | Pending |
| REPT-01 | Phase 2 | Complete |
| REPT-02 | Phase 2 | Complete |
| REPT-03 | Phase 2 | Complete |
| REPT-04 | Phase 2 | Complete |
| INVK-01 | Phase 1 | Complete |
| INVK-02 | Phase 2 | Complete |

**Coverage:**
- v1 requirements: 20 total
- Mapped to phases: 20
- Unmapped: 0

---
*Requirements defined: 2026-04-01*
*Last updated: 2026-04-01 after roadmap creation*
