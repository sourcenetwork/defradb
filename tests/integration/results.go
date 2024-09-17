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
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
)

// AnyOf may be used as `Results` field where the value may
// be one of several values, yet the value of that field must be the same
// across all nodes due to strong eventual consistency.
type AnyOf []any

// assertResultsAnyOf asserts that actual result is equal to at least one of the expected results.
//
// The comparison is relaxed when using client types other than goClientType.
func assertResultsAnyOf(t testing.TB, client ClientType, expected AnyOf, actual any, msgAndArgs ...any) {
	switch client {
	case HTTPClientType, CLIClientType:
		if !areResultsAnyOf(expected, actual) {
			assert.Contains(t, expected, actual, msgAndArgs...)
		}
	default:
		assert.Contains(t, expected, actual, msgAndArgs...)
	}
}

// assertResultsEqual asserts that actual result is equal to the expected result.
//
// The comparison is relaxed when using client types other than goClientType.
func assertResultsEqual(t testing.TB, client ClientType, expected any, actual any, msgAndArgs ...any) {
	switch client {
	case HTTPClientType, CLIClientType:
		if !areResultsEqual(expected, actual) {
			assert.EqualValues(t, expected, actual, msgAndArgs...)
		}
	default:
		assert.EqualValues(t, expected, actual, msgAndArgs...)
	}
}

// areResultsAnyOf returns true if any of the expected results are of equal value.
//
// Values of type json.Number and immutable.Option will be reduced to their underlying types.
func areResultsAnyOf(expected AnyOf, actual any) bool {
	for _, v := range expected {
		if areResultsEqual(v, actual) {
			return true
		}
	}
	return false
}

// areResultsEqual returns true if the expected and actual results are of equal value.
//
// Values of type json.Number and immutable.Option will be reduced to their underlying types.
func areResultsEqual(expected any, actual any) bool {
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
			if !areResultsEqual(v, actualVal[k]) {
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
		return areResultOptionsEqual(expectedVal, actual)
	case immutable.Option[uint64]:
		return areResultOptionsEqual(expectedVal, actual)
	case immutable.Option[int64]:
		return areResultOptionsEqual(expectedVal, actual)
	case immutable.Option[bool]:
		return areResultOptionsEqual(expectedVal, actual)
	case immutable.Option[string]:
		return areResultOptionsEqual(expectedVal, actual)
	case []uint8:
		return areResultsEqual(base64.StdEncoding.EncodeToString(expectedVal), actual)
	case []int64:
		return areResultArraysEqual(expectedVal, actual)
	case []uint64:
		return areResultArraysEqual(expectedVal, actual)
	case []float64:
		return areResultArraysEqual(expectedVal, actual)
	case []string:
		return areResultArraysEqual(expectedVal, actual)
	case []bool:
		return areResultArraysEqual(expectedVal, actual)
	case []any:
		return areResultArraysEqual(expectedVal, actual)
	case []map[string]any:
		return areResultArraysEqual(expectedVal, actual)
	case []immutable.Option[float64]:
		return areResultArraysEqual(expectedVal, actual)
	case []immutable.Option[uint64]:
		return areResultArraysEqual(expectedVal, actual)
	case []immutable.Option[int64]:
		return areResultArraysEqual(expectedVal, actual)
	case []immutable.Option[bool]:
		return areResultArraysEqual(expectedVal, actual)
	case []immutable.Option[string]:
		return areResultArraysEqual(expectedVal, actual)
	case time.Time:
		return areResultsEqual(expectedVal.Format(time.RFC3339Nano), actual)
	default:
		return assert.ObjectsAreEqualValues(expected, actual)
	}
}

// areResultOptionsEqual returns true if the value of the expected immutable.Option
// and actual result are of equal value.
//
// Values of type json.Number and immutable.Option will be reduced to their underlying types.
func areResultOptionsEqual[S any](expected immutable.Option[S], actual any) bool {
	var expectedVal any
	if expected.HasValue() {
		expectedVal = expected.Value()
	}
	return areResultsEqual(expectedVal, actual)
}

// areResultArraysEqual returns true if the array of expected results and actual results
// are of equal value.
//
// Values of type json.Number and immutable.Option will be reduced to their underlying types.
func areResultArraysEqual[S any](expected []S, actual any) bool {
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
		if !areResultsEqual(v, actualVal[i]) {
			return false
		}
	}
	return true
}

func assertCollectionDescriptions(
	s *state,
	expected []client.CollectionDescription,
	actual []client.CollectionDescription,
) {
	require.Equal(s.t, len(expected), len(actual))

	for i, expected := range expected {
		actual := actual[i]
		if expected.ID != 0 {
			require.Equal(s.t, expected.ID, actual.ID)
		}
		if expected.RootID != 0 {
			require.Equal(s.t, expected.RootID, actual.RootID)
		}
		if expected.SchemaVersionID != "" {
			require.Equal(s.t, expected.SchemaVersionID, actual.SchemaVersionID)
		}

		require.Equal(s.t, expected.Name, actual.Name)
		require.Equal(s.t, expected.IsMaterialized, actual.IsMaterialized)

		if expected.Indexes != nil || len(actual.Indexes) != 0 {
			// Dont bother asserting this if the expected is nil and the actual is nil/empty.
			// This is to save each test action from having to bother declaring an empty slice (if there are no indexes)
			require.Equal(s.t, expected.Indexes, actual.Indexes)
		}

		if expected.Sources != nil {
			// Dont bother asserting this if the expected is nil and the actual is nil/empty.
			// This is to save each test action from having to bother declaring an empty slice (if there are no sources)
			require.Equal(s.t, expected.Sources, actual.Sources)
		}

		if expected.Fields != nil {
			require.Equal(s.t, expected.Fields, actual.Fields)
		}
	}
}
