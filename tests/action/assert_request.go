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
	"fmt"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/onsi/gomega"
	"github.com/onsi/gomega/matchers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corelog"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/state"
)

// DocIndex represents a document index for targeting documents in assertions.
type DocIndex struct {
	// CollectionIndex is the index of the collection holding the document to target.
	CollectionIndex int

	// Index is the index within the target collection at which the document exists.
	//
	// This is dependent on the order in which test AddDoc actions were defined.
	Index int
}

// assertStack keeps track of the current assertion path.
// GraphQL response can be traversed by a key of a map and/or an index of an array.
// So whenever we have a mismatch in a large response, we can use this stack to find the exact path.
// Example output: "_commits[2].links[1].cid"
type assertStack struct {
	stack []string
	isMap []bool
}

func (a *assertStack) pushMap(key string) {
	a.stack = append(a.stack, key)
	a.isMap = append(a.isMap, true)
}

func (a *assertStack) pushArray(index int) {
	a.stack = append(a.stack, strconv.Itoa(index))
	a.isMap = append(a.isMap, false)
}

func (a *assertStack) pop() {
	a.stack = a.stack[:len(a.stack)-1]
	a.isMap = a.isMap[:len(a.isMap)-1]
}

func (a *assertStack) String() string {
	var b strings.Builder
	for i, key := range a.stack {
		if a.isMap[i] {
			if i > 0 {
				_, _ = b.WriteString(".")
			}
			_, _ = b.WriteString(key)
		} else {
			_, _ = b.WriteString("[")
			_, _ = b.WriteString(key)
			_, _ = b.WriteString("]")
		}
	}
	return b.String()
}

// assertRequestResults asserts the results of a GQL request.
func assertRequestResults(
	s *state.State,
	result *client.GQLResult,
	expectedResults map[string]any,
	expectedError string,
	asserter ResultAsserter,
	nodeID int,
	ordered bool,
) bool {
	s.CurrentAssertingNodeID = nodeID
	// we skip assertion benchmark because you don't specify expected result for benchmark.
	if assertErrors(s.T, result.Errors, expectedError) || s.IsBench {
		return true
	}

	if expectedResults == nil && result.Data == nil {
		return false
	}

	// Note: if result.Data == nil this panics (the panic seems useful while testing).
	resultantData, ok := result.Data.(map[string]any)
	if !ok {
		return false
	}
	log.InfoContext(s.Ctx, "", corelog.Any("RequestResults", result.Data))

	if asserter != nil {
		asserter.Assert(s.T, resultantData)
		return true
	}

	// merge all keys so we can check for missing values
	keys := make(map[string]struct{})
	for key := range resultantData {
		keys[key] = struct{}{}
	}
	for key := range expectedResults {
		keys[key] = struct{}{}
	}

	stack := &assertStack{}
	for key := range keys {
		stack.pushMap(key)
		expect, ok := expectedResults[key]
		require.True(s.T, ok, "expected key not found: %s", key)

		actual, ok := resultantData[key]
		require.True(s.T, ok, "result key not found: %s", key)

		switch exp := expect.(type) {
		case []map[string]any:
			actualDocs := ConvertToArrayOfMaps(s.T, actual)
			ok := assertRequestResultDocs(
				s,
				nodeID,
				exp,
				actualDocs,
				stack,
				ordered,
			)
			if !ordered {
				require.True(s.T, ok, "non-ordered expected results: %v not matching actual: %v", exp, actualDocs)
			}

		case gomega.OmegaMatcher:
			execGomegaMatcher(exp, s, actual, stack)

		default:
			assertResultsEqual(
				s.T,
				s.ClientType,
				expect,
				actual,
				fmt.Sprintf("node: %v, path: %s", nodeID, stack),
			)
		}
		stack.pop()
	}

	return false
}

