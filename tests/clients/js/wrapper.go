// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build js

package js

import (
	"context"
	sysjs "syscall/js"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sourcenetwork/goji"
	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/js"
	"github.com/sourcenetwork/defradb/node"
)

var _ client.TxnStore = (*Wrapper)(nil)

// Wrapper implements the client.TxnStore
// interface using the JS client.
type Wrapper struct {
	client *js.Client
	value  sysjs.Value
	node   *node.Node
}

func NewWrapper(node *node.Node) (*Wrapper, error) {
	client := js.NewClient(node)
	return &Wrapper{
		client: client,
		value:  client.JSValue(),
		node:   node,
	}, nil
}

func (w *Wrapper) PeerInfo() peer.AddrInfo {
	panic("not implemented")
}

func (w *Wrapper) SetReplicator(ctx context.Context, info peer.AddrInfo, collections ...string) error {
	panic("not implemented")
}

func (w *Wrapper) DeleteReplicator(ctx context.Context, info peer.AddrInfo, collections ...string) error {
	panic("not implemented")
}

func (w *Wrapper) GetAllReplicators(ctx context.Context) ([]client.Replicator, error) {
	panic("not implemented")
}

func (w *Wrapper) AddP2PCollections(ctx context.Context, collectionIDs ...string) error {
	panic("not implemented")
}

func (w *Wrapper) RemoveP2PCollections(ctx context.Context, collectionIDs ...string) error {
	panic("not implemented")
}

func (w *Wrapper) GetAllP2PCollections(ctx context.Context) ([]string, error) {
	panic("not implemented")
}

func (w *Wrapper) AddP2PDocuments(ctx context.Context, docIDs ...string) error {
	panic("not implemented")
}

func (w *Wrapper) RemoveP2PDocuments(ctx context.Context, docIDs ...string) error {
	panic("not implemented")
}

func (w *Wrapper) GetAllP2PDocuments(ctx context.Context) ([]string, error) {
	panic("not implemented")
}

func (w *Wrapper) SyncDocuments(
	ctx context.Context,
	collectionName string,
	docIDs []string,
) error {
	panic("not implemented")
}

func (w *Wrapper) BasicImport(ctx context.Context, filepath string) error {
	panic("not implemented")
}

func (w *Wrapper) BasicExport(ctx context.Context, config *client.BackupConfig) error {
	panic("not implemented")
}

func (w *Wrapper) AddSchema(ctx context.Context, schema string) ([]client.CollectionVersion, error) {
	res, err := execute(ctx, w.value, "addSchema", schema)
	if err != nil {
		return nil, err
	}
	var out []client.CollectionVersion
	if err := goji.UnmarshalJS(res[0], &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (w *Wrapper) AddDACPolicy(
	ctx context.Context,
	policy string,
) (client.AddPolicyResult, error) {
	res, err := execute(ctx, w.value, "addDACPolicy", policy)
	if err != nil {
		return client.AddPolicyResult{}, err
	}
	var out client.AddPolicyResult
	if err := goji.UnmarshalJS(res[0], &out); err != nil {
		return client.AddPolicyResult{}, err
	}
	return out, nil
}

func (w *Wrapper) AddDACActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
) (client.AddActorRelationshipResult, error) {
	res, err := execute(ctx, w.value, "addDACActorRelationship", collectionName, docID, relation, targetActor)
	if err != nil {
		return client.AddActorRelationshipResult{}, err
	}
	var out client.AddActorRelationshipResult
	if err := goji.UnmarshalJS(res[0], &out); err != nil {
		return client.AddActorRelationshipResult{}, err
	}
	return out, nil
}

func (w *Wrapper) DeleteDACActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
) (client.DeleteActorRelationshipResult, error) {
	res, err := execute(ctx, w.value, "deleteDACActorRelationship", collectionName, docID, relation, targetActor)
	if err != nil {
		return client.DeleteActorRelationshipResult{}, err
	}
	var out client.DeleteActorRelationshipResult
	if err := goji.UnmarshalJS(res[0], &out); err != nil {
		return client.DeleteActorRelationshipResult{}, err
	}
	return out, nil
}

func (w *Wrapper) GetNACStatus(ctx context.Context) (client.NACStatusResult, error) {
	res, err := execute(ctx, w.value, "getNACStatus")
	if err != nil {
		return client.NACStatusResult{}, err
	}
	var out client.NACStatusResult
	if err := goji.UnmarshalJS(res[0], &out); err != nil {
		return client.NACStatusResult{}, err
	}
	return out, nil
}

