# Phase 2: Client-Side Tag Generation - Development Plan

## Overview
This phase implements the core searchable encryption functionality where document field values are transformed into searchable tags using HMAC-SHA256. The implementation will integrate with DefraDB's existing document flow to generate, store, and manage encrypted search tags.

**Design Philosophy**: All searchable encryption logic is contained within `internal/se/*` to maintain clean separation of concerns and make it easier to reason about SE functionality. Other packages only make minimal calls to the SE package.

## Execution Flow

```
Document Create/Update (collection.save)
    ↓
Check for Encrypted Fields
    ↓
Inject SE Context (if needed)
    ↓
Continue Normal Processing → AddDelta
    ↓
SE Hook in AddDelta → Generate Tags
    ↓
Create SE Artifacts
    ↓
Store in Transaction → OnSuccess → Queue for Replication
```

## Key Integration Points

### 1. Collection Save Enhancement (`internal/db/collection.go`)

**Location**: `collection.save()` method  
**Purpose**: Pass collection and document to SE for complete handling

```go
import (
    // ... existing imports ...
    "github.com/sourcenetwork/defradb/internal/se"
)

// In collection.save():
func (c *collection) save(ctx context.Context, doc *client.Document) error {
    // ... existing validation ...
    
    // Prepare SE context if configured
    ctx = se.PrepareContextIfConfigured(ctx, c, doc, c.db.searchableEncryptionKey)
    
    // ... continue with existing save logic ...
}
```

### 2. AddDelta SE Integration (`internal/core/block/store.go`)

**Location**: `AddDelta()` function  
**Purpose**: Process block for searchable encryption

```go
import (
    // ... existing imports ...
    "github.com/sourcenetwork/defradb/internal/se"
)

// After existing encryption handling, before ProcessBlock:
func AddDelta(
    ctx context.Context,
    txn datastore.Txn,
    crdt core.ReplicatedData,
    delta core.Delta,
    links ...DAGLink,
) (cidlink.Link, []byte, error) {
    // ... existing logic ...
    
    // Handle searchable encryption
    if err := se.ProcessBlock(ctx, txn, block); err != nil {
        return cidlink.Link{}, nil, err
    }
    
    // ... continue with ProcessBlock ...
}
```

### 3. SE Package Implementation

**Package Structure**:
- `internal/se/` - DB-related SE logic (context, block processing, artifact management)
- `internal/se/core/` - Core SE functionality (tag generation, cryptographic operations)

**Domain Separator Rationale**:
The domain separator in HMAC serves as a cryptographic namespace to prevent collision attacks:
- Without it, the same (key, value) pair would produce identical tags across different contexts
- `eq:` prefix distinguishes equality tags from future range/prefix tags
- `collectionID:fieldName` ensures tags are unique per field per collection
- This string is only used as HMAC input, not stored, so there's no storage overhead
- We avoid prefixes like "defra:se:" to keep the separator minimal while maintaining securit

#### Core Package (`internal/se/core/`)

```go
// internal/se/core/tag.go
package secore

import (
    "crypto/hmac"
    "crypto/sha256"
    "fmt"
)

// GenerateEqualityTag creates a deterministic search tag for equality queries
func GenerateEqualityTag(
    key []byte,
    collectionID string,
    fieldName string,
    value []byte,
) ([]byte, error) {
    // Domain separation explanation:
    // - "eq" indicates equality search (vs future range/prefix)
    // - collectionID ensures tags are unique per collection
    // - fieldName ensures tags are unique per field
    // This prevents cross-field and cross-collection tag collisions
    // Note: This is HMAC input, not stored data
    domainSeparator := fmt.Sprintf("eq:%s:%s", collectionID, fieldName)
    
    // Compute HMAC-SHA256
    h := hmac.New(sha256.New, key)
    h.Write([]byte(domainSeparator))
    h.Write(value)
    tag := h.Sum(nil)
    
    // Truncate to 16 bytes for efficiency (128-bit security)
    return tag[:16], nil
}

// internal/se/core/artifact.go
package secore

type ArtifactType string
type OperationType string

const (
    ArtifactTypeEqualityTag ArtifactType = "equality_tag"
    
    OperationAdd    OperationType = "add"
    OperationDelete OperationType = "delete"
)

// Artifact represents a searchable encryption operation to be replicated
type Artifact struct {
    Type         ArtifactType
    CollectionID string
    FieldName    string
    DocID        string
    Tag          []byte
    Operation    OperationType
}
```

#### SE Package (`internal/se/`)

