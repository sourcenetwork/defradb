// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package net

import (
	"context"
	"sync"
	"time"

	"github.com/ipfs/boxo/bitswap"
	"github.com/ipfs/boxo/bitswap/network"
	"github.com/ipfs/boxo/blockservice"
	exchange "github.com/ipfs/boxo/exchange"
	dag "github.com/ipfs/boxo/ipld/merkledag"
	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	ipld "github.com/ipfs/go-ipld-format"
	gostream "github.com/libp2p/go-libp2p-gostream"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	peerstore "github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/sourcenetwork/corelog"
	"google.golang.org/grpc"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/events"
	"github.com/sourcenetwork/defradb/internal/core"
	corenet "github.com/sourcenetwork/defradb/internal/core/net"
	"github.com/sourcenetwork/defradb/internal/merkle/clock"
	pb "github.com/sourcenetwork/defradb/net/pb"
)

var (
	numWorkers = 5
)

// Peer is a DefraDB Peer node which exposes all the LibP2P host/peer functionality
// to the underlying DefraDB instance.
type Peer struct {
	//config??

	db            client.DB
	updateChannel chan events.Update

	host host.Host
	dht  routing.Routing
	ps   *pubsub.PubSub

	server *server
	p2pRPC *grpc.Server // rpc server over the P2P network

	// Used to close the dagWorker pool for a given document.
	// The string represents a docID.
	closeJob chan string
	sendJobs chan *dagJob

	// outstanding log request currently being processed
	queuedChildren *cidSafeSet

	// replicators is a map from collectionName => peerId
	replicators map[string]map[peer.ID]struct{}
	mu          sync.Mutex

	// peer DAG service
	ipld.DAGService
	exch  exchange.Interface
	bserv blockservice.BlockService

	ctx    context.Context
	cancel context.CancelFunc
}

// NewPeer creates a new instance of the DefraDB server as a peer-to-peer node.
func NewPeer(
	ctx context.Context,
	db client.DB,
	h host.Host,
	dht routing.Routing,
	ps *pubsub.PubSub,
	serverOptions []grpc.ServerOption,
	dialOptions []grpc.DialOption,
) (*Peer, error) {
	if db == nil {
		return nil, ErrNilDB
	}

	ctx, cancel := context.WithCancel(ctx)
	p := &Peer{
		host:           h,
		dht:            dht,
		ps:             ps,
		db:             db,
		p2pRPC:         grpc.NewServer(serverOptions...),
		ctx:            ctx,
		cancel:         cancel,
		closeJob:       make(chan string),
		sendJobs:       make(chan *dagJob),
		replicators:    make(map[string]map[peer.ID]struct{}),
		queuedChildren: newCidSafeSet(),
	}
	var err error
	p.server, err = newServer(p, db, dialOptions...)
	if err != nil {
		return nil, err
	}

	err = p.loadReplicators(p.ctx)
	if err != nil {
		return nil, err
	}

	p.setupBlockService()
	p.setupDAGService()

	return p, nil
}

// Start all the internal workers/goroutines/loops that manage the P2P state.
func (p *Peer) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// reconnect to known peers
	var wg sync.WaitGroup
	for _, id := range p.host.Peerstore().PeersWithAddrs() {
		if id == p.host.ID() {
			continue
		}
		wg.Add(1)
		go func(id peer.ID) {
			defer wg.Done()
			addr := p.host.Peerstore().PeerInfo(id)
			err := p.host.Connect(p.ctx, addr)
			if err != nil {
				log.InfoContext(
					p.ctx,
					"Failure while reconnecting to a known peer",
					corelog.Any("peer", id),
					corelog.Any("error", err),
				)
			}
		}(id)
	}
	wg.Wait()

	p2plistener, err := gostream.Listen(p.host, corenet.Protocol)
	if err != nil {
		return err
	}

	if p.ps != nil {
		if !p.db.Events().Updates.HasValue() {
			return ErrNilUpdateChannel
		}

		updateChannel, err := p.db.Events().Updates.Value().Subscribe()
		if err != nil {
			return err
		}
		p.updateChannel = updateChannel

		log.InfoContext(p.ctx, "Starting internal broadcaster for pubsub network")
		go p.handleBroadcastLoop()
	}

	log.InfoContext(
		p.ctx,
		"Starting P2P node",
		corelog.Any("P2P addresses", p.host.Addrs()))
	// register the P2P gRPC server
	go func() {
		pb.RegisterServiceServer(p.p2pRPC, p.server)
		if err := p.p2pRPC.Serve(p2plistener); err != nil &&
			!errors.Is(err, grpc.ErrServerStopped) {
			log.ErrorContextE(p.ctx, "Fatal P2P RPC server error", err)
		}
	}()

	// start sendJobWorker
	go p.sendJobWorker()

	return nil
}

