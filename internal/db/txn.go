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
	"sync"

	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/internal/datastore"
)

// ensureContextTxn ensures that the returned context has a transaction.
//
// If a transactions exists on the context it will be made explicit,
// otherwise a new implicit transaction will be created.
//
// The returned context will contain the transaction
// along with the copied values from the input context.
func ensureContextTxn(ctx context.Context, db *DB, readOnly bool) (context.Context, *Txn, error) {
	var ctxTxn any
	var existsOnCtx bool
	ctxTxn, existsOnCtx = datastore.CtxTryGetTxn(ctx)
	if !existsOnCtx {
		var err error
		ctxTxn, err = db.NewTxn(readOnly)
		if err != nil {
			return nil, nil, err
		}
	}

	txn, ok := ctxTxn.(*Txn)
	if !ok {
		return nil, nil, NewErrUnsupportedTxnType(ctxTxn)
	}

	// If the txn has already been set on the context but it hasn't already been set as explicit,
	// we create a copy of the txn and mark it as an explicit txn.
	if !txn.explicit && existsOnCtx {
		txn = &Txn{
			BasicTxn: txn.BasicTxn,
			db:       txn.db,
			explicit: true,
			isClosed: txn.isClosed,
			// We do not need to copy the mutex (or a pointer to it), as if we are doing this,
			// we can be sure that this txn clone is a child of the parent context, and so
			// should not be locking anyway.
		}
	}

	return InitContext(ctx, txn), txn, nil
}

func lockForTxn(ctx context.Context, txn *Txn) (context.Context, func()) {
	type txnCtxInProgressKey uint64

	// Defra's public functions may themselves call other public functions, and we use the
	// context to track this. We must not try to lock using this txn within a child function
	// call, as that would deadlock (the parent already holds the lock).
	//
	// To track this we use the `txnCtxInProgressKey` context key - because setting this creates
	// a new child context, we never need to worry about leakage or cleanup beyond the top-level
	// public txn call.  It also means read/writing it in this way is inherently thread safe - if
	// two threads concurrently try to set it up, they will have two different contexts and the
	// inProgressLock will make them run serially (as they should).
	thisContextOwnsTxnLock := ctx.Value(txnCtxInProgressKey(txn.ID())) == nil
	if thisContextOwnsTxnLock {
		ctx = context.WithValue(ctx, txnCtxInProgressKey(txn.ID()), struct{}{})
		txn.inProgressLock.Lock()
	}

	return ctx, func() {
		if thisContextOwnsTxnLock {
			defer txn.inProgressLock.Unlock()
		}
	}
}

type Txn struct {
	*datastore.BasicTxn
	db       *DB
	explicit bool

	// Badger will panic if a transaction is used after it has been committed/discarded, which is not
	// great for users, and is a pain for us testing, so we handle this upfront here, returning an
	// error instead.
	//
	// We handle this here at this level, instead of corekv, as it allows us to not worry about concurrency
	// due to the protection of the inProgressLock.
	isClosed bool

	// The inProgressLock forces top-level txn-actions to execute serially, preventing concurrent action
	// execution within the transaction.
	//
	// Child Defra public function calls from within top level public Defra functions must *not* attempt to
	// lock this, as that will deadlock.  This is protected against using the context (see `lockForTxn`).
	inProgressLock sync.Mutex
}

var _ client.Txn = (*Txn)(nil)

// wrapDatastoreTxn returns a new Txn from the rootstore.
func wrapDatastoreTxn(txn *datastore.BasicTxn, db *DB) *Txn {
	return &Txn{
		BasicTxn: txn,
		db:       db,
	}
}

func (txn *Txn) Commit() error {
	if txn.explicit {
		// If the transaction has been explicitly defined, `Commit` should
		// only be executed by the transaction creator. As such, a call to
		// `Commit` on an explicit transaction should result in a no-op.
		return nil
	}

	// We lock/unlock without checking the context here, as if a child call to a public Defra function
	// is committing, the code is probably already quite broken, and it is a lot of hassle for both us
	// and users to add context as a param to this function.
	txn.inProgressLock.Lock()
	defer txn.inProgressLock.Unlock()

	err := txn.BasicTxn.Commit()
	if err != nil {
		return err
	}

	txn.isClosed = true
	return nil
}

