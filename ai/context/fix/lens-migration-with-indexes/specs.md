# Specifications: Lens Migration with Indexes

**Status**: Review Phase - Implementation Complete

## Problem Statement

~~Currently~~ Previously, when lens migrations were applied to collections that have secondary indexes, the indexed values were not updated to reflect the migrated data. This caused queries using indexes to return incorrect results based on old, unmigrated values.

This issue has been addressed and is now in review phase.

## Background

DefraDB supports:
- **Lens Migrations**: Transformations applied when switching between collection versions (e.g., incrementing a field value)
- **Secondary Indexes**: Optimized data structures for fast field lookups

The problem occurs when these two features interact:
1. A collection has an index on field `age`
2. Documents are created with `age: 30`
3. A lens migration increments `age` by 5
4. Queries filter by `age: 35` (the migrated value)
5. **Expected**: Documents are found via index
6. **Actual**: Documents are NOT found because index still contains `age: 30`

## Requirements

### Functional Requirements

1. **FR-1: Migration Application Triggers Reindexing**
   - When a lens migration is set/configured, if it affects indexed fields, all affected indexes must be rebuilt
   - The reindexing must use migrated values, not original values

2. **FR-2: Active Version Switch Triggers Reindexing**
   - When switching active collection versions (forward or backward), if there's a migration with a transform between versions, indexes must be rebuilt
   - The indexes must reflect values for the newly active version

3. **FR-3: Patch with Migration Triggers Reindexing**
   - When patching a collection with a lens migration (inline), if the destination version becomes active, indexes must be rebuilt
   - This includes the scenario where unknown placeholder versions are materialized

4. **FR-4: Index Creation After Migration**
   - When creating a new index on a collection that has active migrations, the index must be populated with migrated values
   - This should work regardless of whether the migration was applied before or after index creation

5. **FR-5: Distant Version Switching**
   - When switching between non-adjacent versions (e.g., v1 → v5) with migrations in between, reindexing must correctly apply all intermediate migrations
   - The system must traverse the version chain to determine if reindexing is needed

### Non-Functional Requirements

1. **NFR-1: Performance**
   - Reindexing should only occur when necessary (when migrations affect the active version)
   - The system should avoid redundant reindexing operations

2. **NFR-2: Correctness**
   - Index values must always match the values returned by queries on the active version
   - No stale index data should exist after migration operations

3. **NFR-3: Transactional Consistency**
   - Reindexing operations must be transactional
   - Failed reindexing should roll back properly without leaving the system in an inconsistent state

## Acceptance Criteria

### AC-1: Basic Migration with Existing Index
- Given a collection with an indexed field and existing documents
- When a migration is applied that transforms the indexed field
- Then queries using the index return results based on migrated values
- And the explain plan confirms index usage

### AC-2: Setting Active Version Backward
- Given a collection with v1 (original) and v2 (with migration) where v2 is active
- When setting v1 as active
- Then indexes reflect original (unmigrated) values
- And queries return correct results for v1 data

### AC-3: Setting Active Version Forward
- Given a collection with v1 (original) and v2 (with migration) where v1 is active
- When setting v2 as active
- Then indexes reflect migrated values
- And queries return correct results for v2 data

### AC-4: Index Creation After Migration
- Given a collection with existing documents and an active migration
- When creating a new index on a migrated field
- Then the index is populated with migrated values
- And queries immediately work correctly with the new index

### AC-5: Patch with Inline Migration
- Given a collection with an indexed field
- When patching the collection with an inline lens migration
- Then indexes are automatically rebuilt with migrated values
- And subsequent queries use the updated indexes

### AC-6: Distant Version Switch with Migration
- Given a version chain v1 → v2 → v3 → v4 → v5 where v3→v4 has a migration
- When switching from v1 to v5 (or vice versa)
- Then reindexing occurs with correct migrated values
- And queries return expected results

### AC-7: No Reindexing Without Migration
- Given a version chain without migrations
- When switching between versions
- Then reindexing does NOT occur (performance optimization)
- And queries continue to work correctly

### AC-8: Migration Between Old Versions
- Given a version chain where v5 is active
- When adding a migration between v3→v4 (older versions)
- Then reindexing occurs because v5's chain includes the migration
- And queries reflect the new migration

## Test Coverage

All acceptance criteria are covered by integration tests in:
- `tests/integration/collection_version/migrations/query/with_index_test.go`

Tests that validate the implemented behavior:
- ✅ `TestSchemaMigrationQuery_WithIndexOnMigratedField_ShouldUseIndexWithMigratedValues`
- ✅ `TestSchemaMigrationQuery_WithIndexOnMigratedFieldAndSettingOldVersionAsActive_ShouldUseIndexWithOldValues`
- ✅ `TestSchemaMigrationQuery_WithIndexAppliedAfterMigration_ShouldIndexDocsOnLatestVersion`
- ✅ `TestSchemaMigrationQuery_WithIndexAppliedAfterSetActiveVersion_ShouldIndexDocsOnActiveVersion`
- ✅ `TestSchemaMigrationQuery_SwitchToOldDistantVersionWithNoMigrations_ShouldNotReindex`
- ✅ `TestSchemaMigrationQuery_SwitchToNewDistantVersionWithNoMigrations_ShouldNotReindex`
- ✅ `TestSchemaMigrationQuery_SwitchToOldDistantVersionWithMigrationInBetween_ShouldReindexWithOldValues`
- ✅ `TestSchemaMigrationQuery_SwitchToNewDistantVersionWithMigrationInBetween_ShouldReindexWithMigratedValues`
- ✅ `TestSchemaMigrationQuery_ApplyingMigrationBetweenOldVersions_ShouldReindex`
- ✅ `TestSchemaMigrationQuery_ApplyingMigrationBetweenNewVersions_ShouldNotReindex`
- ✅ `TestSchemaMigrationQuery_ApplyingMigrationToUnknownVersionsThenPatch_ShouldReindex`
- ✅ `TestSchemaMigrationQuery_ApplyingMigrationWithPatching_ShouldReindex`

These tests should all pass with the current implementation.

## Out of Scope

- Reindexing for encrypted indexes (searchable encryption)
- Reindexing for vector embeddings or similarity indexes
- Migration rollback/undo functionality
- Real-time incremental index updates during migration
- Multi-collection migration coordination

## Related Issues

This issue relates to the interaction between:
- Collection versioning system (`internal/db/collection_define.go`)
- Secondary indexing system (`internal/db/collection_index.go`)
- Lens migration system (`internal/db/lens.go`)
- Query planning with indexes (`internal/planner/`)

## Success Metrics

1. All tests in `with_index_test.go` pass
2. No performance regression in non-migration scenarios
3. Index rebuild time is proportional to document count
4. Zero stale index entries after migration operations
