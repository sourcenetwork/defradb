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

#ifndef DEFRA_JVMBOOT_H
#define DEFRA_JVMBOOT_H

#include <jni.h>

int defra_start_jvm(
    const char* jvmLibPath,
    const char* classpath,
    const char* extraOpts,
    JavaVM** outVM,
    JNIEnv** outEnv,
    char* errbuf,
    int errbufLen
);

#endif // DEFRA_JVMBOOT_H
