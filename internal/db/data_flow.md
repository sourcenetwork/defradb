# DefraDB Data Flow: Create Document Request Lifecycle

This document provides a comprehensive overview of how data flows through DefraDB for creating and updating a document,from initial client request to distributed synchronization across peers. It serves as a technical reference for developers at all levels working with the codebase.

## Table of Contents

### Part 1: Overview and Architecture
1. [High-Level Data Flow Overview](#high-level-data-flow-overview)
2. [Key Components](#key-components)

### Part 2: Local Operations
3. [Request Entry Points](#request-entry-points)
4. [Document Creation Flow](#document-creation-flow)
5. [Storage Layer](#storage-layer)

### Part 3: Distributed Operations
6. [Event System](#event-system)
7. [Network Synchronization](#network-synchronization)
8. [Merge Process](#merge-process)

### Part 4: System Behaviors
9. [Error Handling and Recovery](#error-handling-and-recovery)
10. [Summary](#summary)

## High-Level Data Flow Overview

```
┌───────────────────────────────────────────────────────────────────────────────┐
│                                   NODE A (Local)                              │
├───────────────────────────────────────────────────────────────────────────────┤
│                                                                               │
│  Client Request                                                               │
│       │                                                                       │
│       ▼                                                                       │
│  ┌─────────┐      ┌─────────────────────────────────────────────────────┐     │
│  │   API   │─────▶│                  Collection.Save()                  │     │
│  │ Handler │      ├─────────────────────────────────────────────────────┤     │
│  └─────────┘      │  ┌─────────────┐  ┌──────────────┐  ┌────────────┐  │     │
│                   │  │    CRDT     │─▶│   AddDelta   │─▶│   Store    │  │     │
│                   │  │  Creation   │  │   Function   │  │   Blocks   │  │     │
│                   │  └─────────────┘  └──────────────┘  └──────┬─────┘  │     │
│                   │                                            │        │     │
│                   │           ┌─────────────────┬──────────────┴┐       │     │
│                   │           ▼                 ▼               ▼       │     │ 
│                   │    ┌─────────────┐    ┌──────────┐    ┌──────────┐  │     │ 
│                   │    │ Blockstore  │    │Datastore │    │Headstore │  │     │ 
│                   │    │   (IPLD)    │    │ (Local)  │    │ (Heads)  │  │     │ 
│                   │    └─────────────┘    └──────────┘    └──────────┘  │     │ 
│                   │                                                     │     │ 
│                   │          After all storage completes:               │     │
│                   │                     │                               │     │
│                   │                     ▼                               │     │
│                   │              ┌─────────────┐                        │     │
│                   │              │event.Update │                        │     │
│                   │              └──────┬──────┘                        │     │
│                   └─────────────────────┼───────────────────────────────┘     │
│                                         ▼                                     │
│                                 ┌──────────────┐                              │
│                                 │    Send      │                              │
│                                 │pushLogRequest│                              │
│                                 └──────┬───────┘                              │
│                                        │                                      │
└────────────────────────────────────────┼──────────────────────────────────────┘
                                         │
                      Network Layer      │         
                      ══════════════════════════════════════════
                                         │                                    
                                         ▼                                    
┌───────────────────────────────────────────────────────────────────────────────┐
│                                   NODE B (Peer)                               │
├───────────────────────────────────────────────────────────────────────────────┤
│                                        │                                      │
│                                 ┌──────────────┐                              │
│                                 │   Handle     │                              │
│                                 │pushLogRequest│                              │
│                                 └──────┬───────┘                              │
│                                        │                                      │
│                                        ▼                                      │
│                                 ┌──────────────┐      ┌─────────────┐         │
│                                 │   syncDAG    │─────▶│ Blockstore  │         │
│                                 │  (Verify &   │      │   (IPLD)    │         │
│                                 │   Fetch)     │◀─────│             │         │
│                                 └──────┬───────┘      └─────────────┘         │
│                                        │                                      │
│                                        ▼                                      │
│                                 ┌──────────────┐                              │
│                                 │ event.Merge  │                              │
│                                 └──────┬───────┘                              │
│                                        │                                      │
│                                        ▼                                      │
│  ┌─────────────────────────────────────────────────────────────────────────┐  │
│  │                            executeMerge()                               │  │
│  ├─────────────────────────────────────────────────────────────────────────┤  │
│  │                                                                         │  │
│  │  ┌────────────┐    ┌─────────────┐    ┌──────────────┐    ┌──────────┐  │  │ 
│  │  │   Create   │───▶│    Load     │───▶│   Process    │───▶│  Update  │  │  │ 
│  │  │   Merge    │    │  Composite  │    │   Blocks     │    │  Indexes │  │  │ 
│  │  │  Targets   │    │   Blocks    │    │ (Recursive)  │    │  & Heads │  │  │ 
│  │  └────────────┘    └─────────────┘    └──────┬───────┘    └──────────┘  │  │ 
│  │                                              │                          │  │ 
│  │                                              ▼                          │  │ 
│  │                                     ┌─────────────────┐                 │  │ 
│  │                                     │   Decryption    │                 │  │ 
│  │                                     │   (if needed)   │                 │  │ 
│  │                                     └─────────────────┘                 │  │ 
│  │                                                                         │  │ 
│  └─────────────────────────────────────────────────────────────────────────┘  │ 
│                                                                               │
└───────────────────────────────────────────────────────────────────────────────┘
```

## Key Components

Before diving into the data flow, it's essential to understand the key components that participate in DefraDB's operation:

### Core Components
- **Collection**: Schema-defined document containers (`internal/db/collection.go`)
- **CRDT**: Conflict-free Replicated Data Types for distributed consistency (`internal/core/crdt/`)
- **Block**: Content-addressed data units in IPLD format (`internal/core/block/`)
- **Event Bus**: Coordinates distributed operations (`event/`)

### Storage Components
- **Blockstore**: IPLD block storage, shared across peers
- **Datastore**: Local materialized view for queries
- **Headstore**: Tracks latest document versions

### Network Components
- **Peer**: P2P node managing connections (`net/peer.go`)
- **Server**: gRPC server for network requests (`net/server.go`)
- **DAG Sync**: Ensures complete block availability (`net/sync_dag.go`)

---

## Part 2: Local Operations

This section covers operations that happen within a single DefraDB node, from receiving a request to storing data locally.

## Request Entry Points

Requests can enter DefraDB through multiple interfaces, all ultimately reaching the same core logic:

### Entry Path Hierarchy
```
Client Request
├── Direct Go Client (embedded)        → client/db.go
└── HTTP API                           → http/handler.go
    ├── REST endpoints                 → http/handler_*.go
    ├── GraphQL endpoint               → http/handler.go
    └── CLI                            → cli/client.go
```

**Key Files:**
- `client/db.go` - Main client interface implementation
- `http/handler.go` - HTTP request routing and handling  
- `http/handler_collection.go` - Collection-specific handlers
- `cli/client.go` - Command-line interface wrapper
- `cli/collection_*.go` - Collection CLI commands

All paths converge at the client DB interface, providing a unified entry point for database operations regardless of the access method.

## Document Creation Flow

When a user creates or updates a document, the request follows this detailed path:

### 1. Collection Save Method

**Related Files:**
- `internal/db/collection.go` - Main collection operations and save method

The save operation performs these steps in sequence (pseudo-code):

```go
// Pseudocode representation of the save flow
func (c *collection) save(ctx context.Context, doc *Document) error {
    // Step 1: Create/update secondary indexes
    if err := c.indexDocument(doc); err != nil {
        return err
    }
    
    // Step 2: Check if document signing is enabled
    if c.db.config.Signing.Enabled {
        // Prepare for signing
    }
    
    // Step 3: Create MerkleTree-based CRDTs for each field
    for fieldName, fieldValue := range doc.Fields() {
        crdt := createFieldCRDT(fieldName, fieldValue)
        if err := c.blockstore.Put(crdt); err != nil {
            return err
        }
    }
    
    // Step 4: Create composite block containing all fields
    compositeBlock := createCompositeBlock(doc)
    if err := c.blockstore.Put(compositeBlock); err != nil {
        return err
    }
    
    // Step 5: Publish update event
    c.eventBus.Publish(event.Update{
        DocID:      doc.ID(),
        Collection: c.Name(),
        Block:      compositeBlock,
    })
    
    return nil
}
```

### 2. CRDT Types and Creation

**Related Files:**
- `internal/core/crdt/merklecrdt.go` - MerkleCRDT factory and core logic
- `internal/core/crdt/lww.go` - LWW Register implementation
- `internal/core/crdt/counter.go` - Counter implementations
- `internal/core/crdt/composite.go` - Composite CRDT
- `internal/core/crdt/base.go` - Base CRDT interfaces

When a document field is modified, DefraDB creates a Conflict-free Replicated Data Type (CRDT):

**CRDT Type Selection:**
- **LWW (Last-Write-Wins) Register**: Used for basic scalar fields (strings, numbers, booleans)
- **PN-Counter**: Used for fields that support increment/decrement operations
- **P-Counter**: Used for positive-only counter fields
- **Composite CRDT**: Used for complex object structures

**Creation Process:**
1. The field type determines which CRDT implementation to use
2. `FieldLevelCRDTWithStore` instantiates the appropriate CRDT (e.g., `NewLWW`, `NewCounter`)
3. The CRDT's `Delta()` method generates a change record
4. This delta is passed to `AddDelta` for storage and distribution

These CRDTs enable conflict-free merging across distributed nodes by ensuring that concurrent updates can be deterministically resolved.

## Storage Layer

DefraDB uses a multi-store architecture to separate different types of data for optimal performance and organization:

### Storage Components

**Implementation Files:**
- `datastore/store.go` - MultiStore interface definition
- `datastore/blockstore.go` - Blockstore wrapper
- `datastore/txn.go` - Transaction handling
- `internal/keys/` - Key formatting for each store type

The `MultiStore` interface defines all available stores:
- Rootstore - base transactional store
- Datastore - document data (/data namespace)
- Encstore - encryption keys (/enc namespace)
- Headstore - document heads (/head namespace)
- Peerstore - peer information (/peers namespace)
- Blockstore - iPLD blocks (/blocks namespace)
- Systemstore - system metadata (/system namespace)

### AddDelta Function Flow

**Related Files:**
- `internal/core/block/store.go` - AddDelta implementation and block storage
- `internal/core/block/block.go` - Block structure definitions
- `internal/core/block/encryption.go` - Encryption block handling
- `internal/core/block/signature.go` - Block signing
- `internal/encryption/encryptor.go` - Encryption implementation

The `AddDelta` function is called by the collection's save method after creating CRDTs:
- **For field updates**: Called once per modified field with the field's delta
- **For composite blocks**: Called with the composite block containing all field references

**When AddDelta is invoked:**
1. During document creation (`collection.create()`)
2. During document updates (`collection.update()`)
3. Both operations call `collection.save()`, which creates CRDTs and calls AddDelta

The `AddDelta` function orchestrates the storage process:

1. **Height Update**: Increments the block height for version tracking
2. **Encryption Check**: 
   - If encryption is enabled (`internal/core/block/encryption.go`):
   - Creates encryption block with decryption key
   - Attaches as IPLD link to main block
   - Encrypts block content (`internal/encryption/encryptor.go`)
   - Replaces plaintext with ciphertext
3. **Signing**: If enabled, signs the block for integrity verification
4. **Storage**:
   - **Encrypted/signed block → Blockstore**: The transformed block (with encryption and/or signature) is stored in the blockstore. This is the version that gets distributed to other peers during synchronization.
   - **Original block → Local processing**: The unencrypted block data is used for local operations like updating indexes and merging CRDT states. This ensures the local node can work with the data without decryption overhead.
5. **Local Processing**:
   - Merges CRDT changes into the current document state in datastore
   - Updates document heads in headstore to track the latest version
   - Triggers index updates in datastore for query optimization

### Storage Types Explained

Each store serves a specific purpose in DefraDB's architecture:

- **Rootstore**: The base transactional store that supports both batching and transactions. All other stores are namespaced wrappers around this root.

- **Blockstore** (`/blocks`): IPLD-based content-addressed storage for all blocks (documents, fields, encryption blocks). This is the primary store for data that needs to be synchronized across peers. Each block is stored by its CID (Content Identifier).

- **Datastore** (`/data`): Local materialized view of documents optimized for queries. Contains the current state of documents after all CRDT merges have been applied. This is what query operations read from.

- **Headstore** (`/head`): Tracks the latest CID (head) for each document. Critical for the merge process as it determines which blocks are new versus already processed. Maps document IDs to their current head CIDs.

- **Encstore** (`/enc`): Stores symmetric encryption keys for document encryption. Keys are organized by document ID and field name, allowing field-level encryption granularity.

- **Peerstore** (`/peers`): Contains information about known peers, replication settings, and peer-specific metadata needed for P2P operations.

- **Systemstore** (`/system`): Stores system-wide metadata including schema definitions, collection information, index definitions, and configuration data.

---

## Part 3: Distributed Operations

This section covers how DefraDB synchronizes data across multiple nodes in the network, ensuring eventual consistency through its event-driven architecture.

## Event System

**Related Files:**
- `event/event.go` - Event type definitions
- `internal/db/db.go` - Event subscriptions setup
- `net/peer.go` - Network event handling

DefraDB's event system coordinates distributed operations by publishing events at key points in the data flow. These events trigger both local and network-wide actions.

### Core Events

1. **event.Update**: Published after successful document creation/update
   - Triggers network synchronization
   - Contains document ID, collection, and block data

2. **event.Merge**: Published when receiving updates from peers
   - Initiates local merge process
   - Contains block data to merge

3. **event.MergeComplete**: Published after successful merge
   - Signals completion of synchronization
   - May trigger dependent operations

4. **encryption.RequestKeysEvent**: Published when encryption blocks needed
   - Handled by KMS components
   - Triggers network requests for decryption keys

### Event Flow Diagram

```
Document Create
    ↓
event.Update → Network Broadcast → Peer Receives
    ↓                                   ↓
Local Storage                      event.Merge
                                        ↓
                                   executeMerge
                                        ↓
                                 event.MergeComplete
```

## Network Synchronization

Network synchronization in DefraDB happens through a push-based model where nodes actively send updates to interested peers.

### Event Publishing

**Related Files:**
- `net/peer.go` - Peer management and log replication
- `net/node.go` - P2P node implementation
- `net/server.go` - gRPC server handling
- `net/grpc.go` - Protocol definitions
- `event/event.go` - Event type definitions

When an `event.Update` occurs:
1. Node sends updates to replicators via `pushLogToReplicators`
2. Broadcasts to peers subscribed to the document/collection
3. Uses `pushLogRequest` (`net/grpc.go`) for network transmission

### Receiving Updates

**Related Files:**
- `net/server.go` - gRPC server and request handlers

When a peer receives an update, it follows this process:
```go
func (s *server) pushLogHandler(req *pushLogRequest) error {
    // 1. Network validation and decoding
    block := decodeBlock(req.Data)
    
    // 2. Synchronize DAG
    if err := s.syncDAG(block); err != nil {
        return err
    }
    
    // 3. Publish merge event
    s.eventBus.Publish(event.Merge{
        Block: block,
    })
    
    return nil
}
```

### DAG Synchronization

**Related Files:**
- `net/sync_dag.go` - DAG synchronization logic
- `internal/core/block/signature.go` - Signature verification

The `syncDAG` function ensures complete block availability:
1. Fetches the requested block and all sub-blocks
2. Verifies signatures before processing
3. Stores blocks in local blockstore
4. Ensures referential integrity

## Merge Process

The merge process is the heart of DefraDB's conflict resolution, ensuring that updates from different nodes are integrated consistently.

### Overview

The merge process:
1. Determines what needs to be synchronized (merge targets)
2. Loads all necessary blocks from storage
3. Processes blocks recursively, handling encryption and CRDTs
4. Updates local state and indexes

### ExecuteMerge Function

**Related Files:**
- `internal/db/merge.go` - Complete merge implementation

The merge process handles incoming blocks from peers and integrates them into the local state through a sophisticated recursive algorithm:

#### 1. Creating Merge Targets

**Merge targets** define the synchronization points for the merge operation:
- Retrieves current document heads from the headstore using `getHeadsAsMergeTarget()`
- Each target contains:
  - A map of head CIDs to their corresponding blocks
  - The priority (height) of these heads
- All heads in a target have the same priority, representing the current "frontier" of the document's DAG

#### 2. Loading Blocks from Blockstore

The `loadComposites()` function recursively traverses the block DAG:
- Starts from the new block received from a peer
- Walks backwards through parent blocks until reaching:
  - Already processed blocks (present in current heads)
  - Genesis blocks (no parents)
  - Blocks with priority lower than the target
- **Priority handling**:
  - If new block priority ≥ target priority: Add to front of processing queue
  - If new block priority < target priority: Create new merge target and recurse with parent blocks
- This ensures blocks are processed in the correct order while handling divergent branches

#### 3. Processing Blocks Recursively

The merge processor handles blocks through several phases:

**Decryption Phase:**
- Checks if block is encrypted
- Attempts to fetch encryption block from local store
- On failure: Caches the failure and continues (will retry later)
- On success: Decrypts and processes the block

**CRDT Merge Phase:**
- Initializes appropriate CRDT type based on field type
- Applies the delta to merge changes
- Handles conflicts according to CRDT semantics (e.g., LWW uses timestamp)

**Recursive Processing:**
- After processing a composite block, recursively processes all linked field blocks
- Maintains processing order to ensure consistency

#### 4. Encryption Key Recovery

**Key Files:**
- `internal/kms/pubsub.go` - KMS event handling
- `internal/encryption/context.go` - Encryption context management

If decryption fails during merge:
- Publishes `encryption.RequestKeysEvent`
- KMS subscribers attempt to fetch keys from the network
- On key retrieval, calls `tryFetchMissingBlocksAndMerge()` which:
  - Re-attempts the merge with newly available keys
  - Can trigger another full merge cycle if successful

#### 5. Final Updates

After successful merge:
- Updates secondary indexes with changed documents
- Updates headstore with new head CIDs
- Publishes `event.MergeComplete` for downstream consumers

---

## Part 4: System Behaviors

This section covers cross-cutting concerns that affect the entire data flow.

## Error Handling and Recovery

DefraDB implements robust error handling throughout the data flow:

### Key Recovery Mechanisms

1. **Encryption Key Recovery**:
   - Caches decryption failures
   - Asynchronously requests keys from network
   - Retries merge after key acquisition

2. **Network Resilience**:
   - Automatic retry for failed synchronizations
   - Peer discovery for alternative sources
   - Eventual consistency guarantees

3. **Transaction Rollback**:
   - Failed operations don't corrupt state
   - Atomic updates ensure consistency
   - Clear error propagation to clients

## Summary

DefraDB's data flow follows this path: Client requests → Document updates → CRDT creation → Block storage → Event publication → Network sync → Peer merge.

The architecture uses content-addressed blocks (IPLD), conflict-free data types (CRDTs), and event-driven synchronization to maintain consistency across distributed nodes without central coordination. A multi-store design separates immutable blocks, materialized views, and metadata for optimal performance and network efficiency.

### Further Reading

For deeper dives into specific components:
- Query execution: See `internal/planner/` for query planning and optimization
- Schema management: See `internal/request/graphql/schema/` for schema operations
- Index operations: See `internal/db/index.go` for secondary index management
- Access control: See `acp/` for permission and encryption systems
