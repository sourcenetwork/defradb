# Index: `tests/integration/collection_version/updates/move`

## Overview

This directory contains integration tests for JSON Patch `move` operations applied to collection versions via `PatchCollection`. The subdirectories verify that moving fields to a different index is not supported in DefraDB, and that the appropriate errors are returned for both the moved field and any fields whose indices are displaced by the operation.

## Subdirectories

| Directory | Summary |
|---|---|
| [`field/`](field/INDEX.md) | Tests that field move operations return unsupported-operation errors, including errors for all displaced fields affected by the move. |
