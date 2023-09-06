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
	"testing"

	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/assert"
)

// AnyOf may be used as `Results` field where the value may
// be one of several values, yet the value of that field must be the same
// across all nodes due to strong eventual consistency.
type AnyOf []any

// assertResultsAnyOf asserts that actual result is equal to at least one of the expected results.
//
// The comparison is relaxed when using client types other than goClientType.
func assertResultsAnyOf(t *testing.T, client ClientType, expected AnyOf, actual any, msgAndArgs ...any) {
	switch client {
	case httpClientType, cliClientType:
		if !resultsAreAnyOf(expected, actual) {
			assert.Contains(t, expected, actual, msgAndArgs...)
		}
	default:
		assert.Contains(t, expected, actual, msgAndArgs...)
	}
}

// assertResultsEqual asserts that actual result is equal to the expected result.
//
// The comparison is relaxed when using client types other than goClientType.
func assertResultsEqual(t *testing.T, client ClientType, expected any, actual any, msgAndArgs ...any) {
	switch client {
	case httpClientType, cliClientType:
		if !resultsAreEqual(expected, actual) {
			assert.EqualValues(t, expected, actual, msgAndArgs...)
		}
	default:
		assert.EqualValues(t, expected, actual, msgAndArgs...)
	}
}

// resultsAreAnyOf returns true if any of the expected results are of equal value.
//
// Values of type json.Number and immutable.Option will be reduced to their underlying types.
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
// Values of type json.Number and immutable.Option will be reduced to their underlying types.
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
	case immutable.Option[float64]:
		return resultOptionsAreEqual(expectedVal, actual)
	case immutable.Option[uint64]:
		return resultOptionsAreEqual(expectedVal, actual)
	case immutable.Option[int64]:
		return resultOptionsAreEqual(expectedVal, actual)
	case immutable.Option[bool]:
		return resultOptionsAreEqual(expectedVal, actual)
	case immutable.Option[string]:
		return resultOptionsAreEqual(expectedVal, actual)
	case []int64:
		return resultArraysAreEqual(expectedVal, actual)
	case []uint64:
		return resultArraysAreEqual(expectedVal, actual)
	case []float64:
		return resultArraysAreEqual(expectedVal, actual)
	case []string:
		return resultArraysAreEqual(expectedVal, actual)
	case []bool:
		return resultArraysAreEqual(expectedVal, actual)
	case []any:
		return resultArraysAreEqual(expectedVal, actual)
	case []map[string]any:
		return resultArraysAreEqual(expectedVal, actual)
	case []immutable.Option[float64]:
		return resultArraysAreEqual(expectedVal, actual)
	case []immutable.Option[uint64]:
		return resultArraysAreEqual(expectedVal, actual)
	case []immutable.Option[int64]:
		return resultArraysAreEqual(expectedVal, actual)
	case []immutable.Option[bool]:
		return resultArraysAreEqual(expectedVal, actual)
	case []immutable.Option[string]:
		return resultArraysAreEqual(expectedVal, actual)
	default:
		return assert.ObjectsAreEqualValues(expected, actual)
	}
}

// resultArraysAreEqual returns true if the value of the expected immutable.Option
// and actual result are of equal value.
//
// Values of type json.Number and immutable.Option will be reduced to their underlying types.
func resultOptionsAreEqual[S any](expected immutable.Option[S], actual any) bool {
	var expectedVal any
	if expected.HasValue() {
		expectedVal = expected.Value()
	}
	return resultsAreEqual(expectedVal, actual)
}

// resultArraysAreEqual returns true if the array of expected results and actual results
// are of equal value.
//
// Values of type json.Number and immutable.Option will be reduced to their underlying types.
func resultArraysAreEqual[S any](expected []S, actual any) bool {
	if len(expected) == 0 && actual == nil {
		return true
	}
	actualVal, ok := actual.([]any)
	if !ok {
		return assert.ObjectsAreEqualValues(expected, actual)
	}
	if len(expected) != len(actualVal) {
		return false
	}
	for i, v := range expected {
		if !resultsAreEqual(v, actualVal[i]) {
			return false
		}
	}
	return true
}
