# Domain Pitfalls

**Domain:** Agentic database debugging skill (Claude Code skill for DefraDB)
**Researched:** 2026-03-31

## Critical Pitfalls

Mistakes that cause the skill to be fundamentally broken or produce worthless output.

### Pitfall 1: Context Window Exhaustion

**What goes wrong:** The main skill context fills up with verbose codebase analysis output and raw HTTP responses, causing Claude to lose track of the testing strategy and produce degraded results.

**Why it happens:** DefraDB's planner package has dozens of files. A single "analyze how joins work" request can produce thousands of lines. Combined with repeated curl responses, context fills fast.

**Consequences:** Skill stops mid-session, produces incomplete reports, or starts hallucinating about previous test results.

**Prevention:** Strict sub-agent delegation. ALL codebase reading goes through the `codebase-analyzer` agent. ALL query execution goes through the `query-executor` agent. The orchestrator only sees summaries. Keep SKILL.md under 500 lines with references in separate files.

**Detection:** If the skill starts repeating queries it already ran or forgets what schemas it created, context is exhausted.

### Pitfall 2: Code-Aligned Correctness (Confirmation Bias)

**What goes wrong:** The skill reads DefraDB source code to understand how a feature works, then generates tests that validate the code's actual behavior rather than correct behavior. Bugs pass because the "expected" output matches the buggy code.

**Why it happens:** Natural tendency to derive expected values from how the code works. If the code says "filter X returns Y," the test checks for Y -- even if Y is wrong.

**Consequences:** Skill reports "no bugs found" when bugs exist. Defeats the entire purpose.

**Prevention:** Strict separation: codebase analysis for WHERE to test (which code paths, edge cases, boundary conditions). First-principles reasoning for WHAT SHOULD HAPPEN (database semantics, GraphQL spec, CRUD consistency). These are independent tracks that only combine at query generation time.

**Detection:** Review generated test cases. If "expected behavior" references specific code variables or function names, the reasoning is code-aligned rather than principles-based.

### Pitfall 3: Orphaned DefraDB Processes

**What goes wrong:** The skill starts a DefraDB instance but fails to stop it -- due to a crash, context exhaustion, user abort, or error in the stop logic. Orphaned processes consume memory and block ports.

**Why it happens:** Claude Code skill execution can be interrupted at any point. If stop logic is only in the main flow (not in cleanup hooks), interruptions leave processes running.

**Consequences:** Port 9181 blocked for next run. Memory leak. User must manually find and kill processes.

**Prevention:**
1. Use skill-scoped `Stop` hook in SKILL.md frontmatter to call `stop-defradb.sh`
2. PID file in `/tmp/` so stop script can find the process
3. Start script kills any existing process on the same PID file before starting new one
4. `kill -0` check before assuming process is running

**Detection:** `lsof -i :9181` or `ps aux | grep defradb` shows unexpected processes.

### Pitfall 4: GraphQL String Escaping in Shell

**What goes wrong:** GraphQL queries contain special characters (double quotes, newlines, backslashes) that break when embedded in shell commands or JSON payloads.

**Why it happens:** Multi-level escaping: GraphQL string -> JSON value -> shell argument -> curl data. Each level has its own escaping rules.

**Consequences:** Queries fail to parse. Subtle bugs where escaped characters change query semantics.

**Prevention:** Use `jq -c -n --arg query "$QUERY" '{"query":$query}'` to construct JSON payloads. The `--arg` flag handles all escaping automatically. Never manually construct JSON strings with embedded GraphQL.

**Detection:** HTTP 400 errors from DefraDB with parse error messages. Queries that work in a GraphQL playground but fail from the skill.

## Moderate Pitfalls

### Pitfall 1: Health Check Race Condition

**What goes wrong:** Skill starts DefraDB and immediately sends queries before the instance is fully ready. First few queries fail, potentially masking real bugs with startup noise.

**Prevention:** Poll `/health-check` endpoint in a loop with timeout (30s). Only proceed after 200 OK response. The `start-defradb.sh` script handles this.

### Pitfall 2: Schema Mutation Ordering

**What goes wrong:** DefraDB requires schemas to be added before documents can be created. If the skill tries to create documents for a collection that doesn't exist yet, the mutation fails silently or with an unhelpful error.

