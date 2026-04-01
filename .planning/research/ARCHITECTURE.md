# Architecture Patterns

**Domain:** Agentic database debugging skill (Claude Code skill for DefraDB)
**Researched:** 2026-03-31

## Recommended Architecture

Two-level orchestration: a main skill (orchestrator) delegates to specialized sub-agents for codebase analysis and query execution, while managing the DefraDB instance lifecycle and correctness reasoning itself.

```
User: /defradb:debug "test join query behavior"
  |
  v
[SKILL.md orchestrator] (opus, full tools)
  |
  |-- 1. Build check + instance start (scripts/)
  |
  |-- 2. Parse user prompt -> identify test area
  |
  |-- 3. PARALLEL:
  |     |-- [codebase-analyzer agent] (haiku, read-only)
  |     |     -> Reads planner/parser source for target area
  |     |     -> Returns: how feature works, edge cases, boundary conditions
  |     |
  |     |-- [query-executor agent] (haiku, Bash only)
  |           -> Executes setup queries (schema, initial docs)
  |           -> Returns: confirmation of setup success
  |
  |-- 4. LOOP (until area exhausted):
  |     |-- Orchestrator generates query based on:
  |     |     - Codebase analysis (from step 3)
  |     |     - First-principles reasoning (orchestrator's own)
  |     |     - Previous results (progressive complexity)
  |     |
  |     |-- [query-executor agent] runs query
  |     |     -> Returns: response JSON, timing, errors
  |     |
  |     |-- Orchestrator validates response:
  |     |     - Does result match first-principles expectation?
  |     |     - Does result match inserted data?
  |     |     - Are error messages appropriate?
  |     |
  |     |-- If anomaly: reproduce, then log to report
  |     |-- If clean: increase complexity, continue
  |
  |-- 5. Write final report + stop instance
```

### Component Boundaries

| Component | Responsibility | Communicates With |
|-----------|---------------|-------------------|
| SKILL.md (orchestrator) | Session lifecycle, test strategy, correctness reasoning, report writing | Sub-agents (via Agent tool), scripts (via Bash), user (via output) |
| `codebase-analyzer` agent | Read DefraDB source code, explain feature internals | Orchestrator (returns structured analysis) |
| `query-executor` agent | Execute curl commands, parse JSON responses | Orchestrator (returns query results) |
| `start-defradb.sh` | Start DefraDB instance, wait for health | Called by orchestrator via Bash |
| `stop-defradb.sh` | Stop DefraDB instance, clean up PID file | Called by orchestrator via Bash or Stop hook |
| `exec-graphql.sh` | Execute single GraphQL query, return JSON | Called by query-executor agent |
| `references/*.md` | Domain knowledge (query patterns, correctness rules, area catalog) | Read by orchestrator on-demand |

### Data Flow

**Inbound:** User prompt (`$ARGUMENTS`) -> orchestrator parses into test area and strategy.

**Internal:** Orchestrator dispatches sub-agents with specific tasks. Sub-agent results flow back as summarized text (not raw tool output). Orchestrator maintains state: what has been tested, what anomalies found, what complexity level reached.

**Outbound:** Progress log (incremental writes during session) + final structured report (markdown file with all findings).

## Patterns to Follow

### Pattern 1: Progressive Disclosure in Skill Content

**What:** SKILL.md contains orchestration logic only. Domain knowledge lives in `references/` files loaded on-demand.

**When:** Always. SKILL.md must stay under 500 lines.

**Example:**
```markdown
# In SKILL.md:
## Query Pattern Reference
For GraphQL query templates and common patterns, read
[references/query-patterns.md](references/query-patterns.md)
when you need to construct queries for a specific area.
```

### Pattern 2: Script-Based Deterministic Operations

**What:** Operations with deterministic behavior (start DB, stop DB, execute HTTP request) live in shell scripts, not inline Bash commands in the skill prompt.

**When:** Any operation that should work identically every time.

**Why:** Scripts are testable independently, don't consume skill prompt space, and can be versioned/debugged outside Claude Code.

**Example:**
```bash
# scripts/exec-graphql.sh
#!/bin/bash
set -euo pipefail
PORT="${1:-9181}"
QUERY="$2"
curl -sf -X POST "http://127.0.0.1:$PORT/api/v0/graphql" \
  -H "Content-Type: application/json" \
  -d "$(jq -c -n --arg query "$QUERY" '{"query":$query}')"
```

