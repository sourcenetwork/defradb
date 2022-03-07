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
	log.Debug("Received AddReplicator requeust")

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
