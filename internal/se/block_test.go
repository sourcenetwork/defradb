// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package se

import (
	"context"
	"testing"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/datastore"
	secore "github.com/sourcenetwork/defradb/internal/se/core"
)

type mockTxn struct {
	datastore.Txn
	onSuccessCallbacks []func()
}

func (m *mockTxn) OnSuccess(fn func()) {
	m.onSuccessCallbacks = append(m.onSuccessCallbacks, fn)
}

func createTestContext() (*Context, *mockTxn) {
	txn := &mockTxn{}
	seCtx := &Context{
		config: Config{
			Key:          []byte("test-key-32-bytes-long-for-hmac!"),
			CollectionID: "collection123",
			EncryptedFields: []client.EncryptedIndexDescription{
				{
					FieldName: "email",
					Type:      client.EncryptedIndexTypeEquality,
				},
			},
		},
		artifacts: make([]secore.Artifact, 0),
		txn:       txn,
	}
	return seCtx, txn
}

func TestContext_StoreAndRetrieve(t *testing.T) {
	ctx := context.Background()
	seCtx, _ := createTestContext()

	// Add context to ctx
	ctx = context.WithValue(ctx, contextKey{}, seCtx)

	// Verify we can retrieve context
	retrievedCtx, ok := ctx.Value(contextKey{}).(*Context)
	if !ok {
		t.Fatal("failed to retrieve SE context")
	}

	if retrievedCtx != seCtx {
		t.Error("retrieved context should be the same as stored context")
	}
}

func TestContext_CollectionID(t *testing.T) {
	ctx := context.Background()
	seCtx, _ := createTestContext()

	ctx = context.WithValue(ctx, contextKey{}, seCtx)

	retrievedCtx, ok := ctx.Value(contextKey{}).(*Context)
	if !ok {
		t.Fatal("failed to retrieve SE context")
	}

	if retrievedCtx.config.CollectionID != "collection123" {
		t.Errorf("expected collection ID 'collection123', got %v", retrievedCtx.config.CollectionID)
	}
}

func TestContext_EncryptedFields(t *testing.T) {
	ctx := context.Background()
	seCtx, _ := createTestContext()

	ctx = context.WithValue(ctx, contextKey{}, seCtx)

	retrievedCtx, ok := ctx.Value(contextKey{}).(*Context)
	if !ok {
		t.Fatal("failed to retrieve SE context")
	}

	if len(retrievedCtx.config.EncryptedFields) != 1 {
		t.Fatalf("expected 1 encrypted field, got %d", len(retrievedCtx.config.EncryptedFields))
	}

	field := retrievedCtx.config.EncryptedFields[0]
	if field.FieldName != "email" {
		t.Errorf("expected field name 'email', got %v", field.FieldName)
	}

	if field.Type != client.EncryptedIndexTypeEquality {
		t.Errorf("expected equality type, got %v", field.Type)
	}
}

func TestContext_ArtifactAddition(t *testing.T) {
	seCtx, _ := createTestContext()

	artifact := secore.Artifact{
		Type:         secore.ArtifactTypeEqualityTag,
		CollectionID: "collection123",
		FieldName:    "email",
		SearchTag:    []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		Operation:    secore.OperationAdd,
	}

	seCtx.artifacts = append(seCtx.artifacts, artifact)

	if len(seCtx.artifacts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(seCtx.artifacts))
	}
}

func TestContext_ArtifactProperties(t *testing.T) {
	seCtx, _ := createTestContext()

	artifact := secore.Artifact{
		Type:         secore.ArtifactTypeEqualityTag,
		CollectionID: "collection123",
		FieldName:    "email",
		SearchTag:    []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		Operation:    secore.OperationAdd,
	}

	seCtx.artifacts = append(seCtx.artifacts, artifact)

	if seCtx.artifacts[0].FieldName != "email" {
		t.Errorf("expected field name 'email', got %v", seCtx.artifacts[0].FieldName)
	}

	if seCtx.artifacts[0].Type != secore.ArtifactTypeEqualityTag {
		t.Errorf("expected artifact type equality tag, got %v", seCtx.artifacts[0].Type)
	}

	if seCtx.artifacts[0].CollectionID != "collection123" {
		t.Errorf("expected collection ID 'collection123', got %v", seCtx.artifacts[0].CollectionID)
	}

	if seCtx.artifacts[0].Operation != secore.OperationAdd {
		t.Errorf("expected operation add, got %v", seCtx.artifacts[0].Operation)
	}

	if len(seCtx.artifacts[0].SearchTag) != 16 {
		t.Errorf("expected tag length 16, got %d", len(seCtx.artifacts[0].SearchTag))
	}
}

func TestContext_MultipleArtifacts(t *testing.T) {
	seCtx, _ := createTestContext()

	artifact1 := secore.Artifact{
		Type:         secore.ArtifactTypeEqualityTag,
		CollectionID: "collection123",
		FieldName:    "email",
		SearchTag:    []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		Operation:    secore.OperationAdd,
	}

	artifact2 := secore.Artifact{
		Type:         secore.ArtifactTypeEqualityTag,
		CollectionID: "collection123",
		FieldName:    "name",
		SearchTag:    []byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
		Operation:    secore.OperationAdd,
	}

	seCtx.artifacts = append(seCtx.artifacts, artifact1, artifact2)

	if len(seCtx.artifacts) != 2 {
		t.Fatalf("expected 2 artifacts, got %d", len(seCtx.artifacts))
	}

	if seCtx.artifacts[0].FieldName != "email" {
		t.Errorf("expected first artifact field name 'email', got %v", seCtx.artifacts[0].FieldName)
	}

	if seCtx.artifacts[1].FieldName != "name" {
		t.Errorf("expected second artifact field name 'name', got %v", seCtx.artifacts[1].FieldName)
	}
}

func TestContext_EmptyContext(t *testing.T) {
	ctx := context.Background()

	// Try to retrieve context that wasn't set
	retrievedCtx, ok := ctx.Value(contextKey{}).(*Context)
	if ok {
		t.Error("expected no context, but found one")
	}

	if retrievedCtx != nil {
		t.Error("expected nil context")
	}
}

func TestContext_TransactionCallback(t *testing.T) {
	seCtx, txn := createTestContext()

	// Register a callback through the context
	seCtx.registerReplicationCallback()

	if len(txn.onSuccessCallbacks) != 1 {
		t.Fatalf("expected 1 callback, got %d", len(txn.onSuccessCallbacks))
	}
}
