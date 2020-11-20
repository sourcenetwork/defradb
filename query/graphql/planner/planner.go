package planner

import (
	"context"

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

type PlanContext struct {
	context.Context
	RequestString string
}

type Statement struct {
	requestString   string
	requestDocument *ast.Document // parser.Statement -> parser.Query - >
	requestQuery    parser.Query
}

// Planner combines session state and databse state to
// produce a query plan, which is run by the exuction context.
type Planner struct {
	ctx         PlanContext
	statement   *Statement
	isFinalized bool

	txn core.Txn

	evalCtx parser.EvalContext
}

func NewPlanner() {}

func (p *Planner) newPlan(doc *ast.Document) {}

func (p *Planner) makePlan(doc parser.Statement) {}

// func (p *Planner) Select() {}

func (p *Planner) Limit() {}

func (p *Planner) OrderBy() {}

func (p *Planner) GroupBy() {}

// full plan graph expansion and optimization
func (p *Planner) expandPlan() {}
