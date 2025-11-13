# Generic Retry Coordinator Plan

## Overview

Create a generic retry coordinator package that can handle retry logic for both SE replication and document replication. The solution uses Go generics for type safety and clean APIs.

**Important**: The document replication (Peer) implementation is currently working and tested, so SE replication should be adjusted to follow its patterns.

## Key Design Decisions

1. **Two-Level Retry Structure** (following document replication pattern):
   - Main retry entry per peer with retry metadata
   - Separate entries for individual items to retry

2. **Reuse existing key types** with different prefixes:
   - Both use `ReplicatorRetryIDKey` for main retry entry
   - Both use `ReplicatorRetryDocIDKey` for items (SE stores combined data as "docID")

3. **Minimal Handler interface** with only essential methods

4. **Key Differences Found**:
   - SE currently uses single-level retry (stores all data in one key)
   - Document replication uses two-level (main retry + individual doc keys)
   - Both check for replicator existence using `keys.NewReplicatorKey(peerID)`

## Revised Generic Retry Package

### Core Coordinator Implementation

```go
// internal/retry/coordinator.go
package retry

import (
    "context"
    "time"
    "github.com/fxamacker/cbor/v2"
    "github.com/sourcenetwork/corekv"
    "github.com/sourcenetwork/corelog"
    "github.com/sourcenetwork/defradb/internal/keys"
    "github.com/sourcenetwork/defradb/internal/datastore"
    "github.com/sourcenetwork/defradb/errors"
)

// RetryInfo is the main retry information (stored at peer level)
type RetryInfo struct {
    NextRetry  time.Time
    NumRetries int
    Retrying   bool
}

// Config defines the retry coordinator configuration
type Config struct {
    Name              string
    RetryKeyPrefix    string  // e.g., "REPLICATOR_RETRY_ID" or "SE_RETRY_ID"
    ItemKeyPrefix     string  // e.g., "REPLICATOR_RETRY_DOCID" or "SE_RETRY_ITEM"
    RetryIntervals    []time.Duration
    CheckInterval     time.Duration
    UpdateStatusFunc  func(ctx context.Context, peerID string, active bool) error  // Optional
    Logger            *corelog.Logger
}

// Handler processes retry items - minimal interface
type Handler interface {
    // Core processing
    ProcessItem(ctx context.Context, peerID string, itemID string) error
    
    // Cleanup helper (for deleteReplicatorRetryAndDocs pattern)
    GetCleanupFunc() func(ctx context.Context, peerID string) error
}

// Coordinator manages retry operations
type Coordinator struct {
    config  Config
    handler Handler
    store   corekv.Store
    closeCh chan struct{}
}

// NewCoordinator creates a new retry coordinator
func NewCoordinator(store corekv.Store, config Config, handler Handler) *Coordinator {
    if config.Logger == nil {
        config.Logger = corelog.NewLogger("retry." + config.Name)
    }
    return &Coordinator{
        config:  config,
        handler: handler,
        store:   store,
        closeCh: make(chan struct{}),
    }
}

// Start begins the retry loop
func (c *Coordinator) Start(ctx context.Context) {
    ticker := time.NewTicker(c.config.CheckInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-c.closeCh:
            return
        case <-ticker.C:
            c.processRetries(ctx)
        }
    }
}

// RegisterFailure creates retry entry and adds item
func (c *Coordinator) RegisterFailure(ctx context.Context, peerID string, itemID string) error {
    // Update status to inactive if function provided
    if c.config.UpdateStatusFunc != nil {
        if err := c.config.UpdateStatusFunc(ctx, peerID, false); err != nil {
            return err
        }
    }

    // Create main retry entry if not exists
    if err := c.createIfNotExistsRetry(ctx, peerID); err != nil {
        return err
    }

    // Store item to retry
    itemKey := c.newItemKey(peerID, itemID)
    return c.store.Set(ctx, itemKey.Bytes(), []byte{})
}

// Key helper methods
func (c *Coordinator) newRetryKey(peerID string) keys.ReplicatorRetryIDKey {
    return keys.NewReplicatorRetryIDKeyWithPrefix(peerID, c.config.RetryKeyPrefix)
}

func (c *Coordinator) newItemKey(peerID string, itemID string) keys.ReplicatorRetryDocIDKey {
    return keys.NewReplicatorRetryDocIDKeyWithPrefix(peerID, itemID, c.config.ItemKeyPrefix)
}

// createIfNotExistsRetry creates main retry entry
func (c *Coordinator) createIfNotExistsRetry(ctx context.Context, peerID string) error {
    key := c.newRetryKey(peerID)
    exists, err := c.store.Has(ctx, key.Bytes())
    if err != nil {
        return err
    }
    if exists {
        return nil
    }
    
    info := RetryInfo{
        NextRetry:  time.Now().Add(c.config.RetryIntervals[0]),
        NumRetries: 0,
    }
    
    b, err := cbor.Marshal(info)
    if err != nil {
        return err
    }
    
    return c.store.Set(ctx, key.Bytes(), b)
}

// processRetries checks for due retries
func (c *Coordinator) processRetries(ctx context.Context) {
    iter, err := c.store.Iterator(ctx, corekv.IterOptions{
        Prefix: []byte(c.config.RetryKeyPrefix),
    })
    if err != nil {
        if errors.Is(err, corekv.ErrDBClosed) {
            return
        }
        c.config.Logger.ErrorContextE(ctx, "Failed to iterate retry keys", err)
        return
    }
    defer c.closeIterator(iter)

    now := time.Now()
    for {
        hasNext, err := iter.Next()
        if err != nil {
            c.config.Logger.ErrorContextE(ctx, "Failed to get next retry key", err)
            break
        }
        if !hasNext {
            break
        }

        key, err := keys.NewReplicatorRetryIDKeyFromString(string(iter.Key()))
        if err != nil {
            c.config.Logger.ErrorContextE(ctx, "Failed to parse retry key", err)
            continue
        }
        peerID := key.PeerID

        value, err := iter.Value()
        if err != nil {
            c.config.Logger.ErrorContextE(ctx, "Failed to get retry value", err)
            continue
        }

        var retryInfo RetryInfo
        if err := cbor.Unmarshal(value, &retryInfo); err != nil {
            c.config.Logger.ErrorContextE(ctx, "Failed to unmarshal retry info", err)
            // Clean up corrupted data
            if cleanupFunc := c.handler.GetCleanupFunc(); cleanupFunc != nil {
                cleanupFunc(ctx, peerID)
            }
            continue
        }

        if now.After(retryInfo.NextRetry) && !retryInfo.Retrying {
            // Mark as retrying
            if err := c.setAsRetrying(ctx, peerID, retryInfo); err != nil {
                c.config.Logger.ErrorContextE(ctx, "Failed to set as retrying", err)
                continue
            }

            // Process asynchronously
            go c.retryPeer(ctx, peerID)
        }
    }
}

// setAsRetrying updates retry info
func (c *Coordinator) setAsRetrying(ctx context.Context, peerID string, info RetryInfo) error {
    info.Retrying = true
    info.NumRetries++
    
    b, err := cbor.Marshal(info)
    if err != nil {
        return err
    }
    
    key := c.newRetryKey(peerID)
    return c.store.Set(ctx, key.Bytes(), b)
}

// retryPeer processes all items for a peer
func (c *Coordinator) retryPeer(ctx context.Context, peerID string) {
    c.config.Logger.InfoContext(ctx, "Retrying peer", corelog.String("PeerID", peerID))

    // Check if replicator exists using ReplicatorKey
    exists, err := c.store.Has(ctx, keys.NewReplicatorKey(peerID).Bytes())
    if err != nil {
        c.config.Logger.ErrorContextE(ctx, "Failed to check if replicator exists", err)
        return
    }
    if !exists {
        // Clean up orphaned retry data
        if cleanupFunc := c.handler.GetCleanupFunc(); cleanupFunc != nil {
            cleanupFunc(ctx, peerID)
        }
        return
    }

    // Get all items to retry
    iter, err := c.store.Iterator(ctx, corekv.IterOptions{
        Prefix:   c.newItemKey(peerID, "").Bytes(),
        KeysOnly: true,
    })
    if err != nil {
        c.config.Logger.ErrorContextE(ctx, "Failed to iterate retry items", err)
        return
    }
    defer c.closeIterator(iter)

    allSuccess := true
    for {
        select {
        case <-ctx.Done():
            return
        default:
        }

        hasNext, err := iter.Next()
        if err != nil {
            c.config.Logger.ErrorContextE(ctx, "Failed to get next retry item", err)
            break
        }
        if !hasNext {
            break
        }

        key, err := keys.NewReplicatorRetryDocIDKeyFromString(string(iter.Key()))
        if err != nil {
            c.config.Logger.ErrorContextE(ctx, "Failed to parse item key", err)
            continue
        }
        itemID := key.DocID

        // Process item
        if err := c.handler.ProcessItem(ctx, peerID, itemID); err != nil {
            c.config.Logger.ErrorContextE(ctx, "Failed to retry item", err,
                corelog.String("ItemID", itemID))
            allSuccess = false
            break  // Stop on first failure
        }

        // Delete successful item
        if err := c.store.Delete(ctx, iter.Key()); err != nil {
            c.config.Logger.ErrorContextE(ctx, "Failed to delete retry item", err)
        }
    }

    // Handle completion
    c.handleCompletedRetry(ctx, peerID, allSuccess)
}

// handleCompletedRetry updates retry status
func (c *Coordinator) handleCompletedRetry(ctx context.Context, peerID string, success bool) {
    if success {
        // Check if more items to retry
        done, err := c.deleteRetryIfNoMoreItems(ctx, peerID)
        if err != nil {
            c.config.Logger.ErrorContextE(ctx, "Failed to check remaining items", err)
            return
        }
        if done {
            // Update status to active if function provided
            if c.config.UpdateStatusFunc != nil {
                c.config.UpdateStatusFunc(ctx, peerID, true)
            }
        } else {
            // Immediate retry for remaining items
            c.setNextRetry(ctx, peerID, []time.Duration{0})
        }
    } else {
        // Schedule next retry with backoff
        c.setNextRetry(ctx, peerID, c.config.RetryIntervals)
    }
}

// deleteRetryIfNoMoreItems checks and deletes retry if no items left
func (c *Coordinator) deleteRetryIfNoMoreItems(ctx context.Context, peerID string) (bool, error) {
    // Check if any items remain
    iter, err := c.store.Iterator(ctx, corekv.IterOptions{
        Prefix:   c.newItemKey(peerID, "").Bytes(),
        KeysOnly: true,
    })
    if err != nil {
        return false, err
    }
    defer c.closeIterator(iter)

    hasNext, _ := iter.Next()
    if !hasNext {
        // No more items, delete main retry entry
        key := c.newRetryKey(peerID)
        return true, c.store.Delete(ctx, key.Bytes())
    }
    return false, nil
}

// setNextRetry updates retry time
func (c *Coordinator) setNextRetry(ctx context.Context, peerID string, intervals []time.Duration) error {
    key := c.newRetryKey(peerID)
    b, err := c.store.Get(ctx, key.Bytes())
    if err != nil {
        return err
    }

    var info RetryInfo
    if err := cbor.Unmarshal(b, &info); err != nil {
        return err
    }

    if info.NumRetries >= len(intervals) {
        info.NextRetry = time.Now().Add(intervals[len(intervals)-1])
    } else {
        info.NextRetry = time.Now().Add(intervals[info.NumRetries])
    }
    info.Retrying = false

    b, err = cbor.Marshal(info)
    if err != nil {
        return err
    }

    return c.store.Set(ctx, key.Bytes(), b)
}

// closeIterator safely closes iterator
func (c *Coordinator) closeIterator(iter corekv.Iterator) {
    if iter == nil {
        return
    }
    if err := iter.Close(); err != nil {
        c.config.Logger.ErrorE("Failed to close iterator", err)
    }
}

// Close stops the coordinator
func (c *Coordinator) Close() {
    close(c.closeCh)
}
```

