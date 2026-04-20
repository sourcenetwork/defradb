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
extern Result VerifyBlockSignature(uintptr_t nodePtr, char* keyType, char* publicKey, char* cid,
uintptr_t identity);
extern Result DescribeCollection(uintptr_t nodePtr, CollectionOptions options, uintptr_t identityPtr);
extern Result PatchCollection(uintptr_t nodePtr, char* patch, char* lensConfig, uintptr_t identityPtr);
extern Result NewIdentity(char* keyType);
extern void FreeIdentity(uintptr_t identityPtr);
extern Result GetNodeIdentity(uintptr_t nodePtr);
extern Result ListIndexes(uintptr_t nodePtr, CollectionOptions options, uintptr_t identityPtr);
extern Result NewEncryptedIndex(uintptr_t nodePtr, char* collectionName, char* fieldName, uintptr_t identity);
extern Result ListEncryptedIndexes(uintptr_t nodePtr, char* collectionName, uintptr_t identityPtr);
extern Result DeleteEncryptedIndex(uintptr_t nodePtr, char* collectionName, char* fieldName, uintptr_t identity);
extern Result SetLens(uintptr_t nodePtr, uintptr_t identity, char* src, char* dst, char* cfg);
extern Result AddLens(uintptr_t nodePtr, uintptr_t identityPtr, char* cfg);
extern Result ListLenses(uintptr_t nodePtr, uintptr_t identityPtr);
extern NewNodeResult NewNode(NodeInitOptions cOptions);
extern Result CloseNode(uintptr_t nodePtr);
extern Result GetP2PInfo(uintptr_t nodePtr, uintptr_t identity);
extern Result ListP2PActivePeers(uintptr_t nodePtr, uintptr_t identity);
extern Result ListP2PReplicators(uintptr_t nodePtr, uintptr_t identity);
extern Result AddP2PReplicator(uintptr_t nodePtr, char* collections, char* addresses, uintptr_t identity);
extern Result DeleteP2PReplicator(uintptr_t nodePtr, char* collections, char* id, uintptr_t identity);
extern Result AddP2PCollection(uintptr_t nodePtr, char* collections, uintptr_t identity);
extern Result DeleteP2PCollection(uintptr_t nodePtr, char* collections, uintptr_t identity);
extern Result ListP2PCollections(uintptr_t nodePtr, uintptr_t identity);
extern Result ConnectP2PPeers(uintptr_t nodePtr, char* peerAddresses, uintptr_t identity);
extern Result AddP2PDocument(uintptr_t nodePtr, char* collections, uintptr_t identity);
extern Result DeleteP2PDocument(uintptr_t nodePtr, char* collections, uintptr_t identity);
extern Result ListP2PDocuments(uintptr_t nodePtr, uintptr_t identity);
extern Result SyncP2PDocuments(uintptr_t nodePtr, char* collection, char* docIDs, char* timeoutStr, uintptr_t identity);
extern Result SyncP2PCollectionVersions(uintptr_t nodePtr, char* versionIDs, char* timeoutStr, uintptr_t identity);
extern Result SyncP2PBranchableCollection(uintptr_t nodePtr, char* collectionID, char* timeoutStr, uintptr_t identity);
extern Result PollSubscription(char* id);
extern Result CloseSubscription(char* id);
extern Result ExecuteQuery(uintptr_t nodePtr, char* query, uintptr_t identity,
char* operationName, char* variables);
extern Result AddCollection(uintptr_t nodePtr, char* schema, uintptr_t identity);
extern Result SetActiveCollection(uintptr_t nodePtr, CollectionOptions options, uintptr_t identityPtr);
extern NewTxnResult CreateTransaction(uintptr_t nodePtr, int isReadOnly);
extern Result GetVersion(int flagFull, int flagJSON);
extern Result AddView(uintptr_t nodePtr, char* query, char* sdl, char* transformCIDStr, uintptr_t identityPtr);
extern Result RefreshView(uintptr_t nodePtr, CollectionOptions options, uintptr_t identityPtr);
*/
import "C"

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime/cgo"
	"strings"
	"time"
	"unsafe"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/utils"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/node"

	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/lens/host-go/config/model"
)

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

