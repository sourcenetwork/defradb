# Implementation Plan: Lens Migration with Indexes

**Status**: Implementation Complete - In Review

## Overview

This plan documents the implementation to ensure secondary indexes are properly updated when lens migrations are applied. The solution integrates reindexing logic at three key points: migration configuration, version switching, and collection patching.

## High-Level Approach

The implementation ensures indexes are rebuilt whenever:
1. A lens migration is added that affects the active collection version
2. The active version is switched and there's a migration with transform between versions
3. A collection is patched with an inline migration that becomes active

The core mechanic reuses the existing `reindexNewActiveVersion()` function from `internal/db/collection_index.go`, which:
- Removes all existing index entries for a collection
- Rebuilds indexes by iterating over all documents
- Applies lens transformations automatically via the fetcher layer

## Implementation Components

### 1. Migration Configuration (internal/db/lens.go:102-145)

**Location**: `setMigration()` function

**What Changed**:
- Added `shouldReindexAfterMigration()` helper function
- After successfully saving a migration, checks if reindexing is needed
- Triggers `reindexNewActiveVersion()` if the destination collection is active or in the active version's history chain

**Key Logic**:
```go
func (db *DB) setMigration(ctx context.Context, cfg client.LensConfig) (string, error) {
    // ... existing migration setup code ...

    // NEW: Check if reindexing is needed
    shouldReindex, activeCol, err := db.shouldReindexAfterMigration(ctx, dstCol)
    if err != nil {
        return "", err
    }

    // NEW: Trigger reindexing if necessary
    if shouldReindex {
        err = db.reindexNewActiveVersion(ctx, activeCol)
        if err != nil {
            return "", err
        }
    }

    return id.String(), nil
}
```

**Helper Function**:
```go
func (db *DB) shouldReindexAfterMigration(
    ctx context.Context,
    dstCol client.CollectionVersion,
) (bool, client.CollectionVersion, error) {
    // Direct activation case
    if dstCol.IsActive {
        return true, dstCol, nil
    }

    // Check if dstCol is in the history chain of the active version
    activeCol, err := description.GetActiveCollectionByCollectionID(ctx, dstCol.CollectionID)
    if err != nil {
        if errors.Is(err, corekv.ErrNotFound) {
            return false, client.CollectionVersion{}, nil
        }
        return false, client.CollectionVersion{}, err
    }

    colsWithRoot, err := description.GetCollectionsByCollectionID(ctx, dstCol.CollectionID)
    if err != nil {
        return false, client.CollectionVersion{}, err
    }

    isInChain := isMigrationInActiveChain(dstCol, activeCol, colsWithRoot)
    return isInChain, activeCol, nil
}
```

**Helper Function**:
```go
func isMigrationInActiveChain(
    dstCol, activeCol client.CollectionVersion,
    colsWithRoot []client.CollectionVersion,
) bool {
    versionsByID := make(map[string]client.CollectionVersion, len(colsWithRoot))
    for _, col := range colsWithRoot {
        versionsByID[col.VersionID] = col
    }

    current := activeCol

    for {
        if current.VersionID == dstCol.VersionID {
            return true
        }

        if !current.PreviousVersion.HasValue() {
            return false
        }

        prevSource := current.PreviousVersion.Value()
        prevVersion, exists := versionsByID[prevSource.SourceCollectionID]
        if !exists {
            return false
        }

        current = prevVersion
    }
}
```

### 2. Active Version Switching (internal/db/collection_define.go:531-543)

**Location**: `setActiveCollectionVersion()` function

**What Changed**:
- Enhanced existing `hasMigrationBetweenVersions()` logic
- Checks bidirectionally (both forward and backward in version chain)
- Triggers reindexing when switching between versions with migrations

