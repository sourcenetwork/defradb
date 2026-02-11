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

	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/internal/datastore"
)

// transactionDB is a db that can create transactions.
type transactionDB interface {
	NewTxn(bool) (client.Txn, error)
}

// ensureContextTxn ensures that the returned context has a transaction.
//
// If a transactions exists on the context it will be made explicit,
// otherwise a new implicit transaction will be created.
//
// The returned context will contain the transaction
// along with the copied values from the input context.
func ensureContextTxn(ctx context.Context, db transactionDB, readOnly bool) (context.Context, datastore.Txn, error) {
	// explicit transaction
	ctxTxn, ok := datastore.CtxTryGetTxn(ctx)
	if ok {
		switch txn := ctxTxn.(type) {
		case *Txn:
			if txn.explicit {
				// if it's already an explicit txn we return it as is.
				return InitContext(ctx, txn), txn, nil
			}
			// If the txn has already been set on the context but it hasn't already been set as explicit,
			// we create a copy of the txn and mark it as an explicit txn.
			explicitTxn := &Txn{
				txn.BasicTxn,
				txn.db,
				true,
			}
			return InitContext(ctx, explicitTxn), explicitTxn, nil
		case *datastore.BasicTxn:
			// There are scenarios where the transaction passed to the `db` methods was created
			// from a separate package (ex: `net`). In that situation the type of transaction passed in
			// will most likely be of type `*datastore.Txn`. We can wrap it in a `*Txn` and mark it as explicit.
			//
			// WARNING: This scenario creates a transaction where `*DB` is nil. Calling any method that requires this
			// will result in a panic.
			explicitTxn := &Txn{
				txn,
				nil,
				true,
			}
			return InitContext(ctx, explicitTxn), explicitTxn, nil
		default:
			return nil, nil, NewErrUnsupportedTxnType(ctxTxn)
		}
	}
	clientTxn, err := db.NewTxn(readOnly)
	if err != nil {
		return nil, nil, err
	}
	txn := clientTxn.(*Txn) //nolint:forcetypeassert
	return InitContext(ctx, txn), txn, nil
}

type Txn struct {
	*datastore.BasicTxn
	db       *DB
	explicit bool
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
	return txn.BasicTxn.Commit()
}

func (txn *Txn) Discard() {
	if txn.explicit {
		// If the transaction has been explicitly defined, `Discard` should
		// only be executed by the transaction creator. As such, a call to
		// `Discard` on an explicit transaction should result in a no-op.
		return
	}
	txn.BasicTxn.Discard()
}

func (txn *Txn) PrintDump(ctx context.Context) error {
	return printStore(ctx, txn.Rootstore())
}

func (txn *Txn) AddDACPolicy(
	ctx context.Context,
	policy string,
	opts ...options.Lister[options.AddDACPolicyOptions],
) (client.AddPolicyResult, error) {
	ctx = InitContext(ctx, txn)
	return txn.db.AddDACPolicy(ctx, policy, opts...)
}

func (txn *Txn) AddDACActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
	opts ...options.Lister[options.AddDACActorRelationshipOptions],
) (client.AddActorRelationshipResult, error) {
	ctx = InitContext(ctx, txn)
	return txn.db.AddDACActorRelationship(ctx, collectionName, docID, relation, targetActor, opts...)
}

func (txn *Txn) DeleteDACActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
	opts ...options.Lister[options.DeleteDACActorRelationshipOptions],
) (client.DeleteActorRelationshipResult, error) {
	ctx = InitContext(ctx, txn)
	return txn.db.DeleteDACActorRelationship(ctx, collectionName, docID, relation, targetActor, opts...)
}

func (txn *Txn) AddNACActorRelationship(
	ctx context.Context,
	relation string,
	targetActor string,
	opts ...options.Lister[options.AddNACActorRelationshipOptions],
) (client.AddActorRelationshipResult, error) {
	ctx = InitContext(ctx, txn)
	return txn.db.AddNACActorRelationship(ctx, relation, targetActor, opts...)
}

func (txn *Txn) DeleteNACActorRelationship(
	ctx context.Context,
	relation string,
	targetActor string,
	opts ...options.Lister[options.DeleteNACActorRelationshipOptions],
) (client.DeleteActorRelationshipResult, error) {
	ctx = InitContext(ctx, txn)
	return txn.db.DeleteNACActorRelationship(ctx, relation, targetActor, opts...)
}

func (txn *Txn) ReEnableNAC(ctx context.Context, opts ...options.Lister[options.ReEnableNACOptions]) error {
	ctx = InitContext(ctx, txn)
	return txn.db.ReEnableNAC(ctx, opts...)
}

