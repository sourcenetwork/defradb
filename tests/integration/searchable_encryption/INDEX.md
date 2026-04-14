# Index: `tests/integration/searchable_encryption`

## Overview

This folder contains integration tests for DefraDB's searchable encryption (SE) feature, which allows equality lookups on encrypted document fields via an `@encryptedIndex` directive and corresponding `encrypted_<Collection>` GraphQL query types. The tests cover creating and deleting encrypted indexes (via schema or API), listing indexes per collection or globally, patch-operation immutability of indexes, P2P generation of the GQL query type, replication of SE metadata between nodes (including offline-retry behaviour), ACP-gated replication access, and the full lifecycle of querying encrypted fields with single or multiple filter conditions.

## Test Index

### `add_test.go`

Tests for creating encrypted indexes, verifying that schema-defined and dynamically-added indexes do not disrupt normal querying, and that error conditions (non-existent field, duplicate index) are correctly reported.

| Test Function | Line | Description |
|---|---|---|
| `TestEncryptedIndexNew_SchemaWithEncryptedIndex_ShouldNotHinderQuerying` | 22-63 | Adding an @encryptedIndex directive in the schema does not prevent normal document querying. |
| `TestEncryptedIndexNew_AfterAddRequest_ShouldNotHinderQuerying` | 65-109 | Adding an encrypted index after documents are already stored does not prevent normal querying. |
| `TestEncryptedIndexNew_IfNonExistentFieldIsGiven_ReturnError` | 111-131 | Creating an encrypted index on a field that does not exist in the collection returns an error. |
| `TestEncryptedIndexNew_IfIndexAlreadyExists_ShouldReturnError` | 133-153 | Attempting to create a duplicate encrypted index on the same field returns an already-exists error. |

### `delete_test.go`

Tests for deleting encrypted indexes, ensuring correct removal, error handling for missing indexes, re-creation after deletion, and selective deletion when multiple indexes exist.

| Test Function | Line | Description |
|---|---|---|
| `TestEncryptedIndexDelete_WithExistingIndex_ShouldDeleteSuccessfully` | 23-55 | Deleting an existing encrypted index removes it from the collection's index list. |
| `TestEncryptedIndexDelete_IfIndexDoesNotExist_ReturnError` | 57-77 | Deleting an encrypted index that does not exist returns a does-not-exist error. |
| `TestEncryptedIndexDelete_AfterDelete_CanMakeNewIndexAnew` | 79-114 | After deleting an encrypted index, a new encrypted index can be created on the same field. |
| `TestEncryptedIndexDelete_MultipleIndexes_ShouldOnlyDeleteSpecified` | 116-166 | Deleting one encrypted index from a collection with multiple indexes leaves the others intact. |

### `list_test.go`

Tests for listing encrypted indexes on a specific collection or across all collections, including indexes added after initial schema creation.

| Test Function | Line | Description |
|---|---|---|
| `TestEncryptedIndexList_ShouldReturnListOfExistingIndexes` | 22-64 | Listing encrypted indexes per collection returns the correct indexes for each collection. |
| `TestEncryptedIndexList_IfIndexAddedLater_ShouldReturnListOfExistingIndexes` | 66-107 | A dynamically added encrypted index appears in the collection's index list alongside schema-defined ones. |
| `TestEncryptedIndexList_WhenRequestingAllIndexes_ShouldReturn` | 109-149 | Listing all encrypted indexes across the database returns indexes grouped by collection name. |

### `patch_test.go`

Tests that encrypted indexes are immutable via JSON patch operations, covering add, remove, and replace operations.

| Test Function | Line | Description |
|---|---|---|
| `TestPatchCollection_NewEncryptedIndex_ShouldError` | 23-56 | Adding an encrypted index via a JSON patch operation on a collection returns an immutability error. |
| `TestPatchCollection_RemoveEncryptedIndex_ShouldError` | 59-86 | Removing an encrypted index via a JSON patch remove operation on a collection returns an immutability error. |
| `TestPatchCollection_ModifyEncryptedIndex_ShouldError` | 89-122 | Replacing an encrypted index via a JSON patch replace operation on a collection returns an immutability error. |

