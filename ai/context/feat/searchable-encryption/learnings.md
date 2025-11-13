# Learnings from Implementation

## 1. DefraDB Uses Datastore Keys for Everything
**Discovery**: The key system in DefraDB is more sophisticated than typical key-value patterns
- All keys implement the `keys.DataStoreKey` interface
- Keys have composite structure with prefixes and suffixes
- The `peerstore.go` file acts as a central registry for key prefixes
- Keys support serialization/deserialization through CBOR

## 2. Event System is Central to Distributed Operations
**Discovery**: DefraDB has a comprehensive event system for coordination
- Events are typed with specific event codes
- The `event.Peer` struct carries peer-specific information
- Events can have success/error channels for async communication
- Retry mechanisms are triggered by event failures

## 3. Two-Level Retry Structure Pattern
**Discovery**: Document replication uses a clever two-level structure
- Main retry entry per peer (stores retry metadata like attempts, next retry time)
- Individual items under each peer (stores actual data to retry)
- This allows batch operations and efficient cleanup
- Pattern wasn't obvious until analyzing the full implementation

## 4. Identity Reconstruction for SE
**Discovery**: SE identity handling is complex
- Identities are decomposed into PublicKey and KeyType for storage
- The `reconstructIdentity` method rebuilds the full identity object
- This is necessary because the identity interface can't be directly serialized
- KeyType determines which concrete type to instantiate

## 5. CBOR Marshaling Throughout
**Discovery**: DefraDB consistently uses CBOR for data persistence
- More efficient than JSON for binary data
- Used for all retry data structures
- The `cbor` package is used directly, not through an abstraction
- Important to handle marshaling errors properly

## 6. Replicator Lifecycle Management
**Discovery**: Replicators have specific lifecycle patterns
- They can be stopped but retry entries remain
- The existence check (`hasReplicator`) prevents processing orphaned retries
- This prevents unnecessary work when replicators are removed
- Pattern wasn't documented but is critical for correctness

## 7. Logger Pattern in DefraDB
**Discovery**: DefraDB has specific logger usage patterns
- Loggers are typically passed as parameters, not stored in configs
- The `corelog` package provides structured logging
- Context is always the first parameter for log methods
- Error logging uses specific methods like `ErrorContextE`