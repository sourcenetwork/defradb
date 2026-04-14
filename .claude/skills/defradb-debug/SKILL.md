---
name: defradb:debug
description: Agentically test and debug DefraDB through end-to-end black-box testing via GraphQL HTTP API
disable-model-invocation: true
allowed-tools: Bash Read Grep Glob Write Agent
argument-hint: "<prompt> [--fixtures path] [--store memory|badger] [--remote url]"
---

# DefraDB Debug Skill

Given a prompt describing an area of concern, this skill builds DefraDB (if stale), starts a fresh instance, executes GraphQL queries against the HTTP API, and documents findings. All testing is black-box via HTTP -- no direct Go API calls. The skill manages the full instance lifecycle: build, start, health check, query execution, and clean shutdown.

The variable `$ARGUMENTS` contains the user's full input string. Parse it for the following flags before proceeding:

- `--fixtures <path>` -- Path to a JSON fixture file. If present, load fixtures after the instance is healthy (see Section 3).
- `--store <memory|badger>` -- Store backend to use. Default: `memory`.
- `--remote <url>` -- URL of an existing DefraDB instance. If present, skip build/start/shutdown and connect directly (see Section 2a).
- `--verbose` -- Include full request/response bodies in the session report. Default: off (brief expectations and pass/fail only).
- Everything remaining after flag extraction is the user's prompt describing what area to test and debug.

Store the parsed values:
- `DEFRA_STORE` -- the store backend (default `memory`)
- `DEFRA_FIXTURES` -- path to fixture file, or empty
- `DEFRA_REMOTE` -- remote URL, or empty
- `DEFRA_VERBOSE` -- `true` if `--verbose` was provided, empty otherwise
- `DEFRA_PROMPT` -- the user's prompt text with flags removed

## Section 1: Overview

This skill performs end-to-end black-box testing of DefraDB. The workflow is:

1. Parse arguments from `$ARGUMENTS`
2. Build DefraDB if the binary is stale or missing (skip if `--remote`)
3. Start a fresh instance (skip if `--remote`)
4. Poll health check until ready
5. Load fixtures if `--fixtures` was provided
6. Introspect schema via `sdl generate` to produce expanded GraphQL schema
7. (Optional) Spawn codebase analysis sub-agent for target area
8. Spawn query generation sub-agent with expanded schema + targets
9. Execute validated queries with hypothesis-based correctness validation
10. Finalize session report with bug summary and reproduction steps
11. Shut down the instance (skip if `--remote`)

## Section 2: Instance Lifecycle

### Step 2a: Remote Check

If `DEFRA_REMOTE` is set (the user provided `--remote <url>`):

- Set `DEFRA_URL="$DEFRA_REMOTE"`
- Skip Steps 2b, 2c, and 2d entirely
- Proceed directly to Step 2e (health check) using `DEFRA_URL`

### Step 2b: Build Staleness Detection

The DefraDB Makefile embeds `git rev-parse HEAD` into the binary via ldflags (`version.GitCommit`). The built binary carries its own source commit -- no marker files are needed.

**Important:** a plain `go build` does NOT inject the ldflags; the resulting binary reports an empty commit and staleness detection would then rebuild every session. The skill must pass the same ldflags the Makefile uses so the binary's `version` command reports the real commit.

Execute the following to determine if a rebuild is needed. Determine the project root dynamically using `git rev-parse --show-toplevel`:

