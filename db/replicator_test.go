// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package db

import (
	"context"
	"testing"

	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/sourcenetwork/defradb/client"
	"github.com/stretchr/testify/assert"
)

func TestAddReplicator(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	if err != nil {
		t.Error(err)
	}
	a, err := ma.NewMultiaddr("/ip4/192.168.1.12/tcp/9000/p2p/12D3KooWNXm3dmrwCYSxGoRUyZstaKYiHPdt8uZH5vgVaEJyzU8B")
	if err != nil {
		t.Error(err)
	}
	// Extract the peer ID from the multiaddr.
	info, err := peer.AddrInfoFromP2pAddr(a)
	if err != nil {
		t.Error(err)
	}
	err = db.AddReplicator(ctx, client.Replicator{
		Info:    *info,
		Schemas: []string{"test"},
	})
	assert.NoError(t, err)
}

func TestGetAllReplicatorsWith2Addition(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	if err != nil {
		t.Error(err)
	}
	a, err := ma.NewMultiaddr("/ip4/192.168.1.12/tcp/9000/p2p/12D3KooWNXm3dmrwCYSxGoRUyZstaKYiHPdt8uZH5vgVaEJyzU8B")
	if err != nil {
		t.Error(err)
	}

	// Extract the peer ID from the multiaddr.
	info, err := peer.AddrInfoFromP2pAddr(a)
	if err != nil {
		t.Error(err)
	}

	err = db.AddReplicator(ctx, client.Replicator{
		Info:    *info,
		Schemas: []string{"test"},
	})
	if err != nil {
		t.Error(err)
	}

	a2, err := ma.NewMultiaddr("/ip4/192.168.1.12/tcp/9000/p2p/12D3KooWNXm3dmrwCYSxGoRUyZstaKYiHPdt8uZH5vgVaEJyzU8C")
	if err != nil {
		t.Error(err)
	}

	// Extract the peer ID from the multiaddr.
	info2, err := peer.AddrInfoFromP2pAddr(a2)
	if err != nil {
		t.Error(err)
	}

	err = db.AddReplicator(ctx, client.Replicator{
		Info:    *info2,
		Schemas: []string{"test", "test2", "test3"},
	})
	if err != nil {
		t.Error(err)
	}

	reps, err := db.GetAllReplicators(ctx)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, []client.Replicator{
		{
			Info:    *info,
			Schemas: []string{"test"},
		},
		{
			Info:    *info2,
			Schemas: []string{"test", "test2", "test3"},
		},
	}, reps)
}

func TestGetAllReplicatorsWith2AddionnsOnSamePeer(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	if err != nil {
		t.Error(err)
	}
	a, err := ma.NewMultiaddr("/ip4/192.168.1.12/tcp/9000/p2p/12D3KooWNXm3dmrwCYSxGoRUyZstaKYiHPdt8uZH5vgVaEJyzU8B")
	if err != nil {
		t.Error(err)
	}

	// Extract the peer ID from the multiaddr.
	info, err := peer.AddrInfoFromP2pAddr(a)
	if err != nil {
		t.Error(err)
	}

	err = db.AddReplicator(ctx, client.Replicator{
		Info:    *info,
		Schemas: []string{"test"},
	})
	if err != nil {
		t.Error(err)
	}

	err = db.AddReplicator(ctx, client.Replicator{
		Info:    *info,
		Schemas: []string{"test", "test2", "test3"},
	})
	if err != nil {
		t.Error(err)
	}

	reps, err := db.GetAllReplicators(ctx)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, []client.Replicator{
		{
			Info:    *info,
			Schemas: []string{"test", "test2", "test3"},
		},
	}, reps)
}

func TestDeleteReplicatorWith2Addition(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	if err != nil {
		t.Error(err)
	}
	a, err := ma.NewMultiaddr("/ip4/192.168.1.12/tcp/9000/p2p/12D3KooWNXm3dmrwCYSxGoRUyZstaKYiHPdt8uZH5vgVaEJyzU8B")
	if err != nil {
		t.Error(err)
	}

	// Extract the peer ID from the multiaddr.
	info, err := peer.AddrInfoFromP2pAddr(a)
	if err != nil {
		t.Error(err)
	}

	err = db.AddReplicator(ctx, client.Replicator{
		Info:    *info,
		Schemas: []string{"test"},
	})
	if err != nil {
		t.Error(err)
	}

	a2, err := ma.NewMultiaddr("/ip4/192.168.1.12/tcp/9000/p2p/12D3KooWNXm3dmrwCYSxGoRUyZstaKYiHPdt8uZH5vgVaEJyzU8C")
	if err != nil {
		t.Error(err)
	}

	// Extract the peer ID from the multiaddr.
	info2, err := peer.AddrInfoFromP2pAddr(a2)
	if err != nil {
		t.Error(err)
	}

	err = db.AddReplicator(ctx, client.Replicator{
		Info:    *info2,
		Schemas: []string{"test", "test2", "test3"},
	})
	if err != nil {
		t.Error(err)
	}

	err = db.DeleteReplicator(ctx, info.ID)
	if err != nil {
		t.Error(err)
	}

	reps, err := db.GetAllReplicators(ctx)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, []client.Replicator{
		{
			Info:    *info2,
			Schemas: []string{"test", "test2", "test3"},
		},
	}, reps)
}
