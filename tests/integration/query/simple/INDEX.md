# Index: `tests/integration/query/simple`

## Overview

This folder contains integration tests for basic (non-relational) collection queries in DefraDB. Tests cover field retrieval, aliases, multi-row results, docID and CID filters, limit/offset pagination, ordering across all scalar types, aggregate functions (AVG, SUM, COUNT, MIN, MAX) with filters/limits/offsets and nested groups, groupBy across field types with GROUP child selections, GraphQL fragments and variables, named operations, similarity queries on vector fields, null-input tolerance, branchable collection versioning, embedded _version commit metadata, and CRDT counter types. A `with_filter/` subdirectory provides exhaustive filter operator coverage.

## Test Index

### `simple_test.go`

Baseline queries verifying field retrieval, aliases, multi-row results, undefined fields, default values, and cross-collection isolation.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimple` | 21-53 | Query a single document returns all requested scalar fields. |
| `TestQuerySimpleWithAlias` | 55-85 | Query with field aliases returns results under the aliased names. |
| `TestQuerySimpleWithMultipleRows` | 87-127 | Query with multiple documents returns all rows. |
| `TestQuerySimpleWithUndefinedField` | 129-146 | Querying a field that does not exist on the type returns an error. |
| `TestQuerySimpleWithSomeDefaultValues` | 148-183 | Document created with only Name set returns nil for all other fields. |
| `TestQuerySimpleWithDefaultValue` | 185-218 | Document created with empty body returns nil for every field. |
| `TestQuerySimple_WithDeletedDocsInCollection2_ShouldNotYieldDeletedDocsOnCollection1Query` | 222-298 | Deleting docs from collection 2 does not affect results for collection 1. |

### `with_average_filter_test.go`

Top-level AVG aggregate combined with field and DateTime filters.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithAverageWithFilter` | 21-55 | Top-level AVG with an Age filter averages only matching documents. |
| `TestQuerySimpleWithAverageWithDateTimeFilter` | 57-94 | Top-level AVG with a DateTime filter averages only documents past the threshold. |

### `with_average_order_test.go`

Ordering documents by an aliased AVG aggregate result.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithAverageWithOrder_Succeeds` | 21-80 | Order documents by an aliased multi-field AVG in ASC and DESC directions. |

### `with_average_test.go`

Top-level AVG aggregate on integer and float fields, empty collections, undefined objects, and aliases.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithAverageOnUndefinedObject` | 21-35 | AVG with no collection argument returns an error. |
| `TestQuerySimpleWithAverageOnUndefinedField` | 37-51 | AVG on a collection without specifying a field returns an error. |
| `TestQuerySimpleWithAverageOnEmptyCollection` | 53-69 | AVG on an empty collection returns zero. |
| `TestQuerySimpleWithAverage` | 71-99 | Top-level AVG of an integer field returns the correct average. |
| `TestQuerySimple_WithAliasedAverage_OnEmptyCollection_Succeeds` | 101-117 | Aliased AVG on an empty collection returns zero under the alias name. |

### `with_cid_branchable_test.go`

Querying a branchable collection at specific CIDs: first, middle, and last.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithCidOfBranchableCollection_FirstCid` | 21-67 | Query a branchable collection at the first CID returns the initial document state. |
| `TestQuerySimpleWithCidOfBranchableCollection_MiddleCid` | 69-115 | Query a branchable collection at a middle CID returns the document at that point. |
| `TestQuerySimpleWithCidOfBranchableCollection_LastCid` | 117-167 | Query a branchable collection at the latest CID returns all documents at that state. |

### `with_cid_doc_id_branchable_test.go`

Querying a branchable collection filtering by both CID and docID.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithCidOfBranchableCollectionAndDocID` | 21-70 | Query a branchable collection with both CID and docID returns only the matching document. |

### `with_cid_doc_id_test.go`

Combining CID and docID filters: invalid inputs, version retrieval after updates, schema version metadata, and CRDT counter types.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithInvalidCidAndInvalidDocID` | 22-48 | Query with an invalid CID and invalid docID returns an invalid cid error. |
| `TestQuerySimpleWithUnknownCidAndInvalidDocID` | 52-78 | Query with an unknown CID and invalid docID returns a block-not-found error. |
| `TestQuerySimpleWithCidAndDocID` | 80-117 | Query with a valid CID and matching docID returns the document at that version. |
| `TestQuerySimpleWithUpdateAndFirstCidAndDocID` | 119-161 | Query with the first CID and docID after an update returns the original document state. |
| `TestQuerySimpleWithUpdateAndLastCidAndDocID` | 163-205 | Query with the latest CID and docID after an update returns the updated state. |
| `TestQuerySimpleWithUpdateAndMiddleCidAndDocID` | 207-266 | Query with a middle CID and docID returns the document at that intermediate version. |
| `TestQuerySimpleWithUpdateAndFirstCidAndDocIDAndSchemaVersion` | 268-318 | Query with first CID returns the correct schema collectionVersionId in _version. |
| `TestCidAndDocIDQuery_ContainsPNCounterWithIntKind_NoError` | 321-375 | Query at the first CID with a pncounter int field returns the initial counter value. |
| `TestCidAndDocIDQuery_ContainsPNCounterWithFloatKind_NoError` | 378-432 | Query at the first CID with a pncounter float field returns the initial float value. |
| `TestCidAndDocIDQuery_ContainsPCounterWithIntKind_NoError` | 435-484 | Query at the first CID with a pcounter int field returns the initial counter value. |
| `TestCidAndDocIDQuery_ContainsPCounterWithFloatKind_NoError` | 487-536 | Query at the first CID with a pcounter float field returns the initial float value. |

### `with_cid_test.go`

Filtering by CID: invalid, valid, unknown, multi-doc, counter types, delete operations, and list inputs.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithInvalidCid` | 22-44 | Query with an invalid CID string returns an invalid cid error. |
| `TestQuerySimpleWithCid` | 46-82 | Query with a valid CID returns the document at that exact version. |
| `TestQuerySimple_UnknownCid` | 84-109 | Query with an unknown but valid-format CID returns a block-not-found error. |
| `TestQuerySimpleWithCid_MultipleDocs` | 111-152 | Query with a CID that belongs to one document only returns that document. |
| `TestQuerySimple_WithCIDAndCounterAfterUpdate_ShouldSucceed` | 154-197 | Query with a CID after updating a counter field returns the cumulative counter value. |
| `TestQuerySimple_WithCidAfterDeleteOperation_ShouldReturnUser` | 199-241 | Query with CID and showDeleted after deletion returns the deleted document. |
| `TestQuerySimple_ListOfOneCID` | 243-279 | Query with a one-element CID list returns the document at that version. |
| `TestQuerySimple_MultipleCIDs` | 281-311 | Query with multiple CIDs in a list returns an unsupported error. |

### `with_count_filter_test.go`

Top-level COUNT aggregate combined with field and DateTime filters.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithCountWithFilter` | 21-55 | Top-level COUNT with an Age filter counts only matching documents. |
| `TestQuerySimpleWithCountWithDateTimeFilter` | 57-94 | Top-level COUNT with a DateTime filter counts only documents past the threshold. |

### `with_count_test.go`

Top-level COUNT aggregate on empty collections, undefined objects, and aliases.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithCountOnUndefined` | 21-35 | COUNT with no collection argument returns an error. |
| `TestQuerySimpleWithCountOnEmptyCollection` | 37-53 | COUNT on an empty collection returns zero. |
| `TestQuerySimpleWithCount` | 55-83 | Top-level COUNT on a collection returns the total number of documents. |
| `TestQuerySimple_WithAliasedCount_OnEmptyCollection_Succeeds` | 85-101 | Aliased COUNT on an empty collection returns zero under the alias name. |

