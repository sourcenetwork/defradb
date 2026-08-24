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
	"math"
	"reflect"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/multiformats/go-multiaddr"

	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/corelog"
	"github.com/sourcenetwork/immutable"
	lens "github.com/sourcenetwork/lens/host-go/node"

	"github.com/sourcenetwork/defradb/acp/dac"
	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/event"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/datastore"
	acpDB "github.com/sourcenetwork/defradb/internal/db/acp"
	"github.com/sourcenetwork/defradb/internal/db/description"
	"github.com/sourcenetwork/defradb/internal/db/id"
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

	// dagSyncWorkers is the number of concurrent pubsub message workers.
	// A bounded pool prevents the OOM caused by spawning an unbounded number of
	// goroutines when the network delivers messages faster than they are processed.
	dagSyncWorkers = 32

	// msgQueueSize is the capacity of the incoming pubsub message queue.
	// Messages that arrive when the queue is full are dropped and counted.
	msgQueueSize = 50_000

	// msgQueueMemDivisor sets the queue's byte budget as this fraction of the process
	// memory limit. A slot count does not bound memory on its own, because a queued
	// message holds its decoded payload and payload sizes are not uniform. The queue
	// only has to absorb bursts, so it is given a fraction rather than the whole limit.
	msgQueueMemDivisor = 4

	// msgQueueFallbackMaxBytes is the byte budget used when the process has no memory
	// limit set, where a fraction of it would be meaningless.
	msgQueueFallbackMaxBytes = 1 << 30

	// msgQueueMinBytes floors the budget so a very small memory limit still admits a message.
	msgQueueMinBytes = maxPubsubMessageSize

	// maxInboundDocuments caps the entries a received batch may carry, sized as headroom over
	// this node's own batchMaxDocs to hold the per-entry cost under a tenth of a message. The
	// cap applies while decoding, since the decode allocates before the message can be charged.
	maxInboundDocuments = 16 * batchMaxDocs

	// statsInterval is how often queue depth and merge counters are reported. Rates are
	// reported per interval rather than per event, which keeps the ingest path quiet
	// under load where per-event logging would dominate the output.
	statsInterval = 30 * time.Second
)

// retainedBytesPerDocument is the fixed cost of one decoded batch entry, whatever it carries.
// Read from the type so it stays right if a field is added.
var retainedBytesPerDocument = int64(reflect.TypeFor[protocol.DocumentInfo]().Size())

// pushLogDecMode rejects a batch carrying more than maxInboundDocuments entries. The options
// are constant, so DecMode can only fail on a rejected option value.
var pushLogDecMode = func() cbor.DecMode {
	mode, err := cbor.DecOptions{MaxArrayElements: maxInboundDocuments}.DecMode()
	if err != nil {
		panic(err)
	}
	return mode
}()

// queuedMessage is a decoded request paired with the bytes reserved for it, so the queue's
// accounting can be reversed once the message is processed. The reservation covers the wire
// bytes, which the decoder copies rather than points into, plus the per-entry cost of a batch.
type queuedMessage struct {
	req  *protocol.PushLogRequest
	size int64
}

