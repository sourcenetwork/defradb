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
	"net/http/httptest"

	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/http"
	"github.com/sourcenetwork/defradb/node"
)

var _ client.TxnStore = (*Wrapper)(nil)
var _ client.P2P = (*Wrapper)(nil)

// Wrapper combines an HTTP client and server into a
// single struct that implements the client.TxnStore interface.
type Wrapper struct {
	node         *node.Node
	handler      *http.Handler
	client       *http.Client
	httpServer   *httptest.Server
	serverCancel context.CancelFunc
}

func NewWrapper(node *node.Node) (*Wrapper, error) {
	handler, err := http.NewHandler(node.DB)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	handlerWithCtx := http.InjectServerContext(ctx)(handler)
	httpServer := httptest.NewServer(handlerWithCtx)
	client, err := http.NewClient(httpServer.URL)
	if err != nil {
		cancel()
		return nil, err
	}

	return &Wrapper{
		node,
		handler,
		client,
		httpServer,
		cancel,
	}, nil
}

func (w *Wrapper) PeerInfo(ctx context.Context, opts ...options.Enumerable[options.PeerInfoOptions]) ([]string, error) {
	return w.client.PeerInfo(ctx, opts...)
}

func (w *Wrapper) ActivePeers(
	ctx context.Context,
	opts ...options.Enumerable[options.ActivePeersOptions],
) ([]string, error) {
	return w.client.ActivePeers(ctx, opts...)
}

func (w *Wrapper) Connect(
	ctx context.Context,
	addresses []string,
	opts ...options.Enumerable[options.ConnectOptions],
) error {
	return w.client.Connect(ctx, addresses, opts...)
}

func (w *Wrapper) AddReplicator(
	ctx context.Context,
	addresses []string,
	opts ...options.Enumerable[options.AddReplicatorOptions],
) error {
	return w.client.AddReplicator(ctx, addresses, opts...)
}

func (w *Wrapper) DeleteReplicator(
	ctx context.Context,
	id string,
	opts ...options.Enumerable[options.DeleteReplicatorOptions],
) error {
	return w.client.DeleteReplicator(ctx, id, opts...)
}

func (w *Wrapper) ListReplicators(
	ctx context.Context,
	opts ...options.Enumerable[options.ListReplicatorsOptions],
) ([]client.Replicator, error) {
	return w.client.ListReplicators(ctx, opts...)
}

func (w *Wrapper) AddP2PCollections(
	ctx context.Context,
	collectionIDs []string,
	opts ...options.Enumerable[options.AddP2PCollectionsOptions],
) error {
	return w.client.AddP2PCollections(ctx, collectionIDs, opts...)
}

func (w *Wrapper) DeleteP2PCollections(
	ctx context.Context,
	collectionIDs []string,
	opts ...options.Enumerable[options.DeleteP2PCollectionsOptions],
) error {
	return w.client.DeleteP2PCollections(ctx, collectionIDs, opts...)
}

func (w *Wrapper) ListP2PCollections(
	ctx context.Context,
	opts ...options.Enumerable[options.ListP2PCollectionsOptions],
) ([]string, error) {
	return w.client.ListP2PCollections(ctx, opts...)
}

func (w *Wrapper) AddP2PDocuments(
	ctx context.Context,
	docIDs []string,
	opts ...options.Enumerable[options.AddP2PDocumentsOptions],
) error {
	return w.client.AddP2PDocuments(ctx, docIDs, opts...)
}

func (w *Wrapper) DeleteP2PDocuments(
	ctx context.Context,
	docIDs []string,
	opts ...options.Enumerable[options.DeleteP2PDocumentsOptions],
) error {
	return w.client.DeleteP2PDocuments(ctx, docIDs, opts...)
}

func (w *Wrapper) ListP2PDocuments(
	ctx context.Context,
	opts ...options.Enumerable[options.ListP2PDocumentsOptions],
) ([]string, error) {
	return w.client.ListP2PDocuments(ctx, opts...)
}

func (w *Wrapper) SyncDocuments(
	ctx context.Context,
	collectionName string,
	docIDs []string,
) error {
	return w.client.SyncDocuments(ctx, collectionName, docIDs)
}

