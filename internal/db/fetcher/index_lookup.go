// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package fetcher

import (
	"context"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// IndexLookup probes a single-field index for a given field value, hiding the index entry's storage
// layout (collection short ID, epoch namespace, field ordering) from callers. It is built once for
// an index via NewIndexLookup and then reused for many values, since only the value changes between
// lookups.
//
// It currently exposes Has for existence checks; a value-returning Fetch/Get can be added later
// without callers needing to learn the key layout.
type IndexLookup struct {
	collectionShortID uint32
	indexID           uint32
	epoch             uint32
	descending        bool
}

// NewIndexLookup resolves everything an index lookup needs once: the collection's short ID and the
// index's live epoch. The index must be single-field; index.Fields[0] supplies the field ordering.
func NewIndexLookup(
	ctx context.Context,
	txn datastore.Txn,
	col client.Collection,
	index client.IndexDescription,
) (*IndexLookup, error) {
	collectionID := col.Version().CollectionID

	collectionShortID, err := id.GetCollectionShortID(ctx, collectionID)
	if err != nil {
		return nil, err
	}

	epoch, err := ReadIndexEpoch(ctx, txn, collectionID, index.ID)
	if err != nil {
		return nil, err
	}

	return &IndexLookup{
		collectionShortID: collectionShortID,
		indexID:           index.ID,
		epoch:             epoch,
		descending:        index.Fields[0].Descending,
	}, nil
}

// Has reports whether an index entry exists for the given field value, using the transaction bound
// to ctx. The caller owns fetch accounting.
func (l *IndexLookup) Has(ctx context.Context, value client.NormalValue) (bool, error) {
	txn := datastore.CtxMustGetTxn(ctx)
	key := keys.NewIndexDataStoreKey(l.collectionShortID, l.indexID, l.epoch, []keys.IndexedField{
		{Value: value, Descending: l.descending},
	})
	return txn.Datastore().Has(ctx, &key)
}
