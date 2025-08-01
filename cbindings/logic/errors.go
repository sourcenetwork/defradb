// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cbindings

const (
	// Node
	errClosingNode            string = "error closing node: %v"
	errCreatingStoreDirectory string = "error creating the store directory: %v"
	errCreatingNode           string = "error creating node: %v"
	errUninitializedNode      string = "error: node is not initialized. Call initNode() first"
	errStoppedNode            string = "error stopping node: node is not initialized, or was already stopped"
	errStoppingNode           string = "error stopping node: %v"
	errParsingReplicatorTimes string = "error parsing replicator retry time intervals: %v"
	errNegativeReplicatorTime string = "error: negative time intervals are not allowed for replicator retries"
	errUnreadyStart           string = "Node is still starting (timeout waiting for readiness)"

	// Schema
	errAddingSchema    string = "error adding schema: %v"
	errGettingSchema   string = "error getting schema: %v"
	errPatchingSchema  string = "error patching schema: %v"
	errSetActiveSchema string = "error setting active version of schema: %v"
	errEmptyPatch      string = "patch cannot be empty"

	// Collection
	errGettingCollection    string = "error getting collection: %v"
	errAmbiguousCollection  string = "error: more than one collection matches the given criteria, could not set context"
	errNoMatchingCollection string = "error: no collection matches the given criteria, could not set context"
	errNoDocIDOrFilter      string = "error: performing the operation requires a DocID or filter"

	// Index
	errInvalidAscensionOrder        string = "invalid ascension order: expected ASC or DESC"
	errInvalidIndexFieldDescription string = "invalid or malformed field descriptiona"

	// Subscription
	errInvalidSubscriptionID string = "error: invalid subscription ID"
	errGEttingSubscription   string = "error: could not retrieve subscription"

	// Txn
	errCreatingTxn     string = "error creating transaction: %v"
	errTxnDoesNotExist string = "error: transaction with ID %v does not exist"

	// Generic
	errInvalidLensConfig string = "invalid lens configuration: %v"
	errMarshallingJSON   string = "error marshalling JSON: %v"
	errInvalidKeyType    string = "invalid key type: %v"
)