func (w *CWrapper) PeerInfo(
	ctx context.Context, opts ...options.Enumerable[options.PeerInfoOptions],
) ([]string, error) {
	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())
	defer C.FreeIdentity(cIdentity)

	res := ConvertAndFreeCResult(C.GetP2PInfo(C.uintptr_t(w.handle), cIdentity))

	if res.Status != 0 {
		return nil, errors.New(res.Error)
	}

	addresses, err := unmarshalResult[[]string](res.Value)
	if err != nil {
		return nil, err
	}
	return addresses, nil
}

func (w *CWrapper) ActivePeers(
	ctx context.Context,
	opts ...options.Enumerable[options.ActivePeersOptions],
) ([]string, error) {
	opt := utils.NewOptions(opts...)
	cIdentity := optionToUintptr(opt.GetIdentity())
	defer C.FreeIdentity(cIdentity)

	res := ConvertAndFreeCResult(C.ListP2PActivePeers(C.uintptr_t(w.handle), cIdentity))

	if res.Status != 0 {
		return nil, errors.New(res.Error)
	}

	peers, err := unmarshalResult[[]string](res.Value)
	if err != nil {
		return nil, err
	}
	return peers, nil
}

func (w *CWrapper) AddReplicator(
	ctx context.Context,
	addresses []string,
	opts ...options.Enumerable[options.AddReplicatorOptions],
) error {
	opt := utils.NewOptions(opts...)
	addrStr := C.CString(strings.Join(addresses, ","))
	colStr := C.CString(strings.Join(opt.CollectionNames, ","))
	cIdentity := optionToUintptr(opt.GetIdentity())
	defer C.free(unsafe.Pointer(addrStr))
	defer C.free(unsafe.Pointer(colStr))
	defer C.FreeIdentity(cIdentity)

	res := ConvertAndFreeCResult(C.AddP2PReplicator(C.uintptr_t(w.handle), colStr, addrStr, cIdentity))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (w *CWrapper) DeleteReplicator(
	ctx context.Context,
	id string,
	opts ...options.Enumerable[options.DeleteReplicatorOptions],
) error {
	opt := utils.NewOptions(opts...)
	peerID := C.CString(id)
	colStr := C.CString(strings.Join(opt.CollectionNames, ","))
	cIdentity := optionToUintptr(opt.GetIdentity())
	defer C.free(unsafe.Pointer(peerID))
	defer C.free(unsafe.Pointer(colStr))
	defer C.FreeIdentity(cIdentity)

	res := ConvertAndFreeCResult(C.DeleteP2PReplicator(C.uintptr_t(w.handle), colStr, peerID, cIdentity))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (w *CWrapper) ListReplicators(
	ctx context.Context,
	opts ...options.Enumerable[options.ListReplicatorsOptions],
) ([]client.Replicator, error) {
	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())
	defer C.FreeIdentity(cIdentity)
	res := ConvertAndFreeCResult(C.ListP2PReplicators(C.uintptr_t(w.handle), cIdentity))

	if res.Status != 0 {
		return nil, errors.New(res.Error)
	}

	replicators, err := unmarshalResult[[]client.Replicator](res.Value)
	if err != nil {
		return nil, err
	}
	return replicators, nil
}

func (w *CWrapper) AddP2PCollections(
	ctx context.Context,
	collectionIDs []string,
	opts ...options.Enumerable[options.AddP2PCollectionsOptions],
) error {
	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())
	colStr := C.CString(strings.Join(collectionIDs, ","))
	defer C.free(unsafe.Pointer(colStr))
	defer C.FreeIdentity(cIdentity)
	res := ConvertAndFreeCResult(C.AddP2PCollection(C.uintptr_t(w.handle), colStr, cIdentity))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (w *CWrapper) DeleteP2PCollections(
	ctx context.Context,
	collectionIDs []string,
	opts ...options.Enumerable[options.DeleteP2PCollectionsOptions],
) error {
	colStr := C.CString(strings.Join(collectionIDs, ","))
	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())
	defer C.free(unsafe.Pointer(colStr))
	defer C.FreeIdentity(cIdentity)

	res := ConvertAndFreeCResult(C.DeleteP2PCollection(C.uintptr_t(w.handle), colStr, cIdentity))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (w *CWrapper) ListP2PCollections(
	ctx context.Context,
	opts ...options.Enumerable[options.ListP2PCollectionsOptions],
) ([]string, error) {
	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())
	defer C.FreeIdentity(cIdentity)
	res := ConvertAndFreeCResult(C.ListP2PCollections(C.uintptr_t(w.handle), cIdentity))

	if res.Status != 0 {
		return nil, errors.New(res.Error)
	}

	collections, err := unmarshalResult[[]string](res.Value)
	if err != nil {
		return nil, err
	}
	return collections, nil
}

