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
			ID:       uint32(1),
			FieldIDs: []uint32{1, 2, 3},
			Fields: []base.FieldDescription{
				base.FieldDescription{
					Name: "_key",
					ID:   uint32(1),
					Kind: base.FieldKind_DocKey,
				},
				base.FieldDescription{
					Name: "Name",
					ID:   uint32(2),
					Kind: base.FieldKind_STRING,
					Typ:  core.LWW_REGISTER,
				},
				base.FieldDescription{
					Name: "Age",
					ID:   uint32(3),
					Kind: base.FieldKind_INT,
					Typ:  core.LWW_REGISTER,
				},
			},
		},
		Indexes: []base.IndexDescription{
			{
				Name:    "primary",
				ID:      uint32(0),
				Primary: true,
				Unique:  true,
			},
		},
	}

	return db.CreateCollection(desc)
}

func TestNewCollection(t *testing.T) {
	db, err := newMemoryDB()
	assert.NoError(t, err)

	_, err = newTestCollection(db)
	assert.NoError(t, err)
}

func TestNewCollectionWithSchema(t *testing.T) {
	db, err := newMemoryDB()
	assert.NoError(t, err)

	_, err = newTestCollectionWithSchema(db)
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
