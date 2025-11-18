# Document Discovery Feature Specifications

## Overview

Implement a document discovery mechanism that allows DefraDB nodes to request potentially unknown documents from the network. This feature enables nodes to discover and retrieve document IDs from peer nodes through an explicit GraphQL query interface.

## Requirements

### Core Functionality
- **Discovery Query Interface**: Create explicit GQL queries like `discover_User` for each collection
- **Network Communication**: Enable P2P discovery requests across the network
- **Response Aggregation**: Consolidate responses from multiple peers
- **DocID-only Results**: Return only document IDs, not materialized documents

### Query Parameters
- **Filter Support**: Accept regular GraphQL filters (same as normal queries)
- **Pagination**: Support `limit` and `offset` parameters
- **No Parameters**: Requesting all documents when no parameters provided
- **No Aggregation**: Explicitly exclude aggregate query support

### Technical Requirements
- **Event-driven Architecture**: Use DefraDB's event bus for coordination
- **P2P Integration**: Leverage existing P2P infrastructure
- **Non-blocking Execution**: Asynchronous query processing
- **Response Timeout**: Handle network timeouts gracefully
- **Integration Tests**: Complete test coverage in `/tests/integration/discover`

## Acceptance Criteria

1. **GraphQL Schema Generation**
   - [ ] Generate `discover_<CollectionName>` queries for each collection
   - [ ] Support standard filter, limit, and offset parameters
   - [ ] Exclude aggregation and mutation operations

2. **Query Processing**
   - [ ] Detect discovery queries during parsing
   - [ ] Route discovery queries to specialized scan node
   - [ ] Return document IDs only (not materialized documents)

3. **Network Layer**
   - [ ] Implement dedicated discovery topic for P2P communication
   - [ ] Handle discovery request transmission and reception
   - [ ] Support multiple peer responses consolidation

4. **Response Handling**
   - [ ] Aggregate responses from multiple peers
   - [ ] Handle partial responses and timeouts
   - [ ] Return consolidated document ID list

5. **Testing**
   - [ ] Integration tests for basic discovery functionality
   - [ ] Tests for filter, limit, offset parameters
   - [ ] Network failure and timeout scenarios
   - [ ] Multi-peer response aggregation tests

## Non-Requirements

- Aggregate query support on discovered documents
- Real-time document synchronization
- Document content retrieval (only IDs)
- Complex relationship traversal in discovery phase

## Architecture Alignment

This feature aligns with DefraDB's:
- **Event-driven architecture**: Uses existing event bus for coordination
- **P2P networking**: Leverages libp2p infrastructure
- **GraphQL compatibility**: Extends existing query language
- **Content-addressable design**: Works with document IDs and CIDs