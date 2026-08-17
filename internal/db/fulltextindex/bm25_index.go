// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package fulltextindex

import (
	"context"
	"errors"
	"fmt"

	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/encoding"
	indexbm25 "github.com/sourcenetwork/defradb/internal/index/bm25"
	"github.com/sourcenetwork/defradb/internal/keys"
)

type bm25Index struct {
	ctx               context.Context
	collectionShortID uint32
	indexID           uint32
	epoch             uint32
	params            client.BM25Params
}

var _ Index = (*bm25Index)(nil)

func (i *bm25Index) Insert(docShortID uint64, text string) error {
	tokens := indexbm25.TokenFrequencies(text)
	if len(tokens) == 0 {
		return nil
	}

	ds := datastore.CtxMustGetTxn(i.ctx).Datastore()
	var length uint64
	for term, frequency := range tokens {
		key := keys.NewFullTextPostingKey(i.collectionShortID, i.indexID, i.epoch, term, docShortID)
		if err := ds.Set(i.ctx, &key, encodeUvarints(uint64(frequency))); err != nil {
			return fmt.Errorf("store full-text posting: %w", err)
		}
		length += uint64(frequency)
	}

	lengthKey := keys.NewFullTextLengthKey(i.collectionShortID, i.indexID, i.epoch, docShortID)
	stored, alreadyIndexed, err := i.readUvarints(&lengthKey, 1)
	if err != nil {
		return err
	}
	if err := ds.Set(i.ctx, &lengthKey, encodeUvarints(length)); err != nil {
		return fmt.Errorf("store full-text document length: %w", err)
	}
	if alreadyIndexed {
		return i.addToTotals(0, int64(length)-int64(stored[0]))
	}
	return i.addToTotals(1, int64(length))
}

func (i *bm25Index) Delete(
	docShortID uint64,
	text string,
	allowMissingPostings bool,
) (bool, error) {
	tokens := indexbm25.TokenFrequencies(text)
	if len(tokens) == 0 {
		return true, nil
	}

	ds := datastore.CtxMustGetTxn(i.ctx).Datastore()
	lengthKey := keys.NewFullTextLengthKey(i.collectionShortID, i.indexID, i.epoch, docShortID)
	stored, found, err := i.readUvarints(&lengthKey, 1)
	if err != nil || !found {
		return found, err
	}

	for term := range tokens {
		key := keys.NewFullTextPostingKey(i.collectionShortID, i.indexID, i.epoch, term, docShortID)
		exists, err := ds.Has(i.ctx, &key)
		if err != nil {
			return true, fmt.Errorf("check full-text posting: %w", err)
		}
		if !exists {
			if allowMissingPostings {
				continue
			}
			return true, ErrMissingPosting
		}
		if err := ds.Delete(i.ctx, &key); err != nil {
			return true, fmt.Errorf("delete full-text posting: %w", err)
		}
	}
	if err := ds.Delete(i.ctx, &lengthKey); err != nil {
		return true, fmt.Errorf("delete full-text document length: %w", err)
	}
	return true, i.addToTotals(-1, -int64(stored[0]))
}

func (i *bm25Index) addToTotals(documents, length int64) error {
	key := keys.NewFullTextTotalsKey(i.collectionShortID, i.indexID, i.epoch)
	totals, found, err := i.readUvarints(&key, 2)
	if err != nil {
		return err
	}
	if !found {
		totals = []uint64{0, 0}
	}
	value := encodeUvarints(addClamped(totals[0], documents), addClamped(totals[1], length))
	if err := datastore.CtxMustGetTxn(i.ctx).Datastore().Set(i.ctx, &key, value); err != nil {
		return fmt.Errorf("store full-text totals: %w", err)
	}
	return nil
}

func (i *bm25Index) readUvarints(key keys.Walkable, count int) ([]uint64, bool, error) {
	data, err := datastore.CtxMustGetTxn(i.ctx).Datastore().Get(i.ctx, key)
	if err != nil {
		if errors.Is(err, corekv.ErrNotFound) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("read full-text index record: %w", err)
	}
	values := make([]uint64, count)
	for index := range values {
		data, values[index], err = encoding.DecodeUvarintAscending(data)
		if err != nil {
			return nil, false, fmt.Errorf("decode full-text index record: %w", err)
		}
	}
	if len(data) != 0 {
		return nil, false, fmt.Errorf("decode full-text index record: trailing bytes")
	}
	return values, true, nil
}

func encodeUvarints(values ...uint64) []byte {
	var result []byte
	for _, value := range values {
		result = encoding.EncodeUvarintAscending(result, value)
	}
	return result
}

func addClamped(total uint64, change int64) uint64 {
	if change >= 0 {
		return total + uint64(change)
	}
	if uint64(-change) > total {
		return 0
	}
	return total - uint64(-change)
}
