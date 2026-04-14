# Index: `tests/integration/query/one_to_many_to_many`

## Overview

This folder tests queries over chained one-to-many-to-many relationship schemas, where a parent type relates to a child type via one-to-many, and that child type in turn relates to a grandchild type via another one-to-many. The tests verify that nested join traversal correctly links documents at each level and returns the right associated records.

## Test Index

### `joins_test.go`

Tests that a chained one-to-many-to-many query traverses both levels of the join and returns correctly linked nested documents.

| Test Function | Line | Description |
|---|---|---|
| `TestOneToManyToManyJoinsAreLinkedProperly` | 21-276 | Chained one-to-many-to-many query verifies nested join traversal returns correct linked documents. |
