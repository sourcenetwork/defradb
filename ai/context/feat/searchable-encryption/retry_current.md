# Current Retry Mechanism Comparison

## Document Replication Retry Flow (net/p2p_replicator.go)

### 1. Failure Registration
```go
func (p *Peer) handleReplicatorFailure(ctx context.Context, peerID, docID string) error {
    // Update replicator status to inactive
    updateReplicatorStatus(ctx, peerID, false)
    
    // Create retry entry if not exists
    createIfNotExistsReplicatorRetry(ctx, peerID, p.retryIntervals)
    
    // Store specific docID for retry
    docIDKey := keys.NewReplicatorRetryDocIDKey(peerID, docID)
    txn.Peerstore().Set(ctx, docIDKey.Bytes(), []byte{})
}
```

### 2. Retry Loop
```go
func (p *Peer) handleReplicatorRetries(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        case <-time.After(retryLoopInterval):
            p.retryReplicators(ctx)
        }
    }
}
```

### 3. Processing Retries
```go
func (p *Peer) retryReplicators(ctx context.Context) {
    // Iterate retry entries
    iter := peerstore.Iterator(prefix: keys.REPLICATOR_RETRY_ID)
    
    for each retry entry {
        // Unmarshal retry info
        if unmarshal fails {
            // Delete corrupted retry and all docs
            p.deleteReplicatorRetryAndDocs(ctx, key.PeerID)
            continue
        }
        
        // Check if retry is due
        if now.After(rInfo.NextRetry) && !rInfo.Retrying {
            // IMPORTANT: Check if replicator still exists
            exists := peerstore.Has(keys.NewReplicatorKey(key.PeerID))
            if !exists {
                // Clean up orphaned retry data
                p.deleteReplicatorRetryAndDocs(ctx, key.PeerID)
                continue
            }
            
            // Mark as retrying
            rInfo.Retrying = true
            rInfo.NumRetries++
            
            // Launch async retry
            go p.retryReplicator(ctx, key.PeerID)
        }
    }
}
```

### 4. Document Retry Execution
```go
func (p *Peer) retryReplicator(ctx context.Context, peerID string) {
    // Get all retry docs for this peer
    iter := peerstore.Iterator(prefix: keys.NewReplicatorRetryDocIDKey(peerID, ""))
    
    for each docID {
        err := p.retryDoc(ctx, peerID, key.DocID)
        if err != nil {
            // On first failure, stop and schedule next retry
            p.handleCompletedReplicatorRetry(ctx, peerID, false)
            return
        }
        // Delete successful doc retry
        peerstore.Delete(key)
    }
    
    // All docs succeeded
    p.handleCompletedReplicatorRetry(ctx, peerID, true)
}
```

### 5. Retry Completion
```go
func (p *Peer) handleCompletedReplicatorRetry(ctx context.Context, peerID string, success bool) {
    if success {
        // Check if more docs to retry
        done := deleteReplicatorRetryIfNoMoreDocs(ctx, peerID)
        if done {
            updateReplicatorStatus(ctx, peerID, true) // Set to active
        } else {
            setReplicatorNextRetry(ctx, peerID, []time.Duration{0}) // Immediate retry
        }
    } else {
        // Schedule next retry with backoff
        setReplicatorNextRetry(ctx, peerID, p.retryIntervals)
    }
}
```

## SE Replication Retry Flow (internal/se/replication_coordinator.go)

### 1. Failure Registration
```go
func (rc *ReplicationCoordinator) handleReplicationFailure(ctx context.Context, evt ReplicationFailureEvent) error {
    retryKey := keys.NewPeerstoreSERetry(evt.PeerID.String(), evt.CollectionID, evt.DocID)
    
    // Extract identity info
    var publicKey, keyType string
    if evt.Identity.HasValue() {
        identity := evt.Identity.Value()
        if pubKey := identity.PublicKey(); pubKey != nil {
            publicKey = hex.EncodeToString(pubKey.Raw())
            keyType = string(pubKey.Type())
        }
    }
    
    // Create retry info
    retryInfo := SERetryInfo{
        DocID:        evt.DocID,
        CollectionID: evt.CollectionID,
        FieldNames:   evt.FieldNames,
        NextRetry:    time.Now().Add(rc.retryIntervals[0]),
        NumRetries:   0,
        PublicKey:    publicKey,
        KeyType:      keyType,
    }
    
    // Store retry info
    ps.Set(ctx, retryKey.Bytes(), cbor.Marshal(retryInfo))
}
```

