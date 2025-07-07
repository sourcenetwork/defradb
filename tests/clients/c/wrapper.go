// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build !cshared
// +build !cshared

package cwrap

/*
#include <stdlib.h>
#include "defra_structs.h"
*/
import "C"

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unsafe"

	"github.com/sourcenetwork/defradb/client"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/event"

	"github.com/lens-vm/lens/host-go/config/model"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sourcenetwork/immutable"
)

var _ client.TxnStore = (*CWrapper)(nil)
var _ client.P2P = (*CWrapper)(nil)

type CWrapper struct{}

func NewCWrapper() *CWrapper {
	setupTests()
	return &CWrapper{}
}

func (w *CWrapper) PeerInfo() peer.AddrInfo {
	result := P2PInfo()
	defer freeCResult(result)
	if result.status != 0 {
		return peer.AddrInfo{}
	}
	addrInfo, err := unmarshalResult[peer.AddrInfo](result.value)
	if err != nil {
		return peer.AddrInfo{}
	}
	return addrInfo
}

func (w *CWrapper) SetReplicator(ctx context.Context, info peer.AddrInfo, collections ...string) error {
	cTxnID := cTxnIDFromContext(ctx)
	peerStr := info.String()
	colStr := strings.Join(collections, ",")
	cPeerStr := C.CString(peerStr)
	cColStr := C.CString(colStr)

	result := P2PsetReplicator(cColStr, cPeerStr, cTxnID)

	defer C.free(unsafe.Pointer(cPeerStr))
	defer C.free(unsafe.Pointer(cColStr))
	defer freeCResult(result)

	if result.status != 0 {
		return errors.New(C.GoString(result.error))
	}
	return nil
}

func (w *CWrapper) DeleteReplicator(ctx context.Context, info peer.AddrInfo, collections ...string) error {
	cTxnID := cTxnIDFromContext(ctx)
	peerStr := info.String()
	colStr := strings.Join(collections, ",")
	cPeerStr := C.CString(peerStr)
	cColStr := C.CString(colStr)

	result := P2PdeleteReplicator(cColStr, cPeerStr, cTxnID)

	defer C.free(unsafe.Pointer(cPeerStr))
	defer C.free(unsafe.Pointer(cColStr))
	defer freeCResult(result)

	if result.status != 0 {
		return errors.New(C.GoString(result.error))
	}
	return nil
}

func (w *CWrapper) GetAllReplicators(ctx context.Context) ([]client.Replicator, error) {
	result := P2PgetAllReplicators()
	defer freeCResult(result)

	if result.status != 0 {
		return nil, errors.New(C.GoString(result.error))
	}

	replicators, err := unmarshalResult[[]client.Replicator](result.value)
	if err != nil {
		return nil, err
	}
	return replicators, nil
}

func (w *CWrapper) AddP2PCollections(ctx context.Context, collectionIDs ...string) error {
	cTxnID := cTxnIDFromContext(ctx)
	colStr := strings.Join(collectionIDs, ",")
	cColStr := C.CString(colStr)

	result := P2PcollectionAdd(cColStr, cTxnID)
	defer C.free(unsafe.Pointer(cColStr))
	defer freeCResult(result)

	if result.status != 0 {
		return errors.New(C.GoString(result.error))
	}
	return nil
}

func (w *CWrapper) RemoveP2PCollections(ctx context.Context, collectionIDs ...string) error {
	cTxnID := cTxnIDFromContext(ctx)
	colStr := strings.Join(collectionIDs, ",")
	cColStr := C.CString(colStr)

	result := P2PcollectionRemove(cColStr, cTxnID)

	defer C.free(unsafe.Pointer(cColStr))
	defer freeCResult(result)

	if result.status != 0 {
		return errors.New(C.GoString(result.error))
	}
	return nil
}

func (w *CWrapper) GetAllP2PCollections(ctx context.Context) ([]string, error) {
	cTxnID := cTxnIDFromContext(ctx)
	result := P2PcollectionGetAll(cTxnID)
	defer freeCResult(result)
	if result.status != 0 {
		return nil, errors.New(C.GoString(result.error))
	}
	collections, err := unmarshalResult[[]string](result.value)
	if err != nil {
		return nil, err
	}
	return collections, nil
}

