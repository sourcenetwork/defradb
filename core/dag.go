package core

import (
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	ipld "github.com/ipfs/go-ipld-format"
)

// DAGStore proxies the ipld.DAGService under the /core namespace for future-proofing
type DAGStore interface {
	ipld.DAGService
	Blockstore() blockstore.Blockstore
}
