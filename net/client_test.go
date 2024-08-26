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
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/event"
)

var def = client.CollectionDefinition{
	Description: client.CollectionDescription{
		Fields: []client.CollectionFieldDescription{
			{
				ID:   1,
				Name: "test",
			},
		},
	},
	Schema: client.SchemaDescription{
		Name: "test",
		Fields: []client.SchemaFieldDescription{
			{
				Name: "test",
				Kind: client.FieldKind_NILLABLE_STRING,
				Typ:  client.LWW_REGISTER,
			},
		},
	},
}

func TestPushlogWithDialFailure(t *testing.T) {
	ctx := context.Background()
	db, p := newTestPeer(ctx, t)
	defer db.Close()
	defer p.Close()

	doc, err := client.NewDocFromJSON([]byte(`{"test": "test"}`), def)
	require.NoError(t, err)
	id, err := doc.GenerateDocID()
	require.NoError(t, err)

	cid, err := createCID(doc)
	require.NoError(t, err)

	p.server.opts = append(
		p.server.opts,
		grpc.WithTransportCredentials(nil),
		grpc.WithCredentialsBundle(nil),
	)

	err = p.server.pushLog(event.Update{
		DocID:      id.String(),
		Cid:        cid,
		SchemaRoot: "test",
		Block:      emptyBlock(),
	}, peer.ID("some-peer-id"))
	require.Contains(t, err.Error(), "no transport security set")
}

func TestPushlogWithInvalidPeerID(t *testing.T) {
	ctx := context.Background()
	db, p := newTestPeer(ctx, t)
	defer db.Close()
	defer p.Close()

	doc, err := client.NewDocFromJSON([]byte(`{"test": "test"}`), def)
	require.NoError(t, err)
	id, err := doc.GenerateDocID()
	require.NoError(t, err)

	cid, err := createCID(doc)
	require.NoError(t, err)

	err = p.server.pushLog(event.Update{
		DocID:      id.String(),
		Cid:        cid,
		SchemaRoot: "test",
		Block:      emptyBlock(),
	}, peer.ID("some-peer-id"))
	require.Contains(t, err.Error(), "failed to parse peer ID")
}

func TestPushlogW_WithValidPeerID_NoError(t *testing.T) {
	ctx := context.Background()
	db1, p1 := newTestPeer(ctx, t)
	defer db1.Close()
	defer p1.Close()
	db2, p2 := newTestPeer(ctx, t)
	defer p2.Close()
	defer db2.Close()

	err := p1.host.Connect(ctx, p2.PeerInfo())
	require.NoError(t, err)

	_, err = db1.AddSchema(ctx, `type User {
		name: String
	}`)
	require.NoError(t, err)

	_, err = db2.AddSchema(ctx, `type User {
		name: String
	}`)
	require.NoError(t, err)

	col, err := db1.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	doc, err := client.NewDocFromJSON([]byte(`{"name": "test"}`), col.Definition())
	require.NoError(t, err)

	err = col.Save(ctx, doc)
	require.NoError(t, err)

	col, err = db2.GetCollectionByName(ctx, "User")
	require.NoError(t, err)
	err = col.Save(ctx, doc)
	require.NoError(t, err)

	headCID, err := getHead(ctx, db1, doc.ID())
	require.NoError(t, err)

	b, err := db1.Blockstore().AsIPLDStorage().Get(ctx, headCID.KeyString())
	require.NoError(t, err)

	err = p1.server.pushLog(event.Update{
		DocID:      doc.ID().String(),
		Cid:        headCID,
		SchemaRoot: col.SchemaRoot(),
		Block:      b,
	}, p2.PeerInfo().ID)
	require.NoError(t, err)
}
