// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build cgo
// +build cgo

package main

// The following comment is to allow use of C structs in the Go code

/*
#include "defra_structs.h"
*/
import "C"

import (
	cbindings "github.com/sourcenetwork/defradb/cbindings/logic"
)

//export ACPAddDACPolicy
func ACPAddDACPolicy(n int, cIdentity *C.char, cPolicy *C.char, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.ACPAddDACPolicy(n, C.GoString(cIdentity), C.GoString(cPolicy), uint64(cTxnID))
	return returnC(gcr)
}

//export ACPAddDACActorRelationship
func ACPAddDACActorRelationship(
	n int,
	cIdentity *C.char,
	cCollection *C.char,
	cDocID *C.char,
	cRelation *C.char,
	cActor *C.char,
	cTxnID C.ulonglong,
) *C.Result {
	gcr := cbindings.ACPAddDACActorRelationship(
		n,
		C.GoString(cIdentity),
		C.GoString(cCollection),
		C.GoString(cDocID),
		C.GoString(cRelation),
		C.GoString(cActor),
		uint64(cTxnID),
	)
	return returnC(gcr)
}

//export ACPDeleteDACActorRelationship
func ACPDeleteDACActorRelationship(
	n int,
	cIdentity *C.char,
	cCollection *C.char,
	cDocID *C.char,
	cRelation *C.char,
	cActor *C.char,
	cTxnID C.ulonglong,
) *C.Result {
	gcr := cbindings.ACPDeleteDACActorRelationship(
		n,
		C.GoString(cIdentity),
		C.GoString(cCollection),
		C.GoString(cDocID),
		C.GoString(cRelation),
		C.GoString(cActor),
		uint64(cTxnID),
	)
	return returnC(gcr)
}

//export ACPDisableNAC
func ACPDisableNAC(n int, cIdentity *C.char, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.ACPDisableNAC(n, C.GoString(cIdentity), uint64(cTxnID))
	return returnC(gcr)
}

//export ACPReEnableNAC
func ACPReEnableNAC(n int, cIdentity *C.char, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.ACPReEnableNAC(n, C.GoString(cIdentity), uint64(cTxnID))
	return returnC(gcr)
}

//export ACPAddNACActorRelationship
func ACPAddNACActorRelationship(
	n int,
	cIdentity *C.char,
	cRelation *C.char,
	cActor *C.char,
	cTxnID C.ulonglong,
) *C.Result {
	gcr := cbindings.ACPAddNACActorRelationship(
		n,
		C.GoString(cIdentity),
		C.GoString(cRelation),
		C.GoString(cActor),
		uint64(cTxnID),
	)
	return returnC(gcr)
}

//export ACPDeleteNACActorRelationship
func ACPDeleteNACActorRelationship(
	n int,
	cIdentity *C.char,
	cRelation *C.char,
	cActor *C.char,
	cTxnID C.ulonglong,
) *C.Result {
	gcr := cbindings.ACPDeleteNACActorRelationship(
		n,
		C.GoString(cIdentity),
		C.GoString(cRelation),
		C.GoString(cActor),
		uint64(cTxnID),
	)
	return returnC(gcr)
}

//export ACPGetNACStatus
func ACPGetNACStatus(n int, cIdentity *C.char, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.ACPGetNACStatus(n, C.GoString(cIdentity), uint64(cTxnID))
	return returnC(gcr)
}

//export BlockVerifySignature
func BlockVerifySignature(n int, cKeyType *C.char, cPublicKey *C.char, cCID *C.char) *C.Result {
	gcr := cbindings.BlockVerifySignature(n, C.GoString(cKeyType), C.GoString(cPublicKey), C.GoString(cCID))
	return returnC(gcr)
}