func (txn *Txn) Discard() {
	if txn.explicit {
		// If the transaction has been explicitly defined, `Discard` should
		// only be executed by the transaction creator. As such, a call to
		// `Discard` on an explicit transaction should result in a no-op.
		return
	}

	// We lock/unlock without checking the context here, as if a child call to a public Defra function
	// is discarding, the code is probably already quite broken, and it is a lot of hassle for both us
	// and users to add context as a param to this function.
	txn.inProgressLock.Lock()
	defer txn.inProgressLock.Unlock()

	txn.BasicTxn.Discard()
	txn.isClosed = true
}

func (txn *Txn) PrintDump(ctx context.Context) error {
	ctx, unlock := lockForTxn(ctx, txn)
	defer unlock()

	if txn.isClosed {
		return ErrTxnDiscarded
	}

	return printStore(ctx, txn.Rootstore())
}

func (txn *Txn) AddDACPolicy(
	ctx context.Context,
	policy string,
	opts ...options.Enumerable[options.AddDACPolicyOptions],
) (client.AddPolicyResult, error) {
	ctx = InitContext(ctx, txn)

	ctx, unlock := lockForTxn(ctx, txn)
	defer unlock()

	if txn.isClosed {
		return client.AddPolicyResult{}, ErrTxnDiscarded
	}

	return txn.db.AddDACPolicy(ctx, policy, opts...)
}

func (txn *Txn) AddDACActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
	opts ...options.Enumerable[options.AddDACActorRelationshipOptions],
) (client.AddActorRelationshipResult, error) {
	ctx = InitContext(ctx, txn)

	ctx, unlock := lockForTxn(ctx, txn)
	defer unlock()

	if txn.isClosed {
		return client.AddActorRelationshipResult{}, ErrTxnDiscarded
	}

	return txn.db.AddDACActorRelationship(ctx, collectionName, docID, relation, targetActor, opts...)
}

func (txn *Txn) DeleteDACActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
	opts ...options.Enumerable[options.DeleteDACActorRelationshipOptions],
) (client.DeleteActorRelationshipResult, error) {
	ctx = InitContext(ctx, txn)

	ctx, unlock := lockForTxn(ctx, txn)
	defer unlock()

	if txn.isClosed {
		return client.DeleteActorRelationshipResult{}, ErrTxnDiscarded
	}

	return txn.db.DeleteDACActorRelationship(ctx, collectionName, docID, relation, targetActor, opts...)
}

func (txn *Txn) AddNACActorRelationship(
	ctx context.Context,
	relation string,
	targetActor string,
	opts ...options.Enumerable[options.AddNACActorRelationshipOptions],
) (client.AddActorRelationshipResult, error) {
	ctx = InitContext(ctx, txn)

	ctx, unlock := lockForTxn(ctx, txn)
	defer unlock()

	if txn.isClosed {
		return client.AddActorRelationshipResult{}, ErrTxnDiscarded
	}

	return txn.db.AddNACActorRelationship(ctx, relation, targetActor, opts...)
}

func (txn *Txn) DeleteNACActorRelationship(
	ctx context.Context,
	relation string,
	targetActor string,
	opts ...options.Enumerable[options.DeleteNACActorRelationshipOptions],
) (client.DeleteActorRelationshipResult, error) {
	ctx = InitContext(ctx, txn)

	ctx, unlock := lockForTxn(ctx, txn)
	defer unlock()

	if txn.isClosed {
		return client.DeleteActorRelationshipResult{}, ErrTxnDiscarded
	}

	return txn.db.DeleteNACActorRelationship(ctx, relation, targetActor, opts...)
}

