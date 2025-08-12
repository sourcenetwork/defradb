// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cbindings

/*
#include <stdlib.h>
#include "defra_structs.h"

extern Result* ACPAddDACPolicy(uintptr_t nodePtr, char* identity, char* policy);
extern Result* ACPAddDACActorRelationship(uintptr_t nodePtr, char* identity, char* collection, char* docID, char* relation, char* actor);
extern Result* ACPDeleteDACActorRelationship(uintptr_t nodePtr, char* identity, char* collection, char* docID, char* relation, char* actor);
extern Result* ACPDisableNAC(uintptr_t nodePtr, char* identity);
extern Result* ACPReEnableNAC(uintptr_t nodePtr, char* identity);
extern Result* ACPAddNACActorRelationship(uintptr_t nodePtr, char* identity, char* relation, char* actor);
extern Result* ACPDeleteNACActorRelationship(uintptr_t nodePtr, char* identity, char* relation, char* actor);
extern Result* ACPGetNACStatus(uintptr_t nodePtr, char* identity);
extern Result* BlockVerifySignature(uintptr_t nodePtr, char* keyType, char* publicKey, char* cid);
extern Result* CollectionCreate(uintptr_t nodePtr, char* json, int isEncrypted, char* encryptedFields, CollectionOptions options);
extern Result* CollectionDelete(uintptr_t nodePtr, char* docIDStr, char* filterStr, CollectionOptions options);
extern Result* CollectionDescribe(uintptr_t nodePtr, CollectionOptions options);
extern Result* CollectionListDocIDs(uintptr_t nodePtr, CollectionOptions options);
extern Result* CollectionGet(uintptr_t nodePtr, char* docIDStr, int showDeleted, CollectionOptions options);
extern Result* CollectionPatch(uintptr_t nodePtr, char* patch, char* lensConfig, CollectionOptions options);
extern Result* CollectionUpdate(uintptr_t nodePtr, char* docIDStr, char* filterStr, char* updaterStr, CollectionOptions options);
extern Result* IdentityNew(char* keyType);
extern Result* NodeIdentity(uintptr_t nodePtr);
extern Result* IndexCreate(uintptr_t nodePtr, char* collectionName, char* indexName, char* fieldsStr, int isUnique);
extern Result* IndexList(uintptr_t nodePtr, char* collectionName);
extern Result* IndexDrop(uintptr_t nodePtr, char* collectionName, char* indexName);
extern Result* LensSet(uintptr_t nodePtr, char* src, char* dst, char* cfg);
extern Result* LensDown(uintptr_t nodePtr, char* collectionID, char* documents);
extern Result* LensUp(uintptr_t nodePtr, char* collectionID, char* documents);
extern Result* LensReload(uintptr_t nodePtr);
extern Result* LensSetRegistry(uintptr_t nodePtr, char* collectionID, char* cfg);
extern NewNodeResult NewNode(NodeInitOptions cOptions);
extern Result* NodeClose(uintptr_t nodePtr);
extern Result* P2PInfo(uintptr_t nodePtr);
extern Result* P2PgetAllReplicators(uintptr_t nodePtr);
extern Result* P2PsetReplicator(uintptr_t nodePtr, char* collections, char* peerInfo);
extern Result* P2PdeleteReplicator(uintptr_t nodePtr, char* collections, char* peerInfo);
extern Result* P2PcollectionAdd(uintptr_t nodePtr, char* collections);
extern Result* P2PcollectionRemove(uintptr_t nodePtr, char* collections);
extern Result* P2PcollectionGetAll(uintptr_t nodePtr);
extern Result* P2PdocumentAdd(uintptr_t nodePtr, char* collections);
extern Result* P2PdocumentRemove(uintptr_t nodePtr, char* collections);
extern Result* P2PdocumentGetAll(uintptr_t nodePtr);
extern Result* P2PdocumentSync(uintptr_t nodePtr, char* collection, char* docIDs, char* timeoutStr);
extern Result* PollSubscription(char* id);
extern Result* CloseSubscription(char* id);
extern Result* ExecuteQuery(uintptr_t nodePtr, char* query, char* identity, char* operationName, char* variables);
extern Result* AddSchema(uintptr_t nodePtr, char* schema);
extern Result* SetActiveCollection(uintptr_t nodePtr, char* version);
extern NewTxnResult TransactionCreate(uintptr_t nodePtr, int isConcurrent, int isReadOnly);
extern Result* TransactionCommit(uintptr_t txnPtr);
extern void TransactionDiscard(uintptr_t txnPtr);
extern Result* VersionGet(int flagFull, int flagJSON);
extern Result* ViewAdd(uintptr_t nodePtr, char* query, char* sdl, char* transformStr);
extern Result* ViewRefresh(uintptr_t nodePtr, char* viewNameStr, char* collectionIDStr, char* versionIDStr, int getInactive);
*/
import "C"
import (
	"context"
	"encoding/json"
	"runtime/cgo"
	"strings"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/node"
	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"
)

type Wrapper struct {
	node   *node.Node
	handle cgo.Handle
}

func NewWrapper(node *node.Node) (*Wrapper, error) {
	return &Wrapper{
		node,
		cgo.NewHandle(node),
	}, nil
}

func (w *Wrapper) PeerInfo() peer.AddrInfo {
	cResult := C.P2PInfo(C.uintptr_t(w.handle))
	info := peer.AddrInfo{}
	_ = json.Unmarshal([]byte(C.GoString(cResult.value)), &info)
	return info
}

func (w *Wrapper) SetReplicator(ctx context.Context, info peer.AddrInfo, collections ...string) error {
	peerStr := info.String()
	colStr := strings.Join(collections, ",")
	cResult := C.P2PsetReplicator(C.uintptr_t(w.handle), C.CString(colStr), C.CString(peerStr))
	err := C.GoString(cResult.error)
	if err != "" {
		return errors.New(err)
	}
	return nil
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

func (w *Wrapper) PatchCollection(
	ctx context.Context,
	migration immutable.Option[model.Lens],
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