// queueByteBudget returns the byte budget for the incoming message queue, taken as a fraction
// of the process memory limit so a node with a smaller limit queues proportionally less.
//
// The limit comes from GOMEMLIMIT. The Go runtime does not read a container's memory ceiling,
// so a node without it set gets the fallback, a fixed size that may sit above the ceiling the
// node runs under. Zero is a limit an operator can set, so only MaxInt64 counts as absent.
func queueByteBudget() int64 {
	limit := debug.SetMemoryLimit(-1)
	if limit == math.MaxInt64 {
		return msgQueueFallbackMaxBytes
	}
	return max(limit/msgQueueMemDivisor, msgQueueMinBytes)
}

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
	GetCollections(
		ctx context.Context,
		opts ...options.Enumerable[options.GetCollectionsOptions],
	) ([]client.Collection, error)
	// Merge initiates a merge of the DAG and caches the resulting values into the datastore.
	Merge(ctx context.Context, evt event.Merge) error
	// MergeBatchWithTxn merges events in bounded chunks. The returned slice is parallel
	// to merges and reports which events committed; the rest are not stored and must not
	// be relayed onward as merged. The error names every dropped event.
	MergeBatchWithTxn(ctx context.Context, merges []event.Merge) ([]bool, error)
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

	ctx                  context.Context
	db                   DB
	lens                 *lens.Node
	collectionRepository *description.CollectionRepository
	host                 client.Host
	kms                  kms.Service

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

	// a cid queue for the processing of Pushlogs
	processQueue *processQueue

	// timeout duration for syncing block links.
	syncBlockLinkTimeout time.Duration

	// seCoordinator manages searchable encryption artifact replication
	seCoordinator *se.Coordinator

	// pushHandlers are called when documents are pushed to replicators
	pushHandlers []PushToReplicatorsHandler

	// replicationFilter, when non-nil, is called for each incoming replicated document.
	// Returning false from the filter drops the document silently.
	replicationFilter client.ReplicationFilter

	// topicPeerCounts tracks the number of known peers per pubsub topic.
	// Updated by peerEventHandler; consulted by SendUpdate to emit P2PNoPeers events.
	topicPeerCounts map[string]int
	topicPeerMu     sync.RWMutex

	// batcher accumulates per-collection pubsub updates into a single batched message.
	batcher *pubsubBatcher

	// msgQueue receives incoming pubsub messages for async processing by a fixed worker pool.
	msgQueue   chan queuedMessage
	msgWorkers sync.WaitGroup

	// msgQueueBytes is the wire size of the messages currently queued or in flight.
	msgQueueBytes atomic.Int64
	// msgQueueMaxBytes caps msgQueueBytes. Zero or less disables the byte bound.
	msgQueueMaxBytes int64

	// Counters reported by reportStats and reset on each report, so each line carries
	// the rate for that interval rather than a running total.
	// statMsgsIn counts every message the dispatcher hands over, including dropped ones,
	// so the drop counters have a denominator.
	statMsgsIn        atomic.Int64
	statDroppedBudget atomic.Int64
	statDroppedFull   atomic.Int64
	statMergedDocs    atomic.Int64
	statDroppedDocs   atomic.Int64
	// statSkippedDocs counts documents deliberately not merged: already held, or
	// excluded by access or the replication filter. Not a loss, so kept apart.
	statSkippedDocs atomic.Int64
	statBatches     atomic.Int64
	// statBatchesWithDrops counts batches in which at least one document dropped. The
	// rest of the batch still commits, so this is not a count of failed batches.
	statBatchesWithDrops atomic.Int64

	// The inbound CAR path: imports attempted, imports abandoned, and the blocks an
	// abandoned import had already written. Every document this node stores arrives this
	// way, so the generation counters below say nothing about it.
	statCARImports      atomic.Int64
	statCARImportFailed atomic.Int64

	// CAR generation counters. A CAR that carries only the root block gives its receiver
	// nothing to import, so the receiver falls back to a per-link BitSwap walk.
	statCARBuilt           atomic.Int64
	statCARFailed          atomic.Int64
	statCARMissing         atomic.Int64
	carFailureReason       failureReasons
	carImportFailureReason failureReasons

	// statSyncDAGCalls counts walks started, which is how often a document could not be
	// merged from the CAR alone and needed blocks fetched.
	statSyncDAGCalls     atomic.Int64
	syncDAGFailureReason failureReasons

	// docDropReason names why an inbound document never reached the merge. The set is
	// fixed: a reason taken from an error string would let a peer grow the map.
	docDropReason failureReasons
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
	collectionRepository *description.CollectionRepository,
	replicationFilter client.ReplicationFilter,
) (*P2P, error) {
	p := P2P{
		ctx:                  ctx,
		db:                   db,
		lens:                 lens,
		collectionRepository: collectionRepository,
		host:                 host,
		identityProtocol:     protocol.NewIdentityProtocol(host, db.GetNodeIdentityToken),
		replicators:          make(map[string]map[string][]string),
		peerIdentities:       make(map[string]identity.Identity),
		retryIntervals:       db.RetryIntervals(),
		processQueue:         newProcessQueue(),
		replicationFilter:    replicationFilter,
		syncBlockLinkTimeout: db.P2PBlockSyncTimeout(),
		topicPeerCounts:      make(map[string]int),
		msgQueue:             make(chan queuedMessage, msgQueueSize),
		msgQueueMaxBytes:     queueByteBudget(),
	}

	// The bounds are fixed for the life of the process, and queueBytes and the drop
	// counters cannot be read without them.
	log.Info("p2p queue bounds",
		corelog.Int("slots", msgQueueSize),
		corelog.Int64("byteBudget", p.msgQueueMaxBytes))

	for i := 0; i < dagSyncWorkers; i++ {
		p.msgWorkers.Add(1)
		go p.processMessageWorker()
	}
	go p.reportStats()

	p.replicatorProtocol = protocol.NewCommChannel(host, "rep", &pushLogCommProcessor{p2p: &p})
	p.batcher = newPubsubBatcher(host.ID(), func(topic string, data []byte) error {
		return host.PublishToTopicAsync(ctx, topic, data)
	})

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
			db.NodeACP,
			db.DocumentACP(),
			collectionRetriever,
			nodeIdentity,
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

// Disconnect closes the connection to the peer(s) identified by the given addresses.
func (p *P2P) Disconnect(ctx context.Context, addresses []string) error {
	seen := make(map[string]struct{})
	for _, addr := range addresses {
		maddr, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			return err
		}
		_, p2ppart := multiaddr.SplitLast(maddr)
		if p2ppart == nil || p2ppart.Protocol().Code != multiaddr.P_P2P {
			return errors.New("multiaddr does not contain peer ID")
		}
		peerID := p2ppart.Value()
		if _, ok := seen[peerID]; ok {
			continue
		}
		seen[peerID] = struct{}{}
		if err := p.host.Disconnect(ctx, peerID); err != nil {
			return err
		}
	}
	return nil
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

	ident, err := p.db.GetNodeIdentity(p.ctx)
	if err != nil {
		log.ErrorE("Failed to get node identity", err)
		return false
	}
	getColOpts := options.GetCollections().SetCollectionID(block.Delta.GetCollectionVersionID())
	if ident.HasValue() {
		getColOpts = getColOpts.SetIdentity(identity.FromDID(ident.Value().DID))
	}

	cols, err := p.db.GetCollections(ctx, getColOpts)
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

	// A block may be owned by several documents (shared field blocks); read access to any one is
	// enough. docIDsForBlockCID returns a single empty docID for collection-level blocks, which
	// CheckDocReadAccessWithIdentityFunc gates on the collection object for a branchable collection.
	docIDs, err := p.docIDsForBlockCID(ctx, c, block)
	if err != nil {
		log.ErrorE("Failed to resolve block doc ID", err)
		return false
	}

	for _, docID := range docIDs {
		peerHasAccess, err := acpDB.CheckDocReadAccessWithIdentityFunc(
			ctx,
			identFunc,
			p.db.NodeACP(),
			p.db.DocumentACP().Value(),
			cols[0], // For now we assume there is only one collection.
			docID,
		)
		if err != nil {
			log.ErrorE("Failed to check access", err)
			return false
		}
		if peerHasAccess {
			return true
		}
	}

	return false
}

