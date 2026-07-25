# Supporting multiple index kinds

Status: draft
Branch: `jsimnz/x/index-v2`

## Problem

DefraDB has exactly one kind of secondary index: an ordered composite key over
order-preserving encoded scalar values, with an optional uniqueness constraint.
`client.IndexDescription` is `{Name, ID, Fields[], Unique}` and has no kind discriminator.

We want to add vector, BM25 full-text, bitmap, trigram, and geospatial indexes.

The repository already contains two features that behave like other index kinds, and
neither is built on the index system:

- Searchable encryption (`internal/se/`) is a wholly parallel vertical — its own
  directive, description type, `CollectionVersion` slice, API methods, and plan node. It
  shares no code with regular indexes.
- Vector similarity (`internal/planner/similarity.go`) has no index at all. It computes a
  score per already-fetched document and filters afterwards.

If we add five more kinds the way searchable encryption was added, we get five more
parallel verticals. This document proposes one path instead.

## The distinction that drives the design

The five requested kinds split into two groups, and the split is not about the data
structure. It is about what the index returns.

**Group A — candidate generators.** The index narrows the document set. It may return
extra documents. Correctness comes from re-checking.

- Trigram: trigram postings give candidates for `_like '%foo%'`, then the real pattern
  match confirms them.
- Geospatial bounding box: geohash or S2 cell prefixes cover the query region and
  overhang it, then exact geometry confirms.
- Bitmap: a bitmap per distinct value yields an exact set for equality, and a superset
  when combined loosely.

This group fits the current architecture well, because `filteredFetcher`
(`internal/db/fetcher/wrapper.go:143-179`) already re-evaluates the **entire original
filter** on every document the index produces. Over-returning is already safe today.

**Group B — ranked indexes.** The index returns documents in score order and the query
wants the best K.

- Vector approximate nearest neighbour.
- BM25 relevance.
- Geospatial `_near` (k-nearest).

This group breaks the current architecture in four places at once: there is no score
channel on results, no LIMIT pushdown, no way for an index to supply ordering, and access
control is applied *after* the scan.

That last one is the real problem. `permissionedFetcher`
(`internal/db/fetcher/permissioned.go:57-85`) drops documents the caller may not read,
after the index has already produced them. If a vector index returns the top 10 and access
control removes 6, the query returns 4. Not 10.

## Design

### 1. One kind discriminator, not five verticals

Add to `client.IndexDescription`:

```go
// Kind selects the index implementation. The empty string means the ordered
// key index that has always existed, so descriptions written before this
// field existed deserialize unchanged.
Kind string

// Options carries kind-specific configuration. The kind's constructor
// validates it; the client package does not know any kind's schema.
Options map[string]any
```

The empty-string default means no data migration and no change to existing persisted
`CollectionVersion` records.

A registry maps a kind name to a constructor:

```go
func RegisterIndexKind(name string, ctor IndexKindConstructor)
```

`wrapCollectionIndex` (`internal/db/index.go:145-151`) currently branches on `desc.Unique`.
It becomes a registry lookup, with the existing ordered index registered under the empty
name.

**Version skew matters here.** Index data is local, but collection versions are shared
across peers. A node running older code can receive a version describing an index kind it
does not implement. It must record that index as failed and continue, not fail to open the
collection. Add `ErrUnknownIndexKind` and handle it at description load.

### 2. Keep the write interface, add optional capabilities

`client.CollectionIndex` (Save / Update / Delete / Name / Description) stays as it is.
Every kind can implement it. Add two optional interfaces that the backfill discovers by
type assertion:

```go
// Implemented by kinds whose build is much cheaper in bulk — one embedding
// call for a batch of documents rather than one per document.
type BatchIndex interface {
    SaveBatch(context.Context, []*Document) error
}

// Implemented by kinds that need a global step once the last batch is done —
// building a term dictionary, computing collection statistics, constructing
// a graph.
type FinalizableIndex interface {
    Finalize(context.Context) error
}
```

