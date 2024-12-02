# Secondary Indexing in DefraDB

DefraDB provides a powerful and flexible secondary indexing system that enables efficient document lookups and queries. This document explains the architecture, implementation details, and usage patterns of the indexing system.

## Overview

The indexing system consists of two main components. The first is index storage, which handles storing and maintaining index information. The second is index-based document fetching, which manages retrieving documents using these indexes. Together, these components provide a robust foundation for efficient data access patterns.

## Index Storage

### Core Types

The indexing system is built around several key types that define how indexes are structured and managed. At its heart is the IndexedFieldDescription, which describes a single field being indexed, including its name and whether it should be ordered in descending order. These field descriptions are combined into an IndexDescription, which provides a complete picture of an index including its name, ID, fields, and whether it enforces uniqueness.

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

The CollectionIndex interface ties everything together by defining the core operations that any index must support. This interface is implemented by different index types such as regular indexes, unique indexes, and array indexes, allowing each to provide specific behaviors while maintaining a consistent interface.

```go
type CollectionIndex interface {
    Save(context.Context, datastore.Txn, *Document) error
    Update(context.Context, datastore.Txn, *Document, *Document) error
    Delete(context.Context, datastore.Txn, *Document) error
    Name() string
    Description() IndexDescription
}
```

### Key Structure

Index keys in DefraDB follow a carefully designed format that enables efficient lookups and range scans. For regular indexes, the key format is:
`<collection_id>/<index_id>/<field_value_1>/..<field_value_n>/<doc_id>` -> empty value. 
Unique indexes follow a similar pattern but store the document ID as the value instead: 
`<collection_id>/<index_id>/<field_value_1>/..<field_value_n>` -> `<doc_id>`.

### Value Encoding

While DefraDB primarily uses CBOR for encoding, the indexing system employs a custom encoding/decoding solution inspired by CockroachDB. This decision was made because CBOR doesn't guarantee ordering preservation, which is crucial for index functionality. Our custom encoding ensures that numeric values maintain their natural ordering, strings are properly collated, and complex types like arrays and objects have deterministic ordering.

### Index Maintenance

Index maintenance happens through three primary operations: document creation, updates, and deletion. When a new document is saved, the system indexes all configured fields, generating entries according to the key format and validating any unique constraints. During updates, the system carefully manages both the removal of old index entries and the creation of new ones, ensuring consistency through atomic transactions. For deletions, all associated index entries are cleaned up along with related metadata.

## Index-Based Document Fetching

The IndexFetcher is the cornerstone of document retrieval, orchestrating the process of fetching documents using indexes. It operates in two phases: first retrieving indexed fields (including document IDs), then using a standard fetcher to get any additional requested fields.

For each query, the system creates specialized result iterators based on the document filter conditions. These iterators are smart about how they handle different types of operations. For simple equality comparisons (`_eq`) or membership tests (`_in`), the iterator can often directly build the exact keys needed. For range operations (`_gt`, `_le`, ...) or pattern matching (`_like`, ...), the system employs dedicated value matchers to validate the results.

The performance characteristics of these operations vary. Direct match operations are typically the fastest as they can precisely target the needed keys. Range and pattern operations require more work as they must scan a range of keys and validate each result. The system is designed to minimize both key-value operations during mutations and memory usage during result streaming.

Note: the index fetcher can not benefit at the moment from ordered indexes, as the underlying storage does not support such range queries yet.

## Performance Considerations

When working with indexes, it's important to understand their impact on system performance. Each index increases write amplification as every document modification must update all relevant indexes. However, this cost is often outweighed by the dramatic improvement in read performance for indexed queries.

Index selection should be driven by your query patterns and data distribution. Indexing fields that are frequently used in query filters can significantly improve performance, but indexing rarely-queried fields only adds overhead. For unique indexes, the additional validation requirements make this trade-off even more important to consider.

## Indexing Related Objects

DefraDB's indexing system provides powerful capabilities for handling relationships between documents. Let's explore how this works with a practical example.

Consider a schema defining a relationship between Users and Addresses:

```graphql
type User {
    name: String 
    age: Int
    address: Address @primary @index
} 

type Address {
    user: User
    city: String @index
    street: String 
}
```

In this schema, we've defined two important indexes. First, we have an index on the `Address`'s city field, and second, we have an index on the relationship between `User` and `Address`. This setup enables extremely efficient querying across the relationship. For example:

```graphql
query {
    User(filter: {
        address: {city: {_eq: "Montreal"}}
    }) {
        name
    }
}
```

For requests on not indexed relations, the normal approach is from top to bottom, meaning that first all `User` documents are fetched and then for each `User` document the corresponding `Address` document is fetched. This can be very inefficient for large collections.
With indexing, we use so called inverted fetching, meaning that we first fetch the `Address` documents with the matching `city` value and then for each `Address` document the corresponding `User` document is fetched. This is much more efficient as we can use the index to directly fetch the `User` document.

### Relationship Cardinality Through Indexes

The indexing system also plays a crucial role in enforcing relationship cardinality. By marking an index as unique, you can enforce one-to-one relationships between documents. Here's how you would modify the schema to ensure each User has exactly one Address:

```graphql
type User {
    name: String 
    age: Int
    address: Address @primary @index(unique: true)
} 

type Address {
    user: User
    city: String @index
    street: String 
}
```

The unique index constraint ensures that no two Users can reference the same Address document. Without the unique constraint, the relationship would be one-to-many by default, allowing multiple Users to reference the same Address.
