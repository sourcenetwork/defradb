// Copyright 2024 Democratized Data Foundation
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
	b58 "github.com/mr-tron/base58/base58"
	"github.com/stretchr/testify/require"
)

func TestSetReplicator_WithEmptyPeerInfo_ShouldError(t *testing.T) {
	ctx := context.Background()
	db, p := newTestPeer(ctx, t)
	defer db.Close()
	err := p.SetReplicator(ctx, peer.AddrInfo{})
	require.ErrorContains(t, err, "empty peer ID")
}

func TestSetReplicator_WithSelfTarget_ShouldError(t *testing.T) {
	ctx := context.Background()
	db, p := newTestPeer(ctx, t)
	defer db.Close()
	err := p.SetReplicator(ctx, peer.AddrInfo{ID: p.PeerID()})
	require.ErrorIs(t, err, ErrSelfTargetForReplicator)
}

func TestSetReplicator_WithInvalidCollection_ShouldError(t *testing.T) {
	ctx := context.Background()
	db, p := newTestPeer(ctx, t)
	defer db.Close()
	err := p.SetReplicator(ctx, peer.AddrInfo{ID: "other"}, "invalidCollection")
	require.ErrorIs(t, err, ErrReplicatorCollections)
}

func TestSetReplicator_WithValidCollection_ShouldSucceed(t *testing.T) {
	ctx := context.Background()
	db, p := newTestPeer(ctx, t)
	defer db.Close()
	_, err := db.AddSchema(ctx, `type User { name: String }`)
	require.NoError(t, err)
	err = p.SetReplicator(ctx, peer.AddrInfo{ID: "other"}, "User")
	require.NoError(t, err)
}

func TestSetReplicator_WithValidCollectionsOnSeparateSet_ShouldSucceed(t *testing.T) {
	b, err := b58.Decode("12D3KooWB8Na2fKhdGtej5GjoVhmBBYFvqXiqFCSkR7fJFWHUbNr")
	require.NoError(t, err)
	peerID, err := peer.IDFromBytes(b)
	require.NoError(t, err)
	ctx := context.Background()
	db, p := newTestPeer(ctx, t)
	require.NoError(t, err)
	defer db.Close()
	_, err = db.AddSchema(ctx, `type User { name: String }`)
	require.NoError(t, err)
	err = p.SetReplicator(ctx, peer.AddrInfo{ID: peerID}, "User")
	require.NoError(t, err)

	_, err = db.AddSchema(ctx, `type Book { name: String }`)
	require.NoError(t, err)
	err = p.SetReplicator(ctx, peer.AddrInfo{ID: peerID}, "Book")
	require.NoError(t, err)
}

func TestDeleteReplicator_WithEmptyPeerInfo_ShouldError(t *testing.T) {
	ctx := context.Background()
	db, p := newTestPeer(ctx, t)
	defer db.Close()
	err := p.DeleteReplicator(ctx, peer.AddrInfo{})
	require.ErrorContains(t, err, "empty peer ID")
}

func TestDeleteReplicator_WithNonExistantReplicator_ShouldError(t *testing.T) {
	ctx := context.Background()
	db, p := newTestPeer(ctx, t)
	defer db.Close()
	err := p.DeleteReplicator(ctx, peer.AddrInfo{ID: "other"})
	require.ErrorIs(t, err, ErrReplicatorNotFound)
}

func TestDeleteReplicator_WithValidCollection_ShouldSucceed(t *testing.T) {
	b, err := b58.Decode("12D3KooWB8Na2fKhdGtej5GjoVhmBBYFvqXiqFCSkR7fJFWHUbNr")
	require.NoError(t, err)
	peerID, err := peer.IDFromBytes(b)
	require.NoError(t, err)
	ctx := context.Background()
	db, p := newTestPeer(ctx, t)
	defer db.Close()
	_, err = db.AddSchema(ctx, `type User { name: String }`)
	require.NoError(t, err)
	err = p.SetReplicator(ctx, peer.AddrInfo{ID: peerID}, "User")
	require.NoError(t, err)
	err = p.DeleteReplicator(ctx, peer.AddrInfo{ID: peerID})
	require.NoError(t, err)
}

func TestDeleteReplicator_PartialWithValidCollections_ShouldSucceed(t *testing.T) {
	b, err := b58.Decode("12D3KooWB8Na2fKhdGtej5GjoVhmBBYFvqXiqFCSkR7fJFWHUbNr")
	require.NoError(t, err)
	peerID, err := peer.IDFromBytes(b)
	require.NoError(t, err)
	ctx := context.Background()
	db, p := newTestPeer(ctx, t)
	defer db.Close()
	_, err = db.AddSchema(ctx, `type User { name: String }`)
	require.NoError(t, err)
	_, err = db.AddSchema(ctx, `type Book { name: String }`)
	require.NoError(t, err)
	err = p.SetReplicator(ctx, peer.AddrInfo{ID: peerID}, "User", "Book")
	require.NoError(t, err)
	err = p.DeleteReplicator(ctx, peer.AddrInfo{ID: peerID}, "User")
	require.NoError(t, err)
}

func TestGetAllReplicators_WithValidCollection_ShouldSucceed(t *testing.T) {
	b, err := b58.Decode("12D3KooWB8Na2fKhdGtej5GjoVhmBBYFvqXiqFCSkR7fJFWHUbNr")
	require.NoError(t, err)
	peerID, err := peer.IDFromBytes(b)
	require.NoError(t, err)
	ctx := context.Background()
	db, p := newTestPeer(ctx, t)
	defer db.Close()
	cols, err := db.AddSchema(ctx, `type User { name: String }`)
	require.NoError(t, err)
	err = p.SetReplicator(ctx, peer.AddrInfo{ID: peerID}, "User")
	require.NoError(t, err)
	reps, err := p.GetAllReplicators(ctx)
	require.NoError(t, err)
	require.Equal(t, peerID, reps[0].Info.ID)
	require.Equal(t, []string{cols[0].CollectionID}, reps[0].CollectionIDs)
}

func TestLoadReplicators_WithValidCollection_ShouldSucceed(t *testing.T) {
	b, err := b58.Decode("12D3KooWB8Na2fKhdGtej5GjoVhmBBYFvqXiqFCSkR7fJFWHUbNr")
	require.NoError(t, err)
	peerID, err := peer.IDFromBytes(b)
	require.NoError(t, err)
	ctx := context.Background()
	db, p := newTestPeer(ctx, t)
	defer db.Close()
	_, err = db.AddSchema(ctx, `type User { name: String }`)
	require.NoError(t, err)
	err = p.SetReplicator(ctx, peer.AddrInfo{ID: peerID}, "User")
	require.NoError(t, err)
	err = p.loadAndPublishReplicators(ctx)
	require.NoError(t, err)
}
