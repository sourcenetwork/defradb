# Debug Explain Tests

This folder contains integration tests for the `debug` explain type. Debug explain returns only the
structural shape of the query plan tree — node names and hierarchy — without any node-specific
attributes such as filters, prefixes, or field names. Each test asserts the expected tree structure
produced when a `@explain(type: debug)` directive is applied to a query or mutation.

---

### add_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainMutationRequestWithAdd | 38 | Debug explain of add mutation shows addNode wrapping selectTopNode and scanNode. |
| TestDebugExplainMutationRequestDoesNotAddDocGivenDuplicate | 62 | Debug explain of add mutation with duplicate input shows addNode plan structure. |

---

### basic_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainRequest | 22 | Debug explain shows basic plan tree with selectTopNode and scanNode. |

---

### dagscan_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainCommitsDagScanQueryOp | 36 | Debug explain of commits query with field filter shows dagScanNode plan tree. |
| TestDebugExplainCommitsDagScanQueryOpWithoutField | 61 | Debug explain of commits query without field filter shows dagScanNode plan tree. |

---

### delete_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainMutationRequestWithDeleteUsingFilter | 38 | Debug explain of delete mutation with name filter shows deleteNode plan tree. |
| TestDebugExplainMutationRequestWithDeleteUsingFilterToMatchEverything | 61 | Debug explain of delete mutation with empty filter shows deleteNode matching all docs. |
| TestDebugExplainMutationRequestWithDeleteUsingId | 84 | Debug explain of delete mutation with single docID shows deleteNode plan tree. |
| TestDebugExplainMutationRequestWithDeleteUsingIds | 107 | Debug explain of delete mutation with multiple docIDs shows deleteNode plan tree. |
| TestDebugExplainMutationRequestWithDeleteUsingNoIds | 133 | Debug explain of delete mutation with empty docID list shows deleteNode plan tree. |
| TestDebugExplainMutationRequestWithDeleteUsingFilterAndIds | 156 | Debug explain of delete mutation with both filter and docIDs shows deleteNode plan tree. |

---

### delete_with_error_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainMutationRequestWithDeleteHavingNoSubSelection | 22 | Debug explain of delete mutation without sub-selection returns a parse error. |

---

### group_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainRequestWithGroupByOnParent | 40 | Debug explain of groupBy on a single parent field shows groupNode plan tree. |
| TestDebugExplainRequestWithGroupByTwoFieldsOnParent | 66 | Debug explain of groupBy on two parent fields shows groupNode plan tree. |

---

### group_with_average_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainRequestWithGroupByWithAverageOnAnInnerField | 46 | Debug explain of groupBy with AVG on inner field shows averageNode wrapping groupNode. |
| TestDebugExplainRequestWithAverageInsideTheInnerGroupOnAField | 70 | Debug explain of nested groupBy with AVG on inner GROUP field shows averageNode plan. |
| TestDebugExplainRequestWithAverageInsideTheInnerGroupOnAFieldAndNestedGroupBy | 98 | Debug explain of nested groupBy with AVG on inner field and nested groupBy shows averageNode. |
| TestDebugExplainRequestWithAverageInsideTheInnerGroupAndNestedGroupByWithAverage | 129 | Debug explain of deeply nested groupBy with multiple AVG nodes shows averageNode plan. |

---

### group_with_doc_id_child_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainRequestWithDocIDsOnInnerGroupSelection | 22 | Debug explain of groupBy with docID filter on inner GROUP shows groupNode plan tree. |

---

### group_with_doc_id_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainRequestWithDocIDOnParentGroupBy | 22 | Debug explain of groupBy with single docID on parent shows groupNode plan tree. |
| TestDebugExplainRequestWithDocIDsAndFilterOnParentGroupBy | 51 | Debug explain of groupBy with docIDs and filter on parent shows groupNode plan tree. |

---

