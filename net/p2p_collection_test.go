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

func TestAddP2PCollection_WithInvalidCollection_ShouldError(t *testing.T) {
	ctx := context.Background()
	db, peer := newTestPeer(ctx, t)
	defer db.Close()
	err := peer.AddP2PCollections(ctx, []string{"invalidCollection"})
	require.ErrorIs(t, err, client.ErrCollectionNotFound)
}

func TestAddP2PCollection_WithValidCollection_ShouldSucceed(t *testing.T) {
	ctx := context.Background()
	db, peer := newTestPeer(ctx, t)
	defer db.Close()
	cols, err := db.AddSchema(ctx, `type User { name: String }`)
	require.NoError(t, err)
	schema, err := db.GetSchemaByVersionID(ctx, cols[0].VersionID)
	require.NoError(t, err)
	err = peer.AddP2PCollections(ctx, []string{schema.Root})
	require.NoError(t, err)
}

func TestAddP2PCollection_WithMultipleValidCollections_ShouldSucceed(t *testing.T) {
	ctx := context.Background()
	db, peer := newTestPeer(ctx, t)
	defer db.Close()
	cols1, err := db.AddSchema(ctx, `type User { name: String }`)
	require.NoError(t, err)
	schema1, err := db.GetSchemaByVersionID(ctx, cols1[0].VersionID)
	require.NoError(t, err)
	cols2, err := db.AddSchema(ctx, `type Books { name: String }`)
	require.NoError(t, err)
	schema2, err := db.GetSchemaByVersionID(ctx, cols2[0].VersionID)
	require.NoError(t, err)
	err = peer.AddP2PCollections(ctx, []string{schema1.Root, schema2.Root})
	require.NoError(t, err)
}

func TestRemoveP2PCollection_WithInvalidCollection_ShouldError(t *testing.T) {
	ctx := context.Background()
	db, peer := newTestPeer(ctx, t)
	defer db.Close()
	err := peer.RemoveP2PCollections(ctx, []string{"invalidCollection"})
	require.ErrorIs(t, err, client.ErrCollectionNotFound)
}

func TestRemoveP2PCollection_WithValidCollection_ShouldSucceed(t *testing.T) {
	ctx := context.Background()
	db, peer := newTestPeer(ctx, t)
	defer db.Close()
	cols, err := db.AddSchema(ctx, `type User { name: String }`)
	require.NoError(t, err)
	schema, err := db.GetSchemaByVersionID(ctx, cols[0].VersionID)
	require.NoError(t, err)
	err = peer.AddP2PCollections(ctx, []string{schema.Root})
	require.NoError(t, err)
	err = peer.RemoveP2PCollections(ctx, []string{schema.Root})
	require.NoError(t, err)
}

func TestGetAllP2PCollections_WithMultipleValidCollections_ShouldSucceed(t *testing.T) {
	ctx := context.Background()
	db, peer := newTestPeer(ctx, t)
	defer db.Close()
	cols1, err := db.AddSchema(ctx, `type User { name: String }`)
	require.NoError(t, err)
	schema1, err := db.GetSchemaByVersionID(ctx, cols1[0].VersionID)
	require.NoError(t, err)
	cols2, err := db.AddSchema(ctx, `type Books { name: String }`)
	require.NoError(t, err)
	schema2, err := db.GetSchemaByVersionID(ctx, cols2[0].VersionID)
	require.NoError(t, err)
	err = peer.AddP2PCollections(ctx, []string{schema1.Root, schema2.Root})
	require.NoError(t, err)
	cols, err := peer.GetAllP2PCollections(ctx)
	require.NoError(t, err)
	require.Equal(t, []string{schema2.Root, schema1.Root}, cols)
}
