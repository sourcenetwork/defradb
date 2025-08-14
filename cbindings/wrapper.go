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
extern Result* ACPAddDACPolicy(uintptr_t nodePtr, char* identity, char* policy);
extern Result* ACPAddDACActorRelationship(uintptr_t nodePtr, char* identity, char* collection, char* docID, char* relation, char* actor);
extern Result* ACPDeleteDACActorRelationship(uintptr_t nodePtr, char* identity, char* collection, char* docID, char* relation, char* actor);
extern Result* ACPDisableNAC(uintptr_t nodePtr, char* identity);
extern Result* ACPReEnableNAC(uintptr_t nodePtr, char* identity);
extern Result* ACPAddNACActorRelationship(uintptr_t nodePtr, char* identity, char* relation, char* actor);
extern Result* ACPDeleteNACActorRelationship(uintptr_t nodePtr, char* identity, char* relation, char* actor);
extern Result* ACPGetNACStatus(uintptr_t nodePtr, char* identity);
extern Result* BlockVerifySignature(uintptr_t nodePtr, char* keyType, char* publicKey, char* cid);
extern Result* CollectionDescribe(uintptr_t nodePtr, CollectionOptions options);
extern Result* CollectionPatch(uintptr_t nodePtr, char* patch, char* lensConfig, CollectionOptions options);
extern Result* IdentityNew(char* keyType);
extern Result* NodeIdentity(uintptr_t nodePtr);
extern Result* IndexList(uintptr_t nodePtr, char* collectionName);
extern Result* LensSet(uintptr_t nodePtr, char* src, char* dst, char* cfg);
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
extern Result* VersionGet(int flagFull, int flagJSON);
extern Result* ViewAdd(uintptr_t nodePtr, char* query, char* sdl, char* transformStr);
extern Result* ViewRefresh(uintptr_t nodePtr, char* viewNameStr, char* collectionIDStr, char* versionIDStr, int getInactive);
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

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/event"

	"github.com/libp2p/go-libp2p/core/peer"
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
	return &CWrapper{
		node,
		cgo.NewHandle(node),
	}, nil
}

func (w *CWrapper) PeerInfo() peer.AddrInfo {

	res := ConvertAndFreeCResult(C.P2PInfo(C.uintptr_t(w.handle)))

	if res.Status != 0 {
		return peer.AddrInfo{}
	}

	addrInfo, err := unmarshalResult[peer.AddrInfo](res.Value)
	if err != nil {
		return peer.AddrInfo{}
	}
	return addrInfo

}