func (w *Wrapper) ReEnableNAC(ctx context.Context) error {
	_, err := execute(ctx, w.value, "reEnableNAC")
	return err
}

func (w *Wrapper) DisableNAC(ctx context.Context) error {
	_, err := execute(ctx, w.value, "disableNAC")
	return err
}

func (w *Wrapper) AddNACActorRelationship(
	ctx context.Context,
	relation string,
	targetActor string,
) (client.AddActorRelationshipResult, error) {
	res, err := execute(ctx, w.value, "addNACActorRelationship", relation, targetActor)
	if err != nil {
		return client.AddActorRelationshipResult{}, err
	}
	var out client.AddActorRelationshipResult
	if err := goji.UnmarshalJS(res[0], &out); err != nil {
		return client.AddActorRelationshipResult{}, err
	}
	return out, nil
}

func (w *Wrapper) DeleteNACActorRelationship(
	ctx context.Context,
	relation string,
	targetActor string,
) (client.DeleteActorRelationshipResult, error) {
	res, err := execute(ctx, w.value, "deleteNACActorRelationship", relation, targetActor)
	if err != nil {
		return client.DeleteActorRelationshipResult{}, err
	}
	var out client.DeleteActorRelationshipResult
	if err := goji.UnmarshalJS(res[0], &out); err != nil {
		return client.DeleteActorRelationshipResult{}, err
	}
	return out, nil
}

func (w *Wrapper) PatchSchema(
	ctx context.Context,
	patch string,
	migration immutable.Option[model.Lens],
	setAsDefaultVersion bool,
) error {
	migrationVal, err := goji.MarshalJS(migration)
	if err != nil {
		return err
	}
	_, err = execute(ctx, w.value, "patchSchema", patch, migrationVal, setAsDefaultVersion)
	return err
}

func (w *Wrapper) PatchCollection(
	ctx context.Context,
	patch string,
) error {
	_, err := execute(ctx, w.value, "patchCollection", patch)
	return err
}

func (w *Wrapper) SetActiveSchemaVersion(ctx context.Context, schemaVersionID string) error {
	_, err := execute(ctx, w.value, "setActiveSchemaVersion", schemaVersionID)
	return err
}

