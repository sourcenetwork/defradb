// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package p2p

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/fxamacker/cbor/v2"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"

	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/corelog"
	"github.com/sourcenetwork/immutable"
	lens "github.com/sourcenetwork/lens/host-go/node"

	"github.com/sourcenetwork/defradb/acp/dac"
	"github.com/sourcenetwork/defradb/acp/identity"
	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/datastore"
	acpDB "github.com/sourcenetwork/defradb/internal/db/acp"
	"github.com/sourcenetwork/defradb/internal/db/p2p/protocol"
	"github.com/sourcenetwork/defradb/internal/kms"
	"github.com/sourcenetwork/defradb/internal/se"
	"github.com/sourcenetwork/defradb/internal/telemetry"
)

var (
	log    = corelog.NewLogger("p2p")
	tracer = telemetry.NewTracer()
)

type (
	peerID        = string
	collectionID  = string
	addresses     = []string
	peerAddresses = map[peerID]addresses
)

const (
	networkRequestTimeout = 10 * time.Second
	// pubsubBatchSize is the number of pubsub messages to batch before publishing.
	pubsubBatchSize = 50
	// pubsubBatchTimeout is the maximum time to wait for a pubsub batch to fill.
	pubsubBatchTimeout = 200 * time.Millisecond
	// pubsubBatchWorkers is the number of parallel pubsub batch processors.
	pubsubBatchWorkers = 16
	// maxPubsubMessageSize is the maximum size of a pubsub message (libp2p default is 1MB).
	maxPubsubMessageSize = 1000 * 1024
	// syncWorkerCount is the number of workers processing sync requests.
	syncWorkerCount = 16
	// syncQueueSize is the maximum number of pending sync requests.
	syncQueueSize = 500000
	// batchSyncSize is the number of sync operations to batch before committing.
	batchSyncSize = 100
	// batchSyncTimeout is the maximum time to wait for a sync batch to fill.
	batchSyncTimeout = 50 * time.Millisecond
	// batchSyncWorkers is the number of parallel sync batch processors.
	batchSyncWorkers = 16
)

// PushToReplicatorsHandler is called when documents are pushed to replicators.
// Implementations can perform additional actions like generating SE artifacts.
type PushToReplicatorsHandler interface {
	HandlePushToReplicators(ctx context.Context, evt event.Update) error
}

// DB hold the database related methods that are required by P2P.
type DB interface {
	// NewTxn returns a new transaction on the root store that may be managed externally.
	NewTxn(readOnly bool) (client.Txn, error)
	// WrapCorekvTxn wraps an existing corekv.Txn into a client.Txn.
	WrapCorekvTxn(txn corekv.Txn) client.Txn
	// InitContext returns a new context with all caches initialized and linked to the given transaction.
	InitContext(ctx context.Context, txn client.Txn) context.Context
	// GetNodeIdentity returns the current node identity.
	GetNodeIdentity(ctx context.Context) (immutable.Option[identity.PublicRawIdentity], error)
	// GetNodeIdentityToken returns an identity token for the given audience.
	GetNodeIdentityToken(ctx context.Context, audience immutable.Option[string]) ([]byte, error)
	// GetCollections returns all collections and their descriptions matching the given options
	// that currently exist within this [Store].
	GetCollections(ctx context.Context, options client.CollectionFetchOptions) ([]client.Collection, error)
	// Merge initiates a merge of the DAG and caches the resulting values into the datastore.
	Merge(ctx context.Context, evt event.Merge) error
	// MergeBatch merges multiple events in a single transaction for improved performance.
	MergeBatch(ctx context.Context, evts []event.Merge) error
	// MergeBatchWithTxn merges multiple events using an existing transaction from context.
	MergeBatchWithTxn(ctx context.Context, evts []event.Merge) error
	// Events returns the event bus for the database.
	Events() event.Bus
	// RetryIntervals returns the replicator retry configuration.
	RetryIntervals() []time.Duration
	// NodeACP returns the NodeACP implementation configured on the database.
	NodeACP() acpDB.NACInfo
	// DocumentACP returns the DocumentACP implementation configured on the database.
	DocumentACP() immutable.Option[dac.DocumentACP]
	// Rootstore returns the rootstore
	Rootstore() corekv.TxnStore
	// Multistore returns the multistore
	Multistore() *datastore.Multistore
	// P2PBlockSyncTimeout is the timeout duration for syncing block links.
	P2PBlockSyncTimeout() time.Duration
	// SearchableEncryptionKey returns the searchable encryption key if configured.
	SearchableEncryptionKey() []byte
	// MaxTxnRetries returns the maximum number of transaction retries.
	MaxTxnRetries() int
}

