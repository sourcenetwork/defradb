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

	"github.com/lens-vm/lens/host-go/config/model"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/internal/datastore"
)

// transactionDB is a db that can create transactions.
type transactionDB interface {
	NewTxn(context.Context, bool) (client.Txn, error)
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
		txn := ctxTxn.(*Txn) //nolint:forcetypeassert
		if txn.explicit {
			// if it's already an explicit txn we return it as is.
			return InitContext(ctx, txn), txn, nil
		}
		// If the txn has already been set on the context buty it hasn't already been set as explicit,
		// we create a copy of the txn and mark it as an explicit txn.
		explicitTxn := &Txn{
			txn.BasicTxn,
			txn.db,
			true,
		}
		return InitContext(ctx, explicitTxn), explicitTxn, nil

	}
	clientTxn, err := db.NewTxn(ctx, readOnly)
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

func (txn *Txn) Commit(ctx context.Context) error {
	if txn.explicit {
		return nil
	}
	return txn.BasicTxn.Commit(ctx)
}

func (txn *Txn) Discard(ctx context.Context) {
	if txn.explicit {
		return
	}
	txn.BasicTxn.Discard(ctx)
}

func (txn *Txn) PrintDump(ctx context.Context) error {
	return printStore(ctx, txn.Rootstore())
}

func (txn *Txn) AddDACPolicy(ctx context.Context, policy string) (client.AddPolicyResult, error) {
	ctx = InitContext(ctx, txn)
	return txn.db.AddDACPolicy(ctx, policy)
}

func (txn *Txn) AddDACActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
) (client.AddActorRelationshipResult, error) {
	ctx = InitContext(ctx, txn)
	return txn.db.AddDACActorRelationship(ctx, collectionName, docID, relation, targetActor)
}

func (txn *Txn) DeleteDACActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
) (client.DeleteActorRelationshipResult, error) {
	ctx = InitContext(ctx, txn)
	return txn.db.DeleteDACActorRelationship(ctx, collectionName, docID, relation, targetActor)
}

func (txn *Txn) GetNodeIdentity(ctx context.Context) (immutable.Option[identity.PublicRawIdentity], error) {
	ctx = InitContext(ctx, txn)
	return txn.db.GetNodeIdentity(ctx)
}

func (txn *Txn) VerifySignature(ctx context.Context, blockCid string, pubKey crypto.PublicKey) error {
	ctx = InitContext(ctx, txn)
	return txn.db.VerifySignature(ctx, blockCid, pubKey)
}

func (txn *Txn) AddSchema(ctx context.Context, sdl string) ([]client.CollectionVersion, error) {
	ctx = InitContext(ctx, txn)
	return txn.db.AddSchema(ctx, sdl)
}

func (txn *Txn) PatchSchema(
	ctx context.Context,
	patch string,
	migration immutable.Option[model.Lens],
	setDefault bool,
) error {
	ctx = InitContext(ctx, txn)
	return txn.db.PatchSchema(ctx, patch, migration, setDefault)
}

func (txn *Txn) PatchCollection(ctx context.Context, patch string) error {
	ctx = InitContext(ctx, txn)
	return txn.db.PatchCollection(ctx, patch)
}

func (txn *Txn) SetActiveSchemaVersion(ctx context.Context, version string) error {
	ctx = InitContext(ctx, txn)
	return txn.db.SetActiveSchemaVersion(ctx, version)
}

func (txn *Txn) AddView(
	ctx context.Context,
	gqlQuery string,
	sdl string,
	transform immutable.Option[model.Lens],
) ([]client.CollectionDefinition, error) {
	ctx = InitContext(ctx, txn)
	return txn.db.AddView(ctx, gqlQuery, sdl, transform)
}

func (txn *Txn) RefreshViews(ctx context.Context, options client.CollectionFetchOptions) error {
	ctx = InitContext(ctx, txn)
	return txn.db.RefreshViews(ctx, options)
}

func (txn *Txn) SetMigration(ctx context.Context, config client.LensConfig) error {
	ctx = InitContext(ctx, txn)
	return txn.db.SetMigration(ctx, config)
}

func (txn *Txn) LensRegistry() client.LensRegistry {
	return txn.db.LensRegistry()
}

func (txn *Txn) GetCollectionByName(ctx context.Context, name client.CollectionName) (client.Collection, error) {
	ctx = InitContext(ctx, txn)
	return txn.db.GetCollectionByName(ctx, name)
}

func (txn *Txn) GetCollections(
	ctx context.Context,
	options client.CollectionFetchOptions,
) ([]client.Collection, error) {
	ctx = InitContext(ctx, txn)
	return txn.db.GetCollections(ctx, options)
}

func (txn *Txn) GetSchemaByVersionID(ctx context.Context, versionID string) (client.SchemaDescription, error) {
	ctx = InitContext(ctx, txn)
	return txn.db.GetSchemaByVersionID(ctx, versionID)
}

func (txn *Txn) GetSchemas(ctx context.Context, options client.SchemaFetchOptions) ([]client.SchemaDescription, error) {
	ctx = InitContext(ctx, txn)
	return txn.db.GetSchemas(ctx, options)
}

func (txn *Txn) GetAllIndexes(ctx context.Context) (map[client.CollectionName][]client.IndexDescription, error) {
	ctx = InitContext(ctx, txn)
	return txn.db.GetAllIndexes(ctx)
}

func (txn *Txn) ExecRequest(ctx context.Context, request string, opts ...client.RequestOption) *client.RequestResult {
	ctx = InitContext(ctx, txn)
	return txn.db.ExecRequest(ctx, request, opts...)
}

func (txn *Txn) BasicImport(ctx context.Context, filepath string) error {
	ctx = InitContext(ctx, txn)
	return txn.db.BasicImport(ctx, filepath)
}

func (txn *Txn) BasicExport(ctx context.Context, config *client.BackupConfig) error {
	ctx = InitContext(ctx, txn)
	return txn.db.BasicExport(ctx, config)
}