//export CollectionCreate
func CollectionCreate(
	n int,
	cJSON *C.char,
	cIsEncrypted C.int,
	cEncryptedFields *C.char,
	cOptions C.CollectionOptions,
) *C.Result {
	gocOptions := convertCOptionsToGoCOptions(cOptions)
	gcr := cbindings.CollectionCreate(n, C.GoString(cJSON), cIsEncrypted != 0, C.GoString(cEncryptedFields), gocOptions)
	return returnC(gcr)
}

//export CollectionDelete
func CollectionDelete(n int, cDocID *C.char, cFilter *C.char, cOptions C.CollectionOptions) *C.Result {
	gocOptions := convertCOptionsToGoCOptions(cOptions)
	gcr := cbindings.CollectionDelete(n, C.GoString(cDocID), C.GoString(cFilter), gocOptions)
	return returnC(gcr)
}

//export CollectionDescribe
func CollectionDescribe(n int, cOptions C.CollectionOptions) *C.Result {
	gocOptions := convertCOptionsToGoCOptions(cOptions)
	gcr := cbindings.CollectionDescribe(n, gocOptions)
	return returnC(gcr)
}

//export CollectionListDocIDs
func CollectionListDocIDs(n int, cOptions C.CollectionOptions) *C.Result {
	gocOptions := convertCOptionsToGoCOptions(cOptions)
	gcr := cbindings.CollectionListDocIDs(n, gocOptions)
	return returnC(gcr)
}

//export CollectionGet
func CollectionGet(n int, cDocID *C.char, cShowDeleted C.int, cOptions C.CollectionOptions) *C.Result {
	gocOptions := convertCOptionsToGoCOptions(cOptions)
	gcr := cbindings.CollectionGet(n, C.GoString(cDocID), cShowDeleted != 0, gocOptions)
	return returnC(gcr)
}

//export CollectionPatch
func CollectionPatch(n int, cPatch *C.char, cOptions C.CollectionOptions) *C.Result {
	gocOptions := convertCOptionsToGoCOptions(cOptions)
	gcr := cbindings.CollectionPatch(n, C.GoString(cPatch), gocOptions)
	return returnC(gcr)
}

//export CollectionUpdate
func CollectionUpdate(
	n int,
	cDocID *C.char,
	cFilter *C.char,
	cUpdater *C.char,
	cOptions C.CollectionOptions,
) *C.Result {
	gocOptions := convertCOptionsToGoCOptions(cOptions)
	gcr := cbindings.CollectionUpdate(n, C.GoString(cDocID), C.GoString(cFilter), C.GoString(cUpdater), gocOptions)
	return returnC(gcr)
}

//export IdentityNew
func IdentityNew(cKeyType *C.char) *C.Result {
	gcr := cbindings.IdentityNew(C.GoString(cKeyType))
	return returnC(gcr)
}

//export NodeIdentity
func NodeIdentity(n int) *C.Result {
	gcr := cbindings.NodeIdentity(n)
	return returnC(gcr)
}

//export IndexCreate
func IndexCreate(
	n int,
	cCollectionName *C.char,
	cIndexName *C.char,
	cFields *C.char,
	cIsUnique C.int,
	cTxnID C.ulonglong,
) *C.Result {
	gcr := cbindings.IndexCreate(
		n,
		C.GoString(cCollectionName),
		C.GoString(cIndexName),
		C.GoString(cFields),
		cIsUnique != 0,
		uint64(cTxnID),
	)
	return returnC(gcr)
}

//export IndexList
func IndexList(n int, cCollectionName *C.char, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.IndexList(n, C.GoString(cCollectionName), uint64(cTxnID))
	return returnC(gcr)
}

//export IndexDrop
func IndexDrop(n int, cCollectionName *C.char, cIndexName *C.char, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.IndexDrop(
		n,
		C.GoString(cCollectionName),
		C.GoString(cIndexName),
		uint64(cTxnID),
	)
	return returnC(gcr)
}

