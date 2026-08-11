// Copyright 2025 Democratized Data Foundation
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

	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/utils/slice"
)

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
		client.FieldKind_FLOAT32_ARRAY,
		client.FieldKind_FLOAT64_ARRAY,
		client.FieldKind_DATETIME_ARRAY,
		client.FieldKind_NILLABLE_JSON,
		client.FieldKind_NILLABLE_STRING,
		client.FieldKind_NILLABLE_INT,
		client.FieldKind_NILLABLE_FLOAT32,
		client.FieldKind_NILLABLE_FLOAT64,
		client.FieldKind_NILLABLE_BOOL,
		client.FieldKind_NILLABLE_BLOB,
		client.FieldKind_NILLABLE_DATETIME,
		client.FieldKind_NILLABLE_BOOL_ARRAY,
		client.FieldKind_NILLABLE_INT_ARRAY,
		client.FieldKind_NILLABLE_FLOAT32_ARRAY,
		client.FieldKind_NILLABLE_FLOAT64_ARRAY,
		client.FieldKind_NILLABLE_STRING_ARRAY,
		client.FieldKind_NILLABLE_DATETIME_ARRAY,
		client.FieldKind_JSON,
		client.FieldKind_STRING,
		client.FieldKind_INT,
		client.FieldKind_FLOAT32,
		client.FieldKind_FLOAT64,
		client.FieldKind_BOOL,
		client.FieldKind_BLOB,
		client.FieldKind_DATETIME:
		return true
	default:
		return false
	}
}

// NewCollectionIndex adds a new collection index.
//
// While building is true the index is being backfilled: Save tolerates a unique-index entry
// already written by a concurrent live write of the same document, and Delete tolerates a
// missing entry for a document the backfill has not yet reached.
func NewCollectionIndex(
	ctx context.Context,
	collection client.Collection,
	desc client.IndexDescription,
	building bool,
) (client.CollectionIndex, error) {
	if len(desc.Fields) == 0 {
		return nil, NewErrIndexDescHasNoFields(desc)
	}
	base := collectionBaseIndex{
		collection:      collection,
		desc:            desc,
		building:        building,
		fieldsDescs:     make([]client.CollectionFieldDescription, len(desc.Fields)),
		fieldGenerators: make([]FieldIndexGenerator, len(desc.Fields)),
	}
	base.stats = statsFor(collection)
	for i := range desc.Fields {
		field, foundField := collection.Version().GetFieldByName(desc.Fields[i].Name)
		if !foundField {
			return nil, client.NewErrFieldNotExist(desc.Fields[i].Name)
		}
		base.fieldsDescs[i] = field
		if !isSupportedKind(field.Kind) {
			return nil, NewErrUnsupportedIndexFieldType(field.Kind)
		}
		if field.Typ == client.PN_COUNTER || field.Typ == client.P_COUNTER {
			return nil, NewErrCannotIndexAccumulatedCRDTField(field.Name, field.Typ.String())
		}
		base.fieldGenerators[i] = getFieldGenerator(field.Kind)
	}

	epoch, err := getIndexEpoch(ctx, collection.Version().CollectionID, desc.ID)
	if err != nil {
		return nil, err
	}
	base.epoch = epoch
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
	if json == nil {
		val, err := client.NewNormalNil(client.FieldKind_NILLABLE_JSON)
		if err != nil {
			return err
		}
		return f(val)
	}
	return client.TraverseJSON(json, func(value client.JSON) error {
		val, err := client.NewNormalValue(value)
		if err != nil {
			return err
		}
		return f(val)
	},
		// we don't want to traverse intermediate nodes, because we encode only values that can be filtered on
		client.TraverseJSONOnlyLeaves(),
		// we want to include array elements' indexes in json path, because we want to differentiate
		// between array elements in order to be able to run array-specific queries like _all, _any and _none
		client.TraverseJSONWithArrayIndexInPath(),
		// we want to traverse array elements, but not recurse into them, because we don't have any way
		// to query nested arrays elements.
		// this effectively means that we traverse only leave array elements (string, float, bool, null)
		client.TraverseJSONVisitArrayElements(false),
	)
}

// getFieldGenerator returns appropriate generator for the field type
func getFieldGenerator(kind client.FieldKind) FieldIndexGenerator {
	if kind.IsArray() {
		return &ArrayFieldGenerator{}
	}
	if kind == client.FieldKind_NILLABLE_JSON || kind == client.FieldKind_JSON {
		return &JSONFieldGenerator{}
	}
	return &SimpleFieldGenerator{}
}