func (w *CWrapper) BasicImport(ctx context.Context, filepath string) error {
	panic("not implemented")
}

func (w *CWrapper) BasicExport(ctx context.Context, config *client.BackupConfig) error {
	panic("not implemented")
}

func (w *CWrapper) AddSchema(ctx context.Context, schema string) ([]client.CollectionVersion, error) {
	cTxnID := cTxnIDFromContext(ctx)
	cSchema := C.CString(schema)
	result := AddSchema(cSchema, cTxnID)

	defer C.free(unsafe.Pointer(cSchema))
	defer freeCResult(result)

	if result.status != 0 {
		return nil, errors.New(C.GoString(result.error))
	}

	collectionVersions, err := unmarshalResult[[]client.CollectionVersion](result.value)
	if err != nil {
		return nil, err
	}
	return collectionVersions, nil
}

func (w *CWrapper) AddDACPolicy(
	ctx context.Context,
	policy string,
) (client.AddPolicyResult, error) {
	cTxnID := cTxnIDFromContext(ctx)
	cIdentity := cIdentityFromContext(ctx)
	cPolicy := C.CString(policy)

	result := AcpAddPolicy(cIdentity, cPolicy, cTxnID)

	defer C.free(unsafe.Pointer(cPolicy))
	defer C.free(unsafe.Pointer(cIdentity))
	defer freeCResult(result)

	if result.status != 0 {
		return client.AddPolicyResult{}, errors.New(C.GoString(result.error))
	}

	// Unmarshall the output from JSON to client.AddPolicyResult
	addPolicyRes, err := unmarshalResult[client.AddPolicyResult](result.value)
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
	cTxnID := cTxnIDFromContext(ctx)
	cIdentity := cIdentityFromContext(ctx)
	cCollection := C.CString(collectionName)
	cDocID := C.CString(docID)
	cRelation := C.CString(relation)
	cActor := C.CString(targetActor)

	result := AcpAddRelationship(cIdentity, cCollection, cDocID, cRelation, cActor, cTxnID)

	defer C.free(unsafe.Pointer(cIdentity))
	defer C.free(unsafe.Pointer(cCollection))
	defer C.free(unsafe.Pointer(cDocID))
	defer C.free(unsafe.Pointer(cRelation))
	defer C.free(unsafe.Pointer(cActor))
	defer freeCResult(result)

	if result.status != 0 {
		return client.AddActorRelationshipResult{}, errors.New(C.GoString(result.error))
	}

	// Unmarshall the output from JSON to client.AddActorRelationshipResult
	addRelationshipRes, err := unmarshalResult[client.AddActorRelationshipResult](result.value)
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
	cTxnID := cTxnIDFromContext(ctx)
	cIdentity := cIdentityFromContext(ctx)
	cCollection := C.CString(collectionName)
	cDocID := C.CString(docID)
	cRelation := C.CString(relation)
	cActor := C.CString(targetActor)

	result := AcpDeleteRelationship(cIdentity, cCollection, cDocID, cRelation, cActor, cTxnID)

	defer C.free(unsafe.Pointer(cIdentity))
	defer C.free(unsafe.Pointer(cCollection))
	defer C.free(unsafe.Pointer(cDocID))
	defer C.free(unsafe.Pointer(cRelation))
	defer C.free(unsafe.Pointer(cActor))
	defer freeCResult(result)

	if result.status != 0 {
		return client.DeleteActorRelationshipResult{}, errors.New(C.GoString(result.error))
	}

	// Unmarshall the output from JSON to client.DeleteActorRelationshipResult
	deleteRelationshipRes, err := unmarshalResult[client.DeleteActorRelationshipResult](result.value)
	if err != nil {
		return client.DeleteActorRelationshipResult{}, err
	}
	return deleteRelationshipRes, nil
}

func (w *CWrapper) PatchSchema(
	ctx context.Context,
	patch string,
	migration immutable.Option[model.Lens],
	setAsDefaultVersion bool,
) error {
	cTxnID := cTxnIDFromContext(ctx)
	cPatch := C.CString(patch)
	cMigration, migrationErr := optionToCString(migration)
	defer C.free(unsafe.Pointer(cPatch))
	defer C.free(unsafe.Pointer(cMigration))
	if migrationErr != nil {
		return migrationErr
	}
	var cSetAsDefaultVersion C.int = 0
	if setAsDefaultVersion {
		cSetAsDefaultVersion = 1
	}

	result := PatchSchema(cPatch, cMigration, cSetAsDefaultVersion, cTxnID)

	defer freeCResult(result)

	if result.status != 0 {
		return errors.New(C.GoString(result.error))
	}

	return nil
}

