// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

/*
Package db provides the implementation of the [client.TxnStore] interface, collection operations,
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

	"github.com/sourcenetwork/defradb/acp/dac"
	"github.com/sourcenetwork/defradb/acp/identity"
	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/permission"
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

	rootstore corekv.TxnStore

	events event.Bus

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

	// Admin ACP system along with it's current state information.
	adminInfo AdminInfo

	// Contains document ACP if it exists
	documentACP immutable.Option[dac.DocumentACP]

	// To be able to close the context passed to NewDB on DB close,
	// we need to keep a reference to the cancel function. Otherwise,
	// some goroutines might leak.
	ctxCancel context.CancelFunc

	// If true, block signing is disabled. By default, block signing is enabled.
	signingDisabled bool
}

var _ client.TxnStore = (*DB)(nil)

// NewDB creates a new instance of the DB using the given options.
func NewDB(
	ctx context.Context,
	rootstore corekv.TxnStore,
	adminACP AdminInfo,
	documentACP immutable.Option[dac.DocumentACP],
	lens client.LensRegistry,
	options ...Option,
) (*DB, error) {
	return newDB(ctx, rootstore, adminACP, documentACP, lens, options...)
}

func newDB(
	ctx context.Context,
	rootstore corekv.TxnStore,
	adminACP AdminInfo,
	documentACP immutable.Option[dac.DocumentACP],
	lens client.LensRegistry,
	options ...Option,
) (*DB, error) {
	parser, err := graphql.NewParser()
	if err != nil {
		return nil, err
	}

	opts := &dbOptions{}
	for _, opt := range options {
		opt(opts)
	}

	ctx, cancel := context.WithCancel(ctx)

	db := &DB{
		rootstore:    rootstore,
		adminInfo:    adminACP,
		documentACP:  documentACP,
		lensRegistry: lens,
		parser:       parser,
		options:      options,
		events:       event.NewChannelBus(commandBufferSize, eventBufferSize),
		ctxCancel:    cancel,
	}

	if opts.maxTxnRetries.HasValue() {
		db.maxTxnRetries = opts.maxTxnRetries
	}

	db.nodeIdentity = opts.identity
	db.signingDisabled = opts.disableSigning

	if lens != nil {
		lens.Init(db)
	}

	err = db.initialize(ctx)
	if err != nil {
		return nil, err
	}

	sub, err := db.events.Subscribe(event.MergeName, event.PeerInfoName)
	if err != nil {
		return nil, err
	}
	go db.handleMessages(ctx, sub)

	return db, nil
}

// NewTxn creates a new transaction.
func (db *DB) NewTxn(ctx context.Context, readonly bool) (client.Txn, error) {
	txnId := db.previousTxnID.Add(1)
	txn := datastore.NewTxnFrom(ctx, db.rootstore, txnId, readonly)
	return wrapDatastoreTxn(txn, db), nil
}

// NewConcurrentTxn creates a new transaction that supports concurrent API calls.
func (db *DB) NewConcurrentTxn(ctx context.Context, readonly bool) (client.Txn, error) {
	txnId := db.previousTxnID.Add(1)
	txn := datastore.NewConcurrentTxnFrom(ctx, db.rootstore, txnId, readonly)
	return wrapDatastoreTxn(txn, db), nil
}

func (db *DB) LensRegistry() client.LensRegistry {
	return db.lensRegistry
}

func (db *DB) DocumentACP() immutable.Option[dac.DocumentACP] {
	return db.documentACP
}

func (db *DB) AddDACPolicy(
	ctx context.Context,
	policy string,
) (client.AddPolicyResult, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	if err := db.checkAdminAccess(ctx, acpTypes.AdminDACPolicyAddPerm); err != nil {
		return client.AddPolicyResult{}, err
	}

	if !db.documentACP.HasValue() {
		return client.AddPolicyResult{}, client.ErrACPOperationButACPNotAvailable
	}

	policyID, err := db.documentACP.Value().AddPolicy(
		ctx,
		identity.FromContext(ctx).Value(),
		policy,
	)
	if err != nil {
		return client.AddPolicyResult{}, err
	}

	return client.AddPolicyResult{PolicyID: policyID}, nil
}

// PurgeDACState purges all document ACP state, and calls [Close()] on the acp instance before returning.
//
// This will close the acp system, reset it's state (purge then restart), and finally close it.
//
// Note: all document ACP state will be lost, and won't be recoverable.
func (db *DB) PurgeDACState(ctx context.Context) error {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	// Purge document acp state and keep it closed.
	if db.documentACP.HasValue() {
		documentACP := db.documentACP.Value()
		err := documentACP.ResetState(ctx)
		if err != nil {
			// for now we will just log this error, since SourceHub ACP doesn't yet
			// implement the ResetState.
			log.ErrorE("Failed to reset document ACP state", err)
		}
	}

	return nil
}

// publishDocUpdateEvent publishes an update event for a document.
// It uses heads iterator to read the document's head blocks directly from the storage, i.e. without
// using a transaction.
func (db *DB) publishDocUpdateEvent(ctx context.Context, docID string, collection client.Collection) error {
	headsIterator, err := NewHeadBlocksIterator(
		ctx,
		datastore.HeadstoreFrom(db.rootstore),
		datastore.BlockstoreFrom(db.rootstore),
		docID,
	)
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
			DocID:        docID,
			Cid:          headsIterator.CurrentCid(),
			CollectionID: collection.Version().CollectionID,
			Block:        headsIterator.CurrentRawBlock(),
		}
		db.events.Publish(event.NewMessage(event.UpdateName, updateEvent))
	}
	return nil
}

func (db *DB) AddDACActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
) (client.AddActorRelationshipResult, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	if err := db.checkAdminAccess(ctx, acpTypes.AdminDACRelationAddPerm); err != nil {
		return client.AddActorRelationshipResult{}, err
	}

	if !db.documentACP.HasValue() {
		return client.AddActorRelationshipResult{}, client.ErrACPOperationButACPNotAvailable
	}

	collection, err := db.GetCollectionByName(ctx, collectionName)
	if err != nil {
		return client.AddActorRelationshipResult{}, err
	}

	policyID, resourceName, hasPolicy := permission.IsPermissioned(collection)
	if !hasPolicy {
		return client.AddActorRelationshipResult{}, client.ErrACPOperationButCollectionHasNoPolicy
	}

	exists, err := db.documentACP.Value().AddDocActorRelationship(
		ctx,
		policyID,
		resourceName,
		docID,
		relation,
		identity.FromContext(ctx).Value(),
		targetActor,
	)

	if err != nil {
		return client.AddActorRelationshipResult{}, err
	}

	if !exists {
		err = db.publishDocUpdateEvent(ctx, docID, collection)
		if err != nil {
			return client.AddActorRelationshipResult{}, err
		}
	}

	return client.AddActorRelationshipResult{ExistedAlready: exists}, nil
}

func (db *DB) DeleteDACActorRelationship(
	ctx context.Context,
	collectionName string,
	docID string,
	relation string,
	targetActor string,
) (client.DeleteActorRelationshipResult, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

	if err := db.checkAdminAccess(ctx, acpTypes.AdminDACRelationDeletePerm); err != nil {
		return client.DeleteActorRelationshipResult{}, err
	}

	if !db.documentACP.HasValue() {
		return client.DeleteActorRelationshipResult{}, client.ErrACPOperationButACPNotAvailable
	}

	collection, err := db.GetCollectionByName(ctx, collectionName)
	if err != nil {
		return client.DeleteActorRelationshipResult{}, err
	}

	policyID, resourceName, hasPolicy := permission.IsPermissioned(collection)
	if !hasPolicy {
		return client.DeleteActorRelationshipResult{}, client.ErrACPOperationButCollectionHasNoPolicy
	}

	recordFound, err := db.documentACP.Value().DeleteDocActorRelationship(
		ctx,
		policyID,
		resourceName,
		docID,
		relation,
		identity.FromContext(ctx).Value(),
		targetActor,
	)

	if err != nil {
		return client.DeleteActorRelationshipResult{}, err
	}

	return client.DeleteActorRelationshipResult{RecordFound: recordFound}, nil
}

func (db *DB) GetNodeIdentity(_ context.Context) (immutable.Option[identity.PublicRawIdentity], error) {
	if db.nodeIdentity.HasValue() {
		return immutable.Some(db.nodeIdentity.Value().ToPublicRawIdentity()), nil
	}
	return immutable.None[identity.PublicRawIdentity](), nil
}

func (db *DB) GetNodeIdentityToken(_ context.Context, audience immutable.Option[string]) ([]byte, error) {
	if !db.nodeIdentity.HasValue() {
		return nil, nil
	}

	ident := db.nodeIdentity.Value()
	fullIdentity, ok := ident.(identity.FullIdentity)
	if !ok || fullIdentity.PrivateKey() == nil {
		return nil, identity.ErrPrivateKeyNotAvailable
	}

	return fullIdentity.NewToken(time.Hour*24, audience, immutable.None[string]())
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

	if err := db.initializeAdminACP(ctx, txn); err != nil {
		return err
	}

	// Start document acp if enabled, this will recover previous state if there is any.
	if db.documentACP.HasValue() {
		// db is responsible to call db.documentACP.Close() to free acp resources while closing.
		if err = db.documentACP.Value().Start(ctx); err != nil {
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

	err = txn.Systemstore().Set(ctx, []byte("/init"), []byte{1})
	if err != nil {
		return err
	}

	return txn.Commit(ctx)
}

func (db *DB) Rootstore() corekv.TxnStore {
	return db.rootstore
}

// Events returns the events Channel.
func (db *DB) Events() event.Bus {
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
	return printStore(ctx, db.rootstore)
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

	if db.adminInfo.AdminACP != nil {
		if err := db.adminInfo.AdminACP.Close(); err != nil {
			log.ErrorE("Failure closing admin acp", err)
		}
	}

	if db.documentACP.HasValue() {
		if err := db.documentACP.Value().Close(); err != nil {
			log.ErrorE("Failure closing acp", err)
		}
	}

	log.Info("Successfully closed running process")
}

func printStore(ctx context.Context, store corekv.ReaderWriter) error {
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