// trySelfHasAccess checks if the local node has access to the given block.
//
// This is a best-effort check and returns true unless we explicitly find that the local node
// doesn't have access or if we get an error. The node sending is ultimately responsible for
// ensuring that the recipient has access.
//
// The collection is resolved from collectionID (the stable root collection id) rather than the
// block's collection version id, because the local node may legitimately hold a different version
// of the collection than the one the block was authored against (e.g. replication to an older
// collection version).
func (p *P2P) trySelfHasAccess(
	ctx context.Context,
	blockCID cid.Cid,
	block *coreblock.Block,
	collectionID string,
	docID string,
) (bool, error) {
	if !p.db.DocumentACP().HasValue() {
		return true, nil
	}

	ident, err := p.db.GetNodeIdentity(ctx)
	if err != nil {
		return false, err
	}

	// The collection lookup is a local operation on this node — authorise it
	// as the node itself so NAC sees a known identity rather than "anonymous".
	getColOpts := options.GetCollections().SetCollectionID(collectionID)
	if ident.HasValue() {
		getColOpts = getColOpts.SetIdentity(identity.FromDID(ident.Value().DID))
	}
	cols, err := p.db.GetCollections(ctx, getColOpts)
	if err != nil {
		return false, err
	}
	if len(cols) == 0 {
		return false, client.ErrCollectionNotFound
	}
	if !ident.HasValue() {
		return true, nil
	}

	docIDs := []string{docID}
	if docID == "" {
		docIDs, err = p.docIDsForBlockCID(ctx, blockCID, block)
		if err != nil {
			return false, err
		}
	}

	for _, docID := range docIDs {
		peerHasAccess, err := acpDB.CheckDocReadAccessWithIdentityFunc(
			ctx,
			func() immutable.Option[identity.Identity] {
				return immutable.Some(identity.FromDID(ident.Value().DID))
			},
			p.db.NodeACP(),
			p.db.DocumentACP().Value(),
			cols[0], // For now we assume there is only one collection.
			docID,
		)
		if err != nil {
			return false, err
		}
		if peerHasAccess {
			return true, nil
		}
	}

	return false, nil
}

