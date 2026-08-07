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
#include "defra_errbuf.h"
#include "errors.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// This file contains the generic JNI marshaling layer for the Java test client. It wraps low-level JNI
// operations with helpful error-reporting, and contains the call dispatchers for actually invoking native
// methods. Every native method call made by the client will utilize the functions here to do so.

// check_pending_exception checks whether an exception has occurred, and returns 0 for false or 1 for true
// If a buffer is passed in, the error message (if present) will be copied to it.
static int check_pending_exception(JNIEnv* env, char* buf, int bufLen) {
    // If nothing is pending, then there is nothing to do. Return early.
    jthrowable ex = (*env)->ExceptionOccurred(env);
    if (ex == NULL) {
        return 0;
    }

    // Print the full exception and stack trace to stderr.
    (*env)->ExceptionDescribe(env);

    // Most other JNI calls aren't safe to make while an exception is pending, so we have to clear it.
    (*env)->ExceptionClear(env);

    // If called without a buffer, the caller only wants to know something threw an exception.
    // We don't have to do anything with it though.
    if (buf == NULL || bufLen <= 0) {
        return 1;
    }

    // In Java, every exception is an Object. Therefore, we can look up and invoke its toString() method
    // by using reflection-style JNI calls.
    // First, try to get the object's class...
    jclass objCls = (*env)->FindClass(env, "java/lang/Object");
    if (objCls == NULL) {
        // If the class was not found, we can only point to the stack trace.
        // We still have to clear the exception though.
        (*env)->ExceptionClear(env);
        defra_set_err(buf, bufLen, ERR_JAVA_EXCEPTION_THROWN);
        return 1;
    }
    // ...then try to get its toString method...
    jmethodID toString = (*env)->GetMethodID(env, objCls, "toString", "()Ljava/lang/String;");
    if (toString == NULL) {
        // If the toString() method was not found, we can only point to the stack trace.
        // We still have to clear the exception though.
        (*env)->ExceptionClear(env);
        defra_set_err(buf, bufLen, ERR_JAVA_EXCEPTION_THROWN);
        return 1;
    }
    // ...then try to call its toString() method.
    jstring msg = (jstring)(*env)->CallObjectMethod(env, ex, toString);
    if (msg == NULL || (*env)->ExceptionCheck(env)) {
        // If something inside the object's toString() method threw an exception, we can only point to the stack trace.
        // We still have to clear the exception though.
        (*env)->ExceptionClear(env);
        defra_set_err(buf, bufLen, ERR_JAVA_EXCEPTION_THROWN);
        return 1;
    }

    // Making it this far means we got a real error message, which we can try to copy to the buffer for return.
    // This can technically fail, for example, if the JVM lacks the memory to allocate a string.
    const char* chars = (*env)->GetStringUTFChars(env, msg, NULL);
    if (chars != NULL) {
        // If it did NOT fail, then we copy it into the buffer, and then release it.
        snprintf(buf, (size_t)bufLen, "%s", chars);
        (*env)->ReleaseStringUTFChars(env, msg, chars);
    } else {
        // If it did fail, we will have to clear the exception.
        (*env)->ExceptionClear(env);
        defra_set_err(buf, bufLen, ERR_JAVA_EXCEPTION_THROWN);
    }

    // Finally, we return 1 because an exception was thrown.
    return 1;
}

// defra_attach_thread attaches the current thread to the JVM, which is necessary to allow us to make calls to any JNI
// methods using it. For more information, see the attach() function inside jvm.go, from which this function is called.
// This returns 0 if it was successful, and 1 if it was unsuccessful, printing any error into the passed-in buffer.
int defra_attach_thread(JavaVM* vm, JNIEnv** outEnv, char* errbuf, int errbufLen) {
    jint retcode = (*vm)->AttachCurrentThread(vm, (void**)outEnv, NULL);
    if (retcode != JNI_OK) {
        char msg[64];
        snprintf(msg, sizeof(msg), ERR_FMT_ATTACH_THREAD_FAILED, (int)retcode);
        defra_set_err(errbuf, errbufLen, msg);
        return 1;
    }
    return 0;
}

// defra_detatch_thread is the counterparrt to defra_attach_thread, detaching the thread from the JVM.
// All calls to defra_attach_thread should be paired with a call to this function afterwards.
void defra_detach_thread(JavaVM* vm) {
    (*vm)->DetachCurrentThread(vm);
}

