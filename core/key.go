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
	COLLECTION                = "/collection/names"
	COLLECTION_SCHEMA         = "/collection/schema"
	COLLECTION_SCHEMA_VERSION = "/collection/version"
	SEQ                       = "/seq"
	PRIMARY_KEY               = "/pk"
	REPLICATOR                = "/replicator/id"
	P2P_COLLECTION            = "/p2p/collection"
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

// CollectionKey points to the current/'head' SchemaVersionId for
// the collection of the given name.
type CollectionKey struct {
	CollectionName string
}

var _ Key = (*CollectionKey)(nil)

// CollectionSchemaKey points to the current/'head' SchemaVersionId for
// the collection of the given schema id.
type CollectionSchemaKey struct {
	SchemaId string
}

var _ Key = (*CollectionSchemaKey)(nil)

// CollectionSchemaVersionKey points to schema of a collection at a given
// version.
type CollectionSchemaVersionKey struct {
	SchemaVersionId string
}

var _ Key = (*CollectionSchemaVersionKey)(nil)

// SchemaHistoryKey holds the pathway through the schema version history for
// any given schema.
//
// The keys point to the schema version id of the next version of the schema.
// If a SchemaHistoryKey does not exist for a given SchemaVersionID it means
// that that SchemaVersionID is for the latest version.
type SchemaHistoryKey struct {
	SchemaID             string
	PriorSchemaVersionId string
}

var _ Key = (*SchemaHistoryKey)(nil)

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
func NewCollectionKey(name string) CollectionKey {
	return CollectionKey{CollectionName: name}
}

func NewCollectionSchemaKey(schemaId string) CollectionSchemaKey {
	return CollectionSchemaKey{SchemaId: schemaId}
}

func NewCollectionSchemaVersionKey(schemaVersionId string) CollectionSchemaVersionKey {
	return CollectionSchemaVersionKey{SchemaVersionId: schemaVersionId}
}

func NewSchemaHistoryKey(schemaId string, priorSchemaVersionId string) SchemaHistoryKey {
	return SchemaHistoryKey{
		SchemaID:             schemaId,
		PriorSchemaVersionId: priorSchemaVersionId,
	}
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
	result := COLLECTION

	if k.CollectionName != "" {
		result = result + "/" + k.CollectionName
	}

	return result
}

func (k CollectionKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k CollectionKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

func (k CollectionSchemaKey) ToString() string {
	result := COLLECTION_SCHEMA

	if k.SchemaId != "" {
		result = result + "/" + k.SchemaId
	}

	return result
}

func (k CollectionSchemaKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k CollectionSchemaKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

func (k CollectionSchemaVersionKey) ToString() string {
	result := COLLECTION_SCHEMA_VERSION + "/" + string(ValueKey)

	if k.SchemaVersionId != "" {
		result = result + "/" + k.SchemaVersionId
	}

	return result
}

func (k CollectionSchemaVersionKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k CollectionSchemaVersionKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

func (k SchemaHistoryKey) ToString() string {
	result := COLLECTION_SCHEMA_VERSION + "/" + string(PriorityKey)

	if k.SchemaID != "" {
		result = result + "/" + k.SchemaID
	}

	if k.PriorSchemaVersionId != "" {
		result = result + "/" + k.PriorSchemaVersionId
	}

	return result
}

func (k SchemaHistoryKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k SchemaHistoryKey) ToDS() ds.Key {
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