func (p *P2P) docIDsForBlockCID(
	ctx context.Context,
	blockCID cid.Cid,
	block *coreblock.Block,
) ([]string, error) {
	if block.Delta.IsCollection() {
		return []string{""}, nil
	}

	docIDs, err := id.GetDocIDsForBlockFromStore(
		ctx,
		p.db.Multistore().Systemstore(),
		blockCID,
	)
	if err != nil {
		return nil, err
	}
	if len(docIDs) > 0 {
		return docIDs, nil
	}
	if block.Delta.IsComposite() && len(block.Heads) == 0 {
		return []string{client.NewDocIDV0(blockCID).String()}, nil
	}
	return nil, nil
}

// pubSubMessageHandler decodes the incoming wire message and enqueues it for
// processing by the bounded worker pool. It never blocks the libp2p pubsub
// dispatcher: if the queue is full the message is dropped and counted.
//
// Most of the budget is claimed before decoding, so a message that cannot be queued does not
// pay for a decode it will not use. The per-entry cost of a batch is only known after the
// decode, and is claimed there.
func (p *P2P) pubSubMessageHandler(from string, topic string, msg []byte) ([]byte, error) {
	size := int64(len(msg))
	p.statMsgsIn.Add(1)
	if !p.claimQueueBytes(size) {
		p.statDroppedBudget.Add(1)
		return nil, nil
	}

	req := &protocol.PushLogRequest{}
	if err := pushLogDecMode.Unmarshal(msg, req); err != nil {
		p.releaseQueueBytes(size)
		return nil, err
	}
	req.SenderID = from

	// A batch entry costs a fixed-size struct however little it carries, so the wire size alone
	// does not bound what the queue holds. It folds into size rather than being tracked apart,
	// because every release path and the worker give back that one value.
	if perDocument := retainedBytesPerDocument * int64(len(req.Documents)); perDocument > 0 {
		if !p.claimQueueBytes(perDocument) {
			p.releaseQueueBytes(size)
			p.statDroppedBudget.Add(1)
			return nil, nil
		}
		size += perDocument
	}

	select {
	case p.msgQueue <- queuedMessage{req: req, size: size}:
	case <-p.ctx.Done():
		p.releaseQueueBytes(size)
	default:
		p.releaseQueueBytes(size)
		p.statDroppedFull.Add(1)
	}
	return nil, nil
}

// claimQueueBytes reserves size against the queue's byte budget, reporting whether the
// reservation fit. A budget of zero or less means the byte bound is disabled.
func (p *P2P) claimQueueBytes(size int64) bool {
	if p.msgQueueMaxBytes <= 0 {
		return true
	}
	if p.msgQueueBytes.Add(size) > p.msgQueueMaxBytes {
		p.msgQueueBytes.Add(-size)
		return false
	}
	return true
}

// releaseQueueBytes returns a reservation made by claimQueueBytes.
func (p *P2P) releaseQueueBytes(size int64) {
	if p.msgQueueMaxBytes <= 0 {
		return
	}
	p.msgQueueBytes.Add(-size)
}

