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

//go:build javaclient

package java

/*
#include <jni.h>
#include "../../../cbindings/defra_structs.h"
#include "jnicall.h"
*/
import "C"

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime/cgo"
	"strings"
	"sync"
	"time"

	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/utils"
	"github.com/sourcenetwork/defradb/node"
)

var _ client.TxnStore = (*Wrapper)(nil)
var _ client.P2P = (*Wrapper)(nil)

// Wrapper implements clients.Client by driving an already-running *node.Node through the DefraJavaWrapper JNI bindings,
// embedded in this process (see doc.go). It mirrors cbindings.CWrapper's behaviour, but every call passes through a
// Java DefraNode object instead of calling cbindings' exported C functions directly.
type Wrapper struct {
	node    *node.Node
	handle  uintptr
	nodeObj C.jobject

	// nodeMu guards nodeObj/closed against Close deleting the JNI global ref while a
	// subscription-polling goroutine (which outlives the call that started it) is still using it.
	nodeMu sync.RWMutex
	closed bool // set by the first call to Close; guards against reusing the deleted JNI ref
}

// NewWrapper wraps an already-constructed *node.Node, reusing its cgo.Handle the same way cbindings.NewCWrapper does
// for the C client, but obtains a DefraNode Java object for it via DefraNode's package-private constructor rather
// than the public constructor (which would start a brand new node.)
func NewWrapper(n *node.Node) (*Wrapper, error) {
	h := cgo.NewHandle(n)
	obj, err := newNodeObject(uintptr(h))
	if err != nil {
		h.Delete()
		return nil, err
	}
	return &Wrapper{node: n, handle: uintptr(h), nodeObj: obj}, nil
}

func identityHandle(opt immutable.Option[identity.Identity]) uintptr {
	if !opt.HasValue() {
		return 0
	}
	return uintptr(cgo.NewHandle(opt.Value()))
}

func freeIdentityHandle(h uintptr) {
	if h != 0 {
		cgo.Handle(h).Delete()
	}
}

func unmarshalResult[T any](value string) (T, error) {
	var result T
	if err := json.Unmarshal([]byte(value), &result); err != nil {
		var zero T
		return zero, fmt.Errorf(errFmtUnmarshalResult, value, err)
	}
	return result, nil
}

// getNodeOrTxnHandle mirrors cbindings' helper of the same name: if a *Txn is attached to ctx, its handle is
// used instead of the node's.
func getNodeOrTxnHandle(nodeHandle uintptr, ctx context.Context) uintptr {
	txn, hadTxn := datastore.CtxTryGetTxn(ctx)
	if !hadTxn {
		return nodeHandle
	}
	if t, ok := txn.(*Txn); ok {
		return t.handle
	}
	return nodeHandle
}

