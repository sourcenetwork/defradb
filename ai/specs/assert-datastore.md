# Assert Datastore Development Plan

## Overview

Add a new `Datastore` test action to the DefraDB integration test framework that allows fetching values from the datastore for specific nodes. This action will use the DB interface's `Datastore()` method and support any key type that implements the `keys.Key` interface, providing a flexible builder pattern for key construction.

## Goals

1. Create a `Datastore` test action that can fetch values from a node's datastore
2. Support any key type that implements `keys.Key` interface (not just DataStoreKey)
3. Use `gomega.OmegaMatcher` for value assertions
4. Provide a builder pattern for flexible key construction
5. Support presence/absence checks for keys

## Design

### New Struct Definition

```go
// Datastore is an action that fetches a value from the datastore for a given key.
type Datastore struct {
    // NodeID may hold the ID (index) of a node to fetch from.
    // If not provided, the fetch will be done on all nodes.
    NodeID immutable.Option[int]

    // Key is built using the KeyBuilder interface.
    Key KeyBuilder

    // Value is the expected value matcher.
    // Uses gomega.OmegaMatcher for flexible assertions.
    // Should only be set if ExpectMissingKey is false.
    Value gomega.OmegaMatcher

    // ExpectMissingKey indicates whether the key should be missing.
    // If true, the test will verify the key doesn't exist.
    // If true and Value is set, test execution will error.
    ExpectMissingKey bool

    // ExpectedError is any error expected from the action.
    ExpectedError string
}

// KeyBuilder is an interface for building datastore keys.
type KeyBuilder interface {
    // Build constructs the actual key using the test state
    Build(s *state) (keys.Key, error)
}
```

### Builder Pattern Design

The following code will be implemented in `tests/integration/store_key.go`:

```go
// KeyFactory is the entry point for building different types of keys
type KeyFactory struct{}

// NewKey creates a new key factory
func NewKey() KeyFactory {
    return KeyFactory{}
}

// DatastoreDoc starts building a DataStoreKey for document data
func (f KeyFactory) DatastoreDoc() *DatastoreDocKeyBuilder {
    return &DatastoreDocKeyBuilder{}
}

// DatastoreIndex starts building an IndexDataStoreKey for index entries
func (f KeyFactory) DatastoreIndex() *DatastoreIndexKeyBuilder {
    return &DatastoreIndexKeyBuilder{}
}

// DatastoreDocKeyBuilder builds keys for document data in the datastore
type DatastoreDocKeyBuilder struct {
    collectionIndex int
    docIndex        int
    fieldName       string
    instanceType    keys.InstanceType
}

// Col sets the collection index
func (b *DatastoreDocKeyBuilder) Col(index int) *DatastoreDocKeyBuilder {
    b.collectionIndex = index
    return b
}

// DocID sets the document index
func (b *DatastoreDocKeyBuilder) DocID(index int) *DatastoreDocKeyBuilder {
    b.docIndex = index
    return b
}

// Field sets the field name
func (b *DatastoreDocKeyBuilder) Field(fieldName string) *DatastoreDocKeyBuilder {
    b.fieldName = fieldName
    return b
}

// InstanceType sets the instance type (default is keys.ValueKey)
func (b *DatastoreDocKeyBuilder) InstanceType(t keys.InstanceType) *DatastoreDocKeyBuilder {
    b.instanceType = t
    return b
}

// Build implements KeyBuilder interface
func (b *DatastoreDocKeyBuilder) Build(s *state) (keys.Key, error) {
    // Resolve collection index to collection ID
    col := s.nodes[0].collections[b.collectionIndex]
    collectionID := col.Version().CollectionID
    
    // Resolve doc index to doc ID
    docID := s.docIDs[b.collectionIndex][b.docIndex]
    
    // Resolve field name to field ID
    fieldID := getFieldIDByName(col.Definition(), b.fieldName)
    
    // Default instance type
    instanceType := b.instanceType
    if instanceType == "" {
        instanceType = keys.ValueKey
    }
    
    // Build DataStoreKey
    key := keys.DataStoreKey{
        CollectionShortID: collectionID,
        InstanceType:     instanceType,
        DocID:            docID.String(),
        FieldID:          fieldID,
    }
    
    return key, nil
}

// DatastoreIndexKeyBuilder builds keys for index entries in the datastore
type DatastoreIndexKeyBuilder struct {
    collectionIndex int
    indexID         int
    fields          []indexFieldValue
}

// Col sets the collection index
func (b *DatastoreIndexKeyBuilder) Col(index int) *DatastoreIndexKeyBuilder {
    b.collectionIndex = index
    return b
}

// IndexID sets the index ID
func (b *DatastoreIndexKeyBuilder) IndexID(id int) *DatastoreIndexKeyBuilder {
    b.indexID = id
    return b
}

// Field adds a field value to the index key.
// Can be called multiple times to add multiple fields in order.
// The order matters - fields should be added in the same order as defined in the index.
func (b *DatastoreIndexKeyBuilder) Field(value client.NormalValue, descending bool) *DatastoreIndexKeyBuilder {
    b.fields = append(b.fields, indexFieldValue{
        value:      value,
        descending: descending,
    })
    return b
}

// Build implements KeyBuilder interface
func (b *DatastoreIndexKeyBuilder) Build(s *state) (keys.Key, error) {
    // Resolve collection index to collection ID
    col := s.nodes[0].collections[b.collectionIndex]
    collectionID := col.Version().CollectionID
    
    // Convert to keys.IndexedField
    indexedFields := make([]keys.IndexedField, len(b.fields))
    for i, f := range b.fields {
        indexedFields[i] = keys.IndexedField{
            Value:      f.value,
            Descending: f.descending,
        }
    }
    
    // Build IndexDataStoreKey
    key := keys.IndexDataStoreKey{
        CollectionShortID: collectionID,
        IndexID:          uint32(b.indexID),
        Fields:           indexedFields,
    }
    
    return key, nil
}

// indexFieldValue represents a field value in an index (internal use only)
type indexFieldValue struct {
    value      client.NormalValue
    descending bool
}
```

