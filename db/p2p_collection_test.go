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

	"github.com/sourcenetwork/defradb/client"
)

func newTestCollection(ctx context.Context, db client.DB, name string) (client.Collection, error) {
	desc := client.CollectionDescription{
		Name: name,
		Schema: client.SchemaDescription{
			Fields: []client.FieldDescription{
				{
					Name: "_key",
					Kind: client.FieldKind_DocKey,
				},
				{
					Name: "Name",
					Kind: client.FieldKind_STRING,
					Typ:  client.LWW_REGISTER,
				},
			},
		},
	}

	return db.CreateCollection(ctx, desc)
}

func TestAddP2PCollection(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	require.NoError(t, err)
	defer db.Close(ctx)

	col, err := newTestCollection(ctx, db, "test")
	require.NoError(t, err)

	err = db.AddP2PCollection(ctx, col.SchemaID())
	require.NoError(t, err)
}

func TestGetAllP2PCollection(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	require.NoError(t, err)
	defer db.Close(ctx)

	col1, err := newTestCollection(ctx, db, "test1")
	require.NoError(t, err)
	err = db.AddP2PCollection(ctx, col1.SchemaID())
	require.NoError(t, err)

	col2, err := newTestCollection(ctx, db, "test2")
	require.NoError(t, err)
	err = db.AddP2PCollection(ctx, col2.SchemaID())
	require.NoError(t, err)

	col3, err := newTestCollection(ctx, db, "test3")
	require.NoError(t, err)
	err = db.AddP2PCollection(ctx, col3.SchemaID())
	require.NoError(t, err)

	collections, err := db.GetAllP2PCollections(ctx)
	require.NoError(t, err)
	require.ElementsMatch(t, collections, []string{col1.SchemaID(), col2.SchemaID(), col3.SchemaID()})
}

func TestRemoveP2PCollection(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	require.NoError(t, err)
	defer db.Close(ctx)

	col1, err := newTestCollection(ctx, db, "test1")
	require.NoError(t, err)
	err = db.AddP2PCollection(ctx, col1.SchemaID())
	require.NoError(t, err)

	col2, err := newTestCollection(ctx, db, "test2")
	require.NoError(t, err)
	err = db.AddP2PCollection(ctx, col2.SchemaID())
	require.NoError(t, err)

	col3, err := newTestCollection(ctx, db, "test3")
	require.NoError(t, err)
	err = db.AddP2PCollection(ctx, col3.SchemaID())
	require.NoError(t, err)

	err = db.RemoveP2PCollection(ctx, col2.SchemaID())
	require.NoError(t, err)

	collections, err := db.GetAllP2PCollections(ctx)
	require.NoError(t, err)
	require.ElementsMatch(t, collections, []string{col1.SchemaID(), col3.SchemaID()})
}
