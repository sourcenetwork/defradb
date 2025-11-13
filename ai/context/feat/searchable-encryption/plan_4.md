# Phase 4: Query API and Execution - Development Plan

## Overview
This phase implements the query interface for searchable encryption, enabling clients to search encrypted data on remote nodes without revealing the plaintext. The implementation extends DefraDB's existing query infrastructure to support SE-specific operations.

## Architecture

### Query Flow
1. **Client Query Submission**: User submits SE query through GraphQL/CLI/HTTP/client API
2. **Query Parsing**: GraphQL schema parses the `<Collection>_encrypted` query
3. **Tag Generation**: Client generates search tags from query values using SE key
4. **SE Query Request**: Client sends query request with search tags to replicator nodes
5. **Remote Execution**: Replicator nodes scan their SE artifact storage for matches
6. **Document ID Collection**: Matching document IDs are returned to client
7. **Document Retrieval**: For each document ID:
   - Check if document exists in local datastore
   - If not, publish document request on pubsub network
   - Nodes with the document respond with pushLog data
   - Process pushLog to trigger merge and populate local datastore
8. **Result Return**: Documents returned to user (already decrypted by normal fetch)

## Implementation Stages

### Stage 1: Client-Side Query API

#### 1.1 GraphQL Schema Generation (`internal/request/graphql/schema/generate.go`)

Generate encrypted query fields for each collection alongside regular query fields:

```go
// In Generator.Generate() after regular query fields
for _, t := range g.typeDefs {
    // Generate regular query field
    f, err := g.GenerateQueryInputForGQLType(ctx, t)
    if err != nil {
        return nil, err
    }
    generatedQueryFields = append(generatedQueryFields, f)
    
    // Generate encrypted query field if collection has encrypted indexes
    encryptedField, err := g.GenerateEncryptedQueryInputForGQLType(ctx, t, collections)
    if err != nil {
        return nil, err
    }
    if encryptedField != nil {
        queryType.AddFieldConfig(encryptedField.Name, encryptedField)
    }
}

// New method to generate encrypted query fields
func (g *Generator) GenerateEncryptedQueryInputForGQLType(
    ctx context.Context,
    obj *gql.Object,
    collections []client.CollectionDefinition,
) (*gql.Field, error) {
    // Find collection for this type
    var collection *client.CollectionDefinition
    for _, col := range collections {
        if col.Version.Name == obj.Name() {
            collection = &col
            break
        }
    }
    
    if collection == nil || len(collection.GetEncryptedIndexes()) == 0 {
        return nil, nil // No encrypted indexes, skip
    }
    
    // Create filter input that only supports _eq on encrypted fields
    filterName := obj.Name() + "EncryptedFilterArg"
    filterInput := g.genEncryptedFilterArgInput(obj, collection.GetEncryptedIndexes())
    g.manager.schema.TypeMap()[filterName] = filterInput
    
    // Create the encrypted query field
    field := &gql.Field{
        Name:        obj.Name() + "_encrypted",
        Description: "Query encrypted fields for " + obj.Name(),
        Type:        gql.NewList(obj),
        Args: gql.FieldConfigArgument{
            "filter": schemaTypes.NewArgConfig(filterInput, "Filter encrypted fields"),
            request.LimitClause:  schemaTypes.NewArgConfig(gql.Int, schemaTypes.LimitArgDescription),
            request.OffsetClause: schemaTypes.NewArgConfig(gql.Int, schemaTypes.OffsetArgDescription),
        },
    }
    
    return field, nil
}

// Generate filter input that only supports _eq on encrypted fields
func (g *Generator) genEncryptedFilterArgInput(
    obj *gql.Object,
    encryptedIndexes []client.EncryptedIndexDescription,
) *gql.InputObject {
    inputCfg := gql.InputObjectConfig{
        Name: obj.Name() + "EncryptedFilterArg",
    }
    
    fields := gql.InputObjectConfigFieldMap{}
    for _, encIdx := range encryptedIndexes {
        // Only support _eq operator for encrypted fields
        fields[encIdx.FieldName] = &gql.InputObjectFieldConfig{
            Type: g.manager.schema.TypeMap()[obj.Fields()[encIdx.FieldName].Type.Name() + "EqOperatorBlock"],
        }
    }
    
    inputCfg.Fields = fields
    return gql.NewInputObject(inputCfg)
}
```

#### 1.2 Request Mapping (`internal/request/graphql/parser/query.go`)

Map encrypted queries to internal representation:

