# Index: `tests/integration/net/simple`

## Overview

This folder contains integration tests for simple P2P networking scenarios in DefraDB. It organises tests by networking topology: plain peer-to-peer sync, peer-plus-replicator (hybrid) topologies, and one-to-one replicator relationships. Together the subdirectories verify that documents and CRDT counter fields are created, updated, and deleted correctly across nodes connected in each of these configurations.

## Test Index

There are no `*_test.go` files directly in this directory. All tests live in the subdirectories listed below.

## Subdirectories

| Directory | Summary |
|---|---|
| [`peer/`](peer/INDEX.md) | Tests bidirectional P2P peer sync for document add, update, delete, schema migrations, chained topologies, and node restarts. |
| [`peer_replicator/`](peer_replicator/crdt/INDEX.md) | Tests hybrid three-node topologies where one node acts as both a peer and a replicator source for CRDT counter fields. |
| [`replicator/`](replicator/crdt/INDEX.md) | Tests one-to-one replicator relationships propagating CRDT counter (pcounter, pncounter) field updates from source to target. |
