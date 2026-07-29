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

//go:build javaclient

// Package java implements clients.Client by driving DefraDB through the Defra Java SDK's JNI bindings (a separate,
// sibling repo. See: https://github.com/sourcenetwork/defradb-java-sdk)
//
// This package embeds a JVM directly inside the Go test process via the JNI Invocation API, rather than shelling
// out to a subprocess or loading the Defra Java SDK's own shared object. Doing either of those would start a
// second, independent Go runtime in a second process or a second dlopen'd copy of this same Go runtime, which
// would not work.
//
// Building this package (with -tags javaclient) requires:
//
//   - CGO_ENABLED=1 and a C compiler.
//
//   - CGO_CFLAGS set to include the JDK's JNI headers. For example,
//     CGO_CFLAGS="-I$JAVA_HOME/include -I$JAVA_HOME/include/linux"
//
//     CGO cannot expand $JAVA_HOME itself inside a #cgo directive, so this must be set by whoever invokes the test.
//
// Running tests against this client additionally requires, at runtime:
//
//   - JAVA_HOME set to a JDK 8+ installation
//
//   - DEFRA_CLIENT_JAVA=true
//
//   - A built defradb.jar from the defradb-java-sdk repo.
//
//     To obtain this jar on Linux: Run `make build-java-client` to clone defradb-java-sdk into .javaclient/ and
//     build one from the current working tree. DEFRA_JAVA_JAR will then defaults to that exact output path
//     and doesn't need to be set explicitly. Set DEFRA_JAVA_JAR to override with a jar built manually.
//
// Note, at JVM startup you will likely see the following:
//
//	OpenJDK VM warning: the use of signal() and sigset() for signal chaining
//	was deprecated in version 16.0 and will be removed in a future release.
//
// This is harmless, but is not something that is easily fixable from this side. It is printed by HotSpot's
// own internal signal-chaining detection. HotSpot is the JVM implementation used by Oracle's JDK (and every
// other OpenJDK distribution) - it's open source, just not part of this repo or defradb-java-sdk, so these
// warnings are not coming from our code. Suppression is potentially possible, but
// LD_PRELOAD=$JAVA_HOME/lib/libjsig.so does not seem to work. Theory: it doesn't work because of the way Go's
// runtime installs its signal handlers(?)
package java
