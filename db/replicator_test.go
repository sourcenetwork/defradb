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

	ds "github.com/ipfs/go-datastore"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
)

func TestSetReplicator(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	require.NoError(t, err)
	defer db.Close()
	a, err := ma.NewMultiaddr("/ip4/192.168.1.12/tcp/9000/p2p/12D3KooWNXm3dmrwCYSxGoRUyZstaKYiHPdt8uZH5vgVaEJyzU8B")
	require.NoError(t, err)
	// Extract the peer ID from the multiaddr.
	info, err := peer.AddrInfoFromP2pAddr(a)
	require.NoError(t, err)
	err = db.SetReplicator(ctx, client.Replicator{
		Info:    *info,
		Schemas: []string{"test"},
	})
	assert.NoError(t, err)
}

func TestGetAllReplicatorsWith2Addition(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	require.NoError(t, err)
	defer db.Close()
	a, err := ma.NewMultiaddr("/ip4/192.168.1.12/tcp/9000/p2p/12D3KooWNXm3dmrwCYSxGoRUyZstaKYiHPdt8uZH5vgVaEJyzU8B")
	require.NoError(t, err)

	// Extract the peer ID from the multiaddr.
	info, err := peer.AddrInfoFromP2pAddr(a)
	require.NoError(t, err)

	err = db.SetReplicator(ctx, client.Replicator{
		Info:    *info,
		Schemas: []string{"test"},
	})
	require.NoError(t, err)

	a2, err := ma.NewMultiaddr("/ip4/192.168.1.12/tcp/9000/p2p/12D3KooWNXm3dmrwCYSxGoRUyZstaKYiHPdt8uZH5vgVaEJyzU8C")
	require.NoError(t, err)

	// Extract the peer ID from the multiaddr.
	info2, err := peer.AddrInfoFromP2pAddr(a2)
	require.NoError(t, err)

	err = db.SetReplicator(ctx, client.Replicator{
		Info:    *info2,
		Schemas: []string{"test", "test2", "test3"},
	})
	require.NoError(t, err)

	reps, err := db.GetAllReplicators(ctx)
	require.NoError(t, err)

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

func TestGetAllReplicatorsWith2AdditionsOnSamePeer(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	require.NoError(t, err)
	defer db.Close()
	a, err := ma.NewMultiaddr("/ip4/192.168.1.12/tcp/9000/p2p/12D3KooWNXm3dmrwCYSxGoRUyZstaKYiHPdt8uZH5vgVaEJyzU8B")
	require.NoError(t, err)

	// Extract the peer ID from the multiaddr.
	info, err := peer.AddrInfoFromP2pAddr(a)
	require.NoError(t, err)

	err = db.SetReplicator(ctx, client.Replicator{
		Info:    *info,
		Schemas: []string{"test"},
	})
	require.NoError(t, err)

	err = db.SetReplicator(ctx, client.Replicator{
		Info:    *info,
		Schemas: []string{"test", "test2", "test3"},
	})
	require.NoError(t, err)

	reps, err := db.GetAllReplicators(ctx)
	require.NoError(t, err)

	assert.Equal(t, []client.Replicator{
		{
			Info:    *info,
			Schemas: []string{"test", "test2", "test3"},
		},
	}, reps)
}

func TestDeleteSchemaForReplicator(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	require.NoError(t, err)
	defer db.Close()
	a, err := ma.NewMultiaddr("/ip4/192.168.1.12/tcp/9000/p2p/12D3KooWNXm3dmrwCYSxGoRUyZstaKYiHPdt8uZH5vgVaEJyzU8B")
	require.NoError(t, err)

	// Extract the peer ID from the multiaddr.
	info, err := peer.AddrInfoFromP2pAddr(a)
	require.NoError(t, err)

	err = db.SetReplicator(ctx, client.Replicator{
		Info:    *info,
		Schemas: []string{"test", "test2", "test3"},
	})
	require.NoError(t, err)

	err = db.DeleteReplicator(ctx, client.Replicator{
		Info:    *info,
		Schemas: []string{"test2"},
	})
	require.NoError(t, err)

	rep, err := db.getReplicator(ctx, *info)
	require.NoError(t, err)

	assert.Equal(t, client.Replicator{
		Info:    *info,
		Schemas: []string{"test", "test3"},
	}, rep)
}

func TestDeleteAllSchemasForReplicator(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	require.NoError(t, err)
	defer db.Close()
	a, err := ma.NewMultiaddr("/ip4/192.168.1.12/tcp/9000/p2p/12D3KooWNXm3dmrwCYSxGoRUyZstaKYiHPdt8uZH5vgVaEJyzU8B")
	require.NoError(t, err)

	// Extract the peer ID from the multiaddr.
	info, err := peer.AddrInfoFromP2pAddr(a)
	require.NoError(t, err)

	err = db.SetReplicator(ctx, client.Replicator{
		Info:    *info,
		Schemas: []string{"test", "test2", "test3"},
	})
	require.NoError(t, err)

	err = db.DeleteReplicator(ctx, client.Replicator{
		Info:    *info,
		Schemas: []string{"test", "test2", "test3"},
	})
	require.NoError(t, err)

	_, err = db.getReplicator(ctx, *info)
	require.ErrorIs(t, err, ds.ErrNotFound)
}

func TestDeleteReplicatorWith2Addition(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	require.NoError(t, err)
	defer db.Close()
	a, err := ma.NewMultiaddr("/ip4/192.168.1.12/tcp/9000/p2p/12D3KooWNXm3dmrwCYSxGoRUyZstaKYiHPdt8uZH5vgVaEJyzU8B")
	require.NoError(t, err)

	// Extract the peer ID from the multiaddr.
	info, err := peer.AddrInfoFromP2pAddr(a)
	require.NoError(t, err)

	err = db.SetReplicator(ctx, client.Replicator{
		Info:    *info,
		Schemas: []string{"test"},
	})
	require.NoError(t, err)

	a2, err := ma.NewMultiaddr("/ip4/192.168.1.12/tcp/9000/p2p/12D3KooWNXm3dmrwCYSxGoRUyZstaKYiHPdt8uZH5vgVaEJyzU8C")
	require.NoError(t, err)

	// Extract the peer ID from the multiaddr.
	info2, err := peer.AddrInfoFromP2pAddr(a2)
	require.NoError(t, err)

	err = db.SetReplicator(ctx, client.Replicator{
		Info:    *info2,
		Schemas: []string{"test", "test2", "test3"},
	})
	require.NoError(t, err)

	reps, err := db.GetAllReplicators(ctx)
	require.NoError(t, err)

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

	err = db.DeleteReplicator(ctx, client.Replicator{Info: *info})
	require.NoError(t, err)

	reps, err = db.GetAllReplicators(ctx)
	require.NoError(t, err)

	assert.Equal(t, []client.Replicator{
		{
			Info:    *info2,
			Schemas: []string{"test", "test2", "test3"},
		},
	}, reps)
}