```bash
DEFRA_ROOT=$(git rev-parse --show-toplevel)
cd "$DEFRA_ROOT"

CURRENT_HEAD=$(git rev-parse HEAD)

if [ -f ./build/defradb ]; then
  BUILT_COMMIT=$(./build/defradb version --format json | jq -r '.commit')
else
  BUILT_COMMIT="none"
fi

if [ "$CURRENT_HEAD" != "$BUILT_COMMIT" ]; then
  echo "Build is stale (HEAD=$CURRENT_HEAD, built=$BUILT_COMMIT). Rebuilding..."
  VERSION_GOINFO=$(go version)
  VERSION_GITCOMMIT=$(git rev-parse HEAD)
  VERSION_GITCOMMITDATE=$(git show -s --date=short --format=%cd HEAD)
  VERSION_GITRELEASE="dev-$(git symbolic-ref -q --short HEAD 2>/dev/null || echo HEAD)"
  CGO_ENABLED=1 go build -trimpath -ldflags "\
    -X 'github.com/sourcenetwork/defradb/version.GoInfo=${VERSION_GOINFO}' \
    -X 'github.com/sourcenetwork/defradb/version.GitRelease=${VERSION_GITRELEASE}' \
    -X 'github.com/sourcenetwork/defradb/version.GitCommit=${VERSION_GITCOMMIT}' \
    -X 'github.com/sourcenetwork/defradb/version.GitCommitDate=${VERSION_GITCOMMITDATE}'" \
    -o build/defradb cmd/defradb/main.go
  if [ $? -ne 0 ]; then
    echo "BUILD FAILED. Cannot proceed."
    exit 1
  fi
  echo "Build complete."
else
  echo "Build is current (commit=$CURRENT_HEAD)."
fi
```

Use a 300-second timeout on the build command to handle cold builds. If the build fails, report the error and stop -- do not proceed with a stale or missing binary.

Alternatively, the equivalent `make build` invocation produces the same binary with ldflags injected. Using `go build` with explicit ldflags (as above) is preferred when the skill needs fine-grained control; fall back to `make build` only if the ldflag list drifts.

### Step 2c: Start Instance

Start a fresh DefraDB instance with isolated state:

```bash
DEFRA_TMPDIR=$(mktemp -d)
./build/defradb start \
  --store=$DEFRA_STORE \
  --no-keyring \
  --no-p2p \
  --development \
  --rootdir "$DEFRA_TMPDIR" \
  --url="127.0.0.1:9281" \
  > "$DEFRA_TMPDIR/defradb.log" 2>&1 &
DEFRA_PID=$!
DEFRA_URL="http://127.0.0.1:9281"
```

- `--store=$DEFRA_STORE` uses the parsed store backend (default `memory`)
- `--no-keyring` skips keyring operations (not needed for testing)
- `--no-p2p` disables peer-to-peer networking (out of scope for v1)
- `--development` enables dev mode (allows purge and other dev operations)
- `--rootdir` points to an isolated temp directory for this session
- `--url="127.0.0.1:9281"` uses port 9281 to avoid conflicting with any user-running instance on the default port 9181. If port 9281 is unavailable (another debug session is running), pick a different unused port in the 9200-9299 range.

### Step 2d: PID Tracking and Cleanup

Immediately after starting the instance, persist session state for cross-Bash-call access. The session file is named by port to avoid collisions between concurrent debug sessions:

```bash
DEFRA_PORT=9281  # or whichever port was chosen
echo "$DEFRA_PID" > "$DEFRA_TMPDIR/defradb.pid"
echo "$DEFRA_TMPDIR" > "/tmp/.defradb-debug-session-${DEFRA_PORT}"
```

Since each Bash tool call runs in a fresh shell, the PID and tmpdir variables do not persist. To access them in subsequent Bash calls, read from the port-specific session file:

```bash
DEFRA_PORT=9281  # use the same port chosen during startup
DEFRA_TMPDIR=$(cat "/tmp/.defradb-debug-session-${DEFRA_PORT}" 2>/dev/null)
DEFRA_PID=$(cat "$DEFRA_TMPDIR/defradb.pid" 2>/dev/null)
```

Set a cleanup trap in the initial Bash call:

```bash
trap 'kill $DEFRA_PID 2>/dev/null; rm -rf "$DEFRA_TMPDIR"' EXIT
```

When the debugging session is complete (or on any error that prevents continuation), execute cleanup explicitly:

```bash
DEFRA_PORT=9281  # use the same port chosen during startup
DEFRA_TMPDIR=$(cat "/tmp/.defradb-debug-session-${DEFRA_PORT}" 2>/dev/null)
if [ -n "$DEFRA_TMPDIR" ]; then
  DEFRA_PID=$(cat "$DEFRA_TMPDIR/defradb.pid" 2>/dev/null)
  if [ -n "$DEFRA_PID" ]; then
    kill "$DEFRA_PID" 2>/dev/null
    sleep 1
    kill -9 "$DEFRA_PID" 2>/dev/null
  fi
  rm -rf "$DEFRA_TMPDIR"
  rm -f "/tmp/.defradb-debug-session-${DEFRA_PORT}"
fi
```

