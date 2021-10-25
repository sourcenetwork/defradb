// Copyright 2020 Source Inc.
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
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/db/fetcher"
	"github.com/sourcenetwork/defradb/document"
	"github.com/sourcenetwork/defradb/document/key"
)

func (c *Collection) Get(key key.DocKey) (*document.Document, error) {
	//create txn
	txn, err := c.getTxn(true)
	if err != nil {
		return nil, err
	}
	defer c.discardImplicitTxn(txn)

	found, err := c.exists(txn, key)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, ErrDocumentNotFound
	}

	doc, err := c.get(txn, key)
	if err != nil {
		return nil, err
	}
	return doc, c.commitImplicitTxn(txn)
}

func (c *Collection) get(txn *Txn, key key.DocKey) (*document.Document, error) {
	// create a new document fetcher
	df := new(fetcher.DocumentFetcher)
	desc := &c.desc
	index := &c.desc.Indexes[0]
	// initialize it with the priamry index
	err := df.Init(&c.desc, &c.desc.Indexes[0], nil, false)
	if err != nil {
		return nil, err
	}

	// construct target key for DocKey
	targetKey := base.MakeIndexKey(desc, index, key.Key)
	// run the doc fetcher
	err = df.Start(txn, core.Spans{core.NewSpan(targetKey, targetKey.PrefixEnd())})
	if err != nil {
		return nil, err
	}

	// return first matched decoded doc
	return df.FetchNextDecoded()
}
