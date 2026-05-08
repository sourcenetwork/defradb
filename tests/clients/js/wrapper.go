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
	iIdentity "github.com/sourcenetwork/defradb/internal/identity"
	"github.com/sourcenetwork/defradb/internal/utils"
	"github.com/sourcenetwork/defradb/js"
	"github.com/sourcenetwork/defradb/node"
)

// identityProvider is any options struct that has a GetIdentity method.
type identityProvider interface {
	GetIdentity() immutable.Option[acpIdentity.Identity]
}

// ctxWithOptIdentity extracts identity from opts and puts it in context,
// so that the JS bridge's execute function can pass it to the JS client.
// Only sets identity if the opts actually have one, to avoid overwriting
// an existing identity in context with None.
func ctxWithOptIdentity(ctx context.Context, opt identityProvider) context.Context {
	if opt == nil {
		return ctx
	}
	ident := opt.GetIdentity()
	if ident.HasValue() {
		return iIdentity.WithContext(ctx, ident)
	}
	return ctx
}

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

func (w *Wrapper) PeerInfo(ctx context.Context, opts ...options.Enumerable[options.PeerInfoOptions]) ([]string, error) {
	return nil, nil
}

func (w *Wrapper) ActivePeers(
	ctx context.Context,
	opts ...options.Enumerable[options.ActivePeersOptions],
) ([]string, error) {
	panic("not implemented")
}

func (w *Wrapper) AddReplicator(
	ctx context.Context,
	addresses []string,
	opts ...options.Enumerable[options.AddReplicatorOptions],
) error {
	panic("not implemented")
}

func (w *Wrapper) DeleteReplicator(
	ctx context.Context,
	id string,
	opts ...options.Enumerable[options.DeleteReplicatorOptions],
) error {
	panic("not implemented")
}

func (w *Wrapper) ListReplicators(
	ctx context.Context,
	opts ...options.Enumerable[options.ListReplicatorsOptions],
) ([]client.Replicator, error) {
	panic("not implemented")
}

func (w *Wrapper) AddP2PCollections(
	ctx context.Context,
	collectionNames []string,
	opts ...options.Enumerable[options.AddP2PCollectionsOptions],
) error {
	panic("not implemented")
}

func (w *Wrapper) DeleteP2PCollections(
	ctx context.Context,
	collectionNames []string,
	opts ...options.Enumerable[options.DeleteP2PCollectionsOptions],
) error {
	panic("not implemented")
}

func (w *Wrapper) ListP2PCollections(
	ctx context.Context,
	opts ...options.Enumerable[options.ListP2PCollectionsOptions],
) ([]string, error) {
	panic("not implemented")
}

func (w *Wrapper) AddP2PDocuments(
	ctx context.Context,
	docIDs []string,
	opts ...options.Enumerable[options.AddP2PDocumentsOptions],
) error {
	panic("not implemented")
}

func (w *Wrapper) DeleteP2PDocuments(
	ctx context.Context,
	docIDs []string,
	opts ...options.Enumerable[options.DeleteP2PDocumentsOptions],
) error {
	panic("not implemented")
}

func (w *Wrapper) ListP2PDocuments(ctx context.Context, opts ...options.Enumerable[options.ListP2PDocumentsOptions]) ([]string, error) {
	panic("not implemented")
}

func (w *Wrapper) SyncDocuments(ctx context.Context, collectionName string, docIDs []string, opts ...options.Enumerable[options.SyncDocumentsOptions]) error {
	panic("not implemented")
}

func (w *Wrapper) SyncCollectionVersions(ctx context.Context, versionIDs []string, opts ...options.Enumerable[options.SyncCollectionVersionsOptions]) error {
	panic("not implemented")
}

func (w *Wrapper) SyncBranchableCollection(ctx context.Context, collectionID string, opts ...options.Enumerable[options.SyncBranchableCollectionOptions]) error {
	panic("not implemented")
}

func (w *Wrapper) BasicImport(ctx context.Context, filepath string) error {
	panic("not implemented")
}

func (w *Wrapper) BasicExport(ctx context.Context, filepath string, opts ...options.Enumerable[options.BasicExportOptions]) error {
	panic("not implemented")
}

