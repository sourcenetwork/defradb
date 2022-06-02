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
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/logging"
	"github.com/sourcenetwork/defradb/merkle/crdt"
	"github.com/sourcenetwork/defradb/request/graphql/planner"
	"github.com/sourcenetwork/defradb/request/graphql/schema"

	ds "github.com/ipfs/go-datastore"
	dsRequest "github.com/ipfs/go-datastore/query"
	request "github.com/ipfs/go-datastore/query"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	corenet "github.com/sourcenetwork/defradb/core/net"
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

	broadcaster corenet.Broadcaster

	schema          *schema.SchemaManager
	requestExecutor *planner.RequestExecutor

	// The options used to init the database
	options interface{}
}

// functional option type
type Option func(*db)

func WithBroadcaster(bs corenet.Broadcaster) Option {
	return func(db *db) {
		db.broadcaster = bs
	}
}

// NewDB creates a new instance of the DB using the given options
func NewDB(ctx context.Context, rootstore ds.Batching, options ...Option) (client.DB, error) {
	return newDB(ctx, rootstore, options...)
}

func newDB(ctx context.Context, rootstore ds.Batching, options ...Option) (*db, error) {
	log.Debug(ctx, "loading: internal datastores")
	root := datastore.AsDSReaderWriter(rootstore)
	multistore := datastore.MultiStoreFrom(root)
	crdtFactory := crdt.DefaultFactory.WithStores(multistore)

	log.Debug(ctx, "loading: schema manager")
	sm, err := schema.NewSchemaManager()
	if err != nil {
		return nil, err
	}

	log.Debug(ctx, "loading: request executor")
	exec, err := planner.NewRequestExecutor(sm)
	if err != nil {
		return nil, err
	}

	db := &db{
		rootstore:  rootstore,
		multistore: multistore,

		crdtFactory: &crdtFactory,

		schema:          sm,
		requestExecutor: exec,
		options:         options,
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

// Blockstore returns the internal DAG store which contains IPLD blocks
func (db *db) Blockstore() blockstore.Blockstore {
	return db.multistore.DAGstore()
}

func (db *db) systemstore() datastore.DSReaderWriter {
	return db.multistore.Systemstore()
}

// Initialize is called when a database is first run and creates all the db global meta data
// like Collection ID counters
func (db *db) initialize(ctx context.Context) error {
	db.glock.Lock()
	defer db.glock.Unlock()

	log.Debug(ctx, "Checking if db has already been initialized...")
	exists, err := db.systemstore().Has(ctx, ds.NewKey("init"))
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

func (db *db) PrintDump(ctx context.Context) {
	printStore(ctx, db.multistore.Rootstore())
}

func (db *db) Executor() *planner.RequestExecutor {
	return db.requestExecutor
}

func (db *db) GetRelationshipIdField(fieldName, targetType, thisType string) (string, error) {
	rm := db.schema.Relations
	rel := rm.GetRelationByDescription(fieldName, targetType, thisType)
	if rel == nil {
		return "", fmt.Errorf("Relation does not exists")
	}
	subtypefieldname, _, ok := rel.GetFieldFromSchemaType(targetType)
	if !ok {
		return "", fmt.Errorf("Relation is missing referenced field")
	}
	return subtypefieldname, nil
}

// Close is called when we are shutting down the database.
// This is the place for any last minute cleanup or releaseing
// of resources (IE: Badger instance)
func (db *db) Close(ctx context.Context) {
	log.Info(ctx, "Closing DefraDB process...")
	err := db.rootstore.Close()
	if err != nil {
		log.ErrorE(ctx, "Failure closing running process", err)
	}
	log.Info(ctx, "Successfully closed running process")
}

func printStore(ctx context.Context, store datastore.DSReaderWriter) {
	req := request.Query{
		Prefix:   "",
		KeysOnly: false,
		Orders:   []dsRequest.Order{dsRequest.OrderByKey{}},
	}

	results, err := store.Query(ctx, req)

	if err != nil {
		panic(err)
	}

	defer func() {
		err := results.Close()
		if err != nil {
			log.ErrorE(ctx, "Failure closing set of request (query store) results", err)
		}
	}()

	for r := range results.Next() {
		log.Info(ctx, "", logging.NewKV(r.Key, r.Value))
	}
}
