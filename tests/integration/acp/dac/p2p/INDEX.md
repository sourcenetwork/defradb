# Index: `tests/integration/acp/dac/p2p`

## Overview

This folder contains integration tests that verify P2P replication and subscription behaviour under Document Access Control (DAC) policies. The tests confirm that private documents remain hidden from unauthorized actors across all nodes, that access granted or revoked via DAC actor relationships is honoured consistently regardless of which node handles the operation, and that both the replicator and pub-sub subscription mechanisms interact correctly with SourceHub and local ACP backends.

## Test Index

### `add_test.go`

Tests that private documents can be added on independent nodes and that a subscribing peer only syncs a private document once the owner grants it a reader relation.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_P2PAddPrivateDocumentsOnDifferentNodes_SourceHubACP` | 24-117 | P2P replication allows owner to add private documents independently on different nodes. |
| `TestACP_P2PAddPrivateDocumentAndSyncAfterAddingRelationship_SourceHubACP` | 119-245 | Peer syncs private document to subscribing node only after a reader relation is granted. |

### `delete_test.go`

Tests that a document owner can delete private documents on separate replicator-connected nodes.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_P2PDeletePrivateDocumentsOnDifferentNodes_SourceHubACP` | 24-144 | Owner can delete private documents on separate nodes after replication via a replicator. |

### `replicator_test.go`

Tests that a one-to-one replicator can be configured with a DAC-permissioned collection and that access control is enforced for all actors after sync.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_P2POneToOneReplicatorWithPermissionedCollection_LocalACP` | 24-77 | A one-to-one replicator can be configured with a DAC-permissioned collection using local ACP. |
| `TestACP_P2POneToOneReplicatorWithPermissionedCollection_SourceHubACP` | 79-185 | Replicator syncs private doc and DAC policy hides it from unauthorized actors on all nodes. |

### `replicator_with_doc_actor_relationship_test.go`

Tests that DAC actor relationships can be managed through any node in a replicator setup, with access changes reflected globally.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_P2PReplicatorWithPermissionedCollectionAddDocActorRelationship_SourceHubACP` | 24-274 | Replicator respects DAC: granting and revoking reader relation controls access across all nodes. |

### `subscribe_test.go`

Tests that a subscribing peer can register against a DAC-permissioned collection and that private documents are inaccessible without an explicit actor relation.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_P2PSubscribeAddGetSingleWithPermissionedCollection_LocalACP` | 24-93 | A collection subscription can be registered on a DAC-permissioned collection using local ACP. |
| `TestACP_P2PSubscribeAddGetSingleWithPermissionedCollection_SourceHubACP` | 95-222 | Subscribed peer cannot access a private doc without a DAC relation; unauthorized actors always see nothing. |

### `subscribe_with_doc_actor_relationship_test.go`

Tests the full lifecycle of granting and revoking a DAC reader relation for a subscribing peer, verifying that sync and visibility change accordingly.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_P2PSubscribeAddGetSingleWithPermissionedCollectionAddDocActorRelationship_SourceHubACP` | 24-252 | Subscription peer syncs private doc after node identity granted reader relation; revocation hides it again. |

### `update_test.go`

Tests that a document owner can update private documents on separate replicator-connected nodes after sync.

| Test Function | Line | Description |
|---|---|---|
| `TestACP_P2PUpdatePrivateDocumentsOnDifferentNodes_SourceHubACP` | 24-156 | Owner can update private documents on separate replicator-connected nodes after sync. |
