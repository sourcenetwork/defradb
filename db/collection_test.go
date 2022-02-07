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
	"reflect"
	"testing"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/stretchr/testify/assert"
)

func newTestCollectionWithSchema(ctx context.Context, db *DB) (*Collection, error) {
	desc := base.CollectionDescription{
		Name: "users",
		Schema: base.SchemaDescription{
			Fields: []base.FieldDescription{
				{
					Name: "_key",
					Kind: base.FieldKind_DocKey,
				},
				{
					Name: "Name",
					Kind: base.FieldKind_STRING,
					Typ:  core.LWW_REGISTER,
				},
				{
					Name: "Age",
					Kind: base.FieldKind_INT,
					Typ:  core.LWW_REGISTER,
				},
				{
					Name: "Weight",
					Kind: base.FieldKind_FLOAT,
					Typ:  core.LWW_REGISTER,
				},
			},
		},
	}

	col, err := db.CreateCollection(ctx, desc)
	return col.(*Collection), err
}

func createNewTestCollection(ctx context.Context, db *DB) (client.Collection, error) {
	return db.CreateCollection(ctx, base.CollectionDescription{
		Name: "test",
	})
}

func TestNewCollection_ReturnsError_GivenNoSchema(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB()
	assert.NoError(t, err)

	_, err = createNewTestCollection(ctx, db)
	assert.Error(t, err)
}

func TestNewCollectionWithSchema(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB()
	assert.NoError(t, err)

	col, err := newTestCollectionWithSchema(ctx, db)
	assert.NoError(t, err)

	schema := col.Schema()
	desc := col.Description()

	assert.True(t, reflect.DeepEqual(schema, desc.Schema))
	assert.Equal(t, "users", col.Name())
	assert.Equal(t, uint32(1), col.ID())
	assert.False(t, reflect.DeepEqual(schema, base.SchemaDescription{}))
	assert.Equal(t, 1, len(col.Indexes()))
	assert.Equal(t, 4, len(schema.Fields))
	assert.Equal(t, 4, len(schema.FieldIDs))

	for i := 0; i < 4; i++ {
		assert.Equal(t, uint32(i), schema.FieldIDs[i])
		assert.Equal(t, base.FieldID(i), schema.Fields[i].ID)
	}
}

func TestGetCollection(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB()
	assert.NoError(t, err)

	_, err = newTestCollectionWithSchema(ctx, db)
	assert.NoError(t, err)

	col, err := db.GetCollection(ctx, "users")
	assert.NoError(t, err)

	schema := col.Schema()
	desc := col.Description()

	assert.True(t, reflect.DeepEqual(schema, desc.Schema))
	assert.Equal(t, "users", col.Name())
	assert.Equal(t, uint32(1), col.ID())
	assert.False(t, reflect.DeepEqual(schema, base.SchemaDescription{}))
	assert.Equal(t, 1, len(col.Indexes()))
	assert.Equal(t, 4, len(schema.Fields))
	assert.Equal(t, 4, len(schema.FieldIDs))

	for i := 0; i < 4; i++ {
		assert.Equal(t, uint32(i), schema.FieldIDs[i])
		assert.Equal(t, base.FieldID(i), schema.Fields[i].ID)
	}
}
