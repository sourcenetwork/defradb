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
extern Result TransactionCommit(uintptr_t txnPtr);
extern void TransactionDiscard(uintptr_t txnPtr);
*/
import "C"

import (
	"context"
	"errors"
	"runtime/cgo"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/internal/datastore"
)

var _ client.Txn = (*Transaction)(nil)

type Transaction struct {
	*CWrapper
	tx     client.Txn
	handle cgo.Handle
}

func (txn *Transaction) ID() uint64 {
	return txn.tx.ID()
}

func (txn *Transaction) Commit() error {
	res := ConvertAndFreeCResult(C.TransactionCommit(C.uintptr_t(txn.handle)))
	txnHandleMap.Delete(txn.ID())
	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (txn *Transaction) Discard() {
	C.TransactionDiscard(C.uintptr_t(txn.handle))
	txnHandleMap.Delete(txn.ID())
}

func (txn *Transaction) PrintDump(ctx context.Context) error {
	return txn.CWrapper.PrintDump(ctx)
}

func (txn *Transaction) AddDACPolicy(ctx context.Context, policy string) (client.AddPolicyResult, error) {
	return txn.CWrapper.AddDACPolicy(ctx, policy)
}

func (txn *Transaction) AddDACActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
) (client.AddActorRelationshipResult, error) {
	return txn.CWrapper.AddDACActorRelationship(ctx, collectionName, docID, relation, targetActor)
}

func (txn *Transaction) DeleteDACActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
) (client.DeleteActorRelationshipResult, error) {
	return txn.CWrapper.DeleteDACActorRelationship(ctx, collectionName, docID, relation, targetActor)
}

func (txn *Transaction) GetNodeIdentity(ctx context.Context) (immutable.Option[identity.PublicRawIdentity], error) {
	return txn.CWrapper.GetNodeIdentity(ctx)
}

func (txn *Transaction) VerifySignature(ctx context.Context, blockCid string, pubKey crypto.PublicKey) error {
	return txn.CWrapper.VerifySignature(ctx, blockCid, pubKey)
}

func (txn *Transaction) AddSchema(ctx context.Context, sdl string) ([]client.CollectionVersion, error) {
	return txn.CWrapper.AddSchema(ctx, sdl)
}

func (txn *Transaction) PatchCollection(
	ctx context.Context,
	patch string,
	migration immutable.Option[model.Lens],
) error {
	return txn.CWrapper.PatchCollection(ctx, patch, migration)
}

func (txn *Transaction) SetActiveCollectionVersion(ctx context.Context, version string) error {
	return txn.CWrapper.SetActiveCollectionVersion(ctx, version)
}

func (txn *Transaction) AddView(
	ctx context.Context,
	gqlQuery string,
	sdl string,
	transform immutable.Option[model.Lens],
) ([]client.CollectionVersion, error) {
	return txn.CWrapper.AddView(ctx, gqlQuery, sdl, transform)
}

func (txn *Transaction) RefreshViews(ctx context.Context, options client.CollectionFetchOptions) error {
	return txn.CWrapper.RefreshViews(ctx, options)
}

func (txn *Transaction) SetMigration(ctx context.Context, config client.LensConfig) (string, error) {
	return txn.CWrapper.SetMigration(ctx, config)
}

func (txn *Transaction) GetCollectionByName(
	ctx context.Context,
	name client.CollectionName,
) (client.Collection, error) {
	return txn.CWrapper.GetCollectionByName(ctx, name)
}

func (txn *Transaction) GetCollections(
	ctx context.Context,
	options client.CollectionFetchOptions,
) ([]client.Collection, error) {
	return txn.CWrapper.GetCollections(ctx, options)
}

func (txn *Transaction) GetAllIndexes(
	ctx context.Context,
) (map[client.CollectionName][]client.IndexDescription, error) {
	return txn.CWrapper.GetAllIndexes(ctx)
}

func (txn *Transaction) ExecRequest(
	ctx context.Context,
	request string,
	opts ...client.RequestOption,
) *client.RequestResult {
	return txn.CWrapper.ExecRequest(ctx, request, opts...)
}

func (txn *Transaction) BasicImport(ctx context.Context, filepath string) error {
	return txn.CWrapper.BasicImport(ctx, filepath)
}

func (txn *Transaction) BasicExport(ctx context.Context, config *client.BackupConfig) error {
	return txn.CWrapper.BasicExport(ctx, config)
}

func (txn *Transaction) Blockstore() datastore.Blockstore {
	return txn.tx.(datastore.Txn).Blockstore() //nolint:forcetypeassert
}

func (txn *Transaction) Datastore() datastore.Keyedstore {
	return txn.tx.(datastore.Txn).Datastore() //nolint:forcetypeassert
}

func (txn *Transaction) Encstore() datastore.Blockstore {
	return txn.tx.(datastore.Txn).Encstore() //nolint:forcetypeassert
}

func (txn *Transaction) Headstore() corekv.ReaderWriter {
	return txn.tx.(datastore.Txn).Headstore() //nolint:forcetypeassert
}

func (txn *Transaction) Peerstore() corekv.ReaderWriter {
	return txn.tx.(datastore.Txn).Peerstore() //nolint:forcetypeassert
}

func (txn *Transaction) Rootstore() corekv.ReaderWriter {
	return txn.tx.(datastore.Txn).Rootstore() //nolint:forcetypeassert
}

func (txn *Transaction) Systemstore() corekv.ReaderWriter {
	return txn.tx.(datastore.Txn).Systemstore() //nolint:forcetypeassert
}

func (txn *Transaction) OnSuccess(fn func()) {
	txn.tx.(datastore.Txn).OnSuccess(fn) //nolint:forcetypeassert
}

func (txn *Transaction) OnError(fn func()) {
	txn.tx.(datastore.Txn).OnError(fn) //nolint:forcetypeassert
}

func (txn *Transaction) OnDiscard(fn func()) {
	txn.tx.(datastore.Txn).OnDiscard(fn) //nolint:forcetypeassert
}

func (txn *Transaction) OnSuccessAsync(fn func()) {
	txn.tx.(datastore.Txn).OnSuccessAsync(fn) //nolint:forcetypeassert
}

func (txn *Transaction) OnErrorAsync(fn func()) {
	txn.tx.(datastore.Txn).OnErrorAsync(fn) //nolint:forcetypeassert
}

func (txn *Transaction) OnDiscardAsync(fn func()) {
	txn.tx.(datastore.Txn).OnDiscardAsync(fn) //nolint:forcetypeassert
}
