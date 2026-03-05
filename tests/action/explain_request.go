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
	"reflect"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corelog"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
)

var (
	allPlanNodeNames = map[string]struct{}{
		// Not a planNode but need it here as this is root of the explain graph.
		"explain": {},

		// These are not planNodes but we need to include them here, because typeIndexJoin wraps some nodes
		// under `root` and `subType` attribute (without these they would be skipped from the ordering pattern).
		"root":    {},
		"subType": {},

		// These are all valid nodes.
		"averageNode":    {},
		"countNode":      {},
		"addNode":        {},
		"dagScanNode":    {},
		"deleteNode":     {},
		"groupNode":      {},
		"limitNode":      {},
		"maxNode":        {},
		"minNode":        {},
		"multiScanNode":  {},
		"orderNode":      {},
		"parallelNode":   {},
		"pipeNode":       {},
		"scanNode":       {},
		"selectNode":     {},
		"selectTopNode":  {},
		"sumNode":        {},
		"topLevelNode":   {},
		"typeIndexJoin":  {},
		"typeJoinMany":   {},
		"typeJoinOne":    {},
		"updateNode":     {},
		"upsertNode":     {},
		"valuesNode":     {},
		"viewNode":       {},
		"lensNode":       {},
		"operationNode":  {},
		"similarityNode": {},
	}

	log = corelog.NewLogger("tests.action")
)

// PlanNodeTargetCase is used to target specific plan nodes in explain output.
type PlanNodeTargetCase struct {
	// Name of the plan node, whose attribute(s) we are targetting to be asserted.
	TargetNodeName string

	// How many occurances of this target name to skip until target (0 means match first).
	OccurancesToSkip uint

	// If set to 'true' will include the nested node(s), with their attribute(s) as well.
	IncludeChildNodes bool

	// Expected value of the target node's attribute(s).
	ExpectedAttributes any
}

// ExplainRequest represents an explain query request.
type ExplainRequest struct {
	stateful

	// NodeID is the node ID (index) of the node in which to explain.
	NodeID immutable.Option[int]

	// The identity of this request.
	Identity string

	// Has to be a valid explain request type (one of: 'simple', 'debug', 'execute', 'predict').
	Request string

	// The raw expected explain graph with everything (helpful for debugging purposes).
	// Note: This is not always asserted (i.e. ignored from the comparison if not provided).
	ExpectedFullGraph map[string]any

	// Pattern is used to assert that the plan nodes are in the correct order (attributes are omitted).
	// Note: - Explain requests of type 'debug' will only have Pattern (as they don't have attributes).
	//       - This is not always asserted (i.e. ignored from the comparison if not provided).
	ExpectedPatterns map[string]any

	// Every target helps assert an individual node somewhere in the explain graph (node's position is omitted).
	// Each target assertion is only responsible to check if the node's attributes are correct.
	// This is the only test that sorts the keys and traverses the map in a deterministic order to ensure
	// that consistent skips occur if there are multiple nodes of matching target name.
	// Note: This is not always asserted (i.e. ignored from the comparison if not provided).
	ExpectedTargets []PlanNodeTargetCase

	// The expected error from the explain request.
	ExpectedError string
}

var _ Action = (*ExplainRequest)(nil)
var _ Stateful = (*ExplainRequest)(nil)

// Execute executes the explain request action.
func (a *ExplainRequest) Execute() {
	// Must have a non-empty request.
	if a.Request == "" {
		require.Fail(a.s.T, "Explain test must have a non-empty request.")
	}

	// If no expected results are provided, then it's invalid use of this explain testing setup.
	if a.ExpectedError == "" &&
		a.ExpectedPatterns == nil &&
		a.ExpectedTargets == nil &&
		a.ExpectedFullGraph == nil {
		require.Fail(a.s.T, "Atleast one expected explain parameter must be provided.")
	}

	// If we expect an error, then all other expected results should be empty (they shouldn't be provided).
	if a.ExpectedError != "" &&
		(a.ExpectedFullGraph != nil ||
			a.ExpectedPatterns != nil ||
			a.ExpectedTargets != nil) {
		require.Fail(a.s.T, "Expected error should not have other expected results with it.")
	}

	_, nodes := getNodesWithIDs(a.NodeID, a.s.Nodes)
	for _, node := range nodes {
		result := node.ExecRequest(
			a.s.Ctx,
			a.Request,
		)
		a.assertExplainRequestResults(&result.GQL)
	}
}

