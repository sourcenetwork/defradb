# Index: `tests/integration/query/one_to_one`

## Overview

This folder contains integration tests for querying one-to-one relations in DefraDB. The tests cover basic traversal in both the primary and secondary direction, nil relation handling, relation ID field access, filtering (boolean, numeric, string, compound), ordering by related fields, groupBy on relation ID fields (with and without aliases), GraphQL fragment spreading, and `_version` metadata queries on linked types. The shared schema uses a `Book`–`Author` pair where `Author` holds the `@primary` key via its `published` field.

## Test Index

### `simple_test.go`

Core traversal tests asserting that one-to-one joins work in both directions, with nil relations, multiple records, and direct relation ID field access.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToOne_PrimaryDirection` | 21-69 | Query one-to-one relation from the primary (Book) side returns linked author. |
| `TestQueryOneToOne_SecondaryDirection` | 71-119 | Query one-to-one relation from the secondary (Author) side returns linked book. |
| `TestQueryOneToOneWithMultipleRecords` | 121-203 | Query one-to-one relation with multiple records returns each book with its author. |
| `TestQueryOneToOneWithMultipleRecordsSecondaryDirection` | 205-282 | Query one-to-one from the secondary side with multiple records returns each author with their book. |
| `TestQueryOneToOneWithNilChild` | 284-316 | Query one-to-one from secondary side returns nil when the related book is absent. |
| `TestQueryOneToOneWithNilParent` | 318-349 | Query one-to-one from primary side returns nil when the related author is absent. |
| `TestQueryOneToOne_WithRelationIDFromPrimarySide` | 351-401 | Querying the relation ID field from the primary side returns the linked document ID. |
| `TestQueryOneToOne_WithRelationIDFromSecondarySide` | 403-453 | Querying the relation ID field from the secondary side returns the linked document ID. |

### `with_clashing_id_field_test.go`

Error-path tests asserting that explicitly declaring a relation ID field that conflicts with the auto-generated one is rejected.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToOneWithClashingIdFieldOnSecondary` | 21-44 | Defining an explicit relation ID field on the secondary side returns a duplicate field error. |
| `TestQueryOneToOneWithClashingIdFieldOnPrimary` | 46-69 | Defining an explicit relation ID field on the primary side returns a duplicate field error. |

### `with_count_filter_test.go`

Tests the COUNT aggregate combined with compound relational filters on a one-to-one schema.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToOneWithCountWithCompoundOrFilterThatIncludesRelation` | 21-107 | COUNT query with a compound _or filter referencing a one-to-one relation returns correct count. |

### `with_filter_order_test.go`

Tests combining a relational filter with a relational order clause while omitting the subtype from the selection set.

| Test Function | Line | Description |
|---|---|---|
| `TestOnetoOneSubTypeDscOrderByQueryWithFilterHavinghNoSubTypeSelections` | 21-84 | Query books ordered DESC by related author age with a filter and no author fields selected. |
| `TestOnetoOneSubTypeAscOrderByQueryWithFilterHavinghNoSubTypeSelections` | 86-149 | Query books ordered ASC by related author age with a filter and no author fields selected. |

### `with_filter_test.go`

Tests for filtering one-to-one query results using scalar, boolean, compound, and negated filter expressions that reference the related type.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToOneWithNumericFilterOnParent` | 21-72 | Filtering the related author subtype by an integer field returns the matching book. |
| `TestQueryOneToOneWithStringFilterOnChild` | 74-127 | Filtering the root Book collection by a string field returns the matching book with its author. |
| `TestQueryOneToOneWithBooleanFilterOnChild` | 129-182 | Filtering books by a boolean field on the related author returns the matching result. |
| `TestQueryOneToOneWithFilterThroughChildBackToParent` | 184-251 | Filtering books by a field on the author's linked book traverses the relation back to parent. |
| `TestQueryOneToOneWithBooleanFilterOnChildWithNoSubTypeSelection` | 253-296 | Filtering books by a boolean on the related author without selecting author fields returns books only. |
| `TestQueryOneToOneWithCompoundAndFilterThatIncludesRelation` | 298-376 | A compound _and filter combining a book rating and a related author field narrows results correctly. |
| `TestQueryOneToOneWithCompoundOrFilterThatIncludesRelation` | 378-501 | A compound _or filter combining book rating and related author age returns the expected subset. |
| `TestQueryOneToOne_WithCompoundFiltersThatIncludesRelation_ShouldReturnResults` | 503-630 | Multiple compound filter shapes (_or, _and, _not) referencing a relation all return correct books. |

