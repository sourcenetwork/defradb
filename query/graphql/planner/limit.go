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
	"math"

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/query/graphql/mapper"
	parserTypes "github.com/sourcenetwork/defradb/query/graphql/parser/types"
)

// Limit the results, yielding only what the limit/offset permits
// @todo: Handle cursor
type hardLimitNode struct {
	docMapper

	p    *Planner
	plan planNode

	limit    int64
	offset   int64
	rowIndex int64
}

// HardLimit creates a new hardLimitNode initalized from the parser.Limit object.
func (p *Planner) HardLimit(parsed *mapper.Select, n *mapper.Limit) (*hardLimitNode, error) {
	if n == nil {
		return nil, nil // nothing to do
	}
	limit := int64(math.MaxInt64)
	if n.Limit > 0 {
		limit = n.Limit
	}
	return &hardLimitNode{
		p:         p,
		limit:     limit,
		offset:    n.Offset,
		rowIndex:  0,
		docMapper: docMapper{&parsed.DocumentMapping},
	}, nil
}

func (n *hardLimitNode) Kind() string {
	return "hardLimitNode"
}

func (n *hardLimitNode) Init() error {
	n.rowIndex = 0
	return n.plan.Init()
}

func (n *hardLimitNode) Start() error           { return n.plan.Start() }
func (n *hardLimitNode) Spans(spans core.Spans) { n.plan.Spans(spans) }
func (n *hardLimitNode) Close() error           { return n.plan.Close() }
func (n *hardLimitNode) Value() core.Doc        { return n.plan.Value() }

func (n *hardLimitNode) Next() (bool, error) {
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

func (n *hardLimitNode) Source() planNode { return n.plan }

func (n *hardLimitNode) Explain() (map[string]interface{}, error) {
	return map[string]interface{}{
		limitLabel:  n.limit,
		offsetLabel: n.offset,
	}, nil
}

// limit the results, flagging any records outside the bounds of limit/offset with
// with a 'hidden' flag blocking rendering.  Used if consumers of the results require
// the full dataset.
type renderLimitNode struct {
	documentIterator
	docMapper

	p    *Planner
	plan planNode

	limit    int64
	offset   int64
	rowIndex int64
}

// RenderLimit creates a new renderLimitNode initalized from
// the parser.Limit object.
func (p *Planner) RenderLimit(docMap *core.DocumentMapping, n *parserTypes.Limit) (*renderLimitNode, error) {
	if n == nil {
		return nil, nil // nothing to do
	}
	limit := int64(math.MaxInt64)
	if n.Limit > 0 {
		limit = n.Limit
	}
	return &renderLimitNode{
		p:         p,
		limit:     limit,
		offset:    n.Offset,
		rowIndex:  0,
		docMapper: docMapper{docMap},
	}, nil
}

func (n *renderLimitNode) Kind() string {
	return "renderLimitNode"
}

func (n *renderLimitNode) Init() error {
	n.rowIndex = 0
	return n.plan.Init()
}

func (n *renderLimitNode) Start() error           { return n.plan.Start() }
func (n *renderLimitNode) Spans(spans core.Spans) { n.plan.Spans(spans) }
func (n *renderLimitNode) Close() error           { return n.plan.Close() }

func (n *renderLimitNode) Next() (bool, error) {
	if next, err := n.plan.Next(); !next {
		return false, err
	}

	n.currentValue = n.plan.Value()

	n.rowIndex++
	if n.rowIndex-n.offset > n.limit || n.rowIndex <= n.offset {
		n.currentValue.Hidden = true
	}
	return true, nil
}

func (n *renderLimitNode) Source() planNode { return n.plan }

func (n *renderLimitNode) Explain() (map[string]interface{}, error) {
	return map[string]interface{}{
		limitLabel:  n.limit,
		offsetLabel: n.offset,
	}, nil
}
