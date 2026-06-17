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

	"github.com/ipfs/go-cid"
	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
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

func (r collectionRetriever) ResolvePublicDocID(ctx context.Context, docID string) (string, error) {
	ctx, txn, err := ensureContextTxn(ctx, r.db, false)
	if err != nil {
		return "", err
	}
	defer txn.Discard()

	publicDocID, err := resolvePublicDocIDFromStore(ctx, txn.Systemstore(), docID)
	if err != nil || publicDocID != docID {
		return publicDocID, err
	}

	blockCID, err := cid.Decode(docID)
	if err != nil {
		return docID, nil
	}
	cols, err := txn.GetCollections(ctx, options.GetCollections())
	if err != nil {
		return "", err
	}
	for _, col := range cols {
		shortID, err := id.GetShortCollectionID(ctx, col.CollectionID())
		if err != nil {
			return "", err
		}
		publicDocID, found, err := id.GetDocIDForBlockFromStore(ctx, txn.Systemstore(), shortID, blockCID)
		if err != nil || found {
			return publicDocID, err
		}
	}
	return docID, nil
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

	docID, err = resolvePublicDocIDFromStore(ctx, txn.Systemstore(), docID)
	if err != nil {
		return nil, err
	}

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

func resolvePublicDocIDFromStore(ctx context.Context, store corekv.Reader, docKey string) (string, error) {
	// Old encryption blocks may carry an encoded local document key.
	localDocID, err := keys.DecodeLocalDocID([]byte(docKey))
	if err == nil {
		publicDocID, found, err := id.GetDocIDFromStore(
			ctx,
			store,
			localDocID.CollectionShortID,
			localDocID.DocShortID,
		)
		if err != nil || found {
			return publicDocID, err
		}
	}

	publicDocID, found, err := id.GetNodeDocIDFromStore(ctx, store, docKey)
	if err != nil {
		return "", err
	}
	if found {
		return publicDocID, nil
	}
	return docKey, nil
}
