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
	"slices"
	"strings"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/db/fetcher"
	"github.com/sourcenetwork/defradb/internal/planner/filter"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
	"github.com/sourcenetwork/immutable"
)

// indexSource indicates what criteria was used to select an index.
type indexSource string

const (
	indexSourceNone       indexSource = "none"
	indexSourceFilter     indexSource = "filter"
	indexSourceRelationID indexSource = "relationID"
	indexSourceOrdering   indexSource = "ordering"
)

// indexSelectionResult contains the result of index selection.
type indexSelectionResult struct {
	// index is the selected index, if any.
	index immutable.Option[client.IndexDescription]
	// canSatisfyOrder indicates whether the selected index can provide the requested ordering.
	canSatisfyOrder bool
	// source indicates what criteria led to this index being selected.
	source indexSource
}

// findIndexByFilter finds an index that can be used for the given filter conditions.
// Returns the first matching index sorted by name for deterministic behavior.
// See https://github.com/sourcenetwork/defradb/issues/2680 for cost-based optimization.
func findIndexByFilter(
	col client.Collection,
	filterConditions map[string]any,
) immutable.Option[client.IndexDescription] {
	if filterConditions == nil {
		return immutable.None[client.IndexDescription]()
	}

	var indexCandidates []client.IndexDescription
	colVersion := col.Version()

	filter.TraverseFields(filterConditions, func(path []string, val any) bool {
		for _, field := range colVersion.Fields {
			if field.Name != path[0] {
				continue
			}
			indexes := colVersion.GetIndexesOnField(field.Name)
			if len(indexes) > 0 {
				indexCandidates = append(indexCandidates, indexes...)
				return true
			}
		}
		return true
	})

	if len(indexCandidates) == 0 {
		return immutable.None[client.IndexDescription]()
	}

	// Sort by name for deterministic selection
	slices.SortFunc(indexCandidates, func(a, b client.IndexDescription) int {
		return strings.Compare(a.Name, b.Name)
	})

	return immutable.Some(indexCandidates[0])
}

// findIndexByFieldName finds an index that starts with the given field name.
// Returns the first matching index if multiple exist.
// See https://github.com/sourcenetwork/defradb/issues/2680 for cost-based optimization.
func findIndexByFieldName(
	col client.Collection,
	fieldName string,
) immutable.Option[client.IndexDescription] {
	colVersion := col.Version()

	for _, field := range colVersion.Fields {
		if field.Name != fieldName {
			continue
		}
		indexes := colVersion.GetIndexesOnField(field.Name)
		if len(indexes) > 0 {
			return immutable.Some(indexes[0])
		}
	}

	return immutable.None[client.IndexDescription]()
}

// findIndexForOrdering finds an index that can satisfy the given ordering conditions.
// It checks all available indexes and returns the first one that can provide the ordering.
func findIndexForOrdering(
	col client.Collection,
	ordering []mapper.OrderCondition,
	docMapping *core.DocumentMapping,
) immutable.Option[client.IndexDescription] {
	if len(ordering) == 0 {
		return immutable.None[client.IndexDescription]()
	}

	indexes := col.Version().Indexes
	for _, idx := range indexes {
		canOrder, _ := fetcher.CanBeOrderedByIndex(ordering, idx, docMapping)
		if canOrder {
			return immutable.Some(idx)
		}
	}

	return immutable.None[client.IndexDescription]()
}

// canIndexSatisfyOrdering checks if the given index can satisfy the ordering conditions.
// Returns true if the index can provide the ordering, false otherwise.
func canIndexSatisfyOrdering(
	index client.IndexDescription,
	ordering []mapper.OrderCondition,
	docMapping *core.DocumentMapping,
) bool {
	if len(ordering) == 0 {
		return true
	}
	canOrder, _ := fetcher.CanBeOrderedByIndex(ordering, index, docMapping)
	return canOrder
}

// selectIndexOptions configures how index selection should be performed.
type selectIndexOptions struct {
	// collection to select index from
	collection client.Collection
	// filter conditions to consider for index selection
	filter *mapper.Filter
	// ordering conditions to consider for index selection
	ordering []mapper.OrderCondition
	// relationIDFieldName is the name of the relation ID field (e.g., "_authorID")
	// If provided, an index on this field will be considered
	relationIDFieldName string
	// docMapping is required for ordering-based index selection
	docMapping *core.DocumentMapping
}

// selectIndex chooses the best available index based on the provided options.
// Priority order: Filter > RelationID > Ordering
//
// This is the main entry point for index selection and should be used instead
// of the individual Find* functions when multiple criteria need to be considered.
func selectIndex(opts selectIndexOptions) indexSelectionResult {
	if opts.filter != nil {
		idx := findIndexByFilter(opts.collection, opts.filter.ExternalConditions)
		if idx.HasValue() {
			canOrder := canIndexSatisfyOrdering(idx.Value(), opts.ordering, opts.docMapping)
			return indexSelectionResult{
				index:           idx,
				canSatisfyOrder: canOrder,
				source:          indexSourceFilter,
			}
		}
	}

	if opts.relationIDFieldName != "" {
		idx := findIndexByFieldName(opts.collection, opts.relationIDFieldName)
		if idx.HasValue() {
			canOrder := canIndexSatisfyOrdering(idx.Value(), opts.ordering, opts.docMapping)
			return indexSelectionResult{
				index:           idx,
				canSatisfyOrder: canOrder,
				source:          indexSourceRelationID,
			}
		}
	}

	if len(opts.ordering) > 0 && opts.docMapping != nil {
		idx := findIndexForOrdering(opts.collection, opts.ordering, opts.docMapping)
		if idx.HasValue() {
			return indexSelectionResult{
				index:           idx,
				canSatisfyOrder: true,
				source:          indexSourceOrdering,
			}
		}
	}

	return indexSelectionResult{
		index:           immutable.None[client.IndexDescription](),
		canSatisfyOrder: false,
		source:          indexSourceNone,
	}
}
