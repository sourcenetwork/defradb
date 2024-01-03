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
	indexedFields     []client.FieldDescription
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
	indexDesc client.IndexDescription,
	indexFilter *mapper.Filter,
) *IndexFetcher {
	return &IndexFetcher{
		docFetcher:  docFetcher,
		indexDesc:   indexDesc,
		indexFilter: indexFilter,
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

	f.indexDataStoreKey.IndexID = f.indexDesc.ID
	f.indexDataStoreKey.CollectionID = f.col.ID()

	for _, indexedField := range f.indexDesc.Fields {
		for _, field := range f.col.Schema().Fields {
			if field.Name == indexedField.Name {
				f.indexedFields = append(f.indexedFields, field)
				break
			}
		}
	}

	f.docFields = make([]client.FieldDescription, 0, len(fields)-len(f.indexedFields))
outer:
	for i := range fields {
		for j := range f.indexedFields {
			if fields[i].Name == f.indexedFields[j].Name {
				continue outer
			}
		}
		f.docFields = append(f.docFields, fields[i])
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

		for i, indexedField := range f.indexedFields {
			property := &encProperty{
				Desc: indexedField,
				Raw:  res.key.FieldValues[i],
			}

			f.doc.properties[indexedField] = property
		}

		if f.indexDesc.Unique {
			f.doc.id = res.value
		} else {
			f.doc.id = res.key.FieldValues[len(res.key.FieldValues)-1]
		}

		f.execInfo.FieldsFetched++

		if f.docFetcher != nil && len(f.docFields) > 0 {
			targetKey := base.MakeDataStoreKeyWithCollectionAndDocID(f.col.Description(), string(f.doc.id))
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
