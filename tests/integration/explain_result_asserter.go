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
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/immutable"
)

const (
	iterationsProp   = "iterations"
	docFetchesProp   = "docFetches"
	fieldFetchesProp = "fieldFetches"
	indexFetchesProp = "indexFetches"

	explainProp          = "explain"
	executionSuccessProp = "executionSuccess"
	sizeOfResultProp     = "sizeOfResult"
	planExecutionsProp   = "planExecutions"
	operationNodeProp    = "operationNode"
	filterMatchesProp    = "filterMatches"

	selectTopNodeProp = "selectTopNode"
	selectNodeProp    = "selectNode"
	scanNodeProp      = "scanNode"
	limitNodeProp     = "limitNode"
	orderNodeProp     = "orderNode"
	typeIndexJoinProp = "typeIndexJoin"
	typeJoinManyProp  = "typeJoinMany"
	typeJoinOneProp   = "typeJoinOne"
	orphanNodeProp    = "orphanNode"
	sequenceNodeProp  = "sequenceNode"
	rootProp          = "root"
	subTypeProp       = "subType"
)

type dataMap = map[string]any

// ExplainAsserter is a helper for asserting the result of an explain query.
// It asserts on metrics at specific levels of the explain tree using path navigation.
//
// For simple queries without joins, use NewExplainAsserter() without path.
// For queries with joins, you must specify the path (e.g., "root" or "subType").
type ExplainAsserter struct {
	path           []string
	iterations     immutable.Option[int]
	docFetches     immutable.Option[int]
	fieldFetches   immutable.Option[int]
	indexFetches   immutable.Option[int]
	filterMatches  immutable.Option[int]
	sizeOfResults  immutable.Option[int]
	planExecutions immutable.Option[uint64]
	nextLevel      *ExplainAsserter
}

// NewExplainAsserter creates an asserter for explain query results.
//
// For simple queries (no joins):
//
//	testUtils.NewExplainAsserter().WithIndexFetches(4)
//
// For queries with joins, specify the path to the level:
//
//	testUtils.NewExplainAsserter("root").WithIndexFetches(0)
//	testUtils.NewExplainAsserter("subType").WithIndexFetches(4)
//	testUtils.NewExplainAsserter("subType", "subType").WithIndexFetches(2) // nested
//
// For orphan node metrics (@exhaustive queries):
//
//	testUtils.NewExplainAsserter("orphanNode").WithDocFetches(5).WithIndexFetches(1)
//
// Path elements: "root" for parent side, "subType" for child side, "orphanNode" for orphan metrics.
func NewExplainAsserter(path ...string) *ExplainAsserter {
	return &ExplainAsserter{path: path}
}

func (a *ExplainAsserter) WithIterations(iterations int) *ExplainAsserter {
	a.iterations = immutable.Some(iterations)
	return a
}

func (a *ExplainAsserter) WithDocFetches(docFetches int) *ExplainAsserter {
	a.docFetches = immutable.Some(docFetches)
	return a
}

func (a *ExplainAsserter) WithFieldFetches(fieldFetches int) *ExplainAsserter {
	a.fieldFetches = immutable.Some(fieldFetches)
	return a
}

func (a *ExplainAsserter) WithIndexFetches(indexFetches int) *ExplainAsserter {
	a.indexFetches = immutable.Some(indexFetches)
	return a
}

func (a *ExplainAsserter) WithFilterMatches(filterMatches int) *ExplainAsserter {
	a.filterMatches = immutable.Some(filterMatches)
	return a
}

func (a *ExplainAsserter) WithSizeOfResults(sizeOfResults int) *ExplainAsserter {
	a.sizeOfResults = immutable.Some(sizeOfResults)
	return a
}

func (a *ExplainAsserter) WithPlanExecutions(planExecutions uint64) *ExplainAsserter {
	a.planExecutions = immutable.Some(planExecutions)
	return a
}

// WithLevel adds another level assertion.
// This allows chaining multiple level assertions:
//
//	testUtils.NewExplainAsserter("root").WithIndexFetches(0).
//		WithLevel("subType").WithIndexFetches(4)
func (a *ExplainAsserter) WithLevel(path ...string) *ExplainAsserter {
	next := &ExplainAsserter{path: path}
	current := a
	for current.nextLevel != nil {
		current = current.nextLevel
	}
	current.nextLevel = next
	return next
}