func (w *CWrapper) PatchCollection(
	ctx context.Context,
	patch string,
) error {
	cTxnID := cTxnIDFromContext(ctx)
	cIdentity := cIdentityFromContext(ctx)
	cPatch := C.CString(patch)
	cVersion := C.CString("")
	cCollectionID := C.CString("")
	cName := C.CString("")

	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(cIdentity))
	defer C.free(unsafe.Pointer(cPatch))

	var opts C.CollectionOptions
	opts.tx = cTxnID
	opts.version = cVersion
	opts.collectionID = cCollectionID
	opts.name = cName
	opts.identity = cIdentity
	opts.getInactive = 0

	result := CollectionPatch(cPatch, opts)
	defer freeCResult(result)

	if result.status != 0 {
		return errors.New(C.GoString(result.error))
	}

	return nil
}

func (w *CWrapper) SetActiveSchemaVersion(ctx context.Context, schemaVersionID string) error {
	cTxnID := cTxnIDFromContext(ctx)
	cVersion := C.CString(schemaVersionID)
	result := SetActiveSchema(cVersion, cTxnID)
	defer C.free(unsafe.Pointer(cVersion))
	defer freeCResult(result)
	if result.status != 0 {
		return errors.New(C.GoString(result.error))
	}
	return nil
}

func (w *CWrapper) AddView(
	ctx context.Context,
	query string,
	sdl string,
	transform immutable.Option[model.Lens],
) ([]client.CollectionDefinition, error) {
	cTxnID := cTxnIDFromContext(ctx)
	cQuery := C.CString(query)
	cSDL := C.CString(sdl)
	cTransform, err := cStringFromLensOption(transform)

	defer C.free(unsafe.Pointer(cQuery))
	defer C.free(unsafe.Pointer(cSDL))
	defer C.free(unsafe.Pointer(cTransform))

	if err != nil {
		return []client.CollectionDefinition{}, err
	}

	result := ViewAdd(cQuery, cSDL, cTransform, cTxnID)
	defer freeCResult(result)

	if result.status != 0 {
		return []client.CollectionDefinition{}, errors.New(C.GoString(result.error))
	}

	// Unmarshall the output from JSON to []client.CollectionDefinition
	colDefRes, err := unmarshalResult[[]client.CollectionDefinition](result.value)
	if err != nil {
		return []client.CollectionDefinition{}, err
	}
	return colDefRes, nil
}

func (w *CWrapper) RefreshViews(ctx context.Context, opts client.CollectionFetchOptions) error {
	cTxnID := cTxnIDFromContext(ctx)
	cVersionID := cStringFromImmutableOptionString(opts.VersionID)
	cCollectionID := cStringFromImmutableOptionString(opts.CollectionID)
	cName := cStringFromImmutableOptionString(opts.Name)
	var cGetInactive C.int = 0
	if opts.IncludeInactive.HasValue() {
		if opts.IncludeInactive.Value() {
			cGetInactive = 1
		}
	}

	result := ViewRefresh(cName, cCollectionID, cVersionID, cGetInactive, cTxnID)

	defer C.free(unsafe.Pointer(cVersionID))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cName))
	defer freeCResult(result)

	if result.status != 0 {
		return errors.New(C.GoString(result.error))
	}
	return nil
}

