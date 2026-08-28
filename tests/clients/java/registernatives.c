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

#include "jnicall.h"
#include <jni.h>

// This file contains tables of the DefraNode and DefraTransaction native methods, and the appropriate
// accessor functions to retrieve the contents of these tables. The purpose is to accommodate the binding of
// the JNI native methods directly to the C functions inside nativewrapper.c. This is the mechanism that lets
// Java-side calls to the native methods reach into the C functions already compiled into the test binary
// without loading a second libdefradb.so (i.e. without a second Go runtime.)

// Forward declarations of the JNI functions implemented in nativewrapper.c.
extern jobject JNICALL Java_source_defra_DefraNode_NodeCloseNative(JNIEnv*, jobject, jlong);
extern jobject JNICALL Java_source_defra_DefraNode_ACPAddDACPolicyNative(JNIEnv*, jobject, jlong, jlong, jstring);
extern jobject JNICALL Java_source_defra_DefraNode_ACPAddDACActorRelationshipNative(
    JNIEnv*, jobject, jlong, jlong, jstring, jstring, jstring, jstring);
extern jobject JNICALL Java_source_defra_DefraNode_ACPDeleteDACActorRelationshipNative(
    JNIEnv*, jobject, jlong, jlong, jstring, jstring, jstring, jstring);
extern jobject JNICALL Java_source_defra_DefraNode_ACPDisableNACNative(JNIEnv*, jobject, jlong, jlong);
extern jobject JNICALL Java_source_defra_DefraNode_ACPReEnableNACNative(JNIEnv*, jobject, jlong, jlong);
extern jobject JNICALL Java_source_defra_DefraNode_ACPAddNACActorRelationshipNative(
    JNIEnv*, jobject, jlong, jlong, jstring, jstring);
extern jobject JNICALL Java_source_defra_DefraNode_ACPDeleteNACActorRelationshipNative(
    JNIEnv*, jobject, jlong, jlong, jstring, jstring);
extern jobject JNICALL Java_source_defra_DefraNode_ACPGetNACStatusNative(JNIEnv*, jobject, jlong, jlong);
extern jobject JNICALL Java_source_defra_DefraNode_AddCollectionNative(JNIEnv*, jobject, jlong, jstring, jlong);
extern jobject JNICALL Java_source_defra_DefraNode_DescribeCollectionNative(JNIEnv*, jobject, jlong, jobject, jlong);
extern jobject JNICALL Java_source_defra_DefraNode_PatchCollectionNative(
    JNIEnv*, jobject, jlong, jstring, jstring, jlong);
extern jobject JNICALL Java_source_defra_DefraNode_SetActiveCollectionNative(JNIEnv*, jobject, jlong, jobject, jlong);
extern jobject JNICALL Java_source_defra_DefraNode_TruncateCollectionNative(
    JNIEnv*, jobject, jlong, jstring, jobject, jlong);
extern jobject JNICALL Java_source_defra_DefraNode_AddDocumentNative(
    JNIEnv*, jobject, jlong, jstring, jboolean, jstring, jobject, jlong);
extern jobject JNICALL Java_source_defra_DefraNode_DeleteDocumentNative(
    JNIEnv*, jobject, jlong, jstring, jstring, jobject, jlong);
extern jobject JNICALL Java_source_defra_DefraNode_GetDocumentNative(
    JNIEnv*, jobject, jlong, jstring, jboolean, jobject, jlong);
extern jobject JNICALL Java_source_defra_DefraNode_UpdateDocumentNative(
    JNIEnv*, jobject, jlong, jstring, jstring, jstring, jobject, jlong);
extern jobject JNICALL Java_source_defra_DefraNode_NewEncryptedIndexNative(
    JNIEnv*, jobject, jlong, jstring, jstring, jlong);
extern jobject JNICALL Java_source_defra_DefraNode_ListEncryptedIndexesNative(JNIEnv*, jobject, jlong, jstring, jlong);
extern jobject JNICALL Java_source_defra_DefraNode_DeleteEncryptedIndexNative(
    JNIEnv*, jobject, jlong, jstring, jstring, jlong);