Type assertion means no existing implementation changes and no kind is forced to
implement what it does not need. The backfill loop
(`internal/db/index_backfill.go:145-254`) gains two `if _, ok := idx.(...)` branches.

### 3. Two read shapes

Group A gets an interface for producing candidates. `indexFetcher` becomes one
implementation of it rather than the only option, and the two-way if/else in
`wrappingFetcher.Start` becomes a registry dispatch. The `filteredFetcher` above it keeps
guaranteeing correctness, so a new Group A kind cannot return wrong results — only slow
ones.

Group B needs a new plan node. The critical design decision:

**The ranked index interface must be a lazy iterator in descending score order, not a
`TopK(k)` call.**

```go
type RankedIndexIterator interface {
    // Next returns the next best match. Callers may pull as many as they need.
    Next(context.Context) (docShortID uint32, score float64, ok bool, err error)
}
```

One decision fixes three problems:

- **Access control.** The plan node pulls until it has K results that survive the
  permission check. K is preserved. A one-shot `TopK` cannot do this.
- **LIMIT.** The node stops pulling at the limit. No over-fetch guesswork.
- **Pagination.** Offset works by pulling and discarding.

HNSW supports continued search naturally. BM25 postings traversal does too. A brute-force
scan trivially does.

If a kind genuinely cannot resume, it can implement the iterator over an internally
doubling `TopK` — but that is the kind's problem, not the architecture's.

### 4. Planner

Four changes, in rough order of necessity:

1. **Ask the index, do not guess.** Today `selectIndex`
   (`internal/planner/index_helpers.go:184-230`) picks on field name, and operator
   compatibility is discovered much later in `shouldFallbackToFullScan`
   (`internal/db/fetcher/indexer_iterators.go:961-1027`) with no fallback to another
   index. Invert this: ask each candidate whether it can serve the condition, and take one
   that says yes.
2. **New operators.** `_near`, `_within`, `_search`. These need parser, request, and
   mapper support alongside the existing connor operator set.
3. **Ordering and LIMIT pushdown for ranked kinds.** `_near` and `_search` inherently
   combine filter, order, and limit. The planner must recognise this shape and hand K down.
4. **Multi-index intersection.** Only bitmaps really want this, and it is the largest
   change of the four. Defer it.

**No cost model.** Selection stays first-match, now operator-aware. Issue #2680 stays open.
Adding kinds does not require solving cost-based optimisation, and pretending otherwise
would stall this work behind a much larger project.

### 5. Storage

Better news than expected. The key prefix
`/[CollectionShortID]/[IndexID]/[Epoch]/...` already namespaces each index, and
`IndexID` comes from a per-collection sequence. Everything below that prefix is the
kind's to define. No change to the prefix layout is needed.

Kinds define their own layout underneath:

- Trigram and BM25: one key per (term, document), which avoids read-modify-write
  contention on hot terms. BM25 statistics (document count, average length, per-term
  document frequency) go in metadata keys under the same prefix.
- Bitmap: one key per distinct value with a roaring bitmap as the value. This *does* need
  read-modify-write, and hot values will contend.
- Vector: graph nodes or a flat vector array as keys under the prefix.
- Geospatial: cell identifiers are strings, so the existing order-preserving encoding
  already handles them. A bounding-box query becomes several prefix scans.

The existing `internal/encoding` codec is only needed by kinds that want ordered scalar
components. Others ignore it.

### 6. Lifecycle

The epoch mechanism, action-record state machine, watermark checkpointing, worker pool,
garbage collection, and crash recovery all generalise, because they operate on the key
prefix and the description — never on entry contents.

One genuine unsolved problem: the `building` flag
(`internal/db/index.go:68-90`) makes live writes and backfill converge by tolerating
missing deletes and same-document duplicates. That reasoning assumes index entries are
independent, idempotent, and keyed by document. It does not hold for a shared mutable
structure where a live write and the backfill would both mutate one graph.