type P2P struct {
	identityProtocol   *protocol.IdentityProtocol
	replicatorProtocol protocol.CommChannel[protocol.PushLogRequest, protocol.PushLogReply]

	ctx  context.Context
	db   DB
	lens *lens.Node
	host client.Host
	kms  kms.Service

	// replicators is a map from collection CollectionID => peerId => list of addresses.
	// This is a cached in-memory copy of the persisted replicators in the database.
	// It is used to quickly find the replicators for a given collection when sending updates.
	// The map is protected by repMu.
	replicators map[collectionID]peerAddresses
	repMu       sync.Mutex

	peerIdentities map[peerID]identity.Identity
	piMu           sync.RWMutex

	// The intervals at which to retry replicator failures.
	// For example, this can define an exponential backoff strategy.
	retryIntervals   []time.Duration
	handleRetryMutex sync.Mutex

	// processQueue is a worker pool for processing sync requests.
	processQueue *processQueue

	// syncBatcher combines CAR import and merge in single transactions.
	syncBatcher *syncBatcher

	// pubsubBatcher batches pubsub publish operations for improved throughput.
	pubsubBatcher *pubsubBatcher

	// timeout duration for syncing block links.
	syncBlockLinkTimeout time.Duration

	// seCoordinator manages searchable encryption artifact replication
	seCoordinator *se.Coordinator

	// pushHandlers are called when documents are pushed to replicators
	pushHandlers []PushToReplicatorsHandler

	// collectionCache caches collection existence checks to reduce DB queries.
	collectionCache   map[string]*cachedCollection
	collectionCacheMu sync.RWMutex
}

// cachedCollection stores a cached collection lookup result with expiry.
type cachedCollection struct {
	exists  bool
	col     client.Collection
	expires time.Time
}

const (
	// collectionCacheTTL is how long to cache collection existence checks.
	collectionCacheTTL = 30 * time.Second
)

// pushLogCommProcessor implements CommProcessor for push log functionality
type pushLogCommProcessor struct {
	p2p *P2P
}

func (proc *pushLogCommProcessor) ProcessRequest(
	ctx context.Context,
	req protocol.PushLogRequest,
) (protocol.PushLogReply, error) {
	return protocol.PushLogReply{}, proc.p2p.processPushlogRequest(ctx, &req, true)
}

// peerEventHandlingHost wraps a Host to add a PeerEventHandler to pubsub topics.
// It's added so that KMS doesn't need to bother with event handling and keeps it independent
// from the event bus.
type peerEventHandlingHost struct {
	client.Host
	eventHandler client.PeerEventHandler
}

func (h *peerEventHandlingHost) AddPubSubTopic(
	topicName string,
	subscribe bool,
	handler client.PubsubMessageHandler,
) error {
	return h.Host.AddPubSubTopic(topicName, subscribe, handler, h.eventHandler)
}

// New returns a new configured P2P instance.
func New(
	ctx context.Context,
	db DB,
	lens *lens.Node,
	host client.Host,
	nodeIdentity immutable.Option[identity.Identity],
	collectionRetriever kms.CollectionRetriever,
) (*P2P, error) {
	p := P2P{
		ctx:                  ctx,
		db:                   db,
		lens:                 lens,
		host:                 host,
		identityProtocol:     protocol.NewIdentityProtocol(host, db.GetNodeIdentityToken),
		replicators:          make(map[string]map[string][]string),
		peerIdentities:       make(map[string]identity.Identity),
		retryIntervals:       db.RetryIntervals(),
		processQueue:         newProcessQueue(syncWorkerCount, syncQueueSize),
		syncBlockLinkTimeout: db.P2PBlockSyncTimeout(),
		collectionCache:      make(map[string]*cachedCollection),
	}
	p.syncBatcher = newSyncBatcher(ctx, &p)
	p.pubsubBatcher = newPubsubBatcher(ctx, &p, pubsubBatchSize, pubsubBatchTimeout)
	p.replicatorProtocol = protocol.NewCommChannel(host, "rep", &pushLogCommProcessor{p2p: &p})

	host.SetBlockAccessFunc(p.hasAccess)

	err := p.host.AddPubSubTopic(docSyncTopic, true, p.docSyncMessageHandler, p.peerEventHandler)
	if err != nil {
		return nil, err
	}

	err = p.host.AddPubSubTopic(syncBranchableCollectionTopic, true, p.syncBranchableCollectionMessageHandler,
		p.peerEventHandler)
	if err != nil {
		return nil, err
	}

	go p.handleReplicatorRetries(ctx)
	err = p.loadAndPublishReplicators(ctx)
	if err != nil {
		return nil, err
	}
	err = p.loadAndPublishP2PCollections(ctx)
	if err != nil {
		return nil, err
	}
	err = p.loadAndPublishP2PDocuments(ctx)
	if err != nil {
		return nil, err
	}

	if nodeIdentity.HasValue() {
		p.kms, err = kms.NewPubSubService(
			ctx,
			host.ID(),
			&peerEventHandlingHost{
				Host:         host,
				eventHandler: p.peerEventHandler,
			},
			datastore.EncstoreFrom(db.Rootstore()),
			db.NodeACP(),
			db.DocumentACP(),
			collectionRetriever,
			nodeIdentity.Value().DID(),
		)
		if err != nil {
			return nil, err
		}
	}

	if len(db.SearchableEncryptionKey()) > 0 {
		coord, err := se.NewCoordinator(&p, host, db, db.SearchableEncryptionKey(), nodeIdentity)
		if err != nil {
			return nil, err
		}
		p.seCoordinator = coord
		p.AddPushToReplicatorsHandler(coord)
	}

	return &p, nil
}

// Close shuts down the P2P system gracefully.
func (p *P2P) Close() {
	p.processQueue.close()
	p.syncBatcher.close()
	p.pubsubBatcher.close()
}

