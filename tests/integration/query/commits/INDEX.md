# Index: `tests/integration/query/commits`

## Overview

This folder contains integration tests for the `_commits` GraphQL query, which exposes the underlying Merkle-DAG commit history of documents in DefraDB. Tests cover basic commit retrieval, filtering by CID, docID, fieldName, depth, grouping, ordering, limit/offset pagination, null-input tolerance, nested link traversal, deletion semantics, and compound filter operators. A `branchables/` subdirectory exercises the same query surface on collections annotated with the `@branchable` directive, including multi-node peer scenarios.

## Test Index

### `simple_test.go`

Basic queries against `_commits` verifying field retrieval (cid, fieldName, collectionVersionId, delta, links, heads) with and without document updates and aliases.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryCommits` | 23-65 | Query all commits for a single created document returns three unique CIDs. |
| `TestQueryCommitsMultipleDocs` | 67-119 | Query all commits across multiple documents returns all field and composite CIDs. |
| `TestQueryCommitsWithCollectionVersionIDField` | 121-161 | Query commits requesting the collectionVersionId field returns the correct schema version CID. |
| `TestQueryCommitsWithFieldNameField` | 163-200 | Query commits requesting fieldName returns age, name, and composite (_C) entries. |
| `TestQueryCommitsWithFieldNameFieldAndUpdate` | 202-250 | Query commits after an update returns duplicate age and composite fieldName entries. |
| `TestQuery_CommitsWithAllFieldsWithUpdate_NoError` | 252-379 | Query all commit fields after an update returns full metadata including links and heads. |
| `TestQueryCommits_WithAlias_Succeeds` | 381-417 | Query commits using a field alias returns results under the aliased name. |

### `with_cid_test.go`

Tests filtering the `_commits` query by a specific CID value, including valid field commit CIDs, invalid CID strings, and list-of-CID inputs.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryCommits_WithFirstCommitCid_ShouldSucceed` | 21-60 | Query commits filtered by a specific composite commit CID returns only that commit. |
| `TestQueryCommits_WithFirstCommitCidForFieldCommit_ShouldSucceed` | 62-100 | Query commits filtered by a field commit CID returns only that field commit. |
| `TestQueryCommitsWithInvalidCid` | 102-131 | Query commits with a totally invalid CID string returns an invalid cid error. |
| `TestQueryCommitsWithInvalidShortCid` | 133-162 | Query commits with a short but syntactically invalid CID returns an invalid cid error. |
| `TestQueryCommitsWithUnknownCid` | 164-193 | Query commits with a valid-format but non-existent CID returns a does-not-exist error. |
| `TestQueryCommits_MultipleCids` | 195-228 | Query commits with an array of multiple CIDs returns an unsupported error. |
| `TestQueryCommits_ListOfOne` | 229-261 | Query commits with a one-element CID array returns the single matching commit. |

### `with_delete_test.go`

Tests that `_commits` still returns commit history after a document has been soft-deleted.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryCommits_AfterDocDeletion_ShouldStillFetch` | 23-101 | Query composite commits after document deletion returns both delete and create composite commits. |

### `with_depth_test.go`

Tests the `depth` parameter, which limits how many commit heights back from each head the query traverses.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryCommitsWithDepth1` | 21-57 | Query commits with depth 1 on a newly created document returns all three head commits. |
| `TestQueryCommitsWithDepth1WithUpdate` | 59-108 | Query commits with depth 1 after an update returns only the current head commits at height 2 and 1. |
| `TestQueryCommitsWithDepth2WithUpdate` | 110-177 | Query commits with depth 2 after two updates returns the two most recent heights for each field. |
| `TestQueryCommitsWithDepth1AndMultipleDocs` | 179-231 | Query commits with depth 1 across two documents returns all head commits for both. |
| `TestQueryCommits_WithFilterFieldNameAndDepth_ReturnsCommitsAtAllHeights` | 233-274 | Filter by fieldName with depth 2 after two updates returns commits at both heights for that field. |

### `with_doc_id_cid_test.go`

