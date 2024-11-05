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
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// HeadFetcher is a utility to incrementally fetch all the MerkleCRDT heads of a given doc/field.
type HeadFetcher struct {
	spans   []core.Span
	fieldId immutable.Option[string]

	kvIter dsq.Results
}

func (hf *HeadFetcher) Start(
	ctx context.Context,
	txn datastore.Txn,
	spans []core.Span,
	fieldId immutable.Option[string],
) error {
	if len(spans) == 0 {
		spans = []core.Span{
			core.NewSpan(
				keys.DataStoreKey{},
				keys.DataStoreKey{}.PrefixEnd(),
			),
		}
	}

	if len(spans) > 1 {
		// if we have multiple spans, we need to sort them by their start position
		// so we can do a single iterative sweep
		sort.Slice(spans, func(i, j int) bool {
			// compare by strings if i < j.
			// apply the '!= df.reverse' to reverse the sort
			// if we need to
			return (strings.Compare(spans[i].Start.ToString(), spans[j].Start.ToString()) < 0)
		})
	}
	hf.spans = spans
	hf.fieldId = fieldId

	q := dsq.Query{
		Prefix: hf.spans[0].Start.ToString(),
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
