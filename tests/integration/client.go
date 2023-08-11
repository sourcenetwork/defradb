package tests

import (
	"context"

	blockstore "github.com/ipfs/boxo/blockstore"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/events"
)

var _ client.Store = (*Client)(nil)
var _ client.DB = (*Client)(nil)

// Client splits the client.DB and client.Store interfaces in two so we can test
// different implementations of client.Store without changing integration tests.
type Client struct {
	db    client.DB
	store client.Store
}

func NewClient(db client.DB, store client.Store) *Client {
	return &Client{db, store}
}

func (c *Client) SetReplicator(ctx context.Context, rep client.Replicator) error {
	return c.store.SetReplicator(ctx, rep)
}

func (c *Client) DeleteReplicator(ctx context.Context, rep client.Replicator) error {
	return c.store.DeleteReplicator(ctx, rep)
}

func (c *Client) GetAllReplicators(ctx context.Context) ([]client.Replicator, error) {
	return c.store.GetAllReplicators(ctx)
}

func (c *Client) AddP2PCollection(ctx context.Context, collectionID string) error {
	return c.store.AddP2PCollection(ctx, collectionID)
}

func (c *Client) RemoveP2PCollection(ctx context.Context, collectionID string) error {
	return c.store.RemoveP2PCollection(ctx, collectionID)
}

func (c *Client) GetAllP2PCollections(ctx context.Context) ([]string, error) {
	return c.store.GetAllP2PCollections(ctx)
}

func (c *Client) BasicImport(ctx context.Context, filepath string) error {
	return c.store.BasicImport(ctx, filepath)
}

func (c *Client) BasicExport(ctx context.Context, config *client.BackupConfig) error {
	return c.store.BasicExport(ctx, config)
}

func (c *Client) AddSchema(ctx context.Context, schema string) ([]client.CollectionDescription, error) {
	return c.store.AddSchema(ctx, schema)
}

func (c *Client) PatchSchema(ctx context.Context, patch string) error {
	return c.store.PatchSchema(ctx, patch)
}

func (c *Client) SetMigration(ctx context.Context, config client.LensConfig) error {
	return c.store.SetMigration(ctx, config)
}

func (c *Client) LensRegistry() client.LensRegistry {
	return c.store.LensRegistry()
}

func (c *Client) GetCollectionByName(ctx context.Context, name client.CollectionName) (client.Collection, error) {
	return c.store.GetCollectionByName(ctx, name)
}

func (c *Client) GetCollectionBySchemaID(ctx context.Context, schemaId string) (client.Collection, error) {
	return c.store.GetCollectionBySchemaID(ctx, schemaId)
}

func (c *Client) GetCollectionByVersionID(ctx context.Context, versionId string) (client.Collection, error) {
	return c.store.GetCollectionByVersionID(ctx, versionId)
}

func (c *Client) GetAllCollections(ctx context.Context) ([]client.Collection, error) {
	return c.store.GetAllCollections(ctx)
}

func (c *Client) GetAllIndexes(ctx context.Context) (map[client.CollectionName][]client.IndexDescription, error) {
	return c.store.GetAllIndexes(ctx)
}

func (c *Client) ExecRequest(ctx context.Context, query string) *client.RequestResult {
	return c.store.ExecRequest(ctx, query)
}

func (c *Client) NewTxn(ctx context.Context, b bool) (datastore.Txn, error) {
	return c.db.NewTxn(ctx, b)
}

func (c *Client) NewConcurrentTxn(ctx context.Context, b bool) (datastore.Txn, error) {
	return c.db.NewConcurrentTxn(ctx, b)
}

func (c *Client) WithTxn(tx datastore.Txn) client.Store {
	return c.db.WithTxn(tx)
}

func (c *Client) Root() datastore.RootStore {
	return c.db.Root()
}

func (c *Client) Blockstore() blockstore.Blockstore {
	return c.db.Blockstore()
}

func (c *Client) Close(ctx context.Context) {
	c.db.Close(ctx)
}

func (c *Client) Events() events.Events {
	return c.db.Events()
}

func (c *Client) MaxTxnRetries() int {
	return c.db.MaxTxnRetries()
}

func (c *Client) PrintDump(ctx context.Context) error {
	return c.db.PrintDump(ctx)
}
