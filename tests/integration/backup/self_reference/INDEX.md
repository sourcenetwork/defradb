# Index: `tests/integration/backup/self_reference`

## Overview

This folder contains integration tests for backup export and import of collections that contain self-referential relations, specifically a `User` type whose `boss` field points back to another `User`. The tests also cover multi-collection scenarios where two different types share split or dual primary relations, validating that document links are preserved (or correctly documented as unlinked) across export/import cycles, including across two separate nodes.

## Test Index

### `export_test.go`

Tests verifying that exporting a self-referential User collection produces correct JSON backup content, including after document updates.

| Test Function | Line | Description |
|---|---|---|
| `TestBackupExport_Simple_NoError` | 22-44 | Export a self-referential User collection with a boss relation. |
| `TestBackupExport_MultipleDocsAndDocUpdate_NoError` | 46-70 | Export self-referential documents after an update, verifying new docIDs are reflected. |

### `import_test.go`

Tests verifying that importing self-referential and multi-collection backup content correctly restores document relations, handles ordering, and documents known edge-case limitations.

| Test Function | Line | Description |
|---|---|---|
| `TestBackupSelfRefImport_Simple_NoError` | 23-74 | Import a self-referential User backup and verify boss relations resolve correctly. |
| `TestBackupSelfRefImport_SelfRef_NoError` | 76-145 | Export and re-import a document with a self-referential boss relation across two nodes. |
| `TestBackupSelfRefImport_PrimaryRelationWithSecondCollection_NoError` | 147-212 | Import backup with two collections sharing primary relations and verify cross-collection links. |
| `TestBackupSelfRefImport_PrimaryRelationWithSecondCollectionWrongOrder_NoError` | 214-279 | Import backup with collections in reverse order and verify primary relations still resolve correctly. |
| `TestBackupSelfRefImport_SplitPrimaryRelationWithSecondCollection_NoError` | 283-387 | Import a backup with split primary relations documents remain unlinked due to a known issue. |