extern jobject JNICALL Java_source_defra_DefraNode_NewIndexNative(
    JNIEnv*, jobject, jlong, jstring, jstring, jboolean, jstring, jobject, jlong);
extern jobject JNICALL Java_source_defra_DefraNode_ListIndexesNative(JNIEnv*, jobject, jlong, jobject, jlong);
extern jobject JNICALL Java_source_defra_DefraNode_DeleteIndexNative(JNIEnv*, jobject, jlong, jstring, jobject, jlong);
extern jobject JNICALL Java_source_defra_DefraNode_GetNodeIdentityNative(JNIEnv*, jobject, jlong);
extern jobject JNICALL Java_source_defra_DefraNode_ListActionsNative(JNIEnv*, jobject, jlong, jlong);
extern jobject JNICALL Java_source_defra_DefraNode_DeleteCollectionNative(
    JNIEnv*, jobject, jlong, jstring, jint, jlong);
extern jobject JNICALL Java_source_defra_DefraNode_SetLensNative(
    JNIEnv*, jobject, jlong, jlong, jstring, jstring, jstring);
extern jobject JNICALL Java_source_defra_DefraNode_AddLensNative(JNIEnv*, jobject, jlong, jlong, jstring);
extern jobject JNICALL Java_source_defra_DefraNode_ListLensesNative(JNIEnv*, jobject, jlong, jlong);
extern jobject JNICALL Java_source_defra_DefraNode_GetP2PInfoNative(JNIEnv*, jobject, jlong, jlong);
extern jobject JNICALL Java_source_defra_DefraNode_ListP2PActivePeersNative(JNIEnv*, jobject, jlong, jlong);
extern jobject JNICALL Java_source_defra_DefraNode_ListP2PReplicatorsNative(JNIEnv*, jobject, jlong, jlong);
extern jobject JNICALL Java_source_defra_DefraNode_AddP2PReplicatorNative(
    JNIEnv*, jobject, jlong, jstring, jstring, jlong);
extern jobject JNICALL Java_source_defra_DefraNode_DeleteP2PReplicatorNative(
    JNIEnv*, jobject, jlong, jstring, jstring, jlong);
extern jobject JNICALL Java_source_defra_DefraNode_AddP2PCollectionNative(JNIEnv*, jobject, jlong, jstring, jlong);
extern jobject JNICALL Java_source_defra_DefraNode_DeleteP2PCollectionNative(JNIEnv*, jobject, jlong, jstring, jlong);
extern jobject JNICALL Java_source_defra_DefraNode_ListP2PCollectionsNative(JNIEnv*, jobject, jlong, jlong);
extern jobject JNICALL Java_source_defra_DefraNode_AddP2PDocumentNative(JNIEnv*, jobject, jlong, jstring, jlong);
extern jobject JNICALL Java_source_defra_DefraNode_DeleteP2PDocumentNative(JNIEnv*, jobject, jlong, jstring, jlong);
extern jobject JNICALL Java_source_defra_DefraNode_ListP2PDocumentsNative(JNIEnv*, jobject, jlong, jlong);
extern jobject JNICALL Java_source_defra_DefraNode_SyncP2PDocumentsNative(
    JNIEnv*, jobject, jlong, jstring, jstring, jstring, jlong);
extern jobject JNICALL Java_source_defra_DefraNode_SyncP2PCollectionVersionsNative(
    JNIEnv*, jobject, jlong, jstring, jstring, jlong);
extern jobject JNICALL Java_source_defra_DefraNode_SyncP2PBranchableCollectionNative(
    JNIEnv*, jobject, jlong, jstring, jstring, jlong);
extern jobject JNICALL Java_source_defra_DefraNode_ConnectP2PPeersNative(JNIEnv*, jobject, jlong, jstring, jlong);
extern jobject JNICALL Java_source_defra_DefraNode_DisconnectP2PPeersNative(JNIEnv*, jobject, jlong, jstring, jlong);
extern jobject JNICALL Java_source_defra_DefraNode_ExecuteQueryNative(
    JNIEnv*, jobject, jlong, jstring, jlong, jstring, jstring);