func (w *Wrapper) PeerInfo(ctx context.Context, opts ...options.Enumerable[options.PeerInfoOptions]) ([]string, error) {
	idH := identityHandle(utils.NewOptions(opts...).GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callNode(w.nodeObj, "GetP2PInfoNative", w.handle, newArgs().argLong(idH))
	if err != nil {
		return nil, err
	}
	if err := res.asError(); err != nil {
		return nil, err
	}
	return unmarshalResult[[]string](res.Value)
}

func (w *Wrapper) ActivePeers(
	ctx context.Context, opts ...options.Enumerable[options.ActivePeersOptions],
) ([]string, error) {
	idH := identityHandle(utils.NewOptions(opts...).GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callNode(w.nodeObj, "ListP2PActivePeersNative", w.handle, newArgs().argLong(idH))
	if err != nil {
		return nil, err
	}
	if err := res.asError(); err != nil {
		return nil, err
	}
	return unmarshalResult[[]string](res.Value)
}

func (w *Wrapper) AddReplicator(
	ctx context.Context, addresses []string, opts ...options.Enumerable[options.AddReplicatorOptions],
) error {
	opt := utils.NewOptions(opts...)
	idH := identityHandle(opt.GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callNode(w.nodeObj, "AddP2PReplicatorNative", w.handle,
		newArgs().argStr(strings.Join(opt.CollectionNames, ",")).argStr(strings.Join(addresses, ",")).argLong(idH))
	if err != nil {
		return err
	}
	return res.asError()
}

func (w *Wrapper) DeleteReplicator(
	ctx context.Context, id string, opts ...options.Enumerable[options.DeleteReplicatorOptions],
) error {
	opt := utils.NewOptions(opts...)
	idH := identityHandle(opt.GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callNode(w.nodeObj, "DeleteP2PReplicatorNative", w.handle,
		newArgs().argStr(strings.Join(opt.CollectionNames, ",")).argStr(id).argLong(idH))
	if err != nil {
		return err
	}
	return res.asError()
}

func (w *Wrapper) ListReplicators(
	ctx context.Context, opts ...options.Enumerable[options.ListReplicatorsOptions],
) ([]client.Replicator, error) {
	idH := identityHandle(utils.NewOptions(opts...).GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callNode(w.nodeObj, "ListP2PReplicatorsNative", w.handle, newArgs().argLong(idH))
	if err != nil {
		return nil, err
	}
	if err := res.asError(); err != nil {
		return nil, err
	}
	return unmarshalResult[[]client.Replicator](res.Value)
}

func (w *Wrapper) AddP2PCollections(
	ctx context.Context, collectionIDs []string, opts ...options.Enumerable[options.AddP2PCollectionsOptions],
) error {
	idH := identityHandle(utils.NewOptions(opts...).GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callNode(w.nodeObj, "AddP2PCollectionNative", w.handle,
		newArgs().argStr(strings.Join(collectionIDs, ",")).argLong(idH))
	if err != nil {
		return err
	}
	return res.asError()
}

func (w *Wrapper) DeleteP2PCollections(
	ctx context.Context, collectionIDs []string, opts ...options.Enumerable[options.DeleteP2PCollectionsOptions],
) error {
	idH := identityHandle(utils.NewOptions(opts...).GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callNode(w.nodeObj, "DeleteP2PCollectionNative", w.handle,
		newArgs().argStr(strings.Join(collectionIDs, ",")).argLong(idH))
	if err != nil {
		return err
	}
	return res.asError()
}

func (w *Wrapper) ListP2PCollections(
	ctx context.Context, opts ...options.Enumerable[options.ListP2PCollectionsOptions],
) ([]string, error) {
	idH := identityHandle(utils.NewOptions(opts...).GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callNode(w.nodeObj, "ListP2PCollectionsNative", w.handle, newArgs().argLong(idH))
	if err != nil {
		return nil, err
	}
	if err := res.asError(); err != nil {
		return nil, err
	}
	return unmarshalResult[[]string](res.Value)
}

func (w *Wrapper) AddP2PDocuments(
	ctx context.Context, docIDs []string, opts ...options.Enumerable[options.AddP2PDocumentsOptions],
) error {
	idH := identityHandle(utils.NewOptions(opts...).GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callNode(w.nodeObj, "AddP2PDocumentNative", w.handle,
		newArgs().argStr(strings.Join(docIDs, ",")).argLong(idH))
	if err != nil {
		return err
	}
	return res.asError()
}

func (w *Wrapper) DeleteP2PDocuments(
	ctx context.Context, docIDs []string, opts ...options.Enumerable[options.DeleteP2PDocumentsOptions],
) error {
	idH := identityHandle(utils.NewOptions(opts...).GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callNode(w.nodeObj, "DeleteP2PDocumentNative", w.handle,
		newArgs().argStr(strings.Join(docIDs, ",")).argLong(idH))
	if err != nil {
		return err
	}
	return res.asError()
}

func (w *Wrapper) ListP2PDocuments(
	ctx context.Context, opts ...options.Enumerable[options.ListP2PDocumentsOptions],
) ([]string, error) {
	idH := identityHandle(utils.NewOptions(opts...).GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callNode(w.nodeObj, "ListP2PDocumentsNative", w.handle, newArgs().argLong(idH))
	if err != nil {
		return nil, err
	}
	if err := res.asError(); err != nil {
		return nil, err
	}
	return unmarshalResult[[]string](res.Value)
}

func ctxTimeoutString(ctx context.Context) string {
	deadline, hasDeadline := ctx.Deadline()
	if !hasDeadline {
		return ""
	}
	return time.Until(deadline).String()
}

func (w *Wrapper) SyncDocuments(
	ctx context.Context,
	collectionName string,
	docIDs []string,
	opts ...options.Enumerable[options.SyncDocumentsOptions],
) error {
	opt := utils.NewOptions(opts...)
	idH := identityHandle(opt.GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callNode(w.nodeObj, "SyncP2PDocumentsNative", w.handle,
		newArgs().argStr(collectionName).argStr(strings.Join(docIDs, ",")).argStr(ctxTimeoutString(ctx)).argLong(idH))
	if err != nil {
		return err
	}
	return res.asError()
}

func (w *Wrapper) SyncCollectionVersions(
	ctx context.Context, versionIDs []string, opts ...options.Enumerable[options.SyncCollectionVersionsOptions],
) error {
	opt := utils.NewOptions(opts...)
	idH := identityHandle(opt.GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callNode(w.nodeObj, "SyncP2PCollectionVersionsNative", w.handle,
		newArgs().argStr(strings.Join(versionIDs, ",")).argStr(ctxTimeoutString(ctx)).argLong(idH))
	if err != nil {
		return err
	}
	return res.asError()
}

func (w *Wrapper) SyncBranchableCollection(
	ctx context.Context, collectionID string, opts ...options.Enumerable[options.SyncBranchableCollectionOptions],
) error {
	opt := utils.NewOptions(opts...)
	idH := identityHandle(opt.GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callNode(w.nodeObj, "SyncP2PBranchableCollectionNative", w.handle,
		newArgs().argStr(collectionID).argStr(ctxTimeoutString(ctx)).argLong(idH))
	if err != nil {
		return err
	}
	return res.asError()
}

func (w *Wrapper) BasicImport(ctx context.Context, filepath string) error {
	panic("not implemented")
}

func (w *Wrapper) BasicExport(
	ctx context.Context, filepath string, opts ...options.Enumerable[options.BasicExportOptions],
) error {
	panic("not implemented")
}

func (w *Wrapper) AddCollection(
	ctx context.Context, sdl string, opts ...options.Enumerable[options.AddCollectionOptions],
) ([]client.CollectionVersion, error) {
	var txn datastore.Txn
	gotTxn, hadTxn := datastore.CtxTryGetTxn(ctx)
	if hadTxn {
		txn = gotTxn
	} else {
		clientTxn, err := w.NewTxn(false)
		if err != nil {
			return nil, err
		}
		var ok bool
		txn, ok = clientTxn.(datastore.Txn)
		if !ok {
			return nil, errors.New(errCastClientTxnFailed)
		}
		defer txn.Discard()
	}
	ctx = datastore.CtxSetTxn(ctx, txn)

	idH := identityHandle(utils.NewOptions(opts...).GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callStore(w, ctx, "AddCollectionNative", newArgs().argStr(sdl).argLong(idH))
	if err != nil {
		return nil, err
	}
	if err := res.asError(); err != nil {
		return nil, err
	}

	collectionVersions, err := unmarshalResult[[]client.CollectionVersion](res.Value)
	if err != nil {
		return nil, err
	}

	if !hadTxn {
		if err := txn.Commit(); err != nil {
			return nil, err
		}
	}

	return collectionVersions, nil
}

func (w *Wrapper) AddDACPolicy(
	ctx context.Context, policy string, opts ...options.Enumerable[options.AddDACPolicyOptions],
) (client.AddPolicyResult, error) {
	idH := identityHandle(utils.NewOptions(opts...).GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callStore(w, ctx, "ACPAddDACPolicyNative", newArgs().argLong(idH).argStr(policy))
	if err != nil {
		return client.AddPolicyResult{}, err
	}
	if err := res.asError(); err != nil {
		return client.AddPolicyResult{}, err
	}
	return unmarshalResult[client.AddPolicyResult](res.Value)
}

func (w *Wrapper) AddDACActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
	opts ...options.Enumerable[options.AddDACActorRelationshipOptions],
) (client.AddActorRelationshipResult, error) {
	idH := identityHandle(utils.NewOptions(opts...).GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callStore(w, ctx, "ACPAddDACActorRelationshipNative",
		newArgs().argLong(idH).argStr(collectionName).argStr(docID).argStr(relation).argStr(targetActor))
	if err != nil {
		return client.AddActorRelationshipResult{}, err
	}
	if err := res.asError(); err != nil {
		return client.AddActorRelationshipResult{}, err
	}
	return unmarshalResult[client.AddActorRelationshipResult](res.Value)
}

func (w *Wrapper) DeleteDACActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
	opts ...options.Enumerable[options.DeleteDACActorRelationshipOptions],
) (client.DeleteActorRelationshipResult, error) {
	idH := identityHandle(utils.NewOptions(opts...).GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callStore(w, ctx, "ACPDeleteDACActorRelationshipNative",
		newArgs().argLong(idH).argStr(collectionName).argStr(docID).argStr(relation).argStr(targetActor))
	if err != nil {
		return client.DeleteActorRelationshipResult{}, err
	}
	if err := res.asError(); err != nil {
		return client.DeleteActorRelationshipResult{}, err
	}
	return unmarshalResult[client.DeleteActorRelationshipResult](res.Value)
}

func (w *Wrapper) GetNACStatus(
	ctx context.Context, opts ...options.Enumerable[options.GetNACStatusOptions],
) (client.NACStatusResult, error) {
	idH := identityHandle(utils.NewOptions(opts...).GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callStore(w, ctx, "ACPGetNACStatusNative", newArgs().argLong(idH))
	if err != nil {
		return client.NACStatusResult{}, err
	}
	if err := res.asError(); err != nil {
		return client.NACStatusResult{}, err
	}
	return unmarshalResult[client.NACStatusResult](res.Value)
}

func (w *Wrapper) ReEnableNAC(ctx context.Context, opts ...options.Enumerable[options.ReEnableNACOptions]) error {
	idH := identityHandle(utils.NewOptions(opts...).GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callStore(w, ctx, "ACPReEnableNACNative", newArgs().argLong(idH))
	if err != nil {
		return err
	}
	return res.asError()
}

func (w *Wrapper) DisableNAC(ctx context.Context, opts ...options.Enumerable[options.DisableNACOptions]) error {
	idH := identityHandle(utils.NewOptions(opts...).GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callStore(w, ctx, "ACPDisableNACNative", newArgs().argLong(idH))
	if err != nil {
		return err
	}
	return res.asError()
}

func (w *Wrapper) AddNACActorRelationship(
	ctx context.Context,
	relation string,
	targetActor string,
	opts ...options.Enumerable[options.AddNACActorRelationshipOptions],
) (client.AddActorRelationshipResult, error) {
	idH := identityHandle(utils.NewOptions(opts...).GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callStore(w, ctx, "ACPAddNACActorRelationshipNative",
		newArgs().argLong(idH).argStr(relation).argStr(targetActor))
	if err != nil {
		return client.AddActorRelationshipResult{}, err
	}
	if err := res.asError(); err != nil {
		return client.AddActorRelationshipResult{}, err
	}
	return unmarshalResult[client.AddActorRelationshipResult](res.Value)
}

func (w *Wrapper) DeleteNACActorRelationship(
	ctx context.Context,
	relation string,
	targetActor string,
	opts ...options.Enumerable[options.DeleteNACActorRelationshipOptions],
) (client.DeleteActorRelationshipResult, error) {
	idH := identityHandle(utils.NewOptions(opts...).GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callStore(w, ctx, "ACPDeleteNACActorRelationshipNative",
		newArgs().argLong(idH).argStr(relation).argStr(targetActor))
	if err != nil {
		return client.DeleteActorRelationshipResult{}, err
	}
	if err := res.asError(); err != nil {
		return client.DeleteActorRelationshipResult{}, err
	}
	return unmarshalResult[client.DeleteActorRelationshipResult](res.Value)
}

func (w *Wrapper) PatchCollection(
	ctx context.Context,
	patch string,
	migration immutable.Option[model.Lens],
	opts ...options.Enumerable[options.PatchCollectionOptions],
) error {
	idH := identityHandle(utils.NewOptions(opts...).GetIdentity())
	defer freeIdentityHandle(idH)

	migrationStr := ""
	if migration.HasValue() {
		b, err := json.Marshal(migration.Value())
		if err != nil {
			return err
		}
		migrationStr = string(b)
	}

	res, err := callStore(w, ctx, "PatchCollectionNative", newArgs().argStr(patch).argStr(migrationStr).argLong(idH))
	if err != nil {
		return err
	}
	return res.asError()
}

func (w *Wrapper) DeleteCollection(
	ctx context.Context, names []string, opts ...options.Enumerable[options.DeleteCollectionOptions],
) error {
	opt := utils.NewOptions(opts...)
	idH := identityHandle(opt.GetIdentity())
	defer freeIdentityHandle(idH)

	activeOnly := 0
	if opt.ActiveOnly {
		activeOnly = 1
	}

	handle := getNodeOrTxnHandle(w.handle, ctx)
	res, err := callNode(w.nodeObj, "DeleteCollectionNative", handle,
		newArgs().argStr(strings.Join(names, ",")).argInt(activeOnly).argLong(idH))
	if err != nil {
		return err
	}
	return res.asError()
}

func (w *Wrapper) SetActiveCollectionVersion(
	ctx context.Context, collectionVersionID string, opts ...options.Enumerable[options.SetActiveCollectionVersionOptions],
) error {
	idH := identityHandle(utils.NewOptions(opts...).GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callStore(w, ctx, "SetActiveCollectionNative",
		newArgs().collOpts("", collectionVersionID, "", false, immutable.None[bool]()).argLong(idH))
	if err != nil {
		return err
	}
	return res.asError()
}

func (w *Wrapper) AddView(
	ctx context.Context, query string, sdl string, opts ...options.Enumerable[options.AddViewOptions],
) ([]client.CollectionVersion, error) {
	opt := utils.NewOptions(opts...)
	transformCID := ""
	if opt.TransformCID.HasValue() {
		transformCID = opt.TransformCID.Value()
	}
	idH := identityHandle(opt.GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callStore(w, ctx, "AddViewNative",
		newArgs().argStr(query).argStr(sdl).argStr(transformCID).argLong(idH))
	if err != nil {
		return []client.CollectionVersion{}, err
	}
	if err := res.asError(); err != nil {
		return []client.CollectionVersion{}, err
	}
	return unmarshalResult[[]client.CollectionVersion](res.Value)
}

func (w *Wrapper) RefreshViews(ctx context.Context, opts ...options.Enumerable[options.RefreshViewsOptions]) error {
	opt := utils.NewOptions(opts...)
	name, version, collectionID := collectionOptionsStrings(opt)
	getInactive := opt.GetInactive.HasValue() && opt.GetInactive.Value()
	idH := identityHandle(opt.GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callStore(w, ctx, "RefreshViewNative",
		newArgs().collOpts(name, version, collectionID, getInactive, immutable.None[bool]()).argLong(idH))
	if err != nil {
		return err
	}
	return res.asError()
}

func (w *Wrapper) ListActions(
	ctx context.Context, opts ...options.Enumerable[options.ListActionsOptions],
) ([]client.ActionExecution, error) {
	idH := identityHandle(utils.NewOptions(opts...).GetIdentity())
	defer freeIdentityHandle(idH)

	handle := getNodeOrTxnHandle(w.handle, ctx)
	res, err := callNode(w.nodeObj, "ListActionsNative", handle, newArgs().argLong(idH))
	if err != nil {
		return nil, err
	}
	if err := res.asError(); err != nil {
		return nil, err
	}
	return unmarshalResult[[]client.ActionExecution](res.Value)
}

func (w *Wrapper) SetMigration(
	ctx context.Context, config client.LensConfig, opts ...options.Enumerable[options.SetMigrationOptions],
) (string, error) {
	lensConfig, err := json.Marshal(config.Lens)
	if err != nil {
		return "", err
	}
	idH := identityHandle(utils.NewOptions(opts...).GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callStore(w, ctx, "SetLensNative",
		newArgs().argLong(idH).argStr(config.SourceCollectionVersionID).argStr(config.DestinationCollectionVersionID).argStr(string(lensConfig)))
	if err != nil {
		return "", err
	}
	if err := res.asError(); err != nil {
		return "", err
	}
	return res.Value, nil
}

func (w *Wrapper) AddLens(
	ctx context.Context, lens model.Lens, opts ...options.Enumerable[options.AddLensOptions],
) (string, error) {
	lensConfig, err := json.Marshal(lens)
	if err != nil {
		return "", err
	}
	idH := identityHandle(utils.NewOptions(opts...).GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callStore(w, ctx, "AddLensNative", newArgs().argLong(idH).argStr(string(lensConfig)))
	if err != nil {
		return "", err
	}
	if err := res.asError(); err != nil {
		return "", err
	}
	return res.Value, nil
}

func (w *Wrapper) ListLenses(
	ctx context.Context, opts ...options.Enumerable[options.ListLensesOptions],
) (map[string]model.Lens, error) {
	idH := identityHandle(utils.NewOptions(opts...).GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callStore(w, ctx, "ListLensesNative", newArgs().argLong(idH))
	if err != nil {
		return nil, err
	}
	if err := res.asError(); err != nil {
		return nil, err
	}
	var lenses map[string]model.Lens
	if err := json.Unmarshal([]byte(res.Value), &lenses); err != nil {
		snippet := res.Value
		if len(snippet) > 200 {
			snippet = snippet[:100] + "..." + snippet[len(snippet)-100:]
		}
		return nil, fmt.Errorf(errFmtListLensesUnmarshal, err, len(res.Value), snippet)
	}
	return lenses, nil
}

func (w *Wrapper) GetCollectionByName(
	ctx context.Context, name client.CollectionName, opts ...options.Enumerable[options.GetCollectionByNameOptions],
) (client.Collection, error) {
	opt := utils.NewOptions(opts...)
	getOpts := options.GetCollections().SetCollectionName(name)
	if opt.GetIdentity().HasValue() {
		getOpts.SetIdentity(opt.GetIdentity().Value())
	}
	cols, err := w.GetCollections(ctx, getOpts)
	if err != nil {
		return nil, err
	}
	if len(cols) == 0 {
		return nil, client.ErrCollectionNotFound
	}
	return cols[0], nil
}

func collectionOptionsStrings(opts *options.GetCollectionsOptions) (name, version, collectionID string) {
	if opts == nil {
		return "", "", ""
	}
	if opts.CollectionName.HasValue() {
		name = opts.CollectionName.Value()
	}
	if opts.VersionID.HasValue() {
		version = opts.VersionID.Value()
	}
	if opts.CollectionID.HasValue() {
		collectionID = opts.CollectionID.Value()
	}
	return name, version, collectionID
}

func (w *Wrapper) GetCollections(
	ctx context.Context, opts ...options.Enumerable[options.GetCollectionsOptions],
) ([]client.Collection, error) {
	opt := utils.NewOptions(opts...)
	name, version, collectionID := collectionOptionsStrings(opt)
	getInactive := opt.GetInactive.HasValue() && opt.GetInactive.Value()

	idH := identityHandle(opt.GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callStore(w, ctx, "DescribeCollectionNative",
		newArgs().collOpts(name, version, collectionID, getInactive, immutable.None[bool]()).argLong(idH))
	if err != nil {
		return nil, err
	}
	if err := res.asError(); err != nil {
		return []client.Collection{}, err
	}
	defs, err := unmarshalResult[[]client.CollectionVersion](res.Value)
	if err != nil {
		return nil, err
	}

	txnOpt := datastore.CtxTryGetTxnOption(ctx)

	cols := make([]client.Collection, len(defs))
	for i, def := range defs {
		cols[i] = &Collection{def: def, w: w, txn: txnOpt}
	}
	return cols, nil
}

func (w *Wrapper) ListIndexes(
	ctx context.Context, opts ...options.Enumerable[options.ListIndexesOptions],
) (map[client.CollectionName][]client.ListIndexesResult, error) {
	idH := identityHandle(utils.NewOptions(opts...).GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callStore(w, ctx, "ListIndexesNative",
		newArgs().collOpts("", "", "", false, immutable.None[bool]()).argLong(idH))
	if err != nil {
		return nil, err
	}
	if err := res.asError(); err != nil {
		return nil, err
	}
	return unmarshalResult[map[client.CollectionName][]client.ListIndexesResult](res.Value)
}

func (w *Wrapper) ListAllEncryptedIndexes(
	ctx context.Context, opts ...options.Enumerable[options.ListAllEncryptedIndexesOptions],
) (map[client.CollectionName][]client.EncryptedIndexDescription, error) {
	idH := identityHandle(utils.NewOptions(opts...).GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callStore(w, ctx, "ListEncryptedIndexesNative", newArgs().argStr("").argLong(idH))
	if err != nil {
		return nil, err
	}
	if err := res.asError(); err != nil {
		return nil, err
	}
	return unmarshalResult[map[client.CollectionName][]client.EncryptedIndexDescription](res.Value)
}

func extractStringsFromRequestOptions(opt *options.ExecRequestOptions) (string, string, error) {
	opName := ""
	if opt.OperationName.HasValue() {
		opName = opt.OperationName.Value()
	}
	varsJSON := ""
	if opt.Variables != nil {
		data, err := json.Marshal(opt.Variables)
		if err != nil {
			return "", "", err
		}
		varsJSON = string(data)
	}
	return opName, varsJSON, nil
}

func (w *Wrapper) ExecRequest(
	ctx context.Context, query string, opts ...options.Enumerable[options.ExecRequestOptions],
) *client.RequestResult {
	execOpts := utils.NewOptions(opts...)
	operation, variables, err := extractStringsFromRequestOptions(execOpts)
	if err != nil {
		return &client.RequestResult{GQL: client.GQLResult{Errors: []error{err}}}
	}

	idH := identityHandle(execOpts.GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callStore(w, ctx, "ExecuteQueryNative",
		newArgs().argStr(query).argLong(idH).argStr(operation).argStr(variables))
	if err != nil {
		return &client.RequestResult{GQL: client.GQLResult{Errors: []error{err}}}
	}

	if res.Status == 2 {
		return &client.RequestResult{Subscription: w.wrapSubscriptionAsChannel(ctx, res.Value)}
	}

	retval := &client.RequestResult{}
	if res.Status != 0 {
		retval.GQL.Errors = append(retval.GQL.Errors, client.ReviveError(res.Error))
		return retval
	}
	if err := json.Unmarshal([]byte(res.Value), &retval.GQL); err != nil {
		retval.GQL.Errors = append(retval.GQL.Errors, err)
	}
	return retval
}

// subscriptionPollInterval is how long wrapSubscriptionAsChannel waits between polls once it finds
// nothing to deliver. Unlike cbindings' identical polling loop, each poll here also attaches to the
// JVM, which is heavy enough (and adds to the same OS-thread churn covered in doc.go's signal-chaining
// notes) that busy-spinning isn't free the way it is for a plain cgo call.
const subscriptionPollInterval = 15 * time.Millisecond

// callNodeIfOpen calls a no-handle native method on nodeObj, guarding against Close having
// deleted (or concurrently deleting) the JNI global ref out from under a long-lived caller such
// as the subscription-polling goroutine. Returns open=false if the wrapper is closed, in which
// case res/err are not meaningful and nodeObj must not be touched.
func (w *Wrapper) callNodeIfOpen(name string, b *argBuilder) (res defraResult, err error, open bool) {
	w.nodeMu.RLock()
	defer w.nodeMu.RUnlock()
	if w.closed {
		return defraResult{}, nil, false
	}
	res, err = callNodeNoHandle(w.nodeObj, name, b)
	return res, err, true
}

// wrapSubscriptionAsChannel mirrors cbindings' helper of the same name, polling the subscription via
// PollSubscriptionNative in a loop until ctx is done. PollSubscriptionNative is an instance method on
// DefraNode, so it's invoked on this Wrapper's cached nodeObj.
//
// PollSubscriptionNative returns status 1 for two different situations: a transport/JNI-level
// problem, and the subscription having naturally ended (its result channel closed, or its ID no longer
// being recognised because that already happened on an earlier poll). Either way there is nothing to
// gain from polling the same ID again, so status 1 is treated as terminal rather than being retried
// forever like the "nothing new yet" case (status 2 / empty value).
func (w *Wrapper) wrapSubscriptionAsChannel(ctx context.Context, subID string) <-chan client.GQLResult {
	ch := make(chan client.GQLResult)
	go func() {
		defer close(ch)
		// Always tell the native side to release the subscription store entry (and cancel the
		// underlying DB-level subscription) once this goroutine stops polling it, however that
		// happens (ctx cancellation, a terminal poll result, or the wrapper being closed.) Without
		// this, every subscription that isn't explicitly closed by the caller leaks both the C-side
		// subscription store entry and the live DB subscription driving it.
		defer func() {
			_, _, _ = w.callNodeIfOpen("CloseSubscriptionNative", newArgs().argStr(subID))
		}()
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			res, err, open := w.callNodeIfOpen("PollSubscriptionNative", newArgs().argStr(subID))
			if !open {
				return
			}
			if err != nil || res.Status == 1 {
				return
			}
			if res.Status == 2 || res.Value == "" {
				select {
				case <-time.After(subscriptionPollInterval):
				case <-ctx.Done():
					return
				}
				continue
			}

			var gql client.GQLResult
			if err := json.Unmarshal([]byte(res.Value), &gql); err != nil {
				gql.Errors = append(gql.Errors, err)
			}
			select {
			case ch <- gql:
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch
}

func (w *Wrapper) NewTxn(readOnly bool) (client.Txn, error) {
	txnPtr, err := createTransactionWithHandle(w.nodeObj, w.handle, readOnly)
	if err != nil {
		return nil, err
	}
	txnObj, err := newTransactionObject(txnPtr)
	if err != nil {
		// The native transaction was created successfully, but we failed to construct its Java
		// wrapper object, so discard it via the handle.
		h := cgo.Handle(txnPtr)
		if dsTxn, ok := h.Value().(datastore.Txn); ok {
			dsTxn.Discard()
		}
		h.Delete()
		return nil, err
	}
	dsTxn := cgo.Handle(txnPtr).Value().(datastore.Txn) //nolint:forcetypeassert
	return &Txn{Wrapper: w, tx: dsTxn, handle: txnPtr, txnObj: txnObj}, nil
}

// Close closes the node. Safe to call more than once (mirroring how io.Closer implementations are
// commonly expected to behave) - a second call would otherwise reuse nodeObj's JNI global ref after
// the first call already deleted it.
//
// Holds nodeMu for the whole close sequence, so it can't run concurrently with a subscription-polling
// goroutine's use of nodeObj (see callNodeIfOpen). Without that, Close deleting the JNI global ref
// while a poll is in flight (or about to start) races with the poll's own use of nodeObj.
func (w *Wrapper) Close() {
	w.nodeMu.Lock()
	defer w.nodeMu.Unlock()
	if w.closed {
		return
	}
	w.closed = true
	_, _ = callNode(w.nodeObj, "NodeCloseNative", w.handle, newArgs())
	if env, detach, err := attach(); err == nil {
		C.defra_delete_global_ref(env, w.nodeObj)
		detach()
	}
}

func (w *Wrapper) Events() event.Bus {
	return w.node.DB.Events()
}

func (w *Wrapper) MaxTxnRetries() int {
	return w.node.DB.MaxTxnRetries()
}

func (w *Wrapper) PrintDump(ctx context.Context) error {
	panic("not implemented")
}

func (w *Wrapper) Connect(
	ctx context.Context, addresses []string, opts ...options.Enumerable[options.ConnectOptions],
) error {
	idH := identityHandle(utils.NewOptions(opts...).GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callStore(w, ctx, "ConnectP2PPeersNative",
		newArgs().argStr(strings.Join(addresses, ",")).argLong(idH))
	if err != nil {
		return err
	}
	return res.asError()
}

func (w *Wrapper) Disconnect(
	ctx context.Context, addresses []string, opts ...options.Enumerable[options.DisconnectOptions],
) error {
	idH := identityHandle(utils.NewOptions(opts...).GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callStore(w, ctx, "DisconnectP2PPeersNative",
		newArgs().argStr(strings.Join(addresses, ",")).argLong(idH))
	if err != nil {
		return err
	}
	return res.asError()
}

func (w *Wrapper) GetNodeIdentity(ctx context.Context) (immutable.Option[identity.PublicRawIdentity], error) {
	handle := getNodeOrTxnHandle(w.handle, ctx)
	res, err := callNode(w.nodeObj, "GetNodeIdentityNative", handle, newArgs())
	if err != nil {
		return immutable.None[identity.PublicRawIdentity](), err
	}
	if err := res.asError(); err != nil {
		return immutable.None[identity.PublicRawIdentity](), err
	}
	if res.Value == "Node has no identity assigned to it." {
		return immutable.None[identity.PublicRawIdentity](), nil
	}
	resVal, err := unmarshalResult[identity.PublicRawIdentity](res.Value)
	if err != nil {
		return immutable.None[identity.PublicRawIdentity](), err
	}
	return immutable.Some(resVal), nil
}

func (w *Wrapper) VerifySignature(
	ctx context.Context,
	blockCid string,
	pubKey crypto.PublicKey,
	opts ...options.Enumerable[options.VerifySignatureOptions],
) error {
	idH := identityHandle(utils.NewOptions(opts...).GetIdentity())
	defer freeIdentityHandle(idH)

	res, err := callStore(w, ctx, "VerifyBlockSignatureNative",
		newArgs().argStr(string(pubKey.Type())).argStr(pubKey.String()).argStr(blockCid).argLong(idH))
	if err != nil {
		return err
	}
	return res.asError()
}
