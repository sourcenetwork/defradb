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
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/defradb/client"
)

func getField(t *testing.T, doc *client.Document, fieldName string) any {
	fVal, err := doc.GetValue(fieldName)
	if err != nil {
		assert.Fail(t, "field %s not found", fieldName)
	}
	return fVal.Value()
}

func getStringField(t *testing.T, doc *client.Document, fieldName string) string {
	val, ok := getField(t, doc, fieldName).(string)
	assert.True(t, ok, "field %s is not of type string", fieldName)
	return val
}

func getIntField(t *testing.T, doc *client.Document, fieldName string) int {
	fVal := getField(t, doc, fieldName)
	switch val := fVal.(type) {
	case int:
		return val
	case float64:
		return int(val)
	case int64:
		return int(val)
	}
	assert.Fail(t, "field %s is not of type int or float64", fieldName)
	return 0
}

func getFloatField(t *testing.T, doc *client.Document, fieldName string) float64 {
	val, ok := getField(t, doc, fieldName).(float64)
	assert.True(t, ok, "field %s is not of type float64", fieldName)
	return val
}

func getBooleanField(t *testing.T, doc *client.Document, fieldName string) bool {
	val, ok := getField(t, doc, fieldName).(bool)
	assert.True(t, ok, "field %s is not of type bool", fieldName)
	return val
}

func getDocIDsFromDocs(docs []*client.Document) []string {
	result := make([]string, 0, len(docs))
	for _, doc := range docs {
		result = append(result, doc.ID().String())
	}
	return result
}

