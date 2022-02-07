package client

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"google.golang.org/grpc"

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
func (c *Client) AddReplicator(ctx context.Context, collection string, paddr ma.Multiaddr) (peer.ID, error) {
	if len(collection) == 0 {
		return "", fmt.Errorf("Collection can't be empty")
	}
	if paddr == nil {
		return "", fmt.Errorf("target address can't be empty")
	}
	resp, err := c.c.AddReplicator(ctx, &pb.AddReplicatorRequest{
		Collection: []byte(collection),
		Addr:       paddr.Bytes(),
	})
	if err != nil {
		return "", fmt.Errorf("AddReplicator request failed: %w", err)
	}
	return peer.IDFromBytes(resp.PeerID)
}
