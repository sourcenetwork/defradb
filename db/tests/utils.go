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
	"fmt"
	"os"
	"strings"
	"testing"

	badger "github.com/dgraph-io/badger/v3"
	ds "github.com/ipfs/go-datastore"
	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/defradb/client"
	badgerds "github.com/sourcenetwork/defradb/datastores/badger/v3"
	"github.com/sourcenetwork/defradb/db"
	"github.com/sourcenetwork/defradb/document"
)

const (
	memoryBadgerEnvName = "DEFRA_BADGER_MEMORY"
	fileBadgerEnvName   = "DEFRA_BADGER_FILE"
	memoryMapEnvName    = "DEFRA_MAP"
)

var (
	badgerInMemory bool
	badgerFile     bool
	mapStore       bool
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
	// The expected content of an expected error
	ExpectedError string
}

type databaseInfo struct {
	name      string
	db        *db.DB
	rootstore ds.Batching
}

func (dbi databaseInfo) Rootstore() ds.Batching {
	return dbi.rootstore
}

func (dbi databaseInfo) DB() *db.DB {
	return dbi.db
}

func init() {
	// We use environment variables instead of flags `go test ./...` throws for all packages that don't have the flag defined
	_, badgerInMemory = os.LookupEnv(memoryBadgerEnvName)
	_, badgerFile = os.LookupEnv(fileBadgerEnvName)
	_, mapStore = os.LookupEnv(memoryMapEnvName)

	// default is to run against all
	if !badgerInMemory && !badgerFile && !mapStore {
		badgerInMemory = true
		// Testing against the file system is off by default
		badgerFile = false
		mapStore = true
	}
}

func NewBadgerMemoryDB() (databaseInfo, error) {
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
		name:      "badger-in-memory",
		db:        db,
		rootstore: rootstore,
	}, nil
}

func NewMapDB() (databaseInfo, error) {
	rootstore := ds.NewMapDatastore()
	db, err := db.NewDB(rootstore, struct{}{})
	if err != nil {
		return databaseInfo{}, err
	}

	return databaseInfo{
		name:      "ipfs-map-datastore",
		db:        db,
		rootstore: rootstore,
	}, nil
}

func NewBadgerFileDB(t testing.TB) (databaseInfo, error) {
	path := t.TempDir()

	opts := badgerds.Options{Options: badger.DefaultOptions(path)}
	rootstore, err := badgerds.NewDatastore(path, &opts)
	if err != nil {
		return databaseInfo{}, err
	}

	db, err := db.NewDB(rootstore, struct{}{})
	if err != nil {
		return databaseInfo{}, err
	}

	return databaseInfo{
		name:      "badger-file-system",
		db:        db,
		rootstore: rootstore,
	}, nil
}

func getDatabases(t *testing.T) ([]databaseInfo, error) {
	databases := []databaseInfo{}

	if badgerInMemory {
		badgerIMDatabase, err := NewBadgerMemoryDB()
		if err != nil {
			return nil, err
		}
		databases = append(databases, badgerIMDatabase)
	}

	if badgerFile {
		badgerIMDatabase, err := NewBadgerFileDB(t)
		if err != nil {
			return nil, err
		}
		databases = append(databases, badgerIMDatabase)
	}

	if mapStore {
		mapDatabase, err := NewMapDB()
		if err != nil {
			return nil, err
		}
		databases = append(databases, mapDatabase)
	}

	return databases, nil
}

func ExecuteQueryTestCase(t *testing.T, schema string, collectionNames []string, test QueryTestCase) {
	ctx := context.Background()
	dbs, err := getDatabases(t)
	if assertError(t, err, test) {
		return
	}
	assert.NotEmpty(t, dbs)

	for _, dbi := range dbs {
		fmt.Println("--------------")
		fmt.Println(fmt.Sprintf("Running tests with database type: %s", dbi.name))

		db := dbi.db
		err = db.AddSchema(ctx, schema)
		if assertError(t, err, test) {
			return
		}

		collections := []client.Collection{}
		for _, collectionName := range collectionNames {
			col, err := db.GetCollection(ctx, collectionName)
			if assertError(t, err, test) {
				return
			}
			collections = append(collections, col)
			fmt.Printf("Collection name:%s id%v\n", col.Name(), col.ID())
		}

		// insert docs
		for cid, docs := range test.Docs {
			for i, docStr := range docs {
				doc, err := document.NewFromJSON([]byte(docStr))
				if assertError(t, err, test) {
					return
				}
				err = collections[cid].Save(ctx, doc)
				if assertError(t, err, test) {
					return
				}

				// check for updates
				updates, ok := test.Updates[i]
				if ok {
					for _, u := range updates {
						err = doc.SetWithJSON([]byte(u))
						if assertError(t, err, test) {
							return
						}
						err = collections[cid].Save(ctx, doc)
						if assertError(t, err, test) {
							return
						}
					}
				}
			}
		}

		// exec query
		result := db.ExecQuery(ctx, test.Query)
		if assertErrors(t, result.Errors, test) {
			return
		}

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

		if test.ExpectedError != "" {
			assert.Fail(t, "Expected an error however none was raised.", test.Description)
		}
	}
}

// Asserts as to whether an error has been raised as expected (or not). If an expected
// error has been raised it will return true, returns false in all other cases.
func assertErrors(t *testing.T, errors []interface{}, testCase QueryTestCase) bool {
	if testCase.ExpectedError == "" {
		assert.Empty(t, errors, testCase.Description)
	} else {
		for _, e := range errors {
			// This is always a string at the moment, add support for other types as and when needed
			errorString := e.(string)
			if !strings.Contains(errorString, testCase.ExpectedError) {
				// We use ErrorIs for clearer failures (is a error comparision even if it is just a string)
				assert.ErrorIs(t, fmt.Errorf(errorString), fmt.Errorf(testCase.ExpectedError))
				continue
			}
			return true
		}
	}
	return false
}

// Asserts as to whether an error has been raised as expected (or not). If an expected
// error has been raised it will return true, returns false in all other cases.
func assertError(t *testing.T, err error, testCase QueryTestCase) bool {
	if err == nil {
		return false
	}

	if testCase.ExpectedError == "" {
		assert.NoError(t, err, testCase.Description)
		return false
	} else {
		if !strings.Contains(err.Error(), testCase.ExpectedError) {
			assert.ErrorIs(t, err, fmt.Errorf(testCase.ExpectedError))
			return false
		}
		return true
	}
}
