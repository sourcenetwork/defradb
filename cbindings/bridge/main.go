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

/*
#include "defra_structs.h"
*/
import "C"

import (
	cbindings "github.com/sourcenetwork/defradb/cbindings/logic"
)

//export ACPAddPolicy
func ACPAddPolicy(cIdentity *C.char, cPolicy *C.char, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.ACPAddPolicy(C.GoString(cIdentity), C.GoString(cPolicy), uint64(cTxnID))
	return returnC(gcr)
}

//export ACPAddRelationship
func ACPAddRelationship(cIdentity *C.char,
	cCollection *C.char,
	cDocID *C.char,
	cRelation *C.char,
	cActor *C.char,
	cTxnID C.ulonglong,
) *C.Result {
	gcr := cbindings.ACPAddRelationship(
		C.GoString(cIdentity),
		C.GoString(cCollection),
		C.GoString(cDocID),
		C.GoString(cRelation),
		C.GoString(cActor),
		uint64(cTxnID),
	)
	return returnC(gcr)
}

//export ACPDeleteRelationship
func ACPDeleteRelationship(
	cIdentity *C.char,
	cCollection *C.char,
	cDocID *C.char,
	cRelation *C.char,
	cActor *C.char,
	cTxnID C.ulonglong,
) *C.Result {
	gcr := cbindings.ACPDeleteRelationship(
		C.GoString(cIdentity),
		C.GoString(cCollection),
		C.GoString(cDocID),
		C.GoString(cRelation),
		C.GoString(cActor),
		uint64(cTxnID),
	)
	return returnC(gcr)
}

//export BlockVerifySignature
func BlockVerifySignature(cKeyType *C.char, cPublicKey *C.char, cCID *C.char) *C.Result {
	gcr := cbindings.BlockVerifySignature(C.GoString(cKeyType), C.GoString(cPublicKey), C.GoString(cCID))
	return returnC(gcr)
}

//export CollectionCreate
func CollectionCreate(
	cJSON *C.char,
	cIsEncrypted C.int,
	cEncryptedFields *C.char,
	cOptions C.CollectionOptions,
) *C.Result {
	gocOptions := convertCOptionsToGoCOptions(cOptions)
	gcr := cbindings.CollectionCreate(C.GoString(cJSON), cIsEncrypted != 0, C.GoString(cEncryptedFields), gocOptions)
	return returnC(gcr)
}

//export CollectionDelete
func CollectionDelete(cDocID *C.char, cFilter *C.char, cOptions C.CollectionOptions) *C.Result {
	gocOptions := convertCOptionsToGoCOptions(cOptions)
	gcr := cbindings.CollectionDelete(C.GoString(cDocID), C.GoString(cFilter), gocOptions)
	return returnC(gcr)
}

//export CollectionDescribe
func CollectionDescribe(cOptions C.CollectionOptions) *C.Result {
	gocOptions := convertCOptionsToGoCOptions(cOptions)
	gcr := cbindings.CollectionDescribe(gocOptions)
	return returnC(gcr)
}

//export CollectionListDocIDs
func CollectionListDocIDs(cOptions C.CollectionOptions) *C.Result {
	gocOptions := convertCOptionsToGoCOptions(cOptions)
	gcr := cbindings.CollectionListDocIDs(gocOptions)
	return returnC(gcr)
}

//export DocumentGet
func DocumentGet(cDocID *C.char, cShowDeleted C.int, cOptions C.CollectionOptions) *C.Result {
	gocOptions := convertCOptionsToGoCOptions(cOptions)
	gcr := cbindings.DocumentGet(C.GoString(cDocID), cShowDeleted != 0, gocOptions)
	return returnC(gcr)
}

//export CollectionPatch
func CollectionPatch(cPatch *C.char, cOptions C.CollectionOptions) *C.Result {
	gocOptions := convertCOptionsToGoCOptions(cOptions)
	gcr := cbindings.CollectionPatch(C.GoString(cPatch), gocOptions)
	return returnC(gcr)
}

//export CollectionUpdate
func CollectionUpdate(cDocID *C.char, cFilter *C.char, cUpdater *C.char, cOptions C.CollectionOptions) *C.Result {
	gocOptions := convertCOptionsToGoCOptions(cOptions)
	gcr := cbindings.CollectionUpdate(C.GoString(cDocID), C.GoString(cFilter), C.GoString(cUpdater), gocOptions)
	return returnC(gcr)
}

