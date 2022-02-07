package api

import (
	"context"

	logging "github.com/ipfs/go-log/v2"
	libpeer "github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sourcenetwork/defradb/net"
	pb "github.com/sourcenetwork/defradb/net/api/pb"
)

var (
	log = logging.Logger("netapi")
)

type Service struct {
	peer *net.Peer
}

func NewService(peer *net.Peer) *Service {
	return &Service{peer: peer}
}

func (s *Service) AddReplicator(ctx context.Context, req *pb.AddReplicatorRequest) (*pb.AddReplicatorReply, error) {
	log.Debug("Recieved AddReplicator requeust")

	collection := string(req.Collection)
	if len(collection) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Collection can't be empty")
	}
	addr, err := ma.NewMultiaddrBytes(req.Addr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	pid, err := s.peer.AddReplicator(ctx, collection, addr)
	if err != nil {
		return nil, err
	}
	return &pb.AddReplicatorReply{
		PeerID: marshalPeerID(pid),
	}, nil
}

func marshalPeerID(id libpeer.ID) []byte {
	b, _ := id.Marshal() // This will never return an error
	return b
}