func filterByCollection(docs []GeneratedDoc, name string) []*client.Document {
	var result []*client.Document
	for _, doc := range docs {
		if doc.Col.Description.Name == name {
			result = append(result, doc.Doc)
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

func assertDocIDsMatch(
	t *testing.T,
	docs []GeneratedDoc,
	primaryCol, secondaryCol string,
	foreignField string,
	allowDuplicates bool,
) {
	primaryDocs := filterByCollection(docs, primaryCol)
	secondaryDocs := filterByCollection(docs, secondaryCol)

	docIDs := getDocIDsFromDocs(primaryDocs)
	foreignValues := make([]string, 0, len(secondaryDocs))
	for _, secDoc := range secondaryDocs {
		foreignValues = append(foreignValues, getStringField(t, secDoc, foreignField))
	}

	if allowDuplicates {
		newValues := removeDuplicateStr(foreignValues)
		foreignValues = newValues
	}

	assert.ElementsMatch(t, docIDs, foreignValues)
}

func assertUniformlyDistributedIntFieldRange(t *testing.T, docs []GeneratedDoc, fieldName string, minVal, maxVal int) {
	vals := make(map[int]bool, len(docs))
	foundMin := math.MaxInt
	foundMax := math.MinInt
	for _, doc := range docs {
		val := getIntField(t, doc.Doc, fieldName)
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
		val := getStringField(t, doc.Doc, fieldName)
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

func assertUniformlyDistributedBoolField(t *testing.T, docs []GeneratedDoc, fieldName string, ratio float64) {
	trueCounter := 0

	for _, doc := range docs {
		if getBooleanField(t, doc.Doc, fieldName) {
			trueCounter++
		}
	}

	const precision = 0.05
	const msg = "values of field %s are not uniformly distributed"

	actualRatio := float64(trueCounter) / float64(len(docs))
	assert.GreaterOrEqual(t, actualRatio+precision, ratio, msg, fieldName)
	assert.LessOrEqual(t, actualRatio-precision, ratio, msg, fieldName)
}

func assertUniformlyDistributedFloatFieldRange(t *testing.T, docs []GeneratedDoc, fieldName string, minVal, maxVal float64) {
	vals := make(map[float64]bool, len(docs))
	foundMin := math.Inf(1)
	foundMax := math.Inf(-1)
	for _, doc := range docs {
		val := getFloatField(t, doc.Doc, fieldName)
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
	primaryColInd, secondaryColInd string,
	foreignField string,
	min, max int,
) {
	primaryCol := filterByCollection(docs, primaryColInd)
	secondaryCol := filterByCollection(docs, secondaryColInd)
	assert.GreaterOrEqual(t, len(secondaryCol), len(primaryCol)*min)
	assert.LessOrEqual(t, len(secondaryCol), len(primaryCol)*max)

	secondaryPerPrimary := make(map[string]int)
	for _, d := range secondaryCol {
		docID := getStringField(t, d, foreignField)
		secondaryPerPrimary[docID]++
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

func TestAutoGenerateFromSchema_Simple(t *testing.T) {
	const numUsers = 1000
	schema := `
		type User {
			name: String
			age: Int
			verified: Boolean
			rating: Float
		}`

	docs, err := AutoGenerateFromSDL(schema, WithTypeDemand("User", numUsers))
	assert.NoError(t, err)
	assert.Len(t, docs, numUsers)

	assertUniformlyDistributedStringField(t, docs, "name", 10)
	assertUniformlyDistributedIntFieldRange(t, docs, "age", 0, 10000)
	assertUniformlyDistributedBoolField(t, docs, "verified", 0.5)
	assertUniformlyDistributedFloatFieldRange(t, docs, "rating", 0.0, 1.0)
}

func TestAutoGenerateFromSchema_ConfigIntRange(t *testing.T) {
	const numUsers = 1000
	schema := `
		type User {
			age: Int # min: 1, max: 120
			money: Int # min: -1000, max: 10000
		}`

	docs, err := AutoGenerateFromSDL(schema, WithTypeDemand("User", numUsers))
	assert.NoError(t, err)
	assert.Len(t, docs, numUsers)

	assertUniformlyDistributedIntFieldRange(t, docs, "age", 1, 120)
	assertUniformlyDistributedIntFieldRange(t, docs, "money", -1000, 10000)
}

func TestAutoGenerateFromSchema_ConfigFloatRange(t *testing.T) {
	const numUsers = 1000
	schema := `
		type User {
			rating: Float # min: 1.5, max: 5.0
			product: Float # min: -1.0, max: 1.0
		}`

	docs, err := AutoGenerateFromSDL(schema, WithTypeDemand("User", numUsers))
	assert.NoError(t, err)
	assert.Len(t, docs, numUsers)

	assertUniformlyDistributedFloatFieldRange(t, docs, "rating", 1.5, 5)
	assertUniformlyDistributedFloatFieldRange(t, docs, "product", -1, 1)
}

func TestAutoGenerateFromSchema_ConfigStringLen(t *testing.T) {
	const numUsers = 1000
	schema := `
		type User {
			name: String # len: 8
			email: String # len: 12
		}`

	docs, err := AutoGenerateFromSDL(schema, WithTypeDemand("User", numUsers))
	assert.NoError(t, err)
	assert.Len(t, docs, numUsers)

	assertUniformlyDistributedStringField(t, docs, "name", 8)
	assertUniformlyDistributedStringField(t, docs, "email", 12)
}

func TestAutoGenerateFromSchema_ConfigBoolRatio(t *testing.T) {
	const numUsers = 1000
	schema := `
		type User {
			name: String # len: 8
			verified: Boolean # ratio: 0.2
		}`

	docs, err := AutoGenerateFromSDL(schema, WithTypeDemand("User", numUsers))
	assert.NoError(t, err)
	assert.Len(t, docs, numUsers)

	assertUniformlyDistributedBoolField(t, docs, "verified", 0.2)
}

func TestAutoGenerateFromSchema_IfNoTypeDemandIsGiven_ShouldUseDefault(t *testing.T) {
	schema := `
		type User {
			name: String 
		}`

	docs, err := AutoGenerateFromSDL(schema)
	assert.NoError(t, err)

	const defaultDemand = 10
	assert.Len(t, filterByCollection(docs, "User"), defaultDemand)
}

func TestAutoGenerateFromSchema_RelationOneToOne(t *testing.T) {
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

	docs, err := AutoGenerateFromSDL(schema, WithTypeDemand("User", numUsers))
	assert.NoError(t, err)

	assert.Len(t, filterByCollection(docs, "User"), numUsers)
	assert.Len(t, filterByCollection(docs, "Device"), numUsers)

	assertDocIDsMatch(t, docs, "User", "Device", "owner_id", false)
}

func TestAutoGenerateFromSchema_RelationOneToMany(t *testing.T) {
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

	docs, err := AutoGenerateFromSDL(schema, WithTypeDemand("User", numUsers))
	assert.NoError(t, err)

	assert.Len(t, filterByCollection(docs, "User"), numUsers)
	assert.Len(t, filterByCollection(docs, "Device"), numUsers*2)

	assertDocIDsMatch(t, docs, "User", "Device", "owner_id", true)
}

func TestAutoGenerateFromSchema_RelationOneToManyWithConfiguredNumberOfElements(t *testing.T) {
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

	docs, err := AutoGenerateFromSDL(schema, WithTypeDemand("User", numUsers))
	assert.NoError(t, err)

	assert.Len(t, filterByCollection(docs, "User"), numUsers)

	assertUniformRelationDistribution(t, docs, "User", "Device", "owner_id", minDevicesPerUser, maxDevicesPerUser)

	assertDocIDsMatch(t, docs, "User", "Device", "owner_id", true)
}

func TestAutoGenerateFromSchema_RelationOneToManyToOneWithConfiguredNumberOfElements(t *testing.T) {
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
			device: Device @primary
			OS: String
		}`

	docs, err := AutoGenerateFromSDL(schema, WithTypeDemand("User", numUsers))
	assert.NoError(t, err)

	assert.Len(t, filterByCollection(docs, "User"), numUsers)
	assert.Len(t, filterByCollection(docs, "Device"), numUsers*devicesPerUser)
	assert.Len(t, filterByCollection(docs, "Specs"), numUsers*devicesPerUser)

	assertUniformRelationDistribution(t, docs, "User", "Device", "owner_id", devicesPerUser, devicesPerUser)

	assertDocIDsMatch(t, docs, "User", "Device", "owner_id", true)
	assertDocIDsMatch(t, docs, "Device", "Specs", "device_id", false)
}

func TestAutoGenerateFromSchema_RelationOneToManyToOnePrimaryWithConfiguredNumberOfElements(t *testing.T) {
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

	docs, err := AutoGenerateFromSDL(schema, WithTypeDemand("User", numUsers))
	assert.NoError(t, err)

	assert.Len(t, filterByCollection(docs, "User"), numUsers)
	assert.Len(t, filterByCollection(docs, "Device"), numUsers*devicesPerUser)
	assert.Len(t, filterByCollection(docs, "Specs"), numUsers*devicesPerUser)

	assertUniformRelationDistribution(t, docs, "User", "Device", "owner_id", devicesPerUser, devicesPerUser)

	assertDocIDsMatch(t, docs, "User", "Device", "owner_id", true)
	assertDocIDsMatch(t, docs, "Specs", "Device", "specs_id", false)
}

func TestAutoGenerateFromSchema_RelationOneToManyToManyWithNumDocsForSecondaryType(t *testing.T) {
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
			device: Device @primary
			OS: String
		}
		
		type Component {
			device: Device
			serialNumber: String
		}`

	docs, err := AutoGenerateFromSDL(schema, WithTypeDemand("Device", numDevices))
	assert.NoError(t, err)

	assert.Len(t, filterByCollection(docs, "User"), numDevices/devicesPerUser)
	assert.Len(t, filterByCollection(docs, "Device"), numDevices)
	assert.Len(t, filterByCollection(docs, "Specs"), numDevices)
	assert.Len(t, filterByCollection(docs, "Component"), numDevices*componentsPerDevice)

	assertUniformRelationDistribution(t, docs, "User", "Device", "owner_id", devicesPerUser, devicesPerUser)
	assertUniformRelationDistribution(t, docs, "Device", "Specs", "device_id", 1, 1)
	assertUniformRelationDistribution(t, docs, "Device", "Component", "device_id", componentsPerDevice, componentsPerDevice)

	assertDocIDsMatch(t, docs, "User", "Device", "owner_id", true)
	assertDocIDsMatch(t, docs, "Device", "Specs", "device_id", false)
	assertDocIDsMatch(t, docs, "Device", "Component", "device_id", true)
}

func TestAutoGenerateFromSchema_DemandsForDifferentRelationTrees(t *testing.T) {
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

	docs, err := AutoGenerateFromSDL(
		schema,
		WithTypeDemand("User", numUsers),
		WithTypeDemand("Device", numDevices),
	)
	assert.NoError(t, err)

	assert.Len(t, filterByCollection(docs, "User"), numUsers)
	assert.Len(t, filterByCollection(docs, "Device"), numDevices)
	assert.Len(t, filterByCollection(docs, "Component"), numDevices*componentsPerDevice)

	assertUniformRelationDistribution(t, docs, "Device", "Component", "device_id", componentsPerDevice, componentsPerDevice)

	assertDocIDsMatch(t, docs, "Device", "Component", "device_id", true)
}

func TestAutoGenerateFromSchema_IfTypeDemandedForSameTreeAddsUp_ShouldGenerate(t *testing.T) {
	schema := `
		type User {
			name: String
			devices: [Device]
			orders: [Order] # min: 10, max: 10
		}

		type Device {
			model: String
			owner: User
		}

		type Order {
			id: String
			user: User
		}`

	docs, err := AutoGenerateFromSDL(
		schema,
		WithTypeDemand("Order", 10),
		WithTypeDemand("Device", 30),
	)
	assert.NoError(t, err)

	assert.Len(t, filterByCollection(docs, "User"), 1)
	assert.Len(t, filterByCollection(docs, "Device"), 30)
	assert.Len(t, filterByCollection(docs, "Order"), 10)
}

func TestAutoGenerateFromSchema_IfNoDemandForPrimaryType_ShouldDeduceFromMaxSecondaryDemand(t *testing.T) {
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

	docs, err := AutoGenerateFromSDL(
		schema,
		WithTypeDemand("Order", 10),
		WithTypeDemand("Device", 30),
	)
	assert.NoError(t, err)

	// users len should be equal to max of secondary type demand
	assert.Len(t, filterByCollection(docs, "User"), 30)
	assert.Len(t, filterByCollection(docs, "Device"), 30)
	assert.Len(t, filterByCollection(docs, "Order"), 10)
}

func TestAutoGenerateFromSchema_IfDemand2TypesWithOptions_ShouldAdjust(t *testing.T) {
	const (
		numUsers   = 100
		numDevices = 300
	)
	schema := `
		type User { 
			name: String 
			devices: [Device]
		}
		
		type Device {
			owner: User
			model: String
		}`

	docs, err := AutoGenerateFromSDL(schema,
		WithTypeDemand("User", numUsers),
		WithTypeDemand("Device", numDevices),
	)
	assert.NoError(t, err)

	assert.Len(t, filterByCollection(docs, "User"), numUsers)
	assert.Len(t, filterByCollection(docs, "Device"), numDevices)

	assertDocIDsMatch(t, docs, "User", "Device", "owner_id", true)
}

func TestAutoGenerateFromSchema_IfDemand2TypesWithOptionsAndFieldDemand_ShouldAdjust(t *testing.T) {
	const (
		numUsers   = 100
		numDevices = 300
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

	docs, err := AutoGenerateFromSDL(schema,
		WithTypeDemand("User", numUsers),
		WithTypeDemand("Device", numDevices),
	)
	assert.NoError(t, err)

	assert.Len(t, filterByCollection(docs, "User"), numUsers)
	assert.Len(t, filterByCollection(docs, "Device"), numDevices)

	assertUniformRelationDistribution(t, docs, "User", "Device", "owner_id", 1, 5)

	assertDocIDsMatch(t, docs, "User", "Device", "owner_id", true)
}

func TestAutoGenerateFromSchema_IfDemand2TypesWithRangeOptions_ShouldAdjust(t *testing.T) {
	const (
		numUsers      = 100
		minNumDevices = 100
		maxNumDevices = 500
	)
	schema := `
		type User { 
			name: String 
			devices: [Device]
		}
		
		type Device {
			owner: User
			model: String
		}`

	docs, err := AutoGenerateFromSDL(schema,
		WithTypeDemand("User", numUsers),
		WithTypeDemandRange("Device", minNumDevices, maxNumDevices),
	)
	assert.NoError(t, err)

	assert.Len(t, filterByCollection(docs, "User"), numUsers)
	assert.Len(t, filterByCollection(docs, "Device"), (maxNumDevices+minNumDevices)/2)

	assertUniformRelationDistribution(t, docs, "User", "Device", "owner_id", 1, 5)

	assertDocIDsMatch(t, docs, "User", "Device", "owner_id", true)
}

func TestAutoGenerateFromSchema_ConfigThatCanNotBySupplied(t *testing.T) {
	testCases := []struct {
		name    string
		schema  string
		options []Option
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
			options: []Option{WithTypeDemand("Device", 1)},
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
			options: []Option{WithTypeDemand("Component", 2)},
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
			options: []Option{WithTypeDemand("User", 4)},
		},
		{
			name: "demand for secondary types from the same tree can not be satisfied",
			schema: `
				type User { 
					name: String 
					devices: [Device] # min: 0, max: 29
					orders: [Order] # min: 10, max: 10
				}
				
				type Device {
					model: String
					owner: User
				}
				
				type Order {
					id: String
					user: User
				}`,
			options: []Option{WithTypeDemand("Order", 10), WithTypeDemand("Device", 30)},
		},
		{
			schema: `
				type User { 
					name: String 
					device: Device
				}
				
				type Device {
					model: String
					owner: User
				}`,
			options: []Option{WithTypeDemand("User", 10), WithTypeDemand("Device", 30)},
		},
		{
			schema: `
				type User { 
					name: String 
					device: Device
					orders: Order
				}
				
				type Device {
					model: String
					owner: User
				}
				
				type Order {
					id: String
					user: User
				}`,
			options: []Option{WithTypeDemand("Order", 10), WithTypeDemand("Device", 30)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := AutoGenerateFromSDL(tc.schema, tc.options...)

			assert.ErrorContains(t, err, errCanNotSupplyTypeDemand)
		})
	}
}

func TestAutoGenerateFromSchema_InvalidConfig(t *testing.T) {
	testCases := []struct {
		name   string
		schema string
	}{
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
		},
		{
			name: "int value: max is less than min",
			schema: `
				type User { 
					age: Int # min: 10, max: 2
				}`,
		},
		{
			name: "int value: missing min",
			schema: `
				type User { 
					age: Int # max: 2
				}`,
		},
		{
			name: "int value: missing max",
			schema: `
				type User { 
					age: Int # min: 2
				}`,
		},
		{
			name: "float value: max is less than min",
			schema: `
				type User { 
					rating: Float # min: 10.5, max: 2.5
				}`,
		},
		{
			name: "float value: missing min",
			schema: `
				type User { 
					rating: Float # max: 2.5
				}`,
		},
		{
			name: "float value: missing max",
			schema: `
				type User { 
					rating: Float # min: 2.5
				}`,
		},
		{
			name: "min/max on not number",
			schema: `
				type User { 
					verified: Boolean # min: 2, max: 8
				}`,
		},
		{
			name: "string: zero len",
			schema: `
				type User { 
					name: String # len: 0
				}`,
		},
		{
			name: "string: negative len",
			schema: `
				type User { 
					name: String # len: -2
				}`,
		},
		{
			name: "string: non-int len",
			schema: `
				type User { 
					name: String # len: 8.5
				}`,
		},
		{
			name: "len on non-string",
			schema: `
				type User { 
					age: Int # len: 8
				}`,
		},
		{
			name: "string: non-int len",
			schema: `
				type User { 
					name: String # len: 8.5
				}`,
		},
		{
			name: "bool: negative ratio",
			schema: `
				type User { 
					verified: Boolean # ratio: -0.1
				}`,
		},
		{
			name: "bool: ratio greater than 1",
			schema: `
				type User { 
					verified: Boolean # ratio: 1.1
				}`,
		},
		{
			name: "bool: ratio with non-float type",
			schema: `
				type User { 
					verified: Boolean # ratio: "0.5"
				}`,
		},
		{
			name: "ratio on non-bool field",
			schema: `
				type User { 
					age: Int # ratio: 0.5
				}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := AutoGenerateFromSDL(tc.schema, WithTypeDemand("User", 4))

			assert.ErrorContains(t, err, errInvalidConfiguration)
		})
	}
}

func TestAutoGenerateFromSchema_CustomFieldValueGenerator(t *testing.T) {
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

	docs, err := AutoGenerateFromSDL(schema,
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
		actualAgeVal := getIntField(t, doc.Doc, "age")
		assert.Equal(t, ageVal, actualAgeVal)
	}
}

func TestAutoGenerateFromSchema_IfOptionOverlapsSchemaConfig_ItShouldOverwrite(t *testing.T) {
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

	docs, err := AutoGenerateFromSDL(
		schema,
		WithTypeDemand("User", numUsers),
		WithFieldRange("User", "devices", 3, 3),
		WithFieldRange("User", "age", 30, 40),
		WithFieldRange("User", "rating", 1.0, 2.0),
		WithFieldLen("User", "name", 6),
	)
	assert.NoError(t, err)

	userDocs := filterByCollection(docs, "User")
	assert.Len(t, userDocs, numUsers)
	assert.Len(t, filterByCollection(docs, "Device"), numUsers*3)

	for _, userDoc := range userDocs {
		actualAgeVal := getIntField(t, userDoc, "age")
		assert.GreaterOrEqual(t, actualAgeVal, 30)
		assert.LessOrEqual(t, actualAgeVal, 40)

		actualRatingVal := getFloatField(t, userDoc, "rating")
		assert.GreaterOrEqual(t, actualRatingVal, 1.0)
		assert.LessOrEqual(t, actualRatingVal, 2.0)

		actualNameVal := getStringField(t, userDoc, "name")
		assert.Len(t, actualNameVal, 6)
	}
}

func TestAutoGenerateFromSchema_WithRandomSeed_ShouldBeDeterministic(t *testing.T) {
	schema := `
		type User { 
			devices: [Device] # min: 0, max: 5
			age: Int 
			rating: Float 
		}
		
		type Device {
			model: String
			owner: User
		}`

	demandOpt := WithTypeDemand("User", 5)
	seed := time.Now().UnixNano()

	docs1, err := AutoGenerateFromSDL(schema, demandOpt, WithRandomSeed(seed))
	assert.NoError(t, err)

	docs2, err := AutoGenerateFromSDL(schema, demandOpt, WithRandomSeed(seed))
	assert.NoError(t, err)

	docs3, err := AutoGenerateFromSDL(schema, demandOpt, WithRandomSeed(time.Now().UnixNano()))
	assert.NoError(t, err)

	assert.Equal(t, docs1, docs2)
	assert.NotEqual(t, docs1, docs3)
}

func TestAutoGenerateFromSchema_InvalidOption(t *testing.T) {
	const schema = `
		type User {
			name: String
		}`
	testCases := []struct {
		name    string
		options []Option
	}{
		{
			name:    "type demand for non-existing type",
			options: []Option{WithTypeDemand("Invalid", 10)},
		},
		{
			name:    "type demand range for non-existing type",
			options: []Option{WithTypeDemandRange("Invalid", 0, 10)},
		},
		{
			name:    "field len for non-existing type",
			options: []Option{WithFieldLen("Invalid", "name", 10)},
		},
		{
			name:    "field len for non-existing field",
			options: []Option{WithFieldLen("User", "invalid", 10)},
		},
		{
			name:    "field range for non-existing type",
			options: []Option{WithFieldRange("Invalid", "name", 0, 10)},
		},
		{
			name:    "field range for non-existing field",
			options: []Option{WithFieldRange("User", "invalid", 0, 10)},
		},
		{
			name:    "field generator for non-existing type",
			options: []Option{WithFieldGenerator("Invalid", "name", func(i int, next func() any) any { return "" })},
		},
		{
			name:    "field generator for non-existing field",
			options: []Option{WithFieldGenerator("User", "invalid", func(i int, next func() any) any { return "" })},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := AutoGenerateFromSDL(schema, tc.options...)

			assert.ErrorContains(t, err, errInvalidConfiguration)
		})
	}
}

func TestAutoGenerate_IfCollectionDefinitionIsIncomplete_ReturnError(t *testing.T) {
	getValidDefs := func() []client.CollectionDefinition {
		return []client.CollectionDefinition{
			{
				Description: client.CollectionDescription{
					Name: "User",
					ID:   0,
				},
				Schema: client.SchemaDescription{
					Name: "User",
					Fields: []client.FieldDescription{
						{
							Name: "name",
							Kind: client.FieldKind_INT,
						},
						{
							Name:         "device",
							Kind:         client.FieldKind_FOREIGN_OBJECT,
							Schema:       "Device",
							RelationType: client.Relation_Type_ONE | client.Relation_Type_ONEONE,
						},
					},
				},
			},
			{
				Description: client.CollectionDescription{
					Name: "Device",
					ID:   1,
				},
				Schema: client.SchemaDescription{
					Name: "Device",
					Fields: []client.FieldDescription{
						{
							Name: "model",
							Kind: client.FieldKind_STRING,
						},
						{
							Name:   "owner",
							Kind:   client.FieldKind_FOREIGN_OBJECT,
							Schema: "User",
							RelationType: client.Relation_Type_ONE |
								client.Relation_Type_ONEONE |
								client.Relation_Type_Primary,
						},
					},
				},
			},
		}
	}

	testCases := []struct {
		name       string
		changeDefs func(defs []client.CollectionDefinition)
	}{
		{
			name: "description name is empty",
			changeDefs: func(defs []client.CollectionDefinition) {
				defs[0].Description.Name = ""
			},
		},
		{
			name: "schema name is empty",
			changeDefs: func(defs []client.CollectionDefinition) {
				defs[0].Schema.Name = ""
			},
		},
		{
			name: "field name is empty",
			changeDefs: func(defs []client.CollectionDefinition) {
				defs[0].Schema.Fields[0].Name = ""
			},
		},
		{
			name: "not matching names",
			changeDefs: func(defs []client.CollectionDefinition) {
				defs[0].Schema.Name = "Device"
			},
		},
		{
			name: "ids are not enumerated",
			changeDefs: func(defs []client.CollectionDefinition) {
				defs[1].Description.ID = 0
			},
		},
		{
			name: "relation field is missing schema name",
			changeDefs: func(defs []client.CollectionDefinition) {
				defs[1].Schema.Fields[1].Schema = ""
			},
		},
		{
			name: "relation field references unknown schema",
			changeDefs: func(defs []client.CollectionDefinition) {
				defs[1].Schema.Fields[1].Schema = "Unknown"
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defs := getValidDefs()
			tc.changeDefs(defs)
			_, err := AutoGenerate(defs)

			assert.ErrorContains(t, err, errIncompleteColDefinition)
		})
	}
}

func TestAutoGenerate_IfColDefinitionsAreValid_ShouldGenerate(t *testing.T) {
	const (
		numUsers = 20
	)

	defs := []client.CollectionDefinition{
		{
			Description: client.CollectionDescription{
				Name: "User",
				ID:   0,
			},
			Schema: client.SchemaDescription{
				Name: "User",
				Fields: []client.FieldDescription{
					{
						Name: "name",
						Kind: client.FieldKind_STRING,
					},
					{
						Name: "age",
						Kind: client.FieldKind_INT,
					},
					{
						Name: "rating",
						Kind: client.FieldKind_FLOAT,
					},
					{
						Name:         "devices",
						Kind:         client.FieldKind_FOREIGN_OBJECT_ARRAY,
						Schema:       "Device",
						RelationType: client.Relation_Type_MANY | client.Relation_Type_ONEMANY,
					},
				},
			},
		},
		{
			Description: client.CollectionDescription{
				Name: "Device",
				ID:   1,
			},
			Schema: client.SchemaDescription{
				Name: "Device",
				Fields: []client.FieldDescription{
					{
						Name: "model",
						Kind: client.FieldKind_STRING,
					},
					{
						Name:   "owner_id",
						Kind:   client.FieldKind_FOREIGN_OBJECT,
						Schema: "User",
						RelationType: client.Relation_Type_ONE |
							client.Relation_Type_ONEMANY |
							client.Relation_Type_Primary,
					},
				},
			},
		},
	}
	docs, err := AutoGenerate(
		defs,
		WithTypeDemand("User", numUsers),
		WithFieldRange("User", "devices", 3, 3),
		WithFieldRange("User", "age", 30, 40),
		WithFieldRange("User", "rating", 1.0, 2.0),
		WithFieldLen("User", "name", 6),
	)
	assert.NoError(t, err)

	userDocs := filterByCollection(docs, "User")
	assert.Len(t, userDocs, numUsers)
	assert.Len(t, filterByCollection(docs, "Device"), numUsers*3)

	for _, userDoc := range userDocs {
		actualAgeVal := getIntField(t, userDoc, "age")
		assert.GreaterOrEqual(t, actualAgeVal, 30)
		assert.LessOrEqual(t, actualAgeVal, 40)

		actualRatingVal := getFloatField(t, userDoc, "rating")
		assert.GreaterOrEqual(t, actualRatingVal, 1.0)
		assert.LessOrEqual(t, actualRatingVal, 2.0)

		actualNameVal := getStringField(t, userDoc, "name")
		assert.Len(t, actualNameVal, 6)
	}
}
