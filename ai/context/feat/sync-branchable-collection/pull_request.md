# Add SyncBranchableCollection API for P2P Collection DAG Sync

## Summary

Implements `SyncBranchableCollection` method to enable on-demand synchronization of branchable collection DAGs from peer nodes. This complements the existing `SyncDocuments` and `FetchCollections` methods, providing a complete P2P sync story.

## Problem

DefraDB lacked a way to sync the collection-level DAG for branchable collections. While `SyncDocuments` syncs individual document DAGs and `FetchCollections` fetches schema definitions, there was no method to sync the collection-level commit history that branchable collections maintain.

This prevented users from catching up on collection history after being offline without setting up full replicators.

## Solution

Added `SyncBranchableCollection(ctx, collectionName)` method following the same pattern as `SyncDocuments`:

- Validates collection exists and is branchable
- Publishes request to `collection-sync` pubsub topic
- Retrieves collection head from responding peer
- Syncs the collection DAG via existing DAG sync infrastructure
- Publishes merge event for local processing

## Changes

### Core Implementation

**P2P Layer** (`internal/db/p2p/`):
- `sync_col.go`: Added sync methods and message handlers
- `sync_col_js.go`: Added stubs for JS/WASM
- `p2p.go`: Registered `collection-sync` pubsub topic
- `errors.go`: Added `ErrTimeoutCollectionSync`

**Database Layer** (`internal/db/`):
- `p2p.go`: Added DB wrapper method
- `txn.go`: Added transaction wrapper method

### API Layers

**HTTP** (`http/`):
- `handler_p2p.go`: Added `SyncBranchableCollection` handler and OpenAPI spec
- `client_p2p.go`: Added HTTP client method
- Endpoint: `POST /p2p/collections/sync-branchable`

**CLI** (`cli/`):
- `p2p_branchable_sync.go`: New command file
- `cli.go`: Registered command
- Usage: `defradb client p2p collection sync-branchable <collection-name> [--timeout=10s]`

**C Bindings** (`cbindings/`):
- `p2p.go`: Added `P2PbranchableCollectionSync` export
- `wrapper.go`: Stubbed pending C header regeneration (TODO)

### Testing

**Test Infrastructure** (`tests/`):
- `action/sync_branchable_collection.go`: New test action
- `integration/net/sync/branchable_collection/simple_test.go`: Integration tests (4 tests)
- `clients/{cli,http,js}/wrapper.go`: Updated test wrappers

**Test Coverage:**
- ✅ Sync branchable collection with single document
- ✅ Sync with multiple documents
- ✅ Error on non-branchable collection
- ✅ Error on non-existent collection

### Interface Updates

**Client Interface** (`client/p2p.go`):
```go
SyncBranchableCollection(ctx context.Context, collectionName string) error
```

**Mocks** (`client/mocks/txn.go`):
- Regenerated via mockery

## Technical Details

### Pubsub Protocol

**Topic:** `collection-sync`

**Request:**
```json
{
  "collectionName": "Users"
}
```

**Reply:**
```json
{
  "collectionName": "Users",
  "head": "<cid-bytes>",
  "sender": "<peer-id>"
}
```

### Collection Head Retrieval

Uses collection headstore with short ID mapping:
```go
shortID := GetShortCollectionID(ctx, collectionID)
key := NewHeadstoreColKey(shortID)
headset := NewHeadSet(txn.Headstore(), key)
cids, _, _ := headset.List(ctx)
return cids[0].Bytes()  // Single head for collections
```

### Context Management

Message handlers require explicit cache initialization:
```go
txnCtx := dbid.InitCollectionShortIDCache(p.ctx)
txnCtx = datastore.CtxSetTxn(txnCtx, txn)
```

## Testing

Run integration tests:
```bash
DEFRA_CLIENT_C=false go test ./tests/integration/net/sync/branchable_collection/... -v
```

**Results:** 4/4 tests passing ✅

Build verification:
```bash
go build ./client/... ./internal/db/... ./http/... ./cli/...
```

## Known Limitations

1. **C Bindings:** Requires building C shared library to generate headers. Currently stubbed with error message.
2. **Schema Prerequisite:** Collection schema must exist on both nodes before syncing DAG.
3. **Single Head Assumption:** Assumes branchable collections have single head (current implementation).

## Migration Notes

No migration needed - this is a new API addition with no breaking changes.

## Related Work

- Follows pattern established by `SyncDocuments` (issue #4155)
- Complements `FetchCollections` (renamed from `SyncCollections` in commit c1ecddd5)
- Part of broader P2P sync capabilities

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
