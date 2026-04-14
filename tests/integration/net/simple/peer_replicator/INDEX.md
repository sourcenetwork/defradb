# Index: `tests/integration/net/simple/peer_replicator`

## Overview

This directory contains integration tests for combined peer-replicator network topologies in DefraDB. Tests use a three-node setup where one source node maintains a replicator relationship with a third node and a peer-only relationship with a second node, verifying that replication boundaries are respected for different CRDT field types.

## Subdirectories

| Directory | Summary |
|---|---|
| [`crdt/`](crdt/INDEX.md) | Tests peer-replicator sync behaviour for CRDT counter field types (`pcounter` and `pncounter`), verifying that the replicator target receives all data while a peer-only node receives only what its subscription delivers. |
