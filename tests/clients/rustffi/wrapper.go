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

//go:build rust_ffi

package rustffi

/*
#include <stdlib.h>
#include "defra.h"
*/
import "C"

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unsafe"

	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/utils"
)

// as runs call with ident registered for it.
func as(ident immutable.Option[identity.Identity], call func(C.uintptr_t) result) result {
	handle, err := identityHandle(ident)
	if err != nil {
		return result{Status: 1, Error: err.Error()}
	}
	defer freeIdentity(handle)
	return call(handle)
}

// timeout is the time left before ctx's deadline, as the C API takes it.
func timeout(ctx context.Context) string {
	if deadline, ok := ctx.Deadline(); ok {
		return time.Until(deadline).String()
	}
	return ""
}

func (w *Wrapper) PeerInfo(
	ctx context.Context,
	opts ...options.Enumerable[options.PeerInfoOptions],
) ([]string, error) {
	res := as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_p2p_info(w.node, ident))
	})
	return unmarshalResult[[]string](res)
}

func (w *Wrapper) ActivePeers(
	ctx context.Context,
	opts ...options.Enumerable[options.ActivePeersOptions],
) ([]string, error) {
	res := as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_p2p_peer_list(w.node, ident))
	})
	return unmarshalResult[[]string](res)
}

func (w *Wrapper) AddReplicator(
	ctx context.Context,
	addresses []string,
	opts ...options.Enumerable[options.AddReplicatorOptions],
) error {
	var args cstrings
	defer args.free()
	opt := utils.NewOptions(opts...)
	return as(opt.GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_p2p_replicator_add(w.node,
			args.str(strings.Join(opt.CollectionNames, ",")),
			args.str(strings.Join(addresses, ",")),
			ident))
	}).err()
}

func (w *Wrapper) DeleteReplicator(
	ctx context.Context,
	id string,
	opts ...options.Enumerable[options.DeleteReplicatorOptions],
) error {
	var args cstrings
	defer args.free()
	opt := utils.NewOptions(opts...)
	return as(opt.GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_p2p_replicator_delete(w.node,
			args.str(strings.Join(opt.CollectionNames, ",")), args.str(id), ident))
	}).err()
}

func (w *Wrapper) ListReplicators(
	ctx context.Context,
	opts ...options.Enumerable[options.ListReplicatorsOptions],
) ([]client.Replicator, error) {
	res := as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_p2p_replicator_list(w.node, ident))
	})
	return unmarshalResult[[]client.Replicator](res)
}

func (w *Wrapper) AddP2PCollections(
	ctx context.Context,
	collectionIDs []string,
	opts ...options.Enumerable[options.AddP2PCollectionsOptions],
) error {
	var args cstrings
	defer args.free()
	return as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_p2p_collection_add(w.node, args.str(strings.Join(collectionIDs, ",")), ident))
	}).err()
}

func (w *Wrapper) DeleteP2PCollections(
	ctx context.Context,
	collectionIDs []string,
	opts ...options.Enumerable[options.DeleteP2PCollectionsOptions],
) error {
	var args cstrings
	defer args.free()
	return as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_p2p_collection_delete(w.node, args.str(strings.Join(collectionIDs, ",")), ident))
	}).err()
}

func (w *Wrapper) ListP2PCollections(
	ctx context.Context,
	opts ...options.Enumerable[options.ListP2PCollectionsOptions],
) ([]string, error) {
	res := as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_p2p_collection_list(w.node, ident))
	})
	return unmarshalResult[[]string](res)
}

func (w *Wrapper) AddP2PDocuments(
	ctx context.Context,
	docIDs []string,
	opts ...options.Enumerable[options.AddP2PDocumentsOptions],
) error {
	var args cstrings
	defer args.free()
	return as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_p2p_document_add(w.node, args.str(strings.Join(docIDs, ",")), ident))
	}).err()
}

func (w *Wrapper) DeleteP2PDocuments(
	ctx context.Context,
	docIDs []string,
	opts ...options.Enumerable[options.DeleteP2PDocumentsOptions],
) error {
	var args cstrings
	defer args.free()
	return as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_p2p_document_delete(w.node, args.str(strings.Join(docIDs, ",")), ident))
	}).err()
}