func (txn *Txn) ReEnableNAC(ctx context.Context, opts ...options.Enumerable[options.ReEnableNACOptions]) error {
	ctx = InitContext(ctx, txn)

	ctx, unlock := lockForTxn(ctx, txn)
	defer unlock()

	if txn.isClosed {
		return ErrTxnDiscarded
	}

	return txn.db.ReEnableNAC(ctx, opts...)
}

func (txn *Txn) DisableNAC(ctx context.Context, opts ...options.Enumerable[options.DisableNACOptions]) error {
	ctx = InitContext(ctx, txn)

	ctx, unlock := lockForTxn(ctx, txn)
	defer unlock()

	if txn.isClosed {
		return ErrTxnDiscarded
	}

	return txn.db.DisableNAC(ctx, opts...)
}

func (txn *Txn) GetNACStatus(
	ctx context.Context,
	opts ...options.Enumerable[options.GetNACStatusOptions],
) (client.NACStatusResult, error) {
	ctx = InitContext(ctx, txn)

	ctx, unlock := lockForTxn(ctx, txn)
	defer unlock()

	if txn.isClosed {
		return client.NACStatusResult{}, ErrTxnDiscarded
	}

	return txn.db.GetNACStatus(ctx, opts...)
}

func (txn *Txn) GetNodeIdentity(ctx context.Context) (immutable.Option[acpIdentity.PublicRawIdentity], error) {
	ctx = InitContext(ctx, txn)

	ctx, unlock := lockForTxn(ctx, txn)
	defer unlock()

	if txn.isClosed {
		return immutable.None[acpIdentity.PublicRawIdentity](), ErrTxnDiscarded
	}

	return txn.db.GetNodeIdentity(ctx)
}

func (txn *Txn) VerifySignature(
	ctx context.Context,
	blockCid string,
	pubKey crypto.PublicKey,
	opts ...options.Enumerable[options.VerifySignatureOptions],
) error {
	ctx = InitContext(ctx, txn)

	ctx, unlock := lockForTxn(ctx, txn)
	defer unlock()

	if txn.isClosed {
		return ErrTxnDiscarded
	}

	return txn.db.VerifySignature(ctx, blockCid, pubKey, opts...)
}

func (txn *Txn) AddCollection(
	ctx context.Context,
	sdl string,
	opts ...options.Enumerable[options.AddCollectionOptions],
) ([]client.CollectionVersion, error) {
	ctx = InitContext(ctx, txn)

	ctx, unlock := lockForTxn(ctx, txn)
	defer unlock()

	if txn.isClosed {
		return nil, ErrTxnDiscarded
	}

	return txn.db.AddCollection(ctx, sdl, opts...)
}

func (txn *Txn) PatchCollection(
	ctx context.Context,
	patch string,
	migration immutable.Option[model.Lens],
	opts ...options.Enumerable[options.PatchCollectionOptions],
) error {
	ctx = InitContext(ctx, txn)

	ctx, unlock := lockForTxn(ctx, txn)
	defer unlock()

	if txn.isClosed {
		return ErrTxnDiscarded
	}

	return txn.db.PatchCollection(ctx, patch, migration, opts...)
}

func (txn *Txn) DeleteCollection(
	ctx context.Context,
	names []string,
	opts ...options.Enumerable[options.DeleteCollectionOptions],
) error {
	ctx = InitContext(ctx, txn)
	return txn.db.DeleteCollection(ctx, names, opts...)
}

func (txn *Txn) SetActiveCollectionVersion(
	ctx context.Context,
	version string,
	opts ...options.Enumerable[options.SetActiveCollectionVersionOptions],
) error {
	ctx = InitContext(ctx, txn)

	ctx, unlock := lockForTxn(ctx, txn)
	defer unlock()

	if txn.isClosed {
		return ErrTxnDiscarded
	}

	return txn.db.SetActiveCollectionVersion(ctx, version, opts...)
}