## SE Implementation

**Important Migration Notes**:
- SE currently uses `PeerstoreSERetry` key which stores all retry data in a single entry
- Need to migrate to two-level structure using `ReplicatorRetryIDKey` and `ReplicatorRetryDocIDKey`
- SE stores CollectionID/DocID in the key itself, but we'll move this data to the item ID

```go
// internal/se/retry.go
package se

import (
    "context"
    "encoding/hex"
    "fmt"
    "strings"
    "time"
    
    "github.com/sourcenetwork/defradb/internal/retry"
    "github.com/sourcenetwork/defradb/internal/datastore"
    "github.com/sourcenetwork/defradb/internal/keys"
    "github.com/sourcenetwork/defradb/event"
    "github.com/sourcenetwork/corelog"
)

// SERetryHandler implements retry.Handler
type SERetryHandler struct {
    rc *ReplicationCoordinator
}

// ProcessItem handles a single SE retry
func (h *SERetryHandler) ProcessItem(ctx context.Context, peerID string, itemID string) error {
    // Parse item ID to get retry data
    parts := strings.Split(itemID, "/")
    if len(parts) != 5 {
        return fmt.Errorf("invalid item ID format")
    }
    
    collectionID := parts[0]
    docID := parts[1]
    fieldNames := strings.Split(parts[2], ",")
    publicKey := parts[3]
    keyType := parts[4]

    successChan := make(chan bool, 1)
    defer close(successChan)

    // Reconstruct identity
    identity, err := h.rc.reconstructIdentity(publicKey, keyType)
    if err != nil {
        log.ErrorContextE(ctx, "Failed to reconstruct identity", err,
            corelog.String("DocID", docID))
    } else if identity.HasValue() {
        ctx = acpIdentity.WithContext(ctx, identity)
    }

    // Generate artifacts
    artifacts, err := h.rc.generateSEArtifacts(ctx, docID, collectionID, fieldNames)
    if err != nil {
        return err
    }

    // Publish event
    h.rc.eventBus.Publish(event.NewMessage(ReplicateEventName, ReplicateEvent{
        DocID:        docID,
        CollectionID: collectionID,
        Artifacts:    artifacts,
        IsRetry:      true,
        Success:      successChan,
        Identity:     identity,
    }))

    // Wait for result
    select {
    case success := <-successChan:
        if !success {
            return fmt.Errorf("SE replication failed")
        }
        return nil
    case <-time.After(retryTimeout):
        return fmt.Errorf("SE replication timeout")
    }
}

// GetCleanupFunc returns cleanup function for SE
func (h *SERetryHandler) GetCleanupFunc() func(ctx context.Context, peerID string) error {
    return func(ctx context.Context, peerID string) error {
        // Delete all SE retry data for peer
        ps := datastore.PeerstoreFrom(h.rc.db.Rootstore())
        
        // Delete main retry entry
        retryKey := keys.NewReplicatorRetryIDKeyWithPrefix(peerID, keys.SE_RETRY_ID)
        if err := ps.Delete(ctx, retryKey.Bytes()); err != nil {
            return err
        }
        
        // Delete all items
        iter, err := ps.Iterator(ctx, corekv.IterOptions{
            Prefix: keys.NewReplicatorRetryDocIDKeyWithPrefix(peerID, "", keys.SE_RETRY_ITEM).Bytes(),
            KeysOnly: true,
        })
        if err != nil {
            return err
        }
        defer iter.Close()
        
        for {
            hasNext, err := iter.Next()
            if err != nil || !hasNext {
                break
            }
            if err := ps.Delete(ctx, iter.Key()); err != nil {
                return err
            }
        }
        
        return nil
    }
}

// Setup in ReplicationCoordinator
func (rc *ReplicationCoordinator) setupRetryCoordinator() {
    config := retry.Config{
        Name:           "se-replication",
        RetryKeyPrefix: keys.SE_RETRY_ID,
        ItemKeyPrefix:  keys.SE_RETRY_ITEM,
        RetryIntervals: rc.retryIntervals,
        CheckInterval:  retryLoopInterval,
        // SE doesn't use status updates
    }

    handler := &SERetryHandler{rc: rc}
    rc.retryCoordinator = retry.NewCoordinator(
        datastore.PeerstoreFrom(rc.db.Rootstore()),
        config,
        handler,
    )
    go rc.retryCoordinator.Start(context.Background())
}

// Adjusted handleReplicationFailure to use item-based approach
func (rc *ReplicationCoordinator) handleReplicationFailure(ctx context.Context, evt ReplicationFailureEvent) error {
    // Create item ID with all necessary data
    var publicKey, keyType string
    if evt.Identity.HasValue() {
        identity := evt.Identity.Value()
        if pubKey := identity.PublicKey(); pubKey != nil {
            publicKey = hex.EncodeToString(pubKey.Raw())
            keyType = string(pubKey.Type())
        }
    }
    
    // Store data in item ID (combined into single string)
    itemID := fmt.Sprintf("%s/%s/%s/%s/%s",
        evt.CollectionID,
        evt.DocID,
        strings.Join(evt.FieldNames, ","),
        publicKey,
        keyType,
    )
    
    return rc.retryCoordinator.RegisterFailure(ctx, evt.PeerID.String(), itemID)
}

// Migration from old SE retry structure
func (rc *ReplicationCoordinator) migrateSERetries(ctx context.Context) error {
    ps := datastore.PeerstoreFrom(rc.db.Rootstore())
    
    // Read old SE retry entries
    iter, err := ps.Iterator(ctx, corekv.IterOptions{
        Prefix: keys.NewPeerstoreSERetry("", "", "").Bytes(),
    })
    if err != nil {
        return err
    }
    defer iter.Close()
    
    for {
        hasNext, err := iter.Next()
        if err != nil || !hasNext {
            break
        }
        
        key, _ := keys.NewPeerstoreSERetryFromString(string(iter.Key()))
        value, _ := iter.Value()
        
        var retryInfo SERetryInfo
        if err := cbor.Unmarshal(value, &retryInfo); err != nil {
            continue
        }
        
        // Create item ID
        itemID := fmt.Sprintf("%s/%s/%s/%s/%s",
            retryInfo.CollectionID,
            retryInfo.DocID,
            strings.Join(retryInfo.FieldNames, ","),
            retryInfo.PublicKey,
            retryInfo.KeyType,
        )
        
        // Register with new coordinator
        rc.retryCoordinator.RegisterFailure(ctx, key.PeerID, itemID)
        
        // Delete old entry
        ps.Delete(ctx, iter.Key())
    }
    
    return nil
}
```

