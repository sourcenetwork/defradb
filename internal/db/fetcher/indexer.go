// Copyright 2025 Democratized Data Foundation
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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/db/base"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/planner/filter"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

// indexFetcher is a fetcher that fetches documents by index.
// It fetches only the indexed field and the rest of the fields are fetched by the internal fetcher.
type indexFetcher struct {
	ctx           context.Context
	txn           datastore.Txn
	col           client.Collection
	indexFilter   *mapper.Filter
	mapping       *core.DocumentMapping
	indexedFields []client.FieldDefinition
	fieldsByID    map[uint32]client.FieldDefinition
	indexDesc     client.IndexDescription
	indexIter     indexIterator
	currentDocID  immutable.Option[string]
	execInfo      ExecInfo
}

// var _ Fetcher = (*IndexFetcher)(nil)
var _ fetcher = (*indexFetcher)(nil)

// newIndexFetcher creates a new IndexFetcher.
func newIndexFetcher(
	ctx context.Context,
	txn datastore.Txn,
	fieldsByID map[uint32]client.FieldDefinition,
	indexDesc client.IndexDescription,
	docFilter *mapper.Filter,
	col client.Collection,
	fields []client.FieldDefinition,
	docMapper *core.DocumentMapping,
) (*indexFetcher, error) {
	f := &indexFetcher{
		ctx:        ctx,
		txn:        txn,
		col:        col,
		mapping:    docMapper,
		indexDesc:  indexDesc,
		fieldsByID: fieldsByID,
	}

	fieldsToCopy := make([]mapper.Field, 0, len(indexDesc.Fields))
	for _, field := range indexDesc.Fields {
		typeIndex := docMapper.FirstIndexOfName(field.Name)
		indexField := mapper.Field{Index: typeIndex, Name: field.Name}
		fieldsToCopy = append(fieldsToCopy, indexField)
	}
	for i := range fieldsToCopy {
		f.indexFilter = filter.Merge(f.indexFilter, filter.CopyField(docFilter, fieldsToCopy[i]))
	}

	for _, indexedField := range f.indexDesc.Fields {
		field, ok := f.col.Definition().GetFieldByName(indexedField.Name)
		if ok {
			f.indexedFields = append(f.indexedFields, field)
		}
	}

	iter, err := f.createIndexIterator()
	if err != nil || iter == nil {
		return nil, err
	}

	f.indexIter = iter
	return f, iter.Init(ctx, txn.Datastore())
}

func (f *indexFetcher) NextDoc() (immutable.Option[string], error) {
	totalExecInfo := f.execInfo
	defer func() { f.execInfo.Add(totalExecInfo) }()
	f.execInfo.Reset()

	f.currentDocID = immutable.None[string]()

	//for {
	res, err := f.indexIter.Next()
	if err != nil || !res.foundKey {
		return immutable.None[string](), err
	}

	hasNilField := false
	for i := range f.indexedFields {
		hasNilField = hasNilField || res.key.Fields[i].Value.IsNil()
	}

	if f.indexDesc.Unique && !hasNilField {
		f.currentDocID = immutable.Some(string(res.value))
	} else {
		lastVal := res.key.Fields[len(res.key.Fields)-1].Value
		if str, ok := lastVal.String(); ok {
			f.currentDocID = immutable.Some(str)
		} else if bytes, ok := lastVal.Bytes(); ok {
			f.currentDocID = immutable.Some(string(bytes))
		} else {
			f.currentDocID = immutable.None[string]()
		}
	}
	return f.currentDocID, nil
	//}
}

func (f *indexFetcher) GetFields() (immutable.Option[EncodedDocument], error) {
	if !f.currentDocID.HasValue() {
		return immutable.Option[EncodedDocument]{}, nil
	}
	var execInfo ExecInfo
	prefix := base.MakeDataStoreKeyWithCollectionAndDocID(f.col.Description(), f.currentDocID.Value())
	prefixFetcher, err := newPrefixFetcher(f.ctx, f.txn, []keys.DataStoreKey{prefix}, f.col,
		f.fieldsByID, client.Active, &execInfo)
	if err != nil {
		return immutable.Option[EncodedDocument]{}, err
	}
	_, err = prefixFetcher.NextDoc()
	if err != nil {
		return immutable.Option[EncodedDocument]{}, err
	}
	return prefixFetcher.GetFields()
}

func (f *indexFetcher) Close() error {
	if f.indexIter != nil {
		return f.indexIter.Close()
	}
	return nil
}

// resetState resets the mutable state of this IndexFetcher, returning the state to how it
// was immediately after construction.
func (f *indexFetcher) resetState() {
	// WARNING: Do not reset properties set in the constructor!

	f.col = nil
	f.mapping = nil
	f.indexedFields = nil
	f.indexIter = nil
	f.execInfo.Reset()
}
