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
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/connor"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/db/fetcher"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
	"github.com/sourcenetwork/immutable"
)

// queryableIndexesProvider is implemented by collection types that track
// index lifecycle status. Indexes returned are safe to plan queries against.
type queryableIndexesProvider interface {
	QueryableIndexes() []client.IndexDescription
}

// queryableIndexes returns the indexes usable for planning on col. Collections
// that do not track status expose all their indexes (legacy behavior).
func queryableIndexes(col client.Collection) []client.IndexDescription {
	if p, ok := col.(queryableIndexesProvider); ok {
		return p.QueryableIndexes()
	}
	return col.Version().Indexes
}

// queryableIndexesOnField mirrors CollectionVersion.GetIndexesOnField over queryableIndexes:
// it returns only ready indexes whose first field matches fieldName.
func queryableIndexesOnField(col client.Collection, fieldName string) []client.IndexDescription {
	var result []client.IndexDescription
	for _, idx := range queryableIndexes(col) {
		if len(idx.Fields) > 0 && idx.Fields[0].Name == fieldName {
			result = append(result, idx)
		}
	}
	return result
}

// opRelatedFilter stands in for a condition that is not applied to the field itself but to a
// field of the collection it relates to, and so has no operator of its own on this field.
const opRelatedFilter = ""

// indexKindServesOperator reports whether an index of the given kind can produce candidate
// documents for a filter condition using op.
//
// The answer is what the kind's read path implements, not a list of kinds known to exist. A
// kind that stores entries but has no read path yet therefore serves nothing and is never
// selected, which is what stops a query silently reading entries it cannot interpret.
func indexKindServesOperator(kind string, op string) bool {
	switch kind {
	case "":
		// The ordered key index turns a condition into a range of keys. A regular expression
		// is not a range of keys, so it is the one operator on a string this kind cannot
		// serve, and it falls back to a full scan. It is also the only kind selected for a
		// filter reached through a relation, as it was before there were other kinds.
		return op != connor.RegexOp
	case client.IndexKindTrigram:
		// A trigram index stores fragments of the value rather than the value, so it can
		// generate candidates only for the operators that match somewhere inside a value.
		// The negated forms are excluded because a candidate set cannot express "every
		// document that does not match".
		return op == connor.LikeOp || op == connor.CaseInsensitiveLikeOp || op == connor.RegexOp
	default:
		return false
	}
}

// traverseFilterFieldOperators calls f once per leaf condition in an external filter, with
// the name of the top-level field the condition sits under and the operator applied to that
// field. A condition on a field of a related collection reports opRelatedFilter, since its
// operator applies to the related field and not to this one.
//
// Conditions under _not are skipped, following what the index fetcher already does with
// them: an index produces candidate documents, and no candidate set can express "every
// document that does not match".
func traverseFilterFieldOperators(field string, conditions any, f func(field, op string)) {
	switch t := conditions.(type) {
	case map[string]any:
		for key, val := range t {
			switch {
			case !isFilterOperator(key):
				if field == "" {
					traverseFilterFieldOperators(key, val, f)
				} else {
					f(field, opRelatedFilter)
				}
			case key == request.FilterOpNot:
				// Skipped, see the comment above.
			case key == request.FilterOpAnd || key == request.FilterOpOr || key == request.AliasFieldName:
				traverseFilterFieldOperators(field, val, f)
			case field != "":
				f(field, key)
			}
		}
	case []any:
		for _, val := range t {
			traverseFilterFieldOperators(field, val, f)
		}
	}
}

// isFilterOperator reports whether a key in an external filter is an operator rather than a
// field name. It matches what filter.TraverseFields treats as an operator.
func isFilterOperator(key string) bool {
	if len(key) == 0 || key[0] != '_' || key == request.DocIDFieldName {
		return false
	}
	_, isRelated := request.ToRelatedObjectName(key)
	return !isRelated
}

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
//
// A candidate has to be able to serve the operator its field is filtered by, not merely be
// an index on that field, since which operators a kind can answer differs by kind.
func findIndexByFilter(
	col client.Collection,
	filterConditions map[string]any,
) immutable.Option[client.IndexDescription] {
	if filterConditions == nil {
		return immutable.None[client.IndexDescription]()
	}

	var indexCandidates []client.IndexDescription
	colVersion := col.Version()

	traverseFilterFieldOperators("", filterConditions, func(fieldName, op string) {
		for _, field := range colVersion.Fields {
			if field.Name != fieldName {
				continue
			}
			for _, idx := range queryableIndexesOnField(col, fieldName) {
				if indexKindServesOperator(idx.Kind, op) {
					indexCandidates = append(indexCandidates, idx)
				}
			}
		}
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
		for _, idx := range queryableIndexesOnField(col, field.Name) {
			// The relation ID is looked up by value, so only a kind that can serve _eq is a
			// candidate.
			if indexKindServesOperator(idx.Kind, connor.EqualOp) {
				return immutable.Some(idx)
			}
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

	for _, idx := range queryableIndexes(col) {
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
