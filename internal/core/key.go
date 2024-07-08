// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package core

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
)

var (
	// KeyMin is a minimum key value which sorts before all other keys.
	KeyMin = []byte{}
	// KeyMax is a maximum key value which sorts after all other keys.
	KeyMax = []byte{0xff, 0xff}
)

// InstanceType is a type that represents the type of instance.
type InstanceType string

const (
	// ValueKey is a type that represents a value instance.
	ValueKey = InstanceType("v")
	// PriorityKey is a type that represents a priority instance.
	PriorityKey = InstanceType("p")
	// DeletedKey is a type that represents a deleted document.
	DeletedKey = InstanceType("d")
)

const (
	COLLECTION                     = "collection"
	COLLECTION_ID                  = "/collection/id"
	COLLECTION_NAME                = "/collection/name"
	COLLECTION_SCHEMA_VERSION      = "/collection/version"
	COLLECTION_INDEX               = "/collection/index"
	SCHEMA_VERSION                 = "/schema/version/v"
	SCHEMA_VERSION_ROOT            = "/schema/version/r"
	COLLECTION_SEQ                 = "/seq/collection"
	INDEX_ID_SEQ                   = "/seq/index"
	FIELD_ID_SEQ                   = "/seq/field"
	PRIMARY_KEY                    = "/pk"
	DATASTORE_DOC_VERSION_FIELD_ID = "v"
	REPLICATOR                     = "/replicator/id"
	P2P_COLLECTION                 = "/p2p/collection"
)

// Key is an interface that represents a key in the database.
type Key interface {
	ToString() string
	Bytes() []byte
	ToDS() ds.Key
}

// DataStoreKey is a type that represents a key in the database.
type DataStoreKey struct {
	CollectionRootID uint32
	InstanceType     InstanceType
	DocID            string
	FieldID          string
}

var _ Key = (*DataStoreKey)(nil)

// IndexedField contains information necessary for storing a single
// value of a field in an index.
type IndexedField struct {
	// Value is the value of the field in the index
	Value client.NormalValue
	// Descending is true if the field is sorted in descending order
	Descending bool
}

// IndexDataStoreKey is key of an indexed document in the database.
type IndexDataStoreKey struct {
	// CollectionID is the id of the collection
	CollectionID uint32
	// IndexID is the id of the index
	IndexID uint32
	// Fields is the values of the fields in the index
	Fields []IndexedField
}

var _ Key = (*IndexDataStoreKey)(nil)

type PrimaryDataStoreKey struct {
	CollectionRootID uint32
	DocID            string
}

var _ Key = (*PrimaryDataStoreKey)(nil)

type HeadStoreKey struct {
	DocID   string
	FieldId string //can be 'C'
	Cid     cid.Cid
}

var _ Key = (*HeadStoreKey)(nil)

// CollectionKey points to the json serialized description of the
// the collection of the given ID.
type CollectionKey struct {
	CollectionID uint32
}

var _ Key = (*CollectionKey)(nil)

// CollectionNameKey points to the ID of the collection of the given
// name.
type CollectionNameKey struct {
	Name string
}

var _ Key = (*CollectionNameKey)(nil)

// CollectionSchemaVersionKey points to nil, but the keys/prefix can be used
// to get collections that are using, or have used a given schema version.
//
// If a collection is updated to a different schema version, the old entry(s)
// of this key will be preserved.
//
// This key should be removed in https://github.com/sourcenetwork/defradb/issues/1085
type CollectionSchemaVersionKey struct {
	SchemaVersionID string
	CollectionID    uint32
}

var _ Key = (*CollectionSchemaVersionKey)(nil)

// CollectionIndexKey to a stored description of an index
type CollectionIndexKey struct {
	// CollectionID is the id of the collection that the index is on
	CollectionID immutable.Option[uint32]
	// IndexName is the name of the index
	IndexName string
}

var _ Key = (*CollectionIndexKey)(nil)

// SchemaVersionKey points to the json serialized schema at the specified version.
//
// It's corresponding value is immutable.
type SchemaVersionKey struct {
	SchemaVersionID string
}

var _ Key = (*SchemaVersionKey)(nil)

// SchemaRootKey indexes schema version ids by their root schema id.
//
// The index is the key, there are no values stored against the key.
type SchemaRootKey struct {
	SchemaRoot      string
	SchemaVersionID string
}

