package schema

import (
	ds "github.com/ipfs/go-datastore"
)

// SchemaStore provides storage for schema, and acts as a cache for generated schemas
// 
type SchemaStore struct {
	store ds.Datastore
}

// NewSchemaStore creates a new instance of a schema store
// using the provided Datastore backend.
func NewSchemaStore(backend ds.Datastore) *SchemaStore {
	
}