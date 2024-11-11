// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package fetcher

import (
	"context"

	"github.com/ipfs/go-cid"
	dsq "github.com/ipfs/go-datastore/query"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// HeadFetcher is a utility to incrementally fetch all the MerkleCRDT heads of a given doc/field.
type HeadFetcher struct {
	fieldId immutable.Option[string]

	kvIter dsq.Results
}

// Start starts/initializes the fetcher, performing all the work it can do outside
// of the main iteration loop/funcs.
//
// prefix - Optional. The headstore prefix to scan across.  If None, the entire
// headstore will be scanned - for example, in order to fetch document and collection
// heads.
func (hf *HeadFetcher) Start(
	ctx context.Context,
	txn datastore.Txn,
	prefix immutable.Option[keys.HeadstoreKey],
	fieldId immutable.Option[string],
) error {
	hf.fieldId = fieldId

	var prefixString string
	if prefix.HasValue() {
		prefixString = prefix.Value().ToString()
	}

	q := dsq.Query{
		Prefix: prefixString,
		Orders: []dsq.Order{dsq.OrderByKey{}},
	}

	var err error
	if hf.kvIter != nil {
		if err := hf.kvIter.Close(); err != nil {
			return err
		}
	}
	hf.kvIter, err = txn.Headstore().Query(ctx, q)
	if err != nil {
		return err
	}

	return nil
}

func (hf *HeadFetcher) FetchNext() (*cid.Cid, error) {
	res, available := hf.kvIter.NextSync()
	if res.Error != nil {
		return nil, res.Error
	}
	if !available {
		return nil, nil
	}

	headStoreKey, err := keys.NewHeadstoreKey(res.Key)
	if err != nil {
		return nil, err
	}

	if hf.fieldId.HasValue() {
		switch typedHeadStoreKey := headStoreKey.(type) {
		case keys.HeadstoreDocKey:
			if hf.fieldId.Value() != typedHeadStoreKey.FieldID {
				// FieldIds do not match, continue to next row
				return hf.FetchNext()
			}

			return &typedHeadStoreKey.Cid, nil

		case keys.HeadstoreColKey:
			if hf.fieldId.Value() == "" {
				return &typedHeadStoreKey.Cid, nil
			} else {
				return nil, nil
			}
		}
	}

	cid := headStoreKey.GetCid()
	return &cid, nil
}

func (hf *HeadFetcher) Close() error {
	if hf.kvIter == nil {
		return nil
	}

	return hf.kvIter.Close()
}
