// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.

package graphql

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	wgast "github.com/wundergraph/graphql-go-tools/v2/pkg/ast"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/astnormalization"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/astparser"
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

func TestExecuteIntrospectionPreservesIntrospectionFieldDescriptions(t *testing.T) {
	parser, err := NewParser(false)
	require.NoError(t, err)
	result := parser.ExecuteIntrospection(context.Background(), `{
		__type(name: "__Schema") { fields { name description } }
	}`)
	require.Empty(t, result.GQL.Errors)
	typeData := result.GQL.Data.(map[string]any)["__type"].(map[string]any)
	fields := typeData["fields"].([]any)
	descriptions := make(map[string]any, len(fields))
	for _, field := range fields {
		item := field.(map[string]any)
		descriptions[item["name"].(string)] = item["description"]
	}
	require.Equal(t, map[string]any{
		"directives":       "A list of all directives supported by this server.",
		"mutationType":     "If this server supports mutation, the type that mutation operations will be rooted at.",
		"queryType":        "The type that query operations will be rooted at.",
		"subscriptionType": "If this server supports subscription, the type that subscription operations will be rooted at.",
		"types":            "A list of all types supported by this server.",
	}, descriptions)
}

func TestExecuteIntrospectionDecodesAllDescriptions(t *testing.T) {
	definition, report := astparser.ParseGraphqlDocumentString(`
		schema { query: Query }
		type Query {
			"A description with a newline.\nAnd a quoted \"value\"."
			described: String
		}
	`)
	require.False(t, report.HasErrors(), report.Error())
	astnormalization.NormalizeDefinition(&definition, &report)
	require.False(t, report.HasErrors(), report.Error())

	result := executeIntrospection(&definition, `{
		__type(name: "Query") { fields { name description } }
	}`)
	require.Empty(t, result.GQL.Errors)
	require.Equal(t, map[string]any{
		"__type": map[string]any{
			"fields": []any{map[string]any{
				"name":        "described",
				"description": "A description with a newline.\nAnd a quoted \"value\".",
			}},
		},
	}, result.GQL.Data)
}

func TestExecuteIntrospectionFiltersDeprecatedMembers(t *testing.T) {
	definition := introspectionTestDefinition(t, `
		type Query {
			active: String
			old: String @deprecated(reason: "replaced")
		}
		enum State { ACTIVE OLD @deprecated(reason: "replaced") }
	`)

	result := executeIntrospection(&definition, `{
		defaultFields: __type(name: "Query") { fields { name } }
		allFields: __type(name: "Query") { fields(includeDeprecated: true) { name } }
		defaultEnum: __type(name: "State") { enumValues { name } }
		allEnum: __type(name: "State") { enumValues(includeDeprecated: true) { name } }
	}`)
	require.Empty(t, result.GQL.Errors)
	require.Equal(t, map[string]any{
		"defaultFields": map[string]any{"fields": []any{map[string]any{"name": "active"}}},
		"allFields": map[string]any{"fields": []any{
			map[string]any{"name": "active"}, map[string]any{"name": "old"},
		}},
		"defaultEnum": map[string]any{"enumValues": []any{map[string]any{"name": "ACTIVE"}}},
		"allEnum": map[string]any{"enumValues": []any{
			map[string]any{"name": "ACTIVE"}, map[string]any{"name": "OLD"},
		}},
	}, result.GQL.Data)
}

func TestExecuteIntrospectionReturnsNullForInapplicableTypeFields(t *testing.T) {
	definition := introspectionTestDefinition(t, `type Query { active: String }`)
	result := executeIntrospection(&definition, `{
		__type(name: "Query") {
			description
			inputFields { name }
			enumValues { name }
			possibleTypes { name }
			ofType { name }
		}
	}`)
	require.Empty(t, result.GQL.Errors)
	require.Equal(t, map[string]any{
		"__type": map[string]any{
			"description":   nil,
			"inputFields":   nil,
			"enumValues":    nil,
			"possibleTypes": nil,
			"ofType":        nil,
		},
	}, result.GQL.Data)
}

func TestExecuteIntrospectionRejectsAmbiguousOperation(t *testing.T) {
	definition := introspectionTestDefinition(t, `type Query { active: String }`)
	result := executeIntrospection(&definition, `
		query First { __type(name: "Query") { name } }
		query Second { __type(name: "Query") { name } }
	`)
	require.Len(t, result.GQL.Errors, 1)
	require.EqualError(t, result.GQL.Errors[0],
		"Must provide operation name if query contains multiple operations.")
}

func introspectionTestDefinition(t *testing.T, sdl string) wgast.Document {
	t.Helper()
	definition, report := astparser.ParseGraphqlDocumentString("schema { query: Query }\n" + sdl)
	require.False(t, report.HasErrors(), report.Error())
	astnormalization.NormalizeDefinition(&definition, &report)
	require.False(t, report.HasErrors(), report.Error())
	return definition
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