// reportStats periodically logs queue occupancy and merge outcomes. Queue depth and the
// split between merged and dropped documents are otherwise only visible from a heap
// profile, which is too invasive to take routinely.
func (p *P2P) reportStats() {
	ticker := time.NewTicker(statsInterval)
	defer ticker.Stop()
	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			droppedOverBudget := p.statDroppedBudget.Swap(0)
			droppedQueueFull := p.statDroppedFull.Swap(0)
			log.Info("p2p stats",
				corelog.Int("queueDepth", len(p.msgQueue)),
				corelog.Int64("queueBytes", p.msgQueueBytes.Load()),
				corelog.Int64("droppedOverBudget", droppedOverBudget),
				corelog.Int64("droppedQueueFull", droppedQueueFull),
				corelog.Int64("batches", p.statBatches.Swap(0)),
				corelog.Int64("batchesWithDrops", p.statBatchesWithDrops.Swap(0)),
				corelog.Int64("docsMerged", p.statMergedDocs.Swap(0)),
				corelog.Int64("docsDropped", p.statDroppedDocs.Swap(0)),
				corelog.Int64("docsSkipped", p.statSkippedDocs.Swap(0)),
				corelog.Int64("msgsIn", p.statMsgsIn.Swap(0)),
				corelog.Int64("carImports", p.statCARImports.Swap(0)),
				corelog.Int64("carImportFailed", p.statCARImportFailed.Swap(0)),
				corelog.Int64("carBuilt", p.statCARBuilt.Swap(0)),
				corelog.Int64("carFailed", p.statCARFailed.Swap(0)),
				corelog.Int64("carMissingLinks", p.statCARMissing.Swap(0)),
				corelog.Int64("syncDAGCalls", p.statSyncDAGCalls.Swap(0)),
			)
			// A drop at the door is data this node will not hold. The stats line above is at
			// info, so a node running at error level sees only this.
			if droppedOverBudget != 0 || droppedQueueFull != 0 {
				log.Error("dropped inbound pubsub messages",
					corelog.Int64("overBudget", droppedOverBudget),
					corelog.Int64("queueFull", droppedQueueFull))
			}

			reportFailureReasons("car failures", p.carFailureReason.drain())
			reportFailureReasons("syncDAG failures", p.syncDAGFailureReason.drain())
			reportFailureReasons("document drops", p.docDropReason.drain())
			reportFailureReasons("CAR import failures", p.carImportFailureReason.drain())
		}
	}
}

// reportFailureReasons logs one line naming every reason that occurred in the interval,
// or nothing when there were none. Reasons come from a fixed set, so the line has a
// bounded width.
func reportFailureReasons(msg string, counts []reasonCount) {
	if len(counts) == 0 {
		return
	}
	fields := make([]slog.Attr, 0, len(counts))
	for _, c := range counts {
		fields = append(fields, corelog.Int64(c.reason, c.count))
	}
	log.Info(msg, fields...)
}

// processMessageWorker is a long-lived goroutine that drains p.msgQueue and
// calls processPushlogRequest for each message. dagSyncWorkers of these run
// concurrently, bounding the goroutine count regardless of inbound message rate.
func (p *P2P) processMessageWorker() {
	defer p.msgWorkers.Done()
	for m := range p.msgQueue {
		err := p.processPushlogRequest(p.ctx, m.req, false)
		// Released here rather than deferred so the budget frees up per message, not
		// when the worker exits.
		p.releaseQueueBytes(m.size)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				log.Info("Context done during pushlog request processing", corelog.Any("Error", err))
				continue
			}
			// A failed pushlog is a document this node did not store, so it is reported at
			// error level.
			log.ErrorE("Failed to process pushlog request", err)
		}
	}
}

