# Index: `tests/integration/mutation/delete/field_kinds/one_to_many`

## Overview

This folder contains integration tests for delete mutation behavior in one-to-many relationships. The tests verify that deleting a document on the "many" side of a one-to-many relation is correctly reflected when querying with the `showDeleted` flag, asserting that the `_deleted` field accurately distinguishes soft-deleted documents from live ones across related collections.

## Test Index

### `with_show_deleted_test.go`

Tests that deleting a book document in a one-to-many Author-Book relation is correctly surfaced via the `showDeleted` query flag.

| Test Function | Line | Description |
|---|---|---|
| `TestDeletionOfADocumentUsingSingleDocIDWithShowDeletedDocumentQuery` | 34-118 | Delete one book in a one-to-many relation and verify showDeleted reflects correct state. |
