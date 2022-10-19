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

package client

import (
	"context"

	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"google.golang.org/grpc"

	"github.com/sourcenetwork/defradb/errors"
	pb "github.com/sourcenetwork/defradb/net/api/pb"
)

type Client struct {
	c    pb.ServiceClient
	conn *grpc.ClientConn
}

// NewClient returns a new defra gRPC client connected to the target address
func NewClient(target string, opts ...grpc.DialOption) (*Client, error) {
	conn, err := grpc.Dial(target, opts...)
	if err != nil {
		return nil, err
	}

	return &Client{
		c:    pb.NewServiceClient(conn),
		conn: conn,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

// AddReplicator sends a request to add a target replicator to the DB peer
func (c *Client) AddReplicator(
	ctx context.Context,
	collection string,
	paddr ma.Multiaddr,
) (peer.ID, error) {
	if len(collection) == 0 {
		return "", errors.New("Collection can't be empty")
	}
	if paddr == nil {
		return "", errors.New("target address can't be empty")
	}
	resp, err := c.c.AddReplicator(ctx, &pb.AddReplicatorRequest{
		Collection: []byte(collection),
		Addr:       paddr.Bytes(),
	})
	if err != nil {
		return "", errors.Wrap("AddReplicator request failed", err)
	}
	return peer.IDFromBytes(resp.PeerID)
}
