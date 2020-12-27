package planner

import (
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/query/graphql/parser"
)

// limit the results
// @todo: Handle cursor
type limitNode struct {
	p    *Planner
	plan planNode

	limit    int64
	offset   int64
	rowIndex int64
}

// Limit creates a new limitNode initalized from
// the parser.Limit object.
func (p *Planner) Limit(n *parser.Limit) (*limitNode, error) {
	if n == nil {
		return nil, nil // nothing to do
	}
	return &limitNode{
		p:        p,
		limit:    n.Limit,
		offset:   n.Offset,
		rowIndex: 0,
	}, nil
}

func (n *limitNode) Init() error {
	n.rowIndex = 0
	return n.plan.Init()
}

func (n *limitNode) Start() error                   { return n.plan.Start() }
func (n *limitNode) Spans(spans core.Spans)         { n.plan.Spans(spans) }
func (n *limitNode) Close()                         { n.plan.Close() }
func (n *limitNode) Values() map[string]interface{} { return n.plan.Values() }

func (n *limitNode) Next() (bool, error) {
	// check if we're passed the limit
	if n.rowIndex-n.offset >= n.limit {
		return false, nil
	}

	for {
		// get next
		if next, err := n.plan.Next(); !next {
			return false, err
		}

		// check if we're beyond the offset
		n.rowIndex++
		if n.rowIndex > n.offset {
			break
		}
	}

	return true, nil
}
