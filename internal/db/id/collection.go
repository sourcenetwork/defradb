// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package id

import (
	"context"
	"strconv"

	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/db/sequence"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// ShortCollectionID returns the local, shortened, internal, collection id, which is used
// only in locations where using the full CID would be a waste of storage space.
//
// If there is no short id found for the given full id, a new one will be generated and saved.
func ShortCollectionID(
	ctx context.Context,
	txn datastore.Txn,
	collectionID string,
) (uint32, error) {
	key := keys.NewCollectionID(collectionID)

	valueBytes, err := txn.Systemstore().Get(ctx, key.Bytes())
	if err != nil {
		if errors.Is(err, corekv.ErrNotFound) {
			colSeq, err := sequence.Get(ctx, txn, keys.CollectionIDSequenceKey{})
			if err != nil {
				return 0, err
			}

			shortID, err := colSeq.Next(ctx, txn)
			if err != nil {
				return 0, err
			}

			err = txn.Systemstore().Set(ctx, key.Bytes(), []byte(strconv.Itoa(int(shortID))))
			if err != nil {
				return 0, err
			}

			return uint32(shortID), nil
		} else {
			return 0, err
		}
	}

	v, err := strconv.ParseUint(string(valueBytes), 10, 0)
	if err != nil {
		return 0, err
	}
	return uint32(v), nil
}
