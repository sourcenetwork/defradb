# Fix: Reindex secondary indexes when lens migrations are applied

This PR ensures secondary indexes are properly updated when lens migrations are applied to collections. Previously, indexes retained old unmigrated values, causing queries to return incorrect results when filtering on migrated fields.

## Changes

The implementation adds reindexing logic at three key integration points:

**1. Migration Configuration** (`internal/db/lens.go`)

When setting a lens migration via `setMigration()`, the system now checks if the destination collection version is active or in the active version's history chain. If so, indexes are automatically rebuilt with migrated values.

Added helper functions:
- `shouldReindexAfterMigration()` - determines if reindexing is needed after adding a migration
- `isMigrationInActiveChain()` - walks version chain to check if a version is in the active version's history

**2. Active Version Switching** (`internal/db/collection_define.go`)

When switching active collection versions via `setActiveCollectionVersion()`, the system now checks bidirectionally for migrations between the old and new active versions. If a migration exists, indexes are rebuilt to reflect the newly active version's values.

Added helper functions:
- `hasMigrationBetweenVersions()` - checks both directions for migrations between versions
- `hasMigrationInChain()` - walks backward through version chain looking for migrations

**3. Collection Patching** (`internal/db/collection_define.go`)

Enhanced the existing placeholder replacer logic in `patchCollection()` to detect when placeholder versions are materialized with migrations. When this occurs and the materialized version becomes active, indexes are automatically rebuilt.

## How It Works

The solution leverages the existing `reindexNewActiveVersion()` function from `internal/db/collection_index.go`. This function removes all index entries and rebuilds them by iterating over documents. The lens transformations are automatically applied by the fetcher layer during iteration, ensuring index values match query results.

## Test Coverage

Comprehensive integration tests in `tests/integration/collection_version/migrations/query/with_index_test.go` validate all scenarios:

- Indexes use migrated values when querying with migrations
- Setting active version backward uses original unmigrated values
- Setting active version forward uses migrated values
- Creating indexes after migration uses migrated values
- Distant version switches correctly apply intermediate migrations
- Reindexing only occurs when necessary (performance optimization)
- Patching with inline migrations triggers reindexing
- Adding migrations between old versions reindexes active version

All tests pass with the current implementation.

## Edge Cases

The implementation handles several important edge cases:

- No active version exists (reindexing skipped)
- Collection has no indexes (reindexing is no-op)
- Switching between non-adjacent versions with migrations in between
- Placeholder version materialization with migrations
- Multiple indexes on the same collection
- Bidirectional version switches (forward and backward in version chain)

## Performance

Reindexing only occurs when necessary - specifically when a migration affects the currently active collection version. This avoids unnecessary work in scenarios like adding migrations between old inactive versions or switching between versions with no migrations.

## Breaking Changes

None. This is a bug fix that makes the system behave as expected. Queries that previously returned incorrect results will now return correct results.

## Related Issues

Fixes the issue where migrations do not affect indexes, requiring proper index updates when schema versions change.
