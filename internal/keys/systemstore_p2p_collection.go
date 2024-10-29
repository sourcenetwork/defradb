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
	"strings"

	ds "github.com/ipfs/go-datastore"

	"github.com/sourcenetwork/defradb/errors"
)

type P2PCollectionKey struct {
	CollectionID string
}

var _ Key = (*P2PCollectionKey)(nil)

// New
func NewP2PCollectionKey(collectionID string) P2PCollectionKey {
	return P2PCollectionKey{CollectionID: collectionID}
}

func NewP2PCollectionKeyFromString(key string) (P2PCollectionKey, error) {
	keyArr := strings.Split(key, "/")
	if len(keyArr) != 4 {
		return P2PCollectionKey{}, errors.WithStack(ErrInvalidKey, errors.NewKV("Key", key))
	}
	return NewP2PCollectionKey(keyArr[3]), nil
}

func (k P2PCollectionKey) ToString() string {
	result := P2P_COLLECTION

	if k.CollectionID != "" {
		result = result + "/" + k.CollectionID
	}

	return result
}

func (k P2PCollectionKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k P2PCollectionKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}
