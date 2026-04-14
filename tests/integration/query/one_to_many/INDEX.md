# Index: `tests/integration/query/one_to_many`

## Overview

This folder contains integration tests for one-to-many GraphQL queries in DefraDB, covering the `Book`–`Author` schema defined in `utils.go`. Tests exercise both directions of the relation (querying from the one side and from the many side) and validate a broad range of query features including filters, ordering, limits, offsets, grouping, aggregates (COUNT, SUM, AVG, MIN, MAX), CID/docID lookups, alias filters, and schema-level constraints.

## Test Index

### `one_sided_test.go`

Tests querying a one-to-many relationship defined with only the primary (one) side of the schema, without the inverse array field.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToMany_OneSided` | 21-73 | One-to-many query with a schema that only defines the one side of the relation. |

---

### `simple_test.go`

Core one-to-many query tests from both the primary (Book) and secondary (Author) sides, including the case of a missing parent.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToMany_PrimaryDirection` | 21-69 | One-to-many query from the many (Book) side resolves the related author. |
| `TestQueryOneToMany_SecondaryDirection` | 71-160 | One-to-many query from the one (Author) side returns all related books. |
| `TestQueryOneToManyWithNonExistantParent` | 162-198 | One-to-many query where the related author does not exist returns nil. |

---

### `with_average_test.go`

Tests filtering on an aliased AVG aggregate of related book ratings.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToMany_WithAverageAliasFilter_ShouldMatchAll` | 21-90 | One-to-many query filtering on an aliased average of related field returns all authors. |
| `TestQueryOneToMany_WithAverageAliasFilter_ShouldMatchOne` | 92-156 | One-to-many query filtering on an aliased average of related field returns one matching author. |

---

### `with_cid_doc_id_test.go`

Tests querying a specific document version by combining a CID and docID, including behaviour after parent or child updates.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToManyWithCidAndDocID` | 67-132 | One-to-many query using a specific cid and docID fetches the book with its author. |
| `TestQueryOneToManyWithChildUpdateAndFirstCidAndDocID` | 138-211 | One-to-many query at original parent cid returns current child state after child update. |
| `TestQueryOneToManyWithParentUpdateAndFirstCidAndDocID` | 213-286 | One-to-many query at the first cid returns the original parent state after a parent update. |
| `TestQueryOneToManyWithParentUpdateAndLastCidAndDocID` | 288-361 | One-to-many query at the latest cid returns the updated parent state. |

---

### `with_count_filter_test.go`

Tests COUNT aggregates on the related collection combined with filters, including JSON metadata filters.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToManyWithCountWithFilter` | 21-90 | One-to-many count of related docs filtered by a rating threshold. |
| `TestQueryOneToManyWithCountWithFilterAndChildFilter` | 92-184 | One-to-many count and child listing both filtered independently by rating presence. |
| `TestQueryOneToMany_WithCountWithJSONFilterAndChildFilter_Succeeds` | 186-267 | Top-level count with combined JSON metadata filter and related-field filter succeeds. |

---

### `with_count_limit_offset_test.go`

Tests COUNT with limit and offset applied independently to the count aggregate and to the child listing.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToManyWithCountAndLimitAndOffset` | 21-118 | Count of all related docs alongside a limited-and-offset child listing. |
| `TestQueryOneToManyWithCountAndDifferentOffsets` | 120-213 | Count with an offset differs from the child listing using a different offset. |
| `TestQueryOneToManyWithCountWithLimitWithOffset` | 215-284 | Count of related docs with both limit and offset applied to the count field. |

---

### `with_count_limit_test.go`

Tests COUNT with independent limits on the count aggregate versus the rendered child listing.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToManyWithCountAndLimit` | 21-103 | Count of all related docs alongside a limited child listing uses different limits. |
| `TestQueryOneToManyWithCountAndDifferentLimits` | 105-195 | Count with its own limit produces a different result than the separately-limited child listing. |
| `TestQueryOneToManyWithCountWithLimit` | 197-266 | Count of related docs with a limit of one returns 1 for every author. |

---

### `with_count_order_test.go`

Tests ordering parent results by an aliased COUNT aggregate.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToMany_WithCountAliasOrder_ShouldOrderResults` | 21-89 | Authors ordered descending by an aliased count of their published books. |

---

### `with_count_test.go`

