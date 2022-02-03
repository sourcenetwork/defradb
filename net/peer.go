package net

import (
	"context"
	"errors"
	"fmt"
	"sync"

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
	busBufferSize = 100

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
func NewPeer(ctx context.Context, db client.DB, h host.Host, ps *pubsub.PubSub, ds DAGSyncer, serverOptions []grpc.ServerOption, dialOptions []grpc.DialOption) (*Peer, error) {
	ctx, cancel := context.WithCancel(ctx)
	if db == nil {
		return nil, fmt.Errorf("Database object can't be empty")
	}
	p := &Peer{
		host:           h,
		ps:             ps,
		db:             db,
		ds:             ds,
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

func (p *Peer) Start() error {
	listener, err := gostream.Listen(p.host, corenet.Protocol)
	if err != nil {
		return err
	}

	if p.ps != nil {
		log.Info("Starting internal broadcaster for pubsub network")
		p.bus = broadcast.NewBroadcaster(busBufferSize)
		p.server.db.SetBroadcaster(p.bus)
		go p.handleBroadcastLoop()
	}

	go func() {
		pb.RegisterServiceServer(p.rpc, p.server)
		if err := p.rpc.Serve(listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			log.Fatal("Fatal serve error:", err)
		}
	}()

	// sendJobWorker + NumWorkers
	p.wg.Add(1 + numWorkers)
	go func() {
		defer p.wg.Done()
		p.sendJobWorker()
	}()
	for i := 0; i < numWorkers; i++ {
		go func() {
			defer p.wg.Done()
			p.dagWorker()
		}()
	}

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
			dockey, err := key.NewFromString(msg.DocKey)
			if err != nil {
				log.Error("Failed to get DocKeyfrom broadcast message:", err)
				continue
			}
			log.Debugf("Preparing pubsub pushLog request from broadcast for %s at %s using %s", dockey, msg.Cid, msg.SchemaID)
			body := &pb.PushLogRequest_Body{
				DocKey:   &pb.ProtoDocKey{DocKey: dockey},
				Cid:      &pb.ProtoCid{Cid: msg.Cid},
				SchemaID: []byte(msg.SchemaID),
				Log: &pb.Document_Log{
					Block: msg.Block.RawData(),
				},
			}
			req := &pb.PushLogRequest{
				Body: body,
			}

			// @todo: push to each peer

			if err := p.server.publishLog(p.ctx, msg.DocKey, req); err != nil {
				log.Errorf("Error publishing log %s for %s: %s", msg.Cid, msg.DocKey, err)
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
