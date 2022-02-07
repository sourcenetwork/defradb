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

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/db/fetcher"
	"github.com/sourcenetwork/defradb/document"
	"github.com/sourcenetwork/defradb/document/key"
)

func (c *Collection) Get(ctx context.Context, key key.DocKey) (*document.Document, error) {
	//create txn
	txn, err := c.getTxn(ctx, true)
	if err != nil {
		return nil, err
	}
	defer c.discardImplicitTxn(ctx, txn)

	found, err := c.exists(ctx, txn, key)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, ErrDocumentNotFound
	}

	doc, err := c.get(ctx, txn, key)
	if err != nil {
		return nil, err
	}
	return doc, c.commitImplicitTxn(ctx, txn)
}

func (c *Collection) get(ctx context.Context, txn *Txn, key key.DocKey) (*document.Document, error) {
	// create a new document fetcher
	df := new(fetcher.DocumentFetcher)
	desc := &c.desc
	index := &c.desc.Indexes[0]
	// initialize it with the primary index
	err := df.Init(&c.desc, &c.desc.Indexes[0], nil, false)
	if err != nil {
		return nil, err
	}

	// construct target key for DocKey
	targetKey := base.MakeIndexKey(desc, index, core.Key{Key: key.Key})
	// run the doc fetcher
	err = df.Start(ctx, txn, core.Spans{core.NewSpan(targetKey, targetKey.PrefixEnd())})
	if err != nil {
		return nil, err
	}

	// return first matched decoded doc
	doc, err := df.FetchNextDecoded(ctx)
	if err != nil {
		return nil, err
	}

	err = df.Close()
	if err != nil {
		return nil, err
	}

	return doc, nil
}
