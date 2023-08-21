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
	"strings"
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type dataMap = map[string]any

type ExplainResultAsserter struct {
	iterations     immutable.Option[int]
	docFetches     immutable.Option[int]
	fieldFetches   immutable.Option[int]
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
	require.True(t, ok, "Expected scanNode")

	if a.iterations.HasValue() {
		iterations, hasIterations := scanNode["iterations"]
		require.True(t, hasIterations, "Expected iterations property")
		assert.Equal(t, iterations, uint64(a.iterations.Value()),
			"Expected %d iterations, got %d", a.iterations.Value(), iterations)
	}
	if a.docFetches.HasValue() {
		docFetches, hasDocFetches := scanNode["docFetches"]
		require.True(t, hasDocFetches, "Expected docFetches property")
		assert.Equal(t, docFetches, uint64(a.docFetches.Value()),
			"Expected %d docFetches, got %d", a.docFetches.Value(), docFetches)
	}
	if a.fieldFetches.HasValue() {
		fieldFetches, hasFieldFetches := scanNode["fieldFetches"]
		require.True(t, hasFieldFetches, "Expected fieldFetches property")
		assert.Equal(t, fieldFetches, uint64(a.fieldFetches.Value()),
			"Expected %d fieldFetches, got %d", a.fieldFetches.Value(), fieldFetches)
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

func getDocs() []map[string]any {
	return []map[string]any{
		{
			"name":     "Shahzad",
			"age":      20,
			"verified": false,
			"email":    "shahzad@gmail.com",
		},
		{
			"name":     "Fred",
			"age":      28,
			"verified": false,
			"email":    "fred@gmail.com",
		},
		{
			"name":     "John",
			"age":      30,
			"verified": false,
			"email":    "john@gmail.com",
		},
		{
			"name":     "Islam",
			"age":      32,
			"verified": false,
			"email":    "islam@gmail.com",
		},
		{
			"name":     "Andy",
			"age":      33,
			"verified": true,
			"email":    "andy@gmail.com",
		},
		{
			"name":     "Addo",
			"age":      42,
			"verified": true,
			"email":    "addo@gmail.com",
		},
		{
			"name":     "Keenan",
			"age":      48,
			"verified": true,
			"email":    "keenan@gmail.com",
		},
		{
			"name":     "Chris",
			"age":      55,
			"verified": true,
			"email":    "chris@gmail.com",
		},
	}
}

// createSchemaWithDocs returns UpdateSchema action and CreateDoc actions
// with the documents that match the schema.
// The schema is parsed to get the list of properties, and the docs
// are created with the same properties.
// This allows us to have only one large list of docs with predefined
// properties, and create schemas with different properties from it.
func createSchemaWithDocs(schema string) []any {
	docs := getDocs()
	actions := make([]any, 0, len(docs)+1)
	actions = append(actions, testUtils.SchemaUpdate{Schema: schema})
	props := getSchemaProps(schema)
	for _, doc := range docs {
		docDesc := makeDocWithProps(doc, props)
		actions = append(actions, testUtils.CreateDoc{CollectionID: 0, Doc: docDesc})
	}
	return actions
}

func makeDocWithProps(doc map[string]any, props []string) string {
	sb := strings.Builder{}
	sb.WriteString("{\n")
	for i := range props {
		format := `"%s": %v`
		if _, isStr := doc[props[i]].(string); isStr {
			format = `"%s": "%v"`
		}
		sb.WriteString(fmt.Sprintf(format, props[i], doc[props[i]]))
		if i != len(props)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	sb.WriteString("}")
	return sb.String()
}

func getSchemaProps(schema string) []string {
	props := make([]string, 0)
	lines := strings.Split(schema, "\n")
	for _, line := range lines {
		pos := strings.Index(line, ":")
		if pos != -1 {
			props = append(props, strings.TrimSpace(line[:pos]))
		}
	}
	return props
}
