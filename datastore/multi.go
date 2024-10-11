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
	encStoreKey    = rootStoreKey.ChildString("enc")
)

type multistore struct {
	root   DSReaderWriter
	data   DSReaderWriter
	enc    Blockstore
	head   DSReaderWriter
	peer   DSReaderWriter
	system DSReaderWriter
	dag    Blockstore
}

var _ MultiStore = (*multistore)(nil)

// MultiStoreFrom creates a MultiStore from a root datastore.
func MultiStoreFrom(rootstore ds.Datastore) MultiStore {
	rootRW := AsDSReaderWriter(rootstore)
	ms := &multistore{
		root:   rootRW,
		data:   prefix(rootRW, dataStoreKey),
		enc:    newBlockstore(prefix(rootRW, encStoreKey)),
		head:   prefix(rootRW, headStoreKey),
		peer:   prefix(rootRW, peerStoreKey),
		system: prefix(rootRW, systemStoreKey),
		dag:    newBlockstore(prefix(rootRW, blockStoreKey)),
	}

	return ms
}

// Datastore implements MultiStore.
func (ms multistore) Datastore() DSReaderWriter {
	return ms.data
}

// Encstore implements MultiStore.
func (ms multistore) Encstore() Blockstore {
	return ms.enc
}

// Headstore implements MultiStore.
func (ms multistore) Headstore() DSReaderWriter {
	return ms.head
}

// Peerstore implements MultiStore.
func (ms multistore) Peerstore() DSReaderWriter {
	return ms.peer
}

// Blockstore implements MultiStore.
func (ms multistore) Blockstore() Blockstore {
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
