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

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
)

// @todo: Generalize all Fetchers into an shared Fetcher utility

type BlockFetcher struct {
}

// HeadFetcher is a utility to incrementally fetch all the MerkleCRDT
// heads of a given doc/field
type HeadFetcher struct {

	// Commented because this code is not used yet according to the linter.
	// txn   datastore.Txn

	// key core.Key
	// curSpanIndex int

	spans core.Spans
	cid   *cid.Cid

	kv     *core.HeadKeyValue
	kvIter dsq.Results
	kvEnd  bool
}

func (hf *HeadFetcher) Start(ctx context.Context, txn datastore.Txn, spans core.Spans) error {
	numspans := len(spans.Value)
	if numspans == 0 {
		return errors.New("HeadFetcher must have at least one span")
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
		return nil, errors.New("Failed to get head, fetcher hasn't been initialized or started")
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

/*
// List returns the list of current heads plus the max height.
// @todo Document Heads.List function
func (hh *heads) List() ([]cid.Cid, uint64, error) {
	q := query.Query{
		Prefix:   hh.namespace.String(),
		KeysOnly: false,
	}

	results, err := hh.store.Query(q)
	if err != nil {
		return nil, 0, err
	}
	defer results.Close()

	heads := make([]cid.Cid, 0)
	var maxHeight uint64
	for r := range results.Next() {
		if r.Error != nil {
			return nil, 0, errors.Wrap("Failed to get next query result ", err)
		}
		headKey := ds.NewKey(strings.TrimPrefix(r.Key, hh.namespace.String()))
		headCid, err := dshelp.DsKeyToCid(headKey)
		if err != nil {
			return nil, 0, errors.Wrap("Failed to get CID from key ", err)
		}
		height, n := binary.Uvarint(r.Value)
		if n <= 0 {
			return nil, 0, errors.New("error decoding height")
		}
		heads = append(heads, headCid)
		if height > maxHeight {
			maxHeight = height
		}
	}
	sort.Slice(heads, func(i, j int) bool {
		ci := heads[i].Bytes()
		cj := heads[j].Bytes()
		return bytes.Compare(ci, cj) < 0
	})

	return heads, maxHeight, nil
}
*/
