# Index: `tests/integration/view/one_to_many`

## Overview

This folder contains integration tests for views defined over one-to-many relations in DefraDB. The tests cover basic view creation and querying, alias support on both the outer and inner types, aggregate functions (COUNT) within view queries, GraphQL introspection of view types, and lens transforms applied to outer types or used to synthesise inner embedded documents.

## Test Index

### `simple_test.go`

Tests covering the core behaviours of one-to-many views, including basic querying, mixed SDL, error cases for querying from the inner side, relation omission in the SDL, multiple views sharing embedded schemas, and stacked views.

| Test Function | Line | Description |
|---|---|---|
| `TestView_OneToMany` | 21-95 | View over a one-to-many relation returns nested documents. |
| `TestView_OneToManyWithMixedSDL` | 97-168 | View over a one-to-many relation with mixed SDL referencing the base collection type. |
| `TestView_OneToManyFromInnerSide_Errors` | 170-220 | Querying a one-to-many view from the inner interface type errors. |
| `TestView_OneToManyOuterToInnerToOuter_Errors` | 222-278 | Querying a back-reference field on an inner view interface type errors. |
| `TestView_OneToManyWithRelationInQueryButNotInSDL` | 280-344 | View query includes a relation not declared in the view SDL. |
| `TestView_OneToManyMultipleViewsWithEmbeddedSchema` | 346-404 | Multiple views with distinct embedded interface schemas can be created together. |
| `TestView_OneToManyWithDoubleSidedRelation_Errors` | 406-499 | A view of a view over a one-to-many relation returns nested results. |

### `with_alias_test.go`

Tests that field aliases in the view query work correctly on both the outer type and the inner embedded interface type.

| Test Function | Line | Description |
|---|---|---|
| `TestView_OneToManyWithAliasOnOuter` | 21-95 | View over one-to-many relation uses a field alias on the outer type. |
| `TestView_OneToManyWithAliasOnInner` | 97-173 | View over one-to-many relation uses a field alias on the inner embedded type. |

### `with_count_test.go`

Tests the behaviour of COUNT aggregates within one-to-many view queries, covering the error case when COUNT is not aliased, successful aliased COUNT, and COUNT present in the query but absent from the SDL.

| Test Function | Line | Description |
|---|---|---|
| `TestView_OneToManyWithCount_Errors` | 23-87 | Using COUNT without an aliased field in a one-to-many view query errors. |
| `TestView_OneToManyWithAliasedCount` | 89-161 | View over one-to-many relation exposes an aliased COUNT aggregate field. |
| `TestView_OneToManyWithCountInQueryButNotSDL` | 163-218 | COUNT in the view query is omitted from the view SDL without error. |

### `with_introspection_test.go`

Tests that GraphQL schema introspection correctly exposes the fields declared on the outer view type and the embedded inner interface type of a one-to-many view.

| Test Function | Line | Description |
|---|---|---|
| `TestView_OneToMany_GQLIntrospectionTest` | 22-133 | GraphQL introspection exposes correct fields for a one-to-many view type. |

### `with_transform_test.go`

Tests lens transforms in one-to-many views, covering transforms on the outer type and transforms that synthesise inner embedded documents from a collection with no declared relation.

| Test Function | Line | Description |
|---|---|---|
| `TestView_OneToManyWithTransformOnOuter` | 25-117 | Lens transform on the outer type copies a field in a one-to-many view. |
| `TestView_OneToManyWithTransformAddingInnerDocs` | 119-204 | Lens transform synthesises inner embedded documents in a one-to-many view. |