### Pattern 3: Dual-Track Reasoning

**What:** Codebase analysis informs WHERE to test. First-principles reasoning determines WHAT SHOULD HAPPEN. These tracks are independent.

**When:** Every test case.

**Why:** If the code is buggy and tests are derived from the code, the tests will pass on buggy behavior. First-principles reasoning breaks this cycle.

**Example flow:**
```
Codebase track: "selectNode uses typeIndexJoin for related collections,
  which builds a scan on the secondary index. Edge case: what if the
  secondary collection has zero documents?"

First-principles track: "A left join with an empty right side should
  return all left-side documents with null for joined fields.
  This is standard relational semantics."

Combined: Test a join query where one side has documents and the other
  is empty. Expected: non-empty result with null joined fields.
```

### Pattern 4: Anomaly Reproduction Before Reporting

**What:** When a query produces unexpected results, re-run it 2-3 times on a fresh instance before reporting.

**When:** Every anomaly detection.

**Why:** Eliminates false positives from timing issues, race conditions in the test harness, or transient state.

### Pattern 5: Incremental Complexity

**What:** Start with the simplest possible query for an area, then progressively add complexity: filters, sorting, pagination, nested fields, multiple operations.

**When:** Testing any new area.

**Why:** Simple-first approach isolates failures to specific complexity layers. If a basic CRUD works but filtered queries fail, the bug is in filter handling.

## Anti-Patterns to Avoid

### Anti-Pattern 1: Inline HTTP in Skill Prompt

**What:** Writing curl commands directly in SKILL.md content.

**Why bad:** Bloats the skill prompt, makes commands hard to maintain, introduces escaping issues with GraphQL strings in markdown.

**Instead:** Use `scripts/exec-graphql.sh` that handles escaping, error checking, and response formatting.

### Anti-Pattern 2: Monolithic Skill (No Sub-agents)

**What:** Running all codebase analysis, query execution, and reasoning in the main skill context.

**Why bad:** Codebase analysis of DefraDB's planner package alone can produce thousands of lines of context. Combined with query responses, the main context exhausts quickly.

**Instead:** Delegate verbose operations to sub-agents. Only summaries return to the orchestrator.

### Anti-Pattern 3: Code-Derived Expected Values

**What:** Reading the test suite or source code to determine what a query "should" return, then validating against that.

**Why bad:** If the code has a bug, the code-derived expectation will match the buggy behavior. The skill will report "no bugs found" when bugs exist.

**Instead:** Derive expected values from first principles: "I inserted document X, a query for all documents should return at least X."

### Anti-Pattern 4: Random Query Generation

**What:** Generating random GraphQL queries and checking if they crash.

**Why bad:** Low signal-to-noise ratio. Most random queries either fail to parse or test trivial paths. Misses the targeted edge cases that real bugs hide in.

**Instead:** Use codebase analysis to identify specific code paths and boundary conditions, then construct queries that exercise those paths.

### Anti-Pattern 5: Stateful Sessions Across Runs

**What:** Reusing a DefraDB instance or its data across separate skill invocations.

**Why bad:** State from previous sessions contaminates results. Bugs may not reproduce because they depend on clean state.

**Instead:** Fresh instance per session. Memory store. Clean start every time.

## Scalability Considerations

| Concern | Current (v1) | Future (v2) |
|---------|-------------|-------------|
| Context window | Sub-agents isolate verbose output; orchestrator stays lean | Could add auto-compaction triggers for very long sessions |
| Test breadth | Single area per invocation, exhaustive | Multi-area with area prioritization |
| Instance management | Single local instance | `--remote` for pre-existing instances, potentially Docker |
| Parallelism | Sequential query execution, parallel codebase analysis | Parallel query batches via multiple executor sub-agents |
| Report size | Single markdown file | Structured directory with per-area reports |

## Sources

- [Claude Code Skills Documentation](https://code.claude.com/docs/en/skills)
- [Claude Code Sub-agents Documentation](https://code.claude.com/docs/en/sub-agents)
- PROJECT.md design decisions and constraints
- DefraDB ARCHITECTURE.md (codebase analysis)
