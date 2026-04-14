# Index: `tests/integration/query/commits/branchables`

## Overview

This folder contains integration tests for the `_commits` query on collections annotated with the `@branchable` directive. The tests verify that collection-level commits are created and linked correctly in the DAG, that filtering and parameter options (cid, docID, fieldName) behave as expected, and that multi-node peer scenarios — including document create, update, delete, and concurrent offline edits — converge to strong eventual consistency.

## Test Index

### `add_test.go`

Tests that adding multiple documents to a branchable collection produces the correct set of collection-level and document-level commits.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryCommitsBranchables_WithMultipleAdd` | 23-145 | Query all commits on a branchable collection after adding two documents. |

### `cid_doc_id_test.go`

Tests that combining the `cid` and `docID` parameters when the cid belongs to the collection (not the document) returns an expected error.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryCommitsBranchables_WithCidAndDocIDParam` | 21-60 | Querying commits with both cid and docID params returns an error for mismatched cid. |

### `cid_test.go`

Tests that querying commits by a known collection-level cid returns that commit with nil docID and fieldName, confirming it is a collection-level commit.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryCommitsBranchables_WithCidParam` | 21-64 | Querying a commit by cid on a branchable collection returns the collection-level commit. |

### `delete_test.go`

Tests the commit DAG structure after a document is created and then deleted, verifying that a collection-level delete commit correctly links to the previous create commit.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryCommitsBranchables_WithDelete` | 23-127 | Querying commits after deleting a document shows correct collection-level DAG with delete commit. |

### `field_id_test.go`

Tests that filtering commits by `fieldName: {_eq: null}` returns only the collection-level commit, which has no associated field or document.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryCommitsBranchables_WithFieldNameFilter` | 21-63 | Filtering commits by null fieldName returns only the collection-level commit. |

### `if_test.go`

Tests the effect of the `@branchable(if: <bool>)` argument, confirming that `true` enables collection-level commits and `false` suppresses them.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryCommitsBranchables_WithIfDirectiveTrue` | 23-75 | Collection with @branchable(if: true) produces a collection-level commit on document create. |
| `TestQueryCommitsBranchables_WithIfDirectiveFalse` | 77-127 | Collection with @branchable(if: false) omits the collection-level commit from query results. |

### `peer_index_test.go`

Tests that a document synced over a peer connection to a subscribing node has its index correctly constructed so that indexed queries work on the receiving node.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryCommitsBranchables_SyncsIndexAcrossPeerConnection` | 23-71 | An indexed field on a branchable collection is correctly usable after syncing across a peer connection. |

### `peer_test.go`

Tests that commits on a branchable collection (including collection-level commits) sync correctly to a subscribing peer node after one or multiple documents are added.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryCommitsBranchables_SyncsAcrossPeerConnection` | 25-114 | Commits on a branchable collection sync correctly across a peer connection to a subscribing node. |
| `TestQueryCommitsBranchables_SyncsMultipleAcrossPeerConnection` | 116-251 | Multiple documents on a branchable collection sync with correct collection DAG across peers. |

### `peer_update_test.go`

Tests that concurrent document updates made on two separate nodes before and after establishing a peer connection eventually converge to strong eventual consistency with a correct full commit DAG on both nodes.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryCommitsBranchables_HandlesConcurrentUpdatesAcrossPeerConnection` | 25-317 | Concurrent updates across two peers on a branchable collection reach strong eventual consistency. |

### `simple_test.go`

Tests basic commit queries on a branchable collection, verifying commit counts and the full set of commit metadata fields for a single document.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryCommitsBranchables` | 23-70 | Querying all commits on a branchable collection returns four unique CIDs including the collection commit. |
| `TestQueryCommitsBranchables_WithAllFields` | 72-178 | Querying all commit fields on a branchable collection returns correct metadata for all commit types. |

### `update_test.go`

Tests the commit DAG structure after a document is created and then updated, verifying that collection-level update and create commits link correctly and that field-level heads are maintained.

| Test Function | Line | Description |
|---|---|---|
| `TestQueryCommitsBranchables_WithDocUpdate` | 23-143 | Querying commits after a document update shows correct collection and document DAG with heads. |
