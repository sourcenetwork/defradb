// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package net

import (
	"context"
	"sync"
	"time"

	"github.com/ipfs/boxo/blockservice"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/storage/bsrvadapter"

	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
)

// syncBlockLinkTimeout is the maximum amount of time
// to wait for a block link to be fetched.
var syncBlockLinkTimeout = 5 * time.Second

func makeLinkSystem(blockService blockservice.BlockService) linking.LinkSystem {
	blockStore := &bsrvadapter.Adapter{Wrapped: blockService}

	linkSys := cidlink.DefaultLinkSystem()
	linkSys.SetWriteStorage(blockStore)
	linkSys.SetReadStorage(blockStore)
	linkSys.TrustedStorage = true

	return linkSys
}

// syncDAG synchronizes the DAG starting with the given block
// using the blockservice to fetch remote blocks.
//
// This process walks the entire DAG until the issue below is resolved.
// https://github.com/sourcenetwork/defradb/issues/2722
func syncDAG(ctx context.Context, blockService blockservice.BlockService, block *coreblock.Block) error {
	// use a session to make remote fetches more efficient
	ctx = blockservice.ContextWithSession(ctx, blockService)

	linkSystem := makeLinkSystem(blockService)

	// Store the block in the DAG store
	_, err := linkSystem.Store(linking.LinkContext{Ctx: ctx}, coreblock.GetLinkPrototype(), block.GenerateNode())
	if err != nil {
		return err
	}

	err = loadBlockLinks(ctx, &linkSystem, block)
	if err != nil {
		return err
	}
	return nil
}

// loadBlockLinks loads the links of a block recursively.
//
// If it encounters errors in the concurrent loading of links, it will return
// the first error it encountered.
func loadBlockLinks(ctx context.Context, linkSys *linking.LinkSystem, block *coreblock.Block) error {
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()
	var wg sync.WaitGroup
	var asyncErr error
	var asyncErrOnce sync.Once

	// TODO: this part is not tested yet because there is not easy way of doing it at the moment.
	// https://github.com/sourcenetwork/defradb/issues/3525
	if block.Signature != nil {
		// we deliberately ignore the first returned value, which indicates whether the signature
		// the block was actually verified or not, because we don't handle it any different here.
		// But we want to keep the API of VerifyBlockSignature explicit about the results.
		_, err := coreblock.VerifyBlockSignature(block, linkSys)
		if err != nil {
			return err
		}
	}

	setAsyncErr := func(err error) {
		asyncErr = err
		cancel()
	}

	for _, lnk := range block.AllLinks() {
		wg.Add(1)
		go func(lnk cidlink.Link) {
			defer wg.Done()
			if ctxWithCancel.Err() != nil {
				return
			}
			ctxWithTimeout, cancel := context.WithTimeout(ctx, syncBlockLinkTimeout)
			defer cancel()
			nd, err := linkSys.Load(linking.LinkContext{Ctx: ctxWithTimeout}, lnk, coreblock.BlockSchemaPrototype)
			if err != nil {
				asyncErrOnce.Do(func() { setAsyncErr(err) })
				return
			}
			linkBlock, err := coreblock.GetFromNode(nd)
			if err != nil {
				asyncErrOnce.Do(func() { setAsyncErr(err) })
				return
			}

			err = loadBlockLinks(ctx, linkSys, linkBlock)
			if err != nil {
				asyncErrOnce.Do(func() { setAsyncErr(err) })
				return
			}
		}(lnk)
	}

	wg.Wait()

	return asyncErr
}