### Implementation Flow

1. **Action Handler** (`performAction` in utils.go):
   - Add case for `Datastore` action
   - Call new `fetchDatastore` function

2. **Fetch Function** (`fetchDatastore`):
   - Validate ExpectMissingKey and Value combination
   - Get nodes based on NodeID
   - Build the key using KeyBuilder
   - Fetch value from node's datastore
   - Handle missing key case
   - Assert value using the gomega matcher if key exists
   - Handle errors appropriately

3. **Key Resolution**:
   - Resolve collection indices to actual collection IDs
   - Resolve document indices to actual document IDs
   - Build the appropriate key type based on builder configuration

### Data Flow

```
Test Case
    ↓
Datastore Action
    ↓
performAction (case Datastore)
    ↓
fetchDatastore()
    ↓
├── Validate ExpectMissingKey vs Value
├── Resolve NodeID(s)
├── Build Key using KeyBuilder
├── Get Datastore from DB
├── Fetch value with key
├── Check if key exists
├── If ExpectMissingKey: verify absence
├── Else: Execute gomega matcher on value
└── Handle errors
```

### Usage Examples

```go
// Example 1: Fetching a document field value (default InstanceType is ValueKey)
testUtils.Datastore{
    Key: testUtils.NewKey().
        DatastoreDoc().
        Col(0).
        DocID(0).
        Field("name"),
    Value: gomega.Equal("John"),
}

// Example 2: Checking an index entry with multiple fields
testUtils.Datastore{
    Key: testUtils.NewKey().
        DatastoreIndex().
        Col(0).
        IndexID(0).
        Field("John", false).
        Field(27, false),
    Value: gomega.Equal(someExpectedValue),
}

// Example 3: Verifying a key doesn't exist
testUtils.Datastore{
    Key: testUtils.NewKey().
        DatastoreDoc().
        Col(0).
        DocID(0).
        Field("deletedField"),
    ExpectMissingKey: true,
}

// Example 4: Using a different InstanceType
testUtils.Datastore{
    Key: testUtils.NewKey().
        DatastoreDoc().
        Col(0).
        DocID(0).
        Field("priority").
        InstanceType(keys.PriorityKey),
    Value: gomega.Equal(somePriorityValue),
}
```

### Integration Points

1. **With existing test state**:
   - Access `s.docIDs` for document index resolution
   - Access `s.nodes[].collections` for collection resolution
   - Use `s.nodes` for node access
   - Follow error handling patterns

2. **With keys package**:
   - Support `keys.DataStoreKey` 
   - Support `keys.IndexDataStoreKey`
   - Extensible to other key types implementing `keys.Key`
   - Use key encoding/building methods

3. **With gomega matchers**:
   - Support all existing gomega matchers
   - Follow the same pattern as Request action

4. **File organization**:
   - `test_case.go` - Contains the `Datastore` action struct and `KeyBuilder` interface
   - `store_key.go` - Contains all key builder implementations and helper functions
   - `utils.go` - Contains the `fetchDatastore` implementation

## Testing

Add a test in `tests/integration/mutation/create/simple_test.go` that:
1. Creates a document
2. Uses the Datastore action to fetch a field value
3. Asserts the value using a gomega matcher
4. Tests the ExpectMissingKey functionality

## Implementation Steps

1. Add `Datastore` struct and `KeyBuilder` interface to `test_case.go`
2. Create `store_key.go` file next to `test_case.go` with all key builder types:
   - `KeyFactory` struct and `NewKey()` function
   - `DatastoreDocKeyBuilder` and its methods
   - `DatastoreIndexKeyBuilder` and its methods
   - `indexFieldValue` struct (private)
   - Helper functions for key building
3. Add `fetchDatastore` function to `utils.go`
4. Update `performAction` to handle Datastore case
5. Create test in `simple_test.go` to verify functionality
6. Update any necessary imports

## Considerations

- The datastore returns raw encoded values, so the matcher may need to handle decoded values
- Different field types may require different decoding strategies
- Error handling should follow existing patterns in the test framework
- Support for all nodes vs specific node patterns
- Extensibility to new key types in the future
- Clear error messages when ExpectMissingKey and Value are both set