extern jobject JNICALL Java_source_defra_DefraNode_PollSubscriptionNative(JNIEnv*, jobject, jstring);
extern jobject JNICALL Java_source_defra_DefraNode_CloseSubscriptionNative(JNIEnv*, jobject, jstring);
extern jobject JNICALL Java_source_defra_DefraNode_AddViewNative(
    JNIEnv*, jobject, jlong, jstring, jstring, jstring, jlong);
extern jobject JNICALL Java_source_defra_DefraNode_RefreshViewNative(JNIEnv*, jobject, jlong, jobject, jlong);
extern jobject JNICALL Java_source_defra_DefraNode_VerifyBlockSignatureNative(
    JNIEnv*, jobject, jlong, jstring, jstring, jstring, jlong);
extern jobject JNICALL Java_source_defra_DefraNode_TransactionCreateNative(JNIEnv*, jobject, jlong, jboolean);
extern jobject JNICALL Java_source_defra_DefraTransaction_TransactionCommitNative(JNIEnv*, jobject, jlong);
extern void JNICALL Java_source_defra_DefraTransaction_TransactionDiscardNative(JNIEnv*, jobject, jlong);
extern jobject JNICALL Java_source_defra_DefraTransaction_ACPAddDACPolicyNative(
    JNIEnv*, jobject, jlong, jlong, jstring);
extern jobject JNICALL Java_source_defra_DefraTransaction_ACPAddDACActorRelationshipNative(
    JNIEnv*, jobject, jlong, jlong, jstring, jstring, jstring, jstring);
extern jobject JNICALL Java_source_defra_DefraTransaction_ACPDeleteDACActorRelationshipNative(
    JNIEnv*, jobject, jlong, jlong, jstring, jstring, jstring, jstring);
extern jobject JNICALL Java_source_defra_DefraTransaction_ACPDisableNACNative(JNIEnv*, jobject, jlong, jlong);
extern jobject JNICALL Java_source_defra_DefraTransaction_ACPReEnableNACNative(JNIEnv*, jobject, jlong, jlong);
extern jobject JNICALL Java_source_defra_DefraTransaction_ACPAddNACActorRelationshipNative(
    JNIEnv*, jobject, jlong, jlong, jstring, jstring);
extern jobject JNICALL Java_source_defra_DefraTransaction_ACPDeleteNACActorRelationshipNative(
    JNIEnv*, jobject, jlong, jlong, jstring, jstring);
extern jobject JNICALL Java_source_defra_DefraTransaction_ACPGetNACStatusNative(JNIEnv*, jobject, jlong, jlong);
extern jobject JNICALL Java_source_defra_DefraTransaction_AddCollectionNative(JNIEnv*, jobject, jlong, jstring, jlong);
extern jobject JNICALL Java_source_defra_DefraTransaction_DescribeCollectionNative(
    JNIEnv*, jobject, jlong, jobject, jlong);
extern jobject JNICALL Java_source_defra_DefraTransaction_PatchCollectionNative(
    JNIEnv*, jobject, jlong, jstring, jstring, jlong);
extern jobject JNICALL Java_source_defra_DefraTransaction_SetActiveCollectionNative(
    JNIEnv*, jobject, jlong, jobject, jlong);
extern jobject JNICALL Java_source_defra_DefraTransaction_TruncateCollectionNative(
    JNIEnv*, jobject, jlong, jstring, jobject, jlong);
extern jobject JNICALL Java_source_defra_DefraTransaction_AddDocumentNative(
    JNIEnv*, jobject, jlong, jstring, jboolean, jstring, jobject, jlong);
extern jobject JNICALL Java_source_defra_DefraTransaction_DeleteDocumentNative(
    JNIEnv*, jobject, jlong, jstring, jstring, jobject, jlong);
extern jobject JNICALL Java_source_defra_DefraTransaction_GetDocumentNative(
    JNIEnv*, jobject, jlong, jstring, jboolean, jobject, jlong);
extern jobject JNICALL Java_source_defra_DefraTransaction_UpdateDocumentNative(
    JNIEnv*, jobject, jlong, jstring, jstring, jstring, jobject, jlong);
extern jobject JNICALL Java_source_defra_DefraTransaction_NewEncryptedIndexNative(
    JNIEnv*, jobject, jlong, jstring, jstring, jlong);