func (w *CWrapper) AddP2PDocuments(
	ctx context.Context,
	docIDs []string,
	opts ...options.Enumerable[options.AddP2PDocumentsOptions],
) error {
	docStr := C.CString(strings.Join(docIDs, ","))
	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())
	defer C.FreeIdentity(cIdentity)
	defer C.free(unsafe.Pointer(docStr))

	res := ConvertAndFreeCResult(C.AddP2PDocument(C.uintptr_t(w.handle), docStr, cIdentity))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (w *CWrapper) DeleteP2PDocuments(
	ctx context.Context,
	docIDs []string,
	opts ...options.Enumerable[options.DeleteP2PDocumentsOptions],
) error {
	docStr := C.CString(strings.Join(docIDs, ","))
	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())
	defer C.FreeIdentity(cIdentity)
	defer C.free(unsafe.Pointer(docStr))

	res := ConvertAndFreeCResult(C.DeleteP2PDocument(C.uintptr_t(w.handle), docStr, cIdentity))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (w *CWrapper) ListP2PDocuments(
	ctx context.Context,
	opts ...options.Enumerable[options.ListP2PDocumentsOptions],
) ([]string, error) {
	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())
	defer C.FreeIdentity(cIdentity)
	res := ConvertAndFreeCResult(C.ListP2PDocuments(C.uintptr_t(w.handle), cIdentity))

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
	opts ...options.Enumerable[options.SyncDocumentsOptions],
) error {
	opt := utils.NewOptions(opts...)
	cIdentity := optionToUintptr(opt.GetIdentity())
	defer C.FreeIdentity(cIdentity)

	docs := C.CString(strings.Join(docIDs, ","))
	defer C.free(unsafe.Pointer(docs))

	deadline, hasDeadline := ctx.Deadline()
	timerStr := ""
	if hasDeadline {
		timerStr = time.Until(deadline).String()
	}
	cTimerStr := C.CString(timerStr)
	cCollectionName := C.CString(collectionName)
	defer C.free(unsafe.Pointer(cTimerStr))
	defer C.free(unsafe.Pointer(cCollectionName))

	res := ConvertAndFreeCResult(C.SyncP2PDocuments(
		C.uintptr_t(w.handle), cCollectionName, docs, cTimerStr, cIdentity))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (w *CWrapper) SyncCollectionVersions(
	ctx context.Context,
	versionIDs []string,
	opts ...options.Enumerable[options.SyncCollectionVersionsOptions],
) error {
	opt := utils.NewOptions(opts...)
	versions := C.CString(strings.Join(versionIDs, ","))
	defer C.free(unsafe.Pointer(versions))

	deadline, hasDeadline := ctx.Deadline()
	timerStr := ""
	if hasDeadline {
		timerStr = time.Until(deadline).String()
	}
	cTimerStr := C.CString(timerStr)
	defer C.free(unsafe.Pointer(cTimerStr))

	cIdentity := optionToUintptr(opt.GetIdentity())
	defer C.FreeIdentity(cIdentity)

	res := ConvertAndFreeCResult(
		C.SyncP2PCollectionVersions(C.uintptr_t(w.handle), versions, cTimerStr, cIdentity))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (w *CWrapper) SyncBranchableCollection(
	ctx context.Context,
	collectionID string,
	opts ...options.Enumerable[options.SyncBranchableCollectionOptions],
) error {
	opt := utils.NewOptions(opts...)
	cCollectionID := C.CString(collectionID)
	defer C.free(unsafe.Pointer(cCollectionID))

	deadline, hasDeadline := ctx.Deadline()
	timerStr := ""
	if hasDeadline {
		timerStr = time.Until(deadline).String()
	}
	cTimerStr := C.CString(timerStr)
	defer C.free(unsafe.Pointer(cTimerStr))

	cIdentity := optionToUintptr(opt.GetIdentity())
	defer C.FreeIdentity(cIdentity)

	res := ConvertAndFreeCResult(
		C.SyncP2PBranchableCollection(C.uintptr_t(w.handle), cCollectionID, cTimerStr, cIdentity))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (w *CWrapper) BasicImport(ctx context.Context, filepath string) error {
	panic("not implemented")
}

func (w *CWrapper) BasicExport(
	ctx context.Context,
	filepath string,
	opts ...options.Enumerable[options.BasicExportOptions],
) error {
	panic("not implemented")
}

func (w *CWrapper) AddCollection(
	ctx context.Context,
	sdl string,
	opts ...options.Enumerable[options.AddCollectionOptions],
) ([]client.CollectionVersion, error) {
	// Attach transaction to context if one was passed in
	var txn datastore.Txn
	gotTxn, hadTxn := datastore.CtxTryGetTxn(ctx)
	if hadTxn {
		txn = gotTxn
	} else {
		clientTxn, _ := w.NewTxn(false)
		var ok bool
		txn, ok = clientTxn.(datastore.Txn)
		if !ok {
			return nil, errors.New("failed to cast clientTxn to datastore.Txn")
		}
	}
	ctx = datastore.CtxSetTxn(ctx, txn)

	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())
	defer C.FreeIdentity(cIdentity)
	cSDL := C.CString(sdl)
	defer C.free(unsafe.Pointer(cSDL))

	callHandle := getNodeOrTxnHandle(w.handle, ctx)
	res := ConvertAndFreeCResult(C.AddCollection(callHandle, cSDL, cIdentity))

	if res.Status != 0 {
		return nil, errors.New(res.Error)
	}

	collectionVersions, err := unmarshalResult[[]client.CollectionVersion](res.Value)
	if err != nil {
		return nil, err
	}

	if !hadTxn {
		defer txn.Discard()
		_ = txn.Commit()
	}

	return collectionVersions, nil
}

func (w *CWrapper) AddDACPolicy(
	ctx context.Context,
	policy string,
	opts ...options.Enumerable[options.AddDACPolicyOptions],
) (client.AddPolicyResult, error) {
	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())
	defer C.FreeIdentity(cIdentity)
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
	opts ...options.Enumerable[options.AddDACActorRelationshipOptions],
) (client.AddActorRelationshipResult, error) {
	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())
	cCollectionName := C.CString(collectionName)
	cDocID := C.CString(docID)
	cRelation := C.CString(relation)
	cTargetActor := C.CString(targetActor)
	defer C.free(unsafe.Pointer(cCollectionName))
	defer C.free(unsafe.Pointer(cDocID))
	defer C.free(unsafe.Pointer(cRelation))
	defer C.free(unsafe.Pointer(cTargetActor))
	defer C.FreeIdentity(cIdentity)

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
	opts ...options.Enumerable[options.DeleteDACActorRelationshipOptions],
) (client.DeleteActorRelationshipResult, error) {
	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())
	cCollectionName := C.CString(collectionName)
	cDocID := C.CString(docID)
	cRelation := C.CString(relation)
	cTargetActor := C.CString(targetActor)
	defer C.free(unsafe.Pointer(cCollectionName))
	defer C.free(unsafe.Pointer(cDocID))
	defer C.free(unsafe.Pointer(cRelation))
	defer C.free(unsafe.Pointer(cTargetActor))
	defer C.FreeIdentity(cIdentity)

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

