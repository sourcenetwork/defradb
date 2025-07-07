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

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
)

const (
	validDocID  = "bae-36fb9a9a-7af6-50cc-bc43-82e4c742a53e"
	validDocID2 = "bae-601875f7-7556-5a87-947e-221a8fd19b38"
)

func TestAddP2PDocument_WithInvalidDocument_ShouldError(t *testing.T) {
	ctx := context.Background()
	db, peer := newTestPeer(ctx, t)
	defer db.Close()
	err := peer.AddP2PDocuments(ctx, "invalidDocument")
	require.ErrorIs(t, err, client.ErrMalformedDocID)
}

func TestAddP2PDocument_WithValidDocument_ShouldSucceed(t *testing.T) {
	ctx := context.Background()
	db, peer := newTestPeer(ctx, t)
	defer db.Close()
	err := peer.AddP2PDocuments(ctx, validDocID)
	require.NoError(t, err)
}

func TestAddP2PDocument_WithMultipleValidDocuments_ShouldSucceed(t *testing.T) {
	ctx := context.Background()
	db, peer := newTestPeer(ctx, t)
	defer db.Close()
	err := peer.AddP2PDocuments(ctx, validDocID, validDocID2)
	require.NoError(t, err)
}

func TestRemoveP2PDocument_WithInvalidDocument_ShouldError(t *testing.T) {
	ctx := context.Background()
	db, peer := newTestPeer(ctx, t)
	defer db.Close()
	err := peer.RemoveP2PDocuments(ctx, "invalidDocument")
	require.ErrorIs(t, err, client.ErrMalformedDocID)
}

func TestRemoveP2PDocument_WithValidDocument_ShouldSucceed(t *testing.T) {
	ctx := context.Background()
	db, peer := newTestPeer(ctx, t)
	defer db.Close()
	err := peer.AddP2PDocuments(ctx, validDocID)
	require.NoError(t, err)
	err = peer.RemoveP2PDocuments(ctx, validDocID)
	require.NoError(t, err)
}

func TestGetAllP2PDocuments_WithMultipleValidDocuments_ShouldSucceed(t *testing.T) {
	ctx := context.Background()
	db, peer := newTestPeer(ctx, t)
	defer db.Close()
	err := peer.AddP2PDocuments(ctx, validDocID, validDocID2)
	require.NoError(t, err)
	cols, err := peer.GetAllP2PDocuments(ctx)
	require.NoError(t, err)
	require.Equal(t, []string{validDocID, validDocID2}, cols)
}
