# Document Discovery Development Plan

## Implementation Strategy

This feature follows the same architectural pattern as Searchable Encryption (SE), creating a parallel discovery path through DefraDB's query execution pipeline. The implementation spans from GraphQL schema generation to P2P network communication.

## Phase 1: GraphQL Schema Generation

### 1.1 Discovery Query Type Generation
**File**: `internal/request/graphql/schema/generate.go`

Add discovery query generation similar to encrypted queries around line 1148:

```go
func (g *Generator) GenerateDiscoveryQueryForCollection(col client.CollectionDescription) (*gql.Object, error) {
    // Generate discover_<CollectionName> query type
    // Include filter, limit, offset parameters
    // Return type: [String!]! (array of docIDs)
}
```

**Integration Point**: Modify existing collection schema generation to include discovery queries alongside normal collection queries.

### 1.2 Query Arguments Configuration
**File**: `internal/request/graphql/schema/generate.go` (around line 1561-1562)

Extend argument configuration to include discovery-specific parameters:
- Reuse existing filter arguments from normal queries
- Include standard `limit` and `offset` parameters
- Add discovery-specific request identification

## Phase 2: Query Detection and Parsing

### 2.1 Discovery Query Detection
**File**: `internal/request/graphql/parser/query.go` (around line 95)

Add discovery query detection similar to encrypted query detection:

```go
const DiscoveryCollectionPrefix = "discover_"

isDiscovery := strings.HasPrefix(field.Name.Value, DiscoveryCollectionPrefix)
```

### 2.2 Request Mapping
**File**: `internal/planner/mapper/mapper.go` (around line 126-127)

Add discovery selection type:

```go
if selectRequest.IsDiscovery {
    rootSelectType = DiscoverySelection
}
```

Define `DiscoverySelection` constant and associated mapping logic.

## Phase 3: Query Planning and Execution

### 3.1 Discovery Scan Node Implementation
**File**: `internal/planner/discovery_scan.go` (new file)

Create `discoveryScanNode` similar to `seScanNode`:

```go
type discoveryScanNode struct {
    documentIterator
    docMapper

    p            *Planner
    collection   client.Collection
    collectionID string
    filter       *mapper.Filter
    limit        int
    offset       int

    remoteDocIDs []string
    hasReturned  bool
}
```

**Key Methods**:
- `Start()`: Initialize discovery request
- `queryRemoteNodes()`: Send discovery request via event bus
- `Next()`: Return consolidated document IDs

### 3.2 Planner Integration
**File**: `internal/planner/select.go` (around line 598, 613)

Add discovery query routing:

```go
if selectReq.IsDiscovery {
    return p.SelectDiscovery(selectReq)
}

func (p *Planner) SelectDiscovery(selectReq *mapper.Select) (planNode, error) {
    discoveryScan := &discoveryScanNode{
        // Initialize discovery scan node
    }
    return discoveryScan, nil
}
```

## Phase 4: Event System Integration

### 4.1 Discovery Events Definition
**File**: `event/event.go`

Add discovery-specific events:

```go
type RequestDocDiscoveryEvent struct {
    CollectionID   string
    Filter         map[string]any
    Limit          int
    Offset         int
    ResponseChan   chan DocDiscoveryResult
}

type DocDiscoveryResult struct {
    DocIDs []string
    Error  error
}
```

### 4.2 Event Publishing
**File**: `internal/planner/discovery_scan.go`

Implement event-based communication:

```go
func (n *discoveryScanNode) queryRemoteNodes() ([]string, error) {
    responseChan := make(chan DocDiscoveryResult)

    msg := event.RequestDocDiscoveryEvent{
        CollectionID: n.collectionID,
        Filter:       n.filter.ExternalConditions,
        Limit:        n.limit,
        Offset:       n.offset,
        ResponseChan: responseChan,
    }

    n.p.db.Events().Publish(msg)

    response := <-responseChan
    return response.DocIDs, response.Error
}
```

## Phase 5: P2P Network Layer

### 5.1 Discovery Protocol Implementation
**File**: `internal/db/p2p/discovery.go` (new file)

Add discovery-specific P2P handling:

```go
type DiscoveryProtocol struct {
    p2p *P2P
}

func (dp *DiscoveryProtocol) handleDiscoveryRequest(ctx context.Context, req DiscoveryRequest) error {
    // Forward to network peers via dedicated discovery topic
    // Aggregate responses from multiple peers
    // Send consolidated response back via ResponseChan
}
```

### 5.2 P2P Integration
**File**: `internal/db/p2p/p2p.go`

Add discovery event subscription:

```go
func New(ctx context.Context, db DB, host client.Host) (*P2P, error) {
    // ... existing initialization ...

    // Subscribe to discovery events
    db.Events().Subscribe(RequestDocDiscoveryEventName, p.handleDiscoveryEvent)

    // Add discovery topic to pubsub
    err = p.host.AddPubSubTopic(discoveryTopic, true, p.discoveryMessageHandler)
}
```

### 5.3 Network Message Handling
**File**: `internal/db/p2p/protocol/discovery.go` (new file)

Define discovery protocol messages:

```go
type DiscoveryNetworkRequest struct {
    CollectionID   string
    Filter         map[string]any
    Limit          int
    Offset         int
    SenderID       string
}

type DiscoveryNetworkResponse struct {
    DocIDs   []string
    SenderID string
    Error    error
}
```

## Phase 6: Query Execution Integration

### 6.1 Network Message Handler
**File**: `internal/db/p2p/p2p.go`

Add discovery message handler to process incoming network discovery requests:

```go
func (p *P2P) discoveryMessageHandler(from string, topic string, msg []byte) ([]byte, error) {
    req := &DiscoveryNetworkRequest{}
    if err := cbor.Unmarshal(msg, req); err != nil {
        return nil, err
    }
    req.SenderID = from

    // Execute discovery request on local node
    docIDs, err := p.executeDiscoveryRequest(p.ctx, *req)
    if err != nil {
        return nil, err
    }

    // Prepare response
    response := DiscoveryNetworkResponse{
        DocIDs:   docIDs,
        SenderID: p.host.ID(),
    }

    return cbor.Marshal(response)
}
```

### 6.2 Request Processing
**File**: `internal/db/p2p/p2p.go`

Add `executeDiscoveryRequest` method **to P2P struct** for processing network requests:

```go
func (p *P2P) executeDiscoveryRequest(ctx context.Context, req DiscoveryNetworkRequest) ([]string, error) {
    // Get collection by ID using GetCollections (similar to cli/collection_describe.go)
    cols, err := p.db.GetCollections(ctx, client.CollectionFetchOptions{
        CollectionID: immutable.Some(req.CollectionID),
    })
    if err != nil {
        return nil, err
    }
    if len(cols) == 0 {
        return nil, client.ErrCollectionNotFound
    }

    // Get collection name for GraphQL query
    collectionName := cols[0].Name().Value()

    // Build GraphQL query string from network request
    query := fmt.Sprintf(`
        query {
            %s(filter: %s, limit: %d, offset: %d) {
                _docID
            }
        }
    `, collectionName, buildFilterString(req.Filter), req.Limit, req.Offset)

    // Execute query using DB interface
    result := p.db.ExecRequest(ctx, query)
    if len(result.GQL.Errors) > 0 {
        return nil, result.GQL.Errors[0]
    }

    // Extract document IDs from result
    return extractDocIDsFromResult(result), nil
}
```

**Call Chain**: `P2P.discoveryMessageHandler()` → `P2P.executeDiscoveryRequest()` → `DB.ExecRequest()`

### 6.3 Discovery Event Handler
**File**: `internal/db/p2p/discovery.go`

Handle discovery events from local scan nodes:

```go
func (dp *DiscoveryProtocol) handleDiscoveryEvent(msg event.Message) error {
    req := msg.Data.(event.RequestDocDiscoveryEvent)

    // Send request to network peers
    networkReq := DiscoveryNetworkRequest{
        CollectionID:   req.CollectionID,
        Filter:         req.Filter,
        Limit:          req.Limit,
        Offset:         req.Offset,
        SenderID:       dp.p2p.host.ID(),
    }

    data, err := cbor.Marshal(networkReq)
    if err != nil {
        req.ResponseChan <- DocDiscoveryResult{Error: err}
        return err
    }

    // Publish to discovery topic and collect responses
    go dp.publishAndCollectResponses(req, data)

    return nil
}
```

### 6.4 Response Collection and Aggregation
**File**: `internal/db/p2p/discovery.go`

Implement multi-peer response consolidation using pubsub response channel pattern (similar to KMS):

```go
func (dp *DiscoveryProtocol) publishAndCollectResponses(req event.RequestDocDiscoveryEvent, data []byte) {
    // Publish to discovery topic with expectResponse=true to get response channel
    respChan, err := dp.p2p.host.PublishToTopic(dp.p2p.ctx, discoveryTopic, data, true)
    if err != nil {
        req.ResponseChan <- DocDiscoveryResult{Error: err}
        return
    }

    // Handle response collection asynchronously (similar to KMS pattern)
    go func() {
        dp.handleDiscoveryResponses(<-respChan, req)
    }()
}

func (dp *DiscoveryProtocol) handleDiscoveryResponses(responses [][]byte, req event.RequestDocDiscoveryEvent) {
    var allDocIDs []string
    seen := make(map[string]bool)

    // Process each peer response
    for _, responseData := range responses {
        var networkResp DiscoveryNetworkResponse
        if err := cbor.Unmarshal(responseData, &networkResp); err != nil {
            log.ErrorE("Failed to unmarshal discovery response", err)
            continue
        }

        // Deduplicate document IDs
        for _, docID := range networkResp.DocIDs {
            if !seen[docID] {
                seen[docID] = true
                allDocIDs = append(allDocIDs, docID)
            }
        }
    }

    // Send final aggregated response
    req.ResponseChan <- DocDiscoveryResult{DocIDs: allDocIDs}
}
```

## Complete Execution Flow

### 1. Client Initiates Discovery Query
```
Client → GraphQL: discover_User(filter: {age: {_gt: 18}}, limit: 10)
```

### 2. Query Processing Pipeline
```
GraphQL Parser → Detects "discover_" prefix → Sets IsDiscovery flag
Planner → Creates discoveryScanNode instead of regular scan
discoveryScanNode.Next() → Calls queryRemoteNodes()
```

### 3. Event Bus Communication
```
discoveryScanNode → Publishes RequestDocDiscoveryEvent with ResponseChan
DiscoveryProtocol → Subscribes to RequestDocDiscoveryEvent events
DiscoveryProtocol.handleDiscoveryEvent() → Processes event
```

### 4. Network Layer
```
DiscoveryProtocol → Marshals request → Publishes to P2P discovery topic
Each Peer → Receives via discoveryMessageHandler()
Each Peer → Calls P2P.executeDiscoveryRequest() → Uses collection ID with GetCollections() to get name
Each Peer → Returns DiscoveryNetworkResponse
```

### 5. Response Aggregation
```
DiscoveryProtocol → Collects all peer responses
DiscoveryProtocol.aggregateResponses() → Deduplicates DocIDs
DiscoveryProtocol → Sends final result to ResponseChan
```

### 6. Result Return
```
discoveryScanNode → Receives from ResponseChan
discoveryScanNode.Next() → Returns document with DocIDs array
GraphQL → Returns formatted response to client
```

## Phase 7: Testing Infrastructure

### 7.1 Integration Test Structure
**Directory**: `tests/integration/discover/`

Create comprehensive test coverage:
- `basic_test.go`: Basic discovery functionality
- `filter_test.go`: Filter parameter testing
- `pagination_test.go`: Limit/offset testing
- `network_test.go`: Multi-peer scenarios
- `timeout_test.go`: Network failure handling

### 7.2 Test Scenarios
- Single peer discovery
- Multi-peer response aggregation
- Filter parameter validation
- Pagination behavior
- Network timeout handling
- Empty result sets
- Large result sets

## Implementation Dependencies

### Prerequisites
- Understanding of existing SE implementation pattern
- Familiarity with DefraDB's event bus architecture
- Knowledge of P2P networking implementation

### Potential Blockers
- Event bus capacity and performance considerations
- P2P network topic management
- Query execution overhead on peer nodes
- Response aggregation complexity

## Success Metrics

- [ ] Discovery queries generate valid GraphQL schema
- [ ] Network requests successfully reach peer nodes
- [ ] Responses are properly aggregated and deduplicated
- [ ] Query execution performance meets expectations
- [ ] Integration tests achieve >90% coverage
- [ ] Feature works with existing DefraDB functionality

## Future Considerations

This implementation provides foundation for:
- Document content synchronization after discovery
- Priority-based peer selection
- Caching of discovery results
- Advanced filtering capabilities