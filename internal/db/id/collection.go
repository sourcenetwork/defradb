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

	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/internal/db/sequence"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// GetShortCollectionID returns the local, shortened, internal, collection id, which is used
// only in locations where using the full CID would be a waste of storage space.
func GetShortCollectionID(
	ctx context.Context,
	txn datastore.Txn,
	collectionID string,
) (uint32, error) {
	key := keys.NewCollectionID(collectionID)

	valueBytes, err := txn.Systemstore().Get(ctx, key.Bytes())
	if err != nil {
		return 0, err
	}

	v, err := strconv.ParseUint(string(valueBytes), 10, 0)
	if err != nil {
		return 0, err
	}
	return uint32(v), nil
}

// SetShortCollectionID sets and stores the short collection id, if it does not already exist.
func SetShortCollectionID(
	ctx context.Context,
	txn datastore.Txn,
	collectionID string,
) error {
	key := keys.NewCollectionID(collectionID)

	hasShortID, err := txn.Systemstore().Has(ctx, key.Bytes())
	if err != nil {
		return err
	}
	if hasShortID {
		return nil
	}

	colSeq, err := sequence.Get(ctx, txn, keys.CollectionIDSequenceKey{})
	if err != nil {
		return err
	}

	shortID, err := colSeq.Next(ctx, txn)
	if err != nil {
		return err
	}

	return txn.Systemstore().Set(ctx, key.Bytes(), []byte(strconv.Itoa(int(shortID))))
}