### `peer_acp_test.go`

Tests that ACP access control is enforced during P2P replication of encrypted documents.

| Test Function | Line | Description |
|---|---|---|
| `TestDocEncryptionPeer_WithACP_ReplicatorShouldNotHaveAccess` | 70-166 | A replicator without ACP access cannot read encrypted documents or their commit blocks. |

### `peer_add_test.go`

Tests that encrypted indexes generate the expected `encrypted_<Collection>` GraphQL query type in a P2P-enabled setup, for both schema-defined and dynamically-added indexes.

| Test Function | Line | Description |
|---|---|---|
| `TestEncryptedIndexNewPeer_SchemaWithEncryptedIndex_ShouldGenerateGQL` | 25-58 | An @encryptedIndex schema directive generates the encrypted GQL query type in a P2P setup. |
| `TestEncryptedIndexNewPeer_AfterAddRequest_ShouldGenerateGQL` | 60-96 | Adding an encrypted index after collection creation generates the encrypted GQL query type in a P2P setup. |

### `peer_delete_test.go`

Tests that deleting an encrypted index removes the corresponding GQL query type from the P2P-enabled schema.

| Test Function | Line | Description |
|---|---|---|
| `TestDocEncryptionPeer_AfterDeletingIndex_SEQueryShouldReturnError` | 24-74 | After deleting an encrypted index, the corresponding searchable encryption GQL query type is removed. |

### `peer_query_test.go`

Tests for executing searchable encryption queries across P2P peers, covering single-field lookups, multiple encrypted fields, multi-document filtering, intersection of multiple filter conditions, and error cases for missing or mismatched indexes.

| Test Function | Line | Description |
|---|---|---|
| `TestDocEncryptionPeer_WithSimpleRequest_ShouldFetchSuccessfully` | 25-74 | A searchable encryption query on an encrypted-indexed field returns matching document IDs. |
| `TestDocEncryptionPeer_WithMultipleEncryptedFields_QueryShouldSucceed` | 76-160 | Searchable encryption queries on each of multiple encrypted-indexed fields all return the correct document ID. |
| `TestDocEncryptionPeer_WithMultipleDocs_ShouldFilterCorrectly` | 162-262 | A searchable encryption query filters multiple encrypted documents by value and returns only matching IDs. |
| `TestDocEncryption_IfThereIsNoIndex_EncryptedQueryShouldError` | 264-301 | Issuing a searchable encryption query on a collection without any encrypted index returns a GraphQL error. |
| `TestDocEncryption_IfThereIsIndexButOnAnotherField_EncryptedQueryShouldError` | 303-340 | Filtering a searchable encryption query on a non-indexed field returns an invalid argument error. |
| `TestDocEncryptionPeer_WithQueryOnMultipleFields_ShouldReturnIntersection` | 342-424 | A searchable encryption query filtering on multiple encrypted fields returns only the intersecting document IDs. |

### `query_test.go`

Tests that a searchable encryption query fails gracefully when P2P networking is disabled and thus no GQL type is generated.

| Test Function | Line | Description |
|---|---|---|
| `TestEncryptedIndexNew_IfP2PIsDisabled_CanNotDoSEQuery` | 21-46 | A searchable encryption query fails with a GraphQL error when P2P networking is disabled. |

### `replicator_test.go`

Tests replicator resilience for searchable encryption, ensuring SE metadata syncs after a target node recovers from an offline period.

| Test Function | Line | Description |
|---|---|---|
| `TestSEReplicator_IfDocAddedWhileReplicatorIsOffline_ShouldRetry` | 26-92 | Documents added to the source node while the replicator target is offline are synced after it restarts. |