//export LensSet
func LensSet(n int, cSrc *C.char, cDst *C.char, cCfg *C.char, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.LensSet(
		n,
		C.GoString(cSrc),
		C.GoString(cDst),
		C.GoString(cCfg),
		uint64(cTxnID),
	)
	return returnC(gcr)
}

//export LensDown
func LensDown(n int, cCollectionID *C.char, cDocuments *C.char, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.LensDown(
		n,
		C.GoString(cCollectionID),
		C.GoString(cDocuments),
		uint64(cTxnID),
	)
	return returnC(gcr)
}

//export LensUp
func LensUp(n int, cCollectionID *C.char, cDocuments *C.char, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.LensUp(
		n,
		C.GoString(cCollectionID),
		C.GoString(cDocuments),
		uint64(cTxnID),
	)
	return returnC(gcr)
}

//export LensReload
func LensReload(n int, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.LensReload(n, uint64(cTxnID))
	return returnC(gcr)
}

//export LensSetRegistry
func LensSetRegistry(n int, cCollectionID *C.char, cLensCfg *C.char, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.LensSetRegistry(
		n,
		C.GoString(cCollectionID),
		C.GoString(cLensCfg),
		uint64(cTxnID),
	)
	return returnC(gcr)
}

//export NodeInit
func NodeInit(n int, cOptions C.NodeInitOptions) *C.Result {
	gocOptions := convertNodeInitOptionsToGoNodeInitOptions(cOptions)
	gcr := cbindings.NodeInit(n, gocOptions)
	return returnC(gcr)
}

//export NodeStop
func NodeStop(n int) *C.Result {
	gcr := cbindings.NodeStop(n)
	return returnC(gcr)
}

//export VersionGet
func VersionGet(cFlagFull C.int, cFlagJSON C.int) *C.Result {
	gcr := cbindings.VersionGet(cFlagFull != 0, cFlagJSON != 0)
	return returnC(gcr)
}

//export ViewAdd
func ViewAdd(n int, cQuery *C.char, cSDL *C.char, cTransform *C.char, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.ViewAdd(
		n,
		C.GoString(cQuery),
		C.GoString(cSDL),
		C.GoString(cTransform),
		uint64(cTxnID),
	)
	return returnC(gcr)
}

//export ViewRefresh
func ViewRefresh(
	n int,
	cViewName *C.char,
	cCollectionID *C.char,
	cVersionID *C.char,
	cGetInactive C.int,
	cTxnID C.ulonglong,
) *C.Result {
	gcr := cbindings.ViewRefresh(
		n,
		C.GoString(cViewName),
		C.GoString(cCollectionID),
		C.GoString(cVersionID),
		cGetInactive != 0,
		uint64(cTxnID),
	)
	return returnC(gcr)
}

//export P2PInfo
func P2PInfo(n int) *C.Result {
	gcr := cbindings.P2PInfo(n)
	return returnC(gcr)
}

//export P2PgetAllReplicators
func P2PgetAllReplicators(n int) *C.Result {
	gcr := cbindings.P2PgetAllReplicators(n)
	return returnC(gcr)
}

//export P2PsetReplicator
func P2PsetReplicator(n int, cCollections *C.char, cPeer *C.char, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.P2PsetReplicator(
		n,
		C.GoString(cCollections),
		C.GoString(cPeer),
		uint64(cTxnID),
	)
	return returnC(gcr)
}

//export P2PdeleteReplicator
func P2PdeleteReplicator(n int, cCollections *C.char, cPeer *C.char, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.P2PdeleteReplicator(
		n,
		C.GoString(cCollections),
		C.GoString(cPeer),
		uint64(cTxnID),
	)
	return returnC(gcr)
}

//export P2PcollectionAdd
func P2PcollectionAdd(n int, cCollections *C.char, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.P2PcollectionAdd(
		n,
		C.GoString(cCollections),
		uint64(cTxnID),
	)
	return returnC(gcr)
}