func (w *Wrapper) ListP2PDocuments(
	ctx context.Context,
	opts ...options.Enumerable[options.ListP2PDocumentsOptions],
) ([]string, error) {
	res := as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_p2p_document_list(w.node, ident))
	})
	return unmarshalResult[[]string](res)
}

func (w *Wrapper) SyncDocuments(
	ctx context.Context,
	collectionName string,
	docIDs []string,
	opts ...options.Enumerable[options.SyncDocumentsOptions],
) error {
	var args cstrings
	defer args.free()
	return as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_p2p_document_sync(w.node, args.str(collectionName),
			args.str(strings.Join(docIDs, ",")), args.str(timeout(ctx)), ident))
	}).err()
}

func (w *Wrapper) SyncCollectionVersions(
	ctx context.Context,
	versionIDs []string,
	opts ...options.Enumerable[options.SyncCollectionVersionsOptions],
) error {
	var args cstrings
	defer args.free()
	return as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_p2p_collection_sync_versions(w.node,
			args.str(strings.Join(versionIDs, ",")), args.str(timeout(ctx)), ident))
	}).err()
}

func (w *Wrapper) SyncBranchableCollection(
	ctx context.Context,
	collectionID string,
	opts ...options.Enumerable[options.SyncBranchableCollectionOptions],
) error {
	var args cstrings
	defer args.free()
	return as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_p2p_collection_sync_branchable(w.node,
			args.str(collectionID), args.str(timeout(ctx)), ident))
	}).err()
}

func (w *Wrapper) Connect(
	ctx context.Context,
	addresses []string,
	opts ...options.Enumerable[options.ConnectOptions],
) error {
	var args cstrings
	defer args.free()
	return as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_p2p_connect(w.target(ctx), args.str(strings.Join(addresses, ",")), ident))
	}).err()
}

func (w *Wrapper) Disconnect(
	ctx context.Context,
	addresses []string,
	opts ...options.Enumerable[options.DisconnectOptions],
) error {
	var args cstrings
	defer args.free()
	return as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_p2p_disconnect(w.target(ctx), args.str(strings.Join(addresses, ",")), ident))
	}).err()
}

func (w *Wrapper) BasicImport(ctx context.Context, filepath string) error {
	var args cstrings
	defer args.free()
	return take(C.defra_backup_import(w.target(ctx), args.str(filepath))).err()
}

func (w *Wrapper) BasicExport(
	ctx context.Context,
	filepath string,
	opts ...options.Enumerable[options.BasicExportOptions],
) error {
	var args cstrings
	defer args.free()
	opt := utils.NewOptions(opts...)
	pretty := C.int(0)
	if opt.Pretty {
		pretty = 1
	}
	return take(C.defra_backup_export(w.target(ctx), args.str(filepath),
		args.str(strings.Join(opt.Collections, ",")), args.str(opt.Format), pretty)).err()
}

func (w *Wrapper) AddCollection(
	ctx context.Context,
	sdl string,
	opts ...options.Enumerable[options.AddCollectionOptions],
) ([]client.CollectionVersion, error) {
	// Go adds the collections of one SDL in one transaction.
	if _, hadTxn := datastore.CtxTryGetClientTxn(ctx); !hadTxn {
		txn, err := w.NewTxn(false)
		if err != nil {
			return nil, err
		}
		defer txn.Discard()
		versions, err := w.AddCollection(datastore.CtxSetFromClientTxn(ctx, txn), sdl, opts...)
		if err != nil {
			return nil, err
		}
		return versions, txn.Commit()
	}

	var args cstrings
	defer args.free()
	res := as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_collection_add(w.target(ctx), args.str(sdl), ident))
	})
	return unmarshalResult[[]client.CollectionVersion](res)
}

func (w *Wrapper) AddDACPolicy(
	ctx context.Context,
	policy string,
	opts ...options.Enumerable[options.AddDACPolicyOptions],
) (client.AddPolicyResult, error) {
	var args cstrings
	defer args.free()
	res := as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_acp_dac_policy_add(w.target(ctx), ident, args.str(policy)))
	})
	return unmarshalResult[client.AddPolicyResult](res)
}

