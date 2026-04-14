# Index: `tests/integration/collection/truncate`

## Overview

This folder contains integration tests for the `Truncate` collection operation in DefraDB. The tests verify that truncating a collection correctly removes all stored documents, commit blocks, and index entries, and that the collection can be re-populated cleanly afterwards. Coverage spans standard collections, branchable collections, materialized views, DAC (decentralised access control) policy enforcement, indexed field filtering, and concurrent truncation scenarios.

## Test Index

### `simple_test.go`

Baseline smoke test confirming that truncating a collection completes without error.

| Test Function | Line | Description |
|---|---|---|
| `TestTruncateCollection` | 21-39 | Truncating an empty collection completes without error. |

---

### `add_test.go`

Tests that verify truncation removes documents, blocks, and metadata, and that re-adding the same document after truncation yields consistent identifiers and block heights.

| Test Function | Line | Description |
|---|---|---|
| `TestTruncateCollectionAdd_RemovesDocument` | 21-55 | Truncating a collection removes a previously added document. |
| `TestTruncateCollectionAdd_RemovesSignedDocument` | 61-96 | Truncating a collection removes a previously added digitally signed document. |
| `TestTruncateCollectionAdd_RemovesEncryptedDocument` | 98-133 | Truncating a collection removes a previously added encrypted document. |
| `TestTruncateCollectionAdd_RemovesBlocks` | 135-169 | Truncating a collection removes all commit blocks associated with a document. |
| `TestTruncateCollectionAdd_AddsDocWithSameDocIDAsOriginal` | 171-233 | A document re-added after truncation has the same docID as the original. |
| `TestTruncateCollectionAdd_AddsDocWithSameCIDAsOriginal` | 235-297 | A document re-added after truncation has the same composite commit CID as the original. |
| `TestTruncateCollectionAdd_AddsDocWithBlocksAtHeight1` | 299-353 | A document re-added after truncation has all its commit blocks at height 1. |

---

### `branchable_add_test.go`

Tests that truncation correctly removes documents and commit blocks from branchable collections.

| Test Function | Line | Description |
|---|---|---|
| `TestTruncateCollectionBranchableAdd_RemovesDocument` | 21-55 | Truncating a branchable collection removes a previously added document. |
| `TestTruncateCollectionBranchableAdd_RemovesBlocks` | 57-91 | Truncating a branchable collection removes all commit blocks for its documents. |

---

### `dac_test.go`

Tests that DAC (decentralised access control) ownership is preserved correctly across truncation and document re-creation.

| Test Function | Line | Description |
|---|---|---|
| `TestTruncateCollectionDAC_RemovedPrivateDocumentRetainsPermissions` | 50-103 | After truncation, re-adding a private document retains the original owner's permissions. |
| `TestTruncateCollectionDAC_RemovedPublicDocumentRetainsPermissions` | 105-176 | After truncation, a public document re-added with an identity is owned by the new identity. |

---

### `index_filter_test.go`

Tests that truncation clears both standard and unique indexes, ensuring indexed queries return no results and unique constraints are lifted after truncation.

| Test Function | Line | Description |
|---|---|---|
| `TestTruncateCollectionIndexFilter_RemovesDocument` | 21-55 | Truncating a collection removes a document from an indexed field filter result. |
| `TestTruncateCollectionIndexFilter_WithUniqueIndex_RemovesDocument` | 57-91 | Truncating a collection removes a document that was stored in a unique index. |
| `TestTruncateCollectionIndexFilter_WithUniqueIndex_AllowsRecreationOfDocument` | 93-139 | Truncating a collection clears the unique index, allowing the same document to be re-added. |

---

### `parallel_test.go`

Tests that truncation running concurrently with document additions correctly removes all pre-existing documents.

| Test Function | Line | Description |
|---|---|---|
| `TestTruncateCollectionParallel_DeletesAllPreviouslyExistingDocuments` | 21-99 | Truncating concurrently with document adds removes all pre-existing documents. |

---

### `view_test.go`

Tests the interaction between truncation and materialized views, covering both truncating the view itself and truncating the underlying source collection.

| Test Function | Line | Description |
|---|---|---|
| `TestTruncateCollectionViewAdd_RemovesDocument` | 24-104 | Truncating a materialized view removes its documents but allows reconstruction on refresh. |
| `TestTruncateCollectionViewAdd_TruncatingSourceDoesNotTruncateView` | 105-171 | Truncating the source collection does not remove documents from a materialized view. |
