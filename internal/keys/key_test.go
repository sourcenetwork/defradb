// Copyright 2022 Democratized Data Foundation
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
	"fmt"
	"testing"

	ds "github.com/ipfs/go-datastore"
	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/encoding"
)

func TestNewDataStoreKey_ReturnsEmptyStruct_GivenEmptyString(t *testing.T) {
	inputString := ""

	result, err := NewDataStoreKey(inputString)

	resultString := result.ToString()

	assert.Equal(t, DataStoreKey{}, result)
	assert.Equal(t, "", resultString)
	assert.ErrorIs(t, ErrEmptyKey, err)
}

func encodePrefix(colID, indexID uint32) []byte {
	return encoding.EncodeUvarintAscending(append(encoding.EncodeUvarintAscending(
		[]byte{'/'}, uint64(colID)), '/'), uint64(indexID))
}

func encodeKey(colID, indexID uint32, fieldParts ...any) []byte {
	b := encodePrefix(colID, indexID)
	const partSize = 2
	if len(fieldParts)%partSize != 0 {
		panic(fmt.Sprintf("fieldParts must be a multiple of %d: value, descending", partSize))
	}
	for i := 0; i < len(fieldParts)/partSize; i++ {
		b = append(b, '/')
		isDescending, _ := fieldParts[i*partSize+1].(bool)
		if fieldParts[i*partSize] == nil {
			if isDescending {
				b = encoding.EncodeNullDescending(b)
			} else {
				b = encoding.EncodeNullAscending(b)
			}
		} else {
			v, _ := fieldParts[i*partSize].(int)
			if isDescending {
				b = encoding.EncodeUvarintDescending(b, uint64(v))
			} else {
				b = encoding.EncodeUvarintAscending(b, uint64(v))
			}
		}
	}
	return b
}

func TestIndexDatastoreKey_Bytes(t *testing.T) {
	cases := []struct {
		Name         string
		CollectionID uint32
		IndexID      uint32
		Fields       []IndexedField
		Expected     []byte
	}{
		{
			Name:     "empty",
			Expected: []byte{},
		},
		{
			Name:         "only collection",
			CollectionID: 1,
			Expected:     encoding.EncodeUvarintAscending([]byte{'/'}, 1),
		},
		{
			Name:         "only collection and index",
			CollectionID: 1,
			IndexID:      2,
			Expected:     encodePrefix(1, 2),
		},
		{
			Name:         "collection, index and one field",
			CollectionID: 1,
			IndexID:      2,
			Fields:       []IndexedField{{Value: client.NewNormalInt(5)}},
			Expected:     encodeKey(1, 2, 5, false),
		},
		{
			Name:         "collection, index and two fields",
			CollectionID: 1,
			IndexID:      2,
			Fields:       []IndexedField{{Value: client.NewNormalInt(5)}, {Value: client.NewNormalInt(7)}},
			Expected:     encodeKey(1, 2, 5, false, 7, false),
		},
		{
			Name:         "no index",
			CollectionID: 1,
			Fields:       []IndexedField{{Value: client.NewNormalInt(5)}},
			Expected:     encoding.EncodeUvarintAscending([]byte{'/'}, 1),
		},
		{
			Name:     "no collection",
			IndexID:  2,
			Fields:   []IndexedField{{Value: client.NewNormalInt(5)}},
			Expected: []byte{},
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			key := NewIndexDataStoreKey(c.CollectionID, c.IndexID, c.Fields)
			actual := key.Bytes()
			assert.Equal(t, c.Expected, actual, "upon calling key.Bytes()")
			encKey := EncodeIndexDataStoreKey(&key)
			assert.Equal(t, c.Expected, encKey, "upon calling EncodeIndexDataStoreKey")
		})
	}
}

func TestIndexDatastoreKey_ToString(t *testing.T) {
	key := NewIndexDataStoreKey(1, 2, []IndexedField{{Value: client.NewNormalInt(5)}})
	assert.Equal(t, key.ToString(), string(encodeKey(1, 2, 5, false)))
}

func TestIndexDatastoreKey_ToDS(t *testing.T) {
	key := NewIndexDataStoreKey(1, 2, []IndexedField{{Value: client.NewNormalInt(5)}})
	assert.Equal(t, key.ToDS(), ds.NewKey(string(encodeKey(1, 2, 5, false))))
}

