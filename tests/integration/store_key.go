// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package tests

import (
	"strconv"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// KeyFactory is the entry point for building different types of keys
type KeyFactory struct{}

// NewKey creates a new key factory
func NewKey() KeyFactory {
	return KeyFactory{}
}

// DatastoreDoc starts building a DataStoreKey for document data
func (f KeyFactory) DatastoreDoc() *DatastoreDocKeyBuilder {
	return &DatastoreDocKeyBuilder{
		instanceType: keys.ValueKey, // Default to ValueKey
	}
}

// DatastoreIndex starts building an IndexDataStoreKey for index entries
func (f KeyFactory) DatastoreIndex() *DatastoreIndexKeyBuilder {
	return &DatastoreIndexKeyBuilder{}
}

// DatastoreDocKeyBuilder builds keys for document data in the datastore
type DatastoreDocKeyBuilder struct {
	collectionIndex int
	docIndex        int
	fieldName       string
	instanceType    keys.InstanceType
}

// Col sets the collection index
func (b *DatastoreDocKeyBuilder) Col(index int) *DatastoreDocKeyBuilder {
	b.collectionIndex = index
	return b
}

// DocID sets the document index
func (b *DatastoreDocKeyBuilder) DocID(index int) *DatastoreDocKeyBuilder {
	b.docIndex = index
	return b
}

// Field sets the field name
func (b *DatastoreDocKeyBuilder) Field(fieldName string) *DatastoreDocKeyBuilder {
	b.fieldName = fieldName
	return b
}

// InstanceType sets the instance type (default is keys.ValueKey)
func (b *DatastoreDocKeyBuilder) InstanceType(t keys.InstanceType) *DatastoreDocKeyBuilder {
	b.instanceType = t
	return b
}

// Build implements KeyBuilder interface
func (b *DatastoreDocKeyBuilder) Build(s *state) (keys.Key, error) {
	// Get the first node's database context
	// We need a proper transaction context to query the datastore
	node := s.nodes[0]
	txn, err := node.NewTxn(s.ctx, true)
	if err != nil {
		return nil, err
	}
	defer txn.Discard(s.ctx)

	ctx := db.InitContext(s.ctx, txn)

	col := s.nodes[0].collections[b.collectionIndex]
	collectionID := col.Version().CollectionID

	shortID, err := id.GetShortCollectionID(ctx, collectionID)
	if err != nil {
		return nil, err
	}

	fieldShortID, err := id.GetShortFieldID(ctx, shortID, b.fieldName)
	if err != nil {
		return nil, err
	}

	docID := s.docIDs[b.collectionIndex][b.docIndex]

	key := keys.DataStoreKey{
		CollectionShortID: shortID,
		InstanceType:      b.instanceType,
		DocID:             docID.String(),
		FieldID:           strconv.FormatUint(uint64(fieldShortID), 10),
	}

	return key, nil
}

// DatastoreIndexKeyBuilder builds keys for index entries in the datastore
type DatastoreIndexKeyBuilder struct {
	collectionIndex int
	indexID         int
	fields          []indexFieldEntry
}

// indexFieldEntry holds field value information for building index keys
type indexFieldEntry struct {
	value      any // Can be client.NormalValue or DocIndex
	descending bool
}

// Col sets the collection index
func (b *DatastoreIndexKeyBuilder) Col(index int) *DatastoreIndexKeyBuilder {
	b.collectionIndex = index
	return b
}

// IndexID sets the index ID
func (b *DatastoreIndexKeyBuilder) IndexID(id int) *DatastoreIndexKeyBuilder {
	b.indexID = id
	return b
}

// Field adds a field value to the index key.
// Can be called multiple times to add multiple fields in order.
// The order matters - fields should be added in the same order as defined in the index.
// Value can be either client.NormalValue or DocIndex.
func (b *DatastoreIndexKeyBuilder) Field(value any, descending bool) *DatastoreIndexKeyBuilder {
	b.fields = append(b.fields, indexFieldEntry{
		value:      value,
		descending: descending,
	})
	return b
}

// Build implements KeyBuilder interface
func (b *DatastoreIndexKeyBuilder) Build(s *state) (keys.Key, error) {
	// Get the first node's database context
	// We need a proper transaction context to query the datastore
	node := s.nodes[0]
	txn, err := node.NewTxn(s.ctx, true)
	if err != nil {
		return nil, err
	}
	defer txn.Discard(s.ctx)

	ctx := db.InitContext(s.ctx, txn)

	col := s.nodes[0].collections[b.collectionIndex]
	collectionID := col.Version().CollectionID

	shortID, err := id.GetShortCollectionID(ctx, collectionID)
	if err != nil {
		return nil, err
	}

	// Convert field entries to IndexedField, resolving DocIndex to actual values
	indexedFields := make([]keys.IndexedField, len(b.fields))
	for i, f := range b.fields {
		var value client.NormalValue
		
		switch v := f.value.(type) {
		case client.NormalValue:
			value = v
		case DocIndex:
			// Resolve DocIndex to actual document ID
			if v.CollectionIndex >= len(s.docIDs) || v.Index >= len(s.docIDs[v.CollectionIndex]) {
				return nil, errors.New(
					"document not found for DocIndex",
					errors.NewKV("CollectionIndex", v.CollectionIndex),
					errors.NewKV("Index", v.Index),
				)
			}
			docID := s.docIDs[v.CollectionIndex][v.Index]
			value = client.NewNormalString(docID.String())
		default:
			return nil, client.NewErrUnexpectedType[any]("Field value", v)
		}
		
		indexedFields[i] = keys.IndexedField{
			Value:      value,
			Descending: f.descending,
		}
	}

	// Build IndexDataStoreKey
	key := &keys.IndexDataStoreKey{
		CollectionShortID: shortID,
		IndexID:           uint32(b.indexID),
		Fields:            indexedFields,
	}

	return key, nil
}
