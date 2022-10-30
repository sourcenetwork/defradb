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
	gqlp "github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
)

var _ core.Parser = (*Parser)(nil)

type Parser struct {
	SchemaManager schema.SchemaManager
}

func NewParser() (*Parser, error) {
	schemaManager, err := schema.NewSchemaManager()
	if err != nil {
		return nil, err
	}

	p := &Parser{
		SchemaManager: *schemaManager,
	}

	return p, nil
}

func (p *Parser) IsIntrospection(request string) bool {
	// todo: This needs to be done properly https://github.com/sourcenetwork/defradb/issues/911
	return strings.Contains(request, "IntrospectionQuery")
}

func (p *Parser) ExecuteIntrospection(request string) *client.QueryResult {
	schema := p.SchemaManager.Schema()
	params := gql.Params{Schema: *schema, RequestString: request}
	r := gql.Do(params)

	res := &client.QueryResult{
		GQL: client.GQLResult{
			Data:   r.Data,
			Errors: make([]any, len(r.Errors)),
		},
	}

	for i, err := range r.Errors {
		res.GQL.Errors[i] = err
	}

	return res
}

func (p *Parser) Parse(request string) (*request.Request, []error) {
	source := source.NewSource(&source.Source{
		Body: []byte(request),
		Name: "GraphQL request",
	})

	ast, err := gqlp.Parse(gqlp.ParseParams{Source: source})
	if err != nil {
		return nil, []error{err}
	}

	schema := p.SchemaManager.Schema()
	validationResult := gql.ValidateDocument(schema, ast, nil)
	if !validationResult.IsValid {
		errors := make([]error, len(validationResult.Errors))
		for i, err := range validationResult.Errors {
			errors[i] = err
		}
		return nil, errors
	}

	query, parsingErrors := defrap.ParseQuery(*schema, ast)
	if len(parsingErrors) > 0 {
		return nil, parsingErrors
	}

	return query, nil
}

func (p *Parser) AddSchema(ctx context.Context, schema string) error {
	_, _, err := p.SchemaManager.Generator.FromSDL(ctx, schema)
	return err
}
