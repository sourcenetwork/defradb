package store

import (
	"context"

	blockservice "github.com/ipfs/go-blockservice"
	ds "github.com/ipfs/go-datastore"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	offline "github.com/ipfs/go-ipfs-exchange-offline"
	ipld "github.com/ipfs/go-ipld-format"
	dag "github.com/ipfs/go-merkledag"
)

// DAGStore is the interface to the underlying BlockStore and BlockService
type DAGStore struct {
	ipld.DAGService // become a DAG service
	ctx             context.Context
	store           ds.Batching
	bstore          blockstore.Blockstore
	bserv           blockservice.BlockService
}

// NewDAGStore creates a new DAGStore with the supplied
// Batching datastore
func NewDAGStore(batcher ds.Batching) *DAGStore {
	dstore := &DAGStore{
		ctx:   context.Background(), // @todo Do we need to properly pass through context chain on DAGStore?
		store: batcher,
	}

	dstore.setupBlockstore()
	dstore.setupBlockService()
	dstore.setupDAGService()

	return dstore
}

func (d *DAGStore) setupBlockstore() error {
	bs := blockstore.NewBlockstore(d.store)
	bs = blockstore.NewIdStore(bs)
	cachedbs, err := blockstore.CachedBlockstore(d.ctx, bs, blockstore.DefaultCacheOpts())
	if err != nil {
		return err
	}
	d.bstore = cachedbs
	return nil
}

func (d *DAGStore) setupBlockService() error {
	// if d.cfg.Offline {
	// 	d.bserv = blockservice.New(d.bstore, offline.Exchange(p.bstore))
	// 	return nil
	// }

	// bswapnet := network.NewFromIpfsHost(p.host, p.dht)
	// bswap := bitswap.New(p.ctx, bswapnet, p.bstore)
	// p.bserv = blockservice.New(p.bstore, bswap)

	// @todo Investigate if we need an Exchanger or if it can stay as nil
	d.bserv = blockservice.New(d.bstore, offline.Exchange(d.bstore))
	return nil
}

func (d *DAGStore) setupDAGService() error {
	d.DAGService = dag.NewDAGService(d.bserv)
	return nil
}
