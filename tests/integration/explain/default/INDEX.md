# Default Explain Tests

This folder contains integration tests for the `default` explain type. Default explain returns the
full query plan tree with node-specific attributes — including filters, prefixes, collection IDs,
field names, and aggregation sources — in addition to the structural hierarchy. Each test asserts
the expected attribute values on specific plan nodes produced when a `@explain` (or `@explain(type:
default)`) directive is applied to a query or mutation.

---

### add_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainMutationRequestWithAdd | 38 | Default explain of add mutation shows addNode attributes including input fields. |
| TestDefaultExplainMutationRequestDoesNotAddDocGivenDuplicate | 76 | Default explain of add mutation with partial input shows addNode attributes without optional fields. |

---

### basic_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainOnWrongFieldDirective_BadUsage | 22 | Default explain directive applied to a field instead of an operation returns an error. |
| TestDefaultExplainRequestWithFullBasicGraph | 46 | Default explain of basic query returns full graph with scanNode attributes and prefixes. |
| TestDefaultExplainWithAlias | 91 | Default explain of query with field aliases shows basic scanNode plan tree. |

---

### dagscan_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainCommitsDagScanQueryOp | 36 | Default explain of commits query with field filter shows dagScanNode prefixes and attributes. |
| TestDefaultExplainCommitsDagScanQueryOpWithoutField | 77 | Default explain of commits query without field filter shows dagScanNode with doc prefix. |

---

### delete_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainMutationRequestWithDeleteUsingFilter | 38 | Default explain of delete mutation with filter shows deleteNode attributes with filter and scanNode. |
| TestDefaultExplainMutationRequestWithDeleteUsingFilterToMatchEverything | 93 | Default explain of delete mutation with empty filter shows deleteNode with nil filter attribute. |
| TestDefaultExplainMutationRequestWithDeleteUsingId | 140 | Default explain of delete mutation with single docID shows deleteNode with docID attribute. |
| TestDefaultExplainMutationRequestWithDeleteUsingIds | 189 | Default explain of delete mutation with multiple docIDs shows deleteNode with all docID attributes. |
| TestDefaultExplainMutationRequestWithDeleteUsingNoIds | 243 | Default explain of delete mutation with empty docID list shows deleteNode with empty docID attribute. |
| TestDefaultExplainMutationRequestWithDeleteUsingFilterAndIds | 290 | Default explain of delete mutation with both filter and docIDs shows deleteNode with combined attributes. |

---

### delete_with_error_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainMutationRequestWithDeleteHavingNoSubSelection | 22 | Default explain of delete mutation without sub-selection returns a parse error. |

---

### group_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainRequestWithGroupByOnParent | 38 | Default explain of groupBy on a single parent field shows groupNode with groupByFields attribute. |
| TestDefaultExplainRequestWithGroupByTwoFieldsOnParent | 77 | Default explain of groupBy on two parent fields shows groupNode with multiple groupByFields. |

---

### group_with_average_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainRequestWithGroupByWithAverageOnAnInnerField | 44 | Default explain of groupBy with AVG on inner field shows averageNode attributes over groupNode. |
| TestDefaultExplainRequestWithAverageInsideTheInnerGroupOnAField | 130 | Default explain of nested groupBy with AVG inside inner GROUP shows averageNode over groupNode. |
| TestDefaultExplainRequestWithAverageInsideTheInnerGroupOnAFieldAndNestedGroupBy | 208 | Default explain of nested groupBy with AVG and nested groupBy shows averageNode plan. |
| TestDefaultExplainRequestWithAverageInsideTheInnerGroupAndNestedGroupByWithAverage | 289 | Default explain of deeply nested groupBy with multiple AVG nodes shows averageNode plan. |

---

### group_with_doc_id_child_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainRequestWithDocIDsOnInnerGroupSelection | 22 | Default explain of groupBy with docID filter on inner GROUP shows groupNode plan tree. |

