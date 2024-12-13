// Copyright 2022 Democratized Data Foundation
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

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/db/base"
	"github.com/sourcenetwork/defradb/internal/db/fetcher"
	"github.com/sourcenetwork/defradb/internal/keys"
)

func (c *collection) Get(
	ctx context.Context,
	docID client.DocID,
	showDeleted bool,
) (*client.Document, error) {
	// create txn
	ctx, txn, err := ensureContextTxn(ctx, c.db, true)
	if err != nil {
		return nil, err
	}
	defer txn.Discard(ctx)
	primaryKey := c.getPrimaryKeyFromDocID(docID)

	found, isDeleted, err := c.exists(ctx, primaryKey)
	if err != nil {
		return nil, err
	}
	if !found || (isDeleted && !showDeleted) {
		return nil, client.ErrDocumentNotFoundOrNotAuthorized
	}

	doc, err := c.get(ctx, primaryKey, nil, showDeleted)
	if err != nil {
		return nil, err
	}

	if doc == nil {
		return nil, client.ErrDocumentNotFoundOrNotAuthorized
	}

	return doc, txn.Commit(ctx)
}

func (c *collection) get(
	ctx context.Context,
	primaryKey keys.PrimaryDataStoreKey,
	fields []client.FieldDefinition,
	showDeleted bool,
) (*client.Document, error) {
	txn := mustGetContextTxn(ctx)
	// create a new document fetcher
	df := c.newFetcher()
	// initialize it with the primary index
	err := df.Init(ctx, identity.FromContext(ctx), txn, c.db.acp, c, fields, nil, nil, false, showDeleted)
	if err != nil {
		_ = df.Close()
		return nil, err
	}

	// construct target DS key from DocID.
	targetKey := base.MakeDataStoreKeyWithCollectionAndDocID(c.Description(), primaryKey.DocID)
	// run the doc fetcher
	err = df.Start(ctx, targetKey)
	if err != nil {
		_ = df.Close()
		return nil, err
	}

	// return first matched decoded doc
	encodedDoc, _, err := df.FetchNext(ctx)
	if err != nil {
		_ = df.Close()
		return nil, err
	}

	err = df.Close()
	if err != nil {
		return nil, err
	}

	if encodedDoc == nil {
		return nil, nil
	}

	doc, err := fetcher.Decode(encodedDoc, c.Definition())
	if err != nil {
		return nil, err
	}

	return doc, nil
}