**Prevention:** Always add schema first, verify collection exists (query for it), then create documents. Treat schema creation failure as a fatal error for that test area.

### Pitfall 3: Working Directory Reset Between Bash Calls

**What goes wrong:** Claude Code resets the working directory between Bash tool invocations. Scripts that assume they're in the project root fail when called from sub-agents or after directory changes.

**Prevention:** All scripts use absolute paths. PID files go to `/tmp/`. Scripts receive all paths as arguments, never rely on `pwd`. The `${CLAUDE_SKILL_DIR}` substitution variable provides the skill directory path for referencing bundled scripts.

### Pitfall 4: Sub-Agent Cannot Spawn Sub-Agents

**What goes wrong:** Design assumes nested delegation (orchestrator -> analyzer -> deeper analyzer), which fails because sub-agents cannot spawn other sub-agents in Claude Code.

**Prevention:** Strictly two-level architecture. Orchestrator is the only entity that spawns sub-agents. If a sub-agent needs more information, it returns what it has and the orchestrator dispatches a new sub-agent call.

### Pitfall 5: Background Sub-Agent Permission Denial

**What goes wrong:** Background sub-agents auto-deny any permission not pre-approved. If the query-executor agent needs a Bash permission that wasn't granted upfront, it silently fails.

**Prevention:** Use `permissionMode: acceptEdits` or explicitly pre-approve all needed tools. For the query-executor, the only tool needed is Bash (for curl), which should be pre-approved.

### Pitfall 6: Memory Store Data Loss on Schema Change

**What goes wrong:** Adding a new schema version or modifying schema in DefraDB with memory store may behave differently than persistent stores. Edge cases around schema migration with memory store could produce misleading results.

**Prevention:** For v1, treat memory store as the canonical test backend. Document any memory-store-specific behaviors found. In v2, optionally test with badger for comparison.

## Minor Pitfalls

### Pitfall 1: jq Not Installed

**What goes wrong:** Some minimal environments lack jq. Scripts fail with "command not found."

**Prevention:** Check for jq availability at skill startup. Print clear error message with install instructions if missing.

### Pitfall 2: Port Conflict

**What goes wrong:** Something else is running on port 9181, or a previous orphaned DefraDB instance blocks the port.

**Prevention:** Start script checks for existing process on PID file and kills it. Could also try alternative ports, but keeping it simple for v1.

### Pitfall 3: Large Response Truncation

**What goes wrong:** A query returns thousands of documents. The curl output exceeds what the sub-agent can process, or gets truncated.

**Prevention:** Always use limit/offset in queries when testing large datasets. Default to small document counts (10-50) for test data.

### Pitfall 4: CGO Build Failure

**What goes wrong:** `make build` fails because CGO_ENABLED=1 is not set or C compiler is not available.

**Prevention:** The Makefile handles CGO settings. If build fails, report the error clearly and stop -- don't try to test with a stale binary.

## Phase-Specific Warnings

| Phase Topic | Likely Pitfall | Mitigation |
|-------------|---------------|------------|
| Instance lifecycle | Orphaned processes | Stop hooks + PID file cleanup |
| Build detection | Stale binary tested | Git HEAD comparison with marker file |
| Schema setup | Ordering errors | Schema-first, verify before documents |
| Query execution | String escaping | jq --arg for JSON construction |
| Codebase analysis | Context exhaustion | Dedicated haiku sub-agent |
| Correctness reasoning | Code-aligned bias | Strict first-principles separation |
| Report generation | Incomplete on crash | Incremental writes + Stop hook |
| Sub-agent orchestration | Nested delegation attempt | Two-level architecture only |
| Parallel execution | Permission issues | Pre-approve tools for background agents |

## Sources

- [Claude Code Skills Documentation](https://code.claude.com/docs/en/skills) -- skill lifecycle, hooks, sub-agent limitations
- [Claude Code Sub-agents Documentation](https://code.claude.com/docs/en/sub-agents) -- permission modes, background execution
- DefraDB source code -- API behavior, build requirements, store backends
- PROJECT.md -- correctness oracle hierarchy, constraints