// defra_find_global_class looks up a Java class by name and returns a reference to it that stays valid
// across future attach/detach cycles, and across different OS threads.
// See the setupJVM() function of jvm.go, which makes calls to this function, caching the results for
// the DefraNode, DefraResult, and DefraTransactionResult classes.
jclass defra_find_global_class(JNIEnv* env, const char* name, char* errbuf, int errbufLen) {
    // Look up the class. On failure this returns NULL and leaves an exception pending.
    jclass local = (*env)->FindClass(env, name);
    if (local == NULL) {
        // Fold the exception detail into a more specific message before reporting it
        char detail[DEFRA_ERRBUF_LEN];
        detail[0] = '\0'; // Defensive: Initialize it as an empty string, not as garbage.
        check_pending_exception(env, detail, sizeof(detail));
        char msg[DEFRA_MSG_LEN];
        snprintf(msg, sizeof(msg), ERR_FMT_CLASS_NOT_FOUND, name, detail);
        defra_set_err(errbuf, errbufLen, msg);
        return NULL;
    }

    // Promote the local reference to a global one, deleting the local reference afterwards
    jclass global = (jclass)(*env)->NewGlobalRef(env, local);
    (*env)->DeleteLocalRef(env, local);
    if (global == NULL) {
        // NewGlobalRef could theoretically fail due to a lack of available memory.
        // It would throw an exception that would need to be cleared.
        (*env)->ExceptionClear(env);
        defra_set_err(errbuf, errbufLen, ERR_FAILED_GLOBAL_REF_CLASS);
    }
    return global;
}

// defra_get_method_id looks up a jmethodID for a instance method with a given name
jmethodID defra_get_method_id(JNIEnv* env, jclass cls, const char* name, const char* sig, char* errbuf, int errbufLen) {
    jmethodID m = (*env)->GetMethodID(env, cls, name, sig);
    if (m == NULL) {
        // The method with that name was not found, so return an error
        char detail[DEFRA_ERRBUF_LEN];
        detail[0] = '\0'; // Defensive: Initialize it as an empty string, not as garbage.
        check_pending_exception(env, detail, sizeof(detail));
        char msg[DEFRA_MSG_LEN];
        snprintf(msg, sizeof(msg), ERR_FMT_METHOD_NOT_FOUND, name, sig, detail);
        defra_set_err(errbuf, errbufLen, msg);
    }
    return m;
}

// deffra_get_static_method_id looks up a jmethodID for a static method with a given name
jmethodID defra_get_static_method_id(JNIEnv* env, jclass cls, const char* name, const char* sig, char* errbuf, int errbufLen) {
    jmethodID m = (*env)->GetStaticMethodID(env, cls, name, sig);
    if (m == NULL) {
        // The method with that name was not found, so return an error
        char detail[DEFRA_ERRBUF_LEN];
        detail[0] = '\0'; // Defensive: Initialize it as an empty string, not as garbage.
        check_pending_exception(env, detail, sizeof(detail));
        char msg[DEFRA_MSG_LEN];
        snprintf(msg, sizeof(msg), ERR_FMT_STATIC_METHOD_NOT_FOUND, name, sig, detail);
        defra_set_err(errbuf, errbufLen, msg);
    }
    return m;
}

// defra_get_field_id looks up a jfieldID for a class with a given name
jfieldID defra_get_field_id(JNIEnv* env, jclass cls, const char* name, const char* sig, char* errbuf, int errbufLen) {
    jfieldID f = (*env)->GetFieldID(env, cls, name, sig);
    if (f == NULL) {
        // The field was not fouund, so return an error
        char detail[DEFRA_ERRBUF_LEN];
        detail[0] = '\0'; // Defensive: Initialize it as an empty string, not as garbage.
        check_pending_exception(env, detail, sizeof(detail));
        char msg[DEFRA_MSG_LEN];
        snprintf(msg, sizeof(msg), ERR_FMT_FIELD_NOT_FOUND, name, sig, detail);
        defra_set_err(errbuf, errbufLen, msg);
    }
    return f;
}

