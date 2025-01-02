// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package db

import (
	"context"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/utils/slice"
)

// CollectionIndex is an interface for collection indexes
// It abstracts away common index functionality to be implemented
// by different index types: non-unique, unique, and composite
type CollectionIndex interface {
	client.CollectionIndex
	// RemoveAll removes all documents from the index
	RemoveAll(context.Context, datastore.Txn) error
}

func isSupportedKind(kind client.FieldKind) bool {
	if kind.IsObject() && !kind.IsArray() {
		return true
	}

	switch kind {
	case
		client.FieldKind_DocID,
		client.FieldKind_STRING_ARRAY,
		client.FieldKind_INT_ARRAY,
		client.FieldKind_BOOL_ARRAY,
		client.FieldKind_FLOAT_ARRAY,
		client.FieldKind_NILLABLE_JSON,
		client.FieldKind_NILLABLE_STRING,
		client.FieldKind_NILLABLE_INT,
		client.FieldKind_NILLABLE_FLOAT,
		client.FieldKind_NILLABLE_BOOL,
		client.FieldKind_NILLABLE_BLOB,
		client.FieldKind_NILLABLE_DATETIME,
		client.FieldKind_NILLABLE_BOOL_ARRAY,
		client.FieldKind_NILLABLE_INT_ARRAY,
		client.FieldKind_NILLABLE_FLOAT_ARRAY,
		client.FieldKind_NILLABLE_STRING_ARRAY:
		return true
	default:
		return false
	}
}

// NewCollectionIndex creates a new collection index
func NewCollectionIndex(
	collection client.Collection,
	desc client.IndexDescription,
) (CollectionIndex, error) {
	if len(desc.Fields) == 0 {
		return nil, NewErrIndexDescHasNoFields(desc)
	}
	base := collectionBaseIndex{
		collection:      collection,
		desc:            desc,
		fieldsDescs:     make([]client.SchemaFieldDescription, len(desc.Fields)),
		fieldGenerators: make([]FieldIndexGenerator, len(desc.Fields)),
	}
	for i := range desc.Fields {
		field, foundField := collection.Schema().GetFieldByName(desc.Fields[i].Name)
		if !foundField {
			return nil, client.NewErrFieldNotExist(desc.Fields[i].Name)
		}
		base.fieldsDescs[i] = field
		if !isSupportedKind(field.Kind) {
			return nil, NewErrUnsupportedIndexFieldType(field.Kind)
		}
		base.fieldGenerators[i] = getFieldGenerator(field.Kind)
	}
	if desc.Unique {
		return &collectionUniqueIndex{collectionBaseIndex: base}, nil
	}
	return &collectionSimpleIndex{collectionBaseIndex: base}, nil
}

// FieldIndexGenerator generates index entries for a single field
type FieldIndexGenerator interface {
	// Generate calls the provided function for each value that should be indexed
	Generate(value client.NormalValue, f func(client.NormalValue) error) error
}

type SimpleFieldGenerator struct{}

func (g *SimpleFieldGenerator) Generate(value client.NormalValue, f func(client.NormalValue) error) error {
	return f(value)
}

type ArrayFieldGenerator struct{}

func (g *ArrayFieldGenerator) Generate(value client.NormalValue, f func(client.NormalValue) error) error {
	normVals, err := client.ToArrayOfNormalValues(value)
	if err != nil {
		return err
	}

	// Remove duplicates to avoid duplicate index entries
	uniqueVals := slice.RemoveDuplicates(normVals)
	for _, val := range uniqueVals {
		if err := f(val); err != nil {
			return err
		}
	}
	return nil
}

type JSONFieldGenerator struct{}

func (g *JSONFieldGenerator) Generate(value client.NormalValue, f func(client.NormalValue) error) error {
	json, _ := value.JSON()
	return client.TraverseJSON(json, func(value client.JSON) error {
		val, err := client.NewNormalValue(value)
		if err != nil {
			return err
		}
		return f(val)
	}, client.TraverseJSONOnlyLeaves(), client.TraverseJSONVisitArrayElements(false))
}

// getFieldGenerator returns appropriate generator for the field type
func getFieldGenerator(kind client.FieldKind) FieldIndexGenerator {
	if kind.IsArray() {
		return &ArrayFieldGenerator{}
	}
	if kind == client.FieldKind_NILLABLE_JSON {
		return &JSONFieldGenerator{}
	}
	return &SimpleFieldGenerator{}
}

type collectionBaseIndex struct {
	collection client.Collection
	desc       client.IndexDescription
	// fieldsDescs is a slice of field descriptions for the fields that form the index
	// If there is more than 1 field, the index is composite
	fieldsDescs     []client.SchemaFieldDescription
	fieldGenerators []FieldIndexGenerator
}

