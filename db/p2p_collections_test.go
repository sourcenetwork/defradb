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

	"github.com/stretchr/testify/require"
)

func TestAddP2PCollection(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	require.NoError(t, err)
	defer db.Close(ctx)

	err = db.AddP2PCollection(ctx, "abc123")
	require.NoError(t, err)
}

func TestGetAllP2PCollection(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	require.NoError(t, err)
	defer db.Close(ctx)

	err = db.AddP2PCollection(ctx, "abc123")
	require.NoError(t, err)
	err = db.AddP2PCollection(ctx, "abc789")
	require.NoError(t, err)
	err = db.AddP2PCollection(ctx, "qwe123")
	require.NoError(t, err)

	collections, err := db.GetAllP2PCollections(ctx)
	require.NoError(t, err)
	require.ElementsMatch(t, collections, []string{"abc123", "abc789", "qwe123"})
}

func TestRemoveP2PCollection(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	require.NoError(t, err)
	defer db.Close(ctx)

	err = db.AddP2PCollection(ctx, "abc123")
	require.NoError(t, err)
	err = db.AddP2PCollection(ctx, "abc789")
	require.NoError(t, err)
	err = db.AddP2PCollection(ctx, "qwe123")
	require.NoError(t, err)

	err = db.RemoveP2PCollection(ctx, "abc789")
	require.NoError(t, err)

	collections, err := db.GetAllP2PCollections(ctx)
	require.NoError(t, err)
	require.ElementsMatch(t, collections, []string{"abc123", "qwe123"})
}
