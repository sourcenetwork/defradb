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

	assert.Equal(t, col.Name(), "test")
	assert.Equal(t, col.ID(), uint32(1))
	assert.True(t, reflect.DeepEqual(col.Schema(), base.SchemaDescription{}))
	assert.Equal(t, len(col.Indexes()), 1)
}
