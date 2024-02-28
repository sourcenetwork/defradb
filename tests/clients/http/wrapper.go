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

	blockstore "github.com/ipfs/boxo/blockstore"
	"github.com/lens-vm/lens/host-go/config/model"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/events"
	"github.com/sourcenetwork/defradb/http"
	"github.com/sourcenetwork/defradb/net"
)

var _ client.P2P = (*Wrapper)(nil)

// Wrapper combines an HTTP client and server into a
// single struct that implements the client.DB interface.
type Wrapper struct {
	node       *net.Node
	handler    *http.Handler
	client     *http.Client
	httpServer *httptest.Server
}

func NewWrapper(node *net.Node) (*Wrapper, error) {
	handler, err := http.NewHandler(node)
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

func (w *Wrapper) SetReplicator(ctx context.Context, rep client.Replicator) error {
	return w.client.SetReplicator(ctx, rep)
}

func (w *Wrapper) DeleteReplicator(ctx context.Context, rep client.Replicator) error {
	return w.client.DeleteReplicator(ctx, rep)
}

func (w *Wrapper) GetAllReplicators(ctx context.Context) ([]client.Replicator, error) {
	return w.client.GetAllReplicators(ctx)
}

func (w *Wrapper) AddP2PCollections(ctx context.Context, collectionIDs []string) error {
	return w.client.AddP2PCollections(ctx, collectionIDs)
}

func (w *Wrapper) RemoveP2PCollections(ctx context.Context, collectionIDs []string) error {
	return w.client.RemoveP2PCollections(ctx, collectionIDs)
}

func (w *Wrapper) GetAllP2PCollections(ctx context.Context) ([]string, error) {
	return w.client.GetAllP2PCollections(ctx)
}

func (w *Wrapper) BasicImport(ctx context.Context, filepath string) error {
	return w.client.BasicImport(ctx, filepath)
}

func (w *Wrapper) BasicExport(ctx context.Context, config *client.BackupConfig) error {
	return w.client.BasicExport(ctx, config)
}

func (w *Wrapper) AddSchema(ctx context.Context, schema string) ([]client.CollectionDescription, error) {
	return w.client.AddSchema(ctx, schema)
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

func (w *Wrapper) ExecRequest(ctx context.Context, query string) *client.RequestResult {
	return w.client.ExecRequest(ctx, query)
}

func (w *Wrapper) NewTxn(ctx context.Context, readOnly bool) (datastore.Txn, error) {
	client, err := w.client.NewTxn(ctx, readOnly)
	if err != nil {
		return nil, err
	}
	server, err := w.handler.Transaction(client.ID())
	if err != nil {
		return nil, err
	}
	return &TxWrapper{server, client}, nil
}

func (w *Wrapper) NewConcurrentTxn(ctx context.Context, readOnly bool) (datastore.Txn, error) {
	client, err := w.client.NewConcurrentTxn(ctx, readOnly)
	if err != nil {
		return nil, err
	}
	server, err := w.handler.Transaction(client.ID())
	if err != nil {
		return nil, err
	}
	return &TxWrapper{server, client}, nil
}

func (w *Wrapper) WithTxn(tx datastore.Txn) client.Store {
	return w.client.WithTxn(tx)
}

func (w *Wrapper) Root() datastore.RootStore {
	return w.node.Root()
}

func (w *Wrapper) Blockstore() blockstore.Blockstore {
	return w.node.Blockstore()
}

func (w *Wrapper) Peerstore() datastore.DSBatching {
	return w.node.Peerstore()
}

func (w *Wrapper) ACPModule(ctx context.Context) (immutable.Option[acp.ACPModule], error) {
	return w.node.ACPModule(ctx)
}

func (w *Wrapper) AddPolicy(
	ctx context.Context,
	creator string,
	policy string,
) (client.AddPolicyResult, error) {
	return w.node.AddPolicy(ctx, creator, policy)
}

func (w *Wrapper) Close() {
	w.httpServer.CloseClientConnections()
	w.httpServer.Close()
	w.node.Close()
}

func (w *Wrapper) Events() events.Events {
	return w.node.Events()
}

func (w *Wrapper) MaxTxnRetries() int {
	return w.node.MaxTxnRetries()
}

func (w *Wrapper) PrintDump(ctx context.Context) error {
	return w.node.PrintDump(ctx)
}

func (w *Wrapper) Bootstrap(addrs []peer.AddrInfo) {
	w.node.Bootstrap(addrs)
}

func (w *Wrapper) WaitForPushLogByPeerEvent(id peer.ID) error {
	return w.node.WaitForPushLogByPeerEvent(id)
}

func (w *Wrapper) WaitForPushLogFromPeerEvent(id peer.ID) error {
	return w.node.WaitForPushLogFromPeerEvent(id)
}
