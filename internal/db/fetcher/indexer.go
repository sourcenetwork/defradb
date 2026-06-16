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
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/connor"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/planner/filter"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

// indexFetcher is a fetcher that fetches documents by index.
// It fetches only the indexed field and the rest of the fields are fetched by the internal fetcher.
type indexFetcher struct {
	ctx               context.Context
	txn               datastore.Txn
	col               client.Collection
	indexFilter       *mapper.Filter
	mapping           *core.DocumentMapping
	indexedFields     []client.CollectionFieldDescription
	fieldsByID        map[uint32]client.CollectionFieldDescription
	indexDesc         client.IndexDescription
	indexIter         indexIterator
	currentDocID      immutable.Option[string]
	currentShortDocID immutable.Option[uint32]
	collectionShortID uint32
	execInfo          *ExecInfo
	ordering          []mapper.OrderCondition
}

var _ fetcher = (*indexFetcher)(nil)

// newIndexFetcher creates a new IndexFetcher.
// It can return nil, if there is no efficient way to fetch indexes with given filter conditions.
func newIndexFetcher(
	ctx context.Context,
	txn datastore.Txn,
	fieldsByID map[uint32]client.CollectionFieldDescription,
	indexDesc client.IndexDescription,
	docFilter *mapper.Filter,
	col client.Collection,
	docMapper *core.DocumentMapping,
	execInfo *ExecInfo,
	ordering []mapper.OrderCondition,
) (*indexFetcher, error) {
	// Check if the filter has an OR at the root level that spans different fields.
	// This check MUST happen here before filter.CopyField strips out non-indexed fields,
	// otherwise the orIndexIterator would only see partial OR branches and return incomplete results.
	if docFilter != nil && hasOrWithMultipleFields(docFilter.Conditions, indexDesc, docMapper) {
		return nil, nil
	}

	collectionShortID, err := id.GetShortCollectionID(ctx, col.Version().CollectionID)
	if err != nil {
		return nil, err
	}

	f := &indexFetcher{
		ctx:               ctx,
		txn:               txn,
		col:               col,
		mapping:           docMapper,
		indexDesc:         indexDesc,
		fieldsByID:        fieldsByID,
		collectionShortID: collectionShortID,
		execInfo:          execInfo,
		ordering:          ordering,
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
		if indexedField.Name == request.DocIDFieldName {
			f.indexedFields = append(f.indexedFields, client.CollectionFieldDescription{
				Name: indexedField.Name,
				Kind: client.FieldKind_DocID,
			})
			continue
		}
		field, ok := f.col.Version().GetFieldByName(indexedField.Name)
		if ok {
			f.indexedFields = append(f.indexedFields, field)
		}
	}

	iter, err := f.createIndexIterator(f.indexFilter)
	if err != nil || iter == nil {
		return nil, err
	}

	f.indexIter = iter
	return f, iter.Init(ctx, txn.Datastore())
}

func (f *indexFetcher) NextDoc() (immutable.Option[string], error) {
	f.currentDocID = immutable.None[string]()
	f.currentShortDocID = immutable.None[uint32]()

	res, err := f.indexIter.Next()
	if err != nil {
		return immutable.None[string](), NewErrGetNextIndexEntry(err, f.indexDesc.Name)
	}
	if !res.foundKey {
		return immutable.None[string](), nil
	}

	hasNilField := false
	for i := range f.indexedFields {
		hasNilField = hasNilField || res.key.Fields[i].Value.IsNil()
	}

	if f.indexDesc.Unique && !hasNilField {
		shortDocID, err := keys.DecodeDocShortID(res.value)
		if err != nil {
			return immutable.None[string](), err
		}
		docID, err := f.publicDocID(shortDocID)
		if err != nil {
			return immutable.None[string](), err
		}
		f.currentDocID = immutable.Some(docID)
		f.currentShortDocID = immutable.Some(shortDocID)
	} else {
		lastVal := res.key.Fields[len(res.key.Fields)-1].Value
		if shortDocID, ok, err := normalShortDocID(lastVal); err != nil {
			return immutable.None[string](), err
		} else if ok {
			docID, err := f.publicDocID(shortDocID)
			if err != nil {
				return immutable.None[string](), err
			}
			f.currentDocID = immutable.Some(docID)
			f.currentShortDocID = immutable.Some(shortDocID)
		} else {
			f.currentDocID = immutable.None[string]()
		}
	}
	return f.currentDocID, nil
}

