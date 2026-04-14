# Index: `tests/integration/query/one_to_one_to_one`

## Overview

This folder contains integration tests for querying three-type chains connected by one-to-one relationships (e.g. Publisher → Book → Author). The tests verify that traversal works correctly across different combinations of primary and secondary relation sides, and that nested ordering on a three-hop chain produces the expected results.

## Test Index

### `simple_test.go`

Tests that verify basic three-hop one-to-one chain traversal under different primary/secondary configurations for each relation side.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToOneToOne` | 21-124 | Query a three-hop one-to-one chain from Publisher through Book to Author. |
| `TestQueryOneToOneToOneSecondaryThenPrimary` | 126-229 | Query a three-hop chain where the first relation is secondary and the second is primary. |
| `TestQueryOneToOneToOnePrimaryThenSecondary` | 231-333 | Query a three-hop chain where the first relation is primary and the second is secondary. |
| `TestQueryOneToOneToOneSecondary` | 335-438 | Query a three-hop one-to-one chain where all relations are secondary. |

### `with_order_test.go`

Tests that verify ordering by a deeply nested relation field across a three-hop one-to-one chain.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToOneToOneWithNestedOrder` | 21-123 | Query a three-hop one-to-one chain ordered by a nested relation field. |
