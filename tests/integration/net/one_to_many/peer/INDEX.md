# Index: `tests/integration/net/one_to_many/peer`

## Overview

This folder contains P2P peer synchronization tests for one-to-many relational document graphs. The tests verify that peer nodes correctly handle document updates that cross relational boundaries, including edge cases where a related document exists on the source node but has not been synced to the destination peer.

## Test Index

### `with_add_update_test.go`

Tests that updating a document to link it to a related document propagates correctly over a peer connection even when the related document is absent on the receiving node.

| Test Function | Line | Description |
|---|---|---|
| `TestP2POneToManyPeerWithAddUpdateLinkingSyncedDocToUnsyncedDoc` | 26-131 | Peer syncs a book update linking it to an author that does not exist on the peer node. |