func (w *Wrapper) AddCollection(
	ctx context.Context,
	sdl string,
	opts ...options.Enumerable[options.AddCollectionOptions],
) ([]client.CollectionVersion, error) {
	opt := utils.NewOptions(opts...)
	ctx = ctxWithOptIdentity(ctx, opt)
	res, err := execute(ctx, w.value, "addCollection", sdl)
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
	opt := utils.NewOptions(opts...)
	ctx = ctxWithOptIdentity(ctx, opt)
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
	opts ...options.Enumerable[options.AddDACActorRelationshipOptions],
) (client.AddActorRelationshipResult, error) {
	opt := utils.NewOptions(opts...)
	ctx = ctxWithOptIdentity(ctx, opt)
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
	opts ...options.Enumerable[options.DeleteDACActorRelationshipOptions],
) (client.DeleteActorRelationshipResult, error) {
	opt := utils.NewOptions(opts...)
	ctx = ctxWithOptIdentity(ctx, opt)
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

func (w *Wrapper) GetNACStatus(
	ctx context.Context,
	opts ...options.Enumerable[options.GetNACStatusOptions],
) (client.NACStatusResult, error) {
	opt := utils.NewOptions(opts...)
	ctx = ctxWithOptIdentity(ctx, opt)
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

func (w *Wrapper) ReEnableNAC(ctx context.Context, opts ...options.Enumerable[options.ReEnableNACOptions]) error {
	opt := utils.NewOptions(opts...)
	ctx = ctxWithOptIdentity(ctx, opt)
	_, err := execute(ctx, w.value, "reEnableNAC")
	return err
}

func (w *Wrapper) DisableNAC(ctx context.Context, opts ...options.Enumerable[options.DisableNACOptions]) error {
	opt := utils.NewOptions(opts...)
	ctx = ctxWithOptIdentity(ctx, opt)
	_, err := execute(ctx, w.value, "disableNAC")
	return err
}

func (w *Wrapper) AddNACActorRelationship(
	ctx context.Context,
	relation string,
	targetActor string,
	opts ...options.Enumerable[options.AddNACActorRelationshipOptions],
) (client.AddActorRelationshipResult, error) {
	opt := utils.NewOptions(opts...)
	ctx = ctxWithOptIdentity(ctx, opt)
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
	opts ...options.Enumerable[options.DeleteNACActorRelationshipOptions],
) (client.DeleteActorRelationshipResult, error) {
	opt := utils.NewOptions(opts...)
	ctx = ctxWithOptIdentity(ctx, opt)
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

func (w *Wrapper) PatchCollection(
	ctx context.Context,
	patch string,
	migration immutable.Option[model.Lens],
	opts ...options.Enumerable[options.PatchCollectionOptions],
) error {
	migrationVal, err := goji.MarshalJS(migration)
	if err != nil {
		return err
	}
	opt := utils.NewOptions(opts...)
	ctx = ctxWithOptIdentity(ctx, opt)
	_, err = execute(ctx, w.value, "patchCollection", patch, migrationVal)
	return err
}

func (w *Wrapper) DeleteCollection(
	ctx context.Context,
	names []string,
	opts ...options.Enumerable[options.DeleteCollectionOptions],
) error {
	opt := utils.NewOptions(opts...)
	ctx = ctxWithOptIdentity(ctx, opt)
	namesVal, err := goji.MarshalJS(names)
	if err != nil {
		return err
	}
	_, err = execute(ctx, w.value, "deleteCollection", namesVal, opt.ActiveOnly)
	return err
}

func (w *Wrapper) SetActiveCollectionVersion(
	ctx context.Context,
	collectionVersionID string,
	opts ...options.Enumerable[options.SetActiveCollectionVersionOptions],
) error {
	opt := utils.NewOptions(opts...)
	ctx = ctxWithOptIdentity(ctx, opt)
	_, err := execute(ctx, w.value, "setActiveCollectionVersion", collectionVersionID)
	return err
}

func (w *Wrapper) AddView(
	ctx context.Context,
	query string,
	sdl string,
	opts ...options.Enumerable[options.AddViewOptions],
) ([]client.CollectionVersion, error) {
	opt := utils.NewOptions(opts...)
	ctx = ctxWithOptIdentity(ctx, opt)

	transformCIDVal, err := goji.MarshalJS(opt.TransformCID)
	if err != nil {
		return nil, err
	}
	res, err := execute(ctx, w.value, "addView", query, sdl, transformCIDVal)
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
	var optsVal sysjs.Value
	var err error
	opt := utils.NewOptions(opts...)
	if opt != nil {
		ctx = ctxWithOptIdentity(ctx, opt)
		optsVal, err = goji.MarshalJS(opt)
		if err != nil {
			return err
		}
	} else {
		optsVal = sysjs.Undefined()
	}
	_, err = execute(ctx, w.value, "refreshViews", optsVal)
	return err
}

func (w *Wrapper) SetMigration(ctx context.Context, config client.LensConfig, opts ...options.Enumerable[options.SetMigrationOptions]) (string, error) {
	opt := utils.NewOptions(opts...)
	if opt != nil {
		ctx = ctxWithOptIdentity(ctx, opt)
	}
	configVal, err := goji.MarshalJS(config)
	if err != nil {
		return "", err
	}
	res, err := execute(ctx, w.value, "setMigration", configVal)
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
	opt := utils.NewOptions(opts...)
	if opt != nil {
		ctx = ctxWithOptIdentity(ctx, opt)
	}
	lensVal, err := goji.MarshalJS(lens)
	if err != nil {
		return "", err
	}
	res, err := execute(ctx, w.value, "addLens", lensVal)
	if err != nil {
		return "", err
	}
	return res[0].String(), err
}

func (w *Wrapper) ListLenses(
	ctx context.Context,
	opts ...options.Enumerable[options.ListLensesOptions],
) (map[string]model.Lens, error) {
	opt := utils.NewOptions(opts...)
	ctx = ctxWithOptIdentity(ctx, opt)
	res, err := execute(ctx, w.value, "listLenses")
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
	opt := utils.NewOptions(opts...)
	ctx = ctxWithOptIdentity(ctx, opt)
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
	opts ...options.Enumerable[options.GetCollectionsOptions],
) ([]client.Collection, error) {
	var optsVal sysjs.Value
	var err error
	opt := utils.NewOptions(opts...)
	if opt != nil {
		ctx = ctxWithOptIdentity(ctx, opt)
		optsVal, err = goji.MarshalJS(opt)
		if err != nil {
			return nil, err
		}
	} else {
		optsVal = sysjs.Undefined()
	}
	res, err := execute(ctx, w.value, "getCollections", optsVal)
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
	opt := utils.NewOptions(opts...)
	ctx = ctxWithOptIdentity(ctx, opt)
	res, err := execute(ctx, w.value, "listIndexes")
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
	var optsVal sysjs.Value
	opt := utils.NewOptions(opts...)
	if opt != nil {
		ctx = ctxWithOptIdentity(ctx, opt)
		var err error
		optsVal, err = goji.MarshalJS(opt)
		if err != nil {
			panic(err)
		}
	} else {
		optsVal = sysjs.Undefined()
	}
	res, err := execute(ctx, w.value, "execRequest", query, optsVal)
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
	return w.node.DB.PrintDump(ctx)
}

func (w *Wrapper) Connect(ctx context.Context, addresses []string, opts ...options.Enumerable[options.ConnectOptions]) error {
	return w.node.DB.Connect(ctx, addresses, opts...)
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
	opt := utils.NewOptions(opts...)
	ctx = ctxWithOptIdentity(ctx, opt)
	_, err := execute(ctx, w.value, "verifySignature", pubKey.String(), string(pubKey.Type()), blockCid)
	return err
}

func (w *Wrapper) ListAllEncryptedIndexes(
	ctx context.Context,
	opts ...options.Enumerable[options.ListAllEncryptedIndexesOptions],
) (map[client.CollectionName][]client.EncryptedIndexDescription, error) {
	opt := utils.NewOptions(opts...)
	ctx = ctxWithOptIdentity(ctx, opt)
	return w.node.DB.ListAllEncryptedIndexes(ctx, opts...)
}