func (w *Wrapper) AddDACActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
	opts ...options.Enumerable[options.AddDACActorRelationshipOptions],
) (client.AddActorRelationshipResult, error) {
	var args cstrings
	defer args.free()
	res := as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_acp_dac_relationship_add(w.target(ctx), ident,
			args.str(collectionName), args.str(docID), args.str(relation), args.str(targetActor)))
	})
	return unmarshalResult[client.AddActorRelationshipResult](res)
}

func (w *Wrapper) DeleteDACActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
	opts ...options.Enumerable[options.DeleteDACActorRelationshipOptions],
) (client.DeleteActorRelationshipResult, error) {
	var args cstrings
	defer args.free()
	res := as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_acp_dac_relationship_delete(w.target(ctx), ident,
			args.str(collectionName), args.str(docID), args.str(relation), args.str(targetActor)))
	})
	return unmarshalResult[client.DeleteActorRelationshipResult](res)
}

func (w *Wrapper) GetNACStatus(
	ctx context.Context,
	opts ...options.Enumerable[options.GetNACStatusOptions],
) (client.NACStatusResult, error) {
	res := as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_acp_nac_status(w.target(ctx), ident))
	})
	return unmarshalResult[client.NACStatusResult](res)
}

func (w *Wrapper) ReEnableNAC(ctx context.Context, opts ...options.Enumerable[options.ReEnableNACOptions]) error {
	return as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_acp_nac_re_enable(w.target(ctx), ident))
	}).err()
}

func (w *Wrapper) DisableNAC(ctx context.Context, opts ...options.Enumerable[options.DisableNACOptions]) error {
	return as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_acp_nac_disable(w.target(ctx), ident))
	}).err()
}

func (w *Wrapper) AddNACActorRelationship(
	ctx context.Context,
	relation string,
	targetActor string,
	opts ...options.Enumerable[options.AddNACActorRelationshipOptions],
) (client.AddActorRelationshipResult, error) {
	var args cstrings
	defer args.free()
	res := as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_acp_nac_relationship_add(w.target(ctx), ident,
			args.str(relation), args.str(targetActor)))
	})
	return unmarshalResult[client.AddActorRelationshipResult](res)
}

func (w *Wrapper) DeleteNACActorRelationship(
	ctx context.Context,
	relation string,
	targetActor string,
	opts ...options.Enumerable[options.DeleteNACActorRelationshipOptions],
) (client.DeleteActorRelationshipResult, error) {
	var args cstrings
	defer args.free()
	res := as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_acp_nac_relationship_delete(w.target(ctx), ident,
			args.str(relation), args.str(targetActor)))
	})
	return unmarshalResult[client.DeleteActorRelationshipResult](res)
}

func (w *Wrapper) PatchCollection(
	ctx context.Context,
	patch string,
	migration immutable.Option[model.Lens],
	opts ...options.Enumerable[options.PatchCollectionOptions],
) error {
	migrationJSON, err := optionJSON(migration)
	if err != nil {
		return err
	}
	var args cstrings
	defer args.free()
	return as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_collection_patch(w.target(ctx), args.str(patch), args.str(migrationJSON), ident))
	}).err()
}

func (w *Wrapper) DeleteCollection(
	ctx context.Context,
	names []string,
	opts ...options.Enumerable[options.DeleteCollectionOptions],
) error {
	var args cstrings
	defer args.free()
	opt := utils.NewOptions(opts...)
	activeOnly := C.int(0)
	if opt.ActiveOnly {
		activeOnly = 1
	}
	return as(opt.GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_collection_delete(w.target(ctx), args.str(strings.Join(names, ",")), activeOnly, ident))
	}).err()
}

func (w *Wrapper) SetActiveCollectionVersion(
	ctx context.Context,
	collectionVersionID string,
	opts ...options.Enumerable[options.SetActiveCollectionVersionOptions],
) error {
	var args cstrings
	defer args.free()
	copts := collectionOptions{Version: collectionVersionID}.c(&args)
	return as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_collection_set_active(w.target(ctx), copts, ident))
	}).err()
}