func (txn *Txn) AddView(
	ctx context.Context,
	gqlQuery string,
	sdl string,
	opts ...options.Enumerable[options.AddViewOptions],
) ([]client.CollectionVersion, error) {
	ctx = InitContext(ctx, txn)

	ctx, unlock := lockForTxn(ctx, txn)
	defer unlock()

	if txn.isClosed {
		return nil, ErrTxnDiscarded
	}

	return txn.db.AddView(ctx, gqlQuery, sdl, opts...)
}

func (txn *Txn) RefreshViews(ctx context.Context, opts ...options.Enumerable[options.RefreshViewsOptions]) error {
	ctx = InitContext(ctx, txn)

	ctx, unlock := lockForTxn(ctx, txn)
	defer unlock()

	if txn.isClosed {
		return ErrTxnDiscarded
	}

	return txn.db.RefreshViews(ctx, opts...)
}

func (txn *Txn) SetMigration(
	ctx context.Context, config client.LensConfig, opts ...options.Enumerable[options.SetMigrationOptions],
) (string, error) {
	ctx = InitContext(ctx, txn)

	ctx, unlock := lockForTxn(ctx, txn)
	defer unlock()

	if txn.isClosed {
		return "", ErrTxnDiscarded
	}

	return txn.db.SetMigration(ctx, config, opts...)
}

func (txn *Txn) AddLens(
	ctx context.Context,
	lens model.Lens,
	opts ...options.Enumerable[options.AddLensOptions],
) (string, error) {
	ctx = InitContext(ctx, txn)

	ctx, unlock := lockForTxn(ctx, txn)
	defer unlock()

	if txn.isClosed {
		return "", ErrTxnDiscarded
	}

	return txn.db.AddLens(ctx, lens, opts...)
}

func (txn *Txn) ListLenses(
	ctx context.Context,
	opts ...options.Enumerable[options.ListLensesOptions],
) (map[string]model.Lens, error) {
	ctx = InitContext(ctx, txn)

	ctx, unlock := lockForTxn(ctx, txn)
	defer unlock()

	if txn.isClosed {
		return nil, ErrTxnDiscarded
	}

	return txn.db.ListLenses(ctx)
}

func (txn *Txn) GetCollectionByName(
	ctx context.Context,
	name client.CollectionName,
	opts ...options.Enumerable[options.GetCollectionByNameOptions],
) (client.Collection, error) {
	ctx = InitContext(ctx, txn)

	ctx, unlock := lockForTxn(ctx, txn)
	defer unlock()

	if txn.isClosed {
		return nil, ErrTxnDiscarded
	}

	col, err := txn.db.GetCollectionByName(ctx, name, opts...)
	if err != nil {
		return nil, err
	}

	return newTxnCollection(txn, col), nil
}

func (txn *Txn) GetCollections(
	ctx context.Context,
	opts ...options.Enumerable[options.GetCollectionsOptions],
) ([]client.Collection, error) {
	ctx = InitContext(ctx, txn)

	ctx, unlock := lockForTxn(ctx, txn)
	defer unlock()

	if txn.isClosed {
		return nil, ErrTxnDiscarded
	}

	cols, err := txn.db.GetCollections(ctx, opts...)
	if err != nil {
		return nil, err
	}

	for i, col := range cols {
		cols[i] = newTxnCollection(txn, col)
	}

	return cols, nil
}

func (txn *Txn) ListIndexes(
	ctx context.Context,
	opts ...options.Enumerable[options.ListIndexesOptions],
) (map[client.CollectionName][]client.IndexDescription, error) {
	ctx = InitContext(ctx, txn)

	ctx, unlock := lockForTxn(ctx, txn)
	defer unlock()

	if txn.isClosed {
		return nil, ErrTxnDiscarded
	}

	return txn.db.ListIndexes(ctx, opts...)
}

func (txn *Txn) ListAllEncryptedIndexes(
	ctx context.Context,
	opts ...options.Enumerable[options.ListAllEncryptedIndexesOptions],
) (map[client.CollectionName][]client.EncryptedIndexDescription, error) {
	ctx = InitContext(ctx, txn)

	ctx, unlock := lockForTxn(ctx, txn)
	defer unlock()

	if txn.isClosed {
		return nil, ErrTxnDiscarded
	}

	return txn.db.ListAllEncryptedIndexes(ctx, opts...)
}

