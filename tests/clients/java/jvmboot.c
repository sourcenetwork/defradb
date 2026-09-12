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

#include "jvmboot.h"
#include "defra_errbuf.h"
#include "errors.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// This file contains the implementation of the function to start the JVM on the C side

#ifdef _WIN32
#include <windows.h>
#include <wchar.h>
typedef HMODULE defra_lib_t;
#else
#include <dlfcn.h>
#include <signal.h>
typedef void* defra_lib_t;
#endif

typedef jint (JNICALL *CreateJavaVMFunc)(JavaVM**, void**, void*);

static defra_lib_t defra_dlopen(const char* path) {
#ifdef _WIN32
    // LoadLibraryA decodes its argument using the process's active ANSI code page,
    // not UTF-8. The path built on the Go side (jvm.go) is a Go string, which is
    // always UTF-8, so a JAVA_HOME containing non-ASCII characters (a non-English
    // username, for example) would be misdecoded and fail to load. Transcode to
    // UTF-16 and use the wide-character loader instead.
    int wlen = MultiByteToWideChar(CP_UTF8, 0, path, -1, NULL, 0);
    if (wlen <= 0) {
        return NULL;
    }
    wchar_t* wpath = (wchar_t*)malloc((size_t)wlen * sizeof(wchar_t));
    if (wpath == NULL) {
        return NULL;
    }
    MultiByteToWideChar(CP_UTF8, 0, path, -1, wpath, wlen);

    // jvm.dll depends on other DLLs (jli.dll and friends) that live in JAVA_HOME's
    // bin directory, one level up from jvm.dll's own "server" subdirectory (true
    // for both the JDK 9+ layout and the legacy JDK 8 jre/ layout - see
    // jvmLibraryPath in jvm.go). The standard DLL search order does not include
    // that directory unless it's already on PATH, so add it explicitly rather
    // than relying on the caller's PATH to happen to contain it.
    wchar_t* binDir = (wchar_t*)malloc((wcslen(wpath) + 1) * sizeof(wchar_t));
    if (binDir != NULL) {
        wcscpy(binDir, wpath);
        wchar_t* slash = wcsrchr(binDir, L'\\'); // strips "\jvm.dll"
        if (slash != NULL) {
            *slash = L'\0';
            slash = wcsrchr(binDir, L'\\'); // strips "\server"
            if (slash != NULL) {
                *slash = L'\0';
                SetDllDirectoryW(binDir);
            }
        }
    }

    HMODULE lib = LoadLibraryExW(wpath, NULL, 0);

    // Restore the default search order so this doesn't leak into any other
    // LoadLibrary call made later in the process.
    SetDllDirectoryW(NULL);
    free(binDir);
    free(wpath);
    return lib;
#else
    return dlopen(path, RTLD_NOW | RTLD_GLOBAL);
#endif
}

static void* defra_dlsym(defra_lib_t lib, const char* name) {
#ifdef _WIN32
    return (void*)GetProcAddress(lib, name);
#else
    return dlsym(lib, name);
#endif
}

// DEFRA_MAX_EXTRA_OPTS is the number of extra Java JVM options that can be provided.
// This is an arbitrary number, but it is a rather generous one.
#define DEFRA_MAX_EXTRA_OPTS 32

