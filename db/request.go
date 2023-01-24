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

// ExecRequest executes a request against the database.
func (db *db) ExecRequest(ctx context.Context, query string) *client.QueryResult {
	res := &client.QueryResult{}
	// check if its Introspection query
	if strings.Contains(query, "IntrospectionQuery") {
		return db.ExecIntrospection(query)
	}

	txn, err := db.NewTxn(ctx, false)
	if err != nil {
		res.GQL.Errors = []any{err.Error()}
		return res
	}
	defer txn.Discard(ctx)

	request, errors := db.parser.Parse(query)
	if len(errors) > 0 {
		errorStrings := make([]any, len(errors))
		for i, err := range errors {
			errorStrings[i] = err.Error()
		}
		res.GQL.Errors = errorStrings
		return res
	}

	pub, subRequest, err := db.checkForClientSubsciptions(request)
	if err != nil {
		res.GQL.Errors = []any{err.Error()}
		return res
	}

	if pub != nil {
		res.Pub = pub
		go db.handleSubscription(ctx, pub, subRequest)
		return res
	}

	planner := planner.New(ctx, db, txn)

	results, err := planner.RunRequest(ctx, request)
	if err != nil {
		res.GQL.Errors = []any{err.Error()}
		return res
	}

	if err := txn.Commit(ctx); err != nil {
		res.GQL.Errors = []any{err.Error()}
		return res
	}

	res.GQL.Data = results
	return res
}

// ExecTransactionalRequest executes a transaction request against the database.
func (db *db) ExecTransactionalRequest(
	ctx context.Context,
	query string,
	txn datastore.Txn,
) *client.QueryResult {
	if db.parser.IsIntrospection(query) {
		return db.ExecIntrospection(query)
	}

	res := &client.QueryResult{}

	request, errors := db.parser.Parse(query)
	if len(errors) > 0 {
		errorStrings := make([]any, len(errors))
		for i, err := range errors {
			errorStrings[i] = err.Error()
		}
		res.GQL.Errors = errorStrings
		return res
	}

	planner := planner.New(ctx, db, txn)
	results, err := planner.RunRequest(ctx, request)
	if err != nil {
		res.GQL.Errors = []any{err.Error()}
		return res
	}

	res.GQL.Data = results
	return res
}

// ExecIntrospection executes an introspection query against the database.
func (db *db) ExecIntrospection(query string) *client.QueryResult {
	return db.parser.ExecuteIntrospection(query)
}
