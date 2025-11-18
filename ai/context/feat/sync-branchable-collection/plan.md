# Implementation Plan: SyncBranchableCollection

## Overview

Implement `SyncBranchableCollection` method to enable on-demand synchronization of branchable collection DAGs from peer nodes, following the same pattern as `SyncDocuments`.

## Architecture

### Key Components

**P2P Layer** (`internal/db/p2p/sync_col.go`):
- `SyncBranchableCollection()` - Main entry point, validates collection is branchable
- `syncBranchableCollection()` - Publishes request to pubsub topic
- `wait AndHandleCollectionSyncResponse()` - Waits for peer response
- `handleCollectionSyncResponse()` - Processes response and extracts collection head CID
- `syncCollectionAndMerge()` - Syncs DAG and publishes merge event
- `syncCollectionDAG()` - Loads and syncs all blocks in the DAG
- `collectionSyncMessageHandler()` - Handles incoming sync requests
- `processCollectionSyncItem()` - Retrieves collection head from local headstore

**Pubsub Protocol**:
- Topic: `collection-sync`
- Request: `{collectionName: string}`
- Reply: `{collectionName: string, head: []byte, sender: string}`

### Data Flow

```
Node1 (Requester)                    Node0 (Responder)
      |                                    |
      | Publish collectionSyncRequest     |
      |-----------------------------------→|
      |                                    | Get collection by name
      |                                    | Check IsBranchable
      |                                    | Get collection head from headstore
      | collectionSyncReply                |
      |←-----------------------------------|
      |                                    |
      | syncCollectionDAG(headCID)         |
      | - Load block via IPLD              |
      | - Recursively sync linked blocks   |
      |                                    |
      | Publish event.Merge                |
      | - DB.Merge() processes             |
```

## Implementation Steps

### Phase 1: Core P2P Implementation ✅

**Files Modified:**
- `client/p2p.go` - Add interface method
- `internal/db/p2p/sync_col.go` - Add implementation
- `internal/db/p2p/sync_col_js.go` - Add stub for JS
- `internal/db/p2p/p2p.go` - Register pubsub topic handler
- `internal/db/p2p/errors.go` - Add `ErrTimeoutCollectionSync`

**Implementation:**
- Created request/reply structs for pubsub communication
- Implemented sync logic following `SyncDocuments` pattern
- Added message handler for responding to sync requests
- Used collection headstore (`HeadstoreColKey`) to find collection heads
- Context management for ID cache initialization

### Phase 2: Database Layer ✅

**Files Modified:**
- `internal/db/p2p.go` - Add DB wrapper method
- `internal/db/txn.go` - Add Txn wrapper method

**Implementation:**
- Added wrapper methods that delegate to P2P layer
- No transaction support (same as `SyncDocuments`)
- Proper error handling for missing P2P

### Phase 3: HTTP API ✅

**Files Modified:**
- `http/handler_p2p.go` - Add handler and OpenAPI spec
- `http/client_p2p.go` - Add HTTP client method

**Endpoint:**
- `POST /p2p/collections/sync-branchable`
- Body: `{collectionName: string, timeout?: string}`

### Phase 4: CLI ✅

**Files Modified:**
- `cli/p2p_branchable_sync.go` - New command file
- `cli/cli.go` - Register command

**Command:**
```bash
defradb client p2p collection sync-branchable Users --timeout=10s
```

### Phase 5: C Bindings ✅

**Files Modified:**
- `cbindings/p2p.go` - Add `P2PbranchableCollectionSync` export
- `cbindings/wrapper.go` - Add wrapper (stubbed pending header regeneration)

**Note:** C bindings require building the shared library to generate headers. Temporarily stubbed to allow other clients to build.

### Phase 6: Test Infrastructure ✅

**Files Created:**
- `tests/action/sync_branchable_collection.go` - Test action
- `tests/integration/net/sync/branchable_collection/simple_test.go` - Integration tests

**Files Modified:**
- `tests/clients/cli/wrapper.go` - Add CLI wrapper
- `tests/clients/http/wrapper.go` - Add HTTP wrapper
- `tests/clients/js/wrapper.go` - Add stub
- `client/mocks/txn.go` - Regenerated via mockery

**Tests:**
1. Sync simple branchable collection (1 doc, 3 commits)
2. Sync multiple documents (2 docs, 8 commits)
3. Error on non-branchable collection
4. Error on non-existent collection

## Technical Decisions

### 1. Pubsub vs. Replicator

**Choice:** Pubsub (request/response pattern)

**Rationale:**
- Matches `SyncDocuments` pattern
- On-demand, not continuous
- Doesn't require persistent replicator setup
- Simpler for user catch-up scenarios

### 2. Collection Head Retrieval

**Choice:** Use `HeadstoreColKey` with collection short ID

**Rationale:**
- Collections have single head (unlike documents with multiple field heads)
- Headstore already tracks collection-level commits for branchable collections
- Short ID lookup requires proper context initialization with ID cache

###  3. Context Management

**Challenge:** `GetShortCollectionID` needs context with ID cache initialized

**Solution:**
```go
txnCtx := dbid.InitCollectionShortIDCache(p.ctx)
txnCtx = datastore.CtxSetTxn(txnCtx, txn)
shortID, err := dbid.GetShortCollectionID(txnCtx, col.CollectionID)
```

### 4. Error Handling

Errors returned for:
- Collection not found
- Collection not branchable
- Timeout waiting for peer response
- No collection heads found
- Network/sync errors

## Affected Components

### Core Implementation
- P2P sync layer (`internal/db/p2p/`)
- Database layer (`internal/db/`)
- Client interface (`client/`)

### API Layers
- HTTP handlers and client
- CLI commands
- C bindings (pending header regeneration)

### Testing
- Test actions
- Integration test suite
- Test client wrappers

## Testing Strategy

### Unit-Level
- Error cases: non-branchable, non-existent collections
- Timeout handling via context

### Integration-Level
- Single document branchable collection sync
- Multiple documents sync
- P2P communication between nodes
- Commit verification on receiving node

### Test Execution
```bash
DEFRA_CLIENT_C=false go test ./tests/integration/net/sync/branchable_collection/... -v
```

**Results:** 4/4 tests passing ✅

## Notes

- C bindings implementation is complete but requires building the C shared library to generate headers
- Temporarily stubbed in `cbindings/wrapper.go` to prevent build failures
- Full C support can be enabled after running `make build-c-shared-linux` or similar
