// Copyright 2020 Source Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
package core

import (
	ds "github.com/ipfs/go-datastore"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
)

// MultiStore is an interface wrapper around the 3 main types of stores needed for
// MerkleCRDTs
type MultiStore interface {
	Datastore() DSReaderWriter
	Headstore() DSReaderWriter
	DAGstore() DAGStore
}

// DSReaderWriter simplifies the interface that is exposed by a
// core.DSReaderWriter into its subcomponents Reader and Writer.
// Using this simplified interface means that both core.DSReaderWriter
// and ds.Txn satisfy the interface. Due to go-datastore#113 and
// go-datastore#114 ds.Txn no longer implements core.DSReaderWriter
// Which means we can't swap between the two for Datastores that
// support TxnDatastore.
type DSReaderWriter interface {
	ds.Read
	ds.Write
}

// DAGStore proxies the ipld.DAGService under the /core namespace for future-proofing
type DAGStore interface {
	blockstore.Blockstore
}

// func MultiStoreFrom(rootstore )
