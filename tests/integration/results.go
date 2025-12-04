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
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/immutable"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/state"
)

func init() {
	format.RegisterCustomFormatter(func(value any) (string, bool) {
		if matcher, ok := value.(*docIDAt); ok {
			return matcher.String(), true
		}
		return "", false
	})
}

// TestState is read-only interface for test state. It allows passing the state to custom matchers
// without allowing them to modify the state.
type TestState interface {
	// GetClientType returns the client type of the test.
	GetClientType() state.ClientType
	// GetCurrentNodeID returns the node id that is currently being asserted.
	GetCurrentNodeID() int
	// GetIdentity returns the identity for the given node index.
	GetIdentity(state.Identity) acpIdentity.Identity
	// GetDocID returns the document ID for the given collection index and document index.
	GetDocID(collectionIndex, docIndex int) client.DocID
}

type testStateMatcher struct {
	s TestState
}

func (matcher *testStateMatcher) SetTestState(s TestState) {
	matcher.s = s
}

// TestStateMatcher is a matcher that requires access to the test state.
type TestStateMatcher interface {
	types.GomegaMatcher
	// SetTestState sets the test state.
	SetTestState(s TestState)
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
	testStateMatcher
	Values []any
}

var _ TestStateMatcher = (*anyOf)(nil)