func (p *P2P) KMS() kms.Service {
	return p.kms
}

func (p *P2P) SECoordinator() *se.Coordinator {
	return p.seCoordinator
}

// AddPushToReplicatorsHandler registers a handler that will be called when documents are pushed to replicators.
func (p *P2P) AddPushToReplicatorsHandler(handler PushToReplicatorsHandler) {
	p.pushHandlers = append(p.pushHandlers, handler)
}

func (p *P2P) PeerInfo() ([]string, error) {
	return p.host.Addresses()
}

func (p *P2P) ActivePeers(ctx context.Context) ([]string, error) {
	return p.host.ActivePeers()
}

// Connect initiates a connection to the peer with the given addresses.
func (p *P2P) Connect(ctx context.Context, addresses []string) error {
	return p.host.Connect(ctx, addresses)
}

func (p *P2P) updateReplicators(ctx context.Context, id string, addresses []string, collectionIDs map[string]struct{}) {
	if len(collectionIDs) == 0 {
		// remove peer from store
		if err := p.host.Disconnect(ctx, id); err != nil {
			log.ErrorE("Failed to disconnect from replicator peer", err)
		}
	} else {
		if err := p.host.Connect(ctx, addresses); err != nil {
			log.ErrorE("Failed to connect to replicator peer", err, corelog.Any("Addresses", addresses))
		}
	}

	// update the cached replicators
	p.repMu.Lock()
	for collectionID, peers := range p.replicators {
		if _, hasID := collectionIDs[collectionID]; hasID {
			p.replicators[collectionID][id] = addresses
			delete(collectionIDs, collectionID)
		} else {
			if _, exists := peers[id]; exists {
				delete(p.replicators[collectionID], id)
			}
		}
	}
	for collectionID := range collectionIDs {
		if _, exists := p.replicators[collectionID]; !exists {
			p.replicators[collectionID] = make(map[string][]string)
		}
		p.replicators[collectionID][id] = addresses
	}
	p.repMu.Unlock()
}

// hasAccess checks if the requesting peer has access to the given cid.
//
// This is used as a filter in bitswap to determine if we should send the block to the requesting peer.
func (p *P2P) hasAccess(ctx context.Context, pid string, c cid.Cid) bool {
	if !p.db.DocumentACP().HasValue() {
		return true
	}

	rawblock, err := p.db.Multistore().Blockstore().Get(ctx, c)
	if err != nil {
		if !ipld.IsNotFound(err) {
			log.ErrorE("Failed to get block", err)
		}
		return false
	}

	_, err = coreblock.GetSignatureBlockFromBytes(rawblock.RawData())
	if err == nil {
		// If the block is a signature block, we can safely send it to the requesting peer.
		return true
	}

	block, err := coreblock.GetFromBytes(rawblock.RawData())
	if err != nil {
		if strings.Contains(err.Error(), "invalid key: \"modules\" is not a field in type Block") ||
			strings.Contains(err.Error(), "invalid key: \"lens\" is not a field in type Block") ||
			strings.Contains(err.Error(), "invalid key: \"wasmBytes\" is not a field in type Block") ||
			strings.Contains(err.Error(), "invalid key: \"chunks\" is not a field in type Block") {
			// There are currently 3 kinds of Lens blocks that may be synced, these three error checks
			// are for those blocks.  If the block is a Lens block, we can safely send it to the
			// requesting peer.
			return true
		}
		log.ErrorE("Failed to get doc from block", err)
		return false
	}

	if block.Delta.IsDefinition() {
		return true
	}

	cols, err := p.db.GetCollections(
		ctx,
		client.CollectionFetchOptions{
			VersionID: immutable.Some(block.Delta.GetCollectionVersionID()),
		},
	)
	if err != nil {
		log.ErrorE("Failed to get collections", err)
		return false
	}
	if len(cols) == 0 {
		log.Info("No collections found",
			corelog.Any("Collection Version ID", block.Delta.GetCollectionVersionID()))
		return false
	}

	// If the requesting peer is in the replicators list for that collection, then they have accesp.
	p.repMu.Lock()
	if peerList, ok := p.replicators[cols[0].CollectionID()]; ok {
		_, exists := peerList[pid]
		if exists {
			p.repMu.Unlock()
			return true
		}
	}
	p.repMu.Unlock()

	identFunc := func() immutable.Option[identity.Identity] {
		p.piMu.RLock()
		ident, ok := p.peerIdentities[pid]
		p.piMu.RUnlock()
		if !ok {
			ctx, cancel := context.WithTimeout(ctx, networkRequestTimeout)
			defer cancel()
			resp, err := p.identityProtocol.GetIdentity(ctx, pid)
			if err != nil {
				log.ErrorE("Failed to get identity", err)
				return immutable.None[identity.Identity]()
			}
			ident, err = identity.FromToken(resp.IdentityToken)
			if err != nil {
				log.ErrorE("Failed to parse identity token", err)
				return immutable.None[identity.Identity]()
			}
			tokenIdent, ok := ident.(identity.TokenIdentity)
			if !ok {
				log.ErrorE("Identity is not of type TokenIdentity", nil, corelog.String("Actual", fmt.Sprintf("%T", ident)))
				return immutable.None[identity.Identity]()
			}
			err = identity.VerifyAuthToken(tokenIdent, p.host.ID())
			if err != nil {
				log.ErrorE("Failed to verify auth token", err)
				return immutable.None[identity.Identity]()
			}
			p.piMu.Lock()
			p.peerIdentities[pid] = ident
			p.piMu.Unlock()
		}
		return immutable.Some(ident)
	}

	peerHasAccess, err := acpDB.CheckDocAccessWithIdentityFunc(
		ctx,
		identFunc,
		p.db.NodeACP(),
		p.db.DocumentACP().Value(),
		cols[0], // For now we assume there is only one collection.
		acpTypes.DocumentReadPerm,
		string(block.Delta.GetDocID()),
	)
	if err != nil {
		log.ErrorE("Failed to check access", err)
		return false
	}

	return peerHasAccess
}

