# Index: `tests/integration/view`

## Overview

This directory contains integration tests for DefraDB's view system — a feature that allows users to define virtual collections backed by a query, an optional lens transform pipeline, and an optional materialized cache. The tests span three categories of view topology: simple single-collection views (cacheless and materialized), views over one-to-one relational schemas, and views over one-to-many relational schemas. Together they verify basic querying, SDL/query field-set rules, alias handling, filter composition, GraphQL introspection, @constraints/@default/@embedding directive semantics in view context, CID-based transform references, cardinality-changing lens transforms, aggregate transforms, and the persistence of transform configuration across node restarts.

## Subdirectories

| Directory | Summary |
|---|---|
| [`simple/`](simple/INDEX.md) | Tests for single-collection cacheless and materialized views, covering basic querying, field subset rules, filter composition, aliases, directive behaviour (@constraints, @default, @embedding), GraphQL introspection, and a broad range of lens transform scenarios including multi-step, cardinality-changing, CID-based, and aggregate transforms. |
| [`one_to_one/`](one_to_one/INDEX.md) | Tests for views over one-to-one relations, covering self-referential view schemas, duplicate embedded schema errors, persistence of embedded interface types across schema updates and node restarts, and lens transforms on the outer type. |
| [`one_to_many/`](one_to_many/INDEX.md) | Tests for views over one-to-many relations, covering basic querying, mixed SDL, alias support on outer and inner types, COUNT aggregates, GraphQL introspection, error cases for querying from the inner side, stacked views, and lens transforms on the outer type or synthesising inner embedded documents. |
