// Copyright 2026 Democratized Data Foundation
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

	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/fetcher"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/identity"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/utils"
)

func (c *collection) Get(
	ctx context.Context,
	docID client.DocID,
	opts ...options.Enumerable[options.CollectionGetOptions],
) (*client.Document, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	opt := utils.NewOptions(opts...)
	showDeleted := opt.ShowDeleted

	if err := c.db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodeDocumentReadPerm); err != nil {
		return nil, err
	}

	ctx = identity.WithContext(ctx, opt.Identity)

	// create txn
	ctx, txn, err := ensureContextTxn(ctx, c.db, true)
	if err != nil {
		return nil, err
	}
	defer txn.Discard()
	primaryKey, err := c.getPrimaryKeyFromDocID(ctx, docID)
	if err != nil {
		return nil, err
	}

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

	return doc, txn.Commit()
}

func (c *collection) get(
	ctx context.Context,
	primaryKey keys.PrimaryDataStoreKey,
	fields []client.CollectionFieldDescription,
	showDeleted bool,
) (*client.Document, error) {
	txn := datastore.CtxMustGetTxn(ctx)
	// create a new document fetcher
	df := c.newFetcher(ctx)
	// initialize it with the primary index
	err := df.Init(
		ctx,
		identity.FromContext(ctx),
		txn,
		c.db.nodeACP,
		c.db.documentACP,
		immutable.Option[client.IndexDescription]{},
		c,
		fields,
		nil,
		nil,
		nil,
		showDeleted,
	)
	if err != nil {
		_ = df.Close()
		return nil, err
	}

	shortID, err := id.GetShortCollectionID(ctx, c.Version().CollectionID)
	if err != nil {
		return nil, err
	}

	// construct target DS key from DocID.
	targetKey := keys.DataStoreKey{
		CollectionShortID: shortID,
		DocID:             primaryKey.DocID,
	}
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

	doc, err := fetcher.Decode(ctx, encodedDoc, c.Version())
	if err != nil {
		return nil, err
	}

	return doc, nil
}
