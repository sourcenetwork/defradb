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
- Everything remaining after flag extraction is the user's prompt describing what area to test and debug.

Store the parsed values:
- `DEFRA_STORE` -- the store backend (default `memory`)
- `DEFRA_FIXTURES` -- path to fixture file, or empty
- `DEFRA_REMOTE` -- remote URL, or empty
- `DEFRA_PROMPT` -- the user's prompt text with flags removed

## Section 1: Overview

This skill performs end-to-end black-box testing of DefraDB. The workflow is:

1. Parse arguments from `$ARGUMENTS`
2. Build DefraDB if the binary is stale or missing (skip if `--remote`)
3. Start a fresh instance (skip if `--remote`)
4. Poll health check until ready
5. Load fixtures if `--fixtures` was provided
6. Analyze the user's prompt and execute targeted GraphQL queries
7. Report findings
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

```bash
cd /home/tacos/Workspace/go/src/github.com/sourcenetwork/defradb.worktrees/jsimnz-feat-db-debug-skill

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
- `--url="127.0.0.1:9281"` uses port 9281 to avoid conflicting with any user-running instance on the default port 9181

### Step 2d: PID Tracking and Cleanup

Immediately after starting the instance, persist session state for cross-Bash-call access:

```bash
echo "$DEFRA_PID" > "$DEFRA_TMPDIR/defradb.pid"
echo "$DEFRA_TMPDIR" > /tmp/.defradb-debug-session
```

Since each Bash tool call runs in a fresh shell, the PID and tmpdir variables do not persist. To access them in subsequent Bash calls, read from these files:

```bash
DEFRA_TMPDIR=$(cat /tmp/.defradb-debug-session 2>/dev/null)
DEFRA_PID=$(cat "$DEFRA_TMPDIR/defradb.pid" 2>/dev/null)
```

Set a cleanup trap in the initial Bash call:

```bash
trap 'kill $DEFRA_PID 2>/dev/null; rm -rf "$DEFRA_TMPDIR"' EXIT
```

When the debugging session is complete (or on any error that prevents continuation), execute cleanup explicitly:

```bash
DEFRA_TMPDIR=$(cat /tmp/.defradb-debug-session 2>/dev/null)
if [ -n "$DEFRA_TMPDIR" ]; then
  DEFRA_PID=$(cat "$DEFRA_TMPDIR/defradb.pid" 2>/dev/null)
  if [ -n "$DEFRA_PID" ]; then
    kill "$DEFRA_PID" 2>/dev/null
    # Wait briefly for clean shutdown
    sleep 1
    kill -9 "$DEFRA_PID" 2>/dev/null
  fi
  rm -rf "$DEFRA_TMPDIR"
  rm -f /tmp/.defradb-debug-session
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
    DEFRA_TMPDIR=$(cat /tmp/.defradb-debug-session 2>/dev/null)
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

For each collection name in the `documents` object, iterate over each document and POST an `add_<CollectionName>` mutation:

```bash
for COLLECTION in $(jq -r '.documents | keys[]' "$DEFRA_FIXTURES"); do
  COUNT=$(jq -r ".documents[\"$COLLECTION\"] | length" "$DEFRA_FIXTURES")
  LOADED=0
  for i in $(seq 0 $(($COUNT - 1))); do
    DOC=$(jq -c ".documents[\"$COLLECTION\"][$i]" "$DEFRA_FIXTURES")
    MUTATION="mutation { add_${COLLECTION}(input: ${DOC}) { _docID } }"
    RESPONSE=$(curl -s -X POST "$DEFRA_URL/api/v0/graphql" \
      -H "Content-Type: application/json" \
      -d "{\"query\": \"$MUTATION\"}")
    ERRORS=$(echo "$RESPONSE" | jq -r '.errors // empty')
    if [ -n "$ERRORS" ] && [ "$ERRORS" != "null" ]; then
      echo "Error creating document in $COLLECTION: $ERRORS"
    else
      LOADED=$((LOADED + 1))
    fi
  done
  echo "Loaded $LOADED/$COUNT documents into $COLLECTION."
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

Analyze the user's prompt (`DEFRA_PROMPT`) to understand what area of DefraDB to test. Then execute a systematic testing workflow:

1. **Understand the target area.** Read the user's prompt. Identify which DefraDB features are involved (CRUD, filtering, relations, aggregations, updates, deletes, etc.).

2. **Create appropriate schemas.** Design GraphQL SDL schemas that exercise the target area. Start simple and increase complexity. POST each schema via `/api/v0/collections`.

3. **Generate and execute test queries.** Build a series of GraphQL operations that probe the target area:
   - Basic CRUD (create, read, update, delete)
   - Edge cases (empty inputs, large values, special characters)
   - Filters and conditions
   - Relations and joins (if applicable)
   - Ordering, limiting, offsetting
   - Aggregations (if applicable)

4. **Evaluate results.** For each query, reason about what the correct result should be based on database first principles and GraphQL semantics -- not based on what the code does. Compare actual results against expected results.

5. **Report findings.** For each test:
   - What was tested (the query)
   - What was expected (and why)
   - What actually happened
   - Whether it passed or failed
   - If failed: is it reproducible? What is the severity?

Correctness validation and structured bug reports will be added in a future update.

## Section 6: Shutdown

If `DEFRA_REMOTE` is set, skip shutdown entirely. Report: "Remote instance at $DEFRA_URL -- no shutdown needed."

Otherwise, execute a clean shutdown:

```bash
DEFRA_TMPDIR=$(cat /tmp/.defradb-debug-session 2>/dev/null)
if [ -n "$DEFRA_TMPDIR" ]; then
  DEFRA_PID=$(cat "$DEFRA_TMPDIR/defradb.pid" 2>/dev/null)
  if [ -n "$DEFRA_PID" ]; then
    kill "$DEFRA_PID" 2>/dev/null
    sleep 1
    kill -9 "$DEFRA_PID" 2>/dev/null
  fi
  rm -rf "$DEFRA_TMPDIR"
  rm -f /tmp/.defradb-debug-session
  echo "DefraDB instance shut down cleanly."
else
  echo "No session file found. Instance may have already been cleaned up."
fi
```

Always execute shutdown before finishing the skill session, even if errors occurred during testing.
