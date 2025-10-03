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
	"context"

	"github.com/ipfs/boxo/blockstore"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime/storage"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corekv/namespace"
)

// Blockstore proxies the ipld.DAGService under the /core namespace for future-proofing
type Blockstore interface {
	blockstore.Blockstore
	AsIPLDStorage() IPLDStorage
	// Mark the block as merged by removing the to-merge index.
	MarkAsMerged(ctx context.Context, k cid.Cid) error
	// Check if the block has been merged. It will return false if either the CID is not found
	// or the CID is found AND the to-mege index is aslo found.
	IsMerged(ctx context.Context, k cid.Cid) (bool, error)
}

// IPLDStorage provides the methods needed for an IPLD LinkSystem.
type IPLDStorage interface {
	storage.ReadableStorage
	storage.WritableStorage
}

func prefix(root corekv.ReaderWriter, prefix []byte) corekv.ReaderWriter {
	return namespace.Wrap(root, prefix)
}
