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
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

// collectDocsWithOrphans fetches docs from the inverted join and merges orphan docs
// based on sort direction, respecting the limit.
//
// Two strategies are used depending on whether the fetched doc is the primary side
// of the ordering relation (i.e. stores the FK for it):
//
// Doc is primary (stores FK, e.g. Book has _publisherID):
//
//	Orphans are self-identifying via FK IS NULL.
//	ASC: fetch orphans first, fill remaining from join.
//	DESC: fetch from join with limit, fill remaining with orphans.
//
// Doc is secondary (no FK, e.g. Book ordered by author.name, but Author stores _bookID):
//
//	Orphans can only be identified by exclusion after seeing join results.
//	Both ASC and DESC: collect join docs first, then fetch all docs for parent,
//	exclude join doc IDs to find orphans.
func (r *primaryObjectsRetriever) collectDocsWithOrphans(direction mapper.SortDirection) ([]core.Doc, error) {
	limit := r.getLimit()

	if r.orderingRelFieldIsPrimary() {
		if direction == mapper.ASC {
			return r.collectDocsASCWithOrphansByFK(limit)
		}
		return r.collectDocsDESCWithOrphansByFK(limit)
	}
	return r.collectDocsWithOrphansByExclusion(direction, limit)
}

// collectDocsASCWithOrphansByFK fetches orphans first via FK IS NULL (they sort before
// non-null values), then fills remaining slots from the inverted join.
// Only works when the fetched doc is the primary side of the ordering relation
// (i.e. stores the FK).
func (r *primaryObjectsRetriever) collectDocsASCWithOrphansByFK(limit uint64) ([]core.Doc, error) {
	orphans, err := r.fetchOrphanDocsByFK()
	if err != nil {
		return nil, err
	}

	if limit > 0 && uint64(len(orphans)) >= limit {
		return orphans[:limit], nil
	}

	remaining := limit
	if remaining > 0 {
		remaining -= uint64(len(orphans))
		r.setLimit(remaining)
		defer r.setLimit(limit)
	}

	joinDocs, err := r.collectDocs()
	if err != nil {
		return nil, err
	}

	return append(orphans, joinDocs...), nil
}

// collectDocsDESCWithOrphansByFK fetches from the inverted join first (non-null values sort first),
// then fills remaining slots with orphans identified via FK IS NULL.
// Only works when the fetched doc is the primary side of the ordering relation
// (i.e. stores the FK).
func (r *primaryObjectsRetriever) collectDocsDESCWithOrphansByFK(limit uint64) ([]core.Doc, error) {
	joinDocs, err := r.collectDocs()
	if err != nil {
		return nil, err
	}

	if limit > 0 && uint64(len(joinDocs)) >= limit {
		return joinDocs, nil
	}

	orphans, err := r.fetchOrphanDocsByFK()
	if err != nil {
		return nil, err
	}

	if limit > 0 {
		remaining := limit - uint64(len(joinDocs))
		if uint64(len(orphans)) > remaining {
			orphans = orphans[:remaining]
		}
	}

	return append(joinDocs, orphans...), nil
}

// collectDocsWithOrphansByExclusion collects ALL join docs (ignoring limit) to build
// a correct exclusion set, then identifies orphans as docs not in the join results.
// The limit is temporarily removed and re-applied after merging.
//
// This is used when the fetched doc is the secondary side of the ordering relation
// (i.e. does not store the FK for it). In that case, relation fields are populated by
// joins at runtime and are not stored in the datastore, so there is no field on the doc
// that can be checked to determine whether it is an orphan. The only way to identify
// orphans is by exclusion: first see which docs the join produced, then fetch all docs
// and subtract.
func (r *primaryObjectsRetriever) collectDocsWithOrphansByExclusion(
	direction mapper.SortDirection,
	limit uint64,
) ([]core.Doc, error) {
	// Must remove limit to get all join docs for exclusion set.
	if limit > 0 {
		r.setLimit(0)
	}

	joinDocs, err := r.collectDocs()
	if err != nil {
		return nil, err
	}

	if limit > 0 {
		r.setLimit(limit)
	}

	orphans, err := r.fetchOrphanDocsByExclusion(joinDocs)
	if err != nil {
		return nil, err
	}

	if direction == mapper.ASC {
		result := append(orphans, joinDocs...)
		if limit > 0 && uint64(len(result)) > limit {
			result = result[:limit]
		}
		return result, nil
	}

	result := append(joinDocs, orphans...)
	if limit > 0 && uint64(len(result)) > limit {
		result = result[:limit]
	}
	return result, nil
}

