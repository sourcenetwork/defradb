// Copyright 2025 Democratized Data Foundation
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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/crypto"
)

const testUserSchema = `type User {
	name: String
	age: Int
}`

func setupTestDB(t *testing.T) *DB {
	db, err := newBadgerDB(context.Background())
	require.NoError(t, err)

	_, err = db.AddSchema(context.Background(), testUserSchema)
	require.NoError(t, err)

	return db
}

func createIdentity(t *testing.T, keyType crypto.KeyType) identity.Identity {
	privKey, err := crypto.GenerateKey(keyType)
	require.NoError(t, err)

	return identity.Identity{
		PublicKey:  privKey.GetPublic(),
		PrivateKey: privKey,
	}
}

func createTestDoc(t *testing.T, db *DB, ctx context.Context, docMap map[string]any) (*client.Document, error) {
	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	doc, err := client.NewDocFromMap(docMap, col.Definition())
	if err != nil {
		return nil, err
	}

	err = col.Create(ctx, doc)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

func TestVerifySignatures_WithValidSignature_SuccessfullyVerifies(t *testing.T) {
	db := setupTestDB(t)
	ident := createIdentity(t, crypto.KeyTypeSecp256k1)

	docMap := map[string]any{
		"name": "John",
		"age":  30,
	}

	ctx := identity.WithContext(context.Background(), immutable.Some(ident))
	doc, err := createTestDoc(t, db, ctx, docMap)
	require.NoError(t, err)

	err = db.VerifyBlock(ctx, doc.Head().String())
	require.NoError(t, err)
}

func TestVerifySignatures_WithUpdateBlock_SuccessfullyVerifies(t *testing.T) {
	db := setupTestDB(t)
	ident := createIdentity(t, crypto.KeyTypeSecp256k1)

	ctx := identity.WithContext(context.Background(), immutable.Some(ident))
	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	docMap := map[string]any{
		"name": "John",
		"age":  30,
	}

	doc, err := client.NewDocFromMap(docMap, col.Definition())
	require.NoError(t, err)

	err = col.Create(ctx, doc)
	require.NoError(t, err)
	createCid := doc.Head()

	err = doc.SetWithJSON([]byte(`{"name": "John Doe"}`))
	require.NoError(t, err)

	err = col.Update(ctx, doc)
	require.NoError(t, err)
	updateCid := doc.Head()

	require.NotEqual(t, createCid, updateCid)

	err = db.VerifyBlock(ctx, createCid.String())
	require.NoError(t, err)

	err = db.VerifyBlock(ctx, updateCid.String())
	require.NoError(t, err)
}

func TestVerifySignatures_WithInvalidCID_ReturnsError(t *testing.T) {
	db := setupTestDB(t)

	err := db.VerifyBlock(context.Background(), "invalid-cid")
	require.Error(t, err)
}

func TestVerifySignatures_WithoutIdentity_ReturnsError(t *testing.T) {
	db := setupTestDB(t)
	ident := createIdentity(t, crypto.KeyTypeSecp256k1)

	docMap := map[string]any{
		"name": "John",
		"age":  30,
	}

	ctx := identity.WithContext(context.Background(), immutable.Some(ident))
	doc, err := createTestDoc(t, db, ctx, docMap)
	require.NoError(t, err)

	err = db.VerifyBlock(context.Background(), doc.Head().String())
	require.ErrorIs(t, err, ErrNoIdentityInContext)
}

func TestVerifySignatures_WithDifferentIdentity_ReturnsError(t *testing.T) {
	db := setupTestDB(t)
	ident1 := createIdentity(t, crypto.KeyTypeSecp256k1)

	docMap := map[string]any{
		"name": "John",
		"age":  30,
	}

	ctx := identity.WithContext(context.Background(), immutable.Some(ident1))
	doc, err := createTestDoc(t, db, ctx, docMap)
	require.NoError(t, err)

	privKey2, err := crypto.GenerateKey(crypto.KeyTypeSecp256k1)
	require.NoError(t, err)

	pubKey2 := privKey2.GetPublic()
	ident2 := identity.Identity{
		PublicKey:  pubKey2,
		PrivateKey: privKey2,
	}

	ctx = identity.WithContext(context.Background(), immutable.Some(ident2))
	err = db.VerifyBlock(ctx, doc.Head().String())
	require.Error(t, err)
}

func TestVerifySignatures_WithDifferentKeyTypes_SuccessfullyVerifies(t *testing.T) {
	db := setupTestDB(t)
	ident := createIdentity(t, crypto.KeyTypeEd25519)

	docMap := map[string]any{
		"name": "John",
		"age":  30,
	}

	ctx := identity.WithContext(context.Background(), immutable.Some(ident))
	doc, err := createTestDoc(t, db, ctx, docMap)
	require.NoError(t, err)

	err = db.VerifyBlock(ctx, doc.Head().String())
	require.NoError(t, err)
}