## Document Replication Handler

```go
// net/retry.go
package net

// DocRetryHandler implements retry.Handler
type DocRetryHandler struct {
    peer *Peer
}

// ProcessItem retries a single doc
func (h *DocRetryHandler) ProcessItem(ctx context.Context, peerID string, docID string) error {
    return h.peer.retryDoc(ctx, peerID, docID)
}

// GetCleanupFunc returns the existing deleteReplicatorRetryAndDocs function
func (h *DocRetryHandler) GetCleanupFunc() func(ctx context.Context, peerID string) error {
    return h.peer.deleteReplicatorRetryAndDocs
}

// Setup in Peer
func (p *Peer) setupRetryCoordinator() {
    config := retry.Config{
        Name:             "doc-replication",
        RetryKeyPrefix:   keys.REPLICATOR_RETRY_ID,
        ItemKeyPrefix:    keys.REPLICATOR_RETRY_DOCID,
        RetryIntervals:   p.retryIntervals,
        CheckInterval:    retryLoopInterval,
        UpdateStatusFunc: updateReplicatorStatus,  // Pass the existing function
    }

    handler := &DocRetryHandler{peer: p}
    p.retryCoordinator = retry.NewCoordinator(
        datastore.PeerstoreFrom(p.db.Rootstore()),
        config,
        handler,
    )
    go p.retryCoordinator.Start(p.ctx)
}

// Clean usage
func (p *Peer) handleReplicatorFailure(ctx context.Context, peerID, docID string) error {
    // The coordinator will handle status update via UpdateStatusFunc
    return p.retryCoordinator.RegisterFailure(ctx, peerID, docID)
}
```

