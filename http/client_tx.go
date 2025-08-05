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
	"fmt"
	"net/http"

	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/internal/datastore"
)

var _ client.Txn = (*Transaction)(nil)

// Transaction implements the client.Txn interface over HTTP.
type Transaction struct {
	*Client
	id uint64
}

func NewTransaction(rawURL string, id uint64) (*Transaction, error) {
	httpClient, err := newHttpClient(rawURL)
	if err != nil {
		return nil, err
	}
	return &Transaction{&Client{httpClient}, id}, nil
}

func (txn *Transaction) ID() uint64 {
	return txn.id
}

func (txn *Transaction) Commit(ctx context.Context) error {
	methodURL := txn.http.apiURL.JoinPath("tx", fmt.Sprintf("%d", txn.id))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, methodURL.String(), nil)
	if err != nil {
		return err
	}
	_, err = txn.http.request(req)
	return err
}

func (txn *Transaction) Discard(ctx context.Context) {
	methodURL := txn.http.apiURL.JoinPath("tx", fmt.Sprintf("%d", txn.id))

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, methodURL.String(), nil)
	if err != nil {
		return
	}
	txn.http.request(req) //nolint:errcheck
}

func (txn *Transaction) PrintDump(ctx context.Context) error {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Client.PrintDump(ctx)
}

func (txn *Transaction) AddDACPolicy(ctx context.Context, policy string) (client.AddPolicyResult, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Client.AddDACPolicy(ctx, policy)
}

func (txn *Transaction) AddDACActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
) (client.AddActorRelationshipResult, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Client.AddDACActorRelationship(ctx, collectionName, docID, relation, targetActor)
}

func (txn *Transaction) DeleteDACActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
) (client.DeleteActorRelationshipResult, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Client.DeleteDACActorRelationship(ctx, collectionName, docID, relation, targetActor)
}

func (txn *Transaction) GetNodeIdentity(ctx context.Context) (immutable.Option[identity.PublicRawIdentity], error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Client.GetNodeIdentity(ctx)
}

func (txn *Transaction) VerifySignature(ctx context.Context, blockCid string, pubKey crypto.PublicKey) error {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Client.VerifySignature(ctx, blockCid, pubKey)
}

func (txn *Transaction) AddSchema(ctx context.Context, sdl string) ([]client.CollectionVersion, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Client.AddSchema(ctx, sdl)
}

func (txn *Transaction) PatchSchema(
	ctx context.Context,
	patch string,
	migration immutable.Option[model.Lens],
	setDefault bool,
) error {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Client.PatchSchema(ctx, patch, migration, setDefault)
}

func (txn *Transaction) PatchCollection(ctx context.Context, patch string) error {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Client.PatchCollection(ctx, patch)
}

func (txn *Transaction) SetActiveSchemaVersion(ctx context.Context, version string) error {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Client.SetActiveSchemaVersion(ctx, version)
}

func (txn *Transaction) AddView(
	ctx context.Context,
	gqlQuery string,
	sdl string,
	transform immutable.Option[model.Lens],
) ([]client.CollectionDefinition, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Client.AddView(ctx, gqlQuery, sdl, transform)
}

func (txn *Transaction) RefreshViews(ctx context.Context, options client.CollectionFetchOptions) error {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Client.RefreshViews(ctx, options)
}

func (txn *Transaction) SetMigration(ctx context.Context, config client.LensConfig) error {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Client.SetMigration(ctx, config)
}

func (txn *Transaction) LensRegistry() client.LensRegistry {
	return txn.Client.LensRegistry()
}

func (txn *Transaction) GetCollectionByName(
	ctx context.Context,
	name client.CollectionName,
) (client.Collection, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Client.GetCollectionByName(ctx, name)
}

func (txn *Transaction) GetCollections(
	ctx context.Context,
	options client.CollectionFetchOptions,
) ([]client.Collection, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Client.GetCollections(ctx, options)
}

func (txn *Transaction) GetSchemaByVersionID(ctx context.Context, versionID string) (client.SchemaDescription, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Client.GetSchemaByVersionID(ctx, versionID)
}

func (txn *Transaction) GetSchemas(
	ctx context.Context,
	options client.SchemaFetchOptions,
) ([]client.SchemaDescription, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Client.GetSchemas(ctx, options)
}

func (txn *Transaction) GetAllIndexes(
	ctx context.Context,
) (map[client.CollectionName][]client.IndexDescription, error) {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Client.GetAllIndexes(ctx)
}

func (txn *Transaction) ExecRequest(
	ctx context.Context,
	request string,
	opts ...client.RequestOption,
) *client.RequestResult {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Client.ExecRequest(ctx, request, opts...)
}

func (txn *Transaction) BasicImport(ctx context.Context, filepath string) error {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Client.BasicImport(ctx, filepath)
}

func (txn *Transaction) BasicExport(ctx context.Context, config *client.BackupConfig) error {
	ctx = datastore.CtxSetFromClientTxn(ctx, txn)
	return txn.Client.BasicExport(ctx, config)
}
