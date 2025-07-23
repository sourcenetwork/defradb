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

type AdminACPKey struct{}

var _ Key = (*AdminACPKey)(nil)

func NewAdminACPKey() AdminACPKey {
	return AdminACPKey{}
}

func (k AdminACPKey) ToString() string {
	return ADMIN_ACP
}

func (k AdminACPKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k AdminACPKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}
