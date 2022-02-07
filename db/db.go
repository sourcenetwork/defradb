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
	"fmt"
	"sync"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/merkle/crdt"
	"github.com/sourcenetwork/defradb/query/graphql/planner"
	"github.com/sourcenetwork/defradb/query/graphql/schema"
	"github.com/sourcenetwork/defradb/store"

	ds "github.com/ipfs/go-datastore"
	ktds "github.com/ipfs/go-datastore/keytransform"
	"github.com/ipfs/go-datastore/namespace"
	"github.com/ipfs/go-datastore/query"
	dsq "github.com/ipfs/go-datastore/query"
	logging "github.com/ipfs/go-log/v2"
)

var (
	// ErrDocVerification occurs when a documents contents fail the verification during a Create()
	// call against the supplied Document Key
	ErrDocVerification = errors.New("The document verificatioin failed")

	ErrOptionsEmpty = errors.New("Empty options configuration provided")

	log = logging.Logger("defra.db")
)

// make sure we match our client interface
var (
	_ client.DB         = (*DB)(nil)
	_ client.Collection = (*Collection)(nil)
)

// DB is the main interface for interacting with the
// DefraDB storage system.
//
type DB struct {
	glock sync.RWMutex

	rootstore ds.Batching // main storage interface

	systemstore    core.DSReaderWriter // wrapped store for system data
	ssKeyTransform ktds.KeyTransform

	datastore      core.DSReaderWriter // wrapped store for data
	dsKeyTransform ktds.KeyTransform

	headstore      core.DSReaderWriter // wrapped store for heads
	hsKeyTransform ktds.KeyTransform

	dagstore        core.DAGStore // wrapped store for dags
	dagKeyTransform ktds.KeyTransform

	crdtFactory *crdt.Factory

	schema        *schema.SchemaManager
	queryExecutor *planner.QueryExecutor

	// indicates if this references an initalized database
	initialized bool

	log logging.StandardLogger

	// The options used to init the database
	options interface{}
}

// NewDB creates a new instance of the DB using the given options
func NewDB(rootstore ds.Batching, options interface{}) (*DB, error) {
	log.Debug("loading: internal datastores")
	systemstore := namespace.Wrap(rootstore, base.SystemStoreKey)
	datastore := namespace.Wrap(rootstore, base.DataStoreKey)
	headstore := namespace.Wrap(rootstore, base.HeadStoreKey)
	blockstore := namespace.Wrap(rootstore, base.BlockStoreKey)
	dagstore := store.NewDAGStore(blockstore)
	crdtFactory := crdt.DefaultFactory.WithStores(datastore, headstore, dagstore)

	log.Debug("loading: schema manager")
	sm, err := schema.NewSchemaManager()
	if err != nil {
		return nil, err
	}

	log.Debug("loading: query executor")
	exec, err := planner.NewQueryExecutor(sm)
	if err != nil {
		return nil, err
	}

	db := &DB{
		rootstore: rootstore,

		systemstore:    systemstore,
		ssKeyTransform: systemstore.KeyTransform,

		datastore:      datastore,
		dsKeyTransform: datastore.KeyTransform,

		headstore:      headstore,
		hsKeyTransform: headstore.KeyTransform,

		dagstore:        dagstore,
		dagKeyTransform: blockstore.KeyTransform,

		crdtFactory: &crdtFactory,
		log:         log,

		schema:        sm,
		queryExecutor: exec,
		options:       options,
	}

	return db, err
}

// Start runs all the initial sub-routines and initialization steps.
func (db *DB) Start(ctx context.Context) error {
	return db.Initialize(ctx)
}

// Initialize is called when a database is first run and creates all the db global meta data
// like Collection ID counters
func (db *DB) Initialize(ctx context.Context) error {
	db.glock.Lock()
	defer db.glock.Unlock()

	// if its already initialized, just load the schema and we're done
	if db.initialized {
		return nil
	}

	log.Debug("Checking if db has already been initialized...")
	exists, err := db.systemstore.Has(ctx, ds.NewKey("init"))
	if err != nil && err != ds.ErrNotFound {
		return err
	}
	// if we're loading an existing database, just load the schema
	// and finish intialization
	if exists {
		log.Debug("db has already been initalized, conitnuing.")
		return db.loadSchema(ctx)
	}

	log.Debug("opened a new db, needs full intialization")
	// init meta data
	// collection sequence
	_, err = db.getSequence(ctx, "collection")
	if err != nil {
		return err
	}

	err = db.systemstore.Put(ctx, ds.NewKey("init"), []byte{1})
	if err != nil {
		return err
	}

	db.initialized = true
	return nil
}

func (db *DB) printDebugDB(ctx context.Context) {
	printStore(ctx, db.rootstore)
}

func (db *DB) PrintDump(ctx context.Context) {
	printStore(ctx, db.rootstore)
}

func (db *DB) Executor() *planner.QueryExecutor {
	return db.queryExecutor
}

// Close is called when we are shutting down the database.
// This is the place for any last minute cleanup or releaseing
// of resources (IE: Badger instance)
func (db *DB) Close() {
	log.Info("Closing DefraDB process...")
	err := db.rootstore.Close()
	if err != nil {
		log.Error("Failure closing running process")
	}
	log.Info("Succesfully closed running process")
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
			log.Error("Failure closing set of query store results")
		}
	}()

	for r := range results.Next() {
		fmt.Println(r.Key, ": ", r.Value)
	}
}
