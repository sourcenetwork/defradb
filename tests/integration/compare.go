// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package tests

import (
	"encoding/json"

	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/assert"
)

// AnyOf may be used as `Results` field where the value may
// be one of several values, yet the value of that field must be the same
// across all nodes due to strong eventual consistency.
type AnyOf []any

// resultsAreAnyOf returns true if any of the expected results are of equal value.
//
// NOTE: Values of type json.Number and immutable.Option will be reduced to their underlying types.
func resultsAreAnyOf(expected AnyOf, actual any) bool {
	for _, v := range expected {
		if resultsAreEqual(v, actual) {
			return true
		}
	}
	return false
}

// resultsAreEqual returns true if the expected and actual results are of equal value.
//
// NOTE: Values of type json.Number and immutable.Option will be reduced to their underlying types.
func resultsAreEqual(expected any, actual any) bool {
	switch expectedVal := expected.(type) {
	case map[string]any:
		if len(expectedVal) == 0 && actual == nil {
			return true
		}
		actualVal, ok := actual.(map[string]any)
		if !ok {
			return assert.ObjectsAreEqualValues(expected, actual)
		}
		if len(expectedVal) != len(actualVal) {
			return false
		}
		for k, v := range expectedVal {
			if !resultsAreEqual(v, actualVal[k]) {
				return false
			}
		}
		return true
	case []int64:
		if len(expectedVal) == 0 && actual == nil {
			return true
		}
		actualVal, ok := actual.([]any)
		if !ok {
			return assert.ObjectsAreEqualValues(expected, actual)
		}
		if len(expectedVal) != len(actualVal) {
			return false
		}
		for i, v := range expectedVal {
			if !resultsAreEqual(v, actualVal[i]) {
				return false
			}
		}
		return true
	case []uint64:
		if len(expectedVal) == 0 && actual == nil {
			return true
		}
		actualVal, ok := actual.([]any)
		if !ok {
			return assert.ObjectsAreEqualValues(expected, actual)
		}
		if len(expectedVal) != len(actualVal) {
			return false
		}
		for i, v := range expectedVal {
			if !resultsAreEqual(v, actualVal[i]) {
				return false
			}
		}
		return true
	case []float64:
		if len(expectedVal) == 0 && actual == nil {
			return true
		}
		actualVal, ok := actual.([]any)
		if !ok {
			return assert.ObjectsAreEqualValues(expected, actual)
		}
		if len(expectedVal) != len(actualVal) {
			return false
		}
		for i, v := range expectedVal {
			if !resultsAreEqual(v, actualVal[i]) {
				return false
			}
		}
		return true
	case []string:
		if len(expectedVal) == 0 && actual == nil {
			return true
		}
		actualVal, ok := actual.([]any)
		if !ok {
			return assert.ObjectsAreEqualValues(expected, actual)
		}
		if len(expectedVal) != len(actualVal) {
			return false
		}
		for i, v := range expectedVal {
			if !resultsAreEqual(v, actualVal[i]) {
				return false
			}
		}
		return true
	case []bool:
		if len(expectedVal) == 0 && actual == nil {
			return true
		}
		actualVal, ok := actual.([]any)
		if !ok {
			return assert.ObjectsAreEqualValues(expected, actual)
		}
		if len(expectedVal) != len(actualVal) {
			return false
		}
		for i, v := range expectedVal {
			if !resultsAreEqual(v, actualVal[i]) {
				return false
			}
		}
		return true
	case []any:
		if len(expectedVal) == 0 && actual == nil {
			return true
		}
		actualVal, ok := actual.([]any)
		if !ok {
			return assert.ObjectsAreEqualValues(expected, actual)
		}
		if len(expectedVal) != len(actualVal) {
			return false
		}
		for i, v := range expectedVal {
			if !resultsAreEqual(v, actualVal[i]) {
				return false
			}
		}
		return true
	case []map[string]any:
		if len(expectedVal) == 0 && actual == nil {
			return true
		}
		actualVal, ok := actual.([]any)
		if !ok {
			return assert.ObjectsAreEqualValues(expected, actual)
		}
		if len(expectedVal) != len(actualVal) {
			return false
		}
		for i, v := range expectedVal {
			if !resultsAreEqual(v, actualVal[i]) {
				return false
			}
		}
		return true
	case uint64, uint32, uint16, uint8, uint, int64, int32, int16, int8, int:
		jsonNum, ok := actual.(json.Number)
		if !ok {
			return assert.ObjectsAreEqualValues(expected, actual)
		}
		actualVal, err := jsonNum.Int64()
		if err != nil {
			return false
		}
		return assert.ObjectsAreEqualValues(expected, actualVal)
	case float32, float64:
		jsonNum, ok := actual.(json.Number)
		if !ok {
			return assert.ObjectsAreEqualValues(expected, actual)
		}
		actualVal, err := jsonNum.Float64()
		if err != nil {
			return false
		}
		return assert.ObjectsAreEqualValues(expected, actualVal)
	case []immutable.Option[float64]:
		if len(expectedVal) == 0 && actual == nil {
			return true
		}
		actualVal, ok := actual.([]any)
		if !ok {
			return assert.ObjectsAreEqualValues(expected, actual)
		}
		if len(expectedVal) != len(actualVal) {
			return false
		}
		for i, v := range expectedVal {
			if !resultsAreEqual(v, actualVal[i]) {
				return false
			}
		}
		return true
	case []immutable.Option[uint64]:
		if len(expectedVal) == 0 && actual == nil {
			return true
		}
		actualVal, ok := actual.([]any)
		if !ok {
			return assert.ObjectsAreEqualValues(expected, actual)
		}
		if len(expectedVal) != len(actualVal) {
			return false
		}
		for i, v := range expectedVal {
			if !resultsAreEqual(v, actualVal[i]) {
				return false
			}
		}
		return true
	case []immutable.Option[int64]:
		if len(expectedVal) == 0 && actual == nil {
			return true
		}
		actualVal, ok := actual.([]any)
		if !ok {
			return assert.ObjectsAreEqualValues(expected, actual)
		}
		if len(expectedVal) != len(actualVal) {
			return false
		}
		for i, v := range expectedVal {
			if !resultsAreEqual(v, actualVal[i]) {
				return false
			}
		}
		return true
	case []immutable.Option[bool]:
		if len(expectedVal) == 0 && actual == nil {
			return true
		}
		actualVal, ok := actual.([]any)
		if !ok {
			return assert.ObjectsAreEqualValues(expected, actual)
		}
		if len(expectedVal) != len(actualVal) {
			return false
		}
		for i, v := range expectedVal {
			if !resultsAreEqual(v, actualVal[i]) {
				return false
			}
		}
		return true
	case []immutable.Option[string]:
		if len(expectedVal) == 0 && actual == nil {
			return true
		}
		actualVal, ok := actual.([]any)
		if !ok {
			return assert.ObjectsAreEqualValues(expected, actual)
		}
		if len(expectedVal) != len(actualVal) {
			return false
		}
		for i, v := range expectedVal {
			if !resultsAreEqual(v, actualVal[i]) {
				return false
			}
		}
		return true
	case immutable.Option[float64]:
		if expectedVal.HasValue() {
			expected = expectedVal.Value()
		} else {
			expected = nil
		}
		return resultsAreEqual(expected, actual)
	case immutable.Option[uint64]:
		if expectedVal.HasValue() {
			expected = expectedVal.Value()
		} else {
			expected = nil
		}
		return resultsAreEqual(expected, actual)
	case immutable.Option[int64]:
		if expectedVal.HasValue() {
			expected = expectedVal.Value()
		} else {
			expected = nil
		}
		return resultsAreEqual(expected, actual)
	case immutable.Option[bool]:
		if expectedVal.HasValue() {
			expected = expectedVal.Value()
		} else {
			expected = nil
		}
		return resultsAreEqual(expected, actual)
	case immutable.Option[string]:
		if expectedVal.HasValue() {
			expected = expectedVal.Value()
		} else {
			expected = nil
		}
		return resultsAreEqual(expected, actual)
	default:
		return assert.ObjectsAreEqualValues(expected, actual)
	}
}