// getDocFieldValues retrieves the values of the indexed fields from the given document.
func (index *collectionBaseIndex) getDocFieldValues(doc *client.Document) ([]client.NormalValue, error) {
	result := make([]client.NormalValue, 0, len(index.fieldsDescs))
	for iter := range index.fieldsDescs {
		fieldVal, err := doc.TryGetValue(index.fieldsDescs[iter].Name)
		if err != nil {
			return nil, err
		}
		if fieldVal == nil || fieldVal.Value() == nil {
			normalNil, err := client.NewNormalNil(index.fieldsDescs[iter].Kind)
			if err != nil {
				return nil, err
			}
			result = append(result, normalNil)
			continue
		}
		result = append(result, fieldVal.NormalValue())
	}
	return result, nil
}

func (index *collectionBaseIndex) getDocumentsIndexKey(
	doc *client.Document,
	appendDocID bool,
) (keys.IndexDataStoreKey, error) {
	fieldValues, err := index.getDocFieldValues(doc)
	if err != nil {
		return keys.IndexDataStoreKey{}, err
	}

	fields := make([]keys.IndexedField, len(index.fieldsDescs))
	for i := range index.fieldsDescs {
		fields[i].Value = fieldValues[i]
		fields[i].Descending = index.desc.Fields[i].Descending
	}

	if appendDocID {
		fields = append(fields, keys.IndexedField{Value: client.NewNormalString(doc.ID().String())})
	}
	return keys.NewIndexDataStoreKey(index.collection.ID(), index.desc.ID, fields), nil
}

func (index *collectionBaseIndex) deleteIndexKey(
	ctx context.Context,
	txn datastore.Txn,
	key keys.IndexDataStoreKey,
) error {
	exists, err := txn.Datastore().Has(ctx, key.ToDS())
	if err != nil {
		return err
	}
	if !exists {
		return NewErrCorruptedIndex(index.desc.Name)
	}
	return txn.Datastore().Delete(ctx, key.ToDS())
}

// RemoveAll remove all artifacts of the index from the storage, i.e. all index
// field values for all documents.
func (index *collectionBaseIndex) RemoveAll(ctx context.Context, txn datastore.Txn) error {
	prefixKey := keys.IndexDataStoreKey{}
	prefixKey.CollectionID = index.collection.ID()
	prefixKey.IndexID = index.desc.ID

	keys, err := datastore.FetchKeysForPrefix(ctx, prefixKey.ToString(), txn.Datastore())
	if err != nil {
		return err
	}

	for _, key := range keys {
		err := txn.Datastore().Delete(ctx, key)
		if err != nil {
			return NewCanNotDeleteIndexedField(err)
		}
	}

	return nil
}

// Name returns the name of the index
func (index *collectionBaseIndex) Name() string {
	return index.desc.Name
}

// Description returns the description of the index
func (index *collectionBaseIndex) Description() client.IndexDescription {
	return index.desc
}

// generateKeysAndProcess generates index keys for the given document and calls the provided function
// for each generated key
func (index *collectionBaseIndex) generateKeysAndProcess(
	doc *client.Document,
	appendDocID bool,
	processKey func(keys.IndexDataStoreKey) error,
) error {
	// Get initial key with base values
	baseKey, err := index.getDocumentsIndexKey(doc, appendDocID)
	if err != nil {
		return err
	}

	// Start with first field
	return index.generateKeysForFieldAndProcess(0, baseKey, processKey)
}

func (index *collectionBaseIndex) generateKeysForFieldAndProcess(
	fieldIdx int,
	baseKey keys.IndexDataStoreKey,
	processKey func(keys.IndexDataStoreKey) error,
) error {
	// If we've processed all fields, call the handler
	if fieldIdx >= len(index.fieldsDescs) {
		return processKey(baseKey)
	}

	// Generate values for current field
	return index.fieldGenerators[fieldIdx].Generate(
		baseKey.Fields[fieldIdx].Value,
		func(val client.NormalValue) error {
			// Create new key with generated value
			newKey := baseKey
			newKey.Fields = make([]keys.IndexedField, len(baseKey.Fields))
			copy(newKey.Fields, baseKey.Fields)
			newKey.Fields[fieldIdx].Value = val

			// Process next field
			return index.generateKeysForFieldAndProcess(fieldIdx+1, newKey, processKey)
		},
	)
}

// collectionSimpleIndex is an non-unique index that indexes documents by a single field.
// Single-field indexes store values only in ascending order.
type collectionSimpleIndex struct {
	collectionBaseIndex
}

var _ CollectionIndex = (*collectionSimpleIndex)(nil)

// Save indexes a document by storing the indexed field value.
func (index *collectionSimpleIndex) Save(
	ctx context.Context,
	txn datastore.Txn,
	doc *client.Document,
) error {
	return index.generateKeysAndProcess(doc, true, func(key keys.IndexDataStoreKey) error {
		return txn.Datastore().Put(ctx, key.ToDS(), []byte{})
	})
}

func (index *collectionSimpleIndex) Update(
	ctx context.Context,
	txn datastore.Txn,
	oldDoc *client.Document,
	newDoc *client.Document,
) error {
	err := index.Delete(ctx, txn, oldDoc)
	if err != nil {
		return err
	}
	return index.Save(ctx, txn, newDoc)
}

