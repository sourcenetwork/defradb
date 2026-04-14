// Copyright 2022 Democratized Data Foundation
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

	"github.com/sourcenetwork/immutable"
	lensStore "github.com/sourcenetwork/lens/host-go/store"

	"github.com/sourcenetwork/defradb/acp/dac"
	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/connor"
	"github.com/sourcenetwork/defradb/internal/core"
	acpDB "github.com/sourcenetwork/defradb/internal/db/acp"
	"github.com/sourcenetwork/defradb/internal/db/description"
	"github.com/sourcenetwork/defradb/internal/db/fetcher"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/planner/filter"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
	"github.com/sourcenetwork/defradb/internal/se"
)

// planNode is an interface all nodes in the plan tree need to implement.
type planNode interface {
	// Initializes or Re-Initializes an existing planNode, often called internally by Start().
	Init() error

	// Starts any internal logic or processes required by the planNode. Should be called *after* Init().
	Start() error

	// Prefixes sets the planNodes target prefixes. This is primarily only used for a scanNode,
	// but based on the tree structure, may need to be propagated Eg. From a selectNode -> scanNode.
	Prefixes([]keys.Walkable)

	// Next processes the next result doc from the request. Can only be called *after* Start().
	// Can't be called again if any previous call returns false.
	Next() (bool, error)

	// Values returns the value of the current doc, should only be called *after* Next().
	Value() core.Doc

	// Source returns the child planNode that generates the source values for this plan.
	// If a plan has no source, nil is returned.
	Source() planNode

	// Kind tells the name of concrete planNode type.
	Kind() string

	DocumentMap() *core.DocumentMapping

	// Close terminates the planNode execution and releases its resources. After this
	// method is called you can only safely call Kind() and Source() methods.
	Close() error
}

type documentIterator struct {
	currentValue core.Doc
}

func (n *documentIterator) Value() core.Doc {
	return n.currentValue
}

type docMapper struct {
	documentMapping *core.DocumentMapping
}

func (d *docMapper) DocumentMap() *core.DocumentMapping {
	return d.documentMapping
}

type ExecutionContext struct {
	context.Context
}

type PlanContext struct {
	context.Context
}

// P2P defines the P2P operations needed by the planner.
type P2P interface {
	// QueryDocIDsWithSETags queries SE artifacts from replicators based on field values.
	QueryDocIDsWithSETags(ctx context.Context, collectionID string, fieldValues []se.FieldValueQuery) ([]string, error)
}

// Planner combines session state and database state to
// produce a request plan, which is run by the execution context.
type Planner struct {
	identity             immutable.Option[acpIdentity.Identity]
	nodeACP              acpDB.NACInfo
	documentACP          immutable.Option[dac.DocumentACP]
	db                   client.TxnStore
	collectionRepository *description.CollectionRepository

	p2p       P2P
	ctx       context.Context
	lensStore lensStore.Store

	// joinExpand holds transient state used only during plan expansion for
	// join optimization and orphan wiring. These fields are set at the start of
	// plan expansion and consumed during the recursive expandPlan walk.
	joinExpand joinExpandState
}

func New(
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
	nodeACP acpDB.NACInfo,
	documentACP immutable.Option[dac.DocumentACP],
	db client.TxnStore,
	p2p P2P,
	lensStore lensStore.Store,
	collectionRepository *description.CollectionRepository,
) *Planner {
	return &Planner{
		identity:             identity,
		nodeACP:              nodeACP,
		documentACP:          documentACP,
		db:                   db,
		p2p:                  p2p,
		lensStore:            lensStore,
		ctx:                  ctx,
		collectionRepository: collectionRepository,
	}
}

func (p *Planner) newObjectMutationPlan(stmt *mapper.Mutation) (planNode, error) {
	switch stmt.Type {
	case mapper.AddObjects:
		return p.AddDocs(stmt)

	case mapper.UpdateObjects:
		return p.UpdateDocs(stmt)

	case mapper.DeleteObjects:
		return p.DeleteDocs(stmt)

	case mapper.UpsertObjects:
		return p.UpsertDocs(stmt)

	default:
		return nil, client.NewErrUnhandledType("mutation", stmt.Type)
	}
}

// optimizePlan optimizes the plan using plan expansion and wiring.
func (p *Planner) optimizePlan(planNode planNode) error {
	err := p.expandPlan(planNode, nil)
	return err
}