**Key Logic**:
```go
func (db *DB) setActiveCollectionVersion(
    ctx context.Context,
    versionID string,
) error {
    // ... existing version switching code ...

    // Check if reindexing is needed
    if newActiveCol.HasValue() && prevActiveCol.HasValue() &&
        hasMigrationBetweenVersions(newActiveCol.Value(), prevActiveCol.Value(), colsWithRoot) {
        err = db.reindexNewActiveVersion(ctx, newActiveCol.Value())
        if err != nil {
            return err
        }
    }

    return db.loadSchema(ctx)
}
```

**Helper Function**:
```go
func hasMigrationBetweenVersions(
    activatedVersion, deactivatedVersion client.CollectionVersion,
    colsWithRoot []client.CollectionVersion,
) bool {
    versionsByID := make(map[string]client.CollectionVersion, len(colsWithRoot))
    for _, col := range colsWithRoot {
        versionsByID[col.VersionID] = col
    }

    // Check both directions since we don't know which is newer
    if hasMigrationInChain(activatedVersion, deactivatedVersion, versionsByID) {
        return true
    }

    return hasMigrationInChain(deactivatedVersion, activatedVersion, versionsByID)
}
```

**Helper Function**:
```go
func hasMigrationInChain(
    startVersion, targetVersion client.CollectionVersion,
    versionsByID map[string]client.CollectionVersion,
) bool {
    current := startVersion

    for {
        if !current.PreviousVersion.HasValue() {
            return false
        }

        prevSource := current.PreviousVersion.Value()

        // Check if this link has a migration (Transform)
        if prevSource.Transform.HasValue() {
            return true
        }

        prevVersion, exists := versionsByID[prevSource.SourceCollectionID]
        if !exists {
            return false
        }

        if prevVersion.VersionID == targetVersion.VersionID {
            return false
        }

        current = prevVersion
    }
}
```

### 3. Collection Patching (internal/db/collection_define.go:310-318)

**Location**: `patchCollection()` function

**What Changed**:
- Enhanced existing placeholder replacer logic
- Handles the case where unknown placeholder versions are materialized with migrations
- Triggers reindexing when placeholders are replaced with actual collections

**Key Logic**:
```go
func (db *DB) patchCollection(
    ctx context.Context,
    patchString string,
    migration immutable.Option[model.Lens],
) error {
    // ... existing patching logic ...

    // Track collections upgraded from placeholders
    var placeholderReplacers []client.CollectionVersion

    for i := 0; i < len(newCollections); i++ {
        placeholder := newCollections[i]
        if placeholder.IsPlaceholder {
            for j, col := range newCollections {
                if col.VersionID == placeholder.VersionID && !col.IsPlaceholder {
                    newCollections[j].PreviousVersion = placeholder.PreviousVersion
                    if col.IsActive {
                        placeholderReplacers = append(placeholderReplacers, newCollections[j])
                    }
                    isFound = true
                    break
                }
            }

            if isFound {
                // Remove placeholder
                newCollections = append(newCollections[:i], newCollections[i+1:]...)
                i--
            }
        }
    }

    // ... save collections and migrations ...

    // Reindex collections that were upgraded from placeholders with migrations
    for _, col := range placeholderReplacers {
        if col.PreviousVersion.HasValue() && col.PreviousVersion.Value().Transform.HasValue() {
            err = db.reindexNewActiveVersion(ctx, col)
            if err != nil {
                return err
            }
        }
    }

    return db.loadSchema(ctx)
}
```

### 4. Reindexing Function (internal/db/collection_index.go:696-718)

**Location**: `reindexNewActiveVersion()` function

**What Exists** (no changes needed):
- Removes all index entries via `RemoveAll()`
- Rebuilds indexes via `indexExistingDocs()`
- The lens transformation is automatically applied by the fetcher layer when iterating documents

**Key Implementation**:
```go
func (db *DB) reindexNewActiveVersion(ctx context.Context, col client.CollectionVersion) error {
    if !col.IsActive {
        return nil
    }

    collection, err := db.newCollection(col)
    if err != nil {
        return err
    }

    for _, colIndex := range collection.indexes {
        // Remove all existing index entries
        err = colIndex.RemoveAll(ctx)
        if err != nil {
            return err
        }

        // Rebuild index with current (possibly migrated) values
        err = collection.indexExistingDocs(ctx, colIndex)
        if err != nil {
            return err
        }
    }

    return nil
}
```

