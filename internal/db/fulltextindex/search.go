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
	"math"
	"slices"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/encoding"
	indexbm25 "github.com/sourcenetwork/defradb/internal/index/bm25"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// SearchTarget is one full-text index contributing to a ranked search.
type SearchTarget struct {
	IndexID     uint32
	Epoch       uint32
	Description client.FullTextIndexDescription
	Boost       float64
}

// Hit is one matching document and the sum of its boosted per-field scores.
type Hit struct {
	DocShortID uint64
	Score      float64
}

// SearchResult contains all ranked hits and the posting entries read to produce them.
type SearchResult struct {
	Hits         []Hit
	PostingsRead uint64
}

type posting struct {
	docShortID uint64
	frequency  uint64
}

type termPostings struct {
	postings []posting
	next     int
}

// Search scores every document matching at least one retained query token. Each target is scored
// against its own corpus statistics and BM25 parameters, then multiplied by its boost and summed.
// It deliberately accumulates all matches before yielding them so permission/filter rejection above
// this module can continue pulling past rejected high-scoring documents without a guessed overfetch.
func Search(
	ctx context.Context,
	collectionShortID uint32,
	query string,
	targets []SearchTarget,
) (SearchResult, error) {
	totals := make(map[uint64]float64)
	var postingsRead uint64

	for _, target := range targets {
		if target.Boost == 0 {
			continue
		}
		index, err := Open(ctx, collectionShortID, target.IndexID, target.Epoch, target.Description)
		if err != nil {
			return SearchResult{}, err
		}
		result, err := index.Search(query)
		if err != nil {
			return SearchResult{}, err
		}
		postingsRead += result.PostingsRead
		for _, hit := range result.Hits {
			totals[hit.DocShortID] += target.Boost * hit.Score
		}
	}

	return rankedResult(totals, postingsRead), nil
}

func (i *bm25Index) Search(query string) (SearchResult, error) {
	termlist := mapsKeys(indexbm25.TokenFrequencies(query))
	slices.Sort(termlist)
	totals := make(map[uint64]float64)

	count, totalLength, found, err := i.readTotals()
	if err != nil {
		return SearchResult{}, err
	}
	if !found || count == 0 {
		return SearchResult{}, nil
	}
	averageLength := float64(totalLength) / float64(count)

	var postingsRead uint64
	terms := make([]termPostings, 0, len(termlist))
	for _, term := range termlist {
		postings, err := i.readPostings(term)
		if err != nil {
			return SearchResult{}, err
		}
		postingsRead += uint64(len(postings))
		if len(postings) > 0 {
			terms = append(terms, termPostings{postings: postings})
		}
	}

	// Posting lists are ordered by document short ID. Merge them so each matched document's
	// length is point-read once, regardless of how many query terms it holds.
	for {
		docShortID := ^uint64(0)
		foundDoc := false
		for termIndex := range terms {
			if terms[termIndex].next < len(terms[termIndex].postings) &&
				terms[termIndex].postings[terms[termIndex].next].docShortID <= docShortID {
				docShortID = terms[termIndex].postings[terms[termIndex].next].docShortID
				foundDoc = true
			}
		}
		if !foundDoc {
			break
		}

		length, foundLength, err := i.readLength(docShortID)
		if err != nil {
			return SearchResult{}, err
		}
		var score float64
		for termIndex := range terms {
			if terms[termIndex].next >= len(terms[termIndex].postings) ||
				terms[termIndex].postings[terms[termIndex].next].docShortID != docShortID {
				continue
			}
			posting := terms[termIndex].postings[terms[termIndex].next]
			terms[termIndex].next++
			if !foundLength {
				continue
			}
			score += indexbm25.ScoreTerm(
				posting.frequency,
				uint64(len(terms[termIndex].postings)),
				count,
				length,
				averageLength,
				i.params,
			)
		}
		if foundLength {
			totals[docShortID] += score
		}
	}

	return rankedResult(totals, postingsRead), nil
}

func rankedResult(totals map[uint64]float64, postingsRead uint64) SearchResult {
	hits := make([]Hit, 0, len(totals))
	for docShortID, score := range totals {
		if math.IsNaN(score) {
			continue
		}
		hits = append(hits, Hit{DocShortID: docShortID, Score: score})
	}
	slices.SortFunc(hits, func(a, b Hit) int {
		if a.Score > b.Score {
			return -1
		}
		if a.Score < b.Score {
			return 1
		}
		if a.DocShortID < b.DocShortID {
			return -1
		}
		if a.DocShortID > b.DocShortID {
			return 1
		}
		return 0
	})
	return SearchResult{Hits: hits, PostingsRead: postingsRead}
}

func mapsKeys(values map[string]int) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	return result
}

func (i *bm25Index) readTotals() (uint64, uint64, bool, error) {
	key := keys.NewFullTextTotalsKey(i.collectionShortID, i.indexID, i.epoch)
	values, found, err := i.readUvarints(&key, 2)
	if err != nil || !found {
		return 0, 0, found, err
	}
	return values[0], values[1], true, nil
}

func (i *bm25Index) readLength(docShortID uint64) (uint64, bool, error) {
	key := keys.NewFullTextLengthKey(i.collectionShortID, i.indexID, i.epoch, docShortID)
	values, found, err := i.readUvarints(&key, 1)
	if err != nil || !found {
		return 0, found, err
	}
	return values[0], true, nil
}

func (i *bm25Index) readPostings(term string) ([]posting, error) {
	prefixKey := keys.NewFullTextPostingKey(i.collectionShortID, i.indexID, i.epoch, term, 0)
	prefix := prefixKey.Bytes()
	iter, err := datastore.CtxMustGetTxn(i.ctx).Datastore().Iterator(
		i.ctx, datastore.IterOptions{Prefix: &prefixKey},
	)
	if err != nil {
		return nil, err
	}
	var result []posting
	for {
		found, err := iter.Next()
		if err != nil || !found {
			return result, errors.Join(err, iter.Close())
		}
		entryKey := iter.Key()
		if len(entryKey) <= len(prefix)+1 {
			return nil, errors.Join(keys.ErrInvalidKey, iter.Close())
		}
		docShortID, err := keys.DecodeDocShortID(entryKey[len(prefix)+1:])
		if err != nil {
			return nil, errors.Join(err, iter.Close())
		}
		value, err := iter.Value()
		if err != nil {
			return nil, errors.Join(err, iter.Close())
		}
		remaining, frequency, err := encoding.DecodeUvarintAscending(value)
		if err != nil || len(remaining) != 0 {
			if err == nil {
				err = errors.New("trailing bytes in full-text posting")
			}
			return nil, errors.Join(err, iter.Close())
		}
		result = append(result, posting{docShortID: docShortID, frequency: frequency})
	}
}
