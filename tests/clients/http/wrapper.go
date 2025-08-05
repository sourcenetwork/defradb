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

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
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
	node       *node.Node
	handler    *http.Handler
	client     *http.Client
	httpServer *httptest.Server
}

func NewWrapper(node *node.Node) (*Wrapper, error) {
	handler, err := http.NewHandler(node.DB, node.Peer)
	if err != nil {
		return nil, err
	}

	httpServer := httptest.NewServer(handler)
	client, err := http.NewClient(httpServer.URL)
	if err != nil {
		return nil, err
	}

	return &Wrapper{
		node,
		handler,
		client,
		httpServer,
	}, nil
}

func (w *Wrapper) PeerInfo() peer.AddrInfo {
	return w.client.PeerInfo()
}

func (w *Wrapper) SetReplicator(ctx context.Context, info peer.AddrInfo, collections ...string) error {
	return w.client.SetReplicator(ctx, info, collections...)
}

func (w *Wrapper) DeleteReplicator(ctx context.Context, info peer.AddrInfo, collections ...string) error {
	return w.client.DeleteReplicator(ctx, info, collections...)
}

func (w *Wrapper) GetAllReplicators(ctx context.Context) ([]client.Replicator, error) {
	return w.client.GetAllReplicators(ctx)
}

func (w *Wrapper) AddP2PCollections(ctx context.Context, collectionIDs ...string) error {
	return w.client.AddP2PCollections(ctx, collectionIDs...)
}

func (w *Wrapper) RemoveP2PCollections(ctx context.Context, collectionIDs ...string) error {
	return w.client.RemoveP2PCollections(ctx, collectionIDs...)
}

func (w *Wrapper) GetAllP2PCollections(ctx context.Context) ([]string, error) {
	return w.client.GetAllP2PCollections(ctx)
}

func (w *Wrapper) AddP2PDocuments(ctx context.Context, collectionIDs ...string) error {
	return w.client.AddP2PDocuments(ctx, collectionIDs...)
}

func (w *Wrapper) RemoveP2PDocuments(ctx context.Context, collectionIDs ...string) error {
	return w.client.RemoveP2PDocuments(ctx, collectionIDs...)
}

func (w *Wrapper) GetAllP2PDocuments(ctx context.Context) ([]string, error) {
	return w.client.GetAllP2PDocuments(ctx)
}

func (w *Wrapper) SyncDocuments(
	ctx context.Context,
	collectionName string,
	docIDs []string,
) error {
	return w.client.SyncDocuments(ctx, collectionName, docIDs)
}

func (w *Wrapper) BasicImport(ctx context.Context, filepath string) error {
	return w.client.BasicImport(ctx, filepath)
}

func (w *Wrapper) BasicExport(ctx context.Context, config *client.BackupConfig) error {
	return w.client.BasicExport(ctx, config)
}

func (w *Wrapper) AddSchema(ctx context.Context, schema string) ([]client.CollectionVersion, error) {
	return w.client.AddSchema(ctx, schema)
}

func (w *Wrapper) AddDACPolicy(
	ctx context.Context,
	policy string,
) (client.AddPolicyResult, error) {
	return w.client.AddDACPolicy(ctx, policy)
}

func (w *Wrapper) AddDACActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
) (client.AddActorRelationshipResult, error) {
	return w.client.AddDACActorRelationship(
		ctx,
		collectionName,
		docID,
		relation,
		targetActor,
	)
}

func (w *Wrapper) DeleteDACActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
) (client.DeleteActorRelationshipResult, error) {
	return w.client.DeleteDACActorRelationship(
		ctx,
		collectionName,
		docID,
		relation,
		targetActor,
	)
}

func (w *Wrapper) AddNACActorRelationship(
	ctx context.Context,
	relation string,
	targetActor string,
) (client.AddActorRelationshipResult, error) {
	return w.client.AddNACActorRelationship(
		ctx,
		relation,
		targetActor,
	)
}

func (w *Wrapper) DeleteNACActorRelationship(
	ctx context.Context,
	relation string,
	targetActor string,
) (client.DeleteActorRelationshipResult, error) {
	return w.client.DeleteNACActorRelationship(
		ctx,
		relation,
		targetActor,
	)
}

