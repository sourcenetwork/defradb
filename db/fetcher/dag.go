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
	"sort"
	"strings"

	"github.com/ipfs/go-cid"
	dsq "github.com/ipfs/go-datastore/query"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
)

// HeadFetcher is a utility to incrementally fetch all the MerkleCRDT
// heads of a given doc/field
type HeadFetcher struct {
	spans   core.Spans
	fieldId client.Option[string]

	kv     *core.HeadKeyValue
	kvIter dsq.Results
	kvEnd  bool
}

func (hf *HeadFetcher) Start(
	ctx context.Context,
	txn datastore.Txn,
	spans core.Spans,
	fieldId client.Option[string],
) error {
	numspans := len(spans.Value)
	if numspans == 0 {
		return errors.New("headFetcher must have at least one span")
	} else if numspans > 1 {
		// if we have multiple spans, we need to sort them by their start position
		// so we can do a single iterative sweep
		sort.Slice(spans.Value, func(i, j int) bool {
			// compare by strings if i < j.
			// apply the '!= df.reverse' to reverse the sort
			// if we need to
			return (strings.Compare(spans.Value[i].Start().ToString(), spans.Value[j].Start().ToString()) < 0)
		})
	}
	hf.spans = spans
	hf.fieldId = fieldId

	q := dsq.Query{
		Prefix: hf.spans.Value[0].Start().ToString(),
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

	return hf.nextKey()
}

func (hf *HeadFetcher) nextKey() error {
	res, available := hf.kvIter.NextSync()
	if res.Error != nil {
		hf.kvEnd = true
		hf.kv = nil
		return res.Error
	}
	if !available {
		hf.kvEnd = true
		hf.kv = nil
		return nil
	}

	headStoreKey, err := core.NewHeadStoreKey(res.Key)
	if err != nil {
		hf.kvEnd = true
		hf.kv = nil
		return err
	}
	hf.kv = &core.HeadKeyValue{
		Key:   headStoreKey,
		Value: res.Value,
	}
	hf.kvEnd = false

	return nil
}

func (hf *HeadFetcher) FetchNext() (*cid.Cid, error) {
	if hf.kvEnd {
		return nil, nil
	}

	if hf.kv == nil {
		return nil, errors.New("failed to get head, fetcher hasn't been initialized or started")
	}

	if hf.fieldId.HasValue() && hf.fieldId.Value() != hf.kv.Key.FieldId {
		// FieldIds do not match, continue to next row
		err := hf.nextKey()
		if err != nil {
			return nil, err
		}
		return hf.FetchNext()
	}

	cid := hf.kv.Key.Cid

	err := hf.nextKey()
	if err != nil {
		return nil, err
	}
	return &cid, nil
}

func (hf *HeadFetcher) Close() error {
	if hf.kvIter == nil {
		return nil
	}

	return hf.kvIter.Close()
}