```go
// Add new request type for encrypted queries
type EncryptedSelect struct {
    Select
    IsEncrypted bool
}
```

### Stage 2: Query Planning and Execution

#### 2.1 SE Query Plan Node (`internal/planner/se_scan.go`)

Create a new plan node for SE queries that implements the planNode interface:

```go
type seScanNode struct {
    documentIterator
    docMapper

    p              *Planner
    collection     client.Collection
    collectionID   string
    filter         *mapper.Filter
    encryptedIndexes []client.EncryptedIndexDescription
    
    // SE specific fields
    searchTags     map[string][]byte // fieldName -> searchTag
    remoteDocIDs   []string
    currentIndex   int
}

func (n *seScanNode) Kind() string { return "seScanNode" }

func (n *seScanNode) Init() error {
    // Initialize SE scan node
    return nil
}

func (n *seScanNode) Start() error {
    // 1. Generate search tags from filter
    if err := n.generateSearchTags(); err != nil {
        return err
    }
    
    // 2. Query remote nodes for matching doc IDs
    docIDs, err := n.queryRemoteNodes()
    if err != nil {
        return err
    }
    
    n.remoteDocIDs = docIDs
    n.currentIndex = 0
    
    return nil
}

func (n *seScanNode) generateSearchTags() error {
    n.searchTags = make(map[string][]byte)
    
    // Extract equality conditions from filter
    for fieldName, condition := range n.filter.Conditions {
        // Check if field has encrypted index
        var encIdx *client.EncryptedIndexDescription
        for _, idx := range n.encryptedIndexes {
            if idx.FieldName == fieldName {
                encIdx = &idx
                break
            }
        }
        
        if encIdx == nil {
            continue
        }
        
        // Extract value from condition (only _eq supported)
        value := condition["_eq"]
        if value == nil {
            return fmt.Errorf("only _eq operator supported for encrypted field %s", fieldName)
        }
        
        // Encode value and generate tag
        normalValue := client.NewNormalValue(value)
        encodedValue := encoding.EncodeFieldValue(nil, normalValue, false)
        
        tag, err := secore.GenerateEqualityTag(
            n.p.db.GetSEKey(), // Need to add method to get SE key
            n.collectionID,
            fieldName,
            encodedValue,
        )
        if err != nil {
            return err
        }
        
        n.searchTags[fieldName] = tag
    }
    
    return nil
}

func (n *seScanNode) queryRemoteNodes() ([]string, error) {
    // Get replicator nodes for this collection
    replicators := n.p.db.GetReplicators(n.collectionID)
    
    var allDocIDs []string
    docIDSet := make(map[string]struct{})
    
    // Query each replicator
    for _, replicator := range replicators {
        docIDs, err := n.querySEArtifacts(replicator)
        if err != nil {
            log.ErrorE("Failed to query SE artifacts", err)
            continue
        }
        
        // Deduplicate
        for _, docID := range docIDs {
            if _, exists := docIDSet[docID]; !exists {
                docIDSet[docID] = struct{}{}
                allDocIDs = append(allDocIDs, docID)
            }
        }
    }
    
    return allDocIDs, nil
}

func (n *seScanNode) Next() (bool, error) {
    if n.currentIndex >= len(n.remoteDocIDs) {
        return false, nil
    }
    
    // Get the next document ID
    docIDStr := n.remoteDocIDs[n.currentIndex]
    docID, err := client.NewDocIDFromString(docIDStr)
    if err != nil {
        n.currentIndex++
        return n.Next() // Skip invalid doc ID
    }
    
    // First, try to get the document from local store (in case it exists)
    doc, err := n.collection.Get(n.p.ctx, docID, false)
    if err == nil {
        // Document exists locally, use it
        n.currentValue = doc.ToMap()
        n.currentIndex++
        return true, nil
    }
    
    // Document not found locally, request it from the network
    // This will trigger a pushLog request and merge process
    if err := n.requestDocumentFromNetwork(docIDStr); err != nil {
        log.ErrorE("Failed to request document from network", err)
        n.currentIndex++
        return n.Next() // Skip this document
    }
    
    // Wait for document to be available after merge
    // In practice, this would use a more sophisticated mechanism
    // like channels or callbacks to know when the merge is complete
    if err := n.waitForDocument(docID); err != nil {
        n.currentIndex++
        return n.Next() // Skip if document never arrives
    }
    
    // Now try to fetch the document again
    doc, err = n.collection.Get(n.p.ctx, docID, false)
    if err != nil {
        n.currentIndex++
        return n.Next() // Skip if still not available
    }
    
    n.currentValue = doc.ToMap()
    n.currentIndex++
    
    return true, nil
}

func (n *seScanNode) requestDocumentFromNetwork(docID string) error {
    // This would call the network layer to request the document
    // via pubsub, similar to how encryption keys are requested
    return n.p.db.RequestDocumentFromNetwork(n.p.ctx, docID)
}

func (n *seScanNode) waitForDocument(docID client.DocID) error {
    // This is a simplified version - in practice, we'd use
    // proper synchronization mechanisms
    ctx, cancel := context.WithTimeout(n.p.ctx, 5*time.Second)
    defer cancel()
    
    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            // Check if document is now available
            if _, err := n.collection.Get(n.p.ctx, docID, false); err == nil {
                return nil
            }
        }
    }
}

func (n *seScanNode) Prefixes(prefixes []keys.Walkable) {}
func (n *seScanNode) Source() planNode { return nil }
func (n *seScanNode) Close() error { return nil }
```

