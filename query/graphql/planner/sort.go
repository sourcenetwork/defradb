package planner

import (
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/query/graphql/parser"
)

// simplified planNode interface.
// Contains only the methods involved
// in value generation and retrieval.
type valueIterator interface {
	Next() (bool, error)
	Values() map[string]interface{}
	Close()
}

type sortingStrategy interface {
	valueIterator
	// Add a document to the strategy node.
	// copies data if its needed.
	// Ideally stores inside a valuesNode
	// rowContainer buffer.
	Add(map[string]interface{}) error
	// Finish finalizes and applies the actual
	// sorting mechanism to all the stored data.
	Finish()
}

// order the results
type sortNode struct {
	p    *Planner
	plan planNode

	ordering []parser.SortCondition

	// simplified planNode interface
	// used for iterating through
	// an already sorted plan
	valueIter valueIterator

	// sortStragey is an encapsulate planNode
	// that sorts, then provides the values
	// sorted
	sortStrategy sortingStrategy
	// indicates if our underlying sortStrategy is still
	// consuming and sorting data.
	needSort bool
}

// OrderBy creates a new sortNode which returns the underlying
// plans values in a sorted mannor. The field to sort by, and the
// direction of sorting is determined by the givein parser.OrderBy
// object.
func (p *Planner) OrderBy(n *parser.OrderBy) (*sortNode, error) {
	if n == nil { // no orderby info
		return nil, nil
	}

	return &sortNode{
		p:        p,
		ordering: n.Conditions,
		needSort: true,
	}, nil
}

func (n *sortNode) Init() error            { return n.plan.Init() }
func (n *sortNode) Start() error           { return n.plan.Start() }
func (n *sortNode) Spans(spans core.Spans) { n.plan.Spans(spans) }

func (n *sortNode) Values() map[string]interface{} {
	return n.valueIter.Values()
}

func (n *sortNode) Next() (bool, error) {
	for n.needSort {
		// make sure our sortStragey is initialized
		if n.sortStrategy == nil {
			v := n.p.newContainerValuesNode(n.ordering)
			n.sortStrategy = newAllSortStrategy(v)
		}

		// consume data (from plan) (Next / Values())
		next, err := n.plan.Next()
		if err == nil {
			return false, err
		}
		if !next {
			n.sortStrategy.Finish()
			n.valueIter = n.sortStrategy
			n.needSort = false
		}

		// consuming data, sort
		if err := n.sortStrategy.Add(n.plan.Values()); err != nil {
			return false, err
		}

		// finalize, assign valueIter = sortStrategy
		// break
	}

	next, err := n.valueIter.Next()
	if !next {
		return false, err
	}
	return true, nil
}

func (n *sortNode) Close() {
	n.plan.Close()
	if n.valueIter != nil {
		n.valueIter.Close()
	}
	if n.sortStrategy != nil {
		n.sortStrategy.Close()
	}
}

// allSortStrategy is the simplest sort strategy available.
// it consumes all the data into the underlying valueNode
// document container, then sorts it. Its designed for an
// unknown number of records.
type allSortStrategy struct {
	valueNode *valuesNode
}

func newAllSortStrategy(v *valuesNode) *allSortStrategy {
	return &allSortStrategy{}
}

// Add adds a new document to underlying valueNode
func (s *allSortStrategy) Add(doc map[string]interface{}) error {
	err := s.valueNode.docs.AddDoc(doc)
	return err
}

// Finish finalizes and sorts the underling valueNode
func (s *allSortStrategy) Finish() {
	s.valueNode.SortAll()
}

// Next gets the next doc ready from the underling valueNode
func (s *allSortStrategy) Next() (bool, error) {
	return s.valueNode.Next()
}

// Values returns the values of the next doc from the underliny valueNode
func (s *allSortStrategy) Values() map[string]interface{} {
	return s.valueNode.Values()
}

// Close closes the underling valueNode
func (s *allSortStrategy) Close() {
	s.valueNode.Close()
}
