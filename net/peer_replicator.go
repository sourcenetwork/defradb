package net

import (
	"context"

	"github.com/sourcenetwork/defradb/client"
)

func (p *Peer) SetReplicator(ctx context.Context, rep client.Replicator) error {
	panic("not implemented")
}

func (p *Peer) DeleteReplicator(ctx context.Context, rep client.Replicator) error {
	panic("not implemented")
}

func (p *Peer) GetAllReplicators(ctx context.Context) ([]client.Replicator, error) {
	panic("not implemented")
}
