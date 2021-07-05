package fetcher_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/db/fetcher"
	"github.com/sourcenetwork/defradb/document"

	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/assert"
)

type update struct {
	payload []byte
	diffOps map[string]interface{}
	cid     string
}

var (
	testStates = []update{
		{
			payload: []byte(`{
				"name": "Alice",
				"age": 31,
				"points": 100,
				"verified": true
			}`),
			cid: "Qmcv2iU3myUBwuFCHe3w97sBMMER2FTY2rpbNBP6cqWb4S",
		},
		{
			payload: []byte(`{
				"name": "Pete",
				"age": 31,
				"points": 99.9,
				"verified": true
			}`),
			diffOps: map[string]interface{}{
				"name":   "Pete",
				"points": 99.9,
			},
			cid: "QmPgnQvhPuLGwVU4ZEcbRy7RNCxSkeS72eKwXusUrAEEXR",
		},
		{
			payload: []byte(`{
				"name": "Pete",
				"age": 22,
				"points": 99.9,
				"verified": false
			}`),
			diffOps: map[string]interface{}{
				"verified": false,
				"age":      22,
			},
			cid: "QmRpMfTzExGrXat5W9uCAEtnSpRTvWBcd1hBYNWVPdN9Xh",
		},
		{
			payload: []byte(`{
				"name": "Pete",
				"age": 22,
				"points": 129.99,
				"verified": false
			}`),
			diffOps: map[string]interface{}{
				"points": 129.99,
			},
			cid: "QmRWYwKadjWqHLrzPKd7MdS4EoQuT2RzWVTaBxxVkeSjFH",
		},
	}
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

	// db.PrintDump()
	// assert.True(t, false) // force printing dump

	vf := &fetcher.VersionedFetcher{}
	desc := col.Description()
	err = vf.Init(&desc, nil, nil, false)
	assert.NoError(t, err)

	txn, err := db.NewTxn(false)
	assert.NoError(t, err)

	key := core.NewKey("bae-ed7f0bd5-3f5b-5e93-9310-4b2e71ac460d")
	version, err := cid.Decode("QmRWYwKadjWqHLrzPKd7MdS4EoQuT2RzWVTaBxxVkeSjFH")
	assert.NoError(t, err)

	span := fetcher.NewVersionedSpan(key, version)
	err = vf.Start(txn, span)
	assert.NoError(t, err)

	// err = vf.SeekTo(version)
	// assert.NoError(t, err)

	// store.PrintStore(vf.Rootstore())
	// assert.True(t, false)
}

func TestVersionedFetcherNextMap(t *testing.T) {
	db, err := newMemoryDB()
	assert.NoError(t, err)

	col, err := newTestCollectionWithSchema(db)
	assert.NoError(t, err)

	err = createDocUpdates(col)
	assert.NoError(t, err)

	// assert.True(t, false) // force printing dump

	vf := &fetcher.VersionedFetcher{}
	desc := col.Description()
	err = vf.Init(&desc, nil, nil, false)
	assert.NoError(t, err)

	txn, err := db.NewTxn(false)
	assert.NoError(t, err)

	key := core.NewKey("bae-ed7f0bd5-3f5b-5e93-9310-4b2e71ac460d")
	version, err := cid.Decode(testStates[0].cid)
	assert.NoError(t, err)

	span := fetcher.NewVersionedSpan(key, version)
	err = vf.Start(txn, span)
	assert.NoError(t, err)

	_, doc, err := vf.FetchNextMap()
	assert.NoError(t, err)

	var state map[string]interface{}
	err = json.Unmarshal(testStates[0].payload, &state)
	assert.NoError(t, err)

	compareVersionedDocs(t, doc, state)

	// fmt.Println(doc)
	// assert.True(t, false)

}

