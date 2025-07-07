// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build cgo
// +build cgo

package main

const (
	// Node
	cerrClosingNode            string = "error closing node: %v"
	cerrCreatingStoreDirectory string = "error creating the store directory: %v"
	cerrCreatingNode           string = "error creating node: %v"
	cerrUninitializedNode      string = "error: node is not initialized. Call initNode() first"
	cerrStoppedNode            string = "error stopping node: node is not initialized, or was already stopped"
	cerrStoppingNode           string = "error stopping node: %v"
	cerrParsingReplicatorTimes string = "error parsing replicator retry time intervals: %v"
	cerrNegativeReplicatorTime string = "error: negative time intervals are not allowed for replicator retries"

	// Schema
	cerrAddingSchema    string = "error adding schema: %v"
	cerrGettingSchema   string = "error getting schema: %v"
	cerrPatchingSchema  string = "error patching schema: %v"
	cerrSetActiveSchema string = "error setting active version of schema: %v"
	cerrEmptyPatch      string = "patch cannot be empty"

	// Collection
	cerrGettingCollection    string = "error getting collection: %v"
	cerrAmbiguousCollection  string = "error: more than one collection matches the given criteria, could not set context"
	cerrNoMatchingCollection string = "error: no collection matches the given criteria, could not set context"
	cerrNoDocIDOrFilter      string = "error: performing the operation requires a DocID or filter"

	// Index
	cerrInvalidAscensionOrder        string = "invalid ascension order: expected ASC or DESC"
	cerrInvalidIndexFieldDescription string = "invalid or malformed field descriptiona"

	// Txn
	cerrCreatingTxn     string = "error creating transaction: %v"
	cerrTxnDoesNotExist string = "error: transaction with ID %v does not exist"

	// Generic
	cerrInvalidLensConfig string = "invalid lens configuration: %v"
	cerrMarshallingJSON   string = "error marshalling JSON: %v"
	cerrInvalidKeyType    string = "invalid key type: %v"
)
