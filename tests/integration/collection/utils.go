// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package collection

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

type TestCase struct {
	Description string

	// docs is a map from Collection name, to a list
	// of docs in stringified JSON format
	Docs map[string][]string

	CollectionCalls map[string][]func(client.Collection) error

	// Any error expected to be returned by collection calls.
	ExpectedError string
}

func ExecuteRequestTestCase(
	t *testing.T,
	schema string,
	testCase TestCase,
) {
	ctx := context.Background()

	db, err := testUtils.NewBadgerMemoryDB(ctx)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.AddSchema(ctx, schema)
	if assertError(t, testCase.Description, err, testCase.ExpectedError) {
		return
	}

	setupDatabase(ctx, t, db, testCase)

	for collectionName, collectionCallSet := range testCase.CollectionCalls {
		col, err := db.GetCollectionByName(ctx, collectionName)
		if assertError(t, testCase.Description, err, testCase.ExpectedError) {
			return
		}

		for _, collectionCall := range collectionCallSet {
			err := collectionCall(col)
			if assertError(t, testCase.Description, err, testCase.ExpectedError) {
				return
			}
		}
	}

	if testCase.ExpectedError != "" {
		assert.Fail(t, "Expected an error however none was raised.", testCase.Description)
	}
}

func setupDatabase(
	ctx context.Context,
	t *testing.T,
	db client.DB,
	testCase TestCase,
) {
	for collectionName, docs := range testCase.Docs {
		col, err := db.GetCollectionByName(ctx, collectionName)
		if assertError(t, testCase.Description, err, testCase.ExpectedError) {
			return
		}

		for _, docStr := range docs {
			doc, err := client.NewDocFromJSON([]byte(docStr), col.Schema())
			if assertError(t, testCase.Description, err, testCase.ExpectedError) {
				return
			}
			err = col.Save(ctx, doc)
			if assertError(t, testCase.Description, err, testCase.ExpectedError) {
				return
			}
		}
	}
}

func assertError(t *testing.T, description string, err error, expectedError string) bool {
	if err == nil {
		return false
	}

	if expectedError == "" {
		assert.NoError(t, err, description)
		return false
	} else {
		if !strings.Contains(err.Error(), expectedError) {
			assert.ErrorIs(t, err, errors.New(expectedError))
			return false
		}
		return true
	}
}
