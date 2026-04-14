# Index: `tests/integration/explain/simple`

## Overview

This folder contains integration tests for the `simple` explain type (`@explain(type: simple)`). Each test issues an explain request and verifies the logical shape of the query plan — the node tree structure, collection metadata, filter settings, and prefix configuration — without executing the query or collecting runtime statistics.

## Test Index

### `basic_test.go`

Tests that a simple explain of a straightforward query returns the full logical plan graph including selectNode, scanNode, collection name, and storage prefixes.

| Test Function | Line | Description |
|---|---|---|
| `TestSimpleExplainRequest` | 24-67 | Simple explain of a basic query shows selectNode, scanNode, collection name, and prefixes. |

---

### `with_index_test.go`

Tests that a simple explain of queries using indexed fields shows the correct plan shape — verifying that index usage eliminates explicit orderNode nodes when ordering is satisfied by the index.

| Test Function | Line | Description |
|---|---|---|
| `TestSimpleExplainWithIndexOnFilter` | 22-61 | Simple explain of an equality filter on an indexed field shows a scanNode in the plan shape. |
| `TestSimpleExplainWithIndexOnOrder` | 63-103 | Simple explain of an ASC order on an indexed field shows a scanNode with no explicit orderNode. |
| `TestSimpleExplainWithIndexOnSubqueryNestedRelationOrder` | 105-183 | Simple explain of subquery ordering by a nested indexed relation shows limitNode without an orderNode. |
