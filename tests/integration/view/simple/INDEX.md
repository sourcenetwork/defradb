# Index: `tests/integration/view/simple`

## Overview

This folder contains integration tests for simple (single-collection) views in DefraDB. The tests span both cacheless and materialized view types and cover basic querying, field subset and SDL/query mismatches, filter composition, aliases, @constraints/@default/@embedding directive behaviour, GraphQL introspection, lens transforms (including multi-step transforms, cardinality-changing transforms, CID-based transform references, and aggregate transforms), and persistence of transform configuration across node restarts.

## Test Index

### `materialized_test.go`

Tests the caching lifecycle of materialized views, including automatic refresh on view creation, correct re-refresh behaviour when data already exists in the cache, and the absence of automatic updates after a refresh.

| Test Function | Line | Description |
|---|---|---|
| `TestView_SimpleMaterialized_AutoUpdatesOnViewAdd` | 23-75 | Materialized view automatically refreshes when the view is first added. |
| `TestView_SimpleMaterialized_RefreshesAfterEarlierRefresh` | 77-141 | Materialized view can be refreshed again after a prior refresh that contained data. |
| `TestView_SimpleMaterialized_DoesNotAutoUpdate` | 143-201 | Materialized view does not include documents added after the last explicit refresh. |

### `simple_test.go`

Core view tests covering basic document retrieval, multiple documents, field subset restrictions, SDL/query field mismatches, and stacked views (a view of a view).

| Test Function | Line | Description |
|---|---|---|
| `TestView_Simple` | 21-67 | Simple cacheless view returns a single document. |
| `TestView_SimpleMultipleDocs` | 69-124 | Simple cacheless view returns all documents when multiple exist. |
| `TestView_SimpleWithFieldSubset_ErrorsSelectingExcludedField` | 126-169 | Querying a field excluded from the view SDL returns an error. |
| `TestView_SimpleWithExtraFieldInViewSDL` | 171-220 | A field declared in the view SDL but absent from the query returns nil. |
| `TestView_SimpleWithExtraFieldInViewQuery` | 222-273 | A field in the view query but absent from the SDL is not exposed to callers. |
| `TestView_SimpleViewOfView` | 275-335 | A view defined over another view returns data correctly. |

### `transform_restart_test.go`

Tests that a lens transform attached to a view is correctly persisted and continues to function after a node restart.

| Test Function | Line | Description |
|---|---|---|
| `TestView_SimpleWithTransformAndRestart` | 25-91 | Lens transform on a view persists and works correctly after a node restart. |

### `with_alias_test.go`

Tests that a field alias in the view query is reflected as the exposed field name on the view type.

| Test Function | Line | Description |
|---|---|---|
| `TestView_SimpleWithAlias` | 21-69 | Simple view with a field alias in the query exposes the aliased field name. |

### `with_contraints_test.go`

Tests that @constraints(size) directives declared on view fields do not enforce size limits at query time for either cacheless or materialized views.

| Test Function | Line | Description |
|---|---|---|
| `TestView_SimpleWithSizeConstraint_CacheLessView_DoesNotErrorOnSizeViolation` | 23-93 | Cacheless view with @constraints(size) does not enforce the size limit. |
| `TestView_SimpleWithSizeConstraint_MaterializedView_DoesNotErrorOnSizeViolation` | 98-168 | Materialized view with @constraints(size) does not enforce the size limit. |

### `with_default_value_test.go`

Tests that a @default directive on a view-only field does not populate a value when the underlying source data has no corresponding field.

| Test Function | Line | Description |
|---|---|---|
| `TestView_SimpleWithDefaultValue_DoesNotSetFieldValue` | 21-72 | A @default directive on a view field does not populate a value at query time. |

### `with_embeddings_test.go`

Tests that an @embedding directive on a view field does not trigger vector embedding generation when documents are queried through the view.

| Test Function | Line | Description |
|---|---|---|
| `TestView_SimpleWithEmbeddings_DoesNotGenerateEmbedding` | 21-73 | An @embedding directive on a view field does not generate a vector embedding. |

### `with_filter_test.go`

Tests filter behaviour in views, covering a filter applied solely in the view definition and the combination of a view-level filter with an additional query-time filter.

| Test Function | Line | Description |
|---|---|---|
| `TestView_SimpleWithFilter` | 21-74 | View query with a filter returns only documents matching the filter condition. |
| `TestView_SimpleWithFilterOnViewAndQuery` | 76-142 | Filters on both the view definition and the query are applied conjunctively. |

### `with_introspection_test.go`

Tests that GraphQL introspection returns the correct type name and field list for a simple view type.

| Test Function | Line | Description |
|---|---|---|
| `TestView_Simple_GQLIntrospectionTest` | 22-79 | GraphQL introspection exposes the correct fields for a simple view type. |

### `with_transform_aggregate_test.go`

Tests that a lens transform can compute an aggregate (standard deviation) across all view documents and expose it as a single result.

| Test Function | Line | Description |
|---|---|---|
| `TestView_SimpleWithTransformAggregate` | 25-97 | Lens transform computes a standard deviation aggregate over view documents. |

### `with_transform_cid_test.go`

Tests CID-based lens references for view transforms, covering successful reuse of an existing lens by CID and the error returned when the CID does not exist.

| Test Function | Line | Description |
|---|---|---|
| `TestView_WithTransformCID_CanReuseExistingLens` | 25-87 | A view can reference an existing lens by CID as its transform. |
| `TestView_WithInvalidTransformCID_ReturnsError` | 89-118 | Adding a view with a non-existent transform CID returns a lens not found error. |

### `with_transform_test.go`

Tests lens transforms in simple views, including single-field copy, multiple chained transforms, transforms that expand the result set, and transforms that filter the result set.

| Test Function | Line | Description |
|---|---|---|
| `TestView_SimpleWithTransform` | 25-100 | Lens transform copies a source field into a destination field in the view. |
| `TestView_SimpleWithMultipleTransforms` | 102-189 | A view applies multiple chained lens transforms to produce additional fields. |
| `TestView_SimpleWithTransformReturningMoreDocsThanInput` | 191-265 | Lens transform that prepends synthetic documents returns more results than source. |
| `TestView_SimpleWithTransformReturningFewerDocsThanInput` | 267-348 | Lens filter transform removes documents that do not match a condition. |