```go
// internal/se/context.go
package se

import (
    "context"
    "github.com/sourcenetwork/defradb/client"
    "github.com/sourcenetwork/defradb/internal/se/core"
)

type contextKey struct{}

type Config struct {
    Key             []byte
    CollectionID    string
    EncryptedFields []client.EncryptedIndexDescription
}

type Context struct {
    config    Config
    artifacts []secore.Artifact
    doc       *client.Document  // Reference to document being processed
    txn       datastore.Txn     // Transaction for OnSuccess callback
}

// PrepareContextIfConfigured checks collection configuration and prepares SE context if needed
func PrepareContextIfConfigured(
    ctx context.Context, 
    col client.Collection, 
    doc *client.Document,
    seKey []byte,
) context.Context {
    // Check if SE is configured
    encryptedIndexes := col.Version().EncryptedIndexes
    
    if len(encryptedIndexes) == 0 || len(seKey) == 0 {
        return ctx // Nothing to do
    }
    
    // Get transaction from context
    txn := txnctx.MustGet(ctx)
    
    // Create SE context
    seCtx := &Context{
        config: Config{
            Key:             seKey,
            CollectionID:    col.Version().VersionID,
            EncryptedFields: encryptedIndexes,
        },
        artifacts: make([]secore.Artifact, 0),
        doc:       doc,
        txn:       txn,
    }
    
    // Register callback to handle artifacts after processing
    seCtx.registerReplicationCallback()
    
    return context.WithValue(ctx, contextKey{}, seCtx)
}

// registerReplicationCallback sets up transaction callback for artifact replication
func (c *Context) registerReplicationCallback() {
    c.txn.OnSuccess(func() {
        if len(c.artifacts) == 0 {
            return
        }
        
        // Set doc ID on all artifacts
        docID := c.doc.ID().String()
        for i := range c.artifacts {
            c.artifacts[i].DocID = docID
        }
        
        // Queue for replication (Phase 3)
        // c.db.QueueSEArtifacts(c.artifacts)
    })
}

// internal/se/block.go
package se

import (
    "context"
    "fmt"
    "github.com/sourcenetwork/defradb/datastore"
    "github.com/sourcenetwork/defradb/internal/core/block"
    "github.com/sourcenetwork/defradb/internal/se/core"
    "github.com/sourcenetwork/defradb/client"
)

// ProcessBlock handles SE for a block
func ProcessBlock(
    ctx context.Context,
    txn datastore.Txn,
    block *coreblock.Block,
) error {
    seCtx, ok := ctx.Value(contextKey{}).(*Context)
    if !ok {
        return nil // SE not enabled, nothing to do
    }
    
    // Only process field-level blocks
    if block.Delta.IsComposite() || block.Delta.IsCollection() {
        return nil
    }
    
    fieldName := block.Delta.GetFieldName()
    
    // Check if field has encrypted index
    var encIdx *client.EncryptedIndexDescription
    for _, idx := range seCtx.config.EncryptedFields {
        if idx.FieldName == fieldName {
            encIdx = &idx
            break
        }
    }
    
    if encIdx == nil {
        return nil
    }
    
    // Generate search tag based on index type
    var tag []byte
    var err error
    
    switch encIdx.Type {
    case client.EncryptedIndexTypeEquality:
        tag, err = secore.GenerateEqualityTag(
            seCtx.config.Key,
            seCtx.config.CollectionID,
            fieldName,
            block.Delta.GetData(),
        )
    default:
        return fmt.Errorf("unsupported index type: %s", encIdx.Type)
    }
    
    if err != nil {
        return err
    }
    
    // Create and store artifact
    artifact := secore.Artifact{
        Type:         secore.ArtifactTypeEqualityTag,
        CollectionID: seCtx.config.CollectionID,
        FieldName:    fieldName,
        Tag:          tag,
        Operation:    secore.OperationAdd,
        // DocID will be set later when available
    }
    
    seCtx.artifacts = append(seCtx.artifacts, artifact)
    return nil
}
```

### 4. Automatic Replication Scheduling

**Location**: Internal to SE package  
**Purpose**: Automatically handled via transaction callback

The SE package automatically schedules replication when the transaction succeeds. No additional calls are needed in `collection.save()` - everything is handled when `PrepareContextIfConfigured()` is called with the document pointer.

## Testing Strategy

1. **Unit Tests** (`internal/se/core/tag_test.go`)
   - Tag generation determinism
   - Domain separator correctness
   - Value encoding edge cases

2. **Integration Tests** (`tests/integration/searchable_encryption/`)
   - Document creation with encrypted fields
   - Update operations generating new tags
   - Delete operations scheduling tag removal

3. **Benchmarks**
   - Tag generation performance
   - Memory overhead of context passing
   - Transaction impact

## Open Questions

1. **Artifact Storage**: Should artifacts be persisted immediately or queued in memory?
2. **Batch Processing**: Should we batch multiple field tags per document?
3. **Error Handling**: How to handle partial failures in tag generation?