func (matcher *anyOf) Match(actual any) (bool, error) {
	switch matcher.s.GetClientType() {
	case state.HTTPClientType, state.CLIClientType, state.JSClientType, state.CClientType:
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
	testStateMatcher
	seenValues       []map[any]bool
	invalidValueType any
}

var _ StatefulMatcher = (*UniqueValue)(nil)

// NewUniqueValue creates a new matcher that verifies each value is unique.
// This matcher will track values across all Match calls and fail if a duplicate is found.
func NewUniqueValue() *UniqueValue {
	return &UniqueValue{}
}

func (matcher *UniqueValue) ResetMatcherState() {
	matcher.seenValues = nil
}

func (matcher *UniqueValue) Match(actual any) (bool, error) {
	nodeID := matcher.s.GetCurrentNodeID()
	for nodeID >= len(matcher.seenValues) {
		matcher.seenValues = append(matcher.seenValues, make(map[any]bool))
	}

	var key any

	if !reflect.TypeOf(actual).Comparable() {
		key = fmt.Sprintf("%v", actual)
	} else {
		key = actual
	}

	if matcher.seenValues[nodeID][key] {
		return false, nil
	}

	matcher.seenValues[nodeID][key] = true
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

// SameValue ensures that values passed to Match are the same as the previous value.
// An instance of this matcher should be given to at least 2 assert result places, otherwise
// the matcher makes no sense.
type SameValue struct {
	value any
}

var _ StatefulMatcher = (*SameValue)(nil)

// NewSameValue creates a new matcher that verifies each value is the same as the previous value.
func NewSameValue() *SameValue {
	return &SameValue{}
}

func (matcher *SameValue) ResetMatcherState() {
	matcher.value = nil
}

func (matcher *SameValue) Match(actual any) (bool, error) {
	var newValue any

	if !reflect.TypeOf(actual).Comparable() {
		newValue = fmt.Sprintf("%v", actual)
	} else {
		newValue = actual
	}

	if matcher.value == nil {
		matcher.value = newValue
		return true, nil
	}

	if matcher.value != newValue {
		return false, nil
	}

	return true, nil
}

func (matcher *SameValue) FailureMessage(actual any) string {
	return fmt.Sprintf("Expected value to be the same as the previous value. \n\tPrevious: %v \n\tCurrent:  %v",
		matcher.value, actual)
}

func (matcher *SameValue) NegatedFailureMessage(actual any) string {
	return fmt.Sprintf("Expected value to be different from the previous value. \n\tPrevious: %v \n\tCurrent:  %v",
		matcher.value, actual)
}

// DocIDAt returns a matcher that checks if the actual value is a document ID
// at the specified collection index and document index.
func DocIDAt(collectionIndex, docIndex int) *docIDAt {
	return &docIDAt{
		collectionIndex: collectionIndex,
		docIndex:        docIndex,
	}
}

// docIDAt is a matcher that checks if the actual value is a document ID
// at the specified collection index and document index.
type docIDAt struct {
	testStateMatcher
	collectionIndex int
	docIndex        int
}

var _ TestStateMatcher = (*docIDAt)(nil)

func (matcher *docIDAt) Match(actual any) (bool, error) {
	actualDocID, ok := actual.(string)
	if !ok {
		return false, fmt.Errorf("expected a document ID string, got %T", actual)
	}
	expectedDocID := matcher.s.GetDocID(matcher.collectionIndex, matcher.docIndex).String()
	return actualDocID == expectedDocID, nil
}

func (matcher *docIDAt) FailureMessage(actual any) string {
	expectedDocID := matcher.s.GetDocID(matcher.collectionIndex, matcher.docIndex).String()
	return fmt.Sprintf("Expected\n\t%v\nto be a doID: %s", actual, expectedDocID)
}

func (matcher *docIDAt) NegatedFailureMessage(actual any) string {
	expectedDocID := matcher.s.GetDocID(matcher.collectionIndex, matcher.docIndex).String()
	return fmt.Sprintf("Expected\n\t%v\nnot to be a doID: %s", actual, expectedDocID)
}

func (matcher *docIDAt) String() string {
	return fmt.Sprintf("DocIDAt(collectionIndex: %d, docIndex: %d): %s", matcher.collectionIndex,
		matcher.docIndex, matcher.s.GetDocID(matcher.collectionIndex, matcher.docIndex).String())
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

// isResultsEqual checks that actual result is equal to the expected result and returns true if they are.
//
// The comparison is relaxed when using client types other than goClientType.
func isResultsEqual(client state.ClientType, expected any, actual any) bool {
	switch client {
	case state.HTTPClientType, state.CLIClientType, state.JSClientType, state.CClientType:
		if !areResultsEqual(expected, actual) {
			return assert.ObjectsAreEqualValues(expected, actual)
		}
		return true
	default:
		return assert.ObjectsAreEqualValues(expected, actual)
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

func assertCollectionVersions(
	s *state.State,
	expected []client.CollectionVersion,
	actual []client.CollectionVersion,
) {
	require.Equal(s.T, len(expected), len(actual), "collection versions count mismatch")

	for i, expected := range expected {
		actual := actual[i]
		require.Equal(s.T, expected.Name, actual.Name, "version name mismatch")

		if expected.CollectionSet.HasValue() {
			require.Equal(s.T, expected.CollectionSet.Value().CollectionSetID,
				actual.CollectionSet.Value().CollectionSetID, "collection set ID mismatch")
			require.Equal(s.T, expected.CollectionSet.Value().RelativeID,
				actual.CollectionSet.Value().RelativeID, "collection set relative ID mismatch")
		}

		if expected.VersionID != "" {
			require.Equal(s.T, expected.VersionID, actual.VersionID, "version %d: version ID mismatch", i)
		}
		if expected.CollectionID != "" {
			require.Equal(s.T, expected.CollectionID, actual.CollectionID, "version %d: collection ID mismatch", i)
		}

		require.Equal(s.T, expected.IsMaterialized, actual.IsMaterialized, "version %d: is materialized mismatch", i)
		require.Equal(s.T, expected.IsBranchable, actual.IsBranchable, "version %d: is branchable mismatch", i)
		require.Equal(s.T, expected.IsActive, actual.IsActive, "version %d: is active mismatch", i)

		if expected.Indexes != nil || len(actual.Indexes) != 0 {
			// Dont bother asserting this if the expected is nil and the actual is nil/empty.
			// This is to save each test action from having to bother declaring an empty slice (if there are no indexes)
			require.Equal(s.T, expected.Indexes, actual.Indexes, "version %d: indexes mismatch", i)
		}

		require.Equal(s.T, expected.PreviousVersion.HasValue(), actual.PreviousVersion.HasValue(),
			"version %d: previous version existence mismatch", i)
		if expected.PreviousVersion.HasValue() {
			require.Equal(
				s.T,
				expected.PreviousVersion.Value().SourceCollectionID,
				actual.PreviousVersion.Value().SourceCollectionID,
				"version %d: previous version source collection ID mismatch", i,
			)
			require.Equal(
				s.T,
				expected.PreviousVersion.Value().Transform.HasValue(),
				actual.PreviousVersion.Value().Transform.HasValue(),
				"version %d: previous version transform existence mismatch", i,
			)

			if expected.PreviousVersion.Value().Transform.HasValue() {
				// Dont bother asserting this by default, the transform object is too complex to bother with in most cases.
				require.Equal(
					s.T,
					expected.PreviousVersion.Value().Transform.Value(),
					actual.PreviousVersion.Value().Transform.Value(),
					"version %d: previous version transform value mismatch", i,
				)
			}
		}

		if expected.Query.HasValue() {
			// Dont bother asserting this by default, the query object is to complex to bother with in most cases.
			require.Equal(s.T, expected.Query, actual.Query, "version %d: query mismatch", i)
		}

		if expected.Fields != nil {
			require.Equal(s.T, len(expected.Fields), len(actual.Fields), "version %d: fields count mismatch", i)
			for j := range expected.Fields {
				expectedField := expected.Fields[j]
				actualField := actual.Fields[j]

				require.Equal(s.T, expectedField.Name, actualField.Name,
					"version %d, field %d: field name mismatch", i, j)
				if expectedField.FieldID != "" {
					require.Equal(s.T, expectedField.FieldID, actualField.FieldID,
						"version %d, field %d: field ID mismatch", i, j)
				}
				require.Equal(s.T, expectedField.IsPrimary, actualField.IsPrimary,
					"version %d, field %d: field is primary mismatch", i, j)
				require.Equal(s.T, expectedField.Kind, actualField.Kind,
					"version %d, field %d: field kind mismatch", i, j)
				require.Equal(s.T, expectedField.Typ, actualField.Typ,
					"version %d, field %d: field type mismatch", i, j)
				require.Equal(s.T, expectedField.DefaultValue, actualField.DefaultValue,
					"version %d, field %d: field default value mismatch", i, j)
				require.Equal(s.T, expectedField.RelationName, actualField.RelationName,
					"version %d, field %d: field relation name mismatch", i, j)
				require.Equal(s.T, expectedField.Size, actualField.Size,
					"version %d, field %d: field size mismatch", i, j)
			}
		}

		if expected.VectorEmbeddings != nil {
			require.Equal(s.T, expected.VectorEmbeddings, actual.VectorEmbeddings,
				"version %d: vector embeddings mismatch", i)
		}
	}
}

// CurrentTimestampMatcher is a matcher that checks if the actual value is a
//
//	time.Time within 5 seconds of the current time.
type CurrentTimestampMatcher struct {
	testStateMatcher
}

var _ TestStateMatcher = (*CurrentTimestampMatcher)(nil)

func CurrentTimestamp() *CurrentTimestampMatcher {
	return &CurrentTimestampMatcher{}
}

func (matcher *CurrentTimestampMatcher) Match(actual any) (bool, error) {
	var ts time.Time

	// We want this to work with time.Time as well as strings that can
	// be parsed into a time.Time
	switch v := actual.(type) {
	case time.Time:
		ts = v

	case string:
		parsed, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return false, fmt.Errorf(
				"expected time.Time or RFC3339 string, got unparsable string %q: %w",
				v, err,
			)
		}
		ts = parsed

	default:
		return false, fmt.Errorf("expected time.Time or string, got %T", actual)
	}

	diff := time.Since(ts)
	if diff < 0 {
		diff = -diff
	}

	if diff > 5*time.Second {
		return false, fmt.Errorf("timestamp %v is more than 5 seconds away from now", ts)
	}

	return true, nil
}

func (matcher *CurrentTimestampMatcher) FailureMessage(actual any) string {
	return fmt.Sprintf("Expected timestamp %v to be within 5 seconds of now", actual)
}

func (matcher *CurrentTimestampMatcher) NegatedFailureMessage(actual any) string {
	return fmt.Sprintf("Expected timestamp %v not to be within 5 seconds of now", actual)
}
