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
	"github.com/sourcenetwork/defradb/internal/db/vectorindex"
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
	base, err := buildIndexBase(collection, desc, building)
	if err != nil {
		return nil, err
	}
	// Read the epoch after validation so an invalid description fails the same way whether or not a
	// transaction is on the context.
	base.epoch, err = getIndexEpoch(ctx, collection.Version().CollectionID, desc.ID)
	if err != nil {
		return nil, err
	}
	return wrapCollectionIndex(base)
}

// newCollectionIndexWithEpoch builds an index instance pinned to a caller-resolved epoch, rather
// than re-reading the sequence. A backfill uses it so every batch writes the same epoch even if a
// concurrent version switch advances the sequence mid-build; splitting one build across two epochs
// would leave the live epoch missing the documents indexed before the advance. Live writes use
// NewCollectionIndex, which always targets the current epoch.
func newCollectionIndexWithEpoch(
	collection client.Collection,
	desc client.IndexDescription,
	building bool,
	epoch uint32,
) (client.CollectionIndex, error) {
	base, err := buildIndexBase(collection, desc, building)
	if err != nil {
		return nil, err
	}
	base.epoch = epoch
	return wrapCollectionIndex(base)
}

// buildIndexBase validates the description against the collection and assembles the shared index
// base, leaving the epoch unset for the caller to resolve.
func buildIndexBase(
	collection client.Collection,
	desc client.IndexDescription,
	building bool,
) (collectionBaseIndex, error) {
	if len(desc.Fields) == 0 {
		return collectionBaseIndex{}, NewErrIndexDescHasNoFields(desc)
	}
	base := collectionBaseIndex{
		collection:      collection,
		desc:            desc,
		building:        building,
		fieldsDescs:     make([]client.CollectionFieldDescription, len(desc.Fields)),
		fieldGenerators: make([]FieldIndexGenerator, len(desc.Fields)),
	}
	for i := range desc.Fields {
		field, foundField := collection.Version().GetFieldByName(desc.Fields[i].Name)
		if !foundField {
			return collectionBaseIndex{}, client.NewErrFieldNotExist(desc.Fields[i].Name)
		}
		base.fieldsDescs[i] = field
		if !isSupportedKind(field.Kind) {
			return collectionBaseIndex{}, NewErrUnsupportedIndexFieldType(field.Kind)
		}
		if field.Typ == client.PN_COUNTER || field.Typ == client.P_COUNTER {
			return collectionBaseIndex{}, NewErrCannotIndexAccumulatedCRDTField(field.Name, field.Typ.String())
		}
		base.fieldGenerators[i] = getFieldGenerator(field.Kind)
	}
	return base, nil
}