// Assert validates metrics in the explain result.
// For simple queries, reads from scanNode directly.
// For join queries, navigates to the specified path level.
func (a *ExplainAsserter) Assert(t testing.TB, result map[string]any) {
	explainNode, ok := result[explainProp].(dataMap)
	require.True(t, ok, "Expected explain node")

	assert.Equal(t, true, explainNode[executionSuccessProp], "Expected executionSuccess property")

	if a.sizeOfResults.HasValue() {
		actual := toUint64(explainNode[sizeOfResultProp])
		assert.Equal(t, uint64(a.sizeOfResults.Value()), actual,
			"Expected %d sizeOfResult, got %d", a.sizeOfResults.Value(), actual)
	}
	if a.planExecutions.HasValue() {
		actual := toUint64(explainNode[planExecutionsProp])
		assert.Equal(t, a.planExecutions.Value(), actual,
			"Expected %d planExecutions, got %d", a.planExecutions.Value(), actual)
	}

	operationNode := ConvertToArrayOfMaps(t, explainNode[operationNodeProp])
	require.Len(t, operationNode, 1)

	node, ok := operationNode[0][selectTopNodeProp].(dataMap)
	require.True(t, ok, "Expected selectTopNode")

	selectNode := navigateToSelectNode(t, node)

	if a.filterMatches.HasValue() {
		filterMatches, hasFilterMatches := selectNode[filterMatchesProp]
		require.True(t, hasFilterMatches, "Expected filterMatches property")
		assert.Equal(t, uint64(a.filterMatches.Value()), toUint64(filterMatches),
			"Expected %d filterMatches, got %d", a.filterMatches.Value(), filterMatches)
	}

	metricsNode := a.findMetricsNode(t, selectNode)
	a.assertMetrics(t, func(prop string) uint64 {
		return getMetric(metricsNode, prop)
	}, a.path)

	if a.nextLevel != nil {
		a.nextLevel.assertLevelOnly(t, selectNode)
	}
}

func (a *ExplainAsserter) assertLevelOnly(t testing.TB, selectNode dataMap) {
	metricsNode := a.findMetricsNode(t, selectNode)
	a.assertMetrics(t, func(prop string) uint64 {
		return getMetric(metricsNode, prop)
	}, a.path)

	if a.nextLevel != nil {
		a.nextLevel.assertLevelOnly(t, selectNode)
	}
}

func (a *ExplainAsserter) findMetricsNode(t testing.TB, selectNode dataMap) dataMap {
	if scanNode, has := selectNode[scanNodeProp].(dataMap); has {
		if len(a.path) > 0 {
			require.Fail(t, "Path specified but no typeIndexJoin found")
		}
		return scanNode
	}

	indexJoin, hasJoin := selectNode[typeIndexJoinProp].(dataMap)
	if !hasJoin {
		require.Fail(t, "Expected scanNode or typeIndexJoin")
		return nil
	}

	if len(a.path) == 0 {
		require.Fail(t, "Query has typeIndexJoin - must specify path (e.g., \"root\" or \"subType\")")
		return nil
	}

	if a.path[0] == orphanNodeProp {
		orphanNode := findOrphanNodeInJoin(indexJoin)
		require.NotNil(t, orphanNode, "Expected orphanNode in typeIndexJoin")
		return orphanNode
	}

	// sequenceNode wraps [joinNode, orphanNode] or [orphanNode, joinNode] for @exhaustive.
	// Find the join child (non-orphan) in the array.
	indexJoin = unwrapSequenceNode(indexJoin)

	// orphanNode (wrapper mode) wraps the join for secondary parent @exhaustive queries.
	if orphan, hasOrphan := indexJoin[orphanNodeProp].(dataMap); hasOrphan {
		indexJoin = orphan
	}

	targetNode := navigateToLevel(indexJoin, a.path)
	require.NotNil(t, targetNode, "Could not navigate to level: %v", a.path)

	scanNode := findScanNodeAtLevel(targetNode)
	require.NotNil(t, scanNode, "No scanNode found at level: %v", a.path)

	return scanNode
}

func (a *ExplainAsserter) assertMetrics(t testing.TB, getMetricFn func(string) uint64, path []string) {
	levelInfo := ""
	if len(path) > 0 {
		levelInfo = " at level " + formatPath(path)
	}

	if a.iterations.HasValue() {
		actual := getMetricFn(iterationsProp)
		assert.Equal(t, uint64(a.iterations.Value()), actual,
			"Expected %d iterations%s, got %d", a.iterations.Value(), levelInfo, actual)
	}
	if a.docFetches.HasValue() {
		actual := getMetricFn(docFetchesProp)
		assert.Equal(t, uint64(a.docFetches.Value()), actual,
			"Expected %d docFetches%s, got %d", a.docFetches.Value(), levelInfo, actual)
	}
	if a.fieldFetches.HasValue() {
		actual := getMetricFn(fieldFetchesProp)
		assert.Equal(t, uint64(a.fieldFetches.Value()), actual,
			"Expected %d fieldFetches%s, got %d", a.fieldFetches.Value(), levelInfo, actual)
	}
	if a.indexFetches.HasValue() {
		actual := getMetricFn(indexFetchesProp)
		assert.Equal(t, uint64(a.indexFetches.Value()), actual,
			"Expected %d indexFetches%s, got %d", a.indexFetches.Value(), levelInfo, actual)
	}
}

func formatPath(path []string) string {
	if len(path) == 0 {
		return ""
	}
	var result strings.Builder
	result.WriteString(path[0])
	for i := 1; i < len(path); i++ {
		result.WriteString("/" + path[i])
	}
	return result.String()
}

