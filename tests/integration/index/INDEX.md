# Index Integration Tests

This directory contains integration tests for DefraDB's index subsystem. The tests cover index creation, querying with various filter operators (`_eq`, `_gt`, `_geq`, `_lt`, `_leq`, `_neq`, `_in`, `_nin`, `_like`, `_nlike`, `_ilike`, `_nilike`), unique and composite indexes, relation-based filtering and ordering, JSON path indexing, and index maintenance during document updates. Each test verifies both the correctness of returned results and, where applicable, the query plan efficiency (number of document and index fetches) using explain assertions.

---

### `json_test.go`

| Test Function | Line | Description |
|---|---|---|
| `TestJSONIndex_WithFilterOnNumberField_ShouldUseIndex` | 21-91 | Equality filter on a numeric field within an indexed JSON field uses the index. |
| `TestJSONIndex_WithGtFilterOnNumberField_ShouldUseIndex` | 93-161 | Greater-than filter on a numeric field within an indexed JSON field uses range optimization. |
| `TestJSONIndex_WithGeFilterOnNumberField_ShouldUseIndex` | 163-244 | Greater-or-equal filter on a numeric field within an indexed JSON field uses range optimization. |
| `TestJSONIndex_WithLtFilterOnNumberField_ShouldUseIndex` | 246-314 | Less-than filter on a numeric field within an indexed JSON field uses range optimization. |
| `TestJSONIndex_WithLeFilterOnNumberField_ShouldUseIndex` | 316-385 | Less-or-equal filter on a numeric field within an indexed JSON field uses range optimization. |
| `TestJSONIndex_WithNeFilterOnNumberField_ShouldUseIndex` | 387-458 | Not-equal filter on a numeric field within an indexed JSON field returns all non-matching docs. |
| `TestJSONIndex_WithEqFilterOnStringField_ShouldUseIndex` | 460-529 | Equality filter on a string field within an indexed JSON field uses the index. |
| `TestJSONIndex_WithLikeFilterOnStringField_ShouldUseIndex` | 531-617 | _like and _ilike filters on a string field within an indexed JSON field scan index entries. |
| `TestJSONIndex_WithNLikeFilterOnStringField_ShouldUseIndex` | 619-705 | _nlike and _nilike filters on a string field within an indexed JSON field return non-matching docs. |
| `TestJSONIndex_WithEqFilterOnBoolField_ShouldUseIndex` | 707-777 | Equality filter on a boolean field within an indexed JSON field uses the index. |
| `TestJSONIndex_WithNeFilterOnBoolField_ShouldUseIndex` | 779-851 | Not-equal filter on a boolean field within an indexed JSON field returns all non-false docs. |
| `TestJSONIndex_WithEqFilterOnNullField_ShouldUseIndex` | 853-922 | Equality filter for null on a field within an indexed JSON field returns docs with that field explicitly null. |
| `TestJSONIndex_WithNeFilterOnNullNestedField_ShouldUseIndex` | 924-981 | Not-equal-null filter on a nested field within an indexed JSON field returns docs with non-null values. |
| `TestJSONIndex_UponUpdate_ShouldUseNewIndexValues` | 983-1050 | Updating a JSON field updates the index so subsequent filters use the new values. |
| `TestJSONIndex_WithInFilter_ShouldUseIndex` | 1052-1116 | _in filter on a numeric field within an indexed JSON field uses the index for each listed value. |
| `TestJSONIndex_WithInFilterOfDifferentTypes_ShouldUseIndex` | 1118-1175 | _in filter with mixed numeric and string values on an indexed JSON field matches each typed value. |
| `TestJSONIndex_WithNinFilter_ShouldUseIndex` | 1177-1234 | _nin filter on a numeric field within an indexed JSON field excludes docs with listed values. |
| `TestJSONIndex_WithNotAndInFilter_ShouldNotUseIndex` | 1236-1290 | _not combined with an _in filter on an indexed JSON field does not use the index. |
| `TestJSONIndex_WithCompoundFilterCondition_ShouldUseIndex` | 1292-1351 | Compound _and filter on two different paths within an indexed JSON field uses the index for both. |
| `TestJSONIndex_WithNeFilterAgainstNumberField_ShouldFetchNullValues` | 1353-1409 | Not-equal filter on a numeric JSON field returns docs with different values including those with null. |
| `TestJSONIndex_WithNeFilterAgainstStringField_ShouldFetchNullValues` | 1411-1467 | Not-equal filter on a string JSON field returns docs with different values including those with null. |
| `TestJSONIndex_WithNeFilterAgainstBoolField_ShouldFetchNullValues` | 1469-1525 | Not-equal filter on a boolean JSON field returns docs with different values including those with null. |
| `TestJSONIndex_WithEqFilterAgainstExplicitNullField_ShouldFetchNullValues` | 1527-1584 | Equality filter for null on a top-level indexed JSON field returns docs with explicit null or missing value. |
| `TestJSONIndex_WithGreaterThanFilterOnTopLevelJSONField_ShouldUseIndex` | 1586-1654 | Greater-than filter directly on a top-level indexed JSON field uses the index. |
| `TestJSONIndex_WithGeqNullFilterOnTopLevelJSONField_ShouldNotUseIndex` | 1656-1717 | _geq null filter on a top-level indexed JSON field does not use the index and returns all docs. |
| `TestJSONIndex_WithGeqNullFilterOnNestedJSONPath_ShouldNotUseIndex` | 1719-1780 | _geq null filter on a nested JSON path does not use the index and returns all docs. |
| `TestJSONIndex_WithLeqNullFilterOnTopLevelJSONField_ShouldUseIndex` | 1782-1847 | _leq null filter on a top-level indexed JSON field uses the index to return docs with null or missing value. |
| `TestJSONIndex_WithLeqNullFilterOnNestedJSONPath_ShouldNotUseIndex` | 1849-1912 | _leq null filter on a nested JSON path falls back to a full scan because the index can't handle all null cases. |
| `TestJSONIndex_WithEqFilterWithObjectValueOnNestedPath_ShouldFilter` | 1914-1972 | _eq filter with an object value on a nested JSON path does not use the index and filters in-memory. |
| `TestJSONIndex_WithNeqFilterWithObjectValueOnNestedPath_ShouldFilter` | 1974-2034 | _neq filter with an object value on a nested JSON path does not use the index and filters in-memory. |
| `TestJSONIndex_WithInFilterWithObjectValueOnNestedPath_ShouldFilter` | 2036-2096 | _in filter with object values on a nested JSON path does not use the index and filters in-memory. |
| `TestJSONIndex_WithNinFilterWithObjectValueOnNestedPath_ShouldFilter` | 2098-2158 | _nin filter with object values on a nested JSON path does not use the index and filters in-memory. |

