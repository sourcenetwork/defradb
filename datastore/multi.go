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

func DatastoreFrom(rootstore corekv.Store) ReaderWriter {
	return prefix(rootstore, dataStoreKey.Bytes())
}

func EncstoreFrom(rootstore corekv.Store) Blockstore {
	return newBlockstore(prefix(rootstore, encStoreKey.Bytes()))
}

func HeadstoreFrom(rootstore corekv.Store) ReaderWriter {
	return prefix(rootstore, headStoreKey.Bytes())
}

func BlockstoreFrom(rootstore corekv.Store) Blockstore {
	return newBlockstore(prefix(rootstore, blockStoreKey.Bytes()))
}

func SystemstoreFrom(rootstore corekv.Store) ReaderWriter {
	return prefix(rootstore, systemStoreKey.Bytes())
}

func PeerstoreFrom(rootstore corekv.Store) ReaderWriter {
	return prefix(rootstore, peerStoreKey.Bytes())
}
