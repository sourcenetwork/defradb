// Copyright 2020 Source Inc.
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

	ds "github.com/ipfs/go-datastore"
	"github.com/pkg/errors"
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
	COLLECTION = "collection"
	SCHEMA     = "schema"
	SEQ        = "seq"
)

type Key interface {
	ToString() string
	Bytes() []byte
	ToDS() ds.Key
}

type DataStoreKey struct {
	CollectionId   string
	PrimaryIndexId string
	DocKey         string
	FieldId        string
	InstanceType   InstanceType
}

type HeadStoreKey struct {
	DocKey  string
	FieldId string //can be 'C'
	Cid     string
}

type CollectionKey struct {
	CollectionName string
}

type SchemaKey struct {
	SchemaName string
}

type SequenceKey struct {
	SequenceName string
}

// Creates a new DataStoreKey from a string as best as it can,
// splitting the input using '/' and ':' as field deliminaters.  It assumes
// that the input string is in one of the following formats:
//
// [DocKey]
// [DocKey]:[InstanceType]
// [DocKey]/[FieldId]
// [DocKey]/[FieldId]:[InstanceType]
// Any of the above prefixed but with one of the below:
// [PrimaryIndexId]/
// [CollectionId]/[PrimaryIndexId]/
//
// Any properties before the above (assuming a '/' deliminator) are ignored
func NewDataStoreKey(key string) DataStoreKey {
	dataStoreKey := DataStoreKey{}
	if key == "" {
		return dataStoreKey
	}

	elements := strings.Split(key, "/")
	numberOfElements := len(elements)
	itemComponents := strings.Split(elements[len(elements)-1], ":")
	indexOfDocKey := 0
	var lastItem string

	if len(itemComponents) == 2 {
		dataStoreKey.InstanceType = InstanceType(itemComponents[1])
		if numberOfElements == 1 {
			dataStoreKey.DocKey = itemComponents[0]
			return dataStoreKey
		}
		lastItem = itemComponents[0]
	} else {
		lastItem = elements[numberOfElements-1]
	}

	// if the 2nd-to-last is longer than the last, then the last must be the field id
	if numberOfElements > 1 && len(lastItem) < len(elements[numberOfElements-2]) {
		dataStoreKey.FieldId = lastItem
		indexOfDocKey = numberOfElements - 2
	} else {
		indexOfDocKey = numberOfElements - 1
	}
	dataStoreKey.DocKey = elements[indexOfDocKey]

	if numberOfElements == 1 {
		return dataStoreKey
	}

	if indexOfDocKey-1 < 0 {
		return dataStoreKey
	}
	dataStoreKey.PrimaryIndexId = elements[indexOfDocKey-1]

	if indexOfDocKey-2 < 0 {
		return dataStoreKey
	}
	dataStoreKey.CollectionId = elements[indexOfDocKey-2]

	return dataStoreKey
}

// Creates a new HeadStoreKey from a string as best as it can,
// splitting the input using '/' as a field deliminater.  It assumes
// that the input string is in the following format:
//
// /[DocKey]/[FieldId]/[Cid]
//
// Any properties before the above are ignored
func NewHeadStoreKey(key string) HeadStoreKey {
	elements := strings.Split(key, "/")

	return HeadStoreKey{
		// elements[0] is empty (key has leading '/')
		DocKey:  elements[1],
		FieldId: elements[2],
		Cid:     elements[3],
	}
}

// Returns a formatted collection key for the system data store.
// it assumes the name of the collection is non-empty.
func NewCollectionKey(name string) CollectionKey {
	return CollectionKey{CollectionName: name}
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

func (k HeadStoreKey) WithCid(cid string) HeadStoreKey {
	newKey := k
	newKey.Cid = cid
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
	if k.PrimaryIndexId != "" {
		result = result + "/" + k.PrimaryIndexId
	}
	if k.DocKey != "" {
		result = result + "/" + k.DocKey
	}
	if k.FieldId != "" {
		result = result + "/" + k.FieldId
	}
	if k.InstanceType != "" {
		result = result + ":" + string(k.InstanceType)
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
		k.PrimaryIndexId == other.PrimaryIndexId &&
		k.DocKey == other.DocKey &&
		k.FieldId == other.FieldId &&
		k.InstanceType == other.InstanceType
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
	if k.Cid != "" {
		result = result + "/" + k.Cid
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

	if len(k.DocKey) == 0 {
		return newKey
	}

	newKey.DocKey = string(bytesPrefixEnd([]byte(k.DocKey)))
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
		return 0, errors.Wrap(err, "Failed to get FieldID of Key")
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
