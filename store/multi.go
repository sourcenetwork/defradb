// Copyright 2021 Source Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
package store

import (
	"fmt"

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/base"

	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/namespace"
	dsq "github.com/ipfs/go-datastore/query"
)

type multistore struct {
	root core.DSReaderWriter
	data core.DSReaderWriter
	head core.DSReaderWriter
	// block core.DSReaderWriter
	dag core.DAGStore
}

func MultiStoreFrom(rootstore ds.Datastore) core.MultiStore {
	ms := &multistore{root: rootstore}
	ms.data = namespace.Wrap(rootstore, base.DataStoreKey)
	ms.head = namespace.Wrap(rootstore, base.HeadStoreKey)
	block := namespace.Wrap(rootstore, base.BlockStoreKey)
	ms.dag = NewDAGStore(block)

	return ms
}

// Datastore implements core.MultiStore
func (ms multistore) Datastore() core.DSReaderWriter {
	return ms.data
}

// Headstore implements core.MultiStore
func (ms multistore) Headstore() core.DSReaderWriter {
	return ms.head
}

// DAGstore implements core.MultiStore
func (ms multistore) DAGstore() core.DAGStore {
	return ms.dag
}

// Rootstore implements core.MultiStore
func (ms multistore) Rootstore() core.DSReaderWriter {
	return ms.root
}

func PrintStore(store core.DSReaderWriter) {
	q := dsq.Query{
		Prefix:   "",
		KeysOnly: false,
		Orders:   []dsq.Order{dsq.OrderByKey{}},
	}

	results, err := store.Query(q)
	defer results.Close()
	if err != nil {
		panic(err)
	}

	for r := range results.Next() {
		fmt.Println(r.Key, ": ", r.Value)
	}
}
