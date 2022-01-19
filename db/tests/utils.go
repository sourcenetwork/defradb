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
	"os"
	"testing"

	badger "github.com/dgraph-io/badger/v3"
	ds "github.com/ipfs/go-datastore"
	"github.com/sourcenetwork/defradb/client"
	badgerds "github.com/sourcenetwork/defradb/datastores/badger/v3"
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

type databaseInfo struct {
	name string
	db   *db.DB
}

var badgerInMemory bool
var mapStore bool

func init() {
	// We use environment variables instead of flags `go test ./...` throws for all packages that don't have the flag defined
	_, badgerInMemory = os.LookupEnv("DEFRA_BADGER_MEMORY")
	_, mapStore = os.LookupEnv("DEFRA_MAP")

	// default is to run against all
	if !badgerInMemory && !mapStore {
		badgerInMemory = true
		mapStore = true
	}
}

func newBadgerMemoryDB() (databaseInfo, error) {
	opts := badgerds.Options{Options: badger.DefaultOptions("").WithInMemory(true)}
	rootstore, err := badgerds.NewDatastore("", &opts)
	if err != nil {
		return databaseInfo{}, err
	}

	db, err := db.NewDB(rootstore, struct{}{})
	if err != nil {
		return databaseInfo{}, err
	}

	return databaseInfo{
		name: "badger-in-memory",
		db:   db,
	}, nil
}

func newMapDB() (databaseInfo, error) {
	rootstore := ds.NewMapDatastore()
	db, err := db.NewDB(rootstore, struct{}{})
	if err != nil {
		return databaseInfo{}, err
	}

	return databaseInfo{
		name: "ipfs-map-datastore",
		db:   db,
	}, nil
}

func getDatabases() ([]databaseInfo, error) {
	databases := []databaseInfo{}

	if badgerInMemory {
		badgerIMDatabase, err := newBadgerMemoryDB()
		if err != nil {
			return nil, err
		}
		databases = append(databases, badgerIMDatabase)
	}

	if mapStore {
		mapDatabase, err := newMapDB()
		if err != nil {
			return nil, err
		}
		databases = append(databases, mapDatabase)
	}

	return databases, nil
}

func ExecuteQueryTestCase(t *testing.T, schema string, collectionNames []string, test QueryTestCase) {
	ctx := context.Background()
	dbs, err := getDatabases()
	assert.NoError(t, err)
	assert.NotEmpty(t, dbs)

	for _, dbi := range dbs {
		fmt.Println("--------------")
		//nolint:gosimple
		fmt.Println(fmt.Sprintf("Running tests with database type: %s", dbi.name))

		db := dbi.db
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
}