### group_with_filter_child_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainRequestWithFilterOnInnerGroupSelection | 22 | Debug explain of groupBy with filter on inner GROUP selection shows groupNode plan tree. |
| TestDebugExplainRequestWithFilterOnParentGroupByAndInnerGroupSelection | 48 | Debug explain of groupBy with filters on both parent and inner GROUP shows groupNode plan. |

---

### group_with_filter_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainRequestWithFilterOnGroupByParent | 22 | Debug explain of groupBy with filter on parent shows groupNode plan tree. |

---

### group_with_limit_child_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainRequestWithLimitAndOffsetOnInnerGroupSelection | 22 | Debug explain of groupBy with limit and offset on inner GROUP shows groupNode plan tree. |
| TestDebugExplainRequestWithLimitAndOffsetOnMultipleInnerGroupSelections | 48 | Debug explain of groupBy with limit/offset on multiple inner GROUPs shows groupNode plan. |

---

### group_with_limit_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainRequestWithLimitAndOffsetOnParentGroupBy | 42 | Debug explain of groupBy with limit and offset on parent shows limitNode wrapping groupNode. |
| TestDebugExplainRequestWithLimitOnParentGroupByAndInnerGroupSelection | 72 | Debug explain of groupBy with limit on both parent and inner GROUP shows limitNode plan. |

---

### group_with_order_child_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainRequestWithDescendingOrderOnInnerGroupSelection | 22 | Debug explain of groupBy with descending order on inner GROUP shows groupNode plan tree. |
| TestDebugExplainRequestWithAscendingOrderOnInnerGroupSelection | 48 | Debug explain of groupBy with ascending order on inner GROUP shows groupNode plan tree. |
| TestDebugExplainRequestWithOrderOnNestedParentGroupByAndOnNestedParentsInnerGroupSelection | 74 | Debug explain of nested groupBy with order on parent GROUP and inner GROUP shows groupNode plan. |

---

### group_with_order_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainRequestWithDescendingOrderOnParentGroupBy | 42 | Debug explain of groupBy with descending order on parent shows orderNode wrapping groupNode. |
| TestDebugExplainRequestWithAscendingOrderOnParentGroupBy | 71 | Debug explain of groupBy with ascending order on parent shows orderNode wrapping groupNode. |
| TestDebugExplainRequestWithOrderOnParentGroupByAndOnInnerGroupSelection | 100 | Debug explain of groupBy with order on both parent and inner GROUP shows orderNode plan. |

---

### top_with_average_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainTopLevelAverageRequest | 49 | Debug explain of top-level AVG query shows topLevelNode with averageNode and scanNode. |
| TestDebugExplainTopLevelAverageRequestWithFilter | 74 | Debug explain of top-level AVG with filter shows topLevelNode with averageNode plan. |

---

### top_with_count_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainTopLevelCountRequest | 43 | Debug explain of top-level COUNT query shows topLevelNode with countNode and scanNode. |
| TestDebugExplainTopLevelCountRequestWithFilter | 64 | Debug explain of top-level COUNT with filter shows topLevelNode with countNode plan. |

---

### top_with_max_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplain_TopLevelMaxRequest_Succeeds | 43 | Debug explain of top-level MAX query shows topLevelNode with maxNode and scanNode. |
| TestDebugExplain_TopLevelMaxRequestWithFilter_Succeeds | 68 | Debug explain of top-level MAX with filter shows topLevelNode with maxNode plan. |

---

### top_with_min_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplain_TopLevelMinRequest_Succeeds | 43 | Debug explain of top-level MIN query shows topLevelNode with minNode and scanNode. |
| TestDebugExplain_TopLevelMinRequestWithFilter_Succeeds | 68 | Debug explain of top-level MIN with filter shows topLevelNode with minNode plan. |

---

### top_with_sum_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainTopLevelSumRequest | 43 | Debug explain of top-level SUM query shows topLevelNode with sumNode and scanNode. |
| TestDebugExplainTopLevelSumRequestWithFilter | 68 | Debug explain of top-level SUM with filter shows topLevelNode with sumNode plan. |