func (a *ExplainRequest) assertExplainRequestResults(actualResult *client.GQLResult) {
	// Check expected error matches actual error. If it does we are done.
	if assertErrors(a.s.T, actualResult.Errors, a.ExpectedError) {
		return
	} else if a.ExpectedError != "" { // If didn't find a match but did expected an error, then fail.
		assert.Fail(a.s.T, "Expected an error however none was raised.")
	}

	// Note: if returned gql result is `nil` this panics (the panic seems useful while testing).
	resultantData, ok := actualResult.Data.(map[string]any)
	if !ok {
		return
	}
	log.InfoContext(a.s.Ctx, "", corelog.Any("FullExplainGraphResult", actualResult.Data))

	// Check if the expected full explain graph (if provided) matches the actual full explain graph
	// that is returned, if doesn't match we would like to still see a diff comparison (handy while debugging).
	if a.ExpectedFullGraph != nil {
		assertResultsEqual(
			a.s.T,
			a.s.ClientType,
			a.ExpectedFullGraph,
			resultantData,
		)
	}

	// Ensure the complete high-level pattern matches, inother words check that all the
	// explain graph nodes are in the correct expected ordering.
	if a.ExpectedPatterns != nil {
		// Trim away all attributes (non-plan nodes) from the returned full explain graph result.
		actualResultWithoutAttributes := trimExplainAttributes(a.s.T, resultantData)
		assertResultsEqual(
			a.s.T,
			a.s.ClientType,
			a.ExpectedPatterns,
			actualResultWithoutAttributes,
		)
	}

	// Match the targeted node's attributes (subset assertions), with the expected attributes.
	// Note: This does not check if the node is in correct location or not.
	if a.ExpectedTargets != nil {
		for _, target := range a.ExpectedTargets {
			a.assertExplainTargetCase(target, resultantData)
		}
	}
}

func (a *ExplainRequest) assertExplainTargetCase(
	targetCase PlanNodeTargetCase,
	actualResults map[string]any,
) {
	for _, actualResult := range actualResults {
		foundActualTarget, _, isFound := findTargetNode(
			targetCase.TargetNodeName,
			targetCase.OccurancesToSkip,
			targetCase.IncludeChildNodes,
			actualResult,
		)

		if !isFound {
			assert.Fail(
				a.s.T,
				"Expected target ["+targetCase.TargetNodeName+"], was not found in the explain graph.",
			)
		}

		assertResultsEqual(
			a.s.T,
			a.s.ClientType,
			targetCase.ExpectedAttributes,
			foundActualTarget,
		)
	}
}

// assertErrors asserts whether the expected error is present in the given errors slice.
// Returns true if the expected error was found.
func assertErrors(t testing.TB, errs []error, expectedError string) bool {
	if expectedError == "" {
		require.Empty(t, errs)
	} else {
		for _, e := range errs {
			errorString := e.Error()
			if errorString == expectedError || contains(errorString, expectedError) {
				return true
			} else {
				require.ErrorContains(t, e, expectedError)
			}
		}
	}
	return false
}

// contains checks if s contains substr.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && searchString(s, substr) >= 0))
}

// searchString returns the index of substr in s, or -1 if not found.
func searchString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// findTargetNode returns true if the targetName is found in the explain graph after skipping given number of
// occurances, 0 means first occurance. The function also returns total occurances it encountered so far. The
// returned count of 'matches' should always be <= occurance argument.
//
// Note: The traversal of the map must be in a deterministic and ordered manner, so we skip the same nodes items
// with every run and the occurances to skip logic behaves consistently.
func findTargetNode(
	targetName string,
	toSkip uint,
	includeChildNodes bool,
	actualResult any,
) (any, uint, bool) {
	var totalMatchedSoFar uint = 0

	switch r := actualResult.(type) {
	case map[string]any:
		// To traverse the unordered map in a deterministic order, we will collect the keys, sort them
		// in increasing order, and then traverse the map in that order.
		sortedKeys := make([]string, len(r))

		var index uint = 0
		for k := range r {
			sortedKeys[index] = k
			index++
		}

		sort.Strings(sortedKeys)

		for _, key := range sortedKeys {
			if isPlanNode(key) {
				value := r[key]
				if key == targetName {
					totalMatchedSoFar++

					if toSkip == 0 {
						if includeChildNodes {
							return value, totalMatchedSoFar, true
						}
						return trimSubNodes(value), totalMatchedSoFar, true
					}

					toSkip--
					target, matches, found := findTargetNode(
						targetName,
						toSkip,
						includeChildNodes,
						value,
					)

					totalMatchedSoFar = totalMatchedSoFar + matches
					toSkip -= matches

					if found {
						if includeChildNodes {
							return target, totalMatchedSoFar, true
						}
						return trimSubNodes(target), totalMatchedSoFar, true
					}
				} else {
					// Not a match, traverse furthur.
					target, matches, found := findTargetNode(
						targetName,
						toSkip,
						includeChildNodes,
						value,
					)

					totalMatchedSoFar = totalMatchedSoFar + matches
					toSkip -= matches

					if found {
						if includeChildNodes {
							return target, totalMatchedSoFar, true
						}
						return trimSubNodes(target), totalMatchedSoFar, true
					}
				}
			}
		}

	case []any:
		return findTargetNodeFromArray(targetName, toSkip, includeChildNodes, r)

	case []map[string]any:
		return findTargetNodeFromArray(targetName, toSkip, includeChildNodes, r)
	}

	return nil, totalMatchedSoFar, false
}

