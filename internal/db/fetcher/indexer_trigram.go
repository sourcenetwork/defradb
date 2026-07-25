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

	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/connor"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

// trigramQueryFromFilter returns the trigram query that every document passing conditions
// must also satisfy, considering only conditions on the field at fieldIndex.
//
// Anything it does not understand becomes trigramAll, "every document is a candidate". That
// is the safe direction: under an AND it simply drops out, and under an OR it takes the
// whole branch with it, which is what an unusable branch must do. It is also how _not is
// handled, since a candidate set cannot express "every document that does not match".
func trigramQueryFromFilter(conditions map[connor.FilterKey]any, fieldIndex int) *trigramQuery {
	q := allQuery
	for key, val := range conditions {
		q = q.and(trigramQueryFromCondition(key, val, fieldIndex))
	}
	return q
}

func trigramQueryFromCondition(key connor.FilterKey, val any, fieldIndex int) *trigramQuery {
	switch k := key.(type) {
	case *mapper.PropertyIndex:
		if k.Index != fieldIndex {
			// A condition on another field says nothing about this one.
			return allQuery
		}
		condMap, ok := val.(map[connor.FilterKey]any)
		if !ok {
			return allQuery
		}
		return trigramQueryFromFieldOperators(condMap)

	case *mapper.Operator:
		branches, ok := val.([]any)
		if !ok {
			return allQuery
		}
		switch k.Operation {
		case opAnd:
			return foldTrigramBranches(branches, fieldIndex, allQuery, (*trigramQuery).and)
		case opOr:
			return foldTrigramBranches(branches, fieldIndex, noneQuery, (*trigramQuery).or)
		}
	}
	return allQuery
}

func foldTrigramBranches(
	branches []any,
	fieldIndex int,
	zero *trigramQuery,
	combine func(*trigramQuery, *trigramQuery) *trigramQuery,
) *trigramQuery {
	q := zero
	for _, branch := range branches {
		branchMap, ok := branch.(map[connor.FilterKey]any)
		if !ok {
			return allQuery
		}
		q = combine(q, trigramQueryFromFilter(branchMap, fieldIndex))
	}
	return q
}

// trigramQueryFromFieldOperators converts the operators applied directly to the indexed field.
// Several operators on one field are all required, so they are ANDed together.
func trigramQueryFromFieldOperators(condMap map[connor.FilterKey]any) *trigramQuery {
	q := allQuery
	for key, val := range condMap {
		op, ok := key.(*mapper.Operator)
		if !ok {
			return allQuery
		}
		pattern, ok := val.(string)
		if !ok {
			return allQuery
		}
		switch op.Operation {
		case opLike, opILike:
			q = q.and(likeTrigramQuery(pattern))
		case opRegex:
			q = q.and(regexTrigramQuery(pattern))
		default:
			return allQuery
		}
	}
	return q
}

// isEvaluable reports whether q can be walked as a set of candidate documents.
//
// A node holding neither a trigram nor a sub-node constrains nothing: as an AND it matches
// every document and as an OR it matches none, and the first of those is not a candidate set.
// Rather than tell them apart, a query containing one falls back to a scan, which is the
// answer for trigramAll as well.
func (q *trigramQuery) isEvaluable() bool {
	if q.op == trigramNone {
		return true
	}
	if len(q.trigram) == 0 && len(q.sub) == 0 {
		return false
	}
	for _, sub := range q.sub {
		if !sub.isEvaluable() {
			return false
		}
	}
	return true
}

// trigramCursor walks the document short IDs satisfying one node of a trigram query, in
// ascending order.
type trigramCursor interface {
	// seek returns the smallest document short ID at or after from that satisfies the node,
	// and false once there are none left.
	seek(from uint64) (uint64, bool, error)
}

// trigramPostingCursor walks the postings of one trigram. They are the keys under the
// trigram's prefix, which the store already holds sorted by document short ID, so this is a
// plain seek into that range and never reads the range into memory.
type trigramPostingCursor struct {
	iter     corekv.Iterator
	prefix   []byte
	execInfo *ExecInfo
	done     bool
}

var _ trigramCursor = (*trigramPostingCursor)(nil)

func (c *trigramPostingCursor) seek(from uint64) (uint64, bool, error) {
	if c.done {
		return 0, false, nil
	}

	target := append(append([]byte{}, c.prefix...), '/')
	target = append(target, keys.EncodeDocShortID(from)...)

	found, err := c.iter.Seek(target)
	if err != nil {
		return 0, false, err
	}
	if !found {
		c.done = true
		return 0, false, nil
	}

	key := c.iter.Key()
	if len(key) <= len(c.prefix)+1 {
		return 0, false, keys.ErrInvalidKey
	}
	docShortID, err := keys.DecodeDocShortID(key[len(c.prefix)+1:])
	if err != nil {
		return 0, false, err
	}

	c.execInfo.IndexesFetched++
	return docShortID, true, nil
}

// trigramAndCursor intersects its sub-cursors by leapfrog: every sub-cursor is asked for the
// current candidate in turn, and whenever one reports a larger ID that becomes the candidate
// and the others have to catch up.
//
// This needs no posting list lengths. A sub-cursor over a short list skips the others ahead
// in one step regardless of the order they are visited in, which is the same effect as
// driving the intersection from the shortest list.
type trigramAndCursor struct {
	subs []trigramCursor
}

var _ trigramCursor = (*trigramAndCursor)(nil)