// defrra_start_jvm starts the JVM, returning 0 if successful, or 1 if unsucessful.
// If it succeeds, outVM and outEnv will be populated, and if it fails, errrbuf will be populated.
int defra_start_jvm(
    const char* jvmLibPath,
    const char* classpath,
    const char* extraOpts,
    JavaVM** outVM,
    JNIEnv** outEnv,
    char* errbuf,
    int errbufLen
) {
    // Firrst, load the JVM library, and make sure it did, in fact, load
    defra_lib_t lib = defra_dlopen(jvmLibPath);
    if (lib == NULL) {
        defra_set_err(errbuf, errbufLen, ERR_FAILED_LOAD_JVM_LIB);
        return 1;
    }

    // Then, as a sanity check, make suure that it contains has the symbol needed to create the Java VM
    // (If this fails, the wrong library is likely being loaded).
    CreateJavaVMFunc createVM = (CreateJavaVMFunc)defra_dlsym(lib, "JNI_CreateJavaVM");
    if (createVM == NULL) {
        defra_set_err(errbuf, errbufLen, ERR_JNI_CREATEJAVAVM_SYMBOL_NOT_FOUND);
        return 1;
    }

    // We will allocate enough space to set some Java JVM Options
    JavaVMOption options[DEFRA_MAX_EXTRA_OPTS + 3];

    // The first Java JVM Option that we have to set, is the one that points at the path for our Defra Jar
    size_t cpOptLen = strlen("-Djava.class.path=") + strlen(classpath) + 1;
    char* cpOpt = (char*)malloc(cpOptLen);
    if (cpOpt == NULL) {
        defra_set_err(errbuf, errbufLen, ERR_OUT_OF_MEMORY_CLASSPATH);
        return 1;
    }
    snprintf(cpOpt, cpOptLen, "-Djava.class.path=%s", classpath);
    options[0].optionString = cpOpt;

    // Next, we pass the option: "-Xrs"
    // This reduce the JVM's installation of OS signal handlers. The JVM normally installs its own SIGSEGV/SIGBUS/etc
    // handlers. Since this process already has an active Go runtime with its own signal handling, minimizing overlap
    // reduces the risk of the two runtimes fighting over the same signals.
    options[1].optionString = (char*)"-Xrs";

    // Next, we pass the option: "-Ddefra.native.external=true"
    // This tells the NativeLoader (in the Java source code) to skip calling System.load. This load never has to
    // occur (and should not occur), because native functions are already being dirrectly registered against the
    // implementations already available to the test binary's
    options[2].optionString = (char*)"-Ddefra.native.external=true";
    int nOptions = 3;

    // extraOpts (DEFRA_JAVA_JVM_OPTS) lets a caller append arbitrary raw
    // JVM options, e.g. "-Xcheck:jni" to turn on extra JNI-usage validation
    // when chasing down a crash. Split on whitespace into a scratch buffer.
    char* scratch = NULL;
    if (extraOpts != NULL && extraOpts[0] != '\0') {
        scratch = strdup(extraOpts);
        if (scratch != NULL) {
            char* tok = strtok(scratch, " \t");
            while (tok != NULL && nOptions < DEFRA_MAX_EXTRA_OPTS + 3) {
                options[nOptions].optionString = tok;
                nOptions++;
                tok = strtok(NULL, " \t");
            }
        }
    }

    // Assemble the JavaVMInitArgs block with the options needed to crreate the VM
    JavaVMInitArgs vmArgs;
    vmArgs.version = JNI_VERSION_1_8;
    vmArgs.options = options;
    vmArgs.nOptions = nOptions;
    vmArgs.ignoreUnrecognized = JNI_TRUE;

#ifndef _WIN32
    // Capture Go's own SIGPIPE handler before JNI_CreateJavaVM runs, so it can be
    // restored afterward instead of being replaced outright.
    struct sigaction preJvmSigpipe;
    sigaction(SIGPIPE, NULL, &preJvmSigpipe);
#endif

    // Star by nulling the JavaVM and JNIEnv values, then try to create the JVM
    JavaVM* vm = NULL;
    JNIEnv* env = NULL;
    jint retcode = createVM(&vm, (void**)&env, &vmArgs);

    // Free allocated memory before returning
    free(cpOpt);
    free(scratch);
    if (retcode != JNI_OK) {
        char msg[128];
        snprintf(msg, sizeof(msg), ERR_FMT_CREATE_JAVAVM_FAILED, (int)retcode);
        defra_set_err(errbuf, errbufLen, msg);
        return 1;
    }

#ifndef _WIN32
    // The JVM installs its own SIGPIPE handler without SA_ONSTACK, which can crash
    // the process (not just print the usual deprecation warning) when a real
    // SIGPIPE arrives. Restore the exact handler Go's runtime had installed before
    // JNI_CreateJavaVM ran.
    sigaction(SIGPIPE, &preJvmSigpipe, NULL);
#endif

    *outVM = vm;
    *outEnv = env;
    return 0;
}
