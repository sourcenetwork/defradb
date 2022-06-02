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

func (db *db) ExecQuery(ctx context.Context, query string) *client.QueryResult {
	res := &client.QueryResult{}
	// check if its Introspection query
	if strings.Contains(query, "IntrospectionQuery") {
		return db.ExecIntrospection(query)
	}

	txn, err := db.NewTxn(ctx, false)
	if err != nil {
		res.Errors = []interface{}{err.Error()}
		return res
	}
	defer txn.Discard(ctx)

	results, err := db.queryExecutor.ExecuteRequest(ctx, db, txn, query)
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

func (db *db) ExecTransactionalQuery(
	ctx context.Context,
	query string,
	txn datastore.Txn,
) *client.QueryResult {
	res := &client.QueryResult{}
	// check if its Introspection query
	if strings.Contains(query, "IntrospectionQuery") {
		return db.ExecIntrospection(query)
	}

	results, err := db.queryExecutor.ExecuteRequest(ctx, db, txn, query)
	if err != nil {
		res.Errors = []interface{}{err.Error()}
		return res
	}

	res.Data = results
	return res
}

func (db *db) ExecIntrospection(query string) *client.QueryResult {
	schema := db.schema.Schema()
	// t := schema.Type("userFilterArg")
	// spew.Dump(t.(*gql.InputObject).Fields())
	params := gql.Params{Schema: *schema, RequestString: query}
	r := gql.Do(params)

	res := &client.QueryResult{
		Data:   r.Data,
		Errors: make([]interface{}, len(r.Errors)),
	}

	for i, err := range r.Errors {
		res.Errors[i] = err
	}

	return res
}