// Close the peer node and all its internal workers/goroutines/loops.
func (p *Peer) Close() {
	// close topics
	if err := p.server.removeAllPubsubTopics(); err != nil {
		log.ErrorContextE(p.ctx, "Error closing pubsub topics", err)
	}

	// stop gRPC server
	for _, c := range p.server.conns {
		if err := c.Close(); err != nil {
			log.ErrorContextE(p.ctx, "Failed closing server RPC connections", err)
		}
	}
	stopGRPCServer(p.ctx, p.p2pRPC)
	// stopGRPCServer(p.tcpRPC)

	// close event emitters
	if p.server.pubSubEmitter != nil {
		if err := p.server.pubSubEmitter.Close(); err != nil {
			log.InfoContext(p.ctx, "Could not close pubsub event emitter", corelog.Any("Error", err.Error()))
		}
	}
	if p.server.pushLogEmitter != nil {
		if err := p.server.pushLogEmitter.Close(); err != nil {
			log.InfoContext(p.ctx, "Could not close push log event emitter", corelog.Any("Error", err.Error()))
		}
	}

	if p.db.Events().Updates.HasValue() {
		p.db.Events().Updates.Value().Unsubscribe(p.updateChannel)
	}

	if err := p.bserv.Close(); err != nil {
		log.ErrorContextE(p.ctx, "Error closing block service", err)
	}

	if err := p.host.Close(); err != nil {
		log.ErrorContextE(p.ctx, "Error closing host", err)
	}

	p.cancel()
}

// handleBroadcast loop manages the transition of messages
// from the internal broadcaster to the external pubsub network
func (p *Peer) handleBroadcastLoop() {
	for {
		update, isOpen := <-p.updateChannel
		if !isOpen {
			return
		}

		// check log priority, 1 is new doc log
		// 2 is update log
		var err error
		if update.Priority == 1 {
			err = p.handleDocCreateLog(update)
		} else if update.Priority > 1 {
			err = p.handleDocUpdateLog(update)
		} else {
			log.InfoContext(p.ctx, "Skipping log with invalid priority of 0", corelog.Any("CID", update.Cid))
		}

		if err != nil {
			log.ErrorContextE(p.ctx, "Error while handling broadcast log", err)
		}
	}
}

// RegisterNewDocument registers a new document with the peer node.
func (p *Peer) RegisterNewDocument(
	ctx context.Context,
	docID client.DocID,
	c cid.Cid,
	nd ipld.Node,
	schemaRoot string,
) error {
	// register topic
	if err := p.server.addPubSubTopic(docID.String(), !p.server.hasPubSubTopic(schemaRoot)); err != nil {
		log.ErrorContextE(
			p.ctx,
			"Failed to create new pubsub topic",
			err,
			corelog.String("DocID", docID.String()),
		)
		return err
	}

	// publish log
	body := &pb.PushLogRequest_Body{
		DocID:      []byte(docID.String()),
		Cid:        c.Bytes(),
		SchemaRoot: []byte(schemaRoot),
		Creator:    p.host.ID().String(),
		Log: &pb.Document_Log{
			Block: nd.RawData(),
		},
	}
	req := &pb.PushLogRequest{
		Body: body,
	}

	return p.server.publishLog(p.ctx, schemaRoot, req)
}

