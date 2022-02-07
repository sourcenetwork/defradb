package utils

import (
	"github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
)

func NewCidV1(data []byte) (cid.Cid, error) {
	pref := cid.Prefix{
		Version:  1,
		Codec:    cid.Raw,
		MhType:   mh.SHA2_256,
		MhLength: -1, // default length
	}

	// And then feed it some data
	return pref.Sum(data)
}
