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
	"github.com/libp2p/go-libp2p/core/peer"
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

const networkRequestTimeout = 10 * time.Second

// accessCacheTTL is how long a positive read-access decision for a (peer, document) pair is
// reused before being re-checked. It is short so that a revoked grant becomes effective quickly,
// while still collapsing the per-block access checks of a single DAG sync into one round-trip.
const accessCacheTTL = 3 * time.Second

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

	// accessCache memoizes positive read-access decisions per (peer, document) so that serving
	// an entire document DAG to a peer does not incur one access-control round-trip per block.
	accessCache *accessCache

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
		syncBlockLinkTimeout: db.P2PBlockSyncTimeout(),
		accessCache:          newAccessCache(accessCacheTTL),
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

func (p *P2P) activePeerIDs(ctx context.Context) (map[string]struct{}, error) {
	activePeers, err := p.ActivePeers(ctx)
	if err != nil {
		return nil, err
	}

	peerIDs := make(map[string]struct{}, len(activePeers))
	for _, address := range activePeers {
		peerInfo, err := peer.AddrInfoFromString(address)
		if err != nil {
			return nil, err
		}
		peerIDs[peerInfo.ID.String()] = struct{}{}
	}

	return peerIDs, nil
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

	// The block is servable if the peer can read any one of its owning documents, so a cached
	// grant for any of them is enough. Caching collapses a document's many per-block checks into
	// one round-trip; only grants are cached, so a denied peer is re-checked and picks up a fresh
	// grant without delay. The collection id is keyed because a collection-level block has an empty
	// docID whose access is decided per collection. See accessCache.
	collectionID := cols[0].CollectionID()
	for _, docID := range docIDs {
		if p.accessCache.allowed(pid, collectionID, docID) {
			return true
		}
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
			p.accessCache.storeAllowed(pid, collectionID, docID)
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

// pubSubMessageHandler handles incoming PushLog messages from the pubsub network.
func (p *P2P) pubSubMessageHandler(from string, topic string, msg []byte) ([]byte, error) {
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

	// Verify the advertised CID actually matches the block contents, so a peer cannot push
	// arbitrary content under a CID of its choosing.
	blockLink, err := block.GenerateLink()
	if err != nil {
		return err
	}
	if blockLink.Cid != headCID {
		return ErrBlockCIDMismatch
	}

	// Calls to syncDAG should not overlap for a given CID. If they do, they will use the same
	// underlying pubsub topic and this brings along potential pitfalls. One of them being that
	// if this initial sync call had a negative response for a given link, the subsequent calls will
	// assume a negative response for that same link without retrying.
	p.processQueue.add(headCID)
	done := p.processQueue.doneOnce(headCID)
	defer done()

	// Check if we've already merged this block. If so, skip the sink process.
	isMerged, err := p.db.Multistore().Blockstore().IsMerged(ctx, headCID)
	if err != nil {
		return err
	}
	if isMerged {
		return nil
	}

	// No need to check access if the message is for replication as the node sending
	// will have done so deliberately.
	if !isReplicator {
		mightHaveAccess, err := p.trySelfHasAccess(ctx, headCID, block, req.CollectionID, req.DocID)
		if err != nil {
			return err
		}
		if !mightHaveAccess {
			// If we know we don't have access, we can skip the rest of the processing.
			return nil
		}
	}

	err = p.syncDAG(ctx, block)
	if err != nil {
		return err
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
		// We don't need to return the error for this side-effect-function call.
		// It's a bonus action that shouldn't affect the caller of `processPuslogRequest`.
		log.ErrorE("Failed to send update after sync", err, slog.Any("PeerID", p.host.ID()))
	}

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
			return NewErrPublishingToCollectionTopic(err, evt.Cid.String(), evt.CollectionID)
		}
	}

	return nil
}

// processQueue is synchronization source to ensure that concurrent
// document merges do not cause transaction conflicts.
type processQueue struct {
	cids  map[cid.Cid]chan struct{}
	mutex sync.Mutex
}

func newProcessQueue() *processQueue {
	return &processQueue{
		cids: make(map[cid.Cid]chan struct{}),
	}
}

// add adds a cid to the queue. If the cid is already in the queue, it will
// wait for the cid to be removed from the queue. For every add call, done must
// be called to remove the cid from the queue. Otherwise, subsequent add calls will
// block forever.
func (m *processQueue) add(cid cid.Cid) {
	for {
		m.mutex.Lock()
		done, ok := m.cids[cid]
		if !ok {
			m.cids[cid] = make(chan struct{})
			m.mutex.Unlock()
			return
		}
		m.mutex.Unlock()
		<-done
	}
}

func (m *processQueue) done(cid cid.Cid) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	done, ok := m.cids[cid]
	if ok {
		delete(m.cids, cid)
		close(done)
	}
}

// doneOnce returns a function that invokes done only once.
func (m *processQueue) doneOnce(cid cid.Cid) func() {
	return sync.OnceFunc(func() {
		m.done(cid)
	})
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