func (w *Wrapper) SyncCollectionVersions(
	ctx context.Context,
	versionIDs []string,
	opts ...options.Enumerable[options.SyncCollectionVersionsOptions],
) error {
	return w.client.SyncCollectionVersions(ctx, versionIDs, opts...)
}

func (w *Wrapper) SyncBranchableCollection(
	ctx context.Context,
	collectionID string,
	opts ...options.Enumerable[options.SyncBranchableCollectionOptions]) error {
	return w.client.SyncBranchableCollection(ctx, collectionID, opts...)
}

func (w *Wrapper) BasicImport(ctx context.Context, filepath string) error {
	return w.client.BasicImport(ctx, filepath)
}

func (w *Wrapper) BasicExport(
	ctx context.Context,
	filepath string,
	opts ...options.Enumerable[options.BasicExportOptions],
) error {
	return w.client.BasicExport(ctx, filepath, opts...)
}

func (w *Wrapper) AddSchema(
	ctx context.Context,
	schema string,
	opts ...options.Enumerable[options.AddSchemaOptions],
) ([]client.CollectionVersion, error) {
	return w.client.AddSchema(ctx, schema, opts...)
}

func (w *Wrapper) AddDACPolicy(
	ctx context.Context,
	policy string,
	opts ...options.Enumerable[options.AddDACPolicyOptions],
) (client.AddPolicyResult, error) {
	return w.client.AddDACPolicy(ctx, policy, opts...)
}

func (w *Wrapper) AddDACActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
	opts ...options.Enumerable[options.AddDACActorRelationshipOptions],
) (client.AddActorRelationshipResult, error) {
	return w.client.AddDACActorRelationship(
		ctx,
		collectionName,
		docID,
		relation,
		targetActor,
		opts...,
	)
}

func (w *Wrapper) DeleteDACActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
	opts ...options.Enumerable[options.DeleteDACActorRelationshipOptions],
) (client.DeleteActorRelationshipResult, error) {
	return w.client.DeleteDACActorRelationship(
		ctx,
		collectionName,
		docID,
		relation,
		targetActor,
		opts...,
	)
}

func (w *Wrapper) AddNACActorRelationship(
	ctx context.Context,
	relation string,
	targetActor string,
	opts ...options.Enumerable[options.AddNACActorRelationshipOptions],
) (client.AddActorRelationshipResult, error) {
	return w.client.AddNACActorRelationship(
		ctx,
		relation,
		targetActor,
		opts...,
	)
}

func (w *Wrapper) DeleteNACActorRelationship(
	ctx context.Context,
	relation string,
	targetActor string,
	opts ...options.Enumerable[options.DeleteNACActorRelationshipOptions],
) (client.DeleteActorRelationshipResult, error) {
	return w.client.DeleteNACActorRelationship(
		ctx,
		relation,
		targetActor,
		opts...,
	)
}

func (w *Wrapper) ReEnableNAC(ctx context.Context, opts ...options.Enumerable[options.ReEnableNACOptions]) error {
	return w.client.ReEnableNAC(ctx, opts...)
}

func (w *Wrapper) DisableNAC(ctx context.Context, opts ...options.Enumerable[options.DisableNACOptions]) error {
	return w.client.DisableNAC(ctx, opts...)
}

func (w *Wrapper) GetNACStatus(
	ctx context.Context,
	opts ...options.Enumerable[options.GetNACStatusOptions],
) (client.NACStatusResult, error) {
	return w.client.GetNACStatus(ctx, opts...)
}

func (w *Wrapper) PatchCollection(
	ctx context.Context,
	patch string,
	migration immutable.Option[model.Lens],
	opts ...options.Enumerable[options.PatchCollectionOptions],
) error {
	return w.client.PatchCollection(ctx, patch, migration, opts...)
}

func (w *Wrapper) SetActiveCollectionVersion(
	ctx context.Context,
	collectionVersionID string,
	opts ...options.Enumerable[options.SetActiveCollectionVersionOptions],
) error {
	return w.client.SetActiveCollectionVersion(ctx, collectionVersionID, opts...)
}

