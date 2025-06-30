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

func prefix(root corekv.ReaderWriter, prefix []byte) corekv.ReaderWriter {
	return namespace.Wrap(root, prefix)
}