// getLimit returns the limit from the primary side's plan, or 0 if no limit.
func (r *primaryObjectsRetriever) getLimit() uint64 {
	if selectTop, ok := r.primarySide.plan.(*selectTopNode); ok {
		if selectTop.limit != nil {
			return selectTop.limit.limit
		}
	}
	return 0
}

// setLimit updates the limit on the primary side's plan.
func (r *primaryObjectsRetriever) setLimit(limit uint64) {
	if selectTop, ok := r.primarySide.plan.(*selectTopNode); ok {
		if selectTop.limit != nil {
			selectTop.limit.limit = limit
		}
	}
}

// isOrderingByRelation returns true if the ordering involves a relation field.
// This is detected by checking if any order condition has more than one field index
// (indicating traversal through a relation).
func (r *primaryObjectsRetriever) isOrderingByRelation() bool {
	for _, order := range r.ordering {
		if len(order.FieldIndexes) > 1 {
			return true
		}
	}
	return false
}

// orderingRelFieldIsPrimary returns true if the fetched doc is the primary side of the
// ordering relation (i.e. stores the FK). When true, orphans can be identified directly
// via FK IS NULL. When false, orphans can only be identified by exclusion from join results.
func (r *primaryObjectsRetriever) orderingRelFieldIsPrimary() bool {
	_, relFieldIndex := r.getOrderingInfo()
	fieldName, ok := r.primaryScan.documentMapping.TryToFindNameFromIndex(relFieldIndex)
	if !ok {
		return false
	}
	fieldDef, ok := r.primarySide.col.Version().GetFieldByName(fieldName)
	if !ok {
		return false
	}
	return fieldDef.IsPrimary
}

// fetchOrphanDocsByFK fetches orphan documents using FK IS NULL on the ordering
// relation's FK field. Only works when the fetched doc is the primary side of the
// ordering relation (e.g., Book has _publisherID).
func (r *primaryObjectsRetriever) fetchOrphanDocsByFK() ([]core.Doc, error) {
	if !r.primarySide.relIDFieldMapIndex.HasValue() {
		return nil, nil
	}

	_, relFieldIndex := r.getOrderingInfo()
	relFieldName, ok := r.primaryScan.documentMapping.TryToFindNameFromIndex(relFieldIndex)
	if !ok {
		return nil, nil
	}

	relIDFieldName := request.ToFieldID(relFieldName)
	relIDFieldMapIndex := r.primaryScan.documentMapping.FirstIndexOfName(relIDFieldName)

	parentFilter := addFilterOnField(r.filter, r.primarySide.relIDFieldMapIndex.Value(),
		r.targetSecondaryDoc.GetID())

	fetcher := newSubQueryFetcher(
		r.primaryScan.p.ctx,
		r.primaryScan.p.identity,
		r.primaryScan.p.nodeACP,
		r.primaryScan.p.documentACP,
		r.primarySide.col,
		r.primaryScan.documentMapping,
		r.primaryScan.p.lensStore,
		r.primaryScan.fields,
		&r.primaryScan.execInfo.fetches,
	)

	return fetcher.fetchOrphans(relIDFieldName, relIDFieldMapIndex, parentFilter)
}

// fetchOrphanDocsByExclusion fetches all docs for the current parent and excludes
// those present in joinDocs. Used when the fetched doc is the secondary side of the
// ordering relation (i.e. does not store the FK for it). In that case, relation fields
// are populated by joins at runtime and not stored in the datastore, so there is no
// field value that can distinguish an orphan from a non-orphan. The exclusion list
// from the join results is the only way to identify them.
func (r *primaryObjectsRetriever) fetchOrphanDocsByExclusion(joinDocs []core.Doc) ([]core.Doc, error) {
	if !r.primarySide.relIDFieldMapIndex.HasValue() {
		return nil, nil
	}

	fetcher := newSubQueryFetcher(
		r.primaryScan.p.ctx,
		r.primaryScan.p.identity,
		r.primaryScan.p.nodeACP,
		r.primaryScan.p.documentACP,
		r.primarySide.col,
		r.primaryScan.documentMapping,
		r.primaryScan.p.lensStore,
		r.primaryScan.fields,
		&r.primaryScan.execInfo.fetches,
	)

	parentFilter := addFilterOnField(r.filter, r.primarySide.relIDFieldMapIndex.Value(),
		r.targetSecondaryDoc.GetID())

	return fetcher.fetchAllExcluding(parentFilter, docsToDocIDs(joinDocs))
}

// getOrderingInfo returns the sort direction and relation field index if the ordering involves a relation field.
func (r *primaryObjectsRetriever) getOrderingInfo() (*mapper.SortDirection, int) {
	for _, order := range r.ordering {
		if len(order.FieldIndexes) > 1 {
			return &order.Direction, order.FieldIndexes[0]
		}
	}
	return nil, 0
}
