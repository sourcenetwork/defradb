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
#include <stdint.h>
#include "defra_structs.h"
extern Result ACPAddDACPolicy(uintptr_t nodePtr, uintptr_t identity, char* policy);
extern Result ACPAddDACActorRelationship(uintptr_t nodePtr, uintptr_t identityPtr,
char* collection, char* docID, char* relation, char* actor);
extern Result ACPDeleteDACActorRelationship(uintptr_t nodePtr, uintptr_t identity,
char* collection, char* docID, char* relation, char* actor);
extern Result ACPDisableNAC(uintptr_t nodePtr, uintptr_t identityPtr);
extern Result ACPReEnableNAC(uintptr_t nodePtr, uintptr_t identity);
extern Result ACPAddNACActorRelationship(uintptr_t nodePtr, uintptr_t identity,
char* relation, char* actor);
extern Result ACPDeleteNACActorRelationship(uintptr_t nodePtr, uintptr_t identity,
char* relation, char* actor);
extern Result ACPGetNACStatus(uintptr_t nodePtr, uintptr_t identity);
extern Result BlockVerifySignature(uintptr_t nodePtr, char* keyType, char* publicKey, char* cid,
uintptr_t identity);
extern Result CollectionDescribe(uintptr_t nodePtr, CollectionOptions options, uintptr_t identityPtr);
extern Result CollectionPatch(uintptr_t nodePtr, char* patch, char* lensConfig, uintptr_t identityPtr);
extern Result IdentityNew(char* keyType);
extern void IdentityFree(uintptr_t identityPtr);
extern Result NodeIdentity(uintptr_t nodePtr);
extern Result IndexList(uintptr_t nodePtr, CollectionOptions options, uintptr_t identityPtr);
extern Result EncryptedIndexCreate(uintptr_t nodePtr, char* collectionName, char* fieldName);
extern Result EncryptedIndexList(uintptr_t nodePtr, char* collectionName);
extern Result EncryptedIndexDelete(uintptr_t nodePtr, char* collectionName, char* fieldName);
extern Result LensSet(uintptr_t nodePtr, char* src, char* dst, char* cfg);
extern Result LensAdd(uintptr_t nodePtr, char* cfg);
extern Result LensList(uintptr_t nodePtr);
extern NewNodeResult NewNode(NodeInitOptions cOptions);
extern Result NodeClose(uintptr_t nodePtr);
extern Result P2PInfo(uintptr_t nodePtr);
extern Result P2PActivePeers(uintptr_t nodePtr, uintptr_t identity);
extern Result P2PlistReplicators(uintptr_t nodePtr, uintptr_t identity);
extern Result P2PcreateReplicator(uintptr_t nodePtr, char* collections, char* addresses, uintptr_t identity);
extern Result P2PdeleteReplicator(uintptr_t nodePtr, char* collections, char* id, uintptr_t identity);
extern Result P2PcollectionCreate(uintptr_t nodePtr, char* collections, uintptr_t identity);
extern Result P2PcollectionDelete(uintptr_t nodePtr, char* collections, uintptr_t identity);
extern Result P2PcollectionList(uintptr_t nodePtr, uintptr_t identity);
extern Result P2Pconnect(uintptr_t nodePtr, char* peerAddresses, uintptr_t identity);
extern Result P2PdocumentCreate(uintptr_t nodePtr, char* collections, uintptr_t identity);
extern Result P2PdocumentDelete(uintptr_t nodePtr, char* collections, uintptr_t identity);
extern Result P2PdocumentList(uintptr_t nodePtr, uintptr_t identity);
extern Result P2PdocumentSync(uintptr_t nodePtr, char* collection, char* docIDs, char* timeoutStr, uintptr_t identity);
extern Result P2PcollectionSyncVersions(uintptr_t nodePtr, char* versionIDs, char* timeoutStr, uintptr_t identity);
extern Result P2PbranchableCollectionSync(uintptr_t nodePtr, char* collectionID, char* timeoutStr, uintptr_t identity);
extern Result PollSubscription(char* id);
extern Result CloseSubscription(char* id);
extern Result ExecuteQuery(uintptr_t nodePtr, char* query, uintptr_t identity,
char* operationName, char* variables);
extern Result AddSchema(uintptr_t nodePtr, char* schema, uintptr_t identity);
extern Result SetActiveCollection(uintptr_t nodePtr, CollectionOptions options, uintptr_t identityPtr);
extern NewTxnResult TransactionCreate(uintptr_t nodePtr, int isConcurrent, int isReadOnly);
extern Result VersionGet(int flagFull, int flagJSON);
extern Result ViewAdd(uintptr_t nodePtr, char* query, char* sdl, char* transformCIDStr, uintptr_t identityPtr);
extern Result ViewRefresh(uintptr_t nodePtr, CollectionOptions options, uintptr_t identityPtr);
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
	"unsafe"

	"github.com/sourcenetwork/defradb/client"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/node"

	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"
)

