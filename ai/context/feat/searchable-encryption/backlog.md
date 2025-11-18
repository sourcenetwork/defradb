# Backlog

## Phase 3 Implementation - COMPLETED ✅

**Summary**: Successfully implemented the P2P replication protocol for SE artifacts following the existing document replication patterns. The system now has a complete pipeline from SE artifact creation to network transmission and storage on remote nodes.

### Major Refactoring Completed:
1. ✅ Updated key generation to use string builder for efficiency
2. ✅ Moved SE retry keys from datastore to peerstore (created peerstore_se.go)
3. ✅ Removed Tag field and unified on SearchTag in Artifact struct
4. ✅ Removed EncValue field - only store keys for search lookups
5. ✅ Removed CollectionID from MergeEvent (artifacts already contain it)
6. ✅ Used constants from doc replicator (retryLoopInterval, retryTimeout)
7. ✅ Added error checking for Delete operations
8. ✅ Updated to NOT store SE artifacts locally (only on replicator nodes)
9. ✅ Enhanced DeleteSEArtifacts documentation with parameter details
10. ✅ Used FetchKeysForPrefix for efficient key iteration

## Phase 4 Implementation Update - IN PROGRESS 🚧

### Completed:
1. ✅ Modified GraphQL schema generation to return docID-only results
   - Created `{CollectionName}EncryptedResult` type with only `_docID` field
   - Updated `GenerateEncryptedQueryInputForGQLType` to use the new result type

2. ✅ Simplified seScanNode implementation
   - Removed document fetching logic
   - Modified to return only docIDs
   - Removed inner scan node dependency
   - Moved query execution to Next() for lazy evaluation

### Remaining Phase 4 Tasks:
1. [ ] Resolve import cycle between `internal/core` and `internal/se`
2. [ ] Complete integration testing
3. [ ] Update documentation

## Phase 3 Implementation Progress

### Completed:
1. ✅ Created SE event system (internal/se/events.go)
   - UpdateEvent, MergeEvent, ReplicationFailureEvent

2. ✅ Created storage keys for SE artifacts (internal/keys/datastore_se.go)
   - DatastoreSE for artifact storage
   - SERetryKey for retry mechanism

3. ✅ Created ReplicationCoordinator (internal/se/replication_coordinator.go)
   - Handles merge events and stores artifacts
   - Implements retry mechanism with exponential backoff
   - Supports partial deletion for field updates

4. ✅ Updated SE context to publish update events
   - Modified registerReplicationCallback to publish UpdateEvent

5. ✅ Added GRPC protocol support (net/grpc.go)
   - pushSEArtifactsRequest/Reply messages
   - Handler registration

6. ✅ Updated network layer
   - Added SE event handling in peer.go
   - Added pushSEArtifacts client method
   - Added pushSEArtifactsHandler server method
   - Added error types

7. ✅ Updated DB to initialize SE coordinator
   - Added seCoordinator field
   - Added missing interface methods
   - Initialize coordinator when SE key is present

### Completed (Phase 3):
✅ 8. Close SE coordinator on DB close - added cleanup in DB.Close() method

### TODO (Next Phase):
1. Add artifact generation in Phase 2 implementation
2. Create integration tests for SE replication
3. Add benchmarks for performance testing
4. Implement access control checks in pushSEArtifactsHandler

## Phase 3 Implementation Notes

### 1. Event System Enhancement
- Consider creating a unified event handler registration system instead of hardcoding switch cases in handleMessageLoop
- This would allow SE and other modules to register their own event handlers
- Location: net/peer.go handleMessageLoop()

### 2. Error Types
- Need to create NewErrPushSEArtifacts and related error types in net/errors.go
- Should follow the existing error pattern for consistency

### 3. Integration Testing
- Need comprehensive integration tests for SE artifact replication
- Should test network failures, retries, and partial updates
- Consider adding benchmarks for large artifact sets

