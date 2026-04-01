# Feature Landscape

**Domain:** Agentic database debugging skill (Claude Code skill for DefraDB)
**Researched:** 2026-03-31

## Table Stakes

Features the skill must have to be useful at all. Without these, it cannot find bugs.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Build staleness detection + auto-rebuild | Skill must test current code, not stale binary | Low | Git HEAD vs marker file comparison |
| Instance lifecycle (start/stop) | Cannot test without a running database | Medium | PID file pattern, health check polling, cleanup on exit |
| Schema creation via GraphQL | Must define collections before testing them | Low | Standard `mutation { addSchema(...) }` calls |
| Document CRUD via GraphQL | Core testing capability -- insert, read, update, delete | Low | Well-documented GraphQL mutations |
| Response validation against first-principles | The correctness oracle -- determines if behavior is a bug | High | Must reason about what *should* happen, not what code does |
| Structured bug reports | Must output actionable findings | Medium | Markdown files with repro steps, expected vs actual, reasoning |
| User prompt interpretation | Must understand what area to test | Medium | Parse `$ARGUMENTS` into testable areas |

## Differentiators

Features that make this skill more valuable than manual testing or existing test suites.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Codebase-aware query generation | Reads planner/parser internals to target edge cases that generic testing would miss | High | Sub-agent reads source, identifies boundary conditions, generates targeted queries |
| Exhaustive area coverage | Tests an area thoroughly before moving on, not broad shallow passes | Medium | Iterative deepening: start with basic cases, progressively add complexity |
| Independent correctness reasoning | Finds bugs that code-aligned tests miss because it reasons from database fundamentals | High | Dual-track: codebase knowledge for *where* to probe, first principles for *what should happen* |
| Parallel sub-agent execution | Codebase analysis runs concurrently with query execution | Medium | Sub-agents in background, main skill orchestrates |
| Auto-generated test schemas | Creates schemas tailored to the area being tested, not just hardcoded fixtures | Medium | Dynamic schema generation based on target area (relations, indexes, nested types) |
| Running progress log | User sees what the skill is doing in real-time, not just final report | Low | Write to progress file incrementally |
| Anomaly reproduction | Confirms bugs are reproducible before reporting | Medium | Re-run failing query multiple times, check consistency |

## Anti-Features

Features to explicitly NOT build.

| Anti-Feature | Why Avoid | What to Do Instead |
|--------------|-----------|-------------------|
| Auto-fix bugs | Skill should be a reliable reporter first; auto-fix risks masking issues or introducing regressions | Report with precise repro steps and reasoning |
| Unit test generation | Different testing paradigm; skill does black-box E2E testing | Leave unit test creation to developers |
| P2P/multi-node testing | Requires complex multi-instance orchestration, out of scope for v1 | Single-node only; defer P2P to v2 |
| ACP/permission testing | Requires identity/policy setup infrastructure | Defer to future version |
| Integration test modification | Separate testing system with its own patterns | Report findings; let humans update integration tests |
| Custom GraphQL client library | Over-engineering; curl+jq is sufficient | Use curl+jq for all HTTP interactions |
| Persistent database state across sessions | Defeats reproducibility | Fresh instance per session, memory store by default |
| Time-bounded sessions | Arbitrary time limits produce incomplete results | Run until area is exhaustively covered |

## Feature Dependencies

```
Build detection -> Instance lifecycle (must have binary before starting)
Instance lifecycle -> Schema creation (must have running instance)
Schema creation -> Document CRUD (must have collections)
Document CRUD -> Response validation (must have responses to validate)
Response validation -> Bug reports (must have validated anomalies)

Codebase analysis (parallel, independent) -> Query generation strategy
Query generation strategy + Document CRUD -> Targeted testing
```

## MVP Recommendation

Phase 1 -- minimum viable debugging skill:
1. Instance lifecycle (start/stop with memory store)
2. Build staleness detection
3. Schema creation + document CRUD via GraphQL
4. Basic response validation (inserted data matches queried data)
5. Structured bug report output

Phase 2 -- intelligent testing:
1. Codebase-aware query generation via sub-agents
2. First-principles correctness reasoning
3. Area catalog with progressive complexity
4. Exhaustive coverage per area

Defer: Parallel sub-agent execution (optimize after basic flow works), custom fixtures via `--fixtures` flag (nice-to-have), `--remote` flag (edge case).

## Sources

- PROJECT.md requirements and constraints
- DefraDB HTTP API verified from source code
- Claude Code skills/sub-agents documentation
