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
	"testing"

	ds "github.com/ipfs/go-datastore"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/db/encoding"
	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/assert"
)

func TestNewDataStoreKey_ReturnsEmptyStruct_GivenEmptyString(t *testing.T) {
	inputString := ""

	result, err := NewDataStoreKey(inputString)

	resultString := result.ToString()

	assert.Equal(t, DataStoreKey{}, result)
	assert.Equal(t, "", resultString)
	assert.ErrorIs(t, ErrEmptyKey, err)
}

func TestNewDataStoreKey_ReturnsCollectionIdAndIndexIdAndDocIDAndFieldIdAndInstanceType_GivenFourItemsWithType(
	t *testing.T,
) {
	instanceType := "anyType"
	fieldID := "f1"
	docID := "docID"
	var collectionRootID uint32 = 2
	inputString := fmt.Sprintf("%v/%s/%s/%s", collectionRootID, instanceType, docID, fieldID)

	result, err := NewDataStoreKey(inputString)
	if err != nil {
		t.Error(err)
	}
	resultString := result.ToString()

	assert.Equal(
		t,
		DataStoreKey{
			CollectionRootID: collectionRootID,
			DocID:            docID,
			FieldId:          fieldID,
			InstanceType:     InstanceType(instanceType)},
		result)
	assert.Equal(t, fmt.Sprintf("/%v/%s/%s/%s", collectionRootID, instanceType, docID, fieldID), resultString)
}

func TestNewDataStoreKey_ReturnsEmptyStruct_GivenAStringWithMissingElements(t *testing.T) {
	inputString := "/0/v"

	_, err := NewDataStoreKey(inputString)

	assert.ErrorIs(t, ErrInvalidKey, err)
}

func TestNewDataStoreKey_GivenAShortObjectMarker(t *testing.T) {
	instanceType := "anyType"
	docID := "docID"
	var collectionRootID uint32 = 2
	inputString := fmt.Sprintf("%v/%s/%s", collectionRootID, instanceType, docID)

	result, err := NewDataStoreKey(inputString)
	if err != nil {
		t.Error(err)
	}
	resultString := result.ToString()

	assert.Equal(
		t,
		DataStoreKey{
			CollectionRootID: collectionRootID,
			DocID:            docID,
			InstanceType:     InstanceType(instanceType)},
		result)
	assert.Equal(t, fmt.Sprintf("/%v/%s/%s", collectionRootID, instanceType, docID), resultString)
}

func TestNewDataStoreKey_GivenAStringWithExtraPrefixes(t *testing.T) {
	instanceType := "anyType"
	fieldId := "f1"
	docID := "docID"
	collectionId := "1"
	inputString := "/db/my_database_name/data/" + collectionId + "/" + instanceType + "/" + docID + "/" + fieldId

	_, err := NewDataStoreKey(inputString)

	assert.ErrorIs(t, ErrInvalidKey, err)
}

func TestNewDataStoreKey_GivenAStringWithExtraSuffix(t *testing.T) {
	instanceType := "anyType"
	fieldId := "f1"
	docID := "docID"
	collectionId := "1"
	inputString := "/db/data/" + collectionId + "/" + instanceType + "/" + docID + "/" + fieldId + "/version_number"

	_, err := NewDataStoreKey(inputString)

	assert.ErrorIs(t, ErrInvalidKey, err)
}

func TestNewIndexKey_IfEmptyParam_ReturnPrefix(t *testing.T) {
	key := NewCollectionIndexKey(immutable.None[uint32](), "")
	assert.Equal(t, "/collection/index", key.ToString())
}

func TestNewIndexKey_IfParamsAreGiven_ReturnFullKey(t *testing.T) {
	key := NewCollectionIndexKey(immutable.Some[uint32](1), "idx")
	assert.Equal(t, "/collection/index/1/idx", key.ToString())
}

func TestNewIndexKey_InNoCollectionName_ReturnJustPrefix(t *testing.T) {
	key := NewCollectionIndexKey(immutable.None[uint32](), "idx")
	assert.Equal(t, "/collection/index", key.ToString())
}

func TestNewIndexKey_InNoIndexName_ReturnWithoutIndexName(t *testing.T) {
	key := NewCollectionIndexKey(immutable.Some[uint32](1), "")
	assert.Equal(t, "/collection/index/1", key.ToString())
}

func TestNewIndexKeyFromString_IfInvalidString_ReturnError(t *testing.T) {
	for _, key := range []string{
		"",
		"/collection",
		"/collection/index",
		"/collection/index/col/idx/extra",
		"/wrong/index/col/idx",
		"/collection/wrong/col/idx",
	} {
		_, err := NewCollectionIndexKeyFromString(key)
		assert.ErrorIs(t, err, ErrInvalidKey)
	}
}

