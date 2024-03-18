// Copyright 2024 Democratized Data Foundation
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
	"encoding/json"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"

	ds "github.com/ipfs/go-datastore"
)

func (db *db) fetchCollectionPolicyDescription(
	ctx context.Context,
	txn datastore.Txn,
	colID uint32,
) (immutable.Option[client.PolicyDescription], error) {
	collectionPolicyKey := core.NewCollectionPolicyKey(colID)
	policyBuf, err := txn.Systemstore().Get(ctx, collectionPolicyKey.ToDS())

	if err != nil && errors.Is(err, ds.ErrNotFound) {
		return immutable.None[client.PolicyDescription](), nil
	}

	if err != nil {
		return immutable.None[client.PolicyDescription](), err
	}

	var policy client.PolicyDescription
	err = json.Unmarshal(policyBuf, &policy)
	if err != nil {
		return immutable.None[client.PolicyDescription](), err
	}

	return immutable.Some[client.PolicyDescription](policy), nil
}

func (c *collection) loadPolicy(ctx context.Context, txn datastore.Txn) error {
	policyDescription, err := c.db.fetchCollectionPolicyDescription(ctx, txn, c.ID())
	if err != nil {
		return err
	}
	c.def.Description.Policy = policyDescription
	return nil
}
