package net

import (
	"context"

	"github.com/sourcenetwork/defradb/client"
	"google.golang.org/grpc"

	format "github.com/ipfs/go-ipld-format"
	"github.com/libp2p/go-libp2p-core/host"
)

// Peer is a DefraDB Peer node which exposes all the LibP2P host/peer functionality
// to the underlying DefraDB instance.
type peer struct {
	//config??

	format.DAGService
	host host.Host

	db client.DB

	rpc    *grpc.Server
	server *server

	ctx context.Context
}
