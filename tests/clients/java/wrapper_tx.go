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

//go:build javaclient

package java

/*
#include <jni.h>
#include "jnicall.h"
*/
import "C"

import (
	"context"
	"sync"
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

var _ client.Txn = (*Txn)(nil)
var _ datastore.Txn = (*Txn)(nil)

type Txn struct {
	*Wrapper
	tx datastore.Txn

	handle uintptr
	txnObj C.jobject // the constructed DefraTransaction Java object

	// We track whether the transaction has been committed or discarded to prevent double-finalization.
	// We protect this with a mutex to defend against concurrent calls to Commit and Discard.
	finalizeMu sync.Mutex
	finalized  bool
}

func (txn *Txn) ID() uint64 {
	return txn.tx.ID()
}

func (txn *Txn) StartTS() time.Time {
	return txn.tx.StartTS()
}

// Commit commits the transaction.
//
// Once this returns, the transaction is always finalized. That mirrors the native CommitTransaction
// it calls into (see: cbindings/transaction_commit.go), which deletes the underlying cgo.Handle in a
// deferred call regardless of whether the commit itself succeeded. By the time this method returns,
// the native handle is gone, so there is nothing left for a later Discard to roll back even on failure.
//
//	A second call to Commit (or a Commit racing a Discard) does not repeat the native call, which would
//
// panic, since the handle it would look up has already been deleted. Instead, it returns an error.
func (txn *Txn) Commit() error {
	txn.finalizeMu.Lock()
	defer txn.finalizeMu.Unlock()
	if txn.finalized {
		return client.ErrTransactionNotFound
	}
	err := commitTransaction(txn.txnObj, txn.handle)
	txn.finalize()
	return err
}

// Discard discards the transaction, unless Commit already finalized it (see Commit's comment) -
// including when Commit failed, since the native side has already released the transaction either
// way by the time Commit returns.
func (txn *Txn) Discard() {
	txn.finalizeMu.Lock()
	defer txn.finalizeMu.Unlock()
	if txn.finalized {
		return
	}
	discardTransaction(txn.txnObj, txn.handle)
	txn.finalize()
}

// finalize marks the transaction finalized, releases txnObj's JNI global reference, and zeroes
// handle/txnObj so that any use that manages to slip past the finalized guard fails on a
// zero/invalid handle rather than silently reusing now-invalid native resources.
// Callers must hold the finalizeMu lock.
func (txn *Txn) finalize() {
	txn.finalized = true
	if env, detach, err := attach(); err == nil {
		C.defra_delete_global_ref(env, txn.txnObj)
		detach()
	}
	txn.handle = 0
	txn.txnObj = 0
}

// isFinalized reports whether Commit or Discard has already run.
func (txn *Txn) isFinalized() bool {
	txn.finalizeMu.Lock()
	defer txn.finalizeMu.Unlock()
	return txn.finalized
}

func (txn *Txn) PrintDump(ctx context.Context) error {
	return txn.Wrapper.PrintDump(ctx)
}

func (txn *Txn) AddDACPolicy(
	ctx context.Context, policy string, opts ...options.Enumerable[options.AddDACPolicyOptions],
) (client.AddPolicyResult, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.AddDACPolicy(ctx, policy, opts...)
}

func (txn *Txn) AddDACActorRelationship(
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

func (txn *Txn) DeleteDACActorRelationship(
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

func (txn *Txn) AddNACActorRelationship(
	ctx context.Context,
	relation string,
	targetActor string,
	opts ...options.Enumerable[options.AddNACActorRelationshipOptions],
) (client.AddActorRelationshipResult, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.AddNACActorRelationship(ctx, relation, targetActor, opts...)
}

func (txn *Txn) DeleteNACActorRelationship(
	ctx context.Context,
	relation string,
	targetActor string,
	opts ...options.Enumerable[options.DeleteNACActorRelationshipOptions],
) (client.DeleteActorRelationshipResult, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.DeleteNACActorRelationship(ctx, relation, targetActor, opts...)
}

func (txn *Txn) ReEnableNAC(ctx context.Context, opts ...options.Enumerable[options.ReEnableNACOptions]) error {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.ReEnableNAC(ctx, opts...)
}

func (txn *Txn) DisableNAC(ctx context.Context, opts ...options.Enumerable[options.DisableNACOptions]) error {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.DisableNAC(ctx, opts...)
}

func (txn *Txn) GetNACStatus(
	ctx context.Context, opts ...options.Enumerable[options.GetNACStatusOptions],
) (client.NACStatusResult, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.GetNACStatus(ctx, opts...)
}

func (txn *Txn) GetNodeIdentity(ctx context.Context) (immutable.Option[identity.PublicRawIdentity], error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.GetNodeIdentity(ctx)
}

func (txn *Txn) VerifySignature(
	ctx context.Context, blockCid string, pubKey crypto.PublicKey, opts ...options.Enumerable[options.VerifySignatureOptions],
) error {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.VerifySignature(ctx, blockCid, pubKey, opts...)
}

func (txn *Txn) AddCollection(
	ctx context.Context, sdl string, opts ...options.Enumerable[options.AddCollectionOptions],
) ([]client.CollectionVersion, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.AddCollection(ctx, sdl, opts...)
}

func (txn *Txn) PatchCollection(
	ctx context.Context,
	patch string,
	migration immutable.Option[model.Lens],
	opts ...options.Enumerable[options.PatchCollectionOptions],
) error {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.PatchCollection(ctx, patch, migration, opts...)
}

func (txn *Txn) DeleteCollection(
	ctx context.Context, names []string, opts ...options.Enumerable[options.DeleteCollectionOptions],
) error {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.DeleteCollection(ctx, names, opts...)
}

func (txn *Txn) SetActiveCollectionVersion(
	ctx context.Context, version string, opts ...options.Enumerable[options.SetActiveCollectionVersionOptions],
) error {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.SetActiveCollectionVersion(ctx, version, opts...)
}

func (txn *Txn) AddView(
	ctx context.Context, gqlQuery string, sdl string, opts ...options.Enumerable[options.AddViewOptions],
) ([]client.CollectionVersion, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.AddView(ctx, gqlQuery, sdl, opts...)
}

func (txn *Txn) RefreshViews(ctx context.Context, opts ...options.Enumerable[options.RefreshViewsOptions]) error {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.RefreshViews(ctx, opts...)
}

func (txn *Txn) SetMigration(
	ctx context.Context, config client.LensConfig, opts ...options.Enumerable[options.SetMigrationOptions],
) (string, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.SetMigration(ctx, config, opts...)
}

func (txn *Txn) AddLens(
	ctx context.Context, lens model.Lens, opts ...options.Enumerable[options.AddLensOptions],
) (string, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.AddLens(ctx, lens, opts...)
}

func (txn *Txn) ListLenses(
	ctx context.Context, opts ...options.Enumerable[options.ListLensesOptions],
) (map[string]model.Lens, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.ListLenses(ctx, opts...)
}

func (txn *Txn) GetCollectionByName(
	ctx context.Context, name client.CollectionName, opts ...options.Enumerable[options.GetCollectionByNameOptions],
) (client.Collection, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.GetCollectionByName(ctx, name, opts...)
}

func (txn *Txn) GetCollections(
	ctx context.Context, opts ...options.Enumerable[options.GetCollectionsOptions],
) ([]client.Collection, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.GetCollections(ctx, opts...)
}

func (txn *Txn) ListIndexes(
	ctx context.Context, opts ...options.Enumerable[options.ListIndexesOptions],
) (map[client.CollectionName][]client.ListIndexesResult, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.ListIndexes(ctx, opts...)
}

func (txn *Txn) ListAllEncryptedIndexes(
	ctx context.Context, opts ...options.Enumerable[options.ListAllEncryptedIndexesOptions],
) (map[client.CollectionName][]client.EncryptedIndexDescription, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.ListAllEncryptedIndexes(ctx, opts...)
}

func (txn *Txn) ExecRequest(
	ctx context.Context, request string, opts ...options.Enumerable[options.ExecRequestOptions],
) *client.RequestResult {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.ExecRequest(ctx, request, opts...)
}

func (txn *Txn) BasicImport(ctx context.Context, filepath string) error {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.BasicImport(ctx, filepath)
}

func (txn *Txn) BasicExport(
	ctx context.Context, filepath string, opts ...options.Enumerable[options.BasicExportOptions],
) error {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.BasicExport(ctx, filepath, opts...)
}

func (txn *Txn) ListActions(
	ctx context.Context, opts ...options.Enumerable[options.ListActionsOptions],
) ([]client.ActionExecution, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.ListActions(ctx, opts...)
}

func (txn *Txn) PeerInfo(ctx context.Context, opts ...options.Enumerable[options.PeerInfoOptions]) ([]string, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.PeerInfo(ctx, opts...)
}

func (txn *Txn) ActivePeers(
	ctx context.Context, opts ...options.Enumerable[options.ActivePeersOptions],
) ([]string, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.ActivePeers(ctx, opts...)
}

func (txn *Txn) Connect(ctx context.Context, addresses []string, opts ...options.Enumerable[options.ConnectOptions]) error {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.Connect(ctx, addresses, opts...)
}

func (txn *Txn) Disconnect(
	ctx context.Context, addresses []string, opts ...options.Enumerable[options.DisconnectOptions],
) error {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.Disconnect(ctx, addresses, opts...)
}

func (txn *Txn) AddReplicator(
	ctx context.Context, addresses []string, opts ...options.Enumerable[options.AddReplicatorOptions],
) error {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.AddReplicator(ctx, addresses, opts...)
}

func (txn *Txn) DeleteReplicator(
	ctx context.Context, id string, opts ...options.Enumerable[options.DeleteReplicatorOptions],
) error {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.DeleteReplicator(ctx, id, opts...)
}

func (txn *Txn) ListReplicators(
	ctx context.Context, opts ...options.Enumerable[options.ListReplicatorsOptions],
) ([]client.Replicator, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.ListReplicators(ctx, opts...)
}

func (txn *Txn) AddP2PCollections(
	ctx context.Context, collectionNames []string, opts ...options.Enumerable[options.AddP2PCollectionsOptions],
) error {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.AddP2PCollections(ctx, collectionNames, opts...)
}

func (txn *Txn) DeleteP2PCollections(
	ctx context.Context, collectionNames []string, opts ...options.Enumerable[options.DeleteP2PCollectionsOptions],
) error {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.DeleteP2PCollections(ctx, collectionNames, opts...)
}

func (txn *Txn) ListP2PCollections(
	ctx context.Context, opts ...options.Enumerable[options.ListP2PCollectionsOptions],
) ([]string, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.ListP2PCollections(ctx, opts...)
}

func (txn *Txn) AddP2PDocuments(
	ctx context.Context, docIDs []string, opts ...options.Enumerable[options.AddP2PDocumentsOptions],
) error {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.AddP2PDocuments(ctx, docIDs, opts...)
}

func (txn *Txn) DeleteP2PDocuments(
	ctx context.Context, docIDs []string, opts ...options.Enumerable[options.DeleteP2PDocumentsOptions],
) error {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.DeleteP2PDocuments(ctx, docIDs, opts...)
}

func (txn *Txn) ListP2PDocuments(
	ctx context.Context, opts ...options.Enumerable[options.ListP2PDocumentsOptions],
) ([]string, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.ListP2PDocuments(ctx, opts...)
}

func (txn *Txn) SyncDocuments(
	ctx context.Context, collectionName string, docIDs []string, opts ...options.Enumerable[options.SyncDocumentsOptions],
) error {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.SyncDocuments(ctx, collectionName, docIDs, opts...)
}

func (txn *Txn) SyncCollectionVersions(
	ctx context.Context, versionIDs []string, opts ...options.Enumerable[options.SyncCollectionVersionsOptions],
) error {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.SyncCollectionVersions(ctx, versionIDs, opts...)
}

func (txn *Txn) SyncBranchableCollection(
	ctx context.Context, collectionID string, opts ...options.Enumerable[options.SyncBranchableCollectionOptions],
) error {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Wrapper.SyncBranchableCollection(ctx, collectionID, opts...)
}

func (txn *Txn) Blockstore() datastore.Blockstore {
	return txn.tx.Blockstore()
}

func (txn *Txn) Datastore() datastore.Keyedstore {
	return txn.tx.Datastore()
}

func (txn *Txn) Encstore() datastore.Blockstore {
	return txn.tx.Encstore()
}

func (txn *Txn) Headstore() corekv.ReaderWriter {
	return txn.tx.Headstore()
}

func (txn *Txn) Peerstore() corekv.ReaderWriter {
	return txn.tx.Peerstore()
}

func (txn *Txn) Rootstore() corekv.ReaderWriter {
	return txn.tx.Rootstore()
}

func (txn *Txn) Systemstore() corekv.ReaderWriter {
	return txn.tx.Systemstore()
}

func (txn *Txn) OnSuccess(fn func()) {
	txn.tx.OnSuccess(fn)
}

func (txn *Txn) OnError(fn func()) {
	txn.tx.OnError(fn)
}

func (txn *Txn) OnDiscard(fn func()) {
	txn.tx.OnDiscard(fn)
}

func (txn *Txn) OnSuccessAsync(fn func()) {
	txn.tx.OnSuccessAsync(fn)
}

func (txn *Txn) OnErrorAsync(fn func()) {
	txn.tx.OnErrorAsync(fn)
}

func (txn *Txn) OnDiscardAsync(fn func()) {
	txn.tx.OnDiscardAsync(fn)
}
