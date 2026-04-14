# Index: `tests/integration/collection_version/updates/move/field`

## Overview

This folder tests the behaviour of JSON-Patch `move` operations applied to collection fields via `PatchCollection`. Moving fields to a different index within a collection is not supported in DefraDB; these tests assert that the appropriate errors are returned both for the directly moved field and for any other fields whose indices are displaced by the operation.

## Test Index

### `simple_test.go`

Tests that a single move operation and its displaced-field side-effects each produce the expected unsupported-operation errors.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionUpdatesMoveFieldErrors` | 21-44 | Moving a field to a different index returns an unsupported error. |
| `TestCollectionVersionUpdatesMoveFieldErrorsMultiple` | 46-69 | Moving a field reports unsupported errors for all displaced fields, not just the moved one. |
