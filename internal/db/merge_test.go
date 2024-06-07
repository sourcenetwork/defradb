// Copyright 2024 Democratized Data Foundation
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

	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/events"
	"github.com/sourcenetwork/defradb/internal/core"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/db/base"
	"github.com/sourcenetwork/defradb/internal/merkle/clock"
)

const userSchema = `
type User {
	name: String
	age: Int
	email: String
	points: Int
}
`

func TestMerge_NoError(t *testing.T) {
	// Test that a merge can be performed up to the provided CID.
	ctx := context.Background()

	// Setup the "local" database
	localDB, err := newDefraMemoryDB(ctx)
	require.NoError(t, err)
	_, err = localDB.AddSchema(ctx, userSchema)
	require.NoError(t, err)
	localCol, err := localDB.GetCollectionByName(ctx, "User")
	require.NoError(t, err)
	docMap := map[string]any{
		"name": "Alice",
		"age":  30,
	}
	doc, err := client.NewDocFromMap(docMap, localCol.Definition())
	require.NoError(t, err)

	err = localCol.Create(ctx, doc)
	require.NoError(t, err)

	// Setup the "remote" database
	remoteDB, err := newDefraMemoryDB(ctx)
	require.NoError(t, err)
	_, err = remoteDB.AddSchema(ctx, userSchema)
	require.NoError(t, err)
	remoteCol, err := remoteDB.GetCollectionByName(ctx, "User")
	require.NoError(t, err)
	doc, err = client.NewDocFromMap(docMap, localCol.Definition())
	require.NoError(t, err)
	err = remoteCol.Create(ctx, doc)
	require.NoError(t, err)

	// Add a few changes to the remote node
	err = doc.Set("points", 100)
	require.NoError(t, err)
	err = remoteCol.Update(ctx, doc)
	require.NoError(t, err)

	// Sync the remote blocks to the local node
	err = syncAndMerge(ctx, remoteDB, localDB, remoteCol, localCol, doc.ID().String())
	require.NoError(t, err)

	// verify the local node has the same data as the remote node
	localDoc, err := localCol.Get(ctx, doc.ID(), false)
	require.NoError(t, err)
	localDocString, err := localDoc.String()
	require.NoError(t, err)
	remoteDoc, err := remoteCol.Get(ctx, doc.ID(), false)
	require.NoError(t, err)
	remoteDocString, err := remoteDoc.String()
	require.NoError(t, err)
	require.Equal(t, remoteDocString, localDocString)
}

func TestMerge_DelayedSync_NoError(t *testing.T) {
	// Test that a merge can be performed up to the provided CID.
	ctx := context.Background()

	// Setup the "local" database
	localDB, err := newDefraMemoryDB(ctx)
	require.NoError(t, err)
	_, err = localDB.AddSchema(ctx, userSchema)
	require.NoError(t, err)
	localCol, err := localDB.GetCollectionByName(ctx, "User")
	require.NoError(t, err)
	docMap := map[string]any{
		"name": "Alice",
		"age":  30,
	}
	doc, err := client.NewDocFromMap(docMap, localCol.Definition())
	require.NoError(t, err)

	err = localCol.Create(ctx, doc)
	require.NoError(t, err)

	// Setup the "remote" database
	remoteDB, err := newDefraMemoryDB(ctx)
	require.NoError(t, err)
	_, err = remoteDB.AddSchema(ctx, userSchema)
	require.NoError(t, err)
	remoteCol, err := remoteDB.GetCollectionByName(ctx, "User")
	require.NoError(t, err)
	doc, err = client.NewDocFromMap(docMap, localCol.Definition())
	require.NoError(t, err)
	err = remoteCol.Create(ctx, doc)
	require.NoError(t, err)

	// Add a few changes to the remote node
	err = doc.Set("points", 100)
	require.NoError(t, err)
	err = remoteCol.Update(ctx, doc)
	require.NoError(t, err)

	err = doc.Set("age", 31)
	require.NoError(t, err)
	err = remoteCol.Update(ctx, doc)
	require.NoError(t, err)

	err = doc.Set("email", "alice@yahoo.com")
	require.NoError(t, err)
	err = remoteCol.Update(ctx, doc)
	require.NoError(t, err)

	// Sync the remote blocks to the local node
	err = syncAndMerge(ctx, remoteDB, localDB, remoteCol, localCol, doc.ID().String())
	require.NoError(t, err)

	// verify the local node has the same data as the remote node
	localDoc, err := localCol.Get(ctx, doc.ID(), false)
	require.NoError(t, err)
	localDocString, err := localDoc.String()
	require.NoError(t, err)
	remoteDoc, err := remoteCol.Get(ctx, doc.ID(), false)
	require.NoError(t, err)
	remoteDocString, err := remoteDoc.String()
	require.NoError(t, err)
	require.Equal(t, remoteDocString, localDocString)
}

