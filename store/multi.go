// Copyright 2022 Democratized Data Foundation.
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
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/base"
)

type multistore struct {
	root   core.DSReaderWriter
	data   core.DSReaderWriter
	head   core.DSReaderWriter
	system core.DSReaderWriter
	// block core.DSReaderWriter
	dag core.DAGStore
}

var _ core.MultiStore = (*multistore)(nil)

func MultiStoreFrom(rootstore core.DSReaderWriter) core.MultiStore {
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

func (ms multistore) Systemstore() core.DSReaderWriter {
	return ms.system
}
