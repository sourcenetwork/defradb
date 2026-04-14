# Index: `tests/integration/collection_version/migrations/query`

## Overview

This folder contains integration tests that verify the behaviour of lens-based schema migrations during query execution in DefraDB. The tests cover forward and inverse migrations across single and multiple schema versions, field mutations, copies, and removals, as well as interactions with indexes, transactions, P2P replication, node restarts, collection branching, and the setting of active collection versions.

## Test Index

### `simple_test.go`

Core migration query tests covering single-doc, multi-doc, field mutation, removal, and copy scenarios across one or more schema versions.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionMigrationQuery` | 26-86 | Migration transforms a single document when querying after schema version upgrade. |
| `TestCollectionMigrationQueryMultipleDocs` | 88-167 | Migration transforms all matching documents when querying with multiple docs. |
| `TestCollectionMigrationQueryWithMigrationRegisteredBeforePatchCollection` | 171-231 | Migration registered before schema patch still transforms documents correctly on query. |
| `TestCollectionMigrationQueryMigratesToIntermediaryVersion` | 233-304 | Partial migration chain applies up to the intermediary version only. |
| `TestCollectionMigrationQueryMigratesFromIntermediaryVersion` | 306-377 | Migration from intermediary version to latest applies correctly when querying. |
| `TestCollectionMigrationQueryMigratesAcrossMultipleVersions` | 379-467 | Chained migrations across multiple schema versions all apply during query. |
| `TestCollectionMigrationQueryMigratesAcrossMultipleVersionsBeforePatches` | 469-555 | Migrations registered before schema patches chain correctly when querying. |
| `TestCollectionMigrationQueryMigratesAcrossMultipleVersionsBeforePatchesWrongOrder` | 557-644 | Migrations registered in reverse order before patches still chain correctly on query. |
| `TestCollectionMigrationQueryWithUnknownCollectionMigration` | 652-712 | Orphan migration for unknown collection version does not block query execution. |
| `TestCollectionMigrationQueryMigrationMutatesExistingScalarField` | 714-775 | Migration overwrites an existing scalar field value on query. |
| `TestCollectionMigrationQueryMigrationMutatesExistingInlineArrayField` | 777-838 | Migration replaces an existing inline array field value on query. |
| `TestCollectionMigrationQueryMigrationRemovesExistingField` | 840-901 | Migration that removes a field causes that field to return nil on query. |
| `TestCollectionMigrationQueryMigrationPreservesExistingFieldWhenFieldNotRequested` | 903-979 | Migration preserves unrequested fields; they remain accessible in subsequent queries. |
| `TestCollectionMigrationQueryMigrationCopiesExistingFieldWhenSrcFieldNotRequested` | 981-1043 | Migration copies a field to a new field even when the source field is not requested. |
| `TestCollectionMigrationQueryMigrationCopiesExistingFieldWhenSrcAndDstFieldNotRequested` | 1045-1123 | Migration copies a field even when neither source nor destination is explicitly requested. |

### `with_collection_branch_test.go`

Tests migration query behaviour when the collection version history has branched into multiple active branches.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionMigrationQuery_WithBranchingCollection` | 25-111 | Querying on a branching collection applies the correct branch migration. |

### `with_doc_id_test.go`

Tests migration query behaviour when documents are fetched by a specific docID prefix.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionMigrationQueryByDocID` | 26-92 | Migration correctly transforms a single document fetched by docID prefix. |
| `TestCollectionMigrationQueryMultipleQueriesByDocID` | 106-280 | Lens pool correctly reuses instances across multiple docID-based queries with migration. |

### `with_filter_test.go`

Tests that query filters interact correctly with migrated field values across schema versions.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionMigrationQuery_WithFilter_ShouldFilterFMigration` | 26-109 | Filter applied after migration correctly matches the migrated field value. |
| `TestCollectionMigrationQuery_WithFilterAndMigrationBetweenOldVersions_ShouldApplyMigration` | 111-212 | Filter with migration between older non-adjacent versions correctly applies the transform. |
| `TestCollectionMigrationQuery_WithFilterAndMigrationInOldPatch_ShouldApplyMigration2` | 214-304 | Filter with migration embedded in an old patch correctly transforms filtered query results. |

### `with_index_test.go`