func (p *P2P) peerEventHandler(peerID string, topic string, eventType string) {
	p.db.Events().Publish(event.NewMessage(event.TopicPeerEventName, event.TopicPeerEvent{
		PeerID:    peerID,
		Topic:     topic,
		EventType: eventType,
	}))

	p.topicPeerMu.Lock()
	switch eventType {
	case "JOINED":
		p.topicPeerCounts[topic]++
	case "LEFT":
		if p.topicPeerCounts[topic] > 0 {
			p.topicPeerCounts[topic]--
		}
	}
	p.topicPeerMu.Unlock()
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

	// Handle batched multi-document push using a single shared transaction.
	if len(req.Documents) > 0 {
		// Counted before the documents are processed, so a batch still counts when none
		// of them survives.
		p.statBatches.Add(1)

		results, dropped, err := p.processBatchedDocuments(ctx, req, isReplicator)
		if err != nil {
			return err
		}
		if len(results) > 0 {
			merges := make([]event.Merge, len(results))
			for i, r := range results {
				merges[i] = r.merge
			}
			merged, err := p.db.MergeBatchWithTxn(ctx, merges)
			for _, ok := range merged {
				if ok {
					p.statMergedDocs.Add(1)
				} else {
					p.dropDoc("mergeFailed")
					dropped++
				}
			}
			if err != nil {
				log.ErrorE("Failed to merge documents in batch", err,
					corelog.String("PeerID", req.SenderID),
					corelog.Int("Documents", len(merges)))
			}
			for i, r := range results {
				if !merged[i] {
					// Not stored here, so relaying it would advertise a document this node
					// cannot serve.
					continue
				}
				updateEvt := event.Update{
					DocID:        r.merge.DocID,
					Cid:          r.merge.Cid,
					CollectionID: r.merge.CollectionID,
					Block:        r.block,
					IsRelay:      true,
				}
				p.db.Events().Publish(event.NewMessage(event.UpdateName, updateEvt))
				if err := p.SendUpdate(updateEvt); err != nil {
					log.ErrorE("Failed to send update after batch sync", err,
						slog.String("DocID", r.merge.DocID),
						slog.Any("PeerID", p.host.ID()),
					)
				}
			}
		}

		if dropped > 0 {
			p.statBatchesWithDrops.Add(1)
		}
		return nil
	}

	headCID, err := cid.Cast(req.CID)
	if err != nil {
		p.dropDoc("invalidCID")
		return err
	}

	// enqueue returns without running the handler when the same document is already in
	// flight, or when the queue is full. Neither case reaches the outcomes recorded inside it.
	handled := false
	err = p.processQueue.enqueue(headCID.String(), func() error {
		handled = true

		// Check if we've already merged this block. If so, skip the sink process.
		isMerged, err := p.db.Multistore().Blockstore().IsMerged(ctx, headCID)
		if err != nil {
			p.dropDoc("isMergedError")
			return err
		}
		if isMerged {
			p.skipDoc("alreadyMerged")
			return nil
		}

		// Decode the root block without writing to the blockstore so that CID verification,
		// access control, and the replication filter all run before any storage occurs.
		// Filtered-out documents will not consume blockstore space.
		var block *coreblock.Block
		if len(req.CAR) > 0 {
			block, err = peekCARRootBlock(req.CAR)
		} else {
			block, err = coreblock.GetFromBytes(req.Block)
		}
		if err != nil {
			p.dropDoc("blockDecode")
			return err
		}

		// Verify the advertised CID actually matches the block contents, so a peer cannot push
		// arbitrary content under a CID of its choosing.
		blockLink, err := block.GenerateLink()
		if err != nil {
			p.dropDoc("generateLink")
			return err
		}
		if blockLink.Cid != headCID {
			p.dropDoc("cidMismatch")
			return ErrBlockCIDMismatch
		}

		// No need to check access if the message is for replication as the node sending
		// will have done so deliberately.
		if !isReplicator {
			mightHaveAccess, err := p.trySelfHasAccess(ctx, headCID, block, req.CollectionID, req.DocID)
			if err != nil {
				p.dropDoc("accessError")
				return err
			}
			if !mightHaveAccess {
				// If we know we don't have access, we can skip the rest of the processing.
				p.skipDoc("noAccess")
				return nil
			}
		}

		// Run the replication filter before writing any blocks to storage.
		if !p.filterAllowsReplication(ctx, req.CollectionID, req.DocID, block) {
			p.skipDoc("filtered")
			return nil
		}

		// All pre-storage checks passed — now write blocks to the blockstore.
		if len(req.CAR) > 0 {
			// CAR contains the full block DAG — import it directly, no round-trip sync needed.
			if _, err = p.importCAR(ctx, req.CAR); err != nil {
				p.dropDoc("importCAR")
				return err
			}
		} else {
			if err = p.syncDAG(ctx, block); err != nil {
				p.dropDoc("syncDAG")
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
		if err = p.db.Merge(ctx, mergeEvt); err != nil {
			p.dropDoc("mergeFailed")
			return err
		}
		p.statMergedDocs.Add(1)

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
			// We don't need to return the error for this side-effect-function call.
			// It's a bonus action that shouldn't affect the caller of `processPushlogRequest`.
			log.ErrorE("Failed to send update after sync", err, slog.Any("PeerID", p.host.ID()))
		}

		return nil
	})

	if !handled {
		if err == nil {
			p.skipDoc("inFlight")
		} else {
			p.dropDoc("syncQueueFull")
		}
	}
	return err
}

// batchedDoc pairs a merge event with its raw block bytes for relay after batch commit.
type batchedDoc struct {
	merge event.Merge
	block []byte // raw composite block bytes; empty when doc was received via CAR
}

// processBatchedDocuments runs the per-document pre-storage checks (CID verification,
// access control, replication filter) and block sync for every document in a batched
// PushLogRequest.  It returns the subset that passed all checks and are ready to be
// committed via MergeBatchWithTxn, paired with their raw block bytes for relay.
func (p *P2P) processBatchedDocuments(
	ctx context.Context,
	req *protocol.PushLogRequest,
	isReplicator bool,
) ([]batchedDoc, int, error) {
	results := make([]batchedDoc, 0, len(req.Documents))

	// The counters are process-wide, so this batch's own drops are tallied separately.
	dropped := 0
	drop := func(reason string) {
		p.dropDoc(reason)
		dropped++
	}

	for _, doc := range req.Documents {
		headCID, err := cid.Cast(doc.CID)
		if err != nil {
			log.ErrorE("Batch: invalid CID", err,
				slog.String("DocID", doc.DocID),
				slog.String("CollectionID", req.CollectionID),
			)
			drop("invalidCID")
			continue
		}

		isMerged, err := p.db.Multistore().Blockstore().IsMerged(ctx, headCID)
		if err != nil {
			log.ErrorE("Batch: IsMerged check failed", err, slog.String("DocID", doc.DocID))
			drop("isMergedError")
			continue
		}
		if isMerged {
			p.skipDoc("alreadyMerged")
			continue
		}

		var block *coreblock.Block
		if len(doc.CAR) > 0 {
			block, err = peekCARRootBlock(doc.CAR)
		} else {
			block, err = coreblock.GetFromBytes(doc.Block)
		}
		if err != nil {
			log.ErrorE("Batch: block decode failed", err, slog.String("DocID", doc.DocID))
			drop("blockDecode")
			continue
		}

		blockLink, err := block.GenerateLink()
		if err != nil {
			log.ErrorE("Batch: GenerateLink failed", err, slog.String("DocID", doc.DocID))
			drop("generateLink")
			continue
		}
		if blockLink.Cid != headCID {
			log.Error("Batch: CID mismatch", slog.String("DocID", doc.DocID))
			drop("cidMismatch")
			continue
		}

		if !isReplicator {
			mightHaveAccess, err := p.trySelfHasAccess(ctx, headCID, block, req.CollectionID, doc.DocID)
			if err != nil {
				log.ErrorE("Batch: access check failed", err, slog.String("DocID", doc.DocID))
				drop("accessError")
				continue
			}
			if !mightHaveAccess {
				p.skipDoc("noAccess")
				continue
			}
		}

		if !p.filterAllowsReplication(ctx, req.CollectionID, doc.DocID, block) {
			p.skipDoc("filtered")
			continue
		}

		if len(doc.CAR) > 0 {
			if _, err = p.importCAR(ctx, doc.CAR); err != nil {
				log.ErrorE("Batch: importCAR failed", err, slog.String("DocID", doc.DocID))
				drop("importCAR")
				continue
			}
		} else {
			if err = p.syncDAG(ctx, block); err != nil {
				log.ErrorE("Batch: syncDAG failed", err, slog.String("DocID", doc.DocID))
				drop("syncDAG")
				continue
			}
		}

		results = append(results, batchedDoc{
			merge: event.Merge{
				DocID:        doc.DocID,
				ByPeer:       req.SenderID,
				FromPeer:     req.Creator,
				Cid:          headCID,
				CollectionID: req.CollectionID,
			},
			block: doc.Block, // empty when CAR was used; relay falls back to DAG sync
		})
	}

	return results, dropped, nil
}
func (p *P2P) SendUpdate(evt event.Update) error {
	// push to each peer (replicator)
	p.pushLogToReplicators(evt)

	// Retries are for replicators only and should not pollute the pubsub network.
	if !evt.IsRetry {
		// Pre-generate a CAR so receivers import the full DAG without a BitSwap round-trip.
		var carData []byte
		if block, err := coreblock.GetFromBytes(evt.Block); err == nil {
			ctx, cancel := context.WithTimeout(p.ctx, networkRequestTimeout)
			if data, err := p.generateCAR(ctx, block); err == nil {
				carData = data
			} else {
				log.ErrorE("Failed to generate CAR for pubsub, receivers will fall back to DAG sync", err)
			}
			cancel()
		}

		req := &protocol.PushLogRequest{
			DocID:        evt.DocID,
			CID:          evt.Cid.Bytes(),
			CollectionID: evt.CollectionID,
			Creator:      p.host.ID(),
			Block:        evt.Block,
			CAR:          carData,
		}

		b, err := cbor.Marshal(req)
		if err != nil {
			return err
		}

		if evt.DocID != "" && !evt.IsRelay {
			if err := p.host.PublishToTopicAsync(p.ctx, evt.DocID, b); err != nil {
				return NewErrPublishingToDocIDTopic(err, evt.Cid.String(), evt.DocID)
			}
			p.topicPeerMu.RLock()
			noPeers := p.topicPeerCounts[evt.DocID] == 0
			p.topicPeerMu.RUnlock()
			if noPeers {
				p.db.Events().Publish(event.NewMessage(event.P2PNoPeersName, event.P2PNoPeers{
					DocID:        evt.DocID,
					CollectionID: evt.CollectionID,
					Topic:        evt.DocID,
				}))
			}
		}

		// Route collection-topic publishes through the batcher so multiple doc updates
		// within the flush window are coalesced into a single pubsub message.
		p.batcher.Add(evt.CollectionID, protocol.DocumentInfo{
			DocID: evt.DocID,
			CID:   evt.Cid.Bytes(),
			Block: evt.Block,
			CAR:   carData,
		})
		p.topicPeerMu.RLock()
		noPeers := p.topicPeerCounts[evt.CollectionID] == 0
		p.topicPeerMu.RUnlock()
		if noPeers {
			p.db.Events().Publish(event.NewMessage(event.P2PNoPeersName, event.P2PNoPeers{
				DocID:        evt.DocID,
				CollectionID: evt.CollectionID,
				Topic:        evt.CollectionID,
			}))
		}
	}

	return nil
}

const (
	syncWorkerCount = 8
	syncQueueSize   = 50_000
)

type syncRequest struct {
	key     string
	handler func() error
	result  chan error
}

// processQueue is a bounded worker pool that serialises sync requests by key,
// preventing unbounded goroutine growth and deduplicating concurrent requests.
type processQueue struct {
	mu       sync.Mutex
	inFlight map[string]struct{}
	queue    chan syncRequest
	wg       sync.WaitGroup
}

func newProcessQueue() *processQueue {
	pq := &processQueue{
		inFlight: make(map[string]struct{}),
		queue:    make(chan syncRequest, syncQueueSize),
	}
	for range syncWorkerCount {
		pq.wg.Add(1)
		go pq.worker()
	}
	return pq
}

func (pq *processQueue) worker() {
	defer pq.wg.Done()
	for req := range pq.queue {
		err := req.handler()
		pq.mu.Lock()
		delete(pq.inFlight, req.key)
		pq.mu.Unlock()
		if req.result != nil {
			req.result <- err
		}
	}
}

// enqueue submits a request and blocks until it completes.
// If the key is already in-flight the request is deduplicated and nil is returned.
func (pq *processQueue) enqueue(key string, handler func() error) error {
	pq.mu.Lock()
	if _, ok := pq.inFlight[key]; ok {
		pq.mu.Unlock()
		return nil
	}
	pq.inFlight[key] = struct{}{}
	result := make(chan error, 1)
	req := syncRequest{key: key, handler: handler, result: result}
	select {
	case pq.queue <- req:
		pq.mu.Unlock()
		return <-result
	default:
		delete(pq.inFlight, key)
		pq.mu.Unlock()
		return ErrSyncQueueFull
	}
}

func (pq *processQueue) close() {
	close(pq.queue)
	pq.wg.Wait()
}

// Close flushes any pending batched pubsub messages, drains in-flight sync requests,
// and waits for all worker goroutines to exit.
// It should be called once when the P2P subsystem is shutting down.
func (p *P2P) Close() {
	close(p.msgQueue)
	done := make(chan struct{})
	go func() {
		p.msgWorkers.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		log.Info("timed out waiting for pubsub workers to drain")
	}
	p.batcher.Close()
	p.processQueue.close()
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
