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

#ifndef DEFRA_ERRORS_H
#define DEFRA_ERRORS_H

// This header collects the static error message/format strings used by this package's C sources

// jnicall.c
#define ERR_JAVA_EXCEPTION_THROWN "java exception thrown (see stderr for the stack trace)"
#define ERR_FAILED_GLOBAL_REF_CLASS "failed to create global ref for class"
#define ERR_FMT_ATTACH_THREAD_FAILED "AttachCurrentThread failed with code %d"
#define ERR_FMT_CLASS_NOT_FOUND "class not found: %s: %s"
#define ERR_FMT_METHOD_NOT_FOUND "method not found: %s%s: %s"
#define ERR_FMT_STATIC_METHOD_NOT_FOUND "static method not found: %s%s: %s"
#define ERR_FMT_FIELD_NOT_FOUND "field not found: %s%s: %s"
#define ERR_FMT_REGISTER_NATIVES_FAILED "RegisterNatives failed: %s"
#define ERR_COLLECTION_OPTIONS_CLASS_NOT_FOUND "DefraCollectionOptions class not found"
#define ERR_COLLECTION_OPTIONS_CTOR_NOT_FOUND "DefraCollectionOptions constructor not found"
#define ERR_COLLECTION_OPTIONS_FIELDS_NOT_FOUND "DefraCollectionOptions field(s) not found"
#define ERR_UNKNOWN_ARG_KIND "unknown DefraArg kind"
#define ERR_TOO_MANY_CTOR_ARGS "too many constructor args"
#define ERR_FAILED_GLOBAL_REF_OBJECT "failed to create global ref for new object"
#define ERR_TOO_MANY_METHOD_ARGS "too many method args"

// jvmboot.c
#define ERR_FAILED_LOAD_JVM_LIB "failed to load JVM shared library"
#define ERR_JNI_CREATEJAVAVM_SYMBOL_NOT_FOUND "JNI_CreateJavaVM symbol not found in JVM library"
#define ERR_OUT_OF_MEMORY_CLASSPATH "out of memory building classpath option"
#define ERR_FMT_CREATE_JAVAVM_FAILED "JNI_CreateJavaVM failed with code %d"

#endif // DEFRA_ERRORS_H
