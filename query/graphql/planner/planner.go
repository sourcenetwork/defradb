// Copyright 2020 Source Inc.
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

	"errors"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/query/graphql/parser"

	"github.com/graphql-go/graphql/language/ast"
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
	// structure, may need to be propogated
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
	Close()
}

// basic plan Node that implements the planNode interface
// can be added to any struct to turn it into a planNode
type baseNode struct {
	plan planNode
}

func (n *baseNode) Init() error                    { return n.plan.Init() }
func (n *baseNode) Start() error                   { return n.plan.Start() }
func (n *baseNode) Next() (bool, error)            { return n.plan.Next() }
func (n *baseNode) Spans(spans core.Spans)         { n.plan.Spans(spans) }
func (n *baseNode) Values() map[string]interface{} { return n.plan.Values() }
func (n *baseNode) Close()                         { n.plan.Close() }
func (n *baseNode) Source() planNode               { return n.plan }

type ExecutionContext struct {
	context.Context
}

type PlanContext struct {
	context.Context
}

type Statement struct {
	requestString   string
	requestDocument *ast.Document // parser.Statement -> parser.Query - >
	requestQuery    parser.Query
}

// Planner combines session state and databse state to
// produce a query plan, which is run by the exuction context.
type Planner struct {
	txn client.Txn
	db  client.DB

	ctx     context.Context
	evalCtx parser.EvalContext

	// isFinalized bool

}

func makePlanner(ctx context.Context, db client.DB, txn client.Txn) *Planner {
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
			return nil, errors.New("Query is missing query or mutation statements")
		}
	case *parser.OperationDefinition:
		if len(n.Selections) == 0 {
			return nil, errors.New("OperationDefinition is missing selections")
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
	return plan, nil
}

// plan optimization. Includes plan expansion and wiring
func (p *Planner) optimizePlan(plan planNode) error {
	err := p.expandPlan(plan)
	return err
}

// full plan graph expansion and optimization
func (p *Planner) expandPlan(plan planNode) error {
	switch n := plan.(type) {
	case *selectTopNode:
		return p.expandSelectTopNodePlan(n)
	case *commitSelectTopNode:
		return p.expandPlan(n.plan)
	case *selectNode:
		return p.expandPlan(n.source)
	case *typeIndexJoin:
		return p.expandTypeIndexJoinPlan(n)
	case *groupNode:
		// We only care about expanding the child source here, it is assumed that the parent source
		// is expanded elsewhere/already
		return p.expandPlan(n.dataSource.childSource)
	case MultiNode:
		return p.expandMultiNode(n)
	case *updateNode:
		return p.expandPlan(n.results)
	default:
		return nil
	}
}

func (p *Planner) expandSelectTopNodePlan(plan *selectTopNode) error {
	if err := p.expandPlan(plan.source); err != nil {
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

	// wire up the render plan
	if plan.render != nil {
		plan.render.plan = plan.plan
		plan.plan = plan.render
	}

	// if order
	if plan.sort != nil {
		plan.sort.plan = plan.plan
		plan.plan = plan.sort
	}

	if plan.limit != nil {
		plan.limit.plan = plan.plan
		plan.plan = plan.limit
	}

	return nil
}

func (p *Planner) expandMultiNode(plan MultiNode) error {
	for _, child := range plan.Children() {
		if err := p.expandPlan(child); err != nil {
			return err
		}
	}
	return nil
}

// func (p *Planner) expandSelectNodePlan(plan *selectNode) error {
// 	fmt.Println("Expanding select plan")
// 	return p.expandPlan(plan.source)
// }

func (p *Planner) expandTypeIndexJoinPlan(plan *typeIndexJoin) error {
	switch node := plan.joinPlan.(type) {
	case *typeJoinOne:
		return p.expandPlan(node.subType)
	case *typeJoinMany:
		return p.expandPlan(node.subType)
	}
	return errors.New("Unknown type index join plan")
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

	return p.expandPlan(childSource)
}

// func (p *Planner) QueryDocs(query parser.Query) {

// }

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
	plan, err := p.query(query)
	if err != nil {
		return nil, err
	}

	defer plan.Close()
	if err := plan.Start(); err != nil {
		return nil, err
	}

	if next, err := plan.Next(); err != nil || !next {
		return nil, err
	}

	var docs []map[string]interface{}
	for {
		if values := plan.Values(); values != nil {
			copy := copyMap(values)
			docs = append(docs, copy)
		}

		next, err := plan.Next()
		if err != nil {
			return nil, err
		}

		if !next {
			break
		}
	}

	return docs, nil
}

func (p *Planner) query(query *parser.Query) (planNode, error) {
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
