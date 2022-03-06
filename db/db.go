// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package db

import (
	"context"
	"errors"
	"sync"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	corenet "github.com/sourcenetwork/defradb/core/net"
	"github.com/sourcenetwork/defradb/merkle/crdt"
	"github.com/sourcenetwork/defradb/query/graphql/planner"
	"github.com/sourcenetwork/defradb/query/graphql/schema"
	"github.com/sourcenetwork/defradb/store"

	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	dsq "github.com/ipfs/go-datastore/query"
	"github.com/sourcenetwork/defradb/logging"
)

var (
	log = logging.MustNewLogger("defra.db")
	// ErrDocVerification occurs when a documents contents fail the verification during a Create()
	// call against the supplied Document Key
	ErrDocVerification = errors.New("The document verification failed")

	ErrOptionsEmpty = errors.New("Empty options configuration provided")
)

// make sure we match our client interface
var (
	_ client.DB         = (*DB)(nil)
	_ client.Collection = (*Collection)(nil)
)

// DB is the main interface for interacting with the
// DefraDB storage system.
type DB struct {
	glock sync.RWMutex

	rootstore  ds.Batching
	multistore core.MultiStore

	crdtFactory *crdt.Factory

	broadcaster corenet.Broadcaster

	schema        *schema.SchemaManager
	queryExecutor *planner.QueryExecutor

	// The options used to init the database
	options interface{}
}

// functional option type
type Option func(*DB)

func WithBroadcaster(bs corenet.Broadcaster) Option {
	return func(db *DB) {
		db.broadcaster = bs
	}
}

// NewDB creates a new instance of the DB using the given options
func NewDB(ctx context.Context, rootstore ds.Batching, options ...Option) (*DB, error) {
	log.Debug(ctx, "loading: internal datastores")
	root := store.AsDSReaderWriter(rootstore)
	multistore := store.MultiStoreFrom(root)
	crdtFactory := crdt.DefaultFactory.WithStores(multistore)

	log.Debug(ctx, "loading: schema manager")
	sm, err := schema.NewSchemaManager()
	if err != nil {
		return nil, err
	}

	log.Debug(ctx, "loading: query executor")
	exec, err := planner.NewQueryExecutor(sm)
	if err != nil {
		return nil, err
	}

	db := &DB{
		rootstore:  rootstore,
		multistore: multistore,

		crdtFactory: &crdtFactory,

		schema:        sm,
		queryExecutor: exec,
		options:       options,
	}

	// apply options
	for _, opt := range options {
		if opt == nil {
			continue
		}
		opt(db)
	}

	err = db.initialize(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (db *DB) NewTxn(ctx context.Context, readonly bool) (core.Txn, error) {
	return store.NewTxnFrom(ctx, db.rootstore, readonly)
}

func (db *DB) Root() ds.Batching {
	return db.rootstore
}

// Rootstore gets the internal rootstore handle
func (db *DB) Rootstore() core.DSReaderWriter {
	return db.multistore.Rootstore()
}

// Headstore returns the internal index store for DAG Heads
func (db *DB) Headstore() core.DSReaderWriter {
	return db.multistore.Headstore()
}

// Datastore returns the internal index store for DAG Heads
func (db *DB) Datastore() core.DSReaderWriter {
	return db.multistore.Datastore()
}

// DAGstore returns the internal DAG store which contains IPLD blocks
func (db *DB) DAGstore() core.DAGStore {
	return db.multistore.DAGstore()
}

func (db *DB) Systemstore() core.DSReaderWriter {
	return db.multistore.Systemstore()
}

// Initialize is called when a database is first run and creates all the db global meta data
// like Collection ID counters
func (db *DB) initialize(ctx context.Context) error {
	db.glock.Lock()
	defer db.glock.Unlock()

	log.Debug(ctx, "Checking if db has already been initialized...")
	exists, err := db.Systemstore().Has(ctx, ds.NewKey("init"))
	if err != nil && err != ds.ErrNotFound {
		return err
	}
	// if we're loading an existing database, just load the schema
	// and finish initialization
	if exists {
		log.Debug(ctx, "DB has already been initialized, continuing.")
		return db.loadSchema(ctx)
	}

	log.Debug(ctx, "Opened a new DB, needs full initialization")
	// init meta data
	// collection sequence
	_, err = db.getSequence(ctx, "collection")
	if err != nil {
		return err
	}

	err = db.Systemstore().Put(ctx, ds.NewKey("init"), []byte{1})
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) printDebugDB(ctx context.Context) {
	printStore(ctx, db.Rootstore())
}

func (db *DB) PrintDump(ctx context.Context) {
	printStore(ctx, db.Rootstore())
}

func (db *DB) Executor() *planner.QueryExecutor {
	return db.queryExecutor
}

// Close is called when we are shutting down the database.
// This is the place for any last minute cleanup or releaseing
// of resources (IE: Badger instance)
func (db *DB) Close(ctx context.Context) {
	log.Info(ctx, "Closing DefraDB process...")
	err := db.rootstore.Close()
	if err != nil {
		log.ErrorE(ctx, "Failure closing running process", err)
	}
	log.Info(ctx, "Successfully closed running process")
}

func printStore(ctx context.Context, store core.DSReaderWriter) {
	q := query.Query{
		Prefix:   "",
		KeysOnly: false,
		Orders:   []dsq.Order{dsq.OrderByKey{}},
	}

	results, err := store.Query(ctx, q)

	if err != nil {
		panic(err)
	}

	defer func() {
		err := results.Close()
		if err != nil {
			log.ErrorE(ctx, "Failure closing set of query store results", err)
		}
	}()

	for r := range results.Next() {
		log.Info(ctx, "", logging.NewKV(r.Key, r.Value))
	}
}
