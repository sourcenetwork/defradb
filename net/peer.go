package net

import (
	"context"
	"errors"
	"fmt"

	"github.com/ipfs/go-cid"
	format "github.com/ipfs/go-ipld-format"
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
)

// Peer is a DefraDB Peer node which exposes all the LibP2P host/peer functionality
// to the underlying DefraDB instance.
type Peer struct {
	//config??

	db client.DB

	host host.Host
	ps   *pubsub.PubSub

	rpc    *grpc.Server
	server *server

	bus *broadcast.Broadcaster

	ctx    context.Context
	cancel context.CancelFunc
}

// NewPeer creates a new instance of the DefraDB server as a peer-to-peer node.
func NewPeer(ctx context.Context, db client.DB, h host.Host, ps *pubsub.PubSub, ds format.DAGService, serverOptions []grpc.ServerOption, dialOptions []grpc.DialOption) (*Peer, error) {
	ctx, cancel := context.WithCancel(ctx)
	if db == nil {
		return nil, fmt.Errorf("Database object can't be empty")
	}
	p := &Peer{
		host:   h,
		ps:     ps,
		db:     db,
		rpc:    grpc.NewServer(serverOptions...),
		ctx:    ctx,
		cancel: cancel,
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
		log.Debug("Handling internal broadcat bus message")
		// filter for only messages intended for the pubsub network
		switch msg := v.(type) {
		case core.Log:
			dockey, err := key.NewFromString(msg.DocKey)
			if err != nil {
				log.Error("Failed to get DocKeyfrom broadcast message:", err)
				continue
			}
			body := &pb.PushLogRequest_Body{
				DocKey: &pb.ProtoDocKey{DocKey: dockey},
				Cid:    &pb.ProtoCid{Cid: msg.Cid},
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

func (p *Peer) RegisterNewDocument(ctx context.Context, dockey key.DocKey, c cid.Cid) error {
	log.Debug("Registering a new document with for our peer node: ", dockey.String())

	block, err := p.db.DAGStore().Get(ctx, c)
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
		DocKey: &pb.ProtoDocKey{DocKey: dockey},
		Cid:    &pb.ProtoCid{Cid: c},
		Log: &pb.Document_Log{
			Block: block.RawData(),
		},
	}
	req := &pb.PushLogRequest{
		Body: body,
	}

	return p.server.publishLog(p.ctx, dockey.String(), req)
}