var _ Key = (*SchemaRootKey)(nil)

type P2PCollectionKey struct {
	CollectionID string
}

var _ Key = (*P2PCollectionKey)(nil)

// CollectionIDSequenceKey is used to key the sequence used to generate collection ids.
type CollectionIDSequenceKey struct{}

var _ Key = (*CollectionIDSequenceKey)(nil)

// IndexIDSequenceKey is used to key the sequence used to generate index ids.
//
// The sequence is specific to each collection version.
type IndexIDSequenceKey struct {
	CollectionID uint32
}

var _ Key = (*IndexIDSequenceKey)(nil)

// FieldIDSequenceKey is used to key the sequence used to generate field ids.
//
// The sequence is specific to each collection root.  Multiple collection of the same root
// must maintain consistent field ids.
type FieldIDSequenceKey struct {
	CollectionRoot uint32
}

var _ Key = (*FieldIDSequenceKey)(nil)

type ReplicatorKey struct {
	ReplicatorID string
}

var _ Key = (*ReplicatorKey)(nil)

// Creates a new DataStoreKey from a string as best as it can,
// splitting the input using '/' as a field deliminator.  It assumes
// that the input string is in the following format:
//
// /[CollectionRootId]/[InstanceType]/[DocID]/[FieldId]
//
// Any properties before the above (assuming a '/' deliminator) are ignored
func NewDataStoreKey(key string) (DataStoreKey, error) {
	return DecodeDataStoreKey([]byte(key))
}

func MustNewDataStoreKey(key string) DataStoreKey {
	dsKey, err := NewDataStoreKey(key)
	if err != nil {
		panic(err)
	}
	return dsKey
}

func DataStoreKeyFromDocID(docID client.DocID) DataStoreKey {
	return DataStoreKey{
		DocID: docID.String(),
	}
}

// Creates a new HeadStoreKey from a string as best as it can,
// splitting the input using '/' as a field deliminator.  It assumes
// that the input string is in the following format:
//
// /[DocID]/[FieldId]/[Cid]
//
// Any properties before the above are ignored
func NewHeadStoreKey(key string) (HeadStoreKey, error) {
	elements := strings.Split(key, "/")
	if len(elements) != 4 {
		return HeadStoreKey{}, ErrInvalidKey
	}

	cid, err := cid.Decode(elements[3])
	if err != nil {
		return HeadStoreKey{}, err
	}

	return HeadStoreKey{
		// elements[0] is empty (key has leading '/')
		DocID:   elements[1],
		FieldId: elements[2],
		Cid:     cid,
	}, nil
}

// Returns a formatted collection key for the system data store.
// It assumes the name of the collection is non-empty.
func NewCollectionKey(id uint32) CollectionKey {
	return CollectionKey{CollectionID: id}
}

func NewCollectionNameKey(name string) CollectionNameKey {
	return CollectionNameKey{Name: name}
}

func NewCollectionSchemaVersionKey(schemaVersionId string, collectionID uint32) CollectionSchemaVersionKey {
	return CollectionSchemaVersionKey{
		SchemaVersionID: schemaVersionId,
		CollectionID:    collectionID,
	}
}

func NewCollectionSchemaVersionKeyFromString(key string) (CollectionSchemaVersionKey, error) {
	elements := strings.Split(key, "/")
	colID, err := strconv.Atoi(elements[len(elements)-1])
	if err != nil {
		return CollectionSchemaVersionKey{}, err
	}

	return CollectionSchemaVersionKey{
		SchemaVersionID: elements[len(elements)-2],
		CollectionID:    uint32(colID),
	}, nil
}

// NewCollectionIndexKey creates a new CollectionIndexKey from a collection name and index name.
func NewCollectionIndexKey(colID immutable.Option[uint32], indexName string) CollectionIndexKey {
	return CollectionIndexKey{CollectionID: colID, IndexName: indexName}
}

