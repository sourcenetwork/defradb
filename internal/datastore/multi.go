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

type Multistore struct {
	block  Blockstore
	data   corekv.ReaderWriter
	enc    Blockstore
	head   corekv.ReaderWriter
	peer   corekv.ReaderWriter
	root   corekv.ReaderWriter
	system corekv.ReaderWriter
}

func NewMultistore(rootstore corekv.ReaderWriter) *Multistore {
	return &Multistore{
		block:  newBlockstore(prefix(rootstore, blockStoreKey.Bytes())),
		data:   prefix(rootstore, dataStoreKey.Bytes()),
		enc:    newBlockstore(prefix(rootstore, encStoreKey.Bytes())),
		head:   prefix(rootstore, headStoreKey.Bytes()),
		peer:   prefix(rootstore, peerStoreKey.Bytes()),
		root:   rootstore,
		system: prefix(rootstore, systemStoreKey.Bytes()),
	}
}

func (m *Multistore) Blockstore() Blockstore {
	return m.block
}

func (m *Multistore) Datastore() corekv.ReaderWriter {
	return m.data
}

func (m *Multistore) Encstore() Blockstore {
	return m.enc
}

func (m *Multistore) Headstore() corekv.ReaderWriter {
	return m.head
}

func (m *Multistore) Peerstore() corekv.ReaderWriter {
	return m.peer
}

func (m *Multistore) Rootstore() corekv.ReaderWriter {
	return m.root
}

func (m *Multistore) Systemstore() corekv.ReaderWriter {
	return m.system
}

func DatastoreFrom(rootstore corekv.Store) corekv.ReaderWriter {
	return prefix(rootstore, dataStoreKey.Bytes())
}

func EncstoreFrom(rootstore corekv.Store) Blockstore {
	return newBlockstore(prefix(rootstore, encStoreKey.Bytes()))
}

func HeadstoreFrom(rootstore corekv.Store) corekv.ReaderWriter {
	return prefix(rootstore, headStoreKey.Bytes())
}

func BlockstoreFrom(rootstore corekv.Store) Blockstore {
	return newBlockstore(prefix(rootstore, blockStoreKey.Bytes()))
}

func SystemstoreFrom(rootstore corekv.Store) corekv.ReaderWriter {
	return prefix(rootstore, systemStoreKey.Bytes())
}

func PeerstoreFrom(rootstore corekv.Store) corekv.ReaderWriter {
	return prefix(rootstore, peerStoreKey.Bytes())
}