---

### `query_with_relation_filter_test.go`

| Test Function | Line | Description |
|---|---|---|
| `TestQueryWithIndexOnOneToManyRelation_IfFilterOnIndexedRelation_ShouldFilter` | 21-88 | Filter on an indexed field of a one-to-many relation correctly returns matching parent docs. |
| `TestQueryWithIndexOnOneToOnesSecondaryRelation_IfFilterOnIndexedRelation_ShouldFilter` | 90-157 | Filter on an indexed field of a one-to-one secondary relation correctly returns matching parent docs. |
| `TestQueryWithIndexOnOneToOnePrimaryRelation_IfFilterOnIndexedFieldOfRelationAndRelation_ShouldFilter` | 159-230 | Filter on an indexed field of a one-to-one primary relation with a unique index on the relation field. |
| `TestQueryWithIndexOnOneToOnePrimaryRelation_IfFilterOnIndexedFieldOfRelation_ShouldFilter` | 232-305 | Filter on an indexed field of a one-to-one primary relation returns correct parent docs. |
| `TestQueryWithIndexOnOneToOnePrimaryRelation_IfFilterOnIndexedRelationWhileIndexedForeignField_ShouldFilter` | 307-354 | Filtering on an indexed relation field uses both the relation and foreign-key indexes. |
| `TestQueryWithIndexOnOneToMany_IfFilterOnIndexedPrimaryDoc_ShouldFilter` | 356-450 | Filter on an indexed child field of a one-to-many relation returns all sibling docs of matched parent. |
| `TestQueryWithIndexOnOneToMany_IfFilterOnIndexedPrimaryDocAndSubFilter_ShouldFilter` | 452-540 | Filter on an indexed child field with a sub-filter on related docs returns only sub-filtered results. |
| `TestQueryWithIndexOnOneToMany_IfFilterOnIndexedRelation_ShouldFilterWithExplain` | 542-634 | Filter on an indexed child relation field returns matching parent docs and verifies index fetch count. |
| `TestQueryWithIndexOnOneToOne_IfFilterOnIndexedRelation_ShouldFilter` | 636-688 | Filter on an indexed secondary relation field in a one-to-one schema returns matching parent docs. |
| `TestQueryWithIndexOnManyToOne_IfFilterOnIndexedField_ShouldFilterWithExplain` | 690-758 | Filter on an indexed scalar field of a many-to-one primary document uses the root index. |
| `TestQueryWithIndexOnManyToOne_IfFilterOnIndexedRelation_ShouldFilterWithExplain` | 760-812 | Filter on an indexed relation owner field uses subType and root indexes to find matching devices. |
| `TestQueryWithIndexOnOneToMany_IfIndexedRelationIsNil_NeNilFilterShouldUseIndex` | 814-894 | Not-equal-nil filter on an indexed relation ID field uses the index to exclude ownerless docs. |
| `TestQueryWithIndexOnOneToMany_IfIndexedRelationIsNil_EqNilFilterShouldUseIndex` | 896-975 | Equal-nil filter on an indexed relation ID field uses the index to return ownerless docs. |
| `TestQueryWithIndexOnManyToOne_MultipleViaOneToMany` | 979-1047 | Multiple indexed relation fields on a child document correctly resolve both parent references. |
| `TestQueryWithUniqueIndex_WithFilterOnChildIndexedField_ShouldFetch` | 1049-1086 | Filter on an indexed field of a child relation returns an empty result when no child docs exist. |
| `TestQueryWithIndex_WithScalarAndRelationFilterAtTopLevel_ShouldApplyBothAsAnd` | 1088-1202 | Scalar and relation filters at top level are combined as an implicit AND condition. |
| `TestQueryWithIndex_WithMultipleScalarsAndRelationFilter_ShouldApplyAllAsAnd` | 1204-1294 | Multiple scalar fields and a relation filter at top level are all applied as an AND condition. |

