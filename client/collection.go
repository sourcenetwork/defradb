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

	"github.com/sourcenetwork/defradb/client/options"
)

// Collection represents a defradb collection.
//
// A Collection is mostly analogous to a SQL table, however a collection is specific to its
// host, and many collections may share the same schema.
//
// Many functions on this object will interact with the underlying datastores.
type Collection interface {
	// Name returns the name of this collection.
	Name() string

	// VersionID returns the VersionID of this Collection.
	VersionID() string

	// CollectionID returns the Root of the Schema used to define this Collection.
	CollectionID() string

	// Version returns the CollectionVersion of this Collection.
	Version() CollectionVersion

	// Create a new document.
	//
	// Will verify the DocID/CID to ensure that the new document is correctly formatted.
	Create(ctx context.Context, doc *Document, opts ...options.Enumerable[options.CollectionCreateOptions]) error

	// CreateMany new documents.
	//
	// Will verify the DocIDs/CIDs to ensure that the new documents are correctly formatted.
	CreateMany(ctx context.Context, docs []*Document, opts ...options.Enumerable[options.CollectionCreateOptions]) error

	// Update an existing document with the new values.
	//
	// Any field that needs to be removed or cleared should call doc.Clear(field) before.
	// Any field that is nil/empty that hasn't called Clear will be ignored.
	//
	// Will return a ErrDocumentNotFound error if the given document is not found.
	Update(ctx context.Context, docs *Document, opts ...options.Enumerable[options.CollectionUpdateOptions]) error

	// Save the given document in the database.
	//
	// If a document exists with the given DocID it will update it. Otherwise a new document
	// will be created.
	Save(ctx context.Context, doc *Document, opts ...options.Enumerable[options.CollectionSaveOptions]) error

	// Delete will attempt to delete a document by DocID.
	//
	// Will return true if a deletion is successful, and return false along with an error
	// if it cannot. If the document doesn't exist, then it will return false and a ErrDocumentNotFound error.
	// This operation will hard-delete all state relating to the given DocID.
	// This includes data, block, and head storage.
	Delete(ctx context.Context, docID DocID, opts ...options.Enumerable[options.CollectionDeleteOptions]) (bool, error)

	// Exists checks if a given document exists with supplied DocID.
	//
	// Will return true if a matching document exists, otherwise will return false.
	Exists(ctx context.Context, docID DocID, opts ...options.Enumerable[options.CollectionExistsOptions]) (bool, error)

	// UpdateWithFilter updates using a filter to target documents for update.
	//
	// The provided updater must be a string Patch, string Merge Patch, a parsed Patch, or parsed Merge Patch
	// else an ErrInvalidUpdater will be returned.
	UpdateWithFilter(
		ctx context.Context,
		filter any,
		updater string,
		opts ...options.Enumerable[options.CollectionUpdateWithFilterOptions],
	) (*UpdateResult, error)

	// DeleteWithFilter deletes documents matching the given filter.
	//
	// This operation will soft-delete documents related to the given filter and update the composite block
	// with a status of `Deleted`.
	DeleteWithFilter(
		ctx context.Context,
		filter any,
		opts ...options.Enumerable[options.CollectionDeleteWithFilterOptions],
	) (*DeleteResult, error)

	// Get returns the document with the given DocID.
	//
	// Returns an ErrDocumentNotFound if a document matching the given DocID is not found.
	Get(
		ctx context.Context,
		docID DocID,
		opts ...options.Enumerable[options.CollectionGetOptions],
	) (*Document, error)

	// GetAllDocIDs returns all the document IDs that exist in the collection.
	GetAllDocIDs(
		ctx context.Context,
		opts ...options.Enumerable[options.CollectionGetAllDocIDsOptions],
	) (<-chan DocIDResult, error)

	// CreateIndex creates a new index on the collection.
	// `IndexDescription` contains the description of the index to be created.
	// `IndexDescription.Name` must start with a letter or an underscore and can
	// only contain letters, numbers, and underscores.
	// If the name of the index is not provided, it will be generated.
	// WARNING: This method can not create index for a collection that has a policy.
	CreateIndex(
		context.Context,
		IndexCreateRequest,
		...options.Enumerable[options.CollectionCreateIndexOptions],
	) (IndexDescription, error)

	// DropIndex drops an index from the collection.
	DropIndex(ctx context.Context, indexName string, opts ...options.Enumerable[options.CollectionDropIndexOptions]) error

	// GetIndexes returns all the indexes that exist on the collection.
	GetIndexes(
		ctx context.Context,
		opts ...options.Enumerable[options.CollectionGetIndexesOptions],
	) ([]IndexDescription, error)

	// AddEncryptedIndex adds a new encrypted index to the collection.
	AddEncryptedIndex(
		ctx context.Context,
		desc EncryptedIndexDescription,
		opts ...options.Enumerable[options.AddEncryptedIndexOptions],
	) (EncryptedIndexDescription, error)

	// DeleteEncryptedIndex deletes an encrypted index from the collection.
	DeleteEncryptedIndex(ctx context.Context, fieldName string) error

	// ListEncryptedIndexes returns all the encrypted indexes that exist on the collection.
	ListEncryptedIndexes(
		ctx context.Context,
		opts ...options.Enumerable[options.CollectionListEncryptedIndexesOptions],
	) ([]EncryptedIndexDescription, error)

	// Truncate this collection, permanently deleting all document state on this node.
	//
	// Changes made by this call will not impact other nodes, and cannot be synced to them over the P2P
	// system.
	//
	// This call will lock the collection, and no other read or write document operations on this collection
	// will progress whilst this is executing.
	Truncate(ctx context.Context, opts ...options.Enumerable[options.CollectionTruncateOptions]) error
}

// DocIDResult wraps the result of an attempt at a DocID retrieval operation.
type DocIDResult struct {
	// If a DocID was successfully retrieved, this will be that DocID.
	ID DocID
	// If an error was generated whilst attempting to retrieve the DocID, this will be the error.
	Err error
}

// UpdateResult wraps the result of an update call.
type UpdateResult struct {
	// Count contains the number of documents updated by the update call.
	Count int64
	// DocIDs contains the DocIDs of all the documents updated by the update call.
	DocIDs []string
}

// DeleteResult wraps the result of an delete call.
type DeleteResult struct {
	// Count contains the number of documents deleted by the delete call.
	Count int64
	// DocIDs contains the DocIDs of all the documents deleted by the delete call.
	DocIDs []string
}

// P2PCollection is the gRPC response representation of a P2P collection topic
type P2PCollection struct {
	// The collection ID
	ID string
	// The Collection name
	Name string
}