Core COUNT aggregate tests including empty sets, all-match, and alias-based filter scenarios.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToMany_WithCount_NothingToCount` | 21-53 | Count of related books for an author with no published books returns zero. |
| `TestQueryOneToMany_WithCount_ShouldMatchAll` | 55-124 | Count of related books for each author returns correct totals for all authors. |
| `TestQueryOneToMany_WithCountAliasFilter_ShouldMatchAll` | 126-195 | Filter on aliased count greater than zero returns all authors with published books. |
| `TestQueryOneToMany_WithCountAliasFilter_ShouldMatchOne` | 197-261 | Filter on aliased count greater than one returns only the author with the most books. |

---

### `with_doc_ids_test.go`

Tests filtering the related child collection to a list of specific docIDs.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToManyWithChildDocIDs` | 21-97 | One-to-many query filtering related docs by a list of specific docIDs. |

---

### `with_doc_id_test.go`

Tests filtering the related child collection to a single specific docID.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToManyWithChildDocID` | 21-77 | One-to-many query filtering the related collection to a single specific docID. |

---

### `with_filter_related_id_test.go`

Tests filtering by the related type's docID field from both sides of the relation, including error cases.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryFromManySideWithEqFilterOnRelatedType` | 21-116 | Books filtered from the many side using an equality filter on the related author docID. |
| `TestQueryFromManySideWithFilterOnRelatedObjectID` | 118-213 | Books filtered using the raw _authorID scalar field equality from the many side. |
| `TestQueryFromManySideWithSameFiltersInDifferentWayOnRelatedType` | 215-315 | Books filtered by the same author via both the relation field and the raw ID field simultaneously. |
| `TestQueryFromSingleSideWithEqFilterOnRelatedType` | 317-411 | Authors filtered from the one side using a docID equality filter on a related book. |
| `TestQueryFromSingleSideWithFilterOnRelatedObjectID_Error` | 413-501 | Filtering authors by _publishedID (a non-existent scalar) returns a schema error. |

---

### `with_filter_test.go`

Tests numeric and compound filter conditions on parent and child fields, including aliased child filters.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToManyWithNumericGreaterThanFilterOnParent` | 21-106 | Authors filtered by age greater than 63 return with their published books. |
| `TestQueryOneToManyWithNumericGreaterThanChildFilterOnParentWithUnrenderedChild` | 108-176 | Authors filtered on age and child rating with the child field not selected in output. |
| `TestQueryOneToManyWithNumericGreaterThanFilterOnParentAndChild` | 178-258 | Authors filtered by age and published books filtered by rating simultaneously. |
| `TestQueryOneToManyWithMultipleAliasedFilteredChildren` | 260-362 | Two aliased child selections each with different rating filters on the same relation. |
| `TestQueryOneToManyWithCompoundOperatorInFilterAndRelation` | 364-475 | Authors filtered using compound _or/_and operators combining parent and related-field conditions. |
| `TestQueryOneToMany_WithCompoundOperatorInFilterAndRelationAndCaseInsensitiveLike_NoError` | 477-570 | Authors filtered with compound operators and a case-insensitive _ilike on related book name. |
| `TestQueryOneToMany_WithAliasFilterOnRelated_Succeeds` | 572-657 | Authors filtered on a _alias that references a related-field rating threshold. |

---

### `with_group_filter_test.go`

Tests filter conditions applied inside grouped queries, including filters on the GROUP and on the join within it.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToManyWithParentJoinGroupNumberAndNumberFilterOnJoin` | 21-157 | Authors grouped by age with related books filtered by rating inside the GROUP. |
| `TestQueryOneToManyWithParentJoinGroupNumberAndNumberFilterOnGroup` | 159-298 | Authors grouped by age with the GROUP filtered to include only those with high-rated books. |
| `TestQueryOneToManyWithParentJoinGroupNumberAndNumberFilterOnGroupAndOnGroupJoin` | 300-416 | Authors grouped by age with a parent filter and a nested child rating filter inside the GROUP. |

---

### `with_group_related_id_alias_test.go`