Tests combining the `docID` and `cid` parameters, including mismatched pairs and combinations with `depth`.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryCommitsWithDocIDAndCidForDifferentDoc` | 21-51 | Query commits with a docID and a CID belonging to a different document returns an error. |
| `TestQueryCommitsWithDocIDAndCidForDifferentDocWithUpdate` | 53-90 | After an update, querying with a mismatched docID and CID still returns an error. |
| `TestQueryCommits_WithDocIDAndCidWithUpdate` | 92-132 | Query commits with matching docID and CID after update returns only that specific commit. |
| `TestQueryCommitsWithDocIDAndCidWithUpdateAndDepth` | 134-181 | Query commits with docID, CID, and depth returns the target commit and its ancestor. |

### `with_doc_id_count_test.go`

Tests using the `COUNT` aggregate on the `links` and `heads` sub-fields within a `_commits` query filtered by `docID`.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryCommitsWithDocIDAndLinkCount` | 21-62 | Query commits with docID requesting COUNT(links) returns correct link counts per commit. |
| `TestQueryCommits_WithDocUpdatesAndLinkHeadCount` | 64-124 | Query commits with aliased COUNT(links) and COUNT(heads) after an update returns correct counts. |

### `with_doc_id_field_test.go`

Tests filtering commits by both `docID` and `fieldName` (including unknown, numeric ID, named field, and composite `_C`).

| Test Function | Line | Description |
|---|---|---|
| `TestQueryCommitsWithDocIDAndUnknownField` | 21-50 | Query commits with a docID and filter on an unknown fieldName returns an empty result. |
| `TestQueryCommitsWithDocIDAndUnknownFieldId` | 52-81 | Query commits with docID and filter on a numeric field ID string returns empty result. |
| `TestQueryCommitsWithDocIDAndField` | 83-116 | Query commits filtered by docID and fieldName 'age' returns only the age field commit. |
| `TestQueryCommitsWithDocIDAndCompositeField` | 118-151 | Query commits filtered by docID and fieldName '_C' returns only the composite commit. |

### `with_doc_id_group_order_test.go`

Tests combining `groupBy` and `order` parameters when filtering commits by `docID`.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryCommitsOrderedAndGroupedByDocID` | 21-61 | Group and order commits by docID descending returns two distinct docIDs in reverse order. |

### `with_doc_id_limit_offset_test.go`

Tests applying `limit` and `offset` together to a `_commits` query filtered by `docID`.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryCommitsWithDocIDAndLimitAndOffset` | 21-76 | Query commits with docID, limit 2, and offset 1 returns the two middle commits. |

### `with_doc_id_limit_test.go`

Tests applying a `limit` parameter to a `_commits` query filtered by `docID`.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryCommitsWithDocIDAndLimit` | 21-69 | Query commits with docID and limit 2 after two updates returns only two commits. |

### `with_doc_id_order_limit_offset_test.go`

Tests combining `order`, `limit`, and `offset` parameters together with a `docID` filter.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryCommitsWithDocIDAndOrderAndLimitAndOffset` | 21-78 | Query commits with docID, ASC order by height, limit 2, and offset 4 returns middle commits. |

### `with_doc_id_order_test.go`

Tests ordering commits by `height` and `cid` in both ascending and descending directions for a single `docID`.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryCommitsWithDocIDAndOrderHeightDesc` | 21-76 | Order commits by height descending places the update commit before create commits. |
| `TestQueryCommitsWithDocIDAndOrderHeightAsc` | 78-133 | Order commits by height ascending places the create commits before the update commit. |
| `TestQueryCommitsWithDocIDAndOrderCidDesc` | 135-190 | Order commits by CID descending returns commits in reverse lexicographic CID order. |
| `TestQueryCommitsWithDocIDAndOrderCidAsc` | 192-247 | Order commits by CID ascending returns commits in lexicographic CID order. |
| `TestQueryCommitsWithDocIDAndOrderAndMultiUpdatesCidAsc` | 249-334 | Order commits by height ascending after three updates returns all commits in chronological order. |

### `with_doc_id_prop_test.go`

Tests that the `docID` property field is correctly returned on every commit in the result set.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryCommitsWithDocIDProperty` | 21-57 | Query commits requesting the docID field returns the same document ID for all commits. |

