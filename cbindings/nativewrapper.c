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

// ConvertCtx threads an ok flag through a sequence of field lookups, so that after the
// first failur every subsequent ctx_get_* call in the same conversion becomes a no-op instead of chaining further
// JNI calls onto a pending exception.
typedef struct {
    JNIEnv* env;
    jobject obj;
    jclass cls;
    int ok;
} ConvertCtx;

// ctx_get_object_field looks up and reads an Object field, pending the context still being ok.
static jstring ctx_get_object_field(ConvertCtx* ctx, const char* name, const char* sig) {
    if (!ctx->ok) return NULL;
    jfieldID fid = (*ctx->env)->GetFieldID(ctx->env, ctx->cls, name, sig);
    if (fid == NULL) {
        ctx->ok = 0;
        return NULL;
    }
    return (jstring)(*ctx->env)->GetObjectField(ctx->env, ctx->obj, fid);
}

// ctx_get_utff looks up and reads a String fffield, pinning its native UTF-8 chars. This happens only
// iff the context is still ok. Otherwise, a nuull string will be returrned.
static const char* ctx_get_utf(ConvertCtx* ctx, const char* name) {
    jstring s = ctx_get_object_field(ctx, name, "Ljava/lang/String;");
    if (!ctx->ok || s == NULL) return NULL;
    const char* chars = (*ctx->env)->GetStringUTFChars(ctx->env, s, NULL);
    if (chars == NULL) ctx->ok = 0; // out of memory
    return chars;
}

// ctx_get_bytearray looks up and reads a byte[] field, pending the context still being ok.
static jbyteArray ctx_get_bytearray(ConvertCtx* ctx, const char* name) {
    return (jbyteArray)ctx_get_object_field(ctx, name, "[B");
}

// ctx_get_bool looks up and reads a boolean field, pending the context still being ok.
static jboolean ctx_get_bool(ConvertCtx* ctx, const char* name) {
    if (!ctx->ok) return JNI_FALSE;
    jfieldID fid = (*ctx->env)->GetFieldID(ctx->env, ctx->cls, name, "Z");
    if (fid == NULL) {
        ctx->ok = 0;
        return JNI_FALSE;
    }
    return (*ctx->env)->GetBooleanField(ctx->env, ctx->obj, fid);
}

// ctx_get_int looks up and reads an int field, pending the context still being ok.
static jint ctx_get_int(ConvertCtx* ctx, const char* name) {
    if (!ctx->ok) return 0;
    jfieldID fid = (*ctx->env)->GetFieldID(ctx->env, ctx->cls, name, "I");
    if (fid == NULL) {
        ctx->ok = 0;
        return 0;
    }
    return (*ctx->env)->GetIntField(ctx->env, ctx->obj, fid);
}

// ctx_get_long looks up and reads a long field, pending the context still being ok.
static jlong ctx_get_long(ConvertCtx* ctx, const char* name) {
    if (!ctx->ok) return 0;
    jfieldID fid = (*ctx->env)->GetFieldID(ctx->env, ctx->cls, name, "J");
    if (fid == NULL) {
        ctx->ok = 0;
        return 0;
    }
    return (*ctx->env)->GetLongField(ctx->env, ctx->obj, fid);
}

// safe_field_id looks up a field ID and immediately clears any resulting pending exception.
static jfieldID safe_field_id(JNIEnv* env, jclass cls, const char* name, const char* sig) {
    jfieldID fid = (*env)->GetFieldID(env, cls, name, sig);
    if (fid == NULL) (*env)->ExceptionClear(env);
    return fid;
}

// throwNullOptions throws a NullPointerException for native methods that received a null options
// object where a real one is required.
static void throwNullOptions(JNIEnv* env, const char* message) {
    jclass nullPointerExcepetionCls = (*env)->FindClass(env, "java/lang/NullPointerException");
    if (nullPointerExcepetionCls != NULL) {
        (*env)->ThrowNew(env, nullPointerExcepetionCls, message);
    }
}

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

