// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cli

import (
	"context"
	"fmt"

	"github.com/lens-vm/lens/host-go/config/model"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/http"
)

var _ client.Txn = (*Transaction)(nil)

type Transaction struct {
	*Wrapper
	tx client.Txn
}

func (txn *Transaction) ID() uint64 {
	return txn.tx.ID()
}

func (txn *Transaction) Commit(ctx context.Context) error {
	args := []string{"client", "tx", "commit"}
	args = append(args, fmt.Sprintf("%d", txn.tx.ID()))

	_, err := txn.cmd.execute(ctx, args)
	return err
}

func (txn *Transaction) Discard(ctx context.Context) {
	args := []string{"client", "tx", "discard"}
	args = append(args, fmt.Sprintf("%d", txn.tx.ID()))

	txn.cmd.execute(ctx, args) //nolint:errcheck
}

func (txn *Transaction) PrintDump(ctx context.Context) error {
	ctx = http.SetContextTxn(ctx, txn)
	return txn.Wrapper.PrintDump(ctx)
}

func (txn *Transaction) AddDACPolicy(ctx context.Context, policy string) (client.AddPolicyResult, error) {
	ctx = http.SetContextTxn(ctx, txn)
	return txn.Wrapper.AddDACPolicy(ctx, policy)
}

func (txn *Transaction) AddDACActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
) (client.AddActorRelationshipResult, error) {
	ctx = http.SetContextTxn(ctx, txn)
	return txn.Wrapper.AddDACActorRelationship(ctx, collectionName, docID, relation, targetActor)
}

func (txn *Transaction) DeleteDACActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
) (client.DeleteActorRelationshipResult, error) {
	ctx = http.SetContextTxn(ctx, txn)
	return txn.Wrapper.DeleteDACActorRelationship(ctx, collectionName, docID, relation, targetActor)
}

func (txn *Transaction) GetNodeIdentity(ctx context.Context) (immutable.Option[identity.PublicRawIdentity], error) {
	ctx = http.SetContextTxn(ctx, txn)
	return txn.Wrapper.GetNodeIdentity(ctx)
}

func (txn *Transaction) VerifySignature(ctx context.Context, blockCid string, pubKey crypto.PublicKey) error {
	ctx = http.SetContextTxn(ctx, txn)
	return txn.Wrapper.VerifySignature(ctx, blockCid, pubKey)
}

func (txn *Transaction) AddSchema(ctx context.Context, sdl string) ([]client.CollectionVersion, error) {
	ctx = http.SetContextTxn(ctx, txn)
	return txn.Wrapper.AddSchema(ctx, sdl)
}

func (txn *Transaction) PatchSchema(
	ctx context.Context,
	patch string,
	migration immutable.Option[model.Lens],
	setDefault bool,
) error {
	ctx = http.SetContextTxn(ctx, txn)
	return txn.Wrapper.PatchSchema(ctx, patch, migration, setDefault)
}

func (txn *Transaction) PatchCollection(ctx context.Context, patch string) error {
	ctx = http.SetContextTxn(ctx, txn)
	return txn.Wrapper.PatchCollection(ctx, patch)
}

func (txn *Transaction) SetActiveSchemaVersion(ctx context.Context, version string) error {
	ctx = http.SetContextTxn(ctx, txn)
	return txn.Wrapper.SetActiveSchemaVersion(ctx, version)
}

func (txn *Transaction) AddView(
	ctx context.Context,
	gqlQuery string,
	sdl string,
	transform immutable.Option[model.Lens],
) ([]client.CollectionDefinition, error) {
	ctx = http.SetContextTxn(ctx, txn)
	return txn.Wrapper.AddView(ctx, gqlQuery, sdl, transform)
}

func (txn *Transaction) RefreshViews(ctx context.Context, options client.CollectionFetchOptions) error {
	ctx = http.SetContextTxn(ctx, txn)
	return txn.Wrapper.RefreshViews(ctx, options)
}

func (txn *Transaction) SetMigration(ctx context.Context, config client.LensConfig) error {
	ctx = http.SetContextTxn(ctx, txn)
	return txn.Wrapper.SetMigration(ctx, config)
}

func (txn *Transaction) LensRegistry() client.LensRegistry {
	return txn.Wrapper.LensRegistry()
}

func (txn *Transaction) GetCollectionByName(
	ctx context.Context,
	name client.CollectionName,
) (client.Collection, error) {
	ctx = http.SetContextTxn(ctx, txn)
	return txn.Wrapper.GetCollectionByName(ctx, name)
}

func (txn *Transaction) GetCollections(
	ctx context.Context,
	options client.CollectionFetchOptions,
) ([]client.Collection, error) {
	ctx = http.SetContextTxn(ctx, txn)
	return txn.Wrapper.GetCollections(ctx, options)
}

func (txn *Transaction) GetSchemaByVersionID(ctx context.Context, versionID string) (client.SchemaDescription, error) {
	ctx = http.SetContextTxn(ctx, txn)
	return txn.Wrapper.GetSchemaByVersionID(ctx, versionID)
}

func (txn *Transaction) GetSchemas(
	ctx context.Context,
	options client.SchemaFetchOptions,
) ([]client.SchemaDescription, error) {
	ctx = http.SetContextTxn(ctx, txn)
	return txn.Wrapper.GetSchemas(ctx, options)
}

func (txn *Transaction) GetAllIndexes(
	ctx context.Context,
) (map[client.CollectionName][]client.IndexDescription, error) {
	ctx = http.SetContextTxn(ctx, txn)
	return txn.Wrapper.GetAllIndexes(ctx)
}

func (txn *Transaction) ExecRequest(
	ctx context.Context,
	request string,
	opts ...client.RequestOption,
) *client.RequestResult {
	ctx = http.SetContextTxn(ctx, txn)
	return txn.Wrapper.ExecRequest(ctx, request, opts...)
}

func (txn *Transaction) BasicImport(ctx context.Context, filepath string) error {
	ctx = http.SetContextTxn(ctx, txn)
	return txn.Wrapper.BasicImport(ctx, filepath)
}

func (txn *Transaction) BasicExport(ctx context.Context, config *client.BackupConfig) error {
	ctx = http.SetContextTxn(ctx, txn)
	return txn.Wrapper.BasicExport(ctx, config)
}