func TestMerge_DelayedSyncTwoBranches_NoError(t *testing.T) {
	// Test that a merge can be performed up to the provided CID.
	ctx := context.Background()

	// Setup the "local" database
	localDB, err := newDefraMemoryDB(ctx)
	require.NoError(t, err)
	_, err = localDB.AddSchema(ctx, userSchema)
	require.NoError(t, err)
	localCol, err := localDB.GetCollectionByName(ctx, "User")
	require.NoError(t, err)
	docMap := map[string]interface{}{
		"name": "Alice",
		"age":  30,
	}
	doc, err := client.NewDocFromMap(docMap, localCol.Definition())
	require.NoError(t, err)

	err = localCol.Create(ctx, doc)
	require.NoError(t, err)

	// Setup the "remote" database
	remoteDB1, err := newDefraMemoryDB(ctx)
	require.NoError(t, err)
	_, err = remoteDB1.AddSchema(ctx, userSchema)
	require.NoError(t, err)
	remoteCol1, err := remoteDB1.GetCollectionByName(ctx, "User")
	require.NoError(t, err)
	doc, err = client.NewDocFromMap(docMap, remoteCol1.Definition())
	require.NoError(t, err)
	err = remoteCol1.Create(ctx, doc)
	require.NoError(t, err)

	// Setup the second "remote" database
	remoteDB2, err := newDefraMemoryDB(ctx)
	require.NoError(t, err)
	_, err = remoteDB2.AddSchema(ctx, userSchema)
	require.NoError(t, err)
	remoteCol2, err := remoteDB2.GetCollectionByName(ctx, "User")
	require.NoError(t, err)
	doc2, err := client.NewDocFromMap(docMap, remoteCol2.Definition())
	require.NoError(t, err)
	err = remoteCol2.Create(ctx, doc2)
	require.NoError(t, err)

	// Add a few changes to the remote nodes creating two branches
	err = doc.Set("points", 200)
	require.NoError(t, err)
	err = remoteCol1.Update(ctx, doc)
	require.NoError(t, err)

	err = doc2.Set("points", 100)
	require.NoError(t, err)
	err = remoteCol2.Update(ctx, doc2)
	require.NoError(t, err)

	err = doc.Set("age", 31)
	require.NoError(t, err)
	err = remoteCol1.Update(ctx, doc)
	require.NoError(t, err)

	err = doc2.Set("age", 32)
	require.NoError(t, err)
	err = remoteCol2.Update(ctx, doc2)
	require.NoError(t, err)

	err = doc.Set("email", "alice@yahoo.com")
	require.NoError(t, err)
	err = remoteCol1.Update(ctx, doc)
	require.NoError(t, err)

	err = doc2.Set("email", "alice-in-wonderland@yahoo.com")
	require.NoError(t, err)
	err = remoteCol2.Update(ctx, doc2)
	require.NoError(t, err)

	// Sync the remote blocks to the local node
	err = syncAndMerge(ctx, remoteDB2, remoteDB1, remoteCol2, remoteCol1, doc.ID().String())
	require.NoError(t, err)
	err = syncAndMerge(ctx, remoteDB1, localDB, remoteCol1, localCol, doc.ID().String())
	require.NoError(t, err)

	// verify the local node has the same data as the remote node
	localDoc, err := localCol.Get(ctx, doc.ID(), false)
	require.NoError(t, err)
	localDocString, err := localDoc.String()
	require.NoError(t, err)
	remoteDoc1, err := remoteCol1.Get(ctx, doc.ID(), false)
	require.NoError(t, err)
	remoteDocString1, err := remoteDoc1.String()
	require.NoError(t, err)
	require.Equal(t, remoteDocString1, localDocString)
}

func syncAndMerge(ctx context.Context, from, to *db, fromCol, toCol client.Collection, docID string) error {
	dsKey := base.MakeDataStoreKeyWithCollectionAndDocID(fromCol.Description(), docID)
	headset := clock.NewHeadSet(
		from.multistore.Headstore(),
		dsKey.WithFieldId(core.COMPOSITE_NAMESPACE).ToHeadStoreKey(),
	)

	cids, _, err := headset.List(ctx)
	if err != nil {
		return err
	}

	for _, cid := range cids {
		blockBytes, err := from.multistore.DAGstore().AsIPLDStorage().Get(ctx, cid.KeyString())
		if err != nil {
			return err
		}
		block, err := coreblock.GetFromBytes(blockBytes)
		if err != nil {
			return err
		}
		err = syncDAG(ctx, from, to, block)
		if err != nil {
			return err
		}
		err = to.executeMerge(ctx, events.DAGMerge{
			Cid:        cid,
			SchemaRoot: toCol.SchemaRoot(),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func syncDAG(ctx context.Context, from, to *db, block *coreblock.Block) error {
	lsys := cidlink.DefaultLinkSystem()
	lsys.SetWriteStorage(to.multistore.DAGstore().AsIPLDStorage())
	_, err := lsys.Store(linking.LinkContext{Ctx: ctx}, coreblock.GetLinkPrototype(), block.GenerateNode())
	if err != nil {
		return err
	}

	for _, link := range block.Links {
		lsys := cidlink.DefaultLinkSystem()
		lsys.SetReadStorage(from.multistore.DAGstore().AsIPLDStorage())
		nd, err := lsys.Load(linking.LinkContext{Ctx: ctx}, link, coreblock.SchemaPrototype)
		if err != nil {
			return err
		}
		block, err := coreblock.GetFromNode(nd)
		if err != nil {
			return err
		}
		err = syncDAG(ctx, from, to, block)
		if err != nil {
			return err
		}
	}
	return nil
}
