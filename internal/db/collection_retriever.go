// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package db

import (
	"context"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/db/description"
)

// collectionRetriever is a helper struct that retrieves a collection from a document ID.
type collectionRetriever struct {
	db client.DB
}

// NewCollectionRetriever creates a new CollectionRetriever.
func NewCollectionRetriever(db client.DB) *collectionRetriever {
	return &collectionRetriever{
		db: db,
	}
}

// RetrieveCollectionFromDocID retrieves a collection from a document ID.
func (r *collectionRetriever) RetrieveCollectionFromDocID(
	ctx context.Context,
	docID string,
) (client.Collection, error) {
	ctx, txn, err := ensureContextTxn(ctx, r.db, false)
	if err != nil {
		return nil, err
	}
	defer txn.Discard(ctx)

	headIterator, err := NewHeadBlocksIteratorFromTxn(ctx, txn, docID)
	if err != nil {
		return nil, err
	}

	hasValue, err := headIterator.Next()
	if err != nil {
		return nil, err
	}

	if !hasValue {
		return nil, NewErrDocIDNotFound(docID)
	}

	schema, err := description.GetSchemaVersion(ctx, txn, headIterator.CurrentBlock().Delta.GetSchemaVersionID())
	if err != nil {
		return nil, err
	}

	cols, err := r.db.GetCollections(
		ctx,
		client.CollectionFetchOptions{
			SchemaRoot: immutable.Some(schema.Root),
		},
	)

	if err != nil {
		return nil, err
	}

	if len(cols) == 0 {
		return nil, NewErrCollectionWithSchemaRootNotFound(schema.Root)
	}

	return cols[0], nil
}
