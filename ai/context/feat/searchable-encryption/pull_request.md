# Refactor: Implement Generic Retry Coordinator for SE and Document Replication

This PR introduces a generic retry coordinator that unifies the retry mechanisms for both Searchable Encryption (SE) and document replication in DefraDB. The implementation follows the existing document replication pattern and extends it to support SE replication with proper type safety.

## Changes

### Core Implementation

Created a generic retry coordinator in `internal/retry/coordinator.go` that supports different data types through Go generics. This allows SE to store `SERetryInfo` structs while document replication uses empty `struct{}` for minimal overhead.

### Key System Updates

Added new constants to `internal/keys/peerstore.go` for SE retry functionality:
- `SE_RETRY_ID`: Main retry entry key prefix for SE
- `SE_RETRY_ITEM`: Individual retry item key prefix for SE

### SE Retry Implementation

Implemented SE-specific retry handler in `internal/se/retry.go` that:
- Stores comprehensive retry information including CollectionID, DocID, FieldNames, PublicKey, and KeyType
- Handles identity reconstruction from stored data
- Integrates with the existing SE replication coordinator

### Document Replication Updates

Updated the document replication system in `net/p2p_replicator.go` to:
- Use the new generic coordinator
- Remove deprecated retry methods
- Simplify the retry failure handling logic

### Common Logic Extraction

Extracted common SE replication logic into `publishSEReplication` method in `internal/se/replication_coordinator.go` to eliminate code duplication between normal operations and retry handling.

## Design Decisions

The implementation makes several key design decisions:
- Generic type parameter for type safety without runtime assertions
- Handler interface for business logic separation
- Centralized cleanup logic in the coordinator
- Logger passed as parameter rather than configuration
- Simple data-only configuration structure

## Testing

All existing tests pass without modification. The refactoring maintains full backward compatibility and doesn't change the external behavior of the retry mechanisms.

## Migration

No data migration is required. The new structure will be used for new retry entries while existing entries continue to work with the current implementation.