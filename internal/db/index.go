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
	"github.com/sourcenetwork/defradb/internal/core"
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
	base := collectionBaseIndex{collection: collection, desc: desc}
	base.fieldsDescs = make([]client.SchemaFieldDescription, len(desc.Fields))
	isArray := false
	for i := range desc.Fields {
		field, foundField := collection.Schema().GetFieldByName(desc.Fields[i].Name)
		if !foundField {
			return nil, client.NewErrFieldNotExist(desc.Fields[i].Name)
		}
		base.fieldsDescs[i] = field
		if !isSupportedKind(field.Kind) {
			return nil, NewErrUnsupportedIndexFieldType(field.Kind)
		}
		isArray = isArray || field.Kind.IsArray()
	}
	if isArray {
		if desc.Unique {
			return newCollectionArrayUniqueIndex(base), nil
		} else {
			return newCollectionArrayIndex(base), nil
		}
	} else if desc.Unique {
		return &collectionUniqueIndex{collectionBaseIndex: base}, nil
	} else {
		return &collectionSimpleIndex{collectionBaseIndex: base}, nil
	}
}

type collectionBaseIndex struct {
	collection  client.Collection
	desc        client.IndexDescription
	fieldsDescs []client.SchemaFieldDescription
}

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
) (core.IndexDataStoreKey, error) {
	fieldValues, err := index.getDocFieldValues(doc)
	if err != nil {
		return core.IndexDataStoreKey{}, err
	}

	fields := make([]core.IndexedField, len(index.fieldsDescs))
	for i := range index.fieldsDescs {
		fields[i].Value = fieldValues[i]
		fields[i].Descending = index.desc.Fields[i].Descending
	}

	if appendDocID {
		fields = append(fields, core.IndexedField{Value: client.NewNormalString(doc.ID().String())})
	}
	return core.NewIndexDataStoreKey(index.collection.ID(), index.desc.ID, fields), nil
}