type collectionBaseIndex struct {
	collection client.Collection
	desc       client.IndexDescription
	// fieldsDescs is a slice of field descriptions for the fields that form the index
	// If there is more than 1 field, the index is composite
	fieldsDescs     []client.CollectionFieldDescription
	fieldGenerators []FieldIndexGenerator
	// building is true while the index is being backfilled. deleteIndexKey tolerates
	// missing entries for documents not yet reached by the backfill.
	building bool
	// epoch is the namespace this instance reads and writes, resolved from the index's epoch
	// sequence at construction. During a rebuild the sequence names the epoch being built, so
	// live writes maintain it.
	epoch uint32
	// stats counts unique-index existence checks. Nil when the index was built for a
	// collection this package did not construct.
	stats *mergeStats
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
	ctx context.Context,
	doc *client.Document,
	appendDocShortID bool,
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

	collectionShortID, err := id.GetCollectionShortID(ctx, index.collection.Version().CollectionID)
	if err != nil {
		return keys.IndexDataStoreKey{}, err
	}
	var docShortID uint64
	if appendDocShortID {
		var found bool
		docShortID, found, err = id.GetDocShortID(ctx, collectionShortID, doc.ID().String())
		if err != nil {
			return keys.IndexDataStoreKey{}, err
		}
		if !found {
			return keys.IndexDataStoreKey{}, client.ErrDocumentNotFoundOrNotAuthorized
		}
	}

	key := keys.NewIndexDataStoreKey(collectionShortID, index.desc.ID, index.epoch, fields)
	key.DocShortID = docShortID
	return key, nil
}

