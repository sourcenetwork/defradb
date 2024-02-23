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

func TestNewPolicyKey_GivenParam_ReturnKey(t *testing.T) {
	key := NewCollectionPolicyKey(1)
	assert.Equal(t, "/collection/policy/1", key.ToString())
}

func TestNewPolicyKeyFromString_GivenValidString_ReturnKey(t *testing.T) {
	key, err := NewCollectionPolicyKeyFromString("/collection/policy/1")
	assert.NoError(t, err)
	assert.Equal(t, uint32(1), key.CollectionID)
}

func TestNewPolicyKeyFromString_InvalidStrings_ReturnError(t *testing.T) {
	for _, badKey := range []string{
		"",
		"/collection",
		"/collection/notpolicy",
		"/policy/collection",
		"/collection/policy",
	} {
		_, err := NewCollectionPolicyKeyFromString(badKey)
		assert.ErrorIs(t, err, ErrInvalidKey)
	}
}

func TestNewPolicyKeyFromString_InvalidStringColID_ReturnError(t *testing.T) {
	_, err := NewCollectionPolicyKeyFromString("/collection/policy/bad")
	assert.NotNil(t, err)
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

func toFieldValues(values ...string) [][]byte {
	var result [][]byte = make([][]byte, 0, len(values))
	for _, value := range values {
		result = append(result, []byte(value))
	}
	return result
}

func TestIndexDatastoreKey_ToString(t *testing.T) {
	cases := []struct {
		Key      IndexDataStoreKey
		Expected string
	}{
		{
			Key:      IndexDataStoreKey{},
			Expected: "",
		},
		{
			Key: IndexDataStoreKey{
				CollectionID: 1,
			},
			Expected: "/1",
		},
		{
			Key: IndexDataStoreKey{
				CollectionID: 1,
				IndexID:      2,
			},
			Expected: "/1/2",
		},
		{
			Key: IndexDataStoreKey{
				CollectionID: 1,
				IndexID:      2,
				FieldValues:  toFieldValues("3"),
			},
			Expected: "/1/2/3",
		},
		{
			Key: IndexDataStoreKey{
				CollectionID: 1,
				IndexID:      2,
				FieldValues:  toFieldValues("3", "4"),
			},
			Expected: "/1/2/3/4",
		},
		{
			Key: IndexDataStoreKey{
				CollectionID: 1,
				FieldValues:  toFieldValues("3"),
			},
			Expected: "/1",
		},
		{
			Key: IndexDataStoreKey{
				IndexID:     2,
				FieldValues: toFieldValues("3"),
			},
			Expected: "",
		},
		{
			Key: IndexDataStoreKey{
				FieldValues: toFieldValues("3"),
			},
			Expected: "",
		},
		{
			Key: IndexDataStoreKey{
				CollectionID: 1,
				IndexID:      2,
				FieldValues:  toFieldValues("", ""),
			},
			Expected: "/1/2",
		},
		{
			Key: IndexDataStoreKey{
				CollectionID: 1,
				IndexID:      2,
				FieldValues:  toFieldValues("", "3"),
			},
			Expected: "/1/2",
		},
		{
			Key: IndexDataStoreKey{
				CollectionID: 1,
				IndexID:      2,
				FieldValues:  toFieldValues("3", "", "4"),
			},
			Expected: "/1/2/3",
		},
	}
	for i, c := range cases {
		assert.Equal(t, c.Key.ToString(), c.Expected, "case %d", i)
	}
}

func TestIndexDatastoreKey_Bytes(t *testing.T) {
	key := IndexDataStoreKey{
		CollectionID: 1,
		IndexID:      2,
		FieldValues:  toFieldValues("3", "4"),
	}
	assert.Equal(t, key.Bytes(), []byte("/1/2/3/4"))
}

func TestIndexDatastoreKey_ToDS(t *testing.T) {
	key := IndexDataStoreKey{
		CollectionID: 1,
		IndexID:      2,
		FieldValues:  toFieldValues("3", "4"),
	}
	assert.Equal(t, key.ToDS(), ds.NewKey("/1/2/3/4"))
}

func TestIndexDatastoreKey_EqualTrue(t *testing.T) {
	cases := [][]IndexDataStoreKey{
		{
			{
				CollectionID: 1,
				IndexID:      2,
				FieldValues:  toFieldValues("3", "4"),
			},
			{
				CollectionID: 1,
				IndexID:      2,
				FieldValues:  toFieldValues("3", "4"),
			},
		},
		{
			{
				CollectionID: 1,
				FieldValues:  toFieldValues("3", "4"),
			},
			{
				CollectionID: 1,
				FieldValues:  toFieldValues("3", "4"),
			},
		},
		{
			{
				CollectionID: 1,
			},
			{
				CollectionID: 1,
			},
		},
	}

	for i, c := range cases {
		assert.True(t, c[0].Equal(c[1]), "case %d", i)
	}
}

func TestCollectionIndexKey_Bytes(t *testing.T) {
	key := CollectionIndexKey{
		CollectionID: immutable.Some[uint32](1),
		IndexName:    "idx",
	}
	assert.Equal(t, []byte(COLLECTION_INDEX+"/1/idx"), key.Bytes())
}

func TestIndexDatastoreKey_EqualFalse(t *testing.T) {
	cases := [][]IndexDataStoreKey{
		{
			{
				CollectionID: 1,
			},
			{
				CollectionID: 2,
			},
		},
		{
			{
				CollectionID: 1,
				IndexID:      2,
			},
			{
				CollectionID: 1,
				IndexID:      3,
			},
		},
		{
			{
				CollectionID: 1,
			},
			{
				IndexID: 1,
			},
		},
		{
			{
				CollectionID: 1,
				IndexID:      2,
				FieldValues:  toFieldValues("4", "3"),
			},
			{
				CollectionID: 1,
				IndexID:      2,
				FieldValues:  toFieldValues("3", "4"),
			},
		},
		{
			{
				CollectionID: 1,
				IndexID:      2,
				FieldValues:  toFieldValues("3"),
			},
			{
				CollectionID: 1,
				IndexID:      2,
				FieldValues:  toFieldValues("3", "4"),
			},
		},
		{
			{
				CollectionID: 1,
				FieldValues:  toFieldValues("3", "", "4"),
			},
			{
				CollectionID: 1,
				FieldValues:  toFieldValues("3", "4"),
			},
		},
	}

	for i, c := range cases {
		assert.False(t, c[0].Equal(c[1]), "case %d", i)
	}
}

func TestNewIndexDataStoreKey_ValidKey(t *testing.T) {
	str, err := NewIndexDataStoreKey("/1/2/3")
	assert.NoError(t, err)
	assert.Equal(t, str, IndexDataStoreKey{
		CollectionID: 1,
		IndexID:      2,
		FieldValues:  toFieldValues("3"),
	})

	str, err = NewIndexDataStoreKey("/1/2/3/4")
	assert.NoError(t, err)
	assert.Equal(t, str, IndexDataStoreKey{
		CollectionID: 1,
		IndexID:      2,
		FieldValues:  toFieldValues("3", "4"),
	})
}

func TestNewIndexDataStoreKey_InvalidKey(t *testing.T) {
	keys := []string{
		"",
		"/",
		"/1",
		"/1/2",
		" /1/2/3",
		"1/2/3",
		"/a/2/3",
		"/1/b/3",
	}
	for i, key := range keys {
		_, err := NewIndexDataStoreKey(key)
		assert.Error(t, err, "case %d: %s", i, key)
	}
}