// navigateToSelectNode finds the selectNode, handling orderNode and limitNode wrappers.
func navigateToSelectNode(t testing.TB, node dataMap) dataMap {
	node = unwrapNode(node, limitNodeProp, orderNodeProp)
	selectNode, ok := node[selectNodeProp].(dataMap)
	require.True(t, ok, "Expected selectNode")
	return selectNode
}

// navigateToLevel follows the path through the explain tree.
func navigateToLevel(node dataMap, path []string) dataMap {
	current := node

	for _, step := range path {
		joinNode := getJoinNode(current)

		switch step {
		case rootProp:
			if root, has := joinNode[rootProp].(dataMap); has {
				current = root
			} else {
				return nil
			}
		case subTypeProp:
			if sub, has := joinNode[subTypeProp].(dataMap); has {
				current = navigateThroughSelectTop(sub)
			} else {
				return nil
			}
		default:
			if next, has := current[step].(dataMap); has {
				current = next
			} else {
				return nil
			}
		}
	}

	return current
}

// navigateThroughSelectTop handles the selectTopNode -> selectNode -> typeIndexJoin chain.
func navigateThroughSelectTop(node dataMap) dataMap {
	node = unwrapNode(node, selectTopNodeProp, limitNodeProp, orderNodeProp, selectNodeProp)
	if indexJoin, has := node[typeIndexJoinProp].(dataMap); has {
		return indexJoin
	}
	return node
}

// unwrapNode navigates through wrapper nodes in order.
func unwrapNode(node dataMap, wrappers ...string) dataMap {
	for _, wrapper := range wrappers {
		if inner, has := node[wrapper].(dataMap); has {
			node = inner
		}
	}
	return node
}

// getJoinNode returns the typeJoinMany or typeJoinOne node, or the node itself.
func getJoinNode(node dataMap) dataMap {
	if jm, has := node[typeJoinManyProp].(dataMap); has {
		return jm
	}
	if jo, has := node[typeJoinOneProp].(dataMap); has {
		return jo
	}
	return node
}

// findOrphanNodeInJoin locates the orphanNode metrics in the typeIndexJoin.
// Handles both wrapper mode (orphanNode directly in the join) and sequenceNode mode
// (orphanNode as a child element in the sequenceNode array).
func findOrphanNodeInJoin(indexJoin dataMap) dataMap {
	if orphan, has := indexJoin[orphanNodeProp].(dataMap); has {
		return orphan
	}
	if seqArr, ok := indexJoin[sequenceNodeProp].([]map[string]any); ok {
		for _, child := range seqArr {
			if orphan, has := child[orphanNodeProp].(dataMap); has {
				return orphan
			}
		}
	}
	if seqArr, ok := indexJoin[sequenceNodeProp].([]any); ok {
		for _, child := range seqArr {
			if childMap, ok := child.(dataMap); ok {
				if orphan, has := childMap[orphanNodeProp].(dataMap); has {
					return orphan
				}
			}
		}
	}
	return nil
}

// unwrapSequenceNode finds the join child (non-orphan) inside a sequenceNode array.
// Returns the original node if no sequenceNode is present.
func unwrapSequenceNode(node dataMap) dataMap {
	// Go client: []map[string]any
	if seqArr, ok := node[sequenceNodeProp].([]map[string]any); ok {
		for _, child := range seqArr {
			if _, isOrphan := child[orphanNodeProp]; !isOrphan {
				return child
			}
		}
	}
	// HTTP/CLI/JS clients: []any (JSON deserialization)
	if seqArr, ok := node[sequenceNodeProp].([]any); ok {
		for _, child := range seqArr {
			if childMap, ok := child.(dataMap); ok {
				if _, isOrphan := childMap[orphanNodeProp]; !isOrphan {
					return childMap
				}
			}
		}
	}
	return node
}

// findScanNodeAtLevel finds the scanNode at the current level (not recursively).
func findScanNodeAtLevel(node dataMap) dataMap {
	if scan, has := node[scanNodeProp].(dataMap); has {
		return scan
	}
	if jm, has := node[typeJoinManyProp].(dataMap); has {
		if root, has := jm[rootProp].(dataMap); has {
			if scan, has := root[scanNodeProp].(dataMap); has {
				return scan
			}
		}
	}
	if jo, has := node[typeJoinOneProp].(dataMap); has {
		if root, has := jo[rootProp].(dataMap); has {
			if scan, has := root[scanNodeProp].(dataMap); has {
				return scan
			}
		}
	}
	return nil
}

// getMetric extracts a metric value from a node.
func getMetric(node dataMap, prop string) uint64 {
	return toUint64(node[prop])
}

// toUint64 converts a value to uint64.
// It handles both uint64 (Go embedded client) and float64 (HTTP/CLI/JS/C clients
// where JSON numbers are deserialized as float64).
func toUint64(val any) uint64 {
	switch v := val.(type) {
	case uint64:
		return v
	case json.Number:
		n, _ := v.Int64()
		return uint64(n)
	case float64:
		return uint64(v)
	default:
		return 0
	}
}