// hasCollection checks if the local node has the given collection.
func (p *P2P) hasCollection(ctx context.Context, collectionID string) (bool, error) {
	p.collectionCacheMu.RLock()
	if cached, ok := p.collectionCache[collectionID]; ok && time.Now().Before(cached.expires) {
		p.collectionCacheMu.RUnlock()
		return cached.exists, nil
	}
	p.collectionCacheMu.RUnlock()

	cols, err := p.db.GetCollections(
		ctx,
		client.CollectionFetchOptions{
			CollectionID: immutable.Some(collectionID),
		},
	)
	if err != nil {
		return false, err
	}

	exists := len(cols) > 0
	var col client.Collection
	if exists {
		col = cols[0]
	}

	p.collectionCacheMu.Lock()
	p.collectionCache[collectionID] = &cachedCollection{
		exists:  exists,
		col:     col,
		expires: time.Now().Add(collectionCacheTTL),
	}
	p.collectionCacheMu.Unlock()

	return exists, nil
}

// getCachedCollection returns a cached collection if available and not expired.
func (p *P2P) getCachedCollection(collectionID string) client.Collection {
	p.collectionCacheMu.RLock()
	defer p.collectionCacheMu.RUnlock()

	if cached, ok := p.collectionCache[collectionID]; ok && time.Now().Before(cached.expires) && cached.exists {
		return cached.col
	}
	return nil
}

// trySelfHasAccess checks if the local node has access to the given block.
//
// This is a best-effort check and returns true unless we explicitly find that the local node
// doesn't have access or if we get an error. The node sending is ultimately responsible for
// ensuring that the recipient has access.
func (p *P2P) trySelfHasAccess(ctx context.Context, block *coreblock.Block, collectionID string) (bool, error) {
	if !p.db.DocumentACP().HasValue() {
		return true, nil
	}

	col := p.getCachedCollection(collectionID)
	if col == nil {
		cols, err := p.db.GetCollections(
			ctx,
			client.CollectionFetchOptions{
				CollectionID: immutable.Some(collectionID),
			},
		)
		if err != nil {
			return false, err
		}
		if len(cols) == 0 {
			return false, nil
		}
		col = cols[0]
		p.collectionCacheMu.Lock()
		p.collectionCache[collectionID] = &cachedCollection{
			exists:  true,
			col:     col,
			expires: time.Now().Add(collectionCacheTTL),
		}
		p.collectionCacheMu.Unlock()
	}

	ident, err := p.db.GetNodeIdentity(ctx)
	if err != nil {
		return false, err
	}
	if !ident.HasValue() {
		return true, nil
	}

	peerHasAccess, err := acpDB.CheckDocAccessWithIdentityFunc(
		ctx,
		func() immutable.Option[identity.Identity] {
			return immutable.Some(identity.FromDID(ident.Value().DID))
		},
		p.db.NodeACP(),
		p.db.DocumentACP().Value(),
		col,
		acpTypes.DocumentReadPerm,
		string(block.Delta.GetDocID()),
	)
	if err != nil {
		return false, err
	}

	return peerHasAccess, nil
}

// pubSubMessageHandler handles incoming PushLog messages from the pubsub network.
func (p *P2P) pubSubMessageHandler(from string, topic string, msg []byte) ([]byte, error) {
	log.Info("Received new pubsub message",
		corelog.String("PeerID", p.host.ID()),
		corelog.Any("SenderId", from),
		corelog.String("Topic", topic))

	req := &protocol.PushLogRequest{}
	if err := cbor.Unmarshal(msg, req); err != nil {
		return nil, err
	}
	req.SenderID = from

	if req.IsBatched() {
		return nil, p.processBatchedPushlogRequest(p.ctx, req)
	}

	if err := p.processPushlogRequest(p.ctx, req, false); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			log.Info("Context done during pushlog request processing", corelog.Any("Error", err))
			return nil, nil
		}
		return nil, errors.Wrap(fmt.Sprintf("Failed to process pushlog request %s", topic), err)
	}

	return nil, nil
}