### `with_deleted_field_test.go`

Querying with showDeleted to return soft-deleted documents.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimple_WithDeletedField` | 21-73 | Query with showDeleted returns deleted documents with _deleted set to true. |

### `with_doc_id_filter_test.go`

Filtering by _docID equality in the filter block.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithDocIDFilterBlock` | 21-57 | Filter by _docID equality returns only the matching document. |

### `with_doc_id_test.go`

Filtering by a single docID: not-found, single-doc, and multi-doc scenarios.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithDocIDFilter_TargetNotFound` | 21-46 | Query with a docID that does not exist returns an empty result. |
| `TestQuerySimpleWithDocIDFilter_SingleDocumentTargetFound` | 48-78 | Query with a valid single docID returns only that document. |
| `TestQuerySimpleWithDocIDFilter_MultipleDocumentsTargetFound` | 80-116 | Query with a docID shared by multiple documents returns all matching documents. |

### `with_doc_ids_test.go`

Filtering by a list of docIDs: single not-found, single found, partial match, full match, and empty list.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryWithDocIDsFilter_SingleTargetNotFound` | 21-46 | Filter by a single docID that does not exist returns an empty result. |
| `TestQueryWithDocIDsFilter_SingleTargetFound` | 48-78 | Filter by a single valid docID returns only that document. |
| `TestQuerySimpleWithDocIDsFilter_OneFoundFromMultipleTargets` | 80-116 | Filter with multiple docIDs where one matches returns only the matching document. |
| `TestQuerySimpleWithDocIDsFilter_AllFoundFromMultipleTargets` | 118-165 | Filter with multiple docIDs that all match returns all matching documents. |
| `TestQuerySimpleReturnsNothinGivenEmptyDocIDsFilter` | 167-192 | Filter with an empty docID list returns no results. |

### `with_fragments_test.go`

GraphQL fragment spreads: named fragments, nested fragments, combined with field selection, missing fragments, invalid fields, aggregates, variables, and inline fragments.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimple_WithFragments_Succeeds` | 23-72 | Named fragment spread returns the same fields as selecting them inline. |
| `TestQuerySimple_WithNestedFragments_Succeeds` | 74-121 | Nested fragment spreads compose correctly and return all referenced fields. |
| `TestQuerySimple_WithFragmentSpreadAndSelect_Succeeds` | 123-167 | Combining fragment spread with explicit field selection returns all fields without duplication. |
| `TestQuerySimple_WithMissingFragment_ReturnsError` | 169-197 | Using an undefined fragment name returns a GraphQL validation error. |
| `TestQuerySimple_WithFragmentWithInvalidField_ReturnsError` | 199-230 | A fragment referencing a non-existent field returns a schema validation error. |
| `TestQuerySimple_WithFragmentWithAggregate_Succeeds` | 232-263 | A fragment containing an aggregate field correctly returns the aggregated value. |
| `TestQuerySimple_WithFragmentWithVariables_Succeeds` | 265-309 | A fragment used with GraphQL variables substitutes values correctly. |
| `TestQuerySimple_WithInlineFragment_Succeeds` | 311-354 | An inline fragment on the correct type returns the selected fields. |

### `with_group_aggregate_alias_filter_test.go`

Filtering grouped results by aliased aggregate values (AVG, SUM, MIN, MAX, COUNT).

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimple_WithGroupAverageAliasFilter_FiltersResults` | 21-75 | Filter grouped results by an aliased AVG value excludes groups below the threshold. |
| `TestQuerySimple_WithGroupSumAliasFilter_FiltersResults` | 77-131 | Filter grouped results by an aliased SUM value excludes groups below the threshold. |
| `TestQuerySimple_WithGroupMinAliasFilter_FiltersResults` | 133-187 | Filter grouped results by an aliased MIN value excludes groups below the threshold. |
| `TestQuerySimple_WithGroupMaxAliasFilter_FiltersResults` | 189-243 | Filter grouped results by an aliased MAX value excludes groups above the threshold. |
| `TestQuerySimple_WithGroupCountAliasFilter_FiltersResults` | 245-305 | Filter grouped results by an aliased COUNT value excludes groups below the threshold. |

### `with_group_average_count_test.go`

Group by string computing both AVG and COUNT on a child integer field.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildIntegerAverageAndCount` | 25-74 | Group by string and compute AVG and COUNT of a child integer field. |

### `with_group_average_filter_test.go`

Child AVG with various filter combinations: simple, DateTime, parent filter, matching filter, different filter, multiple filters, and nil members.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildAverageWithFilter` | 21-67 | Group by string with child AVG filtered on age returns the average of matching sub-documents. |
| `TestQuerySimpleWithGroupByStringWithRenderedGroupAndChildAverageWithFilter` | 69-131 | Group by string with rendered GROUP and filtered child AVG returns the average alongside group members. |
| `TestQuerySimpleWithGroupByStringWithRenderedGroupAndChildAverageWithDateTimeFilter` | 133-199 | Group by string with rendered GROUP and child AVG filtered on a DateTime field returns correct averages. |
| `TestQuerySimpleWithGroupByStringWithRenderedGroupWithFilterAndChildAverageWithMatchingFilter` | 201-256 | Group-level filter and matching child AVG filter return only documents satisfying both conditions. |
| `TestQuerySimpleWithGroupByStringWithRenderedGroupWithFilterAndChildAverageWithMatchingDateTimeFilter` | 258-317 | Group-level DateTime filter and matching child AVG filter return documents satisfying both conditions. |
| `TestQuerySimpleWithGroupByStringWithRenderedGroupWithFilterAndChildAverageWithDifferentFilter` | 319-378 | Group-level filter and different child AVG filter independently narrow the rendered group and average. |
| `TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildAveragesWithDifferentFilters` | 380-429 | Two child AVGs with different filters on the same field return independent averages per group. |
| `TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildAverageWithFilterAndNilItem` | 431-489 | Child AVG with a filter when some group members are nil returns the average of non-nil matching items. |

### `with_group_average_limit_offset_test.go`

Child AVG with limit and offset applied to sub-documents per group.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildIntegerAverageWithLimitAndOffset` | 21-73 | Child AVG with limit and offset considers only the windowed sub-documents. |

### `with_group_average_limit_test.go`

Child AVG with a limit applied to sub-documents per group.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildIntegerAverageWithLimit` | 21-74 | Child AVG with a limit considers only the first N sub-documents per group. |

### `with_group_average_max_test.go`

Combining child AVG and MAX aggregates: nested groups and flat group results.

| Test Function | Line | Description |
|---|---|---|
| `TestQuery_SimpleWithGroupByStringWithInnerGroupBooleanAndMaxOfAverageOfInt_Succeeds` | 21-115 | Nested groups compute the MAX of a child AVG of integers across boolean sub-groups. |
| `TestQuerySimple_WithGroupByStringWithoutRenderedGroupAndChildIntegerAverageAndMax_Succeeds` | 117-167 | Group by string computes both AVG and MAX of a child integer field without rendering the group. |

