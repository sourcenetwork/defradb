---
name: defradb:debug
description: Agentically test and debug DefraDB through end-to-end black-box testing via GraphQL HTTP API
disable-model-invocation: true
allowed-tools: Bash Read Grep Glob Write
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
6. Execute targeted queries with hypothesis-based correctness validation
7. Finalize session report with bug summary and reproduction steps
8. Shut down the instance (skip if `--remote`)

## Section 2: Instance Lifecycle

### Step 2a: Remote Check

If `DEFRA_REMOTE` is set (the user provided `--remote <url>`):

- Set `DEFRA_URL="$DEFRA_REMOTE"`
- Skip Steps 2b, 2c, and 2d entirely
- Proceed directly to Step 2e (health check) using `DEFRA_URL`

### Step 2b: Build Staleness Detection

The DefraDB Makefile embeds `git rev-parse HEAD` into the binary via ldflags (`version.GitCommit`). The built binary carries its own source commit -- no marker files are needed.

Execute the following to determine if a rebuild is needed:

Determine the project root dynamically using `git rev-parse --show-toplevel`:

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
  CGO_ENABLED=1 go build -o build/defradb cmd/defradb/main.go
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
  COUNT=$(echo "$DOCS" | jq 'length')
  MUTATION="mutation { add_${COLLECTION}(input: ${DOCS}) { _docID } }"
  RESPONSE=$(curl -s -X POST "$DEFRA_URL/api/v0/graphql" \
    -H "Content-Type: application/json" \
    -d "{\"query\": $(echo "$MUTATION" | jq -Rs .)}")
  ERRORS=$(echo "$RESPONSE" | jq -r '.errors // empty')
  if [ -n "$ERRORS" ] && [ "$ERRORS" != "null" ]; then
    echo "Error loading documents into $COLLECTION: $ERRORS"
  else
    LOADED=$(echo "$RESPONSE" | jq '.data.add_'"$COLLECTION"' | length')
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

### Step 5b: Query Planning

Analyze `DEFRA_PROMPT` to understand the target area. Identify which DefraDB features are involved (CRUD, filtering, relations, aggregations, updates, deletes, etc.). Design schemas and a sequence of queries that systematically probe the target area. Start simple and increase complexity.

Per D-01: the skill is a reasoning agent, not a test harness. Think about what database operations SHOULD do based on fundamental database semantics (CRUD consistency, referential integrity, filter semantics, GraphQL spec), NOT based on what DefraDB's code does. Code-aligned expectations risk validating buggy behavior.

Create the schemas via the pattern in Section 4. Insert any needed documents via the `add_<CollectionName>` mutation pattern from Section 3. Once the setup is complete, enter the hypothesis-then-verify loop below.

### Step 5c: Hypothesis-Then-Verify Loop

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
echo "$RESPONSE" | jq .
```

**3. Evaluate the response against the hypothesis.**

- If the response matches the hypothesis: log as `PASS` with the one-liner in the chronological log.
- If the response does NOT match: proceed to Step 5d (classification) and Step 5e (reproduction) before continuing.

**4. Append a row to the chronological log** in the report file (via the Write tool -- read the current file, add the row, write it back):

```
| N | `<gql>` | Expect: <prediction> | <brief actual> | PASS |
```

If `DEFRA_VERBOSE` is set to `true`, also include the full request body and full response body below the log table in a fenced code block labeled `Query N`. Per D-05: verbose mode controls whether full request/response bodies appear; non-verbose mode keeps the log compact with only brief expectations and pass/fail status.

### Step 5d: Error Classification

When a query result does not match the hypothesis, classify the failure into exactly ONE of three categories. Per D-08 and CORR-02: strict 3-category classification with NO sub-categories.

Inspect the response structure with `jq`:

```bash
ERRORS=$(echo "$RESPONSE" | jq '.errors // empty')
```

Apply the classification rules in order. Classification is based on response structure, not subjective judgment:

1. **PARSE ERROR** -- The response contains an `errors` array AND the error message indicates a parsing/syntax failure (e.g., contains "Syntax Error", "Unexpected", "unknown field", schema validation failures). The query never reached execution.

2. **RUNTIME ERROR** -- The response contains an `errors` array AND the error message indicates an execution-time failure (e.g., contains "failed to", transaction errors, collection not found at runtime). The query was parsed but failed during execution.

3. **DATA CORRECTNESS ISSUE** -- The response has NO `errors` array (or errors is null/empty), `data` is present, but the returned data does not match the hypothesis. The query succeeded technically but produced wrong results.

Pitfall: an empty result set (`{"data": {"User": []}}`) with no errors is NOT a runtime error -- it is a DATA CORRECTNESS ISSUE if the hypothesis expected non-empty data. Classification reads the response shape, not how the data "feels."

### Step 5e: Anomaly Reproduction

When a mismatch is detected (any of the 3 classifications above), confirm reproducibility before recording a bug. Per D-03: 1 re-run, 2 total executions.

1. Log the first failure with full details (query, expected, actual, classification) in the report file.
2. Re-execute the EXACT SAME query against the SAME instance state. Do NOT perform any intervening mutations that could change state between the first and second run.
3. Compare the second result:
   - If second execution ALSO fails with the same mismatch: mark as **CONFIRMED**. Record for the summary report.
   - If second execution passes (different result): mark as **FLAKY/UNCONFIRMED**. Note in the chronological log but do NOT include in the bug summary.

With memory store, behavior should be deterministic, so confirmed failures are real bugs.

Per D-04 and INVK-02: Do NOT interrupt the user mid-session. Continue testing regardless of how many anomalies are found. All confirmed findings are batched into the end-of-session report produced in Step 5g.

If the query under test is itself a mutation (e.g., an update or delete that changes state), the re-run operates on different state. In that case, reconstruct the precondition state first (re-insert the documents), then re-run the mutation, and note in the reproduction entry that reproduction required re-establishing state.

### Step 5f: @explain Investigation

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

### Step 5g: Session Report Finalization

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

Only bugs marked **CONFIRMED** in Step 5e appear in the Bugs Found section. Unconfirmed/flaky anomalies stay in the chronological log but are excluded from the bug summary.

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