func (txn *Txn) ExecRequest(
	ctx context.Context,
	request string,
	opts ...options.Enumerable[options.ExecRequestOptions],
) *client.RequestResult {
	ctx = InitContext(ctx, txn)

	ctx, unlock := lockForTxn(ctx, txn)
	defer unlock()

	if txn.isClosed {
		return &client.RequestResult{
			GQL: client.GQLResult{
				Errors: []error{ErrTxnDiscarded},
			},
		}
	}

	return txn.db.ExecRequest(ctx, request, opts...)
}

func (txn *Txn) BasicImport(ctx context.Context, filepath string) error {
	ctx = InitContext(ctx, txn)

	ctx, unlock := lockForTxn(ctx, txn)
	defer unlock()

	if txn.isClosed {
		return ErrTxnDiscarded
	}

	return txn.db.BasicImport(ctx, filepath)
}

func (txn *Txn) BasicExport(
	ctx context.Context,
	filepath string,
	opts ...options.Enumerable[options.BasicExportOptions],
) error {
	ctx = InitContext(ctx, txn)

	ctx, unlock := lockForTxn(ctx, txn)
	defer unlock()

	if txn.isClosed {
		return ErrTxnDiscarded
	}

	return txn.db.BasicExport(ctx, filepath, opts...)
}

func (txn *Txn) PeerInfo(ctx context.Context, opts ...options.Enumerable[options.PeerInfoOptions]) ([]string, error) {
	ctx, unlock := lockForTxn(ctx, txn)
	defer unlock()

	if txn.isClosed {
		return nil, ErrTxnDiscarded
	}

	return txn.db.PeerInfo(ctx, opts...)
}

func (txn *Txn) ActivePeers(
	ctx context.Context, opts ...options.Enumerable[options.ActivePeersOptions],
) ([]string, error) {
	ctx, unlock := lockForTxn(ctx, txn)
	defer unlock()

	if txn.isClosed {
		return nil, ErrTxnDiscarded
	}

	return txn.db.ActivePeers(ctx, opts...)
}

func (txn *Txn) Connect(
	ctx context.Context, addresses []string, opts ...options.Enumerable[options.ConnectOptions],
) error {
	ctx, unlock := lockForTxn(ctx, txn)
	defer unlock()

	if txn.isClosed {
		return ErrTxnDiscarded
	}

	return txn.db.Connect(ctx, addresses, opts...)
}

func (txn *Txn) AddReplicator(
	ctx context.Context,
	addresses []string,
	opts ...options.Enumerable[options.AddReplicatorOptions],
) error {
	ctx = InitContext(ctx, txn)

	ctx, unlock := lockForTxn(ctx, txn)
	defer unlock()

	if txn.isClosed {
		return ErrTxnDiscarded
	}

	return txn.db.AddReplicator(ctx, addresses, opts...)
}

func (txn *Txn) DeleteReplicator(
	ctx context.Context,
	id string,
	opts ...options.Enumerable[options.DeleteReplicatorOptions],
) error {
	ctx = InitContext(ctx, txn)

	ctx, unlock := lockForTxn(ctx, txn)
	defer unlock()

	if txn.isClosed {
		return ErrTxnDiscarded
	}

	return txn.db.DeleteReplicator(ctx, id, opts...)
}

func (txn *Txn) ListReplicators(
	ctx context.Context,
	opts ...options.Enumerable[options.ListReplicatorsOptions],
) ([]client.Replicator, error) {
	ctx = InitContext(ctx, txn)

	ctx, unlock := lockForTxn(ctx, txn)
	defer unlock()

	if txn.isClosed {
		return nil, ErrTxnDiscarded
	}

	return txn.db.ListReplicators(ctx, opts...)
}