---

### `query_with_relation_sub_filter_test.go`

| Test Function | Line | Description |
|---|---|---|
| `TestQueryWithIndexOnOneToMany_IfSubFilterOnIndexedField_ShouldFilter` | 21-106 | Sub-filter on an indexed child field in a one-to-many query uses the index to filter devices. |
| `TestQueryWithIndexOnOneToMany_IfSubFilterOnNonIndexedField_ShouldNotUseIndex` | 108-185 | Sub-filter on a non-indexed child field in a one-to-many query does not use any index. |
| `TestQueryWithIndexOnOneToMany_IfSubFilterAndOrderOnIndexedField_ShouldUseIndexForFilter` | 187-280 | Sub-filter with order and limit on an indexed child field uses the index and returns ordered results. |
| `TestQueryWithIndexOnOneToMany_WithOrderOnParentAndSubFilter_ShouldFilterPerParent` | 282-377 | Ordered parent query with an indexed sub-filter correctly filters child docs for each parent. |
| `TestQueryWithIndexOnOneToMany_WithOrderOnParentAndSubFilter_ShouldFilterBothWithIndexes` | 379-468 | Indexed parent filter combined with an indexed sub-filter uses both indexes for efficient fetching. |
| `TestQueryWithIndexOnOneToMany_WithSameFilterOnParentAndSubType_ShouldFilterBothWithIndexes` | 470-552 | Different indexed filters on parent and sub-type use the index for both parent existence and child retrieval. |
| `TestQueryWithIndexOnOneToMany_WithSameFilterValueOnParentAndSubType_ShouldReturnMatchingDocs` | 554-629 | Same indexed filter value on both parent and sub-type correctly returns only matched child docs. |
| `TestQueryWithIndexOnOneToMany_WithParentFilterOnRelationAndSubFilterOnDifferentIndexedField_ShouldUseBothIndexes` | 631-732 | Parent filter on one indexed relation field and sub-filter on a different indexed field each use their index. |
| `TestQueryWithIndexOnOneToMany_WithParentFilterOnRelationAndSubFilterOnNonIndexedField_ShouldUseParentIndex` | 734-816 | Parent filter uses the relation index while sub-filter on a non-indexed field is applied in-memory. |
| `TestQueryWithIndexOnOneToMany_WithParentFilterOnOwnFieldAndRelationAndSubFilter_ShouldCombineAllFilters` | 818-909 | Parent scalar filter, relation filter, and sub-filter on different indexed fields are all combined correctly. |

