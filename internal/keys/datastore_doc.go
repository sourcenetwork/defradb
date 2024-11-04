// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package keys

import (
	"strconv"

	ds "github.com/ipfs/go-datastore"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/encoding"
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

	DATASTORE_DOC_VERSION_FIELD_ID = "v"
)

// DataStoreKey is a type that represents a key in the database.
type DataStoreKey struct {
	CollectionRootID uint32
	InstanceType     InstanceType
	DocID            string
	FieldID          string
}

var _ Key = (*DataStoreKey)(nil)

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

func (k DataStoreKey) WithCollectionRoot(colRoot uint32) DataStoreKey {
	newKey := k
	newKey.CollectionRootID = colRoot
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

func (k DataStoreKey) WithFieldID(fieldID string) DataStoreKey {
	newKey := k
	newKey.FieldID = fieldID
	return newKey
}

func (k DataStoreKey) ToHeadStoreKey() HeadStoreKey {
	return HeadStoreKey{
		DocID:   k.DocID,
		FieldID: k.FieldID,
	}
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

// DecodeDataStoreKey decodes a store key into a [DataStoreKey].
func DecodeDataStoreKey(data []byte) (DataStoreKey, error) {
	if len(data) == 0 {
		return DataStoreKey{}, ErrEmptyKey
	}

	if data[0] != '/' {
		return DataStoreKey{}, ErrInvalidKey
	}
	data = data[1:]

	data, colRootID, err := encoding.DecodeUvarintAscending(data)
	if err != nil {
		return DataStoreKey{}, err
	}

	var instanceType InstanceType
	if len(data) > 1 {
		if data[0] == '/' {
			data = data[1:]
		}
		instanceType = InstanceType(data[0])
		data = data[1:]
	}

	const docKeyLength int = 40
	var docID string
	if len(data) > docKeyLength {
		if data[0] == '/' {
			data = data[1:]
		}
		docID = string(data[:docKeyLength])
		data = data[docKeyLength:]
	}

	var fieldID string
	if len(data) > 1 {
		if data[0] == '/' {
			data = data[1:]
		}
		// Todo: This should be encoded/decoded properly in
		// https://github.com/sourcenetwork/defradb/issues/2818
		fieldID = string(data)
	}

	return DataStoreKey{
		CollectionRootID: uint32(colRootID),
		InstanceType:     (instanceType),
		DocID:            docID,
		FieldID:          fieldID,
	}, nil
}

// EncodeDataStoreKey encodes a [*DataStoreKey] to a byte array suitable for sorting in the store.
func EncodeDataStoreKey(key *DataStoreKey) []byte {
	var result []byte

	if key.CollectionRootID != 0 {
		result = encoding.EncodeUvarintAscending([]byte{'/'}, uint64(key.CollectionRootID))
	}

	if key.InstanceType != "" {
		result = append(result, '/')
		result = append(result, []byte(string(key.InstanceType))...)
	}

	if key.DocID != "" {
		result = append(result, '/')
		result = append(result, []byte(key.DocID)...)
	}

	if key.FieldID != "" {
		result = append(result, '/')
		// Todo: This should be encoded/decoded properly in
		// https://github.com/sourcenetwork/defradb/issues/2818
		result = append(result, []byte(key.FieldID)...)
	}

	return result
}
