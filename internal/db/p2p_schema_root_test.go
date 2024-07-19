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

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/event"
)

func TestAddP2PCollection_WithInvalidCollection_ShouldError(t *testing.T) {
	ctx := context.Background()
	db, err := newDefraMemoryDB(ctx)
	require.NoError(t, err)
	defer db.Close()
	err = db.AddP2PCollections(ctx, []string{"invalidCollection"})
	require.ErrorIs(t, err, client.ErrCollectionNotFound)
}

func TestAddP2PCollection_WithValidCollection_ShouldSucceed(t *testing.T) {
	ctx := context.Background()
	db, err := newDefraMemoryDB(ctx)
	require.NoError(t, err)
	defer db.Close()
	sub, err := db.events.Subscribe(event.P2PTopicName)
	require.NoError(t, err)
	cols, err := db.AddSchema(ctx, `type User { name: String }`)
	require.NoError(t, err)
	schema, err := db.GetSchemaByVersionID(ctx, cols[0].SchemaVersionID)
	require.NoError(t, err)
	err = db.AddP2PCollections(ctx, []string{schema.Root})
	require.NoError(t, err)
	// Check that the event was published
	for msg := range sub.Message() {
		p2pTopic := msg.Data.(event.P2PTopic)
		require.Equal(t, []string{schema.Root}, p2pTopic.ToAdd)
		break
	}
}

func TestAddP2PCollection_WithValidCollectionAndDoc_ShouldSucceed(t *testing.T) {
	ctx := context.Background()
	db, err := newDefraMemoryDB(ctx)
	require.NoError(t, err)
	defer db.Close()
	sub, err := db.events.Subscribe(event.P2PTopicName)
	require.NoError(t, err)
	cols, err := db.AddSchema(ctx, `type User { name: String }`)
	require.NoError(t, err)
	col, err := db.GetCollectionByName(ctx, cols[0].Name.Value())
	require.NoError(t, err)
	doc, err := client.NewDocFromMap(map[string]any{"name": "Alice"}, col.Definition())
	require.NoError(t, err)
	err = col.Create(ctx, doc)
	require.NoError(t, err)

	err = db.AddP2PCollections(ctx, []string{col.SchemaRoot()})
	require.NoError(t, err)
	// Check that the event was published
	for msg := range sub.Message() {
		p2pTopic := msg.Data.(event.P2PTopic)
		require.Equal(t, []string{col.SchemaRoot()}, p2pTopic.ToAdd)
		require.Equal(t, []string{doc.ID().String()}, p2pTopic.ToRemove)
		break
	}
}

func TestAddP2PCollection_WithMultipleValidCollections_ShouldSucceed(t *testing.T) {
	ctx := context.Background()
	db, err := newDefraMemoryDB(ctx)
	require.NoError(t, err)
	defer db.Close()
	sub, err := db.events.Subscribe(event.P2PTopicName)
	require.NoError(t, err)
	cols1, err := db.AddSchema(ctx, `type User { name: String }`)
	require.NoError(t, err)
	schema1, err := db.GetSchemaByVersionID(ctx, cols1[0].SchemaVersionID)
	require.NoError(t, err)
	cols2, err := db.AddSchema(ctx, `type Books { name: String }`)
	require.NoError(t, err)
	schema2, err := db.GetSchemaByVersionID(ctx, cols2[0].SchemaVersionID)
	require.NoError(t, err)
	err = db.AddP2PCollections(ctx, []string{schema1.Root, schema2.Root})
	require.NoError(t, err)
	// Check that the event was published
	for msg := range sub.Message() {
		p2pTopic := msg.Data.(event.P2PTopic)
		require.Equal(t, []string{schema1.Root, schema2.Root}, p2pTopic.ToAdd)
		break
	}
}

func TestRemoveP2PCollection_WithInvalidCollection_ShouldError(t *testing.T) {
	ctx := context.Background()
	db, err := newDefraMemoryDB(ctx)
	require.NoError(t, err)
	defer db.Close()
	err = db.RemoveP2PCollections(ctx, []string{"invalidCollection"})
	require.ErrorIs(t, err, client.ErrCollectionNotFound)
}

func TestRemoveP2PCollection_WithValidCollection_ShouldSucceed(t *testing.T) {
	ctx := context.Background()
	db, err := newDefraMemoryDB(ctx)
	require.NoError(t, err)
	defer db.Close()
	sub, err := db.events.Subscribe(event.P2PTopicName)
	require.NoError(t, err)
	cols, err := db.AddSchema(ctx, `type User { name: String }`)
	require.NoError(t, err)
	schema, err := db.GetSchemaByVersionID(ctx, cols[0].SchemaVersionID)
	require.NoError(t, err)
	err = db.AddP2PCollections(ctx, []string{schema.Root})
	require.NoError(t, err)
	// Check that the event was published
	for msg := range sub.Message() {
		p2pTopic := msg.Data.(event.P2PTopic)
		require.Equal(t, []string{schema.Root}, p2pTopic.ToAdd)
		break
	}
	err = db.RemoveP2PCollections(ctx, []string{schema.Root})
	require.NoError(t, err)
	// Check that the event was published
	for msg := range sub.Message() {
		p2pTopic := msg.Data.(event.P2PTopic)
		require.Equal(t, []string{schema.Root}, p2pTopic.ToRemove)
		break
	}
}

