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
	"sync/atomic"
	"time"

	ds "github.com/ipfs/go-datastore"
	dsq "github.com/ipfs/go-datastore/query"

	"github.com/sourcenetwork/corelog"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/request/graphql"
)

var (
	log = corelog.NewLogger("db")
)

// make sure we match our client interface
var (
	_ client.Collection = (*collection)(nil)
)

const (
	// forwardBufferSize controls the size of the channel used to
	// forward events from the system bus to the subscription bus
	forwardBufferSize = 100
	// subscriptionBufferSize controls the size of the channel used to
	// send events to request subscriptions
	subscriptionBufferSize = 10
	// mergeBufferSize controls the size of the channel used to
	// handle merge events
	mergeBufferSize = 100
	// sysBusTimeout is the duration to wait before discarding
	// messages on the sysBus
	sysBusTimeout = 5 * time.Minute
	// subBusTimeout is the duration to wait before discarding
	// messages on the subBus
	subBusTimeout = 1 * time.Minute
)

// DB is the main interface for interacting with the
// DefraDB storage system.
type db struct {
	glock sync.RWMutex

	rootstore  datastore.RootStore
	multistore datastore.MultiStore

	// sysBus is used to send and receive system events
	sysBus *event.Bus

	// subBus is used to send and receive request subscription events
	subBus *event.Bus

	parser core.Parser

	lensRegistry client.LensRegistry

	// The maximum number of retries per transaction.
	maxTxnRetries immutable.Option[int]

	// The options used to init the database
	options []Option

	// The ID of the last transaction created.
	previousTxnID atomic.Uint64

	// Contains ACP if it exists
	acp immutable.Option[acp.ACP]
}

// NewDB creates a new instance of the DB using the given options.
func NewDB(
	ctx context.Context,
	rootstore datastore.RootStore,
	acp immutable.Option[acp.ACP],
	lens client.LensRegistry,
	options ...Option,
) (client.DB, error) {
	return newDB(ctx, rootstore, acp, lens, options...)
}

