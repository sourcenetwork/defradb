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
	"os"
	"strings"
	"syscall"
	"testing"
	"time"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	badgerds "github.com/sourcenetwork/defradb/datastore/badger/v3"
	"github.com/sourcenetwork/defradb/datastore/memory"
	"github.com/sourcenetwork/defradb/db"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/logging"
	"github.com/sourcenetwork/defradb/node"
)

const (
	memoryBadgerEnvName        = "DEFRA_BADGER_MEMORY"
	fileBadgerEnvName          = "DEFRA_BADGER_FILE"
	fileBadgerPathEnvName      = "DEFRA_BADGER_FILE_PATH"
	inMemoryEnvName            = "DEFRA_IN_MEMORY"
	setupOnlyEnvName           = "DEFRA_SETUP_ONLY"
	detectDbChangesEnvName     = "DEFRA_DETECT_DATABASE_CHANGES"
	repositoryEnvName          = "DEFRA_CODE_REPOSITORY"
	targetBranchEnvName        = "DEFRA_TARGET_BRANCH"
	documentationDirectoryName = "data_format_changes"
)

// The integration tests open many files. This increases the limits on the number of open files of
// the process to fix this issue. This is done by default in Go 1.19.
func init() {
	var lim syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &lim); err == nil && lim.Cur != lim.Max {
		lim.Cur = lim.Max
		err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &lim)
		if err != nil {
			log.ErrorE(context.Background(), "error setting rlimit", err)
		}
	}
}

type DatabaseType string

const (
	badgerIMType   DatabaseType = "badger-in-memory"
	defraIMType    DatabaseType = "defra-memory-datastore"
	badgerFileType DatabaseType = "badger-file-system"
)

var (
	log            = logging.MustNewLogger("defra.tests.integration")
	badgerInMemory bool
	badgerFile     bool
	inMemoryStore  bool
)

const subscriptionTimeout = 1 * time.Second

var databaseDir string

/*
If this is set to true the integration test suite will instead of its normal profile do
the following:

On [package] Init:
  - Get the (local) latest commit from the target/parent branch // code assumes
    git fetch has been done
  - Check to see if a clone of that commit/branch is available in the temp dir, and
    if not clone the target branch
  - Check to see if there are any new .md files in the current branch's data_format_changes
    dir (vs the target branch)

For each test:
  - If new documentation detected, pass the test and exit
  - Create a new (test/auto-deleted) temp dir for defra to live/run in
  - Run the test setup (add initial schema, docs, updates) using the target branch (test is skipped
    if test does not exist in target and is new to this branch)
  - Run the test request and assert results (as per normal tests) using the current branch
*/
var DetectDbChanges bool
var SetupOnly bool

var detectDbChangesCodeDir string
var areDatabaseFormatChangesDocumented bool
var previousTestCaseTestName string

func init() {
	// We use environment variables instead of flags `go test ./...` throws for all packages
	//  that don't have the flag defined
	badgerFileValue, _ := os.LookupEnv(fileBadgerEnvName)
	badgerInMemoryValue, _ := os.LookupEnv(memoryBadgerEnvName)
	databaseDir, _ = os.LookupEnv(fileBadgerPathEnvName)
	detectDbChangesValue, _ := os.LookupEnv(detectDbChangesEnvName)
	inMemoryStoreValue, _ := os.LookupEnv(inMemoryEnvName)
	repositoryValue, repositorySpecified := os.LookupEnv(repositoryEnvName)
	setupOnlyValue, _ := os.LookupEnv(setupOnlyEnvName)
	targetBranchValue, targetBranchSpecified := os.LookupEnv(targetBranchEnvName)

	badgerFile = getBool(badgerFileValue)
	badgerInMemory = getBool(badgerInMemoryValue)
	inMemoryStore = getBool(inMemoryStoreValue)
	DetectDbChanges = getBool(detectDbChangesValue)
	SetupOnly = getBool(setupOnlyValue)

	if !repositorySpecified {
		repositoryValue = "git@github.com:sourcenetwork/defradb.git"
	}

	if !targetBranchSpecified {
		targetBranchValue = "develop"
	}

	// default is to run against all
	if !badgerInMemory && !badgerFile && !inMemoryStore && !DetectDbChanges {
		badgerInMemory = true
		// Testing against the file system is off by default
		badgerFile = false
		inMemoryStore = true
	}

	if DetectDbChanges {
		detectDbChangesInit(repositoryValue, targetBranchValue)
	}
}

