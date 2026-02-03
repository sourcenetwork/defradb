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
)

type dataMap = map[string]any

// ExplainAsserter is a helper for asserting the result of an explain query.
// It can assert on aggregated metrics across all nodes, or on specific levels
// of the explain tree using path navigation.
//
// When path is empty (NewExplainAsserter), metrics are aggregated across all scan nodes.
// When path is set (NewLevelAsserter), metrics are read from the specific level.
type ExplainAsserter struct {
	path           []string
	iterations     immutable.Option[int]
	docFetches     immutable.Option[int]
	fieldFetches   immutable.Option[int]
	indexFetches   immutable.Option[int]
	filterMatches  immutable.Option[int]
	sizeOfResults  immutable.Option[int]
	planExecutions immutable.Option[uint64]
}

// NewExplainAsserter creates an asserter for explain query results.
//
// When called without arguments, metrics are aggregated across all scan nodes:
//
//	testUtils.NewExplainAsserter().WithIndexFetches(4)
//
// When called with path arguments, metrics are read from that specific level:
//
//	testUtils.NewExplainAsserter("root").WithIndexFetches(0)
//	testUtils.NewExplainAsserter("subType").WithIndexFetches(4)
//	testUtils.NewExplainAsserter("subType", "subType").WithIndexFetches(2) // nested
//
// Path elements: "root" for parent side, "subType" for child side.
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

// Deprecated: WithOrder is kept for backward compatibility but no longer affects assertion behavior.
// The new recursive assertion automatically handles orderNode wrappers.
func (a *ExplainAsserter) WithOrder() *ExplainAsserter {
	return a
}

// Deprecated: WithLimit is kept for backward compatibility but no longer affects assertion behavior.
// The new recursive assertion automatically handles limitNode wrappers.
func (a *ExplainAsserter) WithLimit() *ExplainAsserter {
	return a
}

// WithLevel adds another level assertion and returns a MultiLevelAsserter.
// This allows chaining multiple level assertions:
//
//	testUtils.NewLevelAsserter("root").WithIndexFetches(0).
//		WithLevel("subType").WithIndexFetches(4)
func (a *ExplainAsserter) WithLevel(path ...string) *MultiLevelAsserter {
	return &MultiLevelAsserter{
		levels:  []*ExplainAsserter{a},
		current: &ExplainAsserter{path: path},
	}
}

// Assert validates metrics in the explain result.
// If path is empty, aggregates metrics across all scan nodes.
// If path is set, reads metrics from the specific level.
func (a *ExplainAsserter) Assert(t testing.TB, result map[string]any) {
	explainNode, ok := result["explain"].(dataMap)
	require.True(t, ok, "Expected explain node")

	assert.Equal(t, true, explainNode["executionSuccess"], "Expected executionSuccess property")

	if a.sizeOfResults.HasValue() {
		actual := explainNode["sizeOfResult"]
		assert.Equal(t, a.sizeOfResults.Value(), actual,
			"Expected %d sizeOfResult, got %d", a.sizeOfResults.Value(), actual)
	}
	if a.planExecutions.HasValue() {
		actual := explainNode["planExecutions"]
		assert.Equal(t, a.planExecutions.Value(), actual,
			"Expected %d planExecutions, got %d", a.planExecutions.Value(), actual)
	}

	operationNode := ConvertToArrayOfMaps(t, explainNode["operationNode"])
	require.Len(t, operationNode, 1)

	node, ok := operationNode[0]["selectTopNode"].(dataMap)
	require.True(t, ok, "Expected selectTopNode")

	selectNode := navigateToSelectNode(t, node)

	if a.filterMatches.HasValue() {
		filterMatches, hasFilterMatches := selectNode["filterMatches"]
		require.True(t, hasFilterMatches, "Expected filterMatches property")
		assert.Equal(t, uint64(a.filterMatches.Value()), filterMatches,
			"Expected %d filterMatches, got %d", a.filterMatches.Value(), filterMatches)
	}

	// Determine how to get metrics based on whether path is set
	if len(a.path) == 0 {
		// Aggregate mode: sum metrics across all scan nodes
		a.assertAggregatedMetrics(t, selectNode)
	} else {
		// Level mode: get metrics from specific level
		a.assertLevelMetrics(t, selectNode)
	}
}

