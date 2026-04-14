# Index: `tests/integration/mutation/delete/field_kinds/one_to_one_to_one`

## Overview

This folder contains integration tests for delete mutations across a three-type one-to-one-to-one relational chain (Author, Book, Publisher). The tests verify that documents can be deleted by docID on both the primary and non-primary sides of the relationship, that result fields can be aliased, and that transactional isolation is correctly maintained — including the ability of a concurrent transaction to read a record that has been deleted but not yet committed in another transaction.

## Test Index

### `with_id_test.go`

Tests that delete a single related document by docID across the one-to-one-to-one chain, covering plain deletion, aliased result fields, and deletion from a multi-document dataset.

| Test Function | Line | Description |
|---|---|---|
| `TestRelationalDeletionOfADocumentUsingSingleKey_Success` | 21-73 | Delete a related document by single docID in a one-to-one-to-one chain. |
| `TestRelationalDeletionOfADocumentUsingSingleKeyWithAlias_Success` | 75-127 | Delete a related document by docID and return the result using a field alias. |
| `TestRelationalDeletionOfADocumentUsingSingleKeyWithMultiDocumentsWithAlias_Success` | 129-198 | Delete one document by docID from a multi-document one-to-one-to-one chain using an alias. |

### `with_txn_test.go`

Tests that verify transactional delete behaviour across the one-to-one-to-one chain, including both forward and backward query directions and cross-transaction read isolation before commit.

| Test Function | Line | Description |
|---|---|---|
| `TestTxnDeletionOfRelatedDocFromPrimarySideForwardDirection` | 24-92 | Transactional deletion from the primary side verifies the linked doc is unset when queried forward. |
| `TestTxnDeletionOfRelatedDocFromPrimarySideBackwardDirection` | 94-156 | Transactional deletion from the primary side verifies the deleted doc is absent when queried backward. |
| `TestATxnCanReadARecordThatIsDeletedInANonCommitedTxnForwardDirection` | 158-259 | A concurrent transaction reads a record deleted in an uncommitted transaction via forward traversal. |
| `TestATxnCanReadARecordThatIsDeletedInANonCommitedTxnBackwardDirection` | 261-356 | A concurrent transaction reads a record deleted in an uncommitted transaction via backward traversal. |
| `TestTxnDeletionOfRelatedDocFromNonPrimarySideForwardDirection` | 358-421 | Transactional deletion from the non-primary side confirms the publisher is removed when queried forward. |
| `TestTxnDeletionOfRelatedDocFromNonPrimarySideBackwardDirection` | 423-492 | Transactional deletion from the non-primary side leaves the book with a nil publisher when queried backward. |
