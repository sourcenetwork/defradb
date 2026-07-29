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
// host, and many collection versions may share the same definition.
//
// Many functions on this object will interact with the underlying datastores.
type Collection interface {
	// Name returns the name of this collection.
	Name() string

	// VersionID returns the VersionID of this Collection.
	VersionID() string

	// CollectionID returns the Root of the definition used to define this Collection.
	CollectionID() string

	// Version returns the CollectionVersion of this Collection.
	Version() CollectionVersion

	// AddDocument adds a new document.
	//
	// Will verify the DocID/CID to ensure that the new document is correctly formatted.
	AddDocument(ctx context.Context, doc *Document, opts ...options.Enumerable[options.AddDocumentOptions]) error

	// AddManyDocuments adds new documents.
	//
	// Will verify the DocIDs/CIDs to ensure that the new documents are correctly formatted.
	AddManyDocuments(ctx context.Context, docs []*Document, opts ...options.Enumerable[options.AddDocumentOptions]) error

	// UpdateDocument updates an existing document with the new values.
	//
	// Any field that needs to be removed or cleared should call doc.Clear(field) before.
	// Any field that is nil/empty that hasn't called Clear will be ignored.
	//
	// Will return a ErrDocumentNotFound error if the given document is not found.
	UpdateDocument(ctx context.Context, docs *Document, opts ...options.Enumerable[options.UpdateDocumentOptions]) error

	// SaveDocument saves the given document in the database.
	//
	// If a document exists with the given DocID it will update it. Otherwise a new document
	// will be created.
	SaveDocument(ctx context.Context, doc *Document, opts ...options.Enumerable[options.SaveDocumentOptions]) error

	// DeleteDocument will attempt to delete a document by DocID.
	//
	// Will return true if a deletion is successful, and return false along with an error
	// if it cannot. If the document doesn't exist, then it will return false and a ErrDocumentNotFound error.
	// This operation will hard-delete all state relating to the given DocID.
	// This includes data, block, and head storage.
	DeleteDocument(
		ctx context.Context,
		docID DocID,
		opts ...options.Enumerable[options.DeleteDocumentOptions],
	) (bool, error)

	// ExistsDocument checks if a given document exists with supplied DocID.
	//
	// Will return true if a matching document exists, otherwise will return false.
	ExistsDocument(
		ctx context.Context,
		docID DocID,
		opts ...options.Enumerable[options.ExistsDocumentOptions],
	) (bool, error)

	// UpdateDocumentsWithFilter updates using a filter to target documents for update.
	//
	// The provided updater must be a string Patch, string Merge Patch, a parsed Patch, or parsed Merge Patch
	// else an ErrInvalidUpdater will be returned.
	UpdateDocumentsWithFilter(
		ctx context.Context,
		filter any,
		updater string,
		opts ...options.Enumerable[options.UpdateDocumentsWithFilterOptions],
	) (*UpdateResult, error)

	// DeleteDocumentsWithFilter deletes documents matching the given filter.
	//
	// This operation will soft-delete documents related to the given filter and update the composite block
	// with a status of `Deleted`.
	DeleteDocumentsWithFilter(
		ctx context.Context,
		filter any,
		opts ...options.Enumerable[options.DeleteDocumentsWithFilterOptions],
	) (*DeleteResult, error)

	// GetDocument returns the document with the given DocID.
	//
	// Returns an ErrDocumentNotFound if a document matching the given DocID is not found.
	GetDocument(
		ctx context.Context,
		docID DocID,
		opts ...options.Enumerable[options.GetDocumentOptions],
	) (*Document, error)

	// NewIndex makes a new index on the collection.
	// `NewIndexRequest` contains the description of the new index.
	// `NewIndexRequest.Name` must start with a letter or an underscore and can
	// only contain letters, numbers, and underscores.
	// If the name of the index is not provided, it will be generated.
	// WARNING: This method can not make a new index for a collection that has a policy.
	//
	// The index is built in the background: this returns once the index is recorded, before existing
	// documents are indexed. The index starts in a "building" state and becomes "ready" once the
	// backfill completes, or "failed" if it cannot. While building, queries that would use it instead
	// full-scan and still return correct results. A build failure is therefore not returned here;
	// check the index's status via ListIndexes to learn the outcome. A failed index must be deleted
	// before it can be recreated.
	NewIndex(
		context.Context,
		NewIndexRequest,
		...options.Enumerable[options.NewCollectionIndexOptions],
	) (IndexDescription, error)

	// DeleteIndex deletes an index from the collection.
	//
	// The index's entries are removed in the background: this returns once the deletion is recorded,
	// before all entries are removed. The index stops being used for queries immediately.
	DeleteIndex(
		ctx context.Context,
		indexName string,
		opts ...options.Enumerable[options.DeleteCollectionIndexOptions],
	) error

	// ListIndexes returns all the indexes that exist on the collection.
	// Each result's Execution reports the index's build status (building, ready, or failed),
	// which is how the outcome of an asynchronous NewIndex is observed.
	ListIndexes(
		ctx context.Context,
		opts ...options.Enumerable[options.ListCollectionIndexesOptions],
	) ([]ListIndexesResult, error)

	// NewEncryptedIndex makes a new encrypted index on the collection.
	NewEncryptedIndex(
		ctx context.Context,
		desc EncryptedIndexDescription,
		opts ...options.Enumerable[options.NewEncryptedIndexOptions],
	) (EncryptedIndexDescription, error)

	// DeleteEncryptedIndex deletes an encrypted index from the collection.
	DeleteEncryptedIndex(
		ctx context.Context,
		fieldName string,
		opts ...options.Enumerable[options.DeleteEncryptedIndexOptions],
	) error

	// ListEncryptedIndexes returns all the encrypted indexes that exist on the collection.
	ListEncryptedIndexes(
		ctx context.Context,
		opts ...options.Enumerable[options.ListCollectionEncryptedIndexesOptions],
	) ([]EncryptedIndexDescription, error)

	// Truncate this collection, permanently deleting all document state on this node.
	//
	// Changes made by this call will not impact other nodes, and cannot be synced to them over the P2P
	// system.
	//
	// This call will lock the collection, and no other read or write document operations on this collection
	// will progress whilst this is executing.
	//
	// Truncate is not transactional. When called with a transaction context, the transaction is used only to
	// scope the collection lock - the truncate effects are applied immediately.
	//
	// If an error is returned, the truncate may have been partially applied - because of this it is strongly
	// suggested to call `Truncate` until it succeeds.
	//
	// User-managed transactions are not supported by this function if the backing store does not support
	// concurrent open transactions, such as LevelDB.
	Truncate(ctx context.Context, opts ...options.Enumerable[options.TruncateCollectionOptions]) error

	// PurgeByDocIDs permanently removes the selected documents from this node.
	//
	// Unlike DeleteDocument, this is a local hard delete and does not replicate over P2P.
	// When pruneHistory is true, unshared history blocks are removed as well.
	//
	// Without a caller transaction, documents are committed in bounded chunks and an error
	// may leave earlier chunks applied. A peer may reintroduce purged data unless a replicated
	// soft delete has already propagated.
	PurgeByDocIDs(
		ctx context.Context,
		docIDs []DocID,
		pruneHistory bool,
		opts ...options.Enumerable[options.TruncateCollectionOptions],
	) error
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
