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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/mapper"
)

// Query is an external hook into the planNode
// system. It allows outside packages to
// execute and manage a query plan graph directly.
// Instead of using one of the available functions
// like ExecQuery(...).
// Currently, this is used by the collection.Update
// system.
type Query planNode

func MakeSelectQuery(
	ctx context.Context,
	db client.DB,
	txn datastore.Txn,
	selectStmt *mapper.Select,
) (Query, error) {
	if selectStmt == nil {
		return nil, errors.New("cannot create query without a selection")
	}
	planner := makePlanner(ctx, db, txn)
	return planner.makePlan(selectStmt)
}

func ExecQuery(
	ctx context.Context,
	db client.DB,
	txn datastore.Txn,
	query *request.Request,
) ([]map[string]any, error) {
	planner := makePlanner(ctx, db, txn)
	return planner.runRequest(ctx, query)
}

func MakePlanFromParser(
	ctx context.Context,
	db client.DB,
	txn datastore.Txn,
	query *request.Request,
) (planNode, error) {
	planner := makePlanner(ctx, db, txn)
	return planner.makePlan(query)
}
