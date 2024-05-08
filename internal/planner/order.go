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
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

// simplified planNode interface.
// Contains only the methods involved
// in value generation and retrieval.
type valueIterator interface {
	Next() (bool, error)
	Value() core.Doc
	Close() error
}

type orderingStrategy interface {
	valueIterator
	// Add a document to the strategy node.
	// copies data if its needed.
	// Ideally stores inside a valuesNode
	// rowContainer buffer.
	Add(core.Doc) error
	// Finish finalizes and applies the actual
	// ordering mechanism to all the stored data.
	Finish()
}

// order the results
type orderNode struct {
	docMapper

	p    *Planner
	plan planNode

	ordering []mapper.OrderCondition

	// simplified planNode interface
	// used for iterating through
	// an already sorted plan
	valueIter valueIterator

	// orderStrategy is an encapsulate planNode
	// that sorts, then provides the values
	// sorted
	orderStrategy orderingStrategy

	// indicates if our underlying orderStrategy is still
	// consuming and sorting data.
	needSort bool

	execInfo orderExecInfo
}

type orderExecInfo struct {
	// Total number of times orderNode was executed.
	iterations uint64
}

// OrderBy creates a new orderNode which returns the underlying
// plans values in a sorted mannor. The field to sort by, and the
// direction of ordering is determined by the given mapper.OrderBy
// object.
func (p *Planner) OrderBy(parsed *mapper.Select, n *mapper.OrderBy) (*orderNode, error) {
	if n == nil { // no orderby info
		return nil, nil
	}

	return &orderNode{
		p:         p,
		ordering:  n.Conditions,
		needSort:  true,
		docMapper: docMapper{parsed.DocumentMapping},
	}, nil
}

func (n *orderNode) Kind() string {
	return "orderNode"
}

func (n *orderNode) Init() error {
	// reset stateful data
	n.needSort = true
	n.orderStrategy = nil
	return n.plan.Init()
}
func (n *orderNode) Start() error { return n.plan.Start() }

func (n *orderNode) Spans(spans core.Spans) { n.plan.Spans(spans) }

func (n *orderNode) Value() core.Doc {
	return n.valueIter.Value()
}

func (n *orderNode) simpleExplain() (map[string]any, error) {
	orderings := []map[string]any{}

	for _, element := range n.ordering {
		// Build the list containing the corresponding names of all the indexes.
		fieldNames := []string{}

		mapping := n.documentMapping
		for _, fieldIndex := range element.FieldIndexes {
			fieldName, found := mapping.TryToFindNameFromIndex(fieldIndex)
			if !found {
				return nil, client.NewErrFieldIndexNotExist(fieldIndex)
			}

			fieldNames = append(fieldNames, fieldName)
			if fieldIndex < len(mapping.ChildMappings) {
				if childMapping := mapping.ChildMappings[fieldIndex]; childMapping != nil {
					mapping = childMapping
				}
			}
		}

		// Put it all together for this order element.
		orderings = append(orderings,
			map[string]any{
				"fields":    fieldNames,
				"direction": string(element.Direction),
			},
		)
	}

	return map[string]any{
		"orderings": orderings,
	}, nil
}

// Explain method returns a map containing all attributes of this node that
// are to be explained, subscribes / opts-in this node to be an explainablePlanNode.
func (n *orderNode) Explain(explainType request.ExplainType) (map[string]any, error) {
	switch explainType {
	case request.SimpleExplain:
		return n.simpleExplain()

	case request.ExecuteExplain:
		return map[string]any{
			"iterations": n.execInfo.iterations,
		}, nil

	default:
		return nil, ErrUnknownExplainRequestType
	}
}

func (n *orderNode) Next() (bool, error) {
	n.execInfo.iterations++

	for n.needSort {
		// make sure our orderStrategy is initialized
		if n.orderStrategy == nil {
			v := n.p.newContainerValuesNode(n.ordering)
			n.orderStrategy = newAllSortStrategy(v)
		}

		// consume data (from plan) (Next / Values())
		next, err := n.plan.Next()
		if err != nil {
			return false, err
		}
		if !next {
			n.orderStrategy.Finish()
			n.valueIter = n.orderStrategy
			n.needSort = false
			break
		}

		// consuming data, sort
		if err := n.orderStrategy.Add(n.plan.Value()); err != nil {
			return false, err
		}
	}

	next, err := n.valueIter.Next()
	if !next {
		return false, err
	}
	return true, nil
}

func (n *orderNode) Close() error {
	err := n.plan.Close()
	if err != nil {
		return err
	}

	if n.valueIter != nil {
		return n.valueIter.Close()
	}

	if n.orderStrategy != nil {
		return n.orderStrategy.Close()
	}
	return nil
}

func (n *orderNode) Source() planNode { return n.plan }

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
	s.valueNode.docs.AddDoc(doc)
	return nil
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
