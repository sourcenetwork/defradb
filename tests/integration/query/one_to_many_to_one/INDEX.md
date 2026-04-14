# Index: `tests/integration/query/one_to_many_to_one`

## Overview

This folder contains integration tests for a three-collection schema (Author → Book → Publisher) that exercises one-to-many (Author/Book) and one-to-one (Book/Publisher) relationships simultaneously. Tests cover basic join correctness, deep filter propagation through two relation levels, multi-key ordering with depth greater than one, SUM aggregations combined with ordering and filtering across the nested join chain, and combinations of these capabilities with limit and direction switching.

## Test Index

### `joins_test.go`

Tests that the full join chain — author to multiple books, each book optionally linked to a publisher — is resolved and returned correctly.

| Test Function | Line | Description |
|---|---|---|
| `TestOneToManyToOneJoinsAreLinkedProperly` | 21-242 | One-to-many-to-one joins resolve nested publisher links correctly for each author. |

---

### `simple_test.go`

Tests basic querying of books with their associated author and publisher across the one-to-many-to-one schema.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToOneRelations` | 21-148 | Querying Book returns both author and publisher relations, including nil publisher. |

---

### `with_filter_test.go`

Tests filtering authors and books using conditions that span two relation levels, including compound logical operators.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryComplexWithDeepFilterOnRenderedChildren` | 21-130 | Filtering authors by nested publisher yearOpened returns only matching author with books. |
| `TestOneToManyToOneWithSumOfDeepFilterSubTypeOfBothDescAndAsc` | 132-173 | SUM on book rating filtered by publisher yearOpened works with two independent filter conditions. |
| `TestOneToManyToOneWithSumOfDeepFilterSubTypeAndDeepOrderBySubtypeOppositeDirections` | 175-224 | SUM with deep filter and a filtered sub-selection on book return correct independent results. |
| `TestOneToManyToOneWithTwoLevelDeepFilter` | 226-298 | Two-level deep filter on author returns all books with publisher yearOpened in results. |
| `TestOneToManyToOneWithCompoundOperatorInFilterAndRelation` | 300-380 | Compound _and/_or filters on author with nested publisher conditions match correct authors. |
| `TestOneToManyToOneWithCompoundOperatorInSubFilterAndRelation` | 382-424 | Compound filter on author and nested book sub-filter both apply independently and correctly. |

---

### `with_order_limit_test.go`

Tests ordering books by a nested publisher field with a limit applied per author.

| Test Function | Line | Description |
|---|---|---|
| `TestOneToManyToOneDeepOrderBySubTypeOfBothDescAndAsc` | 22-82 | Ordering books by nested publisher yearOpened DESC and ASC with limit 1 returns correct books. |

---

### `with_order_test.go`

Tests multi-key ordering of books where one key is a direct field and the other is a nested relation field.

| Test Function | Line | Description |
|---|---|---|
| `TestMultipleOrderByWithDepthGreaterThanOne` | 21-92 | Books ordered by rating ASC then publisher yearOpened DESC return correctly sorted results. |
| `TestMultipleOrderByWithDepthGreaterThanOneOrderSwitched` | 98-169 | Books ordered by publisher yearOpened DESC then rating ASC places nil-publisher book last. |

---

### `with_sum_order_limit_test.go`

Tests SUM aggregations that combine deep ordering on publisher fields with a limit, including cases where the SUM order and the sub-selection order differ.

| Test Function | Line | Description |
|---|---|---|
| `TestOneToManyToOneWithSumOfDeepOrderBySubTypeAndDeepOrderBySubtypeDescDirections` | 21-71 | SUM and book sub-selection both ordered by publisher yearOpened DESC with limit 2 match. |
| `TestOneToManyToOneWithSumOfDeepOrderBySubTypeAndDeepOrderBySubtypeAscDirections` | 77-130 | SUM and book sub-selection both ordered by publisher yearOpened ASC with limit 2 match. |
| `TestOneToManyToOneWithSumOfDeepOrderBySubTypeOfBothDescAndAsc` | 136-176 | Two SUM aggregates with opposite publisher yearOpened orderings and limit 2 produce correct totals. |
| `TestOneToManyToOneWithSumOfDeepOrderBySubTypeAndDeepOrderBySubtypeOppositeDirections` | 182-234 | SUM ordered DESC and book sub-selection ordered ASC with limit 2 return independent correct results. |

---

### `with_sum_test.go`

Tests SUM aggregations combining inline array fields and one-to-many relation fields in the same query.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryWithSumOnInlineAndSumOnOneToManyField` | 21-119 | SUM on an inline integer array and SUM on a one-to-many relation field both return correct totals. |
