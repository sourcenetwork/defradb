# Index: `tests/integration/query/one_to_two_many`

## Overview

This folder contains integration tests for querying a schema where a single type (Author) participates in two distinct one-to-many relationships with another type (Book), distinguished by named relation directives. The tests cover querying from both the one side and the many side, mixing named and unnamed relations, and applying independent ordering to each relation collection.

## Test Index

### `simple_test.go`

Tests that verify correct result traversal when a type holds two separate named one-to-many relations, queried from both sides and combined with additional unnamed relations.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToTwoManyWithNilUnnamedRelationship_FromOneSide` | 21-143 | Query two named one-to-many relations on one type from the one side. |
| `TestQueryOneToTwoManyWithNilUnnamedRelationship_FromManySide` | 145-271 | Query two named one-to-many relations on one type from the many side. |
| `TestQueryOneToTwoManyWithNamedAndUnnamedRelationships` | 273-435 | Query a type with two named many-to-one relations plus a third unnamed relation. |
| `TestQueryOneToTwoManyWithNamedAndUnnamedRelationships_FromManySide` | 437-599 | Query two named relations and one unnamed relation from the many side. |

### `with_order_test.go`

Tests that verify independent ordering can be applied to each of the two one-to-many relation collections on the same type.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToTwoManyWithOrder` | 21-144 | Query two separate one-to-many relations each with independent ordering applied. |
