# Index: `tests/integration/net/simple/peer/crdt`

## Overview

This folder tests the convergence of CRDT (Conflict-free Replicated Data Type) field semantics across P2P peer nodes. It covers Last-Write-Wins (LWW), positive counter (PCounter), and positive-negative counter (PNCounter) CRDT types, verifying that concurrent updates from multiple peers are accumulated and resolved correctly after synchronisation.

## Test Index

### `lww_test.go`

Tests that concurrent LWW updates across peers either preserve all distinct field changes (different fields) or converge to a single winning value (same field).

| Test Function | Line | Description |
|---|---|---|
| `TestP2PUpdate_WithLWWConcurrentDifferentFields_BothFieldsPreserved` | 24-118 | LWW CRDT concurrent updates to different fields on three peers all converge. |
| `TestP2PUpdate_WithLWWConcurrentSameField_ConvergesToSameValue` | 120-208 | LWW CRDT concurrent updates to the same field across three peers converge to one value. |

### `pcounter_test.go`

Tests that PCounter CRDT fields correctly accumulate increments from simultaneous peer updates across two- and three-node topologies.

| Test Function | Line | Description |
|---|---|---|
| `TestP2PUpdate_WithPCounter_NoError` | 25-85 | PCounter field increments are accumulated and synced correctly across two peers. |
| `TestP2PUpdate_WithPCounterThreeNodeSimultaneousUpdate_NoError` | 87-178 | PCounter simultaneous increments from three peers are all summed correctly. |
| `TestP2PUpdate_WithPCounterSimultaneousUpdate_NoError` | 180-251 | PCounter simultaneous increments from two peers are both accumulated after sync. |

### `pncounter_test.go`

Tests that PNCounter CRDT fields correctly accumulate both positive and negative increments from simultaneous peer updates across two- and three-node topologies.

| Test Function | Line | Description |
|---|---|---|
| `TestP2PUpdate_WithPNCounter_NoError` | 25-85 | PNCounter field increments are accumulated and synced correctly across two peers. |
| `TestP2PUpdate_WithPNCounterThreeNodeSimultaneousUpdate_NoError` | 87-179 | PNCounter simultaneous increments and decrements from three peers are summed correctly. |
| `TestP2PUpdate_WithPNCounterSimultaneousUpdate_NoError` | 181-252 | PNCounter simultaneous increments from two peers are both accumulated after sync. |
