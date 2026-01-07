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
	_ "github.com/sourcenetwork/corekv/chunk"
	"github.com/sourcenetwork/corelog"
	"github.com/sourcenetwork/immutable"
	lensNode "github.com/sourcenetwork/lens/host-go/node"
	lensStore "github.com/sourcenetwork/lens/host-go/store"

	"github.com/sourcenetwork/defradb/acp/dac"
	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/datastore"
	acpDB "github.com/sourcenetwork/defradb/internal/db/acp"
	"github.com/sourcenetwork/defradb/internal/db/p2p"
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

	// WARNING - This property should never be accessed directly, use `db.GetLensStore`
	// in order to ensure any transactions are respected.
	lensNode *lensNode.Node

	blockStoreChunkSize immutable.Option[int]

	// The maximum number of retries per transaction.
	maxTxnRetries immutable.Option[int]

	// The options used to init the database.
	options []Option

	// The ID of the last transaction created.
	previousTxnID atomic.Uint64

	// The identity of the current node.
	nodeIdentity immutable.Option[identity.Identity]

	// Node ACP system along with it's current state information.
	nodeACP acpDB.NACInfo

	// Contains document ACP if it exists.
	documentACP immutable.Option[dac.DocumentACP]

	// To be able to close the context passed to NewDB on DB close,
	// we need to keep a reference to the cancel function. Otherwise,
	// some goroutines might leak.
	ctxCancel context.CancelFunc

	// If true, block signing is disabled. By default, block signing is enabled.
	signingDisabled bool

	// The cryptographic key used to generate search tags for searchable encryption.
	searchableEncryptionKey []byte

	docMergeQueue *mergeQueue
	colMergeQueue *mergeQueue

	p2p *p2p.P2P
	// Retry intervals when a replicator failure occurs.
	retryIntervals []time.Duration
	// timeout duration for syncing block links.
	p2pBlockSyncTimeout time.Duration
}

var _ client.TxnStore = (*DB)(nil)

// NewDB creates a new instance of the DB using the given options.
func NewDB(
	ctx context.Context,
	rootstore corekv.TxnStore,
	nodeACP acpDB.NACInfo,
	documentACP immutable.Option[dac.DocumentACP],
	options ...Option,
) (*DB, error) {
	return newDB(ctx, rootstore, nodeACP, documentACP, options...)
}

