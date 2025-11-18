# Learnings

This document captures genuinely new knowledge discovered during this task that is not documented elsewhere in the codebase.

## Purpose

Record insights about:
- Undocumented codebase patterns
- Hidden dependencies between systems
- Non-obvious behaviors
- Gotchas and pitfalls
- System quirks discovered

**Note**: This should NOT duplicate information from README files, existing docs, or code comments. Only record truly new discoveries.

---

## Insights

### 1. Placeholder Collection Versions

**Discovery**: DefraDB creates "placeholder" collection versions when migrations reference unknown schemas.

**Context**: When calling `ConfigureMigration` with source/destination version IDs that don't exist yet, the system creates placeholder versions with `IsPlaceholder: true`. These are later "materialized" when the actual schema is created via `PatchCollection`.

**Why This Matters**:
- Allows forward-declaring migrations before schemas exist
- Requires special handling during patching to detect placeholder materialization
- The `placeholderReplacers` logic in `patchCollection()` handles this case

**Code Location**: `internal/db/lens.go:59-81` and `internal/db/collection_define.go:242-268`

---

### 2. Lens Transformation in Fetcher Layer

**Discovery**: Lens transformations are automatically applied by the document fetcher, not explicitly in the reindexing logic.

**Context**: When iterating documents during reindexing via `indexExistingDocs()`, the fetcher layer automatically applies any active lens transformations based on the collection version. This means `reindexNewActiveVersion()` doesn't need to explicitly invoke lens logic.

**Why This Matters**:
- Simplifies reindexing implementation
- Ensures consistency between query results and index values
- Relies on proper collection version context being set

**Code Location**: `internal/db/collection_index.go:325-339` uses fetcher which handles lens application transparently

---

### 3. Bidirectional Version Chain Traversal

**Discovery**: When checking for migrations between versions, you must check BOTH directions (v1→v5 and v5→v1) because the system doesn't track which version is "newer".

**Context**: Version relationships are stored as `PreviousVersion` links, but when switching between versions, you don't know a priori which version is older. The `hasMigrationBetweenVersions()` function handles this by checking both directions.

**Why This Matters**:
- Critical for correct reindexing decisions
- Prevents missing migrations when switching backward in version history
- Requires maintaining a version map for efficient traversal

**Code Location**: `internal/db/collection_define.go:545-562`

---

### 4. Active Version History Chain

**Discovery**: A migration can affect the active version even if it's between two older versions.

**Scenario**:
```
v1 → v2 → v3 → v4 → v5 (active)
```
If you add a migration between v3→v4, reindexing IS needed for v5 because v5's data depends on that migration chain.

**Why This Matters**:
- Can't just check if the destination version is active
- Must traverse the entire version chain to the active version
- The `isMigrationInActiveChain()` function implements this check

**Code Location**: `internal/db/lens.go:148-180`

---

### 5. Index Creation Timing

**Discovery**: Creating an index after a migration automatically uses migrated values without special handling.

**Context**: When `CreateIndex` is called on a collection with an active migration, `indexExistingDocs()` is invoked, which uses the fetcher layer that automatically applies lens transformations.

**Why This Matters**:
- No special migration-aware code needed in index creation path
- Consistent behavior regardless of index creation timing
- Simplifies the implementation

**Code Location**: `internal/db/collection_index.go:248-263`

---

## Patterns Observed

### Immutable.Option Pattern

DefraDB extensively uses `immutable.Option[T]` from the immutable package for optional values rather than pointers or Go 1.18+ Option types.

**Usage Pattern**:
```go
if col.PreviousVersion.HasValue() {
    prev := col.PreviousVersion.Value()
    if prev.Transform.HasValue() {
        // migration exists
    }
}
```

### Transaction Context Pattern

DefraDB passes transactions via context rather than explicit parameters:
```go
txn := datastore.CtxMustGetTxn(ctx)
```

This allows nested function calls to share the same transaction without explicitly passing it.

---

## System Behaviors

### Schema Version IDs and Test Infrastructure

Schema version IDs are content-addressed hashes (CIDs) that uniquely identify a collection version. They appear in tests as constants like:
```go
const schemaV1 = "bafyreifnbhwntycylk2l6n4khiocdt3vks46tizjdaz6yx4tsmdjtdtlma"
```

**Critical Discovery**: These constants are schema-specific. If a test creates a schema with different structure than the constant represents, the version IDs won't match. This causes issues when:
- Using `ConfigureMigration` with hardcoded constants
- The actual schema created has different fields/structure
- Result: Migration is created for a placeholder version, not the actual collection

**Implication**: Tests must either:
1. Use correct constants matching their exact schema structure
2. Capture version IDs dynamically after creation
3. Ensure schema structures match the constants being used

**Impact**: This issue was hidden by the old linear chain approach which used in-memory collections and didn't validate against the database. The DAG approach exposed this because it properly checks version relationships.

### CollectionID vs VersionID

- **CollectionID**: Identifies a logical collection across all versions (stays the same)
- **VersionID**: Identifies a specific version of a collection (changes with each schema update)

This distinction is critical for version chain traversal and active version lookups.

---

### 6. DAG Traversal for Version Histories

**Discovery**: Collection version histories form a DAG (Directed Acyclic Graph), not just linear chains.

**Context**: The `internal/lens/history.go` package already has comprehensive DAG traversal logic:
- `getTargetedCollectionHistory()` builds a full DAG relative to a target version
- Uses bidirectional links (`next` and `previous` slices) to handle branches
- `linkForwards()` and `linkBackwards()` recursively build the graph

**Why This Matters**:
- Version histories can branch when switching active versions and patching
- A migration on one branch affects that branch but not siblings
- Proper reindexing requires checking ALL reachable versions, not just one path
- The lens package pattern should be reused for consistency

**Example Branching**:
```
       => v2 (branch A, no migration)
v1 =>
       => v3 (branch B, migration: age+5)
```
If v3 is active and we query, age should be migrated. If we switch to v2, age should be original.

**Code Location**: `internal/lens/history.go:57-144`

---

## Notes

- This document should remain focused and concise
- Avoid documenting things that can be learned from code reading
- Focus on non-obvious interactions and system behaviors
- Update if new discoveries emerge during review or production use
