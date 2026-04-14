# Index: `tests/integration/mutation/add/field_kinds/one_to_one`

## Overview

This folder contains integration tests for the `add` (create/insert) mutation on collections that have one-to-one relation fields. The tests verify correct linking and bidirectional traversal of one-to-one relations, as well as expected error behaviour when mutations are attempted from the secondary side or with invalid fields. Both the standard foreign-key field name (e.g., `_publishedID`) and the aliased relation field name (e.g., `published`) are exercised, along with explicit null values on the primary side.

## Test Index

### `with_simple_test.go`

Tests covering basic add-mutation behaviour on one-to-one relations using the raw foreign-key field, including valid linking, secondary-side rejection, and uniqueness enforcement.

| Test Function | Line | Description |
|---|---|---|
| `TestMutationAddOneToOne_WithInvalidField_Error` | 24-45 | Adding a one-to-one relation doc with an invalid field name returns an error. |
| `TestMutationAddOneToOneNoChild` | 49-77 | Adding a one-to-one relation doc referencing a non-existing child doc succeeds. |
| `TestMutationAddOneToOne` | 79-142 | Adding a one-to-one relation links both sides and is queryable bidirectionally. |
| `TestMutationAddOneToOneSecondarySide_CollectionApi` | 144-170 | Setting a one-to-one relation from the secondary side via collection API returns an error. |
| `TestMutationAddOneToOneSecondarySide_GQL` | 172-197 | Setting a one-to-one relation from the secondary side via GQL mutation returns an error. |
| `TestMutationAddOneToOne_ErrorsGivenRelationAlreadyEstablishedViaPrimary` | 199-228 | Adding a second primary-side doc pointing to the same relation target errors on unique index. |

### `with_alias_test.go`

Tests covering the same one-to-one add-mutation scenarios using the aliased relation field name instead of the raw foreign-key identifier.

| Test Function | Line | Description |
|---|---|---|
| `TestMutationAddOneToOne_UseAliasWithInvalidField_Error` | 24-45 | Adding a one-to-one doc using an aliased relation field with an invalid field name errors. |
| `TestMutationAddOneToOne_UseAliasWithNonExistingRelationPrimarySide_AddedDoc` | 49-77 | Adding a one-to-one doc via alias referencing a non-existing primary-side doc still succeeds. |
| `TestMutationAddOneToOne_UseAliasedRelationNameToLink_QueryFromPrimarySide` | 79-140 | Using an aliased relation name to link a one-to-one relation is queryable from both sides. |
| `TestMutationAddOneToOne_UseAliasedRelationNameToLink_CollectionAPI_Errors` | 142-168 | Using an aliased relation name to set a one-to-one link from the secondary side via collection API errors. |
| `TestMutationAddOneToOne_UseAliasedRelationNameToLink_GQL_Errors` | 170-195 | Using an aliased relation name to set a one-to-one link from the secondary side via GQL errors. |

### `with_null_value_test.go`

Tests covering the behaviour of one-to-one add mutations when an explicit null is supplied for the relation field on the primary side.

| Test Function | Line | Description |
|---|---|---|
| `TestMutationAddOneToOne_WithExplicitNullOnPrimarySide` | 21-85 | Adding a one-to-one doc with an explicit null relation on the primary side stores nil. |