func (p *Peer) pushToReplicator(
	ctx context.Context,
	txn datastore.Txn,
	collection client.Collection,
	docIDsCh <-chan client.DocIDResult,
	pid peer.ID,
) {
	for docIDResult := range docIDsCh {
		if docIDResult.Err != nil {
			log.ErrorContextE(ctx, "Key channel error", docIDResult.Err)
			continue
		}
		docID := core.DataStoreKeyFromDocID(docIDResult.ID)
		headset := clock.NewHeadSet(
			txn.Headstore(),
			docID.WithFieldId(core.COMPOSITE_NAMESPACE).ToHeadStoreKey(),
		)
		cids, priority, err := headset.List(ctx)
		if err != nil {
			log.ErrorContextE(
				ctx,
				"Failed to get heads",
				err,
				corelog.String("DocID", docIDResult.ID.String()),
				corelog.Any("PeerID", pid),
				corelog.Any("Collection", collection.Name()))
			continue
		}
		// loop over heads, get block, make the required logs, and send
		for _, c := range cids {
			blk, err := txn.DAGstore().Get(ctx, c)
			if err != nil {
				log.ErrorContextE(ctx, "Failed to get block", err,
					corelog.Any("CID", c),
					corelog.Any("PeerID", pid),
					corelog.Any("Collection", collection.Name()))
				continue
			}

			// @todo: remove encode/decode loop for core.Log data
			nd, err := dag.DecodeProtobuf(blk.RawData())
			if err != nil {
				log.ErrorContextE(ctx, "Failed to decode protobuf", err, corelog.Any("CID", c))
				continue
			}

			evt := events.Update{
				DocID:      docIDResult.ID.String(),
				Cid:        c,
				SchemaRoot: collection.SchemaRoot(),
				Block:      nd,
				Priority:   priority,
			}
			if err := p.server.pushLog(ctx, evt, pid); err != nil {
				log.ErrorContextE(
					ctx,
					"Failed to replicate log",
					err,
					corelog.Any("CID", c),
					corelog.Any("PeerID", pid),
				)
			}
		}
	}
}

func (p *Peer) loadReplicators(ctx context.Context) error {
	reps, err := p.GetAllReplicators(ctx)
	if err != nil {
		return errors.Wrap("failed to get replicators", err)
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, rep := range reps {
		for _, schema := range rep.Schemas {
			if pReps, exists := p.replicators[schema]; exists {
				if _, exists := pReps[rep.Info.ID]; exists {
					continue
				}
			} else {
				p.replicators[schema] = make(map[peer.ID]struct{})
			}

			// add to replicators list
			p.replicators[schema][rep.Info.ID] = struct{}{}
		}

		// Add the destination's peer multiaddress in the peerstore.
		// This will be used during connection and stream creation by libp2p.
		p.host.Peerstore().AddAddrs(rep.Info.ID, rep.Info.Addrs, peerstore.PermanentAddrTTL)

		log.InfoContext(ctx, "loaded replicators from datastore", corelog.Any("Replicator", rep))
	}

	return nil
}

func (p *Peer) loadP2PCollections(ctx context.Context) (map[string]struct{}, error) {
	collections, err := p.GetAllP2PCollections(ctx)
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return nil, err
	}
	colMap := make(map[string]struct{})
	for _, col := range collections {
		err := p.server.addPubSubTopic(col, true)
		if err != nil {
			return nil, err
		}
		colMap[col] = struct{}{}
	}

	return colMap, nil
}

func (p *Peer) handleDocCreateLog(evt events.Update) error {
	docID, err := client.NewDocIDFromString(evt.DocID)
	if err != nil {
		return NewErrFailedToGetDocID(err)
	}

	// We need to register the document before pushing to the replicators if we want to
	// ensure that we have subscribed to the topic.
	err = p.RegisterNewDocument(p.ctx, docID, evt.Cid, evt.Block, evt.SchemaRoot)
	if err != nil {
		return err
	}
	// push to each peer (replicator)
	p.pushLogToReplicators(evt)

	return nil
}

