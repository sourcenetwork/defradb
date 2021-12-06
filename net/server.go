package net

import (
	"sync"

	libpeer "github.com/libp2p/go-libp2p-core/peer"
	rpc "github.com/textileio/go-libp2p-pubsub-rpc"
	"google.golang.org/grpc"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/document/key"
)

// Server is the request/response instance for all P2P RPC communication.
// Implements gRPC server. See net/pb/net.proto for corresponding service definitions.
//
// Specifically, server handles the push/get request/response aspects of the RPC service
// but not the API calls.
type server struct {
	peer *peer
	opts []grpc.DialOption
	db   client.DB

	topics map[key.DocKey]*rpc.Topic

	conns map[libpeer.ID]*grpc.ClientConn

	sync.Mutex
}

// GetDocGraph recieves a get graph request
// func (s *server) GetDocGraph( GetDocGraphRequest ) (GetDocGraphResponse, error)

// PushDocGraph recieves a push graph request
// func (s *server) PushDocGraph( PushDocGraphRequest ) (PushDocGraphResponse, error)
