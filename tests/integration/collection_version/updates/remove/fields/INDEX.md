# Index: `tests/integration/collection_version/updates/remove/fields`

## Overview

This folder tests the behaviour of JSON-Patch `remove` operations applied to collection fields via `PatchCollection`. It covers both valid removals — removing a single field or the entire fields array — and invalid removals that target individual properties of an existing field (`Name`, `Kind`, `Typ`), which are rejected because mutating field metadata is not supported.

## Test Index

### `simple_test.go`

Tests that removing a field or all fields succeeds with correct query behaviour, and that removing individual field properties returns mutation errors.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionUpdatesRemoveField` | 22-65 | Removing a field from the schema version makes it unqueryable while other fields remain. |
| `TestCollectionVersionUpdatesRemoveAllFields` | 67-101 | Removing all fields from the schema version leaves only the built-in _docID queryable. |
| `TestCollectionVersionUpdatesRemoveFieldNameErrors` | 103-126 | Removing the Name property of an existing field returns a mutation error. |
| `TestCollectionVersionUpdatesRemoveFieldKindErrors` | 128-151 | Removing the Kind property of an existing field returns a mutation error. |
| `TestCollectionVersionUpdatesRemoveFieldTypErrors` | 153-176 | Removing the Typ property of an existing field returns a mutation error. |
