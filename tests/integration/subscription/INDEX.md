# Index: `tests/integration/subscription`

## Overview

This folder contains integration tests for DefraDB's GraphQL subscription feature. Tests verify that subscription requests correctly emit events for document add, update, and delete mutations, that filters (field-value and docID) properly limit which events reach a subscriber, and that commit-level subscriptions return valid commit data including links and heads. Additional tests cover edge cases such as counter CRDT subscriptions, closing a node while a subscription is open, and the `showDeleted` flag.

## Test Index

### `subscription_test.go`

Tests that subscription requests emit the correct events for various mutation types, filters, and lifecycle scenarios.

| Test Function | Line | Description |
|---|---|---|
| `TestSubscriptionWithAddMutations` | 22-87 | Subscription receives one event per add mutation for each new document. |
| `TestSubscriptionWithFilterAndOneAddMutation` | 89-131 | Subscription with age filter receives only the add mutation that satisfies the filter. |
| `TestSubscriptionWithFilterAndOneAddMutationOutsideFilter` | 133-165 | Subscription with age filter receives no events when the mutation falls outside the filter. |
| `TestSubscriptionWithFilterAndAddMutations` | 167-223 | Subscription with age filter receives only the add mutation that satisfies the filter out of two. |
| `TestSubscriptionWithUpdateMutations` | 225-285 | Subscription receives one event for a filtered update mutation targeting a specific document. |
| `TestSubscriptionWithUpdateAllMutations` | 287-359 | Subscription receives one event per document when an update-all mutation is applied. |
| `TestSubscription_WithDocIDFilter_ShouldOnlyGetUpdatesForThatDocID` | 361-419 | Subscription filtered by docID only receives events for that specific document. |
| `TestSubscription_WithClose_WontBlock` | 421-439 | Closing the node while a subscription is open does not cause a deadlock. |
| `TestSubscription_WithCounterCRDT_ShouldSucceed` | 441-493 | Subscription receives incremental counter CRDT values after each update event. |
| `TestSubscription_WithDeleteOperation_ShouldSucceed` | 495-559 | Subscription with showDeleted receives add, update, and delete events in order. |

### `with_commit_test.go`

Tests that commit-level subscriptions (`_commits`) return correct commit data and respect docID filters.

| Test Function | Line | Description |
|---|---|---|
| `TestCommitSubscription_WithAddMutations_ReturnCommits` | 21-70 | Commit subscription returns one commit event per add mutation. |
| `TestCommitSubscription_WithCommitLinksAddMutations_ValidLinks` | 72-177 | Commit subscription returns commits with links and heads matching the mutation response. |
| `TestCommitSubscription_WithDocFilterAndMultipleMutations_FilteredDoc` | 179-266 | Commit subscription filtered by docID only receives commits for that specific document. |
