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
	COLLECTION                     = "/collection/id"
	COLLECTION_NAME                = "/collection/name"
	COLLECTION_SCHEMA_VERSION      = "/collection/version"
	COLLECTION_INDEX               = "/collection/index"
	SCHEMA_MIGRATION               = "/schema/migration"
	SCHEMA_VERSION                 = "/schema/version/v"
	SCHEMA_VERSION_HISTORY         = "/schema/version/h"
	SEQ                            = "/seq"
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
	CollectionID string
	InstanceType InstanceType
	DocKey       string
	FieldId      string
}

var _ Key = (*DataStoreKey)(nil)

// IndexDataStoreKey is key of an indexed document in the database.
type IndexDataStoreKey struct {
	// CollectionID is the id of the collection
	CollectionID uint32
	// IndexID is the id of the index
	IndexID uint32
	// FieldValues is the values of the fields in the index
	FieldValues [][]byte
}

var _ Key = (*IndexDataStoreKey)(nil)

type PrimaryDataStoreKey struct {
	CollectionId string
	DocKey       string
}

var _ Key = (*PrimaryDataStoreKey)(nil)

type HeadStoreKey struct {
	DocKey  string
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
	SchemaVersionId string
	CollectionID    uint32
}

var _ Key = (*CollectionSchemaVersionKey)(nil)

// CollectionIndexKey to a stored description of an index
type CollectionIndexKey struct {
	// CollectionName is the name of the collection that the index is on
	CollectionName string
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

// SchemaHistoryKey holds the pathway through the schema version history for
// any given schema.
//
// The key points to the schema version id of the next version of the schema.
// If a SchemaHistoryKey does not exist for a given SchemaVersionID it means
// that that SchemaVersionID is for the latest version.
type SchemaHistoryKey struct {
	SchemaRoot              string
	PreviousSchemaVersionID string
}

var _ Key = (*SchemaHistoryKey)(nil)

// SchemaVersionMigrationKey points to the jsonified configuration of a lens migration
// for the given source schema version id.
type SchemaVersionMigrationKey struct {
	SourceSchemaVersionID string
}

var _ Key = (*SchemaVersionMigrationKey)(nil)

type P2PCollectionKey struct {
	CollectionID string
}

var _ Key = (*P2PCollectionKey)(nil)

type SequenceKey struct {
	SequenceName string
}

var _ Key = (*SequenceKey)(nil)

type ReplicatorKey struct {
	ReplicatorID string
}

var _ Key = (*ReplicatorKey)(nil)

// Creates a new DataStoreKey from a string as best as it can,
// splitting the input using '/' as a field deliminator.  It assumes
// that the input string is in the following format:
//
// /[CollectionId]/[InstanceType]/[DocKey]/[FieldId]
//
// Any properties before the above (assuming a '/' deliminator) are ignored
func NewDataStoreKey(key string) (DataStoreKey, error) {
	dataStoreKey := DataStoreKey{}
	if key == "" {
		return dataStoreKey, ErrEmptyKey
	}

	elements := strings.Split(strings.TrimPrefix(key, "/"), "/")

	numberOfElements := len(elements)

	// With less than 3 or more than 4 elements, we know it's an invalid key
	if numberOfElements < 3 || numberOfElements > 4 {
		return dataStoreKey, ErrInvalidKey
	}

	dataStoreKey.CollectionID = elements[0]
	dataStoreKey.InstanceType = InstanceType(elements[1])
	dataStoreKey.DocKey = elements[2]
	if numberOfElements == 4 {
		dataStoreKey.FieldId = elements[3]
	}

	return dataStoreKey, nil
}

func MustNewDataStoreKey(key string) DataStoreKey {
	dsKey, err := NewDataStoreKey(key)
	if err != nil {
		panic(err)
	}
	return dsKey
}

func DataStoreKeyFromDocKey(dockey client.DocKey) DataStoreKey {
	return DataStoreKey{
		DocKey: dockey.String(),
	}
}

// Creates a new HeadStoreKey from a string as best as it can,
// splitting the input using '/' as a field deliminator.  It assumes
// that the input string is in the following format:
//
// /[DocKey]/[FieldId]/[Cid]
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
		DocKey:  elements[1],
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
		SchemaVersionId: schemaVersionId,
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
		SchemaVersionId: elements[len(elements)-2],
		CollectionID:    uint32(colID),
	}, nil
}

// NewCollectionIndexKey creates a new CollectionIndexKey from a collection name and index name.
func NewCollectionIndexKey(colID, indexName string) CollectionIndexKey {
	return CollectionIndexKey{CollectionName: colID, IndexName: indexName}
}

