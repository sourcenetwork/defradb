// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"context"

	libpeer "github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sourcenetwork/defradb/logging"
	"github.com/sourcenetwork/defradb/net"
	pb "github.com/sourcenetwork/defradb/net/api/pb"
)

var (
	log = logging.MustNewLogger("defra.netapi")
)

type Service struct {
	peer *net.Peer
}

func NewService(peer *net.Peer) *Service {
	return &Service{peer: peer}
}

func (s *Service) SetReplicator(
	ctx context.Context,
	req *pb.SetReplicatorRequest,
) (*pb.SetReplicatorReply, error) {
	log.Debug(ctx, "Received SetReplicator request")

	addr, err := ma.NewMultiaddrBytes(req.Addr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	pid, err := s.peer.SetReplicator(ctx, addr, req.Collections...)
	if err != nil {
		return nil, err
	}
	return &pb.SetReplicatorReply{
		PeerID: marshalPeerID(pid),
	}, nil
}

func (s *Service) DeleteReplicator(
	ctx context.Context,
	req *pb.DeleteReplicatorRequest,
) (*pb.DeleteReplicatorReply, error) {
	log.Debug(ctx, "Received DeleteReplicator request")
	err := s.peer.DeleteReplicator(ctx, libpeer.ID(req.PeerID))
	if err != nil {
		return nil, err
	}
	return &pb.DeleteReplicatorReply{
		PeerID: req.PeerID,
	}, nil
}

func (s *Service) GetAllReplicators(
	ctx context.Context,
	req *pb.GetAllReplicatorRequest,
) (*pb.GetAllReplicatorReply, error) {
	log.Debug(ctx, "Received GetAllReplicators request")

	reps, err := s.peer.GetAllReplicators(ctx)
	if err != nil {
		return nil, err
	}

	pbReps := []*pb.GetAllReplicatorReply_Replicators{}
	for _, rep := range reps {
		pbReps = append(pbReps, &pb.GetAllReplicatorReply_Replicators{
			Info: &pb.GetAllReplicatorReply_Replicators_Info{
				Id:    []byte(rep.Info.ID),
				Addrs: rep.Info.Addrs[0].Bytes(),
			},
			Schemas: rep.Schemas,
		})
	}

	return &pb.GetAllReplicatorReply{
		Replicators: pbReps,
	}, nil
}

func marshalPeerID(id libpeer.ID) []byte {
	b, _ := id.Marshal() // This will never return an error
	return b
}

// RemoveP2PCollections handles the request to add P2P collecctions to the stored list.
func (s *Service) AddP2PCollections(
	ctx context.Context,
	req *pb.AddP2PCollectionsRequest,
) (*pb.AddP2PCollectionsReply, error) {
	log.Debug(ctx, "Received AddP2PCollections request")

	return nil, s.peer.AddP2PCollections(req.Collections)
}

// RemoveP2PCollections handles the request to remove P2P collecctions from the stored list.
func (s *Service) RemoveP2PCollections(
	ctx context.Context,
	req *pb.RemoveP2PCollectionsRequest,
) (*pb.RemoveP2PCollectionsReply, error) {
	log.Debug(ctx, "Received RemoveP2PCollections request")

	return nil, s.peer.RemoveP2PCollections(req.Collections)
}

// GetAllP2PCollections handles the request to get all P2P collecctions from the stored list.
func (s *Service) GetAllP2PCollections(
	ctx context.Context,
	req *pb.GetAllP2PCollectionsRequest,
) (*pb.GetAllP2PCollectionsReply, error) {
	log.Debug(ctx, "Received GetAllP2PCollections request")
	collections, err := s.peer.GetAllP2PCollections()
	if err != nil {
		return nil, err
	}

	var pbCols []*pb.GetAllP2PCollectionsReply_Collection
	for _, col := range collections {
		pbCols = append(pbCols, &pb.GetAllP2PCollectionsReply_Collection{
			Id:   col.ID,
			Name: col.Name,
		})
	}

	return &pb.GetAllP2PCollectionsReply{
		Collections: pbCols,
	}, nil
}
