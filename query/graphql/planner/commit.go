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
	cid "github.com/ipfs/go-cid"

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

type commitSelectNode struct {
	documentIterator
	docMapper

	p *Planner

	source *dagScanNode
}

func (n *commitSelectNode) Kind() string {
	return "commitSelectNode"
}

func (n *commitSelectNode) Init() error {
	return n.source.Init()
}

func (n *commitSelectNode) Start() error {
	return n.source.Start()
}

func (n *commitSelectNode) Next() (bool, error) {
	if next, err := n.source.Next(); !next {
		return false, err
	}

	n.currentValue = n.source.Value()
	cid, hasCid := n.docMapper.DocumentMap().FirstOfName(n.currentValue, "cid").(*cid.Cid)
	if hasCid {
		// dagScanNode yields cids, but we want to yield strings
		n.docMapper.DocumentMap().SetFirstOfName(&n.currentValue, "cid", cid.String())
	}

	return true, nil
}

func (n *commitSelectNode) Spans(spans core.Spans) {
	n.source.Spans(spans)
}

func (n *commitSelectNode) Close() error {
	return n.source.Close()
}

func (n *commitSelectNode) Source() planNode {
	return n.source
}

// Explain method returns a map containing all attributes of this node that
// are to be explained, subscribes / opts-in this node to be an explainablePlanNode.
func (n *commitSelectNode) Explain() (map[string]any, error) {
	return map[string]any{}, nil
}

func (p *Planner) CommitSelect(parsed *mapper.CommitSelect) (planNode, error) {
	commit, err := p.buildCommitSelectNode(parsed)
	if err != nil {
		return nil, err
	}

	plan, err := p.SelectFromSource(&parsed.Select, commit, false, nil)
	if err != nil {
		return nil, err
	}
	return &commitSelectTopNode{
		p:         p,
		plan:      plan,
		docMapper: docMapper{&parsed.DocumentMapping},
	}, nil
}

func (p *Planner) buildCommitSelectNode(parsed *mapper.CommitSelect) (*commitSelectNode, error) {
	dag := p.DAGScan(parsed)

	// @todo: Get Collection field ID
	if !parsed.FieldName.HasValue() {
		dag.field = core.COMPOSITE_NAMESPACE
	} else {
		dag.field = parsed.FieldName.Value()
	}

	// dag.key = &key
	commit := &commitSelectNode{
		p:         p,
		source:    dag,
		docMapper: docMapper{&parsed.DocumentMapping},
	}

	return commit, nil
}
