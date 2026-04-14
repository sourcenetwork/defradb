# Index: `tests/integration/collection_version/updates/copy`

## Overview

This directory contains integration tests for JSON Patch `copy` operations applied to collection versions via `PatchCollection`. The subdirectories verify that copying a field to a bare new index is rejected, and that the valid template pattern — copying a field and then overriding its name and optionally its kind — produces the expected schema changes and is correctly reflected in GraphQL introspection.

## Subdirectories

| Directory | Summary |
|---|---|
| [`field/`](field/INDEX.md) | Tests that raw field copy operations are rejected and that the copy-then-rename template pattern correctly adds a new typed field to the collection schema. |
