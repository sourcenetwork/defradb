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
	"github.com/sourcenetwork/defradb/planner"
)

func (db *db) ExecQuery(ctx context.Context, query string) *client.QueryResult {
	res := &client.QueryResult{}
	// check if its Introspection query
	if strings.Contains(query, "IntrospectionQuery") {
		return db.ExecIntrospection(query)
	}

	txn, err := db.NewTxn(ctx, false)
	if err != nil {
		res.Errors = []any{err.Error()}
		return res
	}
	defer txn.Discard(ctx)

	request, errors := db.parser.Parse(query)
	if len(errors) > 0 {
		errorStrings := make([]any, len(errors))
		for i, err := range errors {
			errorStrings[i] = err.Error()
		}
		res.Errors = errorStrings
		return res
	}

	results, err := planner.ExecQuery(ctx, db, txn, request)
	if err != nil {
		res.Errors = []any{err.Error()}
		return res
	}

	if err := txn.Commit(ctx); err != nil {
		res.Errors = []any{err.Error()}
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
	if db.parser.IsIntrospectionRequest(query) {
		return db.ExecIntrospection(query)
	}

	res := &client.QueryResult{}

	request, errors := db.parser.Parse(query)
	if len(errors) > 0 {
		errorStrings := make([]any, len(errors))
		for i, err := range errors {
			errorStrings[i] = err.Error()
		}
		res.Errors = errorStrings
		return res
	}

	results, err := planner.ExecQuery(ctx, db, txn, request)
	if err != nil {
		res.Errors = []any{err.Error()}
		return res
	}

	res.Data = results
	return res
}

func (db *db) ExecIntrospection(query string) *client.QueryResult {
	return db.parser.ExecuteIntrospectionRequest(query)
}
