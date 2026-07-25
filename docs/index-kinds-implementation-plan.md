# Implementation plan: index kinds, BM25, and trigram

Branch: `jsimnz/x/index-v2`
Companion to `docs/index-kinds-design.md`.

## Agreed query surface

**BM25** — a functional field, mirroring the existing `_similarity` precedent:

```graphql
query {
  Article(order: {_alias: {rank: DESC}}, limit: 10) {
    title
    rank: _bm25(body: {query: "database indexing"})
  }
}
```

`_bm25` requires a BM25 index on the target field. Without one the query errors.
BM25 needs collection-wide statistics (document frequency, average field length,
document count), so unlike `_similarity` it is not computable per document.

**Trigram** — no new query syntax for substring search. `_like` and `_ilike` are
accelerated when a trigram index exists. One new operator is added for regular
expressions:

```graphql
query {
  User(filter: {name: {_regex: "^Jo.*n$"}}) { name }
}
```

`_regex` works on any string field without an index (full scan, like `_like` today) and
is accelerated when a trigram index exists.

Negated forms (`_nlike`, `_nilike`, and any `_not` wrapper) cannot use a trigram index —
a candidate-generator index cannot produce "everything that does not match". This matches
how `_not` and `_none` are already excluded (`internal/db/fetcher/indexer_iterators.go:814-823`).

**Directive** — one `@index` directive gains a kind and generic options:

```graphql
type Article {
  title: String @index(kind: TRIGRAM)
  body:  String @index(kind: BM25, options: {k1: 1.2, b: 0.75})
}
```

## Design decisions taken up front

### Trigram extraction

Lowercase the value, then take a 3-byte sliding window (stride 1, overlapping). Bytes,
not runes — matching is rechecked afterwards, so byte-level candidates are safe. No
padding.

Consequences, to be documented rather than worked around:

- A value shorter than 3 bytes produces no trigrams and cannot be found through the index.
- A query literal shorter than 3 bytes yields no usable trigrams, so the query falls back
  to a full scan.
- Indexing lowercase means one index serves both `_like` and `_ilike`. The case-sensitive
  form over-returns and `filteredFetcher` rechecks exact case. This is the existing safety
  net doing real work, not a shortcut.

### The regexp-to-trigram conversion

Follow `index/regexp.go` from `github.com/google/codesearch` (simplified). Compute
`{canEmpty, exact, prefix, suffix, match}` over the parsed `regexp/syntax` tree and return
a boolean query of AND/OR over trigrams, or "match everything" when nothing useful can be
extracted.

Two properties must be preserved exactly, because the algorithm's correctness rests on them:

- A prefix/suffix/exact set is only usable when **every** member is at least 3 bytes. The
  empty string in a set therefore neutralises it, which is how "I know nothing" propagates.
- Set-size bounds (`maxExact = 7`, `maxSet = 20`) are a trust boundary — regular
  expressions come from user queries and an unbounded cross product is a denial-of-service
  risk.

Always recheck with `regexp` afterwards. The trigram query is a filter, never an answer.

### BM25 scoring

```
score(d) = Σ_t idf(t) · tf(t,d)·(k1+1) / ( tf(t,d) + k1·(1 - b + b·dl(d)/avgdl) )
idf(t)   = ln( 1 + (N - df(t) + 0.5) / (df(t) + 0.5) )
```

Defaults `k1 = 1.2`, `b = 0.75`, overridable through `options`.

The `1 +` inside the IDF logarithm is required, not cosmetic: without it IDF turns negative
for terms appearing in more than half the documents, and documents get penalised for
matching. Lucene, bleve, and every other implementation checked use this same expression.

Statistics kept per index: document count `N`, summed field length (for `avgdl`),
per-document field length `dl`, and per-term document frequency `df`. Only `tf` must be
exact. The rest are ranking inputs, not filters — staleness degrades ordering, never
correctness. Scores are comparable within one query, not across queries.

### Lazy top-K retrieval

WAND and MaxScore need K fixed up front, so they cannot back a pull iterator. Use a k-way
merge over docID-sorted posting lists, scoring each candidate, pushed through a
`container/heap`, then drained in descending score order. The first `Next()` pays the
traversal; every later one is O(log n).

This satisfies the lazy-pull requirement established earlier: `limitNode` stops calling
`Next()` once it has enough surviving rows, and access-control rejections are absorbed
below it, so top-K stays correct without any over-fetch factor.

Impact-ordered postings are the known upgrade path and are explicitly out of scope.

## Phases

Each phase is one or more commits and ends with a review pass.

### Phase 1 — Index kinds in the description types

Changes the index system itself. Every later phase only adds to it.

- `client.IndexDescription` gains `Kind string` and `Options map[string]any`. The empty
  kind means the ordered key index that exists today, so persisted `CollectionVersion`
  records deserialize unchanged and no migration is needed.
- Kind constants in `client`: `IndexKindBM25 = "bm25"`, `IndexKindTrigram = "trigram"`.
- A compile-time registry in `internal/db` mapping kind to constructor.
  `wrapCollectionIndex` (`internal/db/index.go:145-151`) becomes a registry lookup, with
  today's ordered index registered under the empty kind.
