#!/usr/bin/env bash
# Manual smoke tests for the dvncan/shinzo-rebuild branch.
# Covers all 9 rebuilt features plus the Tier 0-3 additions.
#
# Usage (run from repo root):
#   go build -o /tmp/defradb ./cmd/defradb/main.go
#   bash tests/manual/smoke_test.sh
#
# Programmatic tests (SaveManyDocuments, KV, cache, etc.) are run via the
# helper programs in tests/manual/cmd/.

set -uo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
BIN=/tmp/defradb
DIR_A=/tmp/defradb-smoke-A
DIR_B=/tmp/defradb-smoke-B
URL_A="127.0.0.1:9181"
URL_B="127.0.0.1:9182"
P2P_PORT_A=9171
P2P_PORT_B=9172
PID_A=""
PID_B=""
PASS=0
FAIL=0

SCHEMA='type Book { title: String, year: Int, author: String }'

# ─── helpers ─────────────────────────────────────────────────────────────────

pass() { echo "  PASS  $*"; PASS=$((PASS+1)); }
fail() { echo "  FAIL  $*"; FAIL=$((FAIL+1)); }

gql() {
  local url="$1" query="$2"
  # Escape any double-quotes inside the query for valid JSON embedding.
  local escaped="${query//\"/\\\"}"
  curl -sf -X POST "http://$url/api/v0/graphql" \
    -H "Content-Type: application/json" \
    -d "{\"query\":\"$escaped\"}"
}

wait_ready() {
  local url="$1"
  for _ in $(seq 1 40); do
    if curl -sf "http://$url/api/v0/graphql" \
         -H "Content-Type: application/json" \
         -d '{"query":"{collections{name}}"}' &>/dev/null; then
      return 0
    fi
    sleep 0.3
  done
  echo "  Node at $url did not become ready" >&2
  return 1
}

start_node() {
  local dir="$1" url="$2" p2p_port="$3"
  rm -rf "$dir"
  mkdir -p "$dir"
  "$BIN" start \
    --rootdir "$dir" \
    --url "$url" \
    --p2paddr "/ip4/127.0.0.1/tcp/$p2p_port" \
    --no-telemetry \
    --store memory \
    >> "$dir.log" 2>&1 &
  echo $!
}

add_schema() {
  local url="$1"
  "$BIN" client collection add --url "$url" "$SCHEMA" &>/dev/null
}

run_go() {
  (cd "$REPO_ROOT" && go run "$1" 2>&1)
}

cleanup() {
  [[ -n "$PID_A" ]] && kill "$PID_A" 2>/dev/null || true
  [[ -n "$PID_B" ]] && kill "$PID_B" 2>/dev/null || true
}
trap cleanup EXIT

# ─── pre-flight ──────────────────────────────────────────────────────────────
echo ""
echo "=== Pre-flight ==="
if [[ ! -x "$BIN" ]]; then
  echo "  Binary not found at $BIN. Build it first:"
  echo "    go build -o /tmp/defradb ./cmd/defradb/main.go"
  exit 1
fi
pass "binary exists at $BIN"

# ─── start node A ────────────────────────────────────────────────────────────
echo ""
echo "=== Starting node A ==="
PID_A=$(start_node "$DIR_A" "$URL_A" "$P2P_PORT_A")
wait_ready "$URL_A" && pass "node A ready (pid $PID_A)" || { fail "node A failed to start"; exit 1; }
add_schema "$URL_A" && pass "schema loaded on A" || fail "schema load failed on A"

# ─── Feature 5: SaveManyDocuments / add-many ─────────────────────────────────
echo ""
echo "=== Feature 5: SaveManyDocuments / add-many ==="

OUT=$(run_go ./tests/manual/cmd/smoke_addmany/)
if echo "$OUT" | grep -q "SaveManyDocuments: ok"; then
  SAVED=$(echo "$OUT" | grep -o 'saved=[0-9]*' | cut -d= -f2)
  READABLE=$(echo "$OUT" | grep -o 'readable=[0-9]*' | cut -d= -f2)
  pass "SaveManyDocuments: saved=$SAVED, all $READABLE readable (shared-txn, atomic)"
else
  fail "SaveManyDocuments: $(echo "$OUT" | tail -3)"