### 4. Security Considerations
- Verify peer permissions before accepting SE artifacts
- Add rate limiting for SE artifact push requests
- Consider adding artifact size limits

### 5. Future Optimizations
- Batch SE artifact updates for better performance
- Add compression for SE artifact network transmission
- Consider using a more efficient serialization format than CBOR for network messages

## Phase 3 - Post-Implementation TODO

### 1. Complete regenerateArtifactsForRetry Implementation (PARTIALLY COMPLETED ✅)
- **Description**: The `regenerateArtifactsForRetry` method in `internal/se/replication_coordinator.go` has been implemented
- **Completed**:
  - Created `GenerateArtifactFromBlock` helper function in `internal/se/block.go`
  - Created `regenerateSEArtifacts` method in `internal/db/se_regeneration.go`
  - Updated ReplicationCoordinator to use ArtifactRegenerator callback
  - Updated ProcessBlock to use the helper function
- **Blocked by**: Import cycle between `internal/core` and `internal/se`
- **TODO**: 
  - Resolve import cycle (core imports se, se imports core)
  - The issue is that `internal/core/store.go` calls `se.ProcessBlock`
  - Consider moving SE processing to a different layer or using interfaces

### 2. Add Field Names Collection in pushSEArtifacts (COMPLETED ✅)
- Collect unique field names from failed artifacts
- Include field names in ReplicationFailureEvent
- Used for targeted artifact regeneration during retry

### 3. Implement Key Interface for SE Keys (COMPLETED ✅)
- DatastoreSE now implements Key interface with ToDS() method
- PeerstoreSERetry now implements Key interface with ToString(), Bytes(), and ToDS() methods
- Updated Bytes() to use piece-by-piece assembly via ToString()

### 4. Optimize Retry Mechanism Further
- Consider implementing circuit breaker pattern
- Add jitter to retry intervals to avoid thundering herd
- Implement proper coordination with Phase 2 for artifact regeneration

### 5. Add Integration Tests for Retry Logic
- Test retry with field name regeneration
- Test exponential backoff behavior
- Test cleanup of retry entries

### 6. Document Retry Architecture
- Document how retry regenerates artifacts from document fields
- Document the flow from failure to regeneration to retry
- Add sequence diagram for retry process

## Import Cycle Issue (High Priority)

### Description
There's an import cycle between `internal/core` and `internal/se`:
- `internal/core/store.go` imports `internal/se` to call `se.ProcessBlock`
- `internal/se/block.go` imports `internal/core` for block types and constants

### Potential Solutions
1. **Use Interfaces**: Define an interface in core that se implements
2. **Move SE Processing**: Move the SE processing call to a higher layer (e.g., db package)
3. **Event-Based**: Use events to trigger SE processing instead of direct calls
4. **Separate Package**: Create a separate package for SE block processing that both can import

### Current Impact
- Cannot compile the code due to import cycle
- Blocks testing of the regeneration functionality

## Phase 6 - DocID-Only Results (COMPLETED ✅)

### Summary
Modified the searchable encryption query implementation to return only document IDs instead of full documents. This leverages the new `SyncDocuments` method for document retrieval.

### Changes Made:
1. **GraphQL Schema Generation**:
   - Created a single shared `EncryptedSearchResult` type for all collections
   - Type contains `docIDs: [String!]!` field that returns an array of document IDs
   - Registered the type in the default schema (`internal/request/graphql/schema/types/encrypted_search.go`)

2. **seScanNode Implementation**:
   - Removed document fetching logic and inner scan node dependency
   - Simplified to return all docIDs at once in a single result
   - Removed currentIndex tracking - uses hasReturned flag instead
   - Returns a document with docIDs array field

### Benefits:
- Simpler implementation with shared result type
- Better separation of concerns
- Allows users to selectively sync documents
- Aligns with DefraDB's pattern of specialized queries
- More efficient - returns all results at once instead of iterating