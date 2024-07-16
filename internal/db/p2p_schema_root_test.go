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
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/acp"
	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore/memory"
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

// This test documents that we don't allow adding p2p collections that have a policy
// until the following is implemented:
// TODO-ACP: ACP <> P2P https://github.com/sourcenetwork/defradb/issues/2366
func TestAddP2PCollectionsWithPermissionedCollection_Error(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)
	db, err := newDB(ctx, rootstore, immutable.Some[acp.ACP](acp.NewLocalACP()), nil)
	require.NoError(t, err)

	policy := `
        name: test
        description: a policy
        actor:
          name: actor
        resources:
          user:
            permissions:
              read:
                expr: owner
              write:
                expr: owner
            relations:
              owner:
                types:
                  - actor
    `

	privKeyBytes, err := hex.DecodeString("028d53f37a19afb9a0dbc5b4be30c65731479ee8cfa0c9bc8f8bf198cc3c075f")
	require.NoError(t, err)
	privKey := secp256k1.PrivKeyFromBytes(privKeyBytes)
	identity, err := acpIdentity.FromPrivateKey(privKey, time.Hour, immutable.None[string](), immutable.None[string](), false)
	require.NoError(t, err)

	ctx = SetContextIdentity(ctx, immutable.Some(identity))
	policyResult, err := db.AddPolicy(ctx, policy)
	policyID := policyResult.PolicyID
	require.NoError(t, err)
	require.Equal(t, "7b5ed30570e8d9206027ef6d5469879a6c1ea4595625c6ca33a19063a6ed6214", policyID)

	schema := fmt.Sprintf(`
		type User @policy(id: "%s", resource: "user") {
			name: String
			age: Int
		}
	`, policyID,
	)
	_, err = db.AddSchema(ctx, schema)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	err = db.AddP2PCollections(ctx, []string{col.SchemaRoot()})
	require.Error(t, err)
	require.ErrorIs(t, err, ErrP2PColHasPolicy)
}
