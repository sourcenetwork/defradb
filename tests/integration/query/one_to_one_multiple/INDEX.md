# Index: `tests/integration/query/one_to_one_multiple`

## Overview

This folder contains integration tests for schemas where a single collection (Book) holds multiple independent one-to-one relationships simultaneously — one to Publisher and one to Author. Tests exercise all combinations of which side of each relationship owns the foreign key (@primary), verifying that queries resolve both related documents correctly regardless of whether the join originates from the primary or secondary side. An additional test confirms that schema creation with cyclic, mutually referential relations across three types succeeds without error.

## Test Index

### `simple_test.go`

Tests querying a collection that maintains two separate one-to-one relations under all primary/secondary ownership combinations.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToOneMultiple_FromPrimary` | 21-124 | Querying Book with two @primary relations returns correct publisher and author for each book. |
| `TestQueryOneToOneMultiple_FromMixedPrimaryAndSecondary` | 126-229 | Querying Book with one @primary and one secondary relation resolves both sides correctly. |
| `TestQueryOneToOneMultiple_FromSecondary` | 231-334 | Querying Book with two secondary relations owned by Publisher and Author resolves both correctly. |
| `TestAddCollectionWithCyclicMutuallyReferentialRelations_DoesNotError` | 336-364 | Adding a schema with cyclic mutually referential relations between three types does not error. |
