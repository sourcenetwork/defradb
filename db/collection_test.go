package db

import (
	"reflect"
	"testing"

	"github.com/sourcenetwork/defradb/db/base"
	"github.com/stretchr/testify/assert"
)

func TestNewCollection(t *testing.T) {
	db, err := newMemoryDB()
	assert.NoError(t, err)

	_, err = newTestCollection(db)
	assert.NoError(t, err)
}

func TestCollectionProperties(t *testing.T) {
	db, err := newMemoryDB()
	assert.NoError(t, err)

	col, err := newTestCollection(db)
	assert.NoError(t, err)

	assert.Equal(t, "test", col.Name())
	assert.Equal(t, uint32(1), col.ID())
	assert.True(t, reflect.DeepEqual(col.Schema(), base.SchemaDescription{}))
	assert.Equal(t, 1, len(col.Indexes()))
}

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
