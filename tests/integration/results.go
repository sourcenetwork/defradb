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

package tests

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/immutable"

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

// TestState is a type alias for state.TestState.
type TestState = state.TestState

// TestStateMatcher is a type alias for state.TestStateMatcher.
type TestStateMatcher = state.TestStateMatcher

// StatefulMatcher is a type alias for state.StatefulMatcher.
type StatefulMatcher = state.StatefulMatcher

type testStateMatcher struct {
	s state.TestState
}

func (matcher *testStateMatcher) SetTestState(s state.TestState) {
	matcher.s = s
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
	nodeID := matcher.s.GetCurrentAssertingNodeID()
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

// similarityScoreTolerance absorbs the gap between the exact float64 score and the value the engine
// gets from float32-stored vectors, whose components are not all exactly representable in float32.
const similarityScoreTolerance = 1e-6

// CosineSimilarity matches a _similarity result against the cosine of the two given vectors. Prefer
// it over a hard-coded float: it names the vectors under comparison and cannot go stale.
func CosineSimilarity(source, vector []float64) *similarityScore {
	return &similarityScore{source: source, vector: vector}
}

// SimilarityScore matches a _similarity result under the given metric. Use it when the field being
// queried has a non-cosine vector index, since the score follows the index's metric.
func SimilarityScore(metric client.DistanceMetric, source, vector []float64) *similarityScore {
	return &similarityScore{source: source, vector: vector, metric: metric}
}

type similarityScore struct {
	source []float64
	vector []float64
	// The zero value is cosine, so CosineSimilarity keeps working unchanged.
	metric client.DistanceMetric
}

var _ gomega.OmegaMatcher = (*similarityScore)(nil)

// expected computes the score the two vectors should produce under the metric.
//
// It is written out here rather than calling the production scoring function, so that the assertion
// is an independent check on the value and not just on the ordering. Sharing the implementation would
// let a wrong but order-preserving formula (say a square root left in) satisfy every assertion.
func (m *similarityScore) expected() float64 {
	var dot, sumSq, sourceNorm, vectorNorm float64
	for i := range m.source {
		s, v := m.source[i], m.vector[i]
		dot += s * v
		d := s - v
		sumSq += d * d
		sourceNorm += s * s
		vectorNorm += v * v
	}

	switch m.metric {
	case client.DistanceMetricEuclidean:
		return -sumSq
	case client.DistanceMetricDotProduct:
		return dot
	default:
		if sourceNorm == 0 || vectorNorm == 0 {
			return 0
		}
		return dot / (math.Sqrt(sourceNorm) * math.Sqrt(vectorNorm))
	}
}

func (m *similarityScore) Match(actual any) (bool, error) {
	// The HTTP, CLI and C clients decode the result through JSON, so the similarity value arrives as a
	// json.Number rather than a float64. gomega.BeNumerically only accepts native numeric types, so
	// unwrap it first, the same way the rest of the result-matching does.
	if jsonNum, ok := actual.(json.Number); ok {
		f, err := jsonNum.Float64()
		if err != nil {
			return false, err
		}
		actual = f
	}
	return gomega.BeNumerically("~", m.expected(), similarityScoreTolerance).Match(actual)
}

func (m *similarityScore) FailureMessage(actual any) string {
	return fmt.Sprintf("Expected\n\t%v\nto be the %s similarity score\n\t%v (within %v)",
		actual, m.metricName(), m.expected(), similarityScoreTolerance)
}

func (m *similarityScore) NegatedFailureMessage(actual any) string {
	return fmt.Sprintf("Expected\n\t%v\nnot to be the %s similarity score\n\t%v (within %v)",
		actual, m.metricName(), m.expected(), similarityScoreTolerance)
}

// metricName names the metric in failure messages, so a mismatch says which one was expected.
func (m *similarityScore) metricName() string {
	if m.metric == "" {
		return string(client.DistanceMetricCosine)
	}
	return string(m.metric)
}

func ValidDocID() *validDocID {
	return &validDocID{}
}

type validDocID struct{}

func (m *validDocID) Match(actual any) (bool, error) {
	s, ok := actual.(string)
	if !ok {
		return false, fmt.Errorf("expected a document ID string, got %T", actual)
	}
	if _, err := client.NewDocIDFromString(s); err != nil {
		return false, nil
	}
	return true, nil
}

func (m *validDocID) FailureMessage(actual any) string {
	return fmt.Sprintf("Expected\n\t%v\nto be a valid document ID string", actual)
}

func (m *validDocID) NegatedFailureMessage(actual any) string {
	return fmt.Sprintf("Expected\n\t%v\nnot to be a valid document ID string", actual)
}

// ValidCID returns a matcher that passes if the actual value is a string
// that parses as a valid CID.
//
// Use this instead of hard-coding a specific CID string in test results
// when the test's intent is "a CID is returned", not "this exact CID is
// returned". Hard-coded CIDs over-specify the test and break whenever
// something changes block bytes (e.g. enabling signing via the
// [multiplier.SignedDocs] test multiplier).
func ValidCID() *validCID {
	return &validCID{}
}

type validCID struct{}

func (m *validCID) Match(actual any) (bool, error) {
	s, ok := actual.(string)
	if !ok {
		return false, fmt.Errorf("expected a CID string, got %T", actual)
	}
	if _, err := cid.Decode(s); err != nil {
		return false, nil
	}
	return true, nil
}

func (m *validCID) FailureMessage(actual any) string {
	return fmt.Sprintf("Expected\n\t%v\nto be a valid CID string", actual)
}

func (m *validCID) NegatedFailureMessage(actual any) string {
	return fmt.Sprintf("Expected\n\t%v\nnot to be a valid CID string", actual)
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

// CurrentTimestampMatcher is a matcher that checks if the actual value is a
//
//	time.Time within 120 seconds of the current time. The reason for this window
//
// is to allow for some latency in our test runs.
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

	if diff > 120*time.Second {
		return false, fmt.Errorf("timestamp %v is more than 120 seconds away from now", ts)
	}

	return true, nil
}

func (matcher *CurrentTimestampMatcher) FailureMessage(actual any) string {
	return fmt.Sprintf("Expected timestamp %v to be within 120 seconds of now", actual)
}

func (matcher *CurrentTimestampMatcher) NegatedFailureMessage(actual any) string {
	return fmt.Sprintf("Expected timestamp %v not to be within 120 seconds of now", actual)
}

// ArrayMatcher is a matcher that checks if the actual array of value is a
// match for all the matchers in the ArrayMatcher's matchers field.
//
// The actual value must be an array of the same length as the matchers field,
// and each element of the actual array must match the corresponding matcher in
// the matchers field.
type ArrayMatcher struct {
	matchers []TestStateMatcher
	testStateMatcher
}

var _ TestStateMatcher = (*ArrayMatcher)(nil)

func NewArrayMatcher(matchers ...TestStateMatcher) *ArrayMatcher {
	return &ArrayMatcher{matchers: matchers}
}

func (matcher *ArrayMatcher) Match(actual any) (bool, error) {
	switch v := actual.(type) {
	case []any:
		return matchActual(matcher, v)

	case []map[string]any:
		return matchActual(matcher, v)

	case []string:
		return matchActual(matcher, v)

	case []int64:
		return matchActual(matcher, v)

	case []uint64:
		return matchActual(matcher, v)

	case []float32:
		return matchActual(matcher, v)

	case []float64:
		return matchActual(matcher, v)

	case []bool:
		return matchActual(matcher, v)

	case []immutable.Option[string]:
		return matchActual(matcher, v)

	case []immutable.Option[int64]:
		return matchActual(matcher, v)

	case []immutable.Option[uint64]:
		return matchActual(matcher, v)

	case []immutable.Option[float32]:
		return matchActual(matcher, v)

	case []immutable.Option[float64]:
		return matchActual(matcher, v)

	case []immutable.Option[bool]:
		return matchActual(matcher, v)

	case []time.Time:
		return matchActual(matcher, v)

	case []immutable.Option[time.Time]:
		return matchActual(matcher, v)

	default:
		return false, fmt.Errorf("expected an array, got %T", actual)
	}
}

func matchActual[T any](matcher *ArrayMatcher, actual []T) (bool, error) {
	if len(actual) != len(matcher.matchers) {
		return false, fmt.Errorf("expected array of length %d, got %d", len(matcher.matchers), len(actual))
	}
	for i, matcher := range matcher.matchers {
		if ok, err := matcher.Match(actual[i]); err != nil || !ok {
			return false, fmt.Errorf("element %d: %w", i, err)
		}
	}
	return true, nil
}

func (matcher *ArrayMatcher) FailureMessage(actual any) string {
	return fmt.Sprintf("Expected array %v to match all elements", actual)
}

func (matcher *ArrayMatcher) NegatedFailureMessage(actual any) string {
	return fmt.Sprintf("Expected array %v not to match all elements", actual)
}
