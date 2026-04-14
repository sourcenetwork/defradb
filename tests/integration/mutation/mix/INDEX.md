# Index: `tests/integration/mutation/mix`

## Overview

This folder contains integration tests for mixed mutation scenarios in DefraDB, focusing on how add, update, and delete mutations interact when executed within the same or across different concurrent transactions. The tests verify transaction isolation semantics, including visibility of in-flight writes, conflict detection on concurrent updates, and correct document state after commit.

## Test Index

### `with_txn_test.go`

Tests covering transactional isolation for add, update, and delete mutations across same and different concurrent transactions.

| Test Function | Line | Description |
|---|---|---|
| `TestMutationWithTxnDeletesUserGivenSameTransaction` | 24-70 | Add and delete a document within the same transaction succeeds. |
| `TestMutationWithTxnDoesNotDeletesUserGivenDifferentTransactions` | 72-153 | Delete in a separate transaction does not affect documents added in another open transaction. |
| `TestMutationWithTxnDoesUpdateUserGivenSameTransactions` | 155-211 | Update within the same transaction is visible to subsequent queries in that transaction. |
| `TestMutationWithTxnDoesNotUpdateUserGivenDifferentTransactions` | 213-280 | Update in one transaction is not visible to a concurrent separate transaction. |
| `TestMutationWithTxnDoesNotAllowUpdateInSecondTransactionUser` | 282-376 | Concurrent updates in two transactions cause a conflict error on the second commit. |
