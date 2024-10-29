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

import ds "github.com/ipfs/go-datastore"

type ReplicatorKey struct {
	ReplicatorID string
}

var _ Key = (*ReplicatorKey)(nil)

func NewReplicatorKey(id string) ReplicatorKey {
	return ReplicatorKey{ReplicatorID: id}
}

func (k ReplicatorKey) ToString() string {
	result := REPLICATOR

	if k.ReplicatorID != "" {
		result = result + "/" + k.ReplicatorID
	}

	return result
}

func (k ReplicatorKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k ReplicatorKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}