---

### type_join_many_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainRequestWithAOneToManyJoin | 22 | Debug explain of one-to-many join shows typeIndexJoin with typeJoinMany plan tree. |

---

### type_join_one_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainRequestWithAOneToOneJoin | 22 | Debug explain of one-to-one join shows typeIndexJoin with typeJoinOne plan tree. |
| TestDebugExplainRequestWithTwoLevelDeepNestedJoins | 61 | Debug explain of two-level deep nested joins shows nested typeJoinOne plan trees. |

---

### type_join_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainRequestWith2SingleJoinsAnd1ManyJoin | 50 | Debug explain of two one-to-one joins and one many join shows parallelNode with multiScanNode. |

---

### type_join_with_filter_doc_id_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainRequestWithRelatedAndRegularFilterAndDocIDs | 22 | Debug explain of query with related type filter and docIDs shows typeJoinMany plan. |
| TestDebugExplainRequestWithManyRelatedFiltersAndDocID | 69 | Debug explain of query with multiple related type filters and docID shows parallelNode plan. |

---

### type_join_with_filter_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainRequestWithRelatedAndRegularFilter | 22 | Debug explain of query with related type filter shows typeJoinMany plan tree. |
| TestDebugExplainRequestWithManyRelatedFilters | 65 | Debug explain of query with multiple related type filters shows parallelNode with typeJoinMany. |

---

### update_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainMutationRequestWithUpdateUsingBooleanFilter | 42 | Debug explain of update mutation with boolean filter shows updateNode plan tree. |
| TestDebugExplainMutationRequestWithUpdateUsingIds | 74 | Debug explain of update mutation with multiple docIDs shows updateNode plan tree. |
| TestDebugExplainMutationRequestWithUpdateUsingId | 105 | Debug explain of update mutation with single docID shows updateNode plan tree. |
| TestDebugExplainMutationRequestWithUpdateUsingIdsAndFilter | 133 | Debug explain of update mutation with docIDs and filter shows updateNode plan tree. |

---

### upsert_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainMutationRequest_WithUpsert_Succeeds | 38 | Debug explain of upsert mutation shows upsertNode wrapping selectTopNode and scanNode. |

---

### with_average_join_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainRequestWithAverageOnJoinedField | 44 | Debug explain of AVG on one-to-many joined field shows averageNode over typeJoinMany. |
| TestDebugExplainRequestWithAverageOnMultipleJoinedFieldsWithFilter | 68 | Debug explain of AVG on multiple joined fields with filter shows parallelNode with typeJoinMany. |

---

### with_average_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainRequestWithAverageOnArrayField | 42 | Debug explain of AVG on inline array field shows averageNode over countNode and scanNode. |

---

### with_count_join_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainRequestWithCountOnOneToManyJoinedField | 40 | Debug explain of COUNT on one-to-many joined field shows countNode over typeJoinMany. |
| TestDebugExplainRequestWithCountOnOneToManyJoinedFieldWithManySources | 64 | Debug explain of COUNT on multiple joined fields shows parallelNode with typeJoinMany. |

---

### with_count_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainRequestWithCountOnInlineArrayField | 38 | Debug explain of COUNT on inline array field shows countNode over selectNode and scanNode. |

---

### with_filter_doc_id_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainRequestWithDocIDFilter | 22 | Debug explain of query with single docID filter shows basic scanNode plan tree. |
| TestDebugExplainRequestWithDocIDsFilterUsingOneID | 46 | Debug explain of query with docIDs list containing one ID shows basic scanNode plan tree. |
| TestDebugExplainRequestWithDocIDsFilterUsingMultipleButDuplicateIDs | 70 | Debug explain of query with duplicate docIDs in filter shows basic scanNode plan tree. |
| TestDebugExplainRequestWithDocIDsFilterUsingMultipleUniqueIDs | 99 | Debug explain of query with multiple unique docIDs in filter shows basic scanNode plan tree. |
| TestDebugExplainRequestWithMatchingIDFilter | 128 | Debug explain of query with _docID equality filter in filter clause shows scanNode plan. |

