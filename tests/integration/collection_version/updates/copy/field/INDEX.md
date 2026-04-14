# Index: `tests/integration/collection_version/updates/copy/field`

## Overview

This folder tests the behaviour of JSON-Patch `copy` operations applied to collection fields via `PatchCollection`. Because field positions are immutable in DefraDB, a raw copy that leaves a field at a new index is rejected. The valid use-case — treating an existing field as a template, then overriding its `Name` (and optionally its `Kind`) while removing the `FieldID` — is tested both for correct data storage and for correct schema introspection output.

## Test Index

### `simple_test.go`

Tests that copying a field without renaming it is rejected, that multiple such copies accumulate errors, and that the valid template pattern (copy + rename, with optional kind substitution) works correctly.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionUpdatesCopyFieldErrors` | 21-55 | Copying a field to a new index returns an unsupported error. |
| `TestCollectionVersionUpdatesCopyFieldErrorsMultiple` | 57-92 | Multiple copy field operations in a patch each return unsupported errors. |
| `TestCollectionVersionUpdatesCopyFieldWithAndReplaceName` | 94-132 | Copying a field and replacing its name adds a new field to the schema version. |
| `TestCollectionVersionUpdatesCopyFieldWithReplaceNameAndKindSubstitution` | 135-185 | Copying a field then replacing its name and kind adds a correctly-typed new field. |
| `TestCollectionVersionUpdatesCopyFieldAndReplaceNameAndInvalidKindSubstitution` | 188-214 | Copying a field and replacing its kind with an invalid type returns an error. |

### `with_introspection_test.go`

Tests that a field added via the copy-template pattern is correctly reflected in the GraphQL schema introspection response.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionUpdatesCopyFieldIntrospectionWithRemoveIDAndReplaceName` | 22-83 | Copying a field, removing its ID, and renaming it is visible via schema introspection. |