func (txn *Txn) DisableNAC(ctx context.Context, opts ...options.Lister[options.DisableNACOptions]) error {
	ctx = InitContext(ctx, txn)
	return txn.db.DisableNAC(ctx, opts...)
}

func (txn *Txn) GetNACStatus(
	ctx context.Context,
	opts ...options.Lister[options.GetNACStatusOptions],
) (client.NACStatusResult, error) {
	ctx = InitContext(ctx, txn)
	return txn.db.GetNACStatus(ctx, opts...)
}

func (txn *Txn) GetNodeIdentity(ctx context.Context) (immutable.Option[identity.PublicRawIdentity], error) {
	ctx = InitContext(ctx, txn)
	return txn.db.GetNodeIdentity(ctx)
}

func (txn *Txn) VerifySignature(
	ctx context.Context,
	blockCid string,
	pubKey crypto.PublicKey,
	opts ...options.Lister[options.VerifySignatureOptions],
) error {
	ctx = InitContext(ctx, txn)
	return txn.db.VerifySignature(ctx, blockCid, pubKey, opts...)
}

func (txn *Txn) AddSchema(
	ctx context.Context,
	sdl string,
	opts ...options.Lister[options.AddSchemaOptions],
) ([]client.CollectionVersion, error) {
	ctx = InitContext(ctx, txn)
	return txn.db.AddSchema(ctx, sdl, opts...)
}

func (txn *Txn) PatchCollection(
	ctx context.Context,
	patch string,
	migration immutable.Option[model.Lens],
	opts ...options.Lister[options.PatchCollectionOptions],
) error {
	ctx = InitContext(ctx, txn)
	return txn.db.PatchCollection(ctx, patch, migration, opts...)
}

func (txn *Txn) SetActiveCollectionVersion(
	ctx context.Context,
	version string,
	opts ...options.Lister[options.SetActiveCollectionVersionOptions],
) error {
	ctx = InitContext(ctx, txn)
	return txn.db.SetActiveCollectionVersion(ctx, version, opts...)
}

func (txn *Txn) AddView(
	ctx context.Context,
	gqlQuery string,
	sdl string,
	opts ...options.Lister[options.AddViewOptions],
) ([]client.CollectionVersion, error) {
	ctx = InitContext(ctx, txn)
	return txn.db.AddView(ctx, gqlQuery, sdl, opts...)
}

func (txn *Txn) RefreshViews(ctx context.Context, opts ...options.Lister[options.RefreshViewsOptions]) error {
	ctx = InitContext(ctx, txn)
	return txn.db.RefreshViews(ctx, opts...)
}

func (txn *Txn) SetMigration(ctx context.Context, config client.LensConfig) (string, error) {
	ctx = InitContext(ctx, txn)
	return txn.db.SetMigration(ctx, config)
}

func (txn *Txn) AddLens(
	ctx context.Context,
	lens model.Lens,
	opts ...options.Lister[options.AddLensOptions],
) (string, error) {
	ctx = InitContext(ctx, txn)
	return txn.db.AddLens(ctx, lens, opts...)
}

func (txn *Txn) ListLenses(
	ctx context.Context,
	opts ...options.Lister[options.ListLensesOptions],
) (map[string]model.Lens, error) {
	ctx = InitContext(ctx, txn)
	return txn.db.ListLenses(ctx)
}

func (txn *Txn) GetCollectionByName(
	ctx context.Context,
	name client.CollectionName,
	opts ...options.Lister[options.GetCollectionByNameOptions],
) (client.Collection, error) {
	ctx = InitContext(ctx, txn)
	return txn.db.GetCollectionByName(ctx, name, opts...)
}

func (txn *Txn) GetCollections(
	ctx context.Context,
	opts ...options.Lister[options.GetCollectionsOptions],
) ([]client.Collection, error) {
	ctx = InitContext(ctx, txn)
	return txn.db.GetCollections(ctx, opts...)
}

func (txn *Txn) GetAllIndexes(
	ctx context.Context,
	opts ...options.Lister[options.GetAllIndexesOptions],
) (map[client.CollectionName][]client.IndexDescription, error) {
	ctx = InitContext(ctx, txn)
	return txn.db.GetAllIndexes(ctx, opts...)
}

func (txn *Txn) ListAllEncryptedIndexes(
	ctx context.Context,
) (map[client.CollectionName][]client.EncryptedIndexDescription, error) {
	ctx = InitContext(ctx, txn)
	return txn.db.ListAllEncryptedIndexes(ctx)
}

