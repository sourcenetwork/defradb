// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package kms

import (
	"context"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/sourcenetwork/defradb/datastore"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
)

type ipldEncStorage struct {
	encstore datastore.Blockstore
}

func newIPLDEncryptionStorage(encstore datastore.Blockstore) *ipldEncStorage {
	return &ipldEncStorage{encstore: encstore}
}

func (s *ipldEncStorage) get(ctx context.Context, cidBytes []byte) (*coreblock.Encryption, error) {
	lsys := cidlink.DefaultLinkSystem()
	lsys.SetReadStorage(s.encstore.AsIPLDStorage())

	_, blockCid, err := cid.CidFromBytes(cidBytes)
	if err != nil {
		return nil, err
	}

	nd, err := lsys.Load(linking.LinkContext{Ctx: ctx}, cidlink.Link{Cid: blockCid},
		coreblock.EncryptionSchemaPrototype)
	if err != nil {
		return nil, err
	}

	return coreblock.GetEncryptionBlockFromNode(nd)
}

func (s *ipldEncStorage) put(ctx context.Context, blockBytes []byte) ([]byte, error) {
	lsys := cidlink.DefaultLinkSystem()
	lsys.SetWriteStorage(s.encstore.AsIPLDStorage())

	var encBlock coreblock.Encryption
	err := encBlock.Unmarshal(blockBytes)
	if err != nil {
		return nil, err
	}

	link, err := lsys.Store(linking.LinkContext{Ctx: ctx}, coreblock.GetLinkPrototype(), encBlock.GenerateNode())
	if err != nil {
		return nil, err
	}

	return []byte(link.String()), nil
}
