// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

//go:build rust_ffi

package rustffi

/*
#include <stdlib.h>
#include "defra.h"
*/
import "C"

import (
	"context"
	"time"

	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/internal/datastore"
)

var _ client.Txn = (*Transaction)(nil)

// Transaction is an explicit transaction on a Rust node. The calls Go's
// cbindings.Transaction runs in the transaction run in it here too.
type Transaction struct {
	*Wrapper
	handle  C.uintptr_t
	startTS time.Time
}

func (txn *Transaction) ID() uint64 {
	return uint64(txn.handle)
}

func (txn *Transaction) StartTS() time.Time {
	return txn.startTS
}

func (txn *Transaction) Commit() error {
	return take(C.defra_transaction_commit(txn.handle)).err()
}

func (txn *Transaction) Discard() {
	C.defra_transaction_discard(txn.handle)
}

// in is ctx carrying this transaction.
func (txn *Transaction) in(ctx context.Context) context.Context {
	return datastore.CtxSetFromClientTxn(ctx, txn)
}

func (txn *Transaction) PrintDump(ctx context.Context) error {
	return txn.Wrapper.PrintDump(txn.in(ctx))
}

func (txn *Transaction) GetNodeIdentity(ctx context.Context) (immutable.Option[identity.PublicRawIdentity], error) {
	return txn.Wrapper.GetNodeIdentity(ctx)
}

func (txn *Transaction) VerifySignature(
	ctx context.Context,
	blockCid string,
	pubKey crypto.PublicKey,
	opts ...options.Enumerable[options.VerifySignatureOptions],
) error {
	return txn.Wrapper.VerifySignature(txn.in(ctx), blockCid, pubKey, opts...)
}

func (txn *Transaction) AddCollection(
	ctx context.Context,
	sdl string,
	opts ...options.Enumerable[options.AddCollectionOptions],
) ([]client.CollectionVersion, error) {
	return txn.Wrapper.AddCollection(txn.in(ctx), sdl, opts...)
}

func (txn *Transaction) PatchCollection(
	ctx context.Context,
	patch string,
	migration immutable.Option[model.Lens],
	opts ...options.Enumerable[options.PatchCollectionOptions],
) error {
	return txn.Wrapper.PatchCollection(txn.in(ctx), patch, migration, opts...)
}

func (txn *Transaction) DeleteCollection(
	ctx context.Context,
	names []string,
	opts ...options.Enumerable[options.DeleteCollectionOptions],
) error {
	return txn.Wrapper.DeleteCollection(txn.in(ctx), names, opts...)
}

func (txn *Transaction) SetActiveCollectionVersion(
	ctx context.Context,
	version string,
	opts ...options.Enumerable[options.SetActiveCollectionVersionOptions],
) error {
	return txn.Wrapper.SetActiveCollectionVersion(txn.in(ctx), version, opts...)
}

func (txn *Transaction) AddView(
	ctx context.Context,
	gqlQuery string,
	sdl string,
	opts ...options.Enumerable[options.AddViewOptions],
) ([]client.CollectionVersion, error) {
	return txn.Wrapper.AddView(txn.in(ctx), gqlQuery, sdl, opts...)
}

func (txn *Transaction) RefreshViews(
	ctx context.Context, opts ...options.Enumerable[options.RefreshViewsOptions],
) error {
	return txn.Wrapper.RefreshViews(txn.in(ctx), opts...)
}

func (txn *Transaction) SetMigration(
	ctx context.Context, config client.LensConfig, opts ...options.Enumerable[options.SetMigrationOptions],
) (string, error) {
	return txn.Wrapper.SetMigration(txn.in(ctx), config, opts...)
}

func (txn *Transaction) AddLens(
	ctx context.Context,
	lens model.Lens,
	opts ...options.Enumerable[options.AddLensOptions],
) (string, error) {
	return txn.Wrapper.AddLens(txn.in(ctx), lens, opts...)
}

func (txn *Transaction) ListLenses(
	ctx context.Context,
	opts ...options.Enumerable[options.ListLensesOptions],
) (map[string]model.Lens, error) {
	return txn.Wrapper.ListLenses(txn.in(ctx), opts...)
}

func (txn *Transaction) GetCollectionByName(
	ctx context.Context,
	name client.CollectionName,
	opts ...options.Enumerable[options.GetCollectionByNameOptions],
) (client.Collection, error) {
	return txn.Wrapper.GetCollectionByName(txn.in(ctx), name, opts...)
}

func (txn *Transaction) GetCollections(
	ctx context.Context,
	opts ...options.Enumerable[options.GetCollectionsOptions],
) ([]client.Collection, error) {
	return txn.Wrapper.GetCollections(txn.in(ctx), opts...)
}

func (txn *Transaction) ListIndexes(
	ctx context.Context,
	opts ...options.Enumerable[options.ListIndexesOptions],
) (map[client.CollectionName][]client.ListIndexesResult, error) {
	return txn.Wrapper.ListIndexes(txn.in(ctx), opts...)
}

func (txn *Transaction) ExecRequest(
	ctx context.Context,
	request string,
	opts ...options.Enumerable[options.ExecRequestOptions],
) *client.RequestResult {
	return txn.Wrapper.ExecRequest(txn.in(ctx), request, opts...)
}

func (txn *Transaction) BasicImport(ctx context.Context, filepath string) error {
	return txn.Wrapper.BasicImport(txn.in(ctx), filepath)
}

func (txn *Transaction) BasicExport(
	ctx context.Context,
	filepath string,
	opts ...options.Enumerable[options.BasicExportOptions],
) error {
	return txn.Wrapper.BasicExport(txn.in(ctx), filepath, opts...)
}