func (w *CWrapper) GetNACStatus(
	ctx context.Context,
	opts ...options.Enumerable[options.GetNACStatusOptions],
) (client.NACStatusResult, error) {
	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())
	defer C.FreeIdentity(cIdentity)

	callHandle := getNodeOrTxnHandle(w.handle, ctx)
	res := ConvertAndFreeCResult(C.ACPGetNACStatus(callHandle, cIdentity))

	if res.Status != 0 {
		return client.NACStatusResult{}, errors.New(res.Error)
	}
	return unmarshalResult[client.NACStatusResult](res.Value)
}

func (w *CWrapper) ReEnableNAC(ctx context.Context, opts ...options.Enumerable[options.ReEnableNACOptions]) error {
	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())
	defer C.FreeIdentity(cIdentity)

	callHandle := getNodeOrTxnHandle(w.handle, ctx)
	res := ConvertAndFreeCResult(C.ACPReEnableNAC(callHandle, cIdentity))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (w *CWrapper) DisableNAC(ctx context.Context, opts ...options.Enumerable[options.DisableNACOptions]) error {
	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())
	defer C.FreeIdentity(cIdentity)

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
	opts ...options.Enumerable[options.AddNACActorRelationshipOptions],
) (client.AddActorRelationshipResult, error) {
	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())
	cRelation := C.CString(relation)
	cTargetActor := C.CString(targetActor)
	defer C.free(unsafe.Pointer(cRelation))
	defer C.free(unsafe.Pointer(cTargetActor))
	defer C.FreeIdentity(cIdentity)

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
	opts ...options.Enumerable[options.DeleteNACActorRelationshipOptions],
) (client.DeleteActorRelationshipResult, error) {
	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())
	cRelation := C.CString(relation)
	cTargetActor := C.CString(targetActor)
	defer C.free(unsafe.Pointer(cRelation))
	defer C.free(unsafe.Pointer(cTargetActor))
	defer C.FreeIdentity(cIdentity)

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
	opts ...options.Enumerable[options.PatchCollectionOptions],
) error {
	cPatch := C.CString(patch)
	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())
	cVersion := C.CString("")
	cCollectionID := C.CString("")
	cName := C.CString("")
	defer C.free(unsafe.Pointer(cPatch))
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cName))
	defer C.FreeIdentity(cIdentity)

	migrationStr, migrationErr := optionToString(migration)
	if migrationErr != nil {
		return migrationErr
	}
	cMigration := C.CString(migrationStr)
	defer C.free(unsafe.Pointer(cMigration))

	callHandle := getNodeOrTxnHandle(w.handle, ctx)
	res := ConvertAndFreeCResult(C.PatchCollection(callHandle, cPatch, cMigration, cIdentity))

	if res.Status != 0 {
		return errors.New(res.Error)
	}

	return nil
}