// expandPlan does a full plan graph expansion and other optimizations.
func (p *Planner) expandPlan(planNode planNode, parentPlan *selectTopNode) error {
	switch n := planNode.(type) {
	case *selectTopNode:
		return p.expandSelectTopNodePlan(n, parentPlan)

	case *selectNode:
		return p.expandPlan(n.source, parentPlan)

	case *typeIndexJoin:
		return p.expandTypeIndexJoinPlan(n, parentPlan)

	case *groupNode:
		for _, dataSource := range n.dataSources {
			// We only care about expanding the child source here, it is assumed that the parent source
			// is expanded elsewhere/already
			err := p.expandPlan(dataSource.childSource, parentPlan)
			if err != nil {
				return err
			}
		}
		return nil

	case *topLevelNode:
		for _, child := range n.children {
			switch c := child.(type) {
			case *selectTopNode:
				// We only care about expanding the child source here, it is assumed that the parent source
				// is expanded elsewhere/already
				err := p.expandPlan(child, parentPlan)
				if err != nil {
					return err
				}
			case aggregateNode:
				// top-level aggregates use the top-level node as a source
				c.SetPlan(n)
			}
		}
		return nil

	case MultiNode:
		return p.expandMultiNode(n, parentPlan)

	case *updateNode:
		return p.expandPlan(n.results, parentPlan)

	case *addNode:
		return p.expandPlan(n.results, parentPlan)

	case *deleteNode:
		return p.expandPlan(n.source, parentPlan)

	case *upsertNode:
		return p.expandPlan(n.source, parentPlan)

	case *viewNode:
		return p.expandPlan(n.source, parentPlan)

	case *lensNode:
		return p.expandPlan(n.source, parentPlan)

	default:
		return nil
	}
}

func (p *Planner) expandSelectTopNodePlan(plan *selectTopNode, parentPlan *selectTopNode) error {
	if err := p.expandPlan(plan.selectNode, plan); err != nil {
		return err
	}

	// wire up source to plan
	plan.planNode = plan.selectNode

	// The similarity plan need to be expanded before group, order, aggregate and limit or otherwise
	// it wont be taken into consideration if one of them tries to targets it.
	p.expandSimilarityPlans(plan)

	// if group
	if plan.group != nil {
		err := p.expandGroupNodePlan(plan)
		if err != nil {
			return err
		}
		plan.planNode = plan.group
	}

	p.expandAggregatePlans(plan)

	// if we have an index that can take over ordering, we ignore the order node
	if plan.order != nil && !isOrderedByIndex(plan.selectNode.source) {
		plan.order.plan = plan.planNode
		plan.planNode = plan.order
	}

	if plan.limit != nil {
		p.expandLimitPlan(plan, parentPlan)
	}

	// Process deferred orphan wirings now that the full plan chain is built.
	for _, req := range p.joinExpand.pendingOrphanWirings {
		if req.usePointLookup {
			wireSubQueryOrphanPointLookupPipeline(plan, req.join, req.direction)
		} else {
			wireSubQueryOrphanPipeline(plan, req.join, req.direction)
		}
	}
	p.joinExpand.pendingOrphanWirings = nil

	return nil
}

type aggregateNode interface {
	planNode
	SetPlan(plan planNode)
}

func (p *Planner) expandAggregatePlans(plan *selectTopNode) {
	// Iterate through the aggregates backwards to ensure dependencies
	// execute *before* any aggregate dependent on them.
	for i := len(plan.aggregates) - 1; i >= 0; i-- {
		aggregate := plan.aggregates[i]
		aggregate.SetPlan(plan.planNode)
		plan.planNode = aggregate
	}
}

func (p *Planner) expandSimilarityPlans(plan *selectTopNode) {
	for _, sim := range plan.similarity {
		sim.SetPlan(plan.planNode)
		plan.planNode = sim
	}
}

func (p *Planner) expandMultiNode(multiNode MultiNode, parentPlan *selectTopNode) error {
	for _, child := range multiNode.Children() {
		if err := p.expandPlan(child, parentPlan); err != nil {
			return err
		}
	}
	return nil
}