### `with_doc_id_test.go`

Tests filtering commits by `docID`, including unknown IDs, valid IDs, link/head sub-fields, updates, and list inputs.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryCommitsWithUnknownDocID` | 21-47 | Query commits with an unknown docID returns an empty result set. |
| `TestQueryCommitsWithDocID` | 49-86 | Query commits filtered by a valid docID returns all three commits for that document. |
| `TestQueryCommitsWithDocIDAndLinks` | 88-147 | Query commits with docID requesting links shows composite commit links to field commits. |
| `TestQueryCommitsWithDocIDAndUpdate` | 149-205 | Query commits with docID after an update returns all five commits with correct heights. |
| `TestQueryCommitsWithDocIDAndUpdateAndLinks` | 210-299 | Query commits with links and heads after an update shows updated composite links and heads. |
| `TestQueryCommits_DocIDEmptyList` | 301-330 | Query commits with an empty docID list returns an empty result set. |
| `TestQueryCommits_DocIDListOfOne` | 332-368 | Query commits with a one-element docID list returns commits only for that document. |
| `TestQueryCommits_DocIDListOfMany` | 370-397 | Query commits with multiple docIDs in a list returns an unsupported error. |

### `with_doc_id_typename_test.go`

Tests that the `__typename` introspection field on commits returns the expected `"Commit"` value.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryCommitsWithDocIDWithTypeName` | 21-62 | Query commits with docID requesting __typename returns 'Commit' for every result. |

### `with_field_test.go`

Tests filtering commits by `fieldName` using equality, inequality, and CID combinations, including composite and invalid field names.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryCommitsWithField` | 21-51 | Filter commits by fieldName 'age' returns only the age field commit. |
| `TestQueryCommitsWithFieldId` | 53-79 | Filter commits by a numeric field ID string returns an empty result. |
| `TestQueryCommitsWithCompositeField` | 81-111 | Filter commits by fieldName '_C' returns only the composite commit. |
| `TestQueryCommitsWithCompositeFieldIdWithReturnedCollectionVersionID` | 115-147 | Filter composite commit and return collectionVersionId field shows the schema version CID. |
| `TestQueryCommits_WithFilterFieldNameNotEqualComposite_ReturnsFieldCommits` | 149-181 | Filter commits where fieldName is not '_C' returns only field-level commits. |
| `TestQueryCommitsWithFieldAndCID` | 183-216 | Filter commits by matching fieldName and CID returns the single matching commit. |
| `TestQueryCommits_WithWrongFieldAndCID_ReturnEmptyList` | 218-247 | Filter commits by a fieldName that does not match the CID's field returns empty results. |
| `TestQueryCommits_WithInvalidFieldAndCID_ReturnEmptyList` | 249-278 | Filter commits by a non-existent fieldName and a valid CID returns empty results. |

### `with_filter_compound_test.go`

Tests compound `_or` and `_and` filter conditions on commit fieldName.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryCommits_WithFilterFieldNameOrCondition_ReturnsMatchingCommits` | 21-53 | Filter commits with _or on two fieldNames returns commits for both matched fields. |
| `TestQueryCommits_WithFilterFieldNameAndCondition_ReturnsOnlyNameCommit` | 55-84 | Filter commits with _and excluding composite and age returns only the name commit. |

### `with_filter_in_test.go`

