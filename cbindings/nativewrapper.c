// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build javaclient

#include <jni.h>
#include "libdefradb.h"
#include <stdlib.h>
#include <string.h>

// This is the canonical, ground-truth JNI native implementation for the DefraDB Java SDK. It lives here, in the
// cbindings, instead of in the separate defradb-java-sdk repo for important logistics reasons. This file is necessary 
// to build the Java test client. This would mean that two versions of the file need to be maintained and kept in
// sync with one another. Instead of that, we keep one version here, and the Java SDK project's build script retrieves
// a copy of it dynamically during build (because building the Java SDK requires that a path to the defra directory)
// be passed in.

// Forward declarations
void releaseJavaNodeInitOptions(JNIEnv* env, jobject optionsObj, NodeInitOptions opts);
void releaseJavaCollectionOptions(JNIEnv* env, jobject optionsObj, CollectionOptions opts);

// jstring_from_utf8_bytes builds a Java String from a raw, full-UTF-8-encoded
// byte buffer via new String(byte[], "UTF-8"). NewStringUTF (used elsewhere in
// this file) requires "modified UTF-8", which caps sequences at 3 bytes and
// represents supplementary-plane characters (>= U+10000) as surrogate pairs
// instead of the standard 4-byte encoding; fed a genuine 4-byte UTF-8
// sequence, it silently corrupts/truncates decoding partway through. Go's
// json.Marshal guarantees valid *standard* UTF-8 output (invalid byte runs
// get replaced with U+FFFD) but happily emits real 4-byte sequences when it
// finds them - and a multi-megabyte value embedding raw binary (e.g.
// ListLenses' embedded WASM module bytes) has a real chance of containing
// some by pure coincidence. This constructor has no such restriction.
static jstring jstring_from_utf8_bytes(JNIEnv* env, const char* utf8, size_t len) {
    if (utf8 == NULL) {
        return NULL;
    }
    jbyteArray bytes = (*env)->NewByteArray(env, (jsize)len);
    if (bytes == NULL) {
        return NULL;
    }
    (*env)->SetByteArrayRegion(env, bytes, 0, (jsize)len, (const jbyte*)utf8);
    jclass stringCls = (*env)->FindClass(env, "java/lang/String");
    jmethodID ctor = (*env)->GetMethodID(env, stringCls, "<init>", "([BLjava/lang/String;)V");
    jstring charsetName = (*env)->NewStringUTF(env, "UTF-8");
    jstring result = (jstring)(*env)->NewObject(env, stringCls, ctor, bytes, charsetName);
    (*env)->DeleteLocalRef(env, bytes);
    (*env)->DeleteLocalRef(env, charsetName);
    return result;
}

jobject returnDefraResult(JNIEnv* env, Result res) {
    jstring errorStr = res.error ? jstring_from_utf8_bytes(env, res.error, strlen(res.error)) : NULL;
    jstring valueStr = res.value ? jstring_from_utf8_bytes(env, res.value, strlen(res.value)) : NULL;
    if (res.error) free(res.error);
    if (res.value) free(res.value);
    jclass cls = (*env)->FindClass(env, "source/defra/DefraResult");
    jmethodID ctor = (*env)->GetMethodID(env, cls, "<init>", "(ILjava/lang/String;Ljava/lang/String;)V");
    jobject resultObj = (*env)->NewObject(env, cls, ctor, (jint)res.status, errorStr, valueStr);
    return resultObj;
}

jobject returnDefraNewNodeResult(JNIEnv* env, NewNodeResult res) {
    jstring errorStr = res.error ? (*env)->NewStringUTF(env, res.error) : NULL;
    if (res.error) free(res.error);
    jclass cls = (*env)->FindClass(env, "source/defra/DefraNewNodeResult");
    jmethodID ctor = (*env)->GetMethodID(env, cls, "<init>", "(ILjava/lang/String;J)V");
    jobject resultObj = (*env)->NewObject(env, cls, ctor, (jint)res.status, errorStr, (jlong)res.nodePtr);
    return resultObj;
}

jobject returnDefraIdentityResult(JNIEnv* env, NewIdentityResult res) {
    jstring errorStr = res.error ? (*env)->NewStringUTF(env, res.error) : NULL;
    if (res.error) free(res.error);
    jclass cls = (*env)->FindClass(env, "source/defra/DefraIdentityResult");
    jmethodID ctor = (*env)->GetMethodID(env, cls, "<init>", "(ILjava/lang/String;J)V");
    jobject resultObj = (*env)->NewObject(env, cls, ctor, (jint)res.status, errorStr, (jlong)res.identityPtr);
    return resultObj;
}

jobject returnDefraTransactionResult(JNIEnv* env, NewTxnResult res) {
    jstring errorStr = res.error ? (*env)->NewStringUTF(env, res.error) : NULL;
    if (res.error) free(res.error);
    jclass cls = (*env)->FindClass(env, "source/defra/DefraTransactionResult");
    jmethodID ctor = (*env)->GetMethodID(env, cls, "<init>", "(ILjava/lang/String;J)V");
    jobject resultObj = (*env)->NewObject(env, cls, ctor, (jint)res.status, errorStr, (jlong)res.txnPtr);
    return resultObj;
}