---

### group_with_doc_id_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainRequestWithDocIDOnParentGroupBy | 22 | Default explain of groupBy with single docID on parent shows groupNode plan tree. |
| TestDefaultExplainRequestWithDocIDsAndFilterOnParentGroupBy | 76 | Default explain of groupBy with docIDs and filter on parent shows groupNode plan tree. |

---

### group_with_filter_child_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainRequestWithFilterOnInnerGroupSelection | 22 | Default explain of groupBy with filter on inner GROUP shows groupNode plan tree. |
| TestDefaultExplainRequestWithFilterOnParentGroupByAndInnerGroupSelection | 84 | Default explain of groupBy with filters on both parent and inner GROUP shows groupNode plan. |

---

### group_with_filter_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainRequestWithFilterOnGroupByParent | 22 | Default explain of groupBy with filter on parent shows groupNode plan tree. |

---

### group_with_limit_child_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainRequestWithLimitAndOffsetOnInnerGroupSelection | 22 | Default explain of groupBy with limit and offset on inner GROUP shows groupNode plan tree. |
| TestDefaultExplainRequestWithLimitAndOffsetOnMultipleInnerGroupSelections | 71 | Default explain of groupBy with limit/offset on multiple inner GROUPs shows groupNode plan. |

---

### group_with_limit_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainRequestWithLimitAndOffsetOnParentGroupBy | 40 | Default explain of groupBy with limit and offset on parent shows limitNode wrapping groupNode. |
| TestDefaultExplainRequestWithLimitOnParentGroupByAndInnerGroupSelection | 91 | Default explain of groupBy with limit on both parent and inner GROUP shows limitNode plan. |

---

### group_with_order_child_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainRequestWithDescendingOrderOnInnerGroupSelection | 22 | Default explain of groupBy with descending order on inner GROUP shows groupNode plan tree. |
| TestDefaultExplainRequestWithAscendingOrderOnInnerGroupSelection | 73 | Default explain of groupBy with ascending order on inner GROUP shows groupNode plan tree. |
| TestDefaultExplainRequestWithOrderOnNestedParentGroupByAndOnNestedParentsInnerGroupSelection | 124 | Default explain of nested groupBy with order on parent GROUP and inner GROUP shows groupNode plan. |

---

### group_with_order_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainRequestWithDescendingOrderOnParentGroupBy | 40 | Default explain of groupBy with descending order on parent shows orderNode wrapping groupNode. |
| TestDefaultExplainRequestWithAscendingOrderOnParentGroupBy | 94 | Default explain of groupBy with ascending order on parent shows orderNode wrapping groupNode. |
| TestDefaultExplainRequestWithOrderOnParentGroupByAndOnInnerGroupSelection | 148 | Default explain of groupBy with order on both parent and inner GROUP shows orderNode plan. |

---

### invalid_type_arg_test.go

| Test Function | Line | Description |
|---|---|---|
| TestInvalidExplainRequestTypeReturnsError | 22 | Explain request with an invalid type argument returns a validation error. |

---

### top_with_average_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainTopLevelAverageRequest | 49 | Default explain of top-level AVG shows topLevelNode with averageNode attributes and scanNode. |
| TestDefaultExplainTopLevelAverageRequestWithFilter | 131 | Default explain of top-level AVG with filter shows topLevelNode with averageNode and filter attributes. |

---

### top_with_count_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainTopLevelCountRequest | 43 | Default explain of top-level COUNT shows topLevelNode with countNode attributes and scanNode. |
| TestDefaultExplainTopLevelCountRequestWithFilter | 91 | Default explain of top-level COUNT with filter shows topLevelNode with countNode and filter attributes. |

---

### top_with_max_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplain_WithTopLevelMaxRequest_Succeeds | 43 | Default explain of top-level MAX shows topLevelNode with maxNode attributes and scanNode. |
| TestDefaultExplain_WithTopLevelMaxRequestWithFilter_Succeeds | 96 | Default explain of top-level MAX with filter shows topLevelNode with maxNode and filter attributes. |