// processBatchedPushlogRequest processes a batched push log request containing multiple documents.
func (p *P2P) processBatchedPushlogRequest(ctx context.Context, req *protocol.PushLogRequest) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	log.Info("Processing batched pushlog request",
		corelog.Int("DocumentCount", len(req.Documents)),
		corelog.String("CollectionID", req.CollectionID))

	hasCollection, err := p.hasCollection(ctx, req.CollectionID)
	if err != nil {
		return err
	}
	if !hasCollection {
		return nil
	}

	if len(req.CAR) > 0 {
		if _, err := p.importCAR(ctx, req.CAR); err != nil {
			log.ErrorE("Failed to import batched CAR", err)
		}
	}

	var lastErr error
	for _, doc := range req.Documents {
		docReq := &protocol.PushLogRequest{
			DocID:        doc.DocID,
			CID:          doc.CID,
			CollectionID: req.CollectionID,
			Creator:      req.Creator,
			Block:        doc.Block,
		}
		docReq.SenderID = req.SenderID

		if err := p.processPushlogRequest(ctx, docReq, false); err != nil {
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				log.Info("Context done during batched pushlog request processing", corelog.Any("Error", err))
				return err
			}
			log.ErrorE("Failed to process document in batch", err, corelog.String("DocID", doc.DocID))
			lastErr = err
		}
	}

	return lastErr
}

func (p *P2P) peerEventHandler(peerID string, topic string, eventType string) {
	p.db.Events().Publish(event.NewMessage(event.TopicPeerEventName, event.TopicPeerEvent{
		PeerID:    peerID,
		Topic:     topic,
		EventType: eventType,
	}))
}

// processPushlogRequest processes a push log request
func (p *P2P) processPushlogRequest(
	ctx context.Context,
	req *protocol.PushLogRequest,
	isReplicator bool,
) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	block, err := coreblock.GetFromBytes(req.Block)
	if err != nil {
		return err
	}

	headCID, err := cid.Cast(req.CID)
	if err != nil {
		return err
	}

	// Calls to syncDAG should not overlap for a given CID. If they do, they will use the same
	// underlying pubsub topic and this brings along potential pitfalls. One of them being that
	// if this initial sync call had a negative response for a given link, the subsequent calls will
	// assume a negative response for that same link without retrying.
	// For branchable collections, use collectionID as the sync key to serialize collection-level syncs.
	syncKey := headCID.String()
	if isBranchableSync(req.CollectionID, req.DocID) {
		syncKey = req.CollectionID
	}

	processSync := func() error {
		txn := p.db.Rootstore().NewTxn(true)
		defer txn.Discard()
		txnCtx := corekv.SetCtxTxn(ctx, txn)

		isMerged, err := p.db.Multistore().Blockstore().IsMerged(txnCtx, headCID)
		if err != nil {
			return err
		}
		if isMerged {
			return nil
		}

		// Check if we have the collection
		hasCollection, err := p.hasCollection(txnCtx, req.CollectionID)
		if err != nil {
			return err
		}
		if !hasCollection {
			return nil
		}

		// No need to check access if the message is for replication as the node sending
		// will have done so deliberately.
		if !isReplicator {
			mightHaveAccess, err := p.trySelfHasAccess(txnCtx, block, req.CollectionID)
			if err != nil {
				return err
			}
			if !mightHaveAccess {
				return nil
			}
		}

		// For non-CAR syncs, we still need to sync the DAG first
		if len(req.CAR) == 0 {
			hasBlock, err := p.db.Multistore().Blockstore().Has(ctx, headCID)
			if err != nil {
				return err
			}
			if !hasBlock {
				err = p.syncDAG(ctx, block)
				if err != nil {
					return err
				}
			}
		}

		mergeEvt := event.Merge{
			DocID:        req.DocID,
			ByPeer:       req.SenderID,
			FromPeer:     req.Creator,
			Cid:          headCID,
			CollectionID: req.CollectionID,
		}

		p.syncBatcher.enqueue(req.CAR, mergeEvt, func(updateEvt event.Update) {
			updateEvt.Block = req.Block
			p.db.Events().Publish(event.NewMessage(event.UpdateName, updateEvt))
			if err := p.SendUpdate(updateEvt); err != nil {
				log.ErrorE("Failed to send update after sync", err, slog.Any("PeerID", p.host.ID()))
			}
		})
		return nil
	}

	if isReplicator {
		err = p.processQueue.enqueue(syncKey, processSync)
		if err == ErrSyncDuplicate || err == ErrSyncQueueFull {
			return nil
		}
		return err
	}

	p.processQueue.tryEnqueue(syncKey, processSync)
	return nil
}

func (p *P2P) SendUpdate(evt event.Update) error {
	// push to each peer (replicator)
	p.pushLogToReplicators(evt)

	// Retries are for replicators only and should not pollute the pubsub network.
	if !evt.IsRetry {
		p.pubsubBatcher.enqueue(evt)
	}

	return nil
}

// syncRequest represents a sync operation to be processed by workers.
type syncRequest struct {
	key     string
	process func() error
	result  chan error
}

// processQueue is a worker pool that processes sync operations.
type processQueue struct {
	inFlight map[string]struct{}
	mu       sync.Mutex
	queue    chan syncRequest
	wg       sync.WaitGroup
	closed   bool
}

func newProcessQueue(workerCount, queueSize int) *processQueue {
	pq := &processQueue{
		inFlight: make(map[string]struct{}),
		queue:    make(chan syncRequest, queueSize),
	}
	for range workerCount {
		pq.wg.Add(1)
		go pq.worker()
	}
	return pq
}