// discardOrphanedNode releases a Node handle that was successfully created on the Go side but
// can no longer be returned to Java (because constructing its Java wrapper object failed.)
static void discardOrphanedNode(uintptr_t nodePtr) {
    if (nodePtr == 0) return;
    Result closeRes = CloseNode(nodePtr);
    if (closeRes.error) free(closeRes.error);
    if (closeRes.value) free(closeRes.value);
}

// discardOrphanedIdentity is discardOrphanedNode's counterpart for identity handles.
static void discardOrphanedIdentity(uintptr_t identityPtr) {
    if (identityPtr == 0) return;
    FreeIdentity(identityPtr);
}

// discardOrphanedTransaction is discardOrphanedNode's counterpart for transaction handles.
static void discardOrphanedTransaction(uintptr_t txnPtr) {
    if (txnPtr == 0) return;
    DiscardTransaction(txnPtr);
}

jobject returnDefraNewNodeResult(JNIEnv* env, NewNodeResult res) {
    jstring errorStr = res.error ? (*env)->NewStringUTF(env, res.error) : NULL;
    if (res.error) free(res.error);
    jclass cls = (*env)->FindClass(env, "source/defra/DefraNewNodeResult");
    jmethodID ctor = cls ? (*env)->GetMethodID(env, cls, "<init>", "(ILjava/lang/String;J)V") : NULL;
    jobject resultObj = ctor ? (*env)->NewObject(env, cls, ctor, (jint)res.status, errorStr, (jlong)res.nodePtr) : NULL;
    if (resultObj == NULL) {
        discardOrphanedNode((uintptr_t)res.nodePtr);
    }
    return resultObj;
}

jobject returnDefraIdentityResult(JNIEnv* env, NewIdentityResult res) {
    jstring errorStr = res.error ? (*env)->NewStringUTF(env, res.error) : NULL;
    if (res.error) free(res.error);
    jclass cls = (*env)->FindClass(env, "source/defra/DefraIdentityResult");
    jmethodID ctor = cls ? (*env)->GetMethodID(env, cls, "<init>", "(ILjava/lang/String;J)V") : NULL;
    jobject resultObj = ctor ? (*env)->NewObject(env, cls, ctor, (jint)res.status, errorStr, (jlong)res.identityPtr) : NULL;
    if (resultObj == NULL) {
        discardOrphanedIdentity((uintptr_t)res.identityPtr);
    }
    return resultObj;
}

jobject returnDefraTransactionResult(JNIEnv* env, NewTxnResult res) {
    jstring errorStr = res.error ? (*env)->NewStringUTF(env, res.error) : NULL;
    if (res.error) free(res.error);
    jclass cls = (*env)->FindClass(env, "source/defra/DefraTransactionResult");
    jmethodID ctor = cls ? (*env)->GetMethodID(env, cls, "<init>", "(ILjava/lang/String;J)V") : NULL;
    jobject resultObj = ctor ? (*env)->NewObject(env, cls, ctor, (jint)res.status, errorStr, (jlong)res.txnPtr) : NULL;
    if (resultObj == NULL) {
        discardOrphanedTransaction((uintptr_t)res.txnPtr);
    }
    return resultObj;
}

