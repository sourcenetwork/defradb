# Index: `tests/integration/mutation/add/field_kinds/one_to_many`

## Overview

These tests cover the `add` mutation on one-to-many relation fields, using the shared `Book`/`Author` schema defined in `utils.go`. They assert correct behavior when creating documents that reference related documents via foreign key ID fields or aliased relation names. The tests cover both valid linking scenarios and expected error conditions such as invalid field names or using the one-side relation field as a foreign key.

## Test Index

### `with_alias_test.go`

Tests for adding documents in a one-to-many relation using the aliased relation name instead of the internal foreign key ID field, covering error cases, successful linking, and docID consistency.

| Test Function | Line | Description |
|---|---|---|
| `TestMutationAddOneToMany_AliasedRelationNameWithInvalidField_Error` | 24-44 | Adding a document with an invalid field name returns an error. |
| `TestMutationAddOneToMany_AliasedRelationNameNonExistingRelationSingleSide_NoIDFieldError` | 46-67 | Adding a document using the aliased relation field on the one-side returns an error. |
| `TestMutationAddOneToMany_AliasedRelationNameNonExistingRelationManySide_AddedDoc` | 71-99 | Adding a document with an aliased relation referencing a non-existing document succeeds. |
| `TestMutationAddOneToMany_AliasedRelationNameToLinkFromManySide` | 101-164 | Adding a document using the aliased relation name correctly links the one-to-many relation. |
| `TestMutationUpdateOneToMany_AliasRelationNameAndInternalIDBothProduceSameDocID` | 166-240 | Linking via internal relation ID field produces the same docID as using the aliased name. |

### `with_simple_test.go`

Tests for adding documents in a one-to-many relation using the internal foreign key ID field, covering error cases and successful linking via the many-side.

| Test Function | Line | Description |
|---|---|---|
| `TestMutationAddOneToMany_WithInvalidField_Error` | 24-44 | Adding a document with an unknown field name in a one-to-many relation returns an error. |
| `TestMutationAddOneToMany_NonExistingRelationSingleSide_NoIDFieldError` | 46-67 | Adding a document using the foreign key ID field on the one-side returns an error. |
| `TestMutationAddOneToMany_NonExistingRelationManySide_AddedDoc` | 71-99 | Adding a document with a foreign key referencing a non-existing document succeeds. |
| `TestMutationAddOneToMany_RelationIDToLinkFromManySide` | 101-164 | Adding a document with a foreign key ID correctly links the one-to-many relation. |