Tests covering how indexes interact with schema migrations, including reindexing behaviour when switching active versions or adding migrations between versions.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionMigrationQuery_WithIndexOnNotMigratedDocs_ShouldNotHinder` | 34-111 | Index on non-migrated docs still works correctly after migration is configured. |
| `TestCollectionMigrationQuery_WithIndexOnMigratedField_ShouldUseIndexWithMigratedValues` | 113-199 | Index on a migrated field returns results using the post-migration field values. |
| `TestCollectionMigrationQuery_WithIndexOnMigratedFieldAndSettingOldVersionAsActive_ShouldUseIndexWithOldValues` | 201-290 | Index uses original field values when the active collection version is reverted. |
| `TestCollectionMigrationQuery_WithIndexAppliedAfterMigration_ShouldIndexDocsOnLatestVersion` | 292-381 | Index created after migration is configured indexes documents at the latest version values. |
| `TestCollectionMigrationQuery_WithIndexAppliedAfterSetActiveVersion_ShouldIndexDocsOnActiveVersion` | 383-475 | Index created after setting active version indexes documents at the active version values. |
| `TestCollectionMigrationQuery_SwitchToOldDistantVersionWithNoMigrations_ShouldNotReindex` | 566-600 | Switching to a distant old version without migrations does not trigger reindexing. |
| `TestCollectionMigrationQuery_SwitchToNewDistantVersionWithNoMigrations_ShouldNotReindex` | 603-640 | Switching to a distant new version without migrations does not trigger reindexing. |
| `TestCollectionMigrationQuery_SwitchToOldDistantVersionWithMigrationInBetween_ShouldReindexWithOldValues` | 642-677 | Switching to an old version with an intermediate migration triggers reindex using old values. |
| `TestCollectionMigrationQuery_SwitchToNewDistantVersionWithMigrationInBetween_ShouldReindexWithMigratedValues` | 679-717 | Switching to a new distant version with an intermediate migration reindexes using migrated values. |
| `TestCollectionMigrationQuery_ApplyingMigrationBetweenOldVersions_ShouldReindex` | 719-757 | Adding a migration between old versions triggers reindexing with the migrated values. |
| `TestCollectionMigrationQuery_ApplyingMigrationBetweenNewVersions_ShouldNotReindex` | 760-795 | Adding a migration between newer versions beyond the active version does not reindex. |
| `TestCollectionMigrationQuery_ApplyingMigrationToUnknownVersionsThenPatch_ShouldReindex` | 797-885 | Migration registered for unknown versions triggers reindex when the patch links them. |
| `TestCollectionMigrationQuery_ApplyingMigrationWithPatching_ShouldReindex` | 887-967 | Migration included in a schema patch reindexes documents with the migrated values. |
| `TestCollectionMigrationQuery_WithBranchedVersionsAndMigration_ShouldApplyMigrationCorrectly` | 969-1125 | Switching between branched versions reindexes with the correct per-branch migration. |
| `TestCollectionMigrationQuery_WithThreeBranchedVersions_ShouldApplyCorrectMigrationPerBranch` | 1127-1332 | Three collection branches each apply their own distinct migration independently. |

### `with_inverse_test.go`

Tests that inverse (downward) migrations correctly undo field transformations when querying older schema versions.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionMigrationQueryInversesAcrossMultipleVersions` | 25-118 | Inverse migrations across multiple versions clear fields when querying an older version. |

### `with_p2p_collection_branch_test.go`

Tests migration behaviour for P2P replicated documents when nodes are on different collection branches.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionMigrationQueryWithP2PReplicatedDocOnOtherCollectionBranch` | 26-147 | P2P replicated doc on a different collection branch applies correct inverse then forward migration. |

### `with_p2p_test.go`

Tests migration query results for documents replicated over P2P between nodes at different schema versions.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionMigrationQueryWithP2PReplicatedDocAtOlderSchemaVersion` | 26-119 | P2P replicated doc at an older schema version is migrated up on the receiving node. |
| `TestCollectionMigrationQueryWithP2PReplicatedDocAtMuchOlderSchemaVersion` | 121-241 | P2P replicated doc multiple versions behind is chained-migrated up on the receiving node. |
| `TestCollectionMigrationQueryWithP2PReplicatedDocAtNewerSchemaVersion` | 243-339 | P2P replicated doc at a newer schema version is inverse-migrated down on the older node. |
| `TestCollectionMigrationQueryWithP2PReplicatedDocAtMuchNewerSchemaVersionWithSchemaHistoryGap` | 341-422 | P2P synced doc with a schema history gap is still accessible on the receiving node. |

### `with_restart_test.go`

Tests that configured migrations survive a node restart and continue to transform documents correctly.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionMigrationQueryWithRestart` | 25-86 | Migration persists across a node restart and transforms documents on query. |
| `TestCollectionMigrationQueryWithRestartAndMigrationBeforePatchCollection` | 88-149 | Migration registered before patch persists through restart and transforms documents correctly. |

### `with_set_default_test.go`

Tests migration behaviour when the active collection version is explicitly set to an older or newer version.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionMigrationQuery_WithSetDefaultToLatest_AppliesForwardMigration` | 27-88 | Setting the active version to the latest applies the forward migration on query. |
| `TestCollectionMigrationQuery_WithSetDefaultToOriginal_AppliesInverseMigration` | 90-165 | Reverting to the original collection version applies the inverse migration, clearing added fields. |
| `TestCollectionMigrationQuery_WithSetDefaultToOriginalVersionThatDocWasAddedAt_ClearsMigrations` | 167-241 | Reverting to the original doc version skips inverse migration and returns the original value. |

### `with_txn_test.go`

Tests that migrations configured within transactions are correctly scoped and visible to subsequent queries.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionMigrationQueryWithTxn` | 26-88 | Migration configured within a transaction applies correctly when querying in the same transaction. |
| `TestCollectionMigrationQueryWithTxnAndCommit` | 90-155 | Migration committed in a transaction is visible to a subsequent transaction query. |

### `with_update_test.go`

Tests migration behaviour during and after document update mutations.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionMigrationQueryWithUpdateRequest` | 25-105 | Migration runs during an update mutation and the result is persisted for subsequent queries. |
| `TestCollectionMigrationQueryWithMigrationRegisteredAfterUpdate` | 107-175 | Migration registered after a document update does not retroactively transform the updated doc. |