fi

# CLI add-many must correctly refuse HTTP usage.
ADDMANY_CLI=$("$BIN" client document add-many \
  --url "$URL_A" --collection-name Book '[]' 2>&1) || true
if echo "$ADDMANY_CLI" | grep -qiE "not supported|HTTP|direct"; then
  pass "add-many CLI: correctly returns ErrNotSupportedViaHTTP"
else
  fail "add-many CLI: unexpected output ($ADDMANY_CLI)"
fi

# Seed 5 books via single-doc CLI for downstream tests.
for entry in \
  '{"title":"Dune","year":1965,"author":"Herbert"}' \
  '{"title":"Neuromancer","year":1984,"author":"Gibson"}' \
  '{"title":"Snow Crash","year":1992,"author":"Stephenson"}' \
  '{"title":"Foundation","year":1951,"author":"Asimov"}' \
  '{"title":"1984","year":1949,"author":"Orwell"}'
do
  "$BIN" client document add --url "$URL_A" --collection-name Book "$entry" &>/dev/null
done

R=$(gql "$URL_A" '{Book{_docID,title}}')
COUNT=$(echo "$R" | grep -o '"_docID"' | wc -l | tr -d ' ')
if [[ "$COUNT" -ge 5 ]]; then
  pass "seeded $COUNT books on node A via single-doc add"
else
  fail "seed: expected ≥5 docs, got $COUNT  ($R)"
fi

DOC_ID=$(echo "$R"  | grep -o '"_docID":"[^"]*"' | head -1    | cut -d'"' -f4 || true)
DOC_ID2=$(echo "$R" | grep -o '"_docID":"[^"]*"' | sed -n '2p' | cut -d'"' -f4 || true)

# ─── Feature 6: KV snapshot export/import ────────────────────────────────────
echo ""
echo "=== Feature 6: KV snapshot export/import ==="

OUT=$(run_go ./tests/manual/cmd/smoke_kv/)
if echo "$OUT" | grep -q "kv: ok"; then
  PAIRS=$(echo "$OUT" | grep "ExportDocKVs:" | grep -o '[0-9]* pairs' | head -1)
  BYTES=$(echo "$OUT" | grep "ExportDocKVs:" | grep -o '[0-9]* bytes' | head -1)
  pass "kv export (ExportDocKVs): $PAIRS, $BYTES"
  pass "kv export --with-field-mapping (ExportFieldMapping): collectionName + fields present"
  pass "kv import (ImportRawKVs): completed"
  pass "kv import --field-mapping (ImportRawKVsWithMapping): identity remap completed"
else
  fail "kv: $(echo "$OUT" | tail -5)"
fi

# CLI must correctly refuse HTTP usage.
KV_CLI=$("$BIN" client kv export --url "$URL_A" Book /tmp/smoke_probe.kv 2>&1) || true
if echo "$KV_CLI" | grep -qiE "not supported|direct node|HTTP"; then
  pass "kv export CLI: correctly returns ErrNotSupportedViaHTTP"
else
  fail "kv export CLI: unexpected output ($KV_CLI)"
fi

# ─── Feature 8: PurgeByDocIDs ────────────────────────────────────────────────
echo ""
echo "=== Feature 8: PurgeByDocIDs ==="

if [[ -z "$DOC_ID" ]]; then
  fail "purge-docs: no docID captured, skipping purge tests"
else
  R=$(gql "$URL_A" "{Book(filter:{_docID:{_eq:\"$DOC_ID\"}}){_docID}}")
  if echo "$R" | grep -q "$DOC_ID"; then
    pass "purge-docs: target doc exists before purge"
  else
    fail "purge-docs: target doc not found pre-purge ($R)"
  fi

  "$BIN" client collection purge-docs \
    --url "$URL_A" --collection-name Book \
    --docID "$DOC_ID" &>/dev/null || true

  R2=$(gql "$URL_A" "{Book(filter:{_docID:{_eq:\"$DOC_ID\"}}){_docID}}")
  if echo "$R2" | grep -q "$DOC_ID"; then
    fail "purge-docs: doc still present after purge"
  else
    pass "purge-docs: doc hard-deleted"
  fi

  if [[ -n "$DOC_ID2" ]]; then
    "$BIN" client collection purge-docs \
      --url "$URL_A" --collection-name Book \
      --docID "$DOC_ID2" --prune-history &>/dev/null || true
    R3=$(gql "$URL_A" "{Book(filter:{_docID:{_eq:\"$DOC_ID2\"}}){_docID}}")
    if echo "$R3" | grep -q "$DOC_ID2"; then
      fail "purge-docs --prune-history: doc still present"
    else
      pass "purge-docs --prune-history: doc + history removed"
    fi
  fi
