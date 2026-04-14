# Index: `tests/integration`

## Overview

This directory is the root of DefraDB's integration test suite. Tests are organized into subdirectories by feature area (queries, mutations, networking, encryption, ACP, and more), each exercised against multiple client types (Go, HTTP, CLI) through a shared test harness. The harness, implemented in the support files at this level (`utils.go`, `test_case.go`, `client.go`, etc.), drives test execution declaratively via `TestCase` structs that describe actions, expected results, and client-type constraints. The two direct test files here cover the harness infrastructure itself: retry semantics for flaky tests and the correctness of result-matching utilities.

## Subdirectories

| Directory | Summary |
|---|---|
| [`acp/`](acp/INDEX.md) | Tests the full lifecycle of DefraDB's Access Control Policy system, covering both Document Access Control (DAC) and Node Access Control (NAC) across policy registration, per-document CRUD gating, P2P replication, and the delegated admin relation. |
| [`backup/`](backup/INDEX.md) | Tests backup export and import operations across flat and relational schemas (one-to-one, one-to-many, self-referential), verifying correct JSON output, collection filtering, error handling, and relation restoration. |
| [`collection/`](collection/INDEX.md) | Tests collection management including dynamic schema evolution via SDL and JSON patch, foreign-object relation creation, and the `Truncate` operation across branchable, DAC-protected, indexed, and materialized-view scenarios. |
| [`collection_version/`](collection_version/INDEX.md) | Tests the `AddCollection` and `PatchCollection` APIs across the full collection-version lifecycle, including scalar/relational/CRDT field definitions, GraphQL introspection types, branching, active-version switching, and lens-based migrations. |
| [`encryption/`](encryption/INDEX.md) | Tests at-rest document encryption covering key generation, encrypted commit deltas for LWW and counter CRDTs, decryption on query, and interaction with P2P sync, ACP, and secondary indexes. |
| [`explain/`](explain/INDEX.md) | Tests the `@explain` query directive across all four modes (debug, default, execute, simple), verifying plan-tree structure, node attributes, and runtime statistics for queries, mutations, aggregations, and joins. |
| [`index/`](index/INDEX.md) | Tests the index subsystem covering index creation, the full range of filter operators, unique and composite indexes, relation-based filtering and ordering, JSON path indexing, and index maintenance on updates. |
| [`issues/`](issues/INDEX.md) | Regression tests for specific filed GitHub issues, guarding against reintroduction of fixed bugs such as pncounter float overflow and JSON serialization failures. |
| [`mutation/`](mutation/INDEX.md) | Tests all mutation operations — add, update, delete, upsert, and mixed — including CRDT semantics, schema constraints, counter fields, relational mutations, and transactional isolation. |
| [`net/`](net/INDEX.md) | Tests the P2P networking layer including peer and replicator sync, collection and document subscription events, branchable-collection synchronization, and network resilience across various topologies. |
| [`node/`](node/INDEX.md) | Tests node-level operations, currently verifying that each node in a multi-node setup correctly exposes its own unique cryptographic identity. |
| [`query/`](query/INDEX.md) | Tests all GraphQL query operations across simple and relational schemas (one-to-one, one-to-many, many-to-many, chained variants), covering filters, ordering, pagination, aggregates, groupBy, fragments, versioned queries, and vector similarity. |
| [`searchable_encryption/`](searchable_encryption/INDEX.md) | Tests the searchable encryption feature, including encrypted index creation and deletion, P2P replication of SE metadata, ACP-gated access, and equality lookups on encrypted fields. |
| [`signature/`](signature/INDEX.md) | Tests block-signing of CRDT commit blocks using node and client identities, covering key types, ACP-gated verification, P2P propagation, branchable collections, and direct signature verification. |
| [`subscription/`](subscription/INDEX.md) | Tests GraphQL subscriptions verifying event delivery for add, update, and delete mutations with field-value and docID filters, counter CRDTs, commit-level subscriptions, and node-close behaviour. |
| [`txn/`](txn/INDEX.md) | Tests transactional isolation for every database operation type, verifying that changes are invisible to concurrent transactions until committed and that in-flight changes are visible within the same transaction. |
| [`view/`](view/INDEX.md) | Tests DefraDB's view system — virtual collections backed by a query and optional lens transform — across simple, one-to-one, and one-to-many topologies with materialization, directive semantics, and transform persistence. |