## Data Flow

### Scenario 1: Adding Migration to Active Version

```
User calls ConfigureMigration(v1 -> v2 with lens)
    ↓
setMigration() saves migration to lens store
    ↓
shouldReindexAfterMigration() checks if v2 is active
    ↓ (if v2 is active)
reindexNewActiveVersion(v2) is called
    ↓
For each index in v2:
    - Remove all index entries
    - Iterate all documents (lens applied by fetcher)
    - Save migrated values to index
    ↓
Complete
```

### Scenario 2: Switching Active Version with Migration

```
User calls SetActiveCollectionVersion(v2)
    ↓
setActiveCollectionVersion() activates v2, deactivates v1
    ↓
hasMigrationBetweenVersions(v2, v1) checks for migrations
    ↓ (if migration found)
reindexNewActiveVersion(v2) is called
    ↓
For each index in v2:
    - Remove all index entries
    - Iterate all documents (lens applied by fetcher)
    - Save migrated values to index
    ↓
Complete
```

### Scenario 3: Patching with Inline Migration

```
User calls PatchCollection(patch, lens)
    ↓
patchCollection() applies patch, creates v2
    ↓
Migration saved between v1 -> v2
    ↓
placeholderReplacers logic detects v2 is active with migration
    ↓
reindexNewActiveVersion(v2) is called
    ↓
For each index in v2:
    - Remove all index entries
    - Iterate all documents (lens applied by fetcher)
    - Save migrated values to index
    ↓
Complete
```

## Edge Cases Handled

1. **No Active Version**: If there's no active version, reindexing is skipped
2. **No Indexes**: If the collection has no indexes, reindexing is a no-op
3. **Distant Versions**: Walks the version chain to find migrations between non-adjacent versions
4. **Bidirectional Checking**: Handles both forward (v1->v5) and backward (v5->v1) version switches
5. **Placeholder Materialization**: Handles the case where migrations reference unknown versions
6. **Multiple Indexes**: Correctly reindexes all indexes on the collection

## Performance Considerations

1. **Conditional Reindexing**: Only reindexes when necessary (migration affects active version)
2. **No Redundant Work**: Checks prevent multiple reindexing operations in single transaction
3. **Batch Processing**: Uses efficient bulk operations for removing and rebuilding indexes
4. **Transactional**: All operations within existing transaction context

## Files Modified

1. `internal/db/lens.go`
   - Added `shouldReindexAfterMigration()` helper
   - Added `isMigrationInActiveChain()` helper
   - Modified `setMigration()` to trigger reindexing

2. `internal/db/collection_define.go`
   - Added `hasMigrationBetweenVersions()` helper
   - Added `hasMigrationInChain()` helper
   - Modified `setActiveCollectionVersion()` to trigger reindexing
   - Enhanced `patchCollection()` placeholder replacer logic

3. `internal/db/collection_index.go`
   - No changes needed, `reindexNewActiveVersion()` already exists

## Testing Strategy

All scenarios are covered by comprehensive integration tests in:
`tests/integration/collection_version/migrations/query/with_index_test.go`

Test coverage includes:
- Basic migration with existing index
- Setting active version forward and backward
- Creating index after migration
- Distant version switches
- Migration between old versions
- Placeholder materialization
- Performance optimization (no reindexing when not needed)

## Review Checklist

- [ ] All tests in `with_index_test.go` pass
- [ ] No performance regression in non-migration scenarios
- [ ] Edge cases properly handled (no active version, no indexes, etc.)
- [ ] Code follows DefraDB conventions and patterns
- [ ] Error handling is comprehensive
- [ ] Helper functions are properly documented
- [ ] Transaction semantics are correct
