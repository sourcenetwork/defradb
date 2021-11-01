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

func TestNewDataStoreKey_ReturnsDocKey_GivenSingleItem(t *testing.T) {
	docKey := "docKey"
	inputString := docKey

	result := NewDataStoreKey(inputString)
	resultString := result.ToString()

	assert.Equal(t, DataStoreKey{DocKey: docKey}, result)
	assert.Equal(t, "/"+docKey, resultString)
}

func TestNewDataStoreKey_ReturnsDocKey_GivenSingleItemWithLeadingDeliminator(t *testing.T) {
	docKey := "docKey"
	inputString := "/" + docKey

	result := NewDataStoreKey(inputString)
	resultString := result.ToString()

	assert.Equal(t, DataStoreKey{DocKey: docKey}, result)
	assert.Equal(t, "/"+docKey, resultString)
}

func TestNewDataStoreKey_ReturnsDocKeyAndInstanceType_GivenSingleItemWithType(t *testing.T) {
	instanceType := "anyType"
	docKey := "docKey"
	inputString := docKey + ":" + instanceType

	result := NewDataStoreKey(inputString)
	resultString := result.ToString()

	assert.Equal(t, DataStoreKey{DocKey: docKey, InstanceType: InstanceType(instanceType)}, result)
	assert.Equal(t, "/"+docKey+":"+instanceType, resultString)
}

func TestNewDataStoreKey_ReturnsDocKeyAndFieldIdAndInstanceType_GivenTwoItemsWithType(t *testing.T) {
	instanceType := "anyType"
	fieldId := "f1"
	docKey := "docKey"
	inputString := docKey + "/" + fieldId + ":" + instanceType

	result := NewDataStoreKey(inputString)
	resultString := result.ToString()

	assert.Equal(t, DataStoreKey{DocKey: docKey, FieldId: fieldId, InstanceType: InstanceType(instanceType)}, result)
	assert.Equal(t, "/"+docKey+"/"+fieldId+":"+instanceType, resultString)
}

func TestNewDataStoreKey_ReturnsIndexIdAndDocKeyAndFieldIdAndInstanceType_GivenThreeItemsWithType(t *testing.T) {
	instanceType := "anyType"
	fieldId := "f1"
	docKey := "docKey"
	indexId := "0"
	inputString := indexId + "/" + docKey + "/" + fieldId + ":" + instanceType

	result := NewDataStoreKey(inputString)
	resultString := result.ToString()

	assert.Equal(t, DataStoreKey{PrimaryIndexId: indexId, DocKey: docKey, FieldId: fieldId, InstanceType: InstanceType(instanceType)}, result)
	assert.Equal(t, "/"+indexId+"/"+docKey+"/"+fieldId+":"+instanceType, resultString)
}

func TestNewDataStoreKey_ReturnsCollectionIdAndIndexIdAndDocKeyAndFieldIdAndInstanceType_GivenFourItemsWithType(t *testing.T) {
	instanceType := "anyType"
	fieldId := "f1"
	docKey := "docKey"
	indexId := "0"
	collectionId := "1"
	inputString := collectionId + "/" + indexId + "/" + docKey + "/" + fieldId + ":" + instanceType

	result := NewDataStoreKey(inputString)
	resultString := result.ToString()

	assert.Equal(
		t,
		DataStoreKey{
			CollectionId:   collectionId,
			PrimaryIndexId: indexId,
			DocKey:         docKey,
			FieldId:        fieldId,
			InstanceType:   InstanceType(instanceType)},
		result)
	assert.Equal(t, "/"+collectionId+"/"+indexId+"/"+docKey+"/"+fieldId+":"+instanceType, resultString)
}

func TestNewDataStoreKey_ReturnsCollectionIdAndIndexIdAndDocKeyAndFieldIdAndInstanceType_GivenFourItemsWithTypeWithLeadingDeliminator(t *testing.T) {
	instanceType := "anyType"
	fieldId := "f1"
	docKey := "docKey"
	indexId := "0"
	collectionId := "1"
	inputString := "/" + collectionId + "/" + indexId + "/" + docKey + "/" + fieldId + ":" + instanceType

	result := NewDataStoreKey(inputString)
	resultString := result.ToString()

	assert.Equal(
		t,
		DataStoreKey{
			CollectionId:   collectionId,
			PrimaryIndexId: indexId,
			DocKey:         docKey,
			FieldId:        fieldId,
			InstanceType:   InstanceType(instanceType)},
		result)
	assert.Equal(t, "/"+collectionId+"/"+indexId+"/"+docKey+"/"+fieldId+":"+instanceType, resultString)
}

func TestNewDataStoreKey_ReturnsCollectionIdAndIndexIdAndDocKeyAndFieldIdAndInstanceType_GivenFourItemsWithTypeWithLeadingStuff(t *testing.T) {
	instanceType := "anyType"
	fieldId := "f1"
	docKey := "docKey"
	indexId := "0"
	collectionId := "1"
	inputString := "discarded/" + collectionId + "/" + indexId + "/" + docKey + "/" + fieldId + ":" + instanceType

	result := NewDataStoreKey(inputString)
	resultString := result.ToString()

	assert.Equal(
		t,
		DataStoreKey{
			CollectionId:   collectionId,
			PrimaryIndexId: indexId,
			DocKey:         docKey,
			FieldId:        fieldId,
			InstanceType:   InstanceType(instanceType)},
		result)
	assert.Equal(t, "/"+collectionId+"/"+indexId+"/"+docKey+"/"+fieldId+":"+instanceType, resultString)
}
