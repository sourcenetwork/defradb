// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.

package graphql

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExecuteIntrospectionWithGraphQLGoTools(t *testing.T) {
	parser, err := NewParser(false)
	require.NoError(t, err)

	result := parser.ExecuteIntrospection(context.Background(), `
		query Introspection {
			__schema {
				queryType { name }
				types { ...TypeName }
			}
			query: __type(name: "Query") {
				name
				kind
				fields { name }
			}
		}
		fragment TypeName on __Type { name kind }
	`)
	require.Empty(t, result.GQL.Errors)
	data, ok := result.GQL.Data.(map[string]any)
	require.True(t, ok)
	schema, ok := data["__schema"].(map[string]any)
	require.True(t, ok)
	require.NotEmpty(t, schema["types"])
	query, ok := data["query"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "Query", query["name"])
	require.Equal(t, "OBJECT", query["kind"])
}

func TestExecuteStandardIntrospectionQueryWithGraphQLGoTools(t *testing.T) {
	parser, err := NewParser(false)
	require.NoError(t, err)

	result := parser.ExecuteIntrospection(context.Background(), standardIntrospectionQuery)
	require.Empty(t, result.GQL.Errors)
	data, ok := result.GQL.Data.(map[string]any)
	require.True(t, ok)
	schema, ok := data["__schema"].(map[string]any)
	require.True(t, ok)
	require.NotEmpty(t, schema["types"])
	require.NotEmpty(t, schema["directives"])
}

const standardIntrospectionQuery = `
query IntrospectionQuery {
  __schema {
    queryType { name }
    mutationType { name }
    subscriptionType { name }
    types { ...FullType }
    directives { name description locations args { ...InputValue } }
  }
}
fragment FullType on __Type {
  kind name description
  fields(includeDeprecated: true) {
    name description args { ...InputValue } type { ...TypeRef }
    isDeprecated deprecationReason
  }
  inputFields { ...InputValue }
  interfaces { ...TypeRef }
  enumValues(includeDeprecated: true) { name description isDeprecated deprecationReason }
  possibleTypes { ...TypeRef }
}
fragment InputValue on __InputValue { name description type { ...TypeRef } defaultValue }
fragment TypeRef on __Type {
  kind name ofType { kind name ofType { kind name ofType { kind name ofType { kind name } } } }
}`
