// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package db

import (
	"context"
	"strings"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"

	gql "github.com/graphql-go/graphql"
)

const introspectionRequest string = "IntrospectionQuery"

func (db *db) ExecuteRequest(ctx context.Context, request string) *client.RequestResult {
	res := &client.RequestResult{}
	// check if its an Introspection request.
	if strings.Contains(request, introspectionRequest) {
		return db.ExecuteIntrospection(request)
	}

	txn, err := db.NewTxn(ctx, false)
	if err != nil {
		res.Errors = []interface{}{err.Error()}
		return res
	}
	defer txn.Discard(ctx)

	results, err := db.requestExecutor.ExecuteRequest(ctx, db, txn, request)
	if err != nil {
		res.Errors = []interface{}{err.Error()}
		return res
	}

	if err := txn.Commit(ctx); err != nil {
		res.Errors = []interface{}{err.Error()}
		return res
	}

	res.Data = results
	return res
}

func (db *db) ExecuteTransactionalRequest(
	ctx context.Context,
	request string,
	txn datastore.Txn,
) *client.RequestResult {
	res := &client.RequestResult{}
	// check if its Introspection request.
	if strings.Contains(request, introspectionRequest) {
		return db.ExecuteIntrospection(request)
	}

	results, err := db.requestExecutor.ExecuteRequest(ctx, db, txn, request)
	if err != nil {
		res.Errors = []interface{}{err.Error()}
		return res
	}

	res.Data = results
	return res
}

func (db *db) ExecuteIntrospection(request string) *client.RequestResult {
	schema := db.schema.Schema()
	params := gql.Params{Schema: *schema, RequestString: request}
	r := gql.Do(params)

	res := &client.RequestResult{
		Data:   r.Data,
		Errors: make([]interface{}, len(r.Errors)),
	}

	for i, err := range r.Errors {
		res.Errors[i] = err
	}

	return res
}
