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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/request/graphql/parser"
	"github.com/sourcenetwork/defradb/request/graphql/schema"

	gql "github.com/graphql-go/graphql"
	gqlp "github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
)

// RequestPlan is an external hook into the planNode
// system. It allows outside packages to
// execute and manage a request plan graph directly.
// Instead of using one of the available functions
// like ExecRequest(...).
// Currently, this is used by the collection.Update
// system.
type RequestPlan planNode

type RequestExecutor struct {
	SchemaManager *schema.SchemaManager
}

func NewRequestExecutor(manager *schema.SchemaManager) (*RequestExecutor, error) {
	if manager == nil {
		return nil, fmt.Errorf("SchemaManager cannot be nil.")
	}

	return &RequestExecutor{
		SchemaManager: manager,
	}, nil
}

func (e *RequestExecutor) MakeSelectRequest(
	ctx context.Context,
	db client.DB,
	txn datastore.Txn,
	selectStmt *parser.Select,
) (RequestPlan, error) {
	if selectStmt == nil {
		return nil, fmt.Errorf("Cannot create a request plan without a selection.")
	}
	planner := makePlanner(ctx, db, txn)
	return planner.makePlan(selectStmt)
}

func (e *RequestExecutor) ExecuteRequest(
	ctx context.Context,
	db client.DB,
	txn datastore.Txn,
	request string,
	args ...interface{},
) ([]map[string]interface{}, error) {
	q, err := e.ParseRequestString(request)
	if err != nil {
		return nil, err
	}

	planner := makePlanner(ctx, db, txn)
	return planner.runRequest(ctx, q)

}

func (e *RequestExecutor) MakePlanFromParser(
	ctx context.Context,
	db client.DB,
	txn datastore.Txn,
	request *parser.Request,
) (planNode, error) {
	planner := makePlanner(ctx, db, txn)
	return planner.makePlan(request)
}

func (e *RequestExecutor) ParseRequestString(request string) (*parser.Request, error) {
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
		return nil, fmt.Errorf("%v", validationResult.Errors)
	}

	return parser.ParseRequest(ast)
}
