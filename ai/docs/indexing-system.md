---
description: DefraDB's indexing system architecture and implementation guidelines
globs: "**/index*.go, **/fetcher/**/*.go"
alwaysApply: true
---

# Indexing System Guide

DefraDB's indexing system enables efficient document lookups and queries across collections. This guide explains the architecture, implementation details, and best practices for working with indexes.

## Core Concepts

### Index Types

DefraDB supports several types of indexes:

1. **Regular Indexes**:
   - Enable fast lookups by field values
   - Support composite (multi-field) configurations
   - Can be defined in ascending or descending order

2. **Unique Indexes**:
   - Enforce uniqueness constraints on field values
   - Ensure no duplicate values exist within a collection
   - Essential for implementing one-to-one relationships

3. **Array Indexes**:
   - Index individual elements within array fields
   - Enable efficient queries against array contents
   - Support both regular and unique constraints

4. **JSON Indexes**:
   - Index leaf nodes in JSON document structures
   - Enable deep querying into nested JSON data
   - Use path-aware representation for precision

5. **Relationship Indexes**:
   - Index relationship fields between collections
   - Enable efficient cross-collection queries
   - Support both primary and secondary relationships

## Index Architecture

### Key Components

1. **Index Descriptions**:
   ```go
   type IndexedFieldDescription struct {
       Name string       // Field name being indexed
       Descending bool   // Whether field is indexed in descending order
   }
   
   type IndexDescription struct {
       Name string                      // Index name
       ID uint32                        // Local index identifier
       Fields []IndexedFieldDescription // Fields being indexed
       Unique bool                      // Whether index enforces uniqueness
   }
   ```

2. **Collection Indexes**:
   ```go
   type CollectionIndex interface {
       Save(context.Context, datastore.Txn, *Document) error
       Update(context.Context, datastore.Txn, *Document, *Document) error
       Delete(context.Context, datastore.Txn, *Document) error
       Name() string
       Description() IndexDescription
   }
   ```

3. **Index Fetchers**:
   - Specialized document fetchers that use indexes
   - Generate optimized query plans based on indexes
   - Implement different strategies based on query operators

### Key Format

Indexes use carefully designed key formats:

1. **Regular Indexes**:
   ```
   <collection_id>/<index_id>(/<field_value>)+/<doc_id> -> empty value
   ```

2. **Unique Indexes**:
   ```
   <collection_id>/<index_id>(/<field_value>)+ -> <doc_id>
   ```

3. **JSON Indexes**:
   ```
   <collection_id>/<index_id>(/<json_path>/<json_value>)+/<doc_id>
   ```

### Value Encoding

DefraDB uses custom encoding for index values to ensure:
- Numeric values maintain natural ordering
- Strings are properly collated
- Complex types have deterministic ordering
- Binary comparison produces correct results

## Implementation Details

### Index Management Lifecycle

1. **Creation**:
   - Defined via Schema (with @index directive)
   - Created programmatically via API
   - System builds initial index entries for all documents

2. **Maintenance**:
   - Automatically updated during document modifications
   - Transactions ensure consistency
   - Unique constraints validated during updates

3. **Querying**:
   - Query planner detects applicable indexes
   - Generates optimized execution plan
   - Uses direct lookups when possible

4. **Deletion**:
   - Removes all index entries
   - Cleans up metadata
   - Handled atomically in transactions

### Query Optimization

The indexing system optimizes queries through:

1. **Prefix Matching**:
   - Uses index prefixes for efficient lookups
   - Supports partial index usage for composite indexes
   - Optimizes range queries with index bounds

2. **Operator Mapping**:
   - Maps query operators to index operations:
     - Equality (`_eq`) → Direct lookup
     - Range operators (`_gt`, `_lt`, etc.) → Range scans
     - Membership (`_in`) → Multiple direct lookups
     - Pattern matching (`_like`) → Range scan with filter

3. **Index Selection**:
   - Chooses most selective index when multiple are available
   - Considers filter conditions and index fields
   - May use compound indexes for multi-field conditions

## Implementation Guidelines

When working with the indexing system:

### Adding New Index Types

1. Implement the `CollectionIndex` interface
2. Define clear key format and encoding strategy
3. Ensure transactional consistency
4. Add query planner integration
5. Create comprehensive tests covering:
   - Creation and deletion
   - Document operations (create, update, delete)
   - Query performance
   - Edge cases (null values, special characters)

### Extending Existing Indexes

1. Maintain backward compatibility
2. Consider performance impacts of changes
3. Update both storage and retrieval components
4. Document any format changes

### Performance Considerations

1. **Write Amplification**:
   - Each additional index increases write overhead
   - Consider cost/benefit for frequently updated collections
   - Limit indexes to fields actually used in queries

2. **Index Selectivity**:
   - Indexes on fields with high cardinality (many unique values) are more selective
   - Composite indexes should start with more selective fields
   - Consider value distribution in indexed fields

3. **Memory Usage**:
   - Indexes increase memory footprint
   - Consider index size when working with large collections
   - Profile memory usage with representative data volumes

## Relationship Index Patterns

Indexes are particularly powerful for relationship queries:

```graphql
type User {
    name: String 
    address: Address @primary @index
} 

type Address {
    user: User
    city: String @index
    street: String 
}
```

This enables efficient queries like:

```graphql
query {
    User(filter: {
        address: {city: {_eq: "Montreal"}}
    }) {
        name
    }
}
```

The system uses:
1. The city index to find matching Address documents
2. The relationship index to find corresponding User documents
3. Avoids scanning all User documents

### Relationship Cardinality

Indexes also help enforce relationship cardinality:

```graphql
type User {
    address: Address @primary @index(unique: true)
} 
```

This unique index ensures each Address can only be referenced by one User.

## Testing Index Implementations

When implementing or modifying indexes:

1. **Verify Basic Operations**:
   - Test creation, updates, and deletion
   - Verify correct handling of null values
   - Test with various data types and edge cases

2. **Test Query Performance**:
   - Compare performance with and without indexes
   - Measure performance with different data volumes
   - Test various query patterns (equality, range, pattern matching)

3. **Test Unique Constraints**:
   - Verify unique violations are caught
   - Test edge cases (null values, updates vs. inserts)
   - Test concurrent modifications

4. **Integration Testing**:
   - Verify behavior with the complete DefraDB system
   - Test interaction with other components (queries, transactions, P2P)
   - Confirm correct serialization/deserialization

## JSON Indexing Special Considerations

JSON indexing differs from regular field indexing:

1. **Path-Aware Indexing**:
   - Indexes track both values and their paths
   - Enables precise queries at any nesting level
   - Maintains path information in index keys

2. **Leaf Node Strategy**:
   - Only leaf nodes (final values) are indexed
   - Each JSON document generates multiple index entries
   - Enables queries against any path in the structure

3. **Value Normalization**:
   - JSON values are normalized before indexing
   - Maintains consistent ordering regardless of source format
   - Handles various JSON value types appropriately