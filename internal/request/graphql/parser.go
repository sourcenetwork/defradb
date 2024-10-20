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

	gql "github.com/sourcenetwork/graphql-go"
	"github.com/sourcenetwork/graphql-go/language/ast"
	gqlp "github.com/sourcenetwork/graphql-go/language/parser"
	"github.com/sourcenetwork/graphql-go/language/source"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/internal/core"
	defrap "github.com/sourcenetwork/defradb/internal/request/graphql/parser"
	"github.com/sourcenetwork/defradb/internal/request/graphql/schema"
)

var _ core.Parser = (*parser)(nil)

type parser struct {
	schemaManager *schema.SchemaManager
}

func NewParser() (*parser, error) {
	schemaManager, err := schema.NewSchemaManager()
	if err != nil {
		return nil, err
	}

	p := &parser{
		schemaManager: schemaManager,
	}

	return p, nil
}

func (p *parser) BuildRequestAST(request string) (*ast.Document, error) {
	source := source.NewSource(&source.Source{
		Body: []byte(request),
		Name: "GraphQL request",
	})

	ast, err := gqlp.Parse(gqlp.ParseParams{Source: source})
	if err != nil {
		return nil, err
	}

	return ast, nil
}

func (p *parser) IsIntrospection(ast *ast.Document) bool {
	schema := p.schemaManager.Schema()
	return defrap.IsIntrospectionQuery(*schema, ast)
}

func (p *parser) ExecuteIntrospection(request string) *client.RequestResult {
	schema := p.schemaManager.Schema()
	params := gql.Params{Schema: *schema, RequestString: request}
	r := gql.Do(params)

	res := &client.RequestResult{
		GQL: client.GQLResult{
			Data: r.Data,
		},
	}

	for _, err := range r.Errors {
		res.GQL.Errors = append(res.GQL.Errors, err)
	}

	return res
}

func (p *parser) Parse(ast *ast.Document, options *client.GQLOptions) (*request.Request, []error) {
	schema := p.schemaManager.Schema()
	validationResult := gql.ValidateDocument(schema, ast, nil)
	if !validationResult.IsValid {
		errors := make([]error, len(validationResult.Errors))
		for i, err := range validationResult.Errors {
			errors[i] = err
		}
		return nil, errors
	}

	return defrap.ParseRequest(*schema, ast, options)
}

func (p *parser) ParseSDL(sdl string) ([]client.CollectionDefinition, error) {
	return p.schemaManager.ParseSDL(sdl)
}

func (p *parser) SetSchema(ctx context.Context, txn datastore.Txn, collections []client.CollectionDefinition) error {
	schemaManager, err := schema.NewSchemaManager()
	if err != nil {
		return err
	}

	_, err = schemaManager.Generator.Generate(ctx, collections)
	if err != nil {
		return err
	}

	txn.OnSuccess(
		func() {
			p.schemaManager = schemaManager
		},
	)
	return err
}

func (p *parser) NewFilterFromString(collectionType string, body string) (immutable.Option[request.Filter], error) {
	return defrap.NewFilterFromString(*p.schemaManager.Schema(), collectionType, body)
}
