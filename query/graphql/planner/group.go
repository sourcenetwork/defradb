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
	"github.com/sourcenetwork/defradb/query/graphql/parser"
)

// A node responsible for the grouping of documents by a given selection of fields.
type groupNode struct {
	p *Planner

	// The child select information.  Will be nil if there is no child `_group` item requested.
	childSelect *parser.Select

	// The fields to group by - this must be an ordered collection and
	// will include any parent group-by fields (if any)
	groupByFields []string

	dataSource dataSource

	values       []map[string]interface{}
	currentIndex int
	currentValue map[string]interface{}
}

// Creates a new group node.  The function is recursive and will construct the node-chain for any child (`_group`) collections.
// `groupSelect` is optional and will typically be nil if the child `_group` is not requested.
func (p *Planner) GroupBy(n *parser.GroupBy, childSelect *parser.Select) (*groupNode, error) {
	if n == nil {
		return nil, nil
	}

	if childSelect != nil && childSelect.GroupBy != nil {
		// group by fields have to be propagated downwards to ensure correct sub-grouping, otherwise child
		// groups will only group on the fields they explicitly reference
		childSelect.GroupBy.Fields = append(childSelect.GroupBy.Fields, n.Fields...)
	}

	groupNodeObj := groupNode{
		p:             p,
		childSelect:   childSelect,
		groupByFields: n.Fields,
		dataSource:    newDataSource(parser.GroupFieldName),
	}
	return &groupNodeObj, nil
}

func (n *groupNode) Init() error {
	// We need to make sure state is cleared down on Init,
	// this function may be called multiple times per instance (for example during a join)
	n.values = nil
	n.currentValue = nil
	n.currentIndex = 0
	return n.dataSource.Init()
}

func (n *groupNode) Start() error           { return n.dataSource.Start() }
func (n *groupNode) Spans(spans core.Spans) { n.dataSource.Spans(spans) }
func (n *groupNode) Close() error           { return n.dataSource.Close() }
func (n *groupNode) Source() planNode       { return n.dataSource.Source() }

func (n *groupNode) Values() map[string]interface{} {
	return n.currentValue
}

func (n *groupNode) Next() (bool, error) {
	if n.values == nil {
		values, err := join([]dataSource{n.dataSource}, n.groupByFields)
		if err != nil {
			return false, err
		}

		n.values = values.values
	}

	if n.currentIndex < len(n.values) {
		n.currentValue = n.values[n.currentIndex]
		n.currentIndex++
		return true, nil
	}

	return false, nil
}
