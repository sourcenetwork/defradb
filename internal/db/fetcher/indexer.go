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
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/db/id"
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
	execInfo      *ExecInfo
	ordering      []mapper.OrderCondition
}

var _ fetcher = (*indexFetcher)(nil)

// newIndexFetcher creates a new IndexFetcher.
// It can return nil, if there is no efficient way to fetch indexes with given filter conditions.
func newIndexFetcher(
	ctx context.Context,
	txn datastore.Txn,
	fieldsByID map[uint32]client.FieldDefinition,
	indexDesc client.IndexDescription,
	docFilter *mapper.Filter,
	col client.Collection,
	docMapper *core.DocumentMapping,
	execInfo *ExecInfo,
	ordering []mapper.OrderCondition,
) (*indexFetcher, error) {
	f := &indexFetcher{
		ctx:        ctx,
		txn:        txn,
		col:        col,
		mapping:    docMapper,
		indexDesc:  indexDesc,
		fieldsByID: fieldsByID,
		execInfo:   execInfo,
		ordering:   ordering,
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
	f.currentDocID = immutable.None[string]()

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
		} else {
			f.currentDocID = immutable.None[string]()
		}
	}
	return f.currentDocID, nil
}

func (f *indexFetcher) GetFields() (immutable.Option[EncodedDocument], error) {
	if !f.currentDocID.HasValue() {
		return immutable.Option[EncodedDocument]{}, nil
	}

	shortID, err := id.GetShortCollectionID(f.ctx, f.txn, f.col.Version().CollectionID)
	if err != nil {
		return immutable.None[EncodedDocument](), err
	}

	prefix := keys.DataStoreKey{
		CollectionShortID: shortID,
		DocID:             f.currentDocID.Value(),
	}
	prefixFetcher, err := newPrefixFetcher(f.ctx, f.txn, []keys.DataStoreKey{prefix}, f.col,
		f.fieldsByID, client.Active, f.execInfo)
	if err != nil {
		return immutable.Option[EncodedDocument]{}, err
	}
	_, err = prefixFetcher.NextDoc()
	if err != nil {
		return immutable.Option[EncodedDocument]{}, err
	}
	doc, err := prefixFetcher.GetFields()
	return doc, errors.Join(err, prefixFetcher.Close())
}

func (f *indexFetcher) Close() error {
	if f.indexIter != nil {
		return f.indexIter.Close()
	}
	return nil
}
