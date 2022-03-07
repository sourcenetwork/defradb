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
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ipfs/go-cid"
	dag "github.com/ipfs/go-merkledag"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	peerstore "github.com/libp2p/go-libp2p-core/peerstore"
	gostream "github.com/libp2p/go-libp2p-gostream"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/textileio/go-threads/broadcast"
	"google.golang.org/grpc"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	corenet "github.com/sourcenetwork/defradb/core/net"
	"github.com/sourcenetwork/defradb/document/key"
	"github.com/sourcenetwork/defradb/logging"
	"github.com/sourcenetwork/defradb/merkle/clock"
	pb "github.com/sourcenetwork/defradb/net/pb"
)

var (
	numWorkers = 5
)

// Peer is a DefraDB Peer node which exposes all the LibP2P host/peer functionality
// to the underlying DefraDB instance.
type Peer struct {
	//config??

	db client.DB

	host host.Host
	ps   *pubsub.PubSub
	ds   DAGSyncer

	server *server
	p2pRPC *grpc.Server // rpc server over the p2p network

	bus *broadcast.Broadcaster

	jobQueue chan *dagJob
	sendJobs chan *dagJob

	// outstanding log request currently being processed
	queuedChildren *cidSafeSet

	// replicators is a map from collectionName => peerId
	replicators map[string]map[peer.ID]struct{}
	mu          sync.Mutex

	ctx    context.Context
	cancel context.CancelFunc
}

