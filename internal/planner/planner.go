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

	"github.com/sourcenetwork/defradb/acp"
	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/internal/connor"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/planner/filter"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

// planNode is an interface all nodes in the plan tree need to implement.
type planNode interface {
	// Initializes or Re-Initializes an existing planNode, often called internally by Start().
	Init() error

	// Starts any internal logic or processes required by the planNode. Should be called *after* Init().
	Start() error

	// Spans sets the planNodes target spans. This is primarily only used for a scanNode,
	// but based on the tree structure, may need to be propagated Eg. From a selectNode -> scanNode.
	Spans(core.Spans)

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

// Planner combines session state and database state to
// produce a request plan, which is run by the execution context.
type Planner struct {
	txn      datastore.Txn
	identity immutable.Option[acpIdentity.Identity]
	acp      immutable.Option[acp.ACP]
	db       client.Store

	ctx context.Context
}

func New(
	ctx context.Context,
	identity immutable.Option[acpIdentity.Identity],
	acp immutable.Option[acp.ACP],
	db client.Store,
	txn datastore.Txn,
) *Planner {
	return &Planner{
		txn:      txn,
		identity: identity,
		acp:      acp,
		db:       db,
		ctx:      ctx,
	}
}

func (p *Planner) newObjectMutationPlan(stmt *mapper.Mutation) (planNode, error) {
	switch stmt.Type {
	case mapper.CreateObjects:
		return p.CreateDocs(stmt)

	case mapper.UpdateObjects:
		return p.UpdateDocs(stmt)

	case mapper.DeleteObjects:
		return p.DeleteDocs(stmt)

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

	case *createNode:
		return p.expandPlan(n.results, parentPlan)

	case *deleteNode:
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

	// if group
	if plan.group != nil {
		err := p.expandGroupNodePlan(plan)
		if err != nil {
			return err
		}
		plan.planNode = plan.group
	}

	p.expandAggregatePlans(plan)

	// if order
	if plan.order != nil {
		plan.order.plan = plan.planNode
		plan.planNode = plan.order
	}

	if plan.limit != nil {
		p.expandLimitPlan(plan, parentPlan)
	}

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
	switch node := plan.joinPlan.(type) {
	case *typeJoinOne:
		return p.expandTypeJoin(&node.invertibleTypeJoin, parentPlan)
	case *typeJoinMany:
		return p.expandTypeJoin(&node.invertibleTypeJoin, parentPlan)
	}
	return client.NewErrUnhandledType("join plan", plan.joinPlan)
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

func (p *Planner) tryOptimizeJoinDirection(node *invertibleTypeJoin, parentPlan *selectTopNode) error {
	filteredSubFields := findFilteredByRelationFields(
		parentPlan.selectNode.filter.Conditions,
		node.documentMapping,
	)
	slct := node.childSide.plan.(*selectTopNode).selectNode
	desc := slct.collection.Description()
	for subFieldName, subFieldInd := range filteredSubFields {
		indexes := desc.GetIndexesOnField(subFieldName)
		if len(indexes) > 0 && !filter.IsComplex(parentPlan.selectNode.filter) {
			subInd := node.documentMapping.FirstIndexOfName(node.parentSide.relFieldDef.Name)
			relatedField := mapper.Field{Name: node.parentSide.relFieldDef.Name, Index: subInd}
			fieldFilter := filter.UnwrapRelation(filter.CopyField(
				parentPlan.selectNode.filter,
				relatedField,
				mapper.Field{Name: subFieldName, Index: subFieldInd},
			), relatedField)
			// At the moment we just take the first index, but later we want to run some kind of analysis to
			// determine which index is best to use. https://github.com/sourcenetwork/defradb/issues/2680
			err := node.invertJoinDirectionWithIndex(fieldFilter, indexes[0])
			if err != nil {
				return err
			}
			break
		}
	}

	return nil
}

// expandTypeJoin does a plan graph expansion and other optimizations on invertibleTypeJoin.
func (p *Planner) expandTypeJoin(node *invertibleTypeJoin, parentPlan *selectTopNode) error {
	if parentPlan.selectNode.filter == nil {
		return p.expandPlan(node.childSide.plan, parentPlan)
	}

	err := p.tryOptimizeJoinDirection(node, parentPlan)
	if err != nil {
		return err
	}

	return p.expandPlan(node.childSide.plan, parentPlan)
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
	sel *request.Select,
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
	s, err := mapper.ToSelect(p.ctx, p.db, mapper.ObjectSelection, selection)
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
	m, err := mapper.ToOperation(p.ctx, p.db, operation)
	if err != nil {
		return nil, err
	}
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