func (p *Peer) handleDocUpdateLog(evt events.Update) error {
	docID, err := client.NewDocIDFromString(evt.DocID)
	if err != nil {
		return NewErrFailedToGetDocID(err)
	}

	body := &pb.PushLogRequest_Body{
		DocID:      []byte(docID.String()),
		Cid:        evt.Cid.Bytes(),
		SchemaRoot: []byte(evt.SchemaRoot),
		Creator:    p.host.ID().String(),
		Log: &pb.Document_Log{
			Block: evt.Block.RawData(),
		},
	}
	req := &pb.PushLogRequest{
		Body: body,
	}

	// push to each peer (replicator)
	p.pushLogToReplicators(evt)

	if err := p.server.publishLog(p.ctx, evt.DocID, req); err != nil {
		return NewErrPublishingToDocIDTopic(err, evt.Cid.String(), evt.DocID)
	}

	if err := p.server.publishLog(p.ctx, evt.SchemaRoot, req); err != nil {
		return NewErrPublishingToSchemaTopic(err, evt.Cid.String(), evt.SchemaRoot)
	}

	return nil
}

func (p *Peer) pushLogToReplicators(lg events.Update) {
	// push to each peer (replicator)
	peers := make(map[string]struct{})
	for _, peer := range p.ps.ListPeers(lg.DocID) {
		peers[peer.String()] = struct{}{}
	}
	for _, peer := range p.ps.ListPeers(lg.SchemaRoot) {
		peers[peer.String()] = struct{}{}
	}

	p.mu.Lock()
	reps, exists := p.replicators[lg.SchemaRoot]
	p.mu.Unlock()

	if exists {
		for pid := range reps {
			// Don't push if pid is in the list of peers for the topic.
			// It will be handled by the pubsub system.
			if _, ok := peers[pid.String()]; ok {
				continue
			}
			go func(peerID peer.ID) {
				if err := p.server.pushLog(p.ctx, lg, peerID); err != nil {
					log.ErrorContextE(
						p.ctx,
						"Failed pushing log",
						err,
						corelog.String("DocID", lg.DocID),
						corelog.Any("CID", lg.Cid),
						corelog.Any("PeerID", peerID))
				}
			}(pid)
		}
	}
}

func (p *Peer) setupBlockService() {
	bswapnet := network.NewFromIpfsHost(p.host, p.dht)
	bswap := bitswap.New(p.ctx, bswapnet, p.db.Blockstore())
	p.bserv = blockservice.New(p.db.Blockstore(), bswap)
	p.exch = bswap
}

func (p *Peer) setupDAGService() {
	p.DAGService = dag.NewDAGService(p.bserv)
}

func (p *Peer) newDAGSyncerTxn(txn datastore.Txn) ipld.DAGService {
	return dag.NewDAGService(blockservice.New(txn.DAGstore(), p.exch))
}

// Session returns a session-based NodeGetter.
func (p *Peer) Session(ctx context.Context) ipld.NodeGetter {
	ng := dag.NewSession(ctx, p.DAGService)
	if ng == p.DAGService {
		log.InfoContext(ctx, "DAGService does not support sessions")
	}
	return ng
}

func stopGRPCServer(ctx context.Context, server *grpc.Server) {
	stopped := make(chan struct{})
	go func() {
		server.GracefulStop()
		close(stopped)
	}()
	timer := time.NewTimer(10 * time.Second)
	select {
	case <-timer.C:
		server.Stop()
		log.InfoContext(ctx, "Peer gRPC server was shutdown ungracefully")
	case <-stopped:
		timer.Stop()
	}
}

type EvtReceivedPushLog struct {
	ByPeer   peer.ID
	FromPeer peer.ID
}

type EvtPubSub struct {
	Peer peer.ID
}

// rollbackAddPubSubTopics removes the given topics from the pubsub system.
func (p *Peer) rollbackAddPubSubTopics(topics []string, cause error) error {
	for _, topic := range topics {
		if err := p.server.removePubSubTopic(topic); err != nil {
			return errors.WithStack(err, errors.NewKV("Cause", cause))
		}
	}
	return cause
}

// rollbackRemovePubSubTopics adds back the given topics from the pubsub system.
func (p *Peer) rollbackRemovePubSubTopics(topics []string, cause error) error {
	for _, topic := range topics {
		if err := p.server.addPubSubTopic(topic, true); err != nil {
			return errors.WithStack(err, errors.NewKV("Cause", cause))
		}
	}
	return cause
}
