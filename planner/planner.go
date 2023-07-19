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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/logging"
	"github.com/sourcenetwork/defradb/planner/mapper"
)

var (
	log = logging.MustNewLogger("planner")
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
	txn datastore.Txn
	db  client.Store

	ctx context.Context
}

func New(ctx context.Context, db client.Store, txn datastore.Txn) *Planner {
	return &Planner{
		txn: txn,
		db:  db,
		ctx: ctx,
	}
}

func (p *Planner) newPlan(stmt any) (planNode, error) {
	switch n := stmt.(type) {
	case *request.Request:
		if len(n.Queries) > 0 {
			return p.newPlan(n.Queries[0]) // @todo, handle multiple query operation statements
		} else if len(n.Mutations) > 0 {
			return p.newPlan(n.Mutations[0]) // @todo: handle multiple mutation operation statements
		} else {
			return nil, ErrMissingQueryOrMutation
		}

	case *request.OperationDefinition:
		if len(n.Selections) == 0 {
			return nil, ErrOperationDefinitionMissingSelection
		}
		return p.newPlan(n.Selections[0])

	case *request.Select:
		m, err := mapper.ToSelect(p.ctx, p.txn, n)
		if err != nil {
			return nil, err
		}

		if _, isAgg := request.Aggregates[n.Name]; isAgg {
			// If this Select is an aggregate, then it must be a top-level
			// aggregate and we need to resolve it within the context of a
			// top-level node.
			return p.Top(m)
		}

		return p.Select(m)

	case *request.CommitSelect:
		m, err := mapper.ToCommitSelect(p.ctx, p.txn, n)
		if err != nil {
			return nil, err
		}
		return p.CommitSelect(m)

	case *request.ObjectMutation:
		m, err := mapper.ToMutation(p.ctx, p.txn, n)
		if err != nil {
			return nil, err
		}
		return p.newObjectMutationPlan(m)
	}

	return nil, client.NewErrUnhandledType("statement", stmt)
}

func (p *Planner) newObjectMutationPlan(stmt *mapper.Mutation) (planNode, error) {
	switch stmt.Type {
	case mapper.CreateObjects:
		return p.CreateDoc(stmt)

	case mapper.UpdateObjects:
		return p.UpdateDocs(stmt)

	case mapper.DeleteObjects:
		return p.DeleteDocs(stmt)

	default:
		return nil, client.NewErrUnhandledType("mutation", stmt.Type)
	}
}

// makePlan creates a new plan from the parsed data, optimizes the plan and returns
// it. The caller of makePlan is also responsible of calling Close() on the plan to
// free it's resources.
func (p *Planner) makePlan(stmt any) (planNode, error) {
	planNode, err := p.newPlan(stmt)
	if err != nil {
		return nil, err
	}

	err = p.optimizePlan(planNode)
	if err != nil {
		return nil, err
	}

	return planNode, err
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

func (p *Planner) expandTypeIndexJoinPlan(plan *typeIndexJoin, parentPlan *selectTopNode) error {
	switch node := plan.joinPlan.(type) {
	case *typeJoinOne:
		return p.expandPlan(node.subType, parentPlan)
	case *typeJoinMany:
		return p.expandPlan(node.subType, parentPlan)
	}
	return client.NewErrUnhandledType("join plan", plan.joinPlan)
}

func (p *Planner) expandGroupNodePlan(topNodeSelect *selectTopNode) error {
	var sourceNode planNode
	var hasScanNode bool
	// Find the first scan node in the topNodeSelect, we assume that it will be for the correct collection.
	// This may be a commit node.
	sourceNode, hasScanNode = walkAndFindPlanType[*scanNode](topNodeSelect.planNode)
	if !hasScanNode {
		commitNode, hasCommitNode := walkAndFindPlanType[*dagScanNode](topNodeSelect.planNode)
		if !hasCommitNode {
			return ErrFailedToFindGroupSource
		}
		sourceNode = commitNode
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
			&topNodeSelect.selectNode.sourceInfo,
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
		node.root = replace
	case *typeJoinMany:
		node.root = replace
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
	ctx context.Context,
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

// RunRequest classifies the type of request to run, runs it, and then returns the result(s).
func (p *Planner) RunRequest(
	ctx context.Context,
	req *request.Request,
) (result []map[string]any, err error) {
	planNode, err := p.makePlan(req)
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
	return p.executeRequest(ctx, planNode)
}

// RunSubscriptionRequest plans a request specific to a subscription and returns the result.
func (p *Planner) RunSubscriptionRequest(
	ctx context.Context,
	request *request.Select,
) (result []map[string]any, err error) {
	planNode, err := p.makePlan(request)
	if err != nil {
		return nil, err
	}

	defer func() {
		if e := planNode.Close(); e != nil {
			err = NewErrFailedToClosePlan(e, "running subscription request")
		}
	}()

	err = planNode.Init()
	if err != nil {
		return nil, err
	}

	data, err := p.executeRequest(ctx, planNode)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MakePlan makes a plan from the parsed request.
//
// Note: Caller is responsible to call the `Close()` method to free the allocated
// resources of the returned plan.
//
// @TODO {defradb/issues/368}: Test this exported function.
func (p *Planner) MakePlan(request *request.Request) (planNode, error) {
	return p.makePlan(request)
}
