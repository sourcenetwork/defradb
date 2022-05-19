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
	"strings"

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/db/container"
	"github.com/sourcenetwork/defradb/query/graphql/parser"

	parserTypes "github.com/sourcenetwork/defradb/query/graphql/parser/types"
)

// valuesNode contains a collection
// of documents as values inside a document
// container. It implements the planNode
// interface, but is used slightly differently
// then the rest of the nodes in the graph. It
// has no children planNodes.
type valuesNode struct {
	p *Planner
	// plan planNode

	ordering []parser.SortCondition

	docs     *container.DocumentContainer
	docIndex int
}

func (p *Planner) newContainerValuesNode(ordering []parser.SortCondition) *valuesNode {
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
func (n *valuesNode) Close() {
	if n.docs != nil {
		n.docs.Close()
	}
}

func (n *valuesNode) Next() (bool, error) {
	if n.docIndex >= n.docs.Len()-1 {
		return false, nil
	}
	n.docIndex++
	return true, nil
}

func (n *valuesNode) Value() map[string]interface{} {
	return n.docs.At(n.docIndex)
}

func (n *valuesNode) Source() planNode { return nil }

// SortAll actually sorts all the data within the docContainer object
func (n *valuesNode) SortAll() {
	sort.Sort(n)
}

// Less implements the golang sort.Sort interface.
// It compares the values the ith and jth index
// within the docContainer.
// returns true if i < j.
// returns false if i > j.
func (n *valuesNode) Less(i, j int) bool {
	da, db := n.docs.At(i), n.docs.At(j)
	return n.docValueLess(da, db)
}

// docValueLess extracts and compare field values of a document
func (n *valuesNode) docValueLess(da, db map[string]interface{}) bool {
	var ra, rb interface{}
	for _, order := range n.ordering {
		if order.Direction == parserTypes.ASC {
			ra = getMapProp(da, order.Field)
			rb = getMapProp(db, order.Field)
		} else if order.Direction == parserTypes.DESC { // redundant, just else
			ra = getMapProp(db, order.Field)
			rb = getMapProp(da, order.Field)
		}

		if c := base.Compare(ra, rb); c < 0 {
			return true
		} else if c > 0 {
			return false
		}
	}

	return true
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
func getMapProp(obj map[string]interface{}, prop string) interface{} {
	if prop == "" {
		return nil
	}
	props := strings.Split(prop, ".")
	numProps := len(props)
	return getMapPropList(obj, props, numProps)
}

func getMapPropList(obj map[string]interface{}, props []string, numProps int) interface{} {
	if numProps == 1 {
		val, ok := obj[props[0]]
		if !ok {
			return nil
		}
		return val
	}

	val, ok := obj[props[0]]
	if !ok {
		return nil
	}
	subObj, ok := val.(map[string]interface{})
	if !ok {
		return nil
	}
	return getMapPropList(subObj, props[1:], numProps-1)
}
