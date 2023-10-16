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
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
)

func newTestCollection(
	t *testing.T,
	ctx context.Context,
	db *implicitTxnDB,
	name string,
) client.Collection {
	_, err := db.AddSchema(
		ctx,
		fmt.Sprintf(
			`type %s {
				Name: String
			}`,
			name,
		),
	)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, name)
	require.NoError(t, err)

	return col
}

func TestAddP2PCollection(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	require.NoError(t, err)
	defer db.Close(ctx)

	col := newTestCollection(t, ctx, db, "test")

	err = db.AddP2PCollections(ctx, []string{col.SchemaID()})
	require.NoError(t, err)
}

func TestGetAllP2PCollection(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	require.NoError(t, err)
	defer db.Close(ctx)

	col1 := newTestCollection(t, ctx, db, "test1")
	col2 := newTestCollection(t, ctx, db, "test2")
	col3 := newTestCollection(t, ctx, db, "test3")

	collectionIDs := []string{col1.SchemaID(), col2.SchemaID(), col3.SchemaID()}
	err = db.AddP2PCollections(ctx, collectionIDs)
	require.NoError(t, err)

	collections, err := db.GetAllP2PCollections(ctx)
	require.NoError(t, err)
	require.ElementsMatch(t, collections, collectionIDs)
}

func TestRemoveP2PCollection(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	require.NoError(t, err)
	defer db.Close(ctx)

	col1 := newTestCollection(t, ctx, db, "test1")
	col2 := newTestCollection(t, ctx, db, "test2")
	col3 := newTestCollection(t, ctx, db, "test3")

	collectionIDs := []string{col1.SchemaID(), col2.SchemaID(), col3.SchemaID()}

	err = db.AddP2PCollections(ctx, collectionIDs)
	require.NoError(t, err)

	err = db.RemoveP2PCollections(ctx, []string{col2.SchemaID()})
	require.NoError(t, err)

	collections, err := db.GetAllP2PCollections(ctx)
	require.NoError(t, err)
	require.ElementsMatch(t, collections, []string{col1.SchemaID(), col3.SchemaID()})
}
