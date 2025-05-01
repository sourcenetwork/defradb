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
	"fmt"

	ds "github.com/ipfs/go-datastore"
)

const (
	PRIMARY_KEY = "/pk"
)

type PrimaryDataStoreKey struct {
	CollectionShortID uint32
	DocID             string
}

var _ Key = (*PrimaryDataStoreKey)(nil)

func (k PrimaryDataStoreKey) ToDataStoreKey() DataStoreKey {
	return DataStoreKey{
		CollectionShortID: k.CollectionShortID,
		DocID:             k.DocID,
	}
}

func (k PrimaryDataStoreKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k PrimaryDataStoreKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

func (k PrimaryDataStoreKey) ToString() string {
	result := ""

	if k.CollectionShortID != 0 {
		result = result + "/" + fmt.Sprint(k.CollectionShortID)
	}
	result = result + PRIMARY_KEY
	if k.DocID != "" {
		result = result + "/" + k.DocID
	}

	return result
}