// NewCollectionIndexKeyFromString creates a new CollectionIndexKey from a string.
// It expects the input string is in the following format:
//
// /collection/index/[CollectionID]/[IndexName]
//
// Where [IndexName] might be omitted. Anything else will return an error.
func NewCollectionIndexKeyFromString(key string) (CollectionIndexKey, error) {
	keyArr := strings.Split(key, "/")
	if len(keyArr) < 4 || len(keyArr) > 5 || keyArr[1] != COLLECTION || keyArr[2] != "index" {
		return CollectionIndexKey{}, ErrInvalidKey
	}

	colID, err := strconv.Atoi(keyArr[3])
	if err != nil {
		return CollectionIndexKey{}, err
	}

	result := CollectionIndexKey{CollectionID: immutable.Some(uint32(colID))}
	if len(keyArr) == 5 {
		result.IndexName = keyArr[4]
	}
	return result, nil
}

// ToString returns the string representation of the key
// It is in the following format:
// /collection/index/[CollectionID]/[IndexName]
// if [CollectionID] is empty, the rest is ignored
func (k CollectionIndexKey) ToString() string {
	result := COLLECTION_INDEX

	if k.CollectionID.HasValue() {
		result = result + "/" + fmt.Sprint(k.CollectionID.Value())
		if k.IndexName != "" {
			result = result + "/" + k.IndexName
		}
	}

	return result
}

// Bytes returns the byte representation of the key
func (k CollectionIndexKey) Bytes() []byte {
	return []byte(k.ToString())
}

