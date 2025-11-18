# Document Discovery Feature

Implements network-based document discovery functionality that allows DefraDB nodes to request potentially unknown documents from peer nodes through explicit GraphQL queries.

## Changes

This implementation follows the same architectural pattern as Searchable Encryption, creating a parallel discovery path through DefraDB's query execution pipeline.

### Core Components Added
- Discovery query generation for GraphQL schema
- Discovery-specific scan node for query planning
- Event-driven P2P communication protocol
- Multi-peer response aggregation
- Integration test suite

### Key Features
- `discover_<CollectionName>` queries for each collection
- Standard GraphQL filter, limit, and offset parameter support
- Document ID-only responses (no materialized documents)
- Asynchronous P2P request handling
- Consolidated responses from multiple peers

The feature integrates seamlessly with DefraDB's existing event bus, P2P networking, and query execution infrastructure while maintaining the system's performance and security characteristics.