func getBool(val string) bool {
	switch strings.ToLower(val) {
	case "true":
		return true
	default:
		return false
	}
}

// AssertPanicAndSkipChangeDetection asserts that the code of function actually panics,
//
//	also ensures the change detection is skipped so no false fails happen.
//
//	Usage: AssertPanicAndSkipChangeDetection(t, func() { executeTestCase(t, test) })
func AssertPanicAndSkipChangeDetection(t *testing.T, f assert.PanicTestFunc) bool {
	if IsDetectingDbChanges() {
		// The `assert.Panics` call will falsely fail if this test is executed during
		// a detect changes test run
		t.Skip()
	}
	return assert.Panics(t, f, "expected a panic, but none found.")
}

func NewBadgerMemoryDB(ctx context.Context, dbopts ...db.Option) (client.DB, error) {
	opts := badgerds.Options{Options: badger.DefaultOptions("").WithInMemory(true)}
	rootstore, err := badgerds.NewDatastore("", &opts)
	if err != nil {
		return nil, err
	}

	dbopts = append(dbopts, db.WithUpdateEvents())

	db, err := db.NewDB(ctx, rootstore, dbopts...)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func NewInMemoryDB(ctx context.Context) (client.DB, error) {
	rootstore := memory.NewDatastore(ctx)
	db, err := db.NewDB(ctx, rootstore, db.WithUpdateEvents())
	if err != nil {
		return nil, err
	}

	return db, nil
}

func NewBadgerFileDB(ctx context.Context, t testing.TB) (client.DB, error) {
	var path string
	if databaseDir == "" {
		path = t.TempDir()
	} else {
		path = databaseDir
	}

	return newBadgerFileDB(ctx, t, path)
}

func newBadgerFileDB(ctx context.Context, t testing.TB, path string) (client.DB, error) {
	opts := badgerds.Options{Options: badger.DefaultOptions(path)}
	rootstore, err := badgerds.NewDatastore(path, &opts)
	if err != nil {
		return nil, err
	}

	db, err := db.NewDB(ctx, rootstore, db.WithUpdateEvents())
	if err != nil {
		return nil, err
	}

	return db, nil
}

func GetDatabaseTypes() []DatabaseType {
	databases := []DatabaseType{}

	if badgerInMemory {
		databases = append(databases, badgerIMType)
	}

	if badgerFile {
		databases = append(databases, badgerFileType)
	}

	if inMemoryStore {
		databases = append(databases, defraIMType)
	}

	return databases
}

func GetDatabase(ctx context.Context, t *testing.T, dbt DatabaseType) (client.DB, error) {
	switch dbt {
	case badgerIMType:
		db, err := NewBadgerMemoryDB(ctx)
		if err != nil {
			return nil, err
		}
		return db, nil

	case badgerFileType:
		db, err := NewBadgerFileDB(ctx, t)
		if err != nil {
			return nil, err
		}
		return db, nil

	case defraIMType:
		db, err := NewInMemoryDB(ctx)
		if err != nil {
			return nil, err
		}
		return db, nil
	}

	return nil, nil
}

// ExecuteTestCase executes the given TestCase against the configured database
// instances.
//
// Will also attempt to detect incompatible changes in the persisted data if
// configured to do so (the CI will do so, but disabled by default as it is slow).
func ExecuteTestCase(
	t *testing.T,
	collectionNames []string,
	testCase TestCase,
) {
	if DetectDbChanges && DetectDbChangesPreTestChecks(t, collectionNames) {
		return
	}

	ctx := context.Background()
	dbs := GetDatabaseTypes()
	// Assert that this is not empty to protect against accidental mis-configurations,
	// otherwise an empty set would silently pass all the tests.
	require.NotEmpty(t, dbs)

	for _, db := range dbs {
		executeTestCase(ctx, t, collectionNames, testCase, db)
	}
}

func executeTestCase(
	ctx context.Context,
	t *testing.T,
	collectionNames []string,
	testCase TestCase,
	dbt DatabaseType,
) {
	var done bool
	log.Info(ctx, testCase.Description, logging.NewKV("Database", dbt))

	startActionIndex, endActionIndex := getActionRange(testCase)
	// Documents and Collections may already exist in the database if actions have been split
	// by the change detector so we should fetch them here at the start too (if they exist).
	collections := getCollections(ctx, t, dbi.db, collectionNames)
	documents := getDocuments(ctx, t, testCase, collections, startActionIndex)
	txns := []datastore.Txn{}
	allActionsDone := make(chan struct{})
	resultsChans := []chan func(){}
	nodes, onExit := getStartingNodes(ctx, t, dbt, collectionNames)
	defer onExit()

	for i := startActionIndex; i <= endActionIndex; i++ {
		// declare default database for ease of use
		var db client.DB
		if len(nodes) > 0 {
			db = nodes[0].DB
		}

		switch action := testCase.Actions[i].(type) {
		case SchemaUpdate:
			updateSchema(ctx, t, db, testCase, action)
			// If the schema was updated we need to refresh the collection definitions.
			collections = getCollections(ctx, t, db, collectionNames)

		case SchemaPatch:
			patchSchema(ctx, t, dbi.db, testCase, action)
			// If the schema was updated we need to refresh the collection definitions.
			collections = getCollections(ctx, t, dbi.db, collectionNames)

		case CreateDoc:
			documents = createDoc(ctx, t, testCase, collections, documents, action)

		case UpdateDoc:
			updateDoc(ctx, t, testCase, collections, documents, action)

		case TransactionRequest2:
			txns = executeTransactionRequest(ctx, t, db, txns, testCase, action)

		case TransactionCommit:
			commitTransaction(ctx, t, txns, testCase, action)

		case SubscriptionRequest:
			var resultsChan chan func()
			resultsChan, done = executeSubscriptionRequest(ctx, t, allActionsDone, db, testCase, action)
			if done {
				return
			}
			resultsChans = append(resultsChans, resultsChan)

		case Request:
			executeRequest(ctx, t, db, testCase, action)

		case SetupComplete:
			// no-op, just continue.

		default:
			t.Fatalf("Unknown action type %T", action)
		}
	}

	// Notify any active subscriptions that all requests have been sent.
	close(allActionsDone)

	for _, resultsChan := range resultsChans {
		select {
		case subscriptionAssert := <-resultsChan:
			// We want to assert back in the main thread so failures get recorded properly
			subscriptionAssert()

		// a safety in case the stream hangs - we don't want the tests to run forever.
		case <-time.After(subscriptionTimeout):
			assert.Fail(t, "timeout occured while waiting for data stream", testCase.Description)
		}
	}
}

// getActionRange returns the index of the first action to be run, and the last.
//
// Not all processes will run all actions - if this is a change detector run they
// will be split.
//
// If a SetupComplete action is provided, the actions will be split there, if not
// they will be split at the first non SchemaUpdate/CreateDoc/UpdateDoc action.
func getActionRange(testCase TestCase) (int, int) {
	startIndex := 0
	endIndex := len(testCase.Actions) - 1

	if !DetectDbChanges {
		return startIndex, endIndex
	}

	setupCompleteIndex := -1
	firstNonSetupIndex := -1

ActionLoop:
	for i := range testCase.Actions {
		switch testCase.Actions[i].(type) {
		case SetupComplete:
			setupCompleteIndex = i
			// We dont care about anything else if this has been explicitly provided
			break ActionLoop

		case SchemaUpdate, CreateDoc, UpdateDoc:
			continue

		default:
			firstNonSetupIndex = i
			break ActionLoop
		}
	}

	if SetupOnly {
		if setupCompleteIndex > -1 {
			endIndex = setupCompleteIndex
		} else if firstNonSetupIndex > -1 {
			// -1 to exclude this index
			endIndex = firstNonSetupIndex - 1
		}
	} else {
		if setupCompleteIndex > -1 {
			// +1 to exclude the SetupComplete action
			startIndex = setupCompleteIndex + 1
		} else if firstNonSetupIndex > -1 {
			// We must not set this to -1 :)
			startIndex = firstNonSetupIndex
		}
	}

	return startIndex, endIndex
}

func getStartingNodes(
	ctx context.Context,
	t *testing.T,
	dbt DatabaseType,
	collectionNames []string,
) ([]*node.Node, func()) {
	var db client.DB
	if DetectDbChanges && !SetupOnly {
		// Setup the database using the target branch, and then refresh the current instance
		db = SetupDatabaseUsingTargetBranch(ctx, t, collectionNames)
	} else {
		var err error
		db, err = GetDatabase(ctx, t, dbt)
		require.Nil(t, err)
	}

	return []*node.Node{
		{
			DB: db,
		},
	}, func() { defer db.Close(ctx) }
}

// getCollections returns all the collections of the given names, preserving order.
//
// If a given collection is not present in the database the value at the corresponding
// result-index will be nil.
func getCollections(
	ctx context.Context,
	t *testing.T,
	db client.DB,
	collectionNames []string,
) []client.Collection {
	collections := make([]client.Collection, len(collectionNames))

	allCollections, err := db.GetAllCollections(ctx)
	require.Nil(t, err)

	for i, collectionName := range collectionNames {
		for _, collection := range allCollections {
			if collection.Name() == collectionName {
				collections[i] = collection
				break
			}
		}
	}
	return collections
}

func getDocuments(
	ctx context.Context,
	t *testing.T,
	testCase TestCase,
	collections []client.Collection,
	startActionIndex int,
) [][]*client.Document {
	documentsByCollection := make([][]*client.Document, len(collections))

	for i := range collections {
		documentsByCollection[i] = []*client.Document{}
	}

	for i := 0; i < startActionIndex; i++ {
		switch action := testCase.Actions[i].(type) {
		case CreateDoc:
			// We need to add the existing documents in the order in which the test case lists them
			// otherwise they cannot be referenced correctly by other actions.
			doc, err := client.NewDocFromJSON([]byte(action.Doc))
			if err != nil {
				// If an err has been returned, ignore it - it may be expected and if not
				// the test will fail later anyway
				continue
			}

			collection := collections[action.CollectionID]
			// The document may have been mutated by other actions, so to be sure we have the latest
			// version without having to worry about the individual update mechanics we fetch it.
			doc, err = collection.Get(ctx, doc.Key())
			if err != nil {
				// If an err has been returned, ignore it - it may be expected and if not
				// the test will fail later anyway
				continue
			}

			documentsByCollection[action.CollectionID] = append(documentsByCollection[action.CollectionID], doc)
		}
	}

	return documentsByCollection
}

// updateSchema updates the schema using the given details.
func updateSchema(
	ctx context.Context,
	t *testing.T,
	db client.DB,
	testCase TestCase,
	action SchemaUpdate,
) {
	err := db.AddSchema(ctx, action.Schema)
	expectedErrorRaised := AssertError(t, testCase.Description, err, action.ExpectedError)

	assertExpectedErrorRaised(t, testCase.Description, action.ExpectedError, expectedErrorRaised)
}

func patchSchema(
	ctx context.Context,
	t *testing.T,
	db client.DB,
	testCase TestCase,
	action SchemaPatch,
) {
	err := db.PatchSchema(ctx, action.Patch)
	expectedErrorRaised := AssertError(t, testCase.Description, err, action.ExpectedError)

	assertExpectedErrorRaised(t, testCase.Description, action.ExpectedError, expectedErrorRaised)
}

// createDoc creates a document using the collection api and caches it in the
// given documents slice.
func createDoc(
	ctx context.Context,
	t *testing.T,
	testCase TestCase,
	collections []client.Collection,
	documents [][]*client.Document,
	action CreateDoc,
) [][]*client.Document {
	doc, err := client.NewDocFromJSON([]byte(action.Doc))
	if AssertError(t, testCase.Description, err, action.ExpectedError) {
		return nil
	}

	err = collections[action.CollectionID].Save(ctx, doc)
	expectedErrorRaised := AssertError(t, testCase.Description, err, action.ExpectedError)
	if expectedErrorRaised {
		return nil
	}

	assertExpectedErrorRaised(t, testCase.Description, action.ExpectedError, expectedErrorRaised)

	if action.CollectionID >= len(documents) {
		// Expand the slice if required, so that the document can be accessed by collection index
		documents = append(documents, make([][]*client.Document, action.CollectionID-len(documents)+1)...)
	}
	documents[action.CollectionID] = append(documents[action.CollectionID], doc)

	return documents
}

// updateDoc updates a document using the collection api.
func updateDoc(
	ctx context.Context,
	t *testing.T,
	testCase TestCase,
	collections []client.Collection,
	documents [][]*client.Document,
	action UpdateDoc,
) {
	doc := documents[action.CollectionID][action.DocID]

	err := doc.SetWithJSON([]byte(action.Doc))
	if AssertError(t, testCase.Description, err, action.ExpectedError) {
		return
	}

	err = collections[action.CollectionID].Save(ctx, doc)
	expectedErrorRaised := AssertError(t, testCase.Description, err, action.ExpectedError)

	assertExpectedErrorRaised(t, testCase.Description, action.ExpectedError, expectedErrorRaised)
}

// executeTransactionRequest executes the given transactional request.
//
// It will create and cache a new transaction if it is the first of the given
// TransactionId. If an error is returned the transaction will be discarded before
// this function returns.
func executeTransactionRequest(
	ctx context.Context,
	t *testing.T,
	db client.DB,
	txns []datastore.Txn,
	testCase TestCase,
	action TransactionRequest2,
) []datastore.Txn {
	if action.TransactionID >= len(txns) {
		// Extend the txn slice so this txn can fit and be accessed by TransactionId
		txns = append(txns, make([]datastore.Txn, action.TransactionID-len(txns)+1)...)
	}

	if txns[action.TransactionID] == nil {
		// Create a new transaction if one does not already exist.
		txn, err := db.NewTxn(ctx, false)
		if AssertError(t, testCase.Description, err, action.ExpectedError) {
			txn.Discard(ctx)
			return nil
		}

		txns[action.TransactionID] = txn
	}

	result := db.ExecTransactionalRequest(ctx, action.Request, txns[action.TransactionID])
	expectedErrorRaised := assertRequestResults(
		ctx,
		t,
		testCase.Description,
		&result.GQL,
		action.Results,
		action.ExpectedError,
	)

	assertExpectedErrorRaised(t, testCase.Description, action.ExpectedError, expectedErrorRaised)

	if expectedErrorRaised {
		// Make sure to discard the transaction before exit, else an unwanted error
		// may surface later (e.g. on database close).
		txns[action.TransactionID].Discard(ctx)
		return nil
	}

	return txns
}

// commitTransaction commits the given transaction.
//
// Will panic if the given transaction does not exist. Discards the transaction if
// an error is returned on commit.
func commitTransaction(
	ctx context.Context,
	t *testing.T,
	txns []datastore.Txn,
	testCase TestCase,
	action TransactionCommit,
) {
	err := txns[action.TransactionID].Commit(ctx)
	if err != nil {
		txns[action.TransactionID].Discard(ctx)
	}

	expectedErrorRaised := AssertError(t, testCase.Description, err, action.ExpectedError)

	assertExpectedErrorRaised(t, testCase.Description, action.ExpectedError, expectedErrorRaised)
}

// executeRequest executes the given request.
func executeRequest(
	ctx context.Context,
	t *testing.T,
	db client.DB,
	testCase TestCase,
	action Request,
) {
	result := db.ExecRequest(ctx, action.Request)
	expectedErrorRaised := assertRequestResults(
		ctx,
		t,
		testCase.Description,
		&result.GQL,
		action.Results,
		action.ExpectedError,
	)

	assertExpectedErrorRaised(t, testCase.Description, action.ExpectedError, expectedErrorRaised)
}

// executeSubscriptionRequest executes the given subscription request, returning
// a channel that will receive a single event once the subscription has been completed.
//
// The returned channel will receive a function that asserts that
// the subscription received all its expected results and no more.
// It should be called from the main test routine to ensure that
// failures are recorded properly. It will only yield once, once
// the subscription has terminated.
func executeSubscriptionRequest(
	ctx context.Context,
	t *testing.T,
	allActionsDone chan struct{},
	db client.DB,
	testCase TestCase,
	action SubscriptionRequest,
) (chan func(), bool) {
	subscriptionAssert := make(chan func())

	result := db.ExecRequest(ctx, action.Request)
	if AssertErrors(t, testCase.Description, result.GQL.Errors, action.ExpectedError) {
		return nil, true
	}

	go func() {
		data := []map[string]any{}
		errs := []any{}

		allActionsAreDone := false
		expectedDataRecieved := len(action.Results) == 0
		stream := result.Pub.Stream()
		for {
			select {
			case s := <-stream:
				sResult, _ := s.(client.GQLResult)
				sData, _ := sResult.Data.([]map[string]any)
				errs = append(errs, sResult.Errors...)
				data = append(data, sData...)

				if len(data) >= len(action.Results) {
					expectedDataRecieved = true
				}

			case <-allActionsDone:
				allActionsAreDone = true
			}

			if expectedDataRecieved && allActionsAreDone {
				finalResult := &client.GQLResult{
					Data:   data,
					Errors: errs,
				}

				subscriptionAssert <- func() {
					// This assert should be executed from the main test routine
					// so that failures will be properly handled.
					expectedErrorRaised := assertRequestResults(
						ctx,
						t,
						testCase.Description,
						finalResult,
						action.Results,
						action.ExpectedError,
					)

					assertExpectedErrorRaised(t, testCase.Description, action.ExpectedError, expectedErrorRaised)
				}

				return
			}
		}
	}()

	return subscriptionAssert, false
}

// Asserts as to whether an error has been raised as expected (or not). If an expected
// error has been raised it will return true, returns false in all other cases.
func AssertError(t *testing.T, description string, err error, expectedError string) bool {
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

// Asserts as to whether an error has been raised as expected (or not). If an expected
// error has been raised it will return true, returns false in all other cases.
func AssertErrors(
	t *testing.T,
	description string,
	errs []any,
	expectedError string,
) bool {
	if expectedError == "" {
		assert.Empty(t, errs, description)
	} else {
		for _, e := range errs {
			// This is always a string at the moment, add support for other types as and when needed
			errorString := e.(string)
			if !strings.Contains(errorString, expectedError) {
				// We use ErrorIs for clearer failures (is a error comparision even if it is just a string)
				assert.ErrorIs(t, errors.New(errorString), errors.New(expectedError))
				continue
			}
			return true
		}
	}
	return false
}

func assertRequestResults(
	ctx context.Context,
	t *testing.T,
	description string,
	result *client.GQLResult,
	expectedResults []map[string]any,
	expectedError string,
) bool {
	if AssertErrors(t, description, result.Errors, expectedError) {
		return true
	}

	if expectedResults == nil && result.Data == nil {
		return true
	}

	// Note: if result.Data == nil this panics (the panic seems useful while testing).
	resultantData := result.Data.([]map[string]any)

	log.Info(ctx, "", logging.NewKV("RequestResults", result.Data))

	// compare results
	assert.Equal(t, len(expectedResults), len(resultantData), description)
	if len(expectedResults) == 0 {
		assert.Equal(t, expectedResults, resultantData)
	}
	for i, result := range resultantData {
		if len(expectedResults) > i {
			assert.Equal(t, expectedResults[i], result, description)
		}
	}

	return false
}

func assertExpectedErrorRaised(t *testing.T, description string, expectedError string, wasRaised bool) {
	if expectedError != "" && !wasRaised {
		assert.Fail(t, "Expected an error however none was raised.", description)
	}
}
