# Specifications: Implement SyncBranchableCollection

## GitHub Issue

[#4155](https://github.com/sourcenetwork/defradb/issues/4155) - Add `SyncBranchableCollection` API

## Problem Statement

DefraDB currently has `SyncDocuments` for syncing individual document DAGs and `FetchCollections` for fetching collection schema definitions. However, there's no way to sync the collection-level DAG for branchable collections.

**Background:**
- Regular collections: No DAG structure, documents link to collection
- Branchable collections: Have a single DAG for the entire collection that tracks all document changes
- Documents: Have individual DAGs that can be synced via `SyncDocuments`

**User Need:**
Users need to sync the collection-level commit history for branchable collections without setting up full replicators, enabling catch-up scenarios after being offline.

## Requirements

### Functional Requirements

1. Add `SyncBranchableCollection(ctx, collectionName)` method to the P2P interface
2. Method should sync the collection-level DAG from peers over P2P
3. Only works for collections marked with `@branchable` directive
4. Returns error if collection is not branchable
5. Returns error if collection doesn't exist
6. Uses pubsub networking (similar to `SyncDocuments`)
7. Supports context timeouts for the operation
8. Does NOT automatically subscribe to future updates

### Non-Functional Requirements

1. Follow the same pattern as `SyncDocuments` for consistency
2. Support all client interfaces: Go, HTTP, CLI, C bindings, test wrappers
3. Implement message handlers for pubsub request/response
4. Use collection-level headstore for finding collection heads
5. Leverage existing DAG sync infrastructure

## Acceptance Criteria

- [x] `SyncBranchableCollection` method added to `client.P2P` interface
- [x] Implementation in `internal/db/p2p/sync_col.go`
- [x] HTTP endpoint: `POST /p2p/collections/sync-branchable`
- [x] CLI command: `defradb client p2p collection sync-branchable <name>`
- [x] C bindings export: `P2PbranchableCollectionSync`
- [x] Test wrappers updated (CLI, HTTP, JS)
- [x] Integration tests cover:
  - Successful sync of branchable collection
  - Error for non-branchable collections
  - Error for non-existent collections
  - Multiple documents sync correctly
- [x] All tests pass (4/4)

## API Design

### Method Signature
```go
SyncBranchableCollection(ctx context.Context, collectionName string) error
```

### Parameters
- `ctx`: Context with optional timeout
- `collectionName`: Name of the branchable collection to sync

### Returns
- `error`: nil on success, error if collection not found/not branchable/timeout

### Example Usage
```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

err := db.SyncBranchableCollection(ctx, "Users")
if err != nil {
    log.Fatal(err)
}
```

## Success Metrics

- Method successfully syncs collection-level commits from peers
- Error handling works for edge cases
- Integration with existing P2P infrastructure
- No breaking changes to existing APIs
