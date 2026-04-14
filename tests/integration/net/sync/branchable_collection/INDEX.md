# Index: `tests/integration/net/sync/branchable_collection`

## Overview

This folder contains integration tests for the P2P synchronization of branchable collections in DefraDB. Branchable collections maintain a DAG of collection-level commit heads that peers exchange explicitly via `SyncBranchableCollection`, rather than through automatic subscription; these tests verify correct data propagation across simple two-node pairs, complex multi-hop networks, divergent schema versions, and correct DAG structure after sync.

## Test Index

### `simple_test.go`

Basic two-node branchable collection sync scenarios covering empty-to-populated sync, bidirectional merge, absence of automatic subscription, and error handling for invalid inputs.

| Test Function | Line | Description |
|---|---|---|
| `TestBranchableCollectionSync_OneNodeEmptyAnotherWithDocs_ShouldCopyAll` | 25-86 | Branchable collection syncs all docs from a populated node to an empty peer. |
| `TestBranchableCollectionSync_WithDifferentDocsOnBothNodes_ShouldSync` | 88-155 | Branchable collection bi-directional sync merges disjoint docs from both peers. |
| `TestBranchableCollectionSync_ShouldNotSubscribe` | 157-238 | Branchable collection sync does not auto-subscribe to future docs added after sync. |
| `TestBranchableCollectionSync_WithNonBranchableCollection_ShouldError` | 240-260 | Syncing a non-branchable collection returns an error. |
| `TestBranchableCollectionSync_WithNonExistentCollection_ShouldError` | 262-283 | Syncing a non-existent collection index returns an out-of-range error. |

### `versions_test.go`

Tests that branchable collection sync correctly handles peers that have independently evolved divergent schema versions via `PatchCollection`.

| Test Function | Line | Description |
|---|---|---|
| `TestBranchableCollectionSync_WithBranchedVersionsAndDocs_ShouldSync` | 23-187 | Branchable collection syncs docs and branched schema versions across peers correctly. |

### `complex_network_test.go`

Multi-node network topology tests verifying that sync propagates through intermediate hops, that late-joining nodes receive all prior heads, and that all nodes converge to an identical DAG after sync.

| Test Function | Line | Description |
|---|---|---|
| `TestBranchableCollectionSync_WithMultipleDocsInComplexLinkedNetwork_ShouldSyncAll` | 27-148 | Branchable collection syncs all docs across a multi-hop linked five-node network. |
| `TestBranchableCollectionSync_WithMultipleDocumentHeadsReceivedFromPeers_ShouldSyncAll` | 150-232 | Node joining a synced peer pair receives all heads from both prior peers. |
| `TestBranchableCollectionSync_WithDocumentsFromPeers_ShouldHaveIdenticalDAG` | 234-372 | Fully-connected four-node branchable collection sync produces identical DAG on all nodes. |
| `TestBranchableCollectionSync_WithDocumentsFromPeersAndNewHeadAfterSync_ShouldHaveIdenticalDAG` | 374-557 | New doc added after multi-phase sync links all prior heads and yields identical DAG. |
