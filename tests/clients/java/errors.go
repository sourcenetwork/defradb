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

package java

// This file collects every static error message/format string used throughout this package

const (
	// Shared JNI-call plumbing (call.go)
	errFmtNoRegisteredMethod       = "java client: no registered method %q"
	errFmtJNICall                  = "java client: %s: %s"
	errNoRegisteredTxnCreateMethod = "java client: no registered method \"TransactionCreateNative\""
	errFmtTransactionCreateFailed  = "java client: TransactionCreateNative: %s"
	errFmtTransactionCommitFailed  = "java client: TransactionCommitNative: %s"

	// JVM setup and lifecycle (jvm.go)
	errFmtJavaHomeNotSet = "java client: %s must be set to a JDK 8+ installation"
	errFmtJarNotSet      = "java client: %s must be set to a defradb.jar built from defradb-java-sdk " +
		"(or run `make build-java-client` first)"
	errFmtStartJVMFailed                   = "java client: starting embedded JVM (lib=%s): %s"
	errFmtJNIError                         = "java client: %s"
	errFmtRegisterNodeNativesFailed        = "java client: registering DefraNode natives: %s"
	errFmtRegisterTransactionNativesFailed = "java client: registering DefraTransaction natives: %s"
	errFmtUnsupportedGOOS                  = "java client: unsupported GOOS %q"
	errFmtCreateNodeFailed                 = "java client: creating DefraNode: %s"
	errFmtCreateTransactionFailed          = "java client: creating DefraTransaction: %s"

	// Wrapper-level helpers (wrapper.go, wrapper_collection.go)
	errFmtUnmarshalResult      = "failed to unmarshal JSON %q: %w"
	errCastClientTxnFailed     = "failed to cast clientTxn to datastore.Txn"
	errFmtListLensesUnmarshal  = "%w (value len=%d, snippet=%q)"
	errFmtCollectionNotFound   = "collection with name %q not found"
	errFmtDocumentToJSONFailed = "failed to convert document to JSON: %w"
)
