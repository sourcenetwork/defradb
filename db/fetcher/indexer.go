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
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/planner/mapper"
)

// IndexFetcher is a fetcher that fetches documents by index.
// It fetches only the indexed field and the rest of the fields are fetched by the internal fetcher.
type IndexFetcher struct {
	docFetcher        Fetcher
	col               client.Collection
	txn               datastore.Txn
	indexFilter       *mapper.Filter
	docFilter         *mapper.Filter
	doc               *encodedDocument
	mapping           *core.DocumentMapping
	indexedField      client.FieldDescription
	docFields         []client.FieldDescription
	indexDesc         client.IndexDescription
	indexIter         indexIterator
	indexDataStoreKey core.IndexDataStoreKey
	execInfo          ExecInfo
}

var _ Fetcher = (*IndexFetcher)(nil)

// NewIndexFetcher creates a new IndexFetcher.
func NewIndexFetcher(
	docFetcher Fetcher,
	indexedFieldDesc client.FieldDescription,
	indexFilter *mapper.Filter,
) *IndexFetcher {
	return &IndexFetcher{
		docFetcher:   docFetcher,
		indexedField: indexedFieldDesc,
		indexFilter:  indexFilter,
	}
}

func (f *IndexFetcher) Init(
	ctx context.Context,
	txn datastore.Txn,
	col client.Collection,
	fields []client.FieldDescription,
	filter *mapper.Filter,
	docMapper *core.DocumentMapping,
	reverse bool,
	showDeleted bool,
) error {
	f.col = col
	f.docFilter = filter
	f.doc = &encodedDocument{}
	f.mapping = docMapper
	f.txn = txn

	for _, index := range col.Description().Indexes {
		if index.Fields[0].Name == f.indexedField.Name {
			f.indexDesc = index
			f.indexDataStoreKey.IndexID = index.ID
			break
		}
	}

	f.indexDataStoreKey.CollectionID = f.col.ID()

	for i := range fields {
		if fields[i].Name == f.indexedField.Name {
			f.docFields = append(fields[:i], fields[i+1:]...)
			break
		}
	}

	iter, err := createIndexIterator(f.indexDataStoreKey, f.indexFilter, &f.execInfo, f.indexDesc.Unique)
	if err != nil {
		return err
	}
	f.indexIter = iter

	if f.docFetcher != nil && len(f.docFields) > 0 {
		err = f.docFetcher.Init(ctx, f.txn, f.col, f.docFields, f.docFilter, f.mapping, false, false)
	}

	return err
}

func (f *IndexFetcher) Start(ctx context.Context, spans core.Spans) error {
	err := f.indexIter.Init(ctx, f.txn.Datastore())
	if err != nil {
		return err
	}
	return nil
}

func (f *IndexFetcher) FetchNext(ctx context.Context) (EncodedDocument, ExecInfo, error) {
	totalExecInfo := f.execInfo
	defer func() { f.execInfo.Add(totalExecInfo) }()
	f.execInfo.Reset()
	for {
		f.doc.Reset()

		res, err := f.indexIter.Next()
		if err != nil {
			return nil, ExecInfo{}, err
		}

		if !res.foundKey {
			return nil, f.execInfo, nil
		}

		property := &encProperty{
			Desc: f.indexedField,
			Raw:  res.key.FieldValues[0],
		}

		if f.indexDesc.Unique {
			f.doc.id = res.value
		} else {
			f.doc.id = res.key.FieldValues[1]
		}
		f.doc.properties[f.indexedField] = property
		f.execInfo.FieldsFetched++

		if f.docFetcher != nil && len(f.docFields) > 0 {
			targetKey := base.MakeDSKeyWithCollectionAndDocID(f.col.Description(), string(f.doc.id))
			spans := core.NewSpans(core.NewSpan(targetKey, targetKey.PrefixEnd()))
			err := f.docFetcher.Start(ctx, spans)
			if err != nil {
				return nil, ExecInfo{}, err
			}
			encDoc, execInfo, err := f.docFetcher.FetchNext(ctx)
			if err != nil {
				return nil, ExecInfo{}, err
			}
			err = f.docFetcher.Close()
			if err != nil {
				return nil, ExecInfo{}, err
			}
			f.execInfo.Add(execInfo)
			if encDoc == nil {
				continue
			}
			f.doc.MergeProperties(encDoc)
		} else {
			f.execInfo.DocsFetched++
		}
		return f.doc, f.execInfo, nil
	}
}

func (f *IndexFetcher) Close() error {
	if f.indexIter != nil {
		return f.indexIter.Close()
	}
	return nil
}