- `ErrUnknownIndexKind`. Collection versions are shared between peers, so a node running
  older code can receive a version naming a kind it does not implement. It must record
  that index as failed and keep going, never fail to open the collection.
- SDL: `kind` enum argument and `options` JSON argument on `@index`
  (`internal/request/graphql/schema/types/types.go:197`), parsed in
  `indexFromAST` (`internal/request/graphql/schema/collection.go:262-345`).
- Per-kind validation of options and of the field kind being indexed, run at index
  creation.

Done when the existing test suite passes unchanged and an index can be declared with an
explicit kind that has no implementation yet (and fails cleanly).

### Phase 2 — Trigram: storage, write path, and the `_regex` operator

- `internal/db/index_trigram.go` implementing `client.CollectionIndex`.
  Keys: `<index prefix>/<trigram>/<doc short ID>`, empty value. One key per posting rather
  than one packed list per trigram, so inserts and deletes stay point writes and never
  read-modify-write.
- Trigram extraction with unit tests covering short values, unicode, and duplicates.
- `_regex` as an ordinary connor operator (`internal/connor/regex.go`), working with no
  index at all. Compiled with `regexp` (RE2 — no catastrophic backtracking).
- `_regex` added to the String operator blocks in
  `internal/request/graphql/schema/types/base.go`.

Done when documents write trigram postings, and `_regex` filters correctly with no index
present.

### Phase 3 — Trigram: query conversion, iterator, planner integration, tests

- Regexp-to-trigram-query conversion, and the simpler literal-to-trigram conversion for
  `_like`/`_ilike`.
- A trigram index iterator producing candidate doc short IDs, using leapfrog intersection
  driven from the shortest posting list. `corekv.Iterator` has `Seek`, so this needs no
  new storage primitive.
- Planner: select a trigram index for `_like`, `_ilike`, and `_regex`; never for the
  negated forms. Fall back to a full scan when the query yields no usable trigrams.
- Integration tests under `tests/integration/index/`, following the existing file naming
  and `testUtils.ExecuteTestCase` style.

Done when substring and regex queries return identical results with and without the index,
and `EXPLAIN` shows the index was used.

### Phase 4 — BM25: storage and write path

- Tokenizer: lowercase, split on non-letter and non-digit, drop single-character tokens.
  One shared function used by both indexing and querying.
- `internal/db/index_bm25.go` implementing `client.CollectionIndex`. Key layout under the
  index prefix:
  - `t/<term>/<doc short ID>` → term frequency
  - `d/<doc short ID>` → field length
  - `s` → document count and summed field length
- Save, Update, and Delete maintaining all three consistently within the caller's
  transaction.

Document frequency gets no key of its own. It is the number of entries under `t/<term>/`,
which Phase 5 reads as a range while walking that term's documents anyway. A counter per term
would mean a read-modify-write for every distinct term of every document written, on keys the
most common terms make hot across every writer.

Done when writing, updating, and deleting documents leaves the statistics consistent,
covered by unit tests.

### Phase 5 — BM25: `_bm25` field, ranked scan, planner integration, tests

- `client/request/bm25.go` and the mapper field, mirroring `Similarity`
  (`internal/planner/mapper/mapper.go:979-990`).
- Schema generation for `_bm25`, mirroring `genSimilarityFieldConfig`
  (`internal/request/graphql/schema/generate.go:910-943`).
- Ranked iterator: k-way merge, score, heap, drain lazily.
- Threading the score into the virtual field slot. Default approach: extend
  `Fetcher.Init` with an optional rank specification — it is explicit and typed, and the
  other implementations only need pass-through edits. If that diff proves noisy, fall back
  to a small optional interface plus a setter on the wrapping fetcher.
- Planner: resolve the BM25 index from the `_bm25` field's target; error clearly when the
  field has no BM25 index.
- Optimisation: drop `orderNode` when the ordering is exactly the `_bm25` alias descending,
  so the scan is pulled lazily. Without this the results are still correct — `orderNode`
  simply re-sorts what the scan already ordered — but the whole index is drained. Extend
  `isOrderedByIndex` (`internal/planner/planner.go:423-443`), which today only understands
  ascending/descending over encoded field bytes.
- Integration tests, including ranking order, limit behaviour, and the no-index error.

Done when a `_bm25` query returns correctly ranked results, `limit` stops the scan early,
and access-control filtering does not reduce the returned count below the limit when
enough permitted matches exist.

## Explicitly out of scope

- Cost-based index selection. Issue #2680 stays open.
- `_nregex`. `_not: {field: {_regex: ...}}` covers it and cannot use the index anyway.
- Stemming and stop-word lists. IDF already de-weights common terms, and a stemmer is a
  dependency that silently invalidates the index if it ever changes.
- Impact-ordered postings and WAND-style pruning.
- Ranked indexes on the child side of a join. `typeIndexJoin` drives the child scan once
  per parent row, which is a materially harder problem. Root-level scans only.
- Phrase and proximity queries.
