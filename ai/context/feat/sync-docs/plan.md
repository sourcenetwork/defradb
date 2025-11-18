# Document Synchronization Feature Development Plan

## Overview

Implement a document synchronization feature that allows nodes to request and sync specific documents from the network on-demand. This feature enables nodes to discover and retrieve documents they don't have locally by requesting latest heads from peers and automatically synchronizing the complete document DAG.

## Design Principles
- **Event-driven**: Use DefraDB's existing event system for coordination
- **Non-blocking**: Asynchronous operation with configurable timeout support
- **Batch-efficient**: Support multiple document requests in a single network call
- **Auto-subscribe**: Automatically subscribe to collections after successful sync
- **Multi-response**: Handle multiple responses from different peers

## Implementation Plan

### Phase 1: Core Infrastructure (Files to Create/Modify)

#### 1.1 Event System (`event/event.go`)
```go
// Add new event types for document synchronization
type DocSyncRequest struct {
    CollectionID string
    DocIDs       []string      // Support batch requests
    Timeout      time.Duration // Timeout for the operation
    Response     chan DocSyncResponse
}

type DocSyncResponse struct {
    Results map[string]DocSyncResult // Map of DocID -> Result
    Sender  string                   // Peer ID of the responder
    Error   error
}

type DocSyncResult struct {
    Head   cid.Cid // Latest head CID for the document
    Height uint64  // Block height to handle multiple responses with same docID
}

const (
    DocSyncRequestName  = "doc-sync-request"
    DocSyncTopic       = "doc-sync" // Fixed topic for document sync
)
```

#### 1.2 Network Protocol (`net/grpc.go`)
```go
// Add protocol structures for document sync
type docSyncRequest struct {
    CollectionID string   `json:"collectionID"`
    DocIDs       []string `json:"docIDs"`
}

type docSyncReply struct {
    Results map[string]docSyncItem `json:"results"`
    Sender  string                 `json:"sender"`
}

type docSyncItem struct {
    DocID        string `json:"docID"`
    CID          []byte `json:"cid,omitempty"`    // Empty if document not found
    Height       uint64 `json:"height,omitempty"` // Block height
    CollectionID string `json:"collectionID"`
}
```

#### 1.3 P2P Interface (`client/p2p.go`)
```go
// Add to P2P interface
type P2P interface {
    // ... existing methods ...
    
    // SyncDocuments requests the latest versions of specified documents from the network
    // and synchronizes their DAGs locally. After successful sync, automatically subscribes
    // to the documents and their collection for future updates.
    //
    // Parameters:
    // - ctx: Context for the operation  
    // - collectionID: The collection containing the documents
    // - docIDs: List of document IDs to synchronize
    // - opts: Optional parameters including timeout
    //
    // Returns a map of document ID to sync result with head CIDs and heights.
    SyncDocuments(
        ctx context.Context,
        collectionID string,
        docIDs []string,
        opts ...DocSyncOption,
    ) (map[string]DocSyncResult, error)
}

// DocSyncOption configures the document sync operation
type DocSyncOption func(*DocSyncOptions)

// DocSyncOptions contains options for document sync operations
type DocSyncOptions struct {
    Timeout time.Duration
}

// DocSyncWithTimeout sets the timeout for the sync operation
func DocSyncWithTimeout(timeout time.Duration) DocSyncOption {
    return func(opts *DocSyncOptions) {
        opts.Timeout = timeout
    }
}

// DocSyncResult represents the result of synchronizing a single document
type DocSyncResult struct {
    Head   string // Latest head CID of the document (empty if not found)
    Height uint64 // Block height to prioritize newer responses
    Sender string // ID of the peer that provided the document
}
```

### Phase 2: Core Implementation

