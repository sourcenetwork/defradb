// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package planner

import (
	"context"
	"fmt"

	gql "github.com/graphql-go/graphql"
	gqlp "github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/query/graphql/mapper"
	"github.com/sourcenetwork/defradb/query/graphql/parser"
	"github.com/sourcenetwork/defradb/query/graphql/schema"
)

// Query is an external hook into the planNode
// system. It allows outside packages to
// execute and manage a query plan graph directly.
// Instead of using one of the available functions
// like ExecQuery(...).
// Currently, this is used by the collection.Update
// system.
type Query planNode

type QueryExecutor struct {
	SchemaManager *schema.SchemaManager
}

func NewQueryExecutor(manager *schema.SchemaManager) (*QueryExecutor, error) {
	if manager == nil {
		return nil, errors.New("SchemaManager cannot be nil")
	}

	return &QueryExecutor{
		SchemaManager: manager,
	}, nil
}

func (e *QueryExecutor) MakeSelectQuery(
	ctx context.Context,
	db client.DB,
	txn datastore.Txn,
	selectStmt *mapper.Select,
) (Query, error) {
	if selectStmt == nil {
		return nil, errors.New("Cannot create query without a selection")
	}
	planner := makePlanner(ctx, db, txn)
	return planner.makePlan(selectStmt)
}

func (e *QueryExecutor) ExecQuery(
	ctx context.Context,
	db client.DB,
	txn datastore.Txn,
	query string,
	args ...any,
) ([]map[string]any, error) {
	q, err := e.ParseRequestString(query)
	if err != nil {
		return nil, err
	}

	planner := makePlanner(ctx, db, txn)
	return planner.runRequest(ctx, q)
}

func (e *QueryExecutor) MakePlanFromParser(
	ctx context.Context,
	db client.DB,
	txn datastore.Txn,
	query *parser.Query,
) (planNode, error) {
	planner := makePlanner(ctx, db, txn)
	return planner.makePlan(query)
}

func (e *QueryExecutor) ParseRequestString(request string) (*parser.Query, error) {
	source := source.NewSource(&source.Source{
		Body: []byte(request),
		Name: "GraphQL request",
	})

	ast, err := gqlp.Parse(gqlp.ParseParams{Source: source})
	if err != nil {
		return nil, err
	}

	schema := e.SchemaManager.Schema()
	validationResult := gql.ValidateDocument(schema, ast, nil)
	if !validationResult.IsValid {
		return nil, errors.New(fmt.Sprintf("%v", validationResult.Errors))
	}

	return parser.ParseQuery(ast)
}
