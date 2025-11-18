# Decision Journal

## Decision 1: Collection Sync Topic vs. Reusing Existing Topics

**Date**: 2025-11-14

**Context:**
Need to decide whether to create a new pubsub topic `collection-sync` or reuse `doc-sync` for branchable collection synchronization.

**Decision:** Create dedicated `collection-sync` topic

**Rationale:**
- Separation of concerns: documents vs. collections
- Different message formats (collectionName vs. docIDs)
- Easier to debug and monitor
- Follows DefraDB's pattern of topic-per-entity-type
- Allows different handlers with specific logic

**Alternatives Considered:**
- Reuse `doc-sync`: Would require message type discrimination, more complex
- Use schema topic: Wrong semantic - we're syncing commits, not schema

**Outcome:** Clean separation, easy to understand and maintain

---

## Decision 2: Context Management for ID Cache

**Date:** 2025-11-14

**Context:**
`GetShortCollectionID` requires a context with the collection ID cache initialized, but `p.ctx` doesn't have it in the message handler.

**Problem Encountered:**
```
panic: interface conversion: interface {} is nil, not id.collectionShortIDCache
```

**Decision:** Initialize ID cache explicitly in message handler

**Implementation:**
```go
txnCtx := dbid.InitCollectionShortIDCache(p.ctx)
txnCtx = datastore.CtxSetTxn(txnCtx, txn)
shortID, err := dbid.GetShortCollectionID(txnCtx, col.CollectionID)
```

**Rationale:**
- Message handlers run in goroutines without parent context initialization
- ID cache is per-transaction, not global
- Explicit initialization is safer than assuming context state

**Lesson Learned:** Always initialize caches when working in async handlers

---

## Decision 3: C Bindings Stub vs. Full Implementation

**Date:** 2025-11-14

**Context:**
C bindings require building a shared library to generate the C header file with function declarations. Without the header, cgo fails to compile.

**Decision:** Implement C export but stub the wrapper temporarily

**Implementation:**
- `cbindings/p2p.go`: Full `P2PbranchableCollectionSync` export ✅
- `cbindings/wrapper.go`: Stubbed with error message pending header regeneration

**Rationale:**
- Allows other clients (Go, HTTP, CLI) to build and test
- C bindings can be completed after building shared library
- Maintains forward progress without blocking on C build process
- Clear TODO comment documents what's needed

**Alternatives Considered:**
- Block feature until C build: Would delay testing
- Skip C bindings entirely: Incomplete feature
- Build C library immediately: Out of scope for initial implementation

**Outcome:** Feature is 95% complete, C bindings can be finalized separately

---

## Decision 4: Test Setup - Schema on Both Nodes vs. FetchCollections

**Date:** 2025-11-14

**Context:**
Integration tests needed to verify SyncBranchableCollection works correctly. Initial approach tried using `FetchCollections` to get schema on node1.

**Problem:** `IsBranchable` property wasn't being synced with the collection version

**Decision:** Add schema to both nodes explicitly

**Rationale:**
- `SyncBranchableCollection` syncs the DAG (commits), not the schema definition
- Schema must already exist on both nodes
- Simpler test setup, clearer semantics
- Matches real-world usage: schema deployed first, then sync history

**Original Approach:**
```go
// Node 0: Add schema
// Node 1: FetchCollections (gets schema)
// Node 1: SyncBranchableCollection (sync commits)
```

**Final Approach:**
```go
// Node 0: Add schema
// Node 1: Add schema (same schema)
// Node 1: SyncBranchableCollection (sync commits)
```

**Root Cause Analysis:**
- `IsBranchable` is not in `CollectionDefinitionDelta`
- It's an immutable property set during initial schema creation
- `FetchCollections` doesn't sync the `@branchable` directive semantics
- For P2P sync to work, schemas must match on both nodes

**Outcome:** Tests pass, clear separation between schema sync and commit sync

---

## Decision 5: Commit Count Expectations in Tests

**Date:** 2025-11-14

**Context:**
Needed to determine expected number of commits for branchable collection tests.

**Findings:**
- 1 document with 1 field = 3 commits (collection composite + doc composite + field)
- 2 documents with 2 fields each = 8 commits

**Decision:** Use `testUtils.NewUniqueValue()` for flexible matching

**Rationale:**
- Exact CIDs don't matter, just that commits exist
- Count validation ensures sync completeness
- Follows existing branchable collection test patterns

**Implementation:**
```go
uniqueCid := testUtils.NewUniqueValue()
Results: []map[string]any{
    {"cid": uniqueCid},
    {"cid": uniqueCid},
    // ... one per expected commit
}
```