// assertRequestResultDocs returns true if the assertion was successful.
//
// The returned boolean only matters if the assertion is NOT for an ordered set.
func assertRequestResultDocs(
	s *state.State,
	nodeID int,
	expectedResults []map[string]any,
	actualResults []map[string]any,
	stack *assertStack,
	ordered bool,
) bool {
	// compare results
	if !ordered {
		if len(expectedResults) != len(actualResults) {
			return false
		}
		matchedExpectedDocs := make(map[int]struct{}, len(actualResults))
	actualLoop:
		for _, actualDoc := range actualResults {
			found := false
			for expectedDocIndex, expectedDoc := range expectedResults {
				if _, ok := matchedExpectedDocs[expectedDocIndex]; ok {
					// no need to run the process again if this doc was already matched.
					continue
				}
				if len(expectedDoc) != len(actualDoc) {
					continue
				}
				isEqual := assertRequestResultDoc(s, nodeID, actualDoc, expectedDoc, stack, ordered)
				if isEqual {
					found = true
					matchedExpectedDocs[expectedDocIndex] = struct{}{}
					continue actualLoop
				}
			}
			if !found {
				return false
			}
		}

		return true
	}

	require.Equal(s.T, len(expectedResults), len(actualResults), "number of results don't match for %s", stack)

	for actualDocIndex, actualDoc := range actualResults {
		stack.pushArray(actualDocIndex)
		expectedDoc := expectedResults[actualDocIndex]

		require.Equal(
			s.T,
			len(expectedDoc),
			len(actualDoc),
			fmt.Sprintf(
				"number of properties don't match for %s",
				stack,
			),
		)

		assertRequestResultDoc(s, nodeID, actualDoc, expectedDoc, stack, ordered)

		stack.pop()
	}
	return true
}

// assertRequestResultDoc return true if the assertion was successful.
//
// The returned boolean only matters for non ordered assertions.
func assertRequestResultDoc(
	s *state.State,
	nodeID int,
	actualDoc map[string]any,
	expectedDoc map[string]any,
	stack *assertStack,
	ordered bool,
) bool {
	for field, actualValue := range actualDoc {
		stack.pushMap(field)

		switch expectedValue := expectedDoc[field].(type) {
		case gomega.OmegaMatcher:
			if ordered {
				execGomegaMatcher(expectedValue, s, actualValue, stack)
			} else {
				ok := checkGomegaMatcher(expectedValue, s, actualValue)
				if !ok {
					stack.pop()
					return false
				}
			}

		case DocIndex:
			s.DocIDsLock.RLock()
			expectedDocID := s.DocIDs[expectedValue.CollectionIndex][expectedValue.Index].String()
			s.DocIDsLock.RUnlock()

			if ordered {
				assertResultsEqual(
					s.T,
					s.ClientType,
					expectedDocID,
					actualValue,
					fmt.Sprintf("node: %v, path: %s", nodeID, stack),
				)
			} else {
				ok := isResultsEqual(
					s.ClientType,
					expectedDocID,
					actualValue,
				)
				if !ok {
					stack.pop()
					return false
				}
			}
		case []map[string]any:
			actualValueMap := ConvertToArrayOfMaps(s.T, actualValue)

			ok := assertRequestResultDocs(
				s,
				nodeID,
				expectedValue,
				actualValueMap,
				stack,
				ordered,
			)
			if !ok && !ordered {
				stack.pop()
				return false
			}

		case map[string]any:
			actualMap, ok := actualValue.(map[string]any)
			if ordered {
				require.True(s.T, ok, "expected value to be a map %v. Path: %s", actualValue, stack)
			} else if !ok {
				return false
			}

			ok = assertRequestResultDoc(s, nodeID, actualMap, expectedValue, stack, ordered)
			if !ok && !ordered {
				stack.pop()
				return false
			}

		default:
			if ordered {
				assertResultsEqual(
					s.T,
					s.ClientType,
					expectedValue,
					actualValue,
					fmt.Sprintf("node: %v, path: %s", nodeID, stack),
				)
			} else {
				ok := isResultsEqual(
					s.ClientType,
					expectedValue,
					actualValue,
				)
				if !ok {
					stack.pop()
					return false
				}
			}
		}
		stack.pop()
	}
	return true
}

