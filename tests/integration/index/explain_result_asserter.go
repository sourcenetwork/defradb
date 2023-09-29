// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package index

import (
	"fmt"
	"testing"

	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	iterationsProp   = "iterations"
	docFetchesProp   = "docFetches"
	fieldFetchesProp = "fieldFetches"
	indexFetchesProp = "indexFetches"
)

type dataMap = map[string]any

type ExplainResultAsserter struct {
	iterations     immutable.Option[int]
	docFetches     immutable.Option[int]
	fieldFetches   immutable.Option[int]
	indexFetches   immutable.Option[int]
	filterMatches  immutable.Option[int]
	sizeOfResults  immutable.Option[int]
	planExecutions immutable.Option[uint64]
}

func (a *ExplainResultAsserter) Assert(t *testing.T, result []dataMap) {
	require.Len(t, result, 1, "Expected len(result) = 1, got %d", len(result))
	explainNode, ok := result[0]["explain"].(dataMap)
	require.True(t, ok, "Expected explain none")
	assert.Equal(t, explainNode["executionSuccess"], true, "Expected executionSuccess property")
	if a.sizeOfResults.HasValue() {
		actual := explainNode["sizeOfResult"]
		assert.Equal(t, actual, a.sizeOfResults.Value(),
			"Expected %d sizeOfResult, got %d", a.sizeOfResults.Value(), actual)
	}
	if a.planExecutions.HasValue() {
		actual := explainNode["planExecutions"]
		assert.Equal(t, actual, a.planExecutions.Value(),
			"Expected %d planExecutions, got %d", a.planExecutions.Value(), actual)
	}
	selectTopNode, ok := explainNode["selectTopNode"].(dataMap)
	require.True(t, ok, "Expected selectTopNode")
	selectNode, ok := selectTopNode["selectNode"].(dataMap)
	require.True(t, ok, "Expected selectNode")

	if a.filterMatches.HasValue() {
		filterMatches, hasFilterMatches := selectNode["filterMatches"]
		require.True(t, hasFilterMatches, "Expected filterMatches property")
		assert.Equal(t, filterMatches, uint64(a.filterMatches.Value()),
			"Expected %d filterMatches, got %d", a.filterMatches, filterMatches)
	}

	scanNode, ok := selectNode["scanNode"].(dataMap)
	subScanNode := map[string]any{}
	if indexJoin, isJoin := selectNode["typeIndexJoin"].(dataMap); isJoin {
		scanNode, ok = indexJoin["scanNode"].(dataMap)
		subScanNode, _ = indexJoin["subTypeScanNode"].(dataMap)
	}
	require.True(t, ok, "Expected scanNode")

	getScanNodesProp := func(prop string) uint64 {
		val, hasProp := scanNode[prop]
		require.True(t, hasProp, fmt.Sprintf("Expected %s property", prop))
		actual := val.(uint64)
		if subScanNode[prop] != nil {
			actual += subScanNode[prop].(uint64)
		}
		return actual
	}

	if a.iterations.HasValue() {
		actual := getScanNodesProp(iterationsProp)
		assert.Equal(t, actual, uint64(a.iterations.Value()),
			"Expected %d iterations, got %d", a.iterations.Value(), actual)
	}
	if a.docFetches.HasValue() {
		actual := getScanNodesProp(docFetchesProp)
		assert.Equal(t, actual, uint64(a.docFetches.Value()),
			"Expected %d docFetches, got %d", a.docFetches.Value(), actual)
	}
	if a.fieldFetches.HasValue() {
		actual := getScanNodesProp(fieldFetchesProp)
		assert.Equal(t, actual, uint64(a.fieldFetches.Value()),
			"Expected %d fieldFetches, got %d", a.fieldFetches.Value(), actual)
	}
	if a.indexFetches.HasValue() {
		actual := getScanNodesProp(indexFetchesProp)
		assert.Equal(t, actual, uint64(a.indexFetches.Value()),
			"Expected %d indexFetches, got %d", a.indexFetches.Value(), actual)
	}
}

func (a *ExplainResultAsserter) WithIterations(iterations int) *ExplainResultAsserter {
	a.iterations = immutable.Some[int](iterations)
	return a
}

func (a *ExplainResultAsserter) WithDocFetches(docFetches int) *ExplainResultAsserter {
	a.docFetches = immutable.Some[int](docFetches)
	return a
}

func (a *ExplainResultAsserter) WithFieldFetches(fieldFetches int) *ExplainResultAsserter {
	a.fieldFetches = immutable.Some[int](fieldFetches)
	return a
}

func (a *ExplainResultAsserter) WithIndexFetches(indexFetches int) *ExplainResultAsserter {
	a.indexFetches = immutable.Some[int](indexFetches)
	return a
}

func (a *ExplainResultAsserter) WithFilterMatches(filterMatches int) *ExplainResultAsserter {
	a.filterMatches = immutable.Some[int](filterMatches)
	return a
}

func (a *ExplainResultAsserter) WithSizeOfResults(sizeOfResults int) *ExplainResultAsserter {
	a.sizeOfResults = immutable.Some[int](sizeOfResults)
	return a
}

func (a *ExplainResultAsserter) WithPlanExecutions(planExecutions uint64) *ExplainResultAsserter {
	a.planExecutions = immutable.Some[uint64](planExecutions)
	return a
}

func NewExplainAsserter() *ExplainResultAsserter {
	return &ExplainResultAsserter{}
}
