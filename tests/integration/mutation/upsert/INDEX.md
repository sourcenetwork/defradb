# Index: `tests/integration/mutation/upsert`

## Overview

This folder contains integration tests for the `upsert` mutation operation in DefraDB. The tests verify the core upsert semantics — inserting a new document when no filter match is found and updating the matched document when exactly one is found — as well as error cases for null inputs, multiple filter matches, unique index violations, and correct handling of `DateTime` fields with `UTC_NOW`.

## Test Index

### `date_time_test.go`

Tests that upsert correctly handles `DateTime` fields using the `UTC_NOW` scalar on both the insert and update code paths.

| Test Function | Line | Description |
|---|---|---|
| `TestMutationUpsert_WithDateTimeField_WithUTCNow_ShouldBeEqual` | 21-67 | Upsert sets DateTime fields to UTC_NOW on both insert and update paths. |

### `simple_test.go`

Tests the core upsert behaviour including insert-on-miss, update-on-match, same-field update, multiple-match errors, null-argument errors, unique index violations, and result consistency after update.

| Test Function | Line | Description |
|---|---|---|
| `TestMutationUpsertSimple_WithNoFilterMatch_AddsNewDoc` | 21-84 | Upsert creates a new document when no existing document matches the filter. |
| `TestMutationUpsertSimple_WithFilterMatch_UpdatesDoc` | 86-155 | Upsert updates an existing document when the filter matches exactly one document. |
| `TestMutationUpsertSimple_WithFilterMatchOnSameField_UpdatesDoc` | 157-226 | Upsert updates a document field even when the update targets the same field used in the filter. |
| `TestMutationUpsertSimple_WithFilterMatchMultiple_ReturnsError` | 228-269 | Upsert returns an error when the filter matches more than one document. |
| `TestMutationUpsertSimple_WithNullAddInput_ReturnsError` | 271-300 | Upsert returns an error when the add argument is null. |
| `TestMutationUpsertSimple_WithNullUpdateInput_ReturnsError` | 302-331 | Upsert returns an error when the update argument is null. |
| `TestMutationUpsertSimple_WithNullFilterInput_ReturnsError` | 333-362 | Upsert returns an error when the filter argument is null. |
| `TestMutationUpsertSimple_WithUniqueCompositeIndexAndDuplicateUpdate_ReturnsError` | 364-405 | Upsert returns an error when an update would violate a unique composite index. |
| `TestMutationUpsertSimple_WithFilterMatchAndVersion_UpdatesDoc` | 407-476 | Upsert updates a matched document and the result is consistent with a subsequent query. |
