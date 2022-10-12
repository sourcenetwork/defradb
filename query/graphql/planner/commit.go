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
	"github.com/sourcenetwork/defradb/query/graphql/mapper"
)

// commitSelectTopNode is a wrapper for the selectTopNode
// in the case where the select is actually a CommitSelect
type commitSelectTopNode struct {
	docMapper

	p    *Planner
	plan planNode
}

func (n *commitSelectTopNode) Kind() string { return "commitSelectTopNode" }

func (n *commitSelectTopNode) Init() error { return n.plan.Init() }

func (n *commitSelectTopNode) Start() error { return n.plan.Start() }

func (n *commitSelectTopNode) Next() (bool, error) { return n.plan.Next() }

func (n *commitSelectTopNode) Spans(spans core.Spans) { n.plan.Spans(spans) }

func (n *commitSelectTopNode) Value() core.Doc { return n.plan.Value() }

func (n *commitSelectTopNode) Source() planNode { return n.plan }

func (n *commitSelectTopNode) Close() error {
	if n.plan == nil {
		return nil
	}
	return n.plan.Close()
}

func (n *commitSelectTopNode) Append() bool { return true }

func (p *Planner) CommitSelect(parsed *mapper.CommitSelect) (planNode, error) {
	dagScan := p.DAGScan(parsed)

	plan, err := p.SelectFromSource(&parsed.Select, dagScan, false, nil)
	if err != nil {
		return nil, err
	}
	return &commitSelectTopNode{
		p:         p,
		plan:      plan,
		docMapper: docMapper{&parsed.DocumentMapping},
	}, nil
}
