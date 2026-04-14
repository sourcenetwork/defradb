# Index: `tests/integration/query/one_to_one_to_many`

## Overview

This folder contains integration tests for querying across a mixed one-to-one and one-to-many relation chain in DefraDB. The schema links three types — `Indicator`, `Observable`, and `Observation` — where `Indicator` and `Observable` are connected by a one-to-one relation (with the primary marker on either side depending on the test), and `Observable` has a one-to-many relation to `Observation`. Tests assert that traversals and field resolution are correct when navigating the full three-type chain in different directions and with different primary-side placements.

## Test Index

### `simple_test.go`

Basic traversal tests covering all direction and primary-side combinations of the Indicator–Observable–Observation chain.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToOneToMany` | 21-94 | Query across a one-to-one-to-many chain returns the full linked hierarchy. |
| `TestQueryOneToOneToManyFromSecondaryOnOneToMany` | 96-171 | Query from the secondary one-to-one side traverses into the one-to-many observations list. |
| `TestQueryOneToOneToManyFromSecondaryOnOneToOne` | 173-246 | Query from the secondary one-to-one side traverses back through the one-to-one link to indicator. |
| `TestQueryOneToOneToManyFromSecondary` | 248-323 | Query from the secondary indicator side retrieves the observable with its many observations. |
