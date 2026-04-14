# Index: `tests/integration/explain/execute`

## Overview

This folder contains integration tests for the `execute` explain type (`@explain(type: execute)`). Each test runs a real query or mutation against the database and verifies the actual runtime execution statistics returned by the explain plan, including iteration counts, document fetches, field fetches, index fetches, and filter match counts across every plan node.

## Test Index

### `add_test.go`

Tests that an execute explain of an add mutation returns the expected addNode and scanNode runtime statistics.

| Test Function | Line | Description |
|---|---|---|
| `TestExecuteExplainMutationRequestWithAdd` | 22-67 | Execute explain of an add mutation returns actual node iteration stats. |

---

### `dagscan_test.go`

Tests that an execute explain of a `_commits` DAG-scan query returns accurate dagScanNode iteration counts.

| Test Function | Line | Description |
|---|---|---|
| `TestExecuteExplainCommitsDagScan` | 22-65 | Execute explain of a commits dag scan query returns dagScanNode iteration stats. |

---

### `delete_test.go`

Tests that execute explain of delete mutations by docID and by filter each produce correct deleteNode runtime metrics.

| Test Function | Line | Description |
|---|---|---|
| `TestExecuteExplainMutationRequestWithDeleteUsingID` | 22-70 | Execute explain of a delete mutation by docID returns deleteNode iteration stats. |
| `TestExecuteExplainMutationRequestWithDeleteUsingFilter` | 72-121 | Execute explain of a delete mutation by filter returns deleteNode iteration and scan stats. |

---

### `group_test.go`

Tests that an execute explain of a groupBy query returns correct groupNode statistics including group count and child selections.

| Test Function | Line | Description |
|---|---|---|
| `TestExecuteExplainRequestWithGroup` | 22-78 | Execute explain of a groupBy query returns groupNode with child selection and scan stats. |

---

### `query_deleted_docs_test.go`

Tests that an execute explain with `showDeleted: true` reflects deleted documents in the scan node iteration counts.

| Test Function | Line | Description |
|---|---|---|
| `TestExecuteExplainQueryDeletedDocs` | 22-77 | Execute explain of a query with showDeleted returns scan stats including deleted documents. |

---

### `scan_test.go`

Tests that execute explain of basic scan queries accurately reflects filterMatch counts across all-matching, none-matching, partial-matching, and empty-collection scenarios.

| Test Function | Line | Description |
|---|---|---|
| `TestExecuteExplainRequestWithAllDocumentsMatching` | 22-85 | Execute explain of a scan with all documents matching reports correct filterMatches and docFetches. |
| `TestExecuteExplainRequestWithNoDocuments` | 87-129 | Execute explain of a scan on an empty collection shows zero fetches and zero filterMatches. |
| `TestExecuteExplainRequestWithSomeDocumentsMatching` | 131-194 | Execute explain of a filtered scan where only some documents match reports partial filterMatches. |
| `TestExecuteExplainRequestWithDocumentsButNoMatches` | 196-259 | Execute explain of a scan where no documents match filter reports zero filterMatches. |

---

### `top_level_test.go`

Tests that execute explain of top-level aggregate queries (AVG, COUNT, SUM) returns the correct aggregation node hierarchy and iteration counts.

| Test Function | Line | Description |
|---|---|---|
| `TestExecuteExplainTopLevelAverageRequest` | 22-109 | Execute explain of a top-level AVG aggregate returns averageNode, sumNode, and countNode stats. |
| `TestExecuteExplainTopLevelCountRequest` | 111-181 | Execute explain of a top-level COUNT aggregate returns countNode with scan stats. |
| `TestExecuteExplainTopLevelSumRequest` | 183-257 | Execute explain of a top-level SUM aggregate returns sumNode with scan stats. |

---

### `type_join_test.go`

Tests that execute explain of one-to-one join queries reports correct typeJoinOne statistics at single, parallel, nested, and secondary-side join configurations.