func (w *CWrapper) SetMigration(ctx context.Context, config client.LensConfig) error {
	cTxnID := cTxnIDFromContext(ctx)
	cSrc := C.CString(config.SourceSchemaVersionID)
	cDst := C.CString(config.DestinationSchemaVersionID)
	lensConfig, err := json.Marshal(config.Lens)

	defer C.free(unsafe.Pointer(cSrc))
	defer C.free(unsafe.Pointer(cDst))
	if err != nil {
		return err
	}

	cLens := C.CString(string(lensConfig))

	result := LensSet(cSrc, cDst, cLens, cTxnID)
	defer C.free(unsafe.Pointer(cLens))
	defer freeCResult(result)

	if result.status != 0 {
		return errors.New(C.GoString(result.error))
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
	cTxnID := cTxnIDFromContext(ctx)
	cIdentity := cIdentityFromContext(ctx)

	var cName *C.char
	if options.Name.HasValue() {
		cName = C.CString(options.Name.Value())
	} else {
		cName = C.CString("")
	}

	var cVersion *C.char
	if options.VersionID.HasValue() {
		cVersion = C.CString(options.VersionID.Value())
	} else {
		cVersion = C.CString("")
	}

	var cCollectionID *C.char
	if options.CollectionID.HasValue() {
		cCollectionID = C.CString(options.CollectionID.Value())
	} else {
		cCollectionID = C.CString("")
	}

	var cIncludeInactive C.int = 0
	if options.IncludeInactive.HasValue() {
		if options.IncludeInactive.Value() {
			cIncludeInactive = 1
		}
	}

	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cCollectionID))
	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(cIdentity))

	var opts C.CollectionOptions
	opts.tx = cTxnID
	opts.version = cVersion
	opts.collectionID = cCollectionID
	opts.name = cName
	opts.identity = cIdentity
	opts.getInactive = cIncludeInactive

	result := CollectionDescribe(opts)
	defer freeCResult(result)

	if result.status != 0 {
		return []client.Collection{}, errors.New(C.GoString(result.error))
	}

	defs, err := unmarshalResult[[]client.CollectionDefinition](result.value)
	if err != nil {
		return nil, err
	}

	return collectionsFromDefinitions(defs)
}

func (w *CWrapper) GetSchemaByVersionID(ctx context.Context, versionID string) (client.SchemaDescription, error) {
	schemas, err := w.GetSchemas(ctx, client.SchemaFetchOptions{ID: immutable.Some(versionID)})
	if err != nil {
		return client.SchemaDescription{}, err
	}
	return schemas[0], nil
}

func (w *CWrapper) GetSchemas(
	ctx context.Context,
	options client.SchemaFetchOptions,
) ([]client.SchemaDescription, error) {
	cTxnID := cTxnIDFromContext(ctx)
	cRoot := cStringFromImmutableOptionString(options.Root)
	cVersion := cStringFromImmutableOptionString(options.ID)
	cName := cStringFromImmutableOptionString(options.Name)

	result := DescribeSchema(cName, cRoot, cVersion, cTxnID)
	defer C.free(unsafe.Pointer(cRoot))
	defer C.free(unsafe.Pointer(cVersion))
	defer C.free(unsafe.Pointer(cName))
	defer freeCResult(result)

	if result.status != 0 {
		return []client.SchemaDescription{}, errors.New(C.GoString(result.error))
	}

	res, err := unmarshalResult[[]client.SchemaDescription](result.value)
	if err != nil {
		return []client.SchemaDescription{}, errors.New(C.GoString(result.error))
	}
	return res, nil
}

func (w *CWrapper) GetAllIndexes(ctx context.Context) (map[client.CollectionName][]client.IndexDescription, error) {
	cTxnID := cTxnIDFromContext(ctx)
	cColName := C.CString("")
	result := IndexList(cColName, cTxnID)
	defer C.free(unsafe.Pointer(cColName))
	defer freeCResult(result)

	if result.status != 0 {
		return nil, errors.New(C.GoString(result.error))
	}

	res, err := unmarshalResult[map[client.CollectionName][]client.IndexDescription](result.value)
	if err != nil {
		return nil, errors.New(C.GoString(result.error))
	}

	return res, nil
}

