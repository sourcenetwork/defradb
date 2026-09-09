// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.

package graphql

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/ast"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/astnormalization"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/astparser"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/astprinter"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/asttransform"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/astvalidation"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/operationreport"

	"github.com/sourcenetwork/defradb/client"
)

// These tests cover the graphql-go-tools parsing, operation selection,
// normalization, and validation pipeline used by DefraDB.

const experimentSchema = `
	directive @explain(type: String) on QUERY

	type Query {
		User(filter: UserFilterArg): [User]
	}

	type User {
		name: String
		age: Int
	}

	input UserFilterArg {
		age: IntOperatorBlock
	}

	input IntOperatorBlock {
		_eq: Int
	}
`

const experimentOperation = `
	query Other {
		User { name }
	}

	query Selected($age: Int!) @explain(type: "simple") {
		first: User(filter: {age: {_eq: $age}}) {
			...SelectedUserFields
		}
		User {
			name
		}
	}

	fragment SelectedUserFields on User {
		age
		name
	}
`

func TestGraphQLGoToolsProductionBridge(t *testing.T) {
	parser, err := NewParser(false)
	require.NoError(t, err)

	document, err := parser.BuildRequestAST(context.Background(), `query Schema { __schema { queryType { name } } }`)
	require.NoError(t, err)
	require.Equal(t, "graphql", document.Language())
	require.True(t, parser.IsIntrospection(document))

	document, err = parser.BuildRequestAST(context.Background(), `{ User { name } }`)
	require.NoError(t, err)
	require.False(t, parser.IsIntrospection(document))

	document, err = parser.BuildRequestAST(context.Background(), `
		query Schema { ...SchemaFields }
		fragment SchemaFields on Query { __schema { queryType { name } } }
	`)
	require.NoError(t, err)
	require.True(t, parser.IsIntrospection(document))

	_, err = parser.BuildRequestAST(context.Background(), `{`)
	require.Error(t, err)
}

func TestGraphQLGoToolsExperiment_NormalizesDefraOperation(t *testing.T) {
	definition := parseExperimentDocument(t, experimentSchema)
	require.NoError(t, asttransform.MergeDefinitionWithBaseSchema(&definition))

	definitionReport := operationreport.Report{}
	state := astvalidation.DefaultDefinitionValidator().Validate(&definition, &definitionReport)
	require.Equal(t, astvalidation.Valid, state, definitionReport.Error())

	operation := parseExperimentDocument(t, experimentOperation)
	operation.Input.Variables = []byte(`{"age": 42}`)
	operationReport := operationreport.Report{}

	// This replaces BuildExecutionContext, CollectFields, the custom fragment
	// traversal in orderCollectedFields, and variable cleanup with one call.
	normalizer := astnormalization.NewWithOpts(
		astnormalization.WithRemoveNotMatchingOperationDefinitions(),
		astnormalization.WithRemoveFragmentDefinitions(),
		astnormalization.WithInlineFragmentSpreads(),
		astnormalization.WithRemoveUnusedVariables(),
	)
	normalizer.NormalizeNamedOperation(
		&operation,
		&definition,
		[]byte("Selected"),
		&operationReport,
	)
	require.False(t, operationReport.HasErrors(), operationReport.Error())
	state = astvalidation.DefaultOperationValidator().Validate(&operation, &definition, &operationReport)
	require.Equal(t, astvalidation.Valid, state, operationReport.Error())

	normalized, err := astprinter.PrintStringIndent(&operation, "  ")
	require.NoError(t, err)
	require.Equal(t, `query Selected($age: Int!)@explain(type: "simple") {
  first: User(filter: {age: {_eq: $age}}){
    age
    name
  }
  User {
    name
  }
}`, normalized)
	require.JSONEq(t, `{"age": 42}`, string(operation.Input.Variables))
}

func TestGraphQLGoToolsExperiment_RejectsUnknownField(t *testing.T) {
	definition := parseExperimentDocument(t, experimentSchema)
	require.NoError(t, asttransform.MergeDefinitionWithBaseSchema(&definition))
	operation := parseExperimentDocument(t, `{ User { missing } }`)

	report := operationreport.Report{}
	state := astvalidation.DefaultOperationValidator().Validate(&operation, &definition, &report)
	require.Equal(t, astvalidation.Invalid, state)
	require.Contains(t, report.Error(), `Cannot query field "missing" on type "User"`)
}

func TestGraphQLGoToolsExperiment_ValidationErrorsAreStructured(t *testing.T) {
	definition := parseExperimentDocument(t, experimentSchema)
	require.NoError(t, asttransform.MergeDefinitionWithBaseSchema(&definition))
	tests := []struct {
		name      string
		operation string
		contains  string
	}{
		{name: "unknown field", operation: `{ User { missing } }`, contains: `Cannot query field "missing"`},
		{name: "unknown argument", operation: `{ User(unknown: 1) { name } }`, contains: `Unknown argument "unknown"`},
		{name: "missing selection", operation: `{ User }`, contains: `must have a selection of subfields`},
		{name: "unused variable", operation: `query($age: Int!) { User { name } }`, contains: `variable: age defined on operation:  but never used`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			operation := parseExperimentDocument(t, test.operation)
			report := operationreport.Report{}
			state := astvalidation.DefaultOperationValidator().Validate(&operation, &definition, &report)
			require.Equal(t, astvalidation.Invalid, state)
			require.NotEmpty(t, report.ExternalErrors)
			require.Contains(t, report.Error(), test.contains)
		})
	}
}

