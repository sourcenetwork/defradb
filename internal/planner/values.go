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
	"sort"

	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/db/base"
	"github.com/sourcenetwork/defradb/internal/db/container"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

// valuesNode contains a collection
// of documents as values inside a document
// container. It implements the planNode
// interface, but is used slightly differently
// then the rest of the nodes in the graph. It
// has no children planNodes.
type valuesNode struct {
	docMapper

	p *Planner
	// plan planNode

	ordering []mapper.OrderCondition

	docs     *container.DocumentContainer
	docIndex int
}

func (p *Planner) newContainerValuesNode(ordering []mapper.OrderCondition) *valuesNode {
	return &valuesNode{
		p:        p,
		ordering: ordering,
		docs:     container.NewDocumentContainer(0),
		docIndex: -1,
	}
}

func (n *valuesNode) Init() error            { return nil }
func (n *valuesNode) Start() error           { return nil }
func (n *valuesNode) Spans(spans core.Spans) {}

func (n *valuesNode) Kind() string {
	return "valuesNode"
}

func (n *valuesNode) Close() error {
	if n.docs != nil {
		n.docs.Close()
	}
	return nil
}

func (n *valuesNode) Next() (bool, error) {
	if n.docIndex >= n.docs.Len()-1 {
		return false, nil
	}
	n.docIndex++
	return true, nil
}

func (n *valuesNode) Value() core.Doc {
	return n.docs.At(n.docIndex)
}

func (n *valuesNode) Source() planNode { return nil }

// SortAll actually sorts all the data within the docContainer object
func (n *valuesNode) SortAll() {
	sort.Stable(n)
}

// Less implements the golang sort.Interface.
// Less reports whether the elements within the docContainer at index i must sort before the element with index j.
// Returns true if docs[i] < docs[j].
// Returns false if docs[i] >= docs[j].
// If both Less(i, j) and Less(j, i) are false, then the elements at index i and j are considered equal.
func (n *valuesNode) Less(i, j int) bool {
	da, db := n.docs.At(i), n.docs.At(j)
	return n.docValueLess(da, db)
}

// docValueLess extracts and compare field values of a document, returns true only if strictly less when ASC,
// and true if greater than or equal when DESC, otherwise returns false.
func (n *valuesNode) docValueLess(docA, docB core.Doc) bool {
	for _, order := range n.ordering {
		compare := base.Compare(
			getDocProp(docA, order.FieldIndexes),
			getDocProp(docB, order.FieldIndexes),
		)

		if order.Direction == mapper.DESC {
			if compare > 0 {
				return true
			} else {
				return false
			}
		} else { // Otherwise assume order.Direction == mapper.ASC
			if compare < 0 {
				return true
			} else {
				return false
			}
		}
	}
	return false
}

// Swap implements the golang sort.Sort interface.
// It swaps the values at the ith and jth index
// within the docContainer.
func (n *valuesNode) Swap(i, j int) {
	n.docs.Swap(i, j)
}

// Len implements the golang sort.Sort interface.
// returns the size of the internal document
// container
func (n *valuesNode) Len() int {
	return n.docs.Len()
}

// getMapProp is a utility to easily get a specific
// property from a map object. The map may have further nested maps
// that need to be accessed.
// The prop argument has the entire selection of keys to grab in the
// case of nested objects. The key delimeter is a ".".
// Eg.
// prop = "author.name" -> {author: {name: ...}}
func getDocProp(obj core.Doc, prop []int) any {
	if len(prop) == 0 {
		return nil
	}
	return getMapPropList(obj, prop)
}

func getMapPropList(obj core.Doc, props []int) any {
	if len(props) == 1 {
		return obj.Fields[props[0]]
	}

	val := obj.Fields[props[0]]
	if val == nil {
		return nil
	}
	subObj, ok := val.(core.Doc)
	if !ok {
		return nil
	}
	return getMapPropList(subObj, props[1:])
}