---

### `query_with_relation_sub_order_test.go`

| Test Function | Line | Description |
|---|---|---|
| `TestQueryWithOrderOnOneToMany_WithIndexOnOrderFieldDescending_ShouldOrder` | 21-93 | DESC sub-order on an indexed child field returns the author's books in descending rating order. |
| `TestQueryWithOrderOnOneToMany_WithIndexOnOrderFieldAscending_ShouldOrder` | 95-167 | ASC sub-order on an indexed child field returns the author's books in ascending rating order. |
| `TestQueryWithOrderOnOneToMany_WithIndexOnOrderFieldAscendingWithLimit_ShouldOrderAndLimit` | 169-239 | ASC sub-order with a limit on an indexed child field returns only the lowest-rated book. |
| `TestQueryWithOrderOnOneToMany_WithMultipleAuthors_ShouldOrderEachAuthorsBooks` | 241-337 | DESC sub-order on an indexed rating field independently orders each author's books. |
| `TestQueryWithOrderOnOneToMany_WithMultipleAuthorsAndIndexOnRelation_ShouldOrderEachAuthorsBooks` | 339-435 | Sub-order on rating with an additional relation index uses the relation index to scope each author's books. |
| `TestQueryWithOrderOnOneToMany_WithSubFilterAndOrderAndRelationIndex_ShouldFilterThenOrder` | 437-532 | Sub-filter and DESC sub-order on the same indexed field filters and orders each author's books. |
| `TestQueryWithOrderOnOneToMany_WithParentFilterOnRelationAndSubOrder_ShouldOrderChildren` | 534-624 | Parent filter on a child relation field combined with DESC sub-order returns filtered and ordered books. |
| `TestQueryWithNestedOrderByRelationField_WithDESCAndLimit_RecursiveExplain` | 626-735 | Nested DESC sub-order on a grandchild indexed field with limit returns the 2 most recent books. |
| `TestQueryWithNestedOrderByRelationField_WithASCAndLimit_RecursiveExplain` | 737-843 | Nested ASC sub-order on a grandchild indexed field with limit returns the 2 oldest books. |
| `TestQueryWithOrderByRelationField_ExhaustiveWithParentSecondaryASC_ShouldIncludeOrphans` | 845-903 | Exhaustive ASC order by a secondary relation field includes orphan docs at the top of results. |
| `TestQueryWithOrderByRelationField_ExhaustiveWithParentSecondaryDESC_ShouldIncludeOrphans` | 905-963 | Exhaustive DESC order by a secondary relation field includes orphan docs at the bottom of results. |
| `TestQueryWithOrderByRelationField_ExhaustiveWithParentPrimaryASC_ShouldIncludeOrphans` | 965-1025 | Exhaustive ASC order by a primary relation field includes orphan publishers at the top of results. |
| `TestQueryWithOrderByRelationField_ExhaustiveWithParentPrimaryDESC_ShouldIncludeOrphans` | 1027-1087 | Exhaustive DESC order by a primary relation field appends orphan publishers at the bottom of results. |
| `TestQueryWithOrderByRelationField_WithParentSecondaryASC_ExcludesOrphans` | 1089-1146 | ASC order by a secondary relation field without @exhaustive excludes orphan docs from results. |
| `TestQueryWithOrderByRelationField_WithParentPrimaryASC_ExcludesOrphans` | 1148-1207 | ASC order by a primary relation field without @exhaustive excludes orphan publishers from results. |
| `TestQueryWithNestedOrderByRelationField_WithDESCAndLimit_ExcludesOrphans` | 1209-1322 | Nested DESC sub-order without @exhaustive excludes orphan books and returns the 2 most recent linked books. |
| `TestQueryWithNestedOrderByRelationField_WithASCAndLimit_ExcludesOrphans` | 1327-1443 | Nested ASC sub-order without @exhaustive excludes orphan books and returns the 2 oldest linked books. |
| `TestQueryWithNestedOrderByRelationField_ExhaustiveWithASCAndLimit_ShouldIncludeOrphansFirst` | 1445-1551 | Exhaustive nested ASC sub-order with limit places orphan books first before linked books. |
| `TestQueryWithNestedOrderByRelationField_ExhaustiveWithDESCAndLimit_ShouldAppendOrphansLast` | 1553-1659 | Exhaustive nested DESC sub-order with limit satisfies limit from linked books, orphans not needed. |
| `TestQueryWithOrderByRelationField_WithSomeDocsWithoutRelation_ShouldIncludeAll` | 1661-1717 | Exhaustive order by relation field includes all docs, placing orphans first in ASC order. |
| `TestQueryWithFilterOnNullRelation_SecondaryDocWithoutRelation_ShouldReturnOrphans` | 1719-1778 | Exhaustive order on a secondary relation field returns orphan books (no publisher) before linked ones. |