// convertJavaNodeInitOptions is a helper function to convert a Java DefraNodeInitOptions object to a C NodeInitOptions 
// struct. *ok is set to 1 on success, or 0 if optionsObj was null or any field lookup/read failed. If that happens, 
// the returned struct is already fully released and zeroed out, and a Java exception is left pending.
NodeInitOptions convertJavaNodeInitOptions(JNIEnv* env, jobject optionsObj, int* ok) {
    NodeInitOptions opts;
    memset(&opts, 0, sizeof(NodeInitOptions));

    if (optionsObj == NULL) {
        throwNullOptions(env, "node init options must not be null");
        *ok = 0;
        return opts;
    }

    ConvertCtx ctx = { env, optionsObj, (*env)->GetObjectClass(env, optionsObj), 1 };

    // Core strings
    opts.dbPath = ctx_get_utf(&ctx, "dbPath");
    opts.listeningAddresses = ctx_get_utf(&ctx, "listeningAddresses");
    opts.replicatorRetryIntervals = ctx_get_utf(&ctx, "replicatorRetryIntervals");
    opts.peers = ctx_get_utf(&ctx, "peers");

    // Core booleans/ints
    opts.inMemory = ctx_get_bool(&ctx, "inMemory") ? 1 : 0;
    opts.disableP2P = ctx_get_bool(&ctx, "disableP2P") ? 1 : 0;
    opts.disableAPI = ctx_get_bool(&ctx, "disableAPI") ? 1 : 0;
    opts.enableNodeACP = ctx_get_bool(&ctx, "enableNodeACP") ? 1 : 0;
    opts.maxTransactionRetries = ctx_get_int(&ctx, "maxTransactionRetries");

    // Identity - a nested object field, handled by hand since it needs its own class/field lookup
    if (ctx.ok) {
        jfieldID fid_identity = (*env)->GetFieldID(env, ctx.cls, "identity", "Lsource/defra/DefraIdentity;");
        if (fid_identity == NULL) {
            ctx.ok = 0;
        } else {
            jobject identityObj = (*env)->GetObjectField(env, optionsObj, fid_identity);
            if (identityObj != NULL) {
                jclass identityCls = (*env)->GetObjectClass(env, identityObj);
                jfieldID fid_ptr = (*env)->GetFieldID(env, identityCls, "ptr", "J");
                if (fid_ptr == NULL) {
                    ctx.ok = 0;
                } else {
                    opts.identityPtr = (uintptr_t)(*env)->GetLongField(env, identityObj, fid_ptr);
                }
            }
        }
    }

    // Store options
    opts.storeType = ctx_get_utf(&ctx, "storeType");
    opts.badgerFileSize = ctx_get_long(&ctx, "badgerFileSize");

    jbyteArray badgerEncryptionKeyArr = ctx_get_bytearray(&ctx, "badgerEncryptionKey");
    if (ctx.ok && badgerEncryptionKeyArr != NULL) {
        opts.badgerEncryptionKey = (uint8_t*)(*env)->GetByteArrayElements(env, badgerEncryptionKeyArr, NULL);
        opts.badgerEncryptionKeyLen = (int)(*env)->GetArrayLength(env, badgerEncryptionKeyArr);
    }

    // DB options
    opts.enableSigning = ctx_get_bool(&ctx, "enableSigning") ? 1 : 0;

    jbyteArray searchableEncryptionKeyArr = ctx_get_bytearray(&ctx, "searchableEncryptionKey");
    if (ctx.ok && searchableEncryptionKeyArr != NULL) {
        opts.searchableEncryptionKey = (uint8_t*)(*env)->GetByteArrayElements(env, searchableEncryptionKeyArr, NULL);
        opts.searchableEncryptionKeyLen = (int)(*env)->GetArrayLength(env, searchableEncryptionKeyArr);
    }

    opts.p2pBlockSyncTimeoutMs = ctx_get_long(&ctx, "p2pBlockSyncTimeoutMs");
    opts.lensPoolSize = ctx_get_int(&ctx, "lensPoolSize");
    opts.chunkSize = ctx_get_int(&ctx, "chunkSize");

    // P2P options
    opts.enablePubSub = ctx_get_bool(&ctx, "enablePubSub") ? 1 : 0;
    opts.enableRelay = ctx_get_bool(&ctx, "enableRelay") ? 1 : 0;
    opts.enableClearBackoffOnRetry = ctx_get_bool(&ctx, "enableClearBackoffOnRetry") ? 1 : 0;

    jbyteArray p2pPrivateKeyArr = ctx_get_bytearray(&ctx, "p2pPrivateKey");
    if (ctx.ok && p2pPrivateKeyArr != NULL) {
        opts.p2pPrivateKey = (uint8_t*)(*env)->GetByteArrayElements(env, p2pPrivateKeyArr, NULL);
        opts.p2pPrivateKeyLen = (int)(*env)->GetArrayLength(env, p2pPrivateKeyArr);
    }

    // HTTP options
    opts.httpAddress = ctx_get_utf(&ctx, "httpAddress");
    opts.httpAllowedOrigins = ctx_get_utf(&ctx, "httpAllowedOrigins");
    opts.tlsCertPath = ctx_get_utf(&ctx, "tlsCertPath");
    opts.tlsKeyPath = ctx_get_utf(&ctx, "tlsKeyPath");
    opts.httpReadTimeoutMs = ctx_get_long(&ctx, "httpReadTimeoutMs");
    opts.httpWriteTimeoutMs = ctx_get_long(&ctx, "httpWriteTimeoutMs");
    opts.httpIdleTimeoutMs = ctx_get_long(&ctx, "httpIdleTimeoutMs");

    // Document ACP options
    opts.documentACPType = ctx_get_utf(&ctx, "documentACPType");
    opts.documentACPPath = ctx_get_utf(&ctx, "documentACPPath");
    opts.sourceHubChainID = ctx_get_utf(&ctx, "sourceHubChainID");
    opts.sourceHubGRPCAddress = ctx_get_utf(&ctx, "sourceHubGRPCAddress");
    opts.sourceHubCometRPCAddress = ctx_get_utf(&ctx, "sourceHubCometRPCAddress");

    // Node ACP options
    opts.nodeACPPath = ctx_get_utf(&ctx, "nodeACPPath");

    *ok = ctx.ok;
    if (!ctx.ok) {
        // releaseJavaNodeInitOptions does its own GetFieldID lookups internally, which isn't 
        // documented-safe to do while the exception from our own failure is still pending.
        // Park it for the duration and restore it as the pending exception once cleanup is done.
        jthrowable pending = (*env)->ExceptionOccurred(env);
        if (pending != NULL) (*env)->ExceptionClear(env);
        releaseJavaNodeInitOptions(env, optionsObj, opts);
        if (pending != NULL) (*env)->Throw(env, pending);
        memset(&opts, 0, sizeof(NodeInitOptions));
    }
    return opts;
}