// NewPeer creates a new instance of the DefraDB server as a peer-to-peer node.
func NewPeer(
	ctx context.Context,
	db client.DB,
	h host.Host,
	ps *pubsub.PubSub,
	bs *broadcast.Broadcaster,
	ds DAGSyncer,
	tcpAddr ma.Multiaddr,
	serverOptions []grpc.ServerOption,
	dialOptions []grpc.DialOption,
) (*Peer, error) {
	if db == nil {
		return nil, fmt.Errorf("Database object can't be empty")
	}

	ctx, cancel := context.WithCancel(ctx)
	p := &Peer{
		host:           h,
		ps:             ps,
		db:             db,
		ds:             ds,
		bus:            bs,
		p2pRPC:         grpc.NewServer(serverOptions...),
		ctx:            ctx,
		cancel:         cancel,
		jobQueue:       make(chan *dagJob, numWorkers),
		sendJobs:       make(chan *dagJob),
		replicators:    make(map[string]map[peer.ID]struct{}),
		queuedChildren: newCidSafeSet(),
	}
	var err error
	p.server, err = newServer(p, db, dialOptions...)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// Start all the internal workers/goroutines/loops that manage the P2P
// state
func (p *Peer) Start() error {
	p2plistener, err := gostream.Listen(p.host, corenet.Protocol)
	if err != nil {
		return err
	}

	if p.ps != nil {
		log.Info(p.ctx, "Starting internal broadcaster for pubsub network")
		go p.handleBroadcastLoop()
	}

	// register the p2p gRPC server
	go func() {
		pb.RegisterServiceServer(p.p2pRPC, p.server)
		if err := p.p2pRPC.Serve(p2plistener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			log.FatalE(p.ctx, "Fatal p2p rpc serve error", err)
		}
	}()

	// start sendJobWorker + NumWorkers goroutines
	go p.sendJobWorker()
	for i := 0; i < numWorkers; i++ {
		go p.dagWorker()
	}

	return nil
}

func (p *Peer) Close() error {
	// close topics
	if err := p.server.removeAllPubsubTopics(); err != nil {
		log.ErrorE(p.ctx, "Error closing pubsub topics", err)
	}

	// stop gRPC server
	for _, c := range p.server.conns {
		if err := c.Close(); err != nil {
			log.ErrorE(p.ctx, "Failed closing server RPC connections", err)
		}
	}
	stopGRPCServer(p.ctx, p.p2pRPC)
	// stopGRPCServer(p.tcpRPC)

	p.bus.Discard()
	p.cancel()
	return nil
}

// handleBroadcast loop manages the transition of messages
// from the internal broadcaster to the external pubsub network
func (p *Peer) handleBroadcastLoop() {
	if p.bus == nil {
		log.Warn(p.ctx, "Tried to start internal broadcaster with none defined")
		return
	}

	l := p.bus.Listen()
	log.Debug(p.ctx, "Waiting for messages on internal broadcaster")
	for v := range l.Channel() {
		log.Debug(p.ctx, "Handling internal broadcast bus message")
		// filter for only messages intended for the pubsub network
		switch msg := v.(type) {
		case core.Log:

			// check log priority, 1 is new doc log
			// 2 is update log
			var err error
			if msg.Priority == 1 {
				err = p.handleDocCreateLog(msg)
			} else if msg.Priority > 1 {
				err = p.handleDocUpdateLog(msg)
			} else {
				log.Warn(p.ctx, "Skipping log with invalid priority of 0", logging.NewKV("Cid", msg.Cid))
			}

			if err != nil {
				log.ErrorE(p.ctx, "Error while handling broadcast log", err)
			}
		}
	}
}

func (p *Peer) RegisterNewDocument(ctx context.Context, dockey key.DocKey, c cid.Cid, schemaID string) error {
	log.Debug(p.ctx, "Registering a new document for our peer node", logging.NewKV("DocKey", dockey.String()))

	block, err := p.db.GetBlock(ctx, c)
	if err != nil {
		log.ErrorE(p.ctx, "Failed to get document cid", err)
		return err
	}

	// register topic
	if err := p.server.addPubSubTopic(dockey.String()); err != nil {
		log.ErrorE(p.ctx, "Failed to create new pubsub topic", err, logging.NewKV("DocKey", dockey.String()))
		return err
	}

	// publish log
	body := &pb.PushLogRequest_Body{
		DocKey:   &pb.ProtoDocKey{DocKey: dockey},
		Cid:      &pb.ProtoCid{Cid: c},
		SchemaID: []byte(schemaID),
		Log: &pb.Document_Log{
			Block: block.RawData(),
		},
	}
	req := &pb.PushLogRequest{
		Body: body,
	}

	return p.server.publishLog(p.ctx, dockey.String(), req)
}

// AddReplicator adds a target peer node as a replication destination for documents in our DB
func (p *Peer) AddReplicator(ctx context.Context, collectionName string, paddr ma.Multiaddr) (peer.ID, error) {
	var pid peer.ID

	// verify collection
	col, err := p.db.GetCollectionByName(ctx, collectionName)
	if err != nil {
		return pid, fmt.Errorf("Failed to get collection for replicator: %w", err)
	}

	// extra peerID
	// Extract peer portion
	p2p, err := paddr.ValueForProtocol(ma.P_P2P)
	if err != nil {
		return pid, err
	}
	pid, err = peer.Decode(p2p)
	if err != nil {
		return pid, err
	}

	// make sure it's not ourselves
	if pid == p.host.ID() {
		return pid, fmt.Errorf("Can't target ourselves as a replicator")
	}

	// make sure we're not duplicating things
	p.mu.Lock()
	defer p.mu.Unlock()
	if reps, exists := p.replicators[col.SchemaID()]; exists {
		if _, exists := reps[pid]; exists {
			return pid, fmt.Errorf("Replicator already exists for %s with ID %s", collectionName, pid)
		}
	} else {
		p.replicators[col.SchemaID()] = make(map[peer.ID]struct{})
	}

	// add peer to peerstore
	// Extract the peer ID from the multiaddr.
	info, err := peer.AddrInfoFromP2pAddr(paddr)
	if err != nil {
		return pid, fmt.Errorf("Failed to address info from %s: %w", paddr, err)
	}

	// Add the destination's peer multiaddress in the peerstore.
	// This will be used during connection and stream creation by libp2p.
	p.host.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)

	// add to replicators list
	p.replicators[col.SchemaID()][pid] = struct{}{}

	// create read only txn and assign to col
	txn, err := p.db.NewTxn(ctx, true)
	if err != nil {
		return pid, fmt.Errorf("Failed to get txn: %w", err)
	}
	col = col.WithTxn(txn)

	// get dockeys (all)
	keysCh, err := col.GetAllDocKeys(ctx)
	if err != nil {
		txn.Discard(ctx)
		return pid, fmt.Errorf("Failed to get dockey for replicator %s on %s: %w", pid, collectionName, err)
	}

	// async
	// get all keys and push
	// -> get head
	// -> pushLog(head.block)
	go func() {
		defer txn.Discard(ctx)
		for key := range keysCh {
			if key.Err != nil {
				log.ErrorE(p.ctx, "Key channel error", key.Err)
				continue
			}
			dockey := core.DataStoreKeyFromDocKey(key.Key)
			headset := clock.NewHeadSet(txn.Headstore(), dockey.WithFieldId(core.COMPOSITE_NAMESPACE).ToHeadStoreKey())
			cids, priority, err := headset.List(ctx)
			if err != nil {
				log.ErrorE(
					p.ctx,
					"Failed to get heads",
					err,
					logging.NewKV("DocKey", dockey),
					logging.NewKV("Pid", pid),
					logging.NewKV("Collection", collectionName))
				continue
			}
			// loop over heads, get block, make the required logs, and send
			for _, c := range cids {
				blk, err := txn.DAGstore().Get(ctx, c)
				if err != nil {
					log.ErrorE(p.ctx, "Failed to get block", err,
						logging.NewKV("Cid", c),
						logging.NewKV("Pid", pid),
						logging.NewKV("Collection", collectionName))
					continue
				}

				// @todo: remove encode/decode loop for core.Log data
				nd, err := dag.DecodeProtobuf(blk.RawData())
				if err != nil {
					log.ErrorE(p.ctx, "Failed to decode protobuf", err, logging.NewKV("Cid", c))
					continue
				}

				lg := core.Log{
					DocKey:   dockey.ToString(),
					Cid:      c,
					SchemaID: col.SchemaID(),
					Block:    nd,
					Priority: priority,
				}
				if err := p.server.pushLog(ctx, lg, pid); err != nil {
					log.ErrorE(p.ctx, "Failed to replicate log", err, logging.NewKV("Cid", c), logging.NewKV("Pid", pid))
				}
			}
		}
	}()

	return pid, nil
}

