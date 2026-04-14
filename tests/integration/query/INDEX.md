# Index: `tests/integration/query`

## Overview

This directory contains integration tests for all GraphQL query operations in DefraDB. Coverage spans simple (non-relational) scalar queries, inline array types, JSON field filtering, commit history via `_commits`, and the full matrix of relational schema topologies — one-to-one, one-to-many, many-to-many, and chained multi-hop variants including one-to-many-to-many, one-to-many-to-one, one-to-one-to-many, one-to-one-to-one, one-to-one-multiple, one-to-many-multiple, and one-to-two-many. Across all subdirectories the tests validate field retrieval, filters, ordering, pagination (limit/offset), groupBy, aggregate functions (COUNT, SUM, AVG, MIN, MAX), GraphQL fragments and variables, transactional and versioned (`_version`) queries, branchable collections, CRDT counter types, vector similarity queries, and null/error input handling.

## Subdirectories

| Directory | Summary |
|---|---|
| [`commits/`](commits/INDEX.md) | Tests for the `_commits` query covering basic commit retrieval, filtering by CID/docID/fieldName/depth, grouping, ordering, pagination, link traversal, deletion semantics, compound filters, and branchable collections in multi-node peer scenarios. |
| [`inline_array/`](inline_array/INDEX.md) | Tests for querying inline array fields across all supported types (boolean, integer, float, string) in non-null and nillable variants, including aggregates, element-level quantifier filters, full-array equality filters, and groupBy. |
| [`json/`](json/INDEX.md) | Tests for querying JSON-typed fields with the full range of filter operators (equality, set membership, numeric comparison, string pattern matching, array quantifiers) across all JSON value types and nested object paths. |
| [`many_to_many/`](many_to_many/INDEX.md) | Tests for many-to-many relationship queries via an explicit junction collection, covering traversal in both directions with filters and ordering. |
| [`one_to_many/`](one_to_many/INDEX.md) | Tests for one-to-many queries across both relation directions with filters, ordering, limit, offset, grouping, aggregates, CID/docID lookups, alias filters, and schema-level constraints. |
| [`one_to_many_multiple/`](one_to_many_multiple/INDEX.md) | Tests for schemas where a single parent holds multiple distinct one-to-many relationships, verifying that AVG, COUNT, and SUM aggregates with independent filters on each join produce correct results. |
| [`one_to_many_to_many/`](one_to_many_to_many/INDEX.md) | Tests for chained one-to-many-to-many queries verifying that nested join traversal correctly links documents at both levels. |
| [`one_to_many_to_one/`](one_to_many_to_one/INDEX.md) | Tests for a three-collection Author → Book → Publisher schema exercising join correctness, deep filter propagation, multi-key ordering, SUM aggregations, and limit/direction combinations across two relation levels. |
| [`one_to_one/`](one_to_one/INDEX.md) | Tests for one-to-one relation queries covering both traversal directions, nil relations, relation ID field access, filtering, ordering, groupBy, fragment spreading, and `_version` metadata. |
| [`one_to_one_multiple/`](one_to_one_multiple/INDEX.md) | Tests for schemas where a single collection holds two independent one-to-one relationships, verifying resolution under all primary/secondary ownership combinations and cyclic relation schema creation. |
| [`one_to_one_to_many/`](one_to_one_to_many/INDEX.md) | Tests for mixed one-to-one and one-to-many chained queries across three types (Indicator → Observable → Observation) in all traversal directions and primary-side placements. |
| [`one_to_one_to_one/`](one_to_one_to_one/INDEX.md) | Tests for three-hop one-to-one chain queries (Publisher → Book → Author) under different primary/secondary configurations and with nested ordering. |
| [`one_to_two_many/`](one_to_two_many/INDEX.md) | Tests for schemas where a single type participates in two distinct named one-to-many relationships, queried from both sides with independent ordering and mixed named/unnamed relations. |
| [`simple/`](simple/INDEX.md) | Tests for basic non-relational collection queries covering field retrieval, aliases, pagination, ordering, aggregates, groupBy, GraphQL fragments and variables, vector similarity, branchable versioning, embedded commit metadata, CRDT counters, and exhaustive filter operator coverage. |
