# Index: `tests/integration/net/simple/peer/subscribe/document`

## Overview

This folder tests the P2P document-level subscription API, which lets a peer node subscribe to specific documents by ID so it receives updates (not just new documents) from other peers. It covers adding subscriptions (including error cases with malformed IDs), removing subscriptions (including partial and erroneous removals), and listing the current document subscriptions for a peer.

## Test Index

### `with_add_test.go`

Tests that subscribing to a specific document correctly enables (or, on error, prevents) update sync for that document only.

| Test Function | Line | Description |
|---|---|---|
| `TestP2PDocument_AddSingle_ShouldSync` | 24-106 | Subscribing to a specific document syncs only that document's updates to the peer. |
| `TestP2PDocument_AddSingleErroneousDocID_ShouldNotSync` | 108-159 | Subscribing with a malformed document ID returns an error and no sync occurs. |

### `with_add_get_test.go`

Tests that adding document subscriptions is correctly reflected when listing subscriptions for a peer.

| Test Function | Line | Description |
|---|---|---|
| `TestP2PDocumentAddGetSingle` | 22-60 | Subscribing to a single document lists it in the peer's P2P document subscriptions. |
| `TestP2PDocumentAddGetMultiple` | 62-114 | Subscribing to multiple documents lists all of them in the peer's P2P document subscriptions. |

### `with_add_get_remove_test.go`

Tests that document subscriptions are accurately reflected in the subscription list after add and remove operations.

| Test Function | Line | Description |
|---|---|---|
| `TestP2PDocumentAddRemoveGetSingle` | 22-64 | Subscribing then unsubscribing from a single document leaves the subscription list empty. |
| `TestP2PDocumentAddRemoveGetMultiple` | 66-116 | Unsubscribing from one of two document subscriptions leaves only the other in the list. |

### `with_add_remove_test.go`

Tests that removing document subscriptions stops update sync, including edge cases with malformed IDs and empty removal lists.

| Test Function | Line | Description |
|---|---|---|
| `TestP2PDocumentAddAndRemoveSingle` | 24-85 | Unsubscribing from a document stops its updates from syncing to the peer. |
| `TestP2PDocumentAddAndRemoveMultiple` | 87-185 | Unsubscribing from one document stops its sync while the remaining subscription still syncs. |
| `TestP2PDocumentAddSingleAndRemoveErroneous` | 187-252 | A failed unsubscribe with a malformed document ID does not remove existing subscriptions. |
| `TestP2PDocumentAddSingleAndRemoveNone` | 254-315 | Removing an empty document ID list leaves active subscriptions and sync unchanged. |

### `with_get_test.go`

Tests that listing document subscriptions on a freshly connected peer with no subscriptions returns an empty result.

| Test Function | Line | Description |
|---|---|---|
| `TestP2PDocument_GetAllWithNoneConfigured_ShouldSucceed` | 21-43 | Listing P2P document subscriptions on a peer with none configured returns an empty list. |