func TestRemoveP2PCollection_WithValidCollectionAndDoc_ShouldSucceed(t *testing.T) {
	ctx := context.Background()
	db, err := newDefraMemoryDB(ctx)
	require.NoError(t, err)
	defer db.Close()
	sub, err := db.events.Subscribe(event.P2PTopicName)
	require.NoError(t, err)
	cols, err := db.AddSchema(ctx, `type User { name: String }`)
	require.NoError(t, err)
	col, err := db.GetCollectionByName(ctx, cols[0].Name.Value())
	require.NoError(t, err)
	doc, err := client.NewDocFromMap(map[string]any{"name": "Alice"}, col.Definition())
	require.NoError(t, err)
	err = col.Create(ctx, doc)
	require.NoError(t, err)

	err = db.AddP2PCollections(ctx, []string{col.SchemaRoot()})
	require.NoError(t, err)
	// Check that the event was published
	for msg := range sub.Message() {
		p2pTopic := msg.Data.(event.P2PTopic)
		require.Equal(t, []string{col.SchemaRoot()}, p2pTopic.ToAdd)
		require.Equal(t, []string{doc.ID().String()}, p2pTopic.ToRemove)
		break
	}
	err = db.RemoveP2PCollections(ctx, []string{col.SchemaRoot()})
	require.NoError(t, err)
	// Check that the event was published
	for msg := range sub.Message() {
		p2pTopic := msg.Data.(event.P2PTopic)
		require.Equal(t, []string{col.SchemaRoot()}, p2pTopic.ToRemove)
		require.Equal(t, []string{doc.ID().String()}, p2pTopic.ToAdd)
		break
	}
}

func TestLoadP2PCollection_WithValidCollectionsAndDocs_ShouldSucceed(t *testing.T) {
	ctx := context.Background()
	db, err := newDefraMemoryDB(ctx)
	require.NoError(t, err)
	defer db.Close()
	sub, err := db.events.Subscribe(event.P2PTopicName)
	require.NoError(t, err)
	cols1, err := db.AddSchema(ctx, `type User { name: String }`)
	require.NoError(t, err)
	col1, err := db.GetCollectionByName(ctx, cols1[0].Name.Value())
	require.NoError(t, err)
	doc1, err := client.NewDocFromMap(map[string]any{"name": "Alice"}, col1.Definition())
	require.NoError(t, err)
	err = col1.Create(ctx, doc1)
	require.NoError(t, err)

	cols2, err := db.AddSchema(ctx, `type Book { name: String }`)
	require.NoError(t, err)
	col2, err := db.GetCollectionByName(ctx, cols2[0].Name.Value())
	require.NoError(t, err)
	doc2, err := client.NewDocFromMap(map[string]any{"name": "Some book"}, col2.Definition())
	require.NoError(t, err)
	err = col2.Create(ctx, doc2)
	require.NoError(t, err)

	err = db.AddP2PCollections(ctx, []string{col1.SchemaRoot()})
	require.NoError(t, err)
	// Check that the event was published
	for msg := range sub.Message() {
		p2pTopic := msg.Data.(event.P2PTopic)
		require.Equal(t, []string{col1.SchemaRoot()}, p2pTopic.ToAdd)
		require.Equal(t, []string{doc1.ID().String()}, p2pTopic.ToRemove)
		break
	}
	err = db.loadAndPublishP2PCollections(ctx)
	require.NoError(t, err)
	// Check that the event was published
	msg := <-sub.Message()
	p2pTopic := msg.Data.(event.P2PTopic)
	require.Equal(t, []string{col1.SchemaRoot()}, p2pTopic.ToAdd)
	msg = <-sub.Message()
	p2pTopic = msg.Data.(event.P2PTopic)
	require.Equal(t, []string{doc2.ID().String()}, p2pTopic.ToAdd)
}

func TestGetAllP2PCollections_WithMultipleValidCollections_ShouldSucceed(t *testing.T) {
	ctx := context.Background()
	db, err := newDefraMemoryDB(ctx)
	require.NoError(t, err)
	defer db.Close()
	cols1, err := db.AddSchema(ctx, `type User { name: String }`)
	require.NoError(t, err)
	schema1, err := db.GetSchemaByVersionID(ctx, cols1[0].SchemaVersionID)
	require.NoError(t, err)
	cols2, err := db.AddSchema(ctx, `type Books { name: String }`)
	require.NoError(t, err)
	schema2, err := db.GetSchemaByVersionID(ctx, cols2[0].SchemaVersionID)
	require.NoError(t, err)
	err = db.AddP2PCollections(ctx, []string{schema1.Root, schema2.Root})
	require.NoError(t, err)
	cols, err := db.GetAllP2PCollections(ctx)
	require.NoError(t, err)
	require.Equal(t, []string{schema2.Root, schema1.Root}, cols)
}
