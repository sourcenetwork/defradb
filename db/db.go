// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

/*
Package db provides the implementation of the [client.DB] interface, collection operations,
and related components.
*/
package db

import (
	"context"
	"sync"

	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	dsq "github.com/ipfs/go-datastore/query"
	blockstore "github.com/ipfs/go-ipfs-blockstore"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/events"
	"github.com/sourcenetwork/defradb/logging"
	"github.com/sourcenetwork/defradb/merkle/crdt"
	"github.com/sourcenetwork/defradb/planner"
	"github.com/sourcenetwork/defradb/query/graphql/schema"
)

var (
	log = logging.MustNewLogger("defra.db")
	// ErrDocVerification occurs when a documents contents fail the verification during a Create()
	// call against the supplied Document Key.
	ErrDocVerification = errors.New("The document verification failed")

	ErrOptionsEmpty = errors.New("Empty options configuration provided")
)

// make sure we match our client interface
var (
	_ client.DB         = (*db)(nil)
	_ client.Collection = (*collection)(nil)
)

// DB is the main interface for interacting with the
// DefraDB storage system.
type db struct {
	glock sync.RWMutex

	rootstore  ds.Batching
	multistore datastore.MultiStore

	crdtFactory *crdt.Factory

	events client.Events

	schema        *schema.SchemaManager
	queryExecutor *planner.QueryExecutor

	// The options used to init the database
	options any
}

// functional option type
type Option func(*db)

const updateEventBufferSize = 100

func WithUpdateEvents() Option {
	return func(db *db) {
		db.events = client.Events{
			Updates: client.Some(events.New[client.UpdateEvent](0, updateEventBufferSize)),
		}
	}
}

// NewDB creates a new instance of the DB using the given options.
func NewDB(ctx context.Context, rootstore ds.Batching, options ...Option) (client.DB, error) {
	return newDB(ctx, rootstore, options...)
}

func newDB(ctx context.Context, rootstore ds.Batching, options ...Option) (*db, error) {
	log.Debug(ctx, "Loading: internal datastores")
	root := datastore.AsDSReaderWriter(rootstore)
	multistore := datastore.MultiStoreFrom(root)
	crdtFactory := crdt.DefaultFactory.WithStores(multistore)

	log.Debug(ctx, "Loading: schema manager")
	sm, err := schema.NewSchemaManager()
	if err != nil {
		return nil, err
	}

	log.Debug(ctx, "Loading: query executor")
	exec, err := planner.NewQueryExecutor(sm)
	if err != nil {
		return nil, err
	}

	db := &db{
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

func (db *db) NewTxn(ctx context.Context, readonly bool) (datastore.Txn, error) {
	return datastore.NewTxnFrom(ctx, db.rootstore, readonly)
}

func (db *db) Root() ds.Batching {
	return db.rootstore
}

// Blockstore returns the internal DAG store which contains IPLD blocks.
func (db *db) Blockstore() blockstore.Blockstore {
	return db.multistore.DAGstore()
}

func (db *db) systemstore() datastore.DSReaderWriter {
	return db.multistore.Systemstore()
}

// Initialize is called when a database is first run and creates all the db global meta data
// like Collection ID counters.
func (db *db) initialize(ctx context.Context) error {
	db.glock.Lock()
	defer db.glock.Unlock()

	log.Debug(ctx, "Checking if DB has already been initialized...")
	exists, err := db.systemstore().Has(ctx, ds.NewKey("init"))
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return err
	}
	// if we're loading an existing database, just load the schema
	// and finish initialization
	if exists {
		log.Debug(ctx, "DB has already been initialized, continuing")
		return db.loadSchema(ctx)
	}

	log.Debug(ctx, "Opened a new DB, needs full initialization")
	// init meta data
	// collection sequence
	_, err = db.getSequence(ctx, core.COLLECTION)
	if err != nil {
		return err
	}

	err = db.systemstore().Put(ctx, ds.NewKey("init"), []byte{1})
	if err != nil {
		return err
	}

	return nil
}

func (db *db) Events() client.Events {
	return db.events
}

func (db *db) PrintDump(ctx context.Context) error {
	return printStore(ctx, db.multistore.Rootstore())
}

func (db *db) Executor() *planner.QueryExecutor {
	return db.queryExecutor
}

func (db *db) GetRelationshipIdField(fieldName, targetType, thisType string) (string, error) {
	rm := db.schema.Relations
	rel := rm.GetRelationByDescription(fieldName, targetType, thisType)
	if rel == nil {
		return "", errors.New("Relation does not exists")
	}
	subtypefieldname, _, ok := rel.GetFieldFromSchemaType(targetType)
	if !ok {
		return "", errors.New("Relation is missing referenced field")
	}
	return subtypefieldname, nil
}

// Close is called when we are shutting down the database.
// This is the place for any last minute cleanup or releasing of resources (i.e.: Badger instance).
func (db *db) Close(ctx context.Context) {
	log.Info(ctx, "Closing DefraDB process...")
	if db.events.Updates.HasValue() {
		db.events.Updates.Value().Close()
	}

	err := db.rootstore.Close()
	if err != nil {
		log.ErrorE(ctx, "Failure closing running process", err)
	}
	log.Info(ctx, "Successfully closed running process")
}

func printStore(ctx context.Context, store datastore.DSReaderWriter) error {
	q := query.Query{
		Prefix:   "",
		KeysOnly: false,
		Orders:   []dsq.Order{dsq.OrderByKey{}},
	}

	results, err := store.Query(ctx, q)
	if err != nil {
		return err
	}

	for r := range results.Next() {
		log.Info(ctx, "", logging.NewKV(r.Key, r.Value))
	}

	return results.Close()
}
