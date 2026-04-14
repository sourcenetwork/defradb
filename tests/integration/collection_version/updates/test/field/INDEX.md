# Index: `tests/integration/collection_version/updates/test/field`

## Overview

This folder contains integration tests for the JSON patch `test` operation applied to individual fields within a collection schema version. The tests verify that the `test` operation correctly asserts field property values (name, kind, full field object) and returns errors when the actual values do not match the expected values, using both numeric indices and field name strings as path segments.

## Test Index

### `simple_test.go`

Tests that the JSON patch test operation on collection fields correctly passes when field values match and errors when they do not, covering name assertions, full-object assertions, and field-name-as-path-index variants.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionUpdatesTestFieldNameErrors` | 21-43 | JSON patch test operation on field name with wrong value errors. |
| `TestCollectionVersionUpdatesTestFieldNamePasses` | 45-66 | JSON patch test operation on field name with correct value passes. |
| `TestCollectionVersionUpdatesTestFieldErrors` | 68-90 | JSON patch test operation on a full field object with mismatched kind errors. |
| `TestCollectionVersionUpdatesTestFieldPasses` | 92-116 | JSON patch test operation on a full field object with all correct values passes. |
| `TestCollectionVersionUpdatesTestFieldPasses_UsingFieldNameAsIndex` | 118-142 | JSON patch test operation on a full field using field name as path index passes. |
| `TestCollectionVersionUpdatesTestFieldPasses_TargettingKindUsingFieldNameAsIndex` | 144-165 | JSON patch test operation targeting field Kind via field name as path index passes. |