func TestDecodeIndexDataStoreKey(t *testing.T) {
	const colID, indexID = 1, 2
	cases := []struct {
		name           string
		desc           client.IndexDescription
		inputBytes     []byte
		expectedFields []IndexedField
		fieldKinds     []client.FieldKind
	}{
		{
			name: "one field",
			desc: client.IndexDescription{
				ID:     indexID,
				Fields: []client.IndexedFieldDescription{{}},
			},
			inputBytes:     encodeKey(colID, indexID, 5, false),
			expectedFields: []IndexedField{{Value: client.NewNormalInt(5)}},
		},
		{
			name: "two fields (one descending)",
			desc: client.IndexDescription{
				ID:     indexID,
				Fields: []client.IndexedFieldDescription{{}, {Descending: true}},
			},
			inputBytes: encodeKey(colID, indexID, 5, false, 7, true),
			expectedFields: []IndexedField{
				{Value: client.NewNormalInt(5)},
				{Value: client.NewNormalInt(7), Descending: true},
			},
		},
		{
			name: "last encoded value without matching field description is docID",
			desc: client.IndexDescription{
				ID:     indexID,
				Fields: []client.IndexedFieldDescription{{}},
			},
			inputBytes: encoding.EncodeStringAscending(append(encodeKey(1, indexID, 5, false), '/'), "docID"),
			expectedFields: []IndexedField{
				{Value: client.NewNormalInt(5)},
				{Value: client.NewNormalString("docID")},
			},
			fieldKinds: []client.FieldKind{client.FieldKind_NILLABLE_INT},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expectedKey := NewIndexDataStoreKey(colID, indexID, tc.expectedFields)
			fieldDescs := make([]client.FieldDefinition, len(tc.desc.Fields))
			for i := range tc.fieldKinds {
				fieldDescs[i] = client.FieldDefinition{Kind: tc.fieldKinds[i]}
			}
			key, err := DecodeIndexDataStoreKey(tc.inputBytes, &tc.desc, fieldDescs)
			assert.NoError(t, err)
			assert.Equal(t, expectedKey, key)
		})
	}
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

	const colID, indexID = 1, 2

	cases := []struct {
		name      string
		val       []byte
		numFields int
	}{
		{
			name: "empty",
			val:  []byte{},
		},
		{
			name: "only slash",
			val:  []byte{'/'},
		},
		{
			name: "slash after collection",
			val:  append(encoding.EncodeUvarintAscending([]byte{'/'}, colID), '/'),
		},
		{
			name:      "wrong prefix",
			val:       replace(encodeKey(colID, indexID, 5, false), 0, ' '),
			numFields: 1,
		},
		{
			name:      "no slash before collection",
			val:       encodeKey(colID, indexID, 5, false)[1:],
			numFields: 1,
		},
		{
			name:      "no slash before index",
			val:       replace(encodeKey(colID, indexID, 5, false), 2, ' '),
			numFields: 1,
		},
		{
			name:      "no slash before field value",
			val:       replace(encodeKey(colID, indexID, 5, false), 4, ' '),
			numFields: 1,
		},
		{
			name:      "no field value",
			val:       cutEnd(encodeKey(colID, indexID, 5, false), 1),
			numFields: 1,
		},
		{
			name:      "no field description",
			val:       encodeKey(colID, indexID, 5, false, 7, false, 9, false),
			numFields: 2,
		},
	}
	indexDesc := client.IndexDescription{ID: indexID, Fields: []client.IndexedFieldDescription{{}}}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fieldDescs := make([]client.FieldDefinition, c.numFields)
			for i := 0; i < c.numFields; i++ {
				fieldDescs[i] = client.FieldDefinition{Kind: client.FieldKind_NILLABLE_INT}
			}
			_, err := DecodeIndexDataStoreKey(c.val, &indexDesc, fieldDescs)
			assert.Error(t, err, c.name)
		})
	}
}

func TestIndexDataStoreKey_IsEqual(t *testing.T) {
	const colID, indexID = 1, 2

	cases := []struct {
		name        string
		key1        IndexDataStoreKey
		key2        IndexDataStoreKey
		shouldMatch bool
	}{
		{
			name:        "empty",
			key1:        IndexDataStoreKey{},
			key2:        IndexDataStoreKey{},
			shouldMatch: true,
		},
		{
			name:        "same",
			key1:        NewIndexDataStoreKey(colID, indexID, []IndexedField{{Value: client.NewNormalInt(5)}}),
			key2:        NewIndexDataStoreKey(colID, indexID, []IndexedField{{Value: client.NewNormalInt(5)}}),
			shouldMatch: true,
		},
		{
			name:        "different collection",
			key1:        NewIndexDataStoreKey(colID, indexID, []IndexedField{{Value: client.NewNormalInt(5)}}),
			key2:        NewIndexDataStoreKey(colID+1, indexID, []IndexedField{{Value: client.NewNormalInt(5)}}),
			shouldMatch: false,
		},
		{
			name:        "different index",
			key1:        NewIndexDataStoreKey(colID, indexID, []IndexedField{{Value: client.NewNormalInt(5)}}),
			key2:        NewIndexDataStoreKey(colID, indexID+1, []IndexedField{{Value: client.NewNormalInt(5)}}),
			shouldMatch: false,
		},
		{
			name:        "different field",
			key1:        NewIndexDataStoreKey(colID, indexID, []IndexedField{{Value: client.NewNormalInt(5)}}),
			key2:        NewIndexDataStoreKey(colID, indexID, []IndexedField{{Value: client.NewNormalInt(6)}}),
			shouldMatch: false,
		},
		{
			name: "different field count",
			key1: NewIndexDataStoreKey(colID, indexID, []IndexedField{{Value: client.NewNormalInt(5)}}),
			key2: NewIndexDataStoreKey(colID, indexID,
				[]IndexedField{{Value: client.NewNormalInt(5)}, {Value: client.NewNormalInt(6)}}),
			shouldMatch: false,
		},
		{
			name:        "different field type",
			key1:        NewIndexDataStoreKey(colID, indexID, []IndexedField{{Value: client.NewNormalInt(5)}}),
			key2:        NewIndexDataStoreKey(colID, indexID, []IndexedField{{Value: client.NewNormalString("5")}}),
			shouldMatch: false,
		},
		{
			name: "different field descending",
			key1: NewIndexDataStoreKey(colID, indexID, []IndexedField{{Value: client.NewNormalInt(5)}}),
			key2: NewIndexDataStoreKey(colID, indexID,
				[]IndexedField{{Value: client.NewNormalInt(5), Descending: true}}),
			shouldMatch: false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res := c.key1.Equal(c.key2)
			assert.Equal(t, res, c.shouldMatch, c.name)
		})
	}
}