// Helper to convert a Java DefraNodeInitOptions object to a C NodeInitOptions struct
NodeInitOptions convertJavaNodeInitOptions(JNIEnv* env, jobject optionsObj) {
    NodeInitOptions opts;
    memset(&opts, 0, sizeof(NodeInitOptions));
    jclass cls = (*env)->GetObjectClass(env, optionsObj);

    // Core strings
    jfieldID fid_dbPath = (*env)->GetFieldID(env, cls, "dbPath", "Ljava/lang/String;");
    jfieldID fid_listeningAddresses = (*env)->GetFieldID(env, cls, "listeningAddresses", "Ljava/lang/String;");
    jfieldID fid_replicatorRetryIntervals = (*env)->GetFieldID(env, cls, "replicatorRetryIntervals", "Ljava/lang/String;");
    jfieldID fid_peers = (*env)->GetFieldID(env, cls, "peers", "Ljava/lang/String;");

    jstring dbPathStr = (jstring)(*env)->GetObjectField(env, optionsObj, fid_dbPath);
    jstring listeningAddressesStr = (jstring)(*env)->GetObjectField(env, optionsObj, fid_listeningAddresses);
    jstring replicatorRetryIntervalsStr = (jstring)(*env)->GetObjectField(env, optionsObj, fid_replicatorRetryIntervals);
    jstring peersStr = (jstring)(*env)->GetObjectField(env, optionsObj, fid_peers);

    opts.dbPath = dbPathStr ? (*env)->GetStringUTFChars(env, dbPathStr, 0) : NULL;
    opts.listeningAddresses = listeningAddressesStr ? (*env)->GetStringUTFChars(env, listeningAddressesStr, 0) : NULL;
    opts.replicatorRetryIntervals =
        replicatorRetryIntervalsStr ? (*env)->GetStringUTFChars(env, replicatorRetryIntervalsStr, 0) : NULL;
    opts.peers = peersStr ? (*env)->GetStringUTFChars(env, peersStr, 0) : NULL;

    // Core booleans/ints
    jfieldID fid_inMemory = (*env)->GetFieldID(env, cls, "inMemory", "Z");
    jfieldID fid_disableP2P = (*env)->GetFieldID(env, cls, "disableP2P", "Z");
    jfieldID fid_disableAPI = (*env)->GetFieldID(env, cls, "disableAPI", "Z");
    jfieldID fid_enableNodeACP = (*env)->GetFieldID(env, cls, "enableNodeACP", "Z");

    opts.inMemory = (*env)->GetBooleanField(env, optionsObj, fid_inMemory) ? 1 : 0;
    opts.disableP2P = (*env)->GetBooleanField(env, optionsObj, fid_disableP2P) ? 1 : 0;
    opts.disableAPI = (*env)->GetBooleanField(env, optionsObj, fid_disableAPI) ? 1 : 0;
    opts.enableNodeACP = (*env)->GetBooleanField(env, optionsObj, fid_enableNodeACP) ? 1 : 0;

    jfieldID fid_maxTransactionRetries = (*env)->GetFieldID(env, cls, "maxTransactionRetries", "I");
    opts.maxTransactionRetries = (*env)->GetIntField(env, optionsObj, fid_maxTransactionRetries);

    // Identity
    jfieldID fid_identity = (*env)->GetFieldID(env, cls, "identity", "Lsource/defra/DefraIdentity;");
    jobject identityObj = (*env)->GetObjectField(env, optionsObj, fid_identity);
    if (identityObj != NULL) {
        jclass identityCls = (*env)->GetObjectClass(env, identityObj);
        jfieldID fid_ptr = (*env)->GetFieldID(env, identityCls, "ptr", "J");
        opts.identityPtr = (uintptr_t)(*env)->GetLongField(env, identityObj, fid_ptr);
    }

    // Store options
    jfieldID fid_storeType = (*env)->GetFieldID(env, cls, "storeType", "Ljava/lang/String;");
    jstring storeTypeStr = (jstring)(*env)->GetObjectField(env, optionsObj, fid_storeType);
    opts.storeType = storeTypeStr ? (*env)->GetStringUTFChars(env, storeTypeStr, 0) : NULL;

    jfieldID fid_badgerFileSize = (*env)->GetFieldID(env, cls, "badgerFileSize", "J");
    opts.badgerFileSize = (*env)->GetLongField(env, optionsObj, fid_badgerFileSize);

    jfieldID fid_badgerEncryptionKey = (*env)->GetFieldID(env, cls, "badgerEncryptionKey", "[B");
    jbyteArray badgerEncryptionKeyArr = (jbyteArray)(*env)->GetObjectField(env, optionsObj, fid_badgerEncryptionKey);
    if (badgerEncryptionKeyArr != NULL) {
        opts.badgerEncryptionKey = (uint8_t*)(*env)->GetByteArrayElements(env, badgerEncryptionKeyArr, NULL);
        opts.badgerEncryptionKeyLen = (int)(*env)->GetArrayLength(env, badgerEncryptionKeyArr);
    }

    // DB options
    jfieldID fid_enableSigning = (*env)->GetFieldID(env, cls, "enableSigning", "Z");
    opts.enableSigning = (*env)->GetBooleanField(env, optionsObj, fid_enableSigning) ? 1 : 0;

    jfieldID fid_searchableEncryptionKey = (*env)->GetFieldID(env, cls, "searchableEncryptionKey", "[B");
    jbyteArray searchableEncryptionKeyArr = (jbyteArray)(*env)->GetObjectField(env, optionsObj, fid_searchableEncryptionKey);
    if (searchableEncryptionKeyArr != NULL) {
        opts.searchableEncryptionKey = (uint8_t*)(*env)->GetByteArrayElements(env, searchableEncryptionKeyArr, NULL);
        opts.searchableEncryptionKeyLen = (int)(*env)->GetArrayLength(env, searchableEncryptionKeyArr);
    }

    jfieldID fid_p2pBlockSyncTimeoutMs = (*env)->GetFieldID(env, cls, "p2pBlockSyncTimeoutMs", "J");
    opts.p2pBlockSyncTimeoutMs = (*env)->GetLongField(env, optionsObj, fid_p2pBlockSyncTimeoutMs);

    jfieldID fid_lensPoolSize = (*env)->GetFieldID(env, cls, "lensPoolSize", "I");
    opts.lensPoolSize = (*env)->GetIntField(env, optionsObj, fid_lensPoolSize);

    jfieldID fid_chunkSize = (*env)->GetFieldID(env, cls, "chunkSize", "I");
    opts.chunkSize = (*env)->GetIntField(env, optionsObj, fid_chunkSize);

    // P2P options
    jfieldID fid_enablePubSub = (*env)->GetFieldID(env, cls, "enablePubSub", "Z");
    opts.enablePubSub = (*env)->GetBooleanField(env, optionsObj, fid_enablePubSub) ? 1 : 0;

    jfieldID fid_enableRelay = (*env)->GetFieldID(env, cls, "enableRelay", "Z");
    opts.enableRelay = (*env)->GetBooleanField(env, optionsObj, fid_enableRelay) ? 1 : 0;

    jfieldID fid_enableClearBackoffOnRetry = (*env)->GetFieldID(env, cls, "enableClearBackoffOnRetry", "Z");
    opts.enableClearBackoffOnRetry = (*env)->GetBooleanField(env, optionsObj, fid_enableClearBackoffOnRetry) ? 1 : 0;

    jfieldID fid_p2pPrivateKey = (*env)->GetFieldID(env, cls, "p2pPrivateKey", "[B");
    jbyteArray p2pPrivateKeyArr = (jbyteArray)(*env)->GetObjectField(env, optionsObj, fid_p2pPrivateKey);
    if (p2pPrivateKeyArr != NULL) {
        opts.p2pPrivateKey = (uint8_t*)(*env)->GetByteArrayElements(env, p2pPrivateKeyArr, NULL);
        opts.p2pPrivateKeyLen = (int)(*env)->GetArrayLength(env, p2pPrivateKeyArr);
    }

    // HTTP options
    jfieldID fid_httpAddress = (*env)->GetFieldID(env, cls, "httpAddress", "Ljava/lang/String;");
    jstring httpAddressStr = (jstring)(*env)->GetObjectField(env, optionsObj, fid_httpAddress);
    opts.httpAddress = httpAddressStr ? (*env)->GetStringUTFChars(env, httpAddressStr, 0) : NULL;

    jfieldID fid_httpAllowedOrigins = (*env)->GetFieldID(env, cls, "httpAllowedOrigins", "Ljava/lang/String;");
    jstring httpAllowedOriginsStr = (jstring)(*env)->GetObjectField(env, optionsObj, fid_httpAllowedOrigins);
    opts.httpAllowedOrigins = httpAllowedOriginsStr ? (*env)->GetStringUTFChars(env, httpAllowedOriginsStr, 0) : NULL;

    jfieldID fid_tlsCertPath = (*env)->GetFieldID(env, cls, "tlsCertPath", "Ljava/lang/String;");
    jstring tlsCertPathStr = (jstring)(*env)->GetObjectField(env, optionsObj, fid_tlsCertPath);
    opts.tlsCertPath = tlsCertPathStr ? (*env)->GetStringUTFChars(env, tlsCertPathStr, 0) : NULL;

    jfieldID fid_tlsKeyPath = (*env)->GetFieldID(env, cls, "tlsKeyPath", "Ljava/lang/String;");
    jstring tlsKeyPathStr = (jstring)(*env)->GetObjectField(env, optionsObj, fid_tlsKeyPath);
    opts.tlsKeyPath = tlsKeyPathStr ? (*env)->GetStringUTFChars(env, tlsKeyPathStr, 0) : NULL;

    jfieldID fid_httpReadTimeoutMs = (*env)->GetFieldID(env, cls, "httpReadTimeoutMs", "J");
    opts.httpReadTimeoutMs = (*env)->GetLongField(env, optionsObj, fid_httpReadTimeoutMs);

    jfieldID fid_httpWriteTimeoutMs = (*env)->GetFieldID(env, cls, "httpWriteTimeoutMs", "J");
    opts.httpWriteTimeoutMs = (*env)->GetLongField(env, optionsObj, fid_httpWriteTimeoutMs);

    jfieldID fid_httpIdleTimeoutMs = (*env)->GetFieldID(env, cls, "httpIdleTimeoutMs", "J");
    opts.httpIdleTimeoutMs = (*env)->GetLongField(env, optionsObj, fid_httpIdleTimeoutMs);

    // Document ACP options
    jfieldID fid_documentACPType = (*env)->GetFieldID(env, cls, "documentACPType", "Ljava/lang/String;");
    jstring documentACPTypeStr = (jstring)(*env)->GetObjectField(env, optionsObj, fid_documentACPType);
    opts.documentACPType = documentACPTypeStr ? (*env)->GetStringUTFChars(env, documentACPTypeStr, 0) : NULL;

    jfieldID fid_documentACPPath = (*env)->GetFieldID(env, cls, "documentACPPath", "Ljava/lang/String;");
    jstring documentACPPathStr = (jstring)(*env)->GetObjectField(env, optionsObj, fid_documentACPPath);
    opts.documentACPPath = documentACPPathStr ? (*env)->GetStringUTFChars(env, documentACPPathStr, 0) : NULL;

    jfieldID fid_sourceHubChainID = (*env)->GetFieldID(env, cls, "sourceHubChainID", "Ljava/lang/String;");
    jstring sourceHubChainIDStr = (jstring)(*env)->GetObjectField(env, optionsObj, fid_sourceHubChainID);
    opts.sourceHubChainID = sourceHubChainIDStr ? (*env)->GetStringUTFChars(env, sourceHubChainIDStr, 0) : NULL;

    jfieldID fid_sourceHubGRPCAddress = (*env)->GetFieldID(env, cls, "sourceHubGRPCAddress", "Ljava/lang/String;");
    jstring sourceHubGRPCAddressStr = (jstring)(*env)->GetObjectField(env, optionsObj, fid_sourceHubGRPCAddress);
    opts.sourceHubGRPCAddress = sourceHubGRPCAddressStr ? (*env)->GetStringUTFChars(env, sourceHubGRPCAddressStr, 0) : NULL;

    jfieldID fid_sourceHubCometRPCAddress = (*env)->GetFieldID(env, cls, "sourceHubCometRPCAddress", "Ljava/lang/String;");
    jstring sourceHubCometRPCAddressStr = (jstring)(*env)->GetObjectField(env, optionsObj, fid_sourceHubCometRPCAddress);
    opts.sourceHubCometRPCAddress = sourceHubCometRPCAddressStr ? (*env)->GetStringUTFChars(env, sourceHubCometRPCAddressStr, 0) : NULL;

    // Node ACP options
    jfieldID fid_nodeACPPath = (*env)->GetFieldID(env, cls, "nodeACPPath", "Ljava/lang/String;");
    jstring nodeACPPathStr = (jstring)(*env)->GetObjectField(env, optionsObj, fid_nodeACPPath);
    opts.nodeACPPath = nodeACPPathStr ? (*env)->GetStringUTFChars(env, nodeACPPathStr, 0) : NULL;

    return opts;
}

// Helper to convert a Java DefraCollectionOptions object to a C CollectionOptions struct
CollectionOptions convertJavaCollectionOptions(JNIEnv* env, jobject optionsObj) {
    CollectionOptions opts;
	memset(&opts, 0, sizeof(CollectionOptions));
    jclass cls = (*env)->GetObjectClass(env, optionsObj);

    // Strings
    jfieldID fid_version = (*env)->GetFieldID(env, cls, "version", "Ljava/lang/String;");
    jfieldID fid_collectionID = (*env)->GetFieldID(env, cls, "collectionID", "Ljava/lang/String;");
    jfieldID fid_name = (*env)->GetFieldID(env, cls, "name", "Ljava/lang/String;");

    jstring versionStr = (jstring)(*env)->GetObjectField(env, optionsObj, fid_version);
    jstring collectionIDStr = (jstring)(*env)->GetObjectField(env, optionsObj, fid_collectionID);
    jstring nameStr = (jstring)(*env)->GetObjectField(env, optionsObj, fid_name);

    opts.version = versionStr ? (*env)->GetStringUTFChars(env, versionStr, 0) : NULL;
    opts.collectionID = collectionIDStr ? (*env)->GetStringUTFChars(env, collectionIDStr, 0) : NULL;
    opts.name = nameStr ? (*env)->GetStringUTFChars(env, nameStr, 0) : NULL;

    // Boolean
    jfieldID fid_getInactive = (*env)->GetFieldID(env, cls, "getInactive", "Z");
    opts.getInactive = (*env)->GetBooleanField(env, optionsObj, fid_getInactive) ? 1 : 0;

    // enableSigning is a boxed java.lang.Boolean (nullable), matching CollectionOptions'
    // own tri-state: null means unset (0), otherwise 1 (true) or -1 (false).
    jfieldID fid_enableSigning = (*env)->GetFieldID(env, cls, "enableSigning", "Ljava/lang/Boolean;");
    jobject enableSigningObj = (*env)->GetObjectField(env, optionsObj, fid_enableSigning);
    if (enableSigningObj != NULL) {
        jclass booleanCls = (*env)->GetObjectClass(env, enableSigningObj);
        jmethodID booleanValueMid = (*env)->GetMethodID(env, booleanCls, "booleanValue", "()Z");
        opts.enableSigning = (*env)->CallBooleanMethod(env, enableSigningObj, booleanValueMid) ? 1 : -1;
    } else {
        opts.enableSigning = 0;
    }

    return opts;
}

