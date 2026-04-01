# Project Research Summary

**Project:** DefraDB Debug Skill
**Domain:** Agentic database debugging (Claude Code skill)
**Researched:** 2026-03-31
**Confidence:** HIGH

## Executive Summary

This project builds a Claude Code skill that autonomously debugs DefraDB by starting a live instance, generating targeted GraphQL queries informed by codebase analysis, and validating responses against first-principles database semantics rather than code-derived expectations. The recommended approach is a two-level orchestration architecture: an Opus-powered orchestrator manages the session and correctness reasoning, while Haiku-powered sub-agents handle codebase analysis and query execution in isolation. The entire stack requires zero new dependencies -- just curl, jq, and standard POSIX process management against DefraDB's existing GraphQL API.

The critical insight from research is the dual-track reasoning model: codebase analysis tells the skill WHERE to probe (edge cases, boundary conditions, specific planner paths), while first-principles reasoning tells it WHAT SHOULD HAPPEN (relational semantics, CRUD consistency, GraphQL spec compliance). This separation is what prevents the skill from simply confirming buggy code behavior. Without it, the skill is worthless -- it will validate bugs as correct behavior.

The primary risks are context window exhaustion (mitigated by strict sub-agent delegation), orphaned DefraDB processes (mitigated by PID files and Stop hooks), and GraphQL string escaping in shell pipelines (mitigated by using jq --arg for all JSON construction). All three have concrete, validated prevention strategies.

## Key Findings

### Recommended Stack

No new languages, frameworks, or package dependencies. The skill is pure markdown + shell scripts operating within the Claude Code skill system. See [STACK.md](STACK.md) for full details.

**Core technologies:**
- **Claude Code Skills**: Markdown-driven skill definition with YAML frontmatter -- the native execution platform, no alternative exists
- **Claude Code Sub-agents**: Haiku for codebase analysis (read-only) and query execution (Bash-only); Opus for orchestration and correctness reasoning -- provides context isolation and cost control
- **curl + jq**: HTTP interaction with DefraDB's GraphQL API -- universally available, zero-install, handles all escaping via `jq --arg`
- **PID file + kill**: Process lifecycle management for DefraDB instances -- standard POSIX, no container or service manager overhead
- **Git HEAD comparison**: Build staleness detection -- semantically correct for detecting source changes

### Expected Features

See [FEATURES.md](FEATURES.md) for the full feature landscape and dependency graph.

**Must have (table stakes):**
- Build staleness detection and auto-rebuild
- Instance lifecycle management (start/stop with health check)
- Schema creation and document CRUD via GraphQL
- Response validation against first-principles expectations (the correctness oracle)
- Structured bug reports with repro steps

**Should have (differentiators):**
- Codebase-aware query generation via sub-agents (reads planner internals to target real edge cases)
- Independent correctness reasoning (finds bugs that code-aligned tests miss)
- Progressive complexity testing (simple-first, then filters, joins, pagination)
- Area catalog with exhaustive coverage per area

**Defer (v2+):**
- P2P/multi-node testing (complex orchestration)
- ACP/permission testing (requires identity infrastructure)
- Auto-fix capabilities (report-only first)
- Parallel query batch execution (optimize after single-threaded works)

### Architecture Approach

Two-level orchestration with strict component boundaries. The orchestrator (SKILL.md) owns session lifecycle, test strategy, and correctness reasoning. Sub-agents own verbose operations (codebase reading, HTTP execution) and return only summaries. Shell scripts own deterministic operations (start/stop DB, execute queries). Reference files provide domain knowledge loaded on-demand to keep SKILL.md under 500 lines. See [ARCHITECTURE.md](ARCHITECTURE.md) for the full pattern catalog.

**Major components:**
1. **SKILL.md orchestrator** -- session lifecycle, test strategy, correctness reasoning, report writing
2. **codebase-analyzer agent** (Haiku, read-only) -- reads DefraDB source, returns feature analysis and edge cases
3. **query-executor agent** (Haiku, Bash-only) -- executes curl commands, returns parsed JSON responses
4. **Shell scripts** (start/stop/exec/health) -- deterministic operations, independently testable
5. **Reference files** (query patterns, correctness oracle, area catalog) -- domain knowledge loaded on-demand

### Critical Pitfalls

See [PITFALLS.md](PITFALLS.md) for the complete pitfall catalog with phase-specific warnings.

1. **Context window exhaustion** -- Delegate ALL codebase reading and query execution to sub-agents. The orchestrator only sees summaries. Keep SKILL.md under 500 lines.
2. **Code-aligned correctness bias** -- Strictly separate codebase analysis (where to test) from first-principles reasoning (what should happen). Never derive expected values from how the code works.
3. **Orphaned DefraDB processes** -- Use PID files, Stop hooks in SKILL.md frontmatter, and kill-existing-on-start logic. Triple redundancy for cleanup.
4. **GraphQL string escaping** -- Always use `jq --arg` for JSON payload construction. Never manually embed GraphQL in JSON strings.
5. **Working directory reset** -- All scripts use absolute paths. PID files go to `/tmp/`. Scripts receive paths as arguments.

## Implications for Roadmap

Based on combined research, the feature dependency graph and architecture boundaries suggest four phases.

### Phase 1: Foundation -- Instance Lifecycle and Basic Execution
**Rationale:** Everything depends on being able to start DefraDB, send queries, and stop it cleanly. This is the dependency root.
**Delivers:** Working shell scripts (start, stop, health check, exec-graphql), PID file management, build staleness detection, basic SKILL.md skeleton with Stop hook.
**Addresses:** Instance lifecycle, build detection, schema creation, document CRUD (table stakes)
**Avoids:** Orphaned processes (pitfall 3), port conflicts, CGO build failures, working directory reset issues

