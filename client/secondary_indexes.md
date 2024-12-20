# Secondary indexing in DefraDB

DefraDB provides a powerful and flexible secondary indexing system that enables efficient document lookups and queries. This document explains the architecture, implementation details, and usage patterns of the indexing system.

## Overview

The indexing system consists of two main components:

- Index storage (handles storing and maintaining index information).
- Index-based document fetching (manages retrieving documents using these indexes). 

Together, these components provide a robust foundation for efficient data access patterns.

## Index storage

### Core types

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

### Key structure

Index keys in DefraDB follow a carefully designed format that enables efficient lookups and range scans. For regular indexes, the key format is:
```
<collection_id>/<index_id>(/<field_value>)+/<doc_id> -> empty value
``` 
Unique indexes follow a similar pattern but store the document ID as the value instead: 
```
<collection_id>/<index_id>(/<field_value>)+ -> <doc_id>
```

### Value encoding

DefraDB primarily uses CBOR for encoding. However, the indexing system uses a custom encoding/decoding solution, because CBOR doesn't guarantee ordering preservation, which is crucial for index functionality. 

Our custom encoding ensures that: 
- numeric values maintain their natural ordering,  
- strings are properly collated, and  
- complex types like arrays and objects have deterministic ordering.  

### Index maintenance

Index maintenance happens through three primary operations: document creation, updates, and deletion. When a new document is saved, the system indexes all configured fields, generating entries according to the key format and validating any unique constraints. During updates, the system carefully manages both the removal of old index entries and the creation of new ones, ensuring consistency through atomic transactions. For deletions, all associated index entries are cleaned up along with related metadata.

## Index-based document fetching

The IndexFetcher is central to document retrieval, managing the process through two phases. First, it retrieves indexed fields, such as document IDs. Then, it uses a standard fetcher to obtain any additional requested fields.  

For each query, the system generates specialized result iterators based on the document filter conditions. These iterators optimize how operations are handled:  
- For simple equality (_eq) or membership tests (_in), the iterator often constructs the exact keys directly.  
- For range operations (_gt, _le, etc.) or pattern matching (_like, etc.), dedicated value matchers are used to validate the results.  

The performance of these operations varies:  
- Direct match operations are typically the fastest, as they precisely target the required keys.  
- Range and pattern operations involve more work, scanning a range of keys and validating each result.  

The system is optimized to reduce key-value operations during mutations and minimize memory usage during result streaming.  

Note: the index fetcher can not benefit at the moment from ordered indexes, as the underlying storage does not support such range queries yet.

## Performance considerations

When working with indexes, it's important to understand their impact on system performance. Each index increases write amplification as every document modification must update all relevant indexes. However, this cost is often outweighed by the dramatic improvement in read performance for indexed queries.

Index selection should be driven by your query patterns and data distribution. Indexing fields that are frequently used in query filters can significantly improve performance, but indexing rarely-queried fields only adds overhead. For unique indexes, the additional validation requirements make this trade-off even more important to consider.

## Indexing related objects

DefraDB's indexing system is capable of handling relationships between documents. See example below.  

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

For queries on non-indexed relationships, the standard approach is a top-to-bottom strategy:  
1. All User documents are fetched.  
1. For each User document, the corresponding Address document is retrieved.  
This approach can be highly inefficient for large collections.  

With indexing, an inverted fetching strategy is used instead:  
1. The Address documents with the matching city value are fetched first.  
1. For each Address document, the corresponding User document is retrieved.  
This method is significantly more efficient, as the index allows direct retrieval of the relevant User documents.  

### Relationship cardinality using indexes

The indexing system also plays a crucial role in enforcing relationship cardinality. By marking an index as unique, you can enforce one-to-one relationships between documents. The code below shows how to modify the schema to ensure each user has exactly one address:  

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

## JSON field indexing

DefraDB implements a specialized indexing system for JSON fields that differs from how other field types are handled. While a document in DefraDB can contain various field types (Int, String, Bool, JSON, etc.), JSON fields require special treatment due to their hierarchical nature.

#### JSON interface

The indexing system relies on the `JSON` interface defined in `client/json.go`. This interface is crucial for handling JSON fields as it enables traversal of all leaf nodes within a JSON document. A `JSON` value in DefraDB can represent either an entire JSON document or a single node within it. Each `JSON` value maintains its path information, which is essential for indexing.

For example, given this JSON document:
```json
{
    "user": {
        "device": {
            "model": "iPhone"
        }
    }
}
```

The system can represent the "iPhone" value as a `JSON` type with its complete path `[]string{"user", "device", "model"}`. This path-aware representation is fundamental to how the indexing system works.

#### Inverted indexes for JSON

For JSON fields, DefraDB uses inverted indexes with the following key format:
```
<collection_id>/<index_id>(/<json_path>/<json_value>)+/<doc_id>
```

The term "inverted" comes from how these indexes reverse the typical document-to-value relationship. Instead of starting with a document and finding its values, we start with a value and can quickly find all documents containing that value at any path.

This approach differs from traditional secondary indexes in DefraDB. While regular fields map to single index entries, a JSON field generates multiple index entries - one for each leaf node in its structure. The system traverses the entire JSON structure during indexing, creating entries that combine the path and value information.

#### Value normalization and JSON

The indexing system integrates with DefraDB's value normalization through `client.NormalValue`. While the encoding/decoding package handles scalar types directly, JSON values maintain additional path information. Each JSON node is encoded with both its normalized value and its path information, allowing the system to reconstruct the exact location of any value within the JSON structure.

Similar to how other field types are normalized (e.g., integers to int64), JSON leaf values are normalized based on their type before being included in the index. This ensures consistent ordering and comparison operations.

#### Integration with index infrastructure

When a document with a JSON field is indexed, the system:
1. Uses the JSON interface to traverse the document structure
2. Creates index entries for each leaf node, combining path information with normalized values
3. Maintains all entries in a way that enables efficient querying at any depth

This implementation enables efficient queries like:
```graphql
query {
    Collection(filter: {
        jsonField: {
            user: {
                device: {
                    model: {_eq: "iPhone"}
                }
            }
        }
    })
}
```

The system can directly look up matching documents using the index entries, avoiding the need to scan and parse JSON content during query execution.