---

### `query_with_relation_sub_order_complex_test.go`

| Test Function | Line | Description |
|---|---|---|
| `TestQueryWithOrderByRelationField_ExhaustiveASCWithLimit_ManyOrphansEarlyTermination` | 21-107 | Exhaustive ASC order by a relation field with limit returns orphan docs via early termination. |
| `TestQueryWithOrderByRelationField_ExhaustiveASC_ManyBooksShowsFullPipeline` | 109-202 | Exhaustive ASC order by a relation field without limit runs both the orphan and source phases. |
| `TestQueryWithOrderByRelationField_ExhaustiveDESCWithLimit_ManyOrphansSkipsOrphanPhase` | 204-289 | Exhaustive DESC order by a relation field with limit is satisfied by linked docs, skipping the orphan phase. |

---

### `query_with_unique_composite_index_filter_test.go`

| Test Function | Line | Description |
|---|---|---|
| `TestQueryWithUniqueCompositeIndex_WithEqualFilter_ShouldFetch` | 21-106 | Equality filter on one or both fields of a unique composite index returns correct docs. |
| `TestQueryWithUniqueCompositeIndex_WithGreaterThanFilterOnFirstField_ShouldFetch` | 108-144 | Greater-than filter on the first field of a unique composite index uses range optimization. |
| `TestQueryWithUniqueCompositeIndex_WithGreaterThanFilterOnSecondField_ShouldFetch` | 146-182 | Greater-than filter on the second field of a unique composite index scans all index entries. |
| `TestQueryWithUniqueCompositeIndex_WithGreaterOrEqualFilterOnFirstField_ShouldFetch` | 184-221 | Greater-or-equal filter on the first field of a unique composite index uses range optimization. |
| `TestQueryWithUniqueCompositeIndex_WithGreaterOrEqualFilterOnSecondField_ShouldFetch` | 223-260 | Greater-or-equal filter on the second field of a unique composite index scans all index entries. |
| `TestQueryWithUniqueCompositeIndex_WithLessThanFilterOnFirstField_ShouldFetch` | 262-298 | Less-than filter on the first field of a unique composite index uses range optimization. |
| `TestQueryWithUniqueCompositeIndex_WithLessThanFilterOnSecondField_ShouldFetch` | 300-336 | Less-than filter on the second field of a unique composite index scans all index entries. |
| `TestQueryWithUniqueCompositeIndex_WithLessOrEqualFilterOnFirstField_ShouldFetch` | 338-375 | Less-or-equal filter on the first field of a unique composite index uses range optimization. |
| `TestQueryWithUniqueCompositeIndex_WithLessOrEqualFilterOnSecondField_ShouldFetch` | 377-414 | Less-or-equal filter on the second field of a unique composite index scans all index entries. |
| `TestQueryWithUniqueCompositeIndex_WithNotEqualFilter_ShouldFetch` | 416-459 | Not-equal filters on both fields of a unique composite index scan all entries and exclude matches. |
| `TestQueryWithUniqueCompositeIndex_WithInForFirstAndEqForRest_ShouldFetchEfficiently` | 461-536 | _in filter on the first composite index field with _eq on the second efficiently fetches matching docs. |
| `TestQueryWithUniqueCompositeIndex_WithInFilter_ShouldFetch` | 538-591 | _in filters on both fields of a unique composite index return docs matching any combination. |
| `TestQueryWithUniqueCompositeIndex_WithNotInFilter_ShouldFetch` | 593-631 | _nin filters on both fields of a unique composite index exclude docs matching any listed value. |
| `TestQueryWithUniqueCompositeIndex_WithLikeFilter_ShouldFetch` | 633-758 | _like filters on both fields of a unique composite index scan all entries for pattern matches. |
| `TestQueryWithUniqueCompositeIndex_WithNotLikeFilter_ShouldFetch` | 760-798 | _nlike filters on both fields of a unique composite index return docs not matching either pattern. |
| `TestQueryWithUniqueCompositeIndex_WithNotCaseInsensitiveLikeFilter_ShouldFetch` | 800-839 | _nilike and _nlike filters on a unique composite index exclude docs matching either pattern. |
| `TestQueryWithUniqueCompositeIndex_IfFirstFieldIsNotInFilter_ShouldNotUseIndex` | 841-868 | Filter only on the second field of a unique composite index does not use the index. |
| `TestQueryWithUniqueCompositeIndex_WithEqualFilterOnNilValueOnFirst_ShouldFetch` | 870-915 | Null equality filter on the first field of a unique composite index returns docs with no name set. |
| `TestQueryWithUniqueCompositeIndex_WithMultipleNilOnFirstFieldAndNilFilter_ShouldFetchAll` | 917-978 | Null equality on the first composite field returns all docs with that field unset, even if second differs. |
| `TestQueryWithUniqueCompositeIndex_WithEqualFilterOnNilValueOnSecond_ShouldFetch` | 980-1038 | Null equality filter on the second field of a unique composite index returns docs with no age set. |
| `TestQueryWithUniqueCompositeIndex_WithMultipleNilOnSecondFieldsAndNilFilter_ShouldFetchAll` | 1040-1110 | Null equality on the second composite field returns all docs with that field unset for the given name. |
| `TestQueryWithUniqueCompositeIndex_WithMultipleNilOnBothFieldsAndNilFilter_ShouldFetchAll` | 1112-1211 | Null equality on both composite index fields returns docs with both fields unset. |
| `TestQueryWithUniqueCompositeIndex_AfterUpdateOnNilFields_ShouldFetch` | 1213-1352 | After updating docs to set or clear composite index fields, null filters return the correct updated docs. |
| `TestQueryWithUniqueCompositeIndex_IfMiddleFieldIsNotInFilter_ShouldIgnoreValue` | 1354-1412 | Filter skipping the middle field of a three-field unique composite index still returns correct results. |

