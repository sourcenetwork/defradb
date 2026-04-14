# Index: `tests/integration/net/simple/peer`

## Overview

This folder tests bidirectional P2P peer synchronisation between two or more DefraDB nodes. It covers document creation, update, and deletion propagation via document-level and collection-level subscriptions, including edge cases such as schema version mismatches (older/newer/updated schemas), chained multi-node topologies, concurrent operations, pre-connection history, and node restarts.

## Test Index

### `with_add_test.go`

Tests that document creation via peer sync respects subscription boundaries and propagates correctly through collection subscriptions and node chains.

| Test Function | Line | Description |
|---|---|---|
| `TestP2PAddDoesNotSync` | 23-95 | Adding a document via peer sync does not propagate new documents to peer nodes. |
| `TestP2PAddWithP2PCollection` | 99-203 | Collection subscriber receives new documents; non-subscriber does not. |
| `TestP2PAdd_WithP2PCollectionWithNodeChain_ShouldSucceed` | 205-283 | New document added to the head of a five-node chain reaches all subscribed nodes. |
| `TestP2PAdd_WithP2PCollectionOnLastNodeInNodeChain_ShouldPropagateUpdate` | 285-349 | Only the last node in a five-node chain subscribes and still receives the new document. |
| `TestP2PAdd_WithP2PCollectionAndSubscription_ShouldSucceed` | 351-402 | GraphQL subscription on a collection subscriber receives synced documents from a peer. |

### `with_add_add_field_test.go`

Tests that documents with newly added schema fields sync correctly across peers running older, newer, or matching collection versions.

| Test Function | Line | Description |
|---|---|---|
| `TestP2PPeerAddWithNewFieldSyncsDocsToOlderCollectionVersion` | 23-99 | Document with a new field syncs to a peer node running the older schema version. |
| `TestP2PPeerAddWithNewFieldSyncsDocsToNewerCollectionVersion` | 101-159 | Document without a new field syncs to a peer node running the newer schema version. |
| `TestP2PPeerAddWithNewFieldSyncsDocsToUpdatedCollectionVersion` | 161-218 | Document with a new field syncs correctly when both nodes share the updated schema. |
| `TestP2PPeerAddWithNewFieldDocSyncedBeforeReceivingNodeSchemaUpdatedDoesNotReturnNewField` | 222-313 | Document synced before the receiving node's schema update does not return the new field. |

### `with_delete_test.go`

Tests that document deletions propagate correctly across peers, including concurrent delete-and-update races, multi-document scenarios, and pre-connection history.

| Test Function | Line | Description |
|---|---|---|
| `TestP2PWithSingleDocumentConcurrentDeleteAndUpdate` | 24-91 | Concurrent delete and update on different peers results in a deleted document on both nodes. |
| `TestP2PWithMultipleDocumentsSingleDelete` | 95-160 | Deleting one subscribed document syncs the deletion while the other document remains visible. |
| `TestP2PWithMultipleDocumentsSingleDeleteWithShowDeleted` | 162-233 | Synced deletion is visible in showDeleted query results alongside the remaining live document. |
| `TestP2PWithMultipleDocumentsWithSingleUpdateBeforeConnectSingleDeleteWithShowDeleted` | 235-314 | Pre-connect update followed by deletion syncs the final deleted state to the subscriber. |
| `TestP2PWithMultipleDocumentsWithMultipleUpdatesBeforeConnectSingleDeleteWithShowDeleted` | 316-403 | Multiple pre-connect updates followed by deletion syncs the final deleted state to the peer. |
| `TestP2PWithMultipleDocumentsWithUpdateAndDeleteBeforeConnectSingleDeleteWithShowDeleted` | 405-528 | Pre-connection delete on one peer does not override a post-connection update on the other. |

### `with_update_test.go`

Tests that document updates propagate bidirectionally between nodes via document and collection subscriptions, including concurrent updates and unmapped node isolation.

| Test Function | Line | Description |
|---|---|---|
| `TestP2PWithSingleDocumentSingleUpdateFromChild` | 26-83 | Update from the source node syncs to the subscriber node via document subscription. |
| `TestP2PWithSingleDocumentSingleUpdateFromParent` | 87-144 | Update from the target node syncs back to the source node via document subscription. |
| `TestP2PWithSingleDocumentUpdatePerNode` | 147-217 | Concurrent updates from both nodes converge to one of the two values after sync. |
| `TestP2PWithSingleDocumentSingleUpdateDoesNotSyncToNonPeerNode` | 219-310 | Update syncs to the subscribed peer node but not to an unconnected third node. |
| `TestP2PWithSingleDocumentSingleUpdateDoesNotSyncFromUnmappedNode` | 312-399 | Update from an unmapped node does not propagate to connected peer nodes. |
| `TestP2PWithMultipleDocumentUpdatesPerNode` | 402-494 | Multiple updates from both nodes converge to one of the two latest values after sync. |
| `TestP2PWithSingleDocumentSingleUpdateFromChildWithP2PCollection` | 498-563 | New document and its subsequent update both reach the collection-subscribed peer node. |
| `TestP2PWithMultipleDocumentUpdatesPerNodeWithP2PCollection` | 568-682 | Multiple updates per node plus a new document all reach the collection-subscribed peer. |

### `with_update_add_field_test.go`

Tests that document updates referencing newly added schema fields sync correctly to peers with older collection versions.

| Test Function | Line | Description |
|---|---|---|
| `TestP2PPeerUpdateWithNewFieldSyncsDocsToOlderCollectionVersionMultistep` | 23-111 | Second update to an existing field still syncs even if the first update targeted a new field. |
| `TestP2PPeerUpdateWithNewFieldSyncsDocsToOlderCollectionVersion` | 113-193 | Update including a new field syncs the known fields to a peer with the older schema. |

### `with_update_restart_test.go`

Tests that peer document subscription state is restored correctly after a full node restart.

| Test Function | Line | Description |
|---|---|---|
| `TestP2PWithSingleDocumentSingleUpdateFromChildAndRestart` | 24-82 | After a full node restart, updates from the source peer still sync to the subscriber. |

## Subdirectories

| Directory | Summary |
|---|---|
| [`crdt/`](crdt/INDEX.md) | Tests CRDT field convergence (LWW, PCounter, PNCounter) across peer nodes after concurrent updates. |
| [`subscribe/collection/`](subscribe/collection/INDEX.md) | Tests the collection-level P2P subscription API including add, remove, and list operations. |
| [`subscribe/document/`](subscribe/document/INDEX.md) | Tests the document-level P2P subscription API including add, remove, and list operations. |
