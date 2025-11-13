# Developer Plan: Searchable Encryption

This plan outlines the implementation of Searchable Encryption (SE) for DefraDB, based on the specifications in `/specs/se.md`. The goal is to enable privacy-preserving queries on encrypted data stored on untrusted remote nodes.

## Current Status

### Phase 1: Key Management and Infrastructure ✅ COMPLETED
- [x] SE key generation and storage in keyring
- [x] GraphQL `@encryptedIndex` directive for schema definitions
- [x] Encrypted index metadata storage (FieldName, Type)
- [x] CLI commands for encrypted index management
- [x] HTTP API endpoints for encrypted index operations
- [x] Test infrastructure

### Phase 2: Client-Side Tag Generation 🚧 IN PROGRESS
See: [Phase 2 Development Plan](./se_phase_2_development_plan.md)

### Phase 3: Replication and Remote Storage
See: [Phase 3 Development Plan](./se_phase_3_development_plan.md)

### Phase 4: Query API and Execution
See: [Phase 4 Development Plan](./se_phase_4_development_plan.md)

### Phase 5: Testing and Documentation
See: [Phase 5 Development Plan](./se_phase_5_development_plan.md)

## High-Level Architecture

```
Client Request → Collection.Save() → Context Injection → AddDelta
                                                           ↓
                                                    SE Package Processing
                                                           ↓
                                                    Artifact Generation
                                                           ↓
                                                    txn.OnSuccess → Schedule for Replication
```

## Key Components Implemented

### 1. Key Management (`cli/start.go`, `internal/db/config.go`)
- `searchable-encryption-key` generated and stored in keyring
- 32-byte key for HMAC-SHA256 operations
- CLI flag `--no-searchable-encryption` to disable

### 2. Schema Support (`internal/request/graphql/schema/`)
- `@encryptedIndex` directive parser
- Field validation for encrypted indexes
- Integration with collection definition

### 3. Index Management (`client/encrypted_index.go`)
```go
type EncryptedIndexDescription struct {
    FieldName string
    Type      EncryptedIndexType
}

type EncryptedCollectionIndex interface {
    Save(context.Context, datastore.Txn, *Document) error
    Update(context.Context, datastore.Txn, *Document, *Document) error
    Delete(context.Context, datastore.Txn, *Document) error
    Name() string
    Description() EncryptedIndexDescription
}
```

### 4. Collection Integration (`internal/db/collection_index.go`)
- `CreateEncryptedIndex` implementation
- `GetEncryptedIndexes` for listing
- Validation and persistence logic

## Implementation Status

- Phase 1: ✅ Complete
- Phase 2: 🚧 In Progress
- Phase 3: 📋 Planned
- Phase 4: 📋 Planned
- Phase 5: 📋 Planned

## Future Enhancements (Post-Phase 5)

- Support for range queries
- Full-text search capabilities
- Key rotation mechanisms
- Performance optimizations (batching, caching)
- Integration with Orbis KMS
- Relationship queries across encrypted fields

## References

- [SE Specification](/specs/se.md)
- [DefraDB Architecture Overview](/ai/docs/architecture-overview.md)
- [Data Flow Documentation](/internal/db/data_flow.md)