Do not leave DefraDB processes running. Always clean up before finishing the skill session.

### Step 2e: Health Check Polling

Poll the health check endpoint until the instance is ready:

```bash
DEFRA_URL="http://127.0.0.1:9281"
MAX_RETRIES=30
for i in $(seq 1 $MAX_RETRIES); do
  if curl -sf "$DEFRA_URL/health-check" > /dev/null 2>&1; then
    echo "DefraDB is healthy."
    break
  fi
  if [ "$i" -eq "$MAX_RETRIES" ]; then
    echo "DefraDB failed to start within ${MAX_RETRIES}s."
    DEFRA_TMPDIR=$(cat "/tmp/.defradb-debug-session-${DEFRA_PORT}" 2>/dev/null)
    if [ -n "$DEFRA_TMPDIR" ]; then
      echo "Last 20 lines of log:"
      tail -20 "$DEFRA_TMPDIR/defradb.log"
    fi
    echo "STOPPING: Instance did not become healthy."
    exit 1
  fi
  sleep 1
done
```

After the health check passes, verify full API readiness:

```bash
curl -sf "$DEFRA_URL/api/v0/collections" > /dev/null 2>&1
if [ $? -ne 0 ]; then
  echo "WARNING: Health check passed but API is not responding. Check logs."
fi
```

## Section 3: Fixture Loading

Only execute this section if `DEFRA_FIXTURES` is set (the user provided `--fixtures <path>`).

The fixture file is JSON with two top-level keys:

```json
{
  "schema": "type User { name: String, age: Int }",
  "documents": {
    "User": [
      {"name": "Alice", "age": 30},
      {"name": "Bob", "age": 25}
    ]
  }
}
```

- `"schema"` -- a string containing the GraphQL SDL schema definition
- `"documents"` -- an object mapping collection names to arrays of document objects

Load the fixtures:

**Step 1: Add the schema.**

```bash
SCHEMA=$(jq -r '.schema' "$DEFRA_FIXTURES")
RESPONSE=$(curl -s -X POST "$DEFRA_URL/api/v0/collections" \
  -H "Content-Type: text/plain" \
  -d "$SCHEMA")
echo "Schema response: $RESPONSE"
```

Check the response for errors. If the schema addition fails, report the error and the SDL that was sent, then stop fixture loading.

**Step 2: Create documents for each collection.**

The `add_<CollectionName>` mutation accepts an array of documents in its `input` argument, so all documents for a collection can be loaded in a single request:

```bash
for COLLECTION in $(jq -r '.documents | keys[]' "$DEFRA_FIXTURES"); do
  DOCS=$(jq -c ".documents[\"$COLLECTION\"]" "$DEFRA_FIXTURES")
  COUNT=$(printf '%s\n' "$DOCS" | jq 'length')
  MUTATION="mutation { add_${COLLECTION}(input: ${DOCS}) { _docID } }"
  RESPONSE=$(curl -s -X POST "$DEFRA_URL/api/v0/graphql" \
    -H "Content-Type: application/json" \
    -d "{\"query\": $(printf '%s' "$MUTATION" | jq -Rs .)}")
  ERRORS=$(printf '%s\n' "$RESPONSE" | jq -r '.errors // empty')
  if [ -n "$ERRORS" ] && [ "$ERRORS" != "null" ]; then
    echo "Error loading documents into $COLLECTION: $ERRORS"
  else
    LOADED=$(printf '%s\n' "$RESPONSE" | jq '.data.add_'"$COLLECTION"' | length')
    echo "Loaded $LOADED/$COUNT documents into $COLLECTION."
  fi
done
```

## Section 4: Query Execution

Execute GraphQL queries against the running instance:

**GraphQL query/mutation:**

```bash
curl -s -X POST "$DEFRA_URL/api/v0/graphql" \
  -H "Content-Type: application/json" \
  -d '{"query": "<gql>"}' | jq .
```

**Add a schema:**

```bash
curl -s -X POST "$DEFRA_URL/api/v0/collections" \
  -H "Content-Type: text/plain" \
  -d '<sdl>'
```

