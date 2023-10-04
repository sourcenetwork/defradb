// Copyright 2023 Democratized Data Foundation
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
	"strconv"
	"testing"
	"time"

	badger "github.com/dgraph-io/badger/v4"

	"github.com/sourcenetwork/defradb/client"
	badgerds "github.com/sourcenetwork/defradb/datastore/badger/v4"
	"github.com/sourcenetwork/defradb/datastore/memory"
	"github.com/sourcenetwork/defradb/db"
	"github.com/sourcenetwork/defradb/logging"
	"github.com/sourcenetwork/defradb/net"
	changeDetector "github.com/sourcenetwork/defradb/tests/change_detector"
	"github.com/sourcenetwork/defradb/tests/clients"
	"github.com/sourcenetwork/defradb/tests/clients/cli"
	"github.com/sourcenetwork/defradb/tests/clients/http"
)

const (
	clientGoEnvName       = "DEFRA_CLIENT_GO"
	clientHttpEnvName     = "DEFRA_CLIENT_HTTP"
	clientCliEnvName      = "DEFRA_CLIENT_CLI"
	memoryBadgerEnvName   = "DEFRA_BADGER_MEMORY"
	fileBadgerEnvName     = "DEFRA_BADGER_FILE"
	fileBadgerPathEnvName = "DEFRA_BADGER_FILE_PATH"
	inMemoryEnvName       = "DEFRA_IN_MEMORY"
	mutationTypeEnvName   = "DEFRA_MUTATION_TYPE"
)

type DatabaseType string

const (
	badgerIMType   DatabaseType = "badger-in-memory"
	defraIMType    DatabaseType = "defra-memory-datastore"
	badgerFileType DatabaseType = "badger-file-system"
)

type ClientType string

const (
	// goClientType enables running the test suite using
	// the go implementation of the client.DB interface.
	goClientType ClientType = "go"
	// httpClientType enables running the test suite using
	// the http implementation of the client.DB interface.
	httpClientType ClientType = "http"
	// cliClientType enables running the test suite using
	// the cli implementation of the client.DB interface.
	cliClientType ClientType = "cli"
)

// The MutationType that tests will run using.
//
// For example if set to [CollectionSaveMutationType], all supporting
// actions (such as [UpdateDoc]) will execute via [Collection.Save].
//
// Defaults to CollectionSaveMutationType.
type MutationType string

const (
	// CollectionSaveMutationType will cause all supporting actions
	// to run their mutations via [Collection.Save].
	CollectionSaveMutationType MutationType = "collection-save"

	// CollectionNamedMutationType will cause all supporting actions
	// to run their mutations via their corresponding named [Collection]
	// call.
	//
	// For example, CreateDoc will call [Collection.Create], and
	// UpdateDoc will call [Collection.Update].
	CollectionNamedMutationType MutationType = "collection-named"

	// GQLRequestMutationType will cause all supporting actions to
	// run their mutations using GQL requests, typically these will
	// include a `id` parameter to target the specified document.
	GQLRequestMutationType MutationType = "gql"
)

var (
	log            = logging.MustNewLogger("tests.integration")
	badgerInMemory bool
	badgerFile     bool
	inMemoryStore  bool
	httpClient     bool
	goClient       bool
	cliClient      bool
	mutationType   MutationType
	databaseDir    string
)

const (
	// subscriptionTimeout is the maximum time to wait for subscription results to be returned.
	subscriptionTimeout = 1 * time.Second
	// Instantiating lenses is expensive, and our tests do not benefit from a large number of them,
	// so we explicitly set it to a low value.
	lensPoolSize = 2
)

