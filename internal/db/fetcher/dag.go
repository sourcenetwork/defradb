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

func (hf *HeadFetcher) Start(
	ctx context.Context,
	txn datastore.Txn,
	prefix keys.HeadStoreKey,
	fieldId immutable.Option[string],
) error {
	hf.fieldId = fieldId

	q := dsq.Query{
		Prefix: prefix.ToString(),
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

	headStoreKey, err := keys.NewHeadStoreKey(res.Key)
	if err != nil {
		return nil, err
	}

	if hf.fieldId.HasValue() && hf.fieldId.Value() != headStoreKey.FieldID {
		// FieldIds do not match, continue to next row
		return hf.FetchNext()
	}

	return &headStoreKey.Cid, nil
}

func (hf *HeadFetcher) Close() error {
	if hf.kvIter == nil {
		return nil
	}

	return hf.kvIter.Close()
}