func (w *Wrapper) AddView(
	ctx context.Context,
	query string,
	sdl string,
	opts ...options.Enumerable[options.AddViewOptions],
) ([]client.CollectionVersion, error) {
	var args cstrings
	defer args.free()
	opt := utils.NewOptions(opts...)
	var transformCID *C.char
	if opt.TransformCID.HasValue() {
		transformCID = args.str(opt.TransformCID.Value())
	}
	res := as(opt.GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_view_add(w.target(ctx), args.str(query), args.str(sdl), transformCID, ident))
	})
	return unmarshalResult[[]client.CollectionVersion](res)
}

func (w *Wrapper) RefreshViews(ctx context.Context, opts ...options.Enumerable[options.RefreshViewsOptions]) error {
	var args cstrings
	defer args.free()
	opt := utils.NewOptions(opts...)
	copts := getCollectionsOptions(opt).c(&args)
	return as(opt.GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_view_refresh(w.target(ctx), copts, ident))
	}).err()
}

func (w *Wrapper) ListActions(
	ctx context.Context,
	opts ...options.Enumerable[options.ListActionsOptions],
) ([]client.ActionExecution, error) {
	res := as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_action_list(w.target(ctx), ident))
	})
	return unmarshalResult[[]client.ActionExecution](res)
}

func (w *Wrapper) SetMigration(
	ctx context.Context, config client.LensConfig, opts ...options.Enumerable[options.SetMigrationOptions],
) (string, error) {
	lens, err := json.Marshal(config.Lens)
	if err != nil {
		return "", err
	}
	var args cstrings
	defer args.free()
	res := as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_lens_set(w.target(ctx), ident, args.str(config.SourceCollectionVersionID),
			args.str(config.DestinationCollectionVersionID), args.str(string(lens))))
	})
	return res.Value, res.err()
}

func (w *Wrapper) AddLens(
	ctx context.Context,
	lens model.Lens,
	opts ...options.Enumerable[options.AddLensOptions],
) (string, error) {
	lensJSON, err := json.Marshal(lens)
	if err != nil {
		return "", err
	}
	var args cstrings
	defer args.free()
	res := as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_lens_add(w.target(ctx), ident, args.str(string(lensJSON))))
	})
	return res.Value, res.err()
}

func (w *Wrapper) ListLenses(
	ctx context.Context,
	opts ...options.Enumerable[options.ListLensesOptions],
) (map[string]model.Lens, error) {
	res := as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_lens_list(w.target(ctx), ident))
	})
	return unmarshalResult[map[string]model.Lens](res)
}

func (w *Wrapper) GetCollectionByName(
	ctx context.Context,
	name client.CollectionName,
	opts ...options.Enumerable[options.GetCollectionByNameOptions],
) (client.Collection, error) {
	getOpts := options.GetCollections().SetCollectionName(name)
	if ident := utils.NewOptions(opts...).GetIdentity(); ident.HasValue() {
		getOpts.SetIdentity(ident.Value())
	}
	cols, err := w.GetCollections(ctx, getOpts)
	if err != nil {
		return nil, err
	}
	if len(cols) == 0 {
		return nil, fmt.Errorf("collection with name %q not found", name)
	}
	return cols[0], nil
}

func (w *Wrapper) GetCollections(
	ctx context.Context,
	opts ...options.Enumerable[options.GetCollectionsOptions],
) ([]client.Collection, error) {
	var args cstrings
	defer args.free()
	opt := utils.NewOptions(opts...)
	copts := getCollectionsOptions(opt).c(&args)
	res := as(opt.GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_collection_describe(w.target(ctx), copts, ident))
	})
	versions, err := unmarshalResult[[]client.CollectionVersion](res)
	if err != nil {
		return nil, err
	}

	cols := make([]client.Collection, len(versions))
	for i, version := range versions {
		cols[i] = &Collection{def: version, w: w}
	}
	return cols, nil
}

func (w *Wrapper) ListIndexes(
	ctx context.Context,
	opts ...options.Enumerable[options.ListIndexesOptions],
) (map[client.CollectionName][]client.ListIndexesResult, error) {
	var args cstrings
	defer args.free()
	copts := collectionOptions{}.c(&args)
	res := as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_index_list(w.target(ctx), copts, ident))
	})
	return unmarshalResult[map[client.CollectionName][]client.ListIndexesResult](res)
}