func (a *ExplainAsserter) assertAggregatedMetrics(t testing.TB, selectNode dataMap) {
	_, hasScanNode := selectNode["scanNode"].(dataMap)
	if indexJoin, isJoin := selectNode["typeIndexJoin"].(dataMap); isJoin {
		_, hasScanNode = findScanNodeInJoin(indexJoin)
	}
	require.True(t, hasScanNode, "Expected scanNode")

	if a.iterations.HasValue() {
		actual := aggregateMetricFromNode(selectNode, iterationsProp)
		assert.Equal(t, uint64(a.iterations.Value()), actual,
			"Expected %d iterations, got %d", a.iterations.Value(), actual)
	}
	if a.docFetches.HasValue() {
		actual := aggregateMetricFromNode(selectNode, docFetchesProp)
		assert.Equal(t, uint64(a.docFetches.Value()), actual,
			"Expected %d docFetches, got %d", a.docFetches.Value(), actual)
	}
	if a.fieldFetches.HasValue() {
		actual := aggregateMetricFromNode(selectNode, fieldFetchesProp)
		assert.Equal(t, uint64(a.fieldFetches.Value()), actual,
			"Expected %d fieldFetches, got %d", a.fieldFetches.Value(), actual)
	}
	if a.indexFetches.HasValue() {
		actual := aggregateMetricFromNode(selectNode, indexFetchesProp)
		assert.Equal(t, uint64(a.indexFetches.Value()), actual,
			"Expected %d indexFetches, got %d", a.indexFetches.Value(), actual)
	}
}

func (a *ExplainAsserter) assertLevelMetrics(t testing.TB, selectNode dataMap) {
	indexJoin, hasJoin := selectNode["typeIndexJoin"].(dataMap)
	if !hasJoin {
		require.Fail(t, "Expected typeIndexJoin for level assertion")
	}

	if orphanNode, hasOrphan := indexJoin["orphanNode"].(dataMap); hasOrphan {
		indexJoin = orphanNode
	}

	targetNode := navigateToLevel(indexJoin, a.path)
	require.NotNil(t, targetNode, "Could not navigate to level: %v", a.path)

	scanNode := findScanNodeAtLevel(targetNode)
	require.NotNil(t, scanNode, "No scanNode found at level: %v", a.path)

	if a.iterations.HasValue() {
		actual := getMetric(scanNode, iterationsProp)
		assert.Equal(t, uint64(a.iterations.Value()), actual,
			"Expected %d iterations at level %v, got %d", a.iterations.Value(), a.path, actual)
	}
	if a.docFetches.HasValue() {
		actual := getMetric(scanNode, docFetchesProp)
		assert.Equal(t, uint64(a.docFetches.Value()), actual,
			"Expected %d docFetches at level %v, got %d", a.docFetches.Value(), a.path, actual)
	}
	if a.fieldFetches.HasValue() {
		actual := getMetric(scanNode, fieldFetchesProp)
		assert.Equal(t, uint64(a.fieldFetches.Value()), actual,
			"Expected %d fieldFetches at level %v, got %d", a.fieldFetches.Value(), a.path, actual)
	}
	if a.indexFetches.HasValue() {
		actual := getMetric(scanNode, indexFetchesProp)
		assert.Equal(t, uint64(a.indexFetches.Value()), actual,
			"Expected %d indexFetches at level %v, got %d", a.indexFetches.Value(), a.path, actual)
	}
}

// MultiLevelAsserter allows asserting on multiple levels in a single assertion.
type MultiLevelAsserter struct {
	levels  []*ExplainAsserter
	current *ExplainAsserter
}

func (m *MultiLevelAsserter) WithIterations(iterations int) *MultiLevelAsserter {
	m.current.iterations = immutable.Some(iterations)
	return m
}

func (m *MultiLevelAsserter) WithDocFetches(docFetches int) *MultiLevelAsserter {
	m.current.docFetches = immutable.Some(docFetches)
	return m
}

func (m *MultiLevelAsserter) WithFieldFetches(fieldFetches int) *MultiLevelAsserter {
	m.current.fieldFetches = immutable.Some(fieldFetches)
	return m
}

func (m *MultiLevelAsserter) WithIndexFetches(indexFetches int) *MultiLevelAsserter {
	m.current.indexFetches = immutable.Some(indexFetches)
	return m
}

// WithLevel adds another level assertion.
func (m *MultiLevelAsserter) WithLevel(path ...string) *MultiLevelAsserter {
	m.levels = append(m.levels, m.current)
	m.current = &ExplainAsserter{path: path}
	return m
}

// Assert validates metrics at all specified levels of the explain result.
func (m *MultiLevelAsserter) Assert(t testing.TB, result map[string]any) {
	allLevels := append(m.levels, m.current)
	for _, level := range allLevels {
		level.Assert(t, result)
	}
}

// navigateToSelectNode finds the selectNode, handling orderNode and limitNode wrappers.
func navigateToSelectNode(t testing.TB, node dataMap) dataMap {
	if limitNode, has := node["limitNode"].(dataMap); has {
		node = limitNode
	}
	if orderNode, has := node["orderNode"].(dataMap); has {
		node = orderNode
	}
	selectNode, ok := node["selectNode"].(dataMap)
	require.True(t, ok, "Expected selectNode")
	return selectNode
}

