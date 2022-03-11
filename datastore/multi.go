// Copyright 2022 Democratized Data Foundation.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package datastore

import (
	"github.com/sourcenetwork/defradb/db/base"
)

type multistore struct {
	root   DSReaderWriter
	data   DSReaderWriter
	head   DSReaderWriter
	system DSReaderWriter
	// block DSReaderWriter
	dag DAGStore
}

var _ MultiStore = (*multistore)(nil)

func MultiStoreFrom(rootstore DSReaderWriter) MultiStore {
	block := prefix(rootstore, base.BlockStoreKey)
	ms := &multistore{
		root:   rootstore,
		data:   prefix(rootstore, base.DataStoreKey),
		head:   prefix(rootstore, base.HeadStoreKey),
		system: prefix(rootstore, base.SystemStoreKey),
		dag:    NewDAGStore(block),
	}

	return ms
}

// Datastore implements MultiStore
func (ms multistore) Datastore() DSReaderWriter {
	return ms.data
}

// Headstore implements MultiStore
func (ms multistore) Headstore() DSReaderWriter {
	return ms.head
}

// DAGstore implements MultiStore
func (ms multistore) DAGstore() DAGStore {
	return ms.dag
}

// Rootstore implements MultiStore
func (ms multistore) Rootstore() DSReaderWriter {
	return ms.root
}

func (ms multistore) Systemstore() DSReaderWriter {
	return ms.system
}
