// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package db

import (
	"context"
	"strings"

	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/description"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// nextShortDocID returns the next local storage ID.
func (db *DB) nextShortDocID() string {
	return id.FormatShortDocID(db.docIDSequence.Add(1))
}

// seedDocIDSequence restores the counter from stored primary keys.
func (db *DB) seedDocIDSequence(ctx context.Context) error {
	cols, err := description.GetCollections(ctx, db.collectionRepository)
	if err != nil {
		return err
	}

	txn := datastore.CtxMustGetTxn(ctx)
	var maxSeq uint64
	for _, col := range cols {
		collectionShortID, err := id.GetShortCollectionID(ctx, col.CollectionID)
		if err != nil {
			return err
		}

		prefix := keys.PrimaryDataStoreKey{CollectionShortID: collectionShortID}
		iter, err := txn.Datastore().Iterator(ctx, datastore.IterOptions{
			Prefix:   prefix,
			Reverse:  true,
			KeysOnly: true,
		})
		if err != nil {
			return err
		}

		hasValue, iterErr := iter.Next()
		if iterErr != nil {
			return errors.Join(iterErr, iter.Close())
		}
		if !hasValue {
			if err := iter.Close(); err != nil {
				return err
			}
			continue
		}

		shortDocID := strings.TrimPrefix(string(iter.Key()), prefix.ToString()+"/")
		seq, err := id.ParseShortDocID(shortDocID)
		closeErr := iter.Close()
		if err != nil {
			return errors.Join(err, closeErr)
		}
		if closeErr != nil {
			return closeErr
		}
		if seq > maxSeq {
			maxSeq = seq
		}
	}

	db.docIDSequence.Store(maxSeq)
	return nil
}
