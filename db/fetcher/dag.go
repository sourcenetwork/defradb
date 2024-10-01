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
	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
)

// HeadFetcher is a utility to incrementally fetch all the MerkleCRDT heads of a given doc/field.
type HeadFetcher struct {
	spans   core.Spans
	fieldId immutable.Option[string]

	kvIter corekv.Iterator
}

func (hf *HeadFetcher) Start(
	ctx context.Context,
	txn datastore.Txn,
	spans core.Spans,
	fieldId immutable.Option[string],
) error {
	if len(spans.Value) == 0 {
		spans = core.NewSpans(
			core.NewSpan(
				core.DataStoreKey{},
				core.DataStoreKey{}.PrefixEnd(),
			),
		)
	}

	if len(spans.Value) > 1 {
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

	if hf.kvIter != nil {
		if err := hf.kvIter.Close(ctx); err != nil {
			return err
		}
	}
	hf.kvIter = txn.Headstore().Iterator(ctx, corekv.IterOptions{
		Prefix: hf.spans.Value[0].Start().Bytes(),
	})

	return nil
}

func (hf *HeadFetcher) FetchNext() (*cid.Cid, error) {
	hf.kvIter.Next()
	available := hf.kvIter.Valid()
	if !available {
		return nil, nil
	}

	headStoreKey, err := core.NewHeadStoreKey(string(hf.kvIter.Key()))
	if err != nil {
		return nil, err
	}

	if hf.fieldId.HasValue() && hf.fieldId.Value() != headStoreKey.FieldId {
		// FieldIds do not match, continue to next row
		return hf.FetchNext()
	}

	return &headStoreKey.Cid, nil
}

func (hf *HeadFetcher) Close() error {
	hf.kvIter.Close(context.TODO())
	return nil // clean up
}