### `with_group_average_min_test.go`

Combining child AVG and MIN aggregates: nested groups and flat group results.

| Test Function | Line | Description |
|---|---|---|
| `TestQuery_SimpleWithGroupByStringWithInnerGroupBooleanAndMinOfAverageOfInt_Succeeds` | 21-115 | Nested groups compute the MIN of a child AVG of integers across boolean sub-groups. |
| `TestQuerySimple_WithGroupByStringWithoutRenderedGroupAndChildIntegerAverageAndMin_Succeeds` | 117-167 | Group by string computes both AVG and MIN of a child integer field without rendering the group. |

### `with_group_average_sum_test.go`

Combining child AVG and SUM aggregates: nested groups and flat group results.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithGroupByStringWithInnerGroupBooleanAndSumOfCountOfInt` | 21-115 | Nested groups compute the SUM of a child COUNT of integers across boolean sub-groups. |
| `TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildIntegerAverageAndSum` | 121-171 | Group by string computes both AVG and SUM of a child integer field without rendering the group. |

### `with_group_average_test.go`

Child AVG aggregate on integer and float fields, nil values, nested groups, and empty collections.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndAverageOfUndefined` | 21-44 | Child AVG on an undefined field returns an error. |
| `TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildIntegerAverageOnEmptyCollection` | 46-65 | Group by string with child AVG on an empty collection returns zero. |
| `TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildIntegerAverage` | 67-114 | Group by string with child integer AVG returns the correct per-group average. |
| `TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildNilAverage` | 116-162 | Group by string with child AVG where values include nil returns the average of non-nil values. |
| `TestQuerySimpleWithGroupByStringWithInnerGroupBooleanAndAverageOfAverageOfInt` | 164-258 | Nested groups with AVG of a child AVG returns the correct deeply aggregated value. |
| `TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildEmptyFloatAverage` | 260-305 | Group by string with child float AVG on an empty collection returns zero. |
| `TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildFloatAverage` | 307-353 | Group by string with child float AVG returns the correct per-group float average. |
| `TestQuerySimpleWithGroupByStringWithInnerGroupBooleanAndAverageOfAverageOfFloat` | 355-449 | Nested groups with AVG of a child float AVG returns the correct deeply aggregated value. |
| `TestQuerySimpleWithGroupByStringWithInnerGroupBooleanAndAverageOfAverageOfAverageOfFloat` | 451-582 | Triple-nested groups with AVG of AVG of AVG returns the correct value. |

### `with_group_count_filter_test.go`

Child COUNT with various filter combinations: simple, parent filter, matching filter, different filter, and multiple filters.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithGroupByNumberWithoutRenderedGroupAndChildCountWithFilter` | 21-67 | Group by number with filtered child COUNT counts only matching sub-documents. |
| `TestQuerySimpleWithGroupByNumberWithRenderedGroupAndChildCountWithFilter` | 69-132 | Rendered GROUP and filtered child COUNT shows sub-documents alongside filtered counts. |
| `TestQuerySimpleWithGroupByNumberWithRenderedGroupWithFilterAndChildCountWithMatchingFilter` | 134-189 | Group filter and matching child COUNT filter produce identical filtered counts. |
| `TestQuerySimpleWithGroupByNumberWithRenderedGroupWithFilterAndChildCountWithDifferentFilter` | 191-246 | Group filter and different child COUNT filter produce distinct filtered counts. |
| `TestQuerySimpleWithGroupByNumberWithoutRenderedGroupAndChildCountsWithDifferentFilters` | 248-297 | Two child COUNTs with different filters return independent counts in the same group. |

### `with_group_count_limit_offset_test.go`

Child COUNT with limit and offset applied to sub-documents per group.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithGroupByNumberWithoutRenderedGroupAndChildCountWithLimitAndOffset` | 21-67 | Child COUNT with limit and offset counts only the windowed sub-documents per group. |
| `TestQuerySimpleWithGroupByNumberWithRenderedGroupWithLimitAndChildCountWithLimitAndOffset` | 69-138 | Group limit combined with child COUNT limit and offset independently window the top-level and child results. |

### `with_group_count_limit_test.go`

Child COUNT with a limit applied to sub-documents per group.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithGroupByNumberWithoutRenderedGroupAndChildCountWithLimit` | 21-67 | Child COUNT with a limit counts only the first N sub-documents per group. |
| `TestQuerySimpleWithGroupByNumberWithRenderedGroupWithLimitAndChildCountWithLimit` | 69-138 | Top-level group limit combined with child COUNT limit independently cap their respective result sets. |

### `with_group_count_max_test.go`

Nested groups computing MAX of a child COUNT across boolean sub-groups.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimple_WithGroupByStringWithInnerGroupBooleanAndMaxOfCount_Succeeds` | 21-115 | Nested groups compute the MAX of a child COUNT across boolean sub-groups. |

### `with_group_count_min_test.go`

Nested groups computing MIN of a child COUNT across boolean sub-groups.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimple_WithGroupByStringWithInnerGroupBooleanAndMinOfCount_Succeeds` | 21-115 | Nested groups compute the MIN of a child COUNT across boolean sub-groups. |

### `with_group_count_sum_test.go`

Nested groups computing SUM of a child COUNT across boolean sub-groups.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithGroupByStringWithInnerGroupBooleanAndSumOfCount` | 21-115 | Nested groups compute the SUM of a child COUNT across boolean sub-groups. |

### `with_group_count_test.go`

Child COUNT aggregate: undefined group, empty collections, rendered group, undefined fields, aliases, and duplicated aliased counts.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithoutGroupByWithCountOnGroup` | 21-44 | Using COUNT(GROUP) without groupBy returns a schema error. |
| `TestQuerySimpleWithGroupByNumberWithCountOnInnerNonExistantGroup` | 46-72 | COUNT on a non-existent inner GROUP returns a schema validation error. |
| `TestQuerySimpleWithGroupByNumberWithoutRenderedGroupAndChildCount` | 74-120 | Group by number with child COUNT returns the count of sub-documents per group. |
| `TestQuerySimpleWithGroupByNumberWithoutRenderedGroupAndChildCountOnEmptyCollection` | 122-141 | Group by number with child COUNT on an empty collection returns zero counts. |
| `TestQuerySimpleWithGroupByNumberWithRenderedGroupAndChildCount` | 143-206 | Group by number with rendered GROUP and child COUNT returns the count alongside grouped documents. |
| `TestQuerySimpleWithGroupByNumberWithUndefinedField` | 208-243 | Group by an undefined field returns a schema validation error. |
| `TestQuerySimpleWithGroupByNumberWithoutRenderedGroupAndAliasesChildCount` | 245-291 | Child COUNT with an alias returns the count under the alias name. |

### `with_group_doc_id_test.go`

GROUP child selection returning the docID of each grouped document.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithGroupByWithGroupWithDocID` | 21-74 | Group by a field with GROUP selection returns the docID of each grouped document. |

### `with_group_doc_ids_test.go`