func init() {
	// We use environment variables instead of flags `go test ./...` throws for all packages
	//  that don't have the flag defined
	httpClient, _ = strconv.ParseBool(os.Getenv(clientHttpEnvName))
	goClient, _ = strconv.ParseBool(os.Getenv(clientGoEnvName))
	cliClient, _ = strconv.ParseBool(os.Getenv(clientCliEnvName))
	badgerFile, _ = strconv.ParseBool(os.Getenv(fileBadgerEnvName))
	badgerInMemory, _ = strconv.ParseBool(os.Getenv(memoryBadgerEnvName))
	inMemoryStore, _ = strconv.ParseBool(os.Getenv(inMemoryEnvName))

	if value, ok := os.LookupEnv(mutationTypeEnvName); ok {
		mutationType = MutationType(value)
	} else {
		// Default to testing mutations via Collection.Save - it should be simpler and
		// faster. We assume this is desirable when not explicitly testing any particular
		// mutation type.
		mutationType = CollectionSaveMutationType
	}

	if !goClient && !httpClient && !cliClient {
		// Default is to test go client type.
		goClient = true
	}

	if changeDetector.Enabled {
		// Change detector only uses badger file db type.
		badgerFile = true
		badgerInMemory = false
		inMemoryStore = false
	} else if !badgerInMemory && !badgerFile && !inMemoryStore {
		// Default is to test all but filesystem db types.
		badgerFile = false
		badgerInMemory = true
		inMemoryStore = true
	}
}

func NewBadgerMemoryDB(ctx context.Context, dbopts ...db.Option) (client.DB, error) {
	opts := badgerds.Options{
		Options: badger.DefaultOptions("").WithInMemory(true),
	}
	rootstore, err := badgerds.NewDatastore("", &opts)
	if err != nil {
		return nil, err
	}
	db, err := db.NewDB(ctx, rootstore, dbopts...)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func NewInMemoryDB(ctx context.Context, dbopts ...db.Option) (client.DB, error) {
	db, err := db.NewDB(ctx, memory.NewDatastore(ctx), dbopts...)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func NewBadgerFileDB(ctx context.Context, t testing.TB, dbopts ...db.Option) (client.DB, string, error) {
	var dbPath string
	switch {
	case databaseDir != "":
		// restarting database
		dbPath = databaseDir

	case changeDetector.Enabled:
		// change detector
		dbPath = changeDetector.DatabaseDir(t)

	default:
		// default test case
		dbPath = t.TempDir()
	}

	opts := &badgerds.Options{
		Options: badger.DefaultOptions(dbPath),
	}
	rootstore, err := badgerds.NewDatastore(dbPath, opts)
	if err != nil {
		return nil, "", err
	}
	db, err := db.NewDB(ctx, rootstore, dbopts...)
	if err != nil {
		return nil, "", err
	}
	return db, dbPath, err
}

// setupClient returns the client implementation for the current
// testing state. The client type on the test state is used to
// select the client implementation to use.
func setupClient(s *state, node *net.Node) (impl clients.Client, err error) {
	switch s.clientType {
	case httpClientType:
		impl, err = http.NewWrapper(node)

	case cliClientType:
		impl = cli.NewWrapper(node)

	case goClientType:
		impl = node

	default:
		err = fmt.Errorf("invalid client type: %v", s.dbt)
	}

	if err != nil {
		return nil, err
	}
	return
}

// setupDatabase returns the database implementation for the current
// testing state. The database type on the test state is used to
// select the datastore implementation to use.
func setupDatabase(s *state) (impl client.DB, path string, err error) {
	dbopts := []db.Option{
		db.WithUpdateEvents(),
		db.WithLensPoolSize(lensPoolSize),
	}

	switch s.dbt {
	case badgerIMType:
		impl, err = NewBadgerMemoryDB(s.ctx, dbopts...)

	case badgerFileType:
		impl, path, err = NewBadgerFileDB(s.ctx, s.t, dbopts...)

	case defraIMType:
		impl, err = NewInMemoryDB(s.ctx, dbopts...)

	default:
		err = fmt.Errorf("invalid database type: %v", s.dbt)
	}

	if err != nil {
		return nil, "", err
	}
	return
}
