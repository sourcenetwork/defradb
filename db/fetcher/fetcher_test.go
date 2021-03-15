// Copyright 2020 Source Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
package fetcher

import (
	"testing"

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/base"
)

// func newMemoryDB() (*db.DB, error) {
// 	opts := &db.Options{
// 		Store: "memory",
// 		Memory: db.MemoryOptions{
// 			Size: 1024 * 1000,
// 		},
// 	}

// 	return db.NewDB(opts)
// }

// Create a new Fetcher for a Collection named "users"
// with the following schema:
// Users {
//		Name string
//		Age int
// }
func newFetcher() *docFetcher {
	df := new(docFetcher)
	indexDesc := &base.IndexDescription{
		Name:    "users.index.primary",
		ID:      uint32(0),
		Primary: true,
		Unique:  true,
	}
	collectionDesc := &base.CollectionDescription{
		Name: "users",
		ID:   uint32(1),
		Schema: base.SchemaDescription{
			ID:       uint32(1),
			FieldIDs: []uint32{1, 2, 3},
			Fields: []base.FieldDescription{
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
		Indexes: []base.IndexDescription{*indexDesc},
	}
	err := df.Init(collectionDesc, indexDesc, nil, false)
	if err != nil {
		panic(err)
	}
	return df
}

func TestFetcherInit(t *testing.T) {
	df := newFetcher()
	if df.col == nil {
		t.Error("docFetcher cannot be started without a ColletionDescription")
	}
	if df.index == nil {
		t.Error("docFetcher cannot be started without a IndexDescription")
	}
}

func TestFetcherStart(t *testing.T) {
	db, err := newMemoryDB()
	if err != nil {
		t.Error(err)
		return
	}
	txn, err := db.NewTxn(true)
	if err != nil {
		t.Error(err)
		return
	}
	df := newFetcher()
	err = df.Start(txn, core.Spans{})
	if err != nil {
		t.Error(err)
	}
}
