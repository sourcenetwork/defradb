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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/fulltextindex"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// rankedFetcher fetches documents in the exact score order returned by fulltextindex.Search. It is
// intentionally separate from indexFetcher, whose key decoding is only valid for ordered/trigram
// layouts.
type rankedFetcher struct {
	ctx               context.Context
	txn               datastore.Txn
	col               client.Collection
	fieldsByID        map[uint32]client.CollectionFieldDescription
	collectionShortID uint32
	rank              *Rank
	hits              []fulltextindex.Hit
	next              int
	currentDocShortID uint64
	execInfo          *ExecInfo
	postingsRead      uint64
}

var _ fetcher = (*rankedFetcher)(nil)

func newRankedFetcher(
	ctx context.Context,
	txn datastore.Txn,
	fieldsByID map[uint32]client.CollectionFieldDescription,
	col client.Collection,
	rank *Rank,
	execInfo *ExecInfo,
) (*rankedFetcher, error) {
	collectionShortID, err := id.GetCollectionShortID(ctx, col.Version().CollectionID)
	if err != nil {
		return nil, err
	}
	targets := make([]fulltextindex.SearchTarget, 0, len(rank.Targets))
	for _, target := range rank.Targets {
		epoch, err := ReadIndexEpoch(ctx, txn, col.Version().CollectionID, target.Index.ID)
		if err != nil {
			return nil, err
		}
		desc, ok := target.Index.GetFullText()
		if !ok || desc == nil {
			return nil, NewErrRankTargetNotFullText(target.Index.Name)
		}
		targets = append(targets, fulltextindex.SearchTarget{
			IndexID:     target.Index.ID,
			Epoch:       epoch,
			Description: *desc,
			Boost:       target.Boost,
		})
	}
	result, err := fulltextindex.Search(ctx, collectionShortID, rank.Query, targets)
	if err != nil {
		return nil, err
	}
	return &rankedFetcher{
		ctx:               ctx,
		txn:               txn,
		col:               col,
		fieldsByID:        fieldsByID,
		collectionShortID: collectionShortID,
		rank:              rank,
		hits:              result.Hits,
		execInfo:          execInfo,
		postingsRead:      result.PostingsRead,
	}, nil
}

func (f *rankedFetcher) NextDoc() (immutable.Option[string], error) {
	for f.next < len(f.hits) {
		hit := f.hits[f.next]
		f.next++
		docID, found, err := id.GetDocID(f.ctx, hit.DocShortID)
		if err != nil {
			return immutable.None[string](), err
		}
		if !found {
			continue
		}
		f.currentDocShortID = hit.DocShortID
		f.rank.Score = hit.Score
		f.execInfo.IndexesFetched += f.postingsRead
		f.postingsRead = 0
		return immutable.Some(docID), nil
	}
	return immutable.None[string](), nil
}

func (f *rankedFetcher) GetFields() (immutable.Option[EncodedDocument], error) {
	prefix := keys.DataStoreKey{
		CollectionShortID: f.collectionShortID,
		DocShortID:        f.currentDocShortID,
	}
	prefixFetcher, err := newPrefixFetcher(
		f.ctx, f.txn, []keys.DataStoreKey{prefix}, f.col, f.fieldsByID, client.Active, f.execInfo,
	)
	if err != nil {
		return immutable.None[EncodedDocument](), err
	}
	_, err = prefixFetcher.NextDoc()
	if err != nil {
		return immutable.None[EncodedDocument](), errors.Join(err, prefixFetcher.Close())
	}
	doc, err := prefixFetcher.GetFields()
	return doc, errors.Join(err, prefixFetcher.Close())
}

func (f *rankedFetcher) Close() error {
	f.hits = nil
	return nil
}
