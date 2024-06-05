// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package lens

import (
	"context"

	"github.com/lens-vm/lens/host-go/config/model"
	"github.com/sourcenetwork/immutable/enumerable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
)

type implicitTxnLensRegistry struct {
	registry *lensRegistry
	db       client.TxnSource
}

type explicitTxnLensRegistry struct {
	registry *lensRegistry
	txn      datastore.Txn
}

var _ client.LensRegistry = (*implicitTxnLensRegistry)(nil)
var _ client.LensRegistry = (*explicitTxnLensRegistry)(nil)

func (r *implicitTxnLensRegistry) Init(txnSource client.TxnSource) {
	r.db = txnSource
}

func (r *explicitTxnLensRegistry) Init(txnSource client.TxnSource) {}

func (r *explicitTxnLensRegistry) WithTxn(txn datastore.Txn) client.LensRegistry {
	return &explicitTxnLensRegistry{
		registry: r.registry,
		txn:      txn,
	}
}

func (r *implicitTxnLensRegistry) SetMigration(ctx context.Context, collectionID uint32, cfg model.Lens) error {
	txn, err := r.db.NewTxn(ctx, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)
	txnCtx := r.registry.getCtx(txn, false)

	err = r.registry.setMigration(ctx, txnCtx, collectionID, cfg)
	if err != nil {
		return err
	}

	return txn.Commit(ctx)
}

func (r *explicitTxnLensRegistry) SetMigration(ctx context.Context, collectionID uint32, cfg model.Lens) error {
	return r.registry.setMigration(ctx, r.registry.getCtx(r.txn, false), collectionID, cfg)
}

func (r *implicitTxnLensRegistry) ReloadLenses(ctx context.Context) error {
	txn, err := r.db.NewTxn(ctx, true)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)
	txnCtx := r.registry.getCtx(txn, false)

	err = r.registry.reloadLenses(ctx, txnCtx)
	if err != nil {
		return err
	}

	return txn.Commit(ctx)
}

func (r *explicitTxnLensRegistry) ReloadLenses(ctx context.Context) error {
	return r.registry.reloadLenses(ctx, r.registry.getCtx(r.txn, true))
}

func (r *implicitTxnLensRegistry) MigrateUp(
	ctx context.Context,
	src enumerable.Enumerable[LensDoc],
	collectionID uint32,
) (enumerable.Enumerable[map[string]any], error) {
	txn, err := r.db.NewTxn(ctx, true)
	if err != nil {
		return nil, err
	}
	defer txn.Discard(ctx)
	txnCtx := newTxnCtx(txn)

	return r.registry.migrateUp(txnCtx, src, collectionID)
}

func (r *explicitTxnLensRegistry) MigrateUp(
	ctx context.Context,
	src enumerable.Enumerable[LensDoc],
	collectionID uint32,
) (enumerable.Enumerable[map[string]any], error) {
	return r.registry.migrateUp(r.registry.getCtx(r.txn, true), src, collectionID)
}

func (r *implicitTxnLensRegistry) MigrateDown(
	ctx context.Context,
	src enumerable.Enumerable[LensDoc],
	collectionID uint32,
) (enumerable.Enumerable[map[string]any], error) {
	txn, err := r.db.NewTxn(ctx, true)
	if err != nil {
		return nil, err
	}
	defer txn.Discard(ctx)
	txnCtx := newTxnCtx(txn)

	return r.registry.migrateDown(txnCtx, src, collectionID)
}

func (r *explicitTxnLensRegistry) MigrateDown(
	ctx context.Context,
	src enumerable.Enumerable[LensDoc],
	collectionID uint32,
) (enumerable.Enumerable[map[string]any], error) {
	return r.registry.migrateDown(r.registry.getCtx(r.txn, true), src, collectionID)
}
