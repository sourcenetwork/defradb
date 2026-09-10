// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package action

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/state"
)

// Asserts as to whether an error has been raised as expected (or not). If an expected
// error has been raised it will return true, returns false in all other cases.
// normalizeIndexes returns a copy of the slice with each index's Kind/Unique made consistent, so a
// test expectation written in either style (only Unique, or only Kind) compares against the actual.
func normalizeIndexes(indexes []client.IndexDescription) []client.IndexDescription {
	out := make([]client.IndexDescription, len(indexes))
	for i, idx := range indexes {
		out[i] = idx.Normalize()
	}
	return out
}

func assertError(t testing.TB, err error, expectedError string) bool {
	if err == nil {
		return false
	}

	if expectedError == "" {
		require.NoError(t, err)
		return false
	} else {
		if !strings.Contains(err.Error(), expectedError) {
			// Must be require instead of assert, otherwise will show a fake "error not raised".
			require.ErrorIs(t, err, errors.New(expectedError))
			return false
		}
		return true
	}
}

func assertExpectedErrorRaised(t testing.TB, expectedError string, wasRaised bool) {
	AssertExpectedErrorRaised(t, expectedError, wasRaised)
}

// AssertExpectedErrorRaised asserts that an expected error was raised.
func AssertExpectedErrorRaised(t testing.TB, expectedError string, wasRaised bool) {
	if expectedError != "" && !wasRaised {
		assert.Fail(t, "Expected an error however none was raised.")
	}
}

func assertCollectionVersions(
	s *state.State,
	expected []client.CollectionVersion,
	actual []client.CollectionVersion,
) {
	require.Equal(s.T, len(expected), len(actual))

	for i, expected := range expected {
		actual := actual[i]
		require.Equal(s.T, expected.Name, actual.Name)

		if expected.CollectionSet.HasValue() {
			require.Equal(s.T, expected.CollectionSet.Value().CollectionSetID, actual.CollectionSet.Value().CollectionSetID)
			require.Equal(s.T, expected.CollectionSet.Value().RelativeID, actual.CollectionSet.Value().RelativeID)
		}

		if expected.VersionID != "" {
			require.Equal(s.T, expected.VersionID, actual.VersionID)
		}
		if expected.CollectionID != "" {
			require.Equal(s.T, expected.CollectionID, actual.CollectionID)
		}

		require.Equal(s.T, expected.IsMaterialized, actual.IsMaterialized)
		require.Equal(s.T, expected.IsBranchable, actual.IsBranchable)
		require.Equal(s.T, expected.IsActive, actual.IsActive)

		if expected.Indexes != nil {
			// Dont bother asserting this if the expected is nil and the actual is nil/empty.
			// This is to save each test action from having to bother declaring an empty slice (if there are no indexes)
			// Normalize both sides so an expectation written in the old style (Unique only, no Kind)
			// compares equal to the resolved actual (Kind populated).
			require.Equal(s.T, normalizeIndexes(expected.Indexes), normalizeIndexes(actual.Indexes))
		}

		require.Equal(s.T, expected.PreviousVersion.HasValue(), actual.PreviousVersion.HasValue())
		if expected.PreviousVersion.HasValue() {
			require.Equal(
				s.T,
				expected.PreviousVersion.Value().SourceCollectionID,
				actual.PreviousVersion.Value().SourceCollectionID,
			)
			require.Equal(
				s.T,
				expected.PreviousVersion.Value().Transform.HasValue(),
				actual.PreviousVersion.Value().Transform.HasValue(),
			)

			if expected.PreviousVersion.Value().Transform.HasValue() {
				// Dont bother asserting this by default, the transform object is too complex to bother with in most cases.
				require.Equal(
					s.T,
					expected.PreviousVersion.Value().Transform.Value(),
					actual.PreviousVersion.Value().Transform.Value(),
				)
			}
		}

		if expected.Query.HasValue() {
			// Dont bother asserting this by default, the query object is to complex to bother with in most cases.
			require.Equal(s.T, expected.Query, actual.Query)
		}

		if expected.Fields != nil {
			require.Equal(s.T, len(expected.Fields), len(actual.Fields))
			for i := range expected.Fields {
				expectedField := expected.Fields[i]
				actualField := actual.Fields[i]

				require.Equal(s.T, expectedField.Name, actualField.Name)
				if expectedField.FieldID != "" {
					require.Equal(s.T, expectedField.FieldID, actualField.FieldID)
				}
				require.Equal(s.T, expectedField.IsPrimary, actualField.IsPrimary)
				require.Equal(s.T, expectedField.Kind, actualField.Kind)
				require.Equal(s.T, expectedField.Typ, actualField.Typ)
				assertResultsEqual(s.T, s.ClientType, expectedField.DefaultValue, actualField.DefaultValue)
				require.Equal(s.T, expectedField.RelationName, actualField.RelationName)
				require.Equal(s.T, expectedField.Size, actualField.Size)
			}
		}

		if expected.VectorEmbeddings != nil {
			require.Equal(s.T, expected.VectorEmbeddings, actual.VectorEmbeddings)
		}
	}
}

// clientTypeForNode returns the client type used to talk to the given node.
//
// An external node runs in another process and is always reached over HTTP, so
// its results need the relaxed comparison whatever client the run selected.
func clientTypeForNode(s *state.State, nodeID int) state.ClientType {
	if nodeID >= 0 && nodeID < len(s.Nodes) && s.Nodes[nodeID].IsExternal {
		return state.HTTPClientType
	}
	return s.ClientType
}

// assertResultsEqual asserts that actual result is equal to the expected result.
//
// The comparison is relaxed when using client types other than goClientType.
func assertResultsEqual(t testing.TB, client state.ClientType, expected any, actual any, msgAndArgs ...any) {
	switch client {
	case state.HTTPClientType, state.CLIClientType, state.JSClientType, state.CClientType:
		if !areResultsEqual(expected, actual) {
			assert.EqualValues(t, expected, actual, msgAndArgs...)
		}
	default:
		assert.EqualValues(t, expected, actual, msgAndArgs...)
	}
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
	case float32:
		jsonNum, ok := actual.(json.Number)
		if !ok {
			return assert.ObjectsAreEqualValues(expected, actual)
		}
		actualVal, err := jsonNum.Float64()
		if err != nil {
			return false
		}
		return assert.ObjectsAreEqualValues(expected, float32(actualVal))
	case float64:
		jsonNum, ok := actual.(json.Number)
		if !ok {
			return assert.ObjectsAreEqualValues(expected, actual)
		}
		actualVal, err := jsonNum.Float64()
		if err != nil {
			return false
		}
		return assert.ObjectsAreEqualValues(expected, actualVal)
	case immutable.Option[float32]:
		return areResultOptionsEqual(expectedVal, actual)
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
	case immutable.Option[time.Time]:
		return areResultOptionsEqual(expectedVal, actual)
	case []uint8:
		return areResultsEqual(base64.StdEncoding.EncodeToString(expectedVal), actual)
	case []int64:
		return areResultArraysEqual(expectedVal, actual)
	case []uint64:
		return areResultArraysEqual(expectedVal, actual)
	case []float32:
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
	case []immutable.Option[float32]:
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
	case []time.Time:
		return areResultArraysEqual(expectedVal, actual)
	case []immutable.Option[time.Time]:
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