func (c *trigramAndCursor) seek(from uint64) (uint64, bool, error) {
	candidate := from
	matched := 0
	for i := 0; matched < len(c.subs); i = (i + 1) % len(c.subs) {
		got, found, err := c.subs[i].seek(candidate)
		if err != nil || !found {
			return 0, false, err
		}
		if got == candidate {
			matched++
			continue
		}
		candidate = got
		matched = 1
	}
	return candidate, true, nil
}

// trigramOrCursor merges its sub-cursors, yielding the smallest ID any of them can reach.
type trigramOrCursor struct {
	subs []trigramCursor
}

var _ trigramCursor = (*trigramOrCursor)(nil)

func (c *trigramOrCursor) seek(from uint64) (uint64, bool, error) {
	var best uint64
	found := false
	for _, sub := range c.subs {
		got, subFound, err := sub.seek(from)
		if err != nil {
			return 0, false, err
		}
		if subFound && (!found || got < best) {
			best = got
			found = true
		}
	}
	return best, found, nil
}

// trigramIndexIterator produces the candidate documents of a boolean trigram query.
//
// It over-returns by design: it knows only which trigrams a document holds, not in what order
// or how often, and it does not look at the value at all. filteredFetcher re-evaluates the
// whole original filter on every document it yields, so the result set is decided there.
type trigramIndexIterator struct {
	query      *trigramQuery
	baseKey    keys.IndexDataStoreKey
	descending bool
	indexName  string
	execInfo   *ExecInfo

	// cursor is nil when the query provably matches nothing.
	cursor trigramCursor
	// postings holds every store iterator opened by Init, so Close can release them all.
	postings []corekv.Iterator
	// nextID is the lowest document short ID the next call to Next may return.
	nextID uint64
}

var _ indexIterator = (*trigramIndexIterator)(nil)

func (iter *trigramIndexIterator) Init(ctx context.Context, store datastore.Keyedstore) error {
	if err := iter.Close(); err != nil {
		return err
	}
	iter.cursor = nil
	iter.postings = nil
	iter.nextID = 0

	if iter.query.op == trigramNone {
		return nil
	}

	cursor, err := iter.newCursor(ctx, store, iter.query)
	if err != nil {
		return errors.Join(NewErrCreateIndexIterator(err, iter.indexName), iter.Close())
	}
	iter.cursor = cursor
	return nil
}

// newCursor builds the cursor tree for q, opening one store iterator per trigram.
func (iter *trigramIndexIterator) newCursor(
	ctx context.Context,
	store datastore.Keyedstore,
	q *trigramQuery,
) (trigramCursor, error) {
	subs := make([]trigramCursor, 0, len(q.trigram)+len(q.sub))
	for _, trigram := range q.trigram {
		key := iter.baseKey
		key.Fields = []keys.IndexedField{{
			Value:      client.NewNormalString(trigram),
			Descending: iter.descending,
		}}

		postings, err := store.Iterator(ctx, datastore.IterOptions{Prefix: &key})
		if err != nil {
			return nil, err
		}
		iter.postings = append(iter.postings, postings)
		subs = append(subs, &trigramPostingCursor{
			iter:     postings,
			prefix:   key.Bytes(),
			execInfo: iter.execInfo,
		})
	}

	for _, sub := range q.sub {
		cursor, err := iter.newCursor(ctx, store, sub)
		if err != nil {
			return nil, err
		}
		subs = append(subs, cursor)
	}

	// trigramAll and trigramNone are handled by the caller and never reach a sub-node, since
	// andOr absorbs them rather than nesting them.
	if q.op == trigramOr {
		return &trigramOrCursor{subs: subs}, nil
	}
	return &trigramAndCursor{subs: subs}, nil
}

func (iter *trigramIndexIterator) Next() (indexIterResult, error) {
	if iter.cursor == nil {
		return indexIterResult{}, nil
	}

	docShortID, found, err := iter.cursor.seek(iter.nextID)
	if err != nil {
		return indexIterResult{}, NewErrIterateIndex(err, iter.indexName)
	}
	if !found {
		return indexIterResult{}, nil
	}
	iter.nextID = docShortID + 1

	// The entries hold a trigram of the value rather than the value, so there is nothing to
	// report in the key's fields; the document short ID is the whole result.
	return indexIterResult{
		key:      &keys.IndexDataStoreKey{DocShortID: docShortID},
		foundKey: true,
	}, nil
}

func (iter *trigramIndexIterator) Close() error {
	var err error
	for _, postings := range iter.postings {
		err = errors.Join(err, postings.Close())
	}
	iter.postings = nil
	iter.cursor = nil
	return err
}

// createTrigramIndexIterator returns the iterator for a trigram index over docFilter, or nil
// when the filter yields no usable trigrams and the query should scan instead.
//
// It reads the whole document filter rather than the copy narrowed to the indexed field,
// because that copy drops _or branches on other fields, and an _or missing a branch would
// produce a candidate set missing the documents that branch matches.
func (f *indexFetcher) createTrigramIndexIterator(docFilter *mapper.Filter) (indexIterator, error) {
	if docFilter == nil {
		return nil, nil
	}

	fieldIndex := f.mapping.FirstIndexOfName(f.indexDesc.Fields[0].Name)
	query := trigramQueryFromFilter(docFilter.Conditions, fieldIndex)
	if !query.isEvaluable() {
		return nil, nil
	}

	baseKey, err := f.newIndexDataStoreKey()
	if err != nil {
		return nil, err
	}

	return &trigramIndexIterator{
		query:      query,
		baseKey:    baseKey,
		descending: f.indexDesc.Fields[0].Descending,
		indexName:  f.indexDesc.Name,
		execInfo:   f.execInfo,
	}, nil
}