func (index *collectionSimpleIndex) Delete(
	ctx context.Context,
	txn datastore.Txn,
	doc *client.Document,
) error {
	return index.generateKeysAndProcess(doc, true, func(key keys.IndexDataStoreKey) error {
		return index.deleteIndexKey(ctx, txn, key)
	})
}

// hasIndexKeyNilField returns true if the index key has a field with nil value
func hasIndexKeyNilField(key *keys.IndexDataStoreKey) bool {
	for i := range key.Fields {
		if key.Fields[i].Value.IsNil() {
			return true
		}
	}
	return false
}

type collectionUniqueIndex struct {
	collectionBaseIndex
}

var _ CollectionIndex = (*collectionUniqueIndex)(nil)

func (index *collectionUniqueIndex) Save(
	ctx context.Context,
	txn datastore.Txn,
	doc *client.Document,
) error {
	return index.generateKeysAndProcess(doc, false, func(key keys.IndexDataStoreKey) error {
		return addNewUniqueKey(ctx, txn, doc, key, index.fieldsDescs)
	})
}

func newUniqueIndexError(doc *client.Document, fieldsDescs []client.SchemaFieldDescription) error {
	kvs := make([]errors.KV, 0, len(fieldsDescs))
	for iter := range fieldsDescs {
		fieldVal, err := doc.TryGetValue(fieldsDescs[iter].Name)
		var val any
		if err != nil {
			return err
		}
		// If fieldVal is nil, we leave `val` as is (e.g. nil)
		if fieldVal != nil {
			val = fieldVal.Value()
		}
		kvs = append(kvs, errors.NewKV(fieldsDescs[iter].Name, val))
	}

	return NewErrCanNotIndexNonUniqueFields(doc.ID().String(), kvs...)
}

func makeUniqueKeyValueRecord(
	key keys.IndexDataStoreKey,
	doc *client.Document,
) (keys.IndexDataStoreKey, []byte, error) {
	if hasIndexKeyNilField(&key) {
		key.Fields = append(key.Fields, keys.IndexedField{Value: client.NewNormalString(doc.ID().String())})
		return key, []byte{}, nil
	} else {
		return key, []byte(doc.ID().String()), nil
	}
}

func validateUniqueKeyValue(
	ctx context.Context,
	txn datastore.Txn,
	key keys.IndexDataStoreKey,
	val []byte,
	doc *client.Document,
	fieldsDescs []client.SchemaFieldDescription,
) error {
	if len(val) != 0 {
		exists, err := txn.Datastore().Has(ctx, key.ToDS())
		if err != nil {
			return err
		}
		if exists {
			return newUniqueIndexError(doc, fieldsDescs)
		}
	}
	return nil
}

func addNewUniqueKey(
	ctx context.Context,
	txn datastore.Txn,
	doc *client.Document,
	key keys.IndexDataStoreKey,
	fieldsDescs []client.SchemaFieldDescription,
) error {
	key, val, err := makeUniqueKeyValueRecord(key, doc)
	if err != nil {
		return err
	}
	err = validateUniqueKeyValue(ctx, txn, key, val, doc, fieldsDescs)
	if err != nil {
		return err
	}
	err = txn.Datastore().Put(ctx, key.ToDS(), val)
	if err != nil {
		return NewErrFailedToStoreIndexedField(key.ToString(), err)
	}
	return nil
}

func (index *collectionUniqueIndex) Delete(
	ctx context.Context,
	txn datastore.Txn,
	doc *client.Document,
) error {
	return index.generateKeysAndProcess(doc, false, func(key keys.IndexDataStoreKey) error {
		key, _, err := makeUniqueKeyValueRecord(key, doc)
		if err != nil {
			return err
		}
		return txn.Datastore().Delete(ctx, key.ToDS())
	})
}

func (index *collectionUniqueIndex) Update(
	ctx context.Context,
	txn datastore.Txn,
	oldDoc *client.Document,
	newDoc *client.Document,
) error {
	// We only need to update the index if one of the indexed fields
	// on the document has been changed.
	if !isUpdatingIndexedFields(index, oldDoc, newDoc) {
		return nil
	}

	err := index.Delete(ctx, txn, oldDoc)
	if err != nil {
		return err
	}

	return index.Save(ctx, txn, newDoc)
}

func isUpdatingIndexedFields(index CollectionIndex, oldDoc, newDoc *client.Document) bool {
	for _, indexedFields := range index.Description().Fields {
		oldVal, getOldValErr := oldDoc.GetValue(indexedFields.Name)
		newVal, getNewValErr := newDoc.GetValue(indexedFields.Name)

		// GetValue will return an error when the field doesn't exist.
		// This will happen for oldDoc only if the field hasn't been set
		// when first creating the document. For newDoc, this will happen
		// only if the field hasn't been set when first creating the document
		// AND the field hasn't been set on the update.
		switch {
		case getOldValErr != nil && getNewValErr != nil:
			continue
		case getOldValErr != nil && getNewValErr == nil:
			return true
		case !oldVal.NormalValue().Equal(newVal.NormalValue()):
			return true
		}
	}
	return false
}
