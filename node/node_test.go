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
	pubsub "github.com/libp2p/go-libp2p-pubsub"
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
	if err != nil {
		t.Error(err)
	}
	database, err = db.NewDB(ctx, rootstore, options...)
	if err != nil {
		t.Error(err)
	}
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
	if err != nil {
		t.Error(err)
	}
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
	if err != nil {
		t.Error(err)
	}
	var ps *pubsub.PubSub
	assert.Equal(t, ps, n.pubsub)
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
	if err != nil {
		t.Error(err)
	}

	// overly simple check of validity of pubsub, avoiding the process of creating a PubSub
	var ps *pubsub.PubSub
	assert.NotEqual(t, ps, n.pubsub)
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
	if err != nil {
		t.Error(err)
	}
	err = n.Close()
	if err != nil {
		t.Error(err)
	}
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

func TestInvalidListenTCPAddrStrings(t *testing.T) {
	opt := ListenTCPAddrStrings("/ip4/碎片整理")
	options, err := mergeOptions(opt)
	assert.Error(t, err)
	assert.Equal(t, Options{}, options)
}