// deleteIndexKey removes a single index entry. While the index is building, a missing
// entry is tolerated, since not every document has been backfilled yet.
func (index *collectionBaseIndex) deleteIndexKey(
	ctx context.Context,
	key keys.IndexDataStoreKey,
) error {
	txn := datastore.CtxMustGetTxn(ctx)
	ds := txn.Datastore()
	// An index entry is a datastore key and encodes like a document key, so the store
	// cannot tell them apart. Mark it here, where the type is known.
	datastore.ReadKindsOf(txn).Mark(datastore.ReadIndex)
	exists, err := ds.Has(ctx, &key)
	if err != nil {
		return NewErrCheckIndexKeyExists(err, index.desc.Name)
	}
	if !exists {
		// During backfill, documents not yet reached have no entry, so we skip silently.
		if index.building {
			return nil
		}
		return NewErrCorruptedIndex(index.desc.Name)
	}
	err = ds.Delete(ctx, &key)
	if err != nil {
		return NewErrDeleteIndexKey(err)
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
	ctx context.Context,
	doc *client.Document,
	appendDocShortID bool,
	processKey func(keys.IndexDataStoreKey) error,
) error {
	// Get initial key with base values
	baseKey, err := index.getDocumentsIndexKey(ctx, doc, appendDocShortID)
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

var _ client.CollectionIndex = (*collectionSimpleIndex)(nil)

// Save indexes a document by storing the indexed field value.
func (index *collectionSimpleIndex) Save(
	ctx context.Context,
	doc *client.Document,
) error {
	txn := datastore.CtxMustGetTxn(ctx)

	return index.generateKeysAndProcess(ctx, doc, true, func(key keys.IndexDataStoreKey) error {
		err := txn.Datastore().Set(ctx, &key, []byte{})
		if err != nil {
			return NewErrStoreIndexKey(err)
		}
		return nil
	})
}

func (index *collectionSimpleIndex) Update(
	ctx context.Context,
	oldDoc *client.Document,
	newDoc *client.Document,
) error {
	err := index.Delete(ctx, oldDoc)
	if err != nil {
		return NewErrUpdateIndex(err, index.desc.Name)
	}
	if err := index.Save(ctx, newDoc); err != nil {
		return NewErrUpdateIndex(err, index.desc.Name)
	}
	return nil
}

func (index *collectionSimpleIndex) Delete(
	ctx context.Context,
	doc *client.Document,
) error {
	return index.generateKeysAndProcess(ctx, doc, true, func(key keys.IndexDataStoreKey) error {
		return index.deleteIndexKey(ctx, key)
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

var _ client.CollectionIndex = (*collectionUniqueIndex)(nil)

func (index *collectionUniqueIndex) Save(
	ctx context.Context,
	doc *client.Document,
) error {
	return index.generateKeysAndProcess(ctx, doc, false, func(key keys.IndexDataStoreKey) error {
		return saveUniqueKey(ctx, doc, key, index.fieldsDescs, index.building, index.stats)
	})
}

// saveUniqueKey writes a unique index entry for doc.
//
// Keys whose value is empty embed the docID in the key itself, so they are already
// doc-specific and are written unconditionally. For value-bearing keys, an entry that
// already exists is a uniqueness violation — except while the index is building, where
// an entry for the same doc means a concurrent live write got there first and is skipped.
func saveUniqueKey(
	ctx context.Context,
	doc *client.Document,
	key keys.IndexDataStoreKey,
	fieldsDescs []client.CollectionFieldDescription,
	tolerateSameDoc bool,
	stats *mergeStats,
) error {
	txn := datastore.CtxMustGetTxn(ctx)

	docShortID, found, err := id.GetDocShortID(ctx, key.CollectionShortID, doc.ID().String())
	if err != nil {
		return err
	}
	if !found {
		return client.ErrDocumentNotFoundOrNotAuthorized
	}
	key, val, err := makeUniqueKeyValueRecord(key, docShortID)
	if err != nil {
		return err
	}

	if len(val) != 0 {
		// This read puts the unique key in the transaction's read set, so two transactions
		// writing the same index value conflict at commit rather than at this check.
		datastore.ReadKindsOf(txn).Mark(datastore.ReadUniqueIndex)
		existing, err := txn.Datastore().Get(ctx, &key)
		if err != nil && !errors.Is(err, corekv.ErrNotFound) {
			return NewErrCheckUniqueIndexConstraint(err)
		}
		stats.markUniqueIndexCheck(existing != nil)
		if existing != nil {
			if tolerateSameDoc && string(existing) == string(val) {
				return nil
			}
			return newUniqueIndexError(ctx, doc, fieldsDescs, existing)
		}
	}

	if err := txn.Datastore().Set(ctx, &key, val); err != nil {
		return NewErrFailedToStoreIndexedField(key.ToString(), err)
	}
	return nil
}

// incumbentDocID resolves the document already holding a unique index entry, from the
// short ID the entry stores. Returns an empty string if it cannot be resolved, since this
// only decorates an error that is being returned either way.
func incumbentDocID(ctx context.Context, encodedShortID []byte) string {
	shortID, err := keys.DecodeDocShortID(encodedShortID)
	if err != nil {
		return ""
	}
	docID, found, err := id.GetDocID(ctx, shortID)
	if err != nil || !found {
		return ""
	}
	return docID
}

func newUniqueIndexError(
	ctx context.Context,
	doc *client.Document,
	fieldsDescs []client.CollectionFieldDescription,
	existing []byte,
) error {
	kvs := make([]errors.KV, 0, len(fieldsDescs)+1)
	// Naming the document that already holds the value is what makes the two comparable;
	// without it the error only says which one lost.
	if incumbent := incumbentDocID(ctx, existing); incumbent != "" {
		kvs = append(kvs, errors.NewKV("HeldBy", incumbent))
	}
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
	docShortID uint64,
) (keys.IndexDataStoreKey, []byte, error) {
	encodedDocShortID := keys.EncodeDocShortID(docShortID)
	if hasIndexKeyNilField(&key) {
		key.DocShortID = docShortID
		return key, []byte{}, nil
	} else {
		return key, encodedDocShortID, nil
	}
}

func (index *collectionUniqueIndex) Delete(
	ctx context.Context,
	doc *client.Document,
) error {
	txn := datastore.CtxMustGetTxn(ctx)
	collectionShortID, err := id.GetCollectionShortID(ctx, index.collection.Version().CollectionID)
	if err != nil {
		return err
	}
	docShortID, found, err := id.GetDocShortID(ctx, collectionShortID, doc.ID().String())
	if err != nil {
		return err
	}
	if !found {
		return client.ErrDocumentNotFoundOrNotAuthorized
	}
	return index.generateKeysAndProcess(ctx, doc, false, func(key keys.IndexDataStoreKey) error {
		key, _, err := makeUniqueKeyValueRecord(key, docShortID)
		if err != nil {
			return err
		}
		err = txn.Datastore().Delete(ctx, &key)
		if err != nil {
			return NewErrDeleteIndexKey(err)
		}
		return nil
	})
}

func (index *collectionUniqueIndex) Update(
	ctx context.Context,
	oldDoc *client.Document,
	newDoc *client.Document,
) error {
	// We only need to update the index if one of the indexed fields
	// on the document has been changed.
	if !isUpdatingIndexedFields(index, oldDoc, newDoc) {
		return nil
	}

	err := index.Delete(ctx, oldDoc)
	if err != nil {
		return NewErrUpdateIndex(err, index.desc.Name)
	}

	if err := index.Save(ctx, newDoc); err != nil {
		return NewErrUpdateIndex(err, index.desc.Name)
	}
	return nil
}

func isUpdatingIndexedFields(index client.CollectionIndex, oldDoc, newDoc *client.Document) bool {
	for _, indexedFields := range index.Description().Fields {
		oldVal, getOldValErr := oldDoc.GetValue(indexedFields.Name)
		newVal, getNewValErr := newDoc.GetValue(indexedFields.Name)

		// GetValue will return an error when the field doesn't exist.
		// This will happen for oldDoc only if the field hasn't been set
		// when first adding the document. For newDoc, this will happen
		// only if the field hasn't been set when first adding the document
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