// defra_register_natives binds a table of native method declarations on cls directly to C function pointers, instead of
// the JVM resolving them by the usual Java_pkg_Class_method symbol-naming convention against a loaded shared library. 
// The significance is that these are the exact same C functions a real shared library would export, already compiled into
// the Go test binary. Binding them directly, by addrerss, gives native methods a working implementation, without the JVM
// ever needing to load a .so to find them by name. This resuults in the SDK staying usable, despite a single Go runtime.
int defra_register_natives(JNIEnv* env, jclass cls, JNINativeMethod* methods, int nMethods, char* errbuf, int errbufLen) {
    jint retcode = (*env)->RegisterNatives(env, cls, methods, nMethods);
    if (retcode != 0) {
        // If this error is hit, there is a mismatch between the name/signature pairr in the table, and what the Java
        // class actually declares. 
        char detail[DEFRA_ERRBUF_LEN];
        detail[0] = '\0'; // Defensive: Initialize it as an empty string, not as garbage.
        check_pending_exception(env, detail, sizeof(detail));
        char msg[DEFRA_MSG_LEN];
        snprintf(msg, sizeof(msg), ERR_FMT_REGISTER_NATIVES_FAILED, detail);
        defra_set_err(errbuf, errbufLen, msg);
        return 1;
    }
    return 0;
}

// checked_field_id looks up a field ID and, on failure, clears the resulting pending exception
// and records a fixed error message instead. Letting the caller chain another JNI call onto that pending exception
// isn't documented-safe, and -Xcheck:jni aborts on it.
static jfieldID checked_field_id(JNIEnv* env, jclass cls, const char* name, const char* sig, char* errbuf, int errbufLen) {
    jfieldID fid = (*env)->GetFieldID(env, cls, name, sig);
    if (fid == NULL) {
        (*env)->ExceptionClear(env);
        defra_set_err(errbuf, errbufLen, ERR_COLLECTION_OPTIONS_FIELDS_NOT_FOUND);
    }
    return fid;
}

// build_collection_options constructs a DefrraCollectionOptions Java object from a DefraArg struct
static jobject build_collection_options(JNIEnv* env, const DefraArg* arg, char* errbuf, int errbufLen) {
    // First, try to get the object's class...
    jclass cls = (*env)->FindClass(env, "source/defra/DefraCollectionOptions");
    if (cls == NULL) {
        // If the class was not found, we can only point to the stack trace.
        // We still have to clear the exception though.
        (*env)->ExceptionClear(env);
        defra_set_err(errbuf, errbufLen, ERR_COLLECTION_OPTIONS_CLASS_NOT_FOUND);
        return NULL;
    }
    // ...then try to get its constructor method...
    jmethodID constructor = (*env)->GetMethodID(env, cls, "<init>", "()V");
    if (constructor == NULL) {
        // The constructor *should* always be found, provided the appropriate Defra Jar is used.
        // But to be safe, we can defensively handle this path, and clear a potential exception.
        (*env)->ExceptionClear(env);
        defra_set_err(errbuf, errbufLen, ERR_COLLECTION_OPTIONS_CTOR_NOT_FOUND);
        return NULL;
    }
    // ...then try to create a new object
    jobject opts = (*env)->NewObject(env, cls, constructor);
    if (opts == NULL) {
        // This could potentially fail in the case of running out of memory.
        check_pending_exception(env, errbuf, errbufLen);
        return NULL;
    }

    // Try to get all of the field IDs from the created object, checking each one before making the
    // next lookup.
    jfieldID fName = checked_field_id(env, cls, "name", "Ljava/lang/String;", errbuf, errbufLen);
    if (fName == NULL) return NULL;
    jfieldID fVersion = checked_field_id(env, cls, "version", "Ljava/lang/String;", errbuf, errbufLen);
    if (fVersion == NULL) return NULL;
    jfieldID fCollectionID = checked_field_id(env, cls, "collectionID", "Ljava/lang/String;", errbuf, errbufLen);
    if (fCollectionID == NULL) return NULL;
    jfieldID fGetInactive = checked_field_id(env, cls, "getInactive", "Z", errbuf, errbufLen);
    if (fGetInactive == NULL) return NULL;
    jfieldID fEnableSigning = checked_field_id(env, cls, "enableSigning", "Ljava/lang/Boolean;", errbuf, errbufLen);
    if (fEnableSigning == NULL) return NULL;

    // Now that we have all the field IDs, we can assign the values and return
    jstring name = arg->str ? (*env)->NewStringUTF(env, arg->str) : (*env)->NewStringUTF(env, "");
    jstring version = arg->coVersion ? (*env)->NewStringUTF(env, arg->coVersion) : (*env)->NewStringUTF(env, "");
    jstring collectionID = arg->coCollectionID ? (*env)->NewStringUTF(env, arg->coCollectionID) : (*env)->NewStringUTF(env, "");
    (*env)->SetObjectField(env, opts, fName, name);
    (*env)->SetObjectField(env, opts, fVersion, version);
    (*env)->SetObjectField(env, opts, fCollectionID, collectionID);
    (*env)->SetBooleanField(env, opts, fGetInactive, arg->coGetInactive ? JNI_TRUE : JNI_FALSE);

    // enableSigning is left as its default null (unset) unless the caller explicitly asked
    // for true/false, boxed via Boolean.valueOf so the field's tri-state survives the call.
    if (arg->coEnableSigning != 0) {
        jclass booleanCls = (*env)->FindClass(env, "java/lang/Boolean");
        if (booleanCls == NULL) {
            check_pending_exception(env, errbuf, errbufLen);
            return NULL;
        }
        jmethodID valueOfMid = (*env)->GetStaticMethodID(env, booleanCls, "valueOf", "(Z)Ljava/lang/Boolean;");
        if (valueOfMid == NULL) {
            check_pending_exception(env, errbuf, errbufLen);
            return NULL;
        }
        jobject boxedEnableSigning = (*env)->CallStaticObjectMethod(
            env, booleanCls, valueOfMid, arg->coEnableSigning > 0 ? JNI_TRUE : JNI_FALSE);
        if (boxedEnableSigning == NULL) {
            check_pending_exception(env, errbuf, errbufLen);
            return NULL;
        }
        (*env)->SetObjectField(env, opts, fEnableSigning, boxedEnableSigning);
    }

    return opts;
}