// convertJavaCollectionOptions is convertJavaNodeInitOptions' counterpart for CollectionOptions.
CollectionOptions convertJavaCollectionOptions(JNIEnv* env, jobject optionsObj, int* ok) {
    CollectionOptions opts;
    memset(&opts, 0, sizeof(CollectionOptions));

    if (optionsObj == NULL) {
        throwNullOptions(env, "collection options must not be null");
        *ok = 0;
        return opts;
    }

    ConvertCtx ctx = { env, optionsObj, (*env)->GetObjectClass(env, optionsObj), 1 };

    // Strings
    opts.version = ctx_get_utf(&ctx, "version");
    opts.collectionID = ctx_get_utf(&ctx, "collectionID");
    opts.name = ctx_get_utf(&ctx, "name");

    // Boolean
    opts.getInactive = ctx_get_bool(&ctx, "getInactive") ? 1 : 0;

    // enableSigning is a boxed java.lang.Boolean (nullable), matching CollectionOptions' own
    // tri-state: null means unset (0), otherwise 1 (true) or -1 (false) - handled by hand since it
    // needs a further method lookup (booleanValue) beyond a plain field read.
    if (ctx.ok) {
        jfieldID fid_enableSigning = (*env)->GetFieldID(env, ctx.cls, "enableSigning", "Ljava/lang/Boolean;");
        if (fid_enableSigning == NULL) {
            ctx.ok = 0;
        } else {
            jobject enableSigningObj = (*env)->GetObjectField(env, optionsObj, fid_enableSigning);
            if (enableSigningObj != NULL) {
                jclass booleanCls = (*env)->GetObjectClass(env, enableSigningObj);
                jmethodID booleanValueMid = (*env)->GetMethodID(env, booleanCls, "booleanValue", "()Z");
                if (booleanValueMid == NULL) {
                    ctx.ok = 0;
                } else {
                    opts.enableSigning = (*env)->CallBooleanMethod(env, enableSigningObj, booleanValueMid) ? 1 : -1;
                }
            } else {
                opts.enableSigning = 0;
            }
        }
    }

    *ok = ctx.ok;
    if (!ctx.ok) {
        // releaseJavaCollectionOptions does its own GetFieldID lookups internally, which isn't documented-safe to do 
        // while the exception from our own failure is still pending. Park it for the duration and restore it as the pending exception once cleanup is done.
        jthrowable pending = (*env)->ExceptionOccurred(env);
        if (pending != NULL) (*env)->ExceptionClear(env);
        releaseJavaCollectionOptions(env, optionsObj, opts);
        if (pending != NULL) (*env)->Throw(env, pending);
        memset(&opts, 0, sizeof(CollectionOptions));
    }
    return opts;
}

