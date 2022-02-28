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
	"fmt"
	"reflect"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/query/graphql/parser"
)

// planNode is an interface all nodes in the plan tree need to implement
type planNode interface {
	// Initializes or Re-Initializes an existing planNode
	// Often called internally by Start()
	Init() error

	// Starts any internal logic or processes
	// required by the plan node.
	Start() error

	// Next processes the next result doc from
	// the query. Can only be called *after*
	// Start(). Can't be called again if any
	// previous call returns false.
	Next() (bool, error)

	// Spans sets the planNodes target
	// spans. This is primarily only used
	// for a scanNode, but based on the tree
	// structure, may need to be propagated
	// Eg. From a selectNode -> scanNode.
	Spans(core.Spans)

	// returns the value of the current doc
	// processed by the executor
	Values() map[string]interface{}

	// Source returns the child planNode that
	// generates the source values for this plan.
	// If a plan has no source, return nil
	Source() planNode

	// Close terminates the planNode execution releases its resources.
	Close() error
}

// basic plan Node that implements the planNode interface
// can be added to any struct to turn it into a planNode
type baseNode struct { //nolint:unused
	plan planNode
}

func (n *baseNode) Init() error                    { return n.plan.Init() }   //nolint:unused
func (n *baseNode) Start() error                   { return n.plan.Start() }  //nolint:unused
func (n *baseNode) Next() (bool, error)            { return n.plan.Next() }   //nolint:unused
func (n *baseNode) Spans(spans core.Spans)         { n.plan.Spans(spans) }    //nolint:unused
func (n *baseNode) Values() map[string]interface{} { return n.plan.Values() } //nolint:unused
func (n *baseNode) Close() error                   { return n.plan.Close() }  //nolint:unused
func (n *baseNode) Source() planNode               { return n.plan }          //nolint:unused

type ExecutionContext struct {
	context.Context
}

type PlanContext struct {
	context.Context
}

type Statement struct {
	// Commenting out because unused code (structcheck) according to linter.
	// requestString   string
	// requestDocument *ast.Document parser.Statement -> parser.Query - >
	// requestQuery    parser.Query
}

// Planner combines session state and database state to
// produce a query plan, which is run by the execution context.
type Planner struct {
	txn core.Txn
	db  client.DB

	ctx     context.Context
	evalCtx parser.EvalContext

	// isFinalized bool

}

func makePlanner(ctx context.Context, db client.DB, txn core.Txn) *Planner {
	return &Planner{
		txn: txn,
		db:  db,
		ctx: ctx,
	}
}

func (p *Planner) newPlan(stmt parser.Statement) (planNode, error) {
	switch n := stmt.(type) {
	case *parser.Query:
		if len(n.Queries) > 0 {
			return p.newPlan(n.Queries[0]) // @todo, handle multiple query statements
		} else if len(n.Mutations) > 0 {
			return p.newPlan(n.Mutations[0]) // @todo: handle multiple mutation statements
		} else {
			return nil, fmt.Errorf("Query is missing query or mutation statements")
		}
	case *parser.OperationDefinition:
		if len(n.Selections) == 0 {
			return nil, fmt.Errorf("OperationDefinition is missing selections")
		}
		return p.newPlan(n.Selections[0])
	case *parser.Select:
		return p.Select(n)
	case *parser.CommitSelect:
		return p.CommitSelect(n)
	case *parser.Mutation:
		return p.newObjectMutationPlan(n)
	}
	return nil, fmt.Errorf("unknown statement type %T", stmt)
}

func (p *Planner) newObjectMutationPlan(stmt *parser.Mutation) (planNode, error) {
	switch stmt.Type {
	case parser.CreateObjects:
		return p.CreateDoc(stmt)
	case parser.UpdateObjects:
		return p.UpdateDocs(stmt)
	case parser.DeleteObjects:
		return p.DeleteDocs(stmt)
	default:
		return nil, fmt.Errorf("unknown mutation action %T", stmt.Type)
	}

}

func (p *Planner) makePlan(stmt parser.Statement) (planNode, error) {
	plan, err := p.newPlan(stmt)
	if err != nil {
		return nil, err
	}

	err = p.optimizePlan(plan)
	if err != nil {
		return nil, err
	}

	err = plan.Init()
	return plan, err
}

// plan optimization. Includes plan expansion and wiring
func (p *Planner) optimizePlan(plan planNode) error {
	err := p.expandPlan(plan, nil)
	return err
}

// full plan graph expansion and optimization
func (p *Planner) expandPlan(plan planNode, parentPlan *selectTopNode) error {
	switch n := plan.(type) {
	case *selectTopNode:
		return p.expandSelectTopNodePlan(n, parentPlan)
	case *commitSelectTopNode:
		return p.expandPlan(n.plan, parentPlan)
	case *selectNode:
		return p.expandPlan(n.source, parentPlan)
	case *typeIndexJoin:
		return p.expandTypeIndexJoinPlan(n, parentPlan)
	case *groupNode:
		// We only care about expanding the child source here, it is assumed that the parent source
		// is expanded elsewhere/already
		return p.expandPlan(n.dataSource.childSource, parentPlan)
	case MultiNode:
		return p.expandMultiNode(n, parentPlan)
	case *updateNode:
		return p.expandPlan(n.results, parentPlan)
	default:
		return nil
	}
}

