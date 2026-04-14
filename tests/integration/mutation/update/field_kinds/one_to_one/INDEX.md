# Index: `tests/integration/mutation/update/field_kinds/one_to_one`

## Overview

This folder contains integration tests for mutation update operations on one-to-one relation fields in DefraDB. The tests cover linking documents via the primary side of the relation using both the relation name alias and explicit relation ID fields, verify that setting a relation from the secondary side is correctly rejected, assert that unique-index violations and malformed document IDs produce the expected errors, and include a self-referencing one-to-one scenario as well as GQL mutation results.

## Test Index

### `with_alias_test.go`

Tests that use the relation alias name (e.g. `"published"`) rather than the explicit `_publishedID` field when updating one-to-one relations, covering primary-side unique-index violations, secondary-side rejections for both Collection API and GQL, and invalid-length ID errors.

| Test Function | Line | Description |
|---|---|---|
| `TestMutationUpdateOneToOne_AliasRelationNameToLinkFromPrimarySide` | 25-65 | Update one-to-one relation via alias from primary side errors on unique index violation. |
| `TestMutationUpdateOneToOne_AliasRelationNameToLinkFromSecondarySide_CollectionApi` | 67-111 | Update one-to-one relation via alias from secondary side errors using Collection API. |
| `TestMutationUpdateOneToOne_AliasRelationNameToLinkFromSecondarySide_GQL` | 113-156 | Update one-to-one relation via alias from secondary side errors using GQL. |
| `TestMutationUpdateOneToOne_AliasWithInvalidLengthRelationIDToLink_Error` | 158-193 | Update one-to-one relation via alias with malformed relation ID errors. |

### `with_self_ref_test.go`

Tests that a self-referencing one-to-one relation on a single collection can be set from the primary side and queried correctly from both directions.

| Test Function | Line | Description |
|---|---|---|
| `TestMutationUpdateOneToOne_SelfReferencingFromPrimary` | 22-112 | Update self-referencing one-to-one relation from primary side and verify both traversal directions. |

### `with_simple_test.go`

Tests that cover the full range of simple one-to-one relation update scenarios: linking to non-existent documents, successful primary-side linking, secondary-side rejection for Collection API and GQL, unique-index violations, invalid-length IDs, and GQL mutation result shape.

| Test Function | Line | Description |
|---|---|---|
| `TestMutationUpdateOneToOneNoChild` | 27-66 | Update one-to-one relation to a non-existent document ID succeeds without error. |
| `TestMutationUpdateOneToOne` | 68-142 | Update one-to-one relation from primary side and verify both sides reflect the link. |
| `TestMutationUpdateOneToOneSecondarySide_CollectionApi` | 144-180 | Update one-to-one relation from secondary side errors using Collection API. |
| `TestMutationUpdateOneToOneSecondarySide_GQL` | 182-217 | Update one-to-one relation from secondary side errors using GQL. |
| `TestMutationUpdateOneToOne_RelationIDToLinkFromPrimarySide` | 219-259 | Update one-to-one relation ID from primary side errors when book is already linked. |
| `TestMutationUpdateOneToOne_RelationIDToLinkFromSecondarySide_CollectionApi` | 261-304 | Update one-to-one relation ID from secondary side errors using Collection API. |
| `TestMutationUpdateOneToOne_RelationIDToLinkFromSecondarySide_GQL` | 306-348 | Update one-to-one relation ID from secondary side errors using GQL. |
| `TestMutationUpdateOneToOne_InvalidLengthRelationIDToLink_Error` | 350-384 | Update one-to-one relation with an invalid-length relation ID errors with UUID parse error. |
| `TestMutationUpdateOneToOne_WithGQLRequest_ReturnsResults` | 386-433 | Update one-to-one relation via GQL mutation returns the updated document with linked relation. |
