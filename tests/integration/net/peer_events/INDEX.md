# Index: `tests/integration/net/peer_events`

## Overview

This folder contains integration tests for the P2P peer events system in DefraDB. The tests verify that nodes correctly emit and receive join and left peer events across three topic types: the global doc-sync topic (emitted on any peer connection), per-collection topics (emitted when nodes subscribe to the same collection), and per-document topics (emitted when nodes subscribe to the same document).

## Test Index

### `simple_test.go`

Tests for join peer events on the global doc-sync topic when peers connect directly.

| Test Function | Line | Description |
|---|---|---|
| `TestPeerEvents_OnConnect_ShouldReceiveJoinEventOnDocSyncTopic` | 23-43 | Connecting two peers emits a join peer event on the doc-sync topic. |
| `TestPeerEvents_OnConnectMultiplePeers_ShouldReceiveAllJoinEvents` | 45-70 | Connecting multiple peers emits join peer events for all connected nodes. |
| `TestPeerEvents_OnConnectBidirectional_BothNodesShouldReceiveJoinEvents` | 72-98 | Connecting two peers causes both nodes to receive a join event on the doc-sync topic. |

### `collection_test.go`

Tests for join and left peer events on per-collection topics when nodes subscribe or unsubscribe.

| Test Function | Line | Description |
|---|---|---|
| `TestPeerEvents_OnSubscribeToCollection_ShouldReceiveJoinEventOnCollectionTopic` | 22-57 | Subscribing to a collection topic emits a join peer event for that collection. |
| `TestPeerEvents_OnSubscribeToMultipleCollections_ShouldReceiveJoinEventsOnAllTopics` | 59-98 | Subscribing to multiple collection topics emits join peer events for each collection. |
| `TestPeerEvents_MultipleNodesSubscribedToCollection_ShouldReceiveAllJoinEvents` | 100-144 | A node subscribed to a collection receives join events from all other subscribing nodes. |
| `TestPeerEvents_OnUnsubscribeFromCollection_ShouldReceiveLeftEvent` | 146-192 | Unsubscribing from a collection topic emits a left peer event for that collection. |
| `TestPeerEvents_OnUnsubscribeFromMultipleCollections_ShouldReceiveLeftEvents` | 194-245 | Unsubscribing from multiple collection topics emits a left peer event for each. |

### `document_test.go`

Tests for join and left peer events on per-document topics, including interactions with collection and doc-sync topics.

| Test Function | Line | Description |
|---|---|---|
| `TestPeerEvents_OnSubscribeToDocument_ShouldReceiveJoinEventOnDocumentTopic` | 23-63 | Subscribing to a document topic emits a join peer event for that document. |
| `TestPeerEvents_OnSubscribeToMultipleDocuments_ShouldReceiveJoinEventsOnAllTopics` | 65-117 | Subscribing to multiple document topics emits join peer events for each document. |
| `TestPeerEvents_DocumentAndDocSyncTopics_ShouldReceiveJoinEventsOnBoth` | 119-162 | Subscribing to a document topic also receives a join event on the doc-sync topic. |
| `TestPeerEvents_AllTopicTypes_ShouldReceiveJoinEventsOnAll` | 164-218 | Subscribing to collection, document, and doc-sync topics all emit join peer events. |
| `TestPeerEvents_OnUnsubscribeFromDocument_ShouldReceiveLeftEvent` | 220-271 | Unsubscribing from a document topic emits a left peer event for that document. |
| `TestPeerEvents_OnUnsubscribeFromMultipleDocuments_ShouldReceiveLeftEvents` | 273-340 | Unsubscribing from multiple document topics emits a left peer event for each. |
