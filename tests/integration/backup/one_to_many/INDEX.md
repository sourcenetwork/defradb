# Index: `tests/integration/backup/one_to_many`

## Overview

This folder contains integration tests for backup export and import operations involving one-to-many relationships between collections. The schema used throughout the tests pairs a `User` collection (the "one" side) with a `Book` collection (the "many" side, where each book holds an author foreign key). Tests verify that exporting all or a subset of collections produces correct JSON with up-to-date docIDs, and that importing JSON backup data (with or without explicit docIDs) correctly restores documents and their relational links.

## Test Index

### `export_test.go`

Tests covering backup export of one-to-many related collections, including collection filtering, multi-document scenarios, and document-update handling.

| Test Function | Line | Description |
|---|---|---|
| `TestBackupExport_JustUserCollection_NoError` | 22-40 | Export backup of a single collection filters to only that collection. |
| `TestBackupExport_AllCollectionsMultipleDocsAndDocUpdate_NoError` | 42-73 | Export backup of all collections preserves one-to-many relation after a doc update. |
| `TestBackupExport_AllCollectionsMultipleDocsAndMultipleDocUpdate_NoError` | 75-113 | Export backup of all collections with multiple docs and multiple updates includes correct docIDs. |

### `import_test.go`

Tests covering backup import into one-to-many related collections, including keyless docs and docs carrying updated docIDs and foreign key references.

| Test Function | Line | Description |
|---|---|---|
| `TestBackupImport_WithMultipleNoKeyAndMultipleCollections_NoError` | 21-87 | Import backup with docs lacking explicit keys into multiple collections succeeds. |
| `TestBackupImport_WithMultipleNoKeyAndMultipleCollectionsAndUpdatedDocs_NoError` | 89-179 | Import backup with updated docs restores one-to-many relation linking books to their author. |
