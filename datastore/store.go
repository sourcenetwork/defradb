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

	"github.com/sourcenetwork/defradb/logging"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corekv/namespace"
)

var (
	log = logging.MustNewLogger("store")
)

// RootStore wraps Batching and TxnDatastore requiring datastore to support both batching and transactions.
type RootStore interface {
	// corekv.Batchable
	corekv.TxnStore
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

	// Peerstore is a wrapped root DSReaderWriter
	// as a ds.Batching, embedded into a DSBatching
	// under the /peers namespace
	Peerstore() DSReaderWriter

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
	corekv.Reader
	corekv.Writer
}

// DAGStore proxies the ipld.DAGService under the /core namespace for future-proofing
type DAGStore interface {
	blockstore.Blockstore
}

// DSBatching wraps the Batching interface from go-datastore
type DSBatching interface {
	corekv.Batchable
}

// shim to satisfy the namespace.Wrap method which
// expects a `corekv.Store` interface which includes
// a `Close()` method, but our DSReaderWriter doens't
// have it, so we add this stubbed type to shim it
type shimDSReadWriteCloser struct {
	DSReaderWriter
}

func (shimDSReadWriteCloser) Close() {} // noop

func prefix(root DSReaderWriter, prefix []byte) DSReaderWriter {
	shimCloser := shimDSReadWriteCloser{root}
	return namespace.Wrap(shimCloser, prefix)
}