GROUP child selection returning multiple docIDs of grouped documents.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithGroupByWithGroupWithDocIDs` | 21-83 | Group by a field with GROUP selection returns the docIDs of grouped documents. |

### `with_group_filter_test.go`

Filters applied within GROUP children and at the parent level across single and nested groups.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithGroupByStringWithGroupNumberFilter` | 21-88 | Apply a numeric filter within a GROUP child to return only matching sub-documents. |
| `TestQuerySimpleWithGroupByStringWithGroupNumberWithParentFilter` | 90-153 | Apply a parent-level numeric filter to restrict which groups appear in the result. |
| `TestQuerySimpleWithGroupByStringWithUnrenderedGroupNumberWithParentFilter` | 155-205 | Apply a parent-level filter when the GROUP is not rendered returns correctly filtered group metadata. |
| `TestQuerySimpleWithGroupByStringWithMultipleGroupNumberFilter` | 303-384 | Apply multiple group-level filters on numeric fields. |

### `with_group_limit_offset_test.go`

Group-level limit and offset applied to GROUP children.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithGroupByNumberWithGroupLimitAndOffset` | 21-74 | Apply limit and offset within each GROUP to control nested sub-documents. |
| `TestQuerySimpleWithGroupByNumberWithLimitAndOffsetAndWithGroupLimitAndOffset` | 76-120 | Combine top-level limit and offset with group-level limit and offset. |

### `with_group_limit_test.go`

Group-level limit applied to GROUP children with single and multiple groups at different limits.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithGroupByNumberWithGroupLimit` | 21-78 | Apply a limit within each GROUP to cap the number of nested sub-documents. |
| `TestQuerySimpleWithGroupByNumberWithMultipleGroupsWithDifferentLimits` | 80-153 | Different GROUP limits applied to different groups return independent counts. |
| `TestQuerySimpleWithGroupByNumberWithLimitAndGroupWithHigherLimit` | 155-207 | Top-level limit lower than group limit; top-level limit controls number of groups. |
| `TestQuerySimpleWithGroupByNumberWithLimitAndGroupWithLowerLimit` | 209-272 | Top-level limit higher than group limit; group limit caps sub-documents independently. |

### `with_group_max_filter_test.go`

Child MAX with various filter combinations: simple, parent filter, matching filter, different filter, and multiple filters.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimple_WithGroupByNumberWithoutRenderedGroupAndChildMaxWithFilter_Succeeds` | 21-67 | Group by number with filtered child MAX returns the maximum of matching sub-documents. |
| `TestQuerySimple_WithGroupByNumberWithRenderedGroupAndChildMaxWithFilter_Succeeds` | 69-132 | Rendered GROUP and filtered child MAX shows sub-documents alongside filtered maximums. |
| `TestQuerySimple_WithGroupByNumberWithRenderedGroupWithFilterAndChildMaxWithMatchingFilter_Succeeds` | 134-189 | Group filter and matching child MAX filter produce identical filtered maximums. |
| `TestQuerySimple_WithGroupByNumberWithRenderedGroupWithFilterAndChildMaxWithDifferentFilter_Succeeds` | 191-246 | Group filter and different child MAX filter produce distinct filtered maximums. |
| `TestQuerySimple_WithGroupByNumberWithoutRenderedGroupAndChildMaxWithDifferentFilters_Succeeds` | 248-297 | Two child MAX aggregates with different filters return independent maximums in the same group. |

### `with_group_max_limit_offset_test.go`

Child MAX with limit and offset applied to sub-documents per group.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimple_WithGroupByStringWithoutRenderedGroupAndChildIntegerMaxWithLimitAndOffset_Succeeds` | 21-74 | Child MAX with limit and offset considers only the windowed sub-documents. |

### `with_group_max_limit_test.go`

Child MAX with a limit applied to sub-documents per group.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimple_WithGroupByStringWithoutRenderedGroupAndChildIntegerMaxWithLimit_Succeeds` | 21-74 | Child MAX with a limit considers only the first N sub-documents per group. |

### `with_group_max_test.go`

Child MAX aggregate on integer and float fields, nil values, nested groups, and empty collections.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimple_WithGroupByStringWithoutRenderedGroupAndMaxOfUndefined_ReturnsError` | 21-44 | Child MAX on an undefined field returns an error. |
| `TestQuerySimple_WithGroupByStringWithoutRenderedGroupAndChildIntegerMaxOnEmptyCollection_Succeeds` | 46-65 | Group by string with child MAX on an empty collection returns nil. |
| `TestQuerySimple_WithGroupByStringWithoutRenderedGroupAndChildIntegerMax_Succeeds` | 67-114 | Group by string with child integer MAX returns the correct per-group maximum. |
| `TestQuerySimple_WithGroupByStringWithoutRenderedGroupAndChildNilMax_Succeeds` | 116-162 | Group by string with child MAX where values include nil returns the max of non-nil values. |
| `TestQuerySimple_WithGroupByStringWithInnerGroupBooleanAndMaxOfMaxOfInt_Succeeds` | 164-258 | Nested groups with MAX of child MAX returns the correct deeply aggregated maximum. |
| `TestQuerySimple_WithGroupByStringWithoutRenderedGroupAndChildEmptyFloatMax_Succeeds` | 260-305 | Group by string with child float MAX on an empty collection returns nil. |
| `TestQuerySimple_WithGroupByStringWithoutRenderedGroupAndChildFloatMax_Succeeds` | 307-353 | Group by string with child float MAX returns the correct per-group float maximum. |
| `TestQuerySimple_WithGroupByStringWithInnerGroupBooleanAndMaxOfMaxOfFloat_Succeeds` | 355-449 | Nested groups with MAX of child float MAX returns the correct deeply aggregated maximum. |
| `TestQuerySimple_WithGroupByStringWithInnerGroupBooleanAndMaxOfMaxOfMaxOfFloat_Succeeds` | 451-582 | Triple-nested groups with MAX of MAX of MAX returns the correct deeply aggregated float maximum. |

### `with_group_min_filter_test.go`

Child MIN with various filter combinations: simple, parent filter, matching filter, different filter, and multiple filters.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimple_WithGroupByNumberWithoutRenderedGroupAndChildMinWithFilter_Succeeds` | 21-67 | Group by number with filtered child MIN returns the minimum of matching sub-documents. |
| `TestQuerySimple_WithGroupByNumberWithRenderedGroupAndChildMinWithFilter_Succeeds` | 69-132 | Rendered GROUP and filtered child MIN shows sub-documents alongside filtered minimums. |
| `TestQuerySimple_WithGroupByNumberWithRenderedGroupWithFilterAndChildMinWithMatchingFilter_Succeeds` | 134-189 | Group filter and matching child MIN filter produce identical filtered minimums. |
| `TestQuerySimple_WithGroupByNumberWithRenderedGroupWithFilterAndChildMinWithDifferentFilter_Succeeds` | 191-246 | Group filter and different child MIN filter produce distinct filtered minimums. |
| `TestQuerySimple_WithGroupByNumberWithoutRenderedGroupAndChildMinWithDifferentFilters_Succeeds` | 248-297 | Two child MIN aggregates with different filters return independent minimums in the same group. |

### `with_group_min_limit_offset_test.go`

Child MIN with limit and offset applied to sub-documents per group.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimple_WithGroupByStringWithoutRenderedGroupAndChildIntegerMinWithLimitAndOffset_Succeeds` | 21-74 | Child MIN with limit and offset considers only the windowed sub-documents. |

### `with_group_min_limit_test.go`

