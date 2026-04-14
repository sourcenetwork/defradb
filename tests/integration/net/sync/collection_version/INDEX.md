# Index: `tests/integration/net/sync/collection_version`

## Overview

This folder contains integration tests for the P2P synchronization of collection versions (schema definitions) in DefraDB. Peers exchange collection versions by CID via `SyncCollectionVersions`; received versions arrive as inactive and must be explicitly activated. These tests verify that initial, patched, and view-backed collection versions are correctly transferred, that ancestor versions are fetched transitively, that existing active versions are preserved, and that activated synced versions can be used for querying.

## Test Index

### `simple_test.go`

Core tests for syncing an initial collection version over P2P and verifying that the receiving node can activate it and execute document queries.

| Test Function | Line | Description |
|---|---|---|
| `TestSyncColVersion_WithInitialColVersion` | 25-98 | Syncing an initial collection version transfers it to the peer as inactive. |
| `TestSyncColVersion_WithInitialColVersion_CanBeActivatedAndQueried` | 100-161 | A synced collection version can be activated on the receiving peer and queried. |

### `patch_test.go`

Tests for syncing patched (evolved) collection versions, covering both cases where the base version is unknown and where it already exists and is active on the receiving node.

| Test Function | Line | Description |
|---|---|---|
| `TestSyncColVersion_WithPatchVersionOfUnknownCollection` | 25-110 | Syncing a patched collection version fetches all ancestor versions for an unknown collection. |
| `TestSyncColVersion_WithPatchVersionOfKnownCollection` | 112-197 | Syncing a patched version of a locally-known collection adds it as inactive without deactivating the active version. |

### `view_test.go`

Tests for syncing view collection versions (non-materialized, lens-transformed) over P2P, confirming correct inactive/non-materialized state on receipt and successful activation with query execution.

| Test Function | Line | Description |
|---|---|---|
| `TestSyncColVersion_WithView` | 27-127 | Syncing a view collection version transfers it to the peer as non-materialized and inactive. |
| `TestSyncColVersion_WithView_CanBeActivatedAndQueried` | 129-207 | A synced view collection version can be activated and returns lens-transformed query results. |
