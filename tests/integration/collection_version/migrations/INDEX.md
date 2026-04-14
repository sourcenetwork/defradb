# Index: `tests/integration/collection_version/migrations`

## Overview

This directory tests the behaviour of lens-based schema migrations applied to collection versions in DefraDB. The tests cover how migrations transform documents at query time across single and multiple schema versions, including interactions with indexes, transactions, P2P replication, node restarts, collection branching, and explicit active-version management.

## Subdirectories

| Directory | Summary |
|---|---|
| [`query/`](query/INDEX.md) | Tests for migration behaviour during query execution, covering forward and inverse migrations, field mutations, index reindexing, P2P replication, transactions, and node restarts. |
