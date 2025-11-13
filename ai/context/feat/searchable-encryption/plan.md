# Generic Retry Coordinator Implementation Plan

## Overview
This plan describes the implementation of a generic retry coordinator that unifies SE (Searchable Encryption) and document replication retry mechanisms in DefraDB.

## Architecture

### Core Components

1. **Generic Retry Coordinator (`internal/retry/coordinator.go`)**
   - Generic type `Coordinator[T any]` allowing different data types for retry items
   - Manages two-level retry structure: peer level (RetryInfo) + item level (custom data T)
   - Handles retry scheduling, backoff, and cleanup

2. **Handler Interface**
   ```go
   type Handler[T any] interface {
       ProcessItem(ctx context.Context, peerID string, itemID string, itemData T) error
       UpdateStatus(ctx context.Context, peerID string, active bool) error
   }
   ```

3. **Configuration**
   ```go
   type Config struct {
       Name              string
       RetryKeyPrefix    string          
       ItemKeyPrefix     string          
       RetryIntervals    []time.Duration
       CheckInterval     time.Duration
   }
   ```

### Implementation Approach

1. **Phase 1: Update Key System**
   - Add SE_RETRY_ID and SE_RETRY_ITEM constants to `internal/keys/peerstore.go`
   - Modify existing retry key types to support custom prefixes
   - Enable key reuse across different retry scenarios

2. **Phase 2: Create Generic Coordinator**
   - Implement `Coordinator[T]` with generic type parameter
   - Move UpdateStatus from config to handler interface
   - Pass logger as parameter instead of including in config
   - Implement standard cleanup method (deleteRetryAndItems)

3. **Phase 3: Implement SE Retry Handler**
   - Create SERetryInfo struct with necessary fields
   - Implement Handler[SERetryInfo] interface
   - Extract common SE replication logic to publishSEReplication method

4. **Phase 4: Update Document Retry Handler**
   - Implement Handler[struct{}] interface
   - Use UpdateStatus for replicator status updates
   - Remove deprecated retry methods

## Key Design Decisions

1. **Generic Type Parameter**: Allows type-safe storage of different data types without encoding/decoding in item IDs
2. **Handler Interface**: Keeps business logic separate from infrastructure
3. **Config Simplification**: Config only contains data, behavior is in handlers
4. **Cleanup Centralization**: Common cleanup pattern implemented once in coordinator

## Testing Strategy
- Ensure all existing p2p_replicator tests pass
- No new tests needed as this is a refactoring
- Manual verification of retry functionality for both SE and documents

## Migration
No data migration needed as requested - new structure will be used for new retries only.