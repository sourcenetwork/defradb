package db

import (
	blocks "github.com/ipfs/go-block-format"
	cid "github.com/ipfs/go-cid"
)

func (db *DB) GetBlock(c cid.Cid) (blocks.Block, error) {
	return db.dagstore.Get(c)
}