// build_jvalue_array converts an DefraArg[] array into the jvalue[] arrray that a JNI call needs.
// We can think of DefraArg[] as a representation of a form describing the call, and the jvalue[] array being what JNI
// needs to execute it. This will return 0 on success, and 1 on failure (while filling an error buffer.)
static int build_jvalue_array(JNIEnv* env, DefraArg* args, int nargs, jvalue* argv, char* errbuf, int errbufLen) {
    // Loop through the DefraArg[] array, handling each case type appropriately
    for (int i = 0; i < nargs; i++) {
        DefraArg* a = &args[i];
        switch (a->kind) {
            case DEFRA_ARG_LONG:
                argv[i].j = a->j;
                break;
            case DEFRA_ARG_STRING:
                argv[i].l = a->str ? (*env)->NewStringUTF(env, a->str) : NULL;
                break;
            case DEFRA_ARG_BOOL:
                argv[i].z = a->i ? JNI_TRUE : JNI_FALSE;
                break;
            case DEFRA_ARG_INT:
                argv[i].i = a->i;
                break;
            case DEFRA_ARG_COLLECTION_OPTIONS: {
                jobject opts = build_collection_options(env, a, errbuf, errbufLen);
                if (opts == NULL) {
                    return 1;
                }
                argv[i].l = opts;
                break;
            }
            default:
                // This should never happen if DefraArg is not malformed.
                defra_set_err(errbuf, errbufLen, ERR_UNKNOWN_ARG_KIND);
                return 1;
        }
    }
    return 0;
}

// defra_new_object constructs a new Java object by calling a given constructor. It will return this object on success,
// and on failure it will return NULL and fill an error buffer. The arguments to the constructor, and the number of arguments
// to the constructor must be passed in.
jobject defra_new_object(JNIEnv* env, jclass cls, jmethodID ctor, DefraArg* args, int nargs, char* errbuf, int errbufLen) {
    // First, create an array of jvalue arguments and zero out the memory
    jvalue argv[DEFRA_MAX_ARGS];
    memset(argv, 0, sizeof(argv));

    // Defensive check that the user didn't pass too many arguments to fit in the buffer
    if (nargs > DEFRA_MAX_ARGS) {
        defra_set_err(errbuf, errbufLen, ERR_TOO_MANY_CTOR_ARGS);
        return NULL;
    }

    // Fill the argv buffer with jvalues
    if (build_jvalue_array(env, args, nargs, argv, errbuf, errbufLen) != 0) {
        return NULL;
    }


    // Create a local Java object, and then promote the local reference to a global reference
    jobject local = (*env)->NewObjectA(env, cls, ctor, argv);
    if (local == NULL) {
        check_pending_exception(env, errbuf, errbufLen);
        return NULL;
    }
    jobject global = (*env)->NewGlobalRef(env, local);
    (*env)->DeleteLocalRef(env, local);
    if (global == NULL) {
        defra_set_err(errbuf, errbufLen, ERR_FAILED_GLOBAL_REF_OBJECT);
    }
    
    return global;
}