// expandTypeIndexJoinPlan does a plan graph expansion and other optimizations on typeIndexJoin.
func (p *Planner) expandTypeIndexJoinPlan(plan *typeIndexJoin, parentPlan *selectTopNode) error {
	// expandJoin expands the join and wraps it with orphanNode if @exhaustive is set
	// and ordering by relation field is active.
	// For top-level joins, orphanNode wraps the joinPlan directly.
	// For nested joins with FK IS NULL path, orphanNode is wired into the primary side's
	// selectTopNode so limitNode enforces limits naturally via the pipeline.
	expandJoin := func(node planNode, join *invertibleTypeJoin) error {
		plan.join = join
		plan.joinKind = node.Kind()
		orderDir, err := p.expandTypeJoin(join, parentPlan)
		if err != nil {
			return err
		}
		if orderDir.HasValue() && join.exhaustive {
			if !p.joinExpand.inNestedJoin {
				orphan := newOrphanNode(join)
				if join.parentSide.isPrimary() {
					// Primary parent: orphans self-identify via FK IS NULL.
					// Use sequenceNode for clean pipeline composition.
					if orderDir.Value() == mapper.ASC {
						plan.joinPlan = newSequenceNode(orphan, node)
					} else {
						plan.joinPlan = newSequenceNode(node, orphan)
					}
				} else {
					// Secondary parent: orphans identified via point lookups on child's FK index.
					// Wrap join with orphanNode that handles ordering internally.
					plan.joinPlan = newOrphanPointLookupNode(join, node, orderDir.Value())
				}
			} else if parentPlan != nil {
				p.joinExpand.pendingOrphanWirings = append(p.joinExpand.pendingOrphanWirings, &orphanWiringRequest{
					join:           join,
					direction:      orderDir.Value(),
					usePointLookup: !join.parentSide.isPrimary(),
				})
			}
		}
		return nil
	}

	switch node := plan.joinPlan.(type) {
	case *typeJoinOne:
		return expandJoin(node, &node.invertibleTypeJoin)
	case *typeJoinMany:
		return expandJoin(node, &node.invertibleTypeJoin)
	}
	return client.NewErrUnhandledType("join plan", plan.joinPlan)
}

// wireSubQueryOrphanPipeline inserts a sequenceNode with an orphanNode into the
// selectTopNode for nested join orphan handling via FK IS NULL.
// Called after the full plan chain (order, limit) is built.
func wireSubQueryOrphanPipeline(plan *selectTopNode, join *invertibleTypeJoin, direction mapper.SortDirection) {
	orphan := newOrphanNode(join)

	var seq *sequenceNode
	if direction == mapper.ASC {
		if plan.limit != nil {
			seq = newSequenceNode(orphan, plan.limit.plan)
			plan.limit.plan = seq
		} else {
			seq = newSequenceNode(orphan, plan.planNode)
			plan.planNode = seq
		}
	} else {
		if plan.limit != nil {
			seq = newSequenceNode(plan.limit.plan, orphan)
			plan.limit.plan = seq
		} else {
			seq = newSequenceNode(plan.planNode, orphan)
			plan.planNode = seq
		}
	}
}

// wireSubQueryOrphanPointLookupPipeline inserts an orphanWrapperNode
// into the selectTopNode for nested join orphan handling via point lookups.
// It iterates parents and checks each via point lookup on the child's FK index.
// Called after the full plan chain (order, limit) is built.
func wireSubQueryOrphanPointLookupPipeline(
	plan *selectTopNode,
	join *invertibleTypeJoin,
	direction mapper.SortDirection,
) {
	if plan.limit != nil {
		orphan := newOrphanPointLookupNode(join, plan.limit.plan, direction)
		plan.limit.plan = orphan
	} else {
		orphan := newOrphanPointLookupNode(join, plan.planNode, direction)
		plan.planNode = orphan
	}
}

func findFilteredByRelationFields(
	conditions map[connor.FilterKey]any,
	mapping *core.DocumentMapping,
) map[string]int {
	filterProperties := filter.ExtractProperties(conditions)
	filteredSubFields := make(map[string]int)
	for _, prop := range filterProperties {
		if childMapping := mapping.ChildMappings[prop.Index]; childMapping != nil {
			if !prop.IsRelation() {
				continue
			}
			for _, subProp := range prop.Fields {
				for fieldName, indices := range childMapping.IndexesByName {
					if indices[0] == subProp.Index {
						filteredSubFields[fieldName] = subProp.Index
					}
				}
			}
		}
	}
	return filteredSubFields
}

