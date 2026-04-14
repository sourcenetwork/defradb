# Index: `tests/integration/query/many_to_many`

## Overview

This folder contains integration tests for many-to-many relationship queries in DefraDB. The tests model many-to-many associations through an explicit junction collection (e.g. `Enrollment` joining `Student` and `Course`), and verify that GraphQL queries can traverse these relationships correctly — both by querying directly from the junction collection with filters and ordering, and by querying from a primary collection and navigating through nested relation fields to the other side.

## Test Index

### `simple_test.go`

Tests querying a many-to-many relationship directly from the junction collection, filtering and ordering results in both traversal directions.

| Test Function | Line | Description |
|---|---|---|
| `TestManyToMany_QueryFromJoinCollection_ShouldSucceed` | 21-121 | Many-to-many query via junction collection filtered and ordered in both directions. |

### `with_nested_query_test.go`

Tests querying a many-to-many relationship by starting from a primary collection and traversing through nested enrollment relation fields to reach the related collection.

| Test Function | Line | Description |
|---|---|---|
| `TestManyToMany_QueryFromSecondary_Succeeds` | 21-109 | Many-to-many nested query traversing from primary collection through enrollment to related collection. |