func (w *Wrapper) AddView(
	ctx context.Context,
	query string,
	sdl string,
	transform immutable.Option[model.Lens],
) ([]client.CollectionDefinition, error) {
	transformVal, err := goji.MarshalJS(transform)
	if err != nil {
		return nil, err
	}
	res, err := execute(ctx, w.value, "addView", query, sdl, transformVal)
	if err != nil {
		return nil, err
	}
	var out []client.CollectionDefinition
	if err := goji.UnmarshalJS(res[0], &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (w *Wrapper) RefreshViews(ctx context.Context, opts client.CollectionFetchOptions) error {
	optsVal, err := goji.MarshalJS(opts)
	if err != nil {
		return err
	}
	_, err = execute(ctx, w.value, "refreshViews", optsVal)
	return err
}

func (w *Wrapper) SetMigration(ctx context.Context, config client.LensConfig) error {
	configVal, err := goji.MarshalJS(config)
	if err != nil {
		return err
	}
	_, err = execute(ctx, w.value, "setMigration", configVal)
	return err
}

func (w *Wrapper) LensRegistry() client.LensRegistry {
	res, err := execute(context.Background(), w.value, "lensRegistry")
	if err != nil {
		panic(err)
	}
	return &LensRegistry{
		client: res[0],
	}
}

func (w *Wrapper) GetCollectionByName(ctx context.Context, name client.CollectionName) (client.Collection, error) {
	res, err := execute(ctx, w.value, "getCollectionByName", name)
	if err != nil {
		return nil, err
	}
	return &Collection{
		client: res[0],
	}, nil
}

func (w *Wrapper) GetCollections(
	ctx context.Context,
	options client.CollectionFetchOptions,
) ([]client.Collection, error) {
	optionsVal, err := goji.MarshalJS(options)
	if err != nil {
		return nil, err
	}
	res, err := execute(ctx, w.value, "getCollections", optionsVal)
	if err != nil {
		return nil, err
	}
	out := make([]client.Collection, res[0].Length())
	for i := range out {
		out[i] = &Collection{
			client: res[0].Index(i),
		}
	}
	return out, nil
}

func (w *Wrapper) GetSchemaByVersionID(ctx context.Context, versionID string) (client.SchemaDescription, error) {
	res, err := execute(ctx, w.value, "getSchemaByVersionID", versionID)
	if err != nil {
		return client.SchemaDescription{}, err
	}
	var out client.SchemaDescription
	if err := goji.UnmarshalJS(res[0], &out); err != nil {
		return client.SchemaDescription{}, err
	}
	return out, nil
}

func (w *Wrapper) GetSchemas(
	ctx context.Context,
	options client.SchemaFetchOptions,
) ([]client.SchemaDescription, error) {
	optionsVal, err := goji.MarshalJS(options)
	if err != nil {
		return nil, err
	}
	res, err := execute(ctx, w.value, "getSchemas", optionsVal)
	if err != nil {
		return nil, err
	}
	var out []client.SchemaDescription
	if err := goji.UnmarshalJS(res[0], &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (w *Wrapper) GetAllIndexes(ctx context.Context) (map[client.CollectionName][]client.IndexDescription, error) {
	res, err := execute(ctx, w.value, "getAllIndexes")
	if err != nil {
		return nil, err
	}
	var out map[client.CollectionName][]client.IndexDescription
	if err := goji.UnmarshalJS(res[0], &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (w *Wrapper) ExecRequest(
	ctx context.Context,
	query string,
	opts ...client.RequestOption,
) *client.RequestResult {
	var gqlOpts client.GQLOptions
	for _, o := range opts {
		o(&gqlOpts)
	}
	reqOpts, err := goji.MarshalJS(gqlOpts)
	if err != nil {
		panic(err)
	}
	res, err := execute(ctx, w.value, "execRequest", query, reqOpts)
	if err != nil {
		panic(err)
	}
	var gql client.GQLResult
	if err := goji.UnmarshalJS(res[0].Get("gql"), &gql); err != nil {
		gql.Errors = append(gql.Errors, err)
	}
	out := client.RequestResult{
		GQL: gql,
	}
	if v := res[0].Get("subscription"); v.Type() == sysjs.TypeObject {
		out.Subscription = handleSubscription(v)
	}
	return &out
}

// handleSubscription reads values from the subscription async iterator
// and puts them into a channel.
func handleSubscription(value sysjs.Value) <-chan client.GQLResult {
	iter := goji.ForAwaitOf(value)
	sub := make(chan client.GQLResult)
	go func() {
		defer close(sub)
		for val := range iter {
			var gql client.GQLResult
			if err := goji.UnmarshalJS(val.Value, &gql); err != nil {
				gql.Errors = append(gql.Errors, err)
			}
			if val.Error != nil {
				gql.Errors = append(gql.Errors, val.Error)
			}
			sub <- gql
		}
	}()
	return sub
}

func (w *Wrapper) NewTxn(ctx context.Context, readOnly bool) (client.Txn, error) {
	res, err := execute(ctx, w.value, "newTxn", readOnly)
	if err != nil {
		return nil, err
	}
	client := res[0]
	id := uint64(client.Get("id").Int())
	txn, err := w.client.Transaction(id)
	if err != nil {
		return nil, err
	}
	return &Transaction{w, txn}, nil
}

func (w *Wrapper) NewConcurrentTxn(ctx context.Context, readOnly bool) (client.Txn, error) {
	res, err := execute(ctx, w.value, "newConcurrentTxn", readOnly)
	if err != nil {
		return nil, err
	}
	client := res[0]
	id := uint64(client.Get("id").Int())
	txn, err := w.client.Transaction(id)
	if err != nil {
		return nil, err
	}
	return &Transaction{w, txn}, nil
}

func (w *Wrapper) Close() {
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

func (w *Wrapper) GetNodeIdentity(ctx context.Context) (immutable.Option[identity.PublicRawIdentity], error) {
	res, err := execute(ctx, w.value, "getNodeIdentity")
	if err != nil {
		return immutable.None[identity.PublicRawIdentity](), err
	}
	var out immutable.Option[identity.PublicRawIdentity]
	if err := goji.UnmarshalJS(res[0], &out); err != nil {
		return immutable.None[identity.PublicRawIdentity](), err
	}
	return out, nil
}

func (w *Wrapper) VerifySignature(ctx context.Context, blockCid string, pubKey crypto.PublicKey) error {
	_, err := execute(ctx, w.value, "verifySignature", pubKey.String(), string(pubKey.Type()), blockCid)
	return err
}

func (w *Wrapper) GetAllEncryptedIndexes(
	ctx context.Context,
) (map[client.CollectionName][]client.EncryptedIndexDescription, error) {
	return w.node.DB.GetAllEncryptedIndexes(ctx)
}
