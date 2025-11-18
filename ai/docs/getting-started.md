---
description: Quick reference guide for AI agents working with DefraDB
globs: "**/*"
alwaysApply: true
---

# Getting Started with DefraDB Development

This guide provides a quick reference for AI agents working on DefraDB features and bug fixes.

## What is DefraDB?

DefraDB is a user-centric database that prioritizes:
- Data ownership and privacy
- Decentralized architecture
- Multi-master replication via MerkleCRDTs
- GraphQL-compatible query interface
- Peer-to-peer data synchronization
- Granular access control policies

## Development Environment

### Build and Run

```bash
# Build the binary
make build

# Install to $GOPATH/bin
make install

# Start the database
make start

# Start with development mode
make dev:start

# See CLI help
defradb --help
```

### Testing

```bash
# Run all tests
make test

# Run quick tests (no race detection)
make test:quick

# Run tests with code coverage
make test:coverage

# Run benchmark tests
make test:bench

# Clean test cache and run tests
make test:clean

# Run tests for specific module
make test path="./internal/db"
```

### Code Quality

```bash
# Run linters
make lint

# Fix linting issues automatically
make lint:fix

# Generate mocks
make mocks

# Update dependencies
make tidy

# All-in-one fix command
make fix
```

## Key Architectural Components

- **Node**: Central component that orchestrates all subsystems
- **Database Layer**: Manages collections, documents, and queries
- **Storage Layer**: Handles data persistence (Badger, in-memory, etc.)
- **Query Processing**: Parses, plans, and executes queries
- **P2P Networking**: Enables distributed operation
- **MerkleCRDT**: Provides conflict-free replication
- **Access Control**: Manages permissions and access policies
- **Indexing System**: Enables efficient lookups and queries

## Development Guidelines

### Code Structure

- Keep functions focused and concise
- Follow Go's standard package structure
- Write comprehensive tests for all new code
- Document exported types, functions, and methods
- Maintain backward compatibility when possible

### Pull Request Process

1. Understand the issue or feature requirements
2. Design a solution that aligns with architectural principles
3. Implement with comprehensive tests
4. Run `make test` and `make lint` before submitting
5. Create a focused pull request with clear description

## Common Tasks

### Creating a New Feature

1. Understand how it fits into the architecture
2. Review related components and interfaces
3. Design with distributed operation in mind
4. Implement with comprehensive tests
5. Document the feature for users and developers

### Fixing a Bug

1. Understand the root cause
2. Create a test that reproduces the issue
3. Fix the issue without breaking compatibility
4. Verify the fix with appropriate tests
5. Document the fix in the pull request

## Functional Testing

Test key functionality through CLI:

```bash
# Add a schema
defradb client schema add '
  type User {
    name: String 
    age: Int 
    verified: Boolean 
    points: Float
  }
'

# Create a document
defradb client query '
  mutation {
    create_User(input: {name: "Test", age: 30}) {
      _docID
    }
  }
'

# Query documents
defradb client query '
  query {
    User {
      _docID
      name
      age
    }
  }
'
```

## Key Files for Different Components

- **CLI Commands**: `cli/*.go`
- **HTTP API**: `http/*.go`
- **Database Core**: `internal/db/*.go`
- **Query Processing**: `internal/planner/*.go`, `internal/request/*.go`
- **CRDT Implementation**: `internal/core/crdt/*.go`
- **Schema System**: `internal/request/graphql/schema/*.go`
- **P2P Networking**: `net/*.go`
- **Access Control**: `acp/*.go`
- **Client API**: `client/*.go`

## Additional Resources

For deeper understanding of specific components, refer to the dedicated guides:
- Architecture Overview
- Query System
- MerkleCRDT System
- Access Control System
- Indexing System
- P2P Networking
- Schema System
- Development Guidelines

When implementing features or fixing bugs, focus on:
1. Preserving the user-centric design philosophy
2. Maintaining distributed-first architecture
3. Ensuring backward compatibility
4. Providing comprehensive tests
5. Documenting changes clearly