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
	"github.com/ipld/go-ipld-prime/storage"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corekv/namespace"
	"github.com/sourcenetwork/corelog"
)

var (
	log = corelog.NewLogger("store")
)

// ReaderWriter simplifies the interface that is exposed by a
// ReaderWriter into its sub-components Reader and Writer.
// Using this simplified interface means that both ReaderWriter
// and ds.Txn satisfy the interface. Due to go-datastore#113 and
// go-datastore#114 ds.Txn no longer implements ReaderWriter
// Which means we can't swap between the two for Datastores that
// support TxnDatastore.
type ReaderWriter interface {
	corekv.Reader
	corekv.Writer
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

func prefix(root corekv.Store, prefix []byte) corekv.Store {
	return namespace.Wrap(root, prefix)
}