---

### with_filter_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainRequestWithStringEqualFilter | 22 | Debug explain of query with string equality filter shows basic scanNode plan tree. |
| TestDebugExplainRequestWithIntegerEqualFilter | 46 | Debug explain of query with integer equality filter shows basic scanNode plan tree. |
| TestDebugExplainRequestWithGreaterThanFilter | 70 | Debug explain of query with greater-than filter shows basic scanNode plan tree. |
| TestDebugExplainRequestWithLogicalCompoundAndFilter | 94 | Debug explain of query with logical AND compound filter shows basic scanNode plan tree. |
| TestDebugExplainRequestWithLogicalCompoundOrFilter | 118 | Debug explain of query with logical OR compound filter shows basic scanNode plan tree. |
| TestDebugExplainRequestWithMatchInsideList | 142 | Debug explain of query with _in list filter shows basic scanNode plan tree. |

---

### with_index_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainWithIndexOnFilter | 22 | Debug explain of query with indexed field filter shows scanNode plan tree. |
| TestDebugExplainWithIndexOnOrder | 63 | Debug explain of query ordered by an indexed field shows scanNode plan tree. |
| TestDebugExplainWithIndexOnSubqueryNestedRelationOrder | 108 | Debug explain of subquery ordered by nested relation's indexed field omits orderNode. |

---

### with_limit_count_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainRequestWithOnlyLimitOnRelatedChildWithCount | 22 | Debug explain of COUNT with limit on related child shows parallelNode with limitNode in join. |
| TestDebugExplainRequestWithLimitArgsOnParentAndRelatedChildWithCount | 73 | Debug explain of COUNT with limits on both parent and related child shows limitNode wrapping countNode. |

---

### with_limit_join_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainRequestWithOnlyLimitOnRelatedChild | 54 | Debug explain of join with limit on related child shows limitNode in typeJoinMany subType. |
| TestDebugExplainRequestWithOnlyOffsetOnRelatedChild | 94 | Debug explain of join with only offset on related child shows limitNode in typeJoinMany subType. |
| TestDebugExplainRequestWithBothLimitAndOffsetOnRelatedChild | 134 | Debug explain of join with limit and offset on related child shows limitNode in typeJoinMany. |
| TestDebugExplainRequestWithLimitOnRelatedChildAndBothLimitAndOffsetOnParent | 174 | Debug explain of join with limit on child and limit/offset on parent shows limitNode on both. |

---

### with_limit_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainRequestWithOnlyLimit | 38 | Debug explain of query with only limit shows limitNode wrapping selectNode and scanNode. |
| TestDebugExplainRequestWithOnlyOffset | 61 | Debug explain of query with only offset shows limitNode wrapping selectNode and scanNode. |
| TestDebugExplainRequestWithLimitAndOffset | 84 | Debug explain of query with both limit and offset shows limitNode plan tree. |

---

### with_max_join_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainRequest_WithMaxOnOneToManyJoinedField_Succeeds | 40 | Debug explain of MAX on one-to-many joined field shows maxNode over typeJoinMany. |
| TestDebugExplainRequest_WithMaxOnOneToManyJoinedFieldWithFilter_Succeeds | 67 | Debug explain of MAX on one-to-many joined field with filter shows maxNode over typeJoinMany. |
| TestDebugExplainRequest_WithMaxOnOneToManyJoinedFieldWithManySources_Succeeds | 100 | Debug explain of MAX on multiple one-to-many joined fields shows parallelNode with typeJoinMany. |

---

### with_max_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainRequestWithMaxOnInlineArrayField_ChildFieldWillBeEmpty | 38 | Debug explain of MAX on inline array field shows maxNode over selectNode and scanNode. |

