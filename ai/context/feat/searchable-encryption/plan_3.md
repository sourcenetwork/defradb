# Phase 3: Replication and Remote Storage - Development Plan

## Overview
This phase implements the P2P protocol enhancements to replicate searchable encryption artifacts to remote nodes. It builds on Phase 2's artifact generation to enable distributed storage of encrypted search indexes.

## Key Characteristics
- SE artifacts are only stored on replicator nodes, not on the producer node
- Only keys are needed for search - values are stored as empty byte arrays
- Retry information stored in Peerstore, not Datastore
- Artifacts are regenerated from document fields during retry
- ReplicationFailureEvent includes field names for regeneration

## Architecture Overview

### Event Flow
The system uses an event-driven architecture to handle SE artifact replication:

1. **event.Update** triggers when blocks are committed
2. **ReplicationCoordinator** processes the event and generates artifacts
3. **ReplicateEvent** is published for P2P distribution
4. **StoreArtifactsEvent** triggers storage on remote nodes
5. **ReplicationFailureEvent** triggers retry mechanism on failures

### Storage Structure
- Artifact keys: `/se/<CollectionID>/<IndexID>/<SearchTag>/<DocID>`
- Retry keys: `/se/retry/<PeerID>/<CollectionID>/<DocID>`
- Empty values - only keys needed for search operations

## Core Components

### 1. SE Event System
The event system provides communication between SE components:
- **ReplicateEvent**: Published when artifacts need replication
- **StoreArtifactsEvent**: Published when receiving artifacts from peers  
- **ReplicationFailureEvent**: Published when replication fails

### 2. ReplicationCoordinator
Central component that orchestrates SE operations:
- Listens to `event.Update` for block commits
- Generates artifacts from document field values
- Manages retry mechanism with exponential backoff
- Stores artifacts on remote nodes

### 3. Network Protocol Enhancement
P2P communication uses existing GRPC infrastructure with new message types:
- `pushSEArtifactsRequest`: Contains artifacts and metadata
- `pushSEArtifactsHandler`: Receives and processes artifacts
- Integration with existing peer replicator lists

### 4. Retry Mechanism
The retry system ensures eventual consistency:
- Failed replications stored in Peerstore with retry metadata
- Background goroutine checks every 2 seconds for due retries
- Exponential backoff: 2s, 4s, 8s, 16s...
- Artifacts regenerated from current document values during retry

## Integration Points

- **DB Initialization**: ReplicationCoordinator created during DB setup
- **Event Subscriptions**: Components subscribe to relevant SE events
- **Existing P2P Infrastructure**: Reuses libp2p and GRPC systems

## Current Status

Phase 3 is implemented with the following completed components:
- Event-driven replication architecture
- ReplicationCoordinator with retry mechanism  
- Network protocol extensions for SE artifacts
- Storage structure for remote artifact storage
- Integration with document update flow from Phase 2

## Dependencies
- Phase 2 artifact generation
- DefraDB P2P infrastructure
- Event bus system
- Datastore and Peerstore