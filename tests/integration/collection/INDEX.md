# Index: `tests/integration/collection`

## Overview

This directory contains integration tests for DefraDB collection management. The direct test file covers dynamic schema evolution: adding new collections via SDL and via JSON patch in a single request or across separate calls, and establishing one-to-many foreign-object relations between existing and newly created collections. The `truncate/` subdirectory provides exhaustive coverage of the `Truncate` operation across standard, branchable, DAC-protected, indexed, materialised-view, and concurrent scenarios.

## Test Index

### `add_relation_with_collection_test.go`

Tests that verify creating one-to-many foreign-object relations through `AddCollection` SDL and `PatchCollection` JSON patch operations across various batching and chaining configurations.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionUpdatesAddFieldKindForeignObject_WithAddCollectionCreatingOneToManyRelationToExistingCollection_ShouldSucceed` | 21-94 | Adding a collection with a one-to-many foreign-object relation to an existing collection succeeds and the relation is queryable. |
| `TestCollectionVersionUpdatesAddFieldKindForeignObject_WithAddCollectionCreatingOneToManyRelationsToMultipleExistingCollections_ShouldSucceed` | 96-177 | Adding a collection with one-to-many foreign-object relations to multiple existing collections succeeds and all relations are queryable. |
| `TestCollectionVersionUpdatesAddFieldKindForeignObject_WithPatchAddingOneToManyRelationAfterSeparateAddCollections_ShouldSucceed` | 179-258 | A patch that adds a one-to-many relation field between two separately added collections is resolved correctly and the relation is queryable. |
| `TestCollectionVersionUpdatesAddFieldKindForeignObject_WithMixedBatchHavingRelationToExistingAndNewCollections_ShouldSucceed` | 260-333 | A mixed add-collection batch where one collection references both an existing and a newly added collection creates all relations correctly. |
| `TestCollectionVersionUpdatesAddFieldKindForeignObject_WithChainedOneToManyRelationsAcrossSeparateCollections_ShouldSucceed` | 335-412 | Chained one-to-many relations across three separately added collections are resolved correctly, allowing a three-level nested query. |

## Subdirectories

| Directory | Summary |
|---|---|
| [`truncate/`](truncate/INDEX.md) | Integration tests for the `Truncate` collection operation, verifying that all documents, commit blocks, and index entries are removed and that the collection can be cleanly re-populated, including across branchable collections, DAC-protected docs, indexed fields, materialized views, and concurrent scenarios. |
