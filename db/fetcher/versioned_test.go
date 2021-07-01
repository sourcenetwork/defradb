package fetcher_test

import (
	"fmt"
	"testing"

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/db/fetcher"
	"github.com/sourcenetwork/defradb/document"

	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	dshelp "github.com/ipfs/go-ipfs-ds-help"
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

	db.PrintDump()

	k := ds.NewKey("CIQC6HCXT2MS25VZIBFGYJKFAXMEZVVOJGLXM5DO4XF6WWJK43BB3PA")
	c, err := dshelp.DsKeyToCid(k)
	fmt.Println(c)

	assert.True(t, false) // force printing dump

	vf := &fetcher.VersionedFetcher{}
	desc := col.Description()
	err = vf.Init(&desc, nil, nil, false)
	assert.NoError(t, err)

	txn, err := db.NewTxn(false)
	assert.NoError(t, err)

	key := core.NewKey("bae-ed7f0bd5-3f5b-5e93-9310-4b2e71ac460d")
	version, err := cid.Decode("Qmcv2iU3myUBwuFCHe3w97sBMMER2FTY2rpbNBP6cqWb4S")
	assert.NoError(t, err)

	err = vf.Start(txn, key, version)
	assert.NoError(t, err)
}

func createDocUpdates(col *db.Collection) error {
	// col, err := newTestCollectionWithSchema(db)
	// if err != ni

	// dockey: bae-ed7f0bd5-3f5b-5e93-9310-4b2e71ac460d
	// cid: Qmcv2iU3myUBwuFCHe3w97sBMMER2FTY2rpbNBP6cqWb4S
	// sub:
	//   -age: QmSom35RYVzYTE7nGsudvomv1pi9ffjEfSFsPZgQRM92v1
	//	 -name: QmeKjH2iuNjbWqZ5Lx9hSCiZDeCQvb4tHNyGm99dvB69M9
	// 	 -points: Qmd7mvZJkL9uQoC2YZsQE3ijmyGAaHgSnZMvLY4H71Vmaz
	// 	 -verified: QmNRQwWjTBTDfAFUHkG8yuKmtbprYQtGs4jYxGJ5fCfXtn
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
	// cid: QmPgnQvhPuLGwVU4ZEcbRy7RNCxSkeS72eKwXusUrAEEXR
	// sub:
	// 	- name: QmZzL7AUq1L9whhHvVfbBJho6uAJQnAZWEFWYsTD2PgCKM
	//  - points: Qmejouu71QPjTue2P1gLnrzqApa8cU6NPdBoGrCQdpSC1Q
	doc.Set("name", "Pete")
	doc.Set("points", 99.9)
	if err := col.Update(doc); err != nil {
		return err
	}

	// update #2
	// cid: Qmf38odVhMXXE21La5a5dsS1bdrzqSTohY1j17oEmJQyJf
	// sub:
	// 	- verified: QmNTLb5ChDx3HjeAMuWVm7wmgjbXPzDRdPNnzwRqG71T2Q
	//  - age: QmfJTRSXy1x4VxaVDqSa35b3sXQkCAppPSwfhwKGkV2zez
	doc.Set("verified", false)
	doc.Set("age", 22)
	if err := col.Update(doc); err != nil {
		return err
	}

	// update #3
	// cid: QmZDYKSdBfkhZ8kjnXBmkgh8ad8ncy8tmHDZ6vUsvsMPEL
	// sub:
	// 	- points: QmQGkkF1xpLkMFWtG5fNTGs6VwbNXESrtG2Mj35epLU8do
	doc.Set("points", 129.99)
	err = col.Update(doc)

	fmt.Println(doc.ToMap())
	return err
}

func newTestCollectionWithSchema(d *db.DB) (*db.Collection, error) {
	desc := base.CollectionDescription{
		Name: "users",
		Schema: base.SchemaDescription{
			Fields: []base.FieldDescription{
				{
					Name: "_key",
					Kind: base.FieldKind_DocKey,
				},
				{
					Name: "name",
					Kind: base.FieldKind_STRING,
					Typ:  core.LWW_REGISTER,
				},
				{
					Name: "age",
					Kind: base.FieldKind_INT,
					Typ:  core.LWW_REGISTER,
				},
				{
					Name: "verified",
					Kind: base.FieldKind_BOOL,
					Typ:  core.LWW_REGISTER,
				},
				{
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
