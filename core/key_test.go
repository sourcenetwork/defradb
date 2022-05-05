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
			InstanceType: InstanceType(instanceType),
		},
		result)
	assert.Equal(t, "/"+collectionId+"/"+instanceType+"/"+docKey+"/"+fieldId, resultString)
}
