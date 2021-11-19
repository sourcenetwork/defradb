// Copyright 2020 Source Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
package tests_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/db"
	"github.com/sourcenetwork/defradb/document"
	"github.com/stretchr/testify/assert"
)

type QueryTestCase struct {
	Description string
	Query       string
	// docs is a map from Collection Index, to a list
	// of docs in stringified JSON format
	Docs map[int][]string
	// updates is a map from document index, to a list
	// of changes in strinigied JSON format
	Updates map[int][]string
	Results []map[string]interface{}
}

func NewMemoryDB() (*db.DB, error) {
	opts := &db.Options{
		Store: "memory",
		Memory: db.MemoryOptions{
			Size: 1024 * 1000,
		},
		Badger: db.BadgerOptions{
			Path: "test",
		},
	}

	return db.NewDB(opts)
}

func ExecuteQueryTestCase(t *testing.T, schema string, collectionNames []string, test QueryTestCase) {
	ctx := context.Background()
	db, err := NewMemoryDB()
	assert.NoError(t, err)

	err = db.AddSchema(ctx, schema)
	assert.NoError(t, err)

	collections := []client.Collection{}
	for _, collectionName := range collectionNames {
		col, err := db.GetCollection(ctx, collectionName)
		assert.NoError(t, err)
		collections = append(collections, col)
	}

	// insert docs
	for cid, docs := range test.Docs {
		for i, docStr := range docs {
			doc, err := document.NewFromJSON([]byte(docStr))
			assert.NoError(t, err, test.Description)
			err = collections[cid].Save(ctx, doc)
			assert.NoError(t, err, test.Description)

			// check for updates
			updates, ok := test.Updates[i]
			if ok {
				for _, u := range updates {
					err = doc.SetWithJSON([]byte(u))
					assert.NoError(t, err, test.Description)
					err = collections[cid].Save(ctx, doc)
					assert.NoError(t, err, test.Description)
				}
			}
		}
	}

	// exec query
	result := db.ExecQuery(ctx, test.Query)
	assert.Empty(t, result.Errors, test.Description)

	resultantData := result.Data.([]map[string]interface{})

	fmt.Println(test.Description)
	fmt.Println(result.Data)
	fmt.Println("--------------")
	fmt.Println("")

	// compare results
	assert.Equal(t, len(test.Results), len(resultantData), test.Description)
	for i, result := range resultantData {
		assert.Equal(t, test.Results[i], result, test.Description)
	}
}
