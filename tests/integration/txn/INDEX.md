# Index: `tests/integration/txn`

## Overview

This folder contains integration tests for DefraDB's transactional semantics. Tests verify that operations executed inside transactions — including schema changes (AddCollection, PatchCollection, SetActiveCollectionVersion), document mutations (AddDoc, UpdateDoc, DeleteDoc), index management (NewIndex, DeleteIndex, NewEncryptedIndex, DeleteEncryptedIndex), view and lens operations (AddView, AddLens, RefreshViews), migrations (SetMigration), collection truncation, and signature verification — are correctly isolated from concurrent transactions and only become visible after a commit. Each operation is exercised with at least a "with commit" and "without commit" variant, and many include an additional "transactional isolation" test that confirms operations within the same transaction see each other's in-flight changes.

## Test Index

### `add_collection_test.go`

Tests that adding a collection inside a transaction is only visible after the transaction commits.

| Test Function | Line | Description |
|---|---|---|
| `TestTxn_AddCollection_WithCommit_Succeeds` | 28-59 | Committing a transaction that added a collection makes the collection available. |
| `TestTxn_AddCollection_WithoutCommit_EmptyResults` | 63-91 | An uncommitted transaction that adds a collection leaves the database with no collections. |

### `add_doc_test.go`

Tests that adding a document inside a transaction is only visible after the transaction commits.

| Test Function | Line | Description |
|---|---|---|
| `TestTxn_AddDoc_WithCommit_Succeeds` | 26-72 | Committing a transaction that added a document makes the document queryable. |
| `TestTxn_AddDoc_WithoutCommit_EmptyResults` | 76-120 | An uncommitted transaction that adds a document leaves the collection empty to outside queries. |
| `TestTxn_AddDoc_InsideTxnWithAddCollection_WithCommit_Succeeds` | 124-171 | Committing a transaction containing both AddCollection and AddDoc persists both together. |

### `add_lens_test.go`

Tests that adding a lens inside a transaction is only usable after the transaction commits.

| Test Function | Line | Description |
|---|---|---|
| `TestTxn_AddLens_WithCommit_Succeeds` | 28-103 | Committing a transaction that added a lens makes the lens transform queryable through a view. |
| `TestTxn_AddLens_WithoutCommit_Fails` | 107-157 | An uncommitted transaction that adds a lens causes AddView to fail with lens not found. |

### `add_view_test.go`

Tests that adding a view inside a transaction is only queryable after the transaction commits.

| Test Function | Line | Description |
|---|---|---|
| `TestTxn_AddView_WithCommit_Succeeds` | 25-75 | Committing a transaction that added a view makes the view queryable. |
| `TestTxn_AddView_WithoutCommit_Fails` | 79-127 | An uncommitted transaction that adds a view leaves the view unavailable to queries. |

### `delete_doc_test.go`

Tests that deleting a document inside a transaction is only reflected after the transaction commits.

| Test Function | Line | Description |
|---|---|---|
| `TestTxn_DeleteDoc_WithCommit_Succeeds` | 25-68 | Committing a transaction that deleted a document removes it from the database. |
| `TestTxn_DeleteDoc_WithoutCommit_DoesNotDelete` | 72-123 | An uncommitted transaction that deletes a document leaves the document visible to outside queries. |
| `TestTxn_DeleteDoc_ExhibitsTransactionalIsolation_Succeeds` | 127-178 | A transaction can delete a document that was added within the same transaction. |

### `delete_encrypted_index_test.go`

Tests that deleting an encrypted index inside a transaction is only reflected after the transaction commits.

| Test Function | Line | Description |
|---|---|---|
| `TestTxn_DeleteEncryptedIndex_WithCommit_Succeeds` | 26-64 | Committing a transaction that deleted an encrypted index removes it from the collection. |
| `TestTxn_DeleteEncryptedIndex_WithoutCommit_DoesNotDelete` | 68-115 | An uncommitted transaction that deletes an encrypted index leaves the index visible outside. |
| `TestTxn_DeleteEncryptedIndex_ExhibitsTransactionalIsolation_Succeeds` | 119-163 | A transaction can delete an encrypted index that was created within the same transaction. |

