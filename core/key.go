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
)

var (
	// KeyMin is a minimum key value which sorts before all other keys.
	KeyMin = []byte{}
	// KeyMax is a maximum key value which sorts after all other keys.
	KeyMax = []byte{0xff, 0xff}
)

type InstanceType string

const (
	ValueKey    = InstanceType("v")
	PriorityKey = InstanceType("p")
)

const (
	COLLECTION        = "collection/names"
	COLLECTION_SCHEMA = "/collection/schema"
	SCHEMA            = "schema"
	SEQ               = "seq"
	PRIMARY_KEY       = "pk"
)

type Key interface {
	ToString() string
	Bytes() []byte
	ToDS() ds.Key
}

type DataStoreKey struct {
	CollectionId string
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

type CollectionKey struct {
	CollectionName string
}

var _ Key = (*CollectionKey)(nil)

type CollectionSchemaKey struct {
	SchemaId string
}

var _ Key = (*CollectionSchemaKey)(nil)

type SchemaKey struct {
	SchemaName string
}

var _ Key = (*SchemaKey)(nil)

type SequenceKey struct {
	SequenceName string
}

var _ Key = (*SequenceKey)(nil)

// Creates a new DataStoreKey from a string as best as it can,
// splitting the input using '/' as a field deliminater.  It assumes
// that the input string is in the following format:
//
// /[CollectionId]/[InstanceType]/[DocKey]/[FieldId]
//
// Any properties before the above (assuming a '/' deliminator) are ignored
func NewDataStoreKey(key string) DataStoreKey {
	dataStoreKey := DataStoreKey{}
	if key == "" {
		return dataStoreKey
	}

	elements := strings.Split(key, "/")
	numberOfElements := len(elements)

	return DataStoreKey{
		CollectionId: elements[numberOfElements-4],
		InstanceType: InstanceType(elements[numberOfElements-3]),
		DocKey:       elements[numberOfElements-2],
		FieldId:      elements[numberOfElements-1],
	}
}

func DataStoreKeyFromDocKey(dockey client.DocKey) DataStoreKey {
	return DataStoreKey{
		DocKey: dockey.String(),
	}
}

// Creates a new HeadStoreKey from a string as best as it can,
// splitting the input using '/' as a field deliminater.  It assumes
// that the input string is in the following format:
//
// /[DocKey]/[FieldId]/[Cid]
//
// Any properties before the above are ignored
func NewHeadStoreKey(key string) (HeadStoreKey, error) {
	elements := strings.Split(key, "/")
	if len(elements) != 4 {
		return HeadStoreKey{}, fmt.Errorf("Given headstore key string is not in expected format: %s", key)
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
// it assumes the name of the collection is non-empty.
func NewCollectionKey(name string) CollectionKey {
	return CollectionKey{CollectionName: name}
}

func NewCollectionSchemaKey(schemaId string) CollectionSchemaKey {
	return CollectionSchemaKey{SchemaId: schemaId}
}

// NewSchemaKey returns a formatted schema key for the system data store.
// it assumes the name of the schema is non-empty.
func NewSchemaKey(name string) SchemaKey {
	return SchemaKey{SchemaName: name}
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

	if k.CollectionId != "" {
		result = result + "/" + k.CollectionId
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
	return k.CollectionId == other.CollectionId &&
		k.DocKey == other.DocKey &&
		k.FieldId == other.FieldId &&
		k.InstanceType == other.InstanceType
}

func (k PrimaryDataStoreKey) ToDataStoreKey() DataStoreKey {
	return DataStoreKey{
		CollectionId: k.CollectionId,
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
	result = result + "/" + PRIMARY_KEY
	if k.DocKey != "" {
		result = result + "/" + k.DocKey
	}

	return result
}

func (k CollectionKey) ToString() string {
	result := "/" + COLLECTION

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
	result := "/" + COLLECTION_SCHEMA

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

func (k SchemaKey) ToString() string {
	result := "/" + SCHEMA

	if k.SchemaName != "" {
		result = result + "/" + k.SchemaName
	}

	return result
}

func (k SchemaKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k SchemaKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

func (k SequenceKey) ToString() string {
	result := "/" + SEQ

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

// PrefixEnd determines the end key given key as a prefix, that is the
// key that sorts precisely behind all keys starting with prefix: "1"
// is added to the final byte and the carry propagated. The special
// cases of nil and KeyMin always returns KeyMax.
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
	if k.CollectionId != "" {
		newKey.CollectionId = string(bytesPrefixEnd([]byte(k.CollectionId)))
		return newKey
	}
	return newKey
}

// FieldID extracts the Field Identifier from the Key.
// In a Primary index, the last key path is the FieldID.
// This may be different in Secondary Indexes.
// An error is returned if it can't correct convert the
// field to a uint32.
func (k DataStoreKey) FieldID() (uint32, error) {
	fieldID, err := strconv.Atoi(k.FieldId)
	if err != nil {
		return 0, fmt.Errorf("Failed to get FieldID of Key: %w", err)
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
