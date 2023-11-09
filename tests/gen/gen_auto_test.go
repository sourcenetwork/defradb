// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package gen

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getField(t *testing.T, doc map[string]any, fieldName string) any {
	field, ok := doc[fieldName]
	assert.True(t, ok, "field %s not found", fieldName)
	return field
}

func getStringField(t *testing.T, doc map[string]any, fieldName string) string {
	val, ok := getField(t, doc, fieldName).(string)
	assert.True(t, ok, "field %s is not of type string", fieldName)
	return val
}

func getIntField(t *testing.T, doc map[string]any, fieldName string) int {
	switch val := getField(t, doc, fieldName).(type) {
	case int:
		return val
	case float64:
		return int(val)
	}
	assert.Fail(t, "field %s is not of type int or float64", fieldName)
	return 0
}

func getFloatField(t *testing.T, doc map[string]any, fieldName string) float64 {
	val, ok := getField(t, doc, fieldName).(float64)
	assert.True(t, ok, "field %s is not of type float64", fieldName)
	return val
}

func getBooleanField(t *testing.T, doc map[string]any, fieldName string) bool {
	val, ok := getField(t, doc, fieldName).(bool)
	assert.True(t, ok, "field %s is not of type bool", fieldName)
	return val
}

func getDocKeysFromJSONDocs(jsonDocs []string) []string {
	var result []string
	for _, doc := range jsonDocs {
		result = append(result, getDocKeyFromDocJSON(doc))
	}
	return result
}

func filterByCollection(docs []GeneratedDoc, i int) []string {
	var result []string
	for _, doc := range docs {
		if doc.ColIndex == i {
			result = append(result, doc.JSON)
		}
	}
	return result
}