// NewCollectionIndexKeyFromString creates a new CollectionIndexKey from a string.
// It expects the input string is in the following format:
//
// /collection/index/[CollectionName]/[IndexName]
//
// Where [IndexName] might be omitted. Anything else will return an error.
func NewCollectionIndexKeyFromString(key string) (CollectionIndexKey, error) {
	keyArr := strings.Split(key, "/")
	if len(keyArr) < 4 || len(keyArr) > 5 || keyArr[1] != "collection" || keyArr[2] != "index" {
		return CollectionIndexKey{}, ErrInvalidKey
	}
	result := CollectionIndexKey{CollectionName: keyArr[3]}
	if len(keyArr) == 5 {
		result.IndexName = keyArr[4]
	}
	return result, nil
}

// ToString returns the string representation of the key
// It is in the following format:
// /collection/index/[CollectionName]/[IndexName]
// if [CollectionName] is empty, the rest is ignored
func (k CollectionIndexKey) ToString() string {
	result := COLLECTION_INDEX

	if k.CollectionName != "" {
		result = result + "/" + k.CollectionName
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

func NewSchemaHistoryKey(schemaRoot string, previousSchemaVersionID string) SchemaHistoryKey {
	return SchemaHistoryKey{
		SchemaRoot:              schemaRoot,
		PreviousSchemaVersionID: previousSchemaVersionID,
	}
}

func NewSchemaVersionMigrationKey(schemaVersionID string) SchemaVersionMigrationKey {
	return SchemaVersionMigrationKey{SourceSchemaVersionID: schemaVersionID}
}

func NewSchemaHistoryKeyFromString(keyString string) (SchemaHistoryKey, error) {
	keyString = strings.TrimPrefix(keyString, SCHEMA_VERSION_HISTORY+"/")
	elements := strings.Split(keyString, "/")
	if len(elements) != 2 {
		return SchemaHistoryKey{}, ErrInvalidKey
	}

	return SchemaHistoryKey{
		SchemaRoot:              elements[0],
		PreviousSchemaVersionID: elements[1],
	}, nil
}

func NewSequenceKey(name string) SequenceKey {
	return SequenceKey{SequenceName: name}
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

func (k DataStoreKey) WithDocKey(docKey string) DataStoreKey {
	newKey := k
	newKey.DocKey = docKey
	return newKey
}

func (k DataStoreKey) WithInstanceInfo(key DataStoreKey) DataStoreKey {
	newKey := k
	newKey.DocKey = key.DocKey
	newKey.FieldId = key.FieldId
	newKey.InstanceType = key.InstanceType
	return newKey
}

func (k DataStoreKey) WithFieldId(fieldId string) DataStoreKey {
	newKey := k
	newKey.FieldId = fieldId
	return newKey
}

func (k DataStoreKey) ToHeadStoreKey() HeadStoreKey {
	return HeadStoreKey{
		DocKey:  k.DocKey,
		FieldId: k.FieldId,
	}
}

func (k HeadStoreKey) WithDocKey(docKey string) HeadStoreKey {
	newKey := k
	newKey.DocKey = docKey
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
	var result string

	if k.CollectionID != "" {
		result = result + "/" + k.CollectionID
	}
	if k.InstanceType != "" {
		result = result + "/" + string(k.InstanceType)
	}
	if k.DocKey != "" {
		result = result + "/" + k.DocKey
	}
	if k.FieldId != "" {
		result = result + "/" + k.FieldId
	}

	return result
}

func (k DataStoreKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k DataStoreKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

func (k DataStoreKey) Equal(other DataStoreKey) bool {
	return k.CollectionID == other.CollectionID &&
		k.DocKey == other.DocKey &&
		k.FieldId == other.FieldId &&
		k.InstanceType == other.InstanceType
}

func (k DataStoreKey) ToPrimaryDataStoreKey() PrimaryDataStoreKey {
	return PrimaryDataStoreKey{
		CollectionId: k.CollectionID,
		DocKey:       k.DocKey,
	}
}

// NewIndexDataStoreKey creates a new IndexDataStoreKey from a string.
// It expects the input string is in the following format:
//
// /[CollectionID]/[IndexID]/[FieldValue](/[FieldValue]...)
//
// Where [CollectionID] and [IndexID] are integers
func NewIndexDataStoreKey(key string) (IndexDataStoreKey, error) {
	if key == "" {
		return IndexDataStoreKey{}, ErrEmptyKey
	}

	if !strings.HasPrefix(key, "/") {
		return IndexDataStoreKey{}, ErrInvalidKey
	}

	elements := strings.Split(key[1:], "/")

	// With less than 3 elements, we know it's an invalid key
	if len(elements) < 3 {
		return IndexDataStoreKey{}, ErrInvalidKey
	}

	colID, err := strconv.Atoi(elements[0])
	if err != nil {
		return IndexDataStoreKey{}, ErrInvalidKey
	}

	indexKey := IndexDataStoreKey{CollectionID: uint32(colID)}

	indID, err := strconv.Atoi(elements[1])
	if err != nil {
		return IndexDataStoreKey{}, ErrInvalidKey
	}
	indexKey.IndexID = uint32(indID)

	// first 2 elements are the collection and index IDs, the rest are field values
	for i := 2; i < len(elements); i++ {
		indexKey.FieldValues = append(indexKey.FieldValues, []byte(elements[i]))
	}

	return indexKey, nil
}

// Bytes returns the byte representation of the key
func (k *IndexDataStoreKey) Bytes() []byte {
	return []byte(k.ToString())
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
	sb := strings.Builder{}

	if k.CollectionID == 0 {
		return ""
	}
	sb.WriteByte('/')
	sb.WriteString(strconv.Itoa(int(k.CollectionID)))

	if k.IndexID == 0 {
		return sb.String()
	}
	sb.WriteByte('/')
	sb.WriteString(strconv.Itoa(int(k.IndexID)))

	for _, v := range k.FieldValues {
		if len(v) == 0 {
			break
		}
		sb.WriteByte('/')
		sb.WriteString(string(v))
	}

	return sb.String()
}

// Equal returns true if the two keys are equal
func (k IndexDataStoreKey) Equal(other IndexDataStoreKey) bool {
	if k.CollectionID != other.CollectionID {
		return false
	}
	if k.IndexID != other.IndexID {
		return false
	}
	if len(k.FieldValues) != len(other.FieldValues) {
		return false
	}
	for i := range k.FieldValues {
		if string(k.FieldValues[i]) != string(other.FieldValues[i]) {
			return false
		}
	}
	return true
}

func (k PrimaryDataStoreKey) ToDataStoreKey() DataStoreKey {
	return DataStoreKey{
		CollectionID: k.CollectionId,
		DocKey:       k.DocKey,
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

	if k.CollectionId != "" {
		result = result + "/" + k.CollectionId
	}
	result = result + PRIMARY_KEY
	if k.DocKey != "" {
		result = result + "/" + k.DocKey
	}

	return result
}

func (k CollectionKey) ToString() string {
	return fmt.Sprintf("%s/%s", COLLECTION, strconv.Itoa(int(k.CollectionID)))
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

	if k.SchemaVersionId != "" {
		result = result + "/" + k.SchemaVersionId
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

func (k SchemaHistoryKey) ToString() string {
	result := SCHEMA_VERSION_HISTORY

	if k.SchemaRoot != "" {
		result = result + "/" + k.SchemaRoot
	}

	if k.PreviousSchemaVersionID != "" {
		result = result + "/" + k.PreviousSchemaVersionID
	}

	return result
}

func (k SchemaHistoryKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k SchemaHistoryKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

func (k SchemaVersionMigrationKey) ToString() string {
	result := SCHEMA_MIGRATION

	if k.SourceSchemaVersionID != "" {
		result = result + "/" + k.SourceSchemaVersionID
	}

	return result
}

func (k SchemaVersionMigrationKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k SchemaVersionMigrationKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

func (k SequenceKey) ToString() string {
	result := SEQ

	if k.SequenceName != "" {
		result = result + "/" + k.SequenceName
	}

	return result
}

func (k SequenceKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k SequenceKey) ToDS() ds.Key {
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

	if k.DocKey != "" {
		result = result + "/" + k.DocKey
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

	if k.FieldId != "" {
		newKey.FieldId = string(bytesPrefixEnd([]byte(k.FieldId)))
		return newKey
	}
	if k.DocKey != "" {
		newKey.DocKey = string(bytesPrefixEnd([]byte(k.DocKey)))
		return newKey
	}
	if k.InstanceType != "" {
		newKey.InstanceType = InstanceType(bytesPrefixEnd([]byte(k.InstanceType)))
		return newKey
	}
	if k.CollectionID != "" {
		newKey.CollectionID = string(bytesPrefixEnd([]byte(k.CollectionID)))
		return newKey
	}
	return newKey
}

// FieldID extracts the Field Identifier from the Key.
// In a Primary index, the last key path is the FieldID.
// This may be different in Secondary Indexes.
// An error is returned if it can't correct convert the field to a uint32.
func (k DataStoreKey) FieldID() (uint32, error) {
	fieldID, err := strconv.Atoi(k.FieldId)
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