func (pq *processQueue) worker() {
	defer pq.wg.Done()
	for req := range pq.queue {
		err := req.process()
		pq.mu.Lock()
		delete(pq.inFlight, req.key)
		pq.mu.Unlock()
		if req.result != nil {
			req.result <- err
		} else if err != nil {
			log.ErrorE("Async sync request failed", err)
		}
	}
}

// doEnqueue adds a sync request to the queue if the key is not already in-flight.
func (pq *processQueue) doEnqueue(key string, process func() error, resultCh chan error) error {
	pq.mu.Lock()
	if pq.closed {
		pq.mu.Unlock()
		return ErrSyncQueueClosed
	}
	if _, exists := pq.inFlight[key]; exists {
		pq.mu.Unlock()
		return ErrSyncDuplicate
	}
	pq.inFlight[key] = struct{}{}
	pq.mu.Unlock()
	req := syncRequest{
		key:     key,
		process: process,
		result:  resultCh,
	}
	select {
	case pq.queue <- req:
		if resultCh != nil {
			return <-resultCh
		}
		return nil
	default:
		pq.mu.Lock()
		delete(pq.inFlight, key)
		pq.mu.Unlock()
		return ErrSyncQueueFull
	}
}

// enqueue adds a sync request to the queue and waits for the result.
func (pq *processQueue) enqueue(key string, process func() error) error {
	resultCh := make(chan error, 1)
	return pq.doEnqueue(key, process, resultCh)
}

// tryEnqueue adds a sync request to the queue without waiting for the result.
func (pq *processQueue) tryEnqueue(key string, process func() error) bool {
	return pq.doEnqueue(key, process, nil) == nil
}

// close shuts down the worker pool gracefully.
func (pq *processQueue) close() {
	pq.mu.Lock()
	pq.closed = true
	pq.mu.Unlock()
	close(pq.queue)
	pq.wg.Wait()
}

// QueryDocIDsWithSETags queries SE artifacts from replicators based on field values.
func (p *P2P) QueryDocIDsWithSETags(
	ctx context.Context,
	collectionID string,
	fieldValues []se.FieldValueQuery,
) ([]string, error) {
	if p.seCoordinator == nil {
		return []string{}, nil
	}

	return p.seCoordinator.QueryDocIDsByValues(ctx, collectionID, fieldValues)
}

// isBranchableSync returns true if this is a branchable collection sync operation.
func isBranchableSync(collectionID, docID string) bool {
	return collectionID != "" && docID == ""
}

// syncRequestBatch represents a combined CAR import and merge operation.
type syncRequestBatch struct {
	carData  []byte
	parsed   *parsedCAR
	evt      event.Merge
	updateFn func(event.Update)
	resultCh chan error
}

// syncBatcher combines CAR import and merge operations into single transactions.
type syncBatcher struct {
	ctx       context.Context
	cancel    context.CancelFunc
	p2p       *P2P
	queue     chan syncRequestBatch
	wg        sync.WaitGroup
	batchSize int
	timeout   time.Duration
}

func newSyncBatcher(ctx context.Context, p2p *P2P) *syncBatcher {
	ctx, cancel := context.WithCancel(ctx)
	sb := &syncBatcher{
		ctx:       ctx,
		cancel:    cancel,
		p2p:       p2p,
		queue:     make(chan syncRequestBatch, batchSyncSize*batchSyncWorkers*100),
		batchSize: batchSyncSize,
		timeout:   batchSyncTimeout,
	}
	for i := 0; i < batchSyncWorkers; i++ {
		sb.wg.Add(1)
		go sb.processBatches()
	}
	return sb
}

// enqueue adds a sync request to the batch queue.
func (sb *syncBatcher) enqueue(carData []byte, evt event.Merge, updateFn func(event.Update)) <-chan error {
	resultCh := make(chan error, 1)
	select {
	case sb.queue <- syncRequestBatch{
		carData:  carData,
		evt:      evt,
		updateFn: updateFn,
		resultCh: resultCh,
	}:
	case <-sb.ctx.Done():
		resultCh <- sb.ctx.Err()
	}
	return resultCh
}

