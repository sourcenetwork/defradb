// Copyright 2026 Democratized Data Foundation
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

	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/db/action"
	"github.com/sourcenetwork/defradb/internal/utils"
)

func (db *DB) ListActions(
	ctx context.Context,
	opts ...options.Enumerable[options.ListActionsOptions],
) ([]client.ActionExecution, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	opt := utils.NewOptions(opts...)
	ident := opt.GetIdentity()

	if err := db.checkNodeAccess(ctx, ident, acpTypes.NodeListActionPerm); err != nil {
		return nil, err
	}

	ctx, txn, err := ensureContextTxn(ctx, db, true)
	if err != nil {
		return nil, err
	}

	defer txn.Discard()

	actionInfos, err := db.listActions(ctx)
	if err != nil {
		return nil, err
	}

	return actionInfos, nil
}

func (db *DB) listActions(
	ctx context.Context,
) ([]client.ActionExecution, error) {
	return action.ListExecutions(ctx)
}