func (w *CWrapper) SetActiveCollectionVersion(
	ctx context.Context,
	collectionVersionID string,
	opts ...options.Enumerable[options.SetActiveCollectionVersionOptions],
) error {
	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())
	cVersion := C.CString(collectionVersionID)
	cCollectionID := C.CString("")
	cName := C.CString("")

	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cName))
	defer C.FreeIdentity(cIdentity)

	var copts C.CollectionOptions
	copts.version = cVersion
	copts.collectionID = cCollectionID
	copts.name = cName
	copts.getInactive = 0

	callHandle := getNodeOrTxnHandle(w.handle, ctx)
	res := ConvertAndFreeCResult(C.SetActiveCollection(callHandle, copts, cIdentity))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (w *CWrapper) AddView(
	ctx context.Context,
	query string,
	sdl string,
	opts ...options.Enumerable[options.AddViewOptions],
) ([]client.CollectionVersion, error) {
	opt := utils.NewOptions(opts...)

	cTransformCID := C.CString(stringFromImmutableOptionString(opt.TransformCID))
	cQuery := C.CString(query)
	cSDL := C.CString(sdl)
	defer C.free(unsafe.Pointer(cTransformCID))
	defer C.free(unsafe.Pointer(cQuery))
	defer C.free(unsafe.Pointer(cSDL))

	cIdentity := optionToUintptr(opt.GetIdentity())
	defer C.FreeIdentity(cIdentity)

	callHandle := getNodeOrTxnHandle(w.handle, ctx)
	res := ConvertAndFreeCResult(C.AddView(callHandle, cQuery, cSDL, cTransformCID, cIdentity))

	if res.Status != 0 {
		return []client.CollectionVersion{}, errors.New(res.Error)
	}

	colDefRes, err := unmarshalResult[[]client.CollectionVersion](res.Value)
	if err != nil {
		return []client.CollectionVersion{}, err
	}
	return colDefRes, nil
}

func (w *CWrapper) RefreshViews(ctx context.Context, opts ...options.Enumerable[options.RefreshViewsOptions]) error {
	opt := utils.NewOptions(opts...)
	copts := getCollectionsOptionsToCOptions(opt)
	defer C.free(unsafe.Pointer(copts.version))
	defer C.free(unsafe.Pointer(copts.collectionID))
	defer C.free(unsafe.Pointer(copts.name))

	cIdentity := optionToUintptr(opt.GetIdentity())
	defer C.FreeIdentity(cIdentity)

	callHandle := getNodeOrTxnHandle(w.handle, ctx)
	res := ConvertAndFreeCResult(C.RefreshView(callHandle, copts, cIdentity))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (w *CWrapper) SetMigration(
	ctx context.Context, config client.LensConfig, opts ...options.Enumerable[options.SetMigrationOptions],
) (string, error) {
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

	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())
	defer C.FreeIdentity(cIdentity)

	callHandle := getNodeOrTxnHandle(w.handle, ctx)
	res := ConvertAndFreeCResult(C.SetLens(callHandle, cIdentity, src, dst, lens))

	if res.Status != 0 {
		return "", errors.New(res.Error)
	}
	return res.Value, nil
}