// ToDS returns the datastore key
func (k CollectionIndexKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

func NewSchemaVersionKey(schemaVersionID string) SchemaVersionKey {
	return SchemaVersionKey{SchemaVersionID: schemaVersionID}
}

func NewSchemaRootKey(schemaRoot string, schemaVersionID string) SchemaRootKey {
	return SchemaRootKey{
		SchemaRoot:      schemaRoot,
		SchemaVersionID: schemaVersionID,
	}
}

func NewSchemaRootKeyFromString(keyString string) (SchemaRootKey, error) {
	keyString = strings.TrimPrefix(keyString, SCHEMA_VERSION_ROOT+"/")
	elements := strings.Split(keyString, "/")
	if len(elements) != 2 {
		return SchemaRootKey{}, ErrInvalidKey
	}

	return SchemaRootKey{
		SchemaRoot:      elements[0],
		SchemaVersionID: elements[1],
	}, nil
}

func NewIndexIDSequenceKey(collectionID uint32) IndexIDSequenceKey {
	return IndexIDSequenceKey{CollectionID: collectionID}
}

func NewFieldIDSequenceKey(collectionRoot uint32) FieldIDSequenceKey {
	return FieldIDSequenceKey{CollectionRoot: collectionRoot}
}

func (k DataStoreKey) WithValueFlag() DataStoreKey {
	newKey := k
	newKey.InstanceType = ValueKey
	return newKey
}

func (k DataStoreKey) WithPriorityFlag() DataStoreKey {
	newKey := k
	newKey.InstanceType = PriorityKey
	return newKey
}

func (k DataStoreKey) WithDeletedFlag() DataStoreKey {
	newKey := k
	newKey.InstanceType = DeletedKey
	return newKey
}

func (k DataStoreKey) WithDocID(docID string) DataStoreKey {
	newKey := k
	newKey.DocID = docID
	return newKey
}

func (k DataStoreKey) WithInstanceInfo(key DataStoreKey) DataStoreKey {
	newKey := k
	newKey.DocID = key.DocID
	newKey.FieldID = key.FieldID
	newKey.InstanceType = key.InstanceType
	return newKey
}

func (k DataStoreKey) WithFieldId(fieldId string) DataStoreKey {
	newKey := k
	newKey.FieldID = fieldId
	return newKey
}

func (k DataStoreKey) ToHeadStoreKey() HeadStoreKey {
	return HeadStoreKey{
		DocID:   k.DocID,
		FieldId: k.FieldID,
	}
}

func (k HeadStoreKey) WithDocID(docID string) HeadStoreKey {
	newKey := k
	newKey.DocID = docID
	return newKey
}

func (k HeadStoreKey) WithCid(c cid.Cid) HeadStoreKey {
	newKey := k
	newKey.Cid = c
	return newKey
}

func (k HeadStoreKey) WithFieldId(fieldId string) HeadStoreKey {
	newKey := k
	newKey.FieldId = fieldId
	return newKey
}

func (k DataStoreKey) ToString() string {
	return string(k.Bytes())
}

func (k DataStoreKey) Bytes() []byte {
	return EncodeDataStoreKey(&k)
}

func (k DataStoreKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

func (k DataStoreKey) PrettyPrint() string {
	var result string

	if k.CollectionRootID != 0 {
		result = result + "/" + strconv.Itoa(int(k.CollectionRootID))
	}
	if k.InstanceType != "" {
		result = result + "/" + string(k.InstanceType)
	}
	if k.DocID != "" {
		result = result + "/" + k.DocID
	}
	if k.FieldID != "" {
		result = result + "/" + k.FieldID
	}

	return result
}

func (k DataStoreKey) Equal(other DataStoreKey) bool {
	return k.CollectionRootID == other.CollectionRootID &&
		k.DocID == other.DocID &&
		k.FieldID == other.FieldID &&
		k.InstanceType == other.InstanceType
}

func (k DataStoreKey) ToPrimaryDataStoreKey() PrimaryDataStoreKey {
	return PrimaryDataStoreKey{
		CollectionRootID: k.CollectionRootID,
		DocID:            k.DocID,
	}
}

// NewIndexDataStoreKey creates a new IndexDataStoreKey from a collection ID, index ID and fields.
// It also validates values of the fields.
func NewIndexDataStoreKey(collectionID, indexID uint32, fields []IndexedField) IndexDataStoreKey {
	return IndexDataStoreKey{
		CollectionID: collectionID,
		IndexID:      indexID,
		Fields:       fields,
	}
}

// Bytes returns the byte representation of the key
func (k *IndexDataStoreKey) Bytes() []byte {
	return EncodeIndexDataStoreKey(k)
}

// ToDS returns the datastore key
func (k *IndexDataStoreKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

// ToString returns the string representation of the key
// It is in the following format:
// /[CollectionID]/[IndexID]/[FieldValue](/[FieldValue]...)
// If while composing the string from left to right, a component
// is empty, the string is returned up to that point
func (k *IndexDataStoreKey) ToString() string {
	return string(k.Bytes())
}

func (k PrimaryDataStoreKey) ToDataStoreKey() DataStoreKey {
	return DataStoreKey{
		CollectionRootID: k.CollectionRootID,
		DocID:            k.DocID,
	}
}

func (k PrimaryDataStoreKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k PrimaryDataStoreKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

func (k PrimaryDataStoreKey) ToString() string {
	result := ""

	if k.CollectionRootID != 0 {
		result = result + "/" + fmt.Sprint(k.CollectionRootID)
	}
	result = result + PRIMARY_KEY
	if k.DocID != "" {
		result = result + "/" + k.DocID
	}

	return result
}

func (k CollectionKey) ToString() string {
	return fmt.Sprintf("%s/%s", COLLECTION_ID, strconv.Itoa(int(k.CollectionID)))
}

func (k CollectionKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k CollectionKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

func (k CollectionNameKey) ToString() string {
	return fmt.Sprintf("%s/%s", COLLECTION_NAME, k.Name)
}

func (k CollectionNameKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k CollectionNameKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

func (k CollectionSchemaVersionKey) ToString() string {
	result := COLLECTION_SCHEMA_VERSION

	if k.SchemaVersionID != "" {
		result = result + "/" + k.SchemaVersionID
	}

	if k.CollectionID != 0 {
		result = fmt.Sprintf("%s/%s", result, strconv.Itoa(int(k.CollectionID)))
	}

	return result
}

func (k CollectionSchemaVersionKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k CollectionSchemaVersionKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

func (k SchemaVersionKey) ToString() string {
	result := SCHEMA_VERSION

	if k.SchemaVersionID != "" {
		result = result + "/" + k.SchemaVersionID
	}

	return result
}

func (k SchemaVersionKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k SchemaVersionKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

func (k SchemaRootKey) ToString() string {
	result := SCHEMA_VERSION_ROOT

	if k.SchemaRoot != "" {
		result = result + "/" + k.SchemaRoot
	}

	if k.SchemaVersionID != "" {
		result = result + "/" + k.SchemaVersionID
	}

	return result
}

func (k SchemaRootKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k SchemaRootKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

func (k CollectionIDSequenceKey) ToString() string {
	return COLLECTION_SEQ
}

func (k CollectionIDSequenceKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k CollectionIDSequenceKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

func (k IndexIDSequenceKey) ToString() string {
	return INDEX_ID_SEQ + "/" + strconv.Itoa(int(k.CollectionID))
}

func (k IndexIDSequenceKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k IndexIDSequenceKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

func (k FieldIDSequenceKey) ToString() string {
	return FIELD_ID_SEQ + "/" + strconv.Itoa(int(k.CollectionRoot))
}

func (k FieldIDSequenceKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k FieldIDSequenceKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

// New
func NewP2PCollectionKey(collectionID string) P2PCollectionKey {
	return P2PCollectionKey{CollectionID: collectionID}
}

func NewP2PCollectionKeyFromString(key string) (P2PCollectionKey, error) {
	keyArr := strings.Split(key, "/")
	if len(keyArr) != 4 {
		return P2PCollectionKey{}, errors.WithStack(ErrInvalidKey, errors.NewKV("Key", key))
	}
	return NewP2PCollectionKey(keyArr[3]), nil
}

func (k P2PCollectionKey) ToString() string {
	result := P2P_COLLECTION

	if k.CollectionID != "" {
		result = result + "/" + k.CollectionID
	}

	return result
}

func (k P2PCollectionKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k P2PCollectionKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

func NewReplicatorKey(id string) ReplicatorKey {
	return ReplicatorKey{ReplicatorID: id}
}

func (k ReplicatorKey) ToString() string {
	result := REPLICATOR

	if k.ReplicatorID != "" {
		result = result + "/" + k.ReplicatorID
	}

	return result
}

func (k ReplicatorKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k ReplicatorKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

func (k HeadStoreKey) ToString() string {
	var result string

	if k.DocID != "" {
		result = result + "/" + k.DocID
	}
	if k.FieldId != "" {
		result = result + "/" + k.FieldId
	}
	if k.Cid.Defined() {
		result = result + "/" + k.Cid.String()
	}

	return result
}

func (k HeadStoreKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k HeadStoreKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

// PrefixEnd determines the end key given key as a prefix, that is the key that sorts precisely
// behind all keys starting with prefix: "1" is added to the final byte and the carry propagated.
// The special cases of nil and KeyMin always returns KeyMax.
func (k DataStoreKey) PrefixEnd() DataStoreKey {
	newKey := k

	if k.FieldID != "" {
		newKey.FieldID = string(bytesPrefixEnd([]byte(k.FieldID)))
		return newKey
	}
	if k.DocID != "" {
		newKey.DocID = string(bytesPrefixEnd([]byte(k.DocID)))
		return newKey
	}
	if k.InstanceType != "" {
		newKey.InstanceType = InstanceType(bytesPrefixEnd([]byte(k.InstanceType)))
		return newKey
	}
	if k.CollectionRootID != 0 {
		newKey.CollectionRootID = k.CollectionRootID + 1
		return newKey
	}

	return newKey
}

// FieldIDAsUint extracts the Field Identifier from the Key.
// In a Primary index, the last key path is the FieldIDAsUint.
// This may be different in Secondary Indexes.
// An error is returned if it can't correct convert the field to a uint32.
func (k DataStoreKey) FieldIDAsUint() (uint32, error) {
	fieldID, err := strconv.Atoi(k.FieldID)
	if err != nil {
		return 0, NewErrFailedToGetFieldIdOfKey(err)
	}
	return uint32(fieldID), nil
}

func bytesPrefixEnd(b []byte) []byte {
	end := make([]byte, len(b))
	copy(end, b)
	for i := len(end) - 1; i >= 0; i-- {
		end[i] = end[i] + 1
		if end[i] != 0 {
			return end[:i+1]
		}
	}
	// This statement will only be reached if the key is already a
	// maximal byte string (i.e. already \xff...).
	return b
}

// EncStoreDocKey is a key for the encryption store.
type EncStoreDocKey struct {
	DocID   string
	FieldID uint32
}

var _ Key = (*EncStoreDocKey)(nil)

// NewEncStoreDocKey creates a new EncStoreDocKey from a docID and fieldID.
func NewEncStoreDocKey(docID string, fieldID uint32) EncStoreDocKey {
	return EncStoreDocKey{
		DocID:   docID,
		FieldID: fieldID,
	}
}

func (k EncStoreDocKey) ToString() string {
	if k.FieldID == 0 {
		return k.DocID
	}
	return fmt.Sprintf("%s/%d", k.DocID, k.FieldID)
}

func (k EncStoreDocKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k EncStoreDocKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}
