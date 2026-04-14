# Index: `tests/integration/query/one_to_many_multiple`

## Overview

This folder tests queries over schemas where a single parent type (Author) holds multiple distinct one-to-many relationships simultaneously — for example, both a `books` relation and an `articles` relation. The tests verify that aggregate functions (AVG, COUNT, SUM) and filters applied across those multiple joins produce correct, independently computed results.

## Test Index

### `with_average_filter_test.go`

Tests that AVG correctly handles mixed filter conditions when spanning two one-to-many joins on the same parent type.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToManyMultipleWithAverageOnMultipleJoinsWithAndWithoutFilter` | 21-145 | AVG across two one-to-many joins where only one join has a filter. |
| `TestQueryOneToManyMultipleWithAverageOnMultipleJoinsWithFilters` | 147-271 | AVG across two one-to-many joins where both joins have independent filters applied. |

### `with_average_test.go`

Tests that AVG is correctly computed over fields from two separate one-to-many joins without filtering.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToManyMultipleWithAverageOnMultipleJoins` | 21-145 | AVG aggregated across two distinct one-to-many join relations simultaneously. |

### `with_count_filter_test.go`

Tests that COUNT correctly handles mixed filter conditions when spanning two one-to-many joins on the same parent type.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToManyMultipleWithCountOnMultipleJoinsWithAndWithoutFilter` | 21-145 | COUNT across two one-to-many joins where only one join has a filter. |
| `TestQueryOneToManyMultipleWithCountOnMultipleJoinsWithFilters` | 147-271 | COUNT across two one-to-many joins where both joins have independent filters applied. |

### `with_count_test.go`

Tests that COUNT is correctly computed over two separate one-to-many joins, individually and combined.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToManyMultipleWithCount` | 21-140 | COUNT on each of two separate one-to-many joins with aliased result fields. |
| `TestQueryOneToManyMultipleWithCountOnMultipleJoins` | 142-266 | COUNT aggregated across two distinct one-to-many join relations simultaneously. |

### `with_multiple_filter_test.go`

Tests that filtering on both the parent type and multiple one-to-many child relations in a single query works correctly.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToManyMultipleWithMultipleManyFilters` | 21-138 | Filter on parent and two separate one-to-many relations in the same query. |

### `with_sum_filter_test.go`

Tests that SUM correctly handles mixed filter conditions when spanning two one-to-many joins on the same parent type.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToManyMultipleWithSumOnMultipleJoinsWithAndWithoutFilter` | 21-145 | SUM across two one-to-many joins where only one join has a filter. |
| `TestQueryOneToManyMultipleWithSumOnMultipleJoinsWithFilters` | 147-271 | SUM across two one-to-many joins where both joins have independent filters applied. |

### `with_sum_test.go`

Tests that SUM is correctly computed over fields from two separate one-to-many joins without filtering.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToManyMultipleWithSumOnMultipleJoins` | 21-145 | SUM aggregated across two distinct one-to-many join relations simultaneously. |