func TestVersionedFetcherNextMapV1(t *testing.T) {
	db, err := newMemoryDB()
	assert.NoError(t, err)

	col, err := newTestCollectionWithSchema(db)
	assert.NoError(t, err)

	err = createDocUpdates(col)
	assert.NoError(t, err)

	// assert.True(t, false) // force printing dump

	vf := &fetcher.VersionedFetcher{}
	desc := col.Description()
	err = vf.Init(&desc, nil, nil, false)
	assert.NoError(t, err)

	txn, err := db.NewTxn(false)
	assert.NoError(t, err)

	key := core.NewKey("bae-ed7f0bd5-3f5b-5e93-9310-4b2e71ac460d")
	version, err := cid.Decode("QmPgnQvhPuLGwVU4ZEcbRy7RNCxSkeS72eKwXusUrAEEXR")
	assert.NoError(t, err)

	span := fetcher.NewVersionedSpan(key, version)
	err = vf.Start(txn, span)
	assert.NoError(t, err)

	_, doc, err := vf.FetchNextMap()
	assert.NoError(t, err)

	var state map[string]interface{}
	err = json.Unmarshal(testStates[1].payload, &state)
	assert.NoError(t, err)

	compareVersionedDocs(t, doc, state)

	// fmt.Println(doc)
	// assert.True(t, false)

}

func TestVersionedFetcherNextMapV2(t *testing.T) {
	db, err := newMemoryDB()
	assert.NoError(t, err)

	col, err := newTestCollectionWithSchema(db)
	assert.NoError(t, err)

	err = createDocUpdates(col)
	assert.NoError(t, err)

	// assert.True(t, false) // force printing dump

	vf := &fetcher.VersionedFetcher{}
	desc := col.Description()
	err = vf.Init(&desc, nil, nil, false)
	assert.NoError(t, err)

	txn, err := db.NewTxn(false)
	assert.NoError(t, err)

	key := core.NewKey("bae-ed7f0bd5-3f5b-5e93-9310-4b2e71ac460d")
	version, err := cid.Decode("QmRpMfTzExGrXat5W9uCAEtnSpRTvWBcd1hBYNWVPdN9Xh")
	assert.NoError(t, err)

	span := fetcher.NewVersionedSpan(key, version)
	err = vf.Start(txn, span)
	assert.NoError(t, err)

	_, doc, err := vf.FetchNextMap()
	assert.NoError(t, err)

	var state map[string]interface{}
	err = json.Unmarshal(testStates[2].payload, &state)
	assert.NoError(t, err)

	compareVersionedDocs(t, doc, state)

	// fmt.Println(doc)
	// assert.True(t, false)

}

func TestVersionedFetcherNextMapV3(t *testing.T) {
	db, err := newMemoryDB()
	assert.NoError(t, err)

	col, err := newTestCollectionWithSchema(db)
	assert.NoError(t, err)

	err = createDocUpdates(col)
	assert.NoError(t, err)

	// assert.True(t, false) // force printing dump

	vf := &fetcher.VersionedFetcher{}
	desc := col.Description()
	err = vf.Init(&desc, nil, nil, false)
	assert.NoError(t, err)

	txn, err := db.NewTxn(false)
	assert.NoError(t, err)

	key := core.NewKey("bae-ed7f0bd5-3f5b-5e93-9310-4b2e71ac460d")
	version, err := cid.Decode("QmRWYwKadjWqHLrzPKd7MdS4EoQuT2RzWVTaBxxVkeSjFH")
	assert.NoError(t, err)

	span := fetcher.NewVersionedSpan(key, version)
	err = vf.Start(txn, span)
	assert.NoError(t, err)

	_, doc, err := vf.FetchNextMap()
	assert.NoError(t, err)

	var state map[string]interface{}
	err = json.Unmarshal(testStates[3].payload, &state)
	assert.NoError(t, err)

	compareVersionedDocs(t, doc, state)

	// fmt.Println(doc)
	// assert.True(t, false)
}

