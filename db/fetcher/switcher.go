// Copyright 2023 Democratized Data Foundation
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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/planner/mapper"
)

type FetcherSwitcher struct {
	inner Fetcher
}

var _ Fetcher = (*FetcherSwitcher)(nil)

func (f *FetcherSwitcher) Init(
	ctx context.Context,
	txn datastore.Txn,
	col *client.CollectionDescription,
	fields []client.FieldDescription,
	filter *mapper.Filter,
	docMapper *core.DocumentMapping,
	reverse bool,
	showDeleted bool,
) error {
	var index client.IndexDescription
	var filterCond any
	var indexedFieldDesc client.FieldDescription
	colIndexes := col.Indexes
	if filter != nil {
		for filterFieldName, cond := range filter.ExternalConditions {
			for i := range colIndexes {
				if filterFieldName == colIndexes[i].Fields[0].Name {
					index = colIndexes[i]
					filterCond = cond

					indexedFields := col.CollectIndexedFields()
					for j := range indexedFields {
						if indexedFields[j].Name == filterFieldName {
							indexedFieldDesc = indexedFields[j]
							break
						}
					}
				}
			}
		}
	}

	if index.ID != 0 {
		f.inner = NewIndexFetcher(new(DocumentFetcher), indexedFieldDesc, index, filterCond)
	} else {
		f.inner = new(DocumentFetcher)
	}

	return f.inner.Init(ctx, txn, col, fields, filter, docMapper, reverse, showDeleted)
}

func (f *FetcherSwitcher) Start(ctx context.Context, spans core.Spans) error {
	if f.inner == nil {
		return nil
	}
	return f.inner.Start(ctx, spans)
}

func (f *FetcherSwitcher) FetchNext(ctx context.Context) (EncodedDocument, ExecInfo, error) {
	if f.inner == nil {
		return nil, ExecInfo{}, nil
	}
	return f.inner.FetchNext(ctx)
}

func (f *FetcherSwitcher) Close() error {
	if f.inner == nil {
		return nil
	}
	return f.inner.Close()
}
