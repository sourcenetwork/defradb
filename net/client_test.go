// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package net

import (
	"context"
	"testing"

	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/events"
)

func TestPushlogWithDialFailure(t *testing.T) {
	ctx := context.Background()
	_, n := newTestNode(ctx, t)

	doc, err := client.NewDocFromJSON([]byte(`{"test": "test"}`))
	require.NoError(t, err)

	cid, err := createCID(doc)
	require.NoError(t, err)

	n.server.opts = append(
		n.server.opts,
		grpc.WithTransportCredentials(nil),
		grpc.WithCredentialsBundle(nil),
	)

	err = n.server.pushLog(ctx, events.Update{
		DocKey:   doc.Key().String(),
		Cid:      cid,
		SchemaID: "test",
		Block:    &EmptyNode{},
		Priority: 1,
	}, peer.ID("some-peer-id"))
	require.Contains(t, err.Error(), "no transport security set")
}

func TestPushlogWithInvalidPeerID(t *testing.T) {
	ctx := context.Background()
	_, n := newTestNode(ctx, t)

	doc, err := client.NewDocFromJSON([]byte(`{"test": "test"}`))
	require.NoError(t, err)

	cid, err := createCID(doc)
	require.NoError(t, err)

	err = n.server.pushLog(ctx, events.Update{
		DocKey:   doc.Key().String(),
		Cid:      cid,
		SchemaID: "test",
		Block:    &EmptyNode{},
		Priority: 1,
	}, peer.ID("some-peer-id"))
	require.Contains(t, err.Error(), "failed to parse peer ID")
}

func TestPushlogWithInvalidPeerID2(t *testing.T) {
	ctx := context.Background()
	_, n := newTestNode(ctx, t)
	n.Start()
	_, n2 := newTestNode(ctx, t)
	n2.Start()

	err := n.host.Connect(ctx, peer.AddrInfo{
		ID: n2.PeerID(),
		Addrs: []ma.Multiaddr{
			n2.host.Addrs()[0],
		},
	})
	require.NoError(t, err)

	_, err = n.db.AddSchema(ctx, `type User {
		name: String
	}`)
	require.NoError(t, err)

	_, err = n2.db.AddSchema(ctx, `type User {
		name: String
	}`)
	require.NoError(t, err)

	doc, err := client.NewDocFromJSON([]byte(`{"name": "test"}`))
	require.NoError(t, err)

	col, err := n.db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)
	err = col.Save(ctx, doc)
	require.NoError(t, err)

	col, err = n2.db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)
	err = col.Save(ctx, doc)
	require.NoError(t, err)

	cid, err := createCID(doc)
	require.NoError(t, err)

	err = n.server.pushLog(ctx, events.Update{
		DocKey:   doc.Key().String(),
		Cid:      cid,
		SchemaID: col.SchemaID(),
		Block:    &EmptyNode{},
		Priority: 1,
	}, n2.PeerID())
	require.NoError(t, err)
}
