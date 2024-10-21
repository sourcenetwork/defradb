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
	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/db/base"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

// IndexFetcher is a fetcher that fetches documents by index.
// It fetches only the indexed field and the rest of the fields are fetched by the internal fetcher.
type IndexFetcher struct {
	docFetcher    Fetcher
	col           client.Collection
	txn           datastore.Txn
	indexFilter   *mapper.Filter
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
	identity immutable.Option[acpIdentity.Identity],
	txn datastore.Txn,
	acp immutable.Option[acp.ACP],
	col client.Collection,
	fields []client.FieldDefinition,
	filter *mapper.Filter,
	docMapper *core.DocumentMapping,
	reverse bool,
	showDeleted bool,
) error {
	f.resetState()

	f.col = col
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
			// If the field is array, we want to keep it also for the document fetcher
			// because the index only contains one array elements, not the whole array.
			// The doc fetcher will fetch the whole array for us.
			if fields[i].Name == f.indexedFields[j].Name && !fields[i].Kind.IsArray() {
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

	// if it turns out that we can't use the index, we need to fall back to the document fetcher
	if f.indexIter == nil {
		f.docFields = fields
	}

	if len(f.docFields) > 0 {
		err = f.docFetcher.Init(
			ctx,
			identity,
			f.txn,
			acp,
			f.col,
			f.docFields,
			filter,
			f.mapping,
			false,
			false,
		)
	}

	return err
}

func (f *IndexFetcher) Start(ctx context.Context, spans core.Spans) error {
	if f.indexIter == nil {
		return f.docFetcher.Start(ctx, spans)
	}
	return f.indexIter.Init(ctx, f.txn.Datastore())
}

func (f *IndexFetcher) FetchNext(ctx context.Context) (EncodedDocument, ExecInfo, error) {
	if f.indexIter == nil {
		return f.docFetcher.FetchNext(ctx)
	}
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

			// Index will fetch only 1 array element. So we skip it here and let doc fetcher
			// fetch the whole array.
			if indexedField.Kind.IsArray() {
				continue
			}

			// We need to convert it to cbor bytes as this is what it will be encoded from on value retrieval.
			// In the future we have to either get rid of CBOR or properly handle different encoding
			// for properties in a single document.
			fieldBytes, err := client.NewFieldValue(client.NONE_CRDT, field.Value).Bytes()
			if err != nil {
				return nil, ExecInfo{}, err
			}
			property.Raw = fieldBytes

			f.doc.properties[indexedField.Key()] = property
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

		if len(f.docFields) > 0 {
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
	if f.indexIter == nil {
		return f.docFetcher.Close()
	}
	return f.indexIter.Close()
}

// resetState resets the mutable state of this IndexFetcher, returning the state to how it
// was immediately after construction.
func (f *IndexFetcher) resetState() {
	// WARNING: Do not reset properties set in the constructor!

	f.col = nil
	f.txn = nil
	f.doc = nil
	f.mapping = nil
	f.indexedFields = nil
	f.docFields = nil
	f.indexIter = nil
	f.execInfo.Reset()
}