extern jobject JNICALL Java_source_defra_DefraTransaction_ListEncryptedIndexesNative(
    JNIEnv*, jobject, jlong, jstring, jlong);
extern jobject JNICALL Java_source_defra_DefraTransaction_DeleteEncryptedIndexNative(
    JNIEnv*, jobject, jlong, jstring, jstring, jlong);
extern jobject JNICALL Java_source_defra_DefraTransaction_NewIndexNative(
    JNIEnv*, jobject, jlong, jstring, jstring, jboolean, jstring, jobject, jlong);
extern jobject JNICALL Java_source_defra_DefraTransaction_ListIndexesNative(JNIEnv*, jobject, jlong, jobject, jlong);
extern jobject JNICALL Java_source_defra_DefraTransaction_DeleteIndexNative(
    JNIEnv*, jobject, jlong, jstring, jobject, jlong);
extern jobject JNICALL Java_source_defra_DefraTransaction_SetLensNative(
    JNIEnv*, jobject, jlong, jlong, jstring, jstring, jstring);
extern jobject JNICALL Java_source_defra_DefraTransaction_AddLensNative(JNIEnv*, jobject, jlong, jlong, jstring);
extern jobject JNICALL Java_source_defra_DefraTransaction_ListLensesNative(JNIEnv*, jobject, jlong, jlong);
extern jobject JNICALL Java_source_defra_DefraTransaction_GetP2PInfoNative(JNIEnv*, jobject, jlong, jlong);
extern jobject JNICALL Java_source_defra_DefraTransaction_ListP2PReplicatorsNative(JNIEnv*, jobject, jlong, jlong);
extern jobject JNICALL Java_source_defra_DefraTransaction_AddP2PReplicatorNative(
    JNIEnv*, jobject, jlong, jstring, jstring, jlong);
extern jobject JNICALL Java_source_defra_DefraTransaction_ConnectP2PPeersNative(
    JNIEnv*, jobject, jlong, jstring, jlong);
extern jobject JNICALL Java_source_defra_DefraTransaction_DisconnectP2PPeersNative(
    JNIEnv*, jobject, jlong, jstring, jlong);
extern jobject JNICALL Java_source_defra_DefraTransaction_ExecuteQueryNative(
    JNIEnv*, jobject, jlong, jstring, jlong, jstring, jstring);
extern jobject JNICALL Java_source_defra_DefraTransaction_AddViewNative(
    JNIEnv*, jobject, jlong, jstring, jstring, jstring, jlong);
extern jobject JNICALL Java_source_defra_DefraTransaction_RefreshViewNative(JNIEnv*, jobject, jlong, jobject, jlong);
extern jobject JNICALL Java_source_defra_DefraTransaction_VerifyBlockSignatureNative(
    JNIEnv*, jobject, jlong, jstring, jstring, jstring, jlong);