Response format is standard GraphQL: `{"data": {...}, "errors": [...]}`. Always check both `data` and `errors` fields. Use `jq` to parse responses for readability.

When constructing queries, escape double quotes inside the JSON body properly. For mutations with string values, use backslash-escaped quotes: `\"value\"`.

**Portability note:** when a response JSON is captured into a shell variable and then fed to `jq`, use `printf '%s\n' "$RESPONSE" | jq .` -- NOT `echo "$RESPONSE" | jq .`. In `zsh`, the built-in `echo` interprets backslash escapes by default, so a legitimate `\n` inside a JSON string (common in DefraDB error messages) gets converted to a literal newline and the resulting stream is no longer valid JSON. `printf '%s\n'` is portable across `bash`, `zsh`, and `sh` and preserves the response bytes exactly. The same concern applies to `jq -Rs .` on authored queries when the query happens to contain newline escapes. All response-handling examples in Section 5 use `printf` for this reason.

## Section 5: Debugging Loop

This section replaces passive query execution with active correctness validation. Every query goes through a hypothesis-then-verify cycle. Anomalies are reproduced before reporting. All findings are batched to an end-of-session report -- no mid-session user interrupts.

### Step 5a: Session Report Initialization

At the START of the debugging loop (before any queries), create the session report file. Determine the next available sequence number for today's date:

```bash
DATE=$(date +%Y-%m-%d)
SEQ=1
while [ -f "DEBUG_PROGRESS_${DATE}_$(printf '%02d' $SEQ).md" ]; do
  SEQ=$((SEQ + 1))
done
REPORT_FILE="DEBUG_PROGRESS_${DATE}_$(printf '%02d' $SEQ).md"
echo "$REPORT_FILE"
```

Use the Write tool (not Bash heredoc) to create the initial report file in the working directory (repo root) with this header:

```markdown
# Debug Session: <DATE>

**Prompt:** <DEFRA_PROMPT>
**Store:** <DEFRA_STORE>
**Instance:** <local:9281 or remote URL>
**Started:** <timestamp>

## Chronological Log

| # | Query | Expectation | Result | Status |
|---|-------|-------------|--------|--------|
```

Remember the report filename for use in all subsequent steps. The report file accumulates findings as the session progresses so that state is preserved across Bash tool calls -- do not wait until the end to write the log entries.

### Step 5b: Query Planning and Schema Setup

Analyze `DEFRA_PROMPT` to understand the target area. Identify which DefraDB features are involved (CRUD, filtering, relations, aggregations, updates, deletes, etc.). Design schemas and a sequence of queries that systematically probe the target area. Start simple and increase complexity.

Per D-01: the skill is a reasoning agent, not a test harness. Think about what database operations SHOULD do based on fundamental database semantics (CRUD consistency, referential integrity, filter semantics, GraphQL spec), NOT based on what DefraDB's code does. Code-aligned expectations risk validating buggy behavior.

Create the schemas via the pattern in Section 4. Insert any needed documents via the `add_<CollectionName>` mutation pattern from Section 3.

After loading the schema (either from fixtures in Section 3, or from schemas created above), run `sdl generate` to produce the fully-expanded GraphQL schema. Schema introspection happens on-demand -- after every schema load, not just once at startup. This handles the case where the skill creates new schemas during a testing session.

```bash
# Write SDL to temp file (avoids shell escaping issues with echo)
DEFRA_PORT=9281
SDL_FILE="/tmp/defradb-debug-sdl-${DEFRA_PORT}.sdl"
printf '%s' "$SDL" > "$SDL_FILE"

# Generate expanded schema
DEFRA_ROOT=$(git rev-parse --show-toplevel)
EXPANDED_SCHEMA="/tmp/defradb-debug-schema-${DEFRA_PORT}.graphql"
"$DEFRA_ROOT/build/defradb" sdl generate "$SDL_FILE" -o "$EXPANDED_SCHEMA" -y

if [ $? -ne 0 ]; then
  echo "ERROR: sdl generate failed. Cannot validate queries against schema."
  # Continue without schema validation -- queries will be generated without pre-validation
else
  echo "Expanded schema: $(wc -l < "$EXPANDED_SCHEMA") lines"
fi
```