// isOrderedByIndex checks if the plan is ordered by an index.
func isOrderedByIndex(plan planNode) bool {
	var scan *scanNode
	// the typeIndexJoin has 2 scan nodes for every side of the join
	// so we need to make sure we get the scan node that is scheduled first, i.e. more optimal
	typeJoin := getNode[*typeIndexJoin](plan)
	if typeJoin != nil {
		scan = getNode[*scanNode](typeJoin.join.getFirstSide().plan)
	} else {
		scan = getNode[*scanNode](plan)
	}
	if scan == nil || !scan.index.HasValue() {
		return false
	}

	ok, _ := fetcher.CanBeOrderedByIndex(scan.ordering, scan.index.Value(), scan.documentMapping)
	return ok
}

// tryOptimizeJoinDirection tries to optimize the join direction by using a filter or order on the child side.
// Returns the order direction if the join involves a relation ordering, otherwise returns None.
func (p *Planner) tryOptimizeJoinDirection(
	node *invertibleTypeJoin,
	parentPlan *selectTopNode,
) (immutable.Option[mapper.SortDirection], error) {
	if !node.childSide.relFieldDef.HasValue() {
		// If the relation is one sided we cannot invert the join, so return early
		return immutable.None[mapper.SortDirection](), nil
	}
	optimized, err := p.tryOptimizeJoinDirectionByFilter(node, parentPlan)
	if err != nil {
		return immutable.None[mapper.SortDirection](), err
	}
	if !optimized {
		return p.tryOptimizeJoinDirectionByOrder(node, parentPlan)
	}

	// Filter optimization already inverted the join. Check if there's also
	// a relation ordering to get the direction for orphan node wiring.
	if parentPlan.order != nil && len(parentPlan.order.ordering) > 0 {
		name, err := findOrderedByRelationFields(parentPlan.order.ordering[0], node.documentMapping)
		if err != nil {
			return immutable.None[mapper.SortDirection](), err
		}
		if name != "" {
			return immutable.Some(parentPlan.order.ordering[0].Direction), nil
		}
	}

	return immutable.None[mapper.SortDirection](), nil
}

// tryOptimizeJoinDirectionByFilter tries to optimize the join direction by using a filter on the child side.
// If the child side has an index on a field that is filtered on, we can invert the join direction.
// Returns true if the join direction was optimized, false otherwise.
func (p *Planner) tryOptimizeJoinDirectionByFilter(node *invertibleTypeJoin, parentPlan *selectTopNode) (bool, error) {
	if parentPlan.selectNode.filter == nil {
		return false, nil
	}

	filteredSubFields := findFilteredByRelationFields(
		parentPlan.selectNode.filter.Conditions,
		node.documentMapping,
	)

	slct := node.childSide.plan.(*selectTopNode).selectNode
	desc := slct.collection.Version()

	for subFieldName, subFieldInd := range filteredSubFields {
		indexes := desc.GetIndexesOnField(subFieldName)
		if len(indexes) > 0 && !filter.IsComplex(parentPlan.selectNode.filter) {
			subInd := node.documentMapping.FirstIndexOfName(node.parentSide.relFieldDef.Value().Name)
			relatedField := mapper.Field{Name: node.parentSide.relFieldDef.Value().Name, Index: subInd}
			relevantFilter := filter.CopyField(parentPlan.selectNode.filter, relatedField,
				mapper.Field{Name: subFieldName, Index: subFieldInd})

			fieldFilter := extractRelatedSubFilter(relevantFilter, node.parentSide.plan.DocumentMap(), relatedField)
			// At the moment we just take the first index, but later we want to run some kind of analysis to
			// determine which index is best to use. https://github.com/sourcenetwork/defradb/issues/2680
			node.invertJoinDirectionWithIndex(indexes[0], fieldFilter, nil)
			// If there's a sub-filter on the child side, remove the related field condition from
			// the parent filter. This prevents re-evaluation at the parent level which would fail
			// when the sub-filter modifies the child docs (e.g., filtering by model="Galaxy" when
			// parent filter is model="Walkman").
			//
			// Example:
			// 	User(filter: {devices: {model: {_eq: "Walkman"}}}) {
			// 		name
			// 		devices(filter: {model: {_eq: "Galaxy"}}) {
			// 			model
			// 		}
			// 	}
			if node.subFilter != nil {
				filter.RemoveField(parentPlan.selectNode.filter, relatedField)
			}
			return true, nil
		}
	}
	return false, nil
}

