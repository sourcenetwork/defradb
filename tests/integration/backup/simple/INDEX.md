# Index: `tests/integration/backup/simple`

## Overview

This folder contains integration tests for backup export and import of a flat, single-collection `User` schema (name and age fields, no relations). The export tests validate correct JSON output, collection filtering, and error handling for invalid file paths or collection names. The import tests verify that documents are restored and queryable, cover documents with and without docID keys, empty objects, and confirm atomic failure when any document in a batch contains an invalid field.

## Test Index

### `export_test.go`

Tests verifying that exporting a simple flat User collection produces the correct backup content or returns the expected error under various configurations.

| Test Function | Line | Description |
|---|---|---|
| `TestBackupExport_Simple_NoError` | 22-37 | Export a flat User document and verify the JSON backup content. |
| `TestBackupExport_Empty_NoError` | 39-54 | Export a document with no fields set and verify the backup contains only the docID. |
| `TestBackupExport_WithInvalidFilePath_ReturnError` | 56-74 | Export backup to a non-existent nested path returns a file creation error. |
| `TestBackupExport_WithInvalidCollection_ReturnError` | 76-94 | Export backup filtered to a non-existent collection name returns a collection not found error. |
| `TestBackupExport_JustUserCollection_NoError` | 96-114 | Export backup filtered to the User collection only produces correct scoped output. |

### `import_test.go`

Tests verifying that importing simple flat User documents succeeds or fails as expected across a variety of input shapes and error conditions.

| Test Function | Line | Description |
|---|---|---|
| `TestBackupImport_Simple_NoError` | 21-49 | Import a flat User document from backup and verify it is queryable. |
| `TestBackupImport_WithInvalidFilePath_ReturnError` | 51-63 | Import backup from a non-existent file path returns a file open error. |
| `TestBackupImport_WithInvalidCollection_ReturnError` | 65-77 | Import backup content referencing an unknown collection name returns a collection not found error. |
| `TestBackupImport_WithDocAlreadyExists_ReturnError` | 79-95 | Import backup for a document that already exists returns a duplicate document error. |
| `TestBackupImport_WithNoKeys_NoError` | 97-125 | Import a document without docID keys and verify it is assigned an ID and queryable. |
| `TestBackupImport_WithMultipleNoKeys_NoError` | 127-167 | Import multiple documents without docID keys and verify all are assigned IDs and queryable. |
| `TestBackupImport_EmptyObject_NoError` | 169-195 | Import an empty object document and verify it is created with all fields null. |
| `TestBackupImport_WithMultipleNoKeysAndInvalidField_Errors` | 197-226 | Import multiple documents where one has an invalid field errors and commits no documents. |
