package base

import (
	ds "github.com/ipfs/go-datastore"
)

var (
	// Individual Store Keys
	RootStoreKey   = ds.NewKey("/db")
	SystemStoreKey = RootStoreKey.ChildString("/system")
	DataStoreKey   = RootStoreKey.ChildString("/data")
	HeadStoreKey   = RootStoreKey.ChildString("/heads")
	BlockStoreKey  = RootStoreKey.ChildString("/blocks")
)