func (w *Wrapper) AddView(
	ctx context.Context,
	query string,
	sdl string,
	opts ...options.Enumerable[options.AddViewOptions],
) ([]client.CollectionVersion, error) {
	return w.client.AddView(ctx, query, sdl, opts...)
}

func (w *Wrapper) RefreshViews(ctx context.Context, opts ...options.Enumerable[options.RefreshViewsOptions]) error {
	return w.client.RefreshViews(ctx, opts...)
}

func (w *Wrapper) SetMigration(
	ctx context.Context, config client.LensConfig, opts ...options.Enumerable[options.SetMigrationOptions],
) (string, error) {
	return w.client.SetMigration(ctx, config, opts...)
}

func (w *Wrapper) AddLens(
	ctx context.Context,
	lens model.Lens,
	opts ...options.Enumerable[options.AddLensOptions],
) (string, error) {
	return w.client.AddLens(ctx, lens, opts...)
}

func (w *Wrapper) ListLenses(
	ctx context.Context,
	opts ...options.Enumerable[options.ListLensesOptions],
) (map[string]model.Lens, error) {
	return w.client.ListLenses(ctx, opts...)
}

func (w *Wrapper) GetCollectionByName(
	ctx context.Context,
	name client.CollectionName,
	opts ...options.Enumerable[options.GetCollectionByNameOptions],
) (client.Collection, error) {
	return w.client.GetCollectionByName(ctx, name, opts...)
}

func (w *Wrapper) GetCollections(
	ctx context.Context,
	opts ...options.Enumerable[options.GetCollectionsOptions],
) ([]client.Collection, error) {
	return w.client.GetCollections(ctx, opts...)
}

func (w *Wrapper) GetAllIndexes(
	ctx context.Context,
	opts ...options.Enumerable[options.GetAllIndexesOptions],
) (map[client.CollectionName][]client.IndexDescription, error) {
	return w.client.GetAllIndexes(ctx, opts...)
}

func (w *Wrapper) ListAllEncryptedIndexes(
	ctx context.Context,
	opts ...options.Enumerable[options.ListAllEncryptedIndexesOptions],
) (map[client.CollectionName][]client.EncryptedIndexDescription, error) {
	return w.client.ListAllEncryptedIndexes(ctx, opts...)
}

func (w *Wrapper) ExecRequest(
	ctx context.Context,
	query string,
	opts ...options.Enumerable[options.ExecRequestOptions],
) *client.RequestResult {
	return w.client.ExecRequest(ctx, query, opts...)
}

func (w *Wrapper) NewTxn(readOnly bool) (client.Txn, error) {
	clientTxn, err := w.client.NewTxn(readOnly)
	if err != nil {
		return nil, err
	}
	serverTxn, err := w.handler.Transaction(clientTxn.ID())
	if err != nil {
		return nil, err
	}
	return &Transaction{w, serverTxn}, nil
}

func (w *Wrapper) NewConcurrentTxn(readOnly bool) (client.Txn, error) {
	clientTxn, err := w.client.NewConcurrentTxn(readOnly)
	if err != nil {
		return nil, err
	}
	serverTxn, err := w.handler.Transaction(clientTxn.ID())
	if err != nil {
		return nil, err
	}
	return &Transaction{w, serverTxn}, nil
}

func (w *Wrapper) Close() {
	w.serverCancel()
	w.httpServer.Close()
	_ = w.node.Close(context.Background())
}

func (w *Wrapper) Events() event.Bus {
	return w.node.DB.Events()
}

func (w *Wrapper) MaxTxnRetries() int {
	return w.node.DB.MaxTxnRetries()
}

func (w *Wrapper) PrintDump(ctx context.Context) error {
	return w.node.DB.PrintDump(ctx)
}

func (w *Wrapper) Host() string {
	return w.httpServer.URL
}

func (w *Wrapper) GetNodeIdentity(ctx context.Context) (immutable.Option[identity.PublicRawIdentity], error) {
	return w.client.GetNodeIdentity(ctx)
}

func (w *Wrapper) VerifySignature(
	ctx context.Context,
	cid string,
	pubKey crypto.PublicKey,
	opts ...options.Enumerable[options.VerifySignatureOptions],
) error {
	return w.client.VerifySignature(ctx, cid, pubKey, opts...)
}
