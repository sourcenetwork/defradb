---
description: High-level architecture of DefraDB and its components
globs: "**/*.go"
alwaysApply: true
---

# DefraDB Architecture Overview

DefraDB is a user-centric database that prioritizes data ownership, personal privacy, and information security. Its architecture is designed to enable decentralized data management with strong security and privacy controls.

## Core Architecture Components

### 1. Node

The central orchestrator that manages all subsystems:
- Initializes and manages database lifecycle
- Configures and maintains P2P networking
- Serves the HTTP API
- Manages access control policies
- Handles encryption and key management

### 2. Database Layer

Implements document and collection operations:
- **Collections**: Schema-defined containers for documents
- **Documents**: Core data entities with versioned history
- **Schemas**: Define document structure and relationships
- **Transactions**: Ensure ACID compliance for operations

### 3. Storage Layer

Manages persistent data with multiple backend options:
- **Badger**: Default key-value store (optimized for SSDs)
- **In-Memory**: For testing and ephemeral workloads
- **Blockstore**: For IPLD blocks (data, commits)
- Supports encryption at rest

### 4. Query Processing

Parses, plans, and executes queries through:
- **DQL Parser**: Processes DefraDB Query Language (GraphQL-compatible)
- **Query Planner**: Optimizes execution strategy
- **Execution Engine**: Performs actual data operations
- Supports filtering, aggregation, and relationships

### 5. P2P Networking

Enables distributed operation:
- Built on libp2p for peer discovery and communication
- **Pubsub**: Passive synchronization via broadcasts
- **Replicator**: Active data pushing to specific peers
- **DAG Sync**: Synchronizes document commit history

### 6. Access Control System (ACP)

Manages granular permissions with:
- **Document Access Control (DAC)**: Controls who can access entire documents
- **Field Access Control (FAC)**: Fine-grained field-level permissions
- **Admin Access Control (AAC)**: Controls system-level operations
- **Local or SourceHub-based** policy enforcement

### 7. Indexing System

Enables efficient data access with:
- **Secondary Indexes**: For fast field-value lookups
- **Unique Indexes**: To enforce data constraints
- **Compound Indexes**: For multi-field queries
- **JSON Indexes**: For deep querying of nested data
- **Relationship Indexes**: For efficient cross-collection queries

### 8. MerkleCRDT Layer

Provides conflict-free replication:
- **Last-Write-Wins (LWW)**: For basic field types
- **P-N Counter**: For incrementing/decrementing counters
- **Composite**: For complex object types
- Ensures eventual consistency across distributed nodes

## Data Flow Patterns

1. **Queries**: 
   - Parsed from GraphQL/JSON
   - Planned for optimal execution
   - Executed against collections
   - Results returned via HTTP/GraphQL

2. **Mutations**: 
   - Executed within transactions
   - Generate CRDT commits
   - Update document state
   - Publish changes via P2P if configured

3. **P2P Synchronization**: 
   - Nodes discover each other via libp2p
   - Subscribe to collection topics
   - Exchange document commits
   - Resolve conflicts via MerkleCRDTs

4. **Access Control**: 
   - Policies defined at schema/collection level
   - Identity verified via JWT/signatures
   - Permissions checked before operations
   - Field-level filtering applied to results

## Core Design Philosophies

1. **User-Centric**: Data ownership and privacy as foundational principles
2. **Decentralized**: Multi-write-master architecture with no central authority
3. **Conflict-Free**: Automatically resolves conflicts in distributed scenarios
4. **Schema-Driven**: Strong typing with GraphQL schema definitions
5. **Security-First**: Granular access controls and end-to-end encryption
6. **GraphQL-Compatible**: Familiar query language with extensions
7. **Content-Addressable**: Leverages IPLD for data addressing and linking

When implementing features or fixes, ensure alignment with these architectural principles and maintain the separation of concerns between components.