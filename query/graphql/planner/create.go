package planner

import (
	"errors"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/document"
	"github.com/sourcenetwork/defradb/query/graphql/parser"
)

// createNode is used to construct and execute
// an object create mutation.
//
// Create nodes are the simplest of the object mutations
// Each Iteration of the plan, creates and returns one
// document, until we've exhaused the payload. No filtering
// or Select plans
type createNode struct {
	p *Planner

	// cache information about the original data source
	// collection name, meta-data, etc.
	sourceInfo sourceInfo
	collection client.Collection

	// newDoc is the JSON string of the new document, unpares
	newDocStr string
	doc       *document.Document
	// result is the target document as a map after creation
	result map[string]interface{}

	err error

	returned bool
}

func (n *createNode) Init() error { return nil }

func (n *createNode) Start() error {
	// parse the doc
	if n.newDocStr == "" {
		return errors.New("Invalid document to create")
	}

	doc, err := document.NewFromJSON([]byte(n.newDocStr))
	if err != nil {
		n.err = err
		return err
	}
	n.doc = doc
	return nil
}

// Next only returns once.
func (n *createNode) Next() (bool, error) {
	if n.err != nil {
		return false, n.err
	}

	if n.returned {
		return false, nil
	}

	if err := n.collection.WithTxn(n.p.txn).Create(n.doc); err != nil {
		return false, err
	}

	n.returned = true
	return true, nil
}

func (n *createNode) Spans(spans core.Spans) { /* no-op */ }

func (n *createNode) Values() map[string]interface{} {
	val, _ := n.doc.ToMap()
	return val
}

func (n *createNode) Close() { /* no-op?? */ }

func (p *Planner) CreateDoc(parsed *parser.Mutation) (planNode, error) {
	// create a mutation createNode.
	create := &createNode{
		p:         p,
		newDocStr: parsed.Data,
	}

	// get collection
	col, err := p.db.GetCollection(parsed.Schema)
	if err != nil {
		return nil, err
	}
	create.collection = col

	// last step, create a basic Select statement
	// from the parsed Mutation object
	// and construct a new Select planNode
	// which uses the new create node as its
	// source, instead of a scan node.
	slct := parsed.ToSelect()
	return p.SelectFromSource(slct, create)
}