// Helper to release allocated Java strings after the call
void releaseJavaNodeInitOptions(JNIEnv* env, jobject optionsObj, NodeInitOptions opts) {
    jclass cls = (*env)->GetObjectClass(env, optionsObj);

    // Core strings
    jfieldID fid_dbPath = (*env)->GetFieldID(env, cls, "dbPath", "Ljava/lang/String;");
    jfieldID fid_listeningAddresses = (*env)->GetFieldID(env, cls, "listeningAddresses", "Ljava/lang/String;");
    jfieldID fid_replicatorRetryIntervals = (*env)->GetFieldID(env, cls, "replicatorRetryIntervals", "Ljava/lang/String;");
    jfieldID fid_peers = (*env)->GetFieldID(env, cls, "peers", "Ljava/lang/String;");

    if (opts.dbPath) {
        (*env)->ReleaseStringUTFChars(env, (jstring)(*env)->GetObjectField(env, optionsObj, fid_dbPath), opts.dbPath);
    }
    if (opts.listeningAddresses) {
        (*env)->ReleaseStringUTFChars(
            env, (jstring)(*env)->GetObjectField(env, optionsObj, fid_listeningAddresses), opts.listeningAddresses);
    }
    if (opts.replicatorRetryIntervals) {
        (*env)->ReleaseStringUTFChars(
            env,
            (jstring)(*env)->GetObjectField(env, optionsObj, fid_replicatorRetryIntervals),
            opts.replicatorRetryIntervals);
    }
    if (opts.peers) {
        (*env)->ReleaseStringUTFChars(env, (jstring)(*env)->GetObjectField(env, optionsObj, fid_peers), opts.peers);
    }

    // Store options
    jfieldID fid_storeType = (*env)->GetFieldID(env, cls, "storeType", "Ljava/lang/String;");
    if (opts.storeType) (*env)->ReleaseStringUTFChars(env, (jstring)(*env)->GetObjectField(env, optionsObj, fid_storeType), opts.storeType);

    jfieldID fid_badgerEncryptionKey = (*env)->GetFieldID(env, cls, "badgerEncryptionKey", "[B");
    if (opts.badgerEncryptionKey) (*env)->ReleaseByteArrayElements(env, (jbyteArray)(*env)->GetObjectField(env, optionsObj, fid_badgerEncryptionKey), (jbyte*)opts.badgerEncryptionKey, JNI_ABORT);

    // DB options
    jfieldID fid_searchableEncryptionKey = (*env)->GetFieldID(env, cls, "searchableEncryptionKey", "[B");
    if (opts.searchableEncryptionKey) (*env)->ReleaseByteArrayElements(env, (jbyteArray)(*env)->GetObjectField(env, optionsObj, fid_searchableEncryptionKey), (jbyte*)opts.searchableEncryptionKey, JNI_ABORT);

    // P2P options
    jfieldID fid_p2pPrivateKey = (*env)->GetFieldID(env, cls, "p2pPrivateKey", "[B");
    if (opts.p2pPrivateKey) (*env)->ReleaseByteArrayElements(env, (jbyteArray)(*env)->GetObjectField(env, optionsObj, fid_p2pPrivateKey), (jbyte*)opts.p2pPrivateKey, JNI_ABORT);

    // HTTP options
    jfieldID fid_httpAddress = (*env)->GetFieldID(env, cls, "httpAddress", "Ljava/lang/String;");
    if (opts.httpAddress) (*env)->ReleaseStringUTFChars(env, (jstring)(*env)->GetObjectField(env, optionsObj, fid_httpAddress), opts.httpAddress);

    jfieldID fid_httpAllowedOrigins = (*env)->GetFieldID(env, cls, "httpAllowedOrigins", "Ljava/lang/String;");
    if (opts.httpAllowedOrigins) (*env)->ReleaseStringUTFChars(env, (jstring)(*env)->GetObjectField(env, optionsObj, fid_httpAllowedOrigins), opts.httpAllowedOrigins);

    jfieldID fid_tlsCertPath = (*env)->GetFieldID(env, cls, "tlsCertPath", "Ljava/lang/String;");
    if (opts.tlsCertPath) (*env)->ReleaseStringUTFChars(env, (jstring)(*env)->GetObjectField(env, optionsObj, fid_tlsCertPath), opts.tlsCertPath);

    jfieldID fid_tlsKeyPath = (*env)->GetFieldID(env, cls, "tlsKeyPath", "Ljava/lang/String;");
    if (opts.tlsKeyPath) (*env)->ReleaseStringUTFChars(env, (jstring)(*env)->GetObjectField(env, optionsObj, fid_tlsKeyPath), opts.tlsKeyPath);

    // Document ACP options
    jfieldID fid_documentACPType = (*env)->GetFieldID(env, cls, "documentACPType", "Ljava/lang/String;");
    if (opts.documentACPType) (*env)->ReleaseStringUTFChars(env, (jstring)(*env)->GetObjectField(env, optionsObj, fid_documentACPType), opts.documentACPType);

    jfieldID fid_documentACPPath = (*env)->GetFieldID(env, cls, "documentACPPath", "Ljava/lang/String;");
    if (opts.documentACPPath) (*env)->ReleaseStringUTFChars(env, (jstring)(*env)->GetObjectField(env, optionsObj, fid_documentACPPath), opts.documentACPPath);

    jfieldID fid_sourceHubChainID = (*env)->GetFieldID(env, cls, "sourceHubChainID", "Ljava/lang/String;");
    if (opts.sourceHubChainID) (*env)->ReleaseStringUTFChars(env, (jstring)(*env)->GetObjectField(env, optionsObj, fid_sourceHubChainID), opts.sourceHubChainID);

    jfieldID fid_sourceHubGRPCAddress = (*env)->GetFieldID(env, cls, "sourceHubGRPCAddress", "Ljava/lang/String;");
    if (opts.sourceHubGRPCAddress) (*env)->ReleaseStringUTFChars(env, (jstring)(*env)->GetObjectField(env, optionsObj, fid_sourceHubGRPCAddress), opts.sourceHubGRPCAddress);

    jfieldID fid_sourceHubCometRPCAddress = (*env)->GetFieldID(env, cls, "sourceHubCometRPCAddress", "Ljava/lang/String;");
    if (opts.sourceHubCometRPCAddress) (*env)->ReleaseStringUTFChars(env, (jstring)(*env)->GetObjectField(env, optionsObj, fid_sourceHubCometRPCAddress), opts.sourceHubCometRPCAddress);

    // Node ACP options
    jfieldID fid_nodeACPPath = (*env)->GetFieldID(env, cls, "nodeACPPath", "Ljava/lang/String;");
    if (opts.nodeACPPath) (*env)->ReleaseStringUTFChars(env, (jstring)(*env)->GetObjectField(env, optionsObj, fid_nodeACPPath), opts.nodeACPPath);
}

void releaseJavaCollectionOptions(JNIEnv* env, jobject optionsObj, CollectionOptions opts) {
    jclass cls = (*env)->GetObjectClass(env, optionsObj);
    jfieldID fid_version = (*env)->GetFieldID(env, cls, "version", "Ljava/lang/String;");
    jfieldID fid_collectionID = (*env)->GetFieldID(env, cls, "collectionID", "Ljava/lang/String;");
    jfieldID fid_name = (*env)->GetFieldID(env, cls, "name", "Ljava/lang/String;");

    jstring versionStr = (jstring)(*env)->GetObjectField(env, optionsObj, fid_version);
    jstring collectionIDStr = (jstring)(*env)->GetObjectField(env, optionsObj, fid_collectionID);
    jstring nameStr = (jstring)(*env)->GetObjectField(env, optionsObj, fid_name);

    if (opts.version) (*env)->ReleaseStringUTFChars(env, versionStr, opts.version);
    if (opts.collectionID) (*env)->ReleaseStringUTFChars(env, collectionIDStr, opts.collectionID);
    if (opts.name) (*env)->ReleaseStringUTFChars(env, nameStr, opts.name);
}