---

### `query_with_unique_index_on_relation_filter_test.go`

| Test Function | Line | Description |
|---|---|---|
| `TestQueryWithUniqueCompositeIndex_WithFilterOnIndexedRelation_ShouldFilter` | 21-75 | Filter on a unique composite index combining a relation field and a scalar returns matching docs. |

---

### `query_with_unique_index_only_filter_test.go`

| Test Function | Line | Description |
|---|---|---|
| `TestQueryWithUniqueIndex_WithEqualFilter_ShouldFetch` | 21-55 | Equality filter on a unique indexed field fetches exactly one matching doc. |
| `TestQueryWithUniqueIndex_WithGreaterThanFilter_ShouldFetch` | 57-92 | Greater-than filter on a unique indexed integer field returns matching docs using range optimization. |
| `TestQueryWithUniqueIndex_WithGreaterOrEqualFilter_ShouldFetch` | 94-130 | Greater-or-equal filter on a unique indexed integer field returns all docs meeting the threshold. |
| `TestQueryWithUniqueIndex_WithLessThanFilter_ShouldFetch` | 132-167 | Less-than filter on a unique indexed integer field returns docs below the threshold. |
| `TestQueryWithUniqueIndex_WithLessOrEqualFilter_ShouldFetch` | 169-205 | Less-or-equal filter on a unique indexed integer field returns docs at or below the threshold. |
| `TestQueryWithUniqueIndex_WithNotEqualFilter_ShouldFetch` | 207-250 | Not-equal filter on a unique indexed string field returns all docs except the matching one. |
| `TestQueryWithUniqueIndex_WithInFilter_ShouldFetch` | 252-288 | _in filter on a unique indexed integer field fetches all docs with values in the list. |
| `TestQueryWithUniqueIndex_WithNotInFilter_ShouldFetch` | 290-328 | _nin filter on a unique indexed integer field excludes docs with listed values. |
| `TestQueryWithUniqueIndex_WithLikeFilter_ShouldFetch` | 330-452 | _like filter with various patterns on a unique indexed string field scans all index entries. |
| `TestQueryWithUniqueIndex_WithNotLikeFilter_ShouldFetch` | 454-495 | _nlike filter on a unique indexed string field returns docs not matching the pattern. |
| `TestQueryWithUniqueIndex_WithNotCaseInsensitiveLikeFilter_ShouldFetch` | 497-539 | _nilike filter on a unique indexed field excludes docs matching the case-insensitive pattern. |
| `TestQueryWithUniqueIndex_IfNoMatch_ReturnEmptyResult` | 541-574 | Equality filter on a unique indexed field with no matching doc returns an empty result. |
| `TestQueryWithUniqueIndex_WithEqualFilterOnNilValue_ShouldFetch` | 576-622 | Equality filter for null on a unique indexed field returns docs with no value set. |
| `TestQueryWithUniqueIndex_WithEqualFilterOnZero_ShouldNotFetchNil` | 624-675 | Equality filter for zero on a unique indexed integer field does not return docs with a nil value. |
| `TestQueryWithUniqueIndex_WithNotEqualFilterOnNilValue_ShouldFetch` | 677-729 | Not-equal-null filter on a unique indexed integer field returns docs with any non-nil value. |
| `TestQueryWithUniqueIndex_WithMultipleNilValuesAndEqualFilter_ShouldFetch` | 731-782 | Null equality filter on a unique indexed field returns all docs with no value set. |
| `TestQueryWithUniqueIndex_WithDateTimeField_ShouldIndex` | 784-830 | Equality filter on a unique indexed DateTime field uses the index for exact match lookup. |

---

### `update_unique_composite_test.go`

| Test Function | Line | Description |
|---|---|---|
| `TestUniqueCompositeIndexUpdate_UponUpdatingDocWithExistingFieldValue_ShouldSucceed` | 21-55 | Updating a non-indexed field on a doc with a unique composite index succeeds without conflict. |

---

### `update_unique_test.go`

| Test Function | Line | Description |
|---|---|---|
| `TestUniqueIndexUpdate_UponUpdatingDocNonIndexedField_ShouldSucceed` | 21-53 | Updating a non-indexed field on a doc with a unique index succeeds without conflict. |
