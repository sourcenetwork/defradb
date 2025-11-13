# Plan 6: Refactor SE Queries to Return DocIDs Only

## Overview

With the introduction of the `SyncDocuments` method in the P2P interface, we can simplify the searchable encryption query implementation. Instead of fetching full documents through the SE query, we'll return only document IDs and let users call `SyncDocuments` to retrieve the actual documents.

## Current State

The current implementation:
1. Generates an encrypted query field `{CollectionName}_encrypted` that returns full documents
2. seScanNode attempts to fetch documents both locally and from the network
3. The implementation is incomplete with network operations commented out

## New Approach

### 1. GraphQL Schema Changes

Instead of returning full documents, the encrypted query will return a specialized result type containing only docIDs:

```graphql
# Current
query {
  User_encrypted(filter: {...}) {
    name
    email
    # ... other fields
  }
}

# New
query {
  User_encrypted(filter: {...}) {
    docIDs
  }
}
```

### 2. Implementation Pattern

Following the commits query pattern:
- Define a result type that only contains docID
- Modify the GraphQL field to return this type
- Update the planner to handle this special case

### 3. Modified Components

#### A. GraphQL Schema Generation

1. **Created Shared Result Type** (`internal/request/graphql/schema/types/encrypted_search.go`)
   - Single `EncryptedSearchResult` type shared by all collections
   - Contains `docIDs` field that returns an array of strings

```go
type EncryptedSearchResult {
    docIDs: [String!]!
}
```

2. **Updated Schema Generation** (`internal/request/graphql/schema/generate.go`)
   - Modified `GenerateEncryptedQueryInputForGQLType` to use the shared type
   - Returns the shared `EncryptedSearchResult` type instead of collection-specific types

#### B. seScanNode (`internal/planner/se_scan.go`)

Simplified the implementation:
1. Removed document fetching logic
2. Removed currentIndex tracking - returns all docIDs at once
3. Return a single result containing the docIDs array

```go
func (n *seScanNode) Next() (bool, error) {
    // Return all docIDs at once in a single result
    if n.hasReturned {
        return false, nil
    }

    // Query remote nodes if not done yet
    if n.remoteDocIDs == nil {
        docIDs, err := n.queryRemoteNodes()
        if err != nil {
            return false, err
        }
        n.remoteDocIDs = docIDs
    }

    // Create a single document with the docIDs array
    doc := n.documentMapping.NewDoc()
    n.documentMapping.SetFirstOfName(&doc, "docIDs", n.remoteDocIDs)
    n.currentValue = doc
    n.hasReturned = true

    return true, nil
}
```

#### C. Query Parser Updates

Need to ensure the parser recognizes encrypted queries as a special type that returns docIDs only.

### 4. User Workflow

With this change, users will:
1. Execute encrypted search to get docIDs
2. Call `SyncDocuments` with the returned docIDs to fetch full documents

```go
// Step 1: Search for matching documents
result, err := db.ExecRequest(ctx, `query {
    User_encrypted(filter: {name: {_eq: "encrypted_value"}}) {
        docIDs
    }
}`)

// Step 2: Extract docIDs from result
docIDs := result.User_encrypted.docIDs

// Step 3: Sync documents
err = db.SyncDocuments(ctx, "User", docIDs)

// Step 4: Query synced documents locally
docs, err := db.ExecRequest(ctx, `query {
    User(docIDs: [...]) {
        name
        email
        // ... other fields
    }
}`)
```

### 5. Benefits

1. **Simpler Implementation**: No need for complex document fetching in seScanNode
2. **Better Separation of Concerns**: SE query focuses only on finding matches
3. **Flexibility**: Users can choose which documents to sync
4. **Consistency**: Aligns with DefraDB's pattern of specialized queries (like commits)

### 6. Implementation Steps

- [x] Analyze current implementation
- [ ] Update GraphQL schema generation to return docID-only type
- [ ] Simplify seScanNode to return docIDs without fetching
- [ ] Update query parser if needed
- [ ] Test the implementation
- [ ] Update documentation

## Notes

- This approach significantly simplifies the SE implementation
- Network document fetching is delegated to the existing `SyncDocuments` method
- The change is backward-incompatible but aligns better with DefraDB's architecture