Child MIN with a limit applied to sub-documents per group.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimple_WithGroupByStringWithoutRenderedGroupAndChildIntegerMinWithLimit_Succeeds` | 21-74 | Child MIN with a limit considers only the first N sub-documents per group. |

### `with_group_min_test.go`

Child MIN aggregate on integer and float fields, nil values, nested groups, and empty collections.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimple_WithGroupByStringWithoutRenderedGroupAndMinOfUndefined_ReturnsError` | 21-44 | Child MIN on an undefined field returns an error. |
| `TestQuerySimple_WithGroupByStringWithoutRenderedGroupAndChildIntegerMinOnEmptyCollection_Succeeds` | 46-65 | Group by string with child MIN on an empty collection returns nil. |
| `TestQuerySimple_WithGroupByStringWithoutRenderedGroupAndChildIntegerMin_Succeeds` | 67-114 | Group by string with child integer MIN returns the correct per-group minimum. |
| `TestQuerySimple_WithGroupByStringWithoutRenderedGroupAndChildNilMin_Succeeds` | 116-162 | Group by string with child MIN where values include nil returns the min of non-nil values. |
| `TestQuerySimple_WithGroupByStringWithInnerGroupBooleanAndMinOfMinOfInt_Succeeds` | 164-258 | Nested groups with MIN of child MIN returns the correct deeply aggregated minimum. |
| `TestQuerySimple_WithGroupByStringWithoutRenderedGroupAndChildEmptyFloatMin_Succeeds` | 260-305 | Group by string with child float MIN on an empty collection returns nil. |
| `TestQuerySimple_WithGroupByStringWithoutRenderedGroupAndChildFloatMin_Succeeds` | 307-353 | Group by string with child float MIN returns the correct per-group float minimum. |
| `TestQuerySimple_WithGroupByStringWithInnerGroupBooleanAndMinOfMinOfFloat_Succeeds` | 355-449 | Nested groups with MIN of child float MIN returns the correct deeply aggregated minimum. |
| `TestQuerySimple_WithGroupByStringWithInnerGroupBooleanAndMinOfMinOfMinOfFloat_Succeeds` | 451-582 | Triple-nested groups with MIN of MIN of MIN returns the correct deeply aggregated float minimum. |

### `with_group_order_test.go`

Ordering within GROUP children and combining parent and child order directions across nested groups.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithGroupByStringWithGroupNumberWithGroupOrder` | 21-94 | Order within each GROUP by a numeric field in ascending order. |
| `TestQuerySimpleWithGroupByStringWithGroupNumberWithGroupOrderDescending` | 96-169 | Order within each GROUP by a numeric field in descending order. |
| `TestQuerySimpleWithGroupByStringAndOrderDescendingWithGroupNumberWithGroupOrder` | 171-244 | Top-level descending order combined with ascending GROUP order returns correctly sorted results. |
| `TestQuerySimpleWithGroupByStringWithInnerGroupBooleanThenInnerOrderDescending` | 246-356 | Nested groups with descending inner order returns correctly sorted nested sub-documents. |

### `with_group_sum_filter_test.go`

Child SUM with various filter combinations: simple, parent filter, matching filter, different filter, and multiple filters.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithGroupByNumberWithoutRenderedGroupAndChildSumWithFilter` | 21-67 | Group by number with filtered child SUM returns the sum of matching sub-documents. |
| `TestQuerySimpleWithGroupByNumberWithRenderedGroupAndChildSumWithFilter` | 69-132 | Rendered GROUP and filtered child SUM shows sub-documents alongside filtered sums. |
| `TestQuerySimpleWithGroupByNumberWithRenderedGroupWithFilterAndChildSumWithMatchingFilter` | 134-189 | Group filter and matching child SUM filter produce identical filtered sums. |
| `TestQuerySimpleWithGroupByNumberWithRenderedGroupWithFilterAndChildSumWithDifferentFilter` | 191-246 | Group filter and different child SUM filter produce distinct filtered sums. |
| `TestQuerySimpleWithGroupByNumberWithoutRenderedGroupAndChildSumsWithDifferentFilters` | 248-297 | Two child SUM aggregates with different filters return independent sums in the same group. |

### `with_group_sum_limit_offset_test.go`

Child SUM with limit and offset applied to sub-documents per group.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildIntegerSumWithLimitAndOffset` | 21-74 | Child SUM with limit and offset considers only the windowed sub-documents. |

### `with_group_sum_limit_test.go`

Child SUM with a limit applied to sub-documents per group.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildIntegerSumWithLimit` | 21-74 | Child SUM with a limit sums only the first N sub-documents per group. |

### `with_group_sum_test.go`

Child SUM aggregate on integer and float fields, nil values, nested groups, and empty collections.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndSumOfUndefined` | 21-44 | Child SUM on an undefined field returns an error. |
| `TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildIntegerSumOnEmptyCollection` | 46-65 | Group by string with child SUM on an empty collection returns zero. |
| `TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildIntegerSum` | 67-114 | Group by string with child integer SUM returns the correct per-group sum. |
| `TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildNilSum` | 116-162 | Group by string with child SUM where values include nil sums only non-nil values. |
| `TestQuerySimpleWithGroupByStringWithInnerGroupBooleanAndSumOfSumOfInt` | 164-258 | Nested groups with SUM of child SUM returns the correct deeply aggregated sum. |
| `TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildEmptyFloatSum` | 260-305 | Group by string with child float SUM on an empty collection returns zero. |
| `TestQuerySimpleWithGroupByStringWithoutRenderedGroupAndChildFloatSum` | 307-353 | Group by string with child float SUM returns the correct per-group float sum. |
| `TestQuerySimpleWithGroupByStringWithInnerGroupBooleanAndSumOfSumOfFloat` | 355-449 | Nested groups with SUM of child float SUM returns the correct deeply aggregated sum. |
| `TestQuerySimpleWithGroupByStringWithInnerGroupBooleanAndSumOfSumOfSumOfFloat` | 451-582 | Triple-nested groups with SUM of SUM of SUM returns the correct deeply aggregated float sum. |

### `with_group_test.go`

GroupBy across different field types: number, DateTime, string, boolean, nested groups, compound keys, undefined fields, and error cases.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithGroupByEmpty` | 21-65 | groupBy with an empty field list returns an error. |
| `TestQuerySimpleWithGroupByNumber` | 67-120 | Group documents by an integer field returns one entry per distinct value. |
| `TestQuerySimpleWithGroupByDateTime` | 122-175 | Group documents by a DateTime field returns one entry per distinct datetime. |
| `TestQuerySimpleWithGroupByNumberWithGroupString` | 177-251 | Group by integer with a GROUP sub-selection returns nested documents per group. |
| `TestQuerySimpleWithGroupByWithoutGroupedFieldSelectedWithInnerGroup` | 253-327 | Grouping without selecting the grouped field still returns correct GROUP children. |
| `TestQuerySimpleWithGroupByString` | 329-403 | Group documents by a string field returns one entry per distinct string value. |
| `TestQuerySimpleWithGroupByStringWithInnerGroupBoolean` | 405-516 | Group by string with a nested GROUP by boolean returns doubly-nested groups. |
| `TestQuerySimpleWithGroupByStringThenBoolean` | 518-616 | Group by string then boolean returns correctly nested two-level groups. |
| `TestQuerySimpleWithGroupByBooleanThenNumber` | 618-716 | Group by boolean then number returns correctly nested two-level groups. |
| `TestQuerySimpleWithGroupByNumberOnUndefined` | 718-760 | Group by an undefined field with empty collection returns no groups. |
| `TestQuerySimpleWithGroupByNumberOnUndefinedWithChildren` | 762-820 | Group by an undefined field with children returns no groups and no sub-documents. |
| `TestQuerySimpleErrorsWithNonGroupFieldsSelected` | 822-839 | Selecting a non-grouped field outside of GROUP returns an error. |

