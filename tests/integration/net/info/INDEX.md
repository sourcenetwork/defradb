# Index: `tests/integration/net/info`

## Overview

This folder contains integration tests for the P2P network information API, specifically the `ActivePeers` endpoint. The tests verify that a node correctly reports its connected peers — returning an error when no P2P system is configured, an empty list when configured but isolated, and accurate peer addresses as connections are established between two or more nodes.

## Test Index

### `connect_peers_test.go`

Tests that verify the active peers list is accurate under various connection configurations, from no P2P system to multi-node mesh networks.

| Test Function | Line | Description |
|---|---|---|
| `TestNetInfoPeers_NoP2PConfigured` | 25-47 | Calling ActivePeers returns an error when no P2P system is configured. |
| `TestNetInfoPeers` | 49-62 | A P2P-enabled node with no connections reports an empty active peers list. |
| `TestNetInfoConnectPeers` | 64-82 | Connecting two peers makes the remote peer appear in the active peers list. |
| `TestNetInfoConnectMultiplePeers` | 84-130 | Connecting three nodes causes all peers to appear in each node's active peers list. |
