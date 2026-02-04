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
	"encoding/binary"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/connor"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/planner/filter"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

// ReadIndexEpoch returns the index's current epoch: the value of its epoch sequence.
//
// The sequence is seeded when the index is created, so a missing sequence is an inconsistent
// state and returns an error rather than defaulting, which would scan the wrong namespace.
func ReadIndexEpoch(ctx context.Context, txn datastore.Txn, collectionID string, indexID uint32) (uint32, error) {
	collectionShortID, err := id.GetCollectionShortID(ctx, collectionID)
	if err != nil {
		return 0, err
	}
	val, err := txn.Systemstore().Get(ctx, keys.NewIndexEpochSequenceKey(collectionShortID, indexID).Bytes())
	if err != nil {
		if errors.Is(err, corekv.ErrNotFound) {
			return 0, NewErrIndexEpochNotFound(err, collectionID, indexID)
		}
		return 0, err
	}
	return uint32(binary.BigEndian.Uint64(val)), nil
}

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
	currentDocShortID immutable.Option[uint64]
	collectionShortID uint32
	execInfo          *ExecInfo
	ordering          []mapper.OrderCondition
	// epoch is the namespace this fetcher scans, resolved from the index's epoch sequence.
	epoch uint32
}

var _ fetcher = (*indexFetcher)(nil)

// newIndexFetcher creates a new IndexFetcher.
// It can return nil if there is no efficient way to fetch with given filter conditions.
// If startKey is provided, the iterator opens at that position for cursor-based pagination.
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
	startKey immutable.Option[keys.Walkable],
) (*indexFetcher, error) {
	// Check if the filter has an OR at the root level that spans different fields.
	// This check MUST happen here before filter.CopyField strips out non-indexed fields,
	// otherwise the orIndexIterator would only see partial OR branches and return incomplete results.
	if docFilter != nil && hasOrWithMultipleFields(docFilter.Conditions, indexDesc, docMapper) {
		return nil, nil
	}

	collectionShortID, err := id.GetCollectionShortID(ctx, col.Version().CollectionID)
	if err != nil {
		return nil, err
	}
	epoch, err := ReadIndexEpoch(ctx, txn, col.Version().CollectionID, indexDesc.ID)
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
		epoch:             epoch,
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
		field, ok := f.col.Version().GetFieldByName(indexedField.Name)
		if ok {
			f.indexedFields = append(f.indexedFields, field)
		}
	}

	iter, err := f.createIndexIterator(f.indexFilter)
	if err != nil || iter == nil {
		return nil, err
	}

	if startKey.HasValue() {
		if matchIter, ok := iter.(*indexMatchIterator); ok {
			matchIter.prefixKey = nil
			baseKey, err := f.newIndexDataStoreKey()
			if err != nil {
				return nil, err
			}

			if matchIter.reverse {
				// For reverse iteration, the cursor position becomes the upper bound (endKey).
				// We iterate backward from just before the cursor position.
				// The startKey (lower bound) is the index prefix to iterate down to.
				matchIter.startKey = &baseKey
				matchIter.endKey = startKey.Value()
			} else {
				// For forward iteration, the cursor position becomes the lower bound (startKey).
				// We iterate forward from just after the cursor position.
				// The endKey (upper bound) is the end of the index prefix.
				matchIter.startKey = startKey.Value()
				if matchIter.endKey == nil {
					matchIter.endKey = baseKey.PrefixEnd()
				}
			}
		}
	}

	f.indexIter = iter
	return f, iter.Init(ctx, txn.Datastore())
}

func (f *indexFetcher) NextDoc() (immutable.Option[string], error) {
	f.currentDocID = immutable.None[string]()
	f.currentDocShortID = immutable.None[uint64]()

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
		docShortID, err := keys.DecodeDocShortID(res.value)
		if err != nil {
			return immutable.None[string](), err
		}
		docID, err := f.docIDFromDocShortID(docShortID)
		if err != nil {
			return immutable.None[string](), err
		}
		f.currentDocID = immutable.Some(docID)
		f.currentDocShortID = immutable.Some(docShortID)
	} else {
		// Non-unique index entries must carry the doc suffix.
		if res.key.DocShortID == 0 {
			return immutable.None[string](), NewErrUnexpectedTypeValue[uint64](res.key.DocShortID)
		}
		docID, err := f.docIDFromDocShortID(res.key.DocShortID)
		if err != nil {
			return immutable.None[string](), err
		}
		f.currentDocID = immutable.Some(docID)
		f.currentDocShortID = immutable.Some(res.key.DocShortID)
	}
	return f.currentDocID, nil
}

func (f *indexFetcher) docIDFromDocShortID(docShortID uint64) (string, error) {
	docID, found, err := id.GetDocID(f.ctx, docShortID)
	if err != nil {
		return "", err
	}
	if found {
		return docID, nil
	}
	return "", nil
}

func (f *indexFetcher) GetFields() (immutable.Option[EncodedDocument], error) {
	if !f.currentDocID.HasValue() {
		return immutable.Option[EncodedDocument]{}, nil
	}

	prefix := keys.DataStoreKey{
		CollectionShortID: f.collectionShortID,
		DocShortID:        f.currentDocShortID.Value(),
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