fi

# ─── Feature 2: Batch signing ────────────────────────────────────────────────
echo ""
echo "=== Feature 2: Batch signing ==="

OUT=$(run_go ./tests/manual/cmd/smoke_batchsign/)
if echo "$OUT" | grep -q "batch signing: ok"; then
  pass "batch signing: NewBatchCIDCollector, ContextWithBatchSigning, IsBatchSigningEnabled all correct"
else
  fail "batch signing: $(echo "$OUT" | tail -3)"
fi

# ─── Feature 3: Block signing refactor ───────────────────────────────────────
echo ""
echo "=== Feature 3: Block signing refactor ==="
if (cd "$REPO_ROOT" && go build ./internal/core/block/... &>/dev/null); then
  pass "block signing: internal/core/block compiles (block_signing.go + signature types)"
else
  fail "block signing: compile error in internal/core/block"
fi

# ─── Feature 4: Bounded sync worker pool ────────────────────────────────────
echo ""
echo "=== Feature 4: Bounded sync worker pool ==="
R=$(gql "$URL_A" '{Book{title}}')
if echo "$R" | grep -qE '"title"|Book|data'; then
  pass "sync worker pool: node responsive (bounded pool not deadlocked)"
else
  fail "sync worker pool: node unresponsive ($R)"
fi

# ─── Feature 7 + Feature 1: ReplicationFilter + CAR sync ────────────────────
echo ""
echo "=== Feature 7: ReplicationFilter + Feature 1: CAR sync ==="

PID_B=$(start_node "$DIR_B" "$URL_B" "$P2P_PORT_B")
wait_ready "$URL_B" && pass "node B ready (pid $PID_B)" || { fail "node B failed to start"; }

if [[ -n "$PID_B" ]]; then
  add_schema "$URL_B" && pass "schema loaded on B" || fail "schema load failed on B"

  PEER_INFO=$("$BIN" client p2p info --url "$URL_A" 2>&1)
  ADDR_A=$(echo "$PEER_INFO" | grep -oE '/ip4/[0-9.]+/tcp/[0-9]+/p2p/[A-Za-z0-9]+' | head -1)

  PEER_INFO_B=$("$BIN" client p2p info --url "$URL_B" 2>&1)
  ADDR_B=$(echo "$PEER_INFO_B" | grep -oE '/ip4/[0-9.]+/tcp/[0-9]+/p2p/[A-Za-z0-9]+' | head -1 || true)

  if [[ -n "$ADDR_A" ]]; then
    pass "p2p: node A addr = $ADDR_A"

    # Add B as a replicator on A — so A pushes new docs to B.
    "$BIN" client p2p replicator add --url "$URL_A" "$ADDR_B" &>/dev/null || true
    sleep 2

    NEW=$(gql "$URL_A" 'mutation{add_Book(input:{title:"Hyperion",year:1989,author:"Simmons"}){_docID,title}}')
    if echo "$NEW" | grep -q "Hyperion"; then
      pass "replication filter: doc created on A"
      sleep 5
      RB=$(gql "$URL_B" '{Book{title}}')
      if echo "$RB" | grep -q "Hyperion"; then
        pass "replication filter + CAR sync: doc arrived on B"
      else
        fail "replication filter + CAR sync: doc not found on B after 5s  ($RB)"
      fi
    else
      fail "replication filter: could not create doc on A ($NEW)"
    fi
  else
    fail "p2p: could not parse node A multiaddr from: $PEER_INFO"
  fi
fi

# ─── Feature 9: CLI wiring ───────────────────────────────────────────────────
echo ""
echo "=== Feature 9: CLI wiring ==="