//export P2PcollectionRemove
func P2PcollectionRemove(n int, cCollections *C.char, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.P2PcollectionRemove(
		n,
		C.GoString(cCollections),
		uint64(cTxnID),
	)
	return returnC(gcr)
}

//export P2PcollectionGetAll
func P2PcollectionGetAll(n int, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.P2PcollectionGetAll(n, uint64(cTxnID))
	return returnC(gcr)
}

//export AddSchema
func AddSchema(n int, cSchema *C.char, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.AddSchema(n, C.GoString(cSchema), uint64(cTxnID))
	return returnC(gcr)
}

//export DescribeSchema
func DescribeSchema(n int, cName *C.char, cRoot *C.char, cVersion *C.char, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.DescribeSchema(
		n,
		C.GoString(cName),
		C.GoString(cRoot),
		C.GoString(cVersion),
		uint64(cTxnID),
	)
	return returnC(gcr)
}

//export PatchSchema
func PatchSchema(n int, cPatch *C.char, cLensConfig *C.char, cSetActive C.int, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.PatchSchema(
		n,
		C.GoString(cPatch),
		C.GoString(cLensConfig),
		cSetActive != 0,
		uint64(cTxnID),
	)
	return returnC(gcr)
}

//export SetActiveSchema
func SetActiveSchema(n int, cVersion *C.char, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.SetActiveSchema(
		n,
		C.GoString(cVersion),
		uint64(cTxnID),
	)
	return returnC(gcr)
}

//export TransactionCreate
func TransactionCreate(n int, cIsConcurrent C.int, cIsReadOnly C.int) *C.Result {
	gcr := cbindings.TransactionCreate(n, cIsConcurrent != 0, cIsReadOnly != 0)
	return returnC(gcr)
}

//export TransactionCommit
func TransactionCommit(n int, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.TransactionCommit(n, uint64(cTxnID))
	return returnC(gcr)
}

//export TransactionDiscard
func TransactionDiscard(n int, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.TransactionDiscard(n, uint64(cTxnID))
	return returnC(gcr)
}

//export PollSubscription
func PollSubscription(cID *C.char) *C.Result {
	gcr := cbindings.PollSubscription(C.GoString(cID))
	return returnC(gcr)
}

//export CloseSubscription
func CloseSubscription(cID *C.char) *C.Result {
	gcr := cbindings.CloseSubscription(C.GoString(cID))
	return returnC(gcr)
}

//export ExecuteQuery
func ExecuteQuery(
	n int,
	cQuery *C.char,
	cIdentity *C.char,
	cTxnID C.ulonglong,
	cOperationName *C.char,
	cVariables *C.char,
) *C.Result {
	gcr := cbindings.ExecuteQuery(
		n,
		C.GoString(cQuery),
		C.GoString(cIdentity),
		uint64(cTxnID),
		C.GoString(cOperationName),
		C.GoString(cVariables),
	)
	return returnC(gcr)
}

//export P2PdocumentAdd
func P2PdocumentAdd(n int, cCollections *C.char, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.P2PdocumentAdd(n, C.GoString(cCollections), uint64(cTxnID))
	return returnC(gcr)
}

//export P2PdocumentRemove
func P2PdocumentRemove(n int, cCollections *C.char, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.P2PdocumentRemove(n, C.GoString(cCollections), uint64(cTxnID))
	return returnC(gcr)
}

//export P2PdocumentGetAll
func P2PdocumentGetAll(n int, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.P2PdocumentGetAll(n, uint64(cTxnID))
	return returnC(gcr)
}

//export P2PdocumentSync
func P2PdocumentSync(n int, cCollection *C.char, cDocIDs *C.char, cTxnID C.ulonglong, cTimeout *C.char) *C.Result {
	gcr := cbindings.P2PdocumentSync(n, C.GoString(cCollection), C.GoString(cDocIDs), uint64(cTxnID), C.GoString(cTimeout))
	return returnC(gcr)
}

// Intentionally left blank to allow CGO to build the library
func main() {}