### `delete_index_test.go`

Tests that deleting an index inside a transaction is only reflected after the transaction commits.

| Test Function | Line | Description |
|---|---|---|
| `TestTxn_DeleteIndex_WithCommit_Succeeds` | 26-51 | Committing a transaction that deleted an index removes the index from the collection. |
| `TestTxn_DeleteIndex_WithoutCommit_DoesNotDelete` | 55-94 | An uncommitted transaction that deletes an index leaves the index visible outside the transaction. |
| `TestTxn_DeleteIndex_ExhibitsTransactionalIsolation_Succeeds` | 98-129 | A transaction can delete an index that was added on a schema created within the same transaction. |

### `get_collections_test.go`

Tests that GetCollections within a transaction respects transactional isolation boundaries.

| Test Function | Line | Description |
|---|---|---|
| `TestTxn_GetCollections_InsideTxnWithAddSchema_Succeeds` | 27-56 | GetCollections inside a transaction sees the collection added within the same transaction. |
| `TestTxn_GetCollections_InsideTxnWithoutAddSchema_NoCollections` | 60-88 | GetCollections in a separate transaction does not see a collection added in another transaction. |

### `list_encrypted_indexes_test.go`

Tests that an encrypted index created and committed within a transaction is visible after the commit.

| Test Function | Line | Description |
|---|---|---|
| `TestTxn_AddEncryptedIndex_ExhibitsTransactionalIsolation_Succeeds` | 26-75 | An encrypted index added and committed within a transaction is visible after the commit. |

### `list_indexes_test.go`

Tests that ListIndexes within a transaction sees indexes added in the same transaction.

| Test Function | Line | Description |
|---|---|---|
| `TestTxn_ListIndexes_InsideTxn_Succeeds` | 26-73 | ListIndexes inside a transaction sees the index added within the same transaction. |

### `list_lenses_test.go`

Tests that ListLenses within a transaction respects transactional isolation boundaries.

| Test Function | Line | Description |
|---|---|---|
| `TestTxn_ListLenses_InsideTxnWithAddLens_Succeeds` | 27-79 | ListLenses inside a transaction sees the lens added within the same transaction. |
| `TestTxn_ListLenses_InsideTxnWithoutAddLens_NoLenses` | 83-124 | ListLenses in a separate transaction does not see a lens added in another transaction. |

### `new_encrypted_index_test.go`

Tests that creating an encrypted index inside a transaction is only visible after the transaction commits.

| Test Function | Line | Description |
|---|---|---|
| `TestTxn_NewEncryptedIndex_WithCommit_Succeeds` | 26-66 | Committing a transaction that created an encrypted index makes it visible to ListEncryptedIndexes. |
| `TestTxn_NewEncryptedIndex_WithoutCommit_NoIndexes` | 70-109 | An uncommitted transaction that creates an encrypted index leaves no index visible outside. |
| `TestTxn_ListEncryptedIndexes_InsideTxn_Succeeds` | 113-158 | ListEncryptedIndexes inside a transaction sees the encrypted index added in the same transaction. |

### `new_index_test.go`

Tests that creating an index inside a transaction is only visible after the transaction commits.

| Test Function | Line | Description |
|---|---|---|
| `TestTxn_NewIndex_WithCommit_Succeeds` | 26-70 | Committing a transaction that created a new index makes the index visible to ListIndexes. |
| `TestTxn_NewIndex_WithoutCommit_NoIndexes` | 74-112 | An uncommitted transaction that creates an index leaves no index visible outside the transaction. |
| `TestTxn_NewIndex_ExhibitsTransactionalIsolation_Succeeds` | 116-169 | A transaction can create an index on a collection and document added within the same transaction. |

### `patch_collection_test.go`

Tests that patching a collection schema inside a transaction is only applied after the transaction commits.