After generating the expanded schema, extract a type catalog -- a summary of available types, their fields, field types, and relations. Read the expanded schema file and identify the top-level `Query` and `Mutation` types to understand available operations. This catalog serves as a "menu of what can be queried" for the query generation sub-agent and for the orchestrator's query planning.

When the skill auto-generates schemas (no `--fixtures`), the orchestrator designs the SDL based on the target area, loads it into the running instance via the collections endpoint, AND runs `sdl generate` on the same SDL to get the expanded schema. Both steps use the same SDL string.

If schema introspection fails (`sdl generate` returns non-zero), log the failure and fall back to generating queries without schema validation. The session can still proceed -- queries just will not be pre-validated.

Once the setup is complete, proceed to query generation via sub-agent (Step 5d) or directly to the hypothesis-then-verify loop (Step 5e) if schema introspection is unavailable.

### Step 5c: Codebase Analysis (Optional)

*This step is populated by Task 2. See Step 5c below.*

### Step 5d: Query Generation via Sub-Agent

The orchestrator spawns a query generation sub-agent using the Agent tool. The sub-agent absorbs the full expanded schema (which can be 2600+ lines) so the orchestrator does not need to hold it in context. This implements the context isolation boundary between schema details and the orchestrator's reasoning.

The sub-agent:

1. Reads the expanded schema from `$EXPANDED_SCHEMA` (the temp file produced by `sdl generate`)
2. Receives a description of what to test (derived from `DEFRA_PROMPT` and optionally from codebase analysis)
3. Generates edge-case queries targeting the described area
4. Validates EACH query against the expanded schema using `npx @graphql-inspector/cli validate`
5. Returns ONLY queries that pass validation, with a one-line description of what each tests
6. Does NOT include expected results or correctness criteria (dual-track separation -- codebase knowledge determines WHERE to probe, first-principles reasoning determines WHAT SHOULD HAPPEN)

Spawn the sub-agent with allowed tools: `Bash`, `Read`. The sub-agent uses Bash for running `npx @graphql-inspector/cli validate` and writing temp query files, and Read for reading the expanded schema file.

**Sub-agent prompt template:**

```
You are a GraphQL query generator for DefraDB testing.

Read the expanded GraphQL schema from: /tmp/defradb-debug-schema-{PORT}.graphql

Target area: {description from orchestrator}
Test targets: {list from codebase analysis, if available}

Generate {N} GraphQL queries that:
1. Test edge cases and boundary conditions for the target area
2. Include a mix of queries, mutations, and filters
3. Cover both happy paths and expected-to-fail scenarios

For EACH query:
1. Write it to /tmp/defradb-debug-query-{N}.graphql
2. Run: npx @graphql-inspector/cli validate /tmp/defradb-debug-query-{N}.graphql /tmp/defradb-debug-schema-{PORT}.graphql
3. If validation fails, fix the query and re-validate
4. Only return queries that pass validation

Return a numbered list of valid queries with a one-line description of what each tests.
Do NOT include expected results or correctness criteria -- only what the query tests.
```

After the sub-agent returns, the orchestrator parses the numbered list of validated queries. These queries feed into the hypothesis-then-verify loop (Step 5e). The orchestrator formulates hypotheses for each query using first-principles reasoning BEFORE execution -- exactly as in Phase 2.

If the expanded schema file does not exist (`sdl generate` failed earlier), skip sub-agent query generation and fall back to the orchestrator generating queries inline (the Phase 2 behavior). Log: "Schema introspection unavailable -- generating queries without pre-validation."

### Step 5e: Hypothesis-Then-Verify Loop

For EACH query in the sequence, follow these substeps in strict order.

**1. BEFORE executing the query, formulate a hypothesis.**

Write a brief one-liner prediction: `Expect: <prediction>` (e.g., `Expect: 3 docs matching filter`, `Expect: empty result`, `Expect: error because field does not exist`).

Base the prediction on database first principles: what data was inserted, what the query asks for, and how a correct database should behave. Log the hypothesis BEFORE the curl command. This ordering is critical to avoid confirmation bias -- never look at the response first and then rationalize what was "expected."

Per D-02: keep the expectation short. Detailed reasoning ("because I inserted 3 Users with age > 30 and the filter is age > 30") only surfaces when actual != expected, not for every query.