func (p *Peer) handleDocCreateLog(lg core.Log) error {
	dockey, err := key.NewFromString(lg.DocKey)
	if err != nil {
		return fmt.Errorf("Failed to get DocKey from broadcast message: %w", err)
	}

	// push to each peer (replicator)
	p.pushLogToReplicators(p.ctx, lg)

	return p.RegisterNewDocument(p.ctx, dockey, lg.Cid, lg.SchemaID)
}

func (p *Peer) handleDocUpdateLog(lg core.Log) error {
	dockey, err := key.NewFromString(lg.DocKey)
	if err != nil {
		return fmt.Errorf("Failed to get DocKey from broadcast message: %w", err)
	}
	log.Debug(
		p.ctx,
		"Preparing pubsub pushLog request from broadcast",
		logging.NewKV("DocKey", dockey),
		logging.NewKV("Cid", lg.Cid),
		logging.NewKV("SchemaId", lg.SchemaID))

	body := &pb.PushLogRequest_Body{
		DocKey:   &pb.ProtoDocKey{DocKey: dockey},
		Cid:      &pb.ProtoCid{Cid: lg.Cid},
		SchemaID: []byte(lg.SchemaID),
		Log: &pb.Document_Log{
			Block: lg.Block.RawData(),
		},
	}
	req := &pb.PushLogRequest{
		Body: body,
	}

	// push to each peer (replicator)
	p.pushLogToReplicators(p.ctx, lg)

	if err := p.server.publishLog(p.ctx, lg.DocKey, req); err != nil {
		return fmt.Errorf("Error publishing log %s for %s: %w", lg.Cid, lg.DocKey, err)
	}
	return nil
}

func (p *Peer) pushLogToReplicators(ctx context.Context, lg core.Log) {
	// push to each peer (replicator)
	if reps, exists := p.replicators[lg.SchemaID]; exists {
		for pid := range reps {
			go func(peerID peer.ID) {
				if err := p.server.pushLog(p.ctx, lg, peerID); err != nil {
					log.ErrorE(
						p.ctx,
						"Failed pushing log",
						err,
						logging.NewKV("DocKey", lg.DocKey),
						logging.NewKV("Cid", lg.Cid),
						logging.NewKV("PeerId", peerID))
				}
			}(pid)
		}
	}
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
		log.Warn(ctx, "peer GRPC server was shutdown ungracefully")
	case <-stopped:
		timer.Stop()
	}
}