func (w *CWrapper) AddLens(
	ctx context.Context,
	lens model.Lens,
	opts ...options.Enumerable[options.AddLensOptions],
) (string, error) {
	lensConfig, err := json.Marshal(lens)
	if err != nil {
		return "", err
	}
	lensStr := C.CString(string(lensConfig))
	defer C.free(unsafe.Pointer(lensStr))

	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())
	defer C.FreeIdentity(cIdentity)

	callHandle := getNodeOrTxnHandle(w.handle, ctx)
	res := ConvertAndFreeCResult(C.AddLens(callHandle, cIdentity, lensStr))

	if res.Status != 0 {
		return "", errors.New(res.Error)
	}
	return res.Value, nil
}

func (w *CWrapper) ListLenses(
	ctx context.Context,
	opts ...options.Enumerable[options.ListLensesOptions],
) (map[string]model.Lens, error) {
	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())
	defer C.FreeIdentity(cIdentity)

	callHandle := getNodeOrTxnHandle(w.handle, ctx)
	res := ConvertAndFreeCResult(C.ListLenses(callHandle, cIdentity))

	if res.Status != 0 {
		return nil, errors.New(res.Error)
	}

	var lenses map[string]model.Lens
	if err := json.Unmarshal([]byte(res.Value), &lenses); err != nil {
		return nil, err
	}
	return lenses, nil
}

func (w *CWrapper) GetCollectionByName(
	ctx context.Context,
	name client.CollectionName,
	opts ...options.Enumerable[options.GetCollectionByNameOptions],
) (client.Collection, error) {
	cols, err := w.GetCollections(ctx, options.GetCollections().SetCollectionName(name))
	if err != nil {
		return nil, err
	}

	if len(cols) == 0 {
		return nil, fmt.Errorf("collection with name %q not found", name)
	}

	// cols will always have length == 1 here
	return cols[0], nil
}

// getCollectionsOptionsToCOptions converts GetCollectionsOptions to C.CollectionOptions.
// The caller is responsible for freeing the C strings (version, collectionID, name).
func getCollectionsOptionsToCOptions(opts *options.GetCollectionsOptions) C.CollectionOptions {
	var name, version, collectionID string
	var getInactive C.int = 0

	if opts != nil {
		if opts.CollectionName.HasValue() {
			name = opts.CollectionName.Value()
		}
		if opts.VersionID.HasValue() {
			version = opts.VersionID.Value()
		}
		if opts.CollectionID.HasValue() {
			collectionID = opts.CollectionID.Value()
		}
		if opts.GetInactive.HasValue() && opts.GetInactive.Value() {
			getInactive = 1
		}
	}

	var copts C.CollectionOptions
	copts.version = C.CString(version)
	copts.collectionID = C.CString(collectionID)
	copts.name = C.CString(name)
	copts.getInactive = getInactive
	return copts
}

func (w *CWrapper) GetCollections(
	ctx context.Context,
	opts ...options.Enumerable[options.GetCollectionsOptions],
) ([]client.Collection, error) {
	copts := getCollectionsOptionsToCOptions(utils.NewOptions(opts...))
	defer C.free(unsafe.Pointer(copts.version))
	defer C.free(unsafe.Pointer(copts.collectionID))
	defer C.free(unsafe.Pointer(copts.name))

	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())
	defer C.FreeIdentity(cIdentity)

	callHandle := getNodeOrTxnHandle(w.handle, ctx)
	res := ConvertAndFreeCResult(C.DescribeCollection(callHandle, copts, cIdentity))

	if res.Status != 0 {
		return []client.Collection{}, errors.New(res.Error)
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

func (w *CWrapper) ListIndexes(
	ctx context.Context,
	opts ...options.Enumerable[options.ListIndexesOptions],
) (map[client.CollectionName][]client.IndexDescription, error) {
	cVersion := C.CString("")
	cCollectionID := C.CString("")
	cName := C.CString("")
	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cName))
	defer C.FreeIdentity(cIdentity)

	var copts C.CollectionOptions
	copts.version = cVersion
	copts.collectionID = cCollectionID
	copts.name = cName
	copts.getInactive = 0

	callHandle := getNodeOrTxnHandle(w.handle, ctx)
	res := ConvertAndFreeCResult(C.ListIndexes(callHandle, copts, cIdentity))

	if res.Status != 0 {
		return nil, errors.New(res.Error)
	}

	resValue, err := unmarshalResult[map[client.CollectionName][]client.IndexDescription](res.Value)
	if err != nil {
		return nil, err
	}

	return resValue, nil
}