func TestNewIndexKeyFromString_IfOnlyCollectionName_ReturnKey(t *testing.T) {
	key, err := NewCollectionIndexKeyFromString("/collection/index/1")
	assert.NoError(t, err)
	assert.Equal(t, immutable.Some[uint32](1), key.CollectionID)
	assert.Equal(t, "", key.IndexName)
}

func TestNewIndexKeyFromString_IfFullKeyString_ReturnKey(t *testing.T) {
	key, err := NewCollectionIndexKeyFromString("/collection/index/1/idx")
	assert.NoError(t, err)
	assert.Equal(t, immutable.Some[uint32](1), key.CollectionID)
	assert.Equal(t, "idx", key.IndexName)
}

func encodePrefix(colID, indexID uint32) []byte {
	return encoding.EncodeUvarintAscending(append(encoding.EncodeUvarintAscending(
		[]byte{'/'}, uint64(colID)), '/'), uint64(indexID))
}

func encodeKey(colID, indexID uint32, fieldParts ...any) []byte {
	b := encodePrefix(colID, indexID)
	const partSize = 4
	if len(fieldParts)%partSize != 0 {
		panic(fmt.Sprintf("fieldParts must be a multiple of %d: id, kind, value, descending", partSize))
	}
	for i := 0; i < len(fieldParts)/partSize; i++ {
		b = append(b, '/')
		b = encoding.EncodeUvarintAscending(b, uint64(fieldParts[i*partSize].(int)))
		b = encoding.EncodeUvarintAscending(b, uint64(fieldParts[i*partSize+1].(client.FieldKind)))
		isDescending := fieldParts[i*partSize+3].(bool)
		if fieldParts[i*partSize+2] == nil {
			if isDescending {
				b = encoding.EncodeNullDescending(b)
			} else {
				b = encoding.EncodeNullAscending(b)
			}
		} else {
			if isDescending {
				b = encoding.EncodeUvarintDescending(b, uint64(fieldParts[i*partSize+2].(int)))
			} else {
				b = encoding.EncodeUvarintAscending(b, uint64(fieldParts[i*partSize+2].(int)))
			}
		}
	}
	return b
}

const intKind = client.FieldKind_NILLABLE_INT

func TestIndexDatastoreKey_Bytes(t *testing.T) {
	cases := []struct {
		Name        string
		Key         IndexDataStoreKey
		Expected    []byte
		ExpectError bool
	}{
		{
			Name:     "empty",
			Key:      IndexDataStoreKey{},
			Expected: []byte{},
		},
		{
			Name: "only collection",
			Key: IndexDataStoreKey{
				CollectionID: 1,
			},
			Expected: encoding.EncodeUvarintAscending([]byte{'/'}, 1),
		},
		{
			Name: "only collection and index",
			Key: IndexDataStoreKey{
				CollectionID: 1,
				IndexID:      2,
			},
			Expected: encodePrefix(1, 2),
		},
		{
			Name: "collection, index and one field",
			Key: IndexDataStoreKey{
				CollectionID: 1,
				IndexID:      2,
				Fields: []IndexedField{
					{ID: 3, Value: client.NewFieldValue(client.NONE_CRDT, 5, intKind)},
				},
			},
			Expected: encodeKey(1, 2, 3, intKind, 5, false),
		},
		{
			Name: "collection, index and two fields",
			Key: IndexDataStoreKey{
				CollectionID: 1,
				IndexID:      2,
				Fields: []IndexedField{
					{ID: 3, Value: client.NewFieldValue(client.NONE_CRDT, 5, intKind)},
					{ID: 6, Value: client.NewFieldValue(client.NONE_CRDT, 7, intKind)},
				},
			},
			Expected: encodeKey(1, 2, 3, intKind, 5, false, 6, intKind, 7, false),
		},
		{
			Name: "no index",
			Key: IndexDataStoreKey{
				CollectionID: 1,
				Fields: []IndexedField{
					{ID: 3, Value: client.NewFieldValue(client.NONE_CRDT, 5, intKind)},
				},
			},
			Expected: encoding.EncodeUvarintAscending([]byte{'/'}, 1),
		},
		{
			Name: "no collection",
			Key: IndexDataStoreKey{
				IndexID: 2,
				Fields: []IndexedField{
					{ID: 3, Value: client.NewFieldValue(client.NONE_CRDT, 5, intKind)},
				},
			},
			Expected: []byte{},
		},
		{
			Name: "wrong field value",
			Key: IndexDataStoreKey{
				CollectionID: 1,
				IndexID:      2,
				Fields: []IndexedField{
					{ID: 3, Value: client.NewFieldValue(client.NONE_CRDT, "string", intKind)},
				},
			},
			Expected:    nil,
			ExpectError: true,
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			actual := c.Key.Bytes()
			assert.Equal(t, c.Expected, actual, c.Name+": key.Bytes()")
			encKey, err := EncodeIndexDataStoreKey([]byte{}, &c.Key)
			if c.ExpectError {
				assert.Error(t, err, c.Name)
			} else {
				assert.NoError(t, err, c.Name)
			}
			assert.Equal(t, c.Expected, encKey, c.Name+": EncodeIndexDataStoreKey")
		})
	}
}

