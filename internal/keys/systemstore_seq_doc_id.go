// Copyright 2026 Democratized Data Foundation
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

// DocShortIDSequenceKey is used to key the node-local DocShortID sequence.
type DocShortIDSequenceKey struct{}

var _ Key = (*DocShortIDSequenceKey)(nil)

func NewDocShortIDSequenceKey() DocShortIDSequenceKey {
	return DocShortIDSequenceKey{}
}

func (k DocShortIDSequenceKey) ToString() string {
	return DOC_SHORT_ID_SEQ
}

func (k DocShortIDSequenceKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k DocShortIDSequenceKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}
