# Index: `tests/integration/collection_version/updates/replace`

## Overview

This directory tests the behaviour of JSON-Patch `replace` operations applied to collection schema versions via `PatchCollection`. The tests verify that replacing a field definition correctly updates the schema, makes the old field name inaccessible, and exposes the replacement field for queries.

## Subdirectories

| Directory | Summary |
|---|---|
| [`field/`](field/INDEX.md) | Tests for JSON-Patch `replace` operations on individual collection fields, verifying that a replaced field is correctly reflected in the schema and query results. |
