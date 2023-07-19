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
	blockstore "github.com/ipfs/boxo/blockstore"
	ds "github.com/ipfs/go-datastore"

	"github.com/sourcenetwork/defradb/datastore/iterable"
	"github.com/sourcenetwork/defradb/logging"
)

var (
	log = logging.MustNewLogger("store")
)

// RootStore wraps Batching and TxnDatastore requiring datastore to support both batching and transactions.
type RootStore interface {
	ds.Batching
	ds.TxnDatastore
}

// MultiStore is an interface wrapper around the 3 main types of stores needed for MerkleCRDTs.
type MultiStore interface {
	Rootstore() DSReaderWriter

	// Datastore is a wrapped root DSReaderWriter
	// under the /data namespace
	Datastore() DSReaderWriter

	// Headstore is a wrapped root DSReaderWriter
	// under the /head namespace
	Headstore() DSReaderWriter

	// DAGstore is a wrapped root DSReaderWriter
	// as a Blockstore, embedded into a DAGStore
	// under the /blocks namespace
	DAGstore() DAGStore

	// Headstore is a wrapped root DSReaderWriter
	// under the /system namespace
	Systemstore() DSReaderWriter
}

// DSReaderWriter simplifies the interface that is exposed by a
// DSReaderWriter into its sub-components Reader and Writer.
// Using this simplified interface means that both DSReaderWriter
// and ds.Txn satisfy the interface. Due to go-datastore#113 and
// go-datastore#114 ds.Txn no longer implements DSReaderWriter
// Which means we can't swap between the two for Datastores that
// support TxnDatastore.
type DSReaderWriter interface {
	ds.Read
	ds.Write
	iterable.Iterable
}

// DAGStore proxies the ipld.DAGService under the /core namespace for future-proofing
type DAGStore interface {
	blockstore.Blockstore
}