### `with_fragments_test.go`

Tests that GraphQL named fragments can be spread on the root type and on the related type in a one-to-one query.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToOne_WithFragment` | 21-81 | A GraphQL fragment on the parent type correctly spreads related author fields. |
| `TestQueryOneToOne_WithFragmentWithObjectWithFragment` | 83-146 | Nested fragments on both the parent and the related object type resolve correctly. |

### `with_group_related_id_alias_test.go`

Tests groupBy using the relation field alias (e.g. `author`) rather than the raw `_authorID` field, from both the primary and secondary sides, with and without GROUP and join selections.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToOneWithGroupRelatedIDAlias` | 21-108 | GroupBy the relation alias field from the primary side returns each author's books in a GROUP. |
| `TestQueryOneToOneWithGroupRelatedIDAliasFromSecondaryWithoutInnerGroup` | 110-174 | GroupBy the alias relation field from the secondary side without selecting GROUP returns only IDs. |
| `TestQueryOneToOneWithGroupRelatedIDAliasFromSecondaryWithoutInnerGroupWithJoin` | 176-250 | GroupBy alias relation field from secondary side with a join returns IDs and author names. |
| `TestQueryOneToOneWithGroupRelatedIDAliasFromSecondaryWithInnerGroup` | 252-330 | GroupBy alias relation field from secondary side with GROUP returns each book under its author ID. |
| `TestQueryOneToOneWithGroupRelatedIDAliasFromSecondaryWithInnerGroupWithJoin` | 332-419 | GroupBy alias relation from secondary side with GROUP and join returns IDs, authors, and books. |

### `with_group_related_id_test.go`

Tests groupBy using the raw `_authorID` relation ID field from both the primary and secondary sides, with and without GROUP and join selections.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToOneWithGroupRelatedID` | 21-99 | GroupBy the relation ID field from the primary side groups books under each author ID. |
| `TestQueryOneToOneWithGroupRelatedIDFromSecondaryWithoutGroup` | 101-165 | GroupBy the relation ID from the secondary side without GROUP returns only the ID per group. |
| `TestQueryOneToOneWithGroupRelatedIDFromSecondaryWithoutGroupWithJoin` | 167-241 | GroupBy the relation ID from secondary side with a join returns IDs and author names without GROUP. |
| `TestQueryOneToOneWithGroupRelatedIDFromSecondaryWithGroup` | 243-321 | GroupBy the relation ID from secondary side with GROUP returns each book nested under its author ID. |
| `TestQueryOneToOneWithGroupRelatedIDFromSecondaryWithGroupWithJoin` | 323-410 | GroupBy relation ID from secondary side with GROUP and join returns IDs, authors, and books. |

### `with_order_test.go`

Tests ordering one-to-one query results by fields on the related type, including boolean and integer fields, with and without subtype field selection, and with aliased field references.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToOneWithChildBooleanOrderDescending` | 21-93 | Books ordered DESC by the related author's boolean field returns verified author first. |
| `TestQueryOneToOneWithChildBooleanOrderAscending` | 95-167 | Books ordered ASC by the related author's boolean field returns unverified author first. |
| `TestQueryOneToOneWithChildIntOrderDescendingWithNoSubTypeFieldsSelected` | 169-229 | Books ordered DESC by the author's integer age field with no author fields in the selection. |
| `TestQueryOneToOneWithChildIntOrderAscendingWithNoSubTypeFieldsSelected` | 231-291 | Books ordered ASC by the author's integer age field with no author fields in the selection. |
| `TestQueryOneToOne_WithAliasedChildIntOrderAscending_ShouldOrder` | 293-362 | Books ordered ASC by an aliased child relation's integer age field are correctly sorted. |
| `TestQueryOneToOne_WithChildAliasedIntOrderAscending_ShouldOrder` | 364-433 | Books ordered ASC by an aliased field inside the author subtype are correctly sorted. |

### `with_version_test.go`

Tests that the `_version` metadata field resolves correctly alongside a one-to-one join regardless of its position in the selection set.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToOne_WithVersionOnOuterBeforeJoin` | 21-85 | Requesting _version before the join field on a one-to-one relation returns correct docID. |
| `TestQueryOneToOne_WithVersionOnOuterAfterJoin` | 87-151 | Requesting _version after the join field on a one-to-one relation returns correct docID. |
