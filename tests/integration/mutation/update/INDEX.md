# Index: `tests/integration/mutation/update`

## Overview

This directory contains integration tests for the `update` mutation in DefraDB. The direct test files cover fundamental update behaviour — updating by a single docID or multiple docIDs, filter-based updates (boolean filters, relation filters, filter on the updated field, relation traversal in the result), `_version` CID retrieval after update, default-value non-overwrite semantics, update-after-delete no-op behaviour via `Collection.Save`, underscored collection names, null argument handling for `filter` and `docID`, and error cases for invalid filter types, empty filters passed to the Go client, invalid JSON, and invalid updater shapes. The subdirectories extend coverage to schema-level constraints (`@constraints`), CRDT counter fields (`pcounter`/`pncounter`), automatic vector embedding regeneration (`@embedding`), and relational field kinds (one-to-many and one-to-one).

## Test Index

### `underscored_collection_test.go`

Tests that an update mutation works correctly on a collection whose type name contains an underscore.

| Test Function | Line | Description |
|---|---|---|
| `TestMutationUpdateUnderscoredCollection` | 21-62 | Updating a document in an underscored-name collection correctly stores and returns the new value. |

### `with_default_values_test.go`

Tests that updating a document does not overwrite existing field values with schema defaults.

| Test Function | Line | Description |
|---|---|---|
| `TestMutationUpdate_WithDefaultValues_DoesNotOverwrite` | 21-63 | Updating a subset of fields does not overwrite other fields with their `@default` values. |

### `with_delete_test.go`

Tests that attempting to update a previously deleted document via `Collection.Save` returns an appropriate error.

| Test Function | Line | Description |
|---|---|---|
| `TestUpdateSave_DeletedDoc_DoesNothing` | 24-57 | Saving an update to a deleted document via Collection.Save returns a deleted-document error. |

### `with_filter_action_test.go`

Tests for the lower-level `UpdateWithFilter` action covering invalid filter types, empty filter, invalid JSON, invalid updater shape, patch no-op, and successful filter-based update via the Collection API.

| Test Function | Line | Description |
|---|---|---|
| `TestUpdateWithInvalidFilterType_ReturnsError` | 24-42 | Passing a non-JSON-serialisable filter type via HTTP/CLI client returns a collection-not-found error. |
| `TestUpdateWithInvalidFilterType_WithGoClient_ReturnsError` | 44-59 | Passing an unsupported filter type via the Go client returns an unsupported-filter-type error. |
| `TestUpdateWithEmptyFilter_ReturnsError` | 61-74 | Passing an empty-string filter returns a filter-cannot-be-empty error. |
| `TestUpdateWithInvalidJSON_ReturnsError` | 76-96 | Passing a non-JSON filter string returns a JSON parse error. |
| `TestUpdateWithInvalidUpdater_ReturnsError` | 98-118 | Passing a non-object updater string returns an invalid-updater-type error. |
| `TestUpdateWithPatch_DoesNothing` | 120-152 | Passing a JSON array as the updater is treated as a no-op patch and leaves the document unchanged. |
| `TestUpdateWithFilter_Succeeds` | 154-185 | A valid filter and object updater correctly updates the matching document. |

### `with_filter_test.go`

Tests for GQL filter-based update mutations covering boolean field filters, relation-field filters, filter on the field being updated, and inclusion of related documents in the mutation result.

| Test Function | Line | Description |
|---|---|---|
| `TestMutationUpdate_WithBooleanFilter_ResultNotFilteredOut` | 23-61 | An update that changes the field used in the filter still returns the updated document. |
| `TestMutationUpdate_WithBooleanFilter` | 63-120 | A boolean filter selects and updates only the matching documents, leaving others unchanged. |
| `TestMutationUpdate_WithFilterOnUpdatedField_ReturnsResult` | 125-163 | Filtering on the field being updated returns the post-update value in the result. |
| `TestMutationUpdate_WithRelationSelectInResponse_ReturnsRelation` | 165-224 | An update mutation can include a related document in its result selection. |
| `TestMutationUpdate_WithRelationFilter_CorrectlyFilters` | 226-295 | A filter on a related document's field correctly restricts which documents are updated. |
| `TestMutationUpdate_WithRelationFilterAndRelationSelect_ReturnsBoth` | 297-372 | A relation filter combined with a relation select in the result returns the correct updated document and its relation. |

### `with_id_test.go`

Tests for single-docID updates covering the success path and a non-existent docID (no-op returning empty results).

| Test Function | Line | Description |
|---|---|---|
| `TestMutationUpdate_WithId` | 21-65 | Updating by a single known docID applies the change and returns the updated document. |
| `TestMutationUpdate_WithNonExistantId` | 67-101 | Updating with a docID that does not exist returns an empty result without error. |

### `with_id_and_version_test.go`

Tests that an update mutation including `_version { cid }` in the result returns the full version history with both the new and prior commit CIDs.

| Test Function | Line | Description |
|---|---|---|
| `TestMutationUpdate_WithIdAndVersion_ReturnResults` | 21-76 | Updating by docID and selecting `_version { cid }` returns the update and create commit CIDs. |

### `with_ids_test.go`

Tests for multi-docID updates, verifying that only the specified documents are modified.

| Test Function | Line | Description |
|---|---|---|
| `TestMutationUpdate_WithIds` | 21-80 | Updating with a list of docIDs applies the change to exactly those documents and returns them. |

### `with_null_input_test.go`

Tests that null values for the `filter` and `docID` GQL arguments are treated as match-all, updating all existing documents.

| Test Function | Line | Description |
|---|---|---|
| `TestMutationUpdate_WithNullFilter_Succeeds` | 21-54 | Passing `filter: null` updates all documents in the collection. |
| `TestMutationUpdate_WithNullDocID_Succeeds` | 56-89 | Passing `docID: null` (singular) updates all documents in the collection. |
| `TestMutationUpdate_WithNullDocIDs_Succeeds` | 91-124 | Passing `docID: null` (treated as list null) updates all documents in the collection. |

## Subdirectories

| Directory | Summary |
|---|---|
| [`constraints/`](constraints/INDEX.md) | Tests that the `@constraints(size: N)` directive is enforced during document update, covering both the success path and the error path for mismatched array lengths. |
| [`crdt/`](crdt/INDEX.md) | Tests for PCounter and PNCounter CRDT field updates, verifying correct accumulation, rejection of invalid operations, and overflow edge cases across Int and Float numeric kinds. |
| [`embeddings/`](embeddings/INDEX.md) | Tests that the `@embedding` directive correctly re-generates vector embeddings when source fields are updated, and suppresses generation when the vector or only unrelated fields are supplied. |
| [`field_kinds/`](field_kinds/INDEX.md) | Tests for re-linking relational fields during update across one-to-many and one-to-one relations, covering both alias and raw ID fields, invalid-side rejection, and error cases. |