Tests `_in` and `_nin` filter operators on commit fieldName, including combinations with `_and` and `_or`.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryCommits_WithFilterFieldNameIn_ReturnsMatchingCommits` | 21-49 | Filter commits with _in for age and name returns both field-level commits. |
| `TestQueryCommits_WithFilterFieldNameInComposite_ReturnsCompositeCommit` | 51-78 | Filter commits with _in containing '_C' returns only the composite commit. |
| `TestQueryCommits_WithFilterFieldNameInEmpty_ReturnsNoCommits` | 80-105 | Filter commits with an empty _in list returns no commits. |
| `TestQueryCommits_WithFilterFieldNameNotIn_ExcludesMatchingCommits` | 107-134 | Filter commits with _nin excluding composite and age returns only the name commit. |
| `TestQueryCommits_WithFilterFieldNameNotInComposite_ExcludesCompositeCommit` | 136-164 | Filter commits with _nin excluding '_C' returns only field-level commits. |
| `TestQueryCommits_WithFilterFieldNameNotInEmpty_ReturnsAllCommits` | 166-195 | Filter commits with an empty _nin list returns all commits. |
| `TestQueryCommits_WithFilterFieldNameInAndCondition_ReturnsFilteredCommits` | 197-224 | Filter commits with _in combined with _and to further exclude age returns only name commit. |
| `TestQueryCommits_WithFilterFieldNameNotInOrCondition_ReturnsFilteredCommits` | 226-254 | Filter commits with _nin combined in _or to also include composite returns age and composite. |

### `with_group_test.go`

Tests the `groupBy` parameter on `_commits`, covering grouping by height, cid, docID, and fieldName, with and without GROUP child selections.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryCommitsWithGroupBy` | 21-61 | Group commits by height returns one entry per distinct height value. |
| `TestQueryCommitsWithGroupByHeightWithChild` | 63-125 | Group commits by height with a GROUP child returns individual commits nested under each height. |
| `TestQueryCommitsWithGroupByCidWithChild` | 128-182 | Group commits by CID with a GROUP child returns each commit grouped by its unique CID. |
| `TestQueryCommitsWithGroupByDocID` | 184-238 | Group commits by docID across two documents returns one entry per document. |
| `TestQueryCommitsWithGroupByFieldName` | 240-283 | Group commits by fieldName returns one entry per distinct field including composite. |
| `TestQueryCommitsWithGroupByFieldNameWithChild` | 285-352 | Group commits by fieldName with a GROUP child nests commit heights under each field name. |

### `with_nested_links_test.go`

Tests querying deeply nested `links` and `heads` sub-fields on commits, including filters applied at both the top level and within nested selections.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryCommits_WithSingleAddNestedLinks_Succeed` | 23-97 | Query all commits with nested links shows field links within the composite commit. |
| `TestQueryCommits_WithSingleAddNestedLinksCompositeFilter_Succeed` | 99-145 | Filter to composite commit and query nested links returns correct linked field commits. |
| `TestQueryCommits_WithSingleAddNestedLinksNestedFilter_Succeed` | 147-189 | Filter composite commits and apply a nested filter on links to return only the age link. |
| `TestQueryCommits_WithSingleUpdateDoubleNestedLinks_Succeeds` | 191-352 | Query all commits after an update with double-nested links and heads returns full DAG metadata. |

### `with_null_input_test.go`

Tests that passing `null` for each optional `_commits` parameter (depth, cid, filter fieldName, order, docID, limit, offset, groupBy) behaves as if the parameter were omitted.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryCommitsWithNullDepth` | 21-57 | Query commits with depth set to null returns all commits as if no depth were specified. |
| `TestQueryCommitsWithNullCID` | 59-95 | Query commits with cid set to null returns all commits as if no CID filter were applied. |
| `TestQueryCommitsWithNullField` | 97-123 | Filter commits by fieldName equal to null returns an empty result set. |
| `TestQueryCommitsWithNullOrder` | 125-161 | Query commits with order set to null returns all commits in default order. |
| `TestQueryCommitsWithNullOrderField` | 163-199 | Query commits with docID set to null returns all commits as if no docID filter were applied. |
| `TestQueryCommitsWithNullLimit` | 201-237 | Query commits with limit set to null returns all commits without any limit. |
| `TestQueryCommitsWithNullOffset` | 239-275 | Query commits with offset set to null returns all commits starting from the beginning. |
| `TestQueryCommitsWithNullGroupBy` | 277-313 | Query commits with groupBy set to null returns all commits without any grouping. |

## Subdirectories

| Directory | Summary |
|---|---|
| [`branchables/`](branchables/INDEX.md) | Tests for `_commits` on collections with the `@branchable` directive, covering collection-level commit DAG structure, peer sync, concurrent updates, and filtering across create, update, and delete operations. |
