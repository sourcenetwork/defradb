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
	"testing"

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

func TestNewDataStoreKey_ReturnsCollectionIdAndIndexIdAndDocKeyAndFieldIdAndInstanceType_GivenFourItemsWithType(
	t *testing.T,
) {
	instanceType := "anyType"
	fieldId := "f1"
	docKey := "docKey"
	collectionId := "1"
	inputString := collectionId + "/" + instanceType + "/" + docKey + "/" + fieldId

	result, err := NewDataStoreKey(inputString)
	if err != nil {
		t.Error(err)
	}
	resultString := result.ToString()

	assert.Equal(
		t,
		DataStoreKey{
			CollectionID: collectionId,
			DocKey:       docKey,
			FieldId:      fieldId,
			InstanceType: InstanceType(instanceType)},
		result)
	assert.Equal(t, "/"+collectionId+"/"+instanceType+"/"+docKey+"/"+fieldId, resultString)
}

func TestNewDataStoreKey_ReturnsEmptyStruct_GivenAStringWithMissingElements(t *testing.T) {
	inputString := "/0/v"

	_, err := NewDataStoreKey(inputString)

	assert.ErrorIs(t, ErrInvalidKey, err)
}

func TestNewDataStoreKey_GivenAShortObjectMarker(t *testing.T) {
	instanceType := "anyType"
	docKey := "docKey"
	collectionId := "1"
	inputString := collectionId + "/" + instanceType + "/" + docKey

	result, err := NewDataStoreKey(inputString)
	if err != nil {
		t.Error(err)
	}
	resultString := result.ToString()

	assert.Equal(
		t,
		DataStoreKey{
			CollectionID: collectionId,
			DocKey:       docKey,
			InstanceType: InstanceType(instanceType)},
		result)
	assert.Equal(t, "/"+collectionId+"/"+instanceType+"/"+docKey, resultString)
}

func TestNewDataStoreKey_GivenAStringWithExtraPrefixes(t *testing.T) {
	instanceType := "anyType"
	fieldId := "f1"
	docKey := "docKey"
	collectionId := "1"
	inputString := "/db/my_database_name/data/" + collectionId + "/" + instanceType + "/" + docKey + "/" + fieldId

	_, err := NewDataStoreKey(inputString)

	assert.ErrorIs(t, ErrInvalidKey, err)
}

func TestNewDataStoreKey_GivenAStringWithExtraSuffix(t *testing.T) {
	instanceType := "anyType"
	fieldId := "f1"
	docKey := "docKey"
	collectionId := "1"
	inputString := "/db/data/" + collectionId + "/" + instanceType + "/" + docKey + "/" + fieldId + "/version_number"

	_, err := NewDataStoreKey(inputString)

	assert.ErrorIs(t, ErrInvalidKey, err)
}

func TestNewIndexKey_IfEmptyParam_ReturnPrefix(t *testing.T) {
	key := NewCollectionIndexKey("", "")
	assert.Equal(t, "/collection/index", key.ToString())
}

func TestNewIndexKey_ParamsAreGiven_ReturnFullKey(t *testing.T) {
	key := NewCollectionIndexKey("col", "idx")
	assert.Equal(t, "/collection/index/col/idx", key.ToString())
}

func TestNewIndexKey_InNoCollectionName_ReturnJustPrefix(t *testing.T) {
	key := NewCollectionIndexKey("", "idx")
	assert.Equal(t, "/collection/index", key.ToString())
}

func TestNewIndexKey_InNoIndexName_ReturnWithoutIndexName(t *testing.T) {
	key := NewCollectionIndexKey("col", "")
	assert.Equal(t, "/collection/index/col", key.ToString())
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
	key, err := NewCollectionIndexKeyFromString("/collection/index/col")
	assert.NoError(t, err)
	assert.Equal(t, key.CollectionID, "col")
	assert.Equal(t, key.IndexName, "")
}

func TestNewIndexKeyFromString_IfFullKeyString_ReturnKey(t *testing.T) {
	key, err := NewCollectionIndexKeyFromString("/collection/index/col/idx")
	assert.NoError(t, err)
	assert.Equal(t, key.CollectionID, "col")
	assert.Equal(t, key.IndexName, "idx")
}
