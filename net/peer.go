package net

import (
	"context"

	format "github.com/ipfs/go-ipld-format"
	"github.com/libp2p/go-libp2p-core/host"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"google.golang.org/grpc"
)

// Peer is a DefraDB Peer node which exposes all the LibP2P host/peer functionality
// to the underlying DefraDB instance.
type peer struct {
	//config??

	format.DAGService
	host host.Host
	ps   *pubsub.PubSub

	rpc    *grpc.Server
	server *server

	ctx context.Context
}