//export IdentityNew
func IdentityNew(cKeyType *C.char) *C.Result {
	gcr := cbindings.IdentityNew(C.GoString(cKeyType))
	return returnC(gcr)
}

//export NodeIdentity
func NodeIdentity() *C.Result {
	gcr := cbindings.NodeIdentity()
	return returnC(gcr)
}

//export IndexCreate
func IndexCreate(
	cCollectionName *C.char,
	cIndexName *C.char,
	cFields *C.char,
	cIsUnique C.int,
	cTxnID C.ulonglong,
) *C.Result {
	gcr := cbindings.IndexCreate(
		C.GoString(cCollectionName),
		C.GoString(cIndexName),
		C.GoString(cFields),
		cIsUnique != 0,
		uint64(cTxnID),
	)
	return returnC(gcr)
}

//export IndexList
func IndexList(cCollectionName *C.char, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.IndexList(C.GoString(cCollectionName), uint64(cTxnID))
	return returnC(gcr)
}

//export IndexDrop
func IndexDrop(cCollectionName *C.char, cIndexName *C.char, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.IndexDrop(
		C.GoString(cCollectionName),
		C.GoString(cIndexName),
		uint64(cTxnID),
	)
	return returnC(gcr)
}

//export LensSet
func LensSet(cSrc *C.char, cDst *C.char, cCfg *C.char, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.LensSet(
		C.GoString(cSrc),
		C.GoString(cDst),
		C.GoString(cCfg),
		uint64(cTxnID),
	)
	return returnC(gcr)
}

//export LensDown
func LensDown(cCollectionID *C.char, cDocuments *C.char, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.LensDown(
		C.GoString(cCollectionID),
		C.GoString(cDocuments),
		uint64(cTxnID),
	)
	return returnC(gcr)
}

//export LensUp
func LensUp(cCollectionID *C.char, cDocuments *C.char, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.LensUp(
		C.GoString(cCollectionID),
		C.GoString(cDocuments),
		uint64(cTxnID),
	)
	return returnC(gcr)
}

//export LensReload
func LensReload(cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.LensReload(uint64(cTxnID))
	return returnC(gcr)
}

//export LensSetRegistry
func LensSetRegistry(cCollectionID *C.char, cLensCfg *C.char, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.LensSetRegistry(
		C.GoString(cCollectionID),
		C.GoString(cLensCfg),
		uint64(cTxnID),
	)
	return returnC(gcr)
}

//export NodeInit
func NodeInit(cOptions C.NodeInitOptions) *C.Result {
	gocOptions := convertNodeInitOptionsToGoNodeInitOptions(cOptions)
	gcr := cbindings.NodeInit(gocOptions)
	return returnC(gcr)
}

//export NodeStart
func NodeStart() *C.Result {
	gcr := cbindings.NodeStart()
	return returnC(gcr)
}

//export NodeStop
func NodeStop() *C.Result {
	gcr := cbindings.NodeStop()
	return returnC(gcr)
}

//export VersionGet
func VersionGet(cFlagFull C.int, cFlagJSON C.int) *C.Result {
	gcr := cbindings.VersionGet(cFlagFull != 0, cFlagJSON != 0)
	return returnC(gcr)
}

//export ViewAdd
func ViewAdd(cQuery *C.char, cSDL *C.char, cTransform *C.char, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.ViewAdd(
		C.GoString(cQuery),
		C.GoString(cSDL),
		C.GoString(cTransform),
		uint64(cTxnID),
	)
	return returnC(gcr)
}

//export ViewRefresh
func ViewRefresh(
	cViewName *C.char,
	cCollectionID *C.char,
	cVersionID *C.char,
	cGetInactive C.int,
	cTxnID C.ulonglong,
) *C.Result {
	gcr := cbindings.ViewRefresh(
		C.GoString(cViewName),
		C.GoString(cCollectionID),
		C.GoString(cVersionID),
		cGetInactive != 0,
		uint64(cTxnID),
	)
	return returnC(gcr)
}

//export P2PInfo
func P2PInfo() *C.Result {
	gcr := cbindings.P2PInfo()
	return returnC(gcr)
}