#### 2.1 Network Server Implementation (`net/server.go`)
```go
// Add document sync request handler
func (s *server) handleDocSyncRequest(req event.DocSyncRequest) {
    pubsubReq := &docSyncRequest{
        CollectionID: req.CollectionID,
        DocIDs:       req.DocIDs,
    }

    data, err := cbor.Marshal(pubsubReq)
    if err != nil {
        s.sendDocSyncError(req.Response, err)
        return
    }

    // Use WithMultiResponse to handle multiple peer responses
    respChan, err := s.SendPubSubMessage(
        s.peer.ctx, 
        event.DocSyncTopic, 
        data,
        rpc.WithMultiResponse(true),
    )
    if err != nil {
        s.sendDocSyncError(req.Response, err)
        return
    }

    go s.processDocSyncResponses(req, respChan)
}

// Add pubsub message handler for incoming document sync requests  
func (s *server) docSyncMessageHandler(from libpeer.ID, topic string, msg []byte) ([]byte, error) {
    req := &docSyncRequest{}
    if err := cbor.Unmarshal(msg, req); err != nil {
        return nil, err
    }

    results := make(map[string]docSyncItem)
    
    for _, docID := range req.DocIDs {
        result, err := s.processDocSyncItem(req.CollectionID, docID)
        if err != nil {
            log.ErrorContextE(s.peer.ctx, "Failed to process doc sync item", err, 
                logging.NewKV("DocID", docID), 
                logging.NewKV("CollectionID", req.CollectionID))
            continue // Skip failed items
        }
        results[docID] = result
    }

    reply := &docSyncReply{
        Results: results,
        Sender:  s.peer.host.ID().String(),
    }

    return cbor.Marshal(reply)
}

// Process individual document sync items
func (s *server) processDocSyncItem(collectionID, docID string) (docSyncItem, error) {
    cols, err := s.peer.db.GetCollections(s.peer.ctx, client.CollectionFetchOptions{
        CollectionID: immutable.Some(collectionID),
    })
    
    if err != nil {
        return docSyncItem{}, fmt.Errorf("failed to get collection %s: %w", collectionID, err)
    }
    
    if len(cols) == 0 {
        return docSyncItem{}, fmt.Errorf("collection %s not found", collectionID)
    }

    col := cols[0]
    docIDParsed, err := client.NewDocIDFromString(docID)
    if err != nil {
        return docSyncItem{}, fmt.Errorf("failed to parse docID %s: %w", docID, err)
    }

    doc, err := col.Get(s.peer.ctx, docIDParsed, false)
    if err != nil {
        return docSyncItem{}, fmt.Errorf("failed to get document %s: %w", docID, err)
    }

    return docSyncItem{
        DocID:        docID,
        CID:          doc.Head().Bytes(),
        Height:       doc.Height(),
        CollectionID: collectionID,
    }, nil
}

// Process multiple responses from different peers
func (s *server) processDocSyncResponses(req event.DocSyncRequest, respChan <-chan rpc.Response) {
    timeout := req.Timeout
    if timeout == 0 {
        timeout = 30 * time.Second // Default timeout
    }
    
    ctx, cancel := context.WithTimeout(s.peer.ctx, timeout)
    defer cancel()

    collectedResults := make(map[string]DocSyncResult)
    
    for {
        select {
        case resp := <-respChan:
            if resp.Err != nil {
                log.ErrorContextE(ctx, "Received error response from peer", resp.Err)
                continue // Skip failed responses
            }
            
            if len(resp.Data) > 0 {
                var reply docSyncReply
                if err := cbor.Unmarshal(resp.Data, &reply); err != nil {
                    log.ErrorContextE(ctx, "Failed to unmarshal doc sync reply", err)
                    continue
                }

                // Process each document in the response
                for docID, item := range reply.Results {
                    if len(item.CID) > 0 { // Document found
                        _, docCid, err := cid.CidFromBytes(item.CID)
                        if err != nil {
                            log.ErrorContextE(ctx, "Failed to parse CID from bytes", err, 
                                logging.NewKV("DocID", docID))
                            continue
                        }

                        // Check if we already have a result for this docID with higher height
                        if existing, exists := collectedResults[docID]; exists && existing.Height >= item.Height {
                            continue // Skip lower height responses
                        }

                        // Sync the DAG for this document
                        err = s.syncDocumentDAG(ctx, docCid)
                        if err != nil {
                            log.ErrorContextE(ctx, "Failed to sync DAG for document", err, 
                                logging.NewKV("DocID", docID),
                                logging.NewKV("CID", docCid.String()))
                            continue
                        }

                        // Subscribe to document and collection after successful sync
                        s.subscribeToDocument(req.CollectionID, docID)

                        collectedResults[docID] = DocSyncResult{
                            Head:   docCid.String(),
                            Height: item.Height,
                            Sender: reply.Sender,
                        }
                    }
                }
            }
            
        case <-ctx.Done():
            // Send collected results back
            req.Response <- event.DocSyncResponse{
                Results: collectedResults,
                Sender:  s.peer.host.ID().String(),
                Error:   nil,
            }
            return
        }
    }
}

// Sync document DAG - implementation from net/server.go#L730-759
func (s *server) syncDocumentDAG(ctx context.Context, docCid cid.Cid) error {
    blockStore := &bsrvadapter.Adapter{Wrapped: s.peer.blockService}

    linkSys := cidlink.DefaultLinkSystem()
    linkSys.SetReadStorage(blockStore)
    linkSys.TrustedStorage = true

    nd, err := linkSys.Load(linking.LinkContext{Ctx: ctx}, cidlink.Link{Cid: docCid}, coreblock.BlockSchemaPrototype)
    if err != nil {
        return fmt.Errorf("failed to load document node: %w", err)
    }
    
    linkBlock, err := coreblock.GetFromNode(nd)
    if err != nil {
        return fmt.Errorf("failed to get block from node: %w", err)
    }

    err = syncDAG(ctx, s.peer.blockService, linkBlock)
    if err != nil {
        return fmt.Errorf("failed to sync DAG: %w", err)
    }
    
    return nil
}

// Subscribe to document updates after successful sync
func (s *server) subscribeToDocument(collectionID, docID string) {
    // Check if already subscribed to collection topic
    collectionTopic := collectionID
    if !s.hasPubSubTopicAndSubscribed(collectionTopic) {
        _, err := s.addPubSubTopic(collectionTopic, true, nil)
        if err != nil {
            log.ErrorContextE(s.peer.ctx, "Failed to subscribe to collection topic", err,
                logging.NewKV("CollectionID", collectionID))
        }
    }
    
    // Check if already subscribed to document-specific topic  
    docTopic := docID
    if !s.hasPubSubTopicAndSubscribed(docTopic) {
        _, err := s.addPubSubTopic(docTopic, true, nil)
        if err != nil {
            log.ErrorContextE(s.peer.ctx, "Failed to subscribe to document topic", err,
                logging.NewKV("DocID", docID))
        }
    }
}
```