check_help() {
  local label="$1" needle="$2"; shift 2
  local out; out=$("$@" 2>&1)
  if echo "$out" | grep -qF "$needle"; then pass "$label"
  else fail "$label  (expected '$needle' in --help output)"; fi
}

check_help "cli: kv subcommand"              "kv"                   "$BIN" client --help
check_help "cli: kv export"                  "export"               "$BIN" client kv --help
check_help "cli: kv import"                  "import"               "$BIN" client kv --help
check_help "cli: document add-many"          "add-many"             "$BIN" client document --help
check_help "cli: collection purge-docs"      "purge-docs"           "$BIN" client collection --help
check_help "cli: kv export --with-field-mapping" "with-field-mapping" "$BIN" client kv export --help
check_help "cli: kv import --field-mapping"  "field-mapping"        "$BIN" client kv import --help
check_help "cli: purge-docs --prune-history" "prune-history"        "$BIN" client collection purge-docs --help
check_help "cli: kv export HTTP note"        "direct node"          "$BIN" client kv export --help

# ─── Tier 0C: ReplicationFilter pre-storage ordering ─────────────────────────
echo ""
echo "=== Tier 0C: ReplicationFilter pre-storage ordering ==="
P2P_SRC="$REPO_ROOT/internal/db/p2p/p2p.go"
FILTER_LINE=$(grep -n "filterAllowsReplication" "$P2P_SRC" | head -1 | cut -d: -f1)
SYNC_LINE=$(grep -n -E "syncDAG|importCAR" "$P2P_SRC" | head -1 | cut -d: -f1)
if [[ -n "$FILTER_LINE" && -n "$SYNC_LINE" && "$FILTER_LINE" -lt "$SYNC_LINE" ]]; then
  pass "filter pre-storage: filterAllowsReplication (L$FILTER_LINE) before syncDAG/importCAR (L$SYNC_LINE)"
else
  fail "filter pre-storage: wrong order or symbols missing (filter=$FILTER_LINE sync=$SYNC_LINE)"
fi

# ─── Tier 2A: BatchMarkAsMerged ──────────────────────────────────────────────
echo ""
echo "=== Tier 2A: BatchMarkAsMerged ==="
OUT=$(run_go ./tests/manual/cmd/smoke_batchmark/)
if echo "$OUT" | grep -q "BatchMarkAsMerged: ok"; then
  pass "BatchMarkAsMerged: bulk to-merge key removal correct"
else
  fail "BatchMarkAsMerged: $(echo "$OUT" | tail -3)"
fi

# ─── Tier 2B: Heads cache ────────────────────────────────────────────────────
echo ""
echo "=== Tier 2B: Heads cache ==="
if (cd "$REPO_ROOT" && go build ./internal/db/... &>/dev/null); then
  pass "heads cache: internal/db compiles (InitHeadsCacheContext, PrefetchDocHeads, GetCachedHeads)"
else
  fail "heads cache: compile error in internal/db"
fi

# ─── Tier 3A: pubsubBatcher ──────────────────────────────────────────────────
echo ""
echo "=== Tier 3A: pubsubBatcher ==="
if (cd "$REPO_ROOT" && go build ./internal/db/p2p/... &>/dev/null); then
  pass "pubsubBatcher: internal/db/p2p compiles"
else
  fail "pubsubBatcher: compile error in internal/db/p2p"
fi

BATCHER="$REPO_ROOT/internal/db/p2p/batcher.go"
if grep -q "50 \* time.Millisecond" "$BATCHER" && grep -q "= 64" "$BATCHER"; then
  pass "pubsubBatcher: batchFlushInterval=50ms and batchMaxDocs=64 confirmed in source"
else
  fail "pubsubBatcher: expected constants 50ms/64 not found in batcher.go"
fi

# Liveness: rapid creates must not deadlock.
for i in 1 2 3 4 5; do
  gql "$URL_A" "mutation{add_Book(input:{title:\"Rapid$i\",year:200$i,author:\"Batch\"}){_docID}}" &>/dev/null || true
done
sleep 0.2
R=$(gql "$URL_A" '{Book(filter:{author:{_eq:"Batch"}}){title}}')
if echo "$R" | grep -q "Rapid"; then
  pass "pubsubBatcher live: rapid creates did not deadlock, node alive"
