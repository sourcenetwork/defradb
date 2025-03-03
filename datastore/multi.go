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
	"github.com/sourcenetwork/corekv"
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
func MultiStoreFrom(rootstore corekv.Store) MultiStore {
	ms := &multistore{
		root:   rootstore,
		data:   prefix(rootstore, dataStoreKey.Bytes()),
		enc:    newBlockstore(prefix(rootstore, encStoreKey.Bytes())),
		head:   prefix(rootstore, headStoreKey.Bytes()),
		peer:   prefix(rootstore, peerStoreKey.Bytes()),
		system: prefix(rootstore, systemStoreKey.Bytes()),
		dag:    newBlockstore(prefix(rootstore, blockStoreKey.Bytes())),
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