### `with_group_typename_test.go`

GroupBy with __typename in the group and GROUP child selections.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithGroupByWithTypeName` | 21-50 | Grouped results with __typename return the collection type name for each group. |
| `TestQuerySimpleWithGroupByWithChildTypeName` | 52-87 | GROUP child with __typename returns the type name for each sub-document. |

### `with_limit_offset_test.go`

Limit and offset pagination: zero limit, small limits, offsets beyond collection size, and combined limit+offset.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithLimit0` | 21-59 | Limit of 0 returns an empty result set. |
| `TestQuerySimpleWithLimit1` | 61-97 | Limit of 1 returns only the first document. |
| `TestQuerySimpleWithLimit2` | 99-152 | Limit of 2 returns only the first two documents. |
| `TestQuerySimpleWithLimitBiggerThanTotalDocuments` | 154-184 | Limit larger than the document count returns all documents. |
| `TestQuerySimpleWithOffset0` | 186-227 | Offset of 0 returns all documents from the beginning. |
| `TestQuerySimpleWithOffset1` | 229-265 | Offset of 1 skips the first document and returns the rest. |
| `TestQuerySimpleWithOffset2` | 267-330 | Offset of 2 skips the first two documents and returns the rest. |
| `TestQuerySimpleWithOffsetBiggerThanTotalDocuments` | 332-357 | Offset larger than the document count returns an empty result set. |
| `TestQuerySimpleWithLimit0AndOffset0` | 359-400 | Limit 0 and offset 0 together return an empty result set. |
| `TestQuerySimpleWithLimit1AndOffset1` | 402-438 | Limit 1 and offset 1 return the second document. |
| `TestQuerySimpleWithLimit2AndOffset2` | 440-493 | Limit 2 and offset 2 return the third and fourth documents. |

### `with_max_filter_test.go`

Top-level MAX aggregate combined with a field filter.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimple_WithMaxFilter_Succeeds` | 21-55 | Top-level MAX with a filter computes the maximum of only matching documents. |

### `with_max_order_test.go`

Ordering documents by an aliased MAX aggregate result.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithMaxWithOrder_Succeeds` | 21-80 | Order documents by an aliased MAX value in ASC and DESC directions. |

### `with_max_test.go`

Top-level MAX aggregate on integer fields, empty collections, undefined objects, maximum int values, and aliases.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimple_WithMaxOnUndefinedObject_ReturnsError` | 25-39 | MAX with no collection argument returns an error. |
| `TestQuerySimple_WithMaxOnUndefinedField_ReturnsError` | 41-55 | MAX on a collection without specifying a field returns an error. |
| `TestQuerySimple_WithMaxOnEmptyCollection_Succeeds` | 57-73 | MAX on an empty collection returns nil. |
| `TestQuerySimple_WithMax_Succeeds` | 75-103 | Top-level MAX of an integer field returns the correct maximum value. |
| `TestQuerySimple_WithMaxAndMaxValueInt_Succeeds` | 105-138 | Top-level MAX returns the maximum integer value from multiple documents. |
| `TestQuerySimple_WithAliasedMaxOnEmptyCollection_Succeeds` | 140-156 | Aliased MAX on an empty collection returns nil under the alias name. |

### `with_min_filter_test.go`

Top-level MIN aggregate combined with a field filter.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimple_WithMinFilter_Succeeds` | 21-55 | Top-level MIN with a filter computes the minimum of only matching documents. |

### `with_min_order_test.go`

Ordering documents by an aliased MIN aggregate result.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithMinWithOrder_Succeeds` | 21-80 | Order documents by an aliased MIN value in ASC and DESC directions. |

### `with_min_test.go`

Top-level MIN aggregate on integer fields, empty collections, undefined objects, max-value integers, and aliases.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimple_WithMinOnUndefinedObject_ReturnsError` | 25-39 | MIN with no collection argument returns an error. |
| `TestQuerySimple_WithMinOnUndefinedField_ReturnsError` | 41-55 | MIN on a collection without specifying a field returns an error. |
| `TestQuerySimple_WithMinOnEmptyCollection_Succeeds` | 57-73 | MIN on an empty collection returns nil. |
| `TestQuerySimple_WithMin_Succeeds` | 75-103 | Top-level MIN of an integer field returns the correct minimum value. |
| `TestQuerySimple_WithMinAndMaxValueInt_Succeeds` | 105-138 | Top-level MIN returns the minimum integer value from multiple documents. |
| `TestQuerySimple_WithAliasedMinOnEmptyCollection_Succeeds` | 140-156 | Aliased MIN on an empty collection returns nil under the alias name. |

### `with_multiple_types_test.go`

Querying a target collection when multiple other collection types are defined before or after it.

| Test Function | Line | Description |
|---|---|---|
| `TestSimple_WithSevenDummyTypesBefore` | 29-88 | Query with seven dummy collection types defined before the target type returns correct results. |
| `TestSimple_WithEightDummyTypesBefore` | 90-152 | Query with eight dummy collection types before the target returns correct results. |
| `TestSimple_WithEightDummyTypesBeforeInSplitDeclaration` | 154-219 | Query with eight dummy types before in split schema declarations returns correct results. |
| `TestSimple_WithEightDummyTypesAfter` | 221-283 | Query with eight dummy collection types defined after the target returns correct results. |
| `TestSimple_WithSevenDummyTypesBeforeAndOneAfter` | 285-348 | Query with seven dummy types before and one after the target returns correct results. |

### `with_null_input_test.go`