// findTargetNodeFromArray is a helper that runs findTargetNode for each item in an array.
func findTargetNodeFromArray[T any](
	targetName string,
	toSkip uint,
	includeChildNodes bool,
	actualResult []T,
) (any, uint, bool) {
	var totalMatchedSoFar uint = 0

	for _, item := range actualResult {
		target, matches, found := findTargetNode(
			targetName,
			toSkip,
			includeChildNodes,
			item,
		)

		totalMatchedSoFar = totalMatchedSoFar + matches
		toSkip -= matches

		if found {
			if includeChildNodes {
				return target, totalMatchedSoFar, true
			}
			return trimSubNodes(target), totalMatchedSoFar, true
		}
	}

	return nil, totalMatchedSoFar, false
}

// trimSubNodes returns a graph where all the immediate sub nodes are trimmed (i.e. no nested subnodes remain).
func trimSubNodes(graph any) any {
	checkGraph, ok := graph.(map[string]any)
	if !ok {
		return graph
	}

	// Copying is super important here so we don't trim the actual result (as we might want to continue using it),
	trimGraph := copyMap(checkGraph)
	for key := range trimGraph {
		if isPlanNode(key) {
			delete(trimGraph, key)
		}
	}

	return trimGraph
}

// trimExplainAttributes trims away all keys that aren't plan nodes within the explain graph.
func trimExplainAttributes(
	t testing.TB,
	actualResult any,
) map[string]any {
	resultMap, ok := actualResult.(map[string]any)
	if !ok {
		return nil
	}
	trimmedMap := copyMap(resultMap)

	for key, value := range trimmedMap {
		if !isPlanNode(key) {
			delete(trimmedMap, key)
			continue
		}

		switch v := value.(type) {
		case map[string]any:
			trimmedMap[key] = trimExplainAttributes(t, v)

		case []map[string]any:
			trimmedMap[key] = trimExplainAttributesArray(t, v)

		case []any:
			trimmedMap[key] = trimExplainAttributesArray(t, v)

		default:
			assert.Fail(
				t,
				"Unsupported explain graph key-value type encountered: "+reflect.TypeOf(v).String(),
			)
		}
	}

	return trimmedMap
}

// trimExplainAttributesArray is a helper that runs trimExplainAttributes for each item in an array.
func trimExplainAttributesArray[T any](
	t testing.TB,
	actualResult []T,
) []map[string]any {
	trimmedArrayElements := []map[string]any{}
	for _, valueItem := range actualResult {
		trimmedArrayElements = append(
			trimmedArrayElements,
			trimExplainAttributes(t, valueItem),
		)
	}
	return trimmedArrayElements
}

// isPlanNode returns true if someName matches a plan node name, retruns false otherwise.
func isPlanNode(someName string) bool {
	_, isPlanNode := allPlanNodeNames[someName]
	return isPlanNode
}

func copyMap(originalMap map[string]any) map[string]any {
	newMap := make(map[string]any, len(originalMap))
	for oKey, oValue := range originalMap {
		switch v := oValue.(type) {
		case map[string]any:
			newMap[oKey] = copyMap(v)

		case []map[string]any:
			newList := make([]map[string]any, len(v))
			for index, item := range v {
				newList[index] = copyMap(item)
			}
			newMap[oKey] = newList

		default:
			newMap[oKey] = oValue
		}
	}
	return newMap
}
