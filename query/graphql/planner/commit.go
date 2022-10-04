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

	cid "github.com/ipfs/go-cid"

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/errors"
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
	// check type of commit select (all, latest, one)
	var commit *commitSelectNode
	var err error
	switch parsed.Type {
	case mapper.LatestCommits:
		commit, err = p.commitSelectLatest(parsed)
	case mapper.OneCommit:
		commit, err = p.commitSelectBlock(parsed)
	case mapper.AllCommits:
		commit, err = p.commitSelectAll(parsed)
	default:
		return nil, errors.New("Invalid CommitSelect type")
	}
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

// commitSelectLatest is a CommitSelect node initalized with a headsetScanNode and a DocKey
func (p *Planner) commitSelectLatest(parsed *mapper.CommitSelect) (*commitSelectNode, error) {
	dag := p.DAGScan(parsed)
	headset := p.HeadScan(parsed)
	// @todo: Get Collection field ID
	if !parsed.FieldName.HasValue() {
		dag.field = core.COMPOSITE_NAMESPACE
	} else {
		dag.field = parsed.FieldName.Value()
	}

	if parsed.DocKey != "" {
		key := core.DataStoreKey{}.WithDocKey(parsed.DocKey).WithFieldId(dag.field)
		headset.key = key
	}
	dag.headset = headset
	commit := &commitSelectNode{
		p:      p,
		source: dag,
	}

	return commit, nil
}

// commitSelectBlock is a CommitSelect node initialized without a headsetScanNode, and is expected
// to be given a target CID in the parser.CommitSelect object. It returns a single commit if found.
func (p *Planner) commitSelectBlock(parsed *mapper.CommitSelect) (*commitSelectNode, error) {
	dag := p.DAGScan(parsed)
	if parsed.Cid != "" {
		c, err := cid.Decode(parsed.Cid)
		if err != nil {
			return nil, err
		}
		dag.cid = &c
	} // @todo: handle error if no CID is given

	return &commitSelectNode{
		p:      p,
		source: dag,
	}, nil
}

// commitSelectAll is a CommitSelect initialized with a headsetScanNode, and will
// recursively return all graph commits in order.
func (p *Planner) commitSelectAll(parsed *mapper.CommitSelect) (*commitSelectNode, error) {
	dag := p.DAGScan(parsed)
	headset := p.HeadScan(parsed)

	if parsed.Cid != "" {
		c, err := cid.Decode(parsed.Cid)
		if err != nil {
			return nil, err
		}
		dag.cid = &c
	}

	// @todo: Get Collection field ID
	if !parsed.FieldName.HasValue() {
		dag.field = core.COMPOSITE_NAMESPACE
	} else {
		dag.field = parsed.FieldName.Value()
	}

	if parsed.DocKey != "" {
		key := core.DataStoreKey{}.WithDocKey(parsed.DocKey)

		if parsed.FieldName.HasValue() {
			field := parsed.FieldName.Value()
			key = key.WithFieldId(field)
			dag.field = field
		}

		headset.key = key
	}
	dag.headset = headset
	dag.depthLimit = math.MaxUint32 // infinite depth
	// dag.key = &key
	commit := &commitSelectNode{
		p:      p,
		source: dag,
	}

	return commit, nil
}