## Key Modifications Needed

### 1. Add new constants to peerstore.go

```go
// internal/keys/peerstore.go
const (
    REPLICATOR           = "/rep/id"
    REPLICATOR_RETRY_ID  = "/rep/retry/id"
    REPLICATOR_RETRY_DOC = "/rep/retry/doc"
    SE_RETRY_ID          = "/se/retry/id"    // Add this
    SE_RETRY_ITEM        = "/se/retry/item"  // Add this
)
```

### 2. Update ReplicatorRetryIDKey to support custom prefixes

```go
// internal/keys/peerstore_replicator_retry.go

type ReplicatorRetryIDKey struct {
    PeerID string
    prefix string  // Add this field (unexported)
}

func NewReplicatorRetryIDKeyWithPrefix(peerID, prefix string) ReplicatorRetryIDKey {
    return ReplicatorRetryIDKey{
        PeerID: peerID,
        prefix: prefix,
    }
}

func (k ReplicatorRetryIDKey) ToString() string {
    prefix := k.prefix
    if prefix == "" {
        prefix = REPLICATOR_RETRY_ID
    }
    return prefix + "/" + k.PeerID
}

// Update NewReplicatorRetryIDKeyFromString to handle custom prefixes
func NewReplicatorRetryIDKeyFromString(key string) (ReplicatorRetryIDKey, error) {
    // Check for SE prefix first
    if strings.HasPrefix(key, SE_RETRY_ID+"/") {
        peerID := strings.TrimPrefix(key, SE_RETRY_ID+"/")
        if peerID == "" {
            return ReplicatorRetryIDKey{}, errors.WithStack(ErrInvalidKey, errors.NewKV("Key", key))
        }
        return ReplicatorRetryIDKey{PeerID: peerID, prefix: SE_RETRY_ID}, nil
    }
    
    // Default to REPLICATOR_RETRY_ID
    peerID := strings.TrimPrefix(key, REPLICATOR_RETRY_ID+"/")
    if peerID == "" {
        return ReplicatorRetryIDKey{}, errors.WithStack(ErrInvalidKey, errors.NewKV("Key", key))
    }
    return NewReplicatorRetryIDKey(peerID), nil
}
```

