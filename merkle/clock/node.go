package clock

import (
	cid "github.com/ipfs/go-cid"
)

type Node struct {
	delta []byte
	heads []cid.Cid
}