func newDB(
	ctx context.Context,
	rootstore datastore.RootStore,
	acp immutable.Option[acp.ACP],
	lens client.LensRegistry,
	options ...Option,
) (*db, error) {
	multistore := datastore.MultiStoreFrom(rootstore)

	parser, err := graphql.NewParser()
	if err != nil {
		return nil, err
	}

	db := &db{
		rootstore:    rootstore,
		multistore:   multistore,
		acp:          acp,
		lensRegistry: lens,
		parser:       parser,
		options:      options,
		sysBus:       event.NewBus(sysBusTimeout),
		subBus:       event.NewBus(subBusTimeout),
	}

	// apply options
	for _, opt := range options {
		opt(db)
	}

	if lens != nil {
		lens.Init(db)
	}

	err = db.initialize(ctx)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// NewTxn creates a new transaction.
func (db *db) NewTxn(ctx context.Context, readonly bool) (datastore.Txn, error) {
	txnId := db.previousTxnID.Add(1)
	return datastore.NewTxnFrom(ctx, db.rootstore, txnId, readonly)
}

// NewConcurrentTxn creates a new transaction that supports concurrent API calls.
func (db *db) NewConcurrentTxn(ctx context.Context, readonly bool) (datastore.Txn, error) {
	txnId := db.previousTxnID.Add(1)
	return datastore.NewConcurrentTxnFrom(ctx, db.rootstore, txnId, readonly)
}

// Root returns the root datastore.
func (db *db) Root() datastore.RootStore {
	return db.rootstore
}

// Blockstore returns the internal DAG store which contains IPLD blocks.
func (db *db) Blockstore() datastore.DAGStore {
	return db.multistore.DAGstore()
}

// Peerstore returns the internal DAG store which contains IPLD blocks.
func (db *db) Peerstore() datastore.DSBatching {
	return db.multistore.Peerstore()
}

// Headstore returns the internal DAG store which contains IPLD blocks.
func (db *db) Headstore() ds.Read {
	return db.multistore.Headstore()
}

func (db *db) LensRegistry() client.LensRegistry {
	return db.lensRegistry
}

func (db *db) AddPolicy(
	ctx context.Context,
	policy string,
) (client.AddPolicyResult, error) {
	if !db.acp.HasValue() {
		return client.AddPolicyResult{}, client.ErrPolicyAddFailureNoACP
	}
	identity := GetContextIdentity(ctx)

	policyID, err := db.acp.Value().AddPolicy(
		ctx,
		identity.Value().Address,
		policy,
	)
	if err != nil {
		return client.AddPolicyResult{}, err
	}

	return client.AddPolicyResult{PolicyID: policyID}, nil
}

// Initialize is called when a database is first run and creates all the db global meta data
// like Collection ID counters.
func (db *db) initialize(ctx context.Context) error {
	db.glock.Lock()
	defer db.glock.Unlock()

	// start merge process
	go db.handleMerges(ctx)

	// forward system bus events to the subscriber bus
	// to ensure we never block the system bus for user subscriptions
	go func() {
		sub := db.sysBus.Subscribe(forwardBufferSize, event.WildCardEventName)
		for msg := range sub.Message() {
			db.subBus.Publish(msg)
		}
	}()

	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	// Start acp if enabled, this will recover previous state if there is any.
	if db.acp.HasValue() {
		// db is responsible to call db.acp.Close() to free acp resources while closing.
		if err = db.acp.Value().Start(ctx); err != nil {
			return err
		}
	}

	exists, err := txn.Systemstore().Has(ctx, ds.NewKey("init"))
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return err
	}
	// if we're loading an existing database, just load the schema
	// and migrations and finish initialization
	if exists {
		err = db.loadSchema(ctx)
		if err != nil {
			return err
		}

		err = db.lensRegistry.ReloadLenses(ctx)
		if err != nil {
			return err
		}

		// The query language types are only updated on successful commit
		// so we must not forget to do so on success regardless of whether
		// we have written to the datastores.
		return txn.Commit(ctx)
	}

	// init meta data
	// collection sequence
	_, err = db.getSequence(ctx, core.CollectionIDSequenceKey{})
	if err != nil {
		return err
	}

	err = txn.Systemstore().Put(ctx, ds.NewKey("init"), []byte{1})
	if err != nil {
		return err
	}

	return txn.Commit(ctx)
}

// Events returns the events Channel.
func (db *db) Events() *event.Bus {
	return db.sysBus
}

// MaxRetries returns the maximum number of retries per transaction.
// Defaults to `defaultMaxTxnRetries` if not explicitely set
func (db *db) MaxTxnRetries() int {
	if db.maxTxnRetries.HasValue() {
		return db.maxTxnRetries.Value()
	}
	return defaultMaxTxnRetries
}

// PrintDump prints the entire database to console.
func (db *db) PrintDump(ctx context.Context) error {
	return printStore(ctx, db.multistore.Rootstore())
}

// Close is called when we are shutting down the database.
// This is the place for any last minute cleanup or releasing of resources (i.e.: Badger instance).
func (db *db) Close() {
	log.Info("Closing DefraDB process...")

	db.subBus.Close()
	db.sysBus.Close()

	err := db.rootstore.Close()
	if err != nil {
		log.ErrorE("Failure closing running process", err)
	}

	if db.acp.HasValue() {
		if err := db.acp.Value().Close(); err != nil {
			log.ErrorE("Failure closing acp", err)
		}
	}

	log.Info("Successfully closed running process")
}

func printStore(ctx context.Context, store datastore.DSReaderWriter) error {
	q := dsq.Query{
		Prefix:   "",
		KeysOnly: false,
		Orders:   []dsq.Order{dsq.OrderByKey{}},
	}

	results, err := store.Query(ctx, q)
	if err != nil {
		return err
	}

	for r := range results.Next() {
		log.InfoContext(ctx, "", corelog.Any(r.Key, r.Value))
	}

	return results.Close()
}