func (txn *Txn) AddP2PCollections(
	ctx context.Context,
	collectionNames []string,
	opts ...options.Enumerable[options.AddP2PCollectionsOptions],
) error {
	ctx = InitContext(ctx, txn)

	ctx, unlock := lockForTxn(ctx, txn)
	defer unlock()

	if txn.isClosed {
		return ErrTxnDiscarded
	}

	return txn.db.AddP2PCollections(ctx, collectionNames, opts...)
}

func (txn *Txn) DeleteP2PCollections(
	ctx context.Context,
	collectionNames []string,
	opts ...options.Enumerable[options.DeleteP2PCollectionsOptions],
) error {
	ctx = InitContext(ctx, txn)

	ctx, unlock := lockForTxn(ctx, txn)
	defer unlock()

	if txn.isClosed {
		return ErrTxnDiscarded
	}

	return txn.db.DeleteP2PCollections(ctx, collectionNames, opts...)
}

func (txn *Txn) ListP2PCollections(
	ctx context.Context,
	opts ...options.Enumerable[options.ListP2PCollectionsOptions],
) ([]string, error) {
	ctx = InitContext(ctx, txn)

	ctx, unlock := lockForTxn(ctx, txn)
	defer unlock()

	if txn.isClosed {
		return nil, ErrTxnDiscarded
	}

	return txn.db.ListP2PCollections(ctx, opts...)
}

func (txn *Txn) AddP2PDocuments(
	ctx context.Context,
	docIDs []string,
	opts ...options.Enumerable[options.AddP2PDocumentsOptions],
) error {
	ctx = InitContext(ctx, txn)

	ctx, unlock := lockForTxn(ctx, txn)
	defer unlock()

	if txn.isClosed {
		return ErrTxnDiscarded
	}

	return txn.db.AddP2PDocuments(ctx, docIDs, opts...)
}

func (txn *Txn) DeleteP2PDocuments(
	ctx context.Context,
	docIDs []string,
	opts ...options.Enumerable[options.DeleteP2PDocumentsOptions],
) error {
	ctx = InitContext(ctx, txn)

	ctx, unlock := lockForTxn(ctx, txn)
	defer unlock()

	if txn.isClosed {
		return ErrTxnDiscarded
	}

	return txn.db.DeleteP2PDocuments(ctx, docIDs, opts...)
}

func (txn *Txn) ListP2PDocuments(
	ctx context.Context,
	opts ...options.Enumerable[options.ListP2PDocumentsOptions],
) ([]string, error) {
	ctx = InitContext(ctx, txn)

	ctx, unlock := lockForTxn(ctx, txn)
	defer unlock()

	if txn.isClosed {
		return nil, ErrTxnDiscarded
	}

	return txn.db.ListP2PDocuments(ctx, opts...)
}

func (txn *Txn) SyncDocuments(
	ctx context.Context,
	collectionName string,
	docIDs []string,
	opts ...options.Enumerable[options.SyncDocumentsOptions],
) error {
	ctx = InitContext(ctx, txn)

	ctx, unlock := lockForTxn(ctx, txn)
	defer unlock()

	if txn.isClosed {
		return ErrTxnDiscarded
	}

	return txn.db.SyncDocuments(ctx, collectionName, docIDs, opts...)
}

func (txn *Txn) SyncCollectionVersions(
	ctx context.Context,
	versionIDs []string,
	opts ...options.Enumerable[options.SyncCollectionVersionsOptions],
) error {
	ctx = InitContext(ctx, txn)

	ctx, unlock := lockForTxn(ctx, txn)
	defer unlock()

	if txn.isClosed {
		return ErrTxnDiscarded
	}

	return txn.db.SyncCollectionVersions(ctx, versionIDs, opts...)
}

func (txn *Txn) SyncBranchableCollection(
	ctx context.Context,
	collectionID string,
	opts ...options.Enumerable[options.SyncBranchableCollectionOptions],
) error {
	ctx = InitContext(ctx, txn)

	ctx, unlock := lockForTxn(ctx, txn)
	defer unlock()

	if txn.isClosed {
		return ErrTxnDiscarded
	}

	return txn.db.SyncBranchableCollection(ctx, collectionID, opts...)
}