Tests grouping books by the related author using the relation alias, with various ID and selection combinations, and verifying errors when grouping authors by array fields.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToManyWithParentGroupByOnRelatedTypeFromManySideUsingAlias` | 21-200 | Books grouped by the related author object returns _authorID and GROUP with author fields. |
| `TestQueryOneToManyWithParentGroupByOnRelatedTypeFromManySideUsingAliasAndRelatedSelection` | 202-393 | Books grouped by author alias with the author docID and name selected at the group level. |
| `TestQueryOneToManyWithParentGroupByOnRelatedTypeWithIDSelectionFromManySideUsingAlias` | 395-574 | Books grouped by author alias with _authorID selected returns books partitioned by author. |
| `TestQueryOneToManyWithParentGroupByOnRelatedTypeWithIDSelectionFromManySideUsingAliasAndRelatedSelection` | 576-771 | Books grouped by author alias selecting both _authorID and the related author object at group level. |
| `TestQueryOneToManyWithParentGroupByOnRelatedTypeFromSingleSideUsingAlias` | 773-863 | Grouping authors by the published array field returns an error as arrays cannot be group keys. |
| `TestQueryOneToManyWithParentGroupByOnRelatedTypeWithIDSelectionFromSingleSideUsingAlias` | 865-956 | Selecting _publishedID on Author when grouping by the published array field returns an error. |

---

### `with_group_related_id_test.go`

Tests grouping by the explicit `_authorID` scalar and verifying errors when grouping Author by the non-existent `_publishedID` field.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToManyWithParentGroupByOnRelatedTypeIDFromManySide` | 21-184 | Books grouped by the _authorID scalar field with author fields accessible in the GROUP. |
| `TestQueryOneToManyWithParentGroupByOnRelatedTypeIDWithIDSelectionFromManySide` | 186-349 | Books grouped by _authorID with the ID selected returns each author's books correctly partitioned. |
| `TestQueryOneToManyWithParentGroupByOnRelatedTypeFromSingleSide` | 351-445 | Grouping authors by the _publishedID field (non-existent) returns a schema error. |
| `TestQueryOneToManyWithParentGroupByOnRelatedTypeWithIDSelectionFromSingleSide` | 447-542 | Selecting _publishedID and grouping by _publishedID on Author returns a schema error. |

---

### `with_group_test.go`

Tests grouping related books inside a nested join and grouping authors at the top level, including error cases for non-group fields selected at group level.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToManyWithInnerJoinGroupNumber` | 21-135 | Related books grouped by rating inside a nested join on authors. |
| `TestQueryOneToManyWithParentJoinGroupNumber` | 137-285 | Authors grouped by age with each group containing nested published books. |
| `TestQueryOneToManyWithInnerJoinGroupNumberWithNonGroupFieldsSelected` | 287-311 | Selecting a non-group-by field at the group level in a nested join returns an error. |

---

### `with_id_field_test.go`

Tests schema-level validation that prevents explicitly redeclaring the auto-generated relation ID field.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToManyWithIdFieldOnPrimary` | 21-44 | Defining a duplicate _authorID field explicitly in the schema returns an error. |

---

### `with_limit_test.go`

Tests applying limit to the related child collection, including multiple aliased child selections with different limits.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToManyWithSingleChildLimit` | 21-103 | One-to-many query with a limit of one on the related child collection. |
| `TestQueryOneToManyWithMultipleChildLimits` | 105-207 | Two aliased child selections with different limits on the same relation field. |

---

### `with_max_test.go`

Tests filtering on an aliased MAX aggregate of related book ratings.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToMany_WithMaxAliasFilter_ShouldMatchAll` | 21-90 | Filter on an aliased max rating greater than zero returns all authors. |
| `TestQueryOneToMany_WithMaxAliasFilter_ShouldMatchOne` | 92-156 | Filter on an aliased max rating greater than 4.8 returns only the author with the highest-rated book. |

---

### `with_min_test.go`

Tests filtering on an aliased MIN aggregate of related book ratings.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToMany_WithMinAliasFilter_ShouldMatchAll` | 21-90 | Filter on an aliased min rating greater than zero returns all authors with books. |
| `TestQueryOneToMany_WithMinAliasFilter_ShouldMatchOne` | 92-156 | Filter on an aliased min rating less than 4.7 returns only the author with the lowest-rated book. |

---

### `with_order_filter_limit_test.go`

Tests combining a parent filter with a sorted-and-limited child listing in ascending and descending directions.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToManyWithNumericGreaterThanFilterOnParentAndNumericSortAscendingAndLimitOnChild` | 21-97 | Author filtered by age with related books sorted ascending by rating and limited to one. |
| `TestQueryOneToManyWithNumericGreaterThanFilterOnParentAndNumericSortDescendingAndLimitOnChild` | 99-175 | Author filtered by age with related books sorted descending by rating and limited to one. |

---

### `with_order_filter_test.go`

Tests combining parent filters with child ordering in ascending and descending directions without a limit.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToManyWithNumericGreaterThanFilterOnParentAndNumericSortAscendingOnChild` | 22-102 | Author filtered by age with related books sorted ascending by rating. |
| `TestQueryOneToManyWithNumericGreaterThanFilterAndNumericSortDescendingOnChild` | 104-194 | Authors filtered by child rating with related books sorted descending by rating. |

---

### `with_related_id_test.go`

Tests selecting and querying the `_authorID` scalar from both sides of the relation, including the error case on the Author side.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToManyWithRelatedTypeIDFromManySide` | 21-138 | Books queried from the many side include the _authorID scalar field in results. |
| `TestQueryOneToManyWithRelatedTypeIDFromSingleSide` | 140-229 | Querying _authorID on Author (which only has the many side) returns a schema error. |

