# Index: `tests/integration/backup`

## Overview

This directory contains integration tests for DefraDB's backup export and import operations. Tests are organised by relationship complexity: a `simple/` group covers flat single-collection schemas, while `one_to_one/`, `one_to_many/`, and `self_reference/` extend coverage to collections with foreign-key relations of each shape. Across all groups, tests validate that exporting produces correct JSON (with up-to-date docIDs), that collection filtering scopes output correctly, that invalid paths or collection names return descriptive errors, and that importing restores documents, relation links, and unique-index constraints as expected.

## Subdirectories

| Directory | Summary |
|---|---|
| [`one_to_many/`](one_to_many/INDEX.md) | Backup export and import for one-to-many related collections, covering collection filtering, multi-document exports, and relation restoration on import. |
| [`one_to_one/`](one_to_one/INDEX.md) | Backup export and import for one-to-one related collections, including single-collection filtering, doc updates, double named relations, and unique-index violation detection. |
| [`self_reference/`](self_reference/INDEX.md) | Backup export and import for self-referential and multi-collection schemas, verifying that self-referential and cross-collection relation links are preserved across export/import cycles. |
| [`simple/`](simple/INDEX.md) | Backup export and import for a flat single-collection User schema, covering correct JSON output, collection filtering, error handling, and various document key configurations. |