func (w *CWrapper) ListAllEncryptedIndexes(
	ctx context.Context,
	opts ...options.Enumerable[options.ListAllEncryptedIndexesOptions],
) (map[client.CollectionName][]client.EncryptedIndexDescription, error) {
	colName := C.CString("")
	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())
	defer C.free(unsafe.Pointer(colName))
	defer C.FreeIdentity(cIdentity)

	callHandle := getNodeOrTxnHandle(w.handle, ctx)
	res := ConvertAndFreeCResult(C.ListEncryptedIndexes(callHandle, colName, cIdentity))

	if res.Status != 0 {
		return nil, errors.New(res.Error)
	}

	resValue, err := unmarshalResult[map[client.CollectionName][]client.EncryptedIndexDescription](res.Value)
	if err != nil {
		return nil, err
	}

	return resValue, nil
}

func (w *CWrapper) ExecRequest(
	ctx context.Context,
	query string,
	opts ...options.Enumerable[options.ExecRequestOptions],
) *client.RequestResult {
	execRequestOpts := utils.NewOptions(opts...)
	operation, variables, err := extractStringsFromRequestOptions(execRequestOpts)
	if err != nil {
		return &client.RequestResult{
			GQL: client.GQLResult{
				Errors: []error{err},
			},
		}
	}

	cQuery := C.CString(query)
	cIdentity := optionToUintptr(execRequestOpts.GetIdentity())
	cOperation := C.CString(operation)
	cVariables := C.CString(variables)
	defer C.free(unsafe.Pointer(cQuery))
	defer C.free(unsafe.Pointer(cOperation))
	defer C.free(unsafe.Pointer(cVariables))
	defer C.FreeIdentity(cIdentity)

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
	var cReadOnly C.int = 0
	if readOnly {
		cReadOnly = 1
	}

	res := C.CreateTransaction(C.uintptr_t(w.handle), cReadOnly)
	errText := C.GoString(res.error)
	defer C.free(unsafe.Pointer(res.error))

	if res.status != 0 {
		return nil, errors.New(errText)
	}

	handle := cgo.Handle(res.txnPtr)
	dsTxn := handle.Value().(datastore.Txn) //nolint:forcetypeassert
	retTxn := &Transaction{w, dsTxn, handle}
	return retTxn, nil
}

func (w *CWrapper) Close() {
	C.CloseNode(C.uintptr_t(w.handle))
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

func (w *CWrapper) Connect(
	ctx context.Context,
	addresses []string,
	opts ...options.Enumerable[options.ConnectOptions],
) error {
	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())
	cPeerAddresses := C.CString(strings.Join(addresses, ","))
	defer C.free(unsafe.Pointer(cPeerAddresses))
	defer C.FreeIdentity(cIdentity)
	callHandle := getNodeOrTxnHandle(w.handle, ctx)
	res := ConvertAndFreeCResult(C.ConnectP2PPeers(callHandle, cPeerAddresses, cIdentity))
	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (w *CWrapper) GetNodeIdentity(ctx context.Context) (immutable.Option[identity.PublicRawIdentity], error) {
	callHandle := getNodeOrTxnHandle(w.handle, ctx)
	res := ConvertAndFreeCResult(C.GetNodeIdentity(callHandle))

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

func (w *CWrapper) VerifySignature(
	ctx context.Context,
	blockCid string,
	pubKey crypto.PublicKey,
	opts ...options.Enumerable[options.VerifySignatureOptions],
) error {
	cPubKey := C.CString(pubKey.String())
	cKeyType := C.CString(string(pubKey.Type()))
	cBlockCid := C.CString(blockCid)
	cIdentity := optionToUintptr(utils.NewOptions(opts...).GetIdentity())

	defer C.free(unsafe.Pointer(cPubKey))
	defer C.free(unsafe.Pointer(cKeyType))
	defer C.free(unsafe.Pointer(cBlockCid))
	defer C.FreeIdentity(cIdentity)

	callHandle := getNodeOrTxnHandle(w.handle, ctx)
	res := ConvertAndFreeCResult(C.VerifyBlockSignature(callHandle, cKeyType, cPubKey, cBlockCid, cIdentity))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}
