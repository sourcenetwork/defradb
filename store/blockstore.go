package store

import (
	"context"
	"errors"

	"github.com/sourcenetwork/defradb/core"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	dsq "github.com/ipfs/go-datastore/query"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	dshelp "github.com/ipfs/go-ipfs-ds-help"
)

// Blockstore implementation taken from https://github.com/ipfs/go-ipfs-blockstore/blob/master/blockstore.go

// ErrHashMismatch is an error returned when the hash of a block
// is different than expected.
var ErrHashMismatch = errors.New("block in storage has different hash than requested")

// ErrNotFound is an error returned when a block is not found.
var ErrNotFound = errors.New("blockstore: block not found")

// NewBlockstore returns a default Blockstore implementation
// using the provided datastore.Batching backend.
func NewBlockstore(store core.DSReaderWriter) blockstore.Blockstore {
	return &bstore{
		store: store,
	}
}

type bstore struct {
	store core.DSReaderWriter

	rehash bool
}

func (bs *bstore) HashOnRead(enabled bool) {
	bs.rehash = enabled
}

func (bs *bstore) Get(k cid.Cid) (blocks.Block, error) {
	if !k.Defined() {
		log.Error("undefined cid in blockstore")
		return nil, ErrNotFound
	}
	bdata, err := bs.store.Get(dshelp.CidToDsKey(k))
	if err == ds.ErrNotFound {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if bs.rehash {
		rbcid, err := k.Prefix().Sum(bdata)
		if err != nil {
			return nil, err
		}

		if !rbcid.Equals(k) {
			return nil, ErrHashMismatch
		}

		return blocks.NewBlockWithCid(bdata, rbcid)
	}
	return blocks.NewBlockWithCid(bdata, k)
}

func (bs *bstore) Put(block blocks.Block) error {
	k := dshelp.CidToDsKey(block.Cid())

	// Has is cheaper than Put, so see if we already have it
	exists, err := bs.store.Has(k)
	if err == nil && exists {
		return nil // already stored.
	}
	return bs.store.Put(k, block.RawData())
}

func (bs *bstore) PutMany(blocks []blocks.Block) error {
	for _, b := range blocks {
		k := dshelp.CidToDsKey(b.Cid())
		exists, err := bs.store.Has(k)
		if err == nil && exists {
			continue
		}

		err = bs.store.Put(k, b.RawData())
		if err != nil {
			return err
		}
	}
	return nil
}

func (bs *bstore) Has(k cid.Cid) (bool, error) {
	return bs.store.Has(dshelp.CidToDsKey(k))
}

func (bs *bstore) GetSize(k cid.Cid) (int, error) {
	size, err := bs.store.GetSize(dshelp.CidToDsKey(k))
	if err == ds.ErrNotFound {
		return -1, ErrNotFound
	}
	return size, err
}

func (bs *bstore) DeleteBlock(k cid.Cid) error {
	return bs.store.Delete(dshelp.CidToDsKey(k))
}

// AllKeysChan runs a query for keys from the blockstore.
// this is very simplistic, in the future, take dsq.Query as a param?
//
// AllKeysChan respects context.
func (bs *bstore) AllKeysChan(ctx context.Context) (<-chan cid.Cid, error) {

	// KeysOnly, because that would be _a lot_ of data.
	q := dsq.Query{KeysOnly: true}
	res, err := bs.store.Query(q)
	if err != nil {
		return nil, err
	}

	output := make(chan cid.Cid, dsq.KeysOnlyBufSize)
	go func() {
		defer func() {
			res.Close() // ensure exit (signals early exit, too)
			close(output)
		}()

		for {
			e, ok := res.NextSync()
			if !ok {
				return
			}
			if e.Error != nil {
				log.Errorf("blockstore.AllKeysChan got err: %s", e.Error)
				return
			}

			// need to convert to key.Key using key.KeyFromDsKey.
			bk, err := dshelp.BinaryFromDsKey(ds.RawKey(e.Key))
			if err != nil {
				log.Warningf("error parsing key from binary: %s", err)
				continue
			}
			k := cid.NewCidV1(cid.Raw, bk)
			select {
			case <-ctx.Done():
				return
			case output <- k:
			}
		}
	}()

	return output, nil
}