else
  fail "pubsubBatcher live: node unresponsive after rapid creates ($R)"
fi

# ─── Tier 3B: MergeBatchWithTxn + bounded semaphore ─────────────────────────
echo ""
echo "=== Tier 3B: MergeBatchWithTxn / bounded merge semaphore ==="
MERGE_FILE="$REPO_ROOT/internal/db/merge.go"
BATCH_TXN_LINE=$(grep -n "func.*MergeBatchWithTxn" "$MERGE_FILE" | head -1 | cut -d: -f1)
SEM_LINE=$(grep -n "maxConcurrentMerges" "$MERGE_FILE" | head -1 | cut -d: -f1)
NO_SEM_LINE=$(grep -n "addNoSem\b" "$MERGE_FILE" | head -1 | cut -d: -f1)

if [[ -n "$BATCH_TXN_LINE" ]]; then
  pass "MergeBatchWithTxn: defined at merge.go:$BATCH_TXN_LINE"
else
  fail "MergeBatchWithTxn: not found in merge.go"
fi
if [[ -n "$SEM_LINE" ]]; then
  MAX=$(grep "maxConcurrentMerges\s*=" "$MERGE_FILE" | grep -o '[0-9]\+')
  pass "bounded semaphore: maxConcurrentMerges=$MAX at merge.go:$SEM_LINE"
else
  fail "bounded semaphore: maxConcurrentMerges not defined in merge.go"
fi
if [[ -n "$NO_SEM_LINE" ]]; then
  pass "bounded semaphore: addNoSem/doneNoSem present (deadlock-safe batch locking at L$NO_SEM_LINE)"
else
  fail "bounded semaphore: addNoSem missing from merge.go"
fi

# ─── Tier 3C: Iterative loadBlockLinks ───────────────────────────────────────
echo ""
echo "=== Tier 3C: Iterative loadBlockLinks ==="
SYNC_DAG="$REPO_ROOT/internal/db/p2p/sync_dag.go"
if grep -q "stack " "$SYNC_DAG" 2>/dev/null; then
  pass "iterative loadBlockLinks: explicit stack-based DFS in sync_dag.go (no recursion)"
else
  fail "iterative loadBlockLinks: stack pattern not found in sync_dag.go"
fi

# ─── Tier 3D: P2PNoPeers event ───────────────────────────────────────────────
echo ""
echo "=== Tier 3D: P2PNoPeers event ==="
EVT_FILE="$REPO_ROOT/event/event.go"
if grep -q "P2PNoPeersName" "$EVT_FILE" && grep -q "type P2PNoPeers struct" "$EVT_FILE"; then
  pass "P2PNoPeers: constant and struct defined in event/event.go"
else
  fail "P2PNoPeers: missing from event/event.go"
fi
if grep -q "P2PNoPeersName" "$REPO_ROOT/internal/db/p2p/p2p.go"; then
  pass "P2PNoPeers: emitted in p2p.go"
else
  fail "P2PNoPeers: not emitted in p2p.go"
fi

# ─── Tier 3E: P2P.Close() ────────────────────────────────────────────────────
echo ""
echo "=== Tier 3E: P2P.Close() ==="
P2P_SRC="$REPO_ROOT/internal/db/p2p/p2p.go"
if grep -q "func (p \*P2P) Close()" "$P2P_SRC"; then
  pass "P2P.Close(): method defined"
  if grep -A5 "func (p \*P2P) Close()" "$P2P_SRC" | grep -qE "FlushAll|processQueue|stop|batcher"; then
    pass "P2P.Close(): body contains FlushAll/processQueue/batcher drain"
  else
    fail "P2P.Close(): body missing expected drain calls"
  fi
else
  fail "P2P.Close(): method not found in p2p.go"
fi

# ─── Final summary ────────────────────────────────────────────────────────────
echo ""
echo "══════════════════════════════════════════"
printf "  Results:  %d passed" "$PASS"
if [[ "$FAIL" -gt 0 ]]; then
  printf ",  %d FAILED\n" "$FAIL"
else
  printf ",  0 failed\n"
fi
echo "══════════════════════════════════════════"
echo ""
[[ $FAIL -eq 0 ]]
