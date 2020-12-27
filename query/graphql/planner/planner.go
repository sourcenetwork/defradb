package planner

import (
	"context"

	"github.com/pkg/errors"
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
	txn     core.Txn
	ctx     context.Context
	evalCtx parser.EvalContext

	// isFinalized bool

}

func makePlanner(txn core.Txn) *Planner {
	ctx := context.Background()
	return &Planner{
		txn: txn,
		ctx: ctx,
	}
}

func (p *Planner) newPlan(stmt parser.Statement) (planNode, error) {
	switch n := stmt.(type) {
	case *parser.Query:
		return p.newPlan(n.Queries[0]) // @todo ensure parser.Query as at least 1 query definition
	case *parser.QueryDefinition:
		return p.newPlan(n.Selections[0]) // @todo: ensure parser.QueryDefinition has at least 1 selection
	case *parser.Select:
		return p.Select(n)
	default:
		return nil, errors.Errorf("unknown statement type %T", stmt)
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

func (p *Planner) GroupBy() {}

// plan optimization. Includes plan expansion and wiring
func (p *Planner) optimizePlan(plan planNode) error {
	return p.expandPlan(plan)
}

// full plan graph expansion and optimization
func (p *Planner) expandPlan(plan planNode) error {
	switch n := plan.(type) {
	case *selectTopNode:
		return p.expandSelectTopNodePlan(n)
	default:
		return nil
	}
}

func (p *Planner) expandSelectTopNodePlan(plan *selectTopNode) error {
	// wire up source to plan
	plan.plan = plan.source

	// wire up the render plan
	if plan.render != nil {
		plan.render.plan = plan.plan
		plan.plan = plan.render
	}

	// if group
	// if order

	if plan.limit != nil {
		plan.limit.plan = plan.plan
		plan.plan = plan.limit
	}

	return nil
}

// func (p *Planner) QueryDocs(query parser.Query) {

// }

func (p *Planner) queryDocs(query *parser.Query) ([]map[string]interface{}, error) {
	plan, err := p.query(query)
	if err != nil {
		return nil, err
	}

	defer plan.Close()
	if err := plan.Start(); err != nil {
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