// wrapCollectionIndex returns the concrete index implementation for the base, dispatched by kind.
func wrapCollectionIndex(base collectionBaseIndex) (client.CollectionIndex, error) {
	switch base.desc.Kind() {
	case client.IndexKindVector:
		return newCollectionVectorIndex(base)
	default:
		if base.desc.Secondary.Unique {
			return &collectionUniqueIndex{collectionBaseIndex: base}, nil
		}
		return &collectionSimpleIndex{collectionBaseIndex: base}, nil
	}
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

// collectionVectorIndex is a vector index. Save/Update/Delete maintain it in the same transaction as
// the document write, through the algorithm-agnostic vectorindex package.
//
// The collection short id needs a store read, so it is read on first use and kept. The index handle
// is opened fresh on each call because it holds the request's transaction.
type collectionVectorIndex struct {
	collectionBaseIndex

	collectionShortID uint32
	shortIDResolved   bool
}

var _ client.CollectionIndex = (*collectionVectorIndex)(nil)

func newCollectionVectorIndex(base collectionBaseIndex) (client.CollectionIndex, error) {
	return &collectionVectorIndex{collectionBaseIndex: base}, nil
}

// resolveCollectionShortID returns the collection short id, reading it from the store on the first
// call and reusing it after. The id never changes, so this saves a store read on every write.
func (index *collectionVectorIndex) resolveCollectionShortID(ctx context.Context) (uint32, error) {
	if index.shortIDResolved {
		return index.collectionShortID, nil
	}
	shortID, err := id.GetCollectionShortID(ctx, index.collection.Version().CollectionID)
	if err != nil {
		return 0, err
	}
	index.collectionShortID = shortID
	index.shortIDResolved = true
	return shortID, nil
}

// openIndex opens the vector index for this description, reading and writing through the transaction
// on ctx. The vectorindex package selects the algorithm, so the planner opens the same index when
// searching. An unsupported algorithm or metric fails here, on first use.
func (index *collectionVectorIndex) openIndex(ctx context.Context) (vectorindex.Index, uint32, error) {
	collectionShortID, err := index.resolveCollectionShortID(ctx)
	if err != nil {
		return nil, 0, err
	}

	idx, err := vectorindex.Open(ctx, collectionShortID, index.desc.ID, index.epoch, *index.desc.Vector)
	if err != nil {
		return nil, 0, err
	}
	return idx, collectionShortID, nil
}

// nodeAndVector returns the node id (the document's short id) and the vector to index for doc.
// found is false when the document has no short id, which the callers decide how to treat. vec is
// nil when the document has no value for the field, meaning there is nothing to index.
func (index *collectionVectorIndex) nodeAndVector(
	ctx context.Context,
	collectionShortID uint32,
	doc *client.Document,
) (uint64, []float32, bool, error) {
	docShortID, found, err := id.GetDocShortID(ctx, collectionShortID, doc.ID().String())
	if err != nil || !found {
		return 0, nil, found, err
	}

	fieldVal, err := doc.TryGetValue(index.fieldsDescs[0].Name)
	if err != nil {
		return 0, nil, false, err
	}
	if fieldVal == nil || fieldVal.Value() == nil {
		// No vector on this doc, so nothing to index.
		return docShortID, nil, true, nil
	}

	vec, ok := fieldVal.NormalValue().Float32Array()
	if !ok {
		return 0, nil, false, NewErrVectorIndexFieldNotFloat32Array(index.fieldsDescs[0].Name, doc.ID().String())
	}

	return docShortID, vec, true, nil
}

// Save indexes doc by inserting its vector. If the document has no value for the indexed field,
// there is nothing to index and Save does nothing.
func (index *collectionVectorIndex) Save(ctx context.Context, doc *client.Document) error {
	idx, collectionShortID, err := index.openIndex(ctx)
	if err != nil {
		return err
	}

	nodeID, vec, found, err := index.nodeAndVector(ctx, collectionShortID, doc)
	if err != nil {
		return err
	}
	if !found {
		// A document being written always has a short id by this point. Not finding one means the
		// document is missing or the caller cannot access it, which is an error, not a doc to skip.
		return client.ErrDocumentNotFoundOrNotAuthorized
	}
	if vec == nil {
		return nil
	}

	// Vectors of different lengths in one graph make distances meaningless. Dimensions is set for a
	// directly-written field, so check against it. It is 0 only for an @embedding field, where the
	// model fixes the length, so the only guard left is against an empty vector.
	if index.desc.Vector.Dimensions > 0 && len(vec) != int(index.desc.Vector.Dimensions) {
		return NewErrVectorDimensionMismatch(int(index.desc.Vector.Dimensions), len(vec), doc.ID().String())
	}
	if len(vec) == 0 {
		return NewErrVectorIndexEmptyVector(index.fieldsDescs[0].Name, doc.ID().String())
	}

	return idx.Insert(nodeID, vec)
}

// Update re-indexes a document whose vector changed. A document keeps the same id across an update,
// so the old and new vectors map to the same node; delete-then-save re-inserts it. (The binding
// handles the in-place replacement; see the vectorindex package.)
func (index *collectionVectorIndex) Update(ctx context.Context, oldDoc, newDoc *client.Document) error {
	if err := index.Delete(ctx, oldDoc); err != nil {
		return err
	}
	return index.Save(ctx, newDoc)
}

// Delete removes doc from the index. While the index is still building, the backfill may not have
// reached this document yet, so a missing short id is expected and Delete does nothing. Once the
// index is built, a missing short id means the index is out of step with the data.
func (index *collectionVectorIndex) Delete(ctx context.Context, doc *client.Document) error {
	idx, collectionShortID, err := index.openIndex(ctx)
	if err != nil {
		return err
	}

	// Delete needs only the node id, not the vector, so resolve the short id directly. Going through
	// nodeAndVector would also re-decode the field value and fail if it cannot, blocking the delete of
	// a document whose vector became undecodable.
	nodeID, found, err := id.GetDocShortID(ctx, collectionShortID, doc.ID().String())
	if err != nil {
		return err
	}
	if !found {
		if index.building {
			return nil
		}
		return NewErrCorruptedVectorIndex(index.desc.Name, doc.ID().String())
	}

	return idx.Delete(nodeID)
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
		return saveUniqueKey(ctx, doc, key, index.fieldsDescs, index.building)
	})
}

// saveUniqueKey writes a unique index entry for doc.
//
// Keys whose value is empty embed the docID in the key itself, so they are already
// doc-specific and are written unconditionally. For value-bearing keys, an entry that
// already exists is a uniqueness violation, except while the index is building, where
// an entry for the same doc means a concurrent live write got there first and is skipped.
func saveUniqueKey(
	ctx context.Context,
	doc *client.Document,
	key keys.IndexDataStoreKey,
	fieldsDescs []client.CollectionFieldDescription,
	tolerateSameDoc bool,
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
		existing, err := txn.Datastore().Get(ctx, &key)
		if err != nil && !errors.Is(err, corekv.ErrNotFound) {
			return NewErrCheckUniqueIndexConstraint(err)
		}
		if existing != nil {
			if tolerateSameDoc && string(existing) == string(val) {
				return nil
			}
			return newUniqueIndexError(doc, fieldsDescs)
		}
	}

	if err := txn.Datastore().Set(ctx, &key, val); err != nil {
		return NewErrFailedToStoreIndexedField(key.ToString(), err)
	}
	return nil
}

func newUniqueIndexError(doc *client.Document, fieldsDescs []client.CollectionFieldDescription) error {
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
