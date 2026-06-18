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
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"

	"github.com/sourcenetwork/corekv/blockstore"

	"github.com/sourcenetwork/defradb/errors"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/datastore"
)

type ipldEncStorage struct {
	encstore datastore.Blockstore
}

func newIPLDEncryptionStorage(encstore datastore.Blockstore) *ipldEncStorage {
	return &ipldEncStorage{encstore: encstore}
}

func (s *ipldEncStorage) get(ctx context.Context, cidBytes []byte) (*coreblock.Encryption, error) {
	lsys := cidlink.DefaultLinkSystem()
	lsys.SetReadStorage(blockstore.NewIPLDStore(s.encstore))

	_, blockCid, err := cid.CidFromBytes(cidBytes)
	if err != nil {
		return nil, err
	}

	nd, err := lsys.Load(linking.LinkContext{Ctx: ctx}, cidlink.Link{Cid: blockCid},
		coreblock.EncryptionSchemaPrototype)
	if err != nil {
		// Not-found is non-fatal: the caller's `if encBlock == nil` branch
		// handles it, and a peer without the key must reply empty so the
		// multi-peer fan-out can converge on a holder.
		if errors.Is(err, ipld.ErrNotFound{}) {
			return nil, nil
		}
		return nil, err
	}

	return coreblock.GetEncryptionBlockFromNode(nd)
}

func (s *ipldEncStorage) put(ctx context.Context, blockBytes []byte) ([]byte, error) {
	lsys := cidlink.DefaultLinkSystem()
	lsys.SetWriteStorage(blockstore.NewIPLDStore(s.encstore))

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

func (s *ipldEncStorage) putBlock(ctx context.Context, block coreblock.Encryption) ([]byte, error) {
	lsys := cidlink.DefaultLinkSystem()
	lsys.SetWriteStorage(blockstore.NewIPLDStore(s.encstore))

	link, err := lsys.Store(linking.LinkContext{Ctx: ctx}, coreblock.GetLinkPrototype(), block.GenerateNode())
	if err != nil {
		return nil, err
	}

	return []byte(link.String()), nil
}

func (s *ipldEncStorage) computeLink(ctx context.Context, blockBytes []byte) ([]byte, error) {
	lsys := cidlink.DefaultLinkSystem()
	lsys.SetWriteStorage(blockstore.NewIPLDStore(s.encstore))

	var encBlock coreblock.Encryption
	err := encBlock.Unmarshal(blockBytes)
	if err != nil {
		return nil, err
	}

	link, err := lsys.ComputeLink(coreblock.GetLinkPrototype(), encBlock.GenerateNode())
	if err != nil {
		return nil, err
	}

	return []byte(link.String()), nil
}

func (s *ipldEncStorage) computeBlockLink(ctx context.Context, block coreblock.Encryption) ([]byte, error) {
	lsys := cidlink.DefaultLinkSystem()
	lsys.SetWriteStorage(blockstore.NewIPLDStore(s.encstore))

	link, err := lsys.ComputeLink(coreblock.GetLinkPrototype(), block.GenerateNode())
	if err != nil {
		return nil, err
	}

	return []byte(link.String()), nil
}