---

### `with_same_field_name_test.go`

Tests a one-to-many schema where both sides of the relation use the same field name, queried from each direction.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToManyWithSameFieldName_SingleSide` | 49-90 | One-to-many query where both sides share the same field name works from the one side. |
| `TestQueryOneToManyWithSameFieldName_MultiSide` | 92-136 | One-to-many query where both sides share the same field name works from the many side. |

---

### `with_sum_filter_order_test.go`

Tests SUM aggregates combined with parent-level filters and ordering, including dual aliased sums with distinct limit/order settings.

| Test Function | Line | Description |
|---|---|---|
| `TestOneToManyAscOrderAndFilterOnParentWithAggSumOnSubTypeField` | 21-133 | Sum of published ratings with authors filtered by age and ordered ascending by age. |
| `TestOneToManyDescOrderAndFilterOnParentWithAggSumOnSubTypeField` | 135-247 | Sum of published ratings with authors filtered by age and ordered descending by age. |
| `TestOnetoManySumBySubTypeFieldAndSumBySybTypeFieldWithDescOrderingOnFieldWithLimit` | 249-370 | Two aliased sums of published ratings with different order and limit on the second sum. |
| `TestOnetoManySumBySubTypeFieldAndSumBySybTypeFieldWithAscOrderingOnFieldWithLimit` | 372-493 | Two aliased sums of published ratings with ascending order and limit on the second sum. |
| `TestOneToManyLimitAscOrderSumOfSubTypeAndLimitAscOrderFieldsOfSubtype` | 495-624 | Aliased sum with ascending order and limit alongside a matching limited-and-ordered child listing. |
| `TestOneToManyLimitDescOrderSumOfSubTypeAndLimitAscOrderFieldsOfSubtype` | 626-755 | Aliased sum with descending order and limit alongside a matching limited-and-ordered child listing. |

---

### `with_sum_limit_offset_order_test.go`

Tests SUM with limit, offset, and ordering applied in combinations including ascending, descending, and mixed-field ordering.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToManyWithSumWithLimitWithOffsetWithOrderAsc` | 21-131 | Sum of published ratings with offset, limit, and ascending order by book name. |
| `TestQueryOneToManyWithSumWithLimitWithOffsetWithOrderDesc` | 133-242 | Sum of published ratings with offset, limit, and descending order by book name. |
| `TestQueryOneToManyWithSumWithLimitWithOffsetWithOrderAscAndDesc` | 244-358 | Two aliased sums using the same offset and limit but ascending and descending name order respectively. |
| `TestQueryOneToManyWithSumWithLimitWithOffsetWithOrderOnDifferentFields` | 360-473 | Two aliased sums with the same offset and limit but ordering on different fields (name vs rating). |
| `TestQueryOneToManyWithSumWithLimitWithOffsetWithOrderDescAndRenderedChildrenOrderedAsc` | 475-596 | Sum ordered descending by name with a separately rendered child list ordered ascending by name. |

---

### `with_sum_limit_offset_test.go`

Tests SUM with both limit and offset applied to the aggregation window.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToManyWithSumWithLimitAndOffset` | 21-98 | Sum of published book ratings applying both an offset and a limit to the aggregation. |

---

### `with_sum_limit_test.go`

Tests SUM with a limit applied to the aggregation window.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToManyWithSumWithLimit` | 21-99 | Sum of published book ratings with a limit of two applied to the aggregation. |

---

### `with_sum_order_test.go`

Tests ordering parent results by an aliased SUM aggregate.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToMany_WithSumWithAliasOrder_ShouldOrderResults` | 21-97 | Authors ordered descending by an aliased sum of published book ratings. |

---

### `with_sum_test.go`

Tests filtering on aliased SUM aggregates including Float64 and Float32 field types.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToMany_WithSumAliasFilter_ShouldMatchAll` | 21-90 | Filter on an aliased sum of ratings greater than zero returns all authors with books. |
| `TestQueryOneToMany_WithSumAliasFilter_ShouldMatchOne` | 92-156 | Filter on an aliased sum of ratings greater than five returns only the author with more books. |
| `TestQueryOneToMany_WithSumAliasFilterOnFloat32_ShouldMatchOne` | 158-238 | Filter on an aliased sum of Float32 ratings greater than five returns only the prolific author. |

---

### `with_typename_test.go`

Tests that `__typename` is returned correctly for both sides of a one-to-many relation.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryOneToManyWithTypeName` | 21-69 | One-to-many query requesting __typename returns the correct type names for both sides. |
