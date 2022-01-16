package net

import (
	"context"
	"errors"
	"fmt"

	format "github.com/ipfs/go-ipld-format"
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
)

// Peer is a DefraDB Peer node which exposes all the LibP2P host/peer functionality
// to the underlying DefraDB instance.
type Peer struct {
	//config??

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
	p := &Peer{
		host:   h,
		ps:     ps,
		rpc:    grpc.NewServer(serverOptions...),
		ctx:    ctx,
		cancel: cancel,
	}
	var err error
	p.server, err = newServer(p, db, dialOptions...)
	if err != nil {
		return nil, err
	}

	listener, err := gostream.Listen(h, corenet.Protocol)
	if err != nil {
		return nil, err
	}

	if ps != nil {
		p.bus = broadcast.NewBroadcaster(busBufferSize)
		db.SetBroadcaster(p.bus)
		go p.handleBroadcastLoop()
	}

	go func() {
		pb.RegisterServiceServer(p.rpc, p.server)
		if err := p.rpc.Serve(listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			// @todo: Log fatal
			fmt.Println("Fatal serve error:", err)
		}
	}()
	return p, nil
}

// handleBroadcast loop manages the transition of messages
// from the internal broadcaster to the external pubsub network
func (p *Peer) handleBroadcastLoop() {
	if p.bus == nil {
		return
	}

	l := p.bus.Listen()
	for v := range l.Channel() {
		// filter for only messages intended for the pubsub network
		switch msg := v.(type) {
		case core.Log:
			dockey, err := key.NewFromString(msg.DocKey)
			if err != nil {
				// @todo: log error
				fmt.Println("Failed to get DocKey from broadcast messsage: %s", err)
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
				//@todo Log error
				fmt.Println("Error publishing log %s for %s: %s", msg.Cid, msg.DocKey, err)
			}
		}
	}
}
