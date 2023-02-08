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
	"time"

	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/logging"
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
	isTransactional := len(test.TransactionalRequests) > 0

	if DetectDbChanges && DetectDbChangesPreTestChecks(t, collectionNames, isTransactional) {
		return
	}

	// Must have a non-empty request.
	if !isTransactional && test.Request == "" {
		assert.Fail(t, "Test must have a non-empty request.", test.Description)
	}

	ctx := context.Background()
	dbs, err := GetDatabases(ctx, t)
	if AssertError(t, test.Description, err, test.ExpectedError) {
		return
	}
	require.NotEmpty(t, dbs)

	for _, dbi := range dbs {
		log.Info(ctx, test.Description, logging.NewKV("Database", dbi.name))

		if DetectDbChanges {
			if SetupOnly {
				SetupDatabase(
					ctx,
					t,
					dbi,
					schema,
					collectionNames,
					test.Description,
					test.ExpectedError,
					test.Docs,
					immutable.Some(test.Updates),
				)
				dbi.db.Close(ctx)
				return
			}

			dbi = SetupDatabaseUsingTargetBranch(ctx, t, dbi, collectionNames)
		} else {
			SetupDatabase(
				ctx,
				t,
				dbi,
				schema,
				collectionNames,
				test.Description,
				test.ExpectedError,
				test.Docs,
				immutable.Some(test.Updates),
			)
		}

		// Create the transactions before executing the requests.
		transactions := make([]datastore.Txn, 0, len(test.TransactionalRequests))
		erroredRequests := make([]bool, len(test.TransactionalRequests))
		for i, tq := range test.TransactionalRequests {
			if len(transactions) < tq.TransactionId {
				continue
			}

			txn, err := dbi.db.NewTxn(ctx, false)
			if err != nil {
				if AssertError(t, test.Description, err, tq.ExpectedError) {
					erroredRequests[i] = true
				}
			}
			defer txn.Discard(ctx)
			if len(transactions) <= tq.TransactionId {
				transactions = transactions[:tq.TransactionId+1]
			}
			transactions[tq.TransactionId] = txn
		}

		for i, tq := range test.TransactionalRequests {
			if erroredRequests[i] {
				continue
			}
			result := dbi.db.ExecTransactionalRequest(ctx, tq.Request, transactions[tq.TransactionId])
			if assertRequestResults(ctx, t, test.Description, &result.GQL, tq.Results, tq.ExpectedError) {
				erroredRequests[i] = true
			}
		}

		txnIndexesCommited := map[int]struct{}{}
		for i, tq := range test.TransactionalRequests {
			if erroredRequests[i] {
				continue
			}
			if _, alreadyCommited := txnIndexesCommited[tq.TransactionId]; alreadyCommited {
				continue
			}
			txnIndexesCommited[tq.TransactionId] = struct{}{}

			err := transactions[tq.TransactionId].Commit(ctx)
			if AssertError(t, test.Description, err, tq.ExpectedError) {
				erroredRequests[i] = true
			}
		}

		for i, tq := range test.TransactionalRequests {
			if tq.ExpectedError != "" && !erroredRequests[i] {
				assert.Fail(t, "Expected an error however none was raised.", test.Description)
			}
		}

		// We run the core request after the explicitly transactional ones to permit tests to actually
		// call the request on the commited result of the transactional requests.
		if !isTransactional || (isTransactional && test.Request != "") {
			result := dbi.db.ExecRequest(ctx, test.Request)
			if result.Pub != nil {
				for _, q := range test.PostSubscriptionRequests {
					dbi.db.ExecRequest(ctx, q.Request)
					data := []map[string]any{}
					errs := []any{}
					if len(q.Results) > 1 {
						for range q.Results {
							select {
							case s := <-result.Pub.Stream():
								sResult, _ := s.(client.GQLResult)
								sData, _ := sResult.Data.([]map[string]any)
								errs = append(errs, sResult.Errors...)
								data = append(data, sData...)
							// a safety in case the stream hangs.
							case <-time.After(subscriptionTimeout):
								assert.Fail(t, "timeout occured while waiting for data stream", test.Description)
							}
						}
					} else {
						select {
						case s := <-result.Pub.Stream():
							sResult, _ := s.(client.GQLResult)
							sData, _ := sResult.Data.([]map[string]any)
							errs = append(errs, sResult.Errors...)
							data = append(data, sData...)
						// a safety in case the stream hangs or no results are expected.
						case <-time.After(subscriptionTimeout):
							if q.ExpectedTimout {
								continue
							}
							assert.Fail(t, "timeout occured while waiting for data stream", test.Description)
						}
					}
					gqlResult := &client.GQLResult{
						Data:   data,
						Errors: errs,
					}
					if assertRequestResults(
						ctx,
						t,
						test.Description,
						gqlResult,
						q.Results,
						q.ExpectedError,
					) {
						continue
					}
				}
				result.Pub.Unsubscribe()
			} else {
				if assertRequestResults(
					ctx,
					t,
					test.Description,
					&result.GQL,
					test.Results,
					test.ExpectedError,
				) {
					continue
				}

				if test.ExpectedError != "" {
					assert.Fail(t, "Expected an error however none was raised.", test.Description)
				}
			}
		}

		dbi.db.Close(ctx)
	}
}

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