var txnHandleMap = sync.Map{} // map[client.Txn]cgo.Handle

var _ client.TxnStore = (*CWrapper)(nil)
var _ client.P2P = (*CWrapper)(nil)

type CWrapper struct {
	node   *node.Node
	handle cgo.Handle
}

func NewCWrapper(node *node.Node) (*CWrapper, error) {
	handle := cgo.NewHandle(node)
	return &CWrapper{
		node,
		handle,
	}, nil
}

func (w *CWrapper) PeerInfo() ([]string, error) {
	res := ConvertAndFreeCResult(C.P2PInfo(C.uintptr_t(w.handle)))

	if res.Status != 0 {
		return nil, errors.New(res.Error)
	}

	addresses, err := unmarshalResult[[]string](res.Value)
	if err != nil {
		return nil, err
	}
	return addresses, nil
}

func (w *CWrapper) ActivePeers(ctx context.Context) ([]string, error) {
	cIdentity := identityFromContext(ctx)
	defer C.IdentityFree(cIdentity)

	res := ConvertAndFreeCResult(C.P2PActivePeers(C.uintptr_t(w.handle), cIdentity))

	if res.Status != 0 {
		return nil, errors.New(res.Error)
	}

	peers, err := unmarshalResult[[]string](res.Value)
	if err != nil {
		return nil, err
	}
	return peers, nil
}

