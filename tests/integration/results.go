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
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	cid "github.com/ipfs/go-cid"
	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
)

type stateMatcher struct {
	s *state
}

func (matcher *stateMatcher) SetTestState(s *state) {
	matcher.s = s
}

// TestStateMatcher is a matcher that requires access to the test state.
type TestStateMatcher interface {
	types.GomegaMatcher
	// SetTestState sets the test state.
	SetTestState(s *state)
}

// StatefulMatcher is a matcher that requires state to be reset between tests.
type StatefulMatcher interface {
	types.GomegaMatcher
	// ResetMatcherState resets the state of the matcher.
	ResetMatcherState()
}

// AnyOf may be used as `Results` field where the value may
// be one of several values, yet the value of that field must be the same
// across all nodes due to strong eventual consistency.
func AnyOf(values ...any) *anyOf {
	return &anyOf{
		Values: values,
	}
}

type anyOf struct {
	stateMatcher
	Values []any
}

var _ TestStateMatcher = (*anyOf)(nil)

func (matcher *anyOf) Match(actual any) (bool, error) {
	switch matcher.s.clientType {
	case HTTPClientType, CLIClientType:
		if !areResultsAnyOf(matcher.Values, actual) {
			return gomega.ContainElement(actual).Match(matcher.Values)
		}
	default:
		return gomega.ContainElement(actual).Match(matcher.Values)
	}
	return true, nil
}

func (matcher *anyOf) FailureMessage(actual any) string {
	return fmt.Sprintf("Expected\n\t%v\nto be one of\n\t%v", actual, matcher.Values)
}

func (matcher *anyOf) NegatedFailureMessage(actual any) string {
	return fmt.Sprintf("Expected\n\t%v\nnot to be one of\n\t%v", actual, matcher.Values)
}

// UniqueValue ensures that values passed to Match are unique across all calls.
// It fails if the same value is seen more than once.
// An instance of this matcher should be given to at least 2 assert result places, otherwise
// the matcher makes no sense.
type UniqueValue struct {
	seenValues       map[any]bool
	invalidValueType any
}

var _ StatefulMatcher = (*UniqueValue)(nil)

// NewUniqueValue creates a new matcher that verifies each value is unique.
// This matcher will track values across all Match calls and fail if a duplicate is found.
func NewUniqueValue() *UniqueValue {
	return &UniqueValue{seenValues: make(map[any]bool)}
}

func (matcher *UniqueValue) ResetMatcherState() {
	matcher.seenValues = make(map[any]bool)
}

func (matcher *UniqueValue) Match(actual any) (bool, error) {
	var key any

	if !reflect.TypeOf(actual).Comparable() {
		key = fmt.Sprintf("%v", actual)
	} else {
		key = actual
	}

	if matcher.seenValues[key] {
		return false, nil
	}

	matcher.seenValues[key] = true
	return true, nil
}

func (matcher *UniqueValue) FailureMessage(actual any) string {
	if matcher.invalidValueType != nil {
		return fmt.Sprintf("Expected value to be of type %T, but received: %v", matcher.invalidValueType, actual)
	}
	return fmt.Sprintf("Expected unique value, but received duplicate: %v", actual)
}

func (matcher *UniqueValue) NegatedFailureMessage(actual any) string {
	return fmt.Sprintf("Expected value to be a duplicate, but was unique: %v", actual)
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
func areResultsAnyOf(expected []any, actual any) bool {
	for _, v := range expected {
		if areResultsEqual(v, actual) {
			return true
		}
	}
	return false
}

// UniqueCid allows the referencing of Cids by an arbitrary test-defined ID.
//
// Instead of asserting on a specific Cid value, this type will assert that
// no other [UniqueCid]s with different [ID]s has the first Cid value that this instance
// describes.
//
// It will also ensure that all Cids described by this [UniqueCid] have the same
// valid, Cid value.
type UniqueCid struct {
	stateMatcher
	// id is the arbitrary, but hopefully descriptive, id of this [UniqueCid].
	id any

	valuesMismatch bool
	duplicatedID   any
	castFailed     bool
	cidDecodeErr   error
}

var _ TestStateMatcher = (*UniqueCid)(nil)

// NewUniqueCid creates a new [UniqueCid] of the given arbitrary, but hopefully descriptive,
// id.
//
// All results described by [UniqueCid]s with the given id must have the same valid Cid value.
// No other [UniqueCid] ids may describe the same Cid value.
func NewUniqueCid(id any) *UniqueCid {
	return &UniqueCid{
		id: id,
	}
}

func (matcher *UniqueCid) Match(actual any) (bool, error) {
	isNew := true
	for id, value := range matcher.s.cids {
		if id == matcher.id {
			if value != actual {
				matcher.valuesMismatch = true
				return false, nil
			}
			isNew = false
		} else {
			if value == actual {
				matcher.duplicatedID = id
				return false, nil
			}
		}
	}

	if isNew {
		value, ok := actual.(string)
		if !ok {
			matcher.castFailed = true
			return false, nil
		}

		cid, err := cid.Decode(value)
		if err != nil {
			matcher.cidDecodeErr = err
			return false, nil
		}

		matcher.s.cids[matcher.id] = cid.String()
	}

	return true, nil
}

func (matcher *UniqueCid) FailureMessage(actual any) string {
	if matcher.valuesMismatch {
		return fmt.Sprintf("Expected Cids with the same id %v to match", matcher.id)
	} else if matcher.duplicatedID != nil {
		return fmt.Sprintf("Expected Cid value to be unique, but ids \"%v\" and \"%v\" point to the same cid: %v",
			matcher.id, matcher.duplicatedID, actual)
	} else if matcher.castFailed {
		return fmt.Sprintf("Actual value is expected to be convertible to a string. Actual: %v", actual)
	} else if matcher.cidDecodeErr != nil {
		return fmt.Sprintf("Expected actual value to be a valid Cid. Error: %v", matcher.cidDecodeErr)
	}
	return ""
}

func (matcher *UniqueCid) NegatedFailureMessage(actual any) string {
	panic("UniqueCid cannot be negated")
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
		require.Equal(s.t, expected.IsBranchable, actual.IsBranchable)

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

		if expected.VectorEmbeddings != nil {
			require.Equal(s.t, expected.VectorEmbeddings, actual.VectorEmbeddings)
		}
	}
}
