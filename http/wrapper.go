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
	"net/http/httptest"

	blockstore "github.com/ipfs/boxo/blockstore"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/events"
)

var _ client.DB = (*Wrapper)(nil)

// Wrapper combines an HTTP client and server into a
// single struct that implements the client.DB interface.
type Wrapper struct {
	db         client.DB
	handler    *Handler
	client     *Client
	httpServer *httptest.Server
}

func NewWrapper(db client.DB) (*Wrapper, error) {
	handler := NewHandler(db, ServerOptions{})
	httpServer := httptest.NewServer(handler)

	client, err := NewClient(httpServer.URL)
	if err != nil {
		return nil, err
	}

	return &Wrapper{
		db,
		handler,
		client,
		httpServer,
	}, nil
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

func (w *Wrapper) AddP2PCollection(ctx context.Context, collectionID string) error {
	return w.client.AddP2PCollection(ctx, collectionID)
}

func (w *Wrapper) RemoveP2PCollection(ctx context.Context, collectionID string) error {
	return w.client.RemoveP2PCollection(ctx, collectionID)
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

func (w *Wrapper) PatchSchema(ctx context.Context, patch string) error {
	return w.client.PatchSchema(ctx, patch)
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

func (w *Wrapper) GetCollectionBySchemaID(ctx context.Context, schemaId string) (client.Collection, error) {
	return w.client.GetCollectionBySchemaID(ctx, schemaId)
}

func (w *Wrapper) GetCollectionByVersionID(ctx context.Context, versionId string) (client.Collection, error) {
	return w.client.GetCollectionByVersionID(ctx, versionId)
}

func (w *Wrapper) GetAllCollections(ctx context.Context) ([]client.Collection, error) {
	return w.client.GetAllCollections(ctx)
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
	server, ok := w.handler.txs.Load(client.ID())
	if !ok {
		return nil, fmt.Errorf("failed to get server transaction")
	}
	return &TxWrapper{server.(datastore.Txn), client}, nil
}

func (w *Wrapper) NewConcurrentTxn(ctx context.Context, readOnly bool) (datastore.Txn, error) {
	client, err := w.client.NewConcurrentTxn(ctx, readOnly)
	if err != nil {
		return nil, err
	}
	server, ok := w.handler.txs.Load(client.ID())
	if !ok {
		return nil, fmt.Errorf("failed to get server transaction")
	}
	return &TxWrapper{server.(datastore.Txn), client}, nil
}

func (w *Wrapper) WithTxn(tx datastore.Txn) client.Store {
	return w.client.WithTxn(tx)
}

func (w *Wrapper) Root() datastore.RootStore {
	return w.db.Root()
}

func (w *Wrapper) Blockstore() blockstore.Blockstore {
	return w.db.Blockstore()
}

func (w *Wrapper) Close(ctx context.Context) {
	w.httpServer.CloseClientConnections()
	w.httpServer.Close()
	w.db.Close(ctx)
}

func (w *Wrapper) Events() events.Events {
	return w.db.Events()
}

func (w *Wrapper) MaxTxnRetries() int {
	return w.db.MaxTxnRetries()
}

func (w *Wrapper) PrintDump(ctx context.Context) error {
	return w.db.PrintDump(ctx)
}
