package main

import (
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
)

func main() {
	h1, err := multihash.Sum([]byte("hi"), multihash.SHA2_256, -1)
	if err != nil {
		panic(err)
	}
	c := cid.NewCidV0(h1)
	println(c.String())
}
