package core

import (
	blockstore "github.com/ipfs/go-ipfs-blockstore"
)

// DAGStore proxies the ipld.DAGService under the /core namespace for future-proofing
type DAGStore interface {
	blockstore.Blockstore
}
