---
description: DefraDB schema system architecture and implementation guidelines
globs: "**/schema/**/*.go, **/graphql/schema/**/*.go"
alwaysApply: true
---

# Schema System Guide

DefraDB uses a schema-based approach to define document structure, relationships, and capabilities. This guide explains the schema system architecture, implementation details, and best practices for working with schemas.

## Core Concepts

### Schema Architecture

DefraDB's schema system is built on GraphQL type definitions with extensions:

1. **Collection Definitions**:
   - Define document structure and field types
   - Specify relationships between collections
   - Configure indexes and other features via directives

2. **Schema Versioning**:
   - Each schema has a unique version identifier
   - Multiple schema versions can exist for a collection
   - Collections track their active schema version

3. **Field Types**:
   - Basic scalar types (String, Int, Float, Boolean)
   - Complex types (arrays, JSON)
   - Relationship fields (references to other collections)
   - Special types (like vector embeddings)

## Schema Definition

Schemas are defined using GraphQL type definitions with DefraDB-specific directives:

```graphql
type User {
  name: String
  age: Int
  verified: Boolean
  points: Float
  address: Address @primary @index
  tags: [String]
  metadata: JSON
  embedding: Vector(384)
}

type Address {
  street: String
  city: String @index
  country: String
  user: User
}
```

### Key Directives

1. **@primary**:
   - Marks a relationship field as the primary side
   - Determines relationship ownership

2. **@index**:
   - Creates a secondary index on a field
   - Options: `@index(unique: true)` for unique constraints
   - Can be applied to scalar fields and relationships

3. **@policy**:
   - Associates a collection with an access control policy
   - Format: `@policy(id: "policyID", resource: "resourceName")`

## Implementation Architecture

### Schema Management Components

1. **Schema Parser**:
   - Parses GraphQL SDL into internal representations
   - Validates schema structure and relationships
   - Generates schema version identifiers

2. **Schema Registry**:
   - Stores and retrieves schema definitions
   - Manages schema versions
   - Tracks active schema versions per collection

3. **Collection Descriptions**:
   - Contains metadata about collections
   - Links collections to their schema versions
   - Stores indexing information

### Schema Storage

Schemas are stored in the system datastore:

1. **Schema Persistence**:
   - Schema definitions stored as serialized objects
   - Identified by content-based hashing
   - Referenced by collections

2. **Collection Metadata**:
   - Collection descriptions stored in system tables
   - Links from collection name to schema version
   - Maintains collection state information

### Schema Evolution

DefraDB supports schema evolution through:

1. **Schema Patching**:
   - Apply targeted changes to existing schemas
   - Add, remove, or modify fields
   - Update directives and configurations

2. **Version Transitions**:
   - Set active schema version for collections
   - Maintain backward compatibility
   - Support data migration as needed

## Implementation Guidelines

When working with the schema system:

### Adding New Field Types

1. Extend the type system in `internal/encoding/type.go`
2. Implement encoding/decoding for the new type
3. Update schema parsing to recognize the new type
4. Add validation logic for the new type
5. Update GraphQL generation to support the new type
6. Create comprehensive tests covering:
   - Schema parsing
   - Value validation
   - Query/mutation operations
   - Serialization/deserialization

### Adding New Directives

1. Update the schema parser to recognize the directive
2. Implement directive processing logic
3. Add validation for directive usage
4. Update collection creation to apply directive effects
5. Test directive behavior in various scenarios

### Schema Migration Considerations

1. Design for backward compatibility when possible
2. Create data migration strategies for breaking changes
3. Document migration paths clearly
4. Test migration thoroughly with representative data
5. Consider performance impacts for large collections

## Relationship Modeling

DefraDB supports sophisticated relationship modeling:

### Relationship Types

1. **One-to-One**:
   - Implemented with unique indexes on relationship fields
   - Example: User with one Address

2. **One-to-Many**:
   - Standard relationship without unique constraints
   - Example: Author with multiple Books

3. **Many-to-Many**:
   - Typically implemented with an intermediary collection
   - Example: Students and Courses through Enrollments

### Relationship Implementation

Relationships are stored using document references:

1. **Primary Side**:
   - Marked with `@primary` directive
   - Acts as the "owner" of the relationship
   - Controls relationship lifecycle

2. **Secondary Side**:
   - References back to the primary document
   - Maintains consistent bidirectional navigation
   - Updated automatically when primary changes

### Self-Referential Relationships

DefraDB supports self-referential relationships:

```graphql
type Person {
  name: String
  parent: Person @primary
  children: [Person]
}
```

These enable hierarchical or graph-like data structures within a single collection.

## Testing Schema Implementations

When implementing or modifying schema components:

1. **Schema Parsing Tests**:
   - Test valid schema parsing
   - Verify validation errors for invalid schemas
   - Test directive processing
   - Check schema versioning logic

2. **Schema Evolution Tests**:
   - Test schema patching operations
   - Verify version transitions
   - Test backward compatibility
   - Verify data access across schema versions

3. **Relationship Tests**:
   - Test relationship creation and navigation
   - Verify bidirectional consistency
   - Test cascading operations
   - Verify cardinality constraints

4. **Integration Tests**:
   - Test schema functionality in the complete system
   - Verify interaction with queries, mutations, and indexes
   - Test schema persistence and recovery

## Advanced Schema Features

### Default Values

Fields can specify default values:

```graphql
type Product {
  name: String
  inStock: Boolean = true
  createdAt: DateTime = "now()"
}
```

Default values are applied during document creation when fields are omitted.

### JSON Schema Support

The JSON field type provides flexible schema-less storage:

```graphql
type Document {
  title: String
  content: JSON
}
```

This allows storing arbitrary structured data while maintaining queryability through JSON indexes.

### Vector Embeddings

DefraDB supports vector embeddings for AI applications:

```graphql
type Document {
  title: String
  embedding: Vector(384)
}
```

This enables similarity search and other vector operations on embedding fields.

## GraphQL Schema Generation

DefraDB automatically generates a GraphQL API from collection schemas:

1. **Query Types**:
   - Collection-based query endpoints
   - Filter, ordering, and pagination support
   - Relationship traversal
   - Aggregate operations

2. **Mutation Types**:
   - Create, update, and delete operations
   - Batch mutations
   - Nested mutations for relationships

3. **Input Types**:
   - Input object types for create/update operations
   - Filter input types for queries
   - Specialized input types for operations

The generated schema provides a consistent, intuitive API that maps directly to the defined collections and relationships.