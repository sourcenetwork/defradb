// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package p2p

import (
	"context"
	"sync"

	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"

	"github.com/sourcenetwork/corekv/blockstore"
	"github.com/sourcenetwork/defradb/errors"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/encryption"
)

func makeLinkSystem(blockService blockstore.IPLDStore) linking.LinkSystem {
	linkSys := cidlink.DefaultLinkSystem()
	linkSys.SetWriteStorage(blockService)
	linkSys.SetReadStorage(blockService)
	linkSys.TrustedStorage = true

	return linkSys
}

// syncDAG synchronizes the DAG starting with the given block
// using the blockservice to fetch remote blocks.
//
// This process walks the entire DAG until the issue below is resolved.
// https://github.com/sourcenetwork/defradb/issues/2722
func (p *P2P) syncDAG(ctx context.Context, block *coreblock.Block) error {
	// use a session to make remote fetches more efficient
	ctx = p.host.ContextWithSession(ctx)

	linkSystem := makeLinkSystem(p.host.IPLDStore())

	// Store the block in the DAG store
	_, err := linkSystem.Store(linking.LinkContext{Ctx: ctx}, coreblock.GetLinkPrototype(), block.GenerateNode())
	if err != nil {
		return err
	}

	return p.loadBlockLinks(ctx, &linkSystem, block)
}

// loadBlockLinks loads the links of a block recursively.
//
// If it encounters errors in the concurrent loading of links, it will return
// the first error it encountered.
func (p *P2P) loadBlockLinks(ctx context.Context, linkSys *linking.LinkSystem, block *coreblock.Block) error {
	ctx, cancel := context.WithTimeout(ctx, p.syncBlockLinkTimeout)
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

	var encResults *encryption.Results
	if block.IsEncrypted() {
		results, err := p.kms.GetKeys(ctx, *block.Encryption)
		if err != nil {
			return err
		}
		encResults = results
	}

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
			nd, err := linkSys.Load(linking.LinkContext{Ctx: ctx}, lnk, coreblock.BlockSchemaPrototype)
			if err != nil {
				asyncErrOnce.Do(func() { setAsyncErr(err) })
				return
			}
			linkBlock, err := coreblock.GetFromNode(nd)
			if err != nil {
				asyncErrOnce.Do(func() { setAsyncErr(err) })
				return
			}
			err = p.loadBlockLinks(ctx, linkSys, linkBlock)
			if err != nil {
				asyncErrOnce.Do(func() { setAsyncErr(err) })
				return
			}
		}(lnk)
	}

	wg.Wait()

	if encResults != nil {
		for res := range encResults.Get() {
			asyncErr = errors.Join(asyncErr, res.Error)
		}
	}

	return asyncErr
}