func (w *CWrapper) CreateReplicator(ctx context.Context, addresses []string, collections ...string) error {
	addrStr := C.CString(strings.Join(addresses, ","))
	colStr := C.CString(strings.Join(collections, ","))
	cIdentity := identityFromContext(ctx)
	defer C.free(unsafe.Pointer(addrStr))
	defer C.free(unsafe.Pointer(colStr))
	defer C.IdentityFree(cIdentity)

	res := ConvertAndFreeCResult(C.P2PcreateReplicator(C.uintptr_t(w.handle), colStr, addrStr, cIdentity))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (w *CWrapper) DeleteReplicator(ctx context.Context, id string, collections ...string) error {
	peerID := C.CString(id)
	colStr := C.CString(strings.Join(collections, ","))
	cIdentity := identityFromContext(ctx)
	defer C.free(unsafe.Pointer(peerID))
	defer C.free(unsafe.Pointer(colStr))
	defer C.IdentityFree(cIdentity)

	res := ConvertAndFreeCResult(C.P2PdeleteReplicator(C.uintptr_t(w.handle), colStr, peerID, cIdentity))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (w *CWrapper) ListReplicators(ctx context.Context) ([]client.Replicator, error) {
	cIdentity := identityFromContext(ctx)
	defer C.IdentityFree(cIdentity)
	res := ConvertAndFreeCResult(C.P2PlistReplicators(C.uintptr_t(w.handle), cIdentity))

	if res.Status != 0 {
		return nil, errors.New(res.Error)
	}

	replicators, err := unmarshalResult[[]client.Replicator](res.Value)
	if err != nil {
		return nil, err
	}
	return replicators, nil
}

func (w *CWrapper) CreateP2PCollections(ctx context.Context, collectionIDs ...string) error {
	cIdentity := identityFromContext(ctx)
	colStr := C.CString(strings.Join(collectionIDs, ","))
	defer C.free(unsafe.Pointer(colStr))
	defer C.IdentityFree(cIdentity)
	res := ConvertAndFreeCResult(C.P2PcollectionCreate(C.uintptr_t(w.handle), colStr, cIdentity))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (w *CWrapper) DeleteP2PCollections(ctx context.Context, collectionIDs ...string) error {
	colStr := C.CString(strings.Join(collectionIDs, ","))
	cIdentity := identityFromContext(ctx)
	defer C.free(unsafe.Pointer(colStr))
	defer C.IdentityFree(cIdentity)

	res := ConvertAndFreeCResult(C.P2PcollectionDelete(C.uintptr_t(w.handle), colStr, cIdentity))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (w *CWrapper) ListP2PCollections(ctx context.Context) ([]string, error) {
	cIdentity := identityFromContext(ctx)
	defer C.IdentityFree(cIdentity)
	res := ConvertAndFreeCResult(C.P2PcollectionList(C.uintptr_t(w.handle), cIdentity))

	if res.Status != 0 {
		return nil, errors.New(res.Error)
	}

	collections, err := unmarshalResult[[]string](res.Value)
	if err != nil {
		return nil, err
	}
	return collections, nil
}

func (w *CWrapper) CreateP2PDocuments(ctx context.Context, docIDs ...string) error {
	docStr := C.CString(strings.Join(docIDs, ","))
	cIdentity := identityFromContext(ctx)
	defer C.IdentityFree(cIdentity)
	defer C.free(unsafe.Pointer(docStr))

	res := ConvertAndFreeCResult(C.P2PdocumentCreate(C.uintptr_t(w.handle), docStr, cIdentity))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (w *CWrapper) DeleteP2PDocuments(ctx context.Context, docIDs ...string) error {
	docStr := C.CString(strings.Join(docIDs, ","))
	cIdentity := identityFromContext(ctx)
	defer C.IdentityFree(cIdentity)
	defer C.free(unsafe.Pointer(docStr))

	res := ConvertAndFreeCResult(C.P2PdocumentDelete(C.uintptr_t(w.handle), docStr, cIdentity))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (w *CWrapper) ListP2PDocuments(ctx context.Context) ([]string, error) {
	cIdentity := identityFromContext(ctx)
	defer C.IdentityFree(cIdentity)
	res := ConvertAndFreeCResult(C.P2PdocumentList(C.uintptr_t(w.handle), cIdentity))

	if res.Status != 0 {
		return nil, errors.New(res.Error)
	}

	docs, err := unmarshalResult[[]string](res.Value)
	if err != nil {
		return nil, err
	}
	return docs, nil
}

func (w *CWrapper) SyncDocuments(
	ctx context.Context,
	collectionName string,
	docIDs []string,
) error {
	cIdentity := identityFromContext(ctx)
	docs := C.CString(strings.Join(docIDs, ","))
	defer C.free(unsafe.Pointer(docs))
	defer C.IdentityFree(cIdentity)

	deadline, hasDeadline := ctx.Deadline()
	timerStr := ""
	if hasDeadline {
		timerStr = time.Until(deadline).String()
	}
	cTimerStr := C.CString(timerStr)
	cCollectionName := C.CString(collectionName)
	defer C.free(unsafe.Pointer(cTimerStr))
	defer C.free(unsafe.Pointer(cCollectionName))

	res := ConvertAndFreeCResult(C.P2PdocumentSync(C.uintptr_t(w.handle), cCollectionName, docs, cTimerStr, cIdentity))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (w *CWrapper) SyncCollectionVersions(ctx context.Context, versionIDs ...string) error {
	cIdentity := identityFromContext(ctx)
	versions := C.CString(strings.Join(versionIDs, ","))
	defer C.free(unsafe.Pointer(versions))
	defer C.IdentityFree(cIdentity)

	deadline, hasDeadline := ctx.Deadline()
	timerStr := ""
	if hasDeadline {
		timerStr = time.Until(deadline).String()
	}
	cTimerStr := C.CString(timerStr)
	defer C.free(unsafe.Pointer(cTimerStr))

	res := ConvertAndFreeCResult(C.P2PcollectionSyncVersions(C.uintptr_t(w.handle), versions, cTimerStr, cIdentity))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (w *CWrapper) SyncBranchableCollection(ctx context.Context, collectionID string) error {
	cIdentity := identityFromContext(ctx)
	cCollectionID := C.CString(collectionID)
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.IdentityFree(cIdentity)

	deadline, hasDeadline := ctx.Deadline()
	timerStr := ""
	if hasDeadline {
		timerStr = time.Until(deadline).String()
	}
	cTimerStr := C.CString(timerStr)
	defer C.free(unsafe.Pointer(cTimerStr))

	res := ConvertAndFreeCResult(C.P2PbranchableCollectionSync(C.uintptr_t(w.handle), cCollectionID, cTimerStr,
		cIdentity))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (w *CWrapper) BasicImport(ctx context.Context, filepath string) error {
	panic("not implemented")
}

func (w *CWrapper) BasicExport(ctx context.Context, config *client.BackupConfig) error {
	panic("not implemented")
}

func (w *CWrapper) AddSchema(ctx context.Context, schema string) ([]client.CollectionVersion, error) {
	cIdentity := identityFromContext(ctx)
	defer C.IdentityFree(cIdentity)
	cSchema := C.CString(schema)
	defer C.free(unsafe.Pointer(cSchema))

	callHandle := getNodeOrTxnHandle(w.handle, ctx)
	res := ConvertAndFreeCResult(C.AddSchema(callHandle, cSchema, cIdentity))

	if res.Status != 0 {
		return nil, errors.New(res.Error)
	}

	collectionVersions, err := unmarshalResult[[]client.CollectionVersion](res.Value)
	if err != nil {
		return nil, err
	}
	return collectionVersions, nil
}

func (w *CWrapper) AddDACPolicy(
	ctx context.Context,
	policy string,
) (client.AddPolicyResult, error) {
	cIdentity := identityFromContext(ctx)
	defer C.IdentityFree(cIdentity)
	cPolicy := C.CString(policy)
	defer C.free(unsafe.Pointer(cPolicy))

	callHandle := getNodeOrTxnHandle(w.handle, ctx)
	res := ConvertAndFreeCResult(C.ACPAddDACPolicy(callHandle, cIdentity, cPolicy))

	if res.Status != 0 {
		return client.AddPolicyResult{}, errors.New(res.Error)
	}

	addPolicyRes, err := unmarshalResult[client.AddPolicyResult](res.Value)
	if err != nil {
		return client.AddPolicyResult{}, err
	}
	return addPolicyRes, nil
}

func (w *CWrapper) AddDACActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
) (client.AddActorRelationshipResult, error) {
	cIdentity := identityFromContext(ctx)
	cCollectionName := C.CString(collectionName)
	cDocID := C.CString(docID)
	cRelation := C.CString(relation)
	cTargetActor := C.CString(targetActor)
	defer C.free(unsafe.Pointer(cCollectionName))
	defer C.free(unsafe.Pointer(cDocID))
	defer C.free(unsafe.Pointer(cRelation))
	defer C.free(unsafe.Pointer(cTargetActor))
	defer C.IdentityFree(cIdentity)

	callHandle := getNodeOrTxnHandle(w.handle, ctx)
	res := ConvertAndFreeCResult(C.ACPAddDACActorRelationship(
		callHandle,
		cIdentity,
		cCollectionName,
		cDocID,
		cRelation,
		cTargetActor,
	))

	if res.Status != 0 {
		return client.AddActorRelationshipResult{}, errors.New(res.Error)
	}

	// Unmarshall the output from JSON to client.AddActorRelationshipResult
	addRelationshipRes, err := unmarshalResult[client.AddActorRelationshipResult](res.Value)
	if err != nil {
		return client.AddActorRelationshipResult{}, err
	}
	return addRelationshipRes, nil
}

func (w *CWrapper) DeleteDACActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
) (client.DeleteActorRelationshipResult, error) {
	cIdentity := identityFromContext(ctx)
	cCollectionName := C.CString(collectionName)
	cDocID := C.CString(docID)
	cRelation := C.CString(relation)
	cTargetActor := C.CString(targetActor)
	defer C.free(unsafe.Pointer(cCollectionName))
	defer C.free(unsafe.Pointer(cDocID))
	defer C.free(unsafe.Pointer(cRelation))
	defer C.free(unsafe.Pointer(cTargetActor))
	defer C.IdentityFree(cIdentity)

	callHandle := getNodeOrTxnHandle(w.handle, ctx)
	res := ConvertAndFreeCResult(C.ACPDeleteDACActorRelationship(
		callHandle,
		cIdentity,
		cCollectionName,
		cDocID,
		cRelation,
		cTargetActor,
	))

	if res.Status != 0 {
		return client.DeleteActorRelationshipResult{}, errors.New(res.Error)
	}

	deleteRelationshipRes, err := unmarshalResult[client.DeleteActorRelationshipResult](res.Value)
	if err != nil {
		return client.DeleteActorRelationshipResult{}, err
	}
	return deleteRelationshipRes, nil
}

func (w *CWrapper) GetNACStatus(ctx context.Context) (client.NACStatusResult, error) {
	cIdentity := identityFromContext(ctx)
	defer C.IdentityFree(cIdentity)

	callHandle := getNodeOrTxnHandle(w.handle, ctx)
	res := ConvertAndFreeCResult(C.ACPGetNACStatus(callHandle, cIdentity))

	if res.Status != 0 {
		return client.NACStatusResult{}, errors.New(res.Error)
	}
	return unmarshalResult[client.NACStatusResult](res.Value)
}

func (w *CWrapper) ReEnableNAC(ctx context.Context) error {
	cIdentity := identityFromContext(ctx)
	defer C.IdentityFree(cIdentity)

	callHandle := getNodeOrTxnHandle(w.handle, ctx)
	res := ConvertAndFreeCResult(C.ACPReEnableNAC(callHandle, cIdentity))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (w *CWrapper) DisableNAC(ctx context.Context) error {
	cIdentity := identityFromContext(ctx)
	defer C.IdentityFree(cIdentity)

	callHandle := getNodeOrTxnHandle(w.handle, ctx)
	res := ConvertAndFreeCResult(C.ACPDisableNAC(callHandle, cIdentity))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (w *CWrapper) AddNACActorRelationship(
	ctx context.Context,
	relation string,
	targetActor string,
) (client.AddActorRelationshipResult, error) {
	cIdentity := identityFromContext(ctx)
	cRelation := C.CString(relation)
	cTargetActor := C.CString(targetActor)
	defer C.free(unsafe.Pointer(cRelation))
	defer C.free(unsafe.Pointer(cTargetActor))
	defer C.IdentityFree(cIdentity)

	callHandle := getNodeOrTxnHandle(w.handle, ctx)
	res := ConvertAndFreeCResult(C.ACPAddNACActorRelationship(callHandle, cIdentity, cRelation, cTargetActor))

	if res.Status != 0 {
		return client.AddActorRelationshipResult{}, errors.New(res.Error)
	}

	return unmarshalResult[client.AddActorRelationshipResult](res.Value)
}

func (w *CWrapper) DeleteNACActorRelationship(
	ctx context.Context,
	relation string,
	targetActor string,
) (client.DeleteActorRelationshipResult, error) {
	cIdentity := identityFromContext(ctx)
	cRelation := C.CString(relation)
	cTargetActor := C.CString(targetActor)
	defer C.free(unsafe.Pointer(cRelation))
	defer C.free(unsafe.Pointer(cTargetActor))
	defer C.IdentityFree(cIdentity)

	callHandle := getNodeOrTxnHandle(w.handle, ctx)
	res := ConvertAndFreeCResult(C.ACPDeleteNACActorRelationship(callHandle, cIdentity, cRelation, cTargetActor))
	if res.Status != 0 {
		return client.DeleteActorRelationshipResult{}, errors.New(res.Error)
	}
	return unmarshalResult[client.DeleteActorRelationshipResult](res.Value)
}

func (w *CWrapper) PatchCollection(
	ctx context.Context,
	patch string,
	migration immutable.Option[model.Lens],
) error {
	cPatch := C.CString(patch)
	cIdentity := identityFromContext(ctx)
	cVersion := C.CString("")
	cCollectionID := C.CString("")
	cName := C.CString("")
	defer C.free(unsafe.Pointer(cPatch))
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cName))
	defer C.IdentityFree(cIdentity)

	migrationStr, migrationErr := optionToString(migration)
	if migrationErr != nil {
		return migrationErr
	}
	cMigration := C.CString(migrationStr)
	defer C.free(unsafe.Pointer(cMigration))

	callHandle := getNodeOrTxnHandle(w.handle, ctx)
	res := ConvertAndFreeCResult(C.CollectionPatch(callHandle, cPatch, cMigration, cIdentity))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (w *CWrapper) SetActiveCollectionVersion(ctx context.Context, collectionVersionID string) error {
	cIdentity := identityFromContext(ctx)
	cVersion := C.CString(collectionVersionID)
	cCollectionID := C.CString("")
	cName := C.CString("")

	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cName))
	defer C.IdentityFree(cIdentity)

	var opts C.CollectionOptions
	opts.version = cVersion
	opts.collectionID = cCollectionID
	opts.name = cName
	opts.getInactive = 0

	callHandle := getNodeOrTxnHandle(w.handle, ctx)
	res := ConvertAndFreeCResult(C.SetActiveCollection(callHandle, opts, cIdentity))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (w *CWrapper) AddView(
	ctx context.Context,
	query string,
	sdl string,
	transformCID immutable.Option[string],
) ([]client.CollectionVersion, error) {
	cIdentity := identityFromContext(ctx)
	cTransformCID := C.CString(stringFromImmutableOptionString(transformCID))
	cQuery := C.CString(query)
	cSDL := C.CString(sdl)
	defer C.free(unsafe.Pointer(cTransformCID))
	defer C.free(unsafe.Pointer(cQuery))
	defer C.free(unsafe.Pointer(cSDL))
	defer C.IdentityFree(cIdentity)

	callHandle := getNodeOrTxnHandle(w.handle, ctx)
	res := ConvertAndFreeCResult(C.ViewAdd(callHandle, cQuery, cSDL, cTransformCID, cIdentity))

	if res.Status != 0 {
		return []client.CollectionVersion{}, errors.New(res.Error)
	}

	colDefRes, err := unmarshalResult[[]client.CollectionVersion](res.Value)
	if err != nil {
		return []client.CollectionVersion{}, err
	}
	return colDefRes, nil
}

func (w *CWrapper) RefreshViews(ctx context.Context, opts client.CollectionFetchOptions) error {
	cIdentity := identityFromContext(ctx)
	versionID := C.CString(stringFromImmutableOptionString(opts.VersionID))
	collectionID := C.CString(stringFromImmutableOptionString(opts.CollectionID))
	name := C.CString(stringFromImmutableOptionString(opts.Name))
	var cGetInactive C.int = 0
	if opts.IncludeInactive.HasValue() {
		if opts.IncludeInactive.Value() {
			cGetInactive = 1
		}
	}
	defer C.free(unsafe.Pointer(versionID))
	defer C.free(unsafe.Pointer(collectionID))
	defer C.free(unsafe.Pointer(name))
	defer C.IdentityFree(cIdentity)

	var copts C.CollectionOptions
	copts.version = versionID
	copts.collectionID = collectionID
	copts.name = name
	copts.getInactive = cGetInactive

	callHandle := getNodeOrTxnHandle(w.handle, ctx)
	res := ConvertAndFreeCResult(C.ViewRefresh(callHandle, copts, cIdentity))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (w *CWrapper) SetMigration(ctx context.Context, config client.LensConfig) (string, error) {
	src := C.CString(config.SourceCollectionVersionID)
	dst := C.CString(config.DestinationCollectionVersionID)
	lensConfig, err := json.Marshal(config.Lens)
	if err != nil {
		return "", err
	}
	lens := C.CString(string(lensConfig))
	defer C.free(unsafe.Pointer(src))
	defer C.free(unsafe.Pointer(dst))
	defer C.free(unsafe.Pointer(lens))

	callHandle := getNodeOrTxnHandle(w.handle, ctx)
	res := ConvertAndFreeCResult(C.LensSet(callHandle, src, dst, lens))

	if res.Status != 0 {
		return "", errors.New(res.Error)
	}
	return res.Value, nil
}

func (w *CWrapper) AddLens(ctx context.Context, lens model.Lens) (string, error) {
	lensConfig, err := json.Marshal(lens)
	if err != nil {
		return "", err
	}
	lensStr := C.CString(string(lensConfig))
	defer C.free(unsafe.Pointer(lensStr))

	callHandle := getNodeOrTxnHandle(w.handle, ctx)
	res := ConvertAndFreeCResult(C.LensAdd(callHandle, lensStr))

	if res.Status != 0 {
		return "", errors.New(res.Error)
	}
	return res.Value, nil
}

func (w *CWrapper) ListLenses(ctx context.Context) (map[string]model.Lens, error) {
	callHandle := getNodeOrTxnHandle(w.handle, ctx)
	res := ConvertAndFreeCResult(C.LensList(callHandle))

	if res.Status != 0 {
		return nil, errors.New(res.Error)
	}

	var lenses map[string]model.Lens
	if err := json.Unmarshal([]byte(res.Value), &lenses); err != nil {
		return nil, err
	}
	return lenses, nil
}

func (w *CWrapper) GetCollectionByName(ctx context.Context, name client.CollectionName) (client.Collection, error) {
	cols, err := w.GetCollections(ctx, client.CollectionFetchOptions{Name: immutable.Some(name)})
	if err != nil {
		return nil, err
	}

	if len(cols) == 0 {
		return nil, fmt.Errorf("collection with name %q not found", name)
	}

	// cols will always have length == 1 here
	return cols[0], nil
}

func (w *CWrapper) GetCollections(
	ctx context.Context,
	options client.CollectionFetchOptions,
) ([]client.Collection, error) {
	var name string
	if options.Name.HasValue() {
		name = options.Name.Value()
	} else {
		name = ""
	}

	var version string
	if options.VersionID.HasValue() {
		version = options.VersionID.Value()
	} else {
		version = ""
	}

	var collectionID string
	if options.CollectionID.HasValue() {
		collectionID = options.CollectionID.Value()
	} else {
		collectionID = ""
	}

	var includeInactive int = 0
	if options.IncludeInactive.HasValue() {
		if options.IncludeInactive.Value() {
			includeInactive = 1
		}
	}

	cVersion := C.CString(version)
	cCollectionID := C.CString(collectionID)
	cName := C.CString(name)
	cIdentity := identityFromContext(ctx)
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cName))
	defer C.IdentityFree(cIdentity)

	var opts C.CollectionOptions
	opts.version = cVersion
	opts.collectionID = cCollectionID
	opts.name = cName
	opts.getInactive = C.int(includeInactive)

	callHandle := getNodeOrTxnHandle(w.handle, ctx)
	res := ConvertAndFreeCResult(C.CollectionDescribe(callHandle, opts, cIdentity))

	if res.Status != 0 {
		return []client.Collection{}, errors.New(res.Error)
	}

	defs, err := unmarshalResult[[]client.CollectionVersion](res.Value)
	if err != nil {
		return nil, err
	}

	cols := make([]client.Collection, len(defs))
	for i, def := range defs {
		cols[i] = &Collection{def: def, w: w}
	}
	return cols, nil
}

func (w *CWrapper) GetAllIndexes(ctx context.Context) (map[client.CollectionName][]client.IndexDescription, error) {
	cVersion := C.CString("")
	cCollectionID := C.CString("")
	cName := C.CString("")
	cIdentity := identityFromContext(ctx)
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cName))
	defer C.IdentityFree(cIdentity)

	var opts C.CollectionOptions
	opts.version = cVersion
	opts.collectionID = cCollectionID
	opts.name = cName
	opts.getInactive = 0

	callHandle := getNodeOrTxnHandle(w.handle, ctx)
	res := ConvertAndFreeCResult(C.IndexList(callHandle, opts, cIdentity))

	if res.Status != 0 {
		return nil, errors.New(res.Error)
	}

	resValue, err := unmarshalResult[map[client.CollectionName][]client.IndexDescription](res.Value)
	if err != nil {
		return nil, errors.New(res.Error)
	}

	return resValue, nil
}

func (w *CWrapper) ListAllEncryptedIndexes(
	ctx context.Context,
) (map[client.CollectionName][]client.EncryptedIndexDescription, error) {
	colName := C.CString("")
	defer C.free(unsafe.Pointer(colName))

	callHandle := getNodeOrTxnHandle(w.handle, ctx)
	res := ConvertAndFreeCResult(C.EncryptedIndexList(callHandle, colName))

	if res.Status != 0 {
		return nil, errors.New(res.Error)
	}

	resValue, err := unmarshalResult[map[client.CollectionName][]client.EncryptedIndexDescription](res.Value)
	if err != nil {
		return nil, errors.New(res.Error)
	}

	return resValue, nil
}

func (w *CWrapper) ExecRequest(
	ctx context.Context,
	query string,
	opts ...client.RequestOption,
) *client.RequestResult {
	operation, variables, err := extractStringsFromRequestOptions(opts)
	if err != nil {
		return &client.RequestResult{
			GQL: client.GQLResult{
				Errors: []error{err},
			},
		}
	}

	cQuery := C.CString(query)
	cIdentity := identityFromContext(ctx)
	cOperation := C.CString(operation)
	cVariables := C.CString(variables)
	defer C.free(unsafe.Pointer(cQuery))
	defer C.free(unsafe.Pointer(cOperation))
	defer C.free(unsafe.Pointer(cVariables))
	defer C.IdentityFree(cIdentity)

	callHandle := getNodeOrTxnHandle(w.handle, ctx)
	result := C.ExecuteQuery(callHandle, cQuery, cIdentity, cOperation, cVariables)
	res := ConvertAndFreeCResult(result)

	if res.Status == 2 {
		id := res.Value
		newchan := wrapSubscriptionAsChannel(ctx, id)
		return &client.RequestResult{
			Subscription: newchan,
		}
	}

	retval := &client.RequestResult{}
	if res.Status != 0 {
		retval.GQL.Errors = append(retval.GQL.Errors, fmt.Errorf("%s", res.Error))
		return retval
	}
	if err := json.Unmarshal([]byte(res.Value), &retval.GQL); err != nil {
		retval.GQL.Errors = append(retval.GQL.Errors, err)
	}
	return retval
}

func (w *CWrapper) NewTxn(readOnly bool) (client.Txn, error) {
	var concurrent C.int = 0
	var cReadOnly C.int = 0
	if readOnly {
		cReadOnly = 1
	}

	res := C.TransactionCreate(C.uintptr_t(w.handle), concurrent, cReadOnly)
	errText := C.GoString(res.error)
	defer C.free(unsafe.Pointer(res.error))

	if res.status != 0 {
		return nil, errors.New(errText)
	}

	handle := cgo.Handle(res.txnPtr)
	clientTxn := handle.Value().(client.Txn) //nolint:forcetypeassert
	retTxn := &Transaction{w, clientTxn, handle}
	txnHandleMap.Store(retTxn, handle)

	return retTxn, nil
}

func (w *CWrapper) NewConcurrentTxn(readOnly bool) (client.Txn, error) {
	var concurrent C.int = 1
	var cReadOnly C.int = 0
	if readOnly {
		cReadOnly = 1
	}

	res := C.TransactionCreate(C.uintptr_t(w.handle), concurrent, cReadOnly)
	errText := C.GoString(res.error)
	defer C.free(unsafe.Pointer(res.error))

	if res.status != 0 {
		return nil, errors.New(errText)
	}

	handle := cgo.Handle(res.txnPtr)
	clientTxn := handle.Value().(client.Txn) //nolint:forcetypeassert
	retTxn := &Transaction{w, clientTxn, handle}
	txnHandleMap.Store(retTxn, handle)

	return retTxn, nil
}

func (w *CWrapper) Close() {
	C.NodeClose(C.uintptr_t(w.handle))
}

func (w *CWrapper) Events() event.Bus {
	return w.node.DB.Events()
}

func (w *CWrapper) MaxTxnRetries() int {
	return w.node.DB.MaxTxnRetries()
}

func (w *CWrapper) PrintDump(ctx context.Context) error {
	panic("not implemented")
}

func (w *CWrapper) Connect(ctx context.Context, addresses []string) error {
	cIdentity := identityFromContext(ctx)
	cPeerAddresses := C.CString(strings.Join(addresses, ","))
	defer C.free(unsafe.Pointer(cPeerAddresses))
	defer C.IdentityFree(cIdentity)
	callHandle := getNodeOrTxnHandle(w.handle, ctx)
	res := ConvertAndFreeCResult(C.P2Pconnect(callHandle, cPeerAddresses, cIdentity))
	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (w *CWrapper) GetNodeIdentity(ctx context.Context) (immutable.Option[identity.PublicRawIdentity], error) {
	callHandle := getNodeOrTxnHandle(w.handle, ctx)
	res := ConvertAndFreeCResult(C.NodeIdentity(callHandle))

	if res.Status != 0 {
		return immutable.None[identity.PublicRawIdentity](), errors.New(res.Error)
	}

	if res.Value == "Node has no identity assigned to it." {
		return immutable.None[identity.PublicRawIdentity](), nil
	}

	var resVal identity.PublicRawIdentity
	resVal, err := unmarshalResult[identity.PublicRawIdentity](res.Value)
	if err != nil {
		return immutable.None[identity.PublicRawIdentity](), err
	}
	return immutable.Some(resVal), nil
}

func (w *CWrapper) VerifySignature(ctx context.Context, blockCid string, pubKey crypto.PublicKey) error {
	cPubKey := C.CString(pubKey.String())
	cKeyType := C.CString(string(pubKey.Type()))
	cBlockCid := C.CString(blockCid)
	cIdentity := identityFromContext(ctx)

	defer C.free(unsafe.Pointer(cPubKey))
	defer C.free(unsafe.Pointer(cKeyType))
	defer C.free(unsafe.Pointer(cBlockCid))
	defer C.IdentityFree(cIdentity)

	res := ConvertAndFreeCResult(C.BlockVerifySignature(C.uintptr_t(w.handle), cKeyType, cPubKey, cBlockCid, cIdentity))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}
