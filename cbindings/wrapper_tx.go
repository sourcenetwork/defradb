// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cbindings

/*
#include <stdlib.h>
#include <stdint.h>
#include "defra_structs.h"
extern Result CommitTransaction(uintptr_t txnPtr);
extern void DiscardTransaction(uintptr_t txnPtr);
*/
import "C"

import (
	"context"
	"errors"
	"runtime/cgo"
	"time"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/internal/datastore"
)

var _ client.Txn = (*Transaction)(nil)

type Transaction struct {
	*CWrapper
	tx     datastore.Txn
	handle cgo.Handle
}

func (txn *Transaction) ID() uint64 {
	return txn.tx.ID()
}

func (txn *Transaction) StartTS() time.Time {
	return txn.tx.StartTS()
}

func (txn *Transaction) Commit() error {
	res := ConvertAndFreeCResult(C.CommitTransaction(C.uintptr_t(txn.handle)))
	txnHandleMap.Delete(txn.ID())
	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (txn *Transaction) Discard() {
	C.DiscardTransaction(C.uintptr_t(txn.handle))
	txnHandleMap.Delete(txn.ID())
}

func (txn *Transaction) PrintDump(ctx context.Context) error {
	return txn.CWrapper.PrintDump(ctx)
}

func (txn *Transaction) AddDACPolicy(
	ctx context.Context,
	policy string,
	opts ...options.Enumerable[options.AddDACPolicyOptions],
) (client.AddPolicyResult, error) {
	return txn.CWrapper.AddDACPolicy(ctx, policy, opts...)
}

func (txn *Transaction) AddDACActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
	opts ...options.Enumerable[options.AddDACActorRelationshipOptions],
) (client.AddActorRelationshipResult, error) {
	return txn.CWrapper.AddDACActorRelationship(ctx, collectionName, docID, relation, targetActor, opts...)
}

func (txn *Transaction) DeleteDACActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
	opts ...options.Enumerable[options.DeleteDACActorRelationshipOptions],
) (client.DeleteActorRelationshipResult, error) {
	return txn.CWrapper.DeleteDACActorRelationship(ctx, collectionName, docID, relation, targetActor, opts...)
}

func (txn *Transaction) GetNodeIdentity(ctx context.Context) (immutable.Option[identity.PublicRawIdentity], error) {
	return txn.CWrapper.GetNodeIdentity(ctx)
}

func (txn *Transaction) VerifySignature(
	ctx context.Context,
	blockCid string,
	pubKey crypto.PublicKey,
	opts ...options.Enumerable[options.VerifySignatureOptions],
) error {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.CWrapper.VerifySignature(ctx, blockCid, pubKey, opts...)
}

func (txn *Transaction) AddCollection(
	ctx context.Context,
	sdl string,
	opts ...options.Enumerable[options.AddCollectionOptions],
) ([]client.CollectionVersion, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.CWrapper.AddCollection(ctx, sdl, opts...)
}

func (txn *Transaction) PatchCollection(
	ctx context.Context,
	patch string,
	migration immutable.Option[model.Lens],
	opts ...options.Enumerable[options.PatchCollectionOptions],
) error {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.CWrapper.PatchCollection(ctx, patch, migration, opts...)
}

func (txn *Transaction) SetActiveCollectionVersion(
	ctx context.Context,
	version string,
	opts ...options.Enumerable[options.SetActiveCollectionVersionOptions],
) error {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.CWrapper.SetActiveCollectionVersion(ctx, version, opts...)
}

func (txn *Transaction) AddView(
	ctx context.Context,
	gqlQuery string,
	sdl string,
	opts ...options.Enumerable[options.AddViewOptions],
) ([]client.CollectionVersion, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.CWrapper.AddView(ctx, gqlQuery, sdl, opts...)
}

func (txn *Transaction) RefreshViews(
	ctx context.Context, opts ...options.Enumerable[options.RefreshViewsOptions],
) error {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.CWrapper.RefreshViews(ctx, opts...)
}

func (txn *Transaction) SetMigration(
	ctx context.Context, config client.LensConfig, opts ...options.Enumerable[options.SetMigrationOptions],
) (string, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.CWrapper.SetMigration(ctx, config, opts...)
}

func (txn *Transaction) AddLens(
	ctx context.Context,
	lens model.Lens,
	opts ...options.Enumerable[options.AddLensOptions],
) (string, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.CWrapper.AddLens(ctx, lens, opts...)
}

func (txn *Transaction) ListLenses(
	ctx context.Context,
	opts ...options.Enumerable[options.ListLensesOptions],
) (map[string]model.Lens, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.CWrapper.ListLenses(ctx, opts...)
}

func (txn *Transaction) GetCollectionByName(
	ctx context.Context,
	name client.CollectionName,
	opts ...options.Enumerable[options.GetCollectionByNameOptions],
) (client.Collection, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.CWrapper.GetCollectionByName(ctx, name, opts...)
}

func (txn *Transaction) GetCollections(
	ctx context.Context,
	opts ...options.Enumerable[options.GetCollectionsOptions],
) ([]client.Collection, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.CWrapper.GetCollections(ctx, opts...)
}

func (txn *Transaction) ListIndexes(
	ctx context.Context,
	opts ...options.Enumerable[options.ListIndexesOptions],
) (map[client.CollectionName][]client.IndexDescription, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.CWrapper.ListIndexes(ctx, opts...)
}

func (txn *Transaction) ExecRequest(
	ctx context.Context,
	request string,
	opts ...options.Enumerable[options.ExecRequestOptions],
) *client.RequestResult {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.CWrapper.ExecRequest(ctx, request, opts...)
}

func (txn *Transaction) BasicImport(ctx context.Context, filepath string) error {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.CWrapper.BasicImport(ctx, filepath)
}

func (txn *Transaction) BasicExport(
	ctx context.Context,
	filepath string,
	opts ...options.Enumerable[options.BasicExportOptions],
) error {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.CWrapper.BasicExport(ctx, filepath, opts...)
}

func (txn *Transaction) Blockstore() datastore.Blockstore {
	return txn.tx.Blockstore()
}

func (txn *Transaction) Datastore() datastore.Keyedstore {
	return txn.tx.Datastore()
}

func (txn *Transaction) Encstore() datastore.Blockstore {
	return txn.tx.Encstore()
}

func (txn *Transaction) Headstore() corekv.ReaderWriter {
	return txn.tx.Headstore()
}

func (txn *Transaction) Peerstore() corekv.ReaderWriter {
	return txn.tx.Peerstore()
}

func (txn *Transaction) Rootstore() corekv.ReaderWriter {
	return txn.tx.Rootstore()
}

func (txn *Transaction) Systemstore() corekv.ReaderWriter {
	return txn.tx.Systemstore()
}

func (txn *Transaction) OnSuccess(fn func()) {
	txn.tx.OnSuccess(fn)
}

func (txn *Transaction) OnError(fn func()) {
	txn.tx.OnError(fn)
}

func (txn *Transaction) OnDiscard(fn func()) {
	txn.tx.OnDiscard(fn)
}

func (txn *Transaction) OnSuccessAsync(fn func()) {
	txn.tx.OnSuccessAsync(fn)
}

func (txn *Transaction) OnErrorAsync(fn func()) {
	txn.tx.OnErrorAsync(fn)
}

func (txn *Transaction) OnDiscardAsync(fn func()) {
	txn.tx.OnDiscardAsync(fn)
}