// defra_call_object_method invokes a method with a given ID on a given object, returning another jobject on success,
// or NULL on failure. If it fails, it will also fill the error buffer.
jobject defra_call_object_method(JNIEnv* env, jobject obj, jmethodID m, DefraArg* args, int nargs, char* errbuf, int errbufLen) {
    // First, create an array of jvalue arguments and zero out the memory
    jvalue argv[DEFRA_MAX_ARGS];
    memset(argv, 0, sizeof(argv));

    // Defensive check that the user didn't pass too many arguments to fit in the buffer
    if (nargs > DEFRA_MAX_ARGS) {
        defra_set_err(errbuf, errbufLen, ERR_TOO_MANY_METHOD_ARGS);
        return NULL;
    }

    // Fill the argv buffer with jvalues
    if (build_jvalue_array(env, args, nargs, argv, errbuf, errbufLen) != 0) {
        return NULL;
    }

    // Try making the call, and returning the result
    jobject result = (*env)->CallObjectMethodA(env, obj, m, argv);
    if (check_pending_exception(env, errbuf, errbufLen)) {
        return NULL;
    }
    return result;
}

// defra_call_object_method invokes a method with a given ID on a given object, but expects no return value.
// On failure, it will fill an error buffer.
void defra_call_void_method(JNIEnv* env, jobject obj, jmethodID m, DefraArg* args, int nargs, char* errbuf, int errbufLen) {
    // First, create an array of jvalue arguments and zero out the memory
    jvalue argv[DEFRA_MAX_ARGS];
    memset(argv, 0, sizeof(argv));

    // Defensive check that the user didn't pass too many arguments to fit in the buffer
    if (nargs > DEFRA_MAX_ARGS) {
        defra_set_err(errbuf, errbufLen, ERR_TOO_MANY_METHOD_ARGS);
        return;
    }

    // Fill the argv buffer with jvalues
    if (build_jvalue_array(env, args, nargs, argv, errbuf, errbufLen) != 0) {
        return;
    }

    // Try making the call, and then check for any exceptions that might have been thrown
    (*env)->CallVoidMethodA(env, obj, m, argv);
    check_pending_exception(env, errbuf, errbufLen);
}

// defra_delete_global_ref deletes a global reference to an object. It should be called once forr
// every point that a global reference is created, becauuse it will not be cleaned up automatically.
//
// Note: This will silently no-op if the object *is* null. However, because this is called
// during the node's Close() method, or a Transaction's Discard() method on the Go side, there
// is no meaningful way to propogate back a helpful error. So it's currently left as a no-op.
// This could be changed to return an error, as the Commit() method *does* return an error. This
// is as it currently is just for simplicity and consistency.
void defra_delete_global_ref(JNIEnv* env, jobject obj) {
    if (obj != NULL) {
        (*env)->DeleteGlobalRef(env, obj);
    }
}

// defra_get_int_field gets an integer from a field with a given ID belonging to an object
int defra_get_int_field(JNIEnv* env, jobject obj, jfieldID fid) {
    return (int)(*env)->GetIntField(env, obj, fid);
}

//defra_get_long_field gets a long from a field with a given ID belonging to an object
jlong defra_get_long_field(JNIEnv* env, jobject obj, jfieldID fid) {
    return (*env)->GetLongField(env, obj, fid);
}

//defra_get_string_field_copy gets a C String from a field with a given ID belonging to an object
// The returned string will be freshly malloc'd heap memory that the caller is responsible for freeing.
char* defra_get_string_field_copy(JNIEnv* env, jobject obj, jfieldID fid) {
    jstring s = (jstring)(*env)->GetObjectField(env, obj, fid);
    if (s == NULL) {
        return NULL;
    }
    const char* chars = (*env)->GetStringUTFChars(env, s, NULL);
    if (chars == NULL) {
        return NULL;
    }
    char* copy = strdup(chars); // Allocates memory
    (*env)->ReleaseStringUTFChars(env, s, chars);
    return copy;
}
