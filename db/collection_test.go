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

	"github.com/sourcenetwork/defradb/client"
	"github.com/stretchr/testify/assert"
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

func TestNewCollectionWithDescription(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)

	col, err := newTestCollectionWithSchema(ctx, db)
	assert.NoError(t, err)

	desc := col.Description()

	assert.Equal(t, "users", col.Name())
	assert.Equal(t, uint32(1), col.ID())
	assert.Equal(t, 1, len(col.Indexes()))
	assert.Equal(t, 4, len(desc.Schema.Fields))

	for i := 0; i < 4; i++ {
		assert.Equal(t, client.FieldID(i), desc.Schema.Fields[i].ID)
	}
}

func TestNewCollectionReturnsErrorGivenDuplicateSchema(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)

	_, err = newTestCollectionWithSchema(ctx, db)
	assert.NoError(t, err)

	_, err = newTestCollectionWithSchema(ctx, db)
	assert.Errorf(t, err, "Collection already exists")
}

func TestNewCollectionReturnsErrorGivenNoFields(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)

	desc := client.CollectionDescription{
		Name: "users",
	}

	_, err = db.CreateCollection(ctx, desc)
	assert.EqualError(t, err, "Collection schema has no fields")
}

func TestNewCollectionReturnsErrorGivenNoName(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)

	desc := client.CollectionDescription{
		Name: "",
	}

	_, err = db.CreateCollection(ctx, desc)
	assert.EqualError(t, err, "Collection requires name to not be empty")
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
	assert.EqualError(t, err, "Collection schema first field must be a DocKey")
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
	assert.EqualError(t, err, "Collection schema first field must be a DocKey")
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
	assert.EqualError(t, err, "Collection schema field missing Name")
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
	assert.EqualError(t, err, "Collection schema field missing FieldKind")
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
	assert.EqualError(t, err, "Collection schema field missing CRDT type")
}

func TestGetCollectionByName(t *testing.T) {
	ctx := context.Background()
	db, err := newMemoryDB(ctx)
	assert.NoError(t, err)

	_, err = newTestCollectionWithSchema(ctx, db)
	assert.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "users")
	assert.NoError(t, err)

	desc := col.Description()

	assert.Equal(t, "users", col.Name())
	assert.Equal(t, uint32(1), col.ID())
	assert.Equal(t, 1, len(col.Indexes()))
	assert.Equal(t, 4, len(desc.Schema.Fields))

	for i := 0; i < 4; i++ {
		assert.Equal(t, client.FieldID(i), desc.Schema.Fields[i].ID)
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
	assert.EqualError(t, err, "Collection name can't be empty")
}
