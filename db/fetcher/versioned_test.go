package fetcher_test

import (
	"testing"

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/db/fetcher"
	"github.com/sourcenetwork/defradb/document"

	"github.com/stretchr/testify/assert"
)

func newMemoryDB() (*db.DB, error) {
	opts := &db.Options{
		Store: "memory",
		Memory: db.MemoryOptions{
			Size: 1024 * 1000,
		},
	}

	return db.NewDB(opts)
}

func TestVersionedFetcherInit(t *testing.T) {
	db, err := newMemoryDB()
	assert.NoError(t, err)

	col, err := newTestCollectionWithSchema(db)
	assert.NoError(t, err)

	vf := &fetcher.VersionedFetcher{}
	desc := col.Description()
	err = vf.Init(&desc, nil, nil, false)
	assert.NoError(t, err)
}

func TestVersionedFetcherStart(t *testing.T) {
	db, err := newMemoryDB()
	assert.NoError(t, err)

	col, err := newTestCollectionWithSchema(db)
	assert.NoError(t, err)

	err = createDocUpdates(col)
	assert.NoError(t, err)

	vf := &fetcher.VersionedFetcher{}
	desc := col.Description()
	err = vf.Init(&desc, nil, nil, false)
	assert.NoError(t, err)

	err = vf.Start()
}

func createDocUpdates(col *db.Collection) error {
	// col, err := newTestCollectionWithSchema(db)
	// if err != ni

	// dockey: bae-ed7f0bd5-3f5b-5e93-9310-4b2e71ac460d
	testJSONObj := []byte(`{
		"name": "Alice",
		"age": 31,
		"points": 100,
		"verified": true
	}`)

	doc, err := document.NewFromJSON(testJSONObj)
	if err != nil {
		return err
	}

	if err := col.Save(doc); err != nil {
		return err
	}

	// update #1
	doc.Set("name", "Pete")
	doc.Set("points", 99.9)
	if err := col.Update(doc); err != nil {
		return err
	}

	// update #2
	doc.Set("verified", false)
	doc.Set("age", 22)
	if err := col.Update(doc); err != nil {
		return err
	}

	// update #3
	doc.Set("points", 129.99)
	err = col.Update(doc)
	return err
}

func newTestCollectionWithSchema(d *db.DB) (*db.Collection, error) {
	desc := base.CollectionDescription{
		Name: "users",
		Schema: base.SchemaDescription{
			Fields: []base.FieldDescription{
				base.FieldDescription{
					Name: "_key",
					Kind: base.FieldKind_DocKey,
				},
				base.FieldDescription{
					Name: "name",
					Kind: base.FieldKind_STRING,
					Typ:  core.LWW_REGISTER,
				},
				base.FieldDescription{
					Name: "age",
					Kind: base.FieldKind_INT,
					Typ:  core.LWW_REGISTER,
				},
				base.FieldDescription{
					Name: "verified",
					Kind: base.FieldKind_BOOL,
					Typ:  core.LWW_REGISTER,
				},
				base.FieldDescription{
					Name: "points",
					Kind: base.FieldKind_FLOAT,
					Typ:  core.LWW_REGISTER,
				},
			},
		},
	}

	col, err := d.CreateCollection(desc)
	return col.(*db.Collection), err
}