Null inputs for filter, order, limit, offset, docID, CID, groupBy, showDeleted, and logical operators (_or, _and, _not).

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimple_WithNullFilter_Succeeds` | 21-48 | Passing null as the filter argument returns all documents. |
| `TestQuerySimple_WithNullFilterFields_Succeeds` | 50-77 | Passing null for individual filter fields behaves as if no filter is applied. |
| `TestQuerySimple_WithNullOrder_Succeeds` | 79-106 | Passing null as the order argument returns all documents in default order. |
| `TestQuerySimple_WithNullOrderFields_Succeeds` | 108-135 | Passing null for individual order fields returns documents in default order. |
| `TestQuerySimple_WithNullLimit_Succeed` | 137-164 | Passing null as the limit argument returns all documents without a limit. |
| `TestQuerySimple_WithNullOffset_Succeeds` | 166-193 | Passing null as the offset argument returns all documents from the beginning. |
| `TestQuerySimple_WithNullDocID_Succeeds` | 195-222 | Passing null as the docID argument returns all documents. |
| `TestQuerySimple_WithNullDocIDs_Succeeds` | 224-251 | Passing null for the docIDs argument returns all documents. |
| `TestQuerySimple_WithNullCID_Succeeds` | 253-280 | Passing null as the cid argument returns all documents at the current state. |
| `TestQuerySimple_WithNullGroupBy_Succeeds` | 282-309 | Passing null as the groupBy argument returns all documents ungrouped. |
| `TestQuerySimple_WithNullShowDeleted_Succeeds` | 311-338 | Passing null as showDeleted returns all non-deleted documents. |
| `TestQuerySimple_WithFilterWithNullOr_Succeeds` | 340-367 | Filter with a null _or condition behaves as if no _or filter is applied. |
| `TestQuerySimple_WithFilterWithNullOrElement_ReturnsError` | 369-385 | Filter with a null element inside _or returns a validation error. |
| `TestQuerySimple_WithFilterWithNullOrField_ReturnsError` | 387-414 | Filter with a null field value inside _or returns a validation error. |
| `TestQuerySimple_WithFilterWithNullAnd_Succeeds` | 416-443 | Filter with a null _and condition behaves as if no _and filter is applied. |
| `TestQuerySimple_WithFilterWithNullAndElement_ReturnsError` | 445-461 | Filter with a null element inside _and returns a validation error. |
| `TestQuerySimple_WithFilterWithNullAndField_ReturnsError` | 463-490 | Filter with a null field value inside _and returns a validation error. |
| `TestQuerySimple_WithFilterWithNullNot_Succeeds` | 492-519 | Filter with a null _not condition behaves as if no _not filter is applied. |
| `TestQuerySimple_WithFilterWithNullNotField_Succeeds` | 521-548 | Filter with a null field value inside _not behaves as if no _not filter is applied. |

### `with_operation_alias_test.go`

Using a GraphQL operation alias to rename the query response field.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithOperationAlias` | 21-53 | Multiple aliased query operations in one request return results under each alias. |

### `with_operation_name_test.go`

Named GraphQL operations: selecting one of multiple operations and returning an error when no name is given.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleMultipleOperationsWithOperationName` | 23-91 | Multiple named operations with operationName selector returns only the specified operation. |
| `TestQuerySimpleMultipleOperationsWithNoOperationName_ReturnsError` | 93-126 | Multiple operations with no operationName selector returns an error. |

### `with_order_filter_test.go`

Combining a numeric greater-than filter with descending numeric order.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithNumericGreaterThanFilterAndNumericOrderDescending` | 21-73 | Combine a greater-than numeric filter with descending order returns filtered sorted results. |

### `with_order_test.go`

Ordering by scalar field types (Int, Float32, Float64, Blob, DateTime), compound sort directions, alias ordering, and error cases.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithEmptyOrder` | 21-72 | Passing an empty order object returns all documents in default order. |
| `TestQuerySimpleWithNumericOrderAscending` | 74-134 | Order documents by a numeric field ascending returns them from lowest to highest. |
| `TestQuerySimpleWithFloat32OrderAscending` | 136-202 | Order documents by a float32 field ascending returns them from lowest to highest. |
| `TestQuerySimpleWithFloat64OrderAscending` | 204-270 | Order documents by a float64 field ascending returns them from lowest to highest. |
| `TestQuerySimpleWithBlobOrderAscending` | 272-338 | Order documents by a blob field ascending returns them in lexicographic order. |
| `TestQuerySimpleWithDateTimeOrderAscending` | 340-404 | Order documents by a DateTime field ascending returns them chronologically. |
| `TestQuerySimpleWithNumericOrderDescending` | 406-466 | Order documents by a numeric field descending returns them from highest to lowest. |
| `TestQuerySimpleWithFloat32OrderDescending` | 468-534 | Order documents by a float32 field descending returns them from highest to lowest. |
| `TestQuerySimpleWitFloat64OrderDescending` | 536-602 | Order documents by a float64 field descending returns them from highest to lowest. |
| `TestQuerySimpleWithBlobOrderDescending` | 604-670 | Order documents by a blob field descending returns them in reverse lexicographic order. |
| `TestQuerySimpleWithDateTimeOrderDescending` | 672-736 | Order documents by a DateTime field descending returns them reverse-chronologically. |
| `TestQuerySimpleWithNumericOrderDescendingAndBooleanOrderAscending` | 738-807 | Order by numeric descending then boolean ascending returns the correct compound sort. |
| `TestQuerySimple_WithMultipleOrderFieldsASCAndASC_ShouldOrderCorrectly` | 809-869 | Two ASC order fields produce the expected compound sort. |
| `TestQuerySimple_WithMultipleOrderFieldsACSAndDESC_ShouldOrderCorrectly` | 871-931 | First field ASC, second field DESC produces the expected compound sort. |
| `TestQuerySimple_WithMultipleOrderFieldsDESCAndASC_ShouldOrderCorrectly` | 933-993 | First field DESC, second field ASC produces the expected compound sort. |
| `TestQuerySimple_WithMultipleOrderFieldsDECSAndDESC_ShouldOrderCorrectly` | 995-1055 | Two DESC order fields produce the expected compound sort. |
| `TestQuerySimple_WithInvalidOrderEnum_ReturnsError` | 1057-1075 | Passing an invalid order direction enum value returns a schema validation error. |
| `TestQuerySimple_WithMultipleOrderFields_ReturnsError` | 1077-1094 | Specifying multiple order fields in a single order object returns an error. |
| `TestQuerySimple_WithMultipleOrderFieldsNestedWithinMultpleFields_ReturnsError` | 1096-1113 | Nesting multiple order fields within a compound order object returns an error. |
| `TestQuerySimple_WithAliasOrder_ShouldOrderResults` | 1115-1175 | Ordering by an aliased aggregate field sorts documents by the aggregated value. |
| `TestQuerySimple_WithAliasOrderOnNonAliasedField_ShouldOrderResults` | 1177-1237 | Ordering by an alias that refers to a non-aliased field sorts by the underlying field value. |
| `TestQuerySimple_WithAliasOrderOnNonExistantField_ShouldError` | 1239-1280 | Ordering by an alias that references a non-existent field returns a schema error. |
| `TestQuerySimple_WithInvalidAliasOrder_ShouldError` | 1282-1323 | Ordering by an alias with an invalid direction value returns a schema validation error. |
| `TestQuerySimple_WithEmptyAliasOrder_ShouldDoNothing` | 1325-1386 | Ordering by an alias with an empty order object returns documents in default order. |
| `TestQuerySimple_WithNullAliasOrder_ShouldDoNothing` | 1388-1449 | Ordering by an alias with a null direction returns documents in default order. |
| `TestQuerySimple_WithIntAliasOrder_ShouldError` | 1451-1492 | Ordering by an alias with an integer direction value returns a schema validation error. |
| `TestQuerySimple_WithCompoundAliasOrder_ShouldOrderResults` | 1494-1563 | Ordering by a compound alias that aggregates multiple fields sorts by the combined value. |

### `with_restart_test.go`

Querying after a node restart returns the persisted document.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithRestart` | 21-61 | After a node restart the query returns the same documents that were present before. |

### `with_similarity_test.go`

