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

	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/defradb/client"
)

func newTestCollectionWithSchema(ctx context.Context, db client.DB) (client.Collection, error) {
	desc := client.CollectionDescription{
		Name: "users",
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
				{
					Name: "Age",
					Kind: client.FieldKind_INT,
					Typ:  client.LWW_REGISTER,
				},
				{
					Name: "Weight",
					Kind: client.FieldKind_FLOAT,
					Typ:  client.LWW_REGISTER,
				},
			},
		},
	}

	col, err := db.CreateCollection(ctx, desc)
	return col, err
}

func createNewTestCollection(ctx context.Context, db client.DB) (client.Collection, error) {
	return db.CreateCollection(ctx, client.CollectionDescription{
		Name: "test",
	})
}

func TestNewCollection_ReturnsError_GivenNoSchema(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)

	_, err = createNewTestCollection(ctx, db)
	assert.Error(t, err)
}

func TestNewCollectionWithSchema(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)

	col, err := newTestCollectionWithSchema(ctx, db)
	assert.NoError(t, err)

	schema := col.Schema()
	desc := col.Description()

	assert.True(t, reflect.DeepEqual(schema, desc.Schema))
	assert.Equal(t, "users", col.Name())
	assert.Equal(t, uint32(1), col.ID())
	assert.False(t, reflect.DeepEqual(schema, client.SchemaDescription{}))
	assert.Equal(t, 4, len(schema.Fields))

	for i := 0; i < 4; i++ {
		assert.Equal(t, client.FieldID(i), schema.Fields[i].ID)
	}
}

func TestNewCollectionReturnsErrorGivenDuplicateSchema(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)

	_, err = newTestCollectionWithSchema(ctx, db)
	assert.NoError(t, err)

	_, err = newTestCollectionWithSchema(ctx, db)
	assert.Errorf(t, err, "collection already exists")
}

func TestNewCollectionReturnsErrorGivenNoFields(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)

	desc := client.CollectionDescription{
		Name: "users",
		Schema: client.SchemaDescription{
			Fields: []client.FieldDescription{},
		},
	}

	_, err = db.CreateCollection(ctx, desc)
	assert.EqualError(
		t,
		err,
		"invalid state, required property is uninitialized. Host: Collection, PropertyName: Fields",
	)
}

func TestNewCollectionReturnsErrorGivenNoName(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)

	desc := client.CollectionDescription{
		Name: "",
		Schema: client.SchemaDescription{
			Fields: []client.FieldDescription{},
		},
	}

	_, err = db.CreateCollection(ctx, desc)
	assert.EqualError(
		t,
		err,
		"invalid state, required property is uninitialized. Host: Collection, PropertyName: Name",
	)
}

func TestNewCollectionReturnsErrorGivenNoKeyField(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)

	desc := client.CollectionDescription{
		Name: "users",
		Schema: client.SchemaDescription{
			Fields: []client.FieldDescription{
				{
					Name: "Name",
					Kind: client.FieldKind_STRING,
					Typ:  client.LWW_REGISTER,
				},
			},
		},
	}

	_, err = db.CreateCollection(ctx, desc)
	assert.EqualError(t, err, "collection schema first field must be a DocKey")
}

func TestNewCollectionReturnsErrorGivenKeyFieldIsNotFirstField(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)

	desc := client.CollectionDescription{
		Name: "users",
		Schema: client.SchemaDescription{
			Fields: []client.FieldDescription{
				{
					Name: "Name",
					Kind: client.FieldKind_STRING,
					Typ:  client.LWW_REGISTER,
				},
				{
					Name: "_key",
					Kind: client.FieldKind_DocKey,
				},
			},
		},
	}

	_, err = db.CreateCollection(ctx, desc)
	assert.EqualError(t, err, "collection schema first field must be a DocKey")
}

func TestNewCollectionReturnsErrorGivenFieldWithNoName(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)

	desc := client.CollectionDescription{
		Name: "users",
		Schema: client.SchemaDescription{
			Fields: []client.FieldDescription{
				{
					Name: "_key",
					Kind: client.FieldKind_DocKey,
				},
				{
					Name: "",
					Kind: client.FieldKind_STRING,
					Typ:  client.LWW_REGISTER,
				},
			},
		},
	}

	_, err = db.CreateCollection(ctx, desc)
	assert.EqualError(
		t,
		err,
		"invalid state, required property is uninitialized. Host: Collection.Schema, PropertyName: Name",
	)
}

func TestNewCollectionReturnsErrorGivenFieldWithNoKind(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)

	desc := client.CollectionDescription{
		Name: "users",
		Schema: client.SchemaDescription{
			Fields: []client.FieldDescription{
				{
					Name: "_key",
					Kind: client.FieldKind_DocKey,
				},
				{
					Name: "Name",
					Typ:  client.LWW_REGISTER,
				},
			},
		},
	}

	_, err = db.CreateCollection(ctx, desc)
	assert.EqualError(
		t,
		err,
		"invalid state, required property is uninitialized. Host: Collection.Schema, PropertyName: FieldKind",
	)
}

func TestNewCollectionReturnsErrorGivenFieldWithNoType(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)

	desc := client.CollectionDescription{
		Name: "users",
		Schema: client.SchemaDescription{
			Fields: []client.FieldDescription{
				{
					Name: "_key",
					Kind: client.FieldKind_DocKey,
				},
				{
					Name: "Name",
					Kind: client.FieldKind_STRING,
				},
			},
		},
	}

	_, err = db.CreateCollection(ctx, desc)
	assert.EqualError(
		t,
		err,
		"invalid state, required property is uninitialized. Host: Collection.Schema, PropertyName: CRDT type",
	)
}

func TestGetCollectionByName(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)

	_, err = newTestCollectionWithSchema(ctx, db)
	assert.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "users")
	assert.NoError(t, err)

	schema := col.Schema()
	desc := col.Description()

	assert.True(t, reflect.DeepEqual(schema, desc.Schema))
	assert.Equal(t, "users", col.Name())
	assert.Equal(t, uint32(1), col.ID())
	assert.False(t, reflect.DeepEqual(schema, client.SchemaDescription{}))
	assert.Equal(t, 4, len(schema.Fields))

	for i := 0; i < 4; i++ {
		assert.Equal(t, client.FieldID(i), schema.Fields[i].ID)
	}
}

func TestGetCollectionByNameReturnsErrorGivenNonExistantCollection(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)

	_, err = db.GetCollectionByName(ctx, "doesNotExist")
	assert.EqualError(t, err, "datastore: key not found")
}

func TestGetCollectionByNameReturnsErrorGivenEmptyString(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)

	_, err = db.GetCollectionByName(ctx, "")
	assert.EqualError(t, err, "collection name can't be empty")
}
