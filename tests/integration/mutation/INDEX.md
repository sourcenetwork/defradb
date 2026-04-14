# Index: `tests/integration/mutation`

## Overview

This directory contains integration tests for all mutation operations in DefraDB. Coverage spans the four core mutation types — `add` (document creation), `update` (field modification), `delete` (soft deletion), and `upsert` (insert-or-update) — as well as a `mix` subdirectory for cross-mutation transactional interaction tests and a `special` subdirectory for invalid or edge-case mutation inputs. Together these tests validate correct document lifecycle management, error handling, CRDT semantics, schema-level constraints (`@constraints`, `@default`, `@embedding`), counter field types (`pcounter`/`pncounter`), relational field mutations across one-to-one and one-to-many schemas, and transactional isolation guarantees.

## Subdirectories

| Directory | Summary |
|---|---|
| [`add/`](add/INDEX.md) | Tests for the `add` (create/insert) mutation, covering basic document creation, error cases, `@default` directive resolution, GQL variable support, commit CID retrieval, schema constraints, CRDT counters, vector embedding generation, and relational field kinds. |
| [`delete/`](delete/INDEX.md) | Tests for the `delete` mutation, covering filter-based and docID-based deletion, malformed request errors, null argument handling, transactional isolation, and relational soft-deletion across one-to-many and one-to-one-to-one chains. |
| [`mix/`](mix/INDEX.md) | Tests for mixed add, update, and delete mutations within and across concurrent transactions, verifying isolation semantics, conflict detection, and correct document state after commit. |
| [`special/`](special/INDEX.md) | Tests for edge-case and invalid mutation scenarios, including mutations that reference operation names not recognized by the generated GraphQL schema. |
| [`update/`](update/INDEX.md) | Tests for the `update` mutation, covering single and multi-docID updates, filter-based updates, `_version` CID retrieval, default-value non-overwrite semantics, underscored collection names, null argument handling, error cases, schema constraints, CRDT counters, vector embedding regeneration, and relational field kinds. |
| [`upsert/`](upsert/INDEX.md) | Tests for the `upsert` mutation, covering insert-on-miss, update-on-match, `DateTime` with `UTC_NOW`, null input errors, multiple-match errors, and unique index violations. |
