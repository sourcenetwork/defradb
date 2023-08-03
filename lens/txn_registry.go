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

	"github.com/sourcenetwork/immutable/enumerable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
)

type implicitTxnLensRegistry struct {
	registry *lensRegistry
	db       TxnSource
}

type explicitTxnLensRegistry struct {
	registry *lensRegistry
	txn      datastore.Txn
}

var _ client.LensRegistry = (*implicitTxnLensRegistry)(nil)
var _ client.LensRegistry = (*explicitTxnLensRegistry)(nil)

func (r *implicitTxnLensRegistry) WithTxn(txn datastore.Txn) client.LensRegistry {
	return &explicitTxnLensRegistry{
		registry: r.registry,
		txn:      txn,
	}
}

func (r *explicitTxnLensRegistry) WithTxn(txn datastore.Txn) client.LensRegistry {
	return &explicitTxnLensRegistry{
		registry: r.registry,
		txn:      txn,
	}
}

func (r *implicitTxnLensRegistry) SetMigration(ctx context.Context, cfg client.LensConfig) error {
	txn, err := r.db.NewTxn(ctx, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)
	txnCtx := r.registry.getCtx(txn, false)

	err = r.registry.setMigration(ctx, txnCtx, cfg)
	if err != nil {
		return err
	}

	return txn.Commit(ctx)
}

func (r *explicitTxnLensRegistry) SetMigration(ctx context.Context, cfg client.LensConfig) error {
	return r.registry.setMigration(ctx, r.registry.getCtx(r.txn, false), cfg)
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
	schemaVersionID string,
) (enumerable.Enumerable[map[string]any], error) {
	txn, err := r.db.NewTxn(ctx, true)
	if err != nil {
		return nil, err
	}
	defer txn.Discard(ctx)
	txnCtx := newTxnCtx(txn)

	return r.registry.migrateUp(txnCtx, src, schemaVersionID)
}

func (r *explicitTxnLensRegistry) MigrateUp(
	ctx context.Context,
	src enumerable.Enumerable[LensDoc],
	schemaVersionID string,
) (enumerable.Enumerable[map[string]any], error) {
	return r.registry.migrateUp(r.registry.getCtx(r.txn, true), src, schemaVersionID)
}

func (r *implicitTxnLensRegistry) MigrateDown(
	ctx context.Context,
	src enumerable.Enumerable[LensDoc],
	schemaVersionID string,
) (enumerable.Enumerable[map[string]any], error) {
	txn, err := r.db.NewTxn(ctx, true)
	if err != nil {
		return nil, err
	}
	defer txn.Discard(ctx)
	txnCtx := newTxnCtx(txn)

	return r.registry.migrateDown(txnCtx, src, schemaVersionID)
}

func (r *explicitTxnLensRegistry) MigrateDown(
	ctx context.Context,
	src enumerable.Enumerable[LensDoc],
	schemaVersionID string,
) (enumerable.Enumerable[map[string]any], error) {
	return r.registry.migrateDown(r.registry.getCtx(r.txn, true), src, schemaVersionID)
}

func (r *implicitTxnLensRegistry) Config(ctx context.Context) ([]client.LensConfig, error) {
	txn, err := r.db.NewTxn(ctx, true)
	if err != nil {
		return nil, err
	}
	defer txn.Discard(ctx)
	txnCtx := newTxnCtx(txn)

	return r.registry.config(txnCtx), nil
}

func (r *explicitTxnLensRegistry) Config(ctx context.Context) ([]client.LensConfig, error) {
	return r.registry.config(r.registry.getCtx(r.txn, true)), nil
}

func (r *implicitTxnLensRegistry) HasMigration(ctx context.Context, schemaVersionID string) (bool, error) {
	txn, err := r.db.NewTxn(ctx, true)
	if err != nil {
		return false, err
	}
	defer txn.Discard(ctx)
	txnCtx := newTxnCtx(txn)

	return r.registry.hasMigration(txnCtx, schemaVersionID), nil
}

func (r *explicitTxnLensRegistry) HasMigration(ctx context.Context, schemaVersionID string) (bool, error) {
	return r.registry.hasMigration(r.registry.getCtx(r.txn, true), schemaVersionID), nil
}
