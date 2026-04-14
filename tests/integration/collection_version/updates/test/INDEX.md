# Index: `tests/integration/collection_version/updates/test`

## Overview

This directory tests the behaviour of JSON-Patch `test` operations applied to collection schema versions via `PatchCollection`. The tests verify that the `test` operation correctly asserts field property values and returns errors when actual values do not match expected values, using both numeric indices and field name strings as path segments.

## Subdirectories

| Directory | Summary |
|---|---|
| [`field/`](field/INDEX.md) | Tests for JSON-Patch `test` operations on individual collection fields, covering name assertions, full-object assertions, and field-name-as-path-index variants. |
