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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp"
	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/planner/mapper"
)

// IndexFetcher is a fetcher that fetches documents by index.
// It fetches only the indexed field and the rest of the fields are fetched by the internal fetcher.
type IndexFetcher struct {
	docFetcher    Fetcher
	col           client.Collection
	txn           datastore.Txn
	indexFilter   *mapper.Filter
	docFilter     *mapper.Filter
	doc           *encodedDocument
	mapping       *core.DocumentMapping
	indexedFields []client.FieldDefinition
	docFields     []client.FieldDefinition
	indexDesc     client.IndexDescription
	indexIter     indexIterator
	execInfo      ExecInfo
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
	id immutable.Option[identity.Identity],
	txn datastore.Txn,
	acp immutable.Option[acp.ACP],
	col client.Collection,
	fields []client.FieldDefinition,
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

	for _, indexedField := range f.indexDesc.Fields {
		field, ok := f.col.Definition().GetFieldByName(indexedField.Name)
		if ok {
			f.indexedFields = append(f.indexedFields, field)
		}
	}

	f.docFields = make([]client.FieldDefinition, 0, len(fields))
outer:
	for i := range fields {
		for j := range f.indexedFields {
			if fields[i].Name == f.indexedFields[j].Name {
				continue outer
			}
		}
		f.docFields = append(f.docFields, fields[i])
	}

	iter, err := f.createIndexIterator()
	if err != nil {
		return err
	}
	f.indexIter = iter

	if f.docFetcher != nil && len(f.docFields) > 0 {
		err = f.docFetcher.Init(
			ctx,
			id,
			f.txn,
			acp,
			f.col,
			f.docFields,
			f.docFilter,
			f.mapping,
			false,
			false,
		)
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

		hasNilField := false
		for i, indexedField := range f.indexedFields {
			property := &encProperty{Desc: indexedField}

			field := res.key.Fields[i]
			if field.Value.IsNil() {
				hasNilField = true
			}

			// We need to convert it to cbor bytes as this is what it will be encoded from on value retrieval.
			// In the future we have to either get rid of CBOR or properly handle different encoding
			// for properties in a single document.
			fieldBytes, err := client.NewFieldValue(client.NONE_CRDT, field.Value).Bytes()
			if err != nil {
				return nil, ExecInfo{}, err
			}
			property.Raw = fieldBytes

			f.doc.properties[indexedField] = property
		}

		if f.indexDesc.Unique && !hasNilField {
			f.doc.id = res.value
		} else {
			lastVal := res.key.Fields[len(res.key.Fields)-1].Value
			if str, ok := lastVal.String(); ok {
				f.doc.id = []byte(str)
			} else if bytes, ok := lastVal.Bytes(); ok {
				f.doc.id = bytes
			} else {
				return nil, ExecInfo{}, err
			}
		}

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