### 3. Update ReplicatorRetryDocIDKey similarly

```go
// internal/keys/peerstore_replicator_retry_doc.go

type ReplicatorRetryDocIDKey struct {
    PeerID string
    DocID  string
    prefix string  // Add this field (unexported)
}

func NewReplicatorRetryDocIDKeyWithPrefix(peerID, docID, prefix string) ReplicatorRetryDocIDKey {
    return ReplicatorRetryDocIDKey{
        PeerID: peerID,
        DocID:  docID,
        prefix: prefix,
    }
}

func (k ReplicatorRetryDocIDKey) ToString() string {
    prefix := k.prefix
    if prefix == "" {
        prefix = REPLICATOR_RETRY_DOC
    }
    keyString := prefix + "/" + k.PeerID
    if k.DocID != "" {
        keyString += "/" + k.DocID
    }
    return keyString
}

// Update NewReplicatorRetryDocIDKeyFromString to handle custom prefixes
func NewReplicatorRetryDocIDKeyFromString(key string) (ReplicatorRetryDocIDKey, error) {
    var prefix string
    var trimmedKey string
    
    if strings.HasPrefix(key, SE_RETRY_ITEM+"/") {
        prefix = SE_RETRY_ITEM
        trimmedKey = strings.TrimPrefix(key, SE_RETRY_ITEM+"/")
    } else {
        prefix = REPLICATOR_RETRY_DOC
        trimmedKey = strings.TrimPrefix(key, REPLICATOR_RETRY_DOC+"/")
    }
    
    keyArr := strings.Split(trimmedKey, "/")
    if len(keyArr) != 2 {
        return ReplicatorRetryDocIDKey{}, errors.WithStack(ErrInvalidKey, errors.NewKV("Key", key))
    }
    return ReplicatorRetryDocIDKey{
        PeerID: keyArr[0],
        DocID:  keyArr[1],
        prefix: prefix,
    }, nil
}
```

