# Index: `tests/integration/collection_version/aggregates`

## Overview

This folder verifies that when a collection is registered via `AddCollection`, the GraphQL schema is updated to expose the correct aggregate fields (`COUNT`, `SUM`, `AVG`) both at the type level and at the top-level query root. Tests cover simple (empty) collections as well as collections with inline arrays of every supported scalar type (Boolean, Int, Float, String — both nillable and non-null variants), asserting that the introspected selector types and their filter/limit/offset/order input fields match the expected shapes.

## Test Index

### `inline_array_test.go`

Tests that inline scalar array fields produce correctly typed aggregate selector arguments (`CountSelector`, `NumericSelector`) with appropriate filter input types.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionAggregateInlineArrayAddsUsersCount` | 21-144 | Adding a collection with an inline integer array exposes COUNT aggregate with array and group selectors. |
| `TestCollectionVersionAggregateInlineArrayAddsUsersSum` | 146-267 | Adding a collection with an inline float array exposes SUM aggregate with float field selector args. |
| `TestCollectionVersionAggregateInlineArrayAddsUsersAverage` | 269-390 | Adding a collection with an inline integer array exposes AVG aggregate with array and group selectors. |
| `TestCollectionVersionAggregateInlineArrayAddsUsersNillableBooleanCountFilter` | 484-609 | COUNT selector for a nillable boolean inline array includes BooleanFilterArg with expected operators. |
| `TestCollectionVersionAggregateInlineArrayAddsUsersBooleanCountFilter` | 611-736 | COUNT selector for a non-null boolean inline array includes NotNullBooleanFilterArg with expected operators. |
| `TestCollectionVersionAggregateInlineArrayAddsUsersNillableIntegerCountFilter` | 738-887 | COUNT selector for a nillable integer inline array includes IntFilterArg with expected comparison operators. |
| `TestCollectionVersionAggregateInlineArrayAddsUsersIntegerCountFilter` | 889-1038 | COUNT selector for a non-null integer inline array includes NotNullIntFilterArg with expected comparison operators. |
| `TestCollectionVersionAggregateInlineArrayAddsUsersNillableFloatCountFilter` | 1040-1189 | COUNT selector for a nillable float inline array includes Float64FilterArg with expected comparison operators. |
| `TestCollectionVersionAggregateInlineArrayAddsUsersFloatCountFilter` | 1191-1340 | COUNT selector for a non-null float inline array includes NotNullFloat64FilterArg with expected comparison operators. |
| `TestCollectionVersionAggregateInlineArrayAddsUsersNillableStringCountFilter` | 1342-1491 | COUNT selector for a nillable string inline array includes StringFilterArg with expected string operators. |
| `TestCollectionVersionAggregateInlineArrayAddsUsersStringCountFilter` | 1493-1642 | COUNT selector for a non-null string inline array includes NotNullStringFilterArg with expected string operators. |

### `simple_test.go`

Tests that a bare collection (no custom fields) immediately exposes COUNT, SUM, and AVG aggregates with the correct selector argument structure.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionAggregateSimpleAddsUsersCount` | 21-116 | Adding a collection exposes COUNT aggregate with correct selector args on the type. |
| `TestCollectionVersionAggregateSimpleAddsUsersSum` | 118-315 | Adding a collection exposes SUM aggregate with correct numeric selector args on the type. |
| `TestCollectionVersionAggregateSimpleAddsUsersAverage` | 317-414 | Adding a collection exposes AVG aggregate with correct numeric selector args on the type. |

### `top_level_test.go`

Tests that the top-level query root gains COUNT, SUM, and AVG fields scoped to the new collection after it is registered.

| Test Function | Line | Description |
|---|---|---|
| `TestCollectionVersionAggregateTopLevelAddsCountGivenCollection` | 21-98 | Adding a collection exposes a top-level COUNT query field with correct selector args. |
| `TestCollectionVersionAggregateTopLevelAddsSumGivenCollection` | 100-207 | Adding a collection exposes a top-level SUM query field with correct numeric selector args. |
| `TestCollectionVersionAggregateTopLevelAddsAverageGivenCollection` | 209-316 | Adding a collection exposes a top-level AVG query field with correct numeric selector args. |
