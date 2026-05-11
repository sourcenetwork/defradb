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

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
)

// collectionRetriever is a helper struct that retrieves a collection from a document ID.
type collectionRetriever struct {
	db *DB
}

// NewCollectionRetriever creates a new CollectionRetriever.
func NewCollectionRetriever(db *DB) collectionRetriever {
	return collectionRetriever{
		db: db,
	}
}

// RetrieveCollectionFromDocID retrieves a collection from a document ID.
// The supplied identity is forwarded to the underlying collection lookup, so
// NAC sees the caller's identity rather than the node's. Pass `immutable.None`
// for anonymous lookups; NAC will gate the call accordingly.
func (r collectionRetriever) RetrieveCollectionFromDocID(
	ctx context.Context,
	docID string,
	ident immutable.Option[identity.Identity],
) (client.Collection, error) {
	ctx, txn, err := ensureContextTxn(ctx, r.db, false)
	if err != nil {
		return nil, err
	}

	defer txn.Discard()

	headIterator, err := NewHeadBlocksIteratorFromTxn(ctx, docID)
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

	opt := options.GetCollections().SetVersionID(headIterator.CurrentBlock().Delta.GetCollectionVersionID())
	if ident.HasValue() {
		opt = opt.SetIdentity(ident.Value())
	}

	cols, err := txn.GetCollections(ctx, opt)
	if err != nil {
		return nil, err
	}

	if len(cols) == 0 {
		return nil, client.NewErrCollectionNotFoundForCollectionVersion(
			headIterator.CurrentBlock().Delta.GetCollectionVersionID(),
		)
	}

	return cols[0], nil
}
