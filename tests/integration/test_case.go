// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package tests

// TestCase contains the details of the test case to execute.
type TestCase struct {
	// Test description, optional.
	Description string

	// Actions contains the set of actions and their expected results that
	// this test should execute.  They will execute in the order that they
	// are provided.
	Actions []any
}

// SetupComplete is a flag to explicitly notify the change detector at which point
// setup is complete so that it may split actions across database code-versions.
//
// If a SetupComplete action is not provided the change detector will split before
// the first item that is neither a SchemaUpdate, CreateDoc or UpdateDoc action.
type SetupComplete struct{}

// SchemaUpdate is an action that will update the database schema.
type SchemaUpdate struct {
	// The schema update.
	Schema string

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

// CreateDoc will attempt to create the given document in the given collection
// using the collection api.
type CreateDoc struct {
	// The collection in which this document should be created.
	CollectionID int

	// The document to create, in JSON string format.
	Doc string

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

// UpdateDoc will attempt to update the given document in the given collection
// using the collection api.
type UpdateDoc struct {
	// The collection in which this document exists.
	CollectionID int

	// The index-identifier of the document within the collection.  This is based on
	// the order in which it was created, not the ordering of the document within the
	// database.
	DocID int

	// The document update, in JSON string format. Will only update the properties
	// provided.
	Doc string

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

// Request represents a standard Defra (GQL) request.
type Request struct {
	// The request to execute.
	Request string

	// The expected (data) results of the issued request.
	Results []map[string]any

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

// TransactionRequest2 represents a transactional request.
//
// A new transaction will be created for the first TransactionRequest2 of any given
// TransactionId. TransactionRequest2s will be submitted to the database in the order
// in which they are recieved (interleaving amongst other actions if provided), however
// they will not be commited until a TransactionCommit of matching TransactionId is
// provided.
type TransactionRequest2 struct {
	// Used to identify the transaction for this to run against.
	TransactionID int

	// The request to run against the transaction.
	Request string

	// The expected (data) results of the issued request.
	Results []map[string]any

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

// TransactionCommit represents a commit request for a transaction of the given id.
type TransactionCommit struct {
	// Used to identify the transaction to commit.
	TransactionID int

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

// SubscriptionRequest represents a subscription request.
//
// The subscription will remain active until shortly after all actions have been processed.
// The results of the subscription will then be asserted upon.
type SubscriptionRequest struct {
	// The subscription request to submit.
	Request string

	// The expected (data) results yielded through the subscription across its lifetime.
	Results []map[string]any

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}
