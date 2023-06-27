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
	"fmt"

	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	codec "github.com/planetscale/vtprotobuf/codec/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
	_ "google.golang.org/grpc/encoding/proto"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	pb "github.com/sourcenetwork/defradb/net/pb"
)

func init() {
	encoding.RegisterCodec(codec.Codec{})
}

type Client struct {
	c    pb.Service2Client
	conn *grpc.ClientConn
}

// NewClient returns a new defra gRPC client connected to the target address.
func NewClient(target string, opts ...grpc.DialOption) (*Client, error) {
	conn, err := grpc.Dial(target, opts...)
	if err != nil {
		return nil, err
	}

	return &Client{
		c:    pb.NewService2Client(conn),
		conn: conn,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

// SetReplicator sends a request to add a target replicator to the DB peer.
func (c *Client) SetReplicator(
	ctx context.Context,
	paddr ma.Multiaddr,
	collections ...string,
) (peer.ID, error) {
	if paddr == nil {
		return "", errors.New("target address can't be empty")
	}
	resp, err := c.c.SetReplicator(ctx, &pb.SetReplicatorRequest{
		Collections: collections,
		Addr:        paddr.Bytes(),
	})
	if err != nil {
		return "", errors.Wrap("could not add replicator", err)
	}
	return peer.IDFromBytes(resp.PeerID)
}

// DeleteReplicator sends a request to add a target replicator to the DB peer.
func (c *Client) DeleteReplicator(
	ctx context.Context,
	pid peer.ID,
	collections ...string,
) error {
	_, err := c.c.DeleteReplicator(ctx, &pb.DeleteReplicatorRequest{
		PeerID: []byte(pid),
	})
	return err
}

// GetAllReplicators sends a request to add a target replicator to the DB peer.
func (c *Client) GetAllReplicators(
	ctx context.Context,
) ([]client.Replicator, error) {
	resp, err := c.c.GetAllReplicators(ctx, &pb.GetAllReplicatorRequest{})
	if err != nil {
		return nil, errors.Wrap("could not get replicators", err)
	}
	reps := []client.Replicator{}
	for _, rep := range resp.Replicators {
		addr, err := ma.NewMultiaddrBytes(rep.Info.Addrs)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		pid, err := peer.IDFromBytes(rep.Info.Id)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		reps = append(reps, client.Replicator{
			Info: peer.AddrInfo{
				ID:    pid,
				Addrs: []ma.Multiaddr{addr},
			},
			Schemas: rep.Schemas,
		})
	}
	return reps, nil
}

// AddP2PCollections sends a request to add P2P collecctions to the stored list.
func (c *Client) AddP2PCollections(
	ctx context.Context,
	collections ...string,
) error {
	resp, err := c.c.AddP2PCollections(ctx, &pb.AddP2PCollectionsRequest{
		Collections: collections,
	})
	if err != nil {
		return errors.Wrap("could not add P2P collection topics", err)
	}
	if resp.Err != "" {
		return errors.New(fmt.Sprintf("could not add P2P collection topics: %s", resp))
	}
	return nil
}

// RemoveP2PCollections sends a request to remove P2P collecctions from the stored list.
func (c *Client) RemoveP2PCollections(
	ctx context.Context,
	collections ...string,
) error {
	resp, err := c.c.RemoveP2PCollections(ctx, &pb.RemoveP2PCollectionsRequest{
		Collections: collections,
	})
	if err != nil {
		return errors.Wrap("could not remove P2P collection topics", err)
	}
	if resp.Err != "" {
		return errors.New(fmt.Sprintf("could not remove P2P collection topics: %s", resp))
	}
	return nil
}

// RemoveP2PCollections sends a request to get all P2P collecctions from the stored list.
func (c *Client) GetAllP2PCollections(
	ctx context.Context,
) ([]client.P2PCollection, error) {
	resp, err := c.c.GetAllP2PCollections(ctx, &pb.GetAllP2PCollectionsRequest{})
	if err != nil {
		return nil, errors.Wrap("could not get all P2P collection topics", err)
	}
	var collections []client.P2PCollection
	for _, col := range resp.Collections {
		collections = append(collections, client.P2PCollection{
			ID:   col.Id,
			Name: col.Name,
		})
	}
	return collections, nil
}