---

### top_with_min_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplain_WithTopLevelMinRequest_Succeeds | 43 | Default explain of top-level MIN shows topLevelNode with minNode attributes and scanNode. |
| TestDefaultExplain_WithTopLevelMinRequestWithFilter_Succeeds | 96 | Default explain of top-level MIN with filter shows topLevelNode with minNode and filter attributes. |

---

### top_with_sum_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainTopLevelSumRequest | 43 | Default explain of top-level SUM shows topLevelNode with sumNode attributes and scanNode. |
| TestDefaultExplainTopLevelSumRequestWithFilter | 96 | Default explain of top-level SUM with filter shows topLevelNode with sumNode and filter attributes. |

---

### type_join_many_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainRequestWithAOneToManyJoin | 24 | Default explain of one-to-many join shows typeIndexJoin with typeJoinMany attributes. |

---

### type_join_one_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainRequestWithAOneToOneJoin | 24 | Default explain of one-to-one join shows typeIndexJoin with typeJoinOne attributes. |
| TestDefaultExplainRequestWithTwoLevelDeepNestedJoins | 112 | Default explain of two-level deep nested joins shows nested typeJoinOne attributes. |

---

### type_join_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainRequestWith2SingleJoinsAnd1ManyJoin | 37 | Default explain of two one-to-one joins and one many join shows parallelNode with join attributes. |

---

### type_join_with_filter_doc_id_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainRequestWithRelatedAndRegularFilterAndDocIDs | 22 | Default explain of query with related type filter and docIDs shows typeJoinMany attributes. |
| TestDefaultExplainRequestWithManyRelatedFiltersAndDocID | 103 | Default explain of query with multiple related type filters and docID shows parallelNode attributes. |

---

### type_join_with_filter_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainRequestWithRelatedAndRegularFilter | 22 | Default explain of query with related type filter shows typeJoinMany plan and filter attributes. |
| TestDefaultExplainRequestWithManyRelatedFilters | 95 | Default explain of query with multiple related type filters shows parallelNode with typeJoinMany. |

---

### update_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainMutationRequestWithUpdateUsingBooleanFilter | 42 | Default explain of update mutation with boolean filter shows updateNode attributes with filter. |
| TestDefaultExplainMutationRequestWithUpdateUsingIds | 108 | Default explain of update mutation with multiple docIDs shows updateNode attributes with docIDs. |
| TestDefaultExplainMutationRequestWithUpdateUsingId | 169 | Default explain of update mutation with single docID shows updateNode attributes with docID. |
| TestDefaultExplainMutationRequestWithUpdateUsingIdsAndFilter | 225 | Default explain of update mutation with docIDs and filter shows updateNode with combined attributes. |

---

### upsert_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainMutationRequest_WithUpsert_Succeeds | 38 | Default explain of upsert mutation shows upsertNode wrapping selectTopNode and scanNode. |

---

### with_average_join_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainRequestWithAverageOnJoinedField | 44 | Default explain of average on a joined field shows averageNode with countNode and sumNode. |
| TestDefaultExplainRequestWithAverageOnMultipleJoinedFieldsWithFilter | 148 | Default explain of average on multiple joined fields with filter shows parallelNode with two typeIndexJoins. |
| TestDefaultExplainRequestOneToManyWithAverageAndChildNeNilFilterSharesJoinField | 343 | Default explain of average with matching ne-nil child filter reuses a single typeIndexJoin. |

---

### with_average_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainRequestWithAverageOnArrayField | 42 | Default explain of average on an inline array field shows averageNode with countNode and sumNode. |

---

