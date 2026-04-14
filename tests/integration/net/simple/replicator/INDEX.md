# Index: `tests/integration/net/simple/replicator`

## Overview

This directory contains integration tests for one-to-one P2P replicator behaviour in DefraDB. Tests verify that a replicator configured between a source and a target node correctly propagates documents and field updates, with coverage focused on CRDT field types that require accurate accumulation semantics.

## Subdirectories

| Directory | Summary |
|---|---|
| [`crdt/`](crdt/INDEX.md) | Tests one-to-one replicator sync for CRDT counter field types (`pcounter` and `pncounter`), confirming that counter increment updates are propagated from source to target with consistent accumulated values. |
