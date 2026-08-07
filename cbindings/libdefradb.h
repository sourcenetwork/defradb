// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// IMPORTANT NOTE: This file must not be renamed. It is currently named to match the "libdefradb.h" that
// the c-shared build targets would produce. This is significant, because the sister repo containing the Defra Java
// SDK uses the nativewrapper.c file contained in this package to build its shared object. Therefore, the header
// that the nativewrapper.c source code includes must be of the same name that the other repo (and this one) would
// both have.

#ifndef LIBDEFRADB_H
#define LIBDEFRADB_H

#include <stdint.h>
#include "defra_structs.h"

#ifdef __cplusplus
extern "C" {
#endif

// Prototypes for every cgo-exported ("//export") Go function in this package. Each individual file  only needs this
// header included in its own cgo preamble to see the prototypes of whichever other exported functions it calls,
// preventing the need from duplicating declarations.

extern Result ACPAddDACPolicy(uintptr_t nodePtr, uintptr_t identityPtr, char* policy);
extern Result ACPAddDACActorRelationship(uintptr_t nodePtr, uintptr_t identityPtr, char* collection, char* docID, char* relation, char* actor);
extern Result ACPDeleteDACActorRelationship(uintptr_t nodePtr, uintptr_t identityPtr, char* collection, char* docID, char* relation, char* actor);
extern Result ACPDisableNAC(uintptr_t nodePtr, uintptr_t identityPtr);
extern Result ACPReEnableNAC(uintptr_t nodePtr, uintptr_t identityPtr);
extern Result ACPAddNACActorRelationship(uintptr_t nodePtr, uintptr_t identityPtr, char* relation, char* actor);
extern Result ACPDeleteNACActorRelationship(uintptr_t nodePtr, uintptr_t identityPtr, char* relation, char* actor);
extern Result ACPGetNACStatus(uintptr_t nodePtr, uintptr_t identityPtr);
extern Result ListActions(uintptr_t nodePtr, uintptr_t identityPtr);
extern Result VerifyBlockSignature(uintptr_t nodePtr, char* keyType, char* publicKey, char* cid, uintptr_t identityPtr);
extern Result AddCollection(uintptr_t nodePtr, char* sdl, uintptr_t identityPtr);
extern Result DescribeCollection(uintptr_t nodePtr, CollectionOptions opts, uintptr_t identityPtr);
extern Result PatchCollection(uintptr_t nodePtr, char* patch, char* lensConfig, uintptr_t identityPtr);
extern Result SetActiveCollection(uintptr_t nodePtr, CollectionOptions opts, uintptr_t identityPtr);
extern Result TruncateCollection(uintptr_t nodePtr, CollectionOptions opts, uintptr_t identityPtr);
extern Result DeleteCollection(uintptr_t nodePtr, char* names, int activeOnly, uintptr_t identityPtr);
extern Result AddDocument(uintptr_t nodePtr, char* jsonData, int isEncrypted, char* encryptedFields, CollectionOptions opts, uintptr_t identityPtr);
extern Result DeleteDocument(uintptr_t nodePtr, char* docIDStr, char* filterStr, CollectionOptions opts, uintptr_t identityPtr);
extern Result GetDocument(uintptr_t nodePtr, char* docIDStr, int showDeleted, CollectionOptions opts, uintptr_t identityPtr);
extern Result UpdateDocument(uintptr_t nodePtr, char* docIDStr, char* filterStr, char* updaterStr, CollectionOptions opts, uintptr_t identityPtr);
extern Result DeleteEncryptedIndex(uintptr_t nodePtr, char* collectionName, char* fieldName, uintptr_t identityPtr);
extern Result ListEncryptedIndexes(uintptr_t nodePtr, char* collectionName, uintptr_t identityPtr);
extern Result NewEncryptedIndex(uintptr_t nodePtr, char* collectionName, char* fieldName, uintptr_t identityPtr);
extern Result ExportIdentityPrivateKey(uintptr_t identityPtr);
extern void FreeIdentity(uintptr_t identityPtr);
extern NewIdentityResult NewIdentity(char* keyType);
extern NewIdentityResult NewIdentityFromPrivateKey(char* privateKeyHex);
extern Result DeleteIndex(uintptr_t nodePtr, char* indexName, CollectionOptions options, uintptr_t identityPtr);
extern Result ListIndexes(uintptr_t nodePtr, CollectionOptions options, uintptr_t identityPtr);
extern Result NewIndex(uintptr_t nodePtr, char* indexName, char* fieldsStr, int isUnique, CollectionOptions options, uintptr_t identityPtr);
extern Result AddLens(uintptr_t nodePtr, uintptr_t identityPtr, char* cfg);
extern Result ListLenses(uintptr_t nodePtr, uintptr_t identityPtr);
extern Result SetLens(uintptr_t nodePtr, uintptr_t identityPtr, char* src, char* dst, char* cfg);
extern Result CloseNode(uintptr_t nodePtr);
extern Result GetNodeIdentity(uintptr_t nodePtr);
extern NewNodeResult NewNode(NodeInitOptions cOptions);
extern Result GetNodeOptions(uintptr_t nodePtr);
extern Result AddP2PCollection(uintptr_t nodePtr, char* collections, uintptr_t identityPtr);
extern Result DeleteP2PCollection(uintptr_t nodePtr, char* collections, uintptr_t identityPtr);
extern Result ListP2PCollections(uintptr_t nodePtr, uintptr_t identityPtr);
extern Result SyncP2PBranchableCollection(uintptr_t nodePtr, char* collectionID, char* timeoutStr, uintptr_t identityPtr);
extern Result SyncP2PCollectionVersions(uintptr_t nodePtr, char* versionIDs, char* timeoutStr, uintptr_t identityPtr);
extern Result ConnectP2PPeers(uintptr_t nodePtr, char* peerAddresses, uintptr_t identityPtr);
extern Result DisconnectP2PPeers(uintptr_t nodePtr, char* peerAddresses, uintptr_t identityPtr);
extern Result AddP2PDocument(uintptr_t nodePtr, char* collections, uintptr_t identityPtr);
extern Result DeleteP2PDocument(uintptr_t nodePtr, char* collections, uintptr_t identityPtr);
extern Result ListP2PDocuments(uintptr_t nodePtr, uintptr_t identityPtr);
extern Result SyncP2PDocuments(uintptr_t nodePtr, char* collection, char* docIDs, char* timeoutStr, uintptr_t identityPtr);
extern Result GetP2PInfo(uintptr_t nodePtr, uintptr_t identityPtr);
extern Result ListP2PActivePeers(uintptr_t nodePtr, uintptr_t identityPtr);
extern Result AddP2PReplicator(uintptr_t nodePtr, char* collections, char* addresses, uintptr_t identityPtr);
extern Result DeleteP2PReplicator(uintptr_t nodePtr, char* collections, char* id, uintptr_t identityPtr);
extern Result ListP2PReplicators(uintptr_t nodePtr, uintptr_t identityPtr);
extern Result ExecuteQuery(uintptr_t nodePtr, char* query, uintptr_t identityPtr, char* operationName, char* variables);
extern Result CloseSubscription(char* id);
extern Result PollSubscription(char* id);
extern Result CommitTransaction(uintptr_t txnPtr);
extern NewTxnResult CreateTransaction(uintptr_t nodePtr, int isReadOnly);
extern void DiscardTransaction(uintptr_t txnPtr);
extern Result GetVersion(int flagFull, int flagJSON);
extern Result AddView(uintptr_t nodePtr, char* query, char* sdl, char* transformCIDStr, uintptr_t identityPtr);
extern Result RefreshView(uintptr_t nodePtr, CollectionOptions cOptions, uintptr_t identityPtr);

#ifdef __cplusplus
}
#endif

#endif // LIBDEFRADB_H
