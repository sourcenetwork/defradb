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
	"context"
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
)

// Represents a subscription request.
type SubscriptionRequest struct {
	Request string
	// The expected (data) results of the issued request.
	Results []map[string]any
	// The expected error resulting from the issued request.
	ExpectedError string
	// If set to true, the request should yield no results.
	// The timeout is duration is that of subscriptionTimeout (1 second)
	ExpectedTimout bool
}

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

	// A collection of requests to exucute after the subscriber is listening on the stream
	PostSubscriptionRequests []SubscriptionRequest

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
			TransactionRequest2(request),
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
				TransactionId: request.TransactionId,
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

	for _, request := range test.PostSubscriptionRequests {
		actions = append(
			actions,
			SubscriptionRequest2{
				ExpectedError:   request.ExpectedError,
				Request:         request.Request,
				Results:         request.Results,
				ExpectedTimeout: request.ExpectedTimout,
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

// SetupDatabase is persisted for the sake of the explain tests as they use a different
// test executor that calls this function.
func SetupDatabase(
	ctx context.Context,
	t *testing.T,
	dbi databaseInfo,
	schema string,
	collectionNames []string,
	description string,
	expectedError string,
	documents map[int][]string,
	updates immutable.Option[map[int]map[int][]string],
) {
	db := dbi.db
	err := db.AddSchema(ctx, schema)
	if AssertError(t, description, err, expectedError) {
		return
	}

	collections := []client.Collection{}
	for _, collectionName := range collectionNames {
		col, err := db.GetCollectionByName(ctx, collectionName)
		if AssertError(t, description, err, expectedError) {
			return
		}
		collections = append(collections, col)
	}

	// insert docs
	for collectionIndex, docs := range documents {
		hasCollectionUpdates := false
		collectionUpdates := map[int][]string{}

		if updates.HasValue() {
			collectionUpdates, hasCollectionUpdates = updates.Value()[collectionIndex]
		}

		for documentIndex, docStr := range docs {
			doc, err := client.NewDocFromJSON([]byte(docStr))
			if AssertError(t, description, err, expectedError) {
				return
			}
			err = collections[collectionIndex].Save(ctx, doc)
			if AssertError(t, description, err, expectedError) {
				return
			}

			if hasCollectionUpdates {
				documentUpdates, hasDocumentUpdates := collectionUpdates[documentIndex]

				if hasDocumentUpdates {
					for _, u := range documentUpdates {
						err = doc.SetWithJSON([]byte(u))
						if AssertError(t, description, err, expectedError) {
							return
						}
						err = collections[collectionIndex].Save(ctx, doc)
						if AssertError(t, description, err, expectedError) {
							return
						}
					}
				}
			}
		}
	}
}
