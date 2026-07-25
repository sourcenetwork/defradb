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
	"container/heap"
	"context"
	"maps"
	"math"
	"slices"

	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/encoding"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// bm25Posting is one document a term occurs in, and how often.
type bm25Posting struct {
	docShortID uint64
	frequency  float64
}

// bm25Term is one term of the query: the documents holding it, in ascending document short ID,
// and the weight every one of them gets from it.
type bm25Term struct {
	postings []bm25Posting
	idf      float64
	// next is how far the merge has walked into postings.
	next int
}

// bm25ScoredDoc is a document and its score against the whole query.
type bm25ScoredDoc struct {
	docShortID uint64
	score      float64
}

// bm25ScoreHeap orders scored documents best first, breaking a tie by document short ID so that
// the same query over the same documents always answers in the same order.
type bm25ScoreHeap []bm25ScoredDoc

func (h bm25ScoreHeap) Len() int { return len(h) }
func (h bm25ScoreHeap) Less(i, j int) bool {
	if h[i].score != h[j].score {
		return h[i].score > h[j].score
	}
	return h[i].docShortID < h[j].docShortID
}
func (h bm25ScoreHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h *bm25ScoreHeap) Push(x any)   { *h = append(*h, x.(bm25ScoredDoc)) }
func (h *bm25ScoreHeap) Pop() any {
	old := *h
	last := old[len(old)-1]
	*h = old[:len(old)-1]
	return last
}

// bm25IndexIterator yields the documents matching a query in descending order of how well they
// match it.
//
// It is a lazy iterator rather than a call taking the number of results wanted, which is what
// keeps a limit correct under access control. The limit node pulls one document at a time and
// stops when it has enough, and the documents access control removes below it are simply pulled
// past, so no result is ever lost to an over-fetch factor guessed too low. Pruning schemes such
// as WAND cannot be used for the same reason: their threshold is the current K-th best score, so
// they need K before they start.
//
// ponytail: Init walks every posting of every query term and holds one entry per matching
// document, so the first document costs the whole traversal and a query on a term most documents
// hold pays for all of them. That is the accumulate-then-drain shape, and the way past it is
// impact-ordered postings: key each entry by its quantised contribution so a forward scan reads
// each term's documents best first, which lets the threshold algorithm emit results before the
// lists are exhausted. It also bakes the field length and the scoring parameters into the key,
// so it is a different index rather than a change to this iterator.
type bm25IndexIterator struct {
	// key is the index's own prefix, /<collection short ID>/<index ID>/<epoch>/.
	key   keys.IndexDataStoreKey
	terms []string
	k1    float64
	b     float64
	// rank receives the score of the document Next last returned.
	rank      *Rank
	indexName string
	execInfo  *ExecInfo

	scored bm25ScoreHeap
	// postingsRead is what Init read, carried to the first Next. The execution counters are
	// reset per fetched document, and Init runs before the first of them.
	postingsRead uint64
}

var _ indexIterator = (*bm25IndexIterator)(nil)

func (iter *bm25IndexIterator) Init(ctx context.Context, store datastore.Keyedstore) error {
	iter.scored = nil
	iter.postingsRead = 0

	totalsKey := iter.partKey(core.BM25TotalsPart, "", 0)
	totals, err := readUvarints(ctx, store, &totalsKey, 2)
	if err != nil {
		return NewErrCreateIndexIterator(err, iter.indexName)
	}
	// No totals record means no document has ever been written through this index, so there is
	// nothing to score and nothing to divide by.
	if totals == nil || totals[0] == 0 {
		return nil
	}
	documentCount := float64(totals[0])
	averageLength := float64(totals[1]) / documentCount

	terms := make([]bm25Term, 0, len(iter.terms))
	for _, term := range iter.terms {
		postings, err := iter.readPostings(ctx, store, term)
		if err != nil {
			return NewErrCreateIndexIterator(err, iter.indexName)
		}
		if len(postings) == 0 {
			continue
		}
		// The number of documents holding the term is the number of its postings; nothing stores
		// it separately, and the range has to be walked either way.
		//
		// The one that is added inside the logarithm is not cosmetic. Without it the inverse
		// document frequency of a term held by more than half the documents is negative, and a
		// document is then penalised for matching the query.
		documentFrequency := float64(len(postings))
		terms = append(terms, bm25Term{
			postings: postings,
			idf:      math.Log(1 + (documentCount-documentFrequency+0.5)/(documentFrequency+0.5)),
		})
	}

	if err := iter.score(ctx, store, terms, averageLength); err != nil {
		return NewErrCreateIndexIterator(err, iter.indexName)
	}
	heap.Init(&iter.scored)
	return nil
}