func (p *Planner) expandSelectTopNodePlan(plan *selectTopNode, parentPlan *selectTopNode) error {
	if err := p.expandPlan(plan.source, plan); err != nil {
		return err
	}

	// wire up source to plan
	plan.plan = plan.source

	// if group
	if plan.group != nil {
		err := p.expandGroupNodePlan(plan)
		if err != nil {
			return err
		}
		plan.plan = plan.group
	}

	p.expandAggregatePlans(plan)

	// if order
	if plan.sort != nil {
		plan.sort.plan = plan.plan
		plan.plan = plan.sort
	}

	if plan.limit != nil {
		err := p.expandLimitPlan(plan, parentPlan)
		if err != nil {
			return err
		}
	}

	// wire up the render plan
	if plan.render != nil {
		plan.render.plan = plan.plan
		plan.plan = plan.render
	}

	return nil
}

type aggregateNode interface {
	planNode
	SetPlan(plan planNode)
}

func (p *Planner) expandAggregatePlans(plan *selectTopNode) {
	for _, aggregate := range plan.aggregates {
		aggregate.SetPlan(plan.plan)
		plan.plan = aggregate
	}
}

func (p *Planner) expandMultiNode(plan MultiNode, parentPlan *selectTopNode) error {
	for _, child := range plan.Children() {
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
	return fmt.Errorf("Unknown type index join plan")
}

func (p *Planner) expandGroupNodePlan(plan *selectTopNode) error {
	var childSource planNode
	// Find the first scan node in the plan, we assume that it will be for the correct collection
	scanNode := p.walkAndFindPlanType(plan.plan, &scanNode{}).(*scanNode)
	// Check for any existing pipe nodes in the plan, we should use it if there is one
	pipe, hasPipe := p.walkAndFindPlanType(plan.plan, &pipeNode{}).(*pipeNode)

	if !hasPipe {
		newPipeNode := newPipeNode()
		pipe = &newPipeNode
		pipe.source = scanNode
	}

	if plan.group.childSelect != nil {
		childSelectNode, err := p.SelectFromSource(plan.group.childSelect, pipe, false, &plan.source.(*selectNode).sourceInfo)
		if err != nil {
			return err
		}
		// We need to remove the render so that any child records are preserved on arrival at the parent
		childSelectNode.(*selectTopNode).render = nil

		childSource = childSelectNode
	}

	plan.group.dataSource.childSource = childSource
	plan.group.dataSource.parentSource = plan.plan
	plan.group.dataSource.pipeNode = pipe

	if err := p.walkAndReplacePlan(plan.group, scanNode, pipe); err != nil {
		return err
	}

	return p.expandPlan(childSource, plan)
}

func (p *Planner) expandLimitPlan(plan *selectTopNode, parentPlan *selectTopNode) error {
	switch l := plan.limit.(type) {
	case *hardLimitNode:
		if l == nil {
			return nil
		}

		// if this is a child node, and the parent select has an aggregate then we need to
		// replace the hard limit with a render limit to allow the full set of child records
		// to be aggregated
		if parentPlan != nil && len(parentPlan.aggregates) > 0 {
			renderLimit, err := p.RenderLimit(&parser.Limit{
				Offset: l.offset,
				Limit:  l.limit,
			})
			if err != nil {
				return err
			}
			plan.limit = renderLimit

			renderLimit.plan = plan.plan
			plan.plan = plan.limit
		} else {
			l.plan = plan.plan
			plan.plan = plan.limit
		}
	case *renderLimitNode:
		if l == nil {
			return nil
		}

		l.plan = plan.plan
		plan.plan = plan.limit
	}
	return nil
}

// walkAndReplace walks through the provided plan, and searches for an instance
// of the target plan, and replaces it with the replace plan
func (p *Planner) walkAndReplacePlan(plan, target, replace planNode) error {
	src := plan.Source()
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
	switch node := plan.(type) {
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
		return fmt.Errorf("Unknown plan node type to replace: %T", node)
	}

	return nil
}

// walkAndFindPlanType walks through the plan graph, and returns the first
// instance of a plan, that matches the same type as the target plan
func (p *Planner) walkAndFindPlanType(plan, target planNode) planNode {
	src := plan
	if src == nil {
		return nil
	}

	srcType := reflect.TypeOf(src)
	targetType := reflect.TypeOf(target)
	if srcType != targetType {
		return p.walkAndFindPlanType(plan.Source(), target)
	}

	return src
}

func (p *Planner) queryDocs(query *parser.Query) ([]map[string]interface{}, error) {
	plan, err := p.makePlan(query)
	if err != nil {
		return nil, err
	}

	if err = plan.Start(); err != nil {
		if err2 := (plan.Close()); err2 != nil {
			fmt.Println(err2)
		}
		return nil, err
	}

	var next bool
	if next, err = plan.Next(); err != nil {
		if err2 := (plan.Close()); err2 != nil {
			fmt.Println(err2)
		}
		return nil, err
	}

	if !next {
		return []map[string]interface{}{}, nil
	}

	var docs []map[string]interface{}
	for {
		if values := plan.Values(); values != nil {
			copy := copyMap(values)
			docs = append(docs, copy)
		}

		next, err = plan.Next()
		if err != nil {
			if err2 := (plan.Close()); err2 != nil {
				fmt.Println(err2)
			}
			return nil, err
		}

		if !next {
			break
		}
	}

	err = plan.Close()
	return docs, err
}

func (p *Planner) MakePlan(query *parser.Query) (planNode, error) {
	return p.makePlan(query)
}

func copyMap(m map[string]interface{}) map[string]interface{} {
	cp := make(map[string]interface{})
	for k, v := range m {
		vm, ok := v.(map[string]interface{})
		if ok {
			cp[k] = copyMap(vm)
		} else {
			cp[k] = v
		}
	}

	return cp
}