func TestIndexDatastoreKey_ToString(t *testing.T) {
	key := IndexDataStoreKey{
		CollectionID: 1,
		IndexID:      2,
		Fields: []IndexedField{
			{ID: 3, Value: client.NewFieldValue(client.NONE_CRDT, 5, intKind)},
		},
	}
	assert.Equal(t, key.ToString(), string(encodeKey(1, 2, 3, intKind, 5, false)))
}

func TestIndexDatastoreKey_ToDS(t *testing.T) {
	key := IndexDataStoreKey{
		CollectionID: 1,
		IndexID:      2,
		Fields: []IndexedField{
			{ID: 3, Value: client.NewFieldValue(client.NONE_CRDT, 5, intKind)},
		},
	}
	assert.Equal(t, key.ToDS(), ds.NewKey(string(encodeKey(1, 2, 3, intKind, 5, false))))
}

func TestCollectionIndexKey_Bytes(t *testing.T) {
	key := CollectionIndexKey{
		CollectionID: immutable.Some[uint32](1),
		IndexName:    "idx",
	}
	assert.Equal(t, []byte(COLLECTION_INDEX+"/1/idx"), key.Bytes())
}

func TestDecodeIndexDataStoreKey_ValidKey(t *testing.T) {
	indexDesc := client.IndexDescription{ID: 2, Fields: []client.IndexedFieldDescription{{}}}
	key, err := DecodeIndexDataStoreKey(encodeKey(1, 2, 3, intKind, 5, false), &indexDesc)
	assert.NoError(t, err)
	assert.Equal(t, IndexDataStoreKey{
		CollectionID: 1,
		IndexID:      2,
		Fields: []IndexedField{
			{ID: 3, Value: client.NewFieldValue(client.NONE_CRDT, int64(5), intKind)},
		},
	}, key)

	indexDesc = client.IndexDescription{ID: 2, Fields: []client.IndexedFieldDescription{{}, {}}}
	key, err = DecodeIndexDataStoreKey(encodeKey(1, 2, 3, intKind, 5, false, 6, intKind, 7, false), &indexDesc)
	assert.NoError(t, err)
	assert.Equal(t, IndexDataStoreKey{
		CollectionID: 1,
		IndexID:      2,
		Fields: []IndexedField{
			{ID: 3, Value: client.NewFieldValue(client.NONE_CRDT, int64(5), intKind)},
			{ID: 6, Value: client.NewFieldValue(client.NONE_CRDT, int64(7), intKind)},
		},
	}, key)
}

func TestDecodeIndexDataStoreKey_InvalidKey(t *testing.T) {
	replace := func(b []byte, i int, v byte) []byte {
		b = append([]byte{}, b...)
		b[i] = v
		return b
	}
	cutEnd := func(b []byte, l int) []byte {
		return b[:len(b)-l]
	}

	cases := []struct {
		name string
		val  []byte
	}{
		{name: "empty", val: []byte{}},
		{name: "only slash", val: []byte{'/'}},
		{name: "slash after collection", val: append(encoding.EncodeUvarintAscending([]byte{'/'}, 1), '/')},
		{name: "wrong prefix", val: replace(encodeKey(1, 2, 3, intKind, 5, false), 0, ' ')},
		{name: "no slash before collection", val: encodeKey(1, 2, 3, intKind, 5, false)[1:]},
		{name: "no slash before index", val: replace(encodeKey(1, 2, 3, intKind, 5, false), 2, ' ')},
		{name: "no slash before field", val: replace(encodeKey(1, 2, 3, intKind, 5, false), 4, ' ')},
		{name: "no field id", val: append(encodePrefix(1, 2), '/')},
		{name: "no field kind", val: append(encodePrefix(1, 2), '/', encoding.EncodeUvarintAscending(nil, 3)[0])},
		{name: "no field value", val: cutEnd(encodeKey(1, 2, 3, intKind, 5, false), 1)},
	}
	indexDesc := client.IndexDescription{ID: 2, Fields: []client.IndexedFieldDescription{{}}}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := DecodeIndexDataStoreKey(c.val, &indexDesc)
			assert.Error(t, err, c.name)
		})
	}
}
