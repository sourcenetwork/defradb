---
description: MerkleCRDT architecture and implementation guidelines
globs: "**/crdt/**/*.go, **/core/**/*.go"
alwaysApply: true
---

# MerkleCRDT System Guide

MerkleCRDTs are at the core of DefraDB's data model, enabling multi-write-master architecture with automatic conflict resolution. This guide explains how the MerkleCRDT system works and provides guidance for extending or modifying it.

## Core Concepts

### MerkleCRDT Fundamentals

1. **CRDT (Conflict-free Replicated Data Type)**:
   - Data structures that can be replicated across multiple computers
   - Updates can be applied in any order
   - Automatically resolve conflicts to a consistent state
   - Allow concurrent updates without central coordination

2. **Merkle DAG (Directed Acyclic Graph)**:
   - Content-addressable storage based on cryptographic hashes
   - Provides deduplication and data integrity verification
   - Enables efficient synchronization between peers
   - Similar to Git's underlying data model

3. **MerkleCRDT Combination**:
   - CRDTs stored in a Merkle DAG structure
   - Updates reference previous versions (like Git commits)
   - Content-addressing enables efficient synchronization
   - Provides strong consistency guarantees in distributed systems

## DefraDB's CRDT Types

DefraDB implements several CRDT types:

1. **Last-Write-Wins (LWW) Register**:
   - Resolves conflicts by selecting the update with the latest timestamp
   - Used for basic field types like strings, numbers, booleans
   - Implements `LWWRegister` interface

2. **PN-Counter**:
   - Supports increments and decrements
   - Maintains separate counters for increments (P) and decrements (N)
   - Final value is P-N (difference)
   - Used for numeric fields that need counter semantics

3. **Composite CRDT**:
   - Contains multiple named CRDT fields
   - Used for structured objects with multiple fields
   - Each field can have its own CRDT type

## CRDT Implementation Structure

The CRDT implementation is organized around several key interfaces:

### `MerkleCRDT` Interface

The foundational interface that all CRDT types implement:

```go
type MerkleCRDT interface {
    Update(ctx context.Context, delta []byte) error
    Delta() ([]byte, error)
    Value() (interface{}, error)
    // Additional methods for IPLD integration
}
```

### Field CRDT Interface

Specialized interface for field-level CRDTs:

```go
type FieldCRDT interface {
    MerkleCRDT
    FieldID() uint32
    Type() encoding.FieldType
    // Additional methods specific to fields
}
```

### Composite CRDT

For document-level CRDT management:

```go
type CompositeCRDT interface {
    MerkleCRDT
    GetField(id uint32) (FieldCRDT, error)
    SetField(id uint32, value FieldCRDT) error
    // Methods for field management
}
```

## Data Flow in the CRDT System

1. **Document Updates**:
   - Client submits updates to document fields
   - System converts updates to CRDT deltas
   - Deltas are applied to existing CRDT state
   - New state is persisted in the Merkle DAG

2. **Conflict Resolution**:
   - Multiple nodes may update the same field
   - Updates form a DAG of changes
   - CRDT type determines resolution strategy
   - System guarantees eventual consistency

3. **P2P Synchronization**:
   - Nodes exchange CRDT updates via P2P network
   - Updates are integrated into local state
   - Merkle structure enables efficient delta transfers
   - Only new/changed parts are transmitted

## Implementation Guidelines

When working with the MerkleCRDT system:

### Adding New CRDT Types

1. Implement the `MerkleCRDT` interface
2. Define clear semantics for conflict resolution
3. Create efficient binary delta encoding
4. Implement IPLD serialization/deserialization
5. Add appropriate test cases covering:
   - Basic operations
   - Concurrent updates
   - Various conflict scenarios
   - Serialization/deserialization

### Modifying Existing CRDTs

1. Maintain backward compatibility with existing data
2. Document any semantic changes to resolution rules
3. Ensure changes preserve CRDT properties:
   - Commutativity: Order of operations doesn't matter
   - Associativity: Grouping of operations doesn't matter
   - Idempotence: Repeated application doesn't change result

### Performance Considerations

1. Minimize delta size for efficient network transfer
2. Optimize merge operations for common scenarios
3. Consider memory usage for large documents
4. Design for incremental updates to avoid full document transfers

## Testing CRDT Implementations

When implementing or modifying CRDTs:

1. **Verify CRDT Properties**:
   - Test commutativity by applying operations in different orders
   - Test associativity by grouping operations differently
   - Test idempotency by applying the same operation multiple times

2. **Test Concurrent Scenarios**:
   - Simulate concurrent updates from multiple nodes
   - Verify convergence to the same state
   - Test with various network conditions (delays, reordering)

3. **Performance Testing**:
   - Benchmark update operations
   - Measure memory usage
   - Test with large documents and high operation counts

4. **Integration Testing**:
   - Verify behavior in the full DefraDB system
   - Test P2P synchronization of CRDTs
   - Confirm interaction with other components (queries, indexes, etc.)

## Debugging CRDT Issues

When investigating CRDT-related problems:

1. Examine the DAG structure of updates
2. Check for incomplete merges
3. Verify correct timestamp handling for LWW registers
4. Review delta encoding/decoding for corruption
5. Use the document commit history to track changes over time