func TestVersionedFetcherIncrementalSeekTo(t *testing.T) {
	db, err := newMemoryDB()
	assert.NoError(t, err)

	col, err := newTestCollectionWithSchema(db)
	assert.NoError(t, err)

	err = createDocUpdates(col)
	assert.NoError(t, err)

	// assert.True(t, false) // force printing dump

	vf := &fetcher.VersionedFetcher{}
	desc := col.Description()
	err = vf.Init(&desc, nil, nil, false)
	assert.NoError(t, err)

	txn, err := db.NewTxn(false)
	assert.NoError(t, err)

	key := core.NewKey("bae-ed7f0bd5-3f5b-5e93-9310-4b2e71ac460d")
	version, err := cid.Decode(testStates[0].cid)
	assert.NoError(t, err)

	span := fetcher.NewVersionedSpan(key, version)
	err = vf.Start(txn, span)
	assert.NoError(t, err)

	// loop over updates so we can seek to them
	// skip first (create)
	for _, update := range testStates[1:] {
		c, err := cid.Decode(update.cid)
		assert.NoError(t, err)

		err = vf.SeekTo(c)
		assert.NoError(t, err)

		_, doc, err := vf.FetchNextMap()
		assert.NoError(t, err)

		var state map[string]interface{}
		err = json.Unmarshal(update.payload, &state)
		assert.NoError(t, err)

		compareVersionedDocs(t, doc, state)
	}
}

// func buildTestState() (*db.DB, *db.Collection, error) {

// }

func compareVersionedDocs(t *testing.T, doc, expected map[string]interface{}) {
	for k, v := range doc {
		if k == "_key" {
			continue
		}
		// make sure our floats are converted
		if f, ok := expected[k].(float64); ok {
			if f == float64(int64(f)) {
				expected[k] = int64(f)
			}
		}

		if i, ok := v.(uint64); ok {
			if i == uint64(int64(i)) {
				v = int64(i)
			}
		}
		assert.Equal(t, expected[k], v)
	}
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
	// testJSONObj := []byte(`{
	// 	"name": "Alice",
	// 	"age": 31,
	// 	"points": 100,
	// 	"verified": true
	// }`)

	// doc, err := document.NewFromJSON(testJSONObj)
	// if err != nil {
	// 	return err
	// }

	// if err := col.Save(doc); err != nil {
	// 	return err
	// }

	// // update #1
	// // cid: QmPgnQvhPuLGwVU4ZEcbRy7RNCxSkeS72eKwXusUrAEEXR
	// // sub:
	// // 	- name: QmZzL7AUq1L9whhHvVfbBJho6uAJQnAZWEFWYsTD2PgCKM
	// //  - points: Qmejouu71QPjTue2P1gLnrzqApa8cU6NPdBoGrCQdpSC1Q
	// doc.Set("name", "Pete")
	// doc.Set("points", 99.9)
	// if err := col.Update(doc); err != nil {
	// 	return err
	// }

	// // update #2
	// // cid: QmRpMfTzExGrXat5W9uCAEtnSpRTvWBcd1hBYNWVPdN9Xh
	// // sub:
	// // 	- verified: QmNTLb5ChDx3HjeAMuWVm7wmgjbXPzDRdPNnzwRqG71T2Q
	// //  - age: QmfJTRSXy1x4VxaVDqSa35b3sXQkCAppPSwfhwKGkV2zez
	// doc.Set("verified", false)
	// doc.Set("age", 22)
	// if err := col.Update(doc); err != nil {
	// 	return err
	// }

	// // update #3
	// // cid: QmRWYwKadjWqHLrzPKd7MdS4EoQuT2RzWVTaBxxVkeSjFH
	// // sub:
	// // 	- points: QmQGkkF1xpLkMFWtG5fNTGs6VwbNXESrtG2Mj35epLU8do
	// doc.Set("points", 129.99)
	// err = col.Update(doc)

	var doc *document.Document
	var err error
	for i, update := range testStates {
		if i == 0 { // create
			doc, err = document.NewFromJSON(update.payload)
			if err != nil {
				return err
			}
			if err := col.Save(doc); err != nil {
				return err
			}
		} else {
			if update.diffOps == nil {
				return errors.New("Expecting diffOp for update")
			}

			for k, v := range update.diffOps {
				doc.Set(k, v)
			}
			err = col.Update(doc)
			if err != nil {
				return err
			}
		}
	}

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
