// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

//go:build js

package js

import (
	"context"
	sysjs "syscall/js"

	"github.com/sourcenetwork/goji"
	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/js"
	"github.com/sourcenetwork/defradb/node"
)

var (
	_ client.Store    = (*Wrapper)(nil)
	_ client.TxnStore = (*Wrapper)(nil)
	_ client.P2P      = (*Wrapper)(nil)
)

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

func (w *Wrapper) PeerInfo(ctx context.Context, opts ...options.Enumerable[options.PeerInfoOptions]) ([]string, error) {
	res, err := execute(ctx, w.value, "peerInfo", jsOpts(opts))
	if err != nil {
		return nil, err
	}
	var out []string
	if err := goji.UnmarshalJS(res[0], &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (w *Wrapper) ActivePeers(
	ctx context.Context,
	opts ...options.Enumerable[options.ActivePeersOptions],
) ([]string, error) {
	res, err := execute(ctx, w.value, "activePeers", jsOpts(opts))
	if err != nil {
		return nil, err
	}
	var out []string
	if err := goji.UnmarshalJS(res[0], &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (w *Wrapper) Connect(
	ctx context.Context,
	addresses []string,
	opts ...options.Enumerable[options.ConnectOptions],
) error {
	_, err := execute(ctx, w.value, "connect", goji.MustMarshalJS(addresses), jsOpts(opts))
	return err
}

func (w *Wrapper) AddReplicator(
	ctx context.Context,
	addresses []string,
	opts ...options.Enumerable[options.AddReplicatorOptions],
) error {
	_, err := execute(ctx, w.value, "addReplicator", goji.MustMarshalJS(addresses), jsOpts(opts))
	return err
}

func (w *Wrapper) DeleteReplicator(
	ctx context.Context,
	id string,
	opts ...options.Enumerable[options.DeleteReplicatorOptions],
) error {
	_, err := execute(ctx, w.value, "deleteReplicator", id, jsOpts(opts))
	return err
}

func (w *Wrapper) ListReplicators(
	ctx context.Context,
	opts ...options.Enumerable[options.ListReplicatorsOptions],
) ([]client.Replicator, error) {
	res, err := execute(ctx, w.value, "listReplicators", jsOpts(opts))
	if err != nil {
		return nil, err
	}
	var out []client.Replicator
	if err := goji.UnmarshalJS(res[0], &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (w *Wrapper) AddP2PCollections(
	ctx context.Context,
	collectionNames []string,
	opts ...options.Enumerable[options.AddP2PCollectionsOptions],
) error {
	_, err := execute(ctx, w.value, "addP2PCollections", goji.MustMarshalJS(collectionNames), jsOpts(opts))
	return err
}

func (w *Wrapper) DeleteP2PCollections(
	ctx context.Context,
	collectionNames []string,
	opts ...options.Enumerable[options.DeleteP2PCollectionsOptions],
) error {
	_, err := execute(ctx, w.value, "deleteP2PCollections", goji.MustMarshalJS(collectionNames), jsOpts(opts))
	return err
}

func (w *Wrapper) ListP2PCollections(
	ctx context.Context,
	opts ...options.Enumerable[options.ListP2PCollectionsOptions],
) ([]string, error) {
	res, err := execute(ctx, w.value, "listP2PCollections", jsOpts(opts))
	if err != nil {
		return nil, err
	}
	var out []string
	if err := goji.UnmarshalJS(res[0], &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (w *Wrapper) AddP2PDocuments(
	ctx context.Context,
	docIDs []string,
	opts ...options.Enumerable[options.AddP2PDocumentsOptions],
) error {
	_, err := execute(ctx, w.value, "addP2PDocuments", goji.MustMarshalJS(docIDs), jsOpts(opts))
	return err
}

func (w *Wrapper) DeleteP2PDocuments(
	ctx context.Context,
	docIDs []string,
	opts ...options.Enumerable[options.DeleteP2PDocumentsOptions],
) error {
	_, err := execute(ctx, w.value, "deleteP2PDocuments", goji.MustMarshalJS(docIDs), jsOpts(opts))
	return err
}

func (w *Wrapper) ListP2PDocuments(
	ctx context.Context,
	opts ...options.Enumerable[options.ListP2PDocumentsOptions],
) ([]string, error) {
	res, err := execute(ctx, w.value, "listP2PDocuments", jsOpts(opts))
	if err != nil {
		return nil, err
	}
	var out []string
	if err := goji.UnmarshalJS(res[0], &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (w *Wrapper) SyncDocuments(
	ctx context.Context,
	collectionName string,
	docIDs []string,
	opts ...options.Enumerable[options.SyncDocumentsOptions],
) error {
	_, err := execute(ctx, w.value, "syncDocuments", collectionName, goji.MustMarshalJS(docIDs), jsOpts(opts))
	return err
}

func (w *Wrapper) SyncCollectionVersions(
	ctx context.Context,
	versionIDs []string,
	opts ...options.Enumerable[options.SyncCollectionVersionsOptions],
) error {
	_, err := execute(ctx, w.value, "syncCollectionVersions", goji.MustMarshalJS(versionIDs), jsOpts(opts))
	return err
}

func (w *Wrapper) SyncBranchableCollection(
	ctx context.Context,
	collectionID string,
	opts ...options.Enumerable[options.SyncBranchableCollectionOptions],
) error {
	_, err := execute(ctx, w.value, "syncBranchableCollection", collectionID, jsOpts(opts))
	return err
}

func (w *Wrapper) BasicImport(ctx context.Context, filepath string) error {
	_, err := execute(ctx, w.value, "basicImport", filepath)
	return err
}

func (w *Wrapper) BasicExport(
	ctx context.Context,
	filepath string,
	opts ...options.Enumerable[options.BasicExportOptions],
) error {
	_, err := execute(ctx, w.value, "basicExport", filepath, jsOpts(opts))
	return err
}

func (w *Wrapper) AddCollection(
	ctx context.Context,
	sdl string,
	opts ...options.Enumerable[options.AddCollectionOptions],
) ([]client.CollectionVersion, error) {
	res, err := execute(ctx, w.value, "addCollection", sdl, jsOpts(opts))
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
	opts ...options.Enumerable[options.AddDACPolicyOptions],
) (client.AddPolicyResult, error) {
	res, err := execute(ctx, w.value, "addDACPolicy", policy, jsOpts(opts))
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
	opts ...options.Enumerable[options.AddDACActorRelationshipOptions],
) (client.AddActorRelationshipResult, error) {
	res, err := execute(ctx, w.value, "addDACActorRelationship",
		collectionName, docID, relation, targetActor, jsOpts(opts))
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
	opts ...options.Enumerable[options.DeleteDACActorRelationshipOptions],
) (client.DeleteActorRelationshipResult, error) {
	res, err := execute(ctx, w.value, "deleteDACActorRelationship",
		collectionName, docID, relation, targetActor, jsOpts(opts))
	if err != nil {
		return client.DeleteActorRelationshipResult{}, err
	}
	var out client.DeleteActorRelationshipResult
	if err := goji.UnmarshalJS(res[0], &out); err != nil {
		return client.DeleteActorRelationshipResult{}, err
	}
	return out, nil
}

func (w *Wrapper) GetNACStatus(
	ctx context.Context,
	opts ...options.Enumerable[options.GetNACStatusOptions],
) (client.NACStatusResult, error) {
	res, err := execute(ctx, w.value, "getNACStatus", jsOpts(opts))
	if err != nil {
		return client.NACStatusResult{}, err
	}
	var out client.NACStatusResult
	if err := goji.UnmarshalJS(res[0], &out); err != nil {
		return client.NACStatusResult{}, err
	}
	return out, nil
}

func (w *Wrapper) ReEnableNAC(ctx context.Context, opts ...options.Enumerable[options.ReEnableNACOptions]) error {
	_, err := execute(ctx, w.value, "reEnableNAC", jsOpts(opts))
	return err
}

func (w *Wrapper) DisableNAC(ctx context.Context, opts ...options.Enumerable[options.DisableNACOptions]) error {
	_, err := execute(ctx, w.value, "disableNAC", jsOpts(opts))
	return err
}

func (w *Wrapper) AddNACActorRelationship(
	ctx context.Context,
	relation string,
	targetActor string,
	opts ...options.Enumerable[options.AddNACActorRelationshipOptions],
) (client.AddActorRelationshipResult, error) {
	res, err := execute(ctx, w.value, "addNACActorRelationship", relation, targetActor, jsOpts(opts))
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
	opts ...options.Enumerable[options.DeleteNACActorRelationshipOptions],
) (client.DeleteActorRelationshipResult, error) {
	res, err := execute(ctx, w.value, "deleteNACActorRelationship", relation, targetActor, jsOpts(opts))
	if err != nil {
		return client.DeleteActorRelationshipResult{}, err
	}
	var out client.DeleteActorRelationshipResult
	if err := goji.UnmarshalJS(res[0], &out); err != nil {
		return client.DeleteActorRelationshipResult{}, err
	}
	return out, nil
}

func (w *Wrapper) PatchCollection(
	ctx context.Context,
	patch string,
	migration immutable.Option[model.Lens],
	opts ...options.Enumerable[options.PatchCollectionOptions],
) error {
	_, err := execute(ctx, w.value, "patchCollection", patch, goji.MustMarshalJS(migration), jsOpts(opts))
	return err
}

func (w *Wrapper) DeleteCollection(
	ctx context.Context,
	names []string,
	opts ...options.Enumerable[options.DeleteCollectionOptions],
) error {
	_, err := execute(ctx, w.value, "deleteCollection", goji.MustMarshalJS(names), jsOpts(opts))
	return err
}

func (w *Wrapper) SetActiveCollectionVersion(
	ctx context.Context,
	collectionVersionID string,
	opts ...options.Enumerable[options.SetActiveCollectionVersionOptions],
) error {
	_, err := execute(ctx, w.value, "setActiveCollectionVersion", collectionVersionID, jsOpts(opts))
	return err
}

func (w *Wrapper) AddView(
	ctx context.Context,
	query string,
	sdl string,
	opts ...options.Enumerable[options.AddViewOptions],
) ([]client.CollectionVersion, error) {
	res, err := execute(ctx, w.value, "addView", query, sdl, jsOpts(opts))
	if err != nil {
		return nil, err
	}
	var out []client.CollectionVersion
	if err := goji.UnmarshalJS(res[0], &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (w *Wrapper) RefreshViews(ctx context.Context, opts ...options.Enumerable[options.RefreshViewsOptions]) error {
	_, err := execute(ctx, w.value, "refreshViews", jsOpts(opts))
	return err
}

func (w *Wrapper) SetMigration(
	ctx context.Context,
	config client.LensConfig,
	opts ...options.Enumerable[options.SetMigrationOptions],
) (string, error) {
	res, err := execute(ctx, w.value, "setMigration", goji.MustMarshalJS(config), jsOpts(opts))
	if err != nil {
		return "", err
	}
	return res[0].String(), err
}

func (w *Wrapper) AddLens(
	ctx context.Context,
	lens model.Lens,
	opts ...options.Enumerable[options.AddLensOptions],
) (string, error) {
	res, err := execute(ctx, w.value, "addLens", goji.MustMarshalJS(lens), jsOpts(opts))
	if err != nil {
		return "", err
	}
	return res[0].String(), err
}

func (w *Wrapper) ListLenses(
	ctx context.Context,
	opts ...options.Enumerable[options.ListLensesOptions],
) (map[string]model.Lens, error) {
	res, err := execute(ctx, w.value, "listLenses", jsOpts(opts))
	if err != nil {
		return nil, err
	}
	var lenses map[string]model.Lens
	if err := goji.UnmarshalJS(res[0], &lenses); err != nil {
		return nil, err
	}
	return lenses, nil
}

func (w *Wrapper) GetCollectionByName(
	ctx context.Context,
	name client.CollectionName,
	opts ...options.Enumerable[options.GetCollectionByNameOptions],
) (client.Collection, error) {
	res, err := execute(ctx, w.value, "getCollectionByName", name, jsOpts(opts))
	if err != nil {
		return nil, err
	}
	return &Collection{
		client: res[0],
	}, nil
}

func (w *Wrapper) GetCollections(
	ctx context.Context,
	opts ...options.Enumerable[options.GetCollectionsOptions],
) ([]client.Collection, error) {
	res, err := execute(ctx, w.value, "getCollections", jsOpts(opts))
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

func (w *Wrapper) ListIndexes(
	ctx context.Context,
	opts ...options.Enumerable[options.ListIndexesOptions],
) (map[client.CollectionName][]client.IndexDescription, error) {
	res, err := execute(ctx, w.value, "listIndexes", jsOpts(opts))
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
	opts ...options.Enumerable[options.ExecRequestOptions],
) *client.RequestResult {
	res, err := execute(ctx, w.value, "execRequest", query, jsOpts(opts))
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

func (w *Wrapper) NewTxn(readOnly bool) (client.Txn, error) {
	res, err := execute(context.Background(), w.value, "newTxn", readOnly)
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
	_, err := execute(ctx, w.value, "printDump")
	return err
}

func (w *Wrapper) Disconnect(ctx context.Context, addresses []string, opts ...options.Enumerable[options.DisconnectOptions]) error {
	_, err := execute(ctx, w.value, "disconnect", goji.MustMarshalJS(addresses), jsOpts(opts))
	return err
}

func (w *Wrapper) GetNodeIdentity(ctx context.Context) (immutable.Option[acpIdentity.PublicRawIdentity], error) {
	res, err := execute(ctx, w.value, "getNodeIdentity")
	if err != nil {
		return immutable.None[acpIdentity.PublicRawIdentity](), err
	}
	var out immutable.Option[acpIdentity.PublicRawIdentity]
	if err := goji.UnmarshalJS(res[0], &out); err != nil {
		return immutable.None[acpIdentity.PublicRawIdentity](), err
	}
	return out, nil
}

func (w *Wrapper) VerifySignature(
	ctx context.Context,
	blockCid string,
	pubKey crypto.PublicKey,
	opts ...options.Enumerable[options.VerifySignatureOptions],
) error {
	_, err := execute(ctx, w.value, "verifySignature",
		pubKey.String(), string(pubKey.Type()), blockCid, jsOpts(opts))
	return err
}

func (w *Wrapper) ListAllEncryptedIndexes(
	ctx context.Context,
	opts ...options.Enumerable[options.ListAllEncryptedIndexesOptions],
) (map[client.CollectionName][]client.EncryptedIndexDescription, error) {
	res, err := execute(ctx, w.value, "listAllEncryptedIndexes", jsOpts(opts))
	if err != nil {
		return nil, err
	}
	var out map[client.CollectionName][]client.EncryptedIndexDescription
	if err := goji.UnmarshalJS(res[0], &out); err != nil {
		return nil, err
	}
	return out, nil
}