func removeDuplicateStr(strSlice []string) []string {
	allKeys := make(map[string]bool)
	list := make([]string, 0, len(strSlice))
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func assertDocKeysMatch(
	t *testing.T,
	docs []GeneratedDoc,
	primaryCol, secondaryCol int,
	foreignField string,
	allowDuplicates bool,
) {
	primaryDocs := filterByCollection(docs, primaryCol)
	secondaryDocs := filterByCollection(docs, secondaryCol)

	docKeys := getDocKeysFromJSONDocs(primaryDocs)
	foreignValues := make([]string, 0, len(secondaryDocs))
	for _, secDoc := range secondaryDocs {
		foreignValues = append(foreignValues, getStringField(t, jsonToMap(secDoc), foreignField))
	}

	if allowDuplicates {
		newValues := removeDuplicateStr(foreignValues)
		foreignValues = newValues
	}

	assert.ElementsMatch(t, docKeys, foreignValues)
}

func assertUniformlyDistributedIntFieldRange(t *testing.T, docs []GeneratedDoc, fieldName string, minVal, maxVal int) {
	vals := make(map[int]bool, len(docs))
	foundMin := math.MaxInt
	foundMax := math.MinInt
	for _, doc := range docs {
		val := getIntField(t, jsonToMap(doc.JSON), fieldName)
		vals[val] = true
		if val < foundMin {
			foundMin = val
		}
		if val > foundMax {
			foundMax = val
		}
	}
	intRange := maxVal - minVal
	intPrecision := (intRange * 20 / len(docs))
	assert.LessOrEqual(t, foundMin, minVal+intPrecision, "field %s is not distributed across the range", fieldName)
	assert.GreaterOrEqual(t, foundMax, maxVal-intPrecision-1, "field %s is not distributed across the range", fieldName)

	expectedLen := len(docs)
	if intRange < expectedLen {
		expectedLen = intRange
	}
	expectedLen = int(float64(expectedLen) * 0.7)
	assert.GreaterOrEqual(t, len(vals), expectedLen, "values of field %s are not uniformly distributed", fieldName)
}

func assertUniformlyDistributedStringField(t *testing.T, docs []GeneratedDoc, fieldName string, strLen int) {
	vals := make(map[string]bool, len(docs))
	var wrongStr string
	for _, doc := range docs {
		val := getStringField(t, jsonToMap(doc.JSON), fieldName)
		vals[val] = true
		if len(val) != strLen {
			wrongStr = val
		}
	}
	if wrongStr != "" {
		assert.Fail(t, "unexpected string length", "encountered %s field's value with unexpected len. Example: %s should be of len %d",
			fieldName, wrongStr, strLen)
	}
	assert.GreaterOrEqual(t, len(vals), int(float64(len(docs))*0.99),
		"values of field %s are not uniformly distributed", fieldName)
}

func assertUniformlyDistributedBoolField(t *testing.T, docs []GeneratedDoc, fieldName string) {
	userVerifiedCounter := 0

	for _, doc := range docs {
		docMap := jsonToMap(doc.JSON)

		if getBooleanField(t, docMap, fieldName) {
			userVerifiedCounter++
		}
	}

	assert.GreaterOrEqual(t, userVerifiedCounter, int(float64(len(docs))*0.4), "values of field %s are not uniformly distributed", fieldName)
}

func assertUniformlyDistributedFloatFieldRange(t *testing.T, docs []GeneratedDoc, fieldName string, minVal, maxVal float64) {
	vals := make(map[float64]bool, len(docs))
	foundMin := math.Inf(1)
	foundMax := math.Inf(-1)
	for _, doc := range docs {
		val := getFloatField(t, jsonToMap(doc.JSON), fieldName)
		vals[val] = true
		if val < foundMin {
			foundMin = val
		}
		if val > foundMax {
			foundMax = val
		}
	}
	floatPrecision := ((maxVal - minVal) / float64(len(docs))) * 20
	assert.LessOrEqual(t, foundMin, minVal+floatPrecision, "field %s is not distributed across the range", fieldName)
	assert.GreaterOrEqual(t, foundMax, maxVal-floatPrecision, "field %s is not distributed across the range", fieldName)

	assert.GreaterOrEqual(t, len(vals), int(float64(len(docs))*0.7), "values of field %s are not uniformly distributed", fieldName)
}

func assertUniformRelationDistribution(
	t *testing.T,
	docs []GeneratedDoc,
	primaryColInd, secondaryColInd int,
	foreignField string,
	min, max int,
) {
	primaryCol := filterByCollection(docs, primaryColInd)
	secondaryCol := filterByCollection(docs, secondaryColInd)
	assert.GreaterOrEqual(t, len(secondaryCol), len(primaryCol)*min)
	assert.LessOrEqual(t, len(secondaryCol), len(primaryCol)*max)

	secondaryPerPrimary := make(map[string]int)
	for _, d := range secondaryCol {
		docKey := getStringField(t, jsonToMap(d), foreignField)
		secondaryPerPrimary[docKey]++
	}
	minDocsPerPrimary := math.MaxInt
	maxDocsPerPrimary := math.MinInt
	for _, numDevices := range secondaryPerPrimary {
		if numDevices < minDocsPerPrimary {
			minDocsPerPrimary = numDevices
		}
		if numDevices > maxDocsPerPrimary {
			maxDocsPerPrimary = numDevices
		}
	}

	assert.LessOrEqual(t, minDocsPerPrimary, min+1)
	assert.GreaterOrEqual(t, minDocsPerPrimary, min)

	assert.LessOrEqual(t, maxDocsPerPrimary, max)
	assert.GreaterOrEqual(t, maxDocsPerPrimary, max-1)
}

func TestAutoGenerateDocs_Simple(t *testing.T) {
	const numUsers = 1000
	schema := `
		type User {
			name: String
			age: Int
			verified: Boolean
			rating: Float
		}`

	docs, err := AutoGenerateDocs(schema, WithTypeDemand("User", numUsers))
	assert.NoError(t, err)
	assert.Len(t, docs, numUsers)

	assertUniformlyDistributedStringField(t, docs, "name", 10)
	assertUniformlyDistributedIntFieldRange(t, docs, "age", 0, 10000)
	assertUniformlyDistributedBoolField(t, docs, "verified")
	assertUniformlyDistributedFloatFieldRange(t, docs, "rating", 0.0, 1.0)
}

func TestAutoGenerateDocs_ConfigIntRange(t *testing.T) {
	const numUsers = 1000
	schema := `
		type User {
			age: Int # min: 1, max: 120
			money: Int # min: -1000, max: 10000
		}`

	docs, err := AutoGenerateDocs(schema, WithTypeDemand("User", numUsers))
	assert.NoError(t, err)
	assert.Len(t, docs, numUsers)

	assertUniformlyDistributedIntFieldRange(t, docs, "age", 1, 120)
	assertUniformlyDistributedIntFieldRange(t, docs, "money", -1000, 10000)
}

func TestAutoGenerateDocs_ConfigFloatRange(t *testing.T) {
	const numUsers = 1000
	schema := `
		type User {
			rating: Float # min: 1.5, max: 5.0
			product: Float # min: -1.0, max: 1.0
		}`

	docs, err := AutoGenerateDocs(schema, WithTypeDemand("User", numUsers))
	assert.NoError(t, err)
	assert.Len(t, docs, numUsers)

	assertUniformlyDistributedFloatFieldRange(t, docs, "rating", 1.5, 5)
	assertUniformlyDistributedFloatFieldRange(t, docs, "product", -1, 1)
}

func TestAutoGenerateDocs_ConfigStringLen(t *testing.T) {
	const numUsers = 1000
	schema := `
		type User {
			name: String # len: 8
			email: String # len: 12
		}`

	docs, err := AutoGenerateDocs(schema, WithTypeDemand("User", numUsers))
	assert.NoError(t, err)
	assert.Len(t, docs, numUsers)

	assertUniformlyDistributedStringField(t, docs, "name", 8)
	assertUniformlyDistributedStringField(t, docs, "email", 12)
}

func TestAutoGenerateDocs_IfNoTypeDemandIsGiven_ShouldUseDefault(t *testing.T) {
	schema := `
		type User {
			name: String 
		}`

	docs, err := AutoGenerateDocs(schema)
	assert.NoError(t, err)

	const defaultDemand = 10
	assert.Len(t, filterByCollection(docs, 0), defaultDemand)
}

func TestAutoGenerateDocs_RelationOneToOne(t *testing.T) {
	const numUsers = 10
	schema := `
		type User {
			name: String 
			device: Device
		}
		
		type Device {
			owner: User
			model: String
		}`

	docs, err := AutoGenerateDocs(schema, WithTypeDemand("User", numUsers))
	assert.NoError(t, err)

	assert.Len(t, filterByCollection(docs, 0), numUsers)
	assert.Len(t, filterByCollection(docs, 1), numUsers)

	assertDocKeysMatch(t, docs, 0, 1, "owner_id", false)
}

func TestAutoGenerateDocs_RelationOneToMany(t *testing.T) {
	const numUsers = 10
	schema := `
		type User { 
			name: String 
			devices: [Device] 
		}
		
		type Device {
			owner: User
			model: String
		}`

	docs, err := AutoGenerateDocs(schema, WithTypeDemand("User", numUsers))
	assert.NoError(t, err)

	assert.Len(t, filterByCollection(docs, 0), numUsers)
	assert.Len(t, filterByCollection(docs, 1), numUsers*2)

	assertDocKeysMatch(t, docs, 0, 1, "owner_id", true)
}

func TestAutoGenerateDocs_RelationOneToManyWithConfiguredNumberOfElements(t *testing.T) {
	const (
		numUsers          = 100
		minDevicesPerUser = 1
		maxDevicesPerUser = 5
	)
	schema := `
		type User { 
			name: String 
			devices: [Device] # min: 1, max: 5
		}
		
		type Device {
			owner: User
			model: String
		}`

	docs, err := AutoGenerateDocs(schema, WithTypeDemand("User", numUsers))
	assert.NoError(t, err)

	assert.Len(t, filterByCollection(docs, 0), numUsers)

	assertUniformRelationDistribution(t, docs, 0, 1, "owner_id", minDevicesPerUser, maxDevicesPerUser)

	assertDocKeysMatch(t, docs, 0, 1, "owner_id", true)
}

func TestAutoGenerateDocs_RelationOneToManyToOneWithConfiguredNumberOfElements(t *testing.T) {
	const (
		numUsers       = 100
		devicesPerUser = 2
	)
	schema := `
		type User { 
			name: String 
			devices: [Device] # min: 2, max: 2
		}
		
		type Device {
			owner: User
			model: String
			specs: Specs 
		}
		
		type Specs {
			device: Device
			OS: String
		}`

	docs, err := AutoGenerateDocs(schema, WithTypeDemand("User", numUsers))
	assert.NoError(t, err)

	assert.Len(t, filterByCollection(docs, 0), numUsers)
	assert.Len(t, filterByCollection(docs, 1), numUsers*devicesPerUser)
	assert.Len(t, filterByCollection(docs, 2), numUsers*devicesPerUser)

	assertUniformRelationDistribution(t, docs, 0, 1, "owner_id", devicesPerUser, devicesPerUser)

	assertDocKeysMatch(t, docs, 0, 1, "owner_id", true)
	assertDocKeysMatch(t, docs, 1, 2, "device_id", false)
}

func TestAutoGenerateDocs_RelationOneToManyToOnePrimaryWithConfiguredNumberOfElements(t *testing.T) {
	const (
		numUsers       = 100
		devicesPerUser = 2
	)
	schema := `
		type User { 
			name: String 
			devices: [Device] # min: 2, max: 2
		}
		
		type Device {
			owner: User
			model: String
			specs: Specs @primary
		}
		
		type Specs {
			device: Device 
			OS: String
		}`

	docs, err := AutoGenerateDocs(schema, WithTypeDemand("User", numUsers))
	assert.NoError(t, err)

	assert.Len(t, filterByCollection(docs, 0), numUsers)
	assert.Len(t, filterByCollection(docs, 1), numUsers*devicesPerUser)
	assert.Len(t, filterByCollection(docs, 2), numUsers*devicesPerUser)

	assertUniformRelationDistribution(t, docs, 0, 1, "owner_id", devicesPerUser, devicesPerUser)

	assertDocKeysMatch(t, docs, 0, 1, "owner_id", true)
	assertDocKeysMatch(t, docs, 2, 1, "specs_id", false)
}

func TestAutoGenerateDocs_RelationOneToManyToManyWithNumDocsForSecondaryType(t *testing.T) {
	const (
		numDevices          = 40
		devicesPerUser      = 2
		componentsPerDevice = 5
	)
	schema := `
		type User { 
			name: String 
			devices: [Device] # min: 2, max: 2
		}
		
		type Device {
			owner: User
			model: String
			specs: Specs 
			components: [Component] # min: 5, max: 5
		}
		
		type Specs {
			device: Device
			OS: String
		}
		
		type Component {
			device: Device
			serialNumber: String
		}`

	docs, err := AutoGenerateDocs(schema, WithTypeDemand("Device", numDevices))
	assert.NoError(t, err)

	assert.Len(t, filterByCollection(docs, 0), numDevices/devicesPerUser)
	assert.Len(t, filterByCollection(docs, 1), numDevices)
	assert.Len(t, filterByCollection(docs, 2), numDevices)
	assert.Len(t, filterByCollection(docs, 3), numDevices*componentsPerDevice)

	assertUniformRelationDistribution(t, docs, 0, 1, "owner_id", devicesPerUser, devicesPerUser)
	assertUniformRelationDistribution(t, docs, 1, 2, "device_id", 1, 1)
	assertUniformRelationDistribution(t, docs, 1, 3, "device_id", componentsPerDevice, componentsPerDevice)

	assertDocKeysMatch(t, docs, 0, 1, "owner_id", true)
	assertDocKeysMatch(t, docs, 1, 2, "device_id", false)
	assertDocKeysMatch(t, docs, 1, 3, "device_id", true)
}

func TestAutoGenerateDocs_DemandsForDifferentRelationTrees(t *testing.T) {
	const (
		numUsers            = 20
		numDevices          = 15
		componentsPerDevice = 2
	)
	schema := `
		type User { 
			name: String 
		}
		
		type Device {
			model: String
			components: [Component] # min: 2, max: 2
		}
		
		type Component {
			device: Device
			serialNumber: String
		}`

	docs, err := AutoGenerateDocs(
		schema,
		WithTypeDemand("User", numUsers),
		WithTypeDemand("Device", numDevices),
	)
	assert.NoError(t, err)

	assert.Len(t, filterByCollection(docs, 0), numUsers)
	assert.Len(t, filterByCollection(docs, 1), numDevices)
	assert.Len(t, filterByCollection(docs, 2), numDevices*componentsPerDevice)

	assertUniformRelationDistribution(t, docs, 1, 2, "device_id", componentsPerDevice, componentsPerDevice)

	assertDocKeysMatch(t, docs, 1, 2, "device_id", true)
}

func TestAutoGenerateDocs_IfTypeDemandedForSameTreeAddsUp_ShouldGenerate(t *testing.T) {
	schema := `
		type User { 
			name: String 
			devices: [Device] 
			Orders: [Order] # min: 10, max: 10
		}
		
		type Device {
			model: String
			owner: User
		}
		
		type Order {
			id: String
			user: User
		}`

	docs, err := AutoGenerateDocs(
		schema,
		WithTypeDemand("Order", 10),
		WithTypeDemand("Device", 30),
	)
	assert.NoError(t, err)

	assert.Len(t, filterByCollection(docs, 0), 1)
	assert.Len(t, filterByCollection(docs, 1), 30)
	assert.Len(t, filterByCollection(docs, 2), 10)
}

func TestAutoGenerateDocs_IfTypeDemandedForSameTreeDoesNotAddUp_ShouldReturnError(t *testing.T) {
	schema := `
		type User { 
			name: String 
			devices: [Device] # min: 0, max: 29
			Orders: [Order] # min: 10, max: 10
		}
		
		type Device {
			model: String
			owner: User
		}
		
		type Order {
			id: String
			user: User
		}`

	_, err := AutoGenerateDocs(
		schema,
		WithTypeDemand("Order", 10),
		WithTypeDemand("Device", 30),
	)
	assert.ErrorContains(t, err, errInvalidConfiguration)
}

func TestAutoGenerateDocs_IfNoDemandForPrimaryType_ShouldDeduceFromMaxSecondaryDemand(t *testing.T) {
	schema := `
		type User {
			name: String
			devices: [Device]
			Orders: [Order]
		}

		type Device {
			model: String
			owner: User
		}

		type Order {
			id: String
			user: User
		}`

	docs, err := AutoGenerateDocs(
		schema,
		WithTypeDemand("Order", 10),
		WithTypeDemand("Device", 30),
	)
	assert.NoError(t, err)

	// users len should be equal to max of secondary type demand
	assert.Len(t, filterByCollection(docs, 0), 30)
	assert.Len(t, filterByCollection(docs, 1), 30)
	assert.Len(t, filterByCollection(docs, 2), 10)
}

func TestAutoGenerateDocs_InvalidConfig(t *testing.T) {
	testCases := []struct {
		name     string
		schema   string
		typeName string
		demand   int
	}{
		{
			name: "demand for direct parent can not be satisfied",
			schema: `
				type User {
					name: String
					devices: [Device] # min: 2, max: 2
				}

				type Device {
					model: String
					owner: User
				}`,
			typeName: "Device",
			demand:   1,
		},
		{
			name: "demand for grand parent can not be satisfied",
			schema: `
				type User {
					name: String
					devices: [Device] # min: 2, max: 2
				}

				type Device {
					model: String
					owner: User
					components: [Component] # min: 2, max: 2
				}

				type Component {
					device: Device
					OS: String
				}`,
			typeName: "Component",
			demand:   2,
		},
		{
			name: "demand for sibling primary can not be satisfied",
			schema: `
				type User { 
					name: String 
					devices: [Device] # min: 2, max: 2
				}
				
				type Device {
					model: String
					owner: User
					manufacturer: Manufacturer
				}

				type Manufacturer {
					name: String
					devices: [Device] # min: 10, max: 10
				}`,
			typeName: "User",
			demand:   4,
		},
		{
			name: "array: max is less than min",
			schema: `
				type User { 
					name: String 
					devices: [Device] # min: 2, max: 1
				}
				
				type Device {
					model: String
					owner: User
				}`,
			typeName: "User",
			demand:   4,
		},
		{
			name: "array: min is negative",
			schema: `
				type User { 
					name: String 
					devices: [Device] # min: -1, max: 10
				}
				
				type Device {
					model: String
					owner: User
				}`,
			typeName: "User",
			demand:   4,
		},
		{
			name: "array: missing max",
			schema: `
				type User { 
					name: String 
					devices: [Device] # min: 2
				}
				
				type Device {
					model: String
					owner: User
				}`,
			typeName: "User",
			demand:   4,
		},
		{
			name: "array: missing min",
			schema: `
				type User { 
					name: String 
					devices: [Device] # max: 10
				}
				
				type Device {
					model: String
					owner: User
				}`,
			typeName: "User",
			demand:   4,
		},
		{
			name: "array: min as float",
			schema: `
				type User { 
					name: String 
					devices: [Device] # min: 2.5, max: 10
				}
				
				type Device {
					model: String
					owner: User
				}`,
			typeName: "User",
			demand:   4,
		},
		{
			name: "array: max as float",
			schema: `
				type User { 
					name: String 
					devices: [Device] # min: 2, max: 10.0
				}
				
				type Device {
					model: String
					owner: User
				}`,
			typeName: "User",
			demand:   4,
		},
		{
			name: "int value: max is less than min",
			schema: `
				type User { 
					age: Int # min: 10, max: 2
				}`,
			typeName: "User",
			demand:   4,
		},
		{
			name: "int value: missing min",
			schema: `
				type User { 
					age: Int # max: 2
				}`,
			typeName: "User",
			demand:   4,
		},
		{
			name: "int value: missing max",
			schema: `
				type User { 
					age: Int # min: 2
				}`,
			typeName: "User",
			demand:   4,
		},
		{
			name: "float value: max is less than min",
			schema: `
				type User { 
					rating: Float # min: 10.5, max: 2.5
				}`,
			typeName: "User",
			demand:   4,
		},
		{
			name: "float value: missing min",
			schema: `
				type User { 
					rating: Float # max: 2.5
				}`,
			typeName: "User",
			demand:   4,
		},
		{
			name: "float value: missing max",
			schema: `
				type User { 
					rating: Float # min: 2.5
				}`,
			typeName: "User",
			demand:   4,
		},
		{
			name: "min/max on not number",
			schema: `
				type User { 
					verified: Bool # min: 2, max: 8
				}`,
			typeName: "User",
			demand:   4,
		},
		{
			name: "string: zero len",
			schema: `
				type User { 
					name: String # len: 0
				}`,
			typeName: "User",
			demand:   4,
		},
		{
			name: "string: negative len",
			schema: `
				type User { 
					name: String # len: -2
				}`,
			typeName: "User",
			demand:   4,
		},
		{
			name: "string: non-int len",
			schema: `
				type User { 
					name: String # len: 8.5
				}`,
			typeName: "User",
			demand:   4,
		},
		{
			name: "len on non-string",
			schema: `
				type User { 
					age: Int # len: 8
				}`,
			typeName: "User",
			demand:   4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := AutoGenerateDocs(tc.schema, WithTypeDemand(tc.typeName, tc.demand))

			assert.ErrorContains(t, err, errInvalidConfiguration)
		})
	}
}

