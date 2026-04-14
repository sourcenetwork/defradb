# Index: `tests/integration/net`

## Overview

This directory contains integration tests for DefraDB's P2P networking layer. The tests cover the full breadth of peer-to-peer operations: querying the active peers list via the network info API, peer and replicator synchronization of plain and relational document graphs (one-to-many), collection and document subscription events (join/left peer events across the global doc-sync topic, per-collection topics, and per-document topics), simple peer and replicator topologies for document lifecycle operations, and explicit synchronization primitives for branchable collections and collection versions. Together they verify correct data propagation, event delivery, schema transfer, and network resilience across a wide range of connection configurations.

## Subdirectories

| Directory | Summary |
|---|---|
| [`info/`](info/INDEX.md) | Tests the `ActivePeers` network info endpoint, verifying error behaviour when P2P is unconfigured and correct peer reporting as connections are established between one, two, or three nodes. |
| [`one_to_many/`](one_to_many/INDEX.md) | Tests P2P synchronization of one-to-many relational document graphs in both peer and replicator modes, including edge cases where linked related documents are absent on the receiving node. |
| [`peer_events/`](peer_events/INDEX.md) | Tests the P2P peer events system, verifying join and left events on the global doc-sync topic, per-collection topics, and per-document topics as peers connect, subscribe, and unsubscribe. |
| [`simple/`](simple/INDEX.md) | Tests simple P2P networking topologies — plain peer sync, hybrid peer-plus-replicator, and one-to-one replicator — covering document creation, updates, deletes, CRDT counter fields, schema migrations, chained topologies, and node restarts. |
| [`sync/`](sync/INDEX.md) | Tests explicit P2P synchronization primitives (`SyncBranchableCollection`, `SyncCollectionVersions`), covering DAG head exchange, schema transfer, divergent versions, multi-hop topologies, and activation of synced collection versions. |