#### 2.2 P2P Implementation (`node/node_p2p.go`)
```go
// Implement SyncDocuments method for P2P interface
func (n *Node) SyncDocuments(
    ctx context.Context,
    collectionID string,
    docIDs []string,
    opts ...client.DocSyncOption,
) (map[string]client.DocSyncResult, error) {
    // Apply options with defaults
    options := &client.DocSyncOptions{
        Timeout: 30 * time.Second, // Default timeout
    }
    for _, opt := range opts {
        opt(options)
    }

    responseChan := make(chan event.DocSyncResponse, 1)
    defer close(responseChan)

    request := event.DocSyncRequest{
        CollectionID: collectionID,
        DocIDs:       docIDs,
        Timeout:      options.Timeout,
        Response:     responseChan,
    }

    n.events.Publish(event.NewMessage(event.DocSyncRequestName, request))

    timeoutCtx, cancel := context.WithTimeout(ctx, options.Timeout)
    defer cancel()

    select {
    case response := <-responseChan:
        if response.Error != nil {
            return nil, response.Error
        }
        
        // Convert internal results to client results
        results := make(map[string]client.DocSyncResult, len(response.Results))
        for docID, result := range response.Results {
            results[docID] = client.DocSyncResult{
                Head:   result.Head.String(),
                Height: result.Height,
                Sender: result.Sender,
            }
        }
        return results, nil
        
    case <-timeoutCtx.Done():
        return nil, fmt.Errorf("timeout waiting for document sync: %w", timeoutCtx.Err())
    }
}
```

### Phase 3: Client Implementations

#### 3.1 HTTP Client (`http/client_p2p.go`)
```go
func (c *P2P) SyncDocuments(
    ctx context.Context,
    collectionID string,
    docIDs []string,
    opts ...client.DocSyncOption,
) (map[string]client.DocSyncResult, error) {
    // Apply options with defaults
    options := &client.DocSyncOptions{
        Timeout: 30 * time.Second, // Default timeout
    }
    for _, opt := range opts {
        opt(options)
    }

    req := map[string]any{
        "collectionID": collectionID,
        "docIDs":       docIDs,
        "timeout":      options.Timeout.String(),
    }

    response := make(map[string]client.DocSyncResult)
    if err := c.client.Request(ctx, &response, "POST", "p2p/sync/documents", req); err != nil {
        return nil, err
    }

    return response, nil
}
```