func (w *Wrapper) ReEnableNAC(ctx context.Context) error {
	return w.client.ReEnableNAC(ctx)
}

func (w *Wrapper) DisableNAC(ctx context.Context) error {
	return w.client.DisableNAC(ctx)
}

func (w *Wrapper) GetNACStatus(ctx context.Context) (client.NACStatusResult, error) {
	return w.client.GetNACStatus(ctx)
}

func (w *Wrapper) PatchSchema(
	ctx context.Context,
	patch string,
	migration immutable.Option[model.Lens],
	setAsDefaultVersion bool,
) error {
	return w.client.PatchSchema(ctx, patch, migration, setAsDefaultVersion)
}

func (w *Wrapper) PatchCollection(
	ctx context.Context,
	patch string,
) error {
	return w.client.PatchCollection(ctx, patch)
}

func (w *Wrapper) SetActiveSchemaVersion(ctx context.Context, schemaVersionID string) error {
	return w.client.SetActiveSchemaVersion(ctx, schemaVersionID)
}

func (w *Wrapper) AddView(
	ctx context.Context,
	query string,
	sdl string,
	transform immutable.Option[model.Lens],
) ([]client.CollectionDefinition, error) {
	return w.client.AddView(ctx, query, sdl, transform)
}

func (w *Wrapper) RefreshViews(ctx context.Context, opts client.CollectionFetchOptions) error {
	return w.client.RefreshViews(ctx, opts)
}

func (w *Wrapper) SetMigration(ctx context.Context, config client.LensConfig) error {
	return w.client.SetMigration(ctx, config)
}

func (w *Wrapper) LensRegistry() client.LensRegistry {
	return w.client.LensRegistry()
}

func (w *Wrapper) GetCollectionByName(ctx context.Context, name client.CollectionName) (client.Collection, error) {
	return w.client.GetCollectionByName(ctx, name)
}

func (w *Wrapper) GetCollections(
	ctx context.Context,
	options client.CollectionFetchOptions,
) ([]client.Collection, error) {
	return w.client.GetCollections(ctx, options)
}

func (w *Wrapper) GetSchemaByVersionID(ctx context.Context, versionID string) (client.SchemaDescription, error) {
	return w.client.GetSchemaByVersionID(ctx, versionID)
}

func (w *Wrapper) GetSchemas(
	ctx context.Context,
	options client.SchemaFetchOptions,
) ([]client.SchemaDescription, error) {
	return w.client.GetSchemas(ctx, options)
}

func (w *Wrapper) GetAllIndexes(ctx context.Context) (map[client.CollectionName][]client.IndexDescription, error) {
	return w.client.GetAllIndexes(ctx)
}

func (w *Wrapper) ExecRequest(
	ctx context.Context,
	query string,
	opts ...client.RequestOption,
) *client.RequestResult {
	return w.client.ExecRequest(ctx, query, opts...)
}

func (w *Wrapper) NewTxn(ctx context.Context, readOnly bool) (client.Txn, error) {
	clientTxn, err := w.client.NewTxn(ctx, readOnly)
	if err != nil {
		return nil, err
	}
	serverTxn, err := w.handler.Transaction(clientTxn.ID())
	if err != nil {
		return nil, err
	}
	return &Transaction{w, serverTxn}, nil
}

func (w *Wrapper) NewConcurrentTxn(ctx context.Context, readOnly bool) (client.Txn, error) {
	clientTxn, err := w.client.NewConcurrentTxn(ctx, readOnly)
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
	w.httpServer.CloseClientConnections()
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

func (w *Wrapper) Connect(ctx context.Context, addr peer.AddrInfo) error {
	return w.node.Peer.Connect(ctx, addr)
}

func (w *Wrapper) Host() string {
	return w.httpServer.URL
}

func (w *Wrapper) GetNodeIdentity(ctx context.Context) (immutable.Option[identity.PublicRawIdentity], error) {
	return w.client.GetNodeIdentity(ctx)
}

func (w *Wrapper) VerifySignature(ctx context.Context, cid string, pubKey crypto.PublicKey) error {
	return w.client.VerifySignature(ctx, cid, pubKey)
}