## Key Design Points

1. **Minimal Handler Interface**:
   - Only 2 methods: `ProcessItem` and `GetCleanupFunc`
   - Everything else handled by coordinator or config

2. **Reused Key Types**:
   - Both use `ReplicatorRetryIDKey` with different prefixes
   - Both use `ReplicatorRetryDocIDKey` (SE stores combined data as "docID")
   - Keys need minor modification to support custom prefixes

3. **Status Updates**:
   - Passed as optional function in config
   - Document replication uses it, SE doesn't

4. **Cleanup**:
   - Handler provides cleanup function
   - Coordinator calls it when needed

5. **Benefits**:
   - Much simpler handler interface
   - Reuses existing key types
   - All complexity in coordinator
   - Clean separation of concerns

## Implementation Summary

### Phase 1: Key Modifications
1. Add SE_RETRY_ID and SE_RETRY_ITEM constants to peerstore.go
2. Update ReplicatorRetryIDKey to support custom prefixes with NewReplicatorRetryIDKeyWithPrefix
3. Update ReplicatorRetryDocIDKey similarly
4. Update the FromString methods to handle both doc and SE prefixes

### Phase 2: Generic Retry Package
1. Create internal/retry/coordinator.go with the generic Coordinator
2. Define minimal Handler interface with ProcessItem and GetCleanupFunc
3. Implement all retry logic in the coordinator

### Phase 3: SE Migration
1. Implement SERetryHandler with the two required methods
2. Update ReplicationCoordinator to use the new retry coordinator
3. Migrate existing SE retry data from single-level to two-level structure
4. Delete old PeerstoreSERetry keys after migration

### Phase 4: Document Replication Update
1. Implement DocRetryHandler wrapping existing retry logic
2. Replace current retry loop with the generic coordinator
3. Test to ensure no regression in existing functionality

### Key Insight
The main difference between SE and document replication retry mechanisms is that SE currently stores all retry data in a single key (PeerstoreSERetry), while document replication uses a two-level structure. By migrating SE to the two-level structure and reusing the existing key types with different prefixes, we can unify both implementations under a single generic coordinator.