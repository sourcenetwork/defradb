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

import ds "github.com/ipfs/go-datastore"

// CollectionVersionIDSequenceKey is used to key the sequence used to generate collection version ids.
type CollectionVersionIDSequenceKey struct{}

var _ Key = (*CollectionVersionIDSequenceKey)(nil)

func (k CollectionVersionIDSequenceKey) ToString() string {
	return VERSION_SEQ
}

func (k CollectionVersionIDSequenceKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k CollectionVersionIDSequenceKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}