#### 2.2 Planner Integration (`internal/planner/planner.go`)

Add SE query planning to MakePlan:

```go
// In Planner.MakePlan or appropriate location
func (p *Planner) SelectEncrypted(mapper *mapper.Select) (planNode, error) {
    col, err := p.db.GetCollectionByName(p.ctx, mapper.CollectionName)
    if err != nil {
        return nil, err
    }
    
    encryptedIndexes := col.GetEncryptedIndexes()
    if len(encryptedIndexes) == 0 {
        return nil, errors.New("collection has no encrypted indexes")
    }
    
    seScan := &seScanNode{
        p:                p,
        collection:       col,
        collectionID:     col.ID(),
        filter:          mapper.Filter,
        encryptedIndexes: encryptedIndexes,
        documentMapping:  mapper.DocumentMapping,
    }
    
    return seScan, nil
}
```

### Stage 3: Network Protocol and Document Retrieval

The SE query process involves two phases:
1. Query untrusted nodes for matching document IDs
2. Request the actual documents from the P2P network (since local datastore is empty)

#### 3.1 SE Artifact Query (`net/client.go`)

Query SE artifacts on remote replicator nodes:

```go
// querySEArtifacts queries SE artifacts on a remote node
func (s *server) querySEArtifacts(ctx context.Context, pid peer.ID, req querySEArtifactsRequest) (*querySEArtifactsReply, error) {
    client, err := s.dial(pid)
    if err != nil {
        return nil, NewErrQuerySEArtifacts(err)
    }
    
    ctx, cancel := context.WithTimeout(ctx, PullTimeout)
    defer cancel()
    
    resp := &querySEArtifactsReply{}
    if err := client.Invoke(ctx, serviceQuerySEArtifactsName, req, resp); err != nil {
        return nil, NewErrQuerySEArtifacts(err,
            errors.NewKV("CollectionID", req.CollectionID),
            errors.NewKV("PeerID", pid),
        )
    }
    
    return resp, nil
}
```

#### 3.1.2 Document Request via PubSub (`net/pubsub_doc_request.go`)

Since the local datastore is empty, we need to request documents from the network:

```go
// RequestDocumentFromNetwork requests a document from the P2P network via pubsub
func (s *server) RequestDocumentFromNetwork(ctx context.Context, docID string) error {
    // Create a document request message
    req := &DocRequest{
        DocID:        docID,
        RequesterID:  s.peer.ID(),
        Timestamp:    time.Now().Unix(),
    }
    
    data, err := cbor.Marshal(req)
    if err != nil {
        return errors.Wrap("failed to marshal doc request", err)
    }
    
    // Send request on document topic
    topic := getDocumentTopic(docID)
    respChan, err := s.pubsub.SendPubSubMessage(ctx, topic, data)
    if err != nil {
        return errors.Wrap("failed to publish doc request", err)
    }
    
    // Handle response asynchronously
    go s.handleDocumentResponse(ctx, <-respChan, docID)
    
    return nil
}

// handleDocumentResponse processes incoming document data
func (s *server) handleDocumentResponse(ctx context.Context, resp []byte, docID string) {
    var docResp DocResponse
    if err := cbor.Unmarshal(resp, &docResp); err != nil {
        log.ErrorE("Failed to unmarshal doc response", err)
        return
    }
    
    // Process the pushLog to trigger merge
    if err := s.processPushLog(ctx, docResp.PushLog); err != nil {
        log.ErrorE("Failed to process pushLog", err)
        return
    }
}
```