func newDB(
	ctx context.Context,
	rootstore corekv.TxnStore,
	nodeACP acpDB.NACInfo,
	documentACP immutable.Option[dac.DocumentACP],
	options ...Option,
) (*DB, error) {
	opts := defaultDBOptions()
	for _, opt := range options {
		opt(opts)
	}

	parser, err := graphql.NewParser(len(opts.searchableEncryptionKey) > 0)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)

	db := &DB{
		rootstore:           rootstore,
		blockStoreChunkSize: opts.ChunkSize,
		nodeACP:             nodeACP,
		documentACP:         documentACP,
		parser:              parser,
		options:             options,
		events:              event.NewChannelBus(commandBufferSize, eventBufferSize),
		ctxCancel:           cancel,
		docMergeQueue:       newMergeQueue(),
		colMergeQueue:       newMergeQueue(),
		retryIntervals:      opts.retryIntervals,
		p2pBlockSyncTimeout: opts.p2pBlockSyncTimeout,
	}

	if opts.maxTxnRetries.HasValue() {
		db.maxTxnRetries = opts.maxTxnRetries
	}

	db.nodeIdentity = opts.identity
	db.signingDisabled = opts.disableSigning
	db.searchableEncryptionKey = opts.searchableEncryptionKey

	lensRuntime, err := newLensRuntime(opts.LensRuntimeType)
	if err != nil {
		return nil, err
	}

	// Overwrite a few key Lens options for now, by appending them to the end of the option
	// slice.
	opts.LensOptions = append(opts.LensOptions, lensNode.WithRootstore(rootstore))
	opts.LensOptions = append(opts.LensOptions, lensNode.WithTxnSource(wrapSource(db)))
	opts.LensOptions = append(opts.LensOptions, lensNode.WithRuntime(lensRuntime))

	if opts.p2p.HasValue() {
		opts.LensOptions = appendLensP2POpt(opts.LensOptions, opts)
	} else {
		// If defra has no P2P enabled, it doesn't make sense to enable it for Lens
		opts.LensOptions = append(opts.LensOptions, lensNode.WithP2PDisabled(true))
	}

	node, err := lensNode.New(ctx, opts.LensOptions...)
	if err != nil {
		return nil, err
	}
	db.lensNode = node

	if opts.p2p.HasValue() {
		p, err := p2p.New(ctx, db, node, opts.p2p.Value(), db.nodeIdentity, NewCollectionRetriever(db))
		if err != nil {
			return nil, err
		}
		db.p2p = p
	}

	err = db.initialize(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// NewTxn creates a new transaction.
func (db *DB) NewTxn(readonly bool) (client.Txn, error) {
	txnId := db.previousTxnID.Add(1)
	txn := datastore.NewTxnFrom(db.rootstore, txnId, readonly, db.blockStoreChunkSize)
	return wrapDatastoreTxn(txn, db), nil
}

// NewConcurrentTxn creates a new transaction that supports concurrent API calls.
func (db *DB) NewConcurrentTxn(readonly bool) (client.Txn, error) {
	txnId := db.previousTxnID.Add(1)
	txn := datastore.NewConcurrentTxnFrom(db.rootstore, txnId, readonly, db.blockStoreChunkSize)
	return wrapDatastoreTxn(txn, db), nil
}

// publishDocUpdateEvent publishes an update event for a document.
// It uses heads iterator to read the document's head blocks directly from the storage, i.e. without
// using a transaction.
func (db *DB) publishDocUpdateEvent(ctx context.Context, docID string, collection client.Collection) error {
	headsIterator, err := NewHeadBlocksIterator(
		ctx,
		datastore.HeadstoreFrom(db.rootstore),
		datastore.BlockstoreFrom(db.rootstore, db.blockStoreChunkSize),
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
		db.sendUpdate(updateEvent)
	}
	return nil
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
	defer txn.Discard()

	if err := db.initializeNodeACP(ctx, txn); err != nil {
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
	if err != nil {
		return err
	}
	// if we're loading an existing database, just load the schema
	// and migrations and finish initialization
	if exists {
		err = db.loadSchema(ctx)
		if err != nil {
			return err
		}

		err = db.getLensStore(ctx).Reload(ctx)
		if err != nil {
			return err
		}

		// The query language types are only updated on successful commit
		// so we must not forget to do so on success regardless of whether
		// we have written to the datastores.
		return txn.Commit()
	}

	err = txn.Systemstore().Set(ctx, []byte("/init"), []byte{1})
	if err != nil {
		return err
	}

	return txn.Commit()
}

func (db *DB) Rootstore() corekv.TxnStore {
	return db.rootstore
}

func (db *DB) Multistore() *datastore.Multistore {
	return datastore.NewMultistore(db.rootstore, db.blockStoreChunkSize)
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

// SearchableEncryptionKey returns the searchable encryption key if configured.
func (db *DB) SearchableEncryptionKey() []byte {
	return db.searchableEncryptionKey
}

// RetryIntervals returns the replicator retry configuration.
func (db *DB) RetryIntervals() []time.Duration {
	return db.retryIntervals
}

// P2PBlockSyncTimeout is the timeout duration for syncing block links.
func (db *DB) P2PBlockSyncTimeout() time.Duration {
	return db.p2pBlockSyncTimeout
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

	if db.nodeACP.NodeACP != nil {
		if err := db.nodeACP.NodeACP.Close(); err != nil {
			log.ErrorE("Failure closing node acp", err)
		}
	}

	if db.documentACP.HasValue() {
		if err := db.documentACP.Value().Close(); err != nil {
			log.ErrorE("Failure closing acp", err)
		}
	}

	if db.p2p != nil && db.p2p.SECoordinator() != nil {
		db.p2p.SECoordinator().Close()
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

		key, err := datastore.HumanReadableKey(iter.Key())
		if err != nil {
			return errors.Join(err, iter.Close())
		}

		log.InfoContext(ctx, "", corelog.Any(key, value))
	}

	return iter.Close()
}

type txnSource struct {
	txnSource client.TxnSource
}

var _ lensStore.TxnSource = (*txnSource)(nil)

func wrapSource(s client.TxnSource) *txnSource {
	return &txnSource{
		txnSource: s,
	}
}

func (s *txnSource) NewTxn(readOnly bool) (lensStore.Txn, error) {
	txn, err := s.txnSource.NewTxn(readOnly)
	if err != nil {
		return nil, err
	}

	dsTxn := datastore.MustGetFromClientTxn(txn)

	return &wrappedTxn{
		Txn:          dsTxn,
		ReaderWriter: dsTxn.Rootstore(),
	}, nil
}

type wrappedTxn struct {
	datastore.Txn
	corekv.ReaderWriter
}