| Test Function | Line | Description |
|---|---|---|
| `TestTxn_PatchCollection_WithCommit_Succeeds` | 25-61 | Committing a transaction that patched a collection applies the field removal permanently. |
| `TestTxn_PatchCollection_WithoutCommit_PatchNotApplied` | 65-107 | An uncommitted transaction that patches a collection leaves the schema unchanged for outside queries. |

### `refresh_view_test.go`

Tests that refreshing a materialized view inside a transaction is only applied after the transaction commits.

| Test Function | Line | Description |
|---|---|---|
| `TestTxn_RefreshView_WithCommit_Succeeds` | 25-90 | Committing a transaction that refreshed a view includes documents added after the initial view creation. |
| `TestTxn_RefreshView_WithoutCommit_DoesNotRefresh` | 94-160 | An uncommitted transaction that refreshes a view leaves the view stale to outside queries. |

### `request_test.go`

Tests that a GraphQL mutation request inside a transaction is only persisted after the transaction commits.

| Test Function | Line | Description |
|---|---|---|
| `TestTxn_Request_WithCommit_Succeeds` | 25-83 | Committing a transaction that ran a mutation request makes the document queryable. |
| `TestTxn_Request_WithoutCommit_EmptyResults` | 87-143 | An uncommitted transaction that ran a mutation request leaves the collection empty to outside queries. |

### `set_active_collection_version_test.go`

Tests that setting the active collection version inside a transaction is only applied after the transaction commits.

| Test Function | Line | Description |
|---|---|---|
| `TestTxn_SetActiveCollectionVersion_WithCommit_Succeeds` | 25-62 | Committing a transaction that set the active collection version reverts the schema to that version. |
| `TestTxn_SetActiveCollectionVersion_WithoutCommit_VersionNotChanged` | 66-110 | An uncommitted transaction that set the active collection version leaves the schema unchanged. |

### `set_migration_test.go`

Tests that setting a schema migration inside a transaction does not apply until the transaction commits.

| Test Function | Line | Description |
|---|---|---|
| `TestTxn_SetMigration_WithoutCommit_Succeeds` | 31-99 | An uncommitted transaction that sets a migration leaves documents unmigrated to outside queries. |

### `truncate_test.go`

Tests that truncating a collection inside a transaction is only reflected after the transaction commits.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionTruncate_WithCommit_Succeeds` | 24-62 | Committing a transaction that truncated a collection removes all documents from the collection. |

### `update_doc_test.go`

Tests that updating documents inside a transaction is only reflected after the transaction commits.

| Test Function | Line | Description |
|---|---|---|
| `TestTxn_UpdateDoc_WithCommit_Succeeds` | 25-74 | Committing a transaction that updated a document persists the updated field value. |
| `TestTxn_UpdateDoc_WithoutCommit_DoesNotUpdateDocument` | 78-131 | An uncommitted transaction that updated a document leaves the original value visible outside. |
| `TestTxn_UpdateDoc_ExhibitsTransactionalIsolation_Succeeds` | 135-193 | A transaction can update a document that was added within the same transaction. |
| `TestTxn_UpdateDocWithFilter_WithCommit_Succeeds` | 197-245 | Committing a transaction that updated a document with a filter persists the new field value. |
| `TestTxn_UpdateDocWithFilter_WithoutCommit_DoesNotUpdateDocument` | 249-301 | An uncommitted filtered update transaction leaves the original document value visible outside. |
| `TestTxn_UpdateWithFilter_ExhibitsTransactionalIsolation_Succeeds` | 305-355 | A transaction can use a filter to update a document added within the same transaction. |

### `verify_signature_test.go`

Tests that block signature verification inside a transaction respects transactional isolation.

| Test Function | Line | Description |
|---|---|---|
| `TestTxn_VerifyBlockSignature_InsideTxn_Succeeds` | 25-64 | VerifyBlockSignature succeeds when verifying a block created within the same transaction. |
| `TestTxn_VerifyBlockSignature_OutsideTxn_Fails` | 68-108 | VerifyBlockSignature fails when the block exists only in a different uncommitted transaction. |