### Phase 2: Correctness Engine -- Validation and Reporting
**Rationale:** With a running instance and query execution working, the next priority is the correctness oracle -- the feature that makes the skill actually useful rather than just a query runner.
**Delivers:** First-principles validation logic in the orchestrator, structured bug report generation, anomaly reproduction (re-run 2-3x before reporting), incremental progress logging.
**Addresses:** Response validation, structured bug reports, anomaly reproduction (table stakes + differentiators)
**Avoids:** Code-aligned correctness bias (pitfall 2) -- this phase explicitly establishes the dual-track reasoning pattern

### Phase 3: Intelligence -- Sub-Agent Architecture and Codebase-Aware Testing
**Rationale:** Sub-agent architecture and codebase analysis are the differentiators that make this skill better than manual testing. Requires Phase 1 (execution) and Phase 2 (validation) to be useful.
**Delivers:** codebase-analyzer agent definition, query-executor agent definition, sub-agent orchestration in SKILL.md, area catalog with progressive complexity, targeted edge-case query generation.
**Addresses:** Codebase-aware query generation, exhaustive area coverage, parallel sub-agent execution (differentiators)
**Avoids:** Context window exhaustion (pitfall 1), nested sub-agent delegation (pitfall 4), background permission issues (pitfall 5)

### Phase 4: Polish -- Reference Material and UX
**Rationale:** With the core loop working (start -> analyze -> query -> validate -> report), fill in the reference material that makes the skill's domain knowledge comprehensive.
**Delivers:** query-patterns.md reference, correctness-oracle.md guide, area-catalog.md with all testable areas, user prompt parsing improvements, progress log formatting, optional flags (--store, --remote, --fixtures).
**Addresses:** Auto-generated test schemas, running progress log (differentiators)
**Avoids:** Skill content bloat (keep references separate from SKILL.md)

### Phase Ordering Rationale

- **Dependency-driven:** Each phase builds on the previous. You cannot validate responses (Phase 2) without executing queries (Phase 1). You cannot do codebase-aware testing (Phase 3) without having validation logic (Phase 2).
- **Value-driven:** Phase 1 alone is useless (just starts/stops DB). Phase 2 makes it minimally useful (can find obvious bugs). Phase 3 makes it genuinely valuable (finds subtle bugs via targeted probing). Phase 4 is refinement.
- **Risk-driven:** The highest-risk architectural decision (sub-agent decomposition) is deferred to Phase 3, after the simpler single-context flow is proven in Phases 1-2. If sub-agents don't work well, Phases 1-2 still deliver a functional (if less powerful) skill.

### Research Flags

Phases likely needing deeper research during planning:
- **Phase 2:** The correctness oracle is the hardest design problem. How to encode first-principles database semantics in a way that Claude can reliably apply needs iterative refinement. May need research into DefraDB-specific GraphQL mutation syntax for schema creation and document operations.
- **Phase 3:** Sub-agent orchestration patterns are documented but the specific decomposition (analyzer + executor) needs validation during implementation. Exact agent markdown format and permission model should be verified against latest Claude Code docs.

Phases with standard patterns (skip research-phase):
- **Phase 1:** Shell scripts, PID files, curl+jq, health checks -- all well-documented, standard patterns. No research needed.
- **Phase 4:** Reference file authoring is straightforward content work.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | All tools verified against official docs and DefraDB source code. Zero new dependencies. |
| Features | HIGH | Clear dependency graph. Table stakes vs differentiators well-separated. Anti-features explicitly scoped. |
| Architecture | HIGH (structure), MEDIUM (sub-agents) | Two-level orchestration is sound. Specific sub-agent decomposition needs implementation validation. |
| Pitfalls | HIGH | All critical pitfalls have concrete, tested prevention strategies. Phase-specific warnings are actionable. |

**Overall confidence:** HIGH

### Gaps to Address

- **Sub-agent permission model:** The exact `permissionMode` and tool pre-approval needed for background sub-agents should be verified against current Claude Code behavior during Phase 3 implementation.
- **Correctness oracle depth:** How deep can first-principles reasoning go for complex features like CRDT merges or DAG sync? Phase 2 will need iterative refinement of the oracle's reasoning templates.
- **DefraDB GraphQL mutation syntax:** Need to verify exact syntax for `addSchema`, `createDocument`, `updateDocument`, `deleteDocument` mutations. Existing integration tests are the best reference.
- **Area catalog content:** What specific areas of DefraDB should the skill know how to test? Needs phase-specific research reading the planner, parser, and db packages.
- **Memory store edge cases:** Some DefraDB behaviors may differ between memory and persistent stores. Document any discrepancies found; consider adding badger backend option in v2.
- **Session length limits:** No research on how long a Claude Code skill session can run before hitting platform limits. May need to scope per-invocation to a single test area.

## Sources

### Primary (HIGH confidence)
- [Claude Code Skills Documentation](https://code.claude.com/docs/en/skills) -- skill structure, frontmatter, hooks, lifecycle
- [Claude Code Sub-agents Documentation](https://code.claude.com/docs/en/sub-agents) -- agent definitions, permission modes, background execution
- DefraDB source code (`http/handler_store.go`, `cli/config/config.go`, `cli/start.go`) -- API endpoints, default config, build system

### Secondary (MEDIUM confidence)
- [Claude Skills Deep Dive](https://leehanchung.github.io/blogs/2025/10/26/claude-skills-deep-dive/) -- third-party analysis of skill patterns
- [curl for GraphQL](https://til.simonwillison.net/graphql/graphql-with-curl) -- practical reference
- [GraphQL with curl](https://www.maxivanov.io/make-graphql-requests-with-curl/) -- practical reference

---
*Research completed: 2026-03-31*
*Ready for roadmap: yes*