// Helper to release allocated Java strings after the call
void releaseJavaNodeInitOptions(JNIEnv* env, jobject optionsObj, NodeInitOptions opts) {
    jclass cls = (*env)->GetObjectClass(env, optionsObj);

    // Core strings
    jfieldID fid_dbPath = safe_field_id(env, cls, "dbPath", "Ljava/lang/String;");
    jfieldID fid_listeningAddresses = safe_field_id(env, cls, "listeningAddresses", "Ljava/lang/String;");
    jfieldID fid_replicatorRetryIntervals = safe_field_id(env, cls, "replicatorRetryIntervals", "Ljava/lang/String;");
    jfieldID fid_peers = safe_field_id(env, cls, "peers", "Ljava/lang/String;");

    if (opts.dbPath && fid_dbPath) {
        (*env)->ReleaseStringUTFChars(env, (jstring)(*env)->GetObjectField(env, optionsObj, fid_dbPath), opts.dbPath);
    }
    if (opts.listeningAddresses && fid_listeningAddresses) {
        (*env)->ReleaseStringUTFChars(
            env, (jstring)(*env)->GetObjectField(env, optionsObj, fid_listeningAddresses), opts.listeningAddresses);
    }
    if (opts.replicatorRetryIntervals && fid_replicatorRetryIntervals) {
        (*env)->ReleaseStringUTFChars(
            env,
            (jstring)(*env)->GetObjectField(env, optionsObj, fid_replicatorRetryIntervals),
            opts.replicatorRetryIntervals);
    }
    if (opts.peers && fid_peers) {
        (*env)->ReleaseStringUTFChars(env, (jstring)(*env)->GetObjectField(env, optionsObj, fid_peers), opts.peers);
    }

    // Store options
    jfieldID fid_storeType = safe_field_id(env, cls, "storeType", "Ljava/lang/String;");
    if (opts.storeType && fid_storeType) (*env)->ReleaseStringUTFChars(env, (jstring)(*env)->GetObjectField(env, optionsObj, fid_storeType), opts.storeType);

    jfieldID fid_badgerEncryptionKey = safe_field_id(env, cls, "badgerEncryptionKey", "[B");
    if (opts.badgerEncryptionKey && fid_badgerEncryptionKey) (*env)->ReleaseByteArrayElements(env, (jbyteArray)(*env)->GetObjectField(env, optionsObj, fid_badgerEncryptionKey), (jbyte*)opts.badgerEncryptionKey, JNI_ABORT);

    // DB options
    jfieldID fid_searchableEncryptionKey = safe_field_id(env, cls, "searchableEncryptionKey", "[B");
    if (opts.searchableEncryptionKey && fid_searchableEncryptionKey) (*env)->ReleaseByteArrayElements(env, (jbyteArray)(*env)->GetObjectField(env, optionsObj, fid_searchableEncryptionKey), (jbyte*)opts.searchableEncryptionKey, JNI_ABORT);

    // P2P options
    jfieldID fid_p2pPrivateKey = safe_field_id(env, cls, "p2pPrivateKey", "[B");
    if (opts.p2pPrivateKey && fid_p2pPrivateKey) (*env)->ReleaseByteArrayElements(env, (jbyteArray)(*env)->GetObjectField(env, optionsObj, fid_p2pPrivateKey), (jbyte*)opts.p2pPrivateKey, JNI_ABORT);

    // HTTP options
    jfieldID fid_httpAddress = safe_field_id(env, cls, "httpAddress", "Ljava/lang/String;");
    if (opts.httpAddress && fid_httpAddress) (*env)->ReleaseStringUTFChars(env, (jstring)(*env)->GetObjectField(env, optionsObj, fid_httpAddress), opts.httpAddress);

    jfieldID fid_httpAllowedOrigins = safe_field_id(env, cls, "httpAllowedOrigins", "Ljava/lang/String;");
    if (opts.httpAllowedOrigins && fid_httpAllowedOrigins) (*env)->ReleaseStringUTFChars(env, (jstring)(*env)->GetObjectField(env, optionsObj, fid_httpAllowedOrigins), opts.httpAllowedOrigins);

    jfieldID fid_tlsCertPath = safe_field_id(env, cls, "tlsCertPath", "Ljava/lang/String;");
    if (opts.tlsCertPath && fid_tlsCertPath) (*env)->ReleaseStringUTFChars(env, (jstring)(*env)->GetObjectField(env, optionsObj, fid_tlsCertPath), opts.tlsCertPath);

    jfieldID fid_tlsKeyPath = safe_field_id(env, cls, "tlsKeyPath", "Ljava/lang/String;");
    if (opts.tlsKeyPath && fid_tlsKeyPath) (*env)->ReleaseStringUTFChars(env, (jstring)(*env)->GetObjectField(env, optionsObj, fid_tlsKeyPath), opts.tlsKeyPath);

    // Document ACP options
    jfieldID fid_documentACPType = safe_field_id(env, cls, "documentACPType", "Ljava/lang/String;");
    if (opts.documentACPType && fid_documentACPType) (*env)->ReleaseStringUTFChars(env, (jstring)(*env)->GetObjectField(env, optionsObj, fid_documentACPType), opts.documentACPType);

    jfieldID fid_documentACPPath = safe_field_id(env, cls, "documentACPPath", "Ljava/lang/String;");
    if (opts.documentACPPath && fid_documentACPPath) (*env)->ReleaseStringUTFChars(env, (jstring)(*env)->GetObjectField(env, optionsObj, fid_documentACPPath), opts.documentACPPath);

    jfieldID fid_sourceHubChainID = safe_field_id(env, cls, "sourceHubChainID", "Ljava/lang/String;");
    if (opts.sourceHubChainID && fid_sourceHubChainID) (*env)->ReleaseStringUTFChars(env, (jstring)(*env)->GetObjectField(env, optionsObj, fid_sourceHubChainID), opts.sourceHubChainID);

    jfieldID fid_sourceHubGRPCAddress = safe_field_id(env, cls, "sourceHubGRPCAddress", "Ljava/lang/String;");
    if (opts.sourceHubGRPCAddress && fid_sourceHubGRPCAddress) (*env)->ReleaseStringUTFChars(env, (jstring)(*env)->GetObjectField(env, optionsObj, fid_sourceHubGRPCAddress), opts.sourceHubGRPCAddress);

    jfieldID fid_sourceHubCometRPCAddress = safe_field_id(env, cls, "sourceHubCometRPCAddress", "Ljava/lang/String;");
    if (opts.sourceHubCometRPCAddress && fid_sourceHubCometRPCAddress) (*env)->ReleaseStringUTFChars(env, (jstring)(*env)->GetObjectField(env, optionsObj, fid_sourceHubCometRPCAddress), opts.sourceHubCometRPCAddress);

    // Node ACP options
    jfieldID fid_nodeACPPath = safe_field_id(env, cls, "nodeACPPath", "Ljava/lang/String;");
    if (opts.nodeACPPath && fid_nodeACPPath) (*env)->ReleaseStringUTFChars(env, (jstring)(*env)->GetObjectField(env, optionsObj, fid_nodeACPPath), opts.nodeACPPath);
}

