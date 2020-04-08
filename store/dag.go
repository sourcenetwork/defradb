package store

import (
	"github.com/sourcenetwork/defradb/core"

	blockstore "github.com/ipfs/go-ipfs-blockstore"
	logging "github.com/ipfs/go-log"
)

var (
	log = logging.Logger("defradb.merkle.crdt")
)

// DAGStore is the interface to the underlying BlockStore and BlockService
type dagStore struct {
	blockstore.Blockstore // become a Blockstore
	store                 core.DSReaderWriter
	// bstore          blockstore.Blockstore
	// bserv           blockservice.BlockService
}

// NewDAGStore creates a new DAGStore with the supplied
// Batching datastore
func NewDAGStore(store core.DSReaderWriter) core.DAGStore {
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