//=============================================================================
// DefraNode JNI Functions
//=============================================================================

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_NewNodeNative
(JNIEnv *env, jobject thisObj, jobject optionsObj) {
    NodeInitOptions opts = convertJavaNodeInitOptions(env, optionsObj);
    NewNodeResult res = NewNode(opts);
    releaseJavaNodeInitOptions(env, optionsObj, opts);
    return returnDefraNewNodeResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_NodeCloseNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr
) {
    Result res = CloseNode((uintptr_t)nodePtr);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_ACPAddDACPolicyNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jlong identityPtr,
    jstring policyStr
) {
    const char* policyC = policyStr ? (*env)->GetStringUTFChars(env, policyStr, NULL) : NULL;
    Result res = ACPAddDACPolicy((uintptr_t)nodePtr, (uintptr_t)identityPtr, (char*)policyC);
    if (policyStr) (*env)->ReleaseStringUTFChars(env, policyStr, policyC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_ACPAddDACActorRelationshipNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jlong identityPtr,
    jstring collectionStr,
    jstring docIDStr,
    jstring relationStr,
    jstring actorStr
) {
    const char* collectionC = collectionStr ? (*env)->GetStringUTFChars(env, collectionStr, NULL) : NULL;
    const char* docIDC = docIDStr ? (*env)->GetStringUTFChars(env, docIDStr, NULL) : NULL;
    const char* relationC = relationStr ? (*env)->GetStringUTFChars(env, relationStr, NULL) : NULL;
    const char* actorC = actorStr ? (*env)->GetStringUTFChars(env, actorStr, NULL) : NULL;
    Result res = ACPAddDACActorRelationship((uintptr_t)nodePtr, (uintptr_t)identityPtr, (char*)collectionC, (char*)docIDC, (char*)relationC, (char*)actorC);
    if (collectionStr) (*env)->ReleaseStringUTFChars(env, collectionStr, collectionC);
    if (docIDStr) (*env)->ReleaseStringUTFChars(env, docIDStr, docIDC);
    if (relationStr) (*env)->ReleaseStringUTFChars(env, relationStr, relationC);
    if (actorStr) (*env)->ReleaseStringUTFChars(env, actorStr, actorC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_ACPDeleteDACActorRelationshipNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jlong identityPtr,
    jstring collectionStr,
    jstring docIDStr,
    jstring relationStr,
    jstring actorStr
) {
    const char* collectionC = collectionStr ? (*env)->GetStringUTFChars(env, collectionStr, NULL) : NULL;
    const char* docIDC = docIDStr ? (*env)->GetStringUTFChars(env, docIDStr, NULL) : NULL;
    const char* relationC = relationStr ? (*env)->GetStringUTFChars(env, relationStr, NULL) : NULL;
    const char* actorC = actorStr ? (*env)->GetStringUTFChars(env, actorStr, NULL) : NULL;
    Result res = ACPDeleteDACActorRelationship((uintptr_t)nodePtr, (uintptr_t)identityPtr, (char*)collectionC, (char*)docIDC, (char*)relationC, (char*)actorC);
    if (collectionStr) (*env)->ReleaseStringUTFChars(env, collectionStr, collectionC);
    if (docIDStr) (*env)->ReleaseStringUTFChars(env, docIDStr, docIDC);
    if (relationStr) (*env)->ReleaseStringUTFChars(env, relationStr, relationC);
    if (actorStr) (*env)->ReleaseStringUTFChars(env, actorStr, actorC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_ACPDisableNACNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jlong identityPtr
) {
    Result res = ACPDisableNAC((uintptr_t)nodePtr, (uintptr_t)identityPtr);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_ACPReEnableNACNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jlong identityPtr
) {
    Result res = ACPReEnableNAC((uintptr_t)nodePtr, (uintptr_t)identityPtr);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_ACPAddNACActorRelationshipNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jlong identityPtr,
    jstring relationStr,
    jstring actorStr
) {
    const char* relationC = relationStr ? (*env)->GetStringUTFChars(env, relationStr, NULL) : NULL;
    const char* actorC = actorStr ? (*env)->GetStringUTFChars(env, actorStr, NULL) : NULL;
    Result res = ACPAddNACActorRelationship((uintptr_t)nodePtr, (uintptr_t)identityPtr, (char*)relationC, (char*)actorC);
    if (relationStr) (*env)->ReleaseStringUTFChars(env, relationStr, relationC);
    if (actorStr) (*env)->ReleaseStringUTFChars(env, actorStr, actorC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_ACPDeleteNACActorRelationshipNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jlong identityPtr,
    jstring relationStr,
    jstring actorStr
) {
    const char* relationC = relationStr ? (*env)->GetStringUTFChars(env, relationStr, NULL) : NULL;
    const char* actorC = actorStr ? (*env)->GetStringUTFChars(env, actorStr, NULL) : NULL;
    Result res = ACPDeleteNACActorRelationship((uintptr_t)nodePtr, (uintptr_t)identityPtr, (char*)relationC, (char*)actorC);
    if (relationStr) (*env)->ReleaseStringUTFChars(env, relationStr, relationC);
    if (actorStr) (*env)->ReleaseStringUTFChars(env, actorStr, actorC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_ACPGetNACStatusNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jlong identityPtr
) {
    Result res = ACPGetNACStatus((uintptr_t)nodePtr, (uintptr_t)identityPtr);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_AddCollectionNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring sdlStr,
    jlong identityPtr
) {
    const char* sdlC = sdlStr ? (*env)->GetStringUTFChars(env, sdlStr, NULL) : NULL;
    Result res = AddCollection((uintptr_t)nodePtr, (char*)sdlC, (uintptr_t)identityPtr);
    if (sdlStr) (*env)->ReleaseStringUTFChars(env, sdlStr, sdlC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_DescribeCollectionNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jobject optionsObj,
    jlong identityPtr
) {
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj);
    Result res = DescribeCollection((uintptr_t)nodePtr, opts, (uintptr_t)identityPtr);
    releaseJavaCollectionOptions(env, optionsObj, opts);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_PatchCollectionNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring patchStr,
    jstring lensConfigStr,
    jlong identityPtr
) {
    const char* patchC = patchStr ? (*env)->GetStringUTFChars(env, patchStr, NULL) : NULL;
    const char* lensConfigC = lensConfigStr ? (*env)->GetStringUTFChars(env, lensConfigStr, NULL) : NULL;
    Result res = PatchCollection((uintptr_t)nodePtr, (char*)patchC, (char*)lensConfigC, (uintptr_t)identityPtr);
    if (patchStr) (*env)->ReleaseStringUTFChars(env, patchStr, patchC);
    if (lensConfigStr) (*env)->ReleaseStringUTFChars(env, lensConfigStr, lensConfigC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_SetActiveCollectionNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jobject optionsObj,
    jlong identityPtr
) {
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj);
    Result res = SetActiveCollection((uintptr_t)nodePtr, opts, (uintptr_t)identityPtr);
    releaseJavaCollectionOptions(env, optionsObj, opts);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_TruncateCollectionNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jobject optionsObj,
    jlong identityPtr
) {
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj);
    Result res = TruncateCollection((uintptr_t)nodePtr, opts, (uintptr_t)identityPtr);
    releaseJavaCollectionOptions(env, optionsObj, opts);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_AddDocumentNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring jsonStr,
    jboolean isEncrypted,
    jstring encryptedFieldsStr,
    jobject optionsObj,
    jlong identityPtr
) {
    const char* jsonC = jsonStr ? (*env)->GetStringUTFChars(env, jsonStr, NULL) : NULL;
    const char* encryptedFieldsC = encryptedFieldsStr ? (*env)->GetStringUTFChars(env, encryptedFieldsStr, NULL) : NULL;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj);
    int isEncryptedC = (isEncrypted == JNI_TRUE) ? 1 : 0;
    Result res = AddDocument((uintptr_t)nodePtr, (char*)jsonC, isEncryptedC, (char*)encryptedFieldsC, opts, (uintptr_t)identityPtr);
    if (jsonStr) (*env)->ReleaseStringUTFChars(env, jsonStr, jsonC);
    if (encryptedFieldsStr) (*env)->ReleaseStringUTFChars(env, encryptedFieldsStr, encryptedFieldsC);
    releaseJavaCollectionOptions(env, optionsObj, opts);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_DeleteDocumentNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring docIDStr,
    jstring filterStr,
    jobject optionsObj,
    jlong identityPtr
) {
    const char* docIDC = docIDStr ? (*env)->GetStringUTFChars(env, docIDStr, NULL) : NULL;
    const char* filterC = filterStr ? (*env)->GetStringUTFChars(env, filterStr, NULL) : NULL;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj);
    Result res = DeleteDocument((uintptr_t)nodePtr, (char*)docIDC, (char*)filterC, opts, (uintptr_t)identityPtr);
    if (docIDStr) (*env)->ReleaseStringUTFChars(env, docIDStr, docIDC);
    if (filterStr) (*env)->ReleaseStringUTFChars(env, filterStr, filterC);
    releaseJavaCollectionOptions(env, optionsObj, opts);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_GetDocumentNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring docIDStr,
    jboolean showDeleted,
    jobject optionsObj,
    jlong identityPtr
) {
    const char* docIDC = docIDStr ? (*env)->GetStringUTFChars(env, docIDStr, NULL) : NULL;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj);
    int showDeletedC = (showDeleted == JNI_TRUE) ? 1 : 0;
    Result res = GetDocument((uintptr_t)nodePtr, (char*)docIDC, showDeletedC, opts, (uintptr_t)identityPtr);
    if (docIDStr) (*env)->ReleaseStringUTFChars(env, docIDStr, docIDC);
    releaseJavaCollectionOptions(env, optionsObj, opts);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_UpdateDocumentNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring docIDStr,
    jstring filterStr,
    jstring updaterStr,
    jobject optionsObj,
    jlong identityPtr
) {
    const char* docIDC = docIDStr ? (*env)->GetStringUTFChars(env, docIDStr, NULL) : NULL;
    const char* filterC = filterStr ? (*env)->GetStringUTFChars(env, filterStr, NULL) : NULL;
    const char* updaterC = updaterStr ? (*env)->GetStringUTFChars(env, updaterStr, NULL) : NULL;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj);
    Result res = UpdateDocument((uintptr_t)nodePtr, (char*)docIDC, (char*)filterC, (char*)updaterC, opts, (uintptr_t)identityPtr);
    if (docIDStr) (*env)->ReleaseStringUTFChars(env, docIDStr, docIDC);
    if (filterStr) (*env)->ReleaseStringUTFChars(env, filterStr, filterC);
    if (updaterStr) (*env)->ReleaseStringUTFChars(env, updaterStr, updaterC);
    releaseJavaCollectionOptions(env, optionsObj, opts);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_NewEncryptedIndexNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring collectionNameStr,
    jstring fieldNameStr,
    jlong identityPtr
) {
    const char* collectionNameC = collectionNameStr ? (*env)->GetStringUTFChars(env, collectionNameStr, NULL) : NULL;
    const char* fieldNameC = fieldNameStr ? (*env)->GetStringUTFChars(env, fieldNameStr, NULL) : NULL;
    Result res = NewEncryptedIndex((uintptr_t)nodePtr, (char*)collectionNameC, (char*)fieldNameC, (uintptr_t)identityPtr);
    if (collectionNameStr) (*env)->ReleaseStringUTFChars(env, collectionNameStr, collectionNameC);
    if (fieldNameStr) (*env)->ReleaseStringUTFChars(env, fieldNameStr, fieldNameC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_ListEncryptedIndexesNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring collectionNameStr,
    jlong identityPtr
) {
    const char* collectionNameC = collectionNameStr ? (*env)->GetStringUTFChars(env, collectionNameStr, NULL) : NULL;
    Result res = ListEncryptedIndexes((uintptr_t)nodePtr, (char*)collectionNameC, (uintptr_t)identityPtr);
    if (collectionNameStr) (*env)->ReleaseStringUTFChars(env, collectionNameStr, collectionNameC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_DeleteEncryptedIndexNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring collectionNameStr,
    jstring fieldNameStr,
    jlong identityPtr
) {
    const char* collectionNameC = collectionNameStr ? (*env)->GetStringUTFChars(env, collectionNameStr, NULL) : NULL;
    const char* fieldNameC = fieldNameStr ? (*env)->GetStringUTFChars(env, fieldNameStr, NULL) : NULL;
    Result res = DeleteEncryptedIndex((uintptr_t)nodePtr, (char*)collectionNameC, (char*)fieldNameC, (uintptr_t)identityPtr);
    if (collectionNameStr) (*env)->ReleaseStringUTFChars(env, collectionNameStr, collectionNameC);
    if (fieldNameStr) (*env)->ReleaseStringUTFChars(env, fieldNameStr, fieldNameC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_NewIndexNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring indexNameStr,
    jstring fieldsStr,
    jboolean isUnique,
    jobject optionsObj,
    jlong identityPtr
) {
    const char* indexNameC = indexNameStr ? (*env)->GetStringUTFChars(env, indexNameStr, NULL) : NULL;
    const char* fieldsC = fieldsStr ? (*env)->GetStringUTFChars(env, fieldsStr, NULL) : NULL;
    int isUniqueC = (isUnique == JNI_TRUE) ? 1 : 0;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj);
    Result res = NewIndex((uintptr_t)nodePtr, (char*)indexNameC, (char*)fieldsC, isUniqueC, opts, (uintptr_t)identityPtr);
    if (indexNameStr) (*env)->ReleaseStringUTFChars(env, indexNameStr, indexNameC);
    if (fieldsStr) (*env)->ReleaseStringUTFChars(env, fieldsStr, fieldsC);
    releaseJavaCollectionOptions(env, optionsObj, opts);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_ListIndexesNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jobject optionsObj,
    jlong identityPtr
) {
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj);
    Result res = ListIndexes((uintptr_t)nodePtr, opts, (uintptr_t)identityPtr);
    releaseJavaCollectionOptions(env, optionsObj, opts);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_DeleteIndexNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring indexNameStr,
    jobject optionsObj,
    jlong identityPtr
) {
    const char* indexNameC = indexNameStr ? (*env)->GetStringUTFChars(env, indexNameStr, NULL) : NULL;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj);
    Result res = DeleteIndex((uintptr_t)nodePtr, (char*)indexNameC, opts, (uintptr_t)identityPtr);
    if (indexNameStr) (*env)->ReleaseStringUTFChars(env, indexNameStr, indexNameC);
    releaseJavaCollectionOptions(env, optionsObj, opts);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_IdentityNewNative(
    JNIEnv* env,
    jobject thiz,
    jstring keyTypeStr
) {
    const char* keyTypeC = keyTypeStr ? (*env)->GetStringUTFChars(env, keyTypeStr, NULL) : NULL;
    NewIdentityResult res = NewIdentity((char*)keyTypeC);
    if (keyTypeStr) (*env)->ReleaseStringUTFChars(env, keyTypeStr, keyTypeC);
    return returnDefraIdentityResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_GetNodeIdentityNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr
) {
    Result res = GetNodeIdentity((uintptr_t)nodePtr);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_ListActionsNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jlong identityPtr
) {
    Result res = ListActions((uintptr_t)nodePtr, (uintptr_t)identityPtr);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_DeleteCollectionNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring namesStr,
    jint activeOnly,
    jlong identityPtr
) {
    const char* namesC = namesStr ? (*env)->GetStringUTFChars(env, namesStr, NULL) : NULL;
    Result res = DeleteCollection((uintptr_t)nodePtr, (char*)namesC, (int)activeOnly, (uintptr_t)identityPtr);
    if (namesStr) (*env)->ReleaseStringUTFChars(env, namesStr, namesC);
    return returnDefraResult(env, res);
}

JNIEXPORT void JNICALL Java_source_defra_DefraNode_FreeIdentityNative(
    JNIEnv* env,
    jobject thiz,
    jlong identityPtr
) {
    FreeIdentity((uintptr_t)identityPtr);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_SetLensNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jlong identityPtr,
    jstring srcStr,
    jstring dstStr,
    jstring cfgStr
) {
    const char* srcC = srcStr ? (*env)->GetStringUTFChars(env, srcStr, NULL) : NULL;
    const char* dstC = dstStr ? (*env)->GetStringUTFChars(env, dstStr, NULL) : NULL;
    const char* cfgC = cfgStr ? (*env)->GetStringUTFChars(env, cfgStr, NULL) : NULL;
    Result res = SetLens((uintptr_t)nodePtr, (uintptr_t)identityPtr, (char*)srcC, (char*)dstC, (char*)cfgC);
    if (srcStr) (*env)->ReleaseStringUTFChars(env, srcStr, srcC);
    if (dstStr) (*env)->ReleaseStringUTFChars(env, dstStr, dstC);
    if (cfgStr) (*env)->ReleaseStringUTFChars(env, cfgStr, cfgC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_AddLensNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jlong identityPtr,
    jstring cfgStr
) {
    const char* cfgC = cfgStr ? (*env)->GetStringUTFChars(env, cfgStr, NULL) : NULL;
    Result res = AddLens((uintptr_t)nodePtr, (uintptr_t)identityPtr, (char*)cfgC);
    if (cfgStr) (*env)->ReleaseStringUTFChars(env, cfgStr, cfgC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_ListLensesNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jlong identityPtr
) {
    Result res = ListLenses((uintptr_t)nodePtr, (uintptr_t)identityPtr);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_VerifyBlockSignatureNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring keyTypeStr,
    jstring publicKeyStr,
    jstring cidStr,
    jlong identityPtr
) {
    const char* keyTypeC = keyTypeStr ? (*env)->GetStringUTFChars(env, keyTypeStr, NULL) : NULL;
    const char* publicKeyC = publicKeyStr ? (*env)->GetStringUTFChars(env, publicKeyStr, NULL) : NULL;
    const char* cidC = cidStr ? (*env)->GetStringUTFChars(env, cidStr, NULL) : NULL;
    Result res = VerifyBlockSignature((uintptr_t)nodePtr, (char*)keyTypeC, (char*)publicKeyC, (char*)cidC, (uintptr_t)identityPtr);
    if (keyTypeStr) (*env)->ReleaseStringUTFChars(env, keyTypeStr, keyTypeC);
    if (publicKeyStr) (*env)->ReleaseStringUTFChars(env, publicKeyStr, publicKeyC);
    if (cidStr) (*env)->ReleaseStringUTFChars(env, cidStr, cidC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_GetP2PInfoNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jlong identityPtr
) {
    Result res = GetP2PInfo((uintptr_t)nodePtr, (uintptr_t)identityPtr);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_ListP2PActivePeersNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jlong identityPtr
) {
    Result res = ListP2PActivePeers((uintptr_t)nodePtr, (uintptr_t)identityPtr);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_ListP2PReplicatorsNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jlong identityPtr
) {
    Result res = ListP2PReplicators((uintptr_t)nodePtr, (uintptr_t)identityPtr);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_AddP2PReplicatorNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring collectionsStr,
    jstring addressesStr,
    jlong identityPtr
) {
    const char* collectionsC = collectionsStr ? (*env)->GetStringUTFChars(env, collectionsStr, NULL) : NULL;
    const char* addressesC = addressesStr ? (*env)->GetStringUTFChars(env, addressesStr, NULL) : NULL;
    Result res = AddP2PReplicator((uintptr_t)nodePtr, (char*)collectionsC, (char*)addressesC, (uintptr_t)identityPtr);
    if (collectionsStr) (*env)->ReleaseStringUTFChars(env, collectionsStr, collectionsC);
    if (addressesStr) (*env)->ReleaseStringUTFChars(env, addressesStr, addressesC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_DeleteP2PReplicatorNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring collectionsStr,
    jstring idStr,
    jlong identityPtr
) {
    const char* collectionsC = collectionsStr ? (*env)->GetStringUTFChars(env, collectionsStr, NULL) : NULL;
    const char* idC = idStr ? (*env)->GetStringUTFChars(env, idStr, NULL) : NULL;
    Result res = DeleteP2PReplicator((uintptr_t)nodePtr, (char*)collectionsC, (char*)idC, (uintptr_t)identityPtr);
    if (collectionsStr) (*env)->ReleaseStringUTFChars(env, collectionsStr, collectionsC);
    if (idStr) (*env)->ReleaseStringUTFChars(env, idStr, idC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_AddP2PCollectionNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring collectionsStr,
    jlong identityPtr
) {
    const char* collectionsC = collectionsStr ? (*env)->GetStringUTFChars(env, collectionsStr, NULL) : NULL;
    Result res = AddP2PCollection((uintptr_t)nodePtr, (char*)collectionsC, (uintptr_t)identityPtr);
    if (collectionsStr) (*env)->ReleaseStringUTFChars(env, collectionsStr, collectionsC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_DeleteP2PCollectionNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring collectionsStr,
    jlong identityPtr
) {
    const char* collectionsC = collectionsStr ? (*env)->GetStringUTFChars(env, collectionsStr, NULL) : NULL;
    Result res = DeleteP2PCollection((uintptr_t)nodePtr, (char*)collectionsC, (uintptr_t)identityPtr);
    if (collectionsStr) (*env)->ReleaseStringUTFChars(env, collectionsStr, collectionsC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_ListP2PCollectionsNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jlong identityPtr
) {
    Result res = ListP2PCollections((uintptr_t)nodePtr, (uintptr_t)identityPtr);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_AddP2PDocumentNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring collectionsStr,
    jlong identityPtr
) {
    const char* collectionsC = collectionsStr ? (*env)->GetStringUTFChars(env, collectionsStr, NULL) : NULL;
    Result res = AddP2PDocument((uintptr_t)nodePtr, (char*)collectionsC, (uintptr_t)identityPtr);
    if (collectionsStr) (*env)->ReleaseStringUTFChars(env, collectionsStr, collectionsC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_DeleteP2PDocumentNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring collectionsStr,
    jlong identityPtr
) {
    const char* collectionsC = collectionsStr ? (*env)->GetStringUTFChars(env, collectionsStr, NULL) : NULL;
    Result res = DeleteP2PDocument((uintptr_t)nodePtr, (char*)collectionsC, (uintptr_t)identityPtr);
    if (collectionsStr) (*env)->ReleaseStringUTFChars(env, collectionsStr, collectionsC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_ListP2PDocumentsNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jlong identityPtr
) {
    Result res = ListP2PDocuments((uintptr_t)nodePtr, (uintptr_t)identityPtr);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_SyncP2PDocumentsNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring collectionStr,
    jstring docIDsStr,
    jstring timeoutStr,
    jlong identityPtr
) {
    const char* collectionC = collectionStr ? (*env)->GetStringUTFChars(env, collectionStr, NULL) : NULL;
    const char* docIDsC = docIDsStr ? (*env)->GetStringUTFChars(env, docIDsStr, NULL) : NULL;
    const char* timeoutC = timeoutStr ? (*env)->GetStringUTFChars(env, timeoutStr, NULL) : NULL;
    Result res = SyncP2PDocuments((uintptr_t)nodePtr, (char*)collectionC, (char*)docIDsC, (char*)timeoutC, (uintptr_t)identityPtr);
    if (collectionStr) (*env)->ReleaseStringUTFChars(env, collectionStr, collectionC);
    if (docIDsStr) (*env)->ReleaseStringUTFChars(env, docIDsStr, docIDsC);
    if (timeoutStr) (*env)->ReleaseStringUTFChars(env, timeoutStr, timeoutC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_SyncP2PCollectionVersionsNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring versionIDsStr,
    jstring timeoutStr,
    jlong identityPtr
) {
    const char* versionIDsC = versionIDsStr ? (*env)->GetStringUTFChars(env, versionIDsStr, NULL) : NULL;
    const char* timeoutC = timeoutStr ? (*env)->GetStringUTFChars(env, timeoutStr, NULL) : NULL;
    Result res = SyncP2PCollectionVersions((uintptr_t)nodePtr, (char*)versionIDsC, (char*)timeoutC, (uintptr_t)identityPtr);
    if (versionIDsStr) (*env)->ReleaseStringUTFChars(env, versionIDsStr, versionIDsC);
    if (timeoutStr) (*env)->ReleaseStringUTFChars(env, timeoutStr, timeoutC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_SyncP2PBranchableCollectionNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring collectionIDStr,
    jstring timeoutStr,
    jlong identityPtr
) {
    const char* collectionIDC = collectionIDStr ? (*env)->GetStringUTFChars(env, collectionIDStr, NULL) : NULL;
    const char* timeoutC = timeoutStr ? (*env)->GetStringUTFChars(env, timeoutStr, NULL) : NULL;
    Result res = SyncP2PBranchableCollection((uintptr_t)nodePtr, (char*)collectionIDC, (char*)timeoutC, (uintptr_t)identityPtr);
    if (collectionIDStr) (*env)->ReleaseStringUTFChars(env, collectionIDStr, collectionIDC);
    if (timeoutStr) (*env)->ReleaseStringUTFChars(env, timeoutStr, timeoutC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_ConnectP2PPeersNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring peerAddressesStr,
    jlong identityPtr
) {
    const char* peerAddressesC = peerAddressesStr ? (*env)->GetStringUTFChars(env, peerAddressesStr, NULL) : NULL;
    Result res = ConnectP2PPeers((uintptr_t)nodePtr, (char*)peerAddressesC, (uintptr_t)identityPtr);
    if (peerAddressesStr) (*env)->ReleaseStringUTFChars(env, peerAddressesStr, peerAddressesC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_DisconnectP2PPeersNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring peerAddressesStr,
    jlong identityPtr
) {
    const char* peerAddressesC = peerAddressesStr ? (*env)->GetStringUTFChars(env, peerAddressesStr, NULL) : NULL;
    Result res = DisconnectP2PPeers((uintptr_t)nodePtr, (char*)peerAddressesC, (uintptr_t)identityPtr);
    if (peerAddressesStr) (*env)->ReleaseStringUTFChars(env, peerAddressesStr, peerAddressesC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_ExecuteQueryNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring queryStr,
    jlong identityPtr,
    jstring operationNameStr,
    jstring variablesStr
) {
    const char* queryC = queryStr ? (*env)->GetStringUTFChars(env, queryStr, NULL) : NULL;
    const char* operationNameC = operationNameStr ? (*env)->GetStringUTFChars(env, operationNameStr, NULL) : NULL;
    const char* variablesC = variablesStr ? (*env)->GetStringUTFChars(env, variablesStr, NULL) : NULL;
    Result res = ExecuteQuery((uintptr_t)nodePtr, (char*)queryC, (uintptr_t)identityPtr, (char*)operationNameC, (char*)variablesC);
    if (queryStr) (*env)->ReleaseStringUTFChars(env, queryStr, queryC);
    if (operationNameStr) (*env)->ReleaseStringUTFChars(env, operationNameStr, operationNameC);
    if (variablesStr) (*env)->ReleaseStringUTFChars(env, variablesStr, variablesC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_PollSubscriptionNative(
    JNIEnv* env,
    jobject thiz,
    jstring idStr
) {
    const char* idC = idStr ? (*env)->GetStringUTFChars(env, idStr, NULL) : NULL;
    Result res = PollSubscription((char*)idC);
    if (idStr) (*env)->ReleaseStringUTFChars(env, idStr, idC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_CloseSubscriptionNative(
    JNIEnv* env,
    jobject thiz,
    jstring idStr
) {
    const char* idC = idStr ? (*env)->GetStringUTFChars(env, idStr, NULL) : NULL;
    Result res = CloseSubscription((char*)idC);
    if (idStr) (*env)->ReleaseStringUTFChars(env, idStr, idC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_GetVersionNative(
    JNIEnv* env,
    jobject thiz,
    jboolean flagFull,
    jboolean flagJSON
) {
    int flagFullC = (flagFull == JNI_TRUE) ? 1 : 0;
    int flagJSONC = (flagJSON == JNI_TRUE) ? 1 : 0;
    Result res = GetVersion(flagFullC, flagJSONC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_AddViewNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring queryStr,
    jstring sdlStr,
    jstring transformCIDStr,
    jlong identityPtr
) {
    const char* queryC = queryStr ? (*env)->GetStringUTFChars(env, queryStr, NULL) : NULL;
    const char* sdlC = sdlStr ? (*env)->GetStringUTFChars(env, sdlStr, NULL) : NULL;
    const char* transformCIDC = transformCIDStr ? (*env)->GetStringUTFChars(env, transformCIDStr, NULL) : NULL;
    Result res = AddView((uintptr_t)nodePtr, (char*)queryC, (char*)sdlC, (char*)transformCIDC, (uintptr_t)identityPtr);
    if (queryStr) (*env)->ReleaseStringUTFChars(env, queryStr, queryC);
    if (sdlStr) (*env)->ReleaseStringUTFChars(env, sdlStr, sdlC);
    if (transformCIDStr) (*env)->ReleaseStringUTFChars(env, transformCIDStr, transformCIDC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_RefreshViewNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jobject optionsObj,
    jlong identityPtr
) {
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj);
    Result res = RefreshView((uintptr_t)nodePtr, opts, (uintptr_t)identityPtr);
    releaseJavaCollectionOptions(env, optionsObj, opts);
    return returnDefraResult(env, res);
}

//=============================================================================
// DefraNode transaction-create JNI function
//=============================================================================
// See the file header comment: this is bound under DefraNode (not
// DefraTransaction) with the 2-arg signature DefraNode.java actually
// declares.
JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_TransactionCreateNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jboolean isReadOnly
) {
    int isReadOnlyC = (isReadOnly == JNI_TRUE) ? 1 : 0;
    NewTxnResult res = CreateTransaction((uintptr_t)nodePtr, isReadOnlyC);
    return returnDefraTransactionResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraTransaction_TransactionCommitNative(
    JNIEnv* env,
    jobject thiz,
    jlong txnPtr
) {
    Result res = CommitTransaction((uintptr_t)txnPtr);
    return returnDefraResult(env, res);
}

JNIEXPORT void JNICALL Java_source_defra_DefraTransaction_TransactionDiscardNative(
    JNIEnv* env,
    jobject thiz,
    jlong txnPtr
) {
    DiscardTransaction((uintptr_t)txnPtr);
}

// Transaction ACP Methods
JNIEXPORT jobject JNICALL Java_source_defra_DefraTransaction_ACPAddDACPolicyNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jlong identityPtr,
    jstring policyStr
) {
    const char* policyC = policyStr ? (*env)->GetStringUTFChars(env, policyStr, NULL) : NULL;
    Result res = ACPAddDACPolicy((uintptr_t)nodePtr, (uintptr_t)identityPtr, (char*)policyC);
    if (policyStr) (*env)->ReleaseStringUTFChars(env, policyStr, policyC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraTransaction_ACPAddDACActorRelationshipNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jlong identityPtr,
    jstring collectionStr,
    jstring docIDStr,
    jstring relationStr,
    jstring actorStr
) {
    const char* collectionC = collectionStr ? (*env)->GetStringUTFChars(env, collectionStr, NULL) : NULL;
    const char* docIDC = docIDStr ? (*env)->GetStringUTFChars(env, docIDStr, NULL) : NULL;
    const char* relationC = relationStr ? (*env)->GetStringUTFChars(env, relationStr, NULL) : NULL;
    const char* actorC = actorStr ? (*env)->GetStringUTFChars(env, actorStr, NULL) : NULL;
    Result res = ACPAddDACActorRelationship((uintptr_t)nodePtr, (uintptr_t)identityPtr, (char*)collectionC, (char*)docIDC, (char*)relationC, (char*)actorC);
    if (collectionStr) (*env)->ReleaseStringUTFChars(env, collectionStr, collectionC);
    if (docIDStr) (*env)->ReleaseStringUTFChars(env, docIDStr, docIDC);
    if (relationStr) (*env)->ReleaseStringUTFChars(env, relationStr, relationC);
    if (actorStr) (*env)->ReleaseStringUTFChars(env, actorStr, actorC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraTransaction_ACPDeleteDACActorRelationshipNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jlong identityPtr,
    jstring collectionStr,
    jstring docIDStr,
    jstring relationStr,
    jstring actorStr
) {
    const char* collectionC = collectionStr ? (*env)->GetStringUTFChars(env, collectionStr, NULL) : NULL;
    const char* docIDC = docIDStr ? (*env)->GetStringUTFChars(env, docIDStr, NULL) : NULL;
    const char* relationC = relationStr ? (*env)->GetStringUTFChars(env, relationStr, NULL) : NULL;
    const char* actorC = actorStr ? (*env)->GetStringUTFChars(env, actorStr, NULL) : NULL;
    Result res = ACPDeleteDACActorRelationship((uintptr_t)nodePtr, (uintptr_t)identityPtr, (char*)collectionC, (char*)docIDC, (char*)relationC, (char*)actorC);
    if (collectionStr) (*env)->ReleaseStringUTFChars(env, collectionStr, collectionC);
    if (docIDStr) (*env)->ReleaseStringUTFChars(env, docIDStr, docIDC);
    if (relationStr) (*env)->ReleaseStringUTFChars(env, relationStr, relationC);
    if (actorStr) (*env)->ReleaseStringUTFChars(env, actorStr, actorC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraTransaction_ACPDisableNACNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jlong identityPtr
) {
    Result res = ACPDisableNAC((uintptr_t)nodePtr, (uintptr_t)identityPtr);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraTransaction_ACPReEnableNACNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jlong identityPtr
) {
    Result res = ACPReEnableNAC((uintptr_t)nodePtr, (uintptr_t)identityPtr);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraTransaction_ACPAddNACActorRelationshipNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jlong identityPtr,
    jstring relationStr,
    jstring actorStr
) {
    const char* relationC = relationStr ? (*env)->GetStringUTFChars(env, relationStr, NULL) : NULL;
    const char* actorC = actorStr ? (*env)->GetStringUTFChars(env, actorStr, NULL) : NULL;
    Result res = ACPAddNACActorRelationship((uintptr_t)nodePtr, (uintptr_t)identityPtr, (char*)relationC, (char*)actorC);
    if (relationStr) (*env)->ReleaseStringUTFChars(env, relationStr, relationC);
    if (actorStr) (*env)->ReleaseStringUTFChars(env, actorStr, actorC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraTransaction_ACPDeleteNACActorRelationshipNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jlong identityPtr,
    jstring relationStr,
    jstring actorStr
) {
    const char* relationC = relationStr ? (*env)->GetStringUTFChars(env, relationStr, NULL) : NULL;
    const char* actorC = actorStr ? (*env)->GetStringUTFChars(env, actorStr, NULL) : NULL;
    Result res = ACPDeleteNACActorRelationship((uintptr_t)nodePtr, (uintptr_t)identityPtr, (char*)relationC, (char*)actorC);
    if (relationStr) (*env)->ReleaseStringUTFChars(env, relationStr, relationC);
    if (actorStr) (*env)->ReleaseStringUTFChars(env, actorStr, actorC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraTransaction_ACPGetNACStatusNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jlong identityPtr
) {
    Result res = ACPGetNACStatus((uintptr_t)nodePtr, (uintptr_t)identityPtr);
    return returnDefraResult(env, res);
}

// Transaction Document Methods
JNIEXPORT jobject JNICALL Java_source_defra_DefraTransaction_AddDocumentNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring jsonStr,
    jboolean isEncrypted,
    jstring encryptedFieldsStr,
    jobject optionsObj,
    jlong identityPtr
) {
    const char* jsonC = jsonStr ? (*env)->GetStringUTFChars(env, jsonStr, NULL) : NULL;
    const char* encryptedFieldsC = encryptedFieldsStr ? (*env)->GetStringUTFChars(env, encryptedFieldsStr, NULL) : NULL;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj);
    int isEncryptedC = (isEncrypted == JNI_TRUE) ? 1 : 0;
    Result res = AddDocument((uintptr_t)nodePtr, (char*)jsonC, isEncryptedC, (char*)encryptedFieldsC, opts, (uintptr_t)identityPtr);
    if (jsonStr) (*env)->ReleaseStringUTFChars(env, jsonStr, jsonC);
    if (encryptedFieldsStr) (*env)->ReleaseStringUTFChars(env, encryptedFieldsStr, encryptedFieldsC);
    releaseJavaCollectionOptions(env, optionsObj, opts);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraTransaction_GetDocumentNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring docIDStr,
    jboolean showDeleted,
    jobject optionsObj,
    jlong identityPtr
) {
    const char* docIDC = docIDStr ? (*env)->GetStringUTFChars(env, docIDStr, NULL) : NULL;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj);
    int showDeletedC = (showDeleted == JNI_TRUE) ? 1 : 0;
    Result res = GetDocument((uintptr_t)nodePtr, (char*)docIDC, showDeletedC, opts, (uintptr_t)identityPtr);
    if (docIDStr) (*env)->ReleaseStringUTFChars(env, docIDStr, docIDC);
    releaseJavaCollectionOptions(env, optionsObj, opts);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraTransaction_UpdateDocumentNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring docIDStr,
    jstring filterStr,
    jstring updaterStr,
    jobject optionsObj,
    jlong identityPtr
) {
    const char* docIDC = docIDStr ? (*env)->GetStringUTFChars(env, docIDStr, NULL) : NULL;
    const char* filterC = filterStr ? (*env)->GetStringUTFChars(env, filterStr, NULL) : NULL;
    const char* updaterC = updaterStr ? (*env)->GetStringUTFChars(env, updaterStr, NULL) : NULL;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj);
    Result res = UpdateDocument((uintptr_t)nodePtr, (char*)docIDC, (char*)filterC, (char*)updaterC, opts, (uintptr_t)identityPtr);
    if (docIDStr) (*env)->ReleaseStringUTFChars(env, docIDStr, docIDC);
    if (filterStr) (*env)->ReleaseStringUTFChars(env, filterStr, filterC);
    if (updaterStr) (*env)->ReleaseStringUTFChars(env, updaterStr, updaterC);
    releaseJavaCollectionOptions(env, optionsObj, opts);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraTransaction_DeleteDocumentNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring docIDStr,
    jstring filterStr,
    jobject optionsObj,
    jlong identityPtr
) {
    const char* docIDC = docIDStr ? (*env)->GetStringUTFChars(env, docIDStr, NULL) : NULL;
    const char* filterC = filterStr ? (*env)->GetStringUTFChars(env, filterStr, NULL) : NULL;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj);
    Result res = DeleteDocument((uintptr_t)nodePtr, (char*)docIDC, (char*)filterC, opts, (uintptr_t)identityPtr);
    if (docIDStr) (*env)->ReleaseStringUTFChars(env, docIDStr, docIDC);
    if (filterStr) (*env)->ReleaseStringUTFChars(env, filterStr, filterC);
    releaseJavaCollectionOptions(env, optionsObj, opts);
    return returnDefraResult(env, res);
}

// Transaction Index Methods
JNIEXPORT jobject JNICALL Java_source_defra_DefraTransaction_NewIndexNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring indexNameStr,
    jstring fieldsStr,
    jboolean isUnique,
    jobject optionsObj,
    jlong identityPtr
) {
    const char* indexNameC = indexNameStr ? (*env)->GetStringUTFChars(env, indexNameStr, NULL) : NULL;
    const char* fieldsC = fieldsStr ? (*env)->GetStringUTFChars(env, fieldsStr, NULL) : NULL;
    int isUniqueC = (isUnique == JNI_TRUE) ? 1 : 0;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj);
    Result res = NewIndex((uintptr_t)nodePtr, (char*)indexNameC, (char*)fieldsC, isUniqueC, opts, (uintptr_t)identityPtr);
    if (indexNameStr) (*env)->ReleaseStringUTFChars(env, indexNameStr, indexNameC);
    if (fieldsStr) (*env)->ReleaseStringUTFChars(env, fieldsStr, fieldsC);
    releaseJavaCollectionOptions(env, optionsObj, opts);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraTransaction_ListIndexesNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jobject optionsObj,
    jlong identityPtr
) {
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj);
    Result res = ListIndexes((uintptr_t)nodePtr, opts, (uintptr_t)identityPtr);
    releaseJavaCollectionOptions(env, optionsObj, opts);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraTransaction_DeleteIndexNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring indexNameStr,
    jobject optionsObj,
    jlong identityPtr
) {
    const char* indexNameC = indexNameStr ? (*env)->GetStringUTFChars(env, indexNameStr, NULL) : NULL;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj);
    Result res = DeleteIndex((uintptr_t)nodePtr, (char*)indexNameC, opts, (uintptr_t)identityPtr);
    if (indexNameStr) (*env)->ReleaseStringUTFChars(env, indexNameStr, indexNameC);
    releaseJavaCollectionOptions(env, optionsObj, opts);
    return returnDefraResult(env, res);
}

// Transaction P2P Methods
JNIEXPORT jobject JNICALL Java_source_defra_DefraTransaction_GetP2PInfoNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jlong identityPtr
) {
    Result res = GetP2PInfo((uintptr_t)nodePtr, (uintptr_t)identityPtr);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraTransaction_ListP2PReplicatorsNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jlong identityPtr
) {
    Result res = ListP2PReplicators((uintptr_t)nodePtr, (uintptr_t)identityPtr);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraTransaction_AddP2PReplicatorNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring collectionsStr,
    jstring addressesStr,
    jlong identityPtr
) {
    const char* collectionsC = collectionsStr ? (*env)->GetStringUTFChars(env, collectionsStr, NULL) : NULL;
    const char* addressesC = addressesStr ? (*env)->GetStringUTFChars(env, addressesStr, NULL) : NULL;
    Result res = AddP2PReplicator((uintptr_t)nodePtr, (char*)collectionsC, (char*)addressesC, (uintptr_t)identityPtr);
    if (collectionsStr) (*env)->ReleaseStringUTFChars(env, collectionsStr, collectionsC);
    if (addressesStr) (*env)->ReleaseStringUTFChars(env, addressesStr, addressesC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraTransaction_ConnectP2PPeersNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring peerAddressesStr,
    jlong identityPtr
) {
    const char* peerAddressesC = peerAddressesStr ? (*env)->GetStringUTFChars(env, peerAddressesStr, NULL) : NULL;
    Result res = ConnectP2PPeers((uintptr_t)nodePtr, (char*)peerAddressesC, (uintptr_t)identityPtr);
    if (peerAddressesStr) (*env)->ReleaseStringUTFChars(env, peerAddressesStr, peerAddressesC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraTransaction_DisconnectP2PPeersNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring peerAddressesStr,
    jlong identityPtr
) {
    const char* peerAddressesC = peerAddressesStr ? (*env)->GetStringUTFChars(env, peerAddressesStr, NULL) : NULL;
    Result res = DisconnectP2PPeers((uintptr_t)nodePtr, (char*)peerAddressesC, (uintptr_t)identityPtr);
    if (peerAddressesStr) (*env)->ReleaseStringUTFChars(env, peerAddressesStr, peerAddressesC);
    return returnDefraResult(env, res);
}

// Transaction Query Methods
JNIEXPORT jobject JNICALL Java_source_defra_DefraTransaction_ExecuteQueryNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring queryStr,
    jlong identityPtr,
    jstring operationNameStr,
    jstring variablesStr
) {
    const char* queryC = queryStr ? (*env)->GetStringUTFChars(env, queryStr, NULL) : NULL;
    const char* operationNameC = operationNameStr ? (*env)->GetStringUTFChars(env, operationNameStr, NULL) : NULL;
    const char* variablesC = variablesStr ? (*env)->GetStringUTFChars(env, variablesStr, NULL) : NULL;
    Result res = ExecuteQuery((uintptr_t)nodePtr, (char*)queryC, (uintptr_t)identityPtr, (char*)operationNameC, (char*)variablesC);
    if (queryStr) (*env)->ReleaseStringUTFChars(env, queryStr, queryC);
    if (operationNameStr) (*env)->ReleaseStringUTFChars(env, operationNameStr, operationNameC);
    if (variablesStr) (*env)->ReleaseStringUTFChars(env, variablesStr, variablesC);
    return returnDefraResult(env, res);
}

// Transaction Collection Methods
JNIEXPORT jobject JNICALL Java_source_defra_DefraTransaction_AddCollectionNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring sdlStr,
    jlong identityPtr
) {
    const char* sdlC = sdlStr ? (*env)->GetStringUTFChars(env, sdlStr, NULL) : NULL;
    Result res = AddCollection((uintptr_t)nodePtr, (char*)sdlC, (uintptr_t)identityPtr);
    if (sdlStr) (*env)->ReleaseStringUTFChars(env, sdlStr, sdlC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraTransaction_PatchCollectionNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring patchStr,
    jstring lensConfigStr,
    jlong identityPtr
) {
    const char* patchC = patchStr ? (*env)->GetStringUTFChars(env, patchStr, NULL) : NULL;
    const char* lensConfigC = lensConfigStr ? (*env)->GetStringUTFChars(env, lensConfigStr, NULL) : NULL;
    Result res = PatchCollection((uintptr_t)nodePtr, (char*)patchC, (char*)lensConfigC, (uintptr_t)identityPtr);
    if (patchStr) (*env)->ReleaseStringUTFChars(env, patchStr, patchC);
    if (lensConfigStr) (*env)->ReleaseStringUTFChars(env, lensConfigStr, lensConfigC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraTransaction_SetActiveCollectionNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jobject optionsObj,
    jlong identityPtr
) {
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj);
    Result res = SetActiveCollection((uintptr_t)nodePtr, opts, (uintptr_t)identityPtr);
    releaseJavaCollectionOptions(env, optionsObj, opts);
    return returnDefraResult(env, res);
}

// Transaction View Methods
JNIEXPORT jobject JNICALL Java_source_defra_DefraTransaction_AddViewNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring queryStr,
    jstring sdlStr,
    jstring transformCIDStr,
    jlong identityPtr
) {
    const char* queryC = queryStr ? (*env)->GetStringUTFChars(env, queryStr, NULL) : NULL;
    const char* sdlC = sdlStr ? (*env)->GetStringUTFChars(env, sdlStr, NULL) : NULL;
    const char* transformCIDC = transformCIDStr ? (*env)->GetStringUTFChars(env, transformCIDStr, NULL) : NULL;
    Result res = AddView((uintptr_t)nodePtr, (char*)queryC, (char*)sdlC, (char*)transformCIDC, (uintptr_t)identityPtr);
    if (queryStr) (*env)->ReleaseStringUTFChars(env, queryStr, queryC);
    if (sdlStr) (*env)->ReleaseStringUTFChars(env, sdlStr, sdlC);
    if (transformCIDStr) (*env)->ReleaseStringUTFChars(env, transformCIDStr, transformCIDC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraTransaction_RefreshViewNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jobject optionsObj,
    jlong identityPtr
) {
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj);
    Result res = RefreshView((uintptr_t)nodePtr, opts, (uintptr_t)identityPtr);
    releaseJavaCollectionOptions(env, optionsObj, opts);
    return returnDefraResult(env, res);
}

// The following DefraTransaction natives were missing entirely (declared in
// DefraTransaction.java but never implemented here upstream) - added so this
// client can genuinely exercise DefraTransaction's own local-data/schema
// methods, rather than only ever calling DefraNode's equivalents with a
// transaction handle substituted in. Mirrors the DefraNode implementations
// of the same name exactly, since they call the same cbindings functions.
JNIEXPORT jobject JNICALL Java_source_defra_DefraTransaction_DescribeCollectionNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jobject optionsObj,
    jlong identityPtr
) {
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj);
    Result res = DescribeCollection((uintptr_t)nodePtr, opts, (uintptr_t)identityPtr);
    releaseJavaCollectionOptions(env, optionsObj, opts);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraTransaction_TruncateCollectionNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jobject optionsObj,
    jlong identityPtr
) {
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj);
    Result res = TruncateCollection((uintptr_t)nodePtr, opts, (uintptr_t)identityPtr);
    releaseJavaCollectionOptions(env, optionsObj, opts);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraTransaction_NewEncryptedIndexNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring collectionNameStr,
    jstring fieldNameStr,
    jlong identityPtr
) {
    const char* collectionNameC = collectionNameStr ? (*env)->GetStringUTFChars(env, collectionNameStr, NULL) : NULL;
    const char* fieldNameC = fieldNameStr ? (*env)->GetStringUTFChars(env, fieldNameStr, NULL) : NULL;
    Result res = NewEncryptedIndex((uintptr_t)nodePtr, (char*)collectionNameC, (char*)fieldNameC, (uintptr_t)identityPtr);
    if (collectionNameStr) (*env)->ReleaseStringUTFChars(env, collectionNameStr, collectionNameC);
    if (fieldNameStr) (*env)->ReleaseStringUTFChars(env, fieldNameStr, fieldNameC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraTransaction_ListEncryptedIndexesNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring collectionNameStr,
    jlong identityPtr
) {
    const char* collectionNameC = collectionNameStr ? (*env)->GetStringUTFChars(env, collectionNameStr, NULL) : NULL;
    Result res = ListEncryptedIndexes((uintptr_t)nodePtr, (char*)collectionNameC, (uintptr_t)identityPtr);
    if (collectionNameStr) (*env)->ReleaseStringUTFChars(env, collectionNameStr, collectionNameC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraTransaction_DeleteEncryptedIndexNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring collectionNameStr,
    jstring fieldNameStr,
    jlong identityPtr
) {
    const char* collectionNameC = collectionNameStr ? (*env)->GetStringUTFChars(env, collectionNameStr, NULL) : NULL;
    const char* fieldNameC = fieldNameStr ? (*env)->GetStringUTFChars(env, fieldNameStr, NULL) : NULL;
    Result res = DeleteEncryptedIndex((uintptr_t)nodePtr, (char*)collectionNameC, (char*)fieldNameC, (uintptr_t)identityPtr);
    if (collectionNameStr) (*env)->ReleaseStringUTFChars(env, collectionNameStr, collectionNameC);
    if (fieldNameStr) (*env)->ReleaseStringUTFChars(env, fieldNameStr, fieldNameC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraTransaction_SetLensNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jlong identityPtr,
    jstring srcStr,
    jstring dstStr,
    jstring cfgStr
) {
    const char* srcC = srcStr ? (*env)->GetStringUTFChars(env, srcStr, NULL) : NULL;
    const char* dstC = dstStr ? (*env)->GetStringUTFChars(env, dstStr, NULL) : NULL;
    const char* cfgC = cfgStr ? (*env)->GetStringUTFChars(env, cfgStr, NULL) : NULL;
    Result res = SetLens((uintptr_t)nodePtr, (uintptr_t)identityPtr, (char*)srcC, (char*)dstC, (char*)cfgC);
    if (srcStr) (*env)->ReleaseStringUTFChars(env, srcStr, srcC);
    if (dstStr) (*env)->ReleaseStringUTFChars(env, dstStr, dstC);
    if (cfgStr) (*env)->ReleaseStringUTFChars(env, cfgStr, cfgC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraTransaction_AddLensNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jlong identityPtr,
    jstring cfgStr
) {
    const char* cfgC = cfgStr ? (*env)->GetStringUTFChars(env, cfgStr, NULL) : NULL;
    Result res = AddLens((uintptr_t)nodePtr, (uintptr_t)identityPtr, (char*)cfgC);
    if (cfgStr) (*env)->ReleaseStringUTFChars(env, cfgStr, cfgC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraTransaction_ListLensesNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jlong identityPtr
) {
    Result res = ListLenses((uintptr_t)nodePtr, (uintptr_t)identityPtr);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraTransaction_VerifyBlockSignatureNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring keyTypeStr,
    jstring publicKeyStr,
    jstring cidStr,
    jlong identityPtr
) {
    const char* keyTypeC = keyTypeStr ? (*env)->GetStringUTFChars(env, keyTypeStr, NULL) : NULL;
    const char* publicKeyC = publicKeyStr ? (*env)->GetStringUTFChars(env, publicKeyStr, NULL) : NULL;
    const char* cidC = cidStr ? (*env)->GetStringUTFChars(env, cidStr, NULL) : NULL;
    Result res = VerifyBlockSignature((uintptr_t)nodePtr, (char*)keyTypeC, (char*)publicKeyC, (char*)cidC, (uintptr_t)identityPtr);
    if (keyTypeStr) (*env)->ReleaseStringUTFChars(env, keyTypeStr, keyTypeC);
    if (publicKeyStr) (*env)->ReleaseStringUTFChars(env, publicKeyStr, publicKeyC);
    if (cidStr) (*env)->ReleaseStringUTFChars(env, cidStr, cidC);
    return returnDefraResult(env, res);
}

//=============================================================================
// DefraCollection JNI Functions
//=============================================================================

// Document Methods
JNIEXPORT jobject JNICALL Java_source_defra_DefraCollection_AddDocumentNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring jsonStr,
    jboolean isEncrypted,
    jstring encryptedFieldsStr,
    jobject optionsObj,
    jlong identityPtr
) {
    const char* jsonC = jsonStr ? (*env)->GetStringUTFChars(env, jsonStr, NULL) : NULL;
    const char* encryptedFieldsC = encryptedFieldsStr ? (*env)->GetStringUTFChars(env, encryptedFieldsStr, NULL) : NULL;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj);
    int isEncryptedC = (isEncrypted == JNI_TRUE) ? 1 : 0;
    Result res = AddDocument((uintptr_t)nodePtr, (char*)jsonC, isEncryptedC, (char*)encryptedFieldsC, opts, (uintptr_t)identityPtr);
    if (jsonStr) (*env)->ReleaseStringUTFChars(env, jsonStr, jsonC);
    if (encryptedFieldsStr) (*env)->ReleaseStringUTFChars(env, encryptedFieldsStr, encryptedFieldsC);
    releaseJavaCollectionOptions(env, optionsObj, opts);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraCollection_DeleteDocumentNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring docIDStr,
    jstring filterStr,
    jobject optionsObj,
    jlong identityPtr
) {
    const char* docIDC = docIDStr ? (*env)->GetStringUTFChars(env, docIDStr, NULL) : NULL;
    const char* filterC = filterStr ? (*env)->GetStringUTFChars(env, filterStr, NULL) : NULL;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj);
    Result res = DeleteDocument((uintptr_t)nodePtr, (char*)docIDC, (char*)filterC, opts, (uintptr_t)identityPtr);
    if (docIDStr) (*env)->ReleaseStringUTFChars(env, docIDStr, docIDC);
    if (filterStr) (*env)->ReleaseStringUTFChars(env, filterStr, filterC);
    releaseJavaCollectionOptions(env, optionsObj, opts);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraCollection_GetDocumentNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring docIDStr,
    jboolean showDeleted,
    jobject optionsObj,
    jlong identityPtr
) {
    const char* docIDC = docIDStr ? (*env)->GetStringUTFChars(env, docIDStr, NULL) : NULL;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj);
    int showDeletedC = (showDeleted == JNI_TRUE) ? 1 : 0;
    Result res = GetDocument((uintptr_t)nodePtr, (char*)docIDC, showDeletedC, opts, (uintptr_t)identityPtr);
    if (docIDStr) (*env)->ReleaseStringUTFChars(env, docIDStr, docIDC);
    releaseJavaCollectionOptions(env, optionsObj, opts);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraCollection_UpdateDocumentNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring docIDStr,
    jstring filterStr,
    jstring updaterStr,
    jobject optionsObj,
    jlong identityPtr
) {
    const char* docIDC = docIDStr ? (*env)->GetStringUTFChars(env, docIDStr, NULL) : NULL;
    const char* filterC = filterStr ? (*env)->GetStringUTFChars(env, filterStr, NULL) : NULL;
    const char* updaterC = updaterStr ? (*env)->GetStringUTFChars(env, updaterStr, NULL) : NULL;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj);
    Result res = UpdateDocument((uintptr_t)nodePtr, (char*)docIDC, (char*)filterC, (char*)updaterC, opts, (uintptr_t)identityPtr);
    if (docIDStr) (*env)->ReleaseStringUTFChars(env, docIDStr, docIDC);
    if (filterStr) (*env)->ReleaseStringUTFChars(env, filterStr, filterC);
    if (updaterStr) (*env)->ReleaseStringUTFChars(env, updaterStr, updaterC);
    releaseJavaCollectionOptions(env, optionsObj, opts);
    return returnDefraResult(env, res);
}

// Encrypted Index Methods
JNIEXPORT jobject JNICALL Java_source_defra_DefraCollection_NewEncryptedIndexNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring collectionNameStr,
    jstring fieldNameStr,
    jlong identityPtr
) {
    const char* collectionNameC = collectionNameStr ? (*env)->GetStringUTFChars(env, collectionNameStr, NULL) : NULL;
    const char* fieldNameC = fieldNameStr ? (*env)->GetStringUTFChars(env, fieldNameStr, NULL) : NULL;
    Result res = NewEncryptedIndex((uintptr_t)nodePtr, (char*)collectionNameC, (char*)fieldNameC, (uintptr_t)identityPtr);
    if (collectionNameStr) (*env)->ReleaseStringUTFChars(env, collectionNameStr, collectionNameC);
    if (fieldNameStr) (*env)->ReleaseStringUTFChars(env, fieldNameStr, fieldNameC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraCollection_ListEncryptedIndexesNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring collectionNameStr,
    jlong identityPtr
) {
    const char* collectionNameC = collectionNameStr ? (*env)->GetStringUTFChars(env, collectionNameStr, NULL) : NULL;
    Result res = ListEncryptedIndexes((uintptr_t)nodePtr, (char*)collectionNameC, (uintptr_t)identityPtr);
    if (collectionNameStr) (*env)->ReleaseStringUTFChars(env, collectionNameStr, collectionNameC);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraCollection_DeleteEncryptedIndexNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring collectionNameStr,
    jstring fieldNameStr,
    jlong identityPtr
) {
    const char* collectionNameC = collectionNameStr ? (*env)->GetStringUTFChars(env, collectionNameStr, NULL) : NULL;
    const char* fieldNameC = fieldNameStr ? (*env)->GetStringUTFChars(env, fieldNameStr, NULL) : NULL;
    Result res = DeleteEncryptedIndex((uintptr_t)nodePtr, (char*)collectionNameC, (char*)fieldNameC, (uintptr_t)identityPtr);
    if (collectionNameStr) (*env)->ReleaseStringUTFChars(env, collectionNameStr, collectionNameC);
    if (fieldNameStr) (*env)->ReleaseStringUTFChars(env, fieldNameStr, fieldNameC);
    return returnDefraResult(env, res);
}

// Index Methods
JNIEXPORT jobject JNICALL Java_source_defra_DefraCollection_NewIndexNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring indexNameStr,
    jstring fieldsStr,
    jboolean isUnique,
    jobject optionsObj,
    jlong identityPtr
) {
    const char* indexNameC = indexNameStr ? (*env)->GetStringUTFChars(env, indexNameStr, NULL) : NULL;
    const char* fieldsC = fieldsStr ? (*env)->GetStringUTFChars(env, fieldsStr, NULL) : NULL;
    int isUniqueC = (isUnique == JNI_TRUE) ? 1 : 0;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj);
    Result res = NewIndex((uintptr_t)nodePtr, (char*)indexNameC, (char*)fieldsC, isUniqueC, opts, (uintptr_t)identityPtr);
    if (indexNameStr) (*env)->ReleaseStringUTFChars(env, indexNameStr, indexNameC);
    if (fieldsStr) (*env)->ReleaseStringUTFChars(env, fieldsStr, fieldsC);
    releaseJavaCollectionOptions(env, optionsObj, opts);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraCollection_ListIndexesNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jobject optionsObj,
    jlong identityPtr
) {
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj);
    Result res = ListIndexes((uintptr_t)nodePtr, opts, (uintptr_t)identityPtr);
    releaseJavaCollectionOptions(env, optionsObj, opts);
    return returnDefraResult(env, res);
}

JNIEXPORT jobject JNICALL Java_source_defra_DefraCollection_DeleteIndexNative(
    JNIEnv* env,
    jobject thiz,
    jlong nodePtr,
    jstring indexNameStr,
    jobject optionsObj,
    jlong identityPtr
) {
    const char* indexNameC = indexNameStr ? (*env)->GetStringUTFChars(env, indexNameStr, NULL) : NULL;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj);
    Result res = DeleteIndex((uintptr_t)nodePtr, (char*)indexNameC, opts, (uintptr_t)identityPtr);
    if (indexNameStr) (*env)->ReleaseStringUTFChars(env, indexNameStr, indexNameC);
    releaseJavaCollectionOptions(env, optionsObj, opts);
    return returnDefraResult(env, res);
}
