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

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/planner"
)

// execRequest executes a request against the database.
func (db *DB) execRequest(ctx context.Context, request string, options *client.GQLOptions) *client.RequestResult {
	res := &client.RequestResult{}
	ast, err := db.parser.BuildRequestAST(ctx, request)
	if err != nil {
		res.GQL.Errors = append(res.GQL.Errors, err)
		return res
	}
	if db.parser.IsIntrospection(ast) {
		return db.parser.ExecuteIntrospection(ctx, request)
	}

	parsedRequest, errors := db.parser.Parse(ctx, ast, options)
	if len(errors) > 0 {
		res.GQL.Errors = append(res.GQL.Errors, errors...)
		return res
	}

	pub, err := db.handleSubscription(ctx, parsedRequest)
	if err != nil {
		res.GQL.Errors = append(res.GQL.Errors, err)
		return res
	}

	if pub != nil {
		res.Subscription = pub
		return res
	}

	planner := planner.New(ctx, identity.FromContext(ctx), db.documentACP, db)

	results, err := planner.RunRequest(ctx, parsedRequest)
	if err != nil {
		res.GQL.Errors = append(res.GQL.Errors, err)
	}
	res.GQL.Data = results
	return res
}
