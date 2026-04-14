# Index: `tests/integration/backup/one_to_one`

## Overview

This folder contains integration tests for backup export and import operations involving one-to-one relationships between collections. The default schema pairs a `User` collection with a `Book` collection where each book holds a unique `author` foreign key (enforced by a unique index). Additional tests exercise schemas with two named one-to-one relations on the same pair of collections. Tests verify that exports capture correct docIDs and foreign key fields, and that imports correctly restore documents, their relation links, and expected error behaviour when uniqueness constraints are violated.

## Test Index

### `export_test.go`

Tests covering backup export of one-to-one related collections, including single-collection filtering, doc updates, and schemas with double named relations.

| Test Function | Line | Description |
|---|---|---|
| `TestBackupExport_JustUserCollection_NoError` | 22-40 | Export backup of a single collection filters to only that collection. |
| `TestBackupExport_AllCollectionsMultipleDocsAndDocUpdate_NoError` | 42-73 | Export backup of all collections preserves one-to-one relation after a doc update. |
| `TestBackupExport_DoubleReletionship_NoError` | 75-122 | Export backup of a schema with two named one-to-one relations includes both foreign key fields. |
| `TestBackupExport_DoubleReletionshipWithUpdate_NoError` | 124-175 | Export backup with two named one-to-one relations and a doc update captures correct updated docIDs. |

### `import_test.go`

Tests covering backup import into one-to-one related collections, including keyless docs, author-ID restoration, unique-index violation detection, and double named relations.

| Test Function | Line | Description |
|---|---|---|
| `TestBackupImport_WithMultipleNoKeyAndMultipleCollections_NoError` | 21-87 | Import backup with docs lacking explicit keys into multiple collections succeeds. |
| `TestBackupImport_WithMultipleNoKeyAndMultipleCollectionsAndUpdatedDocs_NoError` | 89-160 | Import backup with an author ID restores the one-to-one relation from book to author. |
| `TestBackupImport_WithMultipleNoKeyAndMultipleCollectionsAndMultipleUpdatedDocs_NoError` | 162-195 | Import backup with multiple books sharing the same author ID fails due to unique index violation. |
| `TestBackupImport_DoubleRelationshipWithUpdate_NoError` | 197-255 | Import backup with two named one-to-one relations correctly restores both author and favourite links. |
