# Index: `tests/integration/explain`

## Overview

This directory contains integration tests for DefraDB's `@explain` query directive across all four supported explain modes. The `debug/` group checks that the structural plan-tree shape (node names and hierarchy only) is correct for every query and mutation pattern. The `default/` group verifies the same plan trees with full node attributes — filters, prefixes, collection IDs, field names, and aggregation sources. The `execute/` group runs queries for real and validates runtime statistics such as iteration counts, document and field fetches, index fetches, and filter-match counts. The `simple/` group covers the logical plan shape including collection metadata and storage prefixes without execution statistics.

## Subdirectories

| Directory | Summary |
|---|---|
| [`debug/`](debug/INDEX.md) | Tests for `@explain(type: debug)`, asserting the structural node-name hierarchy of query plan trees for queries, mutations, aggregations, joins, grouping, ordering, limiting, and views — without node-specific attributes. |
| [`default/`](default/INDEX.md) | Tests for `@explain` (default mode), asserting full plan-tree node attributes — filters, collection prefixes, field names, aggregation sources — for queries, mutations, aggregations, joins, grouping, ordering, and limiting. |
| [`execute/`](execute/INDEX.md) | Tests for `@explain(type: execute)`, asserting actual runtime statistics (iteration counts, docFetches, fieldFetches, indexFetches, filterMatches) returned after executing queries and mutations for real. |
| [`simple/`](simple/INDEX.md) | Tests for `@explain(type: simple)`, asserting the logical plan shape including selectNode, scanNode, collection name, and storage prefixes, and verifying that index usage eliminates redundant plan nodes such as orderNode. |