func (index *collectionBaseIndex) deleteIndexKey(
	ctx context.Context,
	txn datastore.Txn,
	key core.IndexDataStoreKey,
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
	prefixKey := core.IndexDataStoreKey{}
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

// collectionSimpleIndex is an non-unique index that indexes documents by a single field.
// Single-field indexes store values only in ascending order.
type collectionSimpleIndex struct {
	collectionBaseIndex
}

var _ CollectionIndex = (*collectionSimpleIndex)(nil)

func (index *collectionSimpleIndex) getDocumentsIndexKey(
	doc *client.Document,
) (core.IndexDataStoreKey, error) {
	return index.collectionBaseIndex.getDocumentsIndexKey(doc, true)
}

// Save indexes a document by storing the indexed field value.
func (index *collectionSimpleIndex) Save(
	ctx context.Context,
	txn datastore.Txn,
	doc *client.Document,
) error {
	key, err := index.getDocumentsIndexKey(doc)
	if err != nil {
		return err
	}
	err = txn.Datastore().Put(ctx, key.ToDS(), []byte{})
	if err != nil {
		return NewErrFailedToStoreIndexedField(key.ToString(), err)
	}
	return nil
}

func (index *collectionSimpleIndex) Update(
	ctx context.Context,
	txn datastore.Txn,
	oldDoc *client.Document,
	newDoc *client.Document,
) error {
	err := index.deleteDocIndex(ctx, txn, oldDoc)
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
	return index.deleteDocIndex(ctx, txn, doc)
}

func (index *collectionSimpleIndex) deleteDocIndex(
	ctx context.Context,
	txn datastore.Txn,
	doc *client.Document,
) error {
	key, err := index.getDocumentsIndexKey(doc)
	if err != nil {
		return err
	}
	return index.deleteIndexKey(ctx, txn, key)
}

// hasIndexKeyNilField returns true if the index key has a field with nil value
func hasIndexKeyNilField(key *core.IndexDataStoreKey) bool {
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

func (index *collectionUniqueIndex) save(
	ctx context.Context,
	txn datastore.Txn,
	key *core.IndexDataStoreKey,
	val []byte,
) error {
	err := txn.Datastore().Put(ctx, key.ToDS(), val)
	if err != nil {
		return NewErrFailedToStoreIndexedField(key.ToDS().String(), err)
	}
	return nil
}

func (index *collectionUniqueIndex) Save(
	ctx context.Context,
	txn datastore.Txn,
	doc *client.Document,
) error {
	key, val, err := index.prepareUniqueIndexRecordToStore(ctx, txn, doc)
	if err != nil {
		return err
	}
	return index.save(ctx, txn, &key, val)
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

func (index *collectionBaseIndex) getDocumentsUniqueIndexRecord(
	doc *client.Document,
) (core.IndexDataStoreKey, []byte, error) {
	key, err := index.getDocumentsIndexKey(doc, false)
	if err != nil {
		return core.IndexDataStoreKey{}, nil, err
	}
	return makeUniqueKeyValueRecord(key, doc)
}

func makeUniqueKeyValueRecord(
	key core.IndexDataStoreKey,
	doc *client.Document,
) (core.IndexDataStoreKey, []byte, error) {
	if hasIndexKeyNilField(&key) {
		key.Fields = append(key.Fields, core.IndexedField{Value: client.NewNormalString(doc.ID().String())})
		return key, []byte{}, nil
	} else {
		return key, []byte(doc.ID().String()), nil
	}
}

func (index *collectionUniqueIndex) prepareUniqueIndexRecordToStore(
	ctx context.Context,
	txn datastore.Txn,
	doc *client.Document,
) (core.IndexDataStoreKey, []byte, error) {
	key, val, err := index.getDocumentsUniqueIndexRecord(doc)
	if err != nil {
		return core.IndexDataStoreKey{}, nil, err
	}
	return key, val, validateUniqueKeyValue(ctx, txn, key, val, doc, index.fieldsDescs)
}

func validateUniqueKeyValue(
	ctx context.Context,
	txn datastore.Txn,
	key core.IndexDataStoreKey,
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

func (index *collectionUniqueIndex) Delete(
	ctx context.Context,
	txn datastore.Txn,
	doc *client.Document,
) error {
	return index.deleteDocIndex(ctx, txn, doc)
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
	newKey, newVal, err := index.prepareUniqueIndexRecordToStore(ctx, txn, newDoc)
	if err != nil {
		return err
	}
	err = index.deleteDocIndex(ctx, txn, oldDoc)
	if err != nil {
		return err
	}
	return index.save(ctx, txn, &newKey, newVal)
}

func (index *collectionUniqueIndex) deleteDocIndex(
	ctx context.Context,
	txn datastore.Txn,
	doc *client.Document,
) error {
	key, _, err := index.getDocumentsUniqueIndexRecord(doc)
	if err != nil {
		return err
	}
	return index.deleteIndexKey(ctx, txn, key)
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
		case oldVal.Value() != newVal.Value():
			return true
		}
	}
	return false
}

type collectionArrayBaseIndex struct {
	collectionBaseIndex
	arrFieldsIndexes []int
}

func newCollectionArrayBaseIndex(base collectionBaseIndex) collectionArrayBaseIndex {
	ind := collectionArrayBaseIndex{collectionBaseIndex: base}
	for i := range base.fieldsDescs {
		if base.fieldsDescs[i].Kind.IsArray() {
			ind.arrFieldsIndexes = append(ind.arrFieldsIndexes, i)
		}
	}
	if len(ind.arrFieldsIndexes) == 0 {
		return collectionArrayBaseIndex{}
	}
	return ind
}

// newIndexKeyGenerator creates a function that generates index keys for a document
// with multiple array fields.
// All generated keys are unique.
// For example for a doc with these values {{"a", "b", "a"}, {"c", "d", "e"}, {"f", "g"}} it generates:
// "acf", "acg", "adf", "adg", "aef", "aeg", "bcf", "bcg", "bdf", "bdg", "bef", "beg"
// Note: the example is simplified and doesn't include field separation
func (index *collectionArrayBaseIndex) newIndexKeyGenerator(
	doc *client.Document,
	appendDocID bool,
) (func() (core.IndexDataStoreKey, bool), error) {
	key, err := index.getDocumentsIndexKey(doc, appendDocID)
	if err != nil {
		return nil, err
	}

	normValsArr := make([][]client.NormalValue, 0, len(index.arrFieldsIndexes))
	for _, arrFieldIndex := range index.arrFieldsIndexes {
		arrVal := key.Fields[arrFieldIndex].Value
		normVals, err := client.ToArrayOfNormalValues(arrVal)
		if err != nil {
			return nil, err
		}
		sets := make(map[client.NormalValue]struct{})
		for i := len(normVals) - 1; i >= 0; i-- {
			if _, ok := sets[normVals[i]]; ok {
				normVals[i] = normVals[len(normVals)-1]
				normVals = normVals[:len(normVals)-1]
			} else {
				sets[normVals[i]] = struct{}{}
			}
		}
		normValsArr = append(normValsArr, normVals)
	}

	arrFieldCounter := make([]int, len(index.arrFieldsIndexes))
	done := false

	return func() (core.IndexDataStoreKey, bool) {
		if done {
			return core.IndexDataStoreKey{}, false
		}

		resultKey := core.IndexDataStoreKey{
			CollectionID: key.CollectionID,
			IndexID:      key.IndexID,
			Fields:       make([]core.IndexedField, len(key.Fields)),
		}
		copy(resultKey.Fields, key.Fields)

		for i, counter := range arrFieldCounter {
			field := &resultKey.Fields[index.arrFieldsIndexes[i]]
			field.Value = normValsArr[i][counter]
		}

		for i := len(arrFieldCounter) - 1; i >= 0; i-- {
			arrFieldCounter[i]++
			if arrFieldCounter[i] < len(normValsArr[i]) {
				break
			}
			arrFieldCounter[i] = 0
			if i == 0 {
				done = true
			}
		}

		return resultKey, true
	}, nil
}

func (index *collectionArrayBaseIndex) getAllKeys(
	doc *client.Document,
	appendDocID bool,
) ([]core.IndexDataStoreKey, error) {
	getNextOldKey, err := index.newIndexKeyGenerator(doc, appendDocID)
	if err != nil {
		return nil, err
	}
	oldKeys := make([]core.IndexDataStoreKey, 0)
	for {
		key, ok := getNextOldKey()
		if !ok {
			break
		}
		oldKeys = append(oldKeys, key)
	}
	return oldKeys, nil
}

func (index *collectionArrayBaseIndex) deleteRetiredKeysAndReturnNew(
	ctx context.Context,
	txn datastore.Txn,
	oldDoc *client.Document,
	newDoc *client.Document,
	appendDocID bool,
) ([]core.IndexDataStoreKey, error) {
	oldKeys, err := index.getAllKeys(oldDoc, appendDocID)
	if err != nil {
		return nil, err
	}
	newKeys, err := index.getAllKeys(newDoc, appendDocID)
	if err != nil {
		return nil, err
	}

	for _, oldKey := range oldKeys {
		isFound := false
		for i := len(newKeys) - 1; i >= 0; i-- {
			if oldKey.IsEqual(newKeys[i]) {
				newKeys[i] = newKeys[len(newKeys)-1]
				newKeys = newKeys[:len(newKeys)-1]
				isFound = true
				break
			}
		}
		if !isFound {
			err = index.deleteIndexKey(ctx, txn, oldKey)
			if err != nil {
				return nil, err
			}
		}
	}

	return newKeys, nil
}

type collectionArrayIndex struct {
	collectionArrayBaseIndex
}

var _ CollectionIndex = (*collectionArrayIndex)(nil)

func newCollectionArrayIndex(base collectionBaseIndex) *collectionArrayIndex {
	return &collectionArrayIndex{collectionArrayBaseIndex: newCollectionArrayBaseIndex(base)}
}

// Save indexes a document by storing the indexed field value.
func (index *collectionArrayIndex) Save(
	ctx context.Context,
	txn datastore.Txn,
	doc *client.Document,
) error {
	getNextKey, err := index.newIndexKeyGenerator(doc, true)
	if err != nil {
		return err
	}

	for {
		key, ok := getNextKey()
		if !ok {
			break
		}
		err = txn.Datastore().Put(ctx, key.ToDS(), []byte{})
		if err != nil {
			return NewErrFailedToStoreIndexedField(key.ToString(), err)
		}
	}
	return nil
}

func (index *collectionArrayIndex) Update(
	ctx context.Context,
	txn datastore.Txn,
	oldDoc *client.Document,
	newDoc *client.Document,
) error {
	newKeys, err := index.deleteRetiredKeysAndReturnNew(ctx, txn, oldDoc, newDoc, true)
	if err != nil {
		return err
	}

	for _, key := range newKeys {
		err = txn.Datastore().Put(ctx, key.ToDS(), []byte{})
		if err != nil {
			return NewErrFailedToStoreIndexedField(key.ToString(), err)
		}
	}

	return nil
}

func (index *collectionArrayIndex) Delete(
	ctx context.Context,
	txn datastore.Txn,
	doc *client.Document,
) error {
	getNextKey, err := index.newIndexKeyGenerator(doc, true)
	if err != nil {
		return err
	}

	for {
		key, ok := getNextKey()
		if !ok {
			break
		}
		err = index.deleteIndexKey(ctx, txn, key)
		if err != nil {
			return err
		}
	}
	return nil
}

type collectionArrayUniqueIndex struct {
	collectionArrayBaseIndex
}

var _ CollectionIndex = (*collectionArrayUniqueIndex)(nil)

func newCollectionArrayUniqueIndex(base collectionBaseIndex) *collectionArrayUniqueIndex {
	return &collectionArrayUniqueIndex{collectionArrayBaseIndex: newCollectionArrayBaseIndex(base)}
}

func (index *collectionArrayUniqueIndex) Save(
	ctx context.Context,
	txn datastore.Txn,
	doc *client.Document,
) error {
	getNextKey, err := index.newIndexKeyGenerator(doc, false)
	if err != nil {
		return err
	}

	for {
		key, ok := getNextKey()
		if !ok {
			break
		}
		err := index.addNewUniqueKey(ctx, txn, doc, key)
		if err != nil {
			return err
		}
	}
	return nil
}

func (index *collectionArrayUniqueIndex) addNewUniqueKey(
	ctx context.Context,
	txn datastore.Txn,
	doc *client.Document,
	key core.IndexDataStoreKey,
) error {
	key, val, err := makeUniqueKeyValueRecord(key, doc)
	if err != nil {
		return err
	}
	err = validateUniqueKeyValue(ctx, txn, key, val, doc, index.fieldsDescs)
	if err != nil {
		return err
	}
	err = txn.Datastore().Put(ctx, key.ToDS(), val)
	if err != nil {
		return NewErrFailedToStoreIndexedField(key.ToString(), err)
	}
	return nil
}

func (index *collectionArrayUniqueIndex) Update(
	ctx context.Context,
	txn datastore.Txn,
	oldDoc *client.Document,
	newDoc *client.Document,
) error {
	newKeys, err := index.deleteRetiredKeysAndReturnNew(ctx, txn, oldDoc, newDoc, false)
	if err != nil {
		return err
	}

	for _, key := range newKeys {
		err := index.addNewUniqueKey(ctx, txn, newDoc, key)
		if err != nil {
			return err
		}
	}

	return nil
}

func (index *collectionArrayUniqueIndex) Delete(
	ctx context.Context,
	txn datastore.Txn,
	doc *client.Document,
) error {
	getNextKey, err := index.newIndexKeyGenerator(doc, false)
	if err != nil {
		return err
	}

	for {
		key, ok := getNextKey()
		if !ok {
			break
		}
		err = index.deleteIndexKey(ctx, txn, key)
		if err != nil {
			return err
		}
	}
	return nil
}