---

### with_min_join_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainRequest_WithMinOnOneToManyJoinedField_Succeeds | 40 | Debug explain of MIN on one-to-many joined field shows minNode over typeJoinMany. |
| TestDebugExplainRequest_WithMinOnOneToManyJoinedFieldWithFilter_Succeeds | 67 | Debug explain of MIN on one-to-many joined field with filter shows minNode over typeJoinMany. |
| TestDebugExplainRequest_WithMinOnOneToManyJoinedFieldWithManySources_Succeeds | 100 | Debug explain of MIN on multiple one-to-many joined fields shows parallelNode with typeJoinMany. |

---

### with_min_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainRequestWithMinOnInlineArrayField_ChildFieldWillBeEmpty | 38 | Debug explain of MIN on inline array field shows minNode over selectNode and scanNode. |

---

### with_order_join_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainRequestWithOrderFieldOnRelatedChild | 37 | Debug explain of join with order on related child shows orderNode in typeJoinMany subType. |
| TestDebugExplainRequestWithOrderFieldOnParentAndRelatedChild | 77 | Debug explain of join with order on parent and related child shows orderNode on both levels. |
| TestDebugExplainRequestWhereParentIsOrderedByItsRelatedChild | 119 | Debug explain of query ordered by a related child's field returns an error. |
| TestDebugExplainRequestWithSubqueryOrderByNestedRelationField | 179 | Debug explain of subquery ordered by nested relation field shows orderNode and limitNode in join. |
| TestDebugExplainRequestWithSubqueryOrderByNestedRelationFieldASC | 236 | Debug explain of subquery ordered ascending by nested relation field shows orderNode in join. |

---

### with_order_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainRequestWithAscendingOrderOnParent | 38 | Debug explain of query with ascending order shows orderNode wrapping selectNode and scanNode. |
| TestDebugExplainRequestWithMultiOrderFieldsOnParent | 62 | Debug explain of query with multiple order fields shows orderNode wrapping selectNode and scanNode. |

---

### with_orphan_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainRequestWithOrderByRelationFieldWithIndex | 74 | Debug explain of query ordered ascending by indexed relation field shows sequenceNode with orphanNode first. |
| TestDebugExplainRequestWithOrderByRelationFieldWithIndexDESC | 121 | Debug explain of query ordered descending by indexed relation field shows sequenceNode with join first. |
| TestDebugExplainRequestWithOrderByRelationFieldSecondaryParent | 168 | Debug explain of secondary parent ordered by related indexed field shows orphanNode wrapping join. |

---

### with_similarity_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainRequestWith_WithSimilarity | 38 | Debug explain of query with SIMILARITY function shows similarityNode in plan tree. |

---

### with_sum_join_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainRequestWithSumOnOneToManyJoinedField | 40 | Debug explain of SUM on one-to-many joined field shows sumNode over typeJoinMany. |
| TestDebugExplainRequestWithSumOnOneToManyJoinedFieldWithFilter | 67 | Debug explain of SUM on one-to-many joined field with filter shows sumNode over typeJoinMany. |
| TestDebugExplainRequestWithSumOnOneToManyJoinedFieldWithManySources | 100 | Debug explain of SUM on multiple one-to-many joined fields shows parallelNode with typeJoinMany. |

---

### with_sum_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainRequestWithSumOnInlineArrayField_ChildFieldWillBeEmpty | 38 | Debug explain of SUM on inline array field shows sumNode over selectNode and scanNode. |

---

### with_view_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainRequestWithView | 43 | Debug explain of query on a cacheless view shows viewNode wrapping selectTopNode and scanNode. |

---

### with_view_transform_test.go

| Test Function | Line | Description |
|---|---|---|
| TestDebugExplainRequestWithViewWithTransform | 47 | Debug explain of query on a view with lens transform shows lensNode in viewNode plan. |
