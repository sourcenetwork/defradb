// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

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
