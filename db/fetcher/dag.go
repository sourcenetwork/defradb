// Copyright 2020 Source Inc.
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
	"errors"
	"sort"
	"strings"

	"github.com/sourcenetwork/defradb/core"

	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	dsq "github.com/ipfs/go-datastore/query"
	dshelp "github.com/ipfs/go-ipfs-ds-help"
)

// @todo: Generalize all Fetchers into an shared Fetcher utility

type BlockFetcher struct {
}

// HeadFetcher is a utility to incrementally fetch all the MerkleCRDT
// heads of a given doc/field
type HeadFetcher struct {
	// key core.Key

	/* Commented because this code is not used yet according to the linter.
	txn   core.Txn
	*/
	spans core.Spans
	// curSpanIndex int

	cid *cid.Cid

	kv     *core.KeyValue
	kvIter dsq.Results
	kvEnd  bool
}

func (hf *HeadFetcher) Start(ctx context.Context, txn core.Txn, spans core.Spans) error {
	numspans := len(spans)
	if numspans == 0 {
		return errors.New("HeadFetcher must have at least one span")
	} else if numspans > 1 {
		// if we have multiple spans, we need to sort them by their start position
		// so we can do a single iterative sweep
		sort.Slice(spans, func(i, j int) bool {
			// compare by strings if i < j.
			// apply the '!= df.reverse' to reverse the sort
			// if we need to
			return (strings.Compare(spans[i].Start().String(), spans[j].Start().String()) < 0)
		})
	}
	hf.spans = spans

	q := dsq.Query{
		Prefix: hf.spans[0].Start().String(),
		Orders: []dsq.Order{dsq.OrderByKey{}},
	}

	var err error
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

func (hf *HeadFetcher) nextKV() (iterDone bool, kv *core.KeyValue, err error) {
	res, available := hf.kvIter.NextSync()
	if !available {
		return true, nil, nil
	}
	if res.Error != nil {
		return true, nil, err
	}

	kv = &core.KeyValue{
		Key:   core.NewKey(res.Key),
		Value: res.Value,
	}
	return false, kv, nil
}

func (hf *HeadFetcher) processKV(kv *core.KeyValue) error {
	// convert Value from KV value to cid.Cid
	headKey := ds.NewKey(strings.TrimPrefix(kv.Key.String(), hf.spans[0].Start().String()))

	hash, err := dshelp.DsKeyToMultihash(headKey)
	if err != nil {
		return err
	}
	headCid := cid.NewCidV1(cid.Raw, hash)
	hf.cid = &headCid
	return nil
}

func (hf *HeadFetcher) FetchNext() (*cid.Cid, error) {
	if hf.kvEnd {
		return nil, nil
	}

	if hf.kv == nil {
		return nil, errors.New("Failed to get head, fetcher hasn't been initialized or started")
	}

	if err := hf.processKV(hf.kv); err != nil {
		return nil, err
	}

	_, err := hf.nextKey()
	if err != nil {
		return nil, err
	}
	return hf.cid, nil
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
			return nil, 0, fmt.Errorf("Failed to get next query result : %w", err)
		}
		// fmt.Println(r.Key, hh.namespace.String())
		headKey := ds.NewKey(strings.TrimPrefix(r.Key, hh.namespace.String()))
		headCid, err := dshelp.DsKeyToCid(headKey)
		if err != nil {
			return nil, 0, fmt.Errorf("Failed to get CID from key : %w", err)
		}
		height, n := binary.Uvarint(r.Value)
		if n <= 0 {
			return nil, 0, errors.New("error decocding height")
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