### 2. Retry Loop
```go
func (rc *ReplicationCoordinator) retrySEReplicators(ctx context.Context) {
    ticker := time.NewTicker(retryLoopInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            rc.processSERetries(ctx)
        }
    }
}
```

### 3. Processing Retries
```go
func (rc *ReplicationCoordinator) processSERetries(ctx context.Context) {
    // Iterate SE retry entries
    iter := peerstore.Iterator(prefix: keys.NewPeerstoreSERetry("", "", ""))
    
    for each retry entry {
        // Unmarshal retry info
        if unmarshal fails {
            log.Error("Failed to unmarshal SE retry info")
            continue // NOTE: No cleanup like document replication
        }
        
        // Check if retry is due
        if now.After(retryInfo.NextRetry) && !retryInfo.Retrying {
            // NOTE: No check if peer/collection still exists
            
            // Mark as retrying
            retryInfo.Retrying = true
            retryInfo.NumRetries++
            
            // Launch async retry
            go rc.retrySEArtifacts(ctx, key.PeerID, retryInfo)
        }
    }
}
```

### 4. SE Artifact Retry Execution
```go
func (rc *ReplicationCoordinator) retrySEArtifacts(ctx context.Context, peerID string, retryInfo SERetryInfo) {
    successChan := make(chan bool)
    
    // Reconstruct identity
    identity := rc.reconstructIdentity(retryInfo.PublicKey, retryInfo.KeyType)
    if identity.HasValue() {
        ctx = acpIdentity.WithContext(ctx, identity)
    }
    
    // Generate artifacts
    artifacts := rc.generateSEArtifacts(ctx, retryInfo.DocID, retryInfo.CollectionID, retryInfo.FieldNames)
    
    // Publish retry event
    rc.eventBus.Publish(ReplicateEvent{
        DocID:        retryInfo.DocID,
        CollectionID: retryInfo.CollectionID,
        Artifacts:    artifacts,
        IsRetry:      true,
        Success:      successChan,
        Identity:     identity,
    })
    
    // Wait for result with timeout
    select {
    case success := <-successChan:
        rc.updateRetryStatus(ctx, peerID, retryInfo, success)
    case <-time.After(retryTimeout):
        rc.updateRetryStatus(ctx, peerID, retryInfo, false)
    }
}
```

### 5. Retry Completion
```go
func (rc *ReplicationCoordinator) updateRetryStatus(ctx context.Context, peerID string, retryInfo SERetryInfo, success bool) {
    retryKey := keys.NewPeerstoreSERetry(peerID, retryInfo.CollectionID, retryInfo.DocID)
    
    if success {
        // Delete retry entry
        ps.Delete(ctx, retryKey.Bytes())
    } else {
        // Update next retry time with backoff
        if retryInfo.NumRetries >= len(rc.retryIntervals) {
            retryInfo.NextRetry = time.Now().Add(rc.retryIntervals[len(rc.retryIntervals)-1])
        } else {
            retryInfo.NextRetry = time.Now().Add(rc.retryIntervals[retryInfo.NumRetries])
        }
        retryInfo.Retrying = false
        
        // Update retry info
        ps.Set(ctx, retryKey.Bytes(), cbor.Marshal(retryInfo))
    }
}
```

## Key Differences

### 1. Retry Granularity
- **Document**: Tracks individual docIDs per peer (two-level: peer → docs)
- **SE**: Single retry entry per (peer, collection, doc) combination

### 2. Data Storage
- **Document**: 
  - Main retry info: `REPLICATOR_RETRY_ID/{peerID}`
  - Individual docs: `REPLICATOR_RETRY_DOCID/{peerID}/{docID}`
- **SE**: 
  - Single entry: `PEERSTORE_SE_RETRY/{peerID}/{collectionID}/{docID}`

### 3. Identity Handling
- **Document**: No identity tracking
- **SE**: Stores public key and key type for identity reconstruction

### 4. Error Handling
- **Document**: 
  - Cleans up corrupted retry data
  - Checks if replicator still exists before retry
  - Stops on first doc failure
- **SE**: 
  - Logs but continues on corrupted data
  - No existence checks
  - Single artifact retry per entry

### 5. Status Updates
- **Document**: Updates replicator status (active/inactive)
- **SE**: No status tracking

### 6. Retry Strategy
- **Document**: 
  - Immediate retry if more docs after partial success
  - Exponential backoff on failure
- **SE**: 
  - Always exponential backoff
  - No concept of partial success

### 7. Cleanup
- **Document**: Has `deleteReplicatorRetryAndDocs` for orphaned data
- **SE**: No equivalent cleanup mechanism