func (sb *syncBatcher) processBatches() {
	defer sb.wg.Done()

	batch := make([]syncRequestBatch, 0, sb.batchSize)
	timer := time.NewTimer(sb.timeout)
	timer.Stop()
	defer timer.Stop()

	processBatch := func() {
		if len(batch) == 0 {
			return
		}

		var allRegularBlocks []blocks.Block
		var allEncBlocks []blocks.Block

		for i := range batch {
			if len(batch[i].carData) > 0 {
				parsed, err := parseCAR(batch[i].carData)
				if err != nil {
					batch[i].resultCh <- err
					batch[i].parsed = nil
					continue
				}
				batch[i].parsed = parsed
				allRegularBlocks = append(allRegularBlocks, parsed.regularBlocks...)
				allEncBlocks = append(allEncBlocks, parsed.encBlocks...)
			}
		}

		txn := sb.p2p.db.Rootstore().NewTxn(false)
		defer txn.Discard()

		txnCtx := corekv.SetCtxTxn(sb.ctx, txn)

		if len(allRegularBlocks) > 0 {
			bstore := datastore.P2PBlockstoreFrom(sb.p2p.db.Rootstore(), immutable.None[int]())
			if err := bstore.PutMany(txnCtx, allRegularBlocks); err != nil {
				log.ErrorE("Batch CAR import failed", err)
				for _, req := range batch {
					if req.parsed != nil {
						req.resultCh <- err
					}
				}
				batch = batch[:0]
				return
			}
		}

		if len(allEncBlocks) > 0 {
			encStore := datastore.EncstoreFrom(sb.p2p.db.Rootstore())
			if err := encStore.PutMany(txnCtx, allEncBlocks); err != nil {
				log.ErrorE("Batch encrypted block import failed", err)
				for _, req := range batch {
					if req.parsed != nil {
						req.resultCh <- err
					}
				}
				batch = batch[:0]
				return
			}
		}

		var mergeEvts []event.Merge
		var validRequests []syncRequestBatch
		for _, req := range batch {
			if len(req.carData) == 0 || req.parsed != nil {
				mergeEvts = append(mergeEvts, req.evt)
				validRequests = append(validRequests, req)
			}
		}

		if len(mergeEvts) > 0 {
			wrappedTxn := sb.p2p.db.WrapCorekvTxn(txn)
			wrappedCtx := sb.p2p.db.InitContext(txnCtx, wrappedTxn)
			err := sb.p2p.db.MergeBatchWithTxn(wrappedCtx, mergeEvts)
			if err != nil {
				log.ErrorE("Batch merge with txn failed, falling back to individual processing", err)
				for _, req := range validRequests {
					req.resultCh <- err
				}
				batch = batch[:0]
				return
			}
		}

		if err := txn.Commit(); err != nil {
			log.ErrorE("Batch sync commit failed", err)
			for _, req := range validRequests {
				req.resultCh <- err
			}
			batch = batch[:0]
			return
		}

		for _, req := range validRequests {
			req.resultCh <- nil
			if req.updateFn != nil {
				updateEvt := event.Update{
					DocID:        req.evt.DocID,
					Cid:          req.evt.Cid,
					CollectionID: req.evt.CollectionID,
					IsRelay:      true,
				}
				req.updateFn(updateEvt)
			}
		}

		batch = batch[:0]
	}

	for {
		select {
		case <-sb.ctx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			processBatch()
			return

		case req := <-sb.queue:
			batch = append(batch, req)

			if len(batch) == 1 {
				timer.Reset(sb.timeout)
			}

			if len(batch) >= sb.batchSize {
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				processBatch()
			}

		case <-timer.C:
			processBatch()
		}
	}
}

func (sb *syncBatcher) close() {
	sb.cancel()
	sb.wg.Wait()
}

// pubsubRequest represents an update event to be published via pubsub.
type pubsubRequest struct {
	evt event.Update
}

// pubsubBatcher batches pubsub publish operations for improved throughput.
type pubsubBatcher struct {
	ctx       context.Context
	cancel    context.CancelFunc
	p2p       *P2P
	queue     chan pubsubRequest
	wg        sync.WaitGroup
	batchSize int
	timeout   time.Duration
}

func newPubsubBatcher(ctx context.Context, p2p *P2P, batchSize int, timeout time.Duration) *pubsubBatcher {
	ctx, cancel := context.WithCancel(ctx)
	pb := &pubsubBatcher{
		ctx:       ctx,
		cancel:    cancel,
		p2p:       p2p,
		queue:     make(chan pubsubRequest, batchSize*pubsubBatchWorkers*100),
		batchSize: batchSize,
		timeout:   timeout,
	}
	for i := 0; i < pubsubBatchWorkers; i++ {
		pb.wg.Add(1)
		go pb.processBatches()
	}
	return pb
}

// enqueue adds an update event to be published asynchronously.
func (pb *pubsubBatcher) enqueue(evt event.Update) {
	select {
	case pb.queue <- pubsubRequest{evt: evt}:
	case <-pb.ctx.Done():
	default:
		log.Info("Pubsub queue full, dropping event", corelog.String("DocID", evt.DocID))
	}
}

func (pb *pubsubBatcher) processBatches() {
	defer pb.wg.Done()

	batch := make([]pubsubRequest, 0, pb.batchSize)
	timer := time.NewTimer(pb.timeout)
	timer.Stop()
	defer timer.Stop()

	processBatch := func() {
		if len(batch) == 0 {
			return
		}

		activePeers, err := pb.p2p.host.ActivePeers()
		if err != nil || len(activePeers) == 0 {
			pb.p2p.db.Events().Publish(event.NewMessage(event.P2PNoPeersName, event.P2PNoPeers{
				DroppedBatchSize: len(batch),
			}))
			batch = batch[:0]
			return
		}

		collectionBatches := make(map[string][]pubsubRequest)
		for _, req := range batch {
			collectionBatches[req.evt.CollectionID] = append(collectionBatches[req.evt.CollectionID], req)
		}

		for collectionID, collectionReqs := range collectionBatches {
			pb.processCollectionBatch(collectionID, collectionReqs)
		}

		batch = batch[:0]
	}

	for {
		select {
		case <-pb.ctx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			processBatch()
			return

		case req := <-pb.queue:
			batch = append(batch, req)

			if len(batch) == 1 {
				timer.Reset(pb.timeout)
			}

			if len(batch) >= pb.batchSize {
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				processBatch()
			}

		case <-timer.C:
			processBatch()
		}
	}
}

