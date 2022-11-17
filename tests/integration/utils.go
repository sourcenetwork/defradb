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
	"io/fs"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	badger "github.com/dgraph-io/badger/v3"
	ds "github.com/ipfs/go-datastore"
	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	badgerds "github.com/sourcenetwork/defradb/datastore/badger/v3"
	"github.com/sourcenetwork/defradb/datastore/memory"
	"github.com/sourcenetwork/defradb/db"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/logging"
)

const (
	memoryBadgerEnvName        = "DEFRA_BADGER_MEMORY"
	fileBadgerEnvName          = "DEFRA_BADGER_FILE"
	fileBadgerPathEnvName      = "DEFRA_BADGER_FILE_PATH"
	memoryMapEnvName           = "DEFRA_MAP"
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

var (
	log            = logging.MustNewLogger("defra.tests.integration")
	badgerInMemory bool
	badgerFile     bool
	mapStore       bool
)

const subsciptionTimeout = 1 * time.Second

// Represents a query assigned to a particular transaction.
type SubscriptionQuery struct {
	Query string
	// The expected (data) results of the query
	Results []map[string]any
	// The expected error resulting from the query.
	ExpectedError string
	// If set to true, the query should yield no results.
	// The timeout is duration is that of subscriptionTimeout (1 second)
	ExpectedTimout bool
}

// Represents a query assigned to a particular transaction.
type TransactionQuery struct {
	// Used to identify the transaction for this to run against (allows multiple
	//  queries to share a single transaction)
	TransactionId int
	// The query to run against the transaction
	Query string
	// The expected (data) results of the query
	Results []map[string]any
	// The expected error resulting from the query.  Also checked against the txn commit.
	ExpectedError string
}

type QueryTestCase struct {
	Description string
	Query       string

	// A collection of queries to exucute after the subscriber is listening on the stream
	PostSubscriptionQueries []SubscriptionQuery

	// A collection of queries tied to a specific transaction.
	// These will be executed before `Query` (if specified), in the order that they are listed here.
	TransactionalQueries []TransactionQuery

	// docs is a map from Collection Index, to a list
	// of docs in stringified JSON format
	Docs map[int][]string

	// updates is a map from document index, to a list
	// of changes in strinigied JSON format
	Updates map[int]map[int][]string

	Results []map[string]any

	// The expected content of an expected error
	ExpectedError string

	// If this is set to true, test case will not be run against the mapStore.
	// Useful if the functionality under test is not supported by it.
	DisableMapStore bool
}

type databaseInfo struct {
	name      string
	path      string
	db        client.DB
	rootstore ds.Batching
}

func (dbi databaseInfo) Name() string {
	return dbi.name
}

func (dbi databaseInfo) Rootstore() ds.Batching {
	return dbi.rootstore
}

func (dbi databaseInfo) DB() client.DB {
	return dbi.db
}

var databaseDir string

/*
If this is set to true the integration test suite will instead of it's normal profile do
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
- Run the test query and assert results (as per normal tests) using the current branch
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
	mapStoreValue, _ := os.LookupEnv(memoryMapEnvName)
	repositoryValue, repositorySpecified := os.LookupEnv(repositoryEnvName)
	setupOnlyValue, _ := os.LookupEnv(setupOnlyEnvName)
	targetBranchValue, targetBranchSpecified := os.LookupEnv(targetBranchEnvName)

	badgerFile = getBool(badgerFileValue)
	badgerInMemory = getBool(badgerInMemoryValue)
	DetectDbChanges = getBool(detectDbChangesValue)
	mapStore = getBool(mapStoreValue)
	SetupOnly = getBool(setupOnlyValue)

	if !repositorySpecified {
		repositoryValue = "git@github.com:sourcenetwork/defradb.git"
	}

	if !targetBranchSpecified {
		targetBranchValue = "develop"
	}

	// default is to run against all
	if !badgerInMemory && !badgerFile && !mapStore && !DetectDbChanges {
		badgerInMemory = true
		// Testing against the file system is off by default
		badgerFile = false
		mapStore = true
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

func IsDetectingDbChanges() bool {
	return DetectDbChanges
}

// AssertPanicAndSkipChangeDetection asserts that the code of function actually panics,
//  also ensures the change detection is skipped so no false fails happen.
//
//  Usage: AssertPanicAndSkipChangeDetection(t, func() { executeTestCase(t, test) })
func AssertPanicAndSkipChangeDetection(t *testing.T, f assert.PanicTestFunc) bool {
	if IsDetectingDbChanges() {
		// The `assert.Panics` call will falsely fail if this test is executed during
		// a detect changes test run
		t.Skip()
	}
	return assert.Panics(t, f, "expected a panic, but none found.")
}

func NewBadgerMemoryDB(ctx context.Context, dbopts ...db.Option) (databaseInfo, error) {
	opts := badgerds.Options{Options: badger.DefaultOptions("").WithInMemory(true)}
	rootstore, err := badgerds.NewDatastore("", &opts)
	if err != nil {
		return databaseInfo{}, err
	}

	dbopts = append(dbopts, db.WithUpdateEvents())

	db, err := db.NewDB(ctx, rootstore, dbopts...)
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
	rootstore := memory.NewDatastore()
	db, err := db.NewDB(ctx, rootstore, db.WithUpdateEvents())
	if err != nil {
		return databaseInfo{}, err
	}

	return databaseInfo{
		name:      "defra-memory-datastore",
		db:        db,
		rootstore: rootstore,
	}, nil
}

func NewBadgerFileDB(ctx context.Context, t testing.TB) (databaseInfo, error) {
	var path string
	if databaseDir == "" {
		path = t.TempDir()
	} else {
		path = databaseDir
	}

	return newBadgerFileDB(ctx, t, path)
}

func newBadgerFileDB(ctx context.Context, t testing.TB, path string) (databaseInfo, error) {
	opts := badgerds.Options{Options: badger.DefaultOptions(path)}
	rootstore, err := badgerds.NewDatastore(path, &opts)
	if err != nil {
		return databaseInfo{}, err
	}

	db, err := db.NewDB(ctx, rootstore, db.WithUpdateEvents())
	if err != nil {
		return databaseInfo{}, err
	}

	return databaseInfo{
		name:      "badger-file-system",
		path:      path,
		db:        db,
		rootstore: rootstore,
	}, nil
}

func GetDatabases(ctx context.Context, t *testing.T, disableMapStore bool) ([]databaseInfo, error) {
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

	if !disableMapStore && mapStore {
		mapDatabase, err := NewMapDB(ctx)
		if err != nil {
			return nil, err
		}
		databases = append(databases, mapDatabase)
	}

	return databases, nil
}

func ExecuteQueryTestCase(
	t *testing.T,
	schema string,
	collectionNames []string,
	test QueryTestCase,
) {
	isTransactional := len(test.TransactionalQueries) > 0

	if DetectDbChanges && DetectDbChangesPreTestChecks(t, collectionNames, isTransactional) {
		return
	}

	// Must have a non-empty request.
	if !isTransactional && test.Query == "" {
		assert.Fail(t, "Test must have a non-empty request.", test.Description)
	}

	ctx := context.Background()
	dbs, err := GetDatabases(ctx, t, test.DisableMapStore)
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

		// Create the transactions before executing and queries
		transactions := make([]datastore.Txn, 0, len(test.TransactionalQueries))
		erroredQueries := make([]bool, len(test.TransactionalQueries))
		for i, tq := range test.TransactionalQueries {
			if len(transactions) < tq.TransactionId {
				continue
			}

			txn, err := dbi.db.NewTxn(ctx, false)
			if err != nil {
				if AssertError(t, test.Description, err, tq.ExpectedError) {
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
			result := dbi.db.ExecTransactionalQuery(ctx, tq.Query, transactions[tq.TransactionId])
			if assertQueryResults(ctx, t, test.Description, &result.GQL, tq.Results, tq.ExpectedError) {
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
			if AssertError(t, test.Description, err, tq.ExpectedError) {
				erroredQueries[i] = true
			}
		}

		for i, tq := range test.TransactionalQueries {
			if tq.ExpectedError != "" && !erroredQueries[i] {
				assert.Fail(t, "Expected an error however none was raised.", test.Description)
			}
		}

		// We run the core query after the explicitly transactional ones to permit tests to query
		//  the commited result of the transactional queries
		if !isTransactional || (isTransactional && test.Query != "") {
			result := dbi.db.ExecQuery(ctx, test.Query)
			if result.Pub != nil {
				for _, q := range test.PostSubscriptionQueries {
					dbi.db.ExecQuery(ctx, q.Query)
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
							case <-time.After(subsciptionTimeout):
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
						case <-time.After(subsciptionTimeout):
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
					if assertQueryResults(
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
				if assertQueryResults(
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

func detectDbChangesInit(repository string, targetBranch string) {
	badgerFile = true
	badgerInMemory = false
	mapStore = false

	if SetupOnly {
		// Only the primary test process should perform the setup below
		return
	}

	tempDir := os.TempDir()

	latestTargetCommitHash := getLatestCommit(repository, targetBranch)
	detectDbChangesCodeDir = path.Join(tempDir, "defra", latestTargetCommitHash, "code")

	_, err := os.Stat(detectDbChangesCodeDir)
	// Warning - there is a race condition here, where if running multiple packages in
	//  parallel (as per default) against a new target commit multiple test pacakges will
	//  try and clone the target branch at the same time (and will fail).
	// This could be solved by using a file lock or similar, however running the change
	//  detector in parallel is significantly slower than running it serially due to machine
	//  resource constraints, so I am leaving the race condition in and recommending running
	//  the change detector with the CLI args `-p 1`
	if os.IsNotExist(err) {
		cloneCmd := exec.Command(
			"git",
			"clone",
			"-b",
			targetBranch,
			"--single-branch",
			repository,
			detectDbChangesCodeDir,
		)
		cloneCmd.Stdout = os.Stdout
		cloneCmd.Stderr = os.Stderr
		err := cloneCmd.Run()
		if err != nil {
			panic(err)
		}
	} else if err != nil {
		panic(err)
	} else {
		// Cache must be cleaned, or it might not run the test setup!
		// Note: this also acts as a race condition if multiple build are running against the
		//       same target if this happens some tests might be silently skipped if the
		//       child-setup fails.  Currently I think it is worth it for slightly faster build
		//       times, but feel very free to change this!
		goTestCacheCmd := exec.Command("go", "clean", "-testcache")
		goTestCacheCmd.Dir = detectDbChangesCodeDir
		err = goTestCacheCmd.Run()
		if err != nil {
			panic(err)
		}
	}

	areDatabaseFormatChangesDocumented = checkIfDatabaseFormatChangesAreDocumented()
}

// Returns true if test should pass early
func DetectDbChangesPreTestChecks(
	t *testing.T,
	collectionNames []string,
	isTransactional bool,
) bool {
	if previousTestCaseTestName == t.Name() {
		// The database format changer currently only supports running the first test
		//  case, if a second case is detected we return early
		return true
	}
	previousTestCaseTestName = t.Name()

	if areDatabaseFormatChangesDocumented {
		// If we are checking that database formatting changes have been made and
		//  documented, and changes are documented, then the tests can all pass.
		return true
	}

	if isTransactional {
		// Transactional queries are not yet supported by the database change
		//  detector, so we skip the test
		t.SkipNow()
	}

	if len(collectionNames) == 0 {
		// If the test doesn't specify any collections, then we can't use it to check
		//  the database format, so we skip it
		t.SkipNow()
	}

	return false
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

func SetupDatabaseUsingTargetBranch(
	ctx context.Context,
	t *testing.T,
	dbi databaseInfo,
	collectionNames []string,
) databaseInfo {
	// Close this database instance so it may be re-inited in the child process,
	//  and this one post-child
	dbi.db.Close(ctx)

	currentTestPackage, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	targetTestPackage := detectDbChangesCodeDir + "/tests/integration/" + strings.Split(
		currentTestPackage,
		"/tests/integration/",
	)[1]

	// If we are checking for database changes, and we are not seting up the database,
	// then we must be in the main test process, and need to create a new process
	// setting up the database for this test using the old branch We should not setup
	// the database using the current branch/process
	goTestCmd := exec.Command(
		"go",
		"test",
		"./...",
		"--run",
		fmt.Sprintf("^%s$", t.Name()),
		"-v",
	)

	goTestCmd.Dir = targetTestPackage
	goTestCmd.Env = os.Environ()
	goTestCmd.Env = append(
		goTestCmd.Env,
		setupOnlyEnvName+"=true",
		fileBadgerPathEnvName+"="+dbi.path,
	)
	out, err := goTestCmd.Output()

	if err != nil {
		// If file is not found - this must be a new test and
		// doesn't exist in the target branch, so we pass it
		// because the child process tries to run the test, but
		// if it doesnt find it, the parent test should pass (not panic).
		if strings.Contains(err.Error(), ": no such file or directory") {
			t.SkipNow()
		} else {
			// Only log the output if there is an error different from above,
			// logging child test runs confuses the go test runner making it
			// think there are no tests in the parent run (it will still
			// run everything though)!
			log.ErrorE(ctx, string(out), err)
			panic(err)
		}
	}

	refreshedDb, err := newBadgerFileDB(ctx, t, dbi.path)
	if err != nil {
		panic(err)
	}

	_, err = refreshedDb.db.GetCollectionByName(ctx, collectionNames[0])
	if err != nil {
		if err.Error() == "datastore: key not found" {
			// If collection is not found - this must be a new test and
			// doesn't exist in the target branch, so we pass it
			t.SkipNow()
		} else {
			panic(err)
		}
	}
	return refreshedDb
}

func assertQueryResults(
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

	// Note: if result.Data == nil this panics (the panic seems useful while testing).
	resultantData := result.Data.([]map[string]any)

	log.Info(ctx, "", logging.NewKV("QueryResults", result.Data))

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

func checkIfDatabaseFormatChangesAreDocumented() bool {
	previousDbChangeFiles, targetDirFound := getDatabaseFormatDocumentation(
		detectDbChangesCodeDir,
		false,
	)
	if !targetDirFound {
		panic("Documentation directory not found")
	}

	previousDbChanges := make(map[string]struct{}, len(previousDbChangeFiles))
	for _, f := range previousDbChangeFiles {
		// Note: we assume flat directory for now - sub directories are not expanded
		previousDbChanges[f.Name()] = struct{}{}
	}

	_, thisFilePath, _, _ := runtime.Caller(0)
	currentDbChanges, currentDirFound := getDatabaseFormatDocumentation(thisFilePath, true)
	if !currentDirFound {
		panic("Documentation directory not found")
	}

	for _, f := range currentDbChanges {
		if _, isChangeOld := previousDbChanges[f.Name()]; !isChangeOld {
			// If there is a new file in the directory then the change
			// has been documented and the test should pass
			return true
		}
	}

	return false
}

func getDatabaseFormatDocumentation(startPath string, allowDescend bool) ([]fs.DirEntry, bool) {
	startInfo, err := os.Stat(startPath)
	if err != nil {
		panic(err)
	}

	var currentDirectory string
	if startInfo.IsDir() {
		currentDirectory = startPath
	} else {
		currentDirectory = path.Dir(startPath)
	}

	for {
		directoryContents, err := os.ReadDir(currentDirectory)
		if err != nil {
			panic(err)
		}

		for _, directoryItem := range directoryContents {
			directoryItemPath := path.Join(currentDirectory, directoryItem.Name())
			if directoryItem.Name() == documentationDirectoryName {
				probableFormatChangeDirectoryContents, err := os.ReadDir(directoryItemPath)
				if err != nil {
					panic(err)
				}
				for _, possibleDocumentationItem := range probableFormatChangeDirectoryContents {
					if path.Ext(possibleDocumentationItem.Name()) == ".md" {
						// If the directory's name matches the expected, and contains .md files
						// we assume it is the documentation directory
						return probableFormatChangeDirectoryContents, true
					}
				}
			} else {
				if directoryItem.IsDir() {
					childContents, directoryFound := getDatabaseFormatDocumentation(directoryItemPath, false)
					if directoryFound {
						return childContents, true
					}
				}
			}
		}

		if allowDescend {
			// If not found in this directory, continue down the path
			currentDirectory = path.Dir(currentDirectory)

			if currentDirectory == "." || currentDirectory == "/" {
				panic("Database documentation directory not found")
			}
		} else {
			return []fs.DirEntry{}, false
		}
	}
}

func getLatestCommit(repoName string, branchName string) string {
	cmd := exec.Command("git", "ls-remote", repoName, "refs/heads/"+branchName)
	result, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	// This is a tab, not a space!
	seperator := "\t"
	return strings.Split(string(result), seperator)[0]
}
