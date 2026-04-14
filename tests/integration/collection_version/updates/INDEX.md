# Index: `tests/integration/collection_version/updates`

## Overview

This directory contains integration tests for JSON Patch operations applied to collection schema versions via `PatchCollection`, and for managing active collection version branches. The direct test file covers branching scenarios — creating multiple inactive and active version branches, patching across branches, switching the active branch with `SetActiveCollectionVersion`, and verifying that only the active branch's fields are queryable. The subdirectories cover the individual JSON Patch operations (`add`, `copy`, `move`, `remove`, `replace`, `test`) and their effects on collection schema evolution.

## Test Index

### `with_version_branch_test.go`

Tests branching behaviour when `PatchCollection` creates diverging collection version lineages, including querying fields on active vs. inactive branches and the effect of explicitly switching the active version.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionUpdates_WithBranchingCollection` | 25-181 | Branching a collection creates multiple inactive and active version branches, with only the active branch's fields queryable. |
| `TestCollectionVersionUpdates_WithPatchOnBranchedCollection` | 183-325 | Patching an active branched collection version extends the active branch, making the new fields queryable. |
| `TestCollectionVersionUpdates_WithBranchingCollectionAndSetActiveCollectionToOtherBranch` | 327-422 | Setting the active collection version to a different branch makes that branch's fields queryable and the previous branch's fields inaccessible. |
| `TestCollectionVersionUpdates_WithBranchingCollectionAndSetActiveCollectionToOtherBranchThenPatch` | 424-570 | Patching after switching to a different branch extends that branch, with the new version sourced from the switched-to branch. |
| `TestCollectionVersionUpdates_WithBranchingCollectionAndGetCollectionAtVersion` | 572-609 | After creating a new active collection version, the original version can still be retrieved by its version ID and is marked inactive. |

## Subdirectories

| Directory | Summary |
|---|---|
| [`add/`](add/INDEX.md) | Tests adding new fields to a collection version, covering field kind, CRDT type, and constraint validation. |
| [`copy/`](copy/INDEX.md) | Tests that raw field copy operations are rejected and that the copy-then-rename template pattern correctly adds a new typed field to the collection schema. |
| [`move/`](move/INDEX.md) | Tests that field move operations return unsupported-operation errors, including errors for all displaced fields affected by the move. |
| [`remove/`](remove/INDEX.md) | Tests for JSON-Patch `remove` operations on collection fields, covering valid field removals and invalid attempts to remove individual field properties. |
| [`replace/`](replace/INDEX.md) | Tests for JSON-Patch `replace` operations on individual collection fields, verifying that a replaced field is correctly reflected in the schema and query results. |
| [`test/`](test/INDEX.md) | Tests for JSON-Patch `test` operations on individual collection fields, covering name assertions, full-object assertions, and field-name-as-path-index variants. |
