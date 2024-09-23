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
	"github.com/ipfs/boxo/blockstore"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipld/go-ipld-prime/storage"

	"github.com/sourcenetwork/corelog"

	"github.com/sourcenetwork/defradb/datastore/iterable"
)

var (
	log = corelog.NewLogger("store")
)

// Rootstore wraps Batching and TxnDatastore requiring datastore to support both batching and transactions.
type Rootstore interface {
	ds.Batching
	ds.TxnDatastore
}

// MultiStore is an interface wrapper around the 3 main types of stores needed for MerkleCRDTs.
type MultiStore interface {
	Rootstore() DSReaderWriter

	// Datastore is a wrapped root DSReaderWriter under the /data namespace
	Datastore() DSReaderWriter

	// Encstore is a wrapped root DSReaderWriter under the /enc namespace
	// This store is used for storing symmetric encryption keys for doc encryption.
	// The store keys are comprised of docID + field name.
	Encstore() Blockstore

	// Headstore is a wrapped root DSReaderWriter under the /head namespace
	Headstore() DSReaderWriter

	// Peerstore is a wrapped root DSReaderWriter as a ds.Batching, embedded into a DSBatching
	// under the /peers namespace
	Peerstore() DSBatching

	// Blockstore is a wrapped root DSReaderWriter as a Blockstore, embedded into a Blockstore
	// under the /blocks namespace
	Blockstore() Blockstore

	// Headstore is a wrapped root DSReaderWriter under the /system namespace
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

// Blockstore proxies the ipld.DAGService under the /core namespace for future-proofing
type Blockstore interface {
	blockstore.Blockstore
	AsIPLDStorage() IPLDStorage
}

// IPLDStorage provides the methods needed for an IPLD LinkSystem.
type IPLDStorage interface {
	storage.ReadableStorage
	storage.WritableStorage
}

// DSBatching wraps the Batching interface from go-datastore
type DSBatching interface {
	ds.Batching
}
