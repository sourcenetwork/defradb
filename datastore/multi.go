// Copyright 2022 Democratized Data Foundation
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
	ds "github.com/ipfs/go-datastore"
)

var (
	// Individual Store Keys
	rootStoreKey   = ds.NewKey("db")
	systemStoreKey = rootStoreKey.ChildString("system")
	dataStoreKey   = rootStoreKey.ChildString("data")
	headStoreKey   = rootStoreKey.ChildString("heads")
	blockStoreKey  = rootStoreKey.ChildString("blocks")
	peerStoreKey   = rootStoreKey.ChildString("ps")
)

type multistore struct {
	root   DSReaderWriter
	data   DSReaderWriter
	head   DSReaderWriter
	peer   DSReaderWriter
	system DSReaderWriter
	// block DSReaderWriter
	dag DAGStore
}

var _ MultiStore = (*multistore)(nil)

// MultiStoreFrom creates a MultiStore from a root datastore.
func MultiStoreFrom(rootstore DSReaderWriter) MultiStore {
	ms := &multistore{
		root:   rootstore,
		data:   prefix(rootstore, dataStoreKey.Bytes()),
		head:   prefix(rootstore, headStoreKey.Bytes()),
		peer:   prefix(rootstore, peerStoreKey.Bytes()),
		system: prefix(rootstore, systemStoreKey.Bytes()),
		dag:    NewDAGStore(prefix(rootstore, blockStoreKey.Bytes())),
	}

	return ms
}

// Datastore implements MultiStore.
func (ms multistore) Datastore() DSReaderWriter {
	return ms.data
}

// Headstore implements MultiStore.
func (ms multistore) Headstore() DSReaderWriter {
	return ms.head
}

// Peerstore implements MultiStore.
func (ms multistore) Peerstore() DSReaderWriter {
	return ms.peer
}

// DAGstore implements MultiStore.
func (ms multistore) DAGstore() DAGStore {
	return ms.dag
}

// Rootstore implements MultiStore.
func (ms multistore) Rootstore() DSReaderWriter {
	return ms.root
}

// Systemstore implements MultiStore.
func (ms multistore) Systemstore() DSReaderWriter {
	return ms.system
}
