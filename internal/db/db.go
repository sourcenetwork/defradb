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

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corelog"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp"
	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/db/permission"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/request/graphql"
	"github.com/sourcenetwork/defradb/internal/telemetry"
)

var (
	log    = corelog.NewLogger("db")
	tracer = telemetry.NewTracer()
)

// make sure we match our client interface
var (
	_ client.Collection = (*collection)(nil)
)

const (
	// commandBufferSize is the size of the channel buffer used to handle events.
	commandBufferSize = 100_000
	// eventBufferSize is the size of the channel buffer used to subscribe to events.
	eventBufferSize = 100
)

// DB is the main struct for DefraDB's storage layer.
type DB struct {
	glock sync.RWMutex

	rootstore  datastore.Rootstore
	multistore datastore.MultiStore

	events *event.Bus

	parser core.Parser

	lensRegistry client.LensRegistry

	// The maximum number of retries per transaction.
	maxTxnRetries immutable.Option[int]

	// The options used to init the database
	options []Option

	// The ID of the last transaction created.
	previousTxnID atomic.Uint64

	// The identity of the current node
	nodeIdentity immutable.Option[identity.Identity]

	// Contains ACP if it exists
	acp immutable.Option[acp.ACP]

	// The peer ID and network address information for the current node
	// if network is enabled. The `atomic.Value` should hold a `peer.AddrInfo` struct.
	peerInfo atomic.Value

	// To be able to close the context passed to NewDB on DB close,
	// we need to keep a reference to the cancel function. Otherwise,
	// some goroutines might leak.
	ctxCancel context.CancelFunc

	// The intervals at which to retry replicator failures.
	// For example, this can define an exponential backoff strategy.
	retryIntervals []time.Duration

	// The signature algorithm to use for DAG blocks.
	// If None, block signing is disabled.
	signingAlg immutable.Option[crypto.KeyType]
}

var _ client.DB = (*DB)(nil)

// NewDB creates a new instance of the DB using the given options.
func NewDB(
	ctx context.Context,
	rootstore datastore.Rootstore,
	acp immutable.Option[acp.ACP],
	lens client.LensRegistry,
	options ...Option,
) (*DB, error) {
	return newDB(ctx, rootstore, acp, lens, options...)
}

func newDB(
	ctx context.Context,
	rootstore datastore.Rootstore,
	acp immutable.Option[acp.ACP],
	lens client.LensRegistry,
	options ...Option,
) (*DB, error) {
	multistore := datastore.MultiStoreFrom(rootstore)

	parser, err := graphql.NewParser()
	if err != nil {
		return nil, err
	}

	opts := defaultOptions()
	for _, opt := range options {
		opt(opts)
	}

	ctx, cancel := context.WithCancel(ctx)

	db := &DB{
		rootstore:      rootstore,
		multistore:     multistore,
		acp:            acp,
		lensRegistry:   lens,
		parser:         parser,
		options:        options,
		events:         event.NewBus(commandBufferSize, eventBufferSize),
		ctxCancel:      cancel,
		retryIntervals: opts.RetryIntervals,
	}

	if opts.maxTxnRetries.HasValue() {
		db.maxTxnRetries = opts.maxTxnRetries
	}

	db.nodeIdentity = opts.identity
	db.signingAlg = opts.signingAlg

	if lens != nil {
		lens.Init(db)
	}

	err = db.initialize(ctx)
	if err != nil {
		return nil, err
	}

	sub, err := db.events.Subscribe(event.MergeName, event.PeerInfoName, event.ReplicatorFailureName)
	if err != nil {
		return nil, err
	}
	go db.handleMessages(ctx, sub)
	go db.handleReplicatorRetries(ctx)

	return db, nil
}

// NewTxn creates a new transaction.
func (db *DB) NewTxn(ctx context.Context, readonly bool) (datastore.Txn, error) {
	txnId := db.previousTxnID.Add(1)
	return datastore.NewTxnFrom(ctx, db.rootstore, txnId, readonly), nil
}

// NewConcurrentTxn creates a new transaction that supports concurrent API calls.
func (db *DB) NewConcurrentTxn(ctx context.Context, readonly bool) (datastore.Txn, error) {
	txnId := db.previousTxnID.Add(1)
	return datastore.NewConcurrentTxnFrom(ctx, db.rootstore, txnId, readonly), nil
}

// Rootstore returns the root datastore.
func (db *DB) Rootstore() datastore.Rootstore {
	return db.rootstore
}

// Blockstore returns the internal DAG store which contains IPLD blocks.
func (db *DB) Blockstore() datastore.Blockstore {
	return db.multistore.Blockstore()
}

// Encstore returns the internal enc store which contains encryption key for documents and their fields.
func (db *DB) Encstore() datastore.Blockstore {
	return db.multistore.Encstore()
}

// Peerstore returns the internal DAG store which contains IPLD blocks.
func (db *DB) Peerstore() datastore.DSReaderWriter {
	return db.multistore.Peerstore()
}

// Headstore returns the internal DAG store which contains IPLD blocks.
func (db *DB) Headstore() corekv.Reader {
	return db.multistore.Headstore()
}

func (db *DB) LensRegistry() client.LensRegistry {
	return db.lensRegistry
}

func (db *DB) AddPolicy(
	ctx context.Context,
	policy string,
) (client.AddPolicyResult, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	if !db.acp.HasValue() {
		return client.AddPolicyResult{}, client.ErrACPOperationButACPNotAvailable
	}

	policyID, err := db.acp.Value().AddPolicy(
		ctx,
		identity.FromContext(ctx).Value(),
		policy,
	)
	if err != nil {
		return client.AddPolicyResult{}, err
	}

	return client.AddPolicyResult{PolicyID: policyID}, nil
}

