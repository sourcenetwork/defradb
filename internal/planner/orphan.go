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
	"maps"

	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/connor"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/db/fetcher"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/planner/filter"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

// joinExpandState is transient state for plan expansion, consumed during expandPlan.
type joinExpandState struct {
	// Cached from mapper.Operation.Exhaustive during makePlan for use during plan expansion.
	exhaustive bool

	// Nested joins handle orphans via retrievePrimaryDocs, so we skip orphanNode wiring.
	inNestedJoin bool

	// Deferred until after order/limit nodes are wired in expandSelectTopNodePlan.
	pendingOrphanWirings []*orphanWiringRequest
}

// orphanWiringRequest captures what's needed to wire orphan nodes after the plan chain is built.
type orphanWiringRequest struct {
	join      *invertibleTypeJoin
	direction mapper.SortDirection
	// True when the parent doesn't store the FK (secondary side), so we need
	// point lookups on the child's FK index to identify orphans.
	usePointLookup bool
}

type orphanExecInfo struct {
	iterations uint64
	fetches    fetcher.ExecInfo
}

// orphanNode yields parent documents that have no related children by scanning
// for FK IS NULL via a scanNode clone. Used inside a sequenceNode for primary-side
// parents where the FK field is stored directly.
type orphanNode struct {
	docMapper

	join *invertibleTypeJoin

	// Set by retrievePrimaryDocs for nested join context (per-iteration filter scope).
	subQueryFilter           *mapper.Filter
	subQueryRelIDFieldName   string
	subQueryRelIDFieldMapIdx int
	isSubQuery               bool

	// Streams from a scanNode clone with FK IS NULL filter.
	standaloneScan *scanNode

	execInfo orphanExecInfo
}

func newOrphanNode(join *invertibleTypeJoin) *orphanNode {
	return &orphanNode{
		docMapper: join.docMapper,
		join:      join,
	}
}

// setSubQueryContext configures the orphanNode for subquery use (FK IS NULL path).
// Called by retrievePrimaryDocs before each Init() cycle with per-iteration params.
func (n *orphanNode) setSubQueryContext(filter *mapper.Filter, relIDFieldName string, relIDFieldMapIdx int) {
	n.subQueryFilter = filter
	n.subQueryRelIDFieldName = relIDFieldName
	n.subQueryRelIDFieldMapIdx = relIDFieldMapIdx
	n.isSubQuery = true
}

const orphanNodeKind = "orphanNode"

func (n *orphanNode) Kind() string {
	return orphanNodeKind
}

func (n *orphanNode) Init() error {
	if n.standaloneScan != nil {
		if err := n.standaloneScan.Close(); err != nil {
			return err
		}
		n.standaloneScan = nil
	}
	return n.initStandaloneScan()
}

func (n *orphanNode) Start() error {
	return nil
}

func (n *orphanNode) Prefixes(prefixes []keys.Walkable) {}

func (n *orphanNode) Source() planNode {
	return nil
}

func (n *orphanNode) Close() error {
	if n.standaloneScan != nil {
		return n.standaloneScan.Close()
	}
	return nil
}

func (n *orphanNode) Next() (bool, error) {
	n.execInfo.iterations++
	if n.standaloneScan == nil {
		return false, nil
	}
	return n.standaloneScan.Next()
}

func (n *orphanNode) Value() core.Doc {
	if n.standaloneScan != nil {
		return n.standaloneScan.Value()
	}
	return core.Doc{}
}

// initStandaloneScan creates and initializes a scanNode clone with FK IS NULL filter
// for streaming orphan detection.
func (n *orphanNode) initStandaloneScan() error {
	var orphanFilter *mapper.Filter
	var relationIDFieldName string

	if n.isSubQuery {
		orphanFilter = addNullFilterOnField(n.subQueryFilter, n.subQueryRelIDFieldMapIdx)
		relationIDFieldName = n.subQueryRelIDFieldName
	} else if n.join.parentSide.isPrimary() {
		relIDFieldMapIndex := n.join.parentSide.relIDFieldMapIndex.Value()
		orphanFilter = addNullFilterOnField(n.join.subFilter, relIDFieldMapIndex)
		relationIDFieldName = request.ToFieldID(n.join.parentSide.relFieldDef.Value().Name)
	} else {
		return nil
	}

	result := selectIndex(selectIndexOptions{
		collection:          n.join.parentSide.col,
		filter:              orphanFilter,
		relationIDFieldName: relationIDFieldName,
		docMapping:          n.documentMapping,
	})

	parentScan := getNode[*scanNode](n.join.parentSide.plan)
	n.standaloneScan = parentScan.cloneWithFilter(orphanFilter, result.index)

	return n.standaloneScan.Init()
}

func (n *orphanNode) simpleExplain() (map[string]any, error) {
	return map[string]any{}, nil
}

func (n *orphanNode) Explain(explainType request.ExplainType) (map[string]any, error) {
	switch explainType {
	case request.SimpleExplain:
		return n.simpleExplain()

	case request.ExecuteExplain:
		fetches := n.execInfo.fetches
		if n.standaloneScan != nil {
			fetches.Add(n.standaloneScan.execInfo.fetches)
		}
		return map[string]any{
			"iterations":   n.execInfo.iterations,
			"docFetches":   fetches.DocsFetched,
			"fieldFetches": fetches.FieldsFetched,
			"indexFetches": fetches.IndexesFetched,
		}, nil

	default:
		return nil, ErrUnknownExplainRequestType
	}
}

// addFilterOnField returns a new filter with a condition that checks if the field equals the given value.
// It does not mutate the input filter.
func addFilterOnField(f *mapper.Filter, propIndex int, value any) *mapper.Filter {
	result := mapper.NewFilter()
	if f != nil {
		maps.Copy(result.Conditions, f.Conditions)
		result.ExternalConditions = make(map[string]any, len(f.ExternalConditions))
		maps.Copy(result.ExternalConditions, f.ExternalConditions)
	}

	propertyIndex := &mapper.PropertyIndex{Index: propIndex}
	filterConditions := map[connor.FilterKey]any{
		propertyIndex: map[connor.FilterKey]any{
			mapper.FilterEqOp: value,
		},
	}

	filter.RemoveField(result, mapper.Field{Index: propIndex})
	result.Conditions = filter.MergeConditions(result.Conditions, filterConditions)
	return result
}

// addNullFilterOnField adds a filter condition that checks if the field is NULL.
func addNullFilterOnField(f *mapper.Filter, propIndex int) *mapper.Filter {
	return addFilterOnField(f, propIndex, nil)
}