void releaseJavaCollectionOptions(JNIEnv* env, jobject optionsObj, CollectionOptions opts) {
    jclass cls = (*env)->GetObjectClass(env, optionsObj);
    jfieldID fid_version = safe_field_id(env, cls, "version", "Ljava/lang/String;");
    jfieldID fid_collectionID = safe_field_id(env, cls, "collectionID", "Ljava/lang/String;");
    jfieldID fid_name = safe_field_id(env, cls, "name", "Ljava/lang/String;");

    if (opts.version && fid_version) {
        (*env)->ReleaseStringUTFChars(env, (jstring)(*env)->GetObjectField(env, optionsObj, fid_version), opts.version);
    }
    if (opts.collectionID && fid_collectionID) {
        (*env)->ReleaseStringUTFChars(env, (jstring)(*env)->GetObjectField(env, optionsObj, fid_collectionID), opts.collectionID);
    }
    if (opts.name && fid_name) {
        (*env)->ReleaseStringUTFChars(env, (jstring)(*env)->GetObjectField(env, optionsObj, fid_name), opts.name);
    }
}

//=============================================================================
// DefraNode JNI Functions
//=============================================================================

JNIEXPORT jobject JNICALL Java_source_defra_DefraNode_NewNodeNative
(JNIEnv *env, jobject thisObj, jobject optionsObj) {
    int optsOk = 1;
    NodeInitOptions opts = convertJavaNodeInitOptions(env, optionsObj, &optsOk);
    if (!optsOk) return NULL;
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
    int optsOk = 1;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj, &optsOk);
    if (!optsOk) return NULL;
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
    int optsOk = 1;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj, &optsOk);
    if (!optsOk) return NULL;
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
    int optsOk = 1;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj, &optsOk);
    if (!optsOk) return NULL;
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
    int optsOk = 1;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj, &optsOk);
    if (!optsOk) return NULL;
    const char* jsonC = jsonStr ? (*env)->GetStringUTFChars(env, jsonStr, NULL) : NULL;
    const char* encryptedFieldsC = encryptedFieldsStr ? (*env)->GetStringUTFChars(env, encryptedFieldsStr, NULL) : NULL;
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
    int optsOk = 1;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj, &optsOk);
    if (!optsOk) return NULL;
    const char* docIDC = docIDStr ? (*env)->GetStringUTFChars(env, docIDStr, NULL) : NULL;
    const char* filterC = filterStr ? (*env)->GetStringUTFChars(env, filterStr, NULL) : NULL;
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
    int optsOk = 1;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj, &optsOk);
    if (!optsOk) return NULL;
    const char* docIDC = docIDStr ? (*env)->GetStringUTFChars(env, docIDStr, NULL) : NULL;
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
    int optsOk = 1;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj, &optsOk);
    if (!optsOk) return NULL;
    const char* docIDC = docIDStr ? (*env)->GetStringUTFChars(env, docIDStr, NULL) : NULL;
    const char* filterC = filterStr ? (*env)->GetStringUTFChars(env, filterStr, NULL) : NULL;
    const char* updaterC = updaterStr ? (*env)->GetStringUTFChars(env, updaterStr, NULL) : NULL;
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
    int optsOk = 1;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj, &optsOk);
    if (!optsOk) return NULL;
    const char* indexNameC = indexNameStr ? (*env)->GetStringUTFChars(env, indexNameStr, NULL) : NULL;
    const char* fieldsC = fieldsStr ? (*env)->GetStringUTFChars(env, fieldsStr, NULL) : NULL;
    int isUniqueC = (isUnique == JNI_TRUE) ? 1 : 0;
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
    int optsOk = 1;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj, &optsOk);
    if (!optsOk) return NULL;
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
    int optsOk = 1;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj, &optsOk);
    if (!optsOk) return NULL;
    const char* indexNameC = indexNameStr ? (*env)->GetStringUTFChars(env, indexNameStr, NULL) : NULL;
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
    int optsOk = 1;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj, &optsOk);
    if (!optsOk) return NULL;
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
    int optsOk = 1;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj, &optsOk);
    if (!optsOk) return NULL;
    const char* jsonC = jsonStr ? (*env)->GetStringUTFChars(env, jsonStr, NULL) : NULL;
    const char* encryptedFieldsC = encryptedFieldsStr ? (*env)->GetStringUTFChars(env, encryptedFieldsStr, NULL) : NULL;
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
    int optsOk = 1;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj, &optsOk);
    if (!optsOk) return NULL;
    const char* docIDC = docIDStr ? (*env)->GetStringUTFChars(env, docIDStr, NULL) : NULL;
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
    int optsOk = 1;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj, &optsOk);
    if (!optsOk) return NULL;
    const char* docIDC = docIDStr ? (*env)->GetStringUTFChars(env, docIDStr, NULL) : NULL;
    const char* filterC = filterStr ? (*env)->GetStringUTFChars(env, filterStr, NULL) : NULL;
    const char* updaterC = updaterStr ? (*env)->GetStringUTFChars(env, updaterStr, NULL) : NULL;
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
    int optsOk = 1;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj, &optsOk);
    if (!optsOk) return NULL;
    const char* docIDC = docIDStr ? (*env)->GetStringUTFChars(env, docIDStr, NULL) : NULL;
    const char* filterC = filterStr ? (*env)->GetStringUTFChars(env, filterStr, NULL) : NULL;
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
    int optsOk = 1;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj, &optsOk);
    if (!optsOk) return NULL;
    const char* indexNameC = indexNameStr ? (*env)->GetStringUTFChars(env, indexNameStr, NULL) : NULL;
    const char* fieldsC = fieldsStr ? (*env)->GetStringUTFChars(env, fieldsStr, NULL) : NULL;
    int isUniqueC = (isUnique == JNI_TRUE) ? 1 : 0;
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
    int optsOk = 1;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj, &optsOk);
    if (!optsOk) return NULL;
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
    int optsOk = 1;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj, &optsOk);
    if (!optsOk) return NULL;
    const char* indexNameC = indexNameStr ? (*env)->GetStringUTFChars(env, indexNameStr, NULL) : NULL;
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
    int optsOk = 1;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj, &optsOk);
    if (!optsOk) return NULL;
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
    int optsOk = 1;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj, &optsOk);
    if (!optsOk) return NULL;
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
    int optsOk = 1;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj, &optsOk);
    if (!optsOk) return NULL;
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
    int optsOk = 1;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj, &optsOk);
    if (!optsOk) return NULL;
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
    int optsOk = 1;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj, &optsOk);
    if (!optsOk) return NULL;
    const char* jsonC = jsonStr ? (*env)->GetStringUTFChars(env, jsonStr, NULL) : NULL;
    const char* encryptedFieldsC = encryptedFieldsStr ? (*env)->GetStringUTFChars(env, encryptedFieldsStr, NULL) : NULL;
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
    int optsOk = 1;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj, &optsOk);
    if (!optsOk) return NULL;
    const char* docIDC = docIDStr ? (*env)->GetStringUTFChars(env, docIDStr, NULL) : NULL;
    const char* filterC = filterStr ? (*env)->GetStringUTFChars(env, filterStr, NULL) : NULL;
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
    int optsOk = 1;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj, &optsOk);
    if (!optsOk) return NULL;
    const char* docIDC = docIDStr ? (*env)->GetStringUTFChars(env, docIDStr, NULL) : NULL;
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
    int optsOk = 1;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj, &optsOk);
    if (!optsOk) return NULL;
    const char* docIDC = docIDStr ? (*env)->GetStringUTFChars(env, docIDStr, NULL) : NULL;
    const char* filterC = filterStr ? (*env)->GetStringUTFChars(env, filterStr, NULL) : NULL;
    const char* updaterC = updaterStr ? (*env)->GetStringUTFChars(env, updaterStr, NULL) : NULL;
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
    int optsOk = 1;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj, &optsOk);
    if (!optsOk) return NULL;
    const char* indexNameC = indexNameStr ? (*env)->GetStringUTFChars(env, indexNameStr, NULL) : NULL;
    const char* fieldsC = fieldsStr ? (*env)->GetStringUTFChars(env, fieldsStr, NULL) : NULL;
    int isUniqueC = (isUnique == JNI_TRUE) ? 1 : 0;
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
    int optsOk = 1;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj, &optsOk);
    if (!optsOk) return NULL;
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
    int optsOk = 1;
    CollectionOptions opts = convertJavaCollectionOptions(env, optionsObj, &optsOk);
    if (!optsOk) return NULL;
    const char* indexNameC = indexNameStr ? (*env)->GetStringUTFChars(env, indexNameStr, NULL) : NULL;
    Result res = DeleteIndex((uintptr_t)nodePtr, (char*)indexNameC, opts, (uintptr_t)identityPtr);
    if (indexNameStr) (*env)->ReleaseStringUTFChars(env, indexNameStr, indexNameC);
    releaseJavaCollectionOptions(env, optionsObj, opts);
    return returnDefraResult(env, res);
}