// processCollectionBatch handles a batch of documents for a single collection.
func (pb *pubsubBatcher) processCollectionBatch(collectionID string, reqs []pubsubRequest) {
	var newEvents []pubsubRequest
	var relayEvents []pubsubRequest
	for _, req := range reqs {
		if req.evt.IsRelay {
			relayEvents = append(relayEvents, req)
		} else {
			newEvents = append(newEvents, req)
		}
	}

	if len(relayEvents) > 0 {
		documents := make([]protocol.DocumentInfo, len(relayEvents))
		for i, req := range relayEvents {
			documents[i] = protocol.DocumentInfo{
				DocID: req.evt.DocID,
				CID:   req.evt.Cid.Bytes(),
				Block: req.evt.Block,
			}
		}

		pubsubReq := &protocol.PushLogRequest{
			CollectionID: collectionID,
			Creator:      pb.p2p.host.ID(),
			Documents:    documents,
		}

		b, err := cbor.Marshal(pubsubReq)
		if err != nil {
			log.ErrorE("Failed to marshal relay batch request", err)
		} else {
			if err := pb.p2p.host.PublishToTopicAsync(pb.ctx, collectionID, b); err != nil {
				log.ErrorE("Failed to publish relay batch to collection topic", err)
			}
		}
	}

	if len(newEvents) == 0 {
		return
	}

	currentBatch := make([]pubsubRequest, 0, len(newEvents))
	currentSize := 0
	for _, req := range newEvents {
		docSize := len(req.evt.DocID) + len(req.evt.Cid.Bytes()) + len(req.evt.Block) + 50
		if currentSize+docSize > maxPubsubMessageSize && len(currentBatch) > 0 {
			pb.publishBatch(collectionID, currentBatch)
			currentBatch = currentBatch[:0]
			currentSize = 0
		}
		currentBatch = append(currentBatch, req)
		currentSize += docSize
	}

	if len(currentBatch) > 0 {
		pb.publishBatch(collectionID, currentBatch)
	}
}

// publishBatch publishes a batch of documents to the collection topic.
func (pb *pubsubBatcher) publishBatch(collectionID string, batch []pubsubRequest) {
	if len(batch) == 0 {
		return
	}

	rootBlocks := make([]rootBlockWithBytes, 0, len(batch))
	documents := make([]protocol.DocumentInfo, len(batch))

	for i, req := range batch {
		documents[i] = protocol.DocumentInfo{
			DocID: req.evt.DocID,
			CID:   req.evt.Cid.Bytes(),
			Block: req.evt.Block,
		}

		block, err := coreblock.GetFromBytes(req.evt.Block)
		if err != nil {
			log.ErrorE("Failed to parse block for CAR generation", err)
			continue
		}
		rootBlocks = append(rootBlocks, rootBlockWithBytes{
			block:    block,
			rawBytes: req.evt.Block,
		})
	}

	var carData []byte
	if len(rootBlocks) > 0 {
		var err error
		carData, err = pb.p2p.generateCARForBlocksWithBytes(pb.ctx, rootBlocks)
		if err != nil {
			log.ErrorE("Failed to generate CAR for batch", err)
		}
	}

	estimatedSize := len(carData)
	for _, doc := range documents {
		estimatedSize += len(doc.DocID) + len(doc.CID) + len(doc.Block) + 50
	}

	if estimatedSize > maxPubsubMessageSize && len(batch) > 1 {
		mid := len(batch) / 2
		pb.publishBatch(collectionID, batch[:mid])
		pb.publishBatch(collectionID, batch[mid:])
		return
	}

	pubsubReq := &protocol.PushLogRequest{
		CollectionID: collectionID,
		Creator:      pb.p2p.host.ID(),
		CAR:          carData,
		Documents:    documents,
	}

	b, err := cbor.Marshal(pubsubReq)
	if err != nil {
		log.ErrorE("Failed to marshal batched pubsub request", err)
		return
	}

	if err := pb.p2p.host.PublishToTopicAsync(pb.ctx, collectionID, b); err != nil {
		log.ErrorE("Failed to publish batched request to collection topic", err)
	}

	for _, req := range batch {
		if req.evt.DocID != "" {
			docReq := &protocol.PushLogRequest{
				DocID:        req.evt.DocID,
				CID:          req.evt.Cid.Bytes(),
				CollectionID: collectionID,
				Creator:      pb.p2p.host.ID(),
				Block:        req.evt.Block,
			}
			docB, err := cbor.Marshal(docReq)
			if err != nil {
				log.ErrorE("Failed to marshal doc pubsub request", err)
				continue
			}
			if err := pb.p2p.host.PublishToTopicAsync(pb.ctx, req.evt.DocID, docB); err != nil {
				log.ErrorE("Failed to publish to doc topic", err)
			}
		}
	}
}

func (pb *pubsubBatcher) close() {
	pb.cancel()
	pb.wg.Wait()
}