// extractRelatedSubFilter extracts the sub filter from the parent filter.
// Returns nil if the relation field doesn't exist in the document map.
func extractRelatedSubFilter(f *mapper.Filter, docMap *core.DocumentMapping, relField mapper.Field) *mapper.Filter {
	// In groupBy queries with GROUP filters, the docMap may not contain the relation field,
	// so we check existence before accessing to avoid a panic.
	indexes, ok := docMap.IndexesByName[relField.Name]
	if !ok {
		return nil
	}
	subInd := indexes[0]
	relatedField := mapper.Field{Name: relField.Name, Index: subInd}
	subFilter := filter.UnwrapRelation(f, relatedField)
	return subFilter
}

// tryOptimizeJoinDirectionByOrder tries to optimize the join direction by using an order on the child side.
// If the child side has an index on a field that is ordered on, we can invert the join direction.
// Returns the order direction if optimized, otherwise returns None.
func (p *Planner) tryOptimizeJoinDirectionByOrder(
	node *invertibleTypeJoin,
	parentPlan *selectTopNode,
) (immutable.Option[mapper.SortDirection], error) {
	if parentPlan.order == nil || len(parentPlan.order.ordering) == 0 {
		return immutable.None[mapper.SortDirection](), nil
	}

	childFieldName, err := findOrderedByRelationFields(parentPlan.order.ordering[0], node.documentMapping)
	if err != nil {
		return immutable.None[mapper.SortDirection](), err
	}

	slct := node.childSide.plan.(*selectTopNode).selectNode
	desc := slct.collection.Version()
	indexes := desc.GetIndexesOnField(childFieldName)

	if len(indexes) == 0 {
		return immutable.None[mapper.SortDirection](), nil
	}

	ordering := parentPlan.order.ordering[0]
	ordering.FieldIndexes = ordering.FieldIndexes[1:]

	node.invertJoinDirectionWithIndex(indexes[0], nil, []mapper.OrderCondition{ordering})
	return immutable.Some(ordering.Direction), nil
}

// findOrderedByRelationFields finds the field that is ordered on in the order condition.
// Returns the field name and an error if the field is not found.
func findOrderedByRelationFields(
	ordering mapper.OrderCondition,
	mapping *core.DocumentMapping,
) (string, error) {
	fieldIndex := ordering.FieldIndexes[0]
	if fieldIndex < len(mapping.ChildMappings) {
		if childMapping := mapping.ChildMappings[fieldIndex]; childMapping != nil {
			// if fieldIndex is from child mapping, then we need to get the sub field index
			// is must exist, otherwise the query would ill-formed
			subFieldIndex := ordering.FieldIndexes[1]
			childFieldName, found := childMapping.TryToFindNameFromIndex(subFieldIndex)
			if !found {
				return "", client.NewErrFieldIndexNotExist(subFieldIndex)
			}
			return childFieldName, nil
		}
	}
	return "", nil
}

// expandTypeJoin does a plan graph expansion and other optimizations on invertibleTypeJoin.
// Returns the order direction if the join was inverted for ordering, otherwise returns None.
func (p *Planner) expandTypeJoin(
	node *invertibleTypeJoin,
	parentPlan *selectTopNode,
) (immutable.Option[mapper.SortDirection], error) {
	orderDir, err := p.tryOptimizeJoinDirection(node, parentPlan)
	if err != nil {
		return immutable.None[mapper.SortDirection](), err
	}

	ensureOrderNodeForRelationIndex(node)

	// Mark that we're now in a nested join context.
	// Any joins inside the child plan should not add orphanNode because they
	// will be iterated via retrievePrimaryDocs which handles orphans correctly.
	oldInNestedJoin := p.joinExpand.inNestedJoin
	p.joinExpand.inNestedJoin = true
	err = p.expandPlan(node.childSide.plan, parentPlan)
	p.joinExpand.inNestedJoin = oldInNestedJoin

	return orderDir, err
}

