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

// syncDAGTimeout is the maximum amount of time
// to wait for a dag to be fetched.
var syncDAGTimeout = 60 * time.Second

// syncDAG synchronizes the DAG starting with the given block
// using the blockservice to fetch remote blocks.
//
// This process walks the entire DAG until the issue below is resolved.
// https://github.com/sourcenetwork/defradb/issues/2722
func syncDAG(ctx context.Context, bserv blockservice.BlockService, block *coreblock.Block) error {
	// use a session to make remote fetches more efficient
	ctx = blockservice.ContextWithSession(ctx, bserv)
	store := &bsrvadapter.Adapter{Wrapped: bserv}

	lsys := cidlink.DefaultLinkSystem()
	lsys.SetWriteStorage(store)
	lsys.SetReadStorage(store)
	lsys.TrustedStorage = true

	// Store the block in the DAG store
	_, err := lsys.Store(linking.LinkContext{Ctx: ctx}, coreblock.GetLinkPrototype(), block.GenerateNode())
	if err != nil {
		return err
	}

	err = loadBlockLinks(ctx, lsys, block)
	if err != nil {
		return err
	}
	return nil
}

// loadBlockLinks loads the links of a block recursively.
//
// If it encounters errors in the concurrent loading of links, it will return
// the first error it encountered.
func loadBlockLinks(ctx context.Context, lsys linking.LinkSystem, block *coreblock.Block) error {
	ctx, cancel := context.WithTimeout(ctx, syncDAGTimeout)
	defer cancel()

	var wg sync.WaitGroup
	var asyncErr error
	var asyncErrOnce sync.Once

	setAsyncErr := func(err error) {
		asyncErr = err
		cancel()
	}

	for _, lnk := range block.AllLinks() {
		wg.Add(1)
		go func(lnk cidlink.Link) {
			defer wg.Done()
			if ctx.Err() != nil {
				return
			}
			nd, err := lsys.Load(linking.LinkContext{Ctx: ctx}, lnk, coreblock.SchemaPrototype)
			if err != nil {
				asyncErrOnce.Do(func() { setAsyncErr(err) })
				return
			}
			linkBlock, err := coreblock.GetFromNode(nd)
			if err != nil {
				asyncErrOnce.Do(func() { setAsyncErr(err) })
				return
			}
			err = loadBlockLinks(ctx, lsys, linkBlock)
			if err != nil {
				asyncErrOnce.Do(func() { setAsyncErr(err) })
				return
			}
		}(lnk)
	}

	wg.Wait()

	return asyncErr
}
