# Index: `tests/integration/mutation/add/field_kinds/one_to_one_to_one`

## Overview

This folder tests transactional creation and relational linking of documents across a one-to-one-to-one chain (Author — Book — Publisher). The shared schema is defined in `utils.go` and each test exercises concurrent transaction semantics, verifying that relational links are correctly visible within and across transactions and that serializable snapshot isolation (SSI) conflict detection behaves as expected.

## Test Index

### `with_txn_test.go`

Tests that concurrent transactions correctly create and link documents in a one-to-one-to-one relationship, covering both forward (Publisher→Book) and backward (Book→Publisher) query directions, including SSI conflict resolution.

| Test Function | Line | Description |
|---|---|---|
| `TestTransactionalCreationAndLinkingOfRelationalDocumentsForward` | 24-185 | Concurrent transactions creating linked Book-Publisher documents resolve with SSI conflict on second commit. |
| `TestTransactionalCreationAndLinkingOfRelationalDocumentsBackward` | 187-342 | Concurrent transactions creating linked Book-Publisher documents both commit successfully when queried backward. |
