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
	"github.com/sourcenetwork/defradb/core"
	parserTypes "github.com/sourcenetwork/defradb/query/graphql/parser/types"
)

// simplified planNode interface.
// Contains only the methods involved
// in value generation and retrieval.
type valueIterator interface {
	Next() (bool, error)
	Value() core.Doc
	Close() error
}

type sortingStrategy interface {
	valueIterator
	// Add a document to the strategy node.
	// copies data if its needed.
	// Ideally stores inside a valuesNode
	// rowContainer buffer.
	Add(core.Doc) error
	// Finish finalizes and applies the actual
	// sorting mechanism to all the stored data.
	Finish()
}

// order the results
type sortNode struct {
	p    *Planner
	plan planNode

	ordering []parserTypes.SortCondition

	// simplified planNode interface
	// used for iterating through
	// an already sorted plan
	valueIter valueIterator

	// sortStrategy is an encapsulate planNode
	// that sorts, then provides the values
	// sorted
	sortStrategy sortingStrategy
	// indicates if our underlying sortStrategy is still
	// consuming and sorting data.
	needSort bool
}

// OrderBy creates a new sortNode which returns the underlying
// plans values in a sorted mannor. The field to sort by, and the
// direction of sorting is determined by the given parserTypes.OrderBy
// object.
func (p *Planner) OrderBy(n *parserTypes.OrderBy) (*sortNode, error) {
	if n == nil { // no orderby info
		return nil, nil
	}

	return &sortNode{
		p:        p,
		ordering: n.Conditions,
		needSort: true,
	}, nil
}

func (n *sortNode) Kind() string {
	return "sortNode"
}

func (n *sortNode) Init() error {
	// reset stateful data
	n.needSort = true
	n.sortStrategy = nil
	return n.plan.Init()
}
func (n *sortNode) Start() error { return n.plan.Start() }

func (n *sortNode) Spans(spans core.Spans) { n.plan.Spans(spans) }

func (n *sortNode) Value() core.Doc {
	return n.valueIter.Value()
}

// Explain method returns a map containing all attributes of this node that
// are to be explained, subscribes / opts-in this node to be an explainablePlanNode.
func (n *sortNode) Explain() (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

func (n *sortNode) Next() (bool, error) {
	for n.needSort {
		// make sure our sortStrategy is initialized
		if n.sortStrategy == nil {
			v := n.p.newContainerValuesNode(n.ordering)
			n.sortStrategy = newAllSortStrategy(v)
		}

		// consume data (from plan) (Next / Values())
		next, err := n.plan.Next()
		if err != nil {
			return false, err
		}
		if !next {
			n.sortStrategy.Finish()
			n.valueIter = n.sortStrategy
			n.needSort = false
			break
		}

		// consuming data, sort
		if err := n.sortStrategy.Add(n.plan.Value()); err != nil {
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

func (n *sortNode) Close() error {
	err := n.plan.Close()
	if err != nil {
		return err
	}

	if n.valueIter != nil {
		return n.valueIter.Close()
	}

	if n.sortStrategy != nil {
		return n.sortStrategy.Close()
	}
	return nil
}

func (n *sortNode) Source() planNode { return n.plan }

// allSortStrategy is the simplest sort strategy available.
// it consumes all the data into the underlying valueNode
// document container, then sorts it. Its designed for an
// unknown number of records.
type allSortStrategy struct {
	valueNode *valuesNode
}

func newAllSortStrategy(v *valuesNode) *allSortStrategy {
	return &allSortStrategy{
		valueNode: v,
	}
}

// Add adds a new document to underlying valueNode
func (s *allSortStrategy) Add(doc core.Doc) error {
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
func (s *allSortStrategy) Value() core.Doc {
	return s.valueNode.Value()
}

// Close closes the underling valueNode
func (s *allSortStrategy) Close() error {
	return s.valueNode.Close()
}
