# Index: `tests/integration/mutation/delete`

## Overview

This directory contains integration tests for the `delete` mutation in DefraDB. The direct test files cover the core deletion API — malformed GQL requests (missing sub-selection), filter-based deletion (single match, multi-match, empty filter, no match, empty collection), docID-based deletion (single ID, multiple IDs, unknown IDs, empty ID list, mixed known/unknown, with transaction isolation, and with result field aliases), and null argument handling for `filter` and `docID` arguments. The subdirectory extends this to relational field kinds, verifying correct soft-deletion behaviour and transactional isolation across one-to-many and one-to-one-to-one relation chains.

## Test Index

### `simple_test.go`

Tests that malformed delete mutations — missing a sub-selection or providing an empty sub-selection — return appropriate GraphQL errors.

| Test Function | Line | Description |
|---|---|---|
| `TestMutationDeletion_WithoutSubSelection` | 21-41 | A delete mutation with no sub-selection returns a GraphQL sub-selection error. |
| `TestMutationDeletion_WithoutSubSelectionFields` | 43-65 | A delete mutation with an empty sub-selection block returns a GraphQL syntax error. |

### `with_filter_test.go`

Tests for filter-based deletion covering single document match, multiple document match, empty filter (delete all), no match, and empty collection.

| Test Function | Line | Description |
|---|---|---|
| `TestMutationDeletion_WithFilter` | 21-54 | Deleting with a filter matching one document removes it and returns its fields. |
| `TestMutationDeletion_WithFilterMatchingMultipleDocs` | 56-107 | Deleting with a filter matching multiple documents removes all matched documents. |
| `TestMutationDeletion_WithEmptyFilter` | 109-159 | Deleting with an empty filter removes all documents in the collection. |
| `TestMutationDeletion_WithFilterNoMatch` | 161-190 | Deleting with a filter that matches no documents returns an empty result. |
| `TestMutationDeletion_WithFilterOnEmptyCollection` | 192-216 | Deleting with a filter on an empty collection returns an empty result. |

### `with_id_test.go`

Tests for single-docID deletion covering unknown docIDs with and without other records present in the collection.

| Test Function | Line | Description |
|---|---|---|
| `TestMutationDeletion_WithIDUnknownValue` | 21-45 | Deleting with an unknown docID returns an empty result. |
| `TestMutationDeletion_WithIDUnknownValueAndUnrelatedRecordInCollection` | 47-76 | Deleting with an unknown docID while another document exists returns an empty result. |

### `with_id_alias_test.go`

Tests that a single-docID deletion result can use an aliased field name in the selection set.

| Test Function | Line | Description |
|---|---|---|
| `TestMutationDeletion_WithIDAndAlias` | 21-54 | Deleting by docID with a field alias in the selection set returns the aliased field name. |

### `with_ids_test.go`

Tests for multi-docID deletion covering simultaneous deletion of two known IDs, empty ID list (no-op), single unknown ID, multiple unknown IDs, and a mixed list of known and unknown IDs.

| Test Function | Line | Description |
|---|---|---|
| `TestMutationDeletion_WithIDs` | 21-63 | Deleting with two known docIDs removes both documents and returns their IDs. |
| `TestMutationDeletion_WithEmptyIDs` | 65-118 | Deleting with an empty docID list is a no-op and leaves all documents intact. |
| `TestMutationDeletion_WithIDsSingleUnknownID` | 120-144 | Deleting with a single unknown docID returns an empty result. |
| `TestMutationDeletion_WithIDsMultipleUnknownID` | 146-170 | Deleting with multiple unknown docIDs returns an empty result. |
| `TestMutationDeletion_WithIDsKnownAndUnknown` | 172-205 | Deleting with a mixed list of known and unknown docIDs removes only the known document. |

### `with_ids_alias_test.go`

Tests that a multi-docID deletion result can use an aliased field name in the selection set.

| Test Function | Line | Description |
|---|---|---|
| `TestMutationDeletion_WithIDsAndSelectAlias` | 21-72 | Deleting with multiple docIDs using a field alias in the selection set returns the aliased field name. |

### `with_ids_filter_test.go`

Tests that combining a docID list with an additional filter argument correctly intersects the two criteria.

| Test Function | Line | Description |
|---|---|---|
| `TestMutationDeletion_WithIDsAndEmptyFilter` | 21-54 | Deleting with a docID list and an empty filter removes the specified document. |

### `with_ids_txn_test.go`

Tests that deleting multiple documents within a transaction is immediately visible to subsequent queries in the same transaction.

| Test Function | Line | Description |
|---|---|---|
| `TestMutationDeletion_WithIDsAndTxn` | 23-74 | Deleting multiple docs inside a transaction makes them invisible to queries within the same transaction. |

### `with_id_txn_test.go`

Tests that deleting a single document within a transaction is immediately visible to subsequent queries in the same transaction.

| Test Function | Line | Description |
|---|---|---|
| `TestMutationDeletion_WithIDAndTxn` | 23-68 | Deleting a doc by ID inside a transaction makes it invisible to queries within the same transaction. |

### `with_ids_update_alias_test.go`

Tests that deletion using multiple docIDs still works correctly when a document has been updated before deletion and the result uses an alias.

| Test Function | Line | Description |
|---|---|---|
| `TestMutationDeletion_WithUpdateAndIDsAndSelectAlias` | 21-80 | Deleting previously updated documents by multiple docIDs with a result alias returns both deleted documents. |

### `with_null_input_test.go`

Tests that null values for the `filter` and `docID` GQL arguments are treated as match-all (no filtering), deleting all existing documents.

| Test Function | Line | Description |
|---|---|---|
| `TestMutationDelete_WithNullFilter_Succeeds` | 21-54 | Passing `filter: null` deletes all documents and returns them. |
| `TestMutationDelete_WithNullDocID_Succeeds` | 56-89 | Passing `docID: null` (singular) deletes all documents and returns them. |
| `TestMutationDelete_WithNullDocIDs_Succeeds` | 91-124 | Passing `docID: null` (treated as list null) deletes all documents and returns them. |

## Subdirectories

| Directory | Summary |
|---|---|
| [`field_kinds/`](field_kinds/INDEX.md) | Tests that delete mutations correctly soft-delete documents across one-to-many and one-to-one-to-one relation chains, with correct `showDeleted` visibility, result aliasing, and transactional read isolation. |