#### 3.2 HTTP Handler (`http/handler_p2p.go`)
```go
func (h *P2PHandler) syncDocuments(w http.ResponseWriter, r *http.Request) {
    var req struct {
        CollectionID string        `json:"collectionID"`
        DocIDs       []string      `json:"docIDs"`
        Timeout      string        `json:"timeout"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    var opts []client.DocSyncOption
    if req.Timeout != "" {
        timeout, err := time.ParseDuration(req.Timeout)
        if err != nil {
            http.Error(w, fmt.Sprintf("invalid timeout format: %v", err), http.StatusBadRequest)
            return
        }
        opts = append(opts, client.DocSyncWithTimeout(timeout))
    }

    results, err := h.node.SyncDocuments(r.Context(), req.CollectionID, req.DocIDs, opts...)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    h.writeJSON(w, results)
}
```

#### 3.3 CLI Client (`cli/p2p_sync.go`)
```go
// Add to existing cli/p2p.go file
var p2pSyncCmd = &cobra.Command{
    Use:   "sync",
    Short: "P2P document synchronization commands",
}

var p2pSyncDocsCmd = &cobra.Command{
    Use:     "docs [collection-id] [doc-id...]",
    Short:   "Synchronize specific documents from the network",
    Aliases: []string{"documents"},
    Args:    cobra.MinimumNArgs(2),
    RunE: func(cmd *cobra.Command, args []string) error {
        collectionID := args[0]
        docIDs := args[1:]
        
        var opts []client.DocSyncOption
        if timeout, _ := cmd.Flags().GetDuration("timeout"); timeout > 0 {
            opts = append(opts, client.DocSyncWithTimeout(timeout))
        }
        
        p2p := mustGetContextP2P(cmd)
        results, err := p2p.SyncDocuments(cmd.Context(), collectionID, docIDs, opts...)
        if err != nil {
            return err
        }

        return writeJSON(cmd, results)
    },
}

func init() {
    p2pSyncDocsCmd.Flags().Duration("timeout", 0, "Timeout for sync operations")
    p2pSyncCmd.AddCommand(p2pSyncDocsCmd)
    p2pCmd.AddCommand(p2pSyncCmd)
}
```

#### 3.4 JS Client (`js/client_p2p.go`)
```go
func (c *P2P) SyncDocuments(this js.Value, args []js.Value) any {
    collectionID := args[0].String()
    docIDsJS := args[1]
    options := args[2] // Optional options object
    
    // Convert JS array to Go slice
    docIDs := make([]string, docIDsJS.Length())
    for i := 0; i < docIDsJS.Length(); i++ {
        docIDs[i] = docIDsJS.Index(i).String()
    }
    
    // Parse options
    var opts []client.DocSyncOption
    if !options.IsUndefined() && !options.IsNull() {
        if timeoutValue := options.Get("timeout"); !timeoutValue.IsUndefined() {
            timeoutSeconds := timeoutValue.Int()
            timeout := time.Duration(timeoutSeconds) * time.Second
            opts = append(opts, client.DocSyncWithTimeout(timeout))
        }
    }
    
    promise := js.Global().Get("Promise")
    return promise.New(js.FuncOf(func(this js.Value, args []js.Value) any {
        resolve := args[0]
        reject := args[1]
        
        go func() {
            results, err := c.node.SyncDocuments(context.Background(), collectionID, docIDs, opts...)
            if err != nil {
                reject.Invoke(err.Error())
                return
            }
            
            // Convert results to JS object
            jsResults := js.ValueOf(map[string]any{})
            for docID, result := range results {
                jsResults.Set(docID, map[string]any{
                    "head":   result.Head,
                    "height": result.Height,
                    "sender": result.Sender,
                })
            }
            
            resolve.Invoke(jsResults)
        }()
        
        return nil
    }))
}
```

### Phase 4: Testing Strategy

#### 4.1 Integration Tests Structure
```
tests/integration/sync_doc/
├── simple_test.go              # Basic single document sync
├── batch_test.go               # Multiple document sync
├── timeout_test.go             # Timeout handling
├── missing_doc_test.go         # Document not found scenarios
├── network_failure_test.go     # Network error scenarios
└── encryption_test.go          # Document sync with encryption
```

#### 4.2 Test Implementation (`tests/integration/sync_doc/simple_test.go`)
```go
func TestDocSync_WithSingleDocument_ShouldSync(t *testing.T) {
    test := testUtils.TestCase{
        Actions: []any{
            testUtils.RandomNetworkingConfig(),
            testUtils.RandomNetworkingConfig(),
            testUtils.SchemaUpdate{
                Schema: `
                    type Users {
                        Name: String
                        Age: Int
                    }
                `,
            },
            testUtils.CreateDoc{
                NodeID: immutable.Some(0),
                Doc: `{
                    "Name": "John",
                    "Age": 21
                }`,
            },
            // Connect peers for P2P communication
            testUtils.ConnectPeers{
                SourceNodeID: 0,
                TargetNodeID: 1,
            },
            // Test document sync using dedicated test action
            testUtils.SyncDocs{
                NodeID:       immutable.Some(1),
                CollectionID: 0,
                DocIDs:       []int{0}, // Reference to created document
                Timeout:      30 * time.Second,
                ExpectedResults: map[string]testUtils.DocSyncResult{
                    // Will be populated by test framework with actual docID
                    "*": {
                        HasHead:   true,
                        HasSender: true,
                        MinHeight: 1,
                    },
                },
            },
            // Verify document is now available on node 1
            testUtils.Request{
                NodeID: immutable.Some(1),
                Request: `query {
                    Users {
                        Name
                        Age
                    }
                }`,
                Results: map[string]any{
                    "Users": []map[string]any{
                        {
                            "Name": "John",
                            "Age":  int64(21),
                        },
                    },
                },
            },
        },
    }

    testUtils.ExecuteTestCase(t, test)
}
```

#### 4.3 Test Action Definition (`tests/integration/test_case.go`)
```go
// SyncDocs will synchronize documents from the network via P2P
type SyncDocs struct {
    // NodeID may hold the ID (index) of a node to execute the sync on.
    NodeID immutable.Option[int]

    // The collection containing the documents to sync.
    CollectionID int

    // The indices of documents to sync (references to previously created documents).
    // Uses the same DocIndex pattern as other test actions - these will be resolved 
    // to actual document IDs at runtime by the test framework.
    DocIDs []int

    // Timeout for the sync operation.
    Timeout time.Duration

    // Expected results of the sync operation.
    // Key "*" can be used as a wildcard to match any document ID
    ExpectedResults map[string]DocSyncResult

    // Any error expected from the action.
    ExpectedError string
}