**2. Execute the query** using the curl pattern from Section 4:

```bash
RESPONSE=$(curl -s -X POST "$DEFRA_URL/api/v0/graphql" \
  -H "Content-Type: application/json" \
  -d '{"query": "<gql>"}')
printf '%s\n' "$RESPONSE" | jq .
```

**3. Evaluate the response against the hypothesis.**

- If the response matches the hypothesis: log as `PASS` with the one-liner in the chronological log.
- If the response does NOT match: proceed to Step 5f (classification) and Step 5g (reproduction) before continuing.

**4. Append a row to the chronological log** in the report file (via the Write tool -- read the current file, add the row, write it back):

```
| N | `<gql>` | Expect: <prediction> | <brief actual> | PASS |
```

If `DEFRA_VERBOSE` is set to `true`, also include the full request body and full response body below the log table in a fenced code block labeled `Query N`. Per D-05: verbose mode controls whether full request/response bodies appear; non-verbose mode keeps the log compact with only brief expectations and pass/fail status.

### Step 5f: Error Classification

When a query result does not match the hypothesis, classify the failure into exactly ONE of three categories. Per D-08 and CORR-02: strict 3-category classification with NO sub-categories.

Inspect the response structure with `jq`:

```bash
ERRORS=$(printf '%s\n' "$RESPONSE" | jq '.errors // empty')
```

Apply the classification rules in order. Classification is based on response structure, not subjective judgment:

1. **PARSE ERROR** -- The response contains an `errors` array AND the error message indicates a parsing, syntax, or schema validation failure (e.g., contains "Syntax Error", "Unexpected", "unknown field", "Cannot query field", "Expected type", or any GraphQL schema validation failure). The query never reached execution. **Note:** in DefraDB, querying a collection that does not exist, referencing an unknown field, or passing a wrongly-typed argument all surface as schema validation failures and MUST be classified as PARSE ERROR, not RUNTIME ERROR.

2. **RUNTIME ERROR** -- The response contains an `errors` array AND the error message indicates an execution-time failure (e.g., contains "failed to", transaction errors, storage errors, CRDT merge failures, internal panics). The query was parsed and schema-validated but failed during execution. RUNTIME ERRORS are comparatively rare against an in-memory store -- DefraDB catches most problems at parse/validation time. If you are unsure whether an error is a parse or runtime error, check whether the error message references an execution verb ("failed to fetch", "failed to commit", "could not read block") -- that signals RUNTIME ERROR.

3. **DATA CORRECTNESS ISSUE** -- The response has NO `errors` array (or errors is null/empty), `data` is present, but the returned data does not match the hypothesis. The query succeeded technically but produced wrong results.

Pitfall: an empty result set (`{"data": {"User": []}}`) with no errors is NOT a runtime error -- it is a DATA CORRECTNESS ISSUE if the hypothesis expected non-empty data. Classification reads the response shape, not how the data "feels."

### Step 5g: Anomaly Reproduction

When a mismatch is detected (any of the 3 classifications above), confirm reproducibility before recording a bug. Per D-03: 1 re-run, 2 total executions.

1. Log the first failure with full details (query, expected, actual, classification) in the report file.
2. Re-execute the EXACT SAME query against the SAME instance state. Do NOT perform any intervening mutations that could change state between the first and second run.
3. Compare the second result:
   - If second execution ALSO fails with the same mismatch: mark as **CONFIRMED**. Record for the summary report.
   - If second execution passes (different result): mark as **FLAKY/UNCONFIRMED**. Note in the chronological log but do NOT include in the bug summary.

With memory store, behavior should be deterministic, so confirmed failures are real bugs.

Per D-04 and INVK-02: Do NOT interrupt the user mid-session. Continue testing regardless of how many anomalies are found. All confirmed findings are batched into the end-of-session report produced in Step 5i.

If the query under test is itself a mutation (e.g., an update or delete that changes state), the re-run operates on different state. In that case, reconstruct the precondition state first (re-insert the documents), then re-run the mutation, and note in the reproduction entry that reproduction required re-establishing state.

### Step 5h: @explain Investigation

