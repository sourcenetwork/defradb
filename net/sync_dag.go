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
	"fmt"
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
func syncDAG(ctx context.Context, blockService, sigBlockService blockservice.BlockService, block *coreblock.Block) error {
	// use a session to make remote fetches more efficient
	ctx = blockservice.ContextWithSession(ctx, blockService)

	linkSys := makeLinkSystem(blockService)
	sigLinkSys := makeLinkSystem(sigBlockService)

	// Store the block in the DAG store
	_, err := linkSys.Store(linking.LinkContext{Ctx: ctx}, coreblock.GetLinkPrototype(), block.GenerateNode())
	if err != nil {
		return err
	}

	err = loadBlockLinks(ctx, &linkSys, &sigLinkSys, block)
	if err != nil {
		return err
	}
	return nil
}

// loadBlockLinks loads the links of a block recursively.
//
// If it encounters errors in the concurrent loading of links, it will return
// the first error it encountered.
func loadBlockLinks(ctx context.Context, linkSys, sigLinkSys *linking.LinkSystem, block *coreblock.Block) error {
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()
	var wg sync.WaitGroup
	var asyncErr error
	var asyncErrOnce sync.Once

	if block.Signature != nil {
		fmt.Printf(">>>>> loadBlockLinks: Verifying block signature %s\n", block.Signature)
		err := coreblock.VerifyBlockSignature(block, sigLinkSys)
		fmt.Printf(">>>>> loadBlockLinks: Block signature verified\n")
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
		fmt.Printf(">>>>> loadBlockLinks: Loading block link: %s\n", lnk)
		go func(lnk cidlink.Link) {
			fmt.Printf(">>>>> go loadBlockLinks: start %s\n", lnk)
			defer wg.Done()
			if ctxWithCancel.Err() != nil {
				return
			}
			ctxWithTimeout, cancel := context.WithTimeout(ctx, syncBlockLinkTimeout)
			defer cancel()
			fmt.Printf(">>>>> go loadBlockLinks: Loading block link: %s\n", lnk)
			nd, err := linkSys.Load(linking.LinkContext{Ctx: ctxWithTimeout}, lnk, coreblock.BlockSchemaPrototype)
			if err != nil {
				fmt.Printf(">>>>> go loadBlockLinks: Error loading block link: %s, error: %s\n", lnk, err)
				asyncErrOnce.Do(func() { setAsyncErr(err) })
				return
			}
			fmt.Printf(">>>>> go loadBlockLinks: Deserialize block link: %s\n", lnk)
			linkBlock, err := coreblock.GetFromNode(nd)
			if err != nil {
				asyncErrOnce.Do(func() { setAsyncErr(err) })
				return
			}

			fmt.Printf(">>>>> go loadBlockLinks: process parsed block: %s\n", lnk)
			err = loadBlockLinks(ctx, linkSys, sigLinkSys, linkBlock)
			if err != nil {
				asyncErrOnce.Do(func() { setAsyncErr(err) })
				return
			}
			fmt.Printf(">>>>> go loadBlockLinks: successfully processed %s\n", lnk)
		}(lnk)
	}

	wg.Wait()

	return asyncErr
}