// DocSyncResult represents expected sync results for testing
type DocSyncResult struct {
    HasHead   bool   // Whether a head CID should be present
    HasSender bool   // Whether a sender should be present  
    MinHeight uint64 // Minimum expected height
}
```

#### 4.4 Test Action Handler (`tests/integration/utils.go`)
```go
// Add to performAction function switch statement
case SyncDocs:
    syncDocs(s, action)

// Handler function for SyncDocs test action
func syncDocs(s *state, action SyncDocs) {
    nodeIDs, nodes := getNodesWithIDs(action.NodeID, s.nodes)
    
    var expectedErrorRaised bool
    
    for index, node := range nodes {
        nodeID := nodeIDs[index]
        
        // Convert document indices to actual document ID strings
        docIDStrings := make([]string, len(action.DocIDs))
        for i, docIndex := range action.DocIDs {
            docIDStrings[i] = s.docIDs[action.CollectionID][docIndex].String()
        }
        
        collectionIDString := s.nodes[nodeID].collections[action.CollectionID].ID().String()
        
        var opts []client.DocSyncOption
        if action.Timeout > 0 {
            opts = append(opts, client.DocSyncWithTimeout(action.Timeout))
        }
        
        // Execute the sync operation
        results, err := node.P2P().SyncDocuments(
            s.ctx,
            collectionIDString,
            docIDStrings,
            opts...,
        )
        
        expectedErrorRaised = AssertError(s.t, s.testCase.Description, err, action.ExpectedError)
        
        if !expectedErrorRaised && action.ExpectedResults != nil {
            // Validate the results against expectations
            for docIDString, expectedResult := range action.ExpectedResults {
                var actualResult client.DocSyncResult
                var found bool
                
                if docIDString == "*" {
                    // Wildcard match - use any result
                    for _, result := range results {
                        actualResult = result
                        found = true
                        break
                    }
                } else {
                    actualResult, found = results[docIDString]
                }
                
                require.True(s.t, found, "Expected result for docID %s not found", docIDString)
                
                if expectedResult.HasHead {
                    require.NotEmpty(s.t, actualResult.Head, "Expected non-empty head for docID %s", docIDString)
                } else {
                    require.Empty(s.t, actualResult.Head, "Expected empty head for docID %s", docIDString)
                }
                
                if expectedResult.HasSender {
                    require.NotEmpty(s.t, actualResult.Sender, "Expected non-empty sender for docID %s", docIDString)
                } else {
                    require.Empty(s.t, actualResult.Sender, "Expected empty sender for docID %s", docIDString)
                }
                
                if expectedResult.MinHeight > 0 {
                    require.GreaterOrEqual(s.t, actualResult.Height, expectedResult.MinHeight, 
                        "Expected height >= %d for docID %s", expectedResult.MinHeight, docIDString)
                }
            }
        }
    }
    
    assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)
}
```

This approach:
- Adds a dedicated test action to the framework for document sync testing
- Follows existing patterns used by other test actions like `CreateDoc` and `UpdateDoc`
- Uses the same `DocIndex` resolution pattern that converts test indices to actual document IDs at runtime
- Allows comprehensive testing of sync functionality with wildcard matching
- Supports both positive and negative test cases with detailed result validation

### Phase 5: Advanced Features

#### 5.1 Batch Optimization
- Implement efficient batching for multiple document requests
- Add configurable batch size limits
- Implement parallel processing for large batches

#### 5.2 Caching and Deduplication
- Cache recent sync requests to avoid redundant network calls
- Implement request deduplication for concurrent requests
- Add metrics for cache hit/miss rates

#### 5.3 Peer Selection Strategy
- Implement intelligent peer selection based on response times
- Add fallback mechanisms for failed peers
- Include peer reputation tracking

## Risk Assessment and Mitigation

### Technical Risks
1. **Network Partitioning**: Use configurable timeouts and retry mechanisms
2. **Memory Usage**: Implement streaming for large documents and batch size limits
3. **Security**: Validate all network inputs and implement rate limiting

## Success Criteria

1. **Functional**: P2P interface can successfully sync documents from the network
2. **Performance**: Sync operations complete within reasonable timeouts (< 30s for normal cases)
3. **Reliability**: Graceful handling of network failures and missing documents
4. **Testing**: Comprehensive test coverage using custom function actions
5. **Multi-response**: Properly handle multiple peer responses using WithMultiResponse
6. **Auto-subscribe**: Automatically subscribe to synced documents and collections

## Estimated Timeline

- **Phase 1-2 (Core Implementation)**: 5-7 days
- **Phase 3 (Client Implementations)**: 3-4 days  
- **Phase 4 (Testing)**: 4-5 days
- **Phase 5 (Advanced Features)**: 2-3 days
- **Total**: 14-19 days

## Dependencies

- Existing DefraDB networking infrastructure
- Event system framework
- IPLD/DAG synchronization mechanisms
- go-libp2p-pubsub-rpc WithMultiResponse option
- Multi-client testing framework

## Questions and Implementation Concerns

1. **Event System Integration**: How should the event subscription be set up in the server initialization? Should there be a dedicated event handler registration for DocSyncRequest events?

2. **Topic Naming Strategy**: What should be the naming convention for document sync topics? Should it be based on collection ID, a dedicated sync topic, or a combination?

3. **DAG Sync Implementation**: The plan references `s.syncDocumentDAG(ctx, docCid)` - should this use existing DAG sync functionality from `net/sync_dag.go` or need a new implementation?

4. **Multi-Response Handling**: How long should we wait for multiple responses? Should there be a minimum number of responses required before returning, or should we return as soon as we get the first successful response?

5. **Subscription Management**: When auto-subscribing to documents after sync, should we also unsubscribe if the sync fails? How do we handle subscription lifecycle management?

6. **Error Aggregation**: When receiving multiple responses, how should we aggregate errors? Should we prefer successful responses over failed ones, or collect all errors?

7. **Duplicate Response Handling**: If multiple peers respond with the same document head, should we deduplicate or sync from multiple sources for verification?

8. **Timeout Configuration**: Should the timeout be configurable globally, per-request, or both? Should there be different timeouts for network requests vs DAG sync?

9. **Client Interface Consistency**: Should the timeout parameter be required or optional with a default? How should this be consistent across all client implementations?

10. **Testing Framework Integration**: Are there any existing testing utilities for P2P operations that should be leveraged, or patterns that should be followed for the custom function actions?

## Future Enhancements

1. **GraphQL Integration**: Add GraphQL mutations for document sync
2. **Subscription Support**: Real-time sync status updates via subscriptions  
3. **Selective Field Sync**: Sync only specific fields of documents
4. **Conflict Resolution**: Advanced merge strategies for concurrent updates
5. **Metrics and Monitoring**: Detailed sync performance metrics and dashboards