### with_count_join_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainRequestWithCountOnOneToManyJoinedField | 40 | Default explain of count on a one-to-many joined field shows countNode with typeIndexJoin attributes. |
| TestDefaultExplainRequestWithCountOnOneToManyJoinedFieldWithManySources | 114 | Default explain of count with multiple joined field sources shows countNode with parallelNode. |
| TestDefaultExplainRequestOneToManyWithCountWithFilterAndChildFilterSharesJoinField | 256 | Default explain of count with matching child filter reuses a single typeIndexJoin instead of parallelNode. |
| TestDefaultExplainRequestOneToManyWithCountAndChildFilterDoesNotShareJoinField | 299 | Default explain of count with non-matching child filter uses parallelNode with two typeIndexJoins. |

---

### with_count_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainRequestWithCountOnInlineArrayField | 38 | Default explain of count on an inline array field shows countNode with scanNode attributes. |

---

### with_filter_doc_id_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainRequestWithDocIDFilter | 22 | Default explain of query with a single docID filter shows scanNode with matching prefix. |
| TestDefaultExplainRequestWithDocIDsFilterUsingOneID | 70 | Default explain of query with docIDs filter using one ID shows scanNode with single prefix. |
| TestDefaultExplainRequestWithDocIDsFilterUsingMultipleButDuplicateIDs | 118 | Default explain of query with duplicate docIDs filter shows scanNode with deduplicated prefixes. |
| TestDefaultExplainRequestWithDocIDsFilterUsingMultipleUniqueIDs | 173 | Default explain of query with multiple unique docIDs shows scanNode with multiple distinct prefixes. |
| TestDefaultExplainRequestWithMatchingIDFilter | 228 | Default explain of query with ID equality filter shows scanNode with docID-keyed prefix. |

---

### with_filter_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainRequestWithStringEqualFilter | 22 | Default explain of query with string equality filter shows scanNode with filter attributes. |
| TestDefaultExplainRequestWithIntegerEqualFilter | 65 | Default explain of query with integer equality filter shows scanNode with numeric filter attribute. |
| TestDefaultExplainRequestWithGreaterThanFilter | 108 | Default explain of query with greater-than filter shows scanNode with _gt filter attribute. |
| TestDefaultExplainRequestWithLogicalCompoundAndFilter | 151 | Default explain of query with compound AND filter shows scanNode with _and filter attributes. |
| TestDefaultExplainRequestWithLogicalCompoundOrFilter | 203 | Default explain of query with compound OR filter shows scanNode with _or filter attributes. |
| TestDefaultExplainRequestWithMatchInsideList | 255 | Default explain of query with _in list filter shows scanNode with _in filter attribute. |
| TestDefaultExplainRequest_WithJSONEqualFilter_Succeeds | 302 | Default explain of query with JSON equality filter shows scanNode with JSON filter attribute. |

---

### with_index_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainWithIndexOnFilter | 22 | Default explain of query with an indexed field filter shows index-based scanNode plan. |
| TestDefaultExplainWithIndexOnOrder | 63 | Default explain of query with index on the order field shows index-based plan attributes. |
| TestDefaultExplainWithIndexOnSubqueryNestedRelationOrder | 105 | Default explain of nested relation query with index on order shows typeIndexJoin plan. |

---

### with_limit_count_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainRequestWithOnlyLimitOnRelatedChildWithCount | 22 | Default explain of count with limit on related child shows limitNode inside typeIndexJoin. |
| TestDefaultExplainRequestWithLimitArgsOnParentAndRelatedChildWithCount | 92 | Default explain of count with limit on both parent and related child shows limitNode attributes at each level. |

---

### with_limit_join_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainRequestWithOnlyLimitOnRelatedChild | 37 | Default explain of join query with limit on related child shows limitNode inside subType. |
| TestDefaultExplainRequestWithOnlyOffsetOnRelatedChild | 86 | Default explain of join query with only offset on related child shows limitNode with offset attribute. |
| TestDefaultExplainRequestWithBothLimitAndOffsetOnRelatedChild | 135 | Default explain of join query with limit and offset on related child shows limitNode with both attributes. |
| TestDefaultExplainRequestWithLimitOnRelatedChildAndBothLimitAndOffsetOnParent | 184 | Default explain of join query with limit on child and both limit and offset on parent shows limitNode at each level. |