// nodeNativeMethods is the single source of truth for DefraNode's native method table: used both to register them (by
// way of defra_register_node_natives) and, via the accessors below, for jvm.go to look up each one's jmethodID by name.
static JNINativeMethod nodeNativeMethods[] = {
        {"NodeCloseNative", "(J)Lsource/defra/DefraResult;", (void*)Java_source_defra_DefraNode_NodeCloseNative},
        {"ACPAddDACPolicyNative", "(JJLjava/lang/String;)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_ACPAddDACPolicyNative},
        {"ACPAddDACActorRelationshipNative",
            "(JJLjava/lang/String;Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_ACPAddDACActorRelationshipNative},
        {"ACPDeleteDACActorRelationshipNative",
            "(JJLjava/lang/String;Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_ACPDeleteDACActorRelationshipNative},
        {"ACPDisableNACNative", "(JJ)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_ACPDisableNACNative},
        {"ACPReEnableNACNative", "(JJ)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_ACPReEnableNACNative},
        {"ACPAddNACActorRelationshipNative", "(JJLjava/lang/String;Ljava/lang/String;)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_ACPAddNACActorRelationshipNative},
        {"ACPDeleteNACActorRelationshipNative", "(JJLjava/lang/String;Ljava/lang/String;)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_ACPDeleteNACActorRelationshipNative},
        {"ACPGetNACStatusNative", "(JJ)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_ACPGetNACStatusNative},
        {"AddCollectionNative", "(JLjava/lang/String;J)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_AddCollectionNative},
        {"DescribeCollectionNative", "(JLsource/defra/DefraCollectionOptions;J)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_DescribeCollectionNative},
        {"PatchCollectionNative", "(JLjava/lang/String;Ljava/lang/String;J)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_PatchCollectionNative},
        {"SetActiveCollectionNative", "(JLsource/defra/DefraCollectionOptions;J)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_SetActiveCollectionNative},
        {"TruncateCollectionNative",
            "(JLjava/lang/String;Lsource/defra/DefraCollectionOptions;J)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_TruncateCollectionNative},
        {"AddDocumentNative",
            "(JLjava/lang/String;ZLjava/lang/String;Lsource/defra/DefraCollectionOptions;J)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_AddDocumentNative},
        {"DeleteDocumentNative",
            "(JLjava/lang/String;Ljava/lang/String;Lsource/defra/DefraCollectionOptions;J)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_DeleteDocumentNative},
        {"GetDocumentNative", "(JLjava/lang/String;ZLsource/defra/DefraCollectionOptions;J)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_GetDocumentNative},
        {"UpdateDocumentNative",
            "(JLjava/lang/String;Ljava/lang/String;Ljava/lang/String;Lsource/defra/"
            "DefraCollectionOptions;J)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_UpdateDocumentNative},
        {"NewEncryptedIndexNative", "(JLjava/lang/String;Ljava/lang/String;J)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_NewEncryptedIndexNative},
        {"ListEncryptedIndexesNative", "(JLjava/lang/String;J)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_ListEncryptedIndexesNative},
        {"DeleteEncryptedIndexNative", "(JLjava/lang/String;Ljava/lang/String;J)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_DeleteEncryptedIndexNative},
        {"NewIndexNative",
            "(JLjava/lang/String;Ljava/lang/String;ZLjava/lang/String;Lsource/defra/DefraCollectionOptions;J)"
            "Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_NewIndexNative},
        {"ListIndexesNative", "(JLsource/defra/DefraCollectionOptions;J)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_ListIndexesNative},
        {"DeleteIndexNative", "(JLjava/lang/String;Lsource/defra/DefraCollectionOptions;J)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_DeleteIndexNative},
        {"GetNodeIdentityNative", "(J)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_GetNodeIdentityNative},
        {"ListActionsNative", "(JJ)Lsource/defra/DefraResult;", (void*)Java_source_defra_DefraNode_ListActionsNative},
        {"DeleteCollectionNative", "(JLjava/lang/String;IJ)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_DeleteCollectionNative},
        {"SetLensNative", "(JJLjava/lang/String;Ljava/lang/String;Ljava/lang/String;)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_SetLensNative},
        {"AddLensNative", "(JJLjava/lang/String;)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_AddLensNative},
        {"ListLensesNative", "(JJ)Lsource/defra/DefraResult;", (void*)Java_source_defra_DefraNode_ListLensesNative},
        {"GetP2PInfoNative", "(JJ)Lsource/defra/DefraResult;", (void*)Java_source_defra_DefraNode_GetP2PInfoNative},
        {"ListP2PActivePeersNative", "(JJ)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_ListP2PActivePeersNative},
        {"ListP2PReplicatorsNative", "(JJ)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_ListP2PReplicatorsNative},
        {"AddP2PReplicatorNative", "(JLjava/lang/String;Ljava/lang/String;J)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_AddP2PReplicatorNative},
        {"DeleteP2PReplicatorNative", "(JLjava/lang/String;Ljava/lang/String;J)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_DeleteP2PReplicatorNative},
        {"AddP2PCollectionNative", "(JLjava/lang/String;J)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_AddP2PCollectionNative},
        {"DeleteP2PCollectionNative", "(JLjava/lang/String;J)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_DeleteP2PCollectionNative},
        {"ListP2PCollectionsNative", "(JJ)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_ListP2PCollectionsNative},
        {"AddP2PDocumentNative", "(JLjava/lang/String;J)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_AddP2PDocumentNative},
        {"DeleteP2PDocumentNative", "(JLjava/lang/String;J)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_DeleteP2PDocumentNative},
        {"ListP2PDocumentsNative", "(JJ)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_ListP2PDocumentsNative},
        {"SyncP2PDocumentsNative",
            "(JLjava/lang/String;Ljava/lang/String;Ljava/lang/String;J)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_SyncP2PDocumentsNative},
        {"SyncP2PCollectionVersionsNative", "(JLjava/lang/String;Ljava/lang/String;J)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_SyncP2PCollectionVersionsNative},
        {"SyncP2PBranchableCollectionNative", "(JLjava/lang/String;Ljava/lang/String;J)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_SyncP2PBranchableCollectionNative},
        {"ConnectP2PPeersNative", "(JLjava/lang/String;J)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_ConnectP2PPeersNative},
        {"DisconnectP2PPeersNative", "(JLjava/lang/String;J)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_DisconnectP2PPeersNative},
        {"ExecuteQueryNative", "(JLjava/lang/String;JLjava/lang/String;Ljava/lang/String;)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_ExecuteQueryNative},
        {"PollSubscriptionNative", "(Ljava/lang/String;)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_PollSubscriptionNative},
        {"CloseSubscriptionNative", "(Ljava/lang/String;)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_CloseSubscriptionNative},
        {"AddViewNative", "(JLjava/lang/String;Ljava/lang/String;Ljava/lang/String;J)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_AddViewNative},
        {"RefreshViewNative", "(JLsource/defra/DefraCollectionOptions;J)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_RefreshViewNative},
        {"VerifyBlockSignatureNative",
            "(JLjava/lang/String;Ljava/lang/String;Ljava/lang/String;J)Lsource/defra/DefraResult;",
            (void*)Java_source_defra_DefraNode_VerifyBlockSignatureNative},
        {"TransactionCreateNative", "(JZ)Lsource/defra/DefraTransactionResult;",
            (void*)Java_source_defra_DefraNode_TransactionCreateNative},
};
static const int nodeNativeMethodsCount = (int)(sizeof(nodeNativeMethods) / sizeof(nodeNativeMethods[0]));

// defra_register_node_natives passes the node native methods table (and the given class, cls) into
// defra_register_natives inside jnicall.c, resulting in JNI binding each entry to a C function pointer.
int defra_register_node_natives(JNIEnv* env, jclass cls, char* errbuf, int errbufLen) {
    return defra_register_natives(env, cls, nodeNativeMethods, nodeNativeMethodsCount, errbuf, errbufLen);
}

// defra_node_native_method_count returns the number of native node methods
int defra_node_native_method_count(void) {
    return nodeNativeMethodsCount;
}

// defra_node_native_method_name is an accessor letting Go (which can't see the static
// nodeNativeMethods table directly) read back entry i's name by index.
const char* defra_node_native_method_name(int i) {
    return nodeNativeMethods[i].name;
}

// defra_node_native_method_signature is an accessor letting Go (which can't see the static
// nodeNativeMethods table directly) read back entry i's name by index.
const char* defra_node_native_method_signature(int i) {
    return nodeNativeMethods[i].signature;
}

// transactionNativeMethods is the single source of truth for the subset of DefraTransaction's native methods
// that this client registers.
static JNINativeMethod transactionNativeMethods[] = {
    {"TransactionCommitNative", "(J)Lsource/defra/DefraResult;",
        (void*)Java_source_defra_DefraTransaction_TransactionCommitNative},
    {"TransactionDiscardNative", "(J)V", (void*)Java_source_defra_DefraTransaction_TransactionDiscardNative},
    {"ACPAddDACPolicyNative", "(JJLjava/lang/String;)Lsource/defra/DefraResult;",
        (void*)Java_source_defra_DefraTransaction_ACPAddDACPolicyNative},
    {"ACPAddDACActorRelationshipNative",
        "(JJLjava/lang/String;Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;)Lsource/defra/DefraResult;",
        (void*)Java_source_defra_DefraTransaction_ACPAddDACActorRelationshipNative},
    {"ACPDeleteDACActorRelationshipNative",
        "(JJLjava/lang/String;Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;)Lsource/defra/DefraResult;",
        (void*)Java_source_defra_DefraTransaction_ACPDeleteDACActorRelationshipNative},
    {"ACPDisableNACNative", "(JJ)Lsource/defra/DefraResult;",
        (void*)Java_source_defra_DefraTransaction_ACPDisableNACNative},
    {"ACPReEnableNACNative", "(JJ)Lsource/defra/DefraResult;",
        (void*)Java_source_defra_DefraTransaction_ACPReEnableNACNative},
    {"ACPAddNACActorRelationshipNative", "(JJLjava/lang/String;Ljava/lang/String;)Lsource/defra/DefraResult;",
        (void*)Java_source_defra_DefraTransaction_ACPAddNACActorRelationshipNative},
    {"ACPDeleteNACActorRelationshipNative", "(JJLjava/lang/String;Ljava/lang/String;)Lsource/defra/DefraResult;",
        (void*)Java_source_defra_DefraTransaction_ACPDeleteNACActorRelationshipNative},
    {"ACPGetNACStatusNative", "(JJ)Lsource/defra/DefraResult;",
        (void*)Java_source_defra_DefraTransaction_ACPGetNACStatusNative},
    {"AddCollectionNative", "(JLjava/lang/String;J)Lsource/defra/DefraResult;",
        (void*)Java_source_defra_DefraTransaction_AddCollectionNative},
    {"DescribeCollectionNative", "(JLsource/defra/DefraCollectionOptions;J)Lsource/defra/DefraResult;",
        (void*)Java_source_defra_DefraTransaction_DescribeCollectionNative},
    {"PatchCollectionNative", "(JLjava/lang/String;Ljava/lang/String;J)Lsource/defra/DefraResult;",
        (void*)Java_source_defra_DefraTransaction_PatchCollectionNative},
    {"SetActiveCollectionNative", "(JLsource/defra/DefraCollectionOptions;J)Lsource/defra/DefraResult;",
        (void*)Java_source_defra_DefraTransaction_SetActiveCollectionNative},
    {"TruncateCollectionNative",
        "(JLjava/lang/String;Lsource/defra/DefraCollectionOptions;J)Lsource/defra/DefraResult;",
        (void*)Java_source_defra_DefraTransaction_TruncateCollectionNative},
    {"AddDocumentNative",
        "(JLjava/lang/String;ZLjava/lang/String;Lsource/defra/DefraCollectionOptions;J)Lsource/defra/DefraResult;",
        (void*)Java_source_defra_DefraTransaction_AddDocumentNative},
    {"DeleteDocumentNative",
        "(JLjava/lang/String;Ljava/lang/String;Lsource/defra/DefraCollectionOptions;J)Lsource/defra/DefraResult;",
        (void*)Java_source_defra_DefraTransaction_DeleteDocumentNative},
    {"GetDocumentNative", "(JLjava/lang/String;ZLsource/defra/DefraCollectionOptions;J)Lsource/defra/DefraResult;",
        (void*)Java_source_defra_DefraTransaction_GetDocumentNative},
    {"UpdateDocumentNative",
        "(JLjava/lang/String;Ljava/lang/String;Ljava/lang/String;Lsource/defra/"
        "DefraCollectionOptions;J)Lsource/defra/DefraResult;",
        (void*)Java_source_defra_DefraTransaction_UpdateDocumentNative},
    {"NewEncryptedIndexNative", "(JLjava/lang/String;Ljava/lang/String;J)Lsource/defra/DefraResult;",
        (void*)Java_source_defra_DefraTransaction_NewEncryptedIndexNative},
    {"ListEncryptedIndexesNative", "(JLjava/lang/String;J)Lsource/defra/DefraResult;",
        (void*)Java_source_defra_DefraTransaction_ListEncryptedIndexesNative},
    {"DeleteEncryptedIndexNative", "(JLjava/lang/String;Ljava/lang/String;J)Lsource/defra/DefraResult;",
        (void*)Java_source_defra_DefraTransaction_DeleteEncryptedIndexNative},
    {"NewIndexNative",
        "(JLjava/lang/String;Ljava/lang/String;ZLjava/lang/String;Lsource/defra/DefraCollectionOptions;J)"
        "Lsource/defra/DefraResult;",
        (void*)Java_source_defra_DefraTransaction_NewIndexNative},
    {"ListIndexesNative", "(JLsource/defra/DefraCollectionOptions;J)Lsource/defra/DefraResult;",
        (void*)Java_source_defra_DefraTransaction_ListIndexesNative},
    {"DeleteIndexNative", "(JLjava/lang/String;Lsource/defra/DefraCollectionOptions;J)Lsource/defra/DefraResult;",
        (void*)Java_source_defra_DefraTransaction_DeleteIndexNative},
    {"SetLensNative", "(JJLjava/lang/String;Ljava/lang/String;Ljava/lang/String;)Lsource/defra/DefraResult;",
        (void*)Java_source_defra_DefraTransaction_SetLensNative},
    {"AddLensNative", "(JJLjava/lang/String;)Lsource/defra/DefraResult;",
        (void*)Java_source_defra_DefraTransaction_AddLensNative},
    {"ListLensesNative", "(JJ)Lsource/defra/DefraResult;", (void*)Java_source_defra_DefraTransaction_ListLensesNative},
    {"GetP2PInfoNative", "(JJ)Lsource/defra/DefraResult;", (void*)Java_source_defra_DefraTransaction_GetP2PInfoNative},
    {"ListP2PReplicatorsNative", "(JJ)Lsource/defra/DefraResult;",
        (void*)Java_source_defra_DefraTransaction_ListP2PReplicatorsNative},
    {"AddP2PReplicatorNative", "(JLjava/lang/String;Ljava/lang/String;J)Lsource/defra/DefraResult;",
        (void*)Java_source_defra_DefraTransaction_AddP2PReplicatorNative},
    {"ConnectP2PPeersNative", "(JLjava/lang/String;J)Lsource/defra/DefraResult;",
        (void*)Java_source_defra_DefraTransaction_ConnectP2PPeersNative},
    {"DisconnectP2PPeersNative", "(JLjava/lang/String;J)Lsource/defra/DefraResult;",
        (void*)Java_source_defra_DefraTransaction_DisconnectP2PPeersNative},
    {"ExecuteQueryNative", "(JLjava/lang/String;JLjava/lang/String;Ljava/lang/String;)Lsource/defra/DefraResult;",
        (void*)Java_source_defra_DefraTransaction_ExecuteQueryNative},
    {"AddViewNative", "(JLjava/lang/String;Ljava/lang/String;Ljava/lang/String;J)Lsource/defra/DefraResult;",
        (void*)Java_source_defra_DefraTransaction_AddViewNative},
    {"RefreshViewNative", "(JLsource/defra/DefraCollectionOptions;J)Lsource/defra/DefraResult;",
        (void*)Java_source_defra_DefraTransaction_RefreshViewNative},
    {"VerifyBlockSignatureNative",
        "(JLjava/lang/String;Ljava/lang/String;Ljava/lang/String;J)Lsource/defra/DefraResult;",
        (void*)Java_source_defra_DefraTransaction_VerifyBlockSignatureNative},
};
static const int transactionNativeMethodsCount =
    (int)(sizeof(transactionNativeMethods) / sizeof(transactionNativeMethods[0]));

// defra_register_transaction_natives passes the transaction native methods table (and the given class, cls) into
// defra_register_natives inside jnicall.c, resulting in JNI binding each entry to a C function pointer.
int defra_register_transaction_natives(JNIEnv* env, jclass cls, char* errbuf, int errbufLen) {
    return defra_register_natives(env, cls, transactionNativeMethods, transactionNativeMethodsCount, errbuf, errbufLen);
}

// defra_transaction_native_method_count returns the number of native transaction methods
int defra_transaction_native_method_count(void) {
    return transactionNativeMethodsCount;
}

// defra_transaction_native_method_name is an accessor letting Go (which can't see the static
// transactionNativeMethods table directly) read back entry i's name by index.
const char* defra_transaction_native_method_name(int i) {
    return transactionNativeMethods[i].name;
}

// defra_transaction_native_method_name is an accessor letting Go (which can't see the static
// transactionNativeMethods table directly) read back entry i's signature by index.
const char* defra_transaction_native_method_signature(int i) {
    return transactionNativeMethods[i].signature;
}
