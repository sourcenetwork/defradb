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
	"time"

	"github.com/ipfs/boxo/blockservice"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/linking/preload"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
	"github.com/ipld/go-ipld-prime/schema"
	"github.com/ipld/go-ipld-prime/storage/bsrvadapter"
	"github.com/ipld/go-ipld-prime/traversal"
	"github.com/ipld/go-ipld-prime/traversal/selector"
	"github.com/ipld/go-ipld-prime/traversal/selector/builder"

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
	ctx, cancel := context.WithTimeout(ctx, syncDAGTimeout)
	defer cancel()

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

	ssb := builder.NewSelectorSpecBuilder(basicnode.Prototype.Any)
	matchAllSelector, err := ssb.ExploreRecursive(selector.RecursionLimitNone(), ssb.ExploreUnion(
		ssb.Matcher(),
		ssb.ExploreAll(ssb.ExploreRecursiveEdge()),
	)).Selector()
	if err != nil {
		return err
	}

	// prototypeChooser returns the node prototype to use when traversing
	prototypeChooser := func(lnk ipld.Link, lnkCtx ipld.LinkContext) (ipld.NodePrototype, error) {
		if tlnkNd, ok := lnkCtx.LinkNode.(schema.TypedLinkNode); ok {
			return tlnkNd.LinkTargetNodePrototype(), nil
		}
		return basicnode.Prototype.Any, nil
	}
	// preloader is used to asynchronously load blocks before traversing
	//
	// any errors encountered during preload are ignored
	preloader := func(pctx preload.PreloadContext, l preload.Link) {
		go lsys.Load(linking.LinkContext{Ctx: pctx.Ctx}, l.Link, coreblock.SchemaPrototype) //nolint:errcheck
	}
	config := traversal.Config{
		Ctx:                            ctx,
		LinkSystem:                     lsys,
		LinkVisitOnlyOnce:              true,
		LinkTargetNodePrototypeChooser: prototypeChooser,
		Preloader:                      preloader,
	}
	visit := func(p traversal.Progress, n datamodel.Node) error {
		return nil
	}
	return traversal.Progress{Cfg: &config}.WalkMatching(block.GenerateNode(), matchAllSelector, visit)
}
