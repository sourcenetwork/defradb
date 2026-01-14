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
	// syncWorkerCount is the number of workers processing sync requests.
	syncWorkerCount = 200
	// syncQueueSize is the maximum number of pending sync requests.
	syncQueueSize = 50000
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
	// GetNodeIdentity returns the current node identity.
	GetNodeIdentity(ctx context.Context) (immutable.Option[identity.PublicRawIdentity], error)
	// GetNodeIdentityToken returns an identity token for the given audience.
	GetNodeIdentityToken(ctx context.Context, audience immutable.Option[string]) ([]byte, error)
	// GetCollections returns all collections and their descriptions matching the given options
	// that currently exist within this [Store].
	GetCollections(ctx context.Context, options client.CollectionFetchOptions) ([]client.Collection, error)
	// Merge initiates a merge of the DAG and caches the resulting values into the datastore.
	Merge(ctx context.Context, evt event.Merge) error
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

	// timeout duration for syncing block links.
	syncBlockLinkTimeout time.Duration

	// seCoordinator manages searchable encryption artifact replication
	seCoordinator *se.Coordinator

	// pushHandlers are called when documents are pushed to replicators
	pushHandlers []PushToReplicatorsHandler
}

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
	}
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
	cols, err := p.db.GetCollections(
		ctx,
		client.CollectionFetchOptions{
			CollectionID: immutable.Some(collectionID),
		},
	)
	if err != nil {
		return false, err
	}
	return len(cols) > 0, nil
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
		cols[0], // For now we assume there is only one collection.
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

	if err := p.processPushlogRequest(p.ctx, req, false); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			log.Info("Context done during pushlog request processing", corelog.Any("Error", err))
			return nil, nil
		}
		return nil, errors.Wrap(fmt.Sprintf("Failed to process pushlog request %s", topic), err)
	}

	return nil, nil
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
		isMerged, err := p.db.Multistore().Blockstore().IsMerged(ctx, headCID)
		if err != nil {
			return err
		}
		if isMerged {
			return nil
		}

		// Check if we have the collection
		hasCollection, err := p.hasCollection(ctx, req.CollectionID)
		if err != nil {
			return err
		}
		if !hasCollection {
			return nil
		}

		// No need to check access if the message is for replication as the node sending
		// will have done so deliberately.
		if !isReplicator {
			mightHaveAccess, err := p.trySelfHasAccess(ctx, block, req.CollectionID)
			if err != nil {
				return err
			}
			if !mightHaveAccess {
				return nil
			}
		}

		if len(req.CAR) > 0 {
			_, err = p.importCAR(ctx, req.CAR)
			if err != nil {
				return err
			}
		} else {
			err = p.syncDAG(ctx, block)
			if err != nil {
				return err
			}
		}

		mergeEvt := event.Merge{
			DocID:        req.DocID,
			ByPeer:       req.SenderID,
			FromPeer:     req.Creator,
			Cid:          headCID,
			CollectionID: req.CollectionID,
		}
		err = p.db.Merge(ctx, mergeEvt)
		if err != nil {
			return err
		}

		// Notify bus subscribers and the network of peers that we have a new document available.
		updateEvt := event.Update{
			DocID:        req.DocID,
			Cid:          headCID,
			CollectionID: req.CollectionID,
			Block:        req.Block,
			IsRelay:      true,
		}
		p.db.Events().Publish(event.NewMessage(event.UpdateName, updateEvt))
		if err := p.SendUpdate(updateEvt); err != nil {
			log.ErrorE("Failed to send update after sync", err, slog.Any("PeerID", p.host.ID()))
		}

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
		req := &protocol.PushLogRequest{
			DocID:        evt.DocID,
			CID:          evt.Cid.Bytes(),
			CollectionID: evt.CollectionID,
			Creator:      p.host.ID(),
			Block:        evt.Block,
		}

		// Generate CAR containing root block and all linked blocks
		if !evt.IsRelay {
			rootBlock, err := coreblock.GetFromBytes(evt.Block)
			if err != nil {
				return err
			}
			carData, err := p.generateCAR(p.ctx, rootBlock)
			if err != nil {
				return err
			}
			req.CAR = carData
		}

		b, err := cbor.Marshal(req)
		if err != nil {
			return err
		}

		if evt.DocID != "" {
			if err := p.host.PublishToTopicAsync(p.ctx, evt.DocID, b); err != nil {
				return NewErrPublishingToDocIDTopic(err, evt.Cid.String(), evt.DocID)
			}
		}

		if err := p.host.PublishToTopicAsync(p.ctx, evt.CollectionID, b); err != nil {
			return NewErrPublishingToSchemaTopic(err, evt.Cid.String(), evt.CollectionID)
		}
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
