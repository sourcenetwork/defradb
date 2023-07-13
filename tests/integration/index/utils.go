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
