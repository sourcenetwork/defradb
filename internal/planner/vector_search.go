// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package planner

import (
	"context"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/fetcher"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/db/vectorindex"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/planner/filter"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

// tryRouteSimilarityToVectorIndex narrows an otherwise-full scan to the k nearest documents when the
// query is a nearest-neighbour search (a single `_similarity` ordered descending, with a limit) and
// the field has a ready vector index. It feeds the graph search results to the scan as document
// prefixes; the similarity/order/limit nodes are left to score, sort and cap as usual. When the query
// does not match, it leaves the full-scan path in place.
//
// A filter can drop some of the k nearest, leaving fewer than k results. The scan handles that by
// widening the search, see [vectorSearch], but only when it can tell which documents passed - the
// cases where it cannot keep the full-scan path.
//
// Routing also changes which documents can be returned at all: only those in the graph. A document
// whose vector field is null or absent was never indexed, so a routed query never returns it, where
// the full-scan path scores it as 0 and can return it among the k.
func (n *selectNode) tryRouteSimilarityToVectorIndex(origScan *scanNode) error {
	if n.selectReq.Limit == nil || n.selectReq.Limit.Limit <= 0 {
		return nil
	}
	// The limit counts groups, not documents, so the k nearest documents are the wrong thing to ask
	// the graph for.
	if n.selectReq.GroupBy != nil {
		return nil
	}
	// Deleted documents are tombstoned out of the graph, so a routed query could never return them.
	if origScan.showDeleted {
		return nil
	}
	// A filter left on the selectNode belongs to a collection with migrations: it is applied after
	// the lens transformation, so documents are dropped above the scan and the scan cannot count how
	// many passed.
	if n.filter != nil {
		return nil
	}

	sim := n.singleSimilarityField()
	if sim == nil {
		return nil
	}
	if !n.isOrderedBySimilarityDesc(sim) {
		return nil
	}

	index, ok := n.readyVectorIndexOnField(sim.SimilarityTarget.Field.Name)
	if !ok {
		return nil
	}
	// readyVectorIndexOnField only returns a vector index, so GetVector always succeeds here.
	vectorDesc, _ := index.GetVector()

	query, ok := similarityQueryVector(sim.Vector)
	if !ok {
		return nil
	}

	// A wrong-length query would be scored on only its shared leading elements, giving wrong results.
	// The full-scan path errors on this; do the same here.
	if dims := int(vectorDesc.Dimensions); dims > 0 && len(query) != dims {
		return NewErrMismatchLengthOnSimilarity(dims, len(query))
	}

	if origScan.filter != nil && !n.canWidenSearchForFilter(origScan) {
		return nil
	}

	// Offset skips the first documents of the result, so the graph must return them too, not just the
	// Limit that remains after skipping. Ask for Limit+Offset; the limit node applies the offset.
	k := int(n.selectReq.Limit.Limit) + int(n.selectReq.Limit.Offset)
	search, err := n.newVectorSearch(index, query, k)
	if err != nil {
		return err
	}
	prefixes, err := search.searchPrefixes(n.planner.ctx, k)
	if err != nil {
		return err
	}

	origScan.Prefixes(prefixes)
	origScan.vectorIndexed = true
	if origScan.filter != nil {
		origScan.vectorSearch = search
	}
	// Empty prefixes would otherwise let the scan fall back to reading the whole collection.
	if len(prefixes) == 0 {
		origScan.noResults = true
	}
	return nil
}

// canWidenSearchForFilter reports whether the scan can widen the graph search to backfill documents
// its filter drops. It cannot when the filter conditions are not all resolved at the scan, and it
// should not when the filter has an index of its own.
//
// There are no collection statistics to weigh a selective filter against the cost of widening, so
// this stands in for that decision, as noted in
// https://github.com/sourcenetwork/defradb/issues/5071. Statistics would also enable seeding the
// first search near the right k and crossing over to a plain scan when widening approaches the
// collection size; both wait on measurements showing they matter.
func (n *selectNode) canWidenSearchForFilter(origScan *scanNode) bool {
	// Every condition has to be one the fetcher applies to the document, because the scan counts the
	// documents that passed it. Conditions on a related object are split out to a join later on, so
	// they drop documents above the scan.
	for _, prop := range filter.ExtractProperties(origScan.filter.Conditions) {
		if prop.IsRelation() {
			return false
		}
	}

	// A condition on a rendered field - an aggregate, or the similarity itself - is likewise split
	// out to the node that computes the field. initFields does the splitting, after this runs, so
	// look for those conditions the same way the split does.
	for _, field := range n.selectReq.Fields {
		var rendered mapper.Field
		switch f := field.(type) {
		case *mapper.Aggregate:
			rendered = f.Field
		case *mapper.Similarity:
			rendered = f.Field
		default:
			continue
		}
		if filter.CopyField(origScan.filter, rendered) != nil {
			return false
		}
	}

	// The user indexed a filtered field, so resolving the filter through that index and scoring the
	// documents it yields is the better plan. It also keeps the fetcher from being given both a
	// secondary index and vector-search prefixes, which it cannot serve at once.
	result := selectIndex(selectIndexOptions{
		collection: origScan.col,
		filter:     origScan.filter,
		ordering:   origScan.ordering,
		docMapping: origScan.documentMapping,
	})
	return result.source != indexSourceFilter
}

// vectorSearch widens a nearest-neighbour search on demand. A filter can drop documents from the k
// nearest, so the scan searches the graph again with a larger k until enough documents have passed
// the filter or the graph has nothing more to give:
// https://github.com/sourcenetwork/defradb/issues/5071
//
// The result is the best k of the candidates the graph search discovered: every candidate is scored
// exactly and sorted downstream, but a matching document the graph search missed stays missing, the
// same as in an unfiltered search. Ranking is exact over the candidates; recall is the graph's.
//
// Each document is fetched at most once however wide the search grows, so document fetches are
// bounded by the full scan these queries ran before this existed. The graph searches themselves
// repeat, but doubling k keeps the candidates requested by all earlier rounds under what the final
// round requests, and the engine widens its traversal along with k (efSearch is raised to k), so
// each round is a genuinely broader search rather than a longer read of the same one.
type vectorSearch struct {
	collectionShortID uint32
	indexID           uint32
	epoch             uint32
	vectorDesc        client.VectorIndexDescription
	query             []float32

	// target is how many documents must pass the filter (limit+offset) before the search can stop
	// widening.
	target int
	// yielded counts the documents the scan has returned, all of which passed the filter.
	yielded int
	// searched is the k of the last graph search.
	searched int
	// seen holds the documents already handed to the fetcher, so a wider search does not fetch any
	// of them a second time.
	seen map[string]struct{}
	// exhausted is set once the graph has no more documents to give.
	exhausted bool
}

// newVectorSearch reads the index state a search needs. It is read once, at plan time, so every
// round of the search sees the same index.
func (n *selectNode) newVectorSearch(
	index client.IndexDescription,
	query []float32,
	target int,
) (*vectorSearch, error) {
	// This is only reached for a vector index (the caller already checked), so GetVector always succeeds.
	vectorDesc, _ := index.GetVector()

	collectionShortID, err := id.GetCollectionShortID(n.planner.ctx, n.collection.Version().CollectionID)
	if err != nil {
		return nil, err
	}

	epoch, err := fetcher.ReadIndexEpoch(
		n.planner.ctx,
		datastore.CtxMustGetTxn(n.planner.ctx),
		n.collection.Version().CollectionID,
		index.ID,
	)
	if err != nil {
		return nil, err
	}

	return &vectorSearch{
		collectionShortID: collectionShortID,
		indexID:           index.ID,
		epoch:             epoch,
		vectorDesc:        *vectorDesc,
		query:             query,
		target:            target,
		seen:              map[string]struct{}{},
	}, nil
}

// searchPrefixes runs the graph search for the k nearest documents and returns one datastore prefix
// per document that earlier, smaller searches did not already return.
func (s *vectorSearch) searchPrefixes(ctx context.Context, k int) ([]keys.Walkable, error) {
	results, err := vectorindex.Search(
		ctx, s.collectionShortID, s.indexID, s.epoch, s.vectorDesc, s.query, k,
	)
	if err != nil {
		return nil, err
	}
	s.searched = k
	// Fewer hits than asked for means the search reached every document the graph can offer, so a
	// wider one cannot return more. This holds because the engine's search breadth is at least k:
	// returning short means the whole reachable graph was explored, not that a result cap was hit.
	if len(results) < k {
		s.exhausted = true
	}

	prefixes := make([]keys.Walkable, 0, len(results))
	for _, r := range results {
		if _, ok := s.seen[r.DocID]; ok {
			continue
		}
		s.seen[r.DocID] = struct{}{}

		docShortID, found, err := id.GetDocShortID(ctx, s.collectionShortID, r.DocID)
		if err != nil {
			return nil, err
		}
		if !found {
			continue
		}
		prefixes = append(prefixes, keys.DataStoreKey{
			CollectionShortID: s.collectionShortID,
			DocShortID:        docShortID,
		})
	}
	return prefixes, nil
}

// singleSimilarityField returns the sole `_similarity` field, or nil if there are none or several:
// with two, which one drives the search is ambiguous, so such a query keeps the full-scan path.
func (n *selectNode) singleSimilarityField() *mapper.Similarity {
	var found *mapper.Similarity
	for _, field := range n.selectReq.Fields {
		if sim, ok := field.(*mapper.Similarity); ok {
			if found != nil {
				return nil
			}
			found = sim
		}
	}
	return found
}

// isOrderedBySimilarityDesc requires ordering by this similarity alone, descending: descending
// because larger cosine means nearer, alone because a second sort key would need documents beyond
// the k the graph returns.
func (n *selectNode) isOrderedBySimilarityDesc(sim *mapper.Similarity) bool {
	if n.selectReq.OrderBy == nil || len(n.selectReq.OrderBy.Conditions) != 1 {
		return false
	}
	cond := n.selectReq.OrderBy.Conditions[0]
	if cond.Direction != mapper.DESC {
		return false
	}
	return len(cond.FieldIndexes) == 1 && cond.FieldIndexes[0] == sim.Field.Index
}

// readyVectorIndexOnField returns the vector index on the field, if any. queryableIndexesOnField has
// already excluded indexes that are still building or have failed, so a returned index is usable.
func (n *selectNode) readyVectorIndexOnField(fieldName string) (client.IndexDescription, bool) {
	for _, idx := range queryableIndexesOnField(n.collection, fieldName) {
		if idx.IsVector() {
			return idx, true
		}
	}
	return client.IndexDescription{}, false
}

func similarityQueryVector(vector any) ([]float32, bool) {
	vec := convertArray[float32](vector)
	if vec == nil {
		return nil, false
	}
	return vec, true
}