Similarity queries on vector fields: error cases (non-vector field, wrong types), empty collections, int/float32/float64 vectors, JSON creation, filtering, ordering with limit, and multiple similarity fields.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimple_WithSimilarityOnQuery_ShouldError` | 21-41 | Using similarity on a query without a vector field returns an error. |
| `TestQuerySimple_WithSimilarityOnUndefinedField_ShouldError` | 43-64 | Using similarity on an undefined field returns an error. |
| `TestQuerySimple_WithSimilarityAndWrongVectorValueType_ShouldError` | 66-90 | Providing a wrong value type for the similarity vector returns an error. |
| `TestQuerySimple_WithSimilarityAndWrongFieldType_ShouldError` | 92-115 | Using similarity on a non-vector field returns an error. |
| `TestQuerySimple_WithSimilarityOnEmptyCollection_ShouldSucceed` | 117-141 | Similarity query on an empty collection returns an empty result. |
| `TestQuerySimple_WithIntSimilarity_ShouldSucceed` | 143-179 | Similarity query on an integer vector field returns documents ordered by similarity. |
| `TestQuerySimple_WithIntSimilarityDifferentVectorLength_ShouldError` | 181-210 | Similarity query with a vector of different length than the stored field returns an error. |
| `TestQuerySimple_WithFloat32Similarity_ShouldSucceed` | 212-248 | Similarity query on a float32 vector field returns documents ordered by similarity. |
| `TestQuerySimple_WithFloat64Similarity_ShouldSucceed` | 250-286 | Similarity query on a float64 vector field returns documents ordered by similarity. |
| `TestQuerySimple_WithJSONDocCreationSimilarity_ShouldSucceed` | 288-324 | Similarity query on a document created via JSON returns documents ordered by similarity. |
| `TestQuerySimple_WithSimilarityAndFilteringOnSimilarityResult_ShouldSucceed` | 326-378 | Similarity query with a filter on the similarity score returns only qualifying documents. |
| `TestQuerySimple_WithSimilarityAndOrderingWithLimitOnSimilarityResult_ShouldSucceed` | 380-432 | Similarity query with ordering and limit returns top-N most similar documents. |
| `TestQuerySimple_WithTwoSimilarityAndFilteringOnSecond_ShouldSucceed` | 434-484 | Two similarity fields where filtering is applied only to the second returns correct results. |
| `TestQuerySimple_WithTwoSimilarityAndFilteringOnBoth_ShouldSucceed` | 489-533 | Two similarity fields with filters on both return documents satisfying both similarity conditions. |

### `with_sum_filter_test.go`

Top-level SUM aggregate combined with a field filter.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithSumWithFilter` | 21-55 | Top-level SUM with a filter sums only matching documents. |

### `with_sum_order_test.go`

Ordering documents by an aliased SUM aggregate result.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithSumWithOrder_Succeeds` | 21-80 | Order documents by an aliased SUM value in ASC and DESC directions. |

### `with_sum_test.go`

Top-level SUM aggregate on integer fields, empty collections, and undefined objects.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithSumOnUndefinedObject` | 21-35 | SUM with no collection argument returns an error. |
| `TestQuerySimpleWithSumOnUndefinedField` | 37-51 | SUM on a collection without specifying a field returns an error. |
| `TestQuerySimpleWithSumOnEmptyCollection` | 53-69 | SUM on an empty collection returns zero. |
| `TestQuerySimpleWithSum` | 71-99 | Top-level SUM of an integer field returns the correct total. |

### `with_typename_test.go`

Querying __typename and aliased __typename on collection documents.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithTypeName` | 21-50 | Query requesting __typename returns the collection type name for each document. |
| `TestQuerySimpleWithAliasedTypeName` | 52-83 | Query with an aliased __typename returns the type name under the alias. |

### `with_variables_test.go`

GraphQL variables: non-null variables, default values, overrides, null enforcement, order variables, and aggregate count variables.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithNonNullVariable` | 23-64 | Query with a non-null variable correctly passes the value to the query. |
| `TestQuerySimpleWithVariableDefaultValue` | 66-103 | Query with a variable that has a default value uses the default when none is supplied. |
| `TestQuerySimpleWithNonNullVariable_ReturnsErrorWhenNull` | 105-133 | Query with a non-null variable set to null returns a validation error. |
| `TestQuerySimpleWithVariableDefaultValueOverride` | 135-172 | Query with a variable default that is overridden uses the provided value instead. |
| `TestQuerySimpleWithOrderVariable` | 174-217 | Query with an order variable correctly applies the dynamic order direction. |
| `TestQuerySimpleWithAggregateCountVariable` | 219-256 | Query with a variable used inside an aggregate correctly counts matching documents. |

### `with_version_cid_doc_ids_test.go`

Querying _version with CID and docID combinations: correct match, mixed correct/incorrect, and all-incorrect.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithVersionAndCidAndCorrectDocID` | 21-68 | Query with CID and a matching docID returns the version at that CID for that document. |
| `TestQuerySimpleWithVersionAndCidAndCorrectAndIncorrectDocID` | 70-120 | Query with CID and mixed docIDs returns the version only for the matching document. |
| `TestQuerySimpleWithVersionAndCidAndIncorrectDocID` | 122-163 | Query with CID and a non-matching docID returns an empty result. |

### `with_version_cid_test.go`

Querying _version with a CID filter to retrieve a specific schema version.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithVersionAndCid` | 21-68 | Query _version with a CID returns the commit history anchored at that CID. |

### `with_version_order_test.go`

Ordering results by a field within the embedded _version commit metadata.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithVersionAndOrder` | 21-76 | Query _version with order applied returns versions sorted by the specified field. |

### `with_version_test.go`

Querying the embedded _version (latest commit) field: basic, after add/update, collectionVersionId, docID, multiple aliases, interleaved aliases, filtered aliases, all commit fields, and all commit fields after update.

| Test Function | Line | Description |
|---|---|---|
| `TestQuerySimpleWithNestedLatestCommit` | 21-77 | Query _version inline returns the latest commit CID and height for each document. |
| `TestQuery_AddDocWithNestedLatestCommit` | 79-143 | Query _version after adding a document returns the initial commit metadata. |
| `TestQuery_UpdateDocWithNestedLatestCommit` | 145-231 | Query _version after updating a document returns updated commit metadata. |
| `TestQuerySimpleWithEmbeddedLatestCommitWithCollectionVersionID` | 233-270 | _version with collectionVersionId returns the schema version CID for each commit. |
| `TestQuerySimpleWithEmbeddedLatestCommitWithDocID` | 272-312 | _version with docID returns the document ID for each commit. |
| `TestQuerySimpleWithMultipleAliasedEmbeddedLatestCommit` | 314-380 | Multiple aliased _version selections return correct commit data under each alias. |
| `TestQuerySimpleWithMultipleAliasedInterleavedNestedLatestCommit` | 382-489 | Interleaved aliased _version selections in complex queries return correct commit data. |
| `TestQuery_WithMultipleAliasedFilteredEmbeddedLatestCommit` | 491-554 | Multiple aliased _version selections with filters return correct filtered commit data. |
| `TestQuery_WithAllCommitFields_NoError` | 556-628 | Query _version requesting all commit fields returns complete commit metadata. |
| `TestQuery_WithAllCommitFieldsWithUpdate_NoError` | 630-734 | Query _version with all commit fields after an update returns full updated commit metadata. |

## Subdirectories

| Directory | Summary |
|---|---|
| [`with_filter/`](with_filter/INDEX.md) | Tests for all filter comparison operators (`_eq`, `_neq`, `_gt`, `_geq`, `_lt`, `_leq`, `_in`, `_nin`, `_like`, `_ilike`, `_nlike`, `_nilike`) across every supported scalar field type, logical combinators (`_and`, `_or`, `_not`), alias-based filtering, null-value edge cases, and inline array field variants. |
