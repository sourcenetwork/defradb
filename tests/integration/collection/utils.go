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
	"fmt"
	"strings"
	"testing"

	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/stretchr/testify/assert"
)

type TestCase struct {
	// docs is a map from Collection name, to a list
	// of docs in stringified JSON format
	Docs map[string][]string

	CollectionCalls map[string][]func(client.Collection) error

	// Any error expected to be returned by collection calls.
	ExpectedError string
}

type dbInfo interface {
	DB() client.DB
}

func ExecuteQueryTestCase(
	t *testing.T,
	schema string,
	testCase TestCase,
) {
	var err error
	ctx := context.Background()

	var dbi dbInfo
	dbi, err = testUtils.NewBadgerMemoryDB(ctx)
	if err != nil {
		t.Fatal(err)
	}

	db := dbi.DB()

	err = db.AddSchema(ctx, schema)
	if assertError(t, err, testCase.ExpectedError) {
		return
	}

	setupDatabase(ctx, t, db, testCase)

	for collectionName, collectionCallSet := range testCase.CollectionCalls {
		col, err := db.GetCollectionByName(ctx, collectionName)
		if assertError(t, err, testCase.ExpectedError) {
			return
		}

		for _, collectionCall := range collectionCallSet {
			err := collectionCall(col)
			if assertError(t, err, testCase.ExpectedError) {
				return
			}
		}
	}

	if testCase.ExpectedError != "" {
		assert.Fail(t, "Expected an error however none was raised.")
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
		if assertError(t, err, testCase.ExpectedError) {
			return
		}

		for _, docStr := range docs {
			doc, err := client.NewDocFromJSON([]byte(docStr))
			if assertError(t, err, testCase.ExpectedError) {
				return
			}
			err = col.Save(ctx, doc)
			if assertError(t, err, testCase.ExpectedError) {
				return
			}
		}
	}
}

func assertError(t *testing.T, err error, expectedError string) bool {
	if err == nil {
		return false
	}

	if expectedError == "" {
		assert.NoError(t, err)
		return false
	} else {
		if !strings.Contains(err.Error(), expectedError) {
			assert.ErrorIs(t, err, fmt.Errorf(expectedError))
			return false
		}
		return true
	}
}
