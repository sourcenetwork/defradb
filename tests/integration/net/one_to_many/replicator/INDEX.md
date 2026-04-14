# Index: `tests/integration/net/one_to_many/replicator`

## Overview

This folder contains replicator-based synchronization tests for one-to-many relational document graphs. The tests verify that a configured replicator correctly propagates both the parent and child sides of a one-to-many relation from the source node to the target node, and that the relation is preserved and queryable after sync.

## Test Index

### `with_add_test.go`

Tests that a replicator correctly syncs newly added documents on both sides of a one-to-many relation and maintains the relational link on the target node.

| Test Function | Line | Description |
|---|---|---|
| `TestP2POneToManyReplicator` | 24-89 | Replicator propagates both sides of a one-to-many relation to the target node. |
