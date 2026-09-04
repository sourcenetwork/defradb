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


#ifndef DEFRA_JNICALL_H
#define DEFRA_JNICALL_H

#include <jni.h>

// DEFRA_ERRBUF_LEN is the standard size for the errbuf/errbufLen buffers used throughout this package
#define DEFRA_ERRBUF_LEN 1024

// DEFRA_MSG_LEN gives a little wiggle room to wrap an error that is contained in the above error buffer
#define DEFRA_MSG_LEN (DEFRA_ERRBUF_LEN + 256)

// DEFRA_MAX_ARGS is the maximum number of arguments allowed for a function being invoked.
// The number is arbitrarily larger than the maximum we currrently use (~6), but leaves room for the addition of
// functions that take more.
#define DEFRA_MAX_ARGS 16

// Enum represents the different kinds of data a DefraArg can be
// This is needed because the struct itself contains a member holding the type, and has enough fields to hold any data that
// it might need dependent on its kind.
typedef enum {
    DEFRA_ARG_LONG = 0,
    DEFRA_ARG_STRING = 1,
    DEFRA_ARG_BOOL = 2,
    DEFRA_ARG_INT = 3,
    DEFRA_ARG_COLLECTION_OPTIONS = 4
} DefraArgKind;

// The aforementioned struct has all fields it could possibly need, with its type stored in the kind member
// This struct is used in place of a union so that it plays more nicely with Go and CGO
typedef struct {
    int kind;                    // DefraArgKind
    jlong j;                     // Long value if kind = DEFRA_ARG_LONG
    int i;                       // Bool value (0 or 1) if kind = DEFRA_ARG_BOOL
                                 // Integer valuue if kind = DEFRA_ARG_INT
    const char* str;             // String value if kind = DEFRA_ARG_STRING
                                 // String holding DefraCollectionOptions.name if kind = DEFRA_ARG_COLLECTION_OPTIONS
    const char* coVersion;       // String holding DefraCollectionOptions.version if kind = DEFRA_ARG_COLLECTION_OPTIONS
    const char* coCollectionID;  // String holding DefrraCollectionOptions.collectionID if kind = DEFRA_ARG_COLLECTION_OPTIONS
    int coGetInactive;           // Bool value (0 or 1) holding DefraCollectionOptions.getInactive if kind = DEFRRA_ARG_COLLECTION_OPTIONS
    int coEnableSigning;         // Tri-state (0 unset, 1 true, -1 false) holding DefraCollectionOptions.enableSigning if kind = DEFRA_ARG_COLLECTION_OPTIONS
} DefraArg;

// Thread lifecycle function declarrations
int defra_attach_thread(JavaVM* vm, JNIEnv** outEnv, char* errbuf, int errbufLen);
void defra_detach_thread(JavaVM* vm);

// Class / method / field lookup function declarations

jclass defra_find_global_class(JNIEnv* env, const char* name, char* errbuf, int errbufLen);
jmethodID defra_get_method_id(JNIEnv* env, jclass cls, const char* name, const char* sig, char* errbuf, int errbufLen);
jmethodID defra_get_static_method_id(JNIEnv* env, jclass cls, const char* name, const char* sig, char* errbuf, int errbufLen);
jfieldID defra_get_field_id(JNIEnv* env, jclass cls, const char* name, const char* sig, char* errbuf, int errbufLen);

// Result field getter function declarations

int defra_get_int_field(JNIEnv* env, jobject obj, jfieldID fid);
jlong defra_get_long_field(JNIEnv* env, jobject obj, jfieldID fid);
char* defra_get_string_field_copy(JNIEnv* env, jobject obj, jfieldID fid);

// Misc. function declarations

int defra_register_natives(JNIEnv* env, jclass cls, JNINativeMethod* methods, int nMethods, char* errbuf, int errbufLen);
jobject defra_new_object(JNIEnv* env, jclass cls, jmethodID ctor, DefraArg* args, int nargs, char* errbuf, int errbufLen);
jobject defra_call_object_method(JNIEnv* env, jobject obj, jmethodID m, DefraArg* args, int nargs, char* errbuf, int errbufLen);
void defra_call_void_method(JNIEnv* env, jobject obj, jmethodID m, DefraArg* args, int nargs, char* errbuf, int errbufLen);
void defra_delete_global_ref(JNIEnv* env, jobject obj);

#endif // DEFRA_JNICALL_H