func (w *CWrapper) SetReplicator(ctx context.Context, info peer.AddrInfo, collections ...string) error {
	peerStr := C.CString(info.String())
	colStr := C.CString(strings.Join(collections, ","))
	defer C.free(unsafe.Pointer(peerStr))
	defer C.free(unsafe.Pointer(colStr))

	res := ConvertAndFreeCResult(C.P2PsetReplicator(C.uintptr_t(w.handle), colStr, peerStr))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (w *CWrapper) DeleteReplicator(ctx context.Context, info peer.AddrInfo, collections ...string) error {
	peerStr := C.CString(info.String())
	colStr := C.CString(strings.Join(collections, ","))
	defer C.free(unsafe.Pointer(peerStr))
	defer C.free(unsafe.Pointer(colStr))

	res := ConvertAndFreeCResult(C.P2PdeleteReplicator(C.uintptr_t(w.handle), colStr, peerStr))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (w *CWrapper) GetAllReplicators(ctx context.Context) ([]client.Replicator, error) {
	res := ConvertAndFreeCResult(C.P2PgetAllReplicators(C.uintptr_t(w.handle)))

	if res.Status != 0 {
		return nil, errors.New(res.Error)
	}

	replicators, err := unmarshalResult[[]client.Replicator](res.Value)
	if err != nil {
		return nil, err
	}
	return replicators, nil
}

func (w *CWrapper) AddP2PCollections(ctx context.Context, collectionIDs ...string) error {
	colStr := C.CString(strings.Join(collectionIDs, ","))
	defer C.free(unsafe.Pointer(colStr))

	res := ConvertAndFreeCResult(C.P2PcollectionAdd(C.uintptr_t(w.handle), colStr))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (w *CWrapper) RemoveP2PCollections(ctx context.Context, collectionIDs ...string) error {
	colStr := C.CString(strings.Join(collectionIDs, ","))
	defer C.free(unsafe.Pointer(colStr))

	res := ConvertAndFreeCResult(C.P2PcollectionRemove(C.uintptr_t(w.handle), colStr))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (w *CWrapper) GetAllP2PCollections(ctx context.Context) ([]string, error) {
	res := ConvertAndFreeCResult(C.P2PcollectionGetAll(C.uintptr_t(w.handle)))

	if res.Status != 0 {
		return nil, errors.New(res.Error)
	}

	collections, err := unmarshalResult[[]string](res.Value)
	if err != nil {
		return nil, err
	}
	return collections, nil
}

func (w *CWrapper) AddP2PDocuments(ctx context.Context, docIDs ...string) error {
	docStr := C.CString(strings.Join(docIDs, ","))
	defer C.free(unsafe.Pointer(docStr))

	res := ConvertAndFreeCResult(C.P2PdocumentAdd(C.uintptr_t(w.handle), docStr))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (w *CWrapper) RemoveP2PDocuments(ctx context.Context, docIDs ...string) error {
	docStr := C.CString(strings.Join(docIDs, ","))
	defer C.free(unsafe.Pointer(docStr))

	res := ConvertAndFreeCResult(C.P2PdocumentRemove(C.uintptr_t(w.handle), docStr))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (w *CWrapper) GetAllP2PDocuments(ctx context.Context) ([]string, error) {
	res := ConvertAndFreeCResult(C.P2PdocumentGetAll(C.uintptr_t(w.handle)))

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

	res := ConvertAndFreeCResult(C.P2PdocumentSync(C.uintptr_t(w.handle), cCollectionName, docs, cTimerStr))

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
	res := ConvertAndFreeCResult(C.AddSchema(C.uintptr_t(w.handle), C.CString(schema)))

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
	identity := identityFromContext(ctx)
	cIdentity := C.CString(identity)
	cPolicy := C.CString(policy)
	defer C.free(unsafe.Pointer(cIdentity))
	defer C.free(unsafe.Pointer(cPolicy))

	res := ConvertAndFreeCResult(C.ACPAddDACPolicy(C.uintptr_t(w.handle), cIdentity, cPolicy))

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
	identity := identityFromContext(ctx)
	cIdentity := C.CString(identity)
	cCollectionName := C.CString(collectionName)
	cDocID := C.CString(docID)
	cRelation := C.CString(relation)
	cTargetActor := C.CString(targetActor)
	defer C.free(unsafe.Pointer(cIdentity))
	defer C.free(unsafe.Pointer(cCollectionName))
	defer C.free(unsafe.Pointer(cDocID))
	defer C.free(unsafe.Pointer(cRelation))
	defer C.free(unsafe.Pointer(cTargetActor))

	res := ConvertAndFreeCResult(C.ACPAddDACActorRelationship(C.uintptr_t(w.handle), cIdentity, cCollectionName, cDocID, cRelation, cTargetActor))

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

	identity := identityFromContext(ctx)
	cIdentity := C.CString(identity)
	cCollectionName := C.CString(collectionName)
	cDocID := C.CString(docID)
	cRelation := C.CString(relation)
	cTargetActor := C.CString(targetActor)
	defer C.free(unsafe.Pointer(cIdentity))
	defer C.free(unsafe.Pointer(cCollectionName))
	defer C.free(unsafe.Pointer(cDocID))
	defer C.free(unsafe.Pointer(cRelation))
	defer C.free(unsafe.Pointer(cTargetActor))

	res := ConvertAndFreeCResult(C.ACPDeleteDACActorRelationship(
		C.uintptr_t(w.handle),
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
	identity := identityFromContext(ctx)
	cIdentity := C.CString(identity)
	defer C.free(unsafe.Pointer(cIdentity))

	res := ConvertAndFreeCResult(C.ACPGetNACStatus(C.uintptr_t(w.handle), cIdentity))

	if res.Status != 0 {
		return client.NACStatusResult{}, errors.New(res.Error)
	}
	return unmarshalResult[client.NACStatusResult](res.Value)
}

func (w *CWrapper) ReEnableNAC(ctx context.Context) error {
	identity := identityFromContext(ctx)
	cIdentity := C.CString(identity)
	defer C.free(unsafe.Pointer(cIdentity))

	res := ConvertAndFreeCResult(C.ACPReEnableNAC(C.uintptr_t(w.handle), cIdentity))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (w *CWrapper) DisableNAC(ctx context.Context) error {
	identity := identityFromContext(ctx)
	cIdentity := C.CString(identity)
	defer C.free(unsafe.Pointer(cIdentity))

	res := ConvertAndFreeCResult(C.ACPDisableNAC(C.uintptr_t(w.handle), cIdentity))

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
	identity := C.CString(identityFromContext(ctx))
	cRelation := C.CString(relation)
	cTargetActor := C.CString(targetActor)
	defer C.free(unsafe.Pointer(identity))
	defer C.free(unsafe.Pointer(cRelation))
	defer C.free(unsafe.Pointer(cTargetActor))

	res := ConvertAndFreeCResult(C.ACPAddNACActorRelationship(C.uintptr_t(w.handle), identity, cRelation, cTargetActor))

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
	identity := C.CString(identityFromContext(ctx))
	cRelation := C.CString(relation)
	cTargetActor := C.CString(targetActor)
	defer C.free(unsafe.Pointer(identity))
	defer C.free(unsafe.Pointer(cRelation))
	defer C.free(unsafe.Pointer(cTargetActor))

	res := ConvertAndFreeCResult(C.ACPDeleteNACActorRelationship(C.uintptr_t(w.handle), identity, cRelation, cTargetActor))
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
	cIdentity := C.CString(identityFromContext(ctx))
	cVersion := C.CString("")
	cCollectionID := C.CString("")
	cName := C.CString("")
	defer C.free(unsafe.Pointer(cPatch))
	defer C.free(unsafe.Pointer(cIdentity))
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cName))

	var opts C.CollectionOptions
	opts.identity = cIdentity
	opts.version = cVersion
	opts.collectionID = cCollectionID
	opts.name = cName
	opts.getInactive = 0

	migrationStr, migrationErr := optionToString(migration)
	if migrationErr != nil {
		return migrationErr
	}
	cMigration := C.CString(migrationStr)
	defer C.free(unsafe.Pointer(cMigration))

	res := ConvertAndFreeCResult(C.CollectionPatch(C.uintptr_t(w.handle), cPatch, cMigration, opts))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (w *CWrapper) SetActiveCollectionVersion(ctx context.Context, schemaVersionID string) error {

	cSchemaVersionID := C.CString(schemaVersionID)
	defer C.free(unsafe.Pointer(cSchemaVersionID))

	res := ConvertAndFreeCResult(C.SetActiveCollection(C.uintptr_t(w.handle), cSchemaVersionID))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (w *CWrapper) AddView(
	ctx context.Context,
	query string,
	sdl string,
	transform immutable.Option[model.Lens],
) ([]client.CollectionDefinition, error) {

	transformStr, err := stringFromLensOption(transform)
	cTransform := C.CString(transformStr)
	cQuery := C.CString(query)
	cSDL := C.CString(sdl)
	defer C.free(unsafe.Pointer(cTransform))
	defer C.free(unsafe.Pointer(cQuery))
	defer C.free(unsafe.Pointer(cSDL))

	if err != nil {
		return []client.CollectionDefinition{}, err
	}

	res := ConvertAndFreeCResult(C.ViewAdd(C.uintptr_t(w.handle), cQuery, cSDL, cTransform))

	if res.Status != 0 {
		return []client.CollectionDefinition{}, errors.New(res.Error)
	}

	colDefRes, err := unmarshalResult[[]client.CollectionDefinition](res.Value)
	if err != nil {
		return []client.CollectionDefinition{}, err
	}
	return colDefRes, nil
}

func (w *CWrapper) RefreshViews(ctx context.Context, opts client.CollectionFetchOptions) error {
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

	res := ConvertAndFreeCResult(C.ViewRefresh(C.uintptr_t(w.handle), name, collectionID, versionID, cGetInactive))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (w *CWrapper) SetMigration(ctx context.Context, config client.LensConfig) error {
	src := C.CString(config.SourceSchemaVersionID)
	dst := C.CString(config.DestinationSchemaVersionID)
	lensConfig, err := json.Marshal(config.Lens)
	if err != nil {
		return err
	}
	lens := C.CString(string(lensConfig))
	defer C.free(unsafe.Pointer(src))
	defer C.free(unsafe.Pointer(dst))
	defer C.free(unsafe.Pointer(lens))

	res := ConvertAndFreeCResult(C.LensSet(C.uintptr_t(w.handle), src, dst, lens))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}

func (w *CWrapper) LensRegistry() client.LensRegistry {
	return &LensRegistry{}
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
	identity := identityFromContext(ctx)

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
	cIdentity := C.CString(identity)
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(cIdentity))

	var opts C.CollectionOptions
	opts.version = cVersion
	opts.collectionID = cCollectionID
	opts.name = cName
	opts.identity = cIdentity
	opts.getInactive = C.int(includeInactive)

	res := ConvertAndFreeCResult(C.CollectionDescribe(C.uintptr_t(w.handle), opts))

	if res.Status != 0 {
		return []client.Collection{}, errors.New(res.Error) //nolint:goerr113
	}

	defs, err := unmarshalResult[[]client.CollectionDefinition](res.Value)
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
	colName := C.CString("")
	defer C.free(unsafe.Pointer(colName))

	res := ConvertAndFreeCResult(C.IndexList(C.uintptr_t(w.handle), colName))

	if res.Status != 0 {
		return nil, errors.New(res.Error)
	}

	resValue, err := unmarshalResult[map[client.CollectionName][]client.IndexDescription](res.Value)
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
	identity := identityFromContext(ctx)
	operation, variables, err := extractStringsFromRequestOptions(opts)
	if err != nil {
		return &client.RequestResult{
			GQL: client.GQLResult{
				Errors: []error{err},
			},
		}
	}

	cQuery := C.CString(query)
	cIdentity := C.CString(identity)
	cOperation := C.CString(operation)
	cVariables := C.CString(variables)
	defer C.free(unsafe.Pointer(cQuery))
	defer C.free(unsafe.Pointer(cIdentity))
	defer C.free(unsafe.Pointer(cOperation))
	defer C.free(unsafe.Pointer(cVariables))

	result := C.ExecuteQuery(C.uintptr_t(w.handle), cQuery, cIdentity, cOperation, cVariables)
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

func (w *CWrapper) NewTxn(ctx context.Context, readOnly bool) (client.Txn, error) {
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

	clientTxn := cgo.Handle(res.txnPtr).Value().(client.Txn)
	retTxn := &Transaction{w, clientTxn, cgo.Handle(res.txnPtr)}

	return retTxn, nil
}

func (w *CWrapper) NewConcurrentTxn(ctx context.Context, readOnly bool) (client.Txn, error) {
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

	clientTxn := cgo.Handle(res.txnPtr).Value().(client.Txn)
	retTxn := &Transaction{w, clientTxn, cgo.Handle(res.txnPtr)}

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

func (w *CWrapper) Connect(ctx context.Context, addr peer.AddrInfo) error {
	panic("not implemented")
}

func (w *CWrapper) GetNodeIdentity(ctx context.Context) (immutable.Option[identity.PublicRawIdentity], error) {
	res := ConvertAndFreeCResult(C.NodeIdentity(C.uintptr_t(w.handle)))

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
	defer C.free(unsafe.Pointer(cPubKey))
	defer C.free(unsafe.Pointer(cKeyType))
	defer C.free(unsafe.Pointer(cBlockCid))

	res := ConvertAndFreeCResult(C.BlockVerifySignature(C.uintptr_t(w.handle), cKeyType, cPubKey, cBlockCid))

	if res.Status != 0 {
		return errors.New(res.Error)
	}
	return nil
}
