// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package node

import (
	"context"
	"testing"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/sourcenetwork/defradb/client"
	badgerds "github.com/sourcenetwork/defradb/datastore/badger/v3"
	"github.com/sourcenetwork/defradb/db"
	"github.com/stretchr/testify/assert"
	"github.com/textileio/go-threads/broadcast"
)

// Node.Boostrap is not tested because the underlying, *ipfslite.Peer.Bootstrap is a best-effort function.

func FixtureNewMemoryDBWithBroadcaster(t *testing.T) (client.DB, *broadcast.Broadcaster) {
	var database client.DB
	var options []db.Option
	var busBufferSize = 100
	ctx := context.Background()
	bs := broadcast.NewBroadcaster(busBufferSize)
	options = append(options, db.WithBroadcaster(bs))
	opts := badgerds.Options{Options: badger.DefaultOptions("").WithInMemory(true)}
	rootstore, err := badgerds.NewDatastore("", &opts)
	assert.NoError(t, err)
	database, err = db.NewDB(ctx, rootstore, options...)
	assert.NoError(t, err)
	return database, bs
}

func TestNewNode(t *testing.T) {
	db, bs := FixtureNewMemoryDBWithBroadcaster(t)
	_, err := NewNode(
		context.Background(),
		db,
		bs,
		// DataPath() is a required option with the current implementation of key management
		DataPath(t.TempDir()),
	)
	assert.NoError(t, err)
}

func TestNewNodeNoPubSub(t *testing.T) {
	db, bs := FixtureNewMemoryDBWithBroadcaster(t)
	n, err := NewNode(
		context.Background(),
		db,
		bs,
		WithPubSub(false),
		// DataPath() is a required option with the current implementation of key management
		DataPath(t.TempDir()),
	)
	assert.NoError(t, err)
	assert.Nil(t, n.pubsub)
}

func TestNewNodeWithPubSub(t *testing.T) {
	db, bs := FixtureNewMemoryDBWithBroadcaster(t)
	ctx := context.Background()
	n, err := NewNode(
		ctx,
		db,
		bs,
		WithPubSub(true),
		// DataPath() is a required option with the current implementation of key management
		DataPath(t.TempDir()),
	)
	assert.NoError(t, err)
	// overly simple check of validity of pubsub, avoiding the process of creating a PubSub
	assert.NotNil(t, n.pubsub)
}

func TestNewNodeWithPubSubFailsWithoutDataPath(t *testing.T) {
	db, bs := FixtureNewMemoryDBWithBroadcaster(t)
	ctx := context.Background()
	_, err := NewNode(
		ctx,
		db,
		bs,
		WithPubSub(true),
	)
	assert.EqualError(t, err, "1 error occurred:\n\t* mkdir : no such file or directory\n\n")
}

func TestNodeClose(t *testing.T) {
	db, bs := FixtureNewMemoryDBWithBroadcaster(t)
	n, err := NewNode(
		context.Background(),
		db,
		bs,
		// DataPath() is a required option with the current implementation of key management
		DataPath(t.TempDir()),
	)
	assert.NoError(t, err)
	err = n.Close()
	assert.NoError(t, err)
}

func mergeOptions(nodeOpts ...NodeOpt) (Options, error) {
	var options Options
	var nodeOpt NodeOpt
	for _, opt := range append(nodeOpts, nodeOpt) {
		if opt == nil {
			continue
		}
		if err := opt(&options); err != nil {
			return options, err
		}
	}
	return options, nil
}

func TestInvalidListenTCPAddrString(t *testing.T) {
	opt := ListenTCPAddrString("/ip4/碎片整理")
	options, err := mergeOptions(opt)
	assert.EqualError(t, err, "failed to parse multiaddr \"/ip4/碎片整理\": invalid value \"碎片整理\" for protocol ip4: failed to parse ip4 addr: 碎片整理")
	assert.Equal(t, Options{}, options)
}
