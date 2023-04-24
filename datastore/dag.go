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
	blockstore "github.com/ipfs/boxo/blockstore"
)

// DAGStore is the interface to the underlying BlockStore and BlockService.
type dagStore struct {
	blockstore.Blockstore // become a Blockstore
	store                 DSReaderWriter
	// bstore          blockstore.Blockstore
	// bserv           blockservice.BlockService
}

// NewDAGStore creates a new DAGStore with the supplied Batching datastore.
func NewDAGStore(store DSReaderWriter) DAGStore {
	dstore := &dagStore{
		Blockstore: NewBlockstore(store),
		store:      store,
	}

	return dstore
}

// func (d *dagStore) setupBlockstore() error {
// 	bs := blockstore.NewBlockstore(d.store)
// 	// bs = blockstore.NewIdStore(bs)
// 	// cachedbs, err := blockstore.CachedBlockstore(d.ctx, bs, blockstore.DefaultCacheOpts())
// 	// if err != nil {
// 	// 	return err
// 	// }
// 	d.bstore = bs
// 	return nil
// }

// func (d *dagStore) setupBlockService() error {
// 	// if d.cfg.Offline {
// 	// 	d.bserv = blockservice.New(d.bstore, offline.Exchange(p.bstore))
// 	// 	return nil
// 	// }

// 	// bswapnet := network.NewFromIpfsHost(p.host, p.dht)
// 	// bswap := bitswap.New(p.ctx, bswapnet, p.bstore)
// 	// p.bserv = blockservice.New(p.bstore, bswap)

// 	// @todo Investigate if we need an Exchanger or if it can stay as nil
// 	d.bserv = blockservice.New(d.bstore, offline.Exchange(d.bstore))
// 	return nil
// }

// func (d *dagStore) setupDAGService() error {
// 	d.DAGService = dag.NewDAGService(d.bserv)
// 	return nil
// }

// func (d *dagStore) Blockstore() blockstore.Blockstore {
// 	return d.bstore
// }
