package net

import (
	"context"
	"fmt"
	gonet "net"
	"time"

	libpeer "github.com/libp2p/go-libp2p-core/peer"
	gostream "github.com/libp2p/go-libp2p-gostream"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/status"

	corenet "github.com/sourcenetwork/defradb/core/net"
	pb "github.com/sourcenetwork/defradb/net/pb"
)

var (
	// DialTimeout is the max time duration to wait when dialing a peer.
	DialTimeout = time.Second * 10
)

// dial attempts to open a gRPC connection over libp2p to a peer.
func (s *server) dial(peerID libpeer.ID) (pb.ServiceClient, error) {
	s.topicLock.Lock()
	defer s.topicLock.Unlock()
	conn, ok := s.conns[peerID]
	if ok {
		if conn.GetState() == connectivity.Shutdown {
			if err := conn.Close(); err != nil && status.Code(err) != codes.Canceled {
				// log.Errorf("error closing connection: %v", err)
				return nil, err
			}
		} else {
			return pb.NewServiceClient(conn), nil
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), DialTimeout)
	defer cancel()
	conn, err := grpc.DialContext(ctx, peerID.Pretty(), s.opts...)
	if err != nil {
		return nil, err
	}
	s.conns[peerID] = conn
	return pb.NewServiceClient(conn), nil
}

// getLibp2pDialer returns a WithContextDialer option for libp2p dialing.
func (s *server) getLibp2pDialer() grpc.DialOption {
	return grpc.WithContextDialer(func(ctx context.Context, peerIDStr string) (gonet.Conn, error) {
		id, err := libpeer.Decode(peerIDStr)
		if err != nil {
			return nil, fmt.Errorf("grpc tried to dial non peerID: %w", err)
		}

		conn, err := gostream.Dial(ctx, s.peer.host, id, corenet.Protocol)
		if err != nil {
			return nil, fmt.Errorf("gostream dial failed: %w", err)
		}

		return conn, nil
	})
}