func (w *Wrapper) ListAllEncryptedIndexes(
	ctx context.Context,
	opts ...options.Enumerable[options.ListAllEncryptedIndexesOptions],
) (map[client.CollectionName][]client.EncryptedIndexDescription, error) {
	var args cstrings
	defer args.free()
	res := as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_encrypted_index_list(w.target(ctx), args.str(""), ident))
	})
	return unmarshalResult[map[client.CollectionName][]client.EncryptedIndexDescription](res)
}

func (w *Wrapper) ExecRequest(
	ctx context.Context,
	query string,
	opts ...options.Enumerable[options.ExecRequestOptions],
) *client.RequestResult {
	opt := utils.NewOptions(opts...)
	operation, variables, err := requestStrings(opt)
	if err != nil {
		return &client.RequestResult{GQL: client.GQLResult{Errors: []error{err}}}
	}
	var args cstrings
	defer args.free()
	res := as(opt.GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_query_execute(w.target(ctx), args.str(query), ident,
			args.str(operation), args.str(variables)))
	})
	if res.Status == subscriptionPending {
		return &client.RequestResult{Subscription: subscriptionChannel(ctx, res.Value)}
	}
	return &client.RequestResult{GQL: gqlResult(res)}
}

// subscriptionChannel delivers the results of GraphQL subscription id until
// it ends or ctx is done, and then closes it.
func subscriptionChannel(ctx context.Context, id string) <-chan client.GQLResult {
	ch := make(chan client.GQLResult)
	go func() {
		defer close(ch)
		cID := C.CString(id)
		defer C.free(unsafe.Pointer(cID))
		defer func() { _ = take(C.defra_subscription_close(cID)) }()
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			res := take(C.defra_subscription_poll(cID))
			switch res.Status {
			case 0:
				select {
				case ch <- gqlResult(res):
				case <-ctx.Done():
					return
				}
			case subscriptionPending:
				time.Sleep(pollInterval)
			default:
				return
			}
		}
	}()
	return ch
}

func (w *Wrapper) NewTxn(readOnly bool) (client.Txn, error) {
	readOnlyFlag := C.int(0)
	if readOnly {
		readOnlyFlag = 1
	}
	cResult := C.defra_transaction_create(w.node, readOnlyFlag)
	handle := cResult.txn_ptr
	message := C.GoString(cResult.error)
	failed := cResult.status != 0
	C.defra_new_txn_result_free(&cResult)
	if failed {
		return nil, fmt.Errorf("%s", message)
	}
	return &Transaction{Wrapper: w, handle: handle, startTS: time.Now()}, nil
}

func (w *Wrapper) NewConcurrentTxn(readOnly bool) (client.Txn, error) {
	return w.NewTxn(readOnly)
}

func (w *Wrapper) PrintDump(ctx context.Context) error {
	return take(C.defra_print_dump(w.target(ctx))).err()
}

func (w *Wrapper) GetNodeIdentity(ctx context.Context) (immutable.Option[identity.PublicRawIdentity], error) {
	res := take(C.defra_node_identity(w.target(ctx)))
	if err := res.err(); err != nil {
		return immutable.None[identity.PublicRawIdentity](), err
	}
	if res.Value == "Node has no identity assigned to it." {
		return immutable.None[identity.PublicRawIdentity](), nil
	}
	raw, err := unmarshalResult[identity.PublicRawIdentity](res)
	if err != nil {
		return immutable.None[identity.PublicRawIdentity](), err
	}
	return immutable.Some(raw), nil
}

func (w *Wrapper) VerifySignature(
	ctx context.Context,
	blockCid string,
	pubKey crypto.PublicKey,
	opts ...options.Enumerable[options.VerifySignatureOptions],
) error {
	var args cstrings
	defer args.free()
	return as(utils.NewOptions(opts...).GetIdentity(), func(ident C.uintptr_t) result {
		return take(C.defra_block_verify(w.target(ctx), args.str(string(pubKey.Type())),
			args.str(pubKey.String()), args.str(blockCid), ident))
	}).err()
}