func TestAutoGenerateDocs_CustomFieldValueGenerator(t *testing.T) {
	const (
		numUsers = 10
		ageVal   = 7
	)
	schema := `
		type User { 
			name: String 
			age: Int 
		}`

	indexes := make([]int, 0, numUsers)
	intVals := make(map[int]bool)

	docs, err := AutoGenerateDocs(schema,
		WithTypeDemand("User", numUsers),
		WithFieldGenerator("User", "age", func(i int, next func() any) any {
			indexes = append(indexes, i)
			intVals[next().(int)] = true
			return ageVal
		}),
	)
	assert.NoError(t, err)

	expectedIndexes := make([]int, 0, numUsers)
	for i := 0; i < numUsers; i++ {
		expectedIndexes = append(expectedIndexes, i)
	}

	assert.Equal(t, expectedIndexes, indexes)
	assert.GreaterOrEqual(t, len(intVals), numUsers-1)

	for _, doc := range docs {
		actualAgeVal := getIntField(t, jsonToMap(doc.JSON), "age")
		assert.Equal(t, ageVal, actualAgeVal)
	}
}

func TestAutoGenerateDocs_IfOptionOverlapsSchemaConfig_ItShouldOverwrite(t *testing.T) {
	const (
		numUsers = 20
	)
	schema := `
		type User { 
			name: String # len: 8
			devices: [Device] # min: 2, max: 2
			age: Int # min: 10, max: 20
			rating: Float # min: 0.0, max: 1.0
		}
		
		type Device {
			model: String
			owner: User
		}`

	docs, err := AutoGenerateDocs(
		schema,
		WithTypeDemand("User", numUsers),
		WithFieldRange("User", "devices", 3, 3),
		WithFieldRange("User", "age", 30, 40),
		WithFieldRange("User", "rating", 1.0, 2.0),
		WithFieldLen("User", "name", 6),
	)
	assert.NoError(t, err)

	userJSONs := filterByCollection(docs, 0)
	assert.Len(t, userJSONs, numUsers)
	assert.Len(t, filterByCollection(docs, 1), numUsers*3)

	for _, userJSON := range userJSONs {
		userMap := jsonToMap(userJSON)

		actualAgeVal := getIntField(t, userMap, "age")
		assert.GreaterOrEqual(t, actualAgeVal, 30)
		assert.LessOrEqual(t, actualAgeVal, 40)

		actualRatingVal := getFloatField(t, userMap, "rating")
		assert.GreaterOrEqual(t, actualRatingVal, 1.0)
		assert.LessOrEqual(t, actualRatingVal, 2.0)

		actualNameVal := getStringField(t, userMap, "name")
		assert.Len(t, actualNameVal, 6)
	}
}
