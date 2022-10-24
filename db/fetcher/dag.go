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

type BlockFetcher struct {
}

// HeadFetcher is a utility to incrementally fetch all the MerkleCRDT
// heads of a given doc/field
type HeadFetcher struct {
	spans   core.Spans
	cid     *cid.Cid
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

	_, err = hf.nextKey()
	return err
}

func (hf *HeadFetcher) nextKey() (bool, error) {
	var err error
	var done bool
	done, hf.kv, err = hf.nextKV()
	if err != nil {
		return false, err
	}

	hf.kvEnd = done
	if hf.kvEnd {
		return true, nil
	}
	return false, nil
}

func (hf *HeadFetcher) nextKV() (iterDone bool, kv *core.HeadKeyValue, err error) {
	res, available := hf.kvIter.NextSync()
	if !available {
		return true, nil, nil
	}
	if res.Error != nil {
		return true, nil, err
	}

	headStoreKey, err := core.NewHeadStoreKey(res.Key)
	if err != nil {
		return true, nil, err
	}
	kv = &core.HeadKeyValue{
		Key:   headStoreKey,
		Value: res.Value,
	}
	return false, kv, nil
}

func (hf *HeadFetcher) processKV(kv *core.HeadKeyValue) {
	hf.cid = &kv.Key.Cid
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
		_, err := hf.nextKey()
		if err != nil {
			return nil, err
		}
		return hf.FetchNext()
	}

	hf.processKV(hf.kv)

	_, err := hf.nextKey()
	if err != nil {
		return nil, err
	}
	return hf.cid, nil
}

func (hf *HeadFetcher) Close() error {
	if hf.kvIter == nil {
		return nil
	}

	return hf.kvIter.Close()
}
