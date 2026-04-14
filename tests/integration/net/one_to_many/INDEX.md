# Index: `tests/integration/net/one_to_many`

## Overview

This directory contains P2P synchronization tests for one-to-many relational document graphs in DefraDB. Both peer and replicator sync modes are covered, with tests verifying that relational links between parent and child documents are correctly propagated across nodes, including edge cases where related documents are absent on the receiving node.

## Subdirectories

| Directory | Summary |
|---|---|
| [`peer/`](peer/INDEX.md) | Tests peer-based sync of one-to-many relational documents, including the case where a linked related document does not yet exist on the receiving peer. |
| [`replicator/`](replicator/INDEX.md) | Tests replicator-based sync of one-to-many relational documents, verifying that both sides of the relation are propagated and the relational link is preserved and queryable on the target node. |