//export P2PgetAllReplicators
func P2PgetAllReplicators() *C.Result {
	gcr := cbindings.P2PgetAllReplicators()
	return returnC(gcr)
}

//export P2PsetReplicator
func P2PsetReplicator(cCollections *C.char, cPeer *C.char, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.P2PsetReplicator(
		C.GoString(cCollections),
		C.GoString(cPeer),
		uint64(cTxnID),
	)
	return returnC(gcr)
}

//export P2PdeleteReplicator
func P2PdeleteReplicator(cCollections *C.char, cPeer *C.char, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.P2PdeleteReplicator(
		C.GoString(cCollections),
		C.GoString(cPeer),
		uint64(cTxnID),
	)
	return returnC(gcr)
}

//export P2PcollectionAdd
func P2PcollectionAdd(cCollections *C.char, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.P2PcollectionAdd(
		C.GoString(cCollections),
		uint64(cTxnID),
	)
	return returnC(gcr)
}

//export P2PcollectionRemove
func P2PcollectionRemove(cCollections *C.char, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.P2PcollectionRemove(
		C.GoString(cCollections),
		uint64(cTxnID),
	)
	return returnC(gcr)
}

//export P2PcollectionGetAll
func P2PcollectionGetAll(cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.P2PcollectionGetAll(uint64(cTxnID))
	return returnC(gcr)
}

//export AddSchema
func AddSchema(cSchema *C.char, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.AddSchema(C.GoString(cSchema), uint64(cTxnID))
	return returnC(gcr)
}

//export DescribeSchema
func DescribeSchema(cName *C.char, cRoot *C.char, cVersion *C.char, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.DescribeSchema(
		C.GoString(cName),
		C.GoString(cRoot),
		C.GoString(cVersion),
		uint64(cTxnID),
	)
	return returnC(gcr)
}

//export PatchSchema
func PatchSchema(cPatch *C.char, cLensConfig *C.char, cSetActive C.int, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.PatchSchema(
		C.GoString(cPatch),
		C.GoString(cLensConfig),
		cSetActive != 0,
		uint64(cTxnID),
	)
	return returnC(gcr)
}

//export SetActiveSchema
func SetActiveSchema(cVersion *C.char, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.SetActiveSchema(
		C.GoString(cVersion),
		uint64(cTxnID),
	)
	return returnC(gcr)
}

//export TransactionCreate
func TransactionCreate(cIsConcurrent C.int, cIsReadOnly C.int) *C.Result {
	gcr := cbindings.TransactionCreate(cIsConcurrent != 0, cIsReadOnly != 0)
	return returnC(gcr)
}

//export TransactionCommit
func TransactionCommit(cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.TransactionCommit(uint64(cTxnID))
	return returnC(gcr)
}

//export TransactionDiscard
func TransactionDiscard(cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.TransactionDiscard(uint64(cTxnID))
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
	cQuery *C.char,
	cIdentity *C.char,
	cTxnID C.ulonglong,
	cOperationName *C.char,
	cVariables *C.char,
) *C.Result {
	gcr := cbindings.ExecuteQuery(
		C.GoString(cQuery),
		C.GoString(cIdentity),
		uint64(cTxnID),
		C.GoString(cOperationName),
		C.GoString(cVariables),
	)
	return returnC(gcr)
}

//export P2PdocumentAdd
func P2PdocumentAdd(cCollections *C.char, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.P2PdocumentAdd(C.GoString(cCollections), uint64(cTxnID))
	return returnC(gcr)
}

//export P2PdocumentRemove
func P2PdocumentRemove(cCollections *C.char, cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.P2PdocumentRemove(C.GoString(cCollections), uint64(cTxnID))
	return returnC(gcr)
}

//export P2PdocumentGetAll
func P2PdocumentGetAll(cTxnID C.ulonglong) *C.Result {
	gcr := cbindings.P2PdocumentGetAll(uint64(cTxnID))
	return returnC(gcr)
}

//export P2PdocumentSync
func P2PdocumentSync(cCollection *C.char, cDocIDs *C.char, cTxnID C.ulonglong, cTimeout *C.char) *C.Result {
	gcr := cbindings.P2PdocumentSync(C.GoString(cCollection), C.GoString(cDocIDs), uint64(cTxnID), C.GoString(cTimeout))
	return returnC(gcr)
}

func main() {}