// publishDocUpdateEvent publishes an update event for a document.
// It uses heads iterator to read the document's head blocks directly from the storage, i.e. without
// using a transaction.
func (db *DB) publishDocUpdateEvent(ctx context.Context, docID string, collection client.Collection) error {
	headsIterator, err := NewHeadBlocksIterator(ctx, db.multistore.Headstore(), db.Blockstore(), docID)
	if err != nil {
		return err
	}

	for {
		hasValue, err := headsIterator.Next()
		if err != nil {
			return err
		}
		if !hasValue {
			break
		}

		updateEvent := event.Update{
			DocID:      docID,
			Cid:        headsIterator.CurrentCid(),
			SchemaRoot: collection.Schema().Root,
			Block:      headsIterator.CurrentRawBlock(),
		}
		db.events.Publish(event.NewMessage(event.UpdateName, updateEvent))
	}
	return nil
}

func (db *DB) AddDocActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
) (client.AddDocActorRelationshipResult, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	if !db.acp.HasValue() {
		return client.AddDocActorRelationshipResult{}, client.ErrACPOperationButACPNotAvailable
	}

	collection, err := db.GetCollectionByName(ctx, collectionName)
	if err != nil {
		return client.AddDocActorRelationshipResult{}, err
	}

	policyID, resourceName, hasPolicy := permission.IsPermissioned(collection)
	if !hasPolicy {
		return client.AddDocActorRelationshipResult{}, client.ErrACPOperationButCollectionHasNoPolicy
	}

	exists, err := db.acp.Value().AddDocActorRelationship(
		ctx,
		policyID,
		resourceName,
		docID,
		relation,
		identity.FromContext(ctx).Value(),
		targetActor,
	)

	if err != nil {
		return client.AddDocActorRelationshipResult{}, err
	}

	if !exists {
		err = db.publishDocUpdateEvent(ctx, docID, collection)
		if err != nil {
			return client.AddDocActorRelationshipResult{}, err
		}
	}

	return client.AddDocActorRelationshipResult{ExistedAlready: exists}, nil
}

func (db *DB) DeleteDocActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
) (client.DeleteDocActorRelationshipResult, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	if !db.acp.HasValue() {
		return client.DeleteDocActorRelationshipResult{}, client.ErrACPOperationButACPNotAvailable
	}

	collection, err := db.GetCollectionByName(ctx, collectionName)
	if err != nil {
		return client.DeleteDocActorRelationshipResult{}, err
	}

	policyID, resourceName, hasPolicy := permission.IsPermissioned(collection)
	if !hasPolicy {
		return client.DeleteDocActorRelationshipResult{}, client.ErrACPOperationButCollectionHasNoPolicy
	}

	recordFound, err := db.acp.Value().DeleteDocActorRelationship(
		ctx,
		policyID,
		resourceName,
		docID,
		relation,
		identity.FromContext(ctx).Value(),
		targetActor,
	)

	if err != nil {
		return client.DeleteDocActorRelationshipResult{}, err
	}

	return client.DeleteDocActorRelationshipResult{RecordFound: recordFound}, nil
}

func (db *DB) GetNodeIdentity(_ context.Context) (immutable.Option[identity.PublicRawIdentity], error) {
	if db.nodeIdentity.HasValue() {
		return immutable.Some(db.nodeIdentity.Value().IntoRawIdentity().Public()), nil
	}
	return immutable.None[identity.PublicRawIdentity](), nil
}

func (db *DB) GetNodeIdentityToken(_ context.Context, audience immutable.Option[string]) ([]byte, error) {
	if db.nodeIdentity.HasValue() {
		return db.nodeIdentity.Value().NewToken(time.Hour*24, audience, immutable.None[string]())
	}
	return nil, nil
}

// Initialize is called when a database is first run and creates all the db global meta data
// like Collection ID counters.
func (db *DB) initialize(ctx context.Context) error {
	db.glock.Lock()
	defer db.glock.Unlock()

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

	exists, err := txn.Systemstore().Has(ctx, []byte("/init"))
	if err != nil && !errors.Is(err, corekv.ErrNotFound) {
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
	_, err = db.getSequence(ctx, keys.CollectionIDSequenceKey{})
	if err != nil {
		return err
	}

	err = txn.Systemstore().Set(ctx, []byte("/init"), []byte{1})
	if err != nil {
		return err
	}

	return txn.Commit(ctx)
}

// Events returns the events Channel.
func (db *DB) Events() *event.Bus {
	return db.events
}

// MaxRetries returns the maximum number of retries per transaction.
// Defaults to `defaultMaxTxnRetries` if not explicitely set
func (db *DB) MaxTxnRetries() int {
	if db.maxTxnRetries.HasValue() {
		return db.maxTxnRetries.Value()
	}
	return defaultMaxTxnRetries
}

// PrintDump prints the entire database to console.
func (db *DB) PrintDump(ctx context.Context) error {
	return printStore(ctx, db.multistore.Rootstore())
}

// Close is called when we are shutting down the database.
// This is the place for any last minute cleanup or releasing of resources (i.e.: Badger instance).
func (db *DB) Close() {
	log.Info("Closing DefraDB process...")

	db.ctxCancel()

	db.events.Close()

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
	iter, err := store.Iterator(ctx, corekv.IterOptions{})
	if err != nil {
		return err
	}

	for {
		hasNext, err := iter.Next()
		if err != nil {
			return errors.Join(err, iter.Close())
		}

		if !hasNext {
			break
		}

		value, err := iter.Value()
		if err != nil {
			return errors.Join(err, iter.Close())
		}

		log.InfoContext(ctx, "", corelog.Any(string(iter.Key()), value))
	}

	return iter.Close()
}
