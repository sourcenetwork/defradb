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
	"bytes"

	ds "github.com/ipfs/go-datastore"

	"github.com/sourcenetwork/defradb/internal/encoding"
)

const (
	PRIMARY_KEY = "/pk"
)

type PrimaryDataStoreKey struct {
	CollectionShortID uint32
	DocShortID        uint32
}

var _ Key = (*PrimaryDataStoreKey)(nil)
var _ CollectionedKey = PrimaryDataStoreKey{}

func NewPrimaryDataStoreKey(key string) (PrimaryDataStoreKey, error) {
	data := []byte(key)
	if len(data) == 0 || data[0] != '/' {
		return PrimaryDataStoreKey{}, ErrInvalidKey
	}
	data = data[1:]

	data, collectionShortID, err := DecodeCollectionShortIDPrefix(data)
	if err != nil {
		return PrimaryDataStoreKey{}, err
	}
	if !bytes.HasPrefix(data, []byte(PRIMARY_KEY)) {
		return PrimaryDataStoreKey{}, ErrInvalidKey
	}
	data = data[len(PRIMARY_KEY):]

	result := PrimaryDataStoreKey{CollectionShortID: collectionShortID}
	if len(data) > 0 {
		if data[0] != '/' {
			return PrimaryDataStoreKey{}, ErrInvalidKey
		}
		data, docShortID, err := DecodeDocShortIDPrefix(data[1:])
		if err != nil {
			return PrimaryDataStoreKey{}, err
		}
		if len(data) != 0 {
			return PrimaryDataStoreKey{}, ErrInvalidKey
		}
		result.DocShortID = docShortID
	}
	return result, nil
}

func (k PrimaryDataStoreKey) ToDataStoreKey() DataStoreKey {
	return DataStoreKey{
		CollectionShortID: k.CollectionShortID,
		DocShortID:        k.DocShortID,
	}
}

func (k PrimaryDataStoreKey) Bytes() []byte {
	result := []byte{}
	if k.CollectionShortID != 0 {
		result = encoding.EncodeUvarintAscending([]byte{'/'}, uint64(k.CollectionShortID))
	}
	result = append(result, []byte(PRIMARY_KEY)...)
	if k.DocShortID != 0 {
		result = append(result, '/')
		result = append(result, EncodeDocShortID(k.DocShortID)...)
	}
	return result
}

func (k PrimaryDataStoreKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

func (k PrimaryDataStoreKey) ToString() string {
	return string(k.Bytes())
}

func (k PrimaryDataStoreKey) GetCollectionShortID() uint32 {
	return k.CollectionShortID
}
