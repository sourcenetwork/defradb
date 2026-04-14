# Index: `tests/integration/collection_version/updates/replace/field`

## Overview

This folder contains integration tests for the JSON patch `replace` operation applied to individual fields within a collection schema version. The tests verify that replacing a field definition via `PatchCollection` correctly updates the schema, makes the old field name inaccessible, and exposes the new field name for queries.

## Test Index

### `simple_test.go`

Tests that a JSON patch replace operation on a collection field swaps the field's definition in the schema and rejects queries using the old field name.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionUpdatesReplaceField` | 22-66 | Replace a field in a collection schema version using a JSON patch replace operation. |
