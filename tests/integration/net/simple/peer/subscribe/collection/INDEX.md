# Index: `tests/integration/net/simple/peer/subscribe/collection`

## Overview

This folder tests the P2P collection-level subscription API, which lets a peer node subscribe to an entire collection so it receives new documents created by other peers. It covers adding subscriptions (including error cases), removing subscriptions (including partial and erroneous removals), and listing the current subscriptions for a peer.

## Test Index

### `with_add_test.go`

Tests that adding collection subscriptions correctly enables (or, on error, prevents) document sync for the subscribed collections.

| Test Function | Line | Description |
|---|---|---|
| `TestP2PCollectionAddSingle` | 25-99 | Subscribing to a collection syncs new docs to the subscriber but not the non-subscriber. |
| `TestP2PCollectionAddMultiple` | 101-196 | Subscribing to two of three collections only syncs documents from the subscribed collections. |
| `TestP2PCollectionAddSingleErroneousCollectionID` | 198-243 | Subscribing with a non-existent collection ID returns an error and no sync occurs. |
| `TestP2PCollectionAddValidAndErroneousCollectionID` | 245-291 | A batch subscribe with a mix of valid and invalid IDs errors and rolls back all subscriptions. |
| `TestP2PCollectionAddValidThenErroneousCollectionID` | 293-346 | A failed subscribe call does not affect a previously successful collection subscription. |
| `TestP2PCollectionAddNone` | 348-391 | Subscribing to an empty collection ID list causes no sync. |

### `with_add_get_test.go`

Tests that adding collection subscriptions is correctly reflected when listing subscriptions for a peer.

| Test Function | Line | Description |
|---|---|---|
| `TestP2PCollectionAddGetSingle` | 21-50 | Subscribing to a single collection lists it in the peer's P2P collection subscriptions. |
| `TestP2PCollectionAddGetMultiple` | 52-89 | Subscribing to multiple collections lists all subscribed collections for the peer. |

### `with_add_get_remove_test.go`

Tests that collection subscriptions are accurately reflected in the subscription list after add and remove operations.

| Test Function | Line | Description |
|---|---|---|
| `TestP2PCollectionAddRemoveGetSingle` | 21-54 | Subscribing then unsubscribing from a collection leaves the subscription list empty. |
| `TestP2PCollectionAddRemoveGetMultiple` | 56-93 | Unsubscribing from one of two collections leaves only the other collection in the list. |

### `with_add_remove_test.go`

Tests that removing collection subscriptions stops document sync, including edge cases with erroneous IDs and empty removal lists.

| Test Function | Line | Description |
|---|---|---|
| `TestP2PCollectionAddAndRemoveSingle` | 23-81 | Unsubscribing from a collection stops new documents from syncing to the peer. |
| `TestP2PCollectionAddAndRemoveMultiple` | 83-158 | Unsubscribing from one collection stops its docs from syncing while the other still syncs. |
| `TestP2PCollectionAddSingleAndRemoveErroneous` | 160-213 | A failed unsubscribe with a non-existent collection ID does not remove existing subscriptions. |
| `TestP2PCollectionAddSingleAndRemoveNone` | 215-266 | Removing an empty collection ID list leaves active subscriptions and sync unchanged. |

### `with_get_test.go`

Tests that listing collection subscriptions on a freshly connected peer with no subscriptions returns an empty result.

| Test Function | Line | Description |
|---|---|---|
| `TestP2PCollectionGetAll` | 20-42 | Listing P2P collection subscriptions on a peer with none configured returns an empty list. |
