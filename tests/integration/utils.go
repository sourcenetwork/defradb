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

import (
	"testing"
)

// Represents a request assigned to a particular transaction.
type TransactionRequest struct {
	// Used to identify the transaction for this to run against (allows multiple
	//  requtests to share a single transaction)
	TransactionId int
	// The request to run against the transaction
	Request string
	// The expected (data) results of the issued request
	Results []map[string]any
	// The expected error resulting from the issued request. Also checked against the txn commit.
	ExpectedError string
}

type RequestTestCase struct {
	Description string
	Request     string

	// A collection of requests that are tied to a specific transaction.
	// These will be executed before `Request` (if specified), in the order that they are listed here.
	TransactionalRequests []TransactionRequest

	// docs is a map from Collection Index, to a list
	// of docs in stringified JSON format
	Docs map[int][]string

	// updates is a map from document index, to a list
	// of changes in strinigied JSON format
	Updates map[int]map[int][]string

	Results []map[string]any

	// The expected content of an expected error
	ExpectedError string
}

func ExecuteRequestTestCase(
	t *testing.T,
	schema string,
	collectionNames []string,
	test RequestTestCase,
) {
	actions := []any{
		SchemaUpdate{
			Schema: schema,
		},
	}

	for collectionIndex, docs := range test.Docs {
		for _, doc := range docs {
			actions = append(
				actions,
				CreateDoc{
					CollectionID: collectionIndex,
					Doc:          doc,
				},
			)
		}
	}

	for collectionIndex, docUpdates := range test.Updates {
		for docIndex, docs := range docUpdates {
			for _, doc := range docs {
				actions = append(
					actions,
					UpdateDoc{
						CollectionID: collectionIndex,
						DocID:        docIndex,
						Doc:          doc,
					},
				)
			}
		}
	}

	for _, request := range test.TransactionalRequests {
		actions = append(
			actions,
			TransactionRequest2{
				TransactionID: request.TransactionId,
				Request:       request.Request,
				Results:       request.Results,
				ExpectedError: request.ExpectedError,
			},
		)
	}

	// The old test framework commited all the transactions at the end
	// so we can just lump these here, they must however be commited in
	// the order in which they were first recieved.
	txnIndexesCommited := map[int]struct{}{}
	for _, request := range test.TransactionalRequests {
		if _, alreadyCommited := txnIndexesCommited[request.TransactionId]; alreadyCommited {
			// Only commit each transaction once.
			continue
		}

		txnIndexesCommited[request.TransactionId] = struct{}{}
		actions = append(
			actions,
			TransactionCommit{
				TransactionID: request.TransactionId,
				ExpectedError: request.ExpectedError,
			},
		)
	}

	if test.Request != "" {
		actions = append(
			actions,
			Request{
				ExpectedError: test.ExpectedError,
				Request:       test.Request,
				Results:       test.Results,
			},
		)
	}

	ExecuteTestCase(
		t,
		collectionNames,
		TestCase{
			Description: test.Description,
			Actions:     actions,
		},
	)
}
