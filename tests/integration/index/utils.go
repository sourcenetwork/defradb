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
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type dataMap = map[string]any

type explainResultAsserter struct {
	iterations    int
	docFetches    int
	filterMatches int
}

func (a explainResultAsserter) Assert(t *testing.T, result []dataMap) {
	require.Len(t, result, 1, "Expected len(result) = 1, got %d", len(result))
	explainNode, ok := result[0]["explain"].(dataMap)
	require.True(t, ok, "Expected explain none")
	assert.Equal(t, explainNode["executionSuccess"], true, "Expected executionSuccess property")
	assert.Equal(t, explainNode["sizeOfResult"], 1, "Expected sizeOfResult property")
	assert.Equal(t, explainNode["planExecutions"], uint64(2), "Expected planExecutions property")
	selectTopNode, ok := explainNode["selectTopNode"].(dataMap)
	require.True(t, ok, "Expected selectTopNode", "Expected selectTopNode")
	selectNode, ok := selectTopNode["selectNode"].(dataMap)
	require.True(t, ok, "Expected selectNode", "Expected selectNode")
	scanNode, ok := selectNode["scanNode"].(dataMap)
	require.True(t, ok, "Expected scanNode", "Expected scanNode")
	iterations, hasIterations := scanNode["iterations"]
	require.True(t, hasIterations, "Expected iterations property")
	assert.Equal(t, iterations, uint64(a.iterations),
		"Expected %d iterations, got %d", a.iterations, iterations)
	docFetches, hasDocFetches := scanNode["docFetches"]
	require.True(t, hasDocFetches, "Expected docFetches property")
	assert.Equal(t, docFetches, uint64(a.docFetches),
		"Expected %d docFetches, got %d", a.docFetches, docFetches)
	filterMatches, hasFilterMatches := selectNode["filterMatches"]
	require.True(t, hasFilterMatches, "Expected filterMatches property")
	assert.Equal(t, filterMatches, uint64(a.filterMatches),
		"Expected %d filterMatches, got %d", a.filterMatches, filterMatches)
}

func newExplainAsserter(iterations, docFetched, filterMatcher int) *explainResultAsserter {
	return &explainResultAsserter{
		iterations:    iterations,
		docFetches:    docFetched,
		filterMatches: filterMatcher,
	}
}

func createUserDocs() []testUtils.CreateDoc {
	return []testUtils.CreateDoc{
		{
			CollectionID: 0,
			Doc: `{
					"name": "John"
				}`,
		},
		{
			CollectionID: 0,
			Doc: `{
					"name": "Islam"
				}`,
		},
		{
			CollectionID: 0,
			Doc: `{
					"name": "Andy"
				}`,
		},
		{
			CollectionID: 0,
			Doc: `{
					"name": "Shahzad"
				}`,
		},
		{
			CollectionID: 0,
			Doc: `{
					"name": "Fred"
				}`,
		},
		{
			CollectionID: 0,
			Doc: `{
					"name": "Orpheus"
				}`,
		},
		{
			CollectionID: 0,
			Doc: `{
					"name": "Addo"
				}`,
		},
	}
}
