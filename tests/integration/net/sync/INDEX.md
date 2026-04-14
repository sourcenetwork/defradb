# Index: `tests/integration/net/sync`

## Overview

This directory contains integration tests for the explicit P2P synchronization primitives in DefraDB: branchable collection sync and collection version sync. Unlike subscription-based propagation, these mechanisms are invoked directly (`SyncBranchableCollection`, `SyncCollectionVersions`) and govern how peers exchange DAG commit heads and schema definitions, including correct handling of divergent versions, multi-hop network topologies, and activation of received versions.

## Subdirectories

| Directory | Summary |
|---|---|
| [`branchable_collection/`](branchable_collection/INDEX.md) | Tests P2P sync of branchable collections, covering simple two-node sync, multi-hop network topologies, divergent schema versions, and correct DAG structure after sync. |
| [`collection_version/`](collection_version/INDEX.md) | Tests P2P sync of collection versions (schema definitions), verifying that initial, patched, and view-backed versions are transferred as inactive, ancestor versions are fetched transitively, and synced versions can be activated and queried. |
