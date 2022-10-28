// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package graphql

import (
	"context"
	"strings"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/core"
	defrap "github.com/sourcenetwork/defradb/query/graphql/parser"
	"github.com/sourcenetwork/defradb/query/graphql/schema"

	gql "github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	gqlp "github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
)

var _ core.Parser = (*parser)(nil)

type parser struct {
	schemaManager schema.SchemaManager
}

func NewParser() (*parser, error) {
	schemaManager, err := schema.NewSchemaManager()
	if err != nil {
		return nil, err
	}

	p := &parser{
		schemaManager: *schemaManager,
	}

	return p, nil
}

func (p *parser) IsIntrospection(request string) bool {
	// todo: This needs to be done properly https://github.com/sourcenetwork/defradb/issues/911
	return strings.Contains(request, "IntrospectionQuery")
}

func (p *parser) ExecuteIntrospection(request string) *client.QueryResult {
	schema := p.schemaManager.Schema()
	params := gql.Params{Schema: *schema, RequestString: request}
	r := gql.Do(params)

	res := &client.QueryResult{
		Data:   r.Data,
		Errors: make([]any, len(r.Errors)),
	}

	for i, err := range r.Errors {
		res.Errors[i] = err
	}

	return res
}

func (p *parser) Parse(request string) (*request.Request, []error) {
	source := source.NewSource(&source.Source{
		Body: []byte(request),
		Name: "GraphQL request",
	})

	ast, err := gqlp.Parse(gqlp.ParseParams{Source: source})
	if err != nil {
		return nil, []error{err}
	}

	schema := p.schemaManager.Schema()
	validationResult := gql.ValidateDocument(schema, ast, nil)
	if !validationResult.IsValid {
		errors := make([]error, len(validationResult.Errors))
		for i, err := range validationResult.Errors {
			errors[i] = err
		}
		return nil, errors
	}

	query, parsingErrors := defrap.ParseQuery(ast)
	if len(parsingErrors) > 0 {
		return nil, parsingErrors
	}

	return query, nil
}

func (p *parser) AddSchema(ctx context.Context, schema string) error {
	_, _, err := p.schemaManager.Generator.FromSDL(ctx, schema)
	return err
}

func (p *parser) CreateDescriptions(ctx context.Context, schemaString string) ([]client.CollectionDescription, []core.SchemaDefinition, error) {
	schemaManager, err := schema.NewSchemaManager()
	if err != nil {
		return nil, nil, err
	}

	types, astdoc, err := schemaManager.Generator.FromSDL(ctx, schemaString)
	if err != nil {
		return nil, nil, err
	}

	colDesc, err := schemaManager.Generator.CreateDescriptions(types)
	if err != nil {
		return nil, nil, err
	}

	definitions := make([]core.SchemaDefinition, len(astdoc.Definitions))
	for i, astDefinition := range astdoc.Definitions {
		objDef, isObjDef := astDefinition.(*ast.ObjectDefinition)
		if !isObjDef {
			continue
		}

		definitions[i] = core.SchemaDefinition{
			Name: objDef.Name.Value,
			Body: objDef.Loc.Source.Body[objDef.Loc.Start:objDef.Loc.End],
		}
	}

	return colDesc, definitions, nil
}
