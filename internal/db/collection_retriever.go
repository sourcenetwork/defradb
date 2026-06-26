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
	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/utils"
	"github.com/sourcenetwork/immutable"
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

// ResolveBlockDocIDs returns every DocID that owns the given block CID. A block can be
// co-owned by several documents (identical genesis field deltas produce one shared CID), so
// callers must handle more than one owner.
func (r collectionRetriever) ResolveBlockDocIDs(ctx context.Context, blockCID cid.Cid) ([]string, error) {
	ctx, txn, err := ensureContextTxn(ctx, r.db, false)
	if err != nil {
		return nil, err
	}
	defer txn.Discard()

	return id.GetDocIDsForBlockFromStore(ctx, txn.Systemstore(), blockCID)
}

// RetrieveCollectionFromDocID retrieves a collection from a document ID.
// The identity argument is kept for the KMS interface, but collection metadata
// resolution is internal and must not require the requester to have get-collection access.
func (r collectionRetriever) RetrieveCollectionFromDocID(
	ctx context.Context,
	docID string,
	_ immutable.Option[identity.Identity],
) (client.Collection, error) {
	ctx, txn, err := ensureContextTxn(ctx, r.db, false)
	if err != nil {
		return nil, err
	}

	defer txn.Discard()

	docID, err = resolveDocIDFromStore(ctx, txn.Systemstore(), docID)
	if err != nil {
		return nil, err
	}

	docRef, found, err := id.GetDocRefFromStore(ctx, txn.Systemstore(), docID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, NewErrDocIDNotFound(docID)
	}

	cols, err := r.db.getCollections(
		ctx,
		utils.NewOptions(options.GetCollections().SetGetInactive(true)),
		true,
	)
	if err != nil {
		return nil, err
	}

	for _, col := range cols {
		collectionShortID, err := id.GetCollectionShortID(ctx, col.CollectionID())
		if err != nil {
			return nil, err
		}
		if collectionShortID == docRef.CollectionShortID {
			return col, nil
		}
	}

	return nil, client.ErrCollectionNotFound
}

func resolveDocIDFromStore(ctx context.Context, store corekv.Reader, docKey string) (string, error) {
	// Old encryption blocks may carry an encoded DocRef.
	docRef, err := keys.DecodeDocRef([]byte(docKey))
	if err == nil {
		docID, found, err := id.GetDocIDFromStore(
			ctx,
			store,
			docRef.DocShortID,
		)
		if err != nil || found {
			return docID, err
		}
	}

	docRef, found, err := id.GetDocRefFromStore(ctx, store, docKey)
	if err != nil || !found {
		return docKey, err
	}
	docID, found, err := id.GetDocIDFromStore(ctx, store, docRef.DocShortID)
	if err != nil || !found {
		return docKey, err
	}

	return docID, nil
}