func (w *CWrapper) ExecRequest(
	ctx context.Context,
	query string,
	opts ...client.RequestOption,
) *client.RequestResult {
	cTxnID := cTxnIDFromContext(ctx)
	cIdentity := cIdentityFromContext(ctx)
	cQuery := C.CString(query)
	cOperation, cVariables := extractCStringsFromRequestOptions(opts)
	result := ExecuteQuery(cQuery, cIdentity, cTxnID, cOperation, cVariables)

	defer C.free(unsafe.Pointer(cIdentity))
	defer C.free(unsafe.Pointer(cQuery))
	defer C.free(unsafe.Pointer(cOperation))
	defer C.free(unsafe.Pointer(cVariables))
	defer freeCResult(result)

	// Unmarshal the result into a *client.RequestResult
	raw := C.GoString(result.value)
	rawError := C.GoString(result.error)
	retval := &client.RequestResult{}
	if result.status != 0 {
		retval.GQL.Errors = append(retval.GQL.Errors, fmt.Errorf("%s", rawError))
		return retval
	}
	if err := json.Unmarshal([]byte(raw), &retval.GQL); err != nil {
		retval.GQL.Errors = append(retval.GQL.Errors, err)
	}
	return retval
}

func (w *CWrapper) NewTxn(ctx context.Context, readOnly bool) (client.Txn, error) {
	var cConcurrent C.int = 0
	var cReadOnly C.int = 0
	if readOnly {
		cReadOnly = 1
	}

	result := TransactionCreate(cConcurrent, cReadOnly)
	value := C.GoString(result.value)
	defer freeCResult(result)

	if result.status != 0 {
		return nil, errors.New(C.GoString(result.error))
	}

	var data struct {
		ID uint64 `json:"id"`
	}

	err := json.Unmarshal([]byte(value), &data)
	if err != nil {
		return nil, err
	}

	retTxn := GetTxnFromHandle(C.ulonglong(data.ID))
	retTxnCast := retTxn.(client.Txn) //nolint:forcetypeassert
	return retTxnCast, nil
}

func (w *CWrapper) NewConcurrentTxn(ctx context.Context, readOnly bool) (client.Txn, error) {
	var cConcurrent C.int = 1
	var cReadOnly C.int = 0
	if readOnly {
		cReadOnly = 1
	}

	result := TransactionCreate(cConcurrent, cReadOnly)
	value := C.GoString(result.value)
	defer freeCResult(result)

	if result.status != 0 {
		return nil, errors.New(C.GoString(result.error))
	}

	var data struct {
		ID uint64 `json:"id"`
	}

	err := json.Unmarshal([]byte(value), &data)
	if err != nil {
		return nil, err
	}

	retTxn := GetTxnFromHandle(C.ulonglong(data.ID))
	retTxnCast := retTxn.(client.Txn) //nolint:forcetypeassert
	return retTxnCast, nil
}

func (w *CWrapper) Close() {
	NodeStop()
}

func (w *CWrapper) Events() event.Bus {
	return GetGlobalNode().DB.Events()
}

func (w *CWrapper) MaxTxnRetries() int {
	return GetGlobalNode().DB.MaxTxnRetries()
}

func (w *CWrapper) PrintDump(ctx context.Context) error {
	panic("not implemented")
}

func (w *CWrapper) Connect(ctx context.Context, addr peer.AddrInfo) error {
	panic("not implemented")
}

func (w *CWrapper) GetNodeIdentity(ctx context.Context) (immutable.Option[identity.PublicRawIdentity], error) {
	result := NodeIdentity()
	defer freeCResult(result)
	valueStr := C.GoString(result.value)

	if result.status != 0 {
		return immutable.None[identity.PublicRawIdentity](), errors.New(C.GoString(result.error))
	}

	if valueStr == "Node has no identity assigned to it." {
		return immutable.None[identity.PublicRawIdentity](), nil
	}

	var res identity.PublicRawIdentity
	res, err := unmarshalResult[identity.PublicRawIdentity](result.value)
	if err != nil {
		return immutable.None[identity.PublicRawIdentity](), err
	}
	return immutable.Some(res), nil
}

func (w *CWrapper) VerifySignature(ctx context.Context, blockCid string, pubKey crypto.PublicKey) error {
	cCID := C.CString(blockCid)
	cPubKey := C.CString(pubKey.String())
	cKeyType := C.CString(string(pubKey.Type()))

	result := BlockVerifySignature(cKeyType, cPubKey, cCID)
	defer C.free(unsafe.Pointer(cCID))
	defer C.free(unsafe.Pointer(cPubKey))
	defer C.free(unsafe.Pointer(cKeyType))
	defer freeCResult(result)

	if result.status != 0 {
		return errors.New(C.GoString(result.error))
	}
	return nil
}