| Test Function | Line | Description |
|---|---|---|
| `TestExecuteExplainRequestWithAOneToOneJoin` | 22-91 | Execute explain of a one-to-one join query returns typeJoinOne with root and subType scan stats. |
| `TestExecuteExplainWithMultipleOneToOneJoins` | 93-199 | Execute explain of multiple one-to-one joins shows a parallelNode with separate typeIndexJoin stats. |
| `TestExecuteExplainWithTwoLevelDeepNestedJoins` | 201-295 | Execute explain of two-level deep nested one-to-one joins returns nested typeJoinOne stats. |
| `TestExecuteExplain_WithOneToOneJoinFromSecondarySide_ShouldIncludeIndex` | 297-367 | Execute explain of a one-to-one join from secondary side shows non-zero indexFetches in subType. |

---

### `update_test.go`

Tests that execute explain of update mutations by docID list and by filter each return correct updateNode runtime metrics.

| Test Function | Line | Description |
|---|---|---|
| `TestExecuteExplainMutationRequestWithUpdateUsingIDs` | 22-84 | Execute explain of an update mutation by docIDs returns updateNode with update count and scan stats. |
| `TestExecuteExplainMutationRequestWithUpdateUsingFilter` | 86-149 | Execute explain of an update mutation by filter returns updateNode iteration and filtered scan stats. |

---

### `upsert_test.go`

Tests that execute explain of upsert mutations correctly distinguishes update-path vs insert-path statistics based on filter matching.

| Test Function | Line | Description |
|---|---|---|
| `TestExecuteExplainMutationRequest_WithUpsertAndMatchingFilter_Succeeds` | 22-68 | Execute explain of an upsert mutation with a matching filter shows upsertNode update path stats. |
| `TestExecuteExplainMutationRequest_WithUpsertAndNoMatchingFilter_Succeeds` | 70-119 | Execute explain of an upsert mutation with no matching filter shows upsertNode insert path stats. |

---

### `with_average_test.go`

Tests that execute explain of AVG aggregations returns the correct averageNode, countNode, and sumNode hierarchy for both inline array fields and joined relations.

| Test Function | Line | Description |
|---|---|---|
| `TestExecuteExplainAverageRequestOnArrayField` | 22-79 | Execute explain of AVG on an inline array field returns averageNode, countNode, and sumNode stats. |
| `TestExplainExplainAverageRequestOnJoinedField` | 81-159 | Execute explain of AVG on a joined one-to-many field returns averageNode wrapping a typeJoinMany. |

---

### `with_count_test.go`

Tests that execute explain of a COUNT on a one-to-many relation returns the correct countNode wrapping a typeJoinMany.

| Test Function | Line | Description |
|---|---|---|
| `TestExecuteExplainRequestWithCountOnOneToManyRelation` | 22-94 | Execute explain of COUNT on a one-to-many relation returns countNode wrapping a typeJoinMany. |

---

### `with_index_test.go`

Tests that execute explain correctly reports non-zero indexFetches when queries use indexed fields for filtering, ordering, and nested relation ordering.

| Test Function | Line | Description |
|---|---|---|
| `TestExecuteExplainWithIndexOnFilter` | 22-87 | Execute explain of an equality filter on an indexed field reports non-zero indexFetches. |
| `TestExecuteExplainWithIndexOnOrder` | 89-155 | Execute explain of an ASC order on an indexed field reports non-zero indexFetches in scan. |
| `TestExecuteExplainWithIndexOnRelationOrder` | 157-277 | Execute explain of ordering by an indexed relation field shows orphanNode and index fetches in join. |
| `TestExecuteExplainWithIndexOnSubqueryNestedRelationOrder` | 279-425 | Execute explain of subquery ordering by a nested indexed relation shows index fetches eliminating orderNode. |

---

### `with_limit_test.go`

Tests that execute explain of queries with limit and offset returns correct limitNode iteration statistics at the parent level and in nested child joins.

| Test Function | Line | Description |
|---|---|---|
| `TestExecuteExplainRequestWithBothLimitAndOffsetOnParent` | 22-72 | Execute explain of a query with limit and offset on the parent returns limitNode iteration stats. |
| `TestExecuteExplainRequestWithBothLimitAndOffsetOnParentAndLimitOnChild` | 74-151 | Execute explain of a query with limit on parent and child returns nested limitNode stats. |

---

### `with_max_test.go`

