package net

import (
	"context"
	"errors"

	format "github.com/ipfs/go-ipld-format"
	"github.com/libp2p/go-libp2p-core/host"
	gostream "github.com/libp2p/go-libp2p-gostream"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"google.golang.org/grpc"

	"github.com/sourcenetwork/defradb/client"
	corenet "github.com/sourcenetwork/defradb/core/net"
	pb "github.com/sourcenetwork/defradb/net/pb"
)

// Peer is a DefraDB Peer node which exposes all the LibP2P host/peer functionality
// to the underlying DefraDB instance.
type peer struct {
	//config??

	host host.Host
	ps   *pubsub.PubSub

	rpc    *grpc.Server
	server *server

	ctx    context.Context
	cancel context.CancelFunc
}

// NewPeer creates a new instance of the DefraDB server as a peer-to-peer node.
func NewPeer(ctx context.Context, db client.DB, h host.Host, ps *pubsub.PubSub, ds format.DAGService, serverOptions []grpc.ServerOption, dialOptions []grpc.DialOption) (*peer, error) {
	ctx, cancel := context.WithCancel(ctx)
	p := &peer{
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

	go func() {
		pb.RegisterServiceServer(p.rpc, p.server)
		if err := p.rpc.Serve(listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			// @todo: Log fatal
		}
	}()
	return p, nil
}
