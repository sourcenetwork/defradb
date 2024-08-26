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

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore/query"
	"github.com/stretchr/testify/require"
	grpcpeer "google.golang.org/grpc/peer"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/core"
	net_pb "github.com/sourcenetwork/defradb/net/pb"
)

func TestNewServerSimple(t *testing.T) {
	ctx := context.Background()
	db, p := newTestPeer(ctx, t)
	defer db.Close()
	defer p.Close()
	_, err := newServer(p)
	require.NoError(t, err)
}

func TestGetDocGraph(t *testing.T) {
	ctx := context.Background()
	db, p := newTestPeer(ctx, t)
	defer db.Close()
	defer p.Close()
	r, err := p.server.GetDocGraph(ctx, &net_pb.GetDocGraphRequest{})
	require.Nil(t, r)
	require.Nil(t, err)
}

func TestPushDocGraph(t *testing.T) {
	ctx := context.Background()
	db, p := newTestPeer(ctx, t)
	defer db.Close()
	defer p.Close()
	r, err := p.server.PushDocGraph(ctx, &net_pb.PushDocGraphRequest{})
	require.Nil(t, r)
	require.Nil(t, err)
}

func TestGetLog(t *testing.T) {
	ctx := context.Background()
	db, p := newTestPeer(ctx, t)
	defer db.Close()
	defer p.Close()
	r, err := p.server.GetLog(ctx, &net_pb.GetLogRequest{})
	require.Nil(t, r)
	require.Nil(t, err)
}

func TestGetHeadLog(t *testing.T) {
	ctx := context.Background()
	db, p := newTestPeer(ctx, t)
	defer db.Close()
	defer p.Close()
	r, err := p.server.GetHeadLog(ctx, &net_pb.GetHeadLogRequest{})
	require.Nil(t, r)
	require.Nil(t, err)
}

func getHead(ctx context.Context, db client.DB, docID client.DocID) (cid.Cid, error) {
	prefix := core.DataStoreKeyFromDocID(docID).ToHeadStoreKey().WithFieldId(core.COMPOSITE_NAMESPACE).ToString()
	results, err := db.Headstore().Query(ctx, query.Query{Prefix: prefix})
	if err != nil {
		return cid.Undef, err
	}
	entries, err := results.Rest()
	if err != nil {
		return cid.Undef, err
	}

	if len(entries) > 0 {
		hsKey, err := core.NewHeadStoreKey(entries[0].Key)
		if err != nil {
			return cid.Undef, err
		}
		return hsKey.Cid, nil
	}
	return cid.Undef, errors.New("no head found")
}

func TestPushLog(t *testing.T) {
	ctx := context.Background()
	db, p := newTestPeer(ctx, t)
	defer db.Close()
	defer p.Close()

	_, err := db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}`)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	doc, err := client.NewDocFromJSON([]byte(`{"name": "John", "age": 30}`), col.Definition())
	require.NoError(t, err)

	ctx = grpcpeer.NewContext(ctx, &grpcpeer.Peer{
		Addr: addr{p.PeerID()},
	})

	err = col.Create(ctx, doc)
	require.NoError(t, err)

	headCID, err := getHead(ctx, db, doc.ID())
	require.NoError(t, err)

	b, err := db.Blockstore().AsIPLDStorage().Get(ctx, headCID.KeyString())
	require.NoError(t, err)

	_, err = p.server.PushLog(ctx, &net_pb.PushLogRequest{
		Body: &net_pb.PushLogRequest_Body{
			DocID:      []byte(doc.ID().String()),
			Cid:        headCID.Bytes(),
			SchemaRoot: []byte(col.SchemaRoot()),
			Creator:    p.PeerID().String(),
			Log: &net_pb.Document_Log{
				Block: b,
			},
		},
	})
	require.NoError(t, err)
}
