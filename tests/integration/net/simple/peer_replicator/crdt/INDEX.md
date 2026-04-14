# Index: `tests/integration/net/simple/peer_replicator/crdt`

## Overview

This folder contains integration tests for P2P peer-replicator behaviour when collections use CRDT counter field types (`pcounter` and `pncounter`). Each test verifies that a three-node topology — one source node with a replicator relationship to a third node and a peer-only relationship to a second node — correctly scopes document and update propagation: the replicator target receives all data while the peer-only node only receives what a normal peer subscription delivers.

## Test Index

### `pcounter_test.go`

Tests that a positive-only CRDT counter (`pcounter`) field is correctly synced and accumulated across peer-replicator topologies.

| Test Function | Line | Description |
|---|---|---|
| `TestP2PPeerReplicatorWithAdd_PCounter_NoError` | 25-120 | Peer-replicator syncs a new pcounter document to the replicator target but not to a peer-only node. |
| `TestP2PPeerReplicatorWithUpdate_PCounter_NoError` | 122-186 | Peer-replicator propagates a pcounter increment update to all subscribed nodes. |

### `pncounter_test.go`

Tests that a positive-negative CRDT counter (`pncounter`) field is correctly synced and accumulated across peer-replicator topologies.

| Test Function | Line | Description |
|---|---|---|
| `TestP2PPeerReplicatorWithAdd_PNCounter_NoError` | 25-122 | Peer-replicator syncs a new pncounter document to the replicator target but not to a peer-only node. |
| `TestP2PPeerReplicatorWithUpdate_PNCounter_NoError` | 124-188 | Peer-replicator propagates a pncounter increment update to all subscribed nodes. |
