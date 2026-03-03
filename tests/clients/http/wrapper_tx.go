// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package http

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

// Transaction combines a Wrapper and transaction into
// a single struct that implements the client.Txn interface.
type Transaction struct {
	*Wrapper
	txn client.Txn
}

func (txn *Transaction) ID() uint64 {
	return txn.txn.ID()
}

func (txn *Transaction) StartTS() time.Time {
	return txn.txn.StartTS()
}

func (txn *Transaction) Commit() error {
	return txn.txn.Commit()
}

func (txn *Transaction) Discard() {
	txn.txn.Discard()
}

func (txn *Transaction) PrintDump(ctx context.Context) error {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.PrintDump(ctx)
}

func (txn *Transaction) AddDACPolicy(
	ctx context.Context,
	policy string,
	opts ...options.Enumerable[options.AddDACPolicyOptions],
) (client.AddPolicyResult, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.AddDACPolicy(ctx, policy, opts...)
}

func (txn *Transaction) AddDACActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
	opts ...options.Enumerable[options.AddDACActorRelationshipOptions],
) (client.AddActorRelationshipResult, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.AddDACActorRelationship(ctx, collectionName, docID, relation, targetActor, opts...)
}

func (txn *Transaction) DeleteDACActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
	opts ...options.Enumerable[options.DeleteDACActorRelationshipOptions],
) (client.DeleteActorRelationshipResult, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.DeleteDACActorRelationship(ctx, collectionName, docID, relation, targetActor, opts...)
}

func (txn *Transaction) GetNodeIdentity(ctx context.Context) (immutable.Option[identity.PublicRawIdentity], error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.GetNodeIdentity(ctx)
}

func (txn *Transaction) VerifySignature(
	ctx context.Context,
	blockCid string,
	pubKey crypto.PublicKey,
	opts ...options.Enumerable[options.VerifySignatureOptions],
) error {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.VerifySignature(ctx, blockCid, pubKey, opts...)
}

func (txn *Transaction) AddCollection(
	ctx context.Context,
	sdl string,
	opts ...options.Enumerable[options.AddCollectionOptions],
) ([]client.CollectionVersion, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.AddCollection(ctx, sdl, opts...)
}

func (txn *Transaction) PatchCollection(
	ctx context.Context,
	patch string,
	migration immutable.Option[model.Lens],
	opts ...options.Enumerable[options.PatchCollectionOptions],
) error {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.PatchCollection(ctx, patch, migration, opts...)
}

func (txn *Transaction) SetActiveCollectionVersion(
	ctx context.Context,
	version string,
	opts ...options.Enumerable[options.SetActiveCollectionVersionOptions],
) error {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.SetActiveCollectionVersion(ctx, version, opts...)
}

func (txn *Transaction) AddView(
	ctx context.Context,
	gqlQuery string,
	sdl string,
	opts ...options.Enumerable[options.AddViewOptions],
) ([]client.CollectionVersion, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.AddView(ctx, gqlQuery, sdl, opts...)
}

func (txn *Transaction) RefreshViews(
	ctx context.Context, opts ...options.Enumerable[options.RefreshViewsOptions],
) error {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.RefreshViews(ctx, opts...)
}

func (txn *Transaction) SetMigration(
	ctx context.Context, config client.LensConfig, opts ...options.Enumerable[options.SetMigrationOptions],
) (string, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.SetMigration(ctx, config, opts...)
}

func (txn *Transaction) AddLens(
	ctx context.Context,
	lens model.Lens,
	opts ...options.Enumerable[options.AddLensOptions],
) (string, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.AddLens(ctx, lens, opts...)
}

func (txn *Transaction) ListLenses(
	ctx context.Context,
	opts ...options.Enumerable[options.ListLensesOptions],
) (map[string]model.Lens, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.ListLenses(ctx, opts...)
}

func (txn *Transaction) GetCollectionByName(
	ctx context.Context,
	name client.CollectionName,
	opts ...options.Enumerable[options.GetCollectionByNameOptions],
) (client.Collection, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.GetCollectionByName(ctx, name, opts...)
}

func (txn *Transaction) GetCollections(
	ctx context.Context,
	opts ...options.Enumerable[options.GetCollectionsOptions],
) ([]client.Collection, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.GetCollections(ctx, opts...)
}

func (txn *Transaction) ListIndexes(
	ctx context.Context,
	opts ...options.Enumerable[options.ListIndexesOptions],
) (map[client.CollectionName][]client.IndexDescription, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.ListIndexes(ctx, opts...)
}

func (txn *Transaction) ExecRequest(
	ctx context.Context,
	request string,
	opts ...options.Enumerable[options.ExecRequestOptions],
) *client.RequestResult {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.ExecRequest(ctx, request, opts...)
}

func (txn *Transaction) BasicImport(ctx context.Context, filepath string) error {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.BasicImport(ctx, filepath)
}

func (txn *Transaction) BasicExport(
	ctx context.Context,
	filepath string,
	opts ...options.Enumerable[options.BasicExportOptions],
) error {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.BasicExport(ctx, filepath, opts...)
}
