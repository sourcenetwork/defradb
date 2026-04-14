# Index: `tests/integration/net/simple/replicator/crdt`

## Overview

This folder contains integration tests for one-to-one P2P replicator behaviour when collections use CRDT counter field types (`pcounter` and `pncounter`). Each test verifies that a replicator established between a source and a target node correctly propagates counter increment updates, confirming that the accumulated CRDT value is consistent on both nodes after synchronisation.

## Test Index

### `pcounter_test.go`

Tests that a positive-only CRDT counter (`pcounter`) field increment is correctly propagated from a replicator source node to its target.

| Test Function | Line | Description |
|---|---|---|
| `TestP2POneToOneReplicatorUpdate_PCounter_NoError` | 24-80 | Replicator propagates a pcounter increment update from source node to target node. |

### `pncounter_test.go`

Tests that a positive-negative CRDT counter (`pncounter`) field increment is correctly propagated from a replicator source node to its target.

| Test Function | Line | Description |
|---|---|---|
| `TestP2POneToOneReplicatorUpdate_PNCounter_NoError` | 24-80 | Replicator propagates a pncounter increment update from source node to target node. |
