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
	list := []string{}
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
		foreignValues = removeDuplicateStr(foreignValues)
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

func TestAutoGenerateDocs_Simple(t *testing.T) {
	const numDocs = 1000
	schema := `
		type User {
			name: String
			age: Int
			verified: Boolean
			rating: Float
		}`

	docs := AutoGenerateDocs(schema, numDocs)
	assert.Len(t, docs, numDocs)

	assertUniformlyDistributedStringField(t, docs, "name", 10)
	assertUniformlyDistributedIntFieldRange(t, docs, "age", 0, 10000)
	assertUniformlyDistributedBoolField(t, docs, "verified")
	assertUniformlyDistributedFloatFieldRange(t, docs, "rating", 0.0, 1.0)
}

func TestAutoGenerateDocs_ConfigIntRange(t *testing.T) {
	const numDocs = 1000
	schema := `
		type User {
			age: Int # min: 1, max: 120
			money: Int # min: -1000, max: 10000
		}`

	docs := AutoGenerateDocs(schema, numDocs)
	assert.Len(t, docs, numDocs)

	assertUniformlyDistributedIntFieldRange(t, docs, "age", 1, 120)
	assertUniformlyDistributedIntFieldRange(t, docs, "money", -1000, 10000)
}

func TestAutoGenerateDocs_ConfigFloatRange(t *testing.T) {
	const numDocs = 1000
	schema := `
		type User {
			rating: Float # min: 1.5, max: 5.0
			product: Float # min: -1.0, max: 1.0
		}`

	docs := AutoGenerateDocs(schema, numDocs)
	assert.Len(t, docs, numDocs)

	assertUniformlyDistributedFloatFieldRange(t, docs, "rating", 1.5, 5)
	assertUniformlyDistributedFloatFieldRange(t, docs, "product", -1, 1)
}

func TestAutoGenerateDocs_ConfigStringLen(t *testing.T) {
	const numDocs = 1000
	schema := `
		type User {
			name: String # len: 8
			email: String # len: 12
		}`

	docs := AutoGenerateDocs(schema, numDocs)
	assert.Len(t, docs, numDocs)

	assertUniformlyDistributedStringField(t, docs, "name", 8)
	assertUniformlyDistributedStringField(t, docs, "email", 12)
}

func TestAutoGenerateDocs_RelationOneToOne(t *testing.T) {
	const numDocs = 10
	schema := `
		type User {
			name: String 
			device: Device
		}
		
		type Device {
			owner: User
			model: String
		}`

	docs := AutoGenerateDocs(schema, numDocs)

	assert.Len(t, filterByCollection(docs, 0), numDocs)
	assert.Len(t, filterByCollection(docs, 1), numDocs)

	assertDocKeysMatch(t, docs, 0, 1, "owner_id", false)
}

func TestAutoGenerateDocs_RelationOneToMany(t *testing.T) {
	const numDocs = 10
	schema := `
		type User { 
			name: String 
			devices: [Device] 
		}
		
		type Device {
			owner: User
			model: String
		}`

	docs := AutoGenerateDocs(schema, numDocs)

	assert.Len(t, filterByCollection(docs, 0), numDocs)
	assert.Len(t, filterByCollection(docs, 1), numDocs*2)

	assertDocKeysMatch(t, docs, 0, 1, "owner_id", true)
}
