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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/planner"
)

// execRequest executes a request against the database.
func (db *db) execRequest(
	ctx context.Context,
	identity immutable.Option[string],
	request string,
	txn datastore.Txn,
) *client.RequestResult {
	res := &client.RequestResult{}
	ast, err := db.parser.BuildRequestAST(request)
	if err != nil {
		res.GQL.Errors = []error{err}
		return res
	}
	if db.parser.IsIntrospection(ast) {
		return db.parser.ExecuteIntrospection(request)
	}

	parsedRequest, errors := db.parser.Parse(ast)
	if len(errors) > 0 {
		res.GQL.Errors = errors
		return res
	}

	pub, subRequest, err := db.checkForClientSubscriptions(parsedRequest)
	if err != nil {
		res.GQL.Errors = []error{err}
		return res
	}

	if pub != nil {
		res.Pub = pub
		go db.handleSubscription(ctx, identity, pub, subRequest)
		return res
	}

	planner := planner.New(
		ctx,
		identity,
		db.acp,
		db.WithTxn(txn),
		txn,
	)

	results, err := planner.RunRequest(ctx, parsedRequest)
	if err != nil {
		res.GQL.Errors = []error{err}
		return res
	}

	res.GQL.Data = results
	return res
}

// ExecIntrospection executes an introspection request against the database.
func (db *db) ExecIntrospection(request string) *client.RequestResult {
	return db.parser.ExecuteIntrospection(request)
}
