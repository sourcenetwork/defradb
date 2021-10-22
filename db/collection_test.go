// Copyright 2020 Source Inc.
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
	"reflect"
	"testing"

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/stretchr/testify/assert"
)

func newTestCollectionWithSchema(db *DB) (*Collection, error) {
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
			},
		},
	}

	col, err := db.CreateCollection(desc)
	return col.(*Collection), err
}

func TestNewCollection(t *testing.T) {
	db, err := newMemoryDB()
	assert.NoError(t, err)

	col, err := newTestCollection(db)
	assert.NoError(t, err)

	assert.Equal(t, "test", col.Name())
	assert.Equal(t, uint32(1), col.ID())
	assert.True(t, reflect.DeepEqual(col.Schema(), base.SchemaDescription{}))
	assert.Equal(t, 1, len(col.Indexes()))
}

func TestNewCollectionWithSchema(t *testing.T) {
	db, err := newMemoryDB()
	assert.NoError(t, err)

	col, err := newTestCollectionWithSchema(db)
	assert.NoError(t, err)

	desc := col.Description()
	schema := col.Schema()
	// indexes := col.Indexes()

	assert.True(t, reflect.DeepEqual(schema, desc.Schema))
	assert.Equal(t, "users", col.Name())
	assert.Equal(t, uint32(1), col.ID())
	assert.False(t, reflect.DeepEqual(schema, base.SchemaDescription{}))
	assert.Equal(t, 1, len(col.Indexes()))
	assert.Equal(t, 3, len(schema.Fields))
	assert.Equal(t, 3, len(schema.FieldIDs))

	for i := 0; i < 3; i++ {
		assert.Equal(t, uint32(i), schema.FieldIDs[i])
		assert.Equal(t, base.FieldID(i), schema.Fields[i].ID)
	}
}

// func TestCollectionIndexes

func TestGetCollection(t *testing.T) {
	db, err := newMemoryDB()
	assert.NoError(t, err)

	_, err = newTestCollection(db)
	assert.NoError(t, err)

	col, err := db.GetCollection("test")
	assert.NoError(t, err)

	assert.Equal(t, "test", col.Name())
	assert.Equal(t, uint32(1), col.ID())
	assert.True(t, reflect.DeepEqual(col.Schema(), base.SchemaDescription{}))
	assert.Equal(t, 1, len(col.Indexes()))
}
