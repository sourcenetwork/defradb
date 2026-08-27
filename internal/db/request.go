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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/extensions"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/identity"
	"github.com/sourcenetwork/defradb/internal/planner"
)

func (db *DB) parseRequest(
	ctx context.Context,
	query string,
	options *client.GQLOptions,
) (*request.Request, *client.RequestResult) {
	res := &client.RequestResult{}
	ast, err := db.parser.BuildRequestAST(ctx, query)
	if err != nil {
		res.GQL.Errors = append(res.GQL.Errors, err)
		return nil, res
	}
	if db.parser.IsIntrospection(ast) {
		return nil, db.parser.ExecuteIntrospection(ctx, query)
	}

	parsedRequest, errors := db.parser.Parse(ctx, ast, options)
	if len(errors) > 0 {
		res.GQL.Errors = append(res.GQL.Errors, errors...)
		return nil, res
	}
	return parsedRequest, nil
}

func (db *DB) executeRequest(ctx context.Context, parsedRequest *request.Request) *client.RequestResult {
	res := &client.RequestResult{}

	pub, err := db.handleSubscription(ctx, parsedRequest)
	if err != nil {
		res.GQL.Errors = append(res.GQL.Errors, err)
		return res
	}

	if pub != nil {
		res.Subscription = pub
		return res
	}

	// Warnings raised while running the request are collected here and returned in
	// the `extensions` field of the response.
	ctx = extensions.WithAccumulator(ctx)

	planner := planner.New(
		ctx,
		identity.FromContext(ctx),
		db.nodeACP,
		db.documentACP,
		db,
		db.p2p,
		db.getLensStore(ctx),
		db.collectionRepository,
	)

	results, err := planner.RunRequest(ctx, parsedRequest)
	if err != nil {
		res.GQL.Errors = append(res.GQL.Errors, err)
	}
	res.GQL.Data = results
	res.GQL.Extensions = extensions.Collect(ctx)
	return res
}

func truncateMutationState(req *request.Request) (bool, bool) {
	if len(req.Mutations) == 0 {
		return false, false
	}

	hasTruncate := false
	for _, selection := range req.Mutations[0].Selections {
		mutation, ok := selection.(*request.ObjectMutation)
		if ok && mutation.Type == request.TruncateObjects {
			hasTruncate = true
		}
	}
	return hasTruncate, hasTruncate && len(req.Mutations[0].Selections) == 1
}