// score walks the terms' posting lists together, lowest document short ID first, and scores each
// document it reaches against every term holding it.
//
// The store keeps the postings of a term sorted by document short ID, so this is a merge over
// already sorted lists rather than an accumulator keyed by document.
func (iter *bm25IndexIterator) score(
	ctx context.Context,
	store datastore.Keyedstore,
	terms []bm25Term,
	averageLength float64,
) error {
	for {
		docShortID := uint64(math.MaxUint64)
		found := false
		for i := range terms {
			if terms[i].next < len(terms[i].postings) && terms[i].postings[terms[i].next].docShortID <= docShortID {
				docShortID = terms[i].postings[terms[i].next].docShortID
				found = true
			}
		}
		if !found {
			return nil
		}

		lengthKey := iter.partKey(core.BM25LengthPart, "", docShortID)
		length, err := readUvarints(ctx, store, &lengthKey, 1)
		if err != nil {
			return err
		}

		var score float64
		for i := range terms {
			if terms[i].next >= len(terms[i].postings) ||
				terms[i].postings[terms[i].next].docShortID != docShortID {
				continue
			}
			frequency := terms[i].postings[terms[i].next].frequency
			terms[i].next++
			if length == nil {
				continue
			}
			saturation := iter.k1 * (1 - iter.b + iter.b*float64(length[0])/averageLength)
			score += terms[i].idf * frequency * (iter.k1 + 1) / (frequency + saturation)
		}

		if length == nil {
			// A posting with no field length recorded for its document is an index that lost
			// half of a write. There is nothing to normalise the term frequency against, so the
			// document is passed over rather than scored against a made-up length.
			continue
		}
		iter.scored = append(iter.scored, bm25ScoredDoc{docShortID: docShortID, score: score})
	}
}

// readPostings returns every document holding term and how often, in ascending document short ID.
func (iter *bm25IndexIterator) readPostings(
	ctx context.Context,
	store datastore.Keyedstore,
	term string,
) ([]bm25Posting, error) {
	key := iter.partKey(core.BM25PostingPart, term, 0)
	prefix := key.Bytes()

	postings, err := store.Iterator(ctx, datastore.IterOptions{Prefix: &key})
	if err != nil {
		return nil, err
	}

	var result []bm25Posting
	for {
		hasNext, err := postings.Next()
		if err != nil || !hasNext {
			return result, errors.Join(err, postings.Close())
		}

		entryKey := postings.Key()
		if len(entryKey) <= len(prefix)+1 {
			return nil, errors.Join(keys.ErrInvalidKey, postings.Close())
		}
		docShortID, err := keys.DecodeDocShortID(entryKey[len(prefix)+1:])
		if err != nil {
			return nil, errors.Join(err, postings.Close())
		}

		value, err := postings.Value()
		if err != nil {
			return nil, errors.Join(err, postings.Close())
		}
		_, frequency, err := encoding.DecodeUvarintAscending(value)
		if err != nil {
			return nil, errors.Join(err, postings.Close())
		}

		iter.postingsRead++
		result = append(result, bm25Posting{docShortID: docShortID, frequency: float64(frequency)})
	}
}

func (iter *bm25IndexIterator) partKey(part string, term string, docShortID uint64) keys.IndexDataStoreKey {
	return core.BM25Key(iter.key.CollectionShortID, iter.key.IndexID, iter.key.Epoch, part, term, docShortID)
}

func (iter *bm25IndexIterator) Next() (indexIterResult, error) {
	if iter.scored.Len() == 0 {
		return indexIterResult{}, nil
	}
	best := heap.Pop(&iter.scored).(bm25ScoredDoc)
	iter.rank.Score = best.score
	iter.execInfo.IndexesFetched += iter.postingsRead
	iter.postingsRead = 0

	// The entries hold terms of the value rather than the value, so there is nothing to report in
	// the key's fields; the document short ID is the whole result.
	return indexIterResult{
		key:      &keys.IndexDataStoreKey{DocShortID: best.docShortID},
		foundKey: true,
	}, nil
}

func (iter *bm25IndexIterator) Close() error {
	iter.scored = nil
	return nil
}

// readUvarints reads count uvarints from the value stored at key, returning nil when the key does
// not exist.
func readUvarints(
	ctx context.Context,
	store datastore.Keyedstore,
	key datastore.Key,
	count int,
) ([]uint64, error) {
	data, err := store.Get(ctx, key)
	if err != nil {
		if errors.Is(err, corekv.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	values := make([]uint64, count)
	for i := range values {
		data, values[i], err = encoding.DecodeUvarintAscending(data)
		if err != nil {
			return nil, err
		}
	}
	return values, nil
}

// createBM25IndexIterator returns the iterator scoring documents against rank's query.
//
// Unlike a candidate-generating kind this never declines the index and falls back to a scan: the
// score is only available here, and a scan would silently report every document as scoring zero.
func (f *indexFetcher) createBM25IndexIterator(rank *Rank) (indexIterator, error) {
	if rank == nil {
		// The index was selected for something other than a _bm25 field, which nothing in the
		// planner does. Fall back to a scan rather than read entries as if they held values.
		return nil, nil
	}

	baseKey, err := f.newIndexDataStoreKey()
	if err != nil {
		return nil, err
	}

	// Tokenized with the function the indexed values were written with. Anything else silently
	// matches nothing.
	terms := slices.Sorted(maps.Keys(core.Tokens(rank.Query)))

	k1, _ := core.BM25Option(f.indexDesc.Options, core.BM25OptionK1, core.BM25DefaultK1)
	b, _ := core.BM25Option(f.indexDesc.Options, core.BM25OptionB, core.BM25DefaultB)

	return &bm25IndexIterator{
		key:       baseKey,
		terms:     terms,
		k1:        k1,
		b:         b,
		rank:      rank,
		indexName: f.indexDesc.Name,
		execInfo:  f.execInfo,
	}, nil
}