Per D-09: the `@explain` GraphQL directive is a general-purpose investigation tool. When investigating anomalies or when the query plan might be relevant to understanding a failure, run an explain query against the same query body:

```bash
curl -s -X POST "$DEFRA_URL/api/v0/graphql" \
  -H "Content-Type: application/json" \
  -d '{"query": "query @explain(type: simple) { <same_query_body> }"}' | jq .
```

Three explain types are available:

- `simple` -- Plan graph with node attributes (filter values, collection names, prefixes). Most useful for understanding planner decisions. Use this by default.
- `execute` -- Plan graph with execution statistics. Useful for seeing what happened during execution.
- `debug` -- Plan graph structure without attribute details. Quick structural overview.

Include the explain output in the bug report when it adds diagnostic value (regardless of `DEFRA_VERBOSE` mode). The explain output answers "why did the query plan look like this," which is valuable context for bug investigation -- especially for data correctness issues involving filters, joins, or indexes.

### Step 5i: Session Report Finalization

After all queries are complete, finalize the report file by reading it, appending the summary section, and writing it back with the Write tool. The finalized file must contain both the chronological log and the structured summary report, separated by headings (per D-07).

The final report structure:

```markdown
# Debug Session: <DATE>

**Prompt:** ...
**Store:** ...
**Instance:** ...
**Started:** ...
**Completed:** <timestamp>

## Chronological Log

| # | Query | Expectation | Result | Status |
|---|-------|-------------|--------|--------|
| 1 | `query { ... }` | Expect: ... | ... | PASS |
| 2 | `query { ... }` | Expect: ... | ... | FAIL |

### Query 2: Detailed Failure Analysis

**Expected:** <what and why, based on first principles>
**Actual:** <what happened>
**Reproduction:** Re-ran query -- same result. CONFIRMED.
**Classification:** <PARSE ERROR | RUNTIME ERROR | DATA CORRECTNESS ISSUE>

**Explain output:** (if relevant)

<explain plan output>

## Summary

### Session Statistics

- Queries executed: N
- Passed: N
- Failed: N (M confirmed, K unconfirmed)

### Bugs Found: N

#### Bug 1: <descriptive title>

**Classification:** <PARSE ERROR | RUNTIME ERROR | DATA CORRECTNESS ISSUE>
**Description:** <what went wrong and why it is wrong, based on first principles>

**Reproduction Steps:**
1. Add schema: `<SDL>`
2. Create documents: <mutations>
3. Execute: `<query>`
4. Expected: <result>
5. Actual: <result>

**Explain output:** (if relevant)

<query plan>

### Observations

- <things noticed that are not bugs but worth noting>

### Recommendations

- <suggested areas for further investigation>
```

For each confirmed bug (REPT-04): include the MINIMAL reproduction steps. This means the smallest schema, fewest documents, and simplest query that reproduces the issue. Strip away anything not needed for reproduction. The goal is that a reader can copy the steps verbatim and trigger the same bug on a fresh instance.

Only bugs marked **CONFIRMED** in Step 5g appear in the Bugs Found section. Unconfirmed/flaky anomalies stay in the chronological log but are excluded from the bug summary.

## Section 6: Shutdown

If `DEFRA_REMOTE` is set, skip shutdown entirely. Report: "Remote instance at $DEFRA_URL -- no shutdown needed."

Otherwise, execute a clean shutdown:

```bash
DEFRA_PORT=9281  # use the same port chosen during startup
DEFRA_TMPDIR=$(cat "/tmp/.defradb-debug-session-${DEFRA_PORT}" 2>/dev/null)
if [ -n "$DEFRA_TMPDIR" ]; then
  DEFRA_PID=$(cat "$DEFRA_TMPDIR/defradb.pid" 2>/dev/null)
  if [ -n "$DEFRA_PID" ]; then
    kill "$DEFRA_PID" 2>/dev/null
    sleep 1
    kill -9 "$DEFRA_PID" 2>/dev/null
  fi
  rm -rf "$DEFRA_TMPDIR"
  rm -f "/tmp/.defradb-debug-session-${DEFRA_PORT}"
  echo "DefraDB instance shut down cleanly."
else
  echo "No session file found. Instance may have already been cleaned up."
fi
```

Always execute shutdown before finishing the skill session, even if errors occurred during testing.
