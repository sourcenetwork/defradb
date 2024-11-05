// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package keys

import (
	ds "github.com/ipfs/go-datastore"
)

// Key is an interface that represents a key in the database.
type Key interface {
	ToString() string
	Bytes() []byte
	ToDS() ds.Key
}

// Walkable represents a key in the database that can be 'walked along'
// by prefixing the end of the key.
type Walkable interface {
	Key
	PrefixEnd() Walkable
}

// PrettyPrint returns the human readable version of the given key.
func PrettyPrint(k Key) string {
	switch typed := k.(type) {
	case DataStoreKey:
		return typed.PrettyPrint()
	default:
		return typed.ToString()
	}
}