func normalShortDocID(val client.NormalValue) (uint32, bool, error) {
	encodedShortDocID, ok := val.Bytes()
	if !ok {
		return 0, false, nil
	}
	shortDocID, err := keys.DecodeDocShortID(encodedShortDocID)
	if err != nil {
		return 0, false, err
	}
	return shortDocID, true, nil
}

func (f *indexFetcher) publicDocID(shortDocID uint32) (string, error) {
	publicDocID, found, err := id.GetPublicDocID(f.ctx, f.collectionShortID, shortDocID)
	if err != nil {
		return "", err
	}
	if found {
		return publicDocID, nil
	}
	return "", nil
}

func (f *indexFetcher) GetFields() (immutable.Option[EncodedDocument], error) {
	if !f.currentDocID.HasValue() {
		return immutable.Option[EncodedDocument]{}, nil
	}

	docID := ""
	var shortDocID uint32
	if f.currentShortDocID.HasValue() {
		shortDocID = f.currentShortDocID.Value()
	} else {
		var err error
		docID = f.currentDocID.Value()
		var found bool
		shortDocID, found, err = id.GetShortDocID(f.ctx, f.collectionShortID, docID)
		if err != nil {
			return immutable.None[EncodedDocument](), err
		}
		if !found {
			return immutable.None[EncodedDocument](), nil
		}
	}

	prefix := keys.DataStoreKey{
		CollectionShortID: f.collectionShortID,
		DocShortID:        shortDocID,
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

// CanBeOrderedByIndex checks if the index can be used to order by the fields in the ordering array.
// The first return value specifies if index can be used.
// The second one specifies if the index should be reversed to match the ordering.
func CanBeOrderedByIndex(
	ordering []mapper.OrderCondition,
	index client.IndexDescription,
	mapping *core.DocumentMapping,
) (bool, bool) {
	// if there is no ordering in the query or the query requests ordering on more fields, then index
	// contains, we can't use index
	if len(ordering) == 0 || len(ordering) > len(index.Fields) {
		return false, false
	}

	orderMismatchCount := 0

	for i := range len(ordering) {
		fieldIndexes := mapping.IndexesByName[index.Fields[i].Name]

		// if indexed field doesn't match the ordering field, we can't use index
		if len(fieldIndexes) == 0 || fieldIndexes[0] != ordering[i].FieldIndexes[0] {
			return false, false
		}

		isDescending := ordering[i].Direction == mapper.DESC
		if index.Fields[i].Descending != isDescending {
			orderMismatchCount++
		}
	}

	// if ordering of all fields matches, we can use index
	// also if ordering of all indexes doesn't match we can use index by reversing it
	allMismatches := orderMismatchCount == len(ordering)
	return orderMismatchCount == 0 || allMismatches, allMismatches
}

// hasOrWithMultipleFields checks if the filter conditions have an _or operator at the root level
// where branches reference fields not covered by this index.
// This check MUST happen here before filter.CopyField strips out non-indexed fields,
// otherwise the orIndexIterator would only see partial OR branches and return incomplete results.
func hasOrWithMultipleFields(
	conditions map[connor.FilterKey]any,
	indexDesc client.IndexDescription,
	docMapper *core.DocumentMapping,
) bool {
	branches := extractOrBranches(conditions)
	if branches == nil {
		return false
	}

	for _, branch := range branches {
		hasNonIndexedField := false
		filter.TraverseProperties(branch, func(prop *mapper.PropertyIndex, _ map[connor.FilterKey]any) bool {
			for _, field := range indexDesc.Fields {
				if docMapper.FirstIndexOfName(field.Name) == prop.Index {
					return true
				}
			}
			hasNonIndexedField = true
			return false // Field not in index, stop traversal
		})

		if hasNonIndexedField {
			return true
		}
	}
	return false
}
