// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package crdt

import (
	"context"
	"encoding/binary"

	ds "github.com/ipfs/go-datastore"

	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/keys"
)

func setPriority(
	ctx context.Context,
	store datastore.DSReaderWriter,
	key keys.DataStoreKey,
	priority uint64,
) error {
	prioK := key.WithPriorityFlag()
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, priority)
	if n == 0 {
		return ErrEncodingPriority
	}

	return store.Put(ctx, prioK.ToDS(), buf[0:n])
}

// get the current priority for given key
func getPriority(
	ctx context.Context,
	store datastore.DSReaderWriter,
	key keys.DataStoreKey,
) (uint64, error) {
	pKey := key.WithPriorityFlag()
	pbuf, err := store.Get(ctx, pKey.ToDS())
	if err != nil {
		if errors.Is(err, ds.ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}

	prio, num := binary.Uvarint(pbuf)
	if num <= 0 {
		return 0, ErrDecodingPriority
	}
	return prio, nil
}