// navigateToLevel follows the path through the explain tree.
func navigateToLevel(node dataMap, path []string) dataMap {
	current := node

	for _, step := range path {
		var joinNode dataMap
		if jm, has := current["typeJoinMany"].(dataMap); has {
			joinNode = jm
		} else if jo, has := current["typeJoinOne"].(dataMap); has {
			joinNode = jo
		} else {
			joinNode = current
		}

		switch step {
		case "root":
			if root, has := joinNode["root"].(dataMap); has {
				current = root
			} else {
				return nil
			}
		case "subType":
			if subType, has := joinNode["subType"].(dataMap); has {
				current = navigateThroughSelectTop(subType)
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
	if selectTop, has := node["selectTopNode"].(dataMap); has {
		node = selectTop
	}
	if limitNode, has := node["limitNode"].(dataMap); has {
		node = limitNode
	}
	if orderNode, has := node["orderNode"].(dataMap); has {
		node = orderNode
	}
	if selectNode, has := node["selectNode"].(dataMap); has {
		node = selectNode
	}
	if indexJoin, has := node["typeIndexJoin"].(dataMap); has {
		return indexJoin
	}
	return node
}

// findScanNodeAtLevel finds the scanNode at the current level (not recursively).
func findScanNodeAtLevel(node dataMap) dataMap {
	if scanNode, has := node["scanNode"].(dataMap); has {
		return scanNode
	}
	for _, joinType := range []string{"typeJoinMany", "typeJoinOne"} {
		if joinNode, has := node[joinType].(dataMap); has {
			if root, hasRoot := joinNode["root"].(dataMap); hasRoot {
				if scanNode, hasScan := root["scanNode"].(dataMap); hasScan {
					return scanNode
				}
			}
		}
	}
	return nil
}

// getMetric extracts a metric value from a node.
func getMetric(node dataMap, prop string) uint64 {
	if val, has := node[prop]; has {
		if num, ok := val.(uint64); ok {
			return num
		}
	}
	return 0
}

// findScanNodeInJoin finds a scanNode within a join structure (for validation).
func findScanNodeInJoin(indexJoin dataMap) (dataMap, bool) {
	// Check for orphanNode wrapper
	if orphanNode, hasOrphan := indexJoin["orphanNode"].(dataMap); hasOrphan {
		indexJoin = orphanNode
	}

	for _, joinType := range []string{"typeJoinMany", "typeJoinOne"} {
		if joinNode, has := indexJoin[joinType].(dataMap); has {
			if root, hasRoot := joinNode["root"].(dataMap); hasRoot {
				if scanNode, hasScan := root["scanNode"].(dataMap); hasScan {
					return scanNode, true
				}
			}
		}
	}
	return nil, false
}

// aggregateMetricFromNode recursively sums a metric from all scanNodes in the tree.
func aggregateMetricFromNode(node dataMap, prop string) uint64 {
	var total uint64

	// Check if this node has the metric directly (scanNode)
	if scanNode, has := node["scanNode"].(dataMap); has {
		if val, hasVal := scanNode[prop]; hasVal {
			if num, ok := val.(uint64); ok {
				total += num
			}
		}
	}

	// Check for typeIndexJoin
	if indexJoin, has := node["typeIndexJoin"].(dataMap); has {
		total += aggregateMetricFromNode(indexJoin, prop)
	}

	// Check for orphanNode
	if orphanNode, has := node["orphanNode"].(dataMap); has {
		total += aggregateMetricFromNode(orphanNode, prop)
	}

	// Check for join types
	for _, joinType := range []string{"typeJoinMany", "typeJoinOne"} {
		if joinNode, has := node[joinType].(dataMap); has {
			// Process root
			if root, hasRoot := joinNode["root"].(dataMap); hasRoot {
				total += aggregateMetricFromNode(root, prop)
			}
			// Process subType
			if subType, hasSubType := joinNode["subType"].(dataMap); hasSubType {
				total += aggregateMetricFromNode(subType, prop)
			}
		}
	}

	// Handle selectTopNode wrapper
	if selectTop, has := node["selectTopNode"].(dataMap); has {
		total += aggregateMetricFromNode(selectTop, prop)
	}

	// Handle limitNode wrapper
	if limitNode, has := node["limitNode"].(dataMap); has {
		total += aggregateMetricFromNode(limitNode, prop)
	}

	// Handle orderNode wrapper
	if orderNode, has := node["orderNode"].(dataMap); has {
		total += aggregateMetricFromNode(orderNode, prop)
	}

	// Handle selectNode wrapper
	if selectNode, has := node["selectNode"].(dataMap); has {
		total += aggregateMetricFromNode(selectNode, prop)
	}

	return total
}
