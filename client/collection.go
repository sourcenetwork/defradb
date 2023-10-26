// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package client

import (
	"context"

	"github.com/sourcenetwork/defradb/datastore"
)

// CollectionDefinition contains the metadata defining what a Collection is.
type CollectionDefinition struct {
	// Description returns the CollectionDescription of this Collection.
	Description CollectionDescription `json:"description"`
	// Schema returns the SchemaDescription used to define this Collection.
	Schema SchemaDescription `json:"schema"`
}

// Collection represents a defradb collection.
//
// A Collection is mostly analogous to a SQL table, however a collection is specific to its
// host, and many collections may share the same schema.
//
// Many functions on this object will interact with the underlying datastores.
type Collection interface {
	// Name returns the name of this collection.
	Name() string
	// ID returns the ID of this Collection.
	ID() uint32
	// SchemaRoot returns the Root of the Schema used to define this Collection.
	SchemaRoot() string

	// Definition contains the metadata defining what a Collection is.
	Definition() CollectionDefinition
	// Schema returns the SchemaDescription used to define this Collection.
	Schema() SchemaDescription
	// Description returns the CollectionDescription of this Collection.
	Description() CollectionDescription

	// Create a new document.
	//
	// Will verify the DocKey/CID to ensure that the new document is correctly formatted.
	Create(context.Context, *Document) error
	// CreateMany new documents.
	//
	// Will verify the DocKeys/CIDs to ensure that the new documents are correctly formatted.
	CreateMany(context.Context, []*Document) error
	// Update an existing document with the new values.
	//
	// Any field that needs to be removed or cleared should call doc.Clear(field) before.
	// Any field that is nil/empty that hasn't called Clear will be ignored.
	//
	// Will return a ErrDocumentNotFound error if the given document is not found.
	Update(context.Context, *Document) error
	// Save the given document in the database.
	//
	// If a document exists with the given DocKey it will update it. Otherwise a new document
	// will be created.
	Save(context.Context, *Document) error
	// Delete will attempt to delete a document by key.
	//
	// Will return true if a deletion is successful, and return false along with an error
	// if it cannot. If the document doesn't exist, then it will return false and a ErrDocumentNotFound error.
	// This operation will hard-delete all state relating to the given DocKey. This includes data, block, and head storage.
	Delete(context.Context, DocKey) (bool, error)
	// Exists checks if a given document exists with supplied DocKey.
	//
	// Will return true if a matching document exists, otherwise will return false.
	Exists(context.Context, DocKey) (bool, error)

	// UpdateWith updates a target document using the given updater type.
	//
	// Target can be a Filter statement, a single docKey, a single document,
	// an array of docKeys, or an array of documents.
	// It is recommended to use the respective typed versions of Update
	// (e.g. UpdateWithFilter or UpdateWithKey) over this function if you can.
	//
	// Returns an ErrInvalidUpdateTarget error if the target type is not supported.
	// Returns an ErrInvalidUpdater error if the updater type is not supported.
	UpdateWith(ctx context.Context, target any, updater string) (*UpdateResult, error)
	// UpdateWithFilter updates using a filter to target documents for update.
	//
	// The provided updater must be a string Patch, string Merge Patch, a parsed Patch, or parsed Merge Patch
	// else an ErrInvalidUpdater will be returned.
	UpdateWithFilter(ctx context.Context, filter any, updater string) (*UpdateResult, error)
	// UpdateWithKey updates using a DocKey to target a single document for update.
	//
	// The provided updater must be a string Patch, string Merge Patch, a parsed Patch, or parsed Merge Patch
	// else an ErrInvalidUpdater will be returned.
	//
	// Returns an ErrDocumentNotFound if a document matching the given DocKey is not found.
	UpdateWithKey(ctx context.Context, key DocKey, updater string) (*UpdateResult, error)
	// UpdateWithKeys updates documents matching the given DocKeys.
	//
	// The provided updater must be a string Patch, string Merge Patch, a parsed Patch, or parsed Merge Patch
	// else an ErrInvalidUpdater will be returned.
	//
	// Returns an ErrDocumentNotFound if a document is not found for any given DocKey.
	UpdateWithKeys(context.Context, []DocKey, string) (*UpdateResult, error)

	// DeleteWith deletes a target document.
	//
	// Target can be a Filter statement, a single docKey, a single document, an array of docKeys,
	// or an array of documents. It is recommended to use the respective typed versions of Delete
	// (e.g. DeleteWithFilter or DeleteWithKey) over this function if you can.
	// This operation will soft-delete documents related to the given DocKey and update the composite block
	// with a status of `Deleted`.
	//
	// Returns an ErrInvalidDeleteTarget if the target type is not supported.
	DeleteWith(ctx context.Context, target any) (*DeleteResult, error)
	// DeleteWithFilter deletes documents matching the given filter.
	//
	// This operation will soft-delete documents related to the given filter and update the composite block
	// with a status of `Deleted`.
	DeleteWithFilter(ctx context.Context, filter any) (*DeleteResult, error)
	// DeleteWithKey deletes using a DocKey to target a single document for delete.
	//
	// This operation will soft-delete documents related to the given DocKey and update the composite block
	// with a status of `Deleted`.
	//
	// Returns an ErrDocumentNotFound if a document matching the given DocKey is not found.
	DeleteWithKey(context.Context, DocKey) (*DeleteResult, error)
	// DeleteWithKeys deletes documents matching the given DocKeys.
	//
	// This operation will soft-delete documents related to the given DocKeys and update the composite block
	// with a status of `Deleted`.
	//
	// Returns an ErrDocumentNotFound if a document is not found for any given DocKey.
	DeleteWithKeys(context.Context, []DocKey) (*DeleteResult, error)

	// Get returns the document with the given DocKey.
	//
	// Returns an ErrDocumentNotFound if a document matching the given DocKey is not found.
	Get(ctx context.Context, key DocKey, showDeleted bool) (*Document, error)

	// WithTxn returns a new instance of the collection, with a transaction
	// handle instead of a raw DB handle.
	WithTxn(datastore.Txn) Collection

	// GetAllDocKeys returns all the document keys that exist in the collection.
	GetAllDocKeys(ctx context.Context) (<-chan DocKeysResult, error)

	// CreateIndex creates a new index on the collection.
	// `IndexDescription` contains the description of the index to be created.
	// `IndexDescription.Name` must start with a letter or an underscore and can
	// only contain letters, numbers, and underscores.
	// If the name of the index is not provided, it will be generated.
	CreateIndex(context.Context, IndexDescription) (IndexDescription, error)

	// DropIndex drops an index from the collection.
	DropIndex(ctx context.Context, indexName string) error

	// GetIndexes returns all the indexes that exist on the collection.
	GetIndexes(ctx context.Context) ([]IndexDescription, error)
}

// DocKeysResult wraps the result of an attempt at a DocKey retrieval operation.
type DocKeysResult struct {
	// If a DocKey was successfully retrieved, this will be that key.
	Key DocKey
	// If an error was generated whilst attempting to retrieve the DocKey, this will be the error.
	Err error
}

// UpdateResult wraps the result of an update call.
type UpdateResult struct {
	// Count contains the number of documents updated by the update call.
	Count int64
	// DocKeys contains the DocKeys of all the documents updated by the update call.
	DocKeys []string
}

// DeleteResult wraps the result of an delete call.
type DeleteResult struct {
	// Count contains the number of documents deleted by the delete call.
	Count int64
	// DocKeys contains the DocKeys of all the documents deleted by the delete call.
	DocKeys []string
}

// P2PCollection is the gRPC response representation of a P2P collection topic
type P2PCollection struct {
	// The collection ID
	ID string
	// The Collection name
	Name string
}
