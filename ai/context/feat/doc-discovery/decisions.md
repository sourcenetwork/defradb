# Document Discovery Decision Journal

This file will be populated during development to record key decisions, trade-offs, and alternative approaches considered during implementation.

## Decision Log

### 1. Event-Based Architecture Pattern
**Decision**: Follow the same event-driven pattern as Searchable Encryption (SE)
**Rationale**: Proven architecture that integrates well with DefraDB's existing event bus
**Alternative**: Direct P2P calls without events - rejected due to tight coupling

### 2. DB Interface Extension
**Decision**: Added `ExecRequest` method to P2P DB interface
**Rationale**: Needed for query execution in P2P context
**Alternative**: Complex query building from scratch - rejected due to complexity

### 3. Collection ID vs Name in Network Protocol
**Decision**: Use CollectionID in network messages, convert to name locally
**Rationale**: CollectionID is stable and unique across nodes, names are needed for GraphQL
**Alternative**: Send collection name - rejected due to potential naming conflicts

### 4. Response Handling Simplification
**Decision**: Simplified response aggregation to single response per pubsub call
**Rationale**: Aligns with existing pubsub patterns in codebase
**Alternative**: Complex multi-response collection - deferred to future iteration

### 5. Filter and Result Processing Implementation
**Decision**: Implemented JSON-based filter string building and result extraction
**Rationale**: Provides functional filter processing using JSON-to-GraphQL conversion
**Alternative**: Complex GraphQL AST building - rejected due to complexity
**Note**: Enhanced from placeholder to working implementation for filter support

### 6. GraphQL Return Type Structure
**Decision**: Use DiscoveryResult object with docIDs field (similar to EncryptedSearchResult)
**Rationale**: Consistent with encrypted search pattern, enables GraphQL sub-selection
**Alternative**: Direct string array return - rejected due to GraphQL schema constraints