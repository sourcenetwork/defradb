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
	"github.com/stretchr/testify/require"
	grpcpeer "google.golang.org/grpc/peer"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/keys"
)

func TestNewServerSimple(t *testing.T) {
	ctx := context.Background()
	db, p := newTestPeer(ctx, t)
	defer db.Close()
	defer p.Close()
	_, err := newServer(p)
	require.NoError(t, err)
}

func getHead(ctx context.Context, db client.DB, docID client.DocID) (cid.Cid, error) {
	prefix := keys.DataStoreKeyFromDocID(docID).ToHeadStoreKey().WithFieldID(core.COMPOSITE_NAMESPACE).Bytes()

	entries, err := datastore.FetchKeysForPrefix(ctx, prefix, db.Headstore())
	if err != nil {
		return cid.Undef, err
	}

	if len(entries) > 0 {
		hsKey, err := keys.NewHeadstoreDocKey(string(entries[0]))
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

	_, err = p.server.pushLogHandler(ctx, &pushLogRequest{
		DocID:        doc.ID().String(),
		CID:          headCID.Bytes(),
		CollectionID: col.SchemaRoot(),
		Creator:      p.PeerID().String(),
		Block:        b,
	})
	require.NoError(t, err)
}
