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
	"github.com/sourcenetwork/defradb/core"
	badgerds "github.com/sourcenetwork/defradb/datastores/badger/v3"
	"github.com/sourcenetwork/defradb/db"
	"github.com/sourcenetwork/defradb/document"
	"github.com/sourcenetwork/defradb/logging"
)

const (
	memoryBadgerEnvName = "DEFRA_BADGER_MEMORY"
	fileBadgerEnvName   = "DEFRA_BADGER_FILE"
	memoryMapEnvName    = "DEFRA_MAP"
)

var (
	log            = logging.MustNewLogger("defra.db.tests")
	badgerInMemory bool
	badgerFile     bool
	mapStore       bool
)

// Represents a query assigned to a particular transaction.
type TransactionQuery struct {
	// Used to identify the transaction for this to run against (allows multiple queries to share a single transaction)
	TransactionId int
	// The query to run against the transaction
	Query string
	// The expected (data) results of the query
	Results []map[string]interface{}
	// The expected error resulting from the query.  Also checked against the txn commit.
	ExpectedError string
}

type QueryTestCase struct {
	Description string
	Query       string
	// A collection of queries tied to a specific transaction.
	// These will be executed before `Query` (if specified), in the order that they are listed here.
	TransactionalQueries []TransactionQuery

	// docs is a map from Collection Index, to a list
	// of docs in stringified JSON format
	Docs map[int][]string

	// updates is a map from document index, to a list
	// of changes in strinigied JSON format
	Updates map[int][]string
	Results []map[string]interface{}
	// The expected content of an expected error
	ExpectedError string

	// If this is set to true, test case will not be run against the mapStore.
	// Useful if the functionality under test is not supported by it.
	DisableMapStore bool
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

func NewBadgerMemoryDB(ctx context.Context) (databaseInfo, error) {
	opts := badgerds.Options{Options: badger.DefaultOptions("").WithInMemory(true)}
	rootstore, err := badgerds.NewDatastore("", &opts)
	if err != nil {
		return databaseInfo{}, err
	}

	db, err := db.NewDB(ctx, rootstore)
	if err != nil {
		return databaseInfo{}, err
	}

	return databaseInfo{
		name:      "badger-in-memory",
		db:        db,
		rootstore: rootstore,
	}, nil
}

func NewMapDB(ctx context.Context) (databaseInfo, error) {
	rootstore := ds.NewMapDatastore()
	db, err := db.NewDB(ctx, rootstore)
	if err != nil {
		return databaseInfo{}, err
	}

	return databaseInfo{
		name:      "ipfs-map-datastore",
		db:        db,
		rootstore: rootstore,
	}, nil
}

func NewBadgerFileDB(ctx context.Context, t testing.TB) (databaseInfo, error) {
	path := t.TempDir()

	opts := badgerds.Options{Options: badger.DefaultOptions(path)}
	rootstore, err := badgerds.NewDatastore(path, &opts)
	if err != nil {
		return databaseInfo{}, err
	}

	db, err := db.NewDB(ctx, rootstore)
	if err != nil {
		return databaseInfo{}, err
	}

	return databaseInfo{
		name:      "badger-file-system",
		db:        db,
		rootstore: rootstore,
	}, nil
}

func getDatabases(ctx context.Context, t *testing.T, test QueryTestCase) ([]databaseInfo, error) {
	databases := []databaseInfo{}

	if badgerInMemory {
		badgerIMDatabase, err := NewBadgerMemoryDB(ctx)
		if err != nil {
			return nil, err
		}
		databases = append(databases, badgerIMDatabase)
	}

	if badgerFile {
		badgerIMDatabase, err := NewBadgerFileDB(ctx, t)
		if err != nil {
			return nil, err
		}
		databases = append(databases, badgerIMDatabase)
	}

	if !test.DisableMapStore && mapStore {
		mapDatabase, err := NewMapDB(ctx)
		if err != nil {
			return nil, err
		}
		databases = append(databases, mapDatabase)
	}

	return databases, nil
}

func ExecuteQueryTestCase(t *testing.T, schema string, collectionNames []string, test QueryTestCase) {
	ctx := context.Background()
	dbs, err := getDatabases(ctx, t, test)
	if assertError(t, test.Description, err, test.ExpectedError) {
		return
	}
	assert.NotEmpty(t, dbs)

	for _, dbi := range dbs {
		// We log with level warn to highlight this item
		log.Warn(ctx, test.Description, logging.NewKV("Database", dbi.name))

		db := dbi.db
		err = db.AddSchema(ctx, schema)
		if assertError(t, test.Description, err, test.ExpectedError) {
			return
		}

		collections := []client.Collection{}
		for _, collectionName := range collectionNames {
			col, err := db.GetCollection(ctx, collectionName)
			if assertError(t, test.Description, err, test.ExpectedError) {
				return
			}
			collections = append(collections, col)
		}

		// insert docs
		for cid, docs := range test.Docs {
			for i, docStr := range docs {
				doc, err := document.NewFromJSON([]byte(docStr))
				if assertError(t, test.Description, err, test.ExpectedError) {
					return
				}
				err = collections[cid].Save(ctx, doc)
				if assertError(t, test.Description, err, test.ExpectedError) {
					return
				}

				// check for updates
				updates, ok := test.Updates[i]
				if ok {
					for _, u := range updates {
						err = doc.SetWithJSON([]byte(u))
						if assertError(t, test.Description, err, test.ExpectedError) {
							return
						}
						err = collections[cid].Save(ctx, doc)
						if assertError(t, test.Description, err, test.ExpectedError) {
							return
						}
					}
				}
			}
		}

		// Create the transactions before executing and queries
		transactions := make([]core.Txn, 0, len(test.TransactionalQueries))
		erroredQueries := make([]bool, len(test.TransactionalQueries))
		for i, tq := range test.TransactionalQueries {
			if len(transactions) < tq.TransactionId {
				continue
			}

			txn, err := db.NewTxn(ctx, false)
			if err != nil {
				if assertError(t, test.Description, err, tq.ExpectedError) {
					erroredQueries[i] = true
				}
			}
			defer txn.Discard(ctx)
			if len(transactions) <= tq.TransactionId {
				transactions = transactions[:tq.TransactionId+1]
			}
			transactions[tq.TransactionId] = txn
		}

		for i, tq := range test.TransactionalQueries {
			if erroredQueries[i] {
				continue
			}
			result := db.ExecTransactionalQuery(ctx, tq.Query, transactions[tq.TransactionId])
			if assertQueryResults(ctx, t, test.Description, result, tq.Results, tq.ExpectedError) {
				erroredQueries[i] = true
			}
		}

		txnIndexesCommited := map[int]struct{}{}
		for i, tq := range test.TransactionalQueries {
			if erroredQueries[i] {
				continue
			}
			if _, alreadyCommited := txnIndexesCommited[tq.TransactionId]; alreadyCommited {
				continue
			}
			txnIndexesCommited[tq.TransactionId] = struct{}{}

			err := transactions[tq.TransactionId].Commit(ctx)
			if assertError(t, test.Description, err, tq.ExpectedError) {
				erroredQueries[i] = true
			}
		}

		for i, tq := range test.TransactionalQueries {
			if tq.ExpectedError != "" && !erroredQueries[i] {
				assert.Fail(t, "Expected an error however none was raised.", test.Description)
			}
		}

		// We run the core query after the explicitly transactional ones to permit tests to query the commited result of the transactional queries
		if test.Query != "" {
			result := db.ExecQuery(ctx, test.Query)
			if assertQueryResults(ctx, t, test.Description, result, test.Results, test.ExpectedError) {
				continue
			}

			if test.ExpectedError != "" {
				assert.Fail(t, "Expected an error however none was raised.", test.Description)
			}
		}
	}
}

func assertQueryResults(ctx context.Context, t *testing.T, description string, result *client.QueryResult, expectedResults []map[string]interface{}, expectedError string) bool {
	if assertErrors(t, description, result.Errors, expectedError) {
		return true
	}
	resultantData := result.Data.([]map[string]interface{})

	// We log with level warn to highlight this item
	log.Warn(ctx, "", logging.NewKV("QueryResults", result.Data))

	// compare results
	assert.Equal(t, len(expectedResults), len(resultantData), description)
	if len(expectedResults) == 0 {
		assert.Equal(t, expectedResults, resultantData)
	}
	for i, result := range resultantData {
		assert.Equal(t, expectedResults[i], result, description)
	}

	return false
}

// Asserts as to whether an error has been raised as expected (or not). If an expected
// error has been raised it will return true, returns false in all other cases.
func assertError(t *testing.T, description string, err error, expectedError string) bool {
	if err == nil {
		return false
	}

	if expectedError == "" {
		assert.NoError(t, err, description)
		return false
	} else {
		if !strings.Contains(err.Error(), expectedError) {
			assert.ErrorIs(t, err, fmt.Errorf(expectedError))
			return false
		}
		return true
	}
}

// Asserts as to whether an error has been raised as expected (or not). If an expected
// error has been raised it will return true, returns false in all other cases.
func assertErrors(t *testing.T, description string, errors []interface{}, expectedError string) bool {
	if expectedError == "" {
		assert.Empty(t, errors, description)
	} else {
		for _, e := range errors {
			// This is always a string at the moment, add support for other types as and when needed
			errorString := e.(string)
			if !strings.Contains(errorString, expectedError) {
				// We use ErrorIs for clearer failures (is a error comparision even if it is just a string)
				assert.ErrorIs(t, fmt.Errorf(errorString), fmt.Errorf(expectedError))
				continue
			}
			return true
		}
	}
	return false
}
