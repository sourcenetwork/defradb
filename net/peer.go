package net

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/host"
	gostream "github.com/libp2p/go-libp2p-gostream"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/textileio/go-threads/broadcast"
	"google.golang.org/grpc"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	corenet "github.com/sourcenetwork/defradb/core/net"
	"github.com/sourcenetwork/defradb/document/key"
	pb "github.com/sourcenetwork/defradb/net/pb"
)

var (
	log = logging.Logger("net")

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

	rpc    *grpc.Server
	server *server

	bus *broadcast.Broadcaster

	jobQueue chan *dagJob
	sendJobs chan *dagJob

	queuedChildren *cidSafeSet

	wg sync.WaitGroup

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
	serverOptions []grpc.ServerOption,
	dialOptions []grpc.DialOption,
) (*Peer, error) {
	ctx, cancel := context.WithCancel(ctx)
	if db == nil {
		return nil, fmt.Errorf("Database object can't be empty")
	}
	p := &Peer{
		host:           h,
		ps:             ps,
		db:             db,
		ds:             ds,
		bus:            bs,
		rpc:            grpc.NewServer(serverOptions...),
		ctx:            ctx,
		cancel:         cancel,
		jobQueue:       make(chan *dagJob, numWorkers),
		sendJobs:       make(chan *dagJob),
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
	listener, err := gostream.Listen(p.host, corenet.Protocol)
	if err != nil {
		return err
	}

	if p.ps != nil {
		log.Info("Starting internal broadcaster for pubsub network")
		go p.handleBroadcastLoop()
	}

	go func() {
		pb.RegisterServiceServer(p.rpc, p.server)
		if err := p.rpc.Serve(listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			log.Fatal("Fatal serve error:", err)
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
		log.Errorf("Error closing pubsub topics: %w", err)
	}

	// stop grpc server
	for _, c := range p.server.conns {
		if err := c.Close(); err != nil {
			log.Errorf("Failed closing server RPC connections: %w", err)
		}
	}
	stopGRPCServer(p.rpc)

	p.bus.Discard()
	p.cancel()
	return nil
}

// handleBroadcast loop manages the transition of messages
// from the internal broadcaster to the external pubsub network
func (p *Peer) handleBroadcastLoop() {
	if p.bus == nil {
		log.Warn("Tried to start internal broadcaster with none defined")
		return
	}

	l := p.bus.Listen()
	log.Debug("Waiting for messages on internal broadcaster")
	for v := range l.Channel() {
		log.Debug("Handling internal broadcast bus message")
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
				log.Warnf("Skipping log %s with invalid priority of 0", msg.Cid)
			}

			if err != nil {
				log.Errorf("Error while handling broadcast log: %s", err)
			}
		}
	}
}

func (p *Peer) RegisterNewDocument(ctx context.Context, dockey key.DocKey, c cid.Cid, schemaID string) error {
	log.Debug("Registering a new document for our peer node: ", dockey.String())

	block, err := p.db.GetBlock(ctx, c)
	if err != nil {
		log.Error("Failed to get document cid: ", err)
		return err
	}

	// register topic
	if err := p.server.addPubSubTopic(dockey.String()); err != nil {
		log.Errorf("Failed to create new pubsub topic for %s: %s", dockey.String(), err)
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

func (p *Peer) handleDocCreateLog(lg core.Log) error {
	dockey, err := key.NewFromString(lg.DocKey)
	if err != nil {
		return fmt.Errorf("Failed to get DocKey from broadcast message: %w", err)
	}

	return p.RegisterNewDocument(p.ctx, dockey, lg.Cid, lg.SchemaID)
}

func (p *Peer) handleDocUpdateLog(lg core.Log) error {
	dockey, err := key.NewFromString(lg.DocKey)
	if err != nil {
		return fmt.Errorf("Failed to get DocKey from broadcast message: %w", err)
	}
	log.Debugf("Preparing pubsub pushLog request from broadcast for %s at %s using %s", dockey, lg.Cid, lg.SchemaID)
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

	// @todo: push to each peer (replicator)

	if err := p.server.publishLog(p.ctx, lg.DocKey, req); err != nil {
		return fmt.Errorf("Error publishing log %s for %s: %w", lg.Cid, lg.DocKey, err)
	}
	return nil
}

func stopGRPCServer(server *grpc.Server) {
	stopped := make(chan struct{})
	go func() {
		server.GracefulStop()
		close(stopped)
	}()
	timer := time.NewTimer(10 * time.Second)
	select {
	case <-timer.C:
		server.Stop()
		log.Warn("peer GRPC server was shutdown ungracefully")
	case <-stopped:
		timer.Stop()
	}
}