// ConvertToArrayOfMaps converts an interface value to an array of maps.
func ConvertToArrayOfMaps(t testing.TB, value any) []map[string]any {
	valueArrayMap, ok := value.([]map[string]any)
	if ok {
		return valueArrayMap
	}
	valueArray, ok := value.([]any)
	require.True(t, ok, "expected value to be an array of maps %v", value)

	valueArrayMap = make([]map[string]any, len(valueArray))
	for i, v := range valueArray {
		valueArrayMap[i], ok = v.(map[string]any)
		require.True(t, ok, "expected value to be an array of maps %v", value)
	}
	return valueArrayMap
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

// execGomegaMatcher executes the given gomega matcher and asserts the result.
func execGomegaMatcher(exp gomega.OmegaMatcher, s *state.State, actual any, stack *assertStack) {
	traverseGomegaMatchers(exp, s, func(m state.TestStateMatcher) { m.SetTestState(s) })

	success, err := exp.Match(actual)
	if err != nil {
		assert.Fail(s.T, "the matcher exited with error", "Error: %s. Path: %s", err, stack)
	}

	if !success {
		assert.Fail(s.T, exp.FailureMessage(actual), "Path: %s", stack)
	}

	traverseGomegaMatchers(exp, s, func(m state.StatefulMatcher) {
		if !slices.Contains(s.StatefulMatchers, m) {
			s.StatefulMatchers = append(s.StatefulMatchers, m)
		}
	})
}

// checkGomegaMatcher executes the given gomega matcher and returns true if successful.
func checkGomegaMatcher(exp gomega.OmegaMatcher, s *state.State, actual any) bool {
	traverseGomegaMatchers(exp, s, func(m state.TestStateMatcher) { m.SetTestState(s) })

	success, err := exp.Match(actual)
	if err != nil || !success {
		return false
	}

	traverseGomegaMatchers(exp, s, func(m state.StatefulMatcher) {
		if !slices.Contains(s.StatefulMatchers, m) {
			s.StatefulMatchers = append(s.StatefulMatchers, m)
		}
	})

	return true
}

// traverseGomegaMatchers traverses the given gomega matcher and calls the given function
// for each matcher found with the type T.
func traverseGomegaMatchers[T gomega.OmegaMatcher](exp gomega.OmegaMatcher, s *state.State, f func(T)) {
	if m, ok := exp.(T); ok {
		f(m)
		return
	}

	var elements []any
	var matchersList []gomega.OmegaMatcher

	switch exp := exp.(type) {
	case *matchers.AndMatcher:
		matchersList = exp.Matchers
	case *matchers.OrMatcher:
		matchersList = exp.Matchers
	case *matchers.NotMatcher:
		matchersList = []gomega.OmegaMatcher{exp.Matcher}
	case *matchers.ConsistOfMatcher:
		elements = exp.Elements
	case *matchers.ContainElementMatcher:
		elements = []any{exp.Element}
	case *matchers.BeElementOfMatcher:
		elements = exp.Elements
	case *matchers.HaveExactElementsMatcher:
		elements = exp.Elements
	case *matchers.ContainElementsMatcher:
		elements = exp.Elements
	case *matchers.HaveEachMatcher:
		elements = []any{exp.Element}
	case *matchers.WithTransformMatcher:
		matchersList = []gomega.OmegaMatcher{exp.Matcher}
	}

	if len(matchersList) > 0 {
		for _, m := range matchersList {
			traverseGomegaMatchers(m, s, f)
		}
	}

	if len(elements) > 0 {
		for _, el := range elements {
			if m, ok := el.(gomega.OmegaMatcher); ok {
				traverseGomegaMatchers(m, s, f)
			}
		}
	}
}
