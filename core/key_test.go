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

	result := NewDataStoreKey(inputString)
	resultString := result.ToString()

	assert.Equal(t, DataStoreKey{}, result)
	assert.Equal(t, "", resultString)
}

func TestNewDataStoreKey_ReturnsCollectionIdAndIndexIdAndDocKeyAndFieldIdAndInstanceType_GivenFourItemsWithType(
	t *testing.T,
) {
	instanceType := "anyType"
	fieldId := "f1"
	docKey := "docKey"
	collectionId := "1"
	inputString := collectionId + "/" + instanceType + "/" + docKey + "/" + fieldId

	result := NewDataStoreKey(inputString)
	resultString := result.ToString()

	assert.Equal(
		t,
		DataStoreKey{
			CollectionId: collectionId,
			DocKey:       docKey,
			FieldId:      fieldId,
			InstanceType: InstanceType(instanceType)},
		result)
	assert.Equal(t, "/"+collectionId+"/"+instanceType+"/"+docKey+"/"+fieldId, resultString)
}

func TestNewDataStoreKey_ReturnsEmptyStruct_GivenAStringWithMissingElements(t *testing.T) {
	inputString := "/0/v"

	result := NewDataStoreKey(inputString)
	resultString := result.ToString()

	assert.Equal(t, DataStoreKey{}, result)
	assert.Equal(t, "", resultString)
}

func TestNewDataStoreKey_GivenAShortObjectMarker(t *testing.T) {
	instanceType := "anyType"
	docKey := "docKey"
	collectionId := "1"
	inputString := collectionId + "/" + instanceType + "/" + docKey

	result := NewDataStoreKey(inputString)
	resultString := result.ToString()

	assert.Equal(
		t,
		DataStoreKey{
			CollectionId: collectionId,
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

	result := NewDataStoreKey(inputString)
	resultString := result.ToString()

	assert.Equal(
		t,
		DataStoreKey{
			CollectionId: collectionId,
			DocKey:       docKey,
			FieldId:      fieldId,
			InstanceType: InstanceType(instanceType)},
		result)
	assert.Equal(t, "/"+collectionId+"/"+instanceType+"/"+docKey+"/"+fieldId, resultString)
}

func TestNewDataStoreKey_GivenAStringWithExtraSuffix(t *testing.T) {
	instanceType := "anyType"
	fieldId := "f1"
	docKey := "docKey"
	collectionId := "1"
	inputString := "/db/data/" + collectionId + "/" + instanceType + "/" + docKey + "/" + fieldId + "/version_number"

	result := NewDataStoreKey(inputString)
	resultString := result.ToString()

	assert.Equal(
		t,
		DataStoreKey{
			CollectionId: collectionId,
			DocKey:       docKey,
			FieldId:      fieldId,
			InstanceType: InstanceType(instanceType)},
		result)
	assert.Equal(t, "/"+collectionId+"/"+instanceType+"/"+docKey+"/"+fieldId, resultString)
}