func TestGraphQLGoToolsExperiment_ValidatesVariables(t *testing.T) {
	definition := parseExperimentDocument(t, experimentSchema)
	require.NoError(t, asttransform.MergeDefinitionWithBaseSchema(&definition))

	operation := `query Selected($age: Int!) { User(filter: {age: {_eq: $age}}) { name } }`
	tests := []struct {
		name      string
		variables map[string]any
		wantError bool
	}{
		{name: "valid", variables: map[string]any{"age": 42}},
		{name: "missing required", variables: map[string]any{}, wantError: true},
		{name: "wrong scalar type", variables: map[string]any{"age": "old"}, wantError: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			selected, err := selectAndNormalizeOperation(operation, "", test.variables, &definition)
			require.NoError(t, err)
			err = validateVariables(&selected, &client.GQLOptions{Variables: test.variables}, &definition)
			if test.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestGraphQLGoToolsExperiment_ValidatesSelectedOperationVariables(t *testing.T) {
	definition := parseExperimentDocument(t, experimentSchema)
	require.NoError(t, asttransform.MergeDefinitionWithBaseSchema(&definition))

	operation := `
		query WithoutVariables { User { name } }
		query WithVariables($age: Int!) { User(filter: {age: {_eq: $age}}) { name } }
	`
	withoutVariables, err := selectAndNormalizeOperation(operation, "WithoutVariables", nil, &definition)
	require.NoError(t, err)
	require.NoError(t, validateVariables(&withoutVariables, &client.GQLOptions{}, &definition))

	withVariables, err := selectAndNormalizeOperation(operation, "WithVariables", nil, &definition)
	require.NoError(t, err)
	require.Error(t, validateVariables(&withVariables, &client.GQLOptions{}, &definition))
}

func TestGraphQLGoToolsExperiment_SelectsOperation(t *testing.T) {
	definition := parseExperimentDocument(t, experimentSchema)
	require.NoError(t, asttransform.MergeDefinitionWithBaseSchema(&definition))
	operation := `query First { User { name } } query Second { User { age } }`

	_, err := selectAndNormalizeOperation(operation, "", nil, &definition)
	require.ErrorContains(t, err, "Must provide operation name")
	_, err = selectAndNormalizeOperation(operation, "Missing", nil, &definition)
	require.ErrorContains(t, err, "Unknown operation")

	selected, err := selectAndNormalizeOperation(operation, "Second", nil, &definition)
	require.NoError(t, err)
	printed, err := astprinter.PrintString(&selected)
	require.NoError(t, err)
	require.Contains(t, printed, "query Second")
	require.NotContains(t, printed, "query First")
	require.NotContains(t, printed, "fragment ")
}

func TestGraphQLGoToolsExperiment_ResolvesSelectionDirectives(t *testing.T) {
	definition := parseExperimentDocument(t, experimentSchema)
	require.NoError(t, asttransform.MergeDefinitionWithBaseSchema(&definition))
	operation := `query Selected($show: Boolean!, $hide: Boolean!) {
		shown: User @include(if: $show) { name }
		hidden: User @skip(if: $hide) { name }
	}`
	selected, err := selectAndNormalizeOperation(operation, "Selected", map[string]any{
		"show": true,
		"hide": true,
	}, &definition)
	require.NoError(t, err)

	printed, err := astprinter.PrintString(&selected)
	require.NoError(t, err)
	require.Contains(t, printed, "shown: User")
	require.NotContains(t, printed, "hidden: User")
	require.NotContains(t, printed, "@include")
}

func BenchmarkGraphQLParserExperiment(b *testing.B) {
	request := []byte(experimentOperation)
	b.ReportAllocs()
	for range b.N {
		_, report := astparser.ParseGraphqlDocumentBytes(request)
		if report.HasErrors() {
			b.Fatal(report.Error())
		}
	}
}

func BenchmarkGraphQLGeneratedSchemaParserExperiment(b *testing.B) {
	schema, err := os.ReadFile("schema/testfixtures/schema.simple.gen.graphql")
	require.NoError(b, err)

	b.ReportAllocs()
	for range b.N {
		_, report := astparser.ParseGraphqlDocumentBytes(schema)
		if report.HasErrors() {
			b.Fatal(report.Error())
		}
	}
}

func parseExperimentDocument(t *testing.T, input string) ast.Document {
	t.Helper()
	document, report := astparser.ParseGraphqlDocumentString(input)
	require.False(t, report.HasErrors(), report.Error())
	return document
}
