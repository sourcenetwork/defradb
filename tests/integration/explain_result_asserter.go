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
	"encoding/json"
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

// ExplainResultAsserter is a helper for asserting the result of an explain query.
// It allows asserting on a selected set of properties.
type ExplainResultAsserter struct {
	iterations     immutable.Option[int]
	docFetches     immutable.Option[int]
	fieldFetches   immutable.Option[int]
	indexFetches   immutable.Option[int]
	filterMatches  immutable.Option[int]
	sizeOfResults  immutable.Option[int]
	planExecutions immutable.Option[uint64]
}

func readNumberProp(t testing.TB, val any, prop string) uint64 {
	switch v := val.(type) {
	case uint64:
		return v
	case json.Number:
		n, err := v.Int64()
		require.NoError(t, err, fmt.Sprintf("Expected %s property to be a uint64", prop))
		return uint64(n)
	default:
		require.Fail(t, fmt.Sprintf("Unexpected type for %s property: %T", prop, val))
	}
	return 0
}

func (a *ExplainResultAsserter) Assert(t testing.TB, result map[string]any) {
	explainNode, ok := result["explain"].(dataMap)
	require.True(t, ok, "Expected explain none")
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
	selectTopNode, ok := operationNode[0]["selectTopNode"].(dataMap)
	require.True(t, ok, "Expected selectTopNode")
	selectNode, ok := selectTopNode["selectNode"].(dataMap)
	require.True(t, ok, "Expected selectNode")

	if a.filterMatches.HasValue() {
		filterMatches, hasFilterMatches := selectNode["filterMatches"]
		require.True(t, hasFilterMatches, "Expected filterMatches property")
		assert.Equal(t, uint64(a.filterMatches.Value()), filterMatches,
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
		actual := readNumberProp(t, val, prop)
		if subScanNode[prop] != nil {
			actual += readNumberProp(t, subScanNode[prop], "subTypeScanNode."+prop)
		}
		return actual
	}

	if a.iterations.HasValue() {
		actual := getScanNodesProp(iterationsProp)
		assert.Equal(t, uint64(a.iterations.Value()), actual,
			"Expected %d iterations, got %d", a.iterations.Value(), actual)
	}
	if a.docFetches.HasValue() {
		actual := getScanNodesProp(docFetchesProp)
		assert.Equal(t, uint64(a.docFetches.Value()), actual,
			"Expected %d docFetches, got %d", a.docFetches.Value(), actual)
	}
	if a.fieldFetches.HasValue() {
		actual := getScanNodesProp(fieldFetchesProp)
		assert.Equal(t, uint64(a.fieldFetches.Value()), actual,
			"Expected %d fieldFetches, got %d", a.fieldFetches.Value(), actual)
	}
	if a.indexFetches.HasValue() {
		actual := getScanNodesProp(indexFetchesProp)
		assert.Equal(t, uint64(a.indexFetches.Value()), actual,
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
