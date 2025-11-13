# Backlog: Future Improvements

This document captures realistic improvements and technical debt related to lens migration and indexing that were identified but deferred during implementation.

## Potential Optimizations

### 1. Incremental Index Updates

**Context**: Currently, reindexing removes all entries and rebuilds from scratch.

**Improvement**: Implement incremental updates that only modify affected index entries.

**Benefits**:
- Reduced reindexing time for large collections
- Lower resource usage during migration operations
- Better performance for collections with many indexes

**Complexity**: High - requires tracking which documents are affected by migration

**Priority**: Medium

---

### 2. Parallel Index Rebuilding

**Context**: Indexes are rebuilt sequentially in a loop.

**Improvement**: Rebuild multiple indexes in parallel when they don't conflict.

**Benefits**:
- Faster reindexing for collections with multiple indexes
- Better utilization of multi-core systems

**Complexity**: Medium - need to handle transaction isolation

**Priority**: Low

---

### 3. Smart Reindexing Detection

**Context**: Currently checks if any migration exists between versions.

**Improvement**: Analyze migration to determine if indexed fields are actually affected.

**Example**:
- If migration only transforms `name` field but index is on `age`, skip reindexing

**Benefits**:
- Avoid unnecessary reindexing operations
- Improved performance for migrations that don't affect indexed fields

**Complexity**: High - requires migration analysis and field tracking

**Priority**: Medium

---

## Technical Debt

### 1. Reindexing Progress Tracking

**Issue**: No visibility into reindexing progress for large collections.

**Improvement**: Add progress tracking/logging for long-running reindex operations.

**Benefits**:
- Better user experience
- Ability to estimate completion time
- Early detection of performance issues

**Complexity**: Low

**Priority**: Low

---

### 2. Reindexing Metrics

**Issue**: No metrics collected for reindexing operations.

**Improvement**: Add telemetry for:
- Number of documents reindexed
- Time taken per index
- Success/failure rates
- Performance characteristics

**Benefits**:
- Better observability
- Ability to identify performance bottlenecks
- Data for optimization decisions

**Complexity**: Low

**Priority**: Medium

---

### 3. Error Recovery Granularity

**Issue**: If reindexing fails midway, all progress is lost (transaction rollback).

**Improvement**: Implement checkpoint/resume mechanism for reindexing.

**Benefits**:
- Better resilience for large collections
- Faster recovery from transient errors
- Reduced impact of reindexing failures

**Complexity**: High - requires careful transaction management

**Priority**: Low

---

## Out of Scope Items to Reconsider

### 1. Encrypted Index Support

**Context**: Searchable encryption indexes are currently out of scope.

**Future Consideration**: When encrypted indexes are more mature, ensure migration logic applies to them as well.

**Notes**:
- May require coordination with KMS
- Might need special handling for key rotation
- Could have different performance characteristics

---

### 2. Vector Embedding Updates

**Context**: Vector embeddings and similarity indexes not considered.

**Future Consideration**: When vector embeddings support migrations, ensure they're properly reindexed.

**Notes**:
- May need to regenerate embeddings with new values
- Could be very expensive operation
- Might need background processing

---

### 3. Migration Rollback

**Context**: No rollback functionality for migrations.

**Future Consideration**: Support undoing migrations and reverting index state.

**Notes**:
- Would require storing original index state
- Complex transaction semantics
- May not be feasible for all migration types

---

## Documentation Improvements

### 1. Migration Best Practices

**Need**: Document best practices for migrations with indexes.

**Topics to Cover**:
- Performance implications of reindexing
- How to minimize reindexing impact
- When to create indexes (before vs after migration)
- Testing strategies for migrations with indexes

---

### 2. Troubleshooting Guide

**Need**: Guide for debugging index/migration issues.

**Topics to Cover**:
- How to detect stale index data
- How to manually trigger reindexing
- Common error scenarios and resolutions
- Performance tuning tips

---

### 3. API Documentation

**Need**: Better documentation for migration-related APIs.

**Topics to Cover**:
- When reindexing happens automatically
- How to check if reindexing is needed
- Transaction semantics for migration operations
- Limitations and edge cases

---

## Testing Improvements

### 1. Performance Benchmarks

**Need**: Benchmark tests for reindexing operations.

**Purpose**:
- Establish baseline performance
- Detect performance regressions
- Validate optimization attempts

---

### 2. Stress Tests

**Need**: Tests with very large document counts and many indexes.

**Purpose**:
- Ensure scalability
- Identify memory issues
- Validate transaction handling

---

### 3. Concurrency Tests

**Need**: Tests for concurrent migration and query operations.

**Purpose**:
- Ensure proper locking
- Validate transaction isolation
- Detect race conditions

---

## Notes

- Items in this backlog are NOT immediate action items
- They represent potential future work based on user needs and system evolution
- Priority and complexity assessments are preliminary and may change
- Some items may become irrelevant as the system evolves
