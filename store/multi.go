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
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/db/base"
)

type multistore struct {
	root   client.DSReaderWriter
	data   client.DSReaderWriter
	head   client.DSReaderWriter
	system client.DSReaderWriter
	// block client.DSReaderWriter
	dag client.DAGStore
}

var _ client.MultiStore = (*multistore)(nil)

func MultiStoreFrom(rootstore client.DSReaderWriter) client.MultiStore {
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

// Datastore implements client.MultiStore
func (ms multistore) Datastore() client.DSReaderWriter {
	return ms.data
}

// Headstore implements client.MultiStore
func (ms multistore) Headstore() client.DSReaderWriter {
	return ms.head
}

// DAGstore implements client.MultiStore
func (ms multistore) DAGstore() client.DAGStore {
	return ms.dag
}

// Rootstore implements client.MultiStore
func (ms multistore) Rootstore() client.DSReaderWriter {
	return ms.root
}

func (ms multistore) Systemstore() client.DSReaderWriter {
	return ms.system
}