func (txn *Txn) ExecRequest(
	ctx context.Context,
	request string,
	opts ...options.Lister[options.ExecRequestOptions],
) *client.RequestResult {
	ctx = InitContext(ctx, txn)
	return txn.db.ExecRequest(ctx, request, opts...)
}

func (txn *Txn) BasicImport(ctx context.Context, filepath string) error {
	ctx = InitContext(ctx, txn)
	return txn.db.BasicImport(ctx, filepath)
}

func (txn *Txn) BasicExport(
	ctx context.Context,
	filepath string,
	opts ...options.Lister[options.BasicExportOptions],
) error {
	ctx = InitContext(ctx, txn)
	return txn.db.BasicExport(ctx, filepath, opts...)
}

func (txn *Txn) PeerInfo(ctx context.Context, opts ...options.Lister[options.PeerInfoOptions]) ([]string, error) {
	return txn.db.PeerInfo(ctx, opts...)
}

func (txn *Txn) ActivePeers(ctx context.Context, opts ...options.Lister[options.ActivePeersOptions]) ([]string, error) {
	return txn.db.ActivePeers(ctx, opts...)
}

func (txn *Txn) Connect(ctx context.Context, addresses []string, opts ...options.Lister[options.ConnectOptions]) error {
	return txn.db.Connect(ctx, addresses, opts...)
}

func (txn *Txn) CreateReplicator(
	ctx context.Context,
	addresses []string,
	opts ...options.Lister[options.CreateReplicatorOptions],
) error {
	ctx = InitContext(ctx, txn)
	return txn.db.CreateReplicator(ctx, addresses, opts...)
}

func (txn *Txn) DeleteReplicator(
	ctx context.Context,
	id string,
	opts ...options.Lister[options.DeleteReplicatorOptions],
) error {
	ctx = InitContext(ctx, txn)
	return txn.db.DeleteReplicator(ctx, id, opts...)
}

func (txn *Txn) ListReplicators(
	ctx context.Context,
	opts ...options.Lister[options.ListReplicatorsOptions],
) ([]client.Replicator, error) {
	ctx = InitContext(ctx, txn)
	return txn.db.ListReplicators(ctx, opts...)
}

func (txn *Txn) CreateP2PCollections(
	ctx context.Context,
	collectionNames []string,
	opts ...options.Lister[options.CreateP2PCollectionsOptions],
) error {
	ctx = InitContext(ctx, txn)
	return txn.db.CreateP2PCollections(ctx, collectionNames, opts...)
}

func (txn *Txn) DeleteP2PCollections(
	ctx context.Context,
	collectionNames []string,
	opts ...options.Lister[options.DeleteP2PCollectionsOptions],
) error {
	ctx = InitContext(ctx, txn)
	return txn.db.DeleteP2PCollections(ctx, collectionNames, opts...)
}

func (txn *Txn) ListP2PCollections(
	ctx context.Context,
	opts ...options.Lister[options.ListP2PCollectionsOptions],
) ([]string, error) {
	ctx = InitContext(ctx, txn)
	return txn.db.ListP2PCollections(ctx, opts...)
}

func (txn *Txn) CreateP2PDocuments(
	ctx context.Context,
	docIDs []string,
	opts ...options.Lister[options.CreateP2PDocumentsOptions],
) error {
	ctx = InitContext(ctx, txn)
	return txn.db.CreateP2PDocuments(ctx, docIDs, opts...)
}

func (txn *Txn) DeleteP2PDocuments(
	ctx context.Context,
	docIDs []string,
	opts ...options.Lister[options.DeleteP2PDocumentsOptions],
) error {
	ctx = InitContext(ctx, txn)
	return txn.db.DeleteP2PDocuments(ctx, docIDs, opts...)
}

func (txn *Txn) ListP2PDocuments(
	ctx context.Context,
	opts ...options.Lister[options.ListP2PDocumentsOptions],
) ([]string, error) {
	ctx = InitContext(ctx, txn)
	return txn.db.ListP2PDocuments(ctx, opts...)
}

func (txn *Txn) SyncDocuments(ctx context.Context, collectionName string, docIDs []string) error {
	ctx = InitContext(ctx, txn)
	return txn.db.SyncDocuments(ctx, collectionName, docIDs)
}

func (txn *Txn) SyncCollectionVersions(ctx context.Context, versionIDs ...string) error {
	ctx = InitContext(ctx, txn)
	return txn.db.SyncCollectionVersions(ctx, versionIDs...)
}

func (txn *Txn) SyncBranchableCollection(ctx context.Context, collectionID string) error {
	ctx = InitContext(ctx, txn)
	return txn.db.SyncBranchableCollection(ctx, collectionID)
}