---

### with_limit_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainRequestWithOnlyLimit | 38 | Default explain of query with only a limit shows limitNode with limit attribute and no offset. |
| TestDefaultExplainRequestWithOnlyOffset | 72 | Default explain of query with only an offset shows limitNode with offset attribute and no limit. |
| TestDefaultExplainRequestWithLimitAndOffset | 106 | Default explain of query with limit and offset shows limitNode with both limit and offset attributes. |

---

### with_max_join_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainRequest_WithMaxOnOneToManyJoinedField_Succeeds | 40 | Default explain of max on a one-to-many joined field shows maxNode with typeIndexJoin attributes. |
| TestDefaultExplainRequest_WithMaxOnOneToManyJoinedFieldWithFilter_Succeeds | 118 | Default explain of max on a joined field with filter shows maxNode with filter in typeIndexJoin. |
| TestDefaultExplainRequest_WithMaxOnOneToManyJoinedFieldWithManySources_Succeeds | 210 | Default explain of max with multiple joined field sources shows parallelNode with two typeIndexJoins. |

---

### with_max_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainRequest_WithMaxOnInlineArrayField_ChildFieldWillBeEmpty | 38 | Default explain of max on an inline array field shows maxNode with empty child field attribute. |

---

### with_min_join_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainRequest_WithMinOnOneToManyJoinedField_Succeeds | 40 | Default explain of min on a one-to-many joined field shows minNode with typeIndexJoin attributes. |
| TestDefaultExplainRequest_WithMinOnOneToManyJoinedFieldWithFilter_Succeeds | 118 | Default explain of min on a joined field with filter shows minNode with filter in typeIndexJoin. |
| TestDefaultExplainRequest_WithMinOnOneToManyJoinedFieldWithManySources_Succeeds | 210 | Default explain of min with multiple joined field sources shows parallelNode with two typeIndexJoins. |

---

### with_min_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainRequest_WithMinOnInlineArrayField_ChildFieldWillBeEmpty | 38 | Default explain of min on an inline array field shows minNode with empty child field attribute. |

---

### with_order_join_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainRequestWithOrderFieldOnRelatedChild | 37 | Default explain of join query with order on related child shows orderNode inside subType. |
| TestDefaultExplainRequestWithOrderFieldOnParentAndRelatedChild | 92 | Default explain of join query with order on both parent and related child shows orderNode at each level. |
| TestDefaultExplainRequestWhereParentIsOrderedByItsRelatedChild | 165 | Default explain of query where parent is sorted by a related child field shows typeIndexJoin with orderNode. |

---

### with_order_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainRequestWithAscendingOrderOnParent | 38 | Default explain of query with ascending order shows orderNode with ASC direction attribute. |
| TestDefaultExplainRequestWithMultiOrderFieldsOnParent | 79 | Default explain of query with multiple order fields shows orderNode with multiple direction attributes. |

---

### with_sum_join_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainRequestWithSumOnOneToManyJoinedField | 40 | Default explain of sum on a one-to-many joined field shows sumNode with typeIndexJoin attributes. |
| TestDefaultExplainRequestWithSumOnOneToManyJoinedFieldWithFilter | 118 | Default explain of sum on a joined field with filter shows sumNode with filter in typeIndexJoin. |
| TestDefaultExplainRequestWithSumOnOneToManyJoinedFieldWithManySources | 210 | Default explain of sum with multiple joined field sources shows parallelNode with two typeIndexJoins. |

---

### with_sum_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDefaultExplainRequestWithSumOnInlineArrayField_ChildFieldWillBeEmpty | 38 | Default explain of sum on an inline array field shows sumNode with empty child field attribute. |
