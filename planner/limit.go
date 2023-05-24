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
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/planner/mapper"
)

// Limit the results, yielding only what the limit/offset permits
// @todo: Handle cursor
type limitNode struct {
	docMapper

	p    *Planner
	plan planNode

	limit    uint64
	offset   uint64
	rowIndex uint64

	execInfo limitExecInfo
}

type limitExecInfo struct {
	// Total number of times limitNode was executed.
	iterations uint64
}

// Limit creates a new limitNode initalized from the parser.Limit object.
func (p *Planner) Limit(parsed *mapper.Select, n *mapper.Limit) (*limitNode, error) {
	if n == nil {
		return nil, nil // nothing to do
	}
	return &limitNode{
		p:         p,
		limit:     n.Limit,
		offset:    n.Offset,
		rowIndex:  0,
		docMapper: docMapper{parsed.DocumentMapping},
	}, nil
}

func (n *limitNode) Kind() string {
	return "limitNode"
}

func (n *limitNode) Init() error {
	n.rowIndex = 0
	return n.plan.Init()
}

func (n *limitNode) Start() error           { return n.plan.Start() }
func (n *limitNode) Spans(spans core.Spans) { n.plan.Spans(spans) }
func (n *limitNode) Close() error           { return n.plan.Close() }
func (n *limitNode) Value() core.Doc        { return n.plan.Value() }

func (n *limitNode) Next() (bool, error) {
	n.execInfo.iterations++

	// check if we're passed the limit
	if n.limit != 0 && n.rowIndex >= n.limit+n.offset {
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

func (n *limitNode) Source() planNode { return n.plan }

func (n *limitNode) simpleExplain() (map[string]any, error) {
	simpleExplainMap := map[string]any{
		limitLabel:  n.limit,
		offsetLabel: n.offset,
	}

	if n.limit == 0 {
		simpleExplainMap[limitLabel] = nil
	}

	return simpleExplainMap, nil
}

func (n *limitNode) Explain(explainType request.ExplainType) (map[string]any, error) {
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