// ensureOrderNodeForRelationIndex clears the child's index if a relation ID index exists,
// ensuring orderNode is added during plan expansion. Without this, isOrderedByIndex might
// see an ordering index that can satisfy ordering, so orderNode won't be added. But
// retrievePrimaryDocs might use a relation ID index instead, which can't satisfy ordering.
func ensureOrderNodeForRelationIndex(node *invertibleTypeJoin) {
	childTop, ok := node.childSide.plan.(*selectTopNode)
	if !ok || childTop.order == nil {
		return
	}

	if !node.childSide.relFieldDef.HasValue() || !node.childSide.isPrimary() {
		return
	}

	relIDFieldName := request.ToFieldID(node.childSide.relFieldDef.Value().Name)
	relIDIndex := findIndexByFieldName(node.childSide.col, relIDFieldName)
	if !relIDIndex.HasValue() {
		return
	}

	// A relation ID index exists and might be used instead of ordering index.
	// Clear the current index to ensure orderNode is added.
	childScan := getNode[*scanNode](node.childSide.plan)
	if childScan != nil {
		childScan.index = immutable.None[client.IndexDescription]()
	}
}

func (p *Planner) expandGroupNodePlan(topNodeSelect *selectTopNode) error {
	var sourceNode planNode
	var hasJoinNode bool
	// Find the first join, scan, or commit node in the topNodeSelect,
	// we assume that it will be for the correct collection.
	sourceNode, hasJoinNode = walkAndFindPlanType[*typeIndexJoin](topNodeSelect.planNode)
	if !hasJoinNode {
		var hasScanNode bool
		sourceNode, hasScanNode = walkAndFindPlanType[*scanNode](topNodeSelect.planNode)
		if !hasScanNode {
			commitNode, hasCommitNode := walkAndFindPlanType[*dagScanNode](topNodeSelect.planNode)
			if !hasCommitNode {
				return ErrFailedToFindGroupSource
			}
			sourceNode = commitNode
		}
	}

	// Check for any existing pipe nodes in the topNodeSelect, we should use it if there is one
	pipe, hasPipe := walkAndFindPlanType[*pipeNode](topNodeSelect.planNode)

	if !hasPipe {
		newPipeNode := newPipeNode(sourceNode.DocumentMap())
		pipe = &newPipeNode
		pipe.source = sourceNode
	}

	if len(topNodeSelect.group.childSelects) == 0 {
		dataSource := topNodeSelect.group.dataSources[0]
		dataSource.parentSource = topNodeSelect.planNode
		dataSource.pipeNode = pipe
	}

	for i, childSelect := range topNodeSelect.group.childSelects {
		childSelectNode, err := p.SelectFromSource(
			childSelect,
			pipe,
			false,
			topNodeSelect.selectNode.collection,
		)
		if err != nil {
			return err
		}

		dataSource := topNodeSelect.group.dataSources[i]
		dataSource.childSource = childSelectNode
		dataSource.parentSource = topNodeSelect.planNode
		dataSource.pipeNode = pipe
	}

	if err := p.walkAndReplacePlan(topNodeSelect.group, sourceNode, pipe); err != nil {
		return err
	}

	for _, dataSource := range topNodeSelect.group.dataSources {
		err := p.expandPlan(dataSource.childSource, topNodeSelect)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Planner) expandLimitPlan(topNodeSelect *selectTopNode, parentPlan *selectTopNode) {
	if topNodeSelect.limit == nil {
		return
	}

	// Limits get more complicated with groups and have to be handled internally, so we ensure
	// any limit topNodeSelect is disabled here
	if parentPlan != nil && parentPlan.group != nil && len(parentPlan.group.childSelects) != 0 {
		topNodeSelect.limit = nil
		return
	}

	topNodeSelect.limit.plan = topNodeSelect.planNode
	topNodeSelect.planNode = topNodeSelect.limit
}

// walkAndReplace walks through the provided plan, and searches for an instance
// of the target plan, and replaces it with the replace plan
func (p *Planner) walkAndReplacePlan(planNode, target, replace planNode) error {
	src := planNode.Source()
	if src == nil {
		return nil
	}

	// not our target plan
	// walk into the next plan
	if src != target {
		return p.walkAndReplacePlan(src, target, replace)
	}

	// We've found our plan, figure out what type our current plan is
	// and update accordingly
	switch node := planNode.(type) {
	case *selectNode:
		node.source = replace
	case *typeJoinOne:
		node.replaceRoot(replace)
	case *typeJoinMany:
		node.replaceRoot(replace)
	case *pipeNode:
		/* Do nothing - pipe nodes should not be replaced */
	// @todo: add more nodes that apply here
	default:
		return client.NewErrUnhandledType("plan", node)
	}

	return nil
}

// walkAndFindPlanType walks through the plan graph, and returns the first
// instance of a plan, that matches the given type.
func walkAndFindPlanType[T planNode](planNode planNode) (T, bool) {
	src := planNode
	if src == nil {
		var defaultT T
		return defaultT, false
	}

	targetType, isTargetType := src.(T)
	if !isTargetType {
		return walkAndFindPlanType[T](planNode.Source())
	}

	return targetType, true
}

// executeRequest executes the plan graph that represents the request that was made.
func (p *Planner) executeRequest(
	_ context.Context,
	planNode planNode,
) ([]map[string]any, error) {
	if err := planNode.Start(); err != nil {
		return nil, err
	}

	hasNext, err := planNode.Next()
	if err != nil {
		return nil, err
	}

	docs := []map[string]any{}
	docMap := planNode.DocumentMap()

	for hasNext {
		copy := docMap.ToMap(planNode.Value())
		docs = append(docs, copy)

		hasNext, err = planNode.Next()
		if err != nil {
			return nil, err
		}
	}
	return docs, err
}

// RunSelection runs a selection and returns the result(s).
func (p *Planner) RunSelection(
	ctx context.Context,
	sel request.Selection,
) (map[string]any, error) {
	req := &request.Request{
		Queries: []*request.OperationDefinition{{
			Selections: []request.Selection{sel},
		}},
	}
	return p.RunRequest(ctx, req)
}

// RunRequest classifies the type of request to run, runs it, and then returns the result(s).
func (p *Planner) RunRequest(
	ctx context.Context,
	req *request.Request,
) (map[string]any, error) {
	planNode, err := p.MakePlan(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		if e := planNode.Close(); e != nil {
			err = NewErrFailedToClosePlan(e, "running request")
		}
	}()

	err = planNode.Init()
	if err != nil {
		return nil, err
	}

	// Ensure subscription request doesn't ever end up with an explain directive.
	if len(req.Subscription) > 0 && req.Subscription[0].Directives.ExplainType.HasValue() {
		return nil, ErrCantExplainSubscriptionRequest
	}

	if len(req.Queries) > 0 && req.Queries[0].Directives.ExplainType.HasValue() {
		return p.explainRequest(ctx, planNode, req.Queries[0].Directives.ExplainType.Value())
	}

	if len(req.Mutations) > 0 && req.Mutations[0].Directives.ExplainType.HasValue() {
		return p.explainRequest(ctx, planNode, req.Mutations[0].Directives.ExplainType.Value())
	}

	// This won't / should NOT execute if it's any kind of explain request.
	res, err := p.executeRequest(ctx, planNode)
	if err != nil {
		return nil, err
	}

	if len(res) > 0 {
		return res[0], nil
	}

	return nil, nil
}

// MakeSelectionPlan makes a plan for a single selection.
//
// Note: Caller is responsible to call the `Close()` method to free the allocated
// resources of the returned plan.
func (p *Planner) MakeSelectionPlan(selection *request.Select) (planNode, error) {
	s, err := mapper.ToSelect(p.ctx, p.db, p.collectionRepository, mapper.ObjectSelection, selection)
	if err != nil {
		return nil, err
	}
	planNode, err := p.Select(s)
	if err != nil {
		return nil, err
	}
	err = p.optimizePlan(planNode)
	if err != nil {
		return nil, err
	}
	return planNode, err
}

// MakePlan makes a plan from the parsed request.
//
// Note: Caller is responsible to call the `Close()` method to free the allocated
// resources of the returned plan.
//
// @TODO {defradb/issues/368}: Test this exported function.
func (p *Planner) MakePlan(req *request.Request) (planNode, error) {
	// TODO handle multiple operation statements
	// https://github.com/sourcenetwork/defradb/issues/1395
	var operation *request.OperationDefinition
	if len(req.Mutations) > 0 {
		operation = req.Mutations[0]
	} else if len(req.Queries) > 0 {
		operation = req.Queries[0]
	} else {
		return nil, ErrMissingQueryOrMutation
	}
	m, err := mapper.ToOperation(p.ctx, p.db, p.collectionRepository, operation)
	if err != nil {
		return nil, err
	}
	p.joinExpand.exhaustive = m.Exhaustive
	planNode, err := p.Operation(m)
	if err != nil {
		return nil, err
	}
	err = p.optimizePlan(planNode)
	if err != nil {
		return nil, err
	}
	return planNode, err
}
