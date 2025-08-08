// Copyright 2025 Democratized Data Foundation
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

type NodeACPKey struct{}

var _ Key = (*NodeACPKey)(nil)

func NewNodeACPKey() NodeACPKey {
	return NodeACPKey{}
}

func (k NodeACPKey) ToString() string {
	return NODE_ACP
}

func (k NodeACPKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k NodeACPKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}
