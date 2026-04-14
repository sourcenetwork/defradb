# Roadmap: DefraDB Debug Skill

## Overview

This skill goes from nothing to a fully autonomous debugging agent in three phases: first, the mechanical foundation of starting/stopping DefraDB and sending queries; second, the correctness engine that makes those queries meaningful by validating responses and producing actionable bug reports; third, the codebase intelligence layer that targets queries at real edge cases instead of random probing. Each phase delivers a coherent capability that builds on the previous.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [ ] **Phase 1: Foundation** - Instance lifecycle, basic query execution, skill entry point, and fixture loading
- [x] **Phase 2: Correctness and Reporting** - First-principles validation, error classification, structured bug reports, and user interrupt discipline (completed 2026-04-14)
- [ ] **Phase 3: Codebase Intelligence** - Sub-agent architecture, codebase-aware query generation, schema introspection, and edge-case targeting

## Phase Details

### Phase 1: Foundation
**Goal**: The skill can be invoked, build DefraDB if needed, start a fresh instance, execute GraphQL queries against it, load user fixtures, and shut down cleanly
**Depends on**: Nothing (first phase)
**Requirements**: LIFE-01, LIFE-02, LIFE-03, LIFE-04, LIFE-05, QEXE-01, QEXE-03, INVK-01
**Success Criteria** (what must be TRUE):
  1. Running `/defradb:debug "test basic CRUD"` starts a fresh DefraDB instance (or connects to remote via `--remote`), executes at least one GraphQL query, and shuts down without orphaning processes
  2. If DefraDB source has changed since last build, the skill rebuilds before starting the instance
  3. The skill waits for DefraDB health check to pass before sending any queries
  4. User-provided JSON fixtures via `--fixtures` are loaded into the running instance before testing begins
  5. The skill file lives at `.claude/skills/` in the defradb repo and is invocable as a Claude Code skill
**Plans:** 2 plans

Plans:
- [x] 01-01-PLAN.md — Create complete SKILL.md with lifecycle, query execution, fixtures, and remote support
- [ ] 01-02-PLAN.md — Integration test lifecycle commands against real DefraDB instance and human verify

### Phase 2: Correctness and Reporting
**Goal**: The skill validates query responses against first-principles database semantics, classifies anomalies, reproduces them before reporting, and produces structured bug reports with reproduction steps
**Depends on**: Phase 1
**Requirements**: CORR-01, CORR-02, REPT-01, REPT-02, REPT-03, REPT-04, INVK-02
**Success Criteria** (what must be TRUE):
  1. After executing queries, the skill reasons about expected results using database fundamentals (CRUD consistency, filter semantics, referential integrity) rather than trusting code behavior
  2. When the skill detects a discrepancy, it classifies it as parse error, runtime error, or data correctness issue
  3. The skill only interrupts the user when an anomaly is reproducible (re-run 2-3x) and the skill can articulate WHY the behavior is wrong
  4. Each debug session produces a DEBUG_PROGRESS_<DATE>.md folder containing a chronological execution log and a structured summary report with minimal reproduction steps for each bug found
**Plans**: 2 plans

Plans:
- [x] 02-01-PLAN.md — Add --verbose flag and rewrite Section 5 with correctness engine, error classification, anomaly reproduction, and structured reporting
- [x] 02-02-PLAN.md — Integration test correctness engine against live DefraDB instance and human verify

### Phase 3: Codebase Intelligence
**Goal**: The skill uses sub-agents to analyze DefraDB source code, introspects the GraphQL schema, and generates targeted edge-case queries informed by planner/request internals while maintaining the dual-track reasoning separation
**Depends on**: Phase 2
**Requirements**: CODE-01, CODE-02, CODE-03, CODE-04, QEXE-02
**Success Criteria** (what must be TRUE):
  1. Before generating queries, the skill introspects the running instance's GraphQL schema to discover available types, fields, and operations
  2. Sub-agents read relevant DefraDB source code (planner, request, db packages) and return analysis summaries to the orchestrator without bloating its context
  3. The skill generates edge-case queries targeting specific planner paths and boundary conditions identified from codebase analysis
  4. Codebase analysis informs WHERE to probe (query targets) while first-principles reasoning from Phase 2 determines WHAT SHOULD HAPPEN (expected results) -- the two tracks remain strictly separated
**Plans**: TBD

Plans:
- [x] 03-01-PLAN.md -- Schema introspection, query generation sub-agent, codebase analysis sub-agent, dual-track integration
- [ ] 03-02-PLAN.md -- Integration test sub-agent architecture against live DefraDB instance and human verify

## Progress

**Execution Order:**
Phases execute in numeric order: 1 -> 2 -> 3

| Phase | Plans Complete | Status | Completed |
|-------|---------------|--------|-----------|
| 1. Foundation | 0/2 | Planning complete | - |
| 2. Correctness and Reporting | 2/2 | Complete   | 2026-04-14 |
| 3. Codebase Intelligence | 1/2 | In progress | - |