Do not solve this generically. Per kind:

- Structures supporting concurrent insert (HNSW): let live writes through as normal. The
  index is already marked building and not used for queries.
- Structures needing a training or global phase (IVF, term dictionary): buffer live writes
  during the build and drain them in `Finalize`.

### 7. Schema

One directive, extended, rather than a directive per kind:

```graphql
@index(kind: VECTOR, options: {dimensions: 768, metric: COSINE})
```

Keeping one directive keeps one parse path, one identifier sequence, one `ListIndexes`
result, and one lifecycle. That is the specific thing the searchable-encryption vertical
gave up.

Field kinds need less work than expected: vectors reuse the existing
`FLOAT32_ARRAY`/`FLOAT64_ARRAY`/`INT_ARRAY` kinds, which is what
`similarityNode` already reads. Only geospatial needs a new field kind for points and
geometries.

## Sequencing

The order is chosen so each phase proves one seam and ships something usable.

**Phase 1 — seams only.** `Kind` field, registry, optional batch and finalize interfaces,
unknown-kind handling. Re-register the existing index under the empty kind. No new index
kinds, no behaviour change. This phase should be provable by the existing test suite
passing unchanged.

**Phase 2 — trigram.** The easiest Group A kind. Over-return plus the existing recheck
means almost no new machinery, and it ships real value: `_like '%foo%'` stops being a full
scan. Validates the candidate-generator path.

**Phase 3 — geospatial bounding box.** Second Group A kind. Adds a new field kind, a new
operator (`_within`), and multi-range scanning. Still over-return plus recheck, so still
no new plan-node shape.

**Phase 4 — the ranked path, proven with brute-force vector.** Build
`rankedIndexScanNode` with the lazy descending-score iterator, permission checks inside
the pull loop, and LIMIT pushdown. Back it with a brute-force vector scan reusing the
existing similarity maths.

This looks like a detour and is not. It makes `_near` a real planner concept with correct
top-K under access control, which is strictly better than today's post-fetch compute, and
it proves the hardest part of the architecture with zero index-build complexity.

**Phase 5 — HNSW and BM25.** Both drop in behind the Phase 4 iterator interface. HNSW
replaces the brute-force structure with no plan-node change. BM25 exercises
`Finalize` and collection-wide statistics.

**Phase 6 — bitmap.** Last, because its real value is multi-index intersection, which is
the largest planner change. A bitmap index serving a single predicate works before that
lands.

If vector is the business priority, note that Phase 4 already delivers correct,
index-selected, limit-pushed vector search. Phases 2 and 3 can be reordered after it; they
are placed first only because they are cheaper ways to prove the Group A seam.

## Deliberately not building

- A cost-based optimiser. Issue #2680 stays open.
- Multi-index intersection, until bitmaps need it.
- A dynamic plugin system. The registry is compile-time.
- A generic distributed index. Indexes stay local derived state, rebuilt per node from the
  replicated document graph.
- Encryption-aware variants of the new kinds. A content-aware index necessarily sees
  content; searchable encryption's HMAC equality trick does not extend to similarity or
  range search. Out of scope.

## Pre-existing issues this work touches

Independent of new index kinds, found while researching:

1. Searchable-encryption artifacts are never deleted. `secore.OperationDelete`
   (`internal/se/core/artifact.go:25-26`) is declared and used nowhere. Replicators keep
   queryable tags for deleted documents.
2. The searchable-encryption query path performs no per-document access-control check.
   `SelectEncrypted` (`internal/planner/select.go:591-620`) returns `seScanNode` as the
   entire plan, with no `permissionedFetcher`.
3. An encrypted field with a normal `@index` stores plaintext in the index of every node
   that merges the document. Nothing warns against the combination.
4. `cosineSimilarity` (`internal/planner/similarity.go:148-160`) computes a dot product
   with no magnitude normalisation. It equals cosine similarity only for pre-normalised
   vectors. Either the name or the implementation is wrong, and Phase 4 must settle which.