#### 3.2 Server Side (`net/server.go`)

Add SE query handler:

```go
// querySEArtifactsHandler handles SE queries from peers
func (s *server) querySEArtifactsHandler(ctx context.Context, req *querySEArtifactsRequest) (*querySEArtifactsReply, error) {
    pid, err := peerIDFromContext(ctx)
    if err != nil {
        return nil, err
    }
    
    log.InfoContext(ctx, "Received SE query",
        corelog.Any("PeerID", pid.String()),
        corelog.Any("CollectionID", req.CollectionID),
        corelog.Any("QueryCount", len(req.Queries)))
    
    ds := s.peer.db.Datastore()
    matchingDocIDs := make(map[string]struct{})
    
    // For each field query, find matching artifacts
    for _, query := range req.Queries {
        key := keys.DatastoreSE{
            CollectionID: req.CollectionID,
            IndexID:      query.IndexID,
            SearchTag:    query.SearchTag,
        }
        
        // Scan for matching keys
        iter, err := ds.Iterator(ctx, corekv.IterOptions{
            Prefix: key.Bytes(),
        })
        if err != nil {
            return nil, err
        }
        
        for {
            hasNext, err := iter.Next()
            if err != nil || !hasNext {
                break
            }
            
            // Extract DocID from key
            dsKey, err := keys.NewDatastoreSEFromString(string(iter.Key()))
            if err != nil {
                continue
            }
            
            matchingDocIDs[dsKey.DocID] = struct{}{}
        }
        iter.Close()
    }
    
    // Convert to slice
    docIDs := make([]string, 0, len(matchingDocIDs))
    for docID := range matchingDocIDs {
        docIDs = append(docIDs, docID)
    }
    
    return &querySEArtifactsReply{
        DocIDs: docIDs,
    }, nil
}
```

#### 3.3 GRPC Protocol (`net/grpc.go`)

```go
// SE query messages
type querySEArtifactsRequest struct {
    CollectionID string
    Queries      []seFieldQuery
}

type seFieldQuery struct {
    FieldName string
    IndexID   string
    SearchTag []byte
}

type querySEArtifactsReply struct {
    DocIDs []string
}

// Document request messages (for pubsub)
type DocRequest struct {
    DocID       string
    RequesterID peer.ID
    Timestamp   int64
}

type DocResponse struct {
    DocID   string
    PushLog pushLogRequest // Reuse existing pushLog structure
    Error   string
}

// Service registration
const serviceQuerySEArtifactsName = "/defradb.net/querySEArtifacts"

// In registerServices()
s.registerUnaryHandler(serviceQuerySEArtifactsName, s.querySEArtifactsHandler)
```

#### 3.4 Document Retrieval Flow

The complete SE query flow:

1. **Client generates search tags** from query filter using SE key
2. **Query replicator nodes** for matching document IDs via gRPC
3. **For each document ID returned**:
   - Check if document exists locally (datastore)
   - If not, publish document request on pubsub
   - Nodes with the document respond with pushLog
   - Process pushLog to trigger merge
   - Wait for merge completion
   - Fetch document from local datastore
4. **Return documents to client**

### Stage 4: Integration with ExecRequest

#### 4.1 Request Processing (`internal/db/request.go`)

Integrate SE queries into the existing ExecRequest flow:

```go
// In db.execRequest() or appropriate handler
func (db *DB) execRequest(ctx context.Context, request string, options *client.GQLOptions) *client.RequestResult {
    // Parse request
    req, err := parser.ParseRequest(request)
    if err != nil {
        return &client.RequestResult{GQL: client.GQLResult{Errors: []error{err}}}
    }
    
    // Check if this is an encrypted query
    if isEncryptedQuery(req) {
        return db.execEncryptedRequest(ctx, req, options)
    }
    
    // Continue with normal request processing
    return db.execNormalRequest(ctx, req, options)
}

func isEncryptedQuery(req *request.Request) bool {
    // Check if any selection ends with "_encrypted"
    for _, query := range req.Queries {
        for _, selection := range query.Selections {
            if strings.HasSuffix(selection.Name, "_encrypted") {
                return true
            }
        }
    }
    return false
}
```

## Error Handling

1. **Field Not Encrypted**: Return clear error when querying non-encrypted field
2. **Unsupported Operators**: Only `_eq` supported, reject others with clear message
3. **Network Failures**: Continue with other nodes, return partial results
4. **Missing Documents**: Skip deleted documents, continue with available ones

## Security Considerations