Tests that execute explain of MAX aggregations returns the correct maxNode statistics for both inline array fields and one-to-many joined fields.

| Test Function | Line | Description |
|---|---|---|
| `TestExecuteExplainRequest_WithMaxOfInlineArrayField_Succeeds` | 22-73 | Execute explain of MAX on an inline array field returns maxNode with scan iteration stats. |
| `TestExecuteExplainRequest_MaxOfRelatedOneToManyField_Succeeds` | 75-150 | Execute explain of MAX on a one-to-many joined field returns maxNode wrapping a typeJoinMany. |

---

### `with_min_test.go`

Tests that execute explain of MIN aggregations returns the correct minNode statistics for both inline array fields and one-to-many joined fields.

| Test Function | Line | Description |
|---|---|---|
| `TestExecuteExplainRequest_WithMinOfInlineArrayField_Succeeds` | 22-73 | Execute explain of MIN on an inline array field returns minNode with scan iteration stats. |
| `TestExecuteExplainRequest_MinOfRelatedOneToManyField_Succeeds` | 75-150 | Execute explain of MIN on a one-to-many joined field returns minNode wrapping a typeJoinMany. |

---

### `with_order_test.go`

Tests that execute explain of ordered queries reports correct orderNode statistics across single-field, multi-field, child-side, combined parent-and-child, and subquery nested relation ordering scenarios.

| Test Function | Line | Description |
|---|---|---|
| `TestExecuteExplainRequestWithOrderFieldOnParent` | 22-72 | Execute explain of a single-field order on the parent returns orderNode iteration stats. |
| `TestExecuteExplainRequestWithMultiOrderFieldsOnParent` | 74-157 | Execute explain of a multi-field order on the parent returns orderNode with all document iterations. |
| `TestExecuteExplainRequestWithOrderFieldOnChild` | 159-233 | Execute explain of order on a child relation shows orderNode inside the join subType. |
| `TestExecuteExplainRequestWithOrderFieldOnBothParentAndChild` | 235-313 | Execute explain of order on both parent and child returns orderNode at each level of the plan. |
| `TestExecuteExplainRequestWhereParentFieldIsOrderedByChildField` | 315-347 | Execute explain ordering a parent by a child relation field returns an expected error. |
| `TestExecuteExplainRequestWithSubqueryOrderByNestedRelationField` | 349-501 | Execute explain of a subquery ordered DESC by a nested relation field shows limitNode and orderNode stats. |
| `TestExecuteExplainRequestWithSubqueryOrderByNestedRelationFieldASC` | 503-655 | Execute explain of a subquery ordered ASC by a nested relation field shows limitNode and orderNode stats. |

---

### `with_orphan_test.go`

Tests that execute explain with the `@exhaustive` directive correctly reports orphanNode metrics when primary-side and secondary-side join roots have unlinked documents.

| Test Function | Line | Description |
|---|---|---|
| `TestExecuteExplainWithOrphanNode_WithPrimaryParent_ReportsMetrics` | 22-86 | Execute explain with a primary-side join and orphan documents shows orphanNode iteration stats. |
| `TestExecuteExplainWithOrphanNode_WithSecondaryParent_ReportsMetrics` | 88-152 | Execute explain with a secondary-side join and orphan documents shows orphanNode indexFetches stats. |

---

### `with_similarity_test.go`

Tests that execute explain of a SIMILARITY vector query returns the correct similarityNode wrapping the selectNode.

| Test Function | Line | Description |
|---|---|---|
| `TestExecuteExplainRequest_WithSimilarity` | 22-76 | Execute explain of a SIMILARITY query returns similarityNode wrapping the selectNode stats. |

---

### `with_sum_test.go`

Tests that execute explain of SUM aggregations returns the correct sumNode statistics for both inline array fields and one-to-many joined fields.

| Test Function | Line | Description |
|---|---|---|
| `TestExecuteExplainRequestWithSumOfInlineArrayField` | 22-73 | Execute explain of SUM on an inline array field returns sumNode with scan iteration stats. |
| `TestExecuteExplainRequestSumOfRelatedOneToManyField` | 75-150 | Execute explain of SUM on a one-to-many joined field returns sumNode wrapping a typeJoinMany. |
