---
description: DefraDB Peer-to-Peer networking architecture and implementation guidelines
globs: "**/net/**/*.go, **/p2p/**/*.go"
alwaysApply: true
---

# P2P Networking System Guide

DefraDB leverages peer-to-peer (P2P) networking for distributed data exchange, synchronization, and replication. This guide explains the architecture, implementation details, and best practices for working with the P2P system.

## Core Concepts

### P2P Architecture

DefraDB's P2P system is built on libp2p and provides:

1. **Node Identity**:
   - Each node has a unique PeerID derived from a public key
   - Ed25519 keys for node identification and authentication
   - Multiaddress format for addressing (e.g., `/ip4/127.0.0.1/tcp/9171/p2p/<peerID>`)

2. **Network Topology**:
   - Decentralized mesh network
   - No central coordinator
   - Direct peer-to-peer connections
   - Dynamic peer discovery

3. **Synchronization Methods**:
   - **Pubsub**: Passive broadcast-based synchronization
   - **Replicator**: Active targeted synchronization
   - **DAG Sync**: Merkle-based differential synchronization

## Key Components

### Host Management

The `Host` interface manages network connectivity:

```go
type Host interface {
    Start(context.Context) error
    Stop(context.Context) error
    ID() peer.ID
    Addrs() []ma.Multiaddr
    // Additional methods for peer management
}
```

### Pubsub System

The pubsub system enables passive synchronization:

1. **Topic Subscription**:
   - Collections are identified as topics
   - Nodes subscribe to collection topics
   - Document commits are broadcast to subscribers

2. **Message Exchange**:
   - Updates are encoded and broadcast
   - Recipients decode and apply changes
   - Content-addressing enables deduplication

### Replicator System

The replicator system provides active synchronization:

1. **Collection Replication**:
   - Targets specific collections for replication
   - Actively pushes changes to configured peers
   - Can be one-way or bidirectional

2. **Retry Mechanism**:
   - Handles temporary peer unavailability
   - Exponential backoff for repeated failures
   - Persists replication configuration

### DAG Synchronization

DAG sync efficiently exchanges document commit history:

1. **Differential Sync**:
   - Uses Merkle structure to identify differences
   - Only transmits missing commits
   - Efficient for large documents with small changes

2. **Content Addressing**:
   - CIDs (Content Identifiers) used for commit referencing
   - Enables verification of data integrity
   - Supports deduplication across the network

## Implementation Architecture

### Network Configuration

DefraDB's network layer can be configured through:

1. **Listening Addresses**:
   - Configurable via `--p2paddr` flag
   - Supports multiple addresses and protocols
   - Default is `/ip4/127.0.0.1/tcp/9171`

2. **Bootstrapping**:
   - Initial peers via `--peers` flag
   - Auto-connection to specified peers
   - Dynamic peer discovery thereafter

3. **Options**:
   - Enable/disable pubsub: `--pubsubenabled`
   - Enable/disable P2P entirely: `--p2pdisabled`
   - Enable relay: `--relay`

### Protocol Handlers

The system implements several protocol handlers:

1. **Collection Protocol**:
   - Exchange collection metadata
   - Manage collection subscriptions
   - Coordinate collection replication

2. **Document Protocol**:
   - Exchange document commits
   - Synchronize document state
   - Handle document versioning

3. **Block Protocol**:
   - Exchange raw IPLD blocks
   - Low-level data transfer
   - Content-addressed storage

## Implementation Guidelines

When working with the P2P system:

### Extending P2P Capabilities

1. Define clear protocol semantics
2. Implement proper error handling and recovery
3. Consider both connected and disconnected operations
4. Design for consensus-free operation when possible
5. Add comprehensive tests covering:
   - Protocol message exchange
   - Connection/disconnection scenarios
   - Network partition handling
   - Data consistency verification

### Performance Considerations

1. **Message Size Optimization**:
   - Minimize message payload size
   - Use delta encoding when possible
   - Consider batching for small updates

2. **Connection Management**:
   - Limit concurrent connections
   - Implement connection pooling
   - Handle connection backpressure

3. **Resource Control**:
   - Implement rate limiting for peers
   - Monitor bandwidth usage
   - Prioritize critical synchronization traffic

### Security Best Practices

1. **Authentication**:
   - Verify peer identity for sensitive operations
   - Use channel encryption for all communications
   - Implement access control for replicated collections

2. **Denial of Service Prevention**:
   - Limit resources per peer
   - Implement message validation before processing
   - Add circuit breakers for misbehaving peers

3. **Data Validation**:
   - Verify data integrity with content addressing
   - Validate document commits before application
   - Check schema compliance for received documents

## P2P Usage Patterns

DefraDB supports two main P2P synchronization patterns:

### Pubsub Peering

For passive synchronization between peers:

```sh
# Node A (default configuration)
defradb start

# Node B (connects to Node A)
defradb start --rootdir ~/.defradb-nodeB --url localhost:9182 \
  --p2paddr /ip4/127.0.0.1/tcp/9172 \
  --peers /ip4/127.0.0.1/tcp/9171/p2p/<peerID-of-A>
```

Collection subscription enables specific collections to be synchronized:

```sh
# Subscribe to collection on Node B
defradb client p2p collection add --url localhost:9182 <collectionID>
```

### Replicator Peering

For active, targeted synchronization:

```sh
# Set Node A to actively replicate to Node B
defradb client p2p replicator set -c <CollectionName> <nodeB_peer_info_json>
```

This configures Node A to actively push changes for the specified collection to Node B.

## Testing P2P Implementations

When implementing or modifying P2P components:

1. **Protocol Testing**:
   - Test message encoding/decoding
   - Verify protocol state machines
   - Test with invalid/malformed messages
   - Ensure proper error handling

2. **Network Scenario Testing**:
   - Test connection establishment and teardown
   - Simulate network partitions
   - Test with varying latencies and bandwidth
   - Verify behavior with disconnected peers

3. **Data Synchronization Testing**:
   - Verify document consistency across peers
   - Test concurrent updates from multiple peers
   - Measure synchronization performance
   - Verify correct conflict resolution

4. **Integration Testing**:
   - Test interaction with other DefraDB components
   - Verify end-to-end synchronization
   - Test with realistic network conditions
   - Benchmark real-world performance

## Limitations and Caveats

Current limitations to be aware of:

1. P2P doesn't fully support collections with ACP when using Local ACP
2. Replication requires identical schema on all nodes
3. Network discovery is limited to direct peer configuration
4. Performance may degrade with many collections or large documents