1. **Tag Generation**: Ensure SE key is properly protected
2. **Access Control**: Verify requesting peer has query permissions
3. **Result Validation**: Only return doc IDs, full document fetch uses normal access control
4. **Rate Limiting**: Implement query rate limits per peer

## Testing Strategy

### Unit Tests
- Tag generation from filter conditions
- Query plan node implementation
- Network message serialization

### Integration Tests
- End-to-end SE query through GraphQL
- Multi-node query execution
- Partial failure handling
- Performance with large result sets

## Current Status

Phase 4 implementation is partially complete:

### Completed:
1. **Stage 1: Client-Side Query API** ✅
   - GraphQL schema generation for `<Collection>_encrypted` queries (implemented in `generate.go`)
   - Filter input that only supports `_eq` operator on encrypted fields
   - Request mapping with `IsEncrypted` flag added to `request.Select` and `mapper.Select`

2. **Stage 2: Query Planning and Execution** ✅ 
   - Created `seScanNode` in `internal/planner/se_scan.go` implementing the planNode interface
   - Added `SelectEncrypted` method to planner
   - Modified `Select` method to route encrypted queries to SE planning
   - Added `IsEncrypted` field to `mapper.Select` struct

### Completed:
3. **Stage 3: Network Protocol** ✅
   - Client-side SE query method implemented in `net/client.go`
   - Server-side SE query handler implemented in `net/server.go`
   - GRPC protocol messages defined in `net/grpc.go`
   - Basic structure in place, but actual datastore querying needs implementation

### Completed:
4. **Stage 3.4: Document Request via PubSub** ✅
   - Implemented document request mechanism via pubsub with dedicated SE topics
   - Added `DocUpdateRequest` and `DocUpdateResponse` event types
   - Created `handleDocUpdateRequest` in `peer.go` to publish requests to dedicated `se-doc-request/{collectionID}` topics
   - Added `docUpdateMessageHandler` in `server.go` to handle incoming requests
   - Modified `updatePubSubTopics` to automatically create SE document request topics when collections are added
   - Integrated with existing pubsub infrastructure
   - CollectionID and SchemaRoot are confirmed to be the same value, ensuring topic consistency

### In Progress:
5. **Stage 4: ExecRequest Integration** 📋
   - Need to hook into existing query flow

### Key TODOs:
1. **SE Key Access**: Add method to access SE key from Store interface
2. **Replicator Methods**: Add `GetReplicators` method to Store interface
3. **Network Implementation**: Complete the querySEArtifacts methods
4. **Tag Generation**: Currently stubbed out due to SE key access limitation

## Implementation Notes

### Stage 1 Implementation Details:
- `GenerateEncryptedQueryInputForGQLType` method creates encrypted query fields
- `genEncryptedFilterArgInput` creates filter input with only `_eq` operator support
- Uses field thunks to ensure proper field resolution
- Parser updated to detect encrypted queries by `_encrypted` suffix

### Stage 2 Implementation Details:
- `seScanNode` created with stub implementations for:
  - `generateSearchTags()` - needs SE key access
  - `queryRemoteNodes()` - needs replicator access and network protocol
  - `Next()` - updated to handle empty local datastore:
    - First tries local datastore
    - If not found, requests document from P2P network via pubsub
    - Waits for merge completion before accessing document
- Filter processing handles `ObjectProperty` keys and extracts `_eq` conditions
- Collection ID accessed via `col.Version().CollectionID`

### Stage 3 Key Design Changes:
- Documents are not assumed to be locally available
- Document retrieval uses pubsub mechanism similar to encryption key requests
- PushLog processing triggers merge to populate local datastore
- Synchronization mechanism needed to wait for merge completion

### Stage 3 Implementation Details:
- `querySEArtifacts` method added to `net/client.go` for client-side SE queries
- `querySEArtifactsHandler` method added to `net/server.go` for handling SE queries
- `NewErrQuerySEArtifacts` error type added to `net/errors.go`
- GRPC protocol messages already defined in `net/grpc.go` with handler registration
- Actual datastore querying logic is stubbed out pending SE key structure implementation
- Document request via pubsub implemented with dedicated SE topics
- Document response handler (`docUpdateMessageHandler`) currently returns empty response due to DB interface limitations - needs access to headstore/transactions to return actual pushLogRequest data

## Dependencies

- Phase 3 artifact storage on replicator nodes
- Access to SE encryption key in DB instance (need to add to Store interface)
- Existing query planner infrastructure
- P2P network for document retrieval
- Replicator information for collection